# 批量操作设计 — Containers 页面 + 参数化操作

> **状态**: 设计稿已确认(2026-08-05)— 用户已决策 MarkedIDs 隔离与 timeout=0 映射
> **关联**: docs/discussion/select-dialog-design.md(Force 选项与 Form 组件统一)
> **关联 explore**: `bg_3f1111db` 调查

---

## 0. 用户决策(2026-08-05 第二轮)

- ✅ **MarkedIDs 按 panel 隔离**(Option γ-A) — 切换 panel 互不影响
- ✅ **timeout=0 映射为 Kill** — UI Stop Form 的 timeout 字段值为 0 时,内部走 `ContainerKill`
- ✅ **Podman Signal 同步见 §X** — Phase 3 暴露 Stop timeout/signal Form 时,Podman adapter 必须同步支持 Signal 转发

---

## 1. 现状与背景

### 1.1 用户报告

> Container 页面不支持批量停止启动Kill等操作,同样参数不包括格外参数如 Force。

### 1.2 关键结论

**当前代码不是"完全没有批量实现",而是后端存在但 UI 可达性未完成**:

1. **后端批量逻辑完整**:`MarkedIDs` 多选状态、批量 handler、批量结果汇总(`BatchActioned`)和刷新逻辑都存在
2. **UI 可达性断裂**:`ModeMark` 截获后续按键,**只允许** Space/Enter、Up/Down、Tab、Ctrl+D、Esc — **不转发** Start/Stop/Restart/Kill/Pause 给批量 handler
3. **用户实际绕路**:先进入批量删除确认再取消 → 间接回到 `ModeNormal` → 此时 Footer 才显示批量按钮(`action/registry.go:150-158`)— 非预期交互
4. **参数几乎全部未暴露**:Stop/Restart/Kill/Remove/Wait 的所有可选参数(Timeout/Signal/Force/RemoveVolumes/WaitCondition)在 UI 上都没有

---

## 2. 已有的批量后端能力(latent implementation)

来自 `bg_3f1111db` 的全量调查:

### 2.1 多选状态机制

| 层 | 文件 | 能力 |
|---|---|---|
| State | `internal/tui/state/selection.go:5-22` | `SelectionState.MarkedIDs map[string]bool` + `Toggle(id)` |
| View | `internal/tui/ui/pages/containers/view.go:28, 159-175` | 接收 markedIDs,投影为 `MarkedRows` |
| Table | `internal/tui/ui/component/table.go:35-63, 226-237` | `TableData.MarkedRows map[int]bool` + `BuildMarkedRows` |

### 2.2 批量 handler(已实现)

| 操作 | 文件 | 函数 |
|---|---|---|
| Batch Start | `internal/tui/keyboard/container_action.go:66-121` | `doBatchContainerAction` 分支 |
| Batch Stop | 同上 | 同上 |
| Batch Restart | 同上 | 同上 |
| Batch Kill | 同上 | 同上 |
| Batch Pause/Unpause | `internal/tui/keyboard/container_action.go:150-218` | 批量切换 |
| Batch Remove | `internal/tui/keyboard/delete_action.go:13-19` + `mark_action.go:131-199` | `executeBulkDelete` |
| Batch Force | `mark_action.go:56-63` + `confirm.go:27-34` | `ShowBulkDeleteForce` 路由 |
| Batch Stop Force → Batch Kill | `confirm.go:27-34` | 重映射 `ShowBatchStop+Force → ShowBatchKill` |

执行模式:单命令内按 ID 顺序逐个调用,**非并发批量 API**。结果通过 `BatchActioned` 聚合成功/失败/失败 ID/joined error。

### 2.3 操作矩阵(全部单容器)

