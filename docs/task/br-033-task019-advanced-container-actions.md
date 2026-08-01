# TASK-019: 容器高级动作 (Copy / Update / Diff / Export / Commit / Wait)

> 任务卡 / TASK-019 / BR-033

## 元信息

- **关联编号**:TASK-019 + BR-033
- **优先级**:high
- **状态**:runtime 已 done,TUI handler 缺失 → `open`
- **依赖**:TASK-017(高频容器操作)、TASK-021(统一 runtime)
- **关联设计**:[docs/feature-design.md §2.1](../feature-design.md)、[bugfix-requirements.md BR-033](../bugfix-requirements.md)

## 目标

为容器页接入 6 个 TASK-019 高级动作(Copy / Update / Diff / Export / Commit / Wait),runtime 层已实现,只缺 TUI 侧的:

1. `case keys.ActionContainer*` 在 `keyboard/actions.go:handleAction` 中的分支
2. 对应的 `doContainerXxx(m)` 函数 + `containerXxxCmd(client, id, opts)` Cmd 工厂
3. 必要的 dialog / 表单(部分动作需要用户输入参数)
4. Help / Footer 文案
5. keymap / 配置 schema 字段(已存在,需确认)

## 代码结构索引

### 必须读懂的文件

| 文件 | 作用 |
|---|---|
| `internal/tui/keys/action.go` | 6 个 `ActionContainer*` KeyAction 已定义(line 23-28) |
| `internal/tui/keys/registry.go` | 6 个默认键位已注册(line 66-71,**有冲突**,见 §风险) |
| `internal/tui/keys/keys.go` | `KeymapConfig` 字段(ContainerUpdate/Diff/Export/Commit/Wait/Copy)已声明 |
| `internal/tui/keyboard/actions.go` | `handleAction` switch — **6 个 case 全部缺失**(line 16-126) |
| `internal/tui/keyboard/container_action.go` | 现有容器动作的 doXxx 模板(`doContainerAction` 模式) |
| `internal/tui/keyboard/container_advanced.go` | (若已存在)运行时调用包装 |
| `internal/data/runtime/actions.go` | `ActionUpdate/Diff/Export/Commit/Wait/Copy` 字符串常量 + `ActionOptions` 联合类型 |
| `internal/data/runtime/docker/service_container.go` | Docker `ContainerService` 完整实现 6 个动作 |
| `internal/data/runtime/podman/service_container.go` | Podman `ContainerService` 完整实现 6 个动作 |
| `internal/data/runtime/docker/service_action.go` | Docker `ResourceActionService.Execute` 分发(line 71-78 含 `ActionRename/Kill` 分发样板) |
| `internal/data/runtime/podman/service_action.go` | Podman `ActionUpdate/Diff/Export/Commit/Wait/Copy` 分发(line 63-110) |
| `internal/data/runtime/containers.go` | `ContainerService` 接口定义(已含 6 个方法 line 235-241) |
| `internal/tui/ui/widget/dialog/choice.go` | 已有 dialog 模板(可复用做参数表单) |
| `internal/tui/ui/widget/dialog/view.go` | `RenderExec`/`RenderConfirm` 模板 |
| `internal/tui/ui/action/registry.go` | 帮助页 / footer 投影(按 `ActivePanel` + `Mode`) |

### 必须修改的文件

| 文件 | 改动 |
|---|---|
| `internal/tui/keyboard/actions.go` | 加 6 个 `case keys.ActionContainer*` 分支(若不抽 doContainerXxx 则直接写在这里) |
| `internal/tui/keyboard/container_action.go`(或新增 `container_advanced.go`) | 加 `doContainerCopy/Update/Diff/Export/Commit/Wait` + `containerXxxCmd` 工厂 |
| `internal/tui/state/containers.go`(视需要) | 加容器详情页"Copy/Update 状态"或 task 进度 state 字段 |
| `internal/tui/state/dialog.go`(若需 dialog state) | 加 Copy/Update/Commit 等的 dialog kind + input state |

### 必须新增的文件

| 文件 | 作用 |
|---|---|
| `internal/tui/ui/widget/dialog/container_diff.go` | Diff 视图:filesystem 改动列表(替代 dialog.go 的临时实现) |
| `internal/tui/ui/widget/dialog/copy.go`(或合入 choice.go) | Copy / Export / Wait 表单:源路径 / 目标 tar 路径 / 阻塞条件 |
| `internal/tui/ui/widget/dialog/update.go` | Update 表单:Memory / NanoCPUs / RestartPolicy / RestartMaxRetries |
| `internal/tui/ui/widget/dialog/commit.go` | Commit 表单:repository:tag / Comment / Author / Pause |

