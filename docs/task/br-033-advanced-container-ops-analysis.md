# BR-033 高级容器操作分析(只读,不改代码)

> 范围: TASK-019 6 个动作(Update / Diff / Export / Commit / Wait / Copy)的接入策略
> 输出: Rename 可复用结构 + 6 动作的 Dialog/Confirm/Result 分类 + Action Bar 接入顺序
> 配套: BR-034 Runtime + UI-State 分析 / 当前任务卡 [docs/task/br-033-task019-advanced-container-actions.md](br-033-task019-advanced-container-actions.md)

## 1. 当前状态盘点

### 1.1 runtime 层(已 done)

- `internal/data/runtime/actions.go:43-48`: 6 个 `Action` 常量 + `ActionOptions` 联合类型 + 各 `XxxOptions` 子结构(line 73-118)已就位
- `internal/data/runtime/containers.go:235-241`: `ContainerService` 接口已含 6 个方法(Update / Diff / Export / Commit / Wait / Copy)
- `internal/data/runtime/docker/service_container.go:71-96`: Docker adapter 6 方法已实现
- `internal/data/runtime/podman/service_container.go:132-238`: Podman adapter 6 方法已实现
- **`actions.go:17 case ActionQuit: return m, tea.Quit` 之后到 line 123,完全没有 6 动作的 case 分支**(grep 验证:无 ActionContainer* 的 case)

### 1.2 keys 层(已注册 + 配置)

- `keys/action.go:23-28`: `ActionContainerUpdate/Diff/Export/Commit/Wait/Copy` 6 个 KeyAction 已定义
- `keys/keys.go`: `keymap.ContainerUpdate/Diff/Export/Commit/Wait/Copy` 字段已存在(配置可重映射)
- `keys/registry.go:66-71`: 6 个默认键位已注册(见任务卡冲突分析):
  - `Ctrl+W` Update
  - `F2` Diff(与 SwitchRuntime 冲突)
  - `Ctrl+X` Export
  - `Ctrl+K` Commit(与 ContainerKill 冲突)
  - `Ctrl+Y` Wait
  - `Ctrl+O` Copy(与 sort column 冲突)

### 1.3 现有 Dialog 模式(可复用模板)

| Dialog | 入口 | 关键函数 | 复用点 |
|---|---|---|---|
| Rename | `openRenameDialog(m)` | action.go:127 | 单行 input + confirm |
| Port | `openPortDetail(m)` | action.go:131 | 列表 + 选择 |
| Resource Create (Volume/Network) | `openResourceCreate(m, type)` | action.go:105/111 | 多行 form 表单 |
| Image Workflow (Tag/Push/Save/Load) | `openImageWorkflow(m, op)` | action.go:94-100 | 单行 input + 进度条 |
| Top (Process list) | `openTopView(m)` | top.go | 整页 + 实时拉取 |
| Confirm (Stop/Restart/Force) | `confirmAction(m, ...)` | container_action.go:37 | 二选一选项 |

## 2. Rename 模式可复用结构

`openRenameDialog`(`internal/tui/keyboard/rename.go` + `state/dialog.go`)的核心:

- 设置 `m.Dialog.Kind = state.DialogContainerRename` + `Open(spec)` 一次性写
- Mode 设为 `m.Dialog.Kind.Mode()` = `ModeRename`(由 dialog state 自动路由)
- 用户输入字符 → `m.Dialog.Input.Text/Cursor` 累积
- Enter → 写回资源 + 关闭 dialog
- Esc → `clearDialogState(m)`

**BR-033 6 动作可复用的同一模式**:
- 每个动作设独立 `DialogKind`(如 `DialogContainerUpdate` / `DialogContainerDiff` / ...)
- `Mode` 自动从 Kind 派生
- input/select 走 `m.Dialog.Input`
- Enter 触发 `svc.Execute(ctx, ref, ActionXxx, opts)`
- 进度条复用 `ModeImageTransfer` 模式

## 3. 6 动作的 Dialog/Confirm/Result 分类