| 操作 | UI 触发 | Handler | SDK 调用 | 实际参数 |
|---|---|---|---|---|
| Start | `s` (no confirm) | `container_cmd.go:13-18` | `ContainerStart(ctx, id, StartOptions{})` | 无 |
| Stop | `Ctrl+S` (confirm) | `containerStopCmd` | `ContainerStop(ctx, id, StopOptions{Timeout:nil})` | Timeout=nil,**Signal 未传递** |
| Restart | `Ctrl+R` (confirm) | `containerRestartCmd` | `ContainerRestart(ctx, id, StopOptions{Timeout:nil})` | Timeout=nil,**Signal 未传递** |
| Kill | `Ctrl+K` (confirm) | `containerKillCmd` | `ContainerKill(ctx, id, "")` | 空 signal = 引擎默认(SIGKILL),**Signal 未暴露** |
| Pause/Unpause | `p` (no confirm) | `containerPauseCmd(unpause=?)` | `ContainerPause/Unpause(ctx, id)` | 动态按状态切换 Toggle |
| Remove | `Ctrl+D` (confirm) | `containerRemoveCmd(..., true)` | `ContainerRemove(ctx, id, RemoveOptions{Force:true})` | **Force 硬编码 true;RemoveVolumes/RemoveLinks 被丢弃** |
| Wait | `;` ActionBar | `doContainerWait` | `ContainerWait(ctx, id, "")` | condition 固定空 → 适配器转为 `not-running` |
| Logs | `l` / `Enter` | `Engine.Containers().Logs` | Docker `ContainerLogs` | Since/Tail/Timestamps 来自全局配置 |
| Exec | `e` | `Engine.Exec().Open` | `ContainerExecCreate + Attach` | TTY passthrough |
| Inspect/Detail | `i` / `d` | `doInspectAction` | `ContainerInspect` | — |
| Stats | `m` | `doStatsAction` | `ContainerStats` | 单容器 stats |
| Rename | `;` ActionBar / `:rename` | `containerRenameCmd` | `ContainerRename(ctx, id, name)` | — |
| Update Resources | `;` ActionBar | Form → `ContainerUpdateOptions` | `ContainerUpdate` | Memory, NanoCPUs, RestartPolicy, MaxRetries(已 Form 化) |
| Copy / Export / Commit | `;` ActionBar | Form → 各自 SDK | `CopyFromContainer` / `ContainerExport` / `ContainerCommit` | 已 Form 化 |

### 2.4 ActionBar 菜单(只支持单容器高级操作)

`internal/tui/actionbar/registry.go:51-93` 当前 Containers ActionBar 9 项:
1. Rename
2. Top
3. Port
4. Filesystem Diff
5. Wait
6. Copy
7. Update
8. Export
9. Commit

**没有**: Start/Stop/Restart/Kill/Pause/Remove/Logs/Exec/Inspect/Stats(直接快捷键,不走 ActionBar)

底部 Confirm/Cancel 只是菜单项"按 Enter 执行 / 按 Esc 关闭"的提示,不是批量按钮组。

---

## 3. UI 可达性问题根因(ModeMark 截获)

### 3.1 当前 keymap(ModeMark)

`internal/tui/keyboard/mark_mode.go:23-57` 仅处理:
- `Space` / `Enter`: Toggle mark
- `Up/Down`: 移动
- `Tab`: 切换 panel
- `Ctrl+D`: 批量删除
- `Esc`: 清除标记并退出

**未处理**: `s`, `Ctrl+S`, `Ctrl+R`, `Ctrl+K`, `p`

### 3.2 用户实际体验

```
Space        → ModeMark (第一次 Space 不标记当前行)
Space × N    → 标记 N 个容器
Ctrl+S       → ModeMark 截获,什么也不发生
Esc          → 清空标记退出
```

因此报告的"不支持批量 Stop/Start/Kill"基本准确。

### 3.3 间接路径(非预期)

1. Space 进入 Mark mode
2. 多次 Space 标记多容器
4. Ctrl+D 触发批量删除确认(进入 ModeConfirm)
5. 在 Confirm 框按 Esc 取消 — **不清空 MarkedIDs**
6. 回到 `ModeNormal`,但 MarkedIDs 仍存在
7. 此时 Footer 进入 `marked` 分支(`action/registry.go:150-158`)显示批量快捷键
8. 此时 Ctrl+S 才能批量 Stop

这条路径**不应**是用户依赖的批量入口。

### 3.4 MarkedIDs 隔离问题

