# 小模型下一步并发执行简报

> 更新日期: 2026-08-01
> 当前状态: BR-034、BR-039、BR-040 已完成。
> 执行规则: 不 commit、不 push，不修改任务范围外文件；完成后由主模型 review、集成并提交。

## 并发批次一

以下三个子任务文件范围不交叉，可以并行执行。

### 小模型 A: BR-033 高级容器动作

读取:

- `docs/task/br-033-task019-advanced-container-actions.md`
- `docs/task/br-033-advanced-container-ops-analysis.md`
- `internal/data/runtime/actions.go`
- `internal/data/runtime/containers.go`
- `internal/tui/keyboard/container_action.go`
- `internal/tui/keyboard/actions.go`

目标:

1. 先核验 Docker、Podman 的 Copy / Update / Diff / Export / Commit / Wait runtime 实现是否真实可调用。
2. 在独立文件 `internal/tui/keyboard/container_advanced.go` 实现不需要表单的 Cmd、消息和 handler 核心。
3. 优先完成 Diff、Wait；涉及文件保存、复杂表单或流关闭的 Copy、Export、Update、Commit 只做状态与接口设计，不仓促实现。
4. 新增 `container_advanced_test.go`，覆盖无选择、错误、成功消息和 reader 关闭。

禁止:

- 不修改 runtime adapter。
- 不修改 `keyboard/actions.go`、`keys/registry.go`、Action Bar、i18n 等共享集成文件。
- 不自行选择冲突快捷键。

### 小模型 B: BR-035 Events 核心

读取:

- `docs/task/br-035-events-panel-and-network-connect.md`
- `internal/tui/update/update_events.go`
- `internal/data/runtime/streaming.go`
- `internal/tui/state/log.go`
- `internal/tui/ui/pages/logs/`
- `internal/tui/ui/component/table.go`

目标:

1. 仅实现 Events，不实现 Network Connect / Disconnect。
2. 新增 `internal/tui/state/events.go` 及测试，支持 1000 条环形上限、Pause、Filter、Clear、Cursor、ViewOffset。
3. 新增 `internal/tui/ui/pages/events/view.go`、测试和独立 table profile。
4. 复用公共 Flex Table，只构造可见行。
5. 输出主模型需要接入的消息流和 mode 清单。

禁止:

- 不修改 `state/app.go`、`keyboard/actions.go`、`keyboard/keyboard.go`、`ui/app/*`、keys、footer、i18n 等共享集成文件。
- 不修改现有 event subscription 生命周期。
- 不实施 BR-035 的 Network 部分。

### 小模型 C: TASK-020 只读评估

读取:

- `docs/task/task-020-podman-capabilities-evaluation.md`
- `design/podman-capabilities-analysis.md`
- `internal/data/runtime/engine.go`
- Docker / Podman engine 与 service 列表

目标:

1. 核验 Podman pod / secret / kube 与 Docker buildx / context 的现有能力和缺口。
2. 按业务价值、实现成本、双 runtime 一致性、测试成本给出决议建议。
3. 只输出 `docs/task/task-020-capabilities-review.md`。
4. 不直接修改 feature design 或创建新 BR，决议由用户和主模型确认后再落地。

## 后续顺序

1. 主模型 review 并集成 BR-033。
2. 主模型集成 BR-035 Events；Network Connect / Disconnect 单独拆下一批。
3. BR-036 Image Import 与 BR-037 Registry Login 串行实施，两者都会修改 image runtime、Action Bar、dialog 和 i18n。
4. BR-038 Exec PTY 由主模型主导，子模型只适合承担独立 buffer/view 测试。

## 工作区保护

当前工作区已有未提交的 BR-039 文档、Podman fixture 和 BR-033 分析文件。小模型必须保留这些变化，不得 reset、checkout、格式化或重写无关文件。