### 不应触碰的文件

- runtime 层(已实现,只通过 `Execute(ctx, ref, action, opts)` 调用)
- 键位注册表(只调整冲突键位,不破坏其它动作)
- 全局键盘分发路径

## 操作链路

### Copy (容器 ↔ 主机文件)

```text
User 按 Ctrl+O
  → keys/registry.go:71 解析 ActionContainerCopy
  → keyboard/actions.go:handleAction 命中 case (缺失,需新增)
  → doContainerCopy(m) 函数 (缺失)
      - 打开 Copy 表单 (widget/dialog/copy.go)
      - 用户输入 srcPath / dstPath
      - 提交后构造 ActionOptions{Copy: &CopyOptions{SourcePath: srcPath}}
      - 调 containerCopyCmd(client, id, opts) → Cmd
  → tea.Cmd 异步执行
      - Engine.Containers().CopyFromContainer(ctx, id, srcPath) 或 CopyToContainer
      - Podman: service_container.go:237
      - Docker: service_container.go:97 (container_advanced.go)
  → 返回 io.ReadCloser;前端保存为 tar 写到 dstPath
  → 更新 toast / audit
```

### Update (运行时改 Memory / CPU / RestartPolicy)

```text
User 按 Ctrl+W
  → keyboard/actions.go:handleAction 命中 case (缺失)
  → doContainerUpdate(m) 函数 (缺失)
      - 打开 Update 表单 (widget/dialog/update.go)
      - 用户输入 Memory(MB)/ NanoCPUs / RestartPolicy / RestartMaxRetries
      - 提交后构造 ActionOptions{Update: &UpdateOptions{...}}
      - 调 containerUpdateCmd(client, id, opts) → Cmd
  → tea.Cmd 异步执行
      - Docker: ContainerUpdate(ctx, id, options) → client.ContainerUpdate
      - Podman: service_container.go:132 svc.Update(ctx, id, ...)
  → toast + audit
```

### Diff (filesystem 改动)

```text
User 按 F2(注:与 SwitchRuntime 冲突,见 §风险)
  → doContainerDiff(m) → containerDiffCmd
  → Cmd 异步执行
      - Engine.Containers().Diff(ctx, id) → []ContainerDiffChange
      - Docker: service_container.go:76 ContainerDiff(ctx, id)
      - Podman: service_container.go:161 svc.Diff(ctx, id)
  → 返回 []DiffChange → state.Detail.DiffResult 或新 state.DiffState
  → 打开 Diff 视图(类似详情页,但只显示 fs 改动)
  → Esc 返回
```

### Export (容器 → tar)

```text
User 按 Ctrl+X
  → doContainerExport(m) → 打开 Export 表单(目标 tar 路径)
  → 提交 → containerExportCmd
      - Docker: ContainerExport(ctx, id) → io.ReadCloser
      - Podman: service_container.go:181 svc.Export(ctx, id)
  → 写入目标 tar 文件
  → toast + audit
```

### Commit (容器 → 新镜像)

```text
User 按 Ctrl+K(注:与 ContainerKill 冲突)
  → doContainerCommit(m) → 打开 Commit 表单(repository:tag / comment / author / pause)
  → 提交 → containerCommitCmd
      - Docker: ContainerCommit(ctx, id, options)
      - Podman: service_container.go:194 svc.Commit(ctx, id, options)
  → 新镜像 ID 返回 → toast + audit
```

### Wait (阻塞等待)

```text
User 按 Ctrl+Y
  → doContainerWait(m) → 打开 Wait 表单(条件:not-running / next-exit / removed)
  → 提交 → containerWaitCmd
      - Docker: ContainerWait(ctx, id, condition)
      - Podman: service_container.go:217 svc.Wait(ctx, id, condition)
  → 阻塞等待条件触发,返回 exit code
  → toast / 自动跳转到详情页刷新
```

## 验收标准

- [ ] 容器页按 Ctrl+O(Copy)、Ctrl+W(Update)、F2(Diff)、Ctrl+X(Export)、Ctrl+K(Commit)、Ctrl+Y(Wait) 全部能弹出对应表单(dialog)或直接执行。
- [ ] 6 个动作的 `case` 在 `handleAction` switch 中存在,不再 fall-through 到默认 return。
- [ ] Docker 与 Podman 两边都接好 runtime 调用;Docker 端用 client SDK,Podman 端用 service_action.go:63-110 模板。
- [ ] 每个动作执行成功后:
  - 更新 toast(`✓ <Action> <containerID>`)
  - 写 audit(`resource.container.<action>`)
  - 必要时刷新资源列表(Start/Stop/Restart/Kill/Remove 不需要,Copy/Update/Diff/Export/Commit/Wait 大多数也不需要,但 Export/Commit 后建议 `Refresh`)。