| 动作 | 选项复杂度 | UI 类型 | Result 显示 | 进度条 |
|---|---|---|---|---|
| **Update** | 4 字段(Memory/NanoCPUs/RestartPolicy/MaxRetries) | 多行 form | toast OK,无独立页 | 无(同步) |
| **Diff** | 无参数 | 整页 result viewer | **需要独立视图**(类似 detail 页) | 无(同步) |
| **Export** | 1 字段(目标 tar 路径) | 单行 input | toast OK | 同步,无进度条 |
| **Commit** | 4 字段(repo/tag/comment/author) + Pause bool | 多行 form | toast OK(返回新镜像 ID) | 无(同步) |
| **Wait** | 1 字段(condition:not-running/next-exit/removed) | 选择/单行 | toast OK(exit code) | **有**(长阻塞) |
| **Copy** | 2 字段(src path / dst path) + 方向(To/From) | 多行 form | 进度条(流式) | **有**(大文件流) |

### 3.1 详细分类(实施参考)

**Update**(4 表单字段):
- Reuse `internal/data/i18n/lang/zh.jsonc` 现有 memory/cpu/restart 翻译 key
- 4 行 form:Memory(MB)/ CPU cores / Restart policy(select) / Max retries
- 不需要 confirm(destructive? 实际是 hot-reload,只对运行中容器有意义,失败不致命)
- **建议:不确认 dialog,失败 toast**

**Diff**(无参数,但有 result):
- 同步返回 `[]ContainerDiffChange`(A/M/C 三种类型 + Path)
- **必须独立视图页**(`ModeContainerDiff` 或复用 `DetailState.Documents`)
- 类似 detail 页:列表 + section(Added/Modified/Deleted)
- 与 detail 页共享 `m.Detail.Documents[DetailSourceSection]` 缓存机制

**Export**(单字段):
- 1 行 input:目标 tar 路径(默认 `<containerID>.tar`)
- 不需要 confirm
- 成功 toast

**Commit**(4 字段):
- 4 行 form:Repository / Tag / Comment / Author + Pause 复选
- 成功 toast(显示新 image id)
- 不需要 confirm

**Wait**(单字段 + 进度条):
- Select(condition: not-running / next-exit / removed)
- 阻塞调用,长耗时
- **需要进度条**(类似 ImageTransfer 模式)
- 完成后 toast 显示 exit code

**Copy**(2 字段 + 方向 + 进度条):
- 3 行 form:Src path(容器内)/ Dst path(主机) / Direction(To container | From container)
- 流式读写,**需要进度条**
- 类似 `image_transfer.go` 模式

## 4. Action Bar 接入顺序(小模型 B 需配合)

按依赖关系 + 风险排序:

| 顺序 | 动作 | 风险 | Action Bar 项 |
|---|---|---|---|
| 1 | **Commit** | 中(4 字段,简单 form) | `Commit (Ctrl+K)` |
| 2 | **Export** | 低(1 字段) | `Export (Ctrl+X)` |
| 3 | **Update** | 中(4 字段,需 hot-reload 状态) | `Update (Ctrl+W)` |
| 4 | **Wait** | 中(进度条) | `Wait (Ctrl+Y)` |
| 5 | **Copy** | 中(进度条 + 路径) | `Copy (Ctrl+O)` |
| 6 | **Diff** | 高(独立视图页 + Detail 共享) | `Diff (F2)` |

**理由**:
- Commit / Export 风险最低(form 简单,同步返回),先打通 form-dialog 模板
- Update hot-reload 需确认运行中容器状态,中等风险
- Wait / Copy 进度条模式相同(imageTransfer 已实现),参考
- Diff 风险最高(需独立视图页 + 共享 Detail 缓存),最后做

**键位冲突的解决**: 任务卡已标"F2 与 SwitchRuntime 冲突 / Ctrl+K 与 ContainerKill 冲突 / Ctrl+O 与 sort 冲突"—— 建议 **Action Bar 触发,无键位冲突**(与 BR-039-A 设计一致)。6 动作通过 Action Bar 在容器页触发,不直接绑键位。

**Action Bar 正确路径(主模型拍板)**:`internal/tui/actionbar/registry.go`(不是之前 BR-039 用错的 `ui/widget/actionbar/`)。容器页加 6 项在主路径下做,小模型 B 不动这一文件。

