# R08-04 生命周期

## 元信息

- 状态: implemented
- 优先级: high
- 来源: 用户评审结果 + docker compose CLI 官方 spec
- 关联任务: [.omo/compose-todo.md §3.1 / §C.2 / 补 E.1 / 补 G.1](../../../.omo/compose-todo.md)
- 关联约束:
  - [../../constraint/C01-form.md](../../constraint/C01-form.md)
  - [../../constraint/C02-dialog.md](../../constraint/C02-dialog.md)
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)

## ⚠ 重复功能说明

| Compose 命令 | 已实现层 | 本需求处理 |
|---|---|---|
| `start` | R01 容器(`containerStartCmd`) | **保留** compose-级(start 整项目/单 service 一批);与容器级共存,API 路径不同(ComposeService.Start vs 容器 Start),仅结果相同 |
| `stop` | R01 容器 | **保留**;R01 容器级 stop 与本需求 stop **共存**(Action Bar scope 不同,API 路径不同) |
| `restart` | R01 容器 | **保留**;compose-级避免逐个 Tab → 容器 |
| `down` | R01 容器(`containerRemoveCmd`)+ R03 卷删删 + Network 删删(各自实现,**未协同**) | **新增 ComposeService.Down**,一个事务内完成 3 类资源,支持 `-v / --rmi / --remove-orphans` |
| `up` | 无 | **已实现**(2026-08-08,`ComposeService.Up` wrap 自实现,见 §决策记录);`--build/--force-recreate/--no-deps/--scale` 标志入 `UpOptions` |

## 目标

在 `ComposeService` 之上提供完整生命周期能力,补以下三个缺口:

1. 当前 `doComposeStart / doComposeStop / doComposeDown`(internal/tui/keyboard/compose_action.go)是**散装**循环,与 R01 容器级动作耦合紧但缺审计 / 错误处理一致性
2. 当前 `doComposeDown` 通过 Ctrl+D **不可达**:`registry.go:146 resourceDeleteOverrides` 没为 compose 分流,delete 走 `doDeleteAction` 静默返回。**这是 bug**(Compose 面板的 Down 失效)
3. `down` 没有 `-v / --rmi / --remove-orphans` 三个 compose-级标志的 Dialog

## 用户流程

### up(start 状态后的反向)

**当前阶段不实现 compose up**。UI 给出明确引导:Compose panel 状态栏展示 `up` 命令对应"请返回 docker compose CLI"。

### start

1. Compose 面板左栏选中项目(或右栏选中 service)
2. 按 `s` / Action Bar (`:` → 操作列表 → Start)
3. ShowProgress toast:`Starting <project>...`
4. 完成后 toast:`✓ <project> started (X/Y 容器)`
5. 部分失败:toast 显示 `⚠ X succeeded, Y failed, Z skipped`,失败容器以 detail link 形式展示

### stop

1. 类似 start,快捷键 `Ctrl+S`
2. 与 start 共享相同的 audit / toast 模型

### restart

1. 快捷键 `Ctrl+R`(默认未绑,需配置 `composeProjectRestart: [ctrl+r]`)
2. 流程同 stop + start 串联

### down

1. 快捷键 `Ctrl+D`(需 R08-12 注册)
2. **检查 `doComposeDown` 当前 dead code bug**,先修复 R08-12 才能在这里生效
3. 弹 Confirm Dialog(参考 [constraint/C02-dialog](../../constraint/C02-dialog.md))
4. Dialog 含 3 个可选开关:
   - `[-v] 删除卷`(默认 unchecked)
   - `[--rmi all] 删除镜像`(默认 unchecked)
   - `[--remove-orphans] 删除孤立容器`(默认 checked)
5. Confirm 后批量删除;audit 一次 trace(`ComposeTarget.Down`)含资源计数

## UI/UX

仅限本需求特有:

- Confirm Dialog:按 [C02-dialog](../../constraint/C02-dialog.md) 尺寸,3 个 checkbox 居左对齐,默认 `Tab` 聚焦 Confirm 按钮(避免误删)
- 三个开关用 `[ ]` / `[x]` 渲染,空格切换
- 进度条仅在 > 10 个容器时展示(避免短操作抖动)

完整 Dialog 视觉规范留到 UI overhaul 阶段。

## 功能规则

### F1 已有 bug:doComposeDown dead code

按 `.omo/compose-todo.md 补 E.1` 追踪链 — `Ctrl+D` 在 compose 面板**完全被 ActionDelete 拦截后静默 nil**,`doComposeDown` 在键盘路径下不可达。

