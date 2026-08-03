# 未关闭的 BUG / 需求

> 建立日期: 2026-08-01
> 来源: [bugfix-requirements.md](bugfix-requirements.md) 的快照
> 目的: 把 `bugfix-requirements.md` 中**当前未完成**的 BUG / 需求单独切出,
>       作为"待开发工作项"独立跟踪,不与已完成条目混在一起。

## 状态总览

| 状态 | 数量 | 备注 |
|---|---|---|
| `open` | 18 | 未开始或被搁置的需求 |
| `implementing` | 1 | 修复进行中(部分子任务已完成) |
| `pending` | 1 | 用户尚未提供具体内容 |
| `wontfix` | 1 (BR-012 已转移为新需求 BR-030,保留记录) |

## 排序与优先级

按"业务影响 + 优先级"降序:

| 编号 | 标题 | 优先级 | 状态 |
|---|---|---|---|
| [BR-009](#br-009) | 卷详情加载 + 容器子视图选中(部分修复) | high | implementing |
| [BR-025](#br-025) | 容器详情页标题占位符 `{0}` `{1}` 未替换 | high | open |
| [BR-030](#br-030) | 取消鼠标点击行选中,改用滚轮上下选行 | high | open |
| [BR-033](#br-033) | TASK-019 高级容器动作未实现(Copy/Update/Diff/Export/Commit/Wait) | high | open |
| [BR-034](#br-034) | 镜像 History 顶层页 (H 键) | high | open |
| [BR-023](#br-023) | F2 唯一提供 runtime 选择器;其它页面不得用 C 刷新 Conn | medium | open |
| [BR-024](#br-024) | 详情源码视图 Ctrl+C 复制不完整 / 一次性失效 | medium | open |
| [BR-026](#br-026) | 详情页快捷键集合:支持 j/k/PgUP/PgDn/Filter,**不**支持 Space/R | medium | open |
| [BR-027](#br-027) | 容器日志页滚到底后向上划卡顿 | medium | open |
| [BR-029](#br-029) | 筛选框 Enter/Esc 双层退出 | medium | open |
| [BR-031](#br-031) | 容器资源占用(CPU/内存)未实时刷新 | medium | open |
| [BR-032](#br-032) | 表格排序 N/Ctrl+N/鼠标点击表头 | medium | open |
| [BR-035](#br-035) | Events 独立面板 (F3) + Network Connect/Disconnect | medium | open |
| [BR-036](#br-036) | 镜像 tarball 导入功能 | medium | open |
| [BR-037](#br-037) | docker / podman Registry Login | medium | open |
| [BR-038](#br-038) | 容器 Exec 页面 UI 不符合 shell 终端 | medium | open |
| [BR-040](#br-040) | Dialog 风格统一:四周透明 + panel 居中 | medium | open |
| [BR-008](#br-008) | H 键应进 Help;镜像页 H 进入 History | medium | open |
| [BR-028](#br-028) | (TBD - 待用户补充第 6 条) | - | pending |

---

## 详情

<a id="br-008"></a>

### BR-008 H 键应进 Help,F1 行为复用;H 键同时承担镜像 History 入口

- 状态: `open`
- 优先级: `medium`
- 症状:
  1. 在任何页面按 H 当前是 `Viewport.ToggleHeader()`(隐藏/显示 header),用户期望 H 直接进入 Help(与 F1 / `?` 一致)。
  2. 镜像详情页里 `History` 曾是一个分区,用户希望 History 提到镜像页顶层,**详情页不再显示 History**。
- 当前行为:
  `keys.KeyH` 在 `keyboard.go:215` 直接执行 `m.Viewport.ToggleHeader()`,没有走 `ActionHelp`;镜像详情中的 `buildImageDetailDataSections` 和文本回退路径已不再渲染 History 段,Docker / Podman 的详情 inspect 也暂时不再请求 History endpoint。详情滚动会回写合法偏移且只渲染当前可见行,避免底部越界偏移累积和全量样式重绘。
- 代码锚点:
  [internal/tui/keyboard/keyboard.go:215](internal/tui/keyboard/keyboard.go:215) `case keys.KeyH: m.Viewport.ToggleHeader()`
  [internal/tui/keyboard/actions.go:21](internal/tui/keyboard/actions.go:21) `case keys.ActionHelp: ToHelp(m)`
  [internal/tui/ui/pages/detail/image.go:14](internal/tui/ui/pages/detail/image.go:14) 镜像分区构造
- 期望行为:
  1. H 键在所有模式下都映射到 `ActionHelp`,与 F1 / `?` 等价;`ToggleHeader` 行为改由其他键(或显式 `Ctrl+H`)承担(可选,若决定取消 Toggle 入口则记录)。
  2. 新增 `ModeHistory / PanelHistory`,以及 `ActionImageHistory`。在镜像页按下 H(或绑定到具体单字母键)进入 History 页,展示该镜像的 layer history(数据来源复用 `InspectImageDetail()` 的 history 输出)。
  3. 镜像详情页已移除 History 段,仅保留其它字段。
  4. Help 页 / Footer / 默认键位示例同步说明 H = Help、H 在镜像页 = History(单一键位根据上下文路由)。
- 验收标准:
  1. 全局按 H 进入 Help,与 F1 行为一致。
  2. 镜像页内 H 进入镜像 History(可与 Help 区分:Help 是全局提示,History 是当前镜像数据)。
  3. 镜像详情页不再渲染 History 分区;键位说明与帮助文案同步更新。

<a id="br-009"></a>

### BR-009 卷详情页面无法加载数据;按 enter 进入容器子视图后无法上下选择

- 状态: `implementing`
- 优先级: `high`
- 症状:
  1. 卷页面按 D 进入详情页后,长时间卡在加载占位或空白,看不到卷的 inspect 数据。
  2. 按 enter 进入卷的容器子视图(由 `Volumes.DetailName` 触发)后,显示容器列表但**无法用上下键选中行**。
- 当前行为:
  1. `doVolumeInspect` 已改为异步 inspect + `SetVolumeDetail` + `ToDetail`,podman no-cgo 下不再在 `Update` 阶段阻塞；失败时会在详情页给出显式错误占位。
  2. `volumes.RenderList → renderContainers` 渲染容器子表时,**只用了 `cm.Items` 与 `rowLimit`,忽略了 `cm.Cursor` 与 `cm.ViewOffset`**,`RenderTable` 的 `Selected` / `Offset` 字段未传入;此外 `Volumes.DetailName` 进入子视图时 `m.Resources.Volumes.Cursor` 重置为 0,但渲染走的却是 `cm.Cursor`,从未联动,选中态无法显示也无法被上下键驱动。
- 代码锚点:
  [internal/tui/keyboard/volume_action.go:24](internal/tui/keyboard/volume_action.go:24) `doVolumeInspect` (同步 inspect, 无 cmd)
  [internal/tui/ui/pages/volumes/view.go:139](internal/tui/ui/pages/volumes/view.go:139) `renderContainers`
  [internal/tui/keyboard/navigate.go:29](internal/tui/keyboard/navigate.go:29) `ToVolumeDetail`
- 期望行为:
  1. 卷 inspect 走异步 `tea.Cmd` 路径,与 BR-006 容器详情保持一致;失败时给出明确错误占位。
  2. 卷 → 容器子视图要正确联动 `cm.Cursor` 与 `cm.ViewOffset`,渲染时传入 `Selected / Offset / Limit`,上下键能正常切换选中行。
  3. 进入子视图时仍把 `Volumes.Cursor` 归零以避免脏状态,但渲染与交互要基于 `cm`,行为参考容器列表自身的 cursor 模型。
  4. 在子视图里选中具体容器后,按 enter / l 等键应可进一步进入该容器的日志或详情(可选,但要避免目前"能进不能选"的死路)。
- 验收标准:
  1. podman no-cgo 启动后,卷 D 键加载可见,失败时显式错误。
  2. 卷 → 容器子视图中,上下键能逐行选中,Selected 行高亮。
  3. 在容器子视图里选中行后能正常退出回卷列表(BEsc 退到上一层,数据正确清理)。

<a id="br-023"></a>

### BR-023 F2 唯一提供 runtime 连接选择器;其它页面不得用 C 刷新 Conn

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 用户期望:F2 是**唯一**用来切换 runtime 连接的入口,采用**小窗口 + 表格布局**,列出每个连接的 Name / Runtime / Version / Status / Latency / TLS,并在窗口底部或标题区显示快捷键提示。
  - 用户明确要求:任何其它页面**不得**使用 `C` 键刷新连接,`C` 只能用于卷/网络的 Create 动作。
- 当前行为:
  - `openRuntimeSelector` 已经在所有页面统一由 F2 触发,渲染为 overlay 表格化条目,展示 runtime/version/socket/status/latency/security。
  - `KeyF12` / `KeyR` 绑定到 `ActionRefreshConnections`(全局);`KeyC` 仅绑定到 `ActionVolumeCreate` / `ActionNetworkCreate`,并未用于刷新连接。
  - 但选择器当前没有显式标注"↑↓ Navigate / Enter Switch / Esc Close"之类的快捷键提示行,Help / Footer 在 `ModeRuntimeSelect` 下也未必同步说明。
- 代码锚点:
  [internal/tui/keyboard/runtime_selector.go:13](internal/tui/keyboard/runtime_selector.go:13) `openRuntimeSelector`
  [internal/tui/ui/app/layout.go:293](internal/tui/ui/app/layout.go:293) `renderRuntimeSelector`
  [internal/tui/keys/registry.go:79](internal/tui/keys/registry.go:79) `ActionVolumeCreate` ← `KeyC`
  [internal/tui/keys/registry.go:82](internal/tui/keys/registry.go:82) `ActionNetworkCreate` ← `KeyC`
  [internal/tui/keys/registry.go:85](internal/tui/keys/registry.go:85) `ActionRefreshConnections` ← `KeyF12, KeyR`
- 期望行为:
  1. F2 是**唯一**的"切换 runtime 连接"入口;在任意页面按下 F2 都弹出一致的"小窗口 + 表格布局"选择器。
  2. 选择器显示每个连接的运行时类型、版本号、socket、状态、延迟、TLS 状态,统一为表格行;窗口底部或标题区显示快捷键提示(如 `↑↓ Navigate · Enter Switch · Esc Close`)。
  3. **任何页面不得**把 `C` 绑定到 `ActionRefreshConnections`。连接刷新只能由 `R` / `F12` 触发。
  4. Help / Footer 在所有页面下都不得出现 `C = refresh conn` 之类的提示。
- 验收标准:
  1. 任何页面按 F2,弹窗表格行包含 `Name / Runtime / Version / Status / Latency / TLS`;窗口内或底部显示快捷键提示行。
  2. 全局搜索 `KeyC.*[Rr]efresh\|[Cc]onn` 没有 `ActionRefreshConnections` 的绑定。
  3. Help / Footer 在 `ModeRuntimeSelect` / `ModeNormal` 下均不显示 `C = refresh` 之类提示。
  4. `R` / `F12` 在所有页面下仍能触发连接池刷新。

<a id="br-024"></a>

### BR-024 详情源码视图 Ctrl+C 复制不完整 / 一次性失效

- 状态: `open`
- 优先级: `medium`
- 症状:
  - YAML / JSON 源码视图下,`Ctrl+A` 可以"全选"高亮,但 `Ctrl+C` 后续复制不可重复执行或复制内容不完整,无法满足"多段、多次、批量"的复制需求。
- 当前行为:
  - `keyboard/detail.go:41-52` 中 `Ctrl+A` 在 `DetailSourceType != Section` 且 `HasRawSource()` 时把 `SourceSelected = true`;`Ctrl+C` 仅在 `SourceSelected == true` 时复制 `SourceText()`(完整格式化 YAML / JSON)到剪贴板。
  - `SourceSelected` 设置后**从不重置**,但应用没有给用户可见的"已选中"高亮区分,且 `Ctrl+C` 缺乏撤销 / 重选控制;若用户在终端里尝试鼠标拖选或多次 `Ctrl+C`,界面没有任何提示当前到底是"全选态"还是"待重新选择"。
- 代码锚点:
  [internal/tui/keyboard/detail.go:41](internal/tui/keyboard/detail.go:41) `Ctrl+A` 分支
  [internal/tui/keyboard/detail.go:45](internal/tui/keyboard/detail.go:45) `Ctrl+C` 分支
  [internal/tui/state/detail.go:57](internal/tui/state/detail.go:57) `SourceSelected` 字段
- 期望行为:
  1. `Ctrl+C` 应当能从源码视图复制**完整**格式化文本(`SourceText()` 全量输出),支持多次连续执行。
  2. 应用应在源码视图给出"已选中 / 已复制"的可视提示,例如选中行有可视高亮、复制后 toast 显示已复制字节数。
  3. 复制行为不应阻止终端自身的鼠标拖选 / 系统剪贴板路径;当应用处于源码视图时,用户可以用 `Ctrl+C` 触发应用复制,也可以用终端自身选区复制。
- 验收标准:
  1. 在 YAML / JSON 视图,`Ctrl+A` 后多次 `Ctrl+C` 都能成功复制,toast 反馈每次复制字节数。
  2. 复制内容长度等于 `SourceText()` 完整输出,不出现截断 / 单行。
  3. 切换回 section 视图后,`SourceSelected` 重置为 false,避免误复制。

<a id="br-025"></a>

### BR-025 容器详情页标题占位符 `{0}` `{1}` 未替换

- 状态: `open`
- 优先级: `high`
- 症状:
  - 容器详情页标题栏显示为字面值 `Container Detail: {0} ({1})`,而不是 `Container Detail: <name> (<id>)`,占位符未被替换。
- 当前行为:
  - `i18n.T` 使用 `golang.org/x/text/message` 的 `Printer.Sprintf(key, args...)` 进行 ICU 占位符替换,理论上支持 `{0}` / `{1}`。
  - `keyboard/container_action.go:409` 调用 `i18n.T("detail.title.container", ctr.Name, ctr.ID)`,**传入 2 个参数**;en.jsonc / zh.jsonc / ja.jsonc 三份语言文件中 `"detail.title.container"` 都存在,且格式串包含 `{0}` `{1}`。
  - 但运行时仍出现字面值,可能原因:`message.Printer.Sprintf` 实际行为 / args 与模板不匹配 / 渲染路径对 `DetailTitle` 再次进行 i18n 查找并以 `key` 形式返回 / 或者 `Sprintf` 走的是 `fmt` 而非 ICU 解析路径。
- 代码锚点:
  [internal/tui/keyboard/container_action.go:409](internal/tui/keyboard/container_action.go:409) `title := i18n.T("detail.title.container", ctr.Name, ctr.ID)`
  [internal/data/i18n/lang.go:80](internal/data/i18n/lang.go:80) `T(key, args...)`
  [internal/data/i18n/lang/en.jsonc:275](internal/data/i18n/lang/en.jsonc:275) `"detail.title.container": "Container Detail: {0} ({1})"`
  [internal/tui/ui/app/page_templates.go:73](internal/tui/ui/app/page_templates.go:73) `view.title = m.Detail.DetailTitle`
- 期望行为:
  容器详情页标题正确替换为 `Container Detail: <name> (<id>)`;同样适用于 `detail.title.container_yaml` / `detail.title.container_json`(切到 YAML / JSON 视图时也要正确替换);`detail.title.image*` 等其它含占位符的标题一并按相同规则处理。
- 验收标准:
  1. 进入任一容器详情页,标题栏显示完整替换后的文本,不出现字面 `{0}` `{1}`。
  2. 切到 YAML / JSON 视图时标题正确替换。
  3. 三种语言(en/zh/ja)都要正确替换。

<a id="br-026"></a>

### BR-026 详情页快捷键集合需要明确:支持 j/k/PgUP/PgDn/Filter,不支持 Space/R

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 详情页当前把 `Space` 当作 PageDown(向下滚 20 行),用户期望**不提供 Space 滚动**;`R` 在详情页当前不绑定,但用户要求**明确不提供 R 刷新**,且 Help / Footer 必须按页面对应的快捷键集合展示。
- 当前行为:
  - `keyboard/detail.go:33` `case keys.KeySpace, keys.KeyPgDn: m.Detail.Scroll(20)` —— `Space` 与 `PgDn` 共用 20 行滚动分支。
  - `R` 在 `ModeDetail` 下不绑定任何动作(`handleDetailKeys` 未匹配 `KeyR` 时落入 `return true, nil`,实际吞掉按键但不报错),但 Help / Footer 不会主动标出 "R 在此处不可用"。
- 代码锚点:
  [internal/tui/keyboard/detail.go:33](internal/tui/keyboard/detail.go:33) `case keys.KeySpace, keys.KeyPgDn`
  [internal/tui/ui/action/registry.go:235](internal/tui/ui/action/registry.go:235) Detail footer/shortcuts 投影
- 期望行为:
  详情页支持的快捷键集合明确如下:
  - `j` / `k`:上 / 下滚动一行
  - `PgUP` / `PgDn`:向上 / 向下翻页
  - `/` 或约定键:Filter 入口(若当前 detail 模式不支持 Filter 则不投影)
  - `Esc` / `Enter`(`ActionBack`):返回
  - `s`:在 section / YAML / JSON 视图间循环(已有)
  - **不提供** `Space` 滚动
  - **不提供** `R` 刷新
  - **不提供** `C` 刷新连接(继承 BR-023 全局规则)

  Help / Footer 在 `ModeDetail` 下只投影当前可用快捷键,绝不出现 `Space = PageDown` 或 `R = Refresh` 提示。
- 验收标准:
  1. 详情页按 `Space` 不再触发滚动(日志/卷/网络/Compose 等其它页面不受影响)。
  2. 详情页按 `R` 不触发任何动作,既无报错也无 toast。
  3. 详情页 Help / Footer 只显示 `j/k`、`PgUP/PgDn`、`Esc/Enter`、`s Source`、`/ Filter`(按页面能力)。
  4. 若某 detail 资源不支持 Filter,Help / Footer 不出现 `/ Filter` 提示。

<a id="br-027"></a>

### BR-027 容器日志页滚到底后向上划卡顿

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 容器日志页用鼠标滚轮滑到最下面,再向上划会卡一下。多次正反向滚动后卡顿更明显。
- 当前行为:
  - 与 BR-016(详情页滚动卡顿)同类根因:`m.Log.Scroll(delta)` 只增减 offset,不立即把 offset 钳制到合法可见范围;反向滚动时 offset 先回到 0 再开始正向,体感是先"还债"再移动。
  - 长日志下 `ui/pages/logs/view.go` 在每帧重建完整可见行,放大卡顿。
- 代码锚点:
  [internal/tui/state/log.go](internal/tui/state/log.go) `Scroll` / `VisibleOffset`
  [internal/tui/ui/pages/logs/view.go:79](internal/tui/ui/pages/logs/view.go:79) 渲染循环起点使用 `m.Log.VisibleOffset(...)`
  [internal/tui/keyboard/log.go:22](internal/tui/keyboard/log.go:22) `m.Log.Scroll(±1)`
  [internal/tui/keyboard/log.go:31](internal/tui/keyboard/log.go:31) `Scroll(±20)`
- 期望行为:
  1. 日志页滚动状态回写到当前可见边界,不能积累不可见 overscroll。
  2. 向上 / 向下滚动在底部与顶部都应即时生效,不需要先回滚"欠账"。
  3. 长日志(`>=1000` 行)下连续正反向滚动 10+ 次,体感无卡顿;每次渲染只截取可见行,不重建完整内容。
- 验收标准:
  1. 日志页滚到底后立即向上滚 1 行,视图即时变化。
  2. 连续滚动 10 次后,反向滚动立即响应。
  3. 1000+ 行日志下反向滚动无卡顿。

<a id="br-034"></a>

### BR-034 镜像 History 顶层页 (H 键) — 历史因性能取消,需重新设计

- 状态: `open`
- 优先级: `high`
- 症状:
  - 在镜像页按 H 当前是 `Viewport.ToggleHeader()`(`keyboard.go:215`),不是进入镜像 History。
  - 镜像详情页里曾渲染完整 History 段,大镜像(>50 layer)下每次滚动都重建整段 section,因**性能问题**取消。
  - 用户期望:H 键在**镜像页**进入**镜像 History 顶层页**(浏览 layer history),详情页不再渲染 History 段。
- 当前行为:
  - `keys.KeyH` 在 `keyboard.go:215` 直接执行 `m.Viewport.ToggleHeader()`,没有走 `ActionHelp`;镜像详情中的 `buildImageDetailDataSections` 与文本回退路径已不再渲染 History 段,Docker / Podman 的详情 inspect 也暂时不再请求 History endpoint。详情滚动已用 BR-016 / BR-007 优化,但仍未重新引入 History 顶层页。
- 代码锚点:
  [internal/tui/keyboard/keyboard.go:215](internal/tui/keyboard/keyboard.go:215) `KeyH` 当前 ToggleHeader
  [internal/tui/ui/pages/detail/image.go:14](internal/tui/ui/pages/detail/image.go:14) 镜像分区构造(已不渲染 History)
  [internal/data/runtime/image.go:57](internal/data/runtime/image.go:57) `ImageHistoryLayer` 数据结构已定义
  [internal/data/runtime/docker/images.go:107](internal/data/runtime/docker/images.go:107) `InspectImageDetailContext` history 字段
- 期望行为:
  1. 在**镜像页**(`ModeImages`),按 H 键进入 `ModeHistory`(独立顶层页,不是详情页内分区)。
  2. History 顶层页只展示镜像 layer 列表:Created / Size / CreatedBy / Comment;支持 j/k、PgUP/PgDn、`/` Filter。
  3. 数据源复用 `InspectImageDetail()` 的 history 输出;复用 BR-007 的按 revision/source 缓存机制 + BR-016 的 ClampVisibleOffset,避免上次性能问题。
- 验收标准:
  1. 镜像页按 H 键,弹出 History 顶层页,标题为 `<image>:<tag> history`;行数等于镜像 layer 数。
  2. 选中镜像 > 50 layer 时,上下滚动 / PgDn / j 翻页均无卡顿。
  3. 详情页不再渲染 History 分区。
- 设计参考: [docs/feature-design.md §5.1](feature-design.md)

<a id="br-035"></a>

### BR-035 Events 独立面板 (F3 键) + Network Connect/Disconnect

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 当前 `Engine.Events().Subscribe(...)` 已实现后台订阅,合并到资源更新路径,但**没有独立的 Events 浏览面板**。
  - `NetworkService` 接口只有 List/Create/Remove/Prune,**没有** Connect / Disconnect;无法动态调整容器的网络接入。
- 当前行为:
  - F3 当前未绑定任何动作;容器启动后只能使用启动时配置的网络。
- 代码锚点:
  [internal/tui/update/update_events.go:28](internal/tui/update/update_events.go:28) `subscribeEventsCmd`
  [internal/data/runtime/streaming.go:64](internal/data/runtime/streaming.go:64) `EventService` 接口
  [internal/data/runtime/docker/service_network.go](internal/data/runtime/docker/service_network.go) Docker `NetworkService` 缺 Connect/Disconnect
  [internal/data/runtime/podman/service_network.go](internal/data/runtime/podman/service_network.go) Podman 同上
  [internal/tui/keys/registry.go](internal/tui/keys/registry.go) `KeyF3` 未注册
- 期望行为:
  1. **Events 独立面板**:全局 `F3` 进入 `ModeEvents`;UI 类似日志页实时流,展示 Time / Type / Action / Resource 四列;支持按 type/action/时间 Filter;`space` 暂停/恢复;`Esc` 返回;`Ctrl+D` 清空(需 confirm)。
  2. **Network Connect / Disconnect**:网络页加 `Connect Container` / `Disconnect Container` 动作;容器详情页显示已接入网络列表 + 断开按钮;runtime 新增 `NetworkService.Connect` / `Disconnect`。
- 验收标准:
  1. 全局 F3 进入 Events 面板;Docker / Podman events 流持续滚动;按 space 暂停、再按恢复。
  2. 容器启动 / 停止 / 镜像 pull 等事件在 Events 面板实时出现。
  3. 在网络页 / 容器详情页能把运行中容器接入新网络;`docker exec ctr ping new-net-peer` 验证网络可达。
  4. disconnect 后,容器在该网络的接口消失;不影响其它网络。
- 设计参考: [docs/feature-design.md §5.7](feature-design.md) / [§5.3](feature-design.md)

<a id="br-036"></a>

### BR-036 镜像 tarball 导入功能 (Image Import)

- 状态: `open`
- 优先级: `medium`
- 症状: 镜像页缺少 tarball 导入动作;用户拿到 `docker export` 的扁平 tar 或第三方 tar 后,无法在 dtui 内导入为新镜像。
- 当前行为: `ImageService` 只有 List / Inspect,无 Import;`Ctrl+L` (ActionImageLoad) 是 `docker load`,与 `docker import` 语义不同。
- 代码锚点:
  [internal/data/runtime/image.go](internal/data/runtime/image.go) `ImageService` 接口(无 Import)
  [internal/data/runtime/docker/service_image.go](internal/data/runtime/docker/service_image.go) Docker ImageService
  [internal/data/runtime/podman/service_image.go](internal/data/runtime/podman/service_image.go) Podman ImageService
- 期望行为:
  1. 镜像页 Action Bar 中加 "Import tarball" 动作(走 BR-039)。
  2. 触发表单:源 tarball 路径、目标 repository:tag、提交消息 / 作者(可选)。
  3. 提交后调 `Engine.Images().Import(...)`;Docker 用 `cli.ImageImport`,Podman 用等价 REST。
  4. 进度反馈复用 `ImageTransfer` 框架的进度 / toast / 取消机制。
- 验收标准:
  1. 镜像页 Action Bar 中 "Import tarball" 可点击;触发表单。
  2. 提交后,新镜像出现在镜像列表,inspect 可见但只有 1 层(history 为空)。
  3. `docker save` 多 layer tarball 走 `Load` (Ctrl+L),`docker export` 扁平 tarball 走 Import,两条路径独立。
- 设计参考: [docs/feature-design.md §5.10](feature-design.md)

<a id="br-037"></a>

### BR-037 docker / podman Registry Login(私库认证)

- 状态: `open`
- 优先级: `medium`
- 症状: 当前无 `docker login` 等价入口;用户访问私有 registry 时必须在外部 CLI 登录。
- 当前行为: `ImageService` 没有 `Login` 方法;`keys.Action*` 无 `ActionRegistryLogin` 注册。
- 代码锚点:
  [internal/data/runtime/image.go](internal/data/runtime/image.go) `ImageService` 接口(无 Login)
  [internal/tui/keys/action.go](internal/tui/keys/action.go) 无 `ActionRegistryLogin` 定义
- 期望行为:
  1. Action Bar 中加 "Registry Login" 动作(走 BR-039)。
  2. 触发表单:服务器 URL(默认 `docker.io`)、用户名、密码(掩码输入)、高级 TLS 选项。
  3. 提交后调 `Engine.Login(ctx, server, user, password, opts)`;Docker 用 `cli.RegistryLogin`,Podman 用等价 REST。
  4. **不存储密码到 dtui 配置**;依赖引擎自身的 `~/.docker/config.json`。
- 验收标准:
  1. Action Bar 中 "Registry Login" 可点击,表单可填写。
  2. 登录成功后,dtui 内 `ImagePull` (`Ctrl+P`) 私有仓库镜像成功,无需外部 `docker login`。
  3. 错误凭证返回明确 toast(`Unauthorized` 等),不泄露密码。
  4. dtui 进程退出后,引擎凭证仍存在(由引擎自身管理)。
- 设计参考: [docs/feature-design.md §5.9](feature-design.md)

<a id="br-038"></a>

### BR-038 容器 Exec 页面 UI 不符合 shell 终端

- 状态: `open`
- 优先级: `medium`
- 症状: 容器页按 `e` 进入 Exec 后,Exec 页面 UI 不像 shell 终端:缺少提示符 / 路径前缀 / 颜色;历史命令 / Tab 补全 / 方向键(↑↓)历史回看 / Ctrl+L 清屏等 shell 特性未提供。
- 当前行为: `internal/tui/ui/widget/dialog/exec.go` 提供 dialog 风格的 exec,输出是单行输入框 + 输出区,不是全屏 terminal。
- 代码锚点:
  [internal/tui/ui/widget/dialog/exec.go](internal/tui/ui/widget/dialog/exec.go) `RenderExecDialog` / `RenderTextInput`
  [internal/tui/keyboard/container_action.go](internal/tui/keyboard/container_action.go) `doAutoExecAction`
  [internal/data/runtime/docker/service_exec.go](internal/data/runtime/docker/service_exec.go) Docker exec PTY 透传
  [internal/data/runtime/podman/service_exec.go](internal/data/runtime/podman/service_exec.go) Podman exec PTY 透传
- 期望行为:
  1. Exec 进入新模式 `ModeExec`,**全屏渲染**,不再是 floating dialog。
  2. 真正的 PTY 透传:`Engine.Exec().Start(ctx, containerID, []string{"/bin/sh"}, TTY=true)`。
  3. 输入处理:↑↓ 翻历史、Tab 补全、Ctrl+L 清屏、Ctrl+C 转发 SIGINT、Ctrl+D 退出前台 shell 但不杀容器。
  4. 输出:ANSI 颜色 / 宽字符 / 进度条正确渲染。
  5. `Esc` 退出 dtui 的 exec mode,**不发信号给前台进程**。
- 验收标准:
  1. 容器页按 `e` 进入 exec,页面**全屏 shell**,提示符正确。
  2. 在 shell 内执行 `ls /` / `cat /etc/hostname` / `top` 等命令,输出正确(包括 ANSI 颜色)。
  3. ↑↓ 翻历史、Tab 补全、Ctrl+L 清屏、Ctrl+C 发 SIGINT 均工作。
  4. Ctrl+D 退出前台 shell 但容器继续运行。
  5. Esc 退出 dtui 的 exec mode,**不**杀掉容器前台进程。
- 设计参考: [docs/feature-design.md §5.4](feature-design.md)

<a id="br-040"></a>

### BR-040 Dialog 风格统一:四周透明 + panel 居中

- 状态: `open`
- 优先级: `medium`
- 症状: dialog 风格不统一:有的占满整屏 / 有的浮在中间 / 有的位置在 panel 外;边框 / 透明 / 位置算法各不相同。
- 当前行为: `internal/tui/ui/widget/dialog/` 下各 dialog(confirm / notification / selection / exec / view)各自渲染。
- 代码锚点:
  [internal/tui/ui/widget/dialog/overlay.go](internal/tui/ui/widget/dialog/overlay.go) `PlaceOverlay`
  [internal/tui/ui/widget/dialog/view.go](internal/tui/ui/widget/dialog/view.go) 各 dialog 渲染
  [internal/tui/ui/widget/dialog/confirm.go](internal/tui/ui/widget/dialog/confirm.go) confirm dialog
  [internal/tui/ui/widget/dialog/notification.go](internal/tui/ui/widget/dialog/notification.go) notification
  [internal/tui/ui/widget/dialog/selection.go](internal/tui/ui/widget/dialog/selection.go) selection
  [internal/tui/ui/widget/dialog/exec.go](internal/tui/ui/widget/dialog/exec.go) exec dialog
- 期望行为:
  1. **统一接口**:新增 `widget/dialog/centered.go`,提供 `CenterOnPanel(panelRect, content, w, h) string`。
  2. **对话框渲染规范**:
     - 边框:lipgloss `Border` 样式,**不**带 `Background` 填充(四周透明)。
     - 内边距:统一 `Padding(1, 2)`。
     - 标题:统一在顶部。
     - 内容区:背景透明或半透明。
  3. **应用范围**:confirm / notification / selection / exec / filter / 自定义表单(Copy / Update / Commit / Wait / Import / Login / Events Filter)。
- 验收标准:
  1. 所有 dialog 在 panel 内居中显示,四周透明,不覆盖 header / footer / message rail。
  2. 同类 dialog(例如两个 confirm)风格一致:边框 / 内边距 / 标题位置相同。
  3. panel 缩窄 / 拉宽时,dialog 跟着居中且不溢出 panel。
  4. dialog 内的按钮 / 输入框 / 文本行版式统一。
- 设计参考: [docs/feature-design.md §5.8](feature-design.md)

<a id="br-028"></a>

### BR-028 (TBD - 待用户补充)

- 状态: `pending`
- 优先级: -
- 症状: 用户第 6 条 BUG 内容未提供。
- 期望行为: 用户明确补充第 6 条 BUG 描述后,按 `新增条目模板` 拆解为症状、当前行为、代码锚点、期望行为、验收标准。

<a id="br-029"></a>

### BR-029 筛选框 Enter/Esc 双层退出:Enter 应用并保留框,首次 Esc 回退到框,二次 Esc 退出筛选模式

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 在资源页按 `/` 进入筛选模式后,按 Enter 当前会触发"应用 + 退出"语义,但用户期望 **Enter 仅应用过滤,筛选框保留可见**。
  - 在资源页筛选模式中按 Esc 当前直接退出筛选模式 (`actions.go:138-140` 直接调用 `BackFromFilter`),用户期望 **首次 Esc 仅把焦点退回到筛选输入框**,**第二次 Esc 才退出筛选模式**。
- 当前行为:
  - `keyboard/actions.go:25-26`: `ActionFilter` 调用 `ToFilter(m)` 进入筛选模式。
  - `keyboard/actions.go:138-140`: 在 `ModeFilter` 下按 Esc,直接调用 `BackFromFilter(m)` 退出整个筛选模式,单 Esc 即退出。
  - 资源列表过滤已实现"输入即时过滤"(BR-004 修复记录),与 Enter 应用语义等价,因此 Enter 不必"额外触发退出"。
- 代码锚点:
  [internal/tui/keyboard/actions.go:25](internal/tui/keyboard/actions.go:25) `ActionFilter` 入口
  [internal/tui/keyboard/actions.go:138](internal/tui/keyboard/actions.go:138) `ModeFilter` 下 Esc 单次退出
  [internal/tui/state/filter.go](internal/tui/state/filter.go) (或对应 model) `ToFilter` / `BackFromFilter` / 焦点状态
- 期望行为:
  1. **筛选模式下的按键语义明确分层**:
     - `Enter`:把当前输入应用到过滤(过滤已实时生效时,Enter 仅"确认提交");筛选输入框**继续可见**,用户可继续编辑。
     - 首次 `Esc`:把焦点从其它位置退回到筛选输入框(若焦点本就在输入框内则 no-op),**不退出 ModeFilter**,输入框继续可见可编辑。
     - 第二次 `Esc`:退出整个筛选模式,清空输入并恢复 ModeNormal。
  2. 资源过滤仍然"输入即时过滤"(沿用 BR-004 行为)。
  3. 与日志页的 `Search`(`ModeSearch`)保持一致的双 Esc 退出语义(若适用)。
- 验收标准:
  1. 在容器 / 镜像 / 卷 / 网络任一页面按 `/`,筛选框出现并获得焦点。
  2. 输入过滤词后立即过滤;按 Enter,过滤仍然生效,筛选框继续可见,焦点保留在输入框。
  3. 按 Esc:若焦点不在输入框,焦点回到输入框;再按 Esc,ModeFilter 退出,输入框消失,ModeNormal 恢复。
  4. 在筛选输入框中按 Esc 同样计数为"已回到输入框"的一次,下一次 Esc 退出整个筛选模式。

<a id="br-030"></a>

### BR-030 取消鼠标点击表格行选中,改用滚轮上下选行

- 状态: `open`
- 优先级: `high`
- 症状:
  - 镜像 / 卷 / 网络 / 容器 / 审计列表上,鼠标点击行尝试选中时位置错位(BR-012)。
  - 用户决定**取消鼠标点击行选中**,改为只用鼠标**滚轮**驱动上下选行;终端鼠标键盘边界因此更清晰。
- 当前行为:
  - `internal/tui/ui/app/mouse.go:174` `HitTest` 把鼠标 `(x,y)` 映射到 panel 行号;`mouse.go:254` `clickListCursor` 写入选中行;镜像 panel 在 standard 模式下与整条 rail 几何不一致,导致点击命中错位。
  - 滚轮已经在日志/详情页支持,但**资源列表页**未把滚轮事件路由到 cursor 上下移动。
- 代码锚点:
  [internal/tui/ui/app/mouse.go:174](internal/tui/ui/app/mouse.go:174) `HitTest`
  [internal/tui/ui/app/mouse.go:223](internal/tui/ui/app/mouse.go:223) `clickListCursor` 调用点
  [internal/tui/ui/app/mouse.go:254](internal/tui/ui/app/mouse.go:254) `clickListCursor`
  [internal/tui/update/update_events.go](internal/tui/update/update_events.go) (或等价) 鼠标滚轮事件分发
- 期望行为:
  1. **取消**鼠标点击行选中:BR-012 已 wontfix,`clickListCursor` 路径移除或退化为 no-op;HitTest 对 panel body 的点击不再触发 cursor 移动。
  2. **新增**滚轮上下选行:在资源列表(容器 / 镜像 / 卷 / 网络 / 审计)的 panel body 上,鼠标滚轮向下滚动一格 → cursor 向下 1 行,向上滚动一格 → cursor 向上 1 行;与键盘 `j` / `k` 行为完全一致,同步刷新 `ViewOffset`。
  3. 终端原生复制 / 选区行为不被应用拦截;只在 panel body 区域生效,panel 之外不抢滚轮事件。
- 验收标准:
  1. 资源列表的鼠标点击不再移动 cursor;HitTest 对 panel 点击不调用 `clickListCursor`。
  2. 在容器列表 panel 内滚动鼠标滚轮,选中行实时上下移动,与键盘 `j` / `k` 行为一致;`ViewOffset` 同步推进。
  3. 在头部 / 消息栏 / footer 区域滚动滚轮,不触发列表 cursor 变化。
  4. 终端原生选区复制(鼠标拖选)在 panel 内仍可用。

<a id="br-031"></a>

### BR-031 容器页容器资源占用(CPU / 内存)未实时刷新

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 在容器页按 `m` 开启 stats 后,选中容器的 CPU / 内存数据没有按预期实时刷新;视图上的数值停留在初始值或上一次刷新时的快照。
- 当前行为:
  - 2026-08-02 BR-042 已修复 ticker 生命周期:离开 Containers 页面时仍保留下一轮调度,返回后恢复请求,不再因一次页面切换永久停止。该条目继续保持 `open`,等待真实 Podman 下 30 秒刷新验证与 stats 错误提示补齐。
  - `internal/tui/update/update_tick.go:32` `handleStatsTick` 已实现:容器页激活时 `StatsActive=true`,按 `Docker.StatsPollSec`(默认 3 秒) 周期为**当前过滤后**的容器列表派发 `FetchStats` 命令;`update_tick.go:17` `handleStatsReceived` 把数据写入 `m.Resources.Containers.Stats[msg.ContainerID]`。
  - `internal/tui/ui/pages/containers/view.go:90` 渲染时通过 `cm.Stats[c.ID]` 读取最新值。
  - 但实测中数据没有"看起来"实时刷新:可能原因包括 (a) `tea.Tick` 被其它 ticker 阻塞;(b) 长轮询或 socket I/O 在 podman no-cgo 模式下阻塞过久;(c) 数据写入路径触发了不必要的全屏重绘;(d) `FetchStats` 出错被静默忽略。
- 代码锚点:
  [internal/tui/update/update_tick.go:17](internal/tui/update/update_tick.go:17) `handleStatsReceived`
  [internal/tui/update/update_tick.go:32](internal/tui/update/update_tick.go:32) `handleStatsTick`
  [internal/tui/update/update.go:129](internal/tui/update/update.go:129) `StatsTick` 分发
  [internal/tui/ui/pages/containers/view.go:90](internal/tui/ui/pages/containers/view.go:90) stats 渲染读取
  [internal/tui/state/containers.go](internal/tui/state/containers.go) `Stats` map 与 `StatsActive` 字段
- 期望行为:
  1. 容器页开启 stats 后,选中的(以及当前过滤后可见的)容器 CPU / 内存数据按 `StatsPollSec` 周期稳定刷新;视觉上能看到数值跳动。
  2. stats tick 与其它 ticker(toast / runtime selector / detail revision)独立,任一 ticker 卡顿不影响 stats。
  3. stats 拉取失败(socket 超时 / podman REST 错误)有明确 toast 或状态栏提示,不再静默忽略。
  4. stats 数据更新路径避免触发全屏重绘,只刷新 stats 列;`BenchmarkRenderCachedContainerDetail` 等现有滚动 benchmark 不应回退。
- 验收标准:
  1. 容器页按 `m` 开启 stats,等待 3 秒以上,选中的容器 CPU% / 内存% 至少有一次刷新;数值变化可观察。
  2. 在容器页停留 30 秒以上,stats 数据按 `StatsPollSec` 持续刷新,不被其它 ticker 阻塞。
  3. 拔掉容器运行时后,UI 显示明确的 stats 拉取失败提示,不再显示陈旧值。
  4. stats 更新不影响滚动 / 选中行高亮 / 详情页 footer。

<a id="br-032"></a>

### BR-032 表格排序:N / Ctrl+N 正反排序,鼠标点击表头排序

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 当前表格排序由 `O`(`toggleSortColumn` 切换列)与 `Ctrl+O`(`toggleSortDirection` 切换升降序)承担。
  - 用户期望:`N` = 正排序(升序),`Ctrl+N` = 反排序(降序);此外,**鼠标点击表头**也能排序。
- 当前行为:
  - `internal/tui/keyboard/keyboard.go:221-234`:`KeyO` 调 `toggleSortColumn`,`KeyCtrlO` 调 `toggleSortDirection`。
  - **没有** `KeyN` / `KeyCtrlN` 在正常模式下的排序绑定;`KeyN` 只在 dialog 中被当作"关闭"(`keyboard.go:145`)。
  - **没有**鼠标点击表头的处理;`mouse.go` 中只有 `clickListCursor` / `HitTest` / 滚轮事件,**没有**表头点击路由。
- 代码锚点:
  [internal/tui/keyboard/keyboard.go:221](internal/tui/keyboard/keyboard.go:221) `KeyO` 入口
  [internal/tui/keyboard/keyboard.go:228](internal/tui/keyboard/keyboard.go:228) `KeyCtrlO` 入口
  [internal/tui/keyboard/keyboard.go:248](internal/tui/keyboard/keyboard.go:248) `toggleSortColumn`
  [internal/tui/keyboard/keyboard.go:268](internal/tui/keyboard/keyboard.go:268) `toggleSortDirection`
  [internal/tui/ui/app/mouse.go:154](internal/tui/ui/app/mouse.go:154) `HitTest` 缺表头分支
  [internal/tui/ui/component/table.go:54](internal/tui/ui/component/table.go:54) `SortColKey` / `SortAsc` 数据结构已存在
- 期望行为:
  1. `N` 键 = 正排序;`Ctrl+N` 键 = 反排序。推荐语义:`N` 把当前排序列设为升序,`Ctrl+N` 把当前排序列设为降序(不切换列)。
  2. 鼠标点击表头:
     - 在升序列上点击 → 切到降序。
     - 在降序列上点击 → 切到升序。
     - 在非当前排序列上点击 → 切到该列,默认升序。
     - 表头必须显示当前排序状态(箭头 / 高亮)。
  3. 现有的 `O` / `Ctrl+O` 是否保留由用户决定:可以并存(`O` 切列,`N` 改方向)或废弃(只保留 `N` / `Ctrl+N` + 鼠标)。
  4. 排序改动后 `Cursor` 不丢失选中行;`ViewOffset` 跟随 cursor 调整;Help / Footer 显示当前可用的排序快捷键。
- 验收标准:
  1. 按 `N` 后,排序列标记为升序,排序箭头在表头显示,行顺序按升序排列。
  2. 按 `Ctrl+N` 后,排序列标记为降序,排序箭头反向,行顺序按降序排列。
  3. 鼠标点击表头某列:若非当前列,切到该列升序;若当前列升序,切到降序;若当前列降序,切到升序。
  4. 排序时 `Cursor` 不丢失选中行;`ViewOffset` 跟随 cursor 调整。
  5. Help / Footer 准确显示当前可用的排序快捷键。

<a id="br-033"></a>

### BR-033 TASK-019 高级容器动作未实现:Copy / Update / Diff / Export / Commit / Wait

- 状态: `open`
- 优先级: `high`
- 症状:
  - 用户在容器页按 `Ctrl+O`(预期 Copy)、`Ctrl+W`(预期 Update)等快捷键,**没有任何反应**(无 toast、无错误、按键被静默吞掉)。
  - 用户期望实现:容器 cp(从容器复制文件到主机或反向)、Update(运行时改 Memory / CPU / 重启策略)、Diff(查看容器文件系统改动)、Export(导出容器文件系统为 tar)、Commit(把容器提交为新镜像)、Wait(阻塞等待容器状态)。
  - `docs/requirements.md` 第 12 行声称这些动作"已实现",但实际上**只是注册了 keymap,handler 缺失**。
- 当前行为:
  - `internal/tui/keys/registry.go:66-71` 把以下 6 个动作注册到默认键位:
    - `ActionContainerUpdate` ← `Ctrl+W`
    - `ActionContainerDiff` ← `F2`(注:与 `ActionSwitchRuntime` 的 `F2` 冲突)
    - `ActionContainerExport` ← `Ctrl+X`
    - `ActionContainerCommit` ← `Ctrl+K`(注:与 `ActionContainerKill` 的 `Ctrl+K` 冲突)
    - `ActionContainerWait` ← `Ctrl+Y`
    - `ActionContainerCopy` ← `Ctrl+O`(注:与 O 切换排序列的现有逻辑冲突)
  - `internal/tui/keyboard/actions.go` 的 `handleAction` switch **完全没有**这些 `case`,默认 fall-through 到 `return m, tea.Batch(cmds...)` —— 即按键被吞掉、无反馈。
  - `internal/data/runtime/podman/service_action.go:63-110` 已经实现了 Update / Diff / Export / Commit / Wait 的 Podman runtime 调用。`runtimeapi.ActionCopy` 等也已存在。
- 代码锚点:
  [internal/tui/keys/registry.go:66](internal/tui/keys/registry.go:66) TASK-019 默认键位注册
  [internal/tui/keyboard/actions.go:16](internal/tui/keyboard/actions.go:16) `handleAction` switch 缺 TASK-019 case
  [internal/tui/keys/action.go:23](internal/tui/keys/action.go:23) TASK-019 Action 定义
  [internal/tui/keys/keys.go:34](internal/tui/keys/keys.go:34) `KeymapConfig` TASK-019 字段
  [internal/data/runtime/podman/service_action.go:63](internal/data/runtime/podman/service_action.go:63) Podman Update/Diff/Export/Commit/Wait/Copy 分发
  [internal/data/runtime/docker/containers.go](internal/data/runtime/docker/containers.go) (待核对 Docker 端 Copy/Update 等是否实现)
- 期望行为:
  1. **取消 docs/requirements.md 的"已实现"声明**:这些动作实际未实现,文档需改回"待实现"。
  2. **接入 handler**:为每个 TASK-019 动作补一个 `case keys.ActionContainerCopy` 等分支,根据操作打开对应表单(dialog)或直接执行:
     - `Copy`:打开"源路径 / 目标路径"对话框,执行 `Engine.Containers().CopyToContainer` / `CopyFromContainer`。
     - `Update`:打开"内存 / CPU / 重启策略"表单,执行 `Engine.Containers().Update`。
     - `Diff`:打开 diff 视图(类似详情但只显示 filesystem diff)。
     - `Export`:打开"目标 tar 路径"对话框,执行 `Engine.Containers().Export`。
     - `Commit`:打开"repository:tag"对话框,执行 `Engine.Containers().Commit`。
     - `Wait`:打开"阻塞条件(运行 / 停止 / 删除)"对话框,执行 `Engine.Containers().Wait`。
  3. **解决键位冲突**:`F2` 与 `ActionSwitchRuntime` / `ActionContainerDiff` 冲突;`Ctrl+K` 与 `ActionContainerKill` / `ActionContainerCommit` 冲突;`Ctrl+O` 与排序 `toggleSortColumn` 冲突。建议 TASK-019 默认绑定由用户重设,不沿用冲突值。
  4. **Docker runtime 支持**:核对 Docker engine 是否提供 Copy/Update/Commit/Wait 的等价 API,缺失时在 helper 中给出 "Docker 引擎不支持" 的明确 toast,而不是静默无响应。
- 验收标准:
  1. 在容器页按 TASK-019 的默认键位,要么弹出对应表单(dialog),要么执行动作并反馈结果,**不再静默吞键**。
  2. 文档 `requirements.md` 中"已实现"的能力与代码保持一致;TASK-019 这一类"已实现"但实际 handler 缺失的条目必须修订。
  3. TASK-019 动作的键位与现有 `O` / `Ctrl+O` / `F2` / `Ctrl+K` 不冲突;Help / Footer 显示当前生效键位。
  4. 至少 Podman runtime 下 `Copy` (容器 → 主机) 与 `Export` (tar 导出) 完整跑通;Docker runtime 下给出明确"未支持"提示或同等实现。
  5. 不与 BR-032 的排序键位 (`N` / `Ctrl+N` / 鼠标点击表头) 冲突;两个 BUG 修复时统一规划键位分配。

---

## 维护说明

- 本文件由 `bugfix-requirements.md` 同步生成。任何条目状态升级为 `done` / `wontfix` 后,应:
  1. 在 `bugfix-requirements.md` 中更新对应条目状态与修复记录
  2. 从本文件中移除该条目
- 新 BUG 仍按"先在 `bugfix-requirements.md` 添加 → 若未完成再列入本文件"的流程处理。
- 优先级(`high` / `medium` / `low`)由业务影响决定,与状态无关。

## 关键事实:已 wontfix 的条目

- **BR-012** 鼠标点击行位置不对 — 用户决定取消鼠标点击行选中,改用滚轮上下选行(见 BR-030)。`clickListCursor` 与 `HitTest` 中与点击相关的分支不再维护。