## 5. 复用现成组件(小模型 B 不重复造)

- `internal/tui/ui/widget/dialog/choice.go` — `ChoiceDialog` (2-3 选项,如 Wait 条件)
- `internal/tui/ui/widget/dialog/centered.go` — `PlaceDialogInPanel` 居中(BR-040)
- `internal/tui/actionbar/registry.go` — 容器页加 6 项(主模型 B 集成;**注意正确路径,不是 `ui/widget/actionbar/`**)
- `internal/tui/ui/widget/dialog/notification.go` — `NotificationDialog`(单信息,所有动作成功后用)
- `image_transfer.go` 模式(tea.Tick + 进度条 + 完成 toast) — Wait / Copy 复用

## 6. 文件清单

### 6.1 必须新增 (6)

| 文件 | 作用 | 约行数 |
|---|---|---|
| `internal/tui/ui/widget/dialog/container_update.go` | Update 表单(4 字段) | ~120 |
| `internal/tui/ui/widget/dialog/container_diff.go` | Diff result viewer(独立视图) | ~150 |
| `internal/tui/ui/widget/dialog/container_export.go` | Export 单字段 input | ~80 |
| `internal/tui/ui/widget/dialog/container_commit.go` | Commit 表单(4 字段 + Pause) | ~120 |
| `internal/tui/ui/widget/dialog/container_wait.go` | Wait 选择 + 进度条 | ~100 |
| `internal/tui/ui/widget/dialog/container_copy.go` | Copy 表单 + 进度条 | ~120 |

(每动作一个 dialog 文件,共 6 个)

### 6.2 必须修改 (3)

| 文件 | 改动 |
|---|---|
| `internal/tui/state/dialog.go` | 加 6 个 `DialogKind` 常量(`DialogContainerUpdate` / `DialogContainerDiff` / `DialogContainerExport` / `DialogContainerCommit` / `DialogContainerWait` / `DialogContainerCopy`) + `Mode()` switch case |
| `internal/tui/state/app.go` | 加 6 个 `AppMode` 枚举(`ModeContainerDiff` 等),不需新增的(Commit/Export/Update/Copy/Wait 是 dialog mode;Diff 需独立 Mode) |
| `internal/tui/keyboard/dialog_keys.go`(新,聚合) | `handleContainerAdvancedDialogKeys` 调度 6 个 dialog 内的按键 |
| `internal/tui/keyboard/container_action.go` | 加 6 个 `doContainerXxxAction` 工厂,沿用 `doContainerAction` 模式 |

### 6.3 不应触碰(主模型集成文件)

- `internal/tui/keys/action.go` / `keys/registry.go` — Action 定义 / 键位注册已就位
- `internal/tui/ui/action/registry.go` — Help / Footer 投影已配置
- `internal/tui/actionbar/registry.go` — Action Bar 项(**正确路径,不是 `ui/widget/actionbar/`**)
- `internal/tui/ui/app/page_templates.go` — page 路由
- `internal/tui/keyboard/keyboard.go` — `dispatchByMode` 全局分发
- `internal/data/runtime/*` — runtime API 已完整
- `internal/tui/keyboard/actions.go:case ActionQuit` — 保持 `tea.Quit`

## 7. 操作链路(Update 样例)

```
User 在容器页 → Action Bar → "Update (Ctrl+W)"
  → keys/action.go:ActionContainerUpdate 命中
  → keyboard/actions.go:case ActionContainerUpdate: doContainerUpdateAction(m)
  → doContainerUpdateAction(m):
      1. 校验 m.Connection.Engine + m.Navigation.ActivePanel == PanelContainers
      2. 校验 ctr := m.Resources.Containers.Selected() (无选 → toast)
      3. 校验 ctr.State == "running" (非 running 提示 toast)
      4. 打开 dialog: m.Dialog.Open(state.DialogSpec{Kind: DialogContainerUpdate, ...})
         m.Navigation.Mode = ModeContainerUpdate
  → dialog: 用户填 4 字段 → Enter
  → handleContainerAdvancedDialogKeys → dialog Update path:
      opts := ActionOptions{Update: &UpdateOptions{Memory: &mem, ...}}
      cmd := serviceActionCmd(m.Connection.Engine, ref, ActionUpdate, opts)
  → tea.Cmd 异步执行 ContainerService.Update(...)
  → 完成后 toast 成功/失败
```