当前 `MarkedIDs` 没有按 panel 隔离:在 Containers 标记的 ID,切到 Images 面板后**仍存在**,可能被错误地用于批量操作。建议每个 Panel 独立维护自己的 MarkedIDs。

---

## 4. 参数化操作现状(Force/Signal/Timeout)

### 4.1 单容器 Force/Signal/Timeout 处理

| 操作 | SDK 字段 | 当前实现 | 缺失 |
|---|---|---|---|
| ContainerStop | `Timeout *int`, `Signal string` | 仅 Timeout=nil | **Signal 未传递** |
| ContainerRestart | `Timeout *int`, `Signal string` | 仅 Timeout=nil | **Signal 未传递** |
| ContainerKill | `Signal string` | 空 signal(默认 SIGKILL)| **Signal 未暴露** |
| ContainerRemove | `Force bool`, `RemoveVolumes bool`, `RemoveLinks bool` | 仅 Force 硬编码 true | **Force 不可选;RemoveVolumes/RemoveLinks 被丢弃** |
| ContainerWait | `WaitCondition string`(not-running/next-exit/removed)| 空 → `not-running` | **Condition 未暴露** |

### 4.2 批量 Force/Signal 现状

- **Batch Stop + Force** → ~~`confirm.go:27-34` 重映射为 `Batch Kill`~~ — **用户已确认改为 timeout=0 映射为 Kill**
- **Batch Remove + Force** → 共享 `executeBulkDelete` 的 Force 参数,所有目标用同一 Force
- **Batch Kill** → 直接逐个 ContainerKill("")(默认 SIGKILL)
- **Batch Start** → 不需要 Force
- **Batch Restart** → 不需要 Force,不需要 Signal(当前)
- **Batch Pause/Unpause** → 按每个容器状态自动 Toggle(语义模糊,应该拆为 Pause + Unpause 两个操作)

#### timeout=0 映射为 Kill(用户已确认)

**用户决策**: timeout=0 在 UI Stop Form 上,语义上等价于 "立即终止" = **走 ContainerKill,不走 ContainerStop**。

实现:
- Stop Form 的 `Timeout`(FormInt, 单位秒)默认 10
- 当用户提交时,如果 `timeout == 0`:
  - 内部走 `ContainerKill(ctx, id, signal)`(默认 signal="",即 SIGKILL)
  - UI toast 显示 "Force killing container"
  - audit trace 标记为 `ActionKill` 而非 `ActionStop`
- 当 `timeout > 0`:
  - 走 `ContainerStop(ctx, id, StopOptions{Timeout: &timeout, Signal: signal})`
  - Docker 引擎先发 signal(默认 SIGTERM),等 timeout 秒后 SIGKILL
- 批量场景同样适用:`Batch Stop with timeout=0` 实际执行 `Batch Kill`

**移除** 当前的 `confirm.go:27-34` 的 `ShowBatchStop+Force → ShowBatchKill` 重映射逻辑 — 改为 Stop Form 提交时按 timeout 值分发。

### 4.3 Podman 对齐

`internal/driver/podman/containers_actions.go:15-20` 当前 Podman Stop 只处理 Timeout,未处理 Signal;Kill 只处理 Signal,未处理 Force。Docker 与 Podman 的能力差异需要 runtime capability gating(参考现有 Runtime capability 模式)。

---

## 5. 设计选项

### 5.1 UI 路由修复(Mark mode 转发)

#### Option α-A: 在 `mark_mode.go` 显式处理批量键

`internal/tui/keyboard/mark_mode.go` 增加分支:
```go
case "s":                  return doBatchContainerAction(m, ActionStart)
case key.CtrlS:            return doBatchContainerAction(m, ActionStop)
case key.CtrlR:            return doBatchContainerAction(m, ActionRestart)
case key.CtrlK:            return doBatchContainerAction(m, ActionKill)
case "p":                  return doBatchContainerAction(m, ActionPause)
case key.CtrlD:            return doDeleteAction(m)  // 已有
```

**优点**:改动小(`mark_mode.go` ~30 行 + 测试)
**缺点**:每个新批量操作要记得加 case

