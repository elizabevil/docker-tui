# BR-033 / BR-035 审查问题修复记录

> 更新日期: 2026-08-02
> 状态: 5 项审查问题已修复并完成主模型集成复核
> 关联: `docs/task/br-033-task019-advanced-container-actions.md`、`docs/task/br-035-events-panel-and-network-connect.md`

## 背景

主模型审查发现 BR-033 / BR-035 小模型产出存在 5 项问题(2 高 / 2 中 / 1 低),
要求修复后才可进入集成。本记录逐项说明修复方案与落盘位置。

## 修复清单

### 1. 高: Events 状态字段冲突(EventState vs EventsPanelState)

**问题**: `AppModel.Events` 已是订阅生命周期 `EventState`(update_events.go 依赖其
`Current/MarkDirty`),新面板若再加同名字段 `Events EventsPanelState` 无法共存。

**修复**:
- 明确拆分为两个独立 AppModel 字段:
  - `Events EventState` — 订阅生命周期(不动,update_events.go 语义保持)
  - `EventPanel EventsPanelState` — BR-035 面板缓冲(新增)
- `ui/pages/events/view.go` 集成说明改为 `EventPanel` 字段名,并显式标注
  "Do NOT reuse the existing `Events state.EventState` field"。
- `state/events_panel.go` 文档注释同步说明字段拆分。

**文件**: `internal/tui/state/app.go`、`internal/tui/state/events_panel.go`、`internal/tui/ui/pages/events/view.go`

### 2. 高: Container Wait 无法取消(永久后台 Cmd)

**问题**: `containerWaitCmd` 使用 `context.Background()`,容器长期运行时切换
runtime / 退出应用都无法终止。

**修复**:
- 新增 `state.ContainerWaitState`(镜像 `ImageTransferState` 模式):
  `Begin() (ctx, generation)` / `Stop()` / `Current(generation)`。
- `AppModel.ContainerWait ContainerWaitState` 字段已加。
- `containerWaitCmd` 改为接收 `ctx context.Context`;`doContainerWait` 通过
  `m.ContainerWait.Begin()` 取得可取消 context。
- 集成层在 runtime 切换 / 应用退出时调用 `m.ContainerWait.Stop()` 即可取消
  (见 container_advanced.go 注释说明)。

**文件**: `internal/tui/state/container_ops.go`(新增)、`internal/tui/state/app.go`、`internal/tui/keyboard/container_advanced.go`

### 3. 中: Export/Copy 直接截断目标文件,失败留损坏 tar

**问题**: `os.Create(dst)` 立即覆盖已有文件;`io.Copy` 中途失败仍保留部分内容;
忽略目标文件 Close 错误。

**修复**: `copyStreamToFile` 改为原子写:
1. 在目标同目录创建临时文件(`os.CreateTemp(dir, ".dtui-stream-*")`)
2. `io.Copy` 完成后显式 `tmp.Close()` 并检查错误
3. 全部成功后才 `os.Rename(tmpName, dst)` 原子替换
4. 失败路径 defer 清理临时文件;已有目标文件保持原样不被截断

**文件**: `internal/tui/keyboard/container_advanced.go`

### 4. 中: Events 上限实际可达 2000 条

**问题**: Events 与 Pending 各限 1000,暂停期间总内存 2000 条,违反"最近 1000 条"。

**修复**: `append` 改为对 live+pending **合并容量** 上限 MaxEvents:
- 超出时优先丢弃 live 头部(最旧),再丢弃 Pending 头部
- 暂停期间新事件保留在 Pending(live 先被淘汰),恢复时 flush 不超上限

**文件**: `internal/tui/state/events_panel.go`

### 5. 低: Events renderer 钳制局部 offset,不修正 state

**问题**: view.go 局部 clamp `offset`,但 `ev.ViewOffset` 仍保留非法值,后续
键盘计算可能基于旧 offset 跳动。

**修复**: 新增 `EventsPanelState.ClampOffset(total, visible)`:
- 钳制 `ViewOffset` 到 `[0, max(0, total-visible)]`
- 钳制 `Cursor` 到 `[0, total-1]`
- renderer 渲染前调用,state 与实际视口保持一致

**文件**: `internal/tui/state/events_panel.go`、`internal/tui/ui/pages/events/view.go`

## 测试补充

| 测试 | 覆盖 |
|---|---|
| `TestCombinedLiveAndPendingCap` | Fix 4: live+pending 总容量 ≤ 1000,最旧 live 先淘汰 |
| `TestClampOffset` | Fix 5: 越界 offset/cursor 被钳制到合法范围 |
| `TestExportFailedCopyPreservesExistingDestination` | Fix 3: 失败复制不截断已有文件、无临时文件泄漏、reader 关闭 |
| `TestContainerWaitUsesCancellableContext` | Fix 2: Stop() 后 Wait 返回 context.Canceled |
| `TestContainerWaitUsesCancellableContext`(并发) | Fix 2: cmd 并发执行期间取消生效 |

## 验证结果

```text
go build ./...            exit 0
go vet ./...              exit 0
go test -short ./...      exit 0 (全部通过)
go test ./internal/tui/state/ ./internal/tui/keyboard/ ./internal/tui/ui/pages/events/  exit 0
```

## 集成结果（2026-08-02）

- [x] BR-035 Events：`ModeEvents`、F3、事件投递、Pause / Filter / Clear、页面和 footer 已接入。
- [x] BR-033 Diff / Wait：已接入 Action Bar；其余参数化动作保留为后续子任务。
- [x] runtime 切换/退出会取消 Container Wait，且取消结果会正确结束 audit。
- [x] 高级动作不再分配冲突的默认快捷键。
- [ ] TASK-020 报告仍仅作决策输入，不直接标记任务完成。
