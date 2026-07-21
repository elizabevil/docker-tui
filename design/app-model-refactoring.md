# AppModel 状态域重构方案

> 对应任务: [TASK-015](future-requirements-task-list.md#当前架构任务)
> 状态: `approved / in_progress`
> 最近更新: 2026-07-21

## 1. 背景

`AppModel` 是 Bubble Tea 的顶层状态，但当前同时承担连接、导航、输入、资源列表、日志、详情、对话框、Exec、Compose、通知、统计和动画状态。字段虽然按注释分组，所有权仍然是平铺的。

当前主要问题：

1. 修改一个功能时需要了解大量无关字段和模式。
2. 相关字段由调用方逐个赋值，容易产生只更新一半的非法状态。
3. `FilterText` / `FilterCursor` 被命令、镜像拉取和 Exec Shell 等不同输入复用。
4. `DialogFocus`、`DialogCursor` 的含义取决于当前 `Mode`，类型本身无法表达约束。
5. `state` 持有 `component.Spinner` 和 `component.ToastLevel`，形成 `state -> ui/component` 依赖。
6. 测试常用不完整的 `AppModel` 字面量，无法保证初始化不变量。

`ConnectionState` 已作为第一阶段从 `AppModel` 中拆出，并通过方法维护连接成功、失败、选择失败、探测和健康转换。当前匿名嵌入仅用于渐进迁移，不是最终 API。

## 2. 目标与非目标

### 2.1 目标

- `AppModel` 只负责组合状态域和协调 Bubble Tea 更新。
- 每个状态域拥有自己的字段、初始化函数和状态转换方法。
- 调用方表达业务意图，例如 `Dialog.Open(spec)`，而不是逐字段赋值。
- 数据状态不依赖 UI 渲染组件、Lip Gloss 或 i18n 文案。
- 配置在 `config.Load()` 解码后完成校验，状态层直接使用有效配置。
- 每一步可以独立测试、提交和回退，不进行一次性重写。

### 2.2 非目标

- 不把每个状态域都实现为独立 Bubble Tea `tea.Model`。
- 不在本任务中重做页面布局、键位体系或 runtime API。
- 不为了减少字段数量创建只有一两个字段且没有不变量的结构体。
- 不在状态方法中执行网络请求、文件 IO、渲染或翻译。

## 3. 设计原则

### 3.1 状态拥有行为

字段组合存在不变量时，由结构体方法完成转换：

```go
dialog.Open(DialogSpec{Kind: DialogImagePull, Input: image})
dialog.MoveFocus(1)
dialog.Close()
```

不再使用：

```go
m.DialogTitle = "..."
m.DialogBody = "..."
m.DialogFocus = 0
m.DialogCursor = len(m.FilterText)
m.Mode = ModeImagePull
```

### 3.2 值对象提供自身能力

可以由结构体回答的问题不拆成外部工具函数。例如：

- `ConnectionSpec.Key()` 负责连接 identity。
- `RuntimeConn.Validate()` 负责单连接配置约束。
- `RuntimeHealthConfig.Interval()` 返回已验证的时长。
- `QueryInput.Insert()`、`DeleteBackward()`、`Reset()` 维护 rune-safe 光标。

只有跨状态域协调或产生副作用的逻辑留在 update / keyboard 层。

### 3.3 事实与展示分离

状态层保存事实和语义枚举，不保存已翻译文案或 UI 组件：

- 保存 `Notification{Level, MessageKey, Args}` 或领域错误，不保存渲染样式。
- 保存 spinner frame / active，不保存 `component.Spinner`。
- 保存 `RuntimeKind`、`ConnectionStatus`，不通过展示字符串判断分支。

### 3.4 显式所有权

最终结构使用命名字段：

```go
type AppModel struct {
	Dependencies Dependencies
	Navigation   NavigationState
	Connection   ConnectionState
	Resources    ResourceState
	Log          LogState
	Detail       DetailState
	Dialog       DialogState
	Exec         ExecState
	Compose      ComposeState
	Feedback     FeedbackState
	Metrics      MetricsState
	Viewport     ViewportState
}
```

迁移期间允许匿名嵌入以减少一次性调用点修改；对应阶段完成后必须改成命名字段，例如 `m.Connection.Connected`，避免字段来源再次变得模糊。

## 4. 状态域边界

| 状态域 | 主要字段 | 应提供的方法 | 不负责 |
|---|---|---|---|
| `Dependencies` | Config、Theme、Audit、AppVersion | 配置只读访问 | 业务状态、渲染 |
| `NavigationState` | ActivePanel、PrevPanel、Mode、FilterInput、退出窗口 | `OpenPanel`、`EnterMode`、`LeaveMode`、`BeginFilterExit` | 加载资源、绘制页面 |
| `ConnectionState` | Client、Pool、连接状态、目标、错误、健康和选择框 | `Begin`、`ConnectedTo`、`Failed`、`SetProbeResult`、`ApplyHealthResult` | 建立网络连接、Toast |
| `ResourceState` | Containers、Images、Volumes、Networks、MarkedIDs | `ResetForRuntime`、`ClearMarks` | runtime API 调用 |
| `LogState` | ContainerID、Lines、Offset、Search、Wrap | `Open`、`Append`、`ApplySearch`、`Scroll`、`Close` | 拉取日志、渲染高亮 |
| `DetailState` | Resource、TitleKey、Raw、StructuredData、Source、Offset | `Open`、`SetData`、`SetError`、`SwitchSource`、`Scroll`、`Close` | Inspect API、YAML/JSON 渲染 |
| `DialogState` | Kind、Body、Preview、Input、Focus、Action | `Open`、`EditInput`、`MoveFocus`、`Confirm`、`Close` | 执行业务动作 |
| `ExecState` | SessionID、Conn、Output、Buffer、Scroll、Shell、Audit | `Start`、`Append`、`Scroll`、`Finish`、`Reset` | Hijack、读写 goroutine 生命周期 |
| `ComposeState` | 项目/服务光标、过滤、焦点、容器子视图 | `SelectProject`、`SelectService`、`OpenContainers`、`Reset` | labels 聚合和 API 请求 |
| `FeedbackState` | 当前通知、计时、错误计数、操作摘要、按键提示 | `Show`、`Tick`、`Clear`、`RecordError` | 样式和最终文案排版 |
| `MetricsState` | StatsActive、HostCPU、内存、磁盘 | `ApplyHostStats`、`EnableContainerStats` | 采样 IO |
| `ViewportState` | Width、Height、HeaderVisible | `Resize` | 页面布局计算 |

不是所有数据都必须封装为方法。纯快照且没有状态转换约束的数据可以直接赋值；只有成组变化、需要校验或存在生命周期的字段必须通过方法维护。

## 5. 输入模型

共享 `FilterText` 是当前最需要消除的隐式耦合。目标是每个交互拥有自己的 `QueryInputState`：

```go
type QueryInputState struct {
	Text   []rune
	Cursor int
}

func (q *QueryInputState) Insert(text string)
func (q *QueryInputState) DeleteBackward()
func (q *QueryInputState) DeleteForward()
func (q *QueryInputState) Move(delta int)
func (q *QueryInputState) Reset()
func (q QueryInputState) String() string
```

建议归属：

| 输入 | 所有者 |
|---|---|
| 资源过滤 | `Navigation.FilterInput` |
| 日志搜索 | `Log.Search.Input` |
| 命令模式 | `Navigation.CommandInput`，后续可独立为 `CommandState` |
| 镜像拉取引用 | 对应 `Dialog.Input` |
| Exec Shell | `Dialog.Input`，确认后写入 `Exec.Shell` |
| Compose 项目/服务过滤 | `Compose.ProjectInput` / `ServiceInput` |

输入方法只编辑文本和光标，不判断当前按键绑定。按键到编辑动作的映射仍属于 `keyboard`。

## 6. Mode 与覆盖层

本任务暂时保留 `AppMode`，避免同时重写所有键盘分发。拆分过程中遵循：

1. `Mode` 只用于顶层路由，不再作为子状态字段含义的唯一来源。
2. Dialog 使用 `DialogKind` 区分确认、拉取、导出、Debug、Exec Shell 等语义。
3. Detail、Log、Exec 是否打开由各自状态显式表达，不能仅通过 `Mode` 推导。
4. 子状态关闭时必须清空其生命周期数据，顶层再切换 `Mode`。

所有状态域完成后，再评估将 `AppMode` 收敛为 `BaseMode + OverlayKind`。该评估不阻塞当前拆分。

## 7. 副作用边界

状态方法必须是同步、确定且容易测试的。调用链保持：

```text
keyboard / update controller
  -> 调用状态方法表达意图
  -> 返回 tea.Cmd 执行 runtime / IO
  -> 消息回到 update
  -> 调用状态方法应用结果
  -> UI 只读取状态并渲染
```

以下逻辑不得进入状态方法：

- `tea.Cmd`、`tea.Tick` 创建。
- Docker / Podman API 调用。
- `net.Conn.Read/Write/Close` 等副作用；状态只管理引用和生命周期事实。
- i18n 翻译和终端宽度渲染。
- Toast、Footer、Header 的最终字符串拼接。

## 8. 配置边界

`NewAppModel` 接收的配置必须来自 `config.Load()`，或在测试中使用通过校验的 `config.DefaultConfig()`。生产代码不允许用零值配置启动后再在 tick / handler 中回退非法参数。

约束如下：

1. 每个配置子结构实现自己的 `Validate()`。
2. `Config.Validate()` 负责组合校验并添加字段路径上下文。
3. 时长、identity 等派生值通过结构体方法获取。
4. runtime handler 不再检查 `<= 0` 后替换默认值。
5. 测试需要非法配置时，应直接测试加载或校验失败，不把非法配置传入运行态。

## 9. 迁移阶段

### Phase 1：ConnectionState（已完成）

- 连接、选择器和健康字段下沉。
- 建立连接状态转换方法。
- 配置健康参数改为加载后校验。
- 当前仍匿名嵌入，待最终清理为命名字段。

### Phase 2：输入、Navigation 与 Dialog

- 状态: `done`
- 已完成: 为 `QueryInputState` 增加 rune-safe 编辑方法；删除 `FilterText` / `FilterCursor` 共享输入；建立迁移期 `NavigationState` 与 `DialogState`；命令、Filter、Search、镜像 Pull 和 Exec 使用独立输入；使用 `DialogKind` 和 `DialogSpec` 收敛对话框语义。
- 后续清理: 迁移期匿名嵌入统一在 Phase 6 改为命名字段。

实施范围：

- `DialogKind` 是对话框语义来源，`Mode` 仅保留顶层路由职责。

这是下一阶段优先项，因为它影响命令、镜像拉取和 Exec，并且是当前最明显的跨功能耦合。

### Phase 3：Feedback 与 UI 依赖清理

- 拆 `FeedbackState`。
- 合并现有 Toast 字段与未接入的 `ToastQueue` 方案，只保留一套通知状态。
- 用 state 层语义类型替代 `component.ToastLevel`。
- 将 spinner 数据与渲染对象分离，移除 `state -> ui/component` 依赖。

### Phase 4：Detail 与 Log

- 拆 `DetailState`、`LogState`。
- 将打开、加载、失败、滚动、搜索和关闭实现为显式转换。
- 保持 Detail / Log 页面模板和现有用户交互不变。

### Phase 5：Exec 与 Compose

- 拆 `ExecState`、`ComposeState`。
- controller 继续负责 goroutine、连接读写和 runtime 请求。
- 为启动、结束、返回和异常退出补资源释放测试。

### Phase 6：其余状态与显式组合

- 拆 `ResourceState`、`MetricsState`、`ViewportState`。
- 将 `ConnectionState` 从匿名嵌入改为命名字段。
- 删除迁移期兼容字段和代理方法。
- 更新架构文档中的最终状态树。

## 10. 每阶段实施步骤

每个状态域遵循同一迁移模板：

1. 搜索该组字段的全部读写点并记录行为矩阵。
2. 新增子状态结构、构造函数和纯状态转换方法。
3. 先迁移写入点，再迁移读取点。
4. 将 UI 调整为只读取新状态，不在渲染函数中修正光标或数据。
5. 删除旧字段及兼容路径。
6. 补充子状态单元测试和相关 handler / renderer 回归测试。
7. 执行 `gofmt`、`git diff --check`、`just check`。
8. 独立提交，并更新 `TASK-015` 的已完成范围。

## 11. 测试策略

### 11.1 子状态测试

每个子状态至少覆盖：

- 零值或构造后的有效初态。
- 正常打开、更新和关闭。
- 重复操作的幂等性。
- 越界光标、空数据、过期异步结果。
- 失败后是否保留应保留的数据、清理应清理的数据。

### 11.2 集成回归

| 场景 | 重点 |
|---|---|
| Runtime 切换 | 成功后清理旧资源瞬态状态，失败时选择框保持打开 |
| Filter / Search | 输入独立、rune-safe、退出不污染其他模式 |
| Dialog | 不同 Kind 的输入、焦点和关闭清理互不串联 |
| Detail / Log | 打开、滚动、返回后恢复正确页面 |
| Exec | 正常结束、异常结束、Esc 返回和资源释放 |
| Compose | 项目/服务/容器子视图光标独立且越界安全 |
| i18n | 中英文切换不改变状态分支和布局高度 |

禁止只依赖完整 `AppModel` 字面量测试子状态。子状态行为优先直接测试其构造函数和方法；顶层测试只验证跨域协调。

## 12. 完成标准

`TASK-015` 完成必须同时满足：

1. `AppModel` 只保留命名状态域和少量顶层协调依赖。
2. Connection、Navigation、Dialog、Feedback、Detail、Log、Exec、Compose 均有独立结构和生命周期方法。
3. 不再存在 `FilterText` / `FilterCursor` 这类跨功能共享输入。
4. `internal/tui/state` 不再导入 `internal/tui/ui/component`。
5. 状态方法不创建 `tea.Cmd`，不执行 runtime / IO，不渲染和翻译。
6. 生产调用不依赖不完整的 `AppModel` 字面量。
7. 所有阶段测试、`git diff --check` 和 `just check` 通过。
8. [架构说明](../docs/architecture.md) 与任务台账更新为最终状态树。

## 13. 风险与控制

| 风险 | 控制方式 |
|---|---|
| 一次修改大量调用点造成回归 | 按状态域迁移，每阶段独立提交 |
| 匿名嵌入长期保留，所有权仍不清晰 | Phase 6 强制改为命名字段 |
| 把副作用塞入状态方法 | 状态方法限制为同步转换，controller 管理 `tea.Cmd` |
| 为拆分而创建空壳结构 | 只有存在生命周期、不变量或成组字段时才拆 |
| Mode 与子状态出现双重真相 | 迁移期由 controller 原子更新，最终评估 BaseMode / OverlayKind |
| UI 为修复越界直接修改状态 | 越界修正移入状态方法，renderer 保持只读 |