#### Option α-B: 第一次 Space 直接标记并回 ModeNormal

`mark_mode.go` 改造:第一次 Space 直接 Toggle 当前行并回 `ModeNormal`,**保留** MarkedIDs。后续 Space 继续 toggle。Esc 才清空并退出。

**优点**:UX 更直接 — 选完即可触发批量,无需"完成选择"按钮
**缺点**:Esc 误按会清空所有标记(可在 Esc 上加 confirm)

#### Option α-C: 增加"完成选择"动作键(如 `Enter` 二次确认)

第一次 Space 进入 Mark mode,Space toggle,**Enter** 退出 Mark mode 但保留 MarkedIDs。然后按 `s/Ctrl+S/Ctrl+K` 触发批量。

**优点**:语义清晰
**缺点**:多一个步骤,UX 不够直观

**推荐**:Option α-A(最小)+ Option α-B(改良 UX)的组合 — mark_mode 转发所有批量键,同时让首次 Space 直接 toggle 并退出,保留 MarkedIDs。

### 5.2 参数化操作 UI(Force/Signal/Timeout)

#### Option β-A: 全部走 Form(对齐 Update Resources 模式)

每个参数化操作打开 Form:
- Stop: Timeout (FormInt, 单位秒) + Signal (FormSelect/FormText)
- Restart: Timeout + Signal
- Kill: Signal
- Remove: Force (FormBool) + RemoveVolumes (FormBool)
- Wait: Condition (FormSelect)

**优点**:与现有 Update Resources/Copy/Export/Commit 风格一致,可扩展
**缺点**:每次 Stop 都要弹 Form,打断快速操作流

#### Option β-B: 快速执行(快捷键无 Form)+ 高级执行(Form)

- `Ctrl+S` / `Ctrl+R` / `Ctrl+K` / `Ctrl+D`:快速执行,**默认参数**(Signal=""/Timeout=nil/Force=false)
- `;` ActionBar 增加 `Stop with options` / `Restart with options` / `Kill with options` / `Remove with options` / `Wait with options`,:打开 Form
- 批量场景同样:快捷键批量快速执行,ActionBar 二级菜单批量带参数

**优点**:兼顾快速路径与高级路径;用户可按需
**缺点**:ActionBar 项数增加

#### Option β-C: 走 ChoiceDialog(同 select-dialog-design Option A)

