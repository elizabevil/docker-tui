# R08-03 项目 / 服务 / 配置发现

## 元信息

- 状态: planned
- 优先级: high
- 来源: R08-02 候选 + 用户评审结果
- 关联任务: [.omo/compose-todo.md 附录 B.2 / §3.2 / 补 G.3 / 补 G.6](../../../.omo/compose-todo.md)
- 关联约束:
  - [../../constraint/C03-table.md](../../constraint/C03-table.md)
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)

## ⚠ 重复功能说明

| 能力 | 已实现层 | 本需求增量 |
|---|---|---|
| `docker compose ps` | R01-容器-列表(从容器 label 聚合) | 本需求统一为 `ComposeService.ListProjects`(带 Source 字段) |
| `docker compose ls` | 无空项目支持 | **与 CLI 一致:无容器则隐藏**(2026-08-08 已定,见 §决策记录) |
| `docker compose config` | 无 | R08-02 决策 B:`Config` 返回 `ErrComposeUnsupported`,本需求给出**替代方案**:`WorkingDir / ConfigFiles / Version` 元信息 |
| `docker compose images` | R02 镜像列表 | 不在 Compose-级重复实现 |

## 目标

`ComposeService.ListProjects / InspectProject` 项目级发现能力 + 仅元信息(`WorkingDir / ConfigFiles / Version`)的轻量"config 替代"能力。具体能力分解:

1. 当前 UI 只能看到"有容器的项目"。空项目(已 down 完的)在不被任何容器体现时,**不可见** —— 与 `docker compose ls` 权威行为一致,本需求**保留**此语义,不补空项目来源。
2. `docker compose config` 真正需要 yaml 解析,R08-02 决策 B 不实现。本需求提供"只读元信息"替代,足够 dtui 用户判断"这个项目有没有 working_dir、有没有 compose 文件、哪份 compose 文件"。
3. 服务级信息(每个 service 的 image / replicas / 状态)从 label + 容器聚合,不重新解析 yaml。

## 用户流程

1. Compose 面板左栏(`gatherComposeProjects` 渲染处)显示 "项目名 ⓘ"(`ⓘ` icon 触发详情)
2. Enter 项目 → 右栏(`inspect` 视图)显示:
   - 项目基础:Name / Source / HasRunning / HasStopped
   - 元信息:WorkingDir / ConfigFiles(逐行,点击复制到剪贴板)/ Version
   - 服务列表(image / replicas running/stopped)
   - 一行附加"compose YAML 不可读,完整 config 需 `docker compose config`"(R08-02 决策 B 透明)
3. 返回(`Esc`)→ 左栏

## UI/UX

只写本需求特有的部分(占位 — UI 详细设计留到 R08-12 之后的 UI overhaul):

- 右栏 inspect 标题:`Compose Detail: <project>`(已与 BR-017 修复保持一致)
- 元信息空值降级显示"(unknown)"而非空白
- ConfigFiles 多行展示在 sub-table 区,宽屏拼接 / 窄屏折叠
- 空项目态:不显示(与 `docker compose ls` 语义一致,见 §决策记录)

详细版式在 UI overhaul 阶段统一治理,本需求**只列出**字段与状态。

## 功能规则

### F1 ProjectSummary 聚合

```go
// 来自 R08-02,但本需求细化聚合顺序
type ComposeProjectSummary struct {
    Name         string
    Source       string // "docker" | "podman" | "aggregated"
    WorkingDir   string
    ConfigFiles  []string
    Version      string
    CreatedAt    time.Time
    Services     []ComposeServiceSummary
    HasRunning   bool   // 至少 1 个 running 容器
    HasStopped   bool   // 至少 1 个 stopped 容器
}
```

### F2 空项目语义(2026-08-08 已定:无容器则隐藏)

- **完全无容器:不显示**(无 `IsOrphan` / `(orphan)` 概念,与 `docker compose ls` 一致 —— 容器 label 是项目存在的唯一载体)
- 有容器但所有 stopped:`HasStopped=true, HasRunning=false`,正常显示(需 `ListProjectsOptions{All:true}`,见 §决策记录)

### F3 WorkingDir / ConfigFiles 来源