修复方案(主模型决策优先):

**方案 A**:`resourceDeleteOverrides` 在 `registry.go:146` 增加 `case "compose", "compose-containers"` 返回 `true`(现成的"加 2 case"最小补丁)

**方案 B**:在 `ComposeService.Down` 走完后仍调用 `doComposeDown`(已存在),但必须先让 `doComposeDown` 在键盘路径可达(A 是前提)。

**结论**:无论 R08-04 是否新写, A 是**必须**先做的修复。

### F2 ComposeService.Down 行为契约

```go
// 来源:R08-02 决策 B
type DownOptions struct {
    RemoveVolumes    bool
    RemoveImages     string // "" | "local" | "all"
    RemoveOrphans    bool
    TimeoutSec       int
    ProjectTimeoutMs int
}

func (s *DockerComposeService) Down(ctx context.Context, project string, opts DownOptions) error
```

实现:

1. 列出所有 `com.docker.compose.project=<p>` 的容器(`ContainerList` + label filter)
2. 列出该项目所有网络(按 label)
3. 列出该项目所有卷(按 label)
4. 找 orphans:有 `container-number` label 但 config-hash 在 `com.docker.compose.config-hash` 不匹配(需要 sentinel 复杂度,可在 F3 简化)
5. 调用 `ContainerRemove(volumes=false)` + `NetworkRemove` + 可选 `VolumeRemove` + 可选 `ImageRemove`,批量(语义原子,失败有部分成功报告)
6. `BatchActioned` 结果封装,同 R01 容器级 `BatchActioned`

### F3 Orphans 检测(简化)

完整 orphans 检测需要"先后台运行 compose 的 config-hash"参照物。在 dtui 此参照物缺失。

**简化**:把 orphans 定义为:

> "容器带 `com.docker.compose.project=<project>` label,但 `com.docker.compose.service` 不在该项目的 service 集合(由 label 聚合的反向推出)中"

即:`service label` 不在项目其它容器宣告的 service 集合内 → orphan。

实现:从所有项目容器聚合 `project→services` 映射,然后对该 project 列表每个容器校验 `service∈services`,不在即 orphan。

### F4 start / stop / restart 行为

```go
func (s *DockerComposeService) Start(ctx context.Context, project string, services []string) error
```

- `services` 空 → 所有 `Stopped` / `Created` 容器全部 start
- `services` 非空 → 只对 `c.ComposeService ∈ services` 的容器 start(对其它该项目的容器不动)
- 返回 `BatchActioned`,失败容器 ID 列表

`stop / restart` 同模式。

### F5 up 实现(2026-08-08 落定)

```
up      -> ComposeService.Up(已实现,wrap 自实现 per R08-02 Q3)
build   -> R08-05 处理;不在 R08-04 范围
config  -> ErrComposeUnsupported  (R08-03 给替代)
pull    -> R08-05 处理;不在 R08-04 范围
```

UI 需在 Action Bar / 帮助文档明示,参考 R08-11。

### F6 audit 与 trace

- start / stop / restart → `ComposeTarget{Name: project, Meta: {Containers: N}}` 单 trace
- down → `ComposeTarget{Name: project, Meta: {Containers: N, Volumes: M, Networks: K}}` 单 trace,**容器 / 卷 / 网络合并**,不拆 3 条

## 实现设计

涉及模块:

- 新文件:`internal/data/runtime/docker/compose_service.go`(R08-02 已建)增加 `Down / Start / Stop / Restart` 实现
- 新文件:`internal/data/runtime/podman/compose_service.go` 同上
- 修改:`internal/tui/keyboard/compose_action.go:67-239`(全部 `doCompose*` 改为调用 `ComposeService.*`,移除直接 `containerStartCmd` 循环)
- 新文件:`internal/tui/ui/pages/compose/down_confirm_dialog.go` — Confirm Dialog with 3 checkboxes(若 R03-03 Dialog 框架允许复用,优先复用)
- 修改:`internal/tui/keys/registry.go:146 resourceDeleteOverrides` 加 `case "compose", "compose-containers": return true`(F1 修复)
- 修改:`internal/tui/keyboard/compose_nav.go:69` 重新确认 Ctrl+D 路径走通
- 修改:`internal/tui/keyboard/audit.go` 增 `ComposeDown` audit
- 新增 ui/state:`internal/tui/state/compose.go` 加 `DownOptions DownOptions` 临时态字段

调用链:

```text
Ctrl+D (compose 面板)
  -> resolveAction → ActionDelete 被 resourceDeleteOverrides 跳过(Compose)
  -> handlePanelFallbacks → handleComposePanelKeys
  -> case KeyCtrlD → doComposeDown → 展示 Confirm Dialog
     -> Confirm → ComposeService.Down(project, opts)
        -> BatchActioned (resources + counts)
           -> Audit + Refresh Containers/Volumes/Networks
```

## 验收标准

1. Ctrl+D 在 compose 面板触发 Down Dialog(不再被 ActionDelete 拦截)
2. Down Dialog 含 3 个 checkbox,默认值符合 F2 默认
3. 部分失败:toast 报告 succeeded / skipped / failed 三计数
4. `DownOptions.RemoveVolumes=true` 时,带 `com.docker.compose.volume` label 的卷一并删
5. `DownOptions.RemoveImages="local"` 仅删 compose build 的镜像(scratch + 中间层),**不**删拉取的 docker hub 镜像(`com.docker.compose.image` label 存在时)
6. Audit 一次 trace,type=`compose_project.down`
7. start / stop 与 R01 容器级 stop 区分 audit 类型,但行为不冲突(并发场景下 action 只针对项目容器)
8. `Up` 命令在 UI 可见(2026-08-08 起已实现,`ComposeService.Up`),Action Bar 明示支持标志

## 非目标

- 实现 compose build(留 R08-05 / future)
- 跨 runtime 编排(Docker 项目和 Podman 项目同名处理) — 留 R08-13
- Compose Spec 多文件 `include`(单 compose 项目)

## R08-14 集成

`DownOptions` 与 CoLocated group 关系(详见 [R08-14 §F6](./R08-14-co-located-group.md)):

- `doComposeDown(m)` 默认行为**不变**(down 整个项目)。当调用方传入 group ID(R08-12 增 `ActionComposeGroupDown`),Down 范围仅限该 group 内 services。
- audit `ComposeTarget.Meta` 增字段 `GroupID / Runtime / Source`(R08-14 §F7)。
- UI 行为:用户按 `Ctrl+D` = down 整个项目;用户在 Action Bar 选中 group 后按 Shift+Ctrl+D = down group(toast 明示 "down group <id> 含 N services")。

Down 内 services 列表由 [`ComposeProjectSummary.CoLocatedGroups`](./R08-14-co-located-group.md) 反查;无需重新 list 容器。

## 迁移记录

- 旧代码:`internal/tui/keyboard/compose_action.go:67-239`(doComposeStart/Stop/Down 当前实现)
- 旧 bug:`doComposeDown` dead code(CTRL+D 不可达)
- 保留信息:`composeProjectContainers / composeProjectVolumes / composeProjectNetworks` helper(底层 label 匹配,在 `ComposeService` 实现里复用)
- 废弃信息:无
- 待确认:`Options.RemoveImages="local"` 判定方法;`Orphans` 简化检测是否对其它项目场景足够

### 2026-08-08 阶段 B 落地

**已完成**(本轮):

- `doComposeRestart`(internal/tui/keyboard/compose_action.go):Stop + Start 链。两条独立 BatchActioned 各自携带 audit trace,各自 toast,不共用 trace(R08-04 F6:start/stop 各自 trace)。
- `doComposeDown` 接 Confirm Dialog(R08-04 F2 全实现):
  - `internal/tui/state/confirm.go::ChoiceOption` 加 `Checked bool` 字段,支持 checkbox 行。
  - `internal/tui/keyboard/confirm.go`:Space 键切换当前 focus 的 checkbox 状态。
  - `internal/tui/keyboard/compose_action.go::openComposeDownConfirm`:打开 5 选项 dialog(3 checkbox + Cancel + Confirm),`ComposeDownRemoveOrphans` 默认 checked。
  - `doConfirmYes`(mark_action.go)从 `m.Confirm.Options` 拷贝 Checked 状态到 `m.Compose.ComposeDownRemove*`,置 `ComposeDownSkipConfirm=true`,递归调 `doComposeDown`。
  - 第二次进入 `doComposeDown` 时按 `ComposeDownRemoveVolumes / ComposeDownRemoveOrphans` 过滤:`--remove-orphans` unchecked 时只删 project 容器,checked 时也删孤儿(项目 service 集外的容器)。

**测试**:`TestComposeDownAggregatesResources` 加 `ComposeDownSkipConfirm=true` 绕过 dialog 直接验证 down 路径;35 packages 全 PASS。