Stop/Kill/Restart:3 选项 ChoiceDialog
- `[ Cancel
- `[ Stop ]` (默认参数,timeout=10s 等)
- `[ Force Stop ]` (timeout=0 + signal=KILL)

**优点**:与 Bulk Delete 模式一致,改动小
**缺点**:无法同时暴露 Signal 选择

**推荐**:Option β-B — 快速 + 高级 Form 双路径。理由:
- 与现有"直接快捷键 + ; ActionBar 高级"模式对齐
- 不打断快速操作
- Signal/Timeout 需要 FormInt/FormSelect/FormText,ChoiceDialog 不适合

### 5.3 MarkedIDs 隔离

#### Option γ-A: 每个 Panel 独立 MarkedIDs

`SelectionState.MarkedIDs` → `SelectionState.PanelMarks map[PanelType]map[string]bool`,切换 panel 时互不影响。

**优点**:防止跨 panel 误操作
**缺点**:改动稍大(`SelectionState` 结构 + 所有引用)

#### Option γ-B: 切换 panel 时自动清空

`onPanelSwitch()` 钩子清空 `MarkedIDs`。

**优点**:简单
**缺点**:用户切走再切回丢失选择,不友好

**推荐**:Option γ-A(结构清晰,长期可扩展) — **用户已确认采用**

**用户已确认采用 Option γ-A**:
- `SelectionState.MarkedIDs` → `SelectionState.PanelMarks map[PanelType]map[string]bool`
- 每个 Panel 独立维护自己的 marked IDs
- 切换 panel 时 `len(PanelMarks[newPanel])` 不变,UI 提示"已标记 N 个"
- `Esc` 不再清空所有 panel 的标记,只清当前 panel
- 提供专门的 "clear marks" ActionBar 项(可选)

---

## 6. 推荐的完整方案

### 6.1 三阶段路线

**Phase 1:UI 路由修复(让批量可达)**
- `mark_mode.go` 转发 Start/Stop/Restart/Kill/Pause 键(`Option α-A`)
- 首次 Space 直接 toggle 并退出 Mark mode 但保留 MarkedIDs(`Option α-B`)
- MarkedIDs 按 panel 隔离(Option γ-A) — **用户已确认**

**Phase 2:批量操作启用 ActionBar 二级菜单**
- `actionbar/registry.go` 增加 `Stop with options` / `Restart with options` / `Kill with options` / `Remove with options` / `Wait with options`
- 各菜单项打开对应 Form
- 当 `len(MarkedIDs)>0` 时,ActionBar 切换为批量视图(Batch Start/Stop/Restart/Kill/Pause/Remove/Force Remove)

**Phase 3:参数化操作 Form 实施(与 select-dialog-design 协调)** — **用户已确认**
- 单容器与批量操作共用同一 Form 组件(`Option β-B`)
- 新增 FormKind:`FormContainerStop`, `FormContainerRestart`, `FormContainerKill`, `FormContainerRemove`, `FormContainerWait`
- 扩展 `runtimeapi.LifecycleOptions`:`Signal string`, `Timeout time.Duration`, `Force bool`, `RemoveVolumes bool`, `WaitCondition string`
- **timeout=0 → 内部走 ContainerKill** — **用户已确认**
- **Podman Signal 同步**(强制):
  - Stop API: Podman 4.5+(API ≥1.41)支持 signal query 参数
  - 当前 Podman adapter(`internal/driver/podman/containers_actions.go`)Stop 分支仅 timeout
  - 修改: Stop 分支加 `signal: lc.Signal` query 参数(若非空)
  - Capability gate: `runtimeapi.Capabilities.SupportsSignalInStop` 在 Podman 连接握手时探测
  - 早期 Podman(API <1.41)降级: Signal UI 字段显示但提交时忽略 + toast 警告
- i18n 新增:
  ```
  container.action.form.stop.title:        Stop container
  container.action.form.stop.timeout:      Grace period
  container.action.form.stop.timeout.unit: seconds
  container.action.form.stop.signal:        Signal
  container.action.form.stop.signal_help:  e.g. SIGTERM, leave empty for engine default
  container.action.form.restart.title:     Restart container
  ...
  container.action.form.kill.title:       Kill container
  container.action.form.kill.signal:       Signal
  container.action.form.kill.signal_help:  default is SIGKILL
  container.action.form.remove.title:     Remove container
  container.action.form.remove.force:      Force remove
  container.action.form.remove.volumes:    Also remove anonymous volumes
  container.action.form.wait.title:       Wait for container
  container.action.form.wait.condition:    Condition
  container.action.form.wait.condition.not-running: Stop running
  container.action.form.wait.condition.next-exit:    Next exit
  container.action.form.wait.condition.removed:      Until removed
  ```

### 6.2 与 select-dialog-design 的协调

`FormContainerRemove` 与 select-dialog-design 的 Option B(`FormImageRemove`)对齐:
- 都用 FormSelect `{ Delete, Force Delete }` + Form 自带 Cancel 按钮
- Container Remove 额外加 `FormBool RemoveVolumes`(镜像无此参数)
- Image Remove 额外加 `FormBool PruneChildren`(容器无此参数)
- Volume Remove 同样模式
- Network Remove 不进 Form(SDK 无 Force,保持 2 选项 ChoiceDialog)

### 6.3 不建议批量化的操作

(直接保留单容器实现)
- Logs(多日志流需要合并、排序、目标标识)
- Exec(交互 TTY 无法合并)
- Inspect/Detail(输出目标不明确)
- Stats(可做聚合但当前 UI 是单容器,改造量大)
- Top(多进程表需要新多目标界面)
- Port(单容器详情)
- Copy(每容器 source/dest 不同)
- Export(每容器需要独立路径)
- Update Resources(每容器现有资源不同)
- Commit(每容器对应不同 repository/tag)
- Wait(多阻塞 wait 需要并发任务管理)

---

## 7. 影响范围(预计改动)

### 7.1 Phase 1 改动(UI 路由修复)

| 文件 | 改动 |
|---|---|
| `internal/tui/state/selection.go` | `MarkedIDs` → `PanelMarks map[PanelType]map[string]bool` |
| `internal/tui/keyboard/mark_mode.go` | 转发批量键 + 首次 Space toggle 后退出 |
| `internal/tui/keyboard/keyboard.go` | 调整 Mark mode 状态转换 |
| `internal/tui/ui/pages/containers/view.go` | 适配新的 `PanelMarks` 字段 |
| `internal/tui/ui/component/table.go` | 同上 |
| `internal/tui/keyboard/mark_action.go` | 适配新结构 |
| `internal/tui/ui/action/registry.go` | 适配 footer 批量项显示 |

预计 ~150-200 行。

### 7.2 Phase 2 改动(ActionBar 二级菜单)

| 文件 | 改动 |
|---|---|
| `internal/tui/actionbar/registry.go` | 增加 5 个带参数的 ActionBar 项 |
| `internal/tui/state/form.go` | 新增 `FormContainerStop` / `Restart` / `Kill` / `Remove` / `Wait` 5 个 FormKind |
| `internal/tui/keyboard/container_form.go` | 新增对应 buildForm 函数 |

预计 ~200-300 行。

### 7.3 Phase 3 改动(参数化 Form 实施)

| 文件 | 改动 |
|---|---|
| `internal/tui/keyboard/container_form.go` | `submitContainerForm` 新增 5 个 case 分支 |
| `internal/tui/keyboard/container_cmd.go` | 各 cmd 函数接受新 options |
| `internal/data/runtime/actions.go` | 扩展 `LifecycleOptions` 加 Signal / RemoveVolumes / WaitCondition |
| `internal/data/runtime/docker/service_action.go` | Stop/Restart/Kill/Remove 映射新字段 |
| `internal/data/runtime/podman/service_action.go` | 同上 |
| `internal/driver/podman/containers_actions.go` | Signal 处理(若 runtime capability 允许) |
| `internal/data/i18n/lang/{en,zh,ja}.jsonc` | 上述新增 keys |

预计 ~300-400 行。

### 7.4 测试

- `mark_mode_test.go`:新批量键转发测试
- `selection_test.go`:`PanelMarks` 隔离测试
- `container_form_test.go`:新增 5 个 Form 测试
- PTY 视觉验证:light 主题下批量操作 ActionBar 切换、批量 Stop/Kill 确认

---

## 8. 用户已决策 + 剩余开放问题

### 已决策
- ✅ MarkedIDs 按 panel 隔离(Option γ-A)
- ✅ timeout=0 映射为 Kill(移除 Batch Stop Force 重映射)
- ✅ Phase 3 暴露 Stop/Restart/Kill/Remove/Wait 全部参数 Form
- ✅ 参数化操作走 Option β-B(快速 + 高级 Form 双路径)

### 剩余开放
1. **Phase 1 立即做吗?** (纯 UI 路由修复,低风险)
2. **Phase 2 / Phase 3 一起做还是分批?**
3. **Kill Form 的默认 signal** (当前 SDK 默认 SIGKILL) — UI 默认显示什么?
4. **Wait Condition 是否本期暴露?** (SDK: not-running / next-exit / removed)
5. **PruneChildren Podman 兼容性** — 需查 Podman REST API 是否支持
6. **i18n 翻译批次** — 中文/日文同期新增?

---

## 9. 后续行动(待用户决定后)

1. 决定 Phase 1 / Phase 2 / Phase 3 范围与顺序
2. 决定 select-dialog-design.md 的 Option A/B/C(与 Phase 3 协调)
3. 决定 MarkedIDs 隔离方案 γ-A/B
4. 决定批量 Stop Force 语义(Stop 真 timeout=0 vs Kill 重映射)
5. 决定 Podman Signal 同步范围
6. 实施并 PTY 视觉验证