## 8. 关键设计决策点

| 决策 | 选择 | 理由 |
|---|---|---|
| Update 失败回滚 | 不回滚(运行时资源限制失败不致命,提示重试) | 简单,符合 docker update 语义 |
| Diff 大文件量 | 截断显示 + "more" 提示 | 避免单次拉几 MB |
| Commit Pause | 默认 true(暂停期间提交) | 与 docker 默认行为一致 |
| Wait 阻塞时长 | 设 60s timeout,失败 toast 提示 | 不阻塞 UI 主循环 |
| Copy 双向 | 一个 dialog 用 select 切方向 | 不开两个 dialog |
| 6 动作都走 Action Bar 触发 | **不** 直接绑键位(避免 §1.2 冲突) | BR-039-A Action Bar 优先 |

## 9. 关键风险

| 风险 | 说明 | 缓解 |
|---|---|---|
| **键位冲突未解决** | F2/Ctrl+K/Ctrl+O 6 动作中 3 个冲突 | 全部走 Action Bar 触发,不绑默认键位(§8) |
| **form 字段超宽** | Update 4 字段在小 panel 显示不全 | 复用 `dialog.PlaceDialogInPanel` 居中 + `ClampWidth` 截断 |
| **Diff result 巨大** | 一个容器 fs 改动上万行 | 截断 + 提示"按 / 搜索" |
| **Wait 60s 阻塞** | 60s 内 UI 不响应 | 实际上 `tea.Cmd` 不阻塞 Update loop(已是异步) |
| **Copy 大文件流** | 流式读写超时 | 跟 `imageTransfer` 一样进度条 + 取消 |
| **Podman REST 6 端点 4.x / 5.x 差异** | 各端点可能版本不同 | `RESTClient.ensureVersion` 已有,沿用 |
| **Audit 字段缺失** | `audit.ContainerTarget` 不够新动作(Export 输出路径?) | 后续扩 `audit.Target` 类型,本轮用通用 ContainerTarget + Detail string |

## 10. 输出给主模型(实施时)

1. **DialogKind 加 6 个常量**(`state/dialog.go`)+ `Mode()` switch 分发
2. **action.go 不加新 case**——6 动作由 `widget/dialog/container_*.go` 内 Enter 处理后调 `serviceActionCmd`
3. **form 字段**只用现有 i18n key(memory/cpu/restart 已有),新增的字段名(repository/tag/comment 等)按 BR-039-D 集中加
4. **Mode**——Diff 需独立 Mode(`ModeContainerDiff`),其他 5 个复用 `ModeResourceCreate` 模式(dialog-only)
5. **不碰 actions.go:case ActionQuit**——主模型硬约束
6. **不绑默认键位**——6 动作只走 Action Bar 触发(避免 §1.2 冲突)
7. **Diff 是最复杂**——最后做;独立视图页 + 共享 Detail 缓存

## 11. 测试清单(本轮小模型 B 交付范围)

### 11.1 单元测试

- `internal/tui/ui/widget/dialog/container_update_test.go` — 4 字段 form 渲染
- `internal/tui/ui/widget/dialog/container_export_test.go` — 单字段 input
- `internal/tui/ui/widget/dialog/container_commit_test.go` — 4 字段 + Pause
- `internal/tui/ui/widget/dialog/container_wait_test.go` — select 选项
- `internal/tui/ui/widget/dialog/container_copy_test.go` — 3 字段
- `internal/tui/ui/widget/dialog/container_diff_test.go` — result viewer(包含 truncation 测试)

### 11.2 集成测试(端到端)

- `internal/tui/keyboard/container_action_advanced_test.go` — 6 动作 enter → handler 链 → toast/result
- Mock engine 验证:`svc.Execute(ctx, ref, ActionXxx, opts)` 收到正确参数

报告完成。**未修改任何代码**。