- Docker:从 `com.docker.compose.project.working_dir` / `com.docker.compose.project.config_files` label 读取
- Podman:无此 label,WorkingDir 与 ConfigFiles 均**留空**(R08-01 §F4)
- 解析:`ConfigFiles` 按 `\\n` 拆分(`yaml\\nyaml.override.yml`),过滤空行
- 缺失时 UI 展示 `(unknown)` + 帮助信息

### F4 Version 字段

- 仅 Docker 侧:`com.docker.compose.version` 例如 `"2.21.0"`
- Podman 侧无对应 label,留空 + 标记 `Source=podman`

### F5 Content-Hash 与 Oneoff

`com.docker.compose.config-hash` 与 `com.docker.compose.oneoff` 暂不暴露给 UI(留给 R08-04 a future 增强用):

- config-hash 用于"配置是否变了"检测(用户重新编辑 compose 后,dtui 是否重读)→ 本需求不实现"刷新元信息"
- oneoff 标记一次性容器(`docker compose run` 创建)→ 仅在 R08-08 用到

## 实现设计

涉及模块:

- 新文件:`internal/data/runtime/compose_inspect.go` — `composeProjectsWithMeta` 业务逻辑(Docker / Podman 共享)
- 修改:`internal/data/runtime/docker/compose_service.go` — ListProjects 实现(已落地,2026-08-08)
- 修改:`internal/data/runtime/podman/compose_service.go` — 同上(已落地)
- 修改:`internal/tui/ui/pages/compose/view.go:36 gatherComposeProjects` 改为调用 `ComposeService.ListProjects(All:true)`(而非直接遍历 `m.Resources.Containers.Items`)
- 状态:`m.Resources.Compose.Projects` 新增字段(具体实现位置待定 — 详见 R08-12 §状态机)

调用链:

```text
view.RenderPanel
  -> ComposeService.ListProjects(ListProjectsOptions{All:true})
     -> 项目聚合(containers + volumes + networks)
        -> 元信息 working_dir / config_files / version(从 label 聚合)
           -> annotate HasRunning / HasStopped
```

## 验收标准

1. `ComposeService.ListProjects` 返回包含 `Source / WorkingDir / ConfigFiles / Version / HasRunning / HasStopped`
2. `ListProjects` 以 `ListProjectsOptions{All bool}` 控制 stopped 容器范围:`All=false`(默认)仅聚合 running,`All=true` 聚合 running + stopped;实现不得写死 `All=true`
3. Podman runtime:WorkingDir / ConfigFiles / Version 字段在所有情况下为空,且 `Source=podman`
4. Docker runtime:testdata 注入的 `working_dir = /home/u/myapp` 单行展示;`config_files = /home/u/myapp/compose.yml\\n/home/u/myapp/compose.override.yml` 双行展示
5. 完全无容器项目不显示(`IsOrphan` 不存在,无 orphan 标记)
6. `InspectProject` 只读,不修改任何资源
7. `Config` 仍返回 `ErrComposeUnsupported` 错误(透明给 UI:显示"complete yaml unavailable, see CLI")

## 非目标

- 在线解析 compose yaml(`Config` 方法保留为 unsupported)
- 编辑 compose yaml(不在本项目范围 — 用户用 IDE / 编辑器)
- 检测 `compose.yaml` 文件更改后再聚合(留给 future enhancement,本需求接受 stale 元信息)

## R08-14 集成

UI 渲染层在以下位置接入 CoLocated group 标记:

- 右栏 services 列表:每行末尾折叠 group badge(`<N>`),hover 展开组成员
- 项目 detail 头部:增字段 "Groups: N"(若 N==0 隐藏)
- service 子视图(`renderComposeContainers`):顶部一行 `本 service 属于 group <id>,含 <services>`

数据来源全部由 [`ComposeProjectSummary.CoLocatedGroups`](./R08-14-co-located-group.md) 提供,R08-03 不增加新查询。

## 迁移记录

- 旧文档:[.omo/compose-todo.md §3.2 information lookup / 补 G.6 元信息增强](../../../.omo/compose-todo.md)
- 旧代码:`internal/tui/ui/pages/compose/view.go:36 gatherComposeProjects`(无元信息)
- 保留信息:`project / services / running count / total count` 现有字段
- 废弃信息:`IsOrphan`(无容器但有 working_dir 的孤儿标记)— 2026-08-08 决策"无容器则隐藏"后不再需要
