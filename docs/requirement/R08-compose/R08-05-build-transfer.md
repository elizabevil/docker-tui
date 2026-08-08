# R08-05 镜像构建与传输

## 元信息

- 状态: planned
- 优先级: medium
- 来源: R08-04 引用 + 用户评审结果
- 关联任务: [.omo/compose-todo.md 附录 A.3 / §3.3 / 补 F.1](../../../.omo/compose-todo.md)
- 关联约束:
  - [../../constraint/C01-form.md](../../constraint/C01-form.md)
  - [../../constraint/C02-dialog.md](../../constraint/C02-dialog.md)
  - [../../constraint/C05-path.md](../../constraint/C05-path.md)

## ⚠ 重复功能说明

| Compose 命令 | 已实现层 | 本需求增量 |
|---|---|---|
| `build` | R02 镜像 `imageBuild`(历史 feat-design §3 明确取消) | **R08 重审后采纳** R08-05,Compose-级 build = 服务级循环 build |
| `pull` | R02 镜像 `imagePull` | 增量:Compose-级按 service 列表循环 pull,进度聚合报告 |
| `push` | R02 镜像 `imagePush` | 增量:按 service 列表循环 push,标签已就绪时**不**重打 tag |
| `compose images` | R02 镜像列表 | 不在 Compose-级重复展示 |

## 目标

补 R08-05 实现 Compose-级 build / pull / push。规格:

1. **build**:逐 service 构建 `service.image`(若 `build:` 段存在),否则(纯 image 引用)跳过。每个 service 输出独立进度 + 行末错误汇总。本需求**不**解析 yaml,先靠 `ComposeService.Inspect` 拿到的 service 级 metadata,然后仅对"service.image 在本地可见"的容器聚合,**不**调用 docker build CLI;**对**确实是 Dockerfile 构建的服务,**返回 ErrComposeUnsupported**(因为没 yaml 解析)
2. **pull**:直接用 service.image 名走 `imagePull`。逐 service 输出进度 + 总体进度条
3. **push**:仅当 service.image 已经 push 过(本地有 `.docker/registry` 痕迹)或者用户在 Form 里显式补 tag 时启用

## 用户流程

### build

1. Compose 面板 → Action Bar (`;`) → Build,或快捷键 `Ctrl+B`(注册在 R08-12)
2. 弹出 Confirm Dialog(可不需 Form,但显示"此项目 N 个 service 中 M 个将构建"):`Build <project>?`
3. 用户 OK → 进度 toast,逐 service 进度
4. 含 `build:` 段的服务返回 ErrComposeUnsupported,toast:`⚠ <service>: requires compose CLI,请用 docker compose CLI`
5. 其它服务按 image 拉取策略,等价 pull

### pull

1. 选中项目,Action Bar → Pull,快捷键 `Ctrl+P`(与单镜像 pull 同键,scope 控制)
2. Form 简化为"全部 / 限定 service"
3. 进度按 service 列表展示

### push

1. Action Bar → Push,快捷键 `Ctrl+U`(与单镜像 push 同键)
2. Form:逐 service list(checkbox),用户限定提交范围
3. Push 完成后 toast:`✓ X pushed, Y skipped (no auth), Z failed`

## UI/UX

仅本需求特有:

- 进度显示:during 操作 footer 显示 `[项目名 · build · 3/8]`,失败的 service 名称灰显
- Confirm Dialog(C02 中等尺寸)
- 用户限定 service 时:弹 Sub-Dialog(列 service 列表 + 复选框,**复用 R03-03 框架**)

完整版式在 UI overhaul 阶段。

## 功能规则

### F1 build 与 R02 imageBuild 的边界

| 维度 | R02 imageBuild | R08-05 composeBuild |
|---|---|---|
| 入口 | 单镜像:有 / 列表 / Form | 整个项目(per-service) |
| 是否解析 yaml | 是 | 否 |
| 是否支持 `build:` | 是 | 否(返回 Err) |
| 用户场景 | 容器开发:改 Dockerfile 后构建 | 项目级 batch 构建 |

故 R08-05 的 `build` 在语义上**不是 R02 imageBuild 的覆盖**,而是**并行的简化入口**:compose 层不解析 yaml,只对"image 已能等价 pull"的 service 起作用;真正的 Dockerfile 构建仍走 docker compose CLI。

### F2 pull 增量

Compose-级 pull 等价"对每个 service 走一次 R02 imagePull",但有以下增量:

- **进度聚合**:每个 service 独立报告 bytes pulled / total,总进度为 sum / max
- **失败模式**:单个 service 拉取失败不阻塞其它 service,toast 末尾汇总
- **重复检测**:同 image 被多个 service 引用,只拉一次(去重)

### F3 push 增量

Compose-级 push 等价"对每个 service 走一次 R02 imagePush",但有:

- **范围限定**:默认全 service,可 Form 限定子集
- **registry 检查**:无认证的 service 跳过 + toast 提示
- **增量同步**:如果 image tag 与目标 tag 不同,先提示用户决定(沿用 R02 的 tag Form)

### F4 audit

每个 service 一条 audit trace,action `compose_service.build` / `compose_service.pull` / `compose_service.push`。汇总 toast 显示在 UI 但 audit 不汇总。

## 实现设计

涉及模块:

- 新方法:`ComposeService.Build(ctx, project, opts)` / `Pull` / `Push`(R08-02)
- 修改:`internal/tui/keyboard/compose_action.go`(doComposeBuild / doComposePull / doComposePush 新增,或经 R08-11 Action Bar 接管)
- 新文件:`internal/tui/ui/pages/compose/build_form.go` — 选项 Form(数量 / service list / force 选项)
- 修改:`internal/data/runtime/image.go`(R02:`imageBuild` / `imagePull` / `imagePush`)的复用
- audit:`internal/tui/keyboard/audit.go` 加 3 个新动作

## 验收标准

1. 项目的 service 列表空 / 仅有 build: 段 → 返回 `ErrComposeUnsupported` + UI 明确提示
2. pull 在 `pull 1` (单 image)与 `compose pull`(批)行为一致,统计粒度不同
3. push 在 Form 限定 → 仅提交选中的 service;未限定 → 全服务
4. Action Bar scope 注册:`OperationScopeCompose` 含 build / pull / push 三项
5. 每个 service 是独立 audit,失败 service 进入 audit history 列表

## 非目标

- 解析 compose yaml 中 build 段(R08-02 决策 B 不做)
- 自动镜像重打 tag(由用户在 Form 中决定,沿用 R02 行为)
- 支持 image buildx / multi-stage / multi-platform(超过本项目范围)
- 镜像缓存清理 / `image prune` 的 compose 级联(走 R02 镜像级)

## 迁移记录

- 旧决策:`docs/feature-design.md §3` Image Build 取消;`feature-todo-list.md §3` Compose up/build/pull 取消
- 重新提出:R08-05 由用户评审改判为采纳
- 保留信息:R02 imagePull / imagePush 实现(Compose-级复用)
- 废弃信息:取消决策(由本需求撤销)
- 待确认:用户评审后是否同时撤销 `docs/feature-design.md §3` Image Build 取消条款 — 留主模型决策