- [ ] 每个动作失败时显示明确错误 toast,不静默吞。
- [ ] 帮助页 / Footer 显示这 6 个动作的当前键位。
- [ ] Help 文案区分 Diff / Update / Commit 等易混淆的动名。
- [ ] 解决键位冲突(F2 与 SwitchRuntime、Ctrl+K 与 ContainerKill、Ctrl+O 与排序 O)。

## 风险

| 风险 | 说明 |
|---|---|
| **键位冲突** | `F2` 已用于 SwitchRuntime;`Ctrl+K` 已用于 ContainerKill;`Ctrl+O` 已用于排序 O;这 6 个新动作的默认键位必须重排。建议:TASK-019 走 Action Bar (BR-039) 触发,不直接占快捷键 |
| **runtime 双适配器一致性** | Docker SDK + Podman REST 必须都接;EngineActions() 走 Execute 统一,确认两边都注册了对应分支(`service_action.go` Podman 已写,Docker 需复核) |
| **Docker exec 缺失** | `docker/service_action.go:63-110` 是 Podman 端的;Docker 端需要确认 `Execute` 分发是否含 `ActionCopy/Update/Diff/Export/Commit/Wait` 全部 6 个 |
| **Exec/Copy 流关闭** | CopyFromContainer 返回 `io.ReadCloser`(tar stream),必须正确关闭避免 FD 泄漏 |
| **Help/Footer 漂移** | 6 个新动作 + 重排键位会导致 Footer 行动提示变化,需同步 docs/navigation.md |
| **i18n 漂移** | 每个动作的 toast / dialog 标题 / 按钮都需要 i18n key,参考 `internal/data/i18n/lang/zh.jsonc` 添加 |
| **批量操作语义** | 容器页 mark 模式下,这 6 个动作是否支持批量?与现有 mark_action.go 的批量模式如何交互?见 BR-039 Action Bar 决策 |

## 建议任务分解

按"最小可交付单元"拆分:

1. **TASK-019-A:Update + Wait 两个无表单动作(无新 dialog)**
   - `case ActionContainerUpdate` / `ActionContainerWait` + `doContainerUpdate` / `doContainerWait` + Cmd 工厂
   - 表单复用现有 choice.go 或简单 inline input
   - **预估**:小模型独立可完成
2. **TASK-019-B:Export + Commit 两个简单表单动作**
   - 表单 widget:dialog/export.go + dialog/commit.go
   - **预估**:小模型独立可完成
3. **TASK-019-C:Copy 动作(复杂,流式读写)**
   - CopyFromContainer 返回 ReadCloser,需写 tar 到本地文件
   - 表单 widget:dialog/copy.go
   - **预估**:小模型辅助,主模型审流处理
4. **TASK-019-D:Diff 动作(视图独占)**
   - 新增 Diff 视图 widget:container_diff.go
   - 类似 detail 页 + revision/source 缓存
   - **预估**:小模型只读定位,主模型实现
5. **TASK-019-E:键位冲突解决 + Help/Footer + i18n 同步**
   - 把 6 个动作改为 Action Bar 触发(BR-039 协同)
   - docs/navigation.md 同步
   - **预估**:小模型机械同步,主模型决策键位

## 待确认项

- [ ] **键位方案**:直接绑键 vs Action Bar?Action Bar (BR-039) 还没开始;如果直接绑键,F2/Ctrl+K/Ctrl+O 三个冲突键必须重排。建议等 BR-039 落地。
- [ ] **Copy 双向 vs 单向**:CopyToContainer 是否同时支持?Docker / Podman 两边都支持 CopyToContainer,需要额外实现。
- [ ] **Diff 显示位置**:详情页 section vs 独立 ModeDiff?独立更清晰,与 Logs 视图类似。
- [ ] **Wait 阻塞语义**:ContainerWait 是阻塞调用,tea.Cmd 中阻塞时间可能很长(容器未退出前不返回)。需要确认 tea.Cmd 的超时 / 取消机制是否能优雅处理。
- [ ] **Commit 后的镜像刷新**:Commit 完成后新镜像是否立即出现在镜像列表?需要触发 Refresh 还是 Engine 主动推送 events?
- [ ] **批量模式语义**:BR-039 设计的 Action Bar 是单选还是支持批量?