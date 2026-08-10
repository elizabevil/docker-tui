# 未关闭的 BUG / 需求

> 建立日期: 2026-08-01
> 最后同步: 2026-08-10 (BR-048 第一次修复入库;新增 BR-049..BR-053 共 5 条 UI / 排版问题)
> 来源: [bugfix-requirements.md](bugfix-requirements.md) 的快照
> 目的: 把 `bugfix-requirements.md` 中**当前未完成**的 BUG / 需求单独切出,
>       作为"待开发工作项"独立跟踪,不与已完成条目混在一起。

## 状态总览

| 状态 | 数量 | 备注 |
|---|---|---|
| `open` | 22 | 未开始或被搁置的需求(含本轮新增 BR-049..BR-053) |
| `partial` | 4 | 部分完成(BR-008 H=History / BR-032 键盘排序 / BR-035 Events 面板 / BR-041 form 部分修复) |
| `implementing` | 1 | 修复进行中(BR-009 卷详情 / 容器子视图) |
| `pending` | 1 | 用户尚未提供具体内容 |
| `wontfix` | 2 (BR-012 由 BR-048 重新激活;BR-030 被用户否决转 wontfix) |

## 排序与优先级

按"业务影响 + 优先级"降序:

| 编号 | 标题 | 优先级 | 状态 |
|---|---|---|---|
| [BR-048](#br-048) | 鼠标点击选中表格行不精确(点击行与选中行错位)— 已修复 click 命中，仍存在偏移残留 | high | partial |
| [BR-049](#br-049) | 容器页面 IP 列大部分为空,实际容器 inspect 可拿到 IP | high | open |
| [BR-050](#br-050) | Image 选中项展示与详情页 Registry 显示字段不完整 | medium | open |
| [BR-051](#br-051) | 表头点击排序只命中列名 / 鼠标命中区域需对齐列中心 | medium | open |
| [BR-009](#br-009) | 卷详情加载 + 容器子视图选中(部分修复) | high | implementing |
| [BR-033](#br-033) | TASK-019 高级容器动作未实现(Copy/Update/Diff/Export/Commit/Wait) | high | open |
| [BR-041](#br-041) | action bar / 输入框 / 长 label wrap / select 图标(部分修复) | high | partial |
| [BR-052](#br-052) | 表格鼠标点击仍存在偏移 — 渲染与命中合约未完全同步 | high | open |
| [BR-053](#br-053) | 表格列宽排版不均匀 — 例如 Network 页 ID 与 CREATED 间隙应均分 | medium | open |
| [BR-023](#br-023) | F2 唯一提供 runtime 选择器;其它页面不得用 C 刷新 Conn | medium | open |
| [BR-024](#br-024) | 详情源码视图 Ctrl+C 复制不完整 / 一次性失效 | medium | open |
| [BR-026](#br-026) | 详情页快捷键集合:支持 j/k/PgUP/PgDn/Filter,**不**支持 Space/R | medium | open |
| [BR-027](#br-027) | 容器日志页滚到底后向上划卡顿 | medium | open |
| [BR-029](#br-029) | 筛选框 Enter/Esc 双层退出 | medium | open |
| [BR-031](#br-031) | 容器资源占用(CPU/内存)未实时刷新 | medium | open |
| [BR-032](#br-032) | 表格排序 N/Ctrl+N/鼠标点击表头(键盘部分已完成) | medium | partial |
| [BR-035](#br-035) | Events 独立面板 (F3,已实现) + Network Connect/Disconnect(未实现) | medium | partial |
| [BR-036](#br-036) | 镜像 tarball 导入功能 | medium | open |
| [BR-037](#br-037) | docker / podman Registry Login | medium | open |
| [BR-038](#br-038) | 容器 Exec 页面 UI 不符合 shell 终端 | medium | open |
| [BR-040](#br-040) | Dialog 风格统一:四周透明 + panel 居中 | medium | open |
| [BR-042](#br-042) | 容器高级动作快捷键缺失(Rename/Top/Port 无绑定;TASK-019 默认键 nil) | medium | open |
| [BR-043](#br-043) | 表格多选标记后,光标移动至标记行无视觉区分 | medium | open |
| [BR-046](#br-046) | Audit 详情页无法进入:Enter 被全局拦截、d 键上下文不含 audit;Trace ID 截断 | medium | open |
| [BR-008](#br-008) | H 键应进 Help(未完成);镜像页 H 进入 History(已完成) | medium | partial |
| [BR-045](#br-045) | 容器日志入口:区分运行中 / 已停止容器(需求讨论) | low | open |
| [BR-028](#br-028) | (TBD - 待用户补充第 6 条) | - | pending |

---

## 详情

<a id="br-008"></a>

### BR-008 H 键应进 Help,F1 行为复用;H 键同时承担镜像 History 入口

- 状态: `partial` (2026-08-07 核查:镜像页 H 进 History 已完成 commit 9df9a39;H 全局未映射到 ActionHelp,仅 `?`/F1 进入 Help)
- 优先级: `medium`
- 症状:
  1. 在任何页面按 H 当前是 `Viewport.ToggleHeader()`(隐藏/显示 header),用户期望 H 直接进入 Help(与 F1 / `?` 一致)。
  2. 镜像详情页里 `History` 曾是一个分区,用户希望 History 提到镜像页顶层,**详情页不再显示 History**。
- 当前行为:
  - `keys/registry.go:81` 镜像上下文注册 `{ActionImageHistory, []string{KeyH}, images}`;H 键进入 `ModeHistory` 顶层页(commit 9df9a39),**不再触发 `ToggleHeader`**;`ToggleHeader` 仅存在于 `state/viewport.go:13`,无任何键绑定。
  - `keys/registry.go:39` `ActionHelp` 仅绑定 `KeyQmark` / `KeyF1`;**H 在非镜像上下文无绑定**(静默无操作)。
  - `ui/action/registry.go:185` 镜像页 footer 显示 `H = history.title`(已同步)。
  - 镜像详情页不再渲染 History 段(`ui/pages/detail/image.go` 分区构造不含 History)。
- 代码锚点:
  [internal/tui/keys/registry.go:81](internal/tui/keys/registry.go:81) `ActionImageHistory` ← `KeyH`(镜像上下文)
  [internal/tui/keys/registry.go:39](internal/tui/keys/registry.go:39) `ActionHelp` ← `KeyQmark` / `KeyF1`
  [internal/tui/state/viewport.go:13](internal/tui/state/viewport.go:13) `ToggleHeader`(无键绑定)
  [internal/tui/ui/pages/detail/image.go:14](internal/tui/ui/pages/detail/image.go:14) 镜像分区构造(不渲染 History)
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

<a id="br-035"></a>

### BR-035 Events 独立面板 (F3 键) + Network Connect/Disconnect

- 状态: `partial`
- 优先级: `medium`
- 症状:
  - 当前 `Engine.Events().Subscribe(...)` 已实现后台订阅,合并到资源更新路径,但**没有独立的 Events 浏览面板**。
  - `NetworkService` 接口只有 List/Create/Remove/Prune,**没有** Connect / Disconnect;无法动态调整容器的网络接入。
- 当前行为:
  - **Events 面板已实现**:`keys/registry.go:52` 全局注册 `{ActionEvents, []string{KeyF3}, app}`;`keyboard/events_keys.go`、`state/events_panel.go`、`ui/pages/events/view.go`、`page_templates` 路由 `ModeEvents` 均已存在。
  - **Network Connect / Disconnect 未实现**:`NetworkService` 接口仍只有 List/Inspect/Create/Remove/Prune;容器启动后只能使用启动时配置的网络。
- 代码锚点:
  [internal/tui/update/update_events.go:28](internal/tui/update/update_events.go:28) `subscribeEventsCmd`
  [internal/tui/keys/registry.go:52](internal/tui/keys/registry.go:52) `ActionEvents` ← `KeyF3`(已注册)
  [internal/data/runtime/streaming.go:64](internal/data/runtime/streaming.go:64) `EventService` 接口
  [internal/data/runtime/docker/service_network.go](internal/data/runtime/docker/service_network.go) Docker `NetworkService` 缺 Connect/Disconnect
  [internal/data/runtime/podman/service_network.go](internal/data/runtime/podman/service_network.go) Podman 同上
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

<a id="br-041"></a>

### BR-041 action bar / 输入框 / 长 label wrap / select 图标 — Form Dialog UI 一组修复

- 状态: `partial` (2026-08-07 核查:select marker 单行与两列布局已修复并有测试通过;TestLongPathCursorStaysVisible / TestFormRendersInSmallViewports 两个验收测试尚未编写)
- 优先级: `high`
- 症状(共 4 个):
  1. action bar(`Esc ▶ Cancel   Enter Confirm` 行)的背景色没有覆盖整行,左/右 padding 与 border 内侧出现缺口或断裂。
  2. `Container path` 等输入字段的值区没有下划线,与 label 视觉区分度不够。
  3. `Local destination (tar)`(label 长度 23,超过 `labelWidth=14`)的 label 与 value 视觉上分两行,value 跑到 padding 起始位置而非 value 列,行内未铺满。
  4. `FormSelect` / `FormMultiSelect` 的下拉 marker `▾` 应该和 value 同行单行显示,而不是出现在独立下一行。
- 当前行为(2026-08-07 核查):
  - `internal/tui/ui/widget/dialog/form_test.go` 现有 18 个测试;`TestFormDialogTwoColumnLayout`(label 完整可见)与 `TestFormDialogSelectCollapsedWithMarker`(▾ 单行可见)**已存在且通过**(`go test ./internal/tui/ui/widget/dialog/...` 通过)。
  - **验收测试缺口**:`TestLongPathCursorStaysVisible`(caret 可见 + 行宽 ≤ 60)与 `TestFormRendersInSmallViewports`(40/80/160 viewport 不溢出)在代码库中**不存在**(grep 无匹配),BR-041 验收标准 §1 的 4 个测试只覆盖了 2 个。
  - 修复提交:dc468a7 / 7c8ad03 / 828396a / 04d9bdd(select marker 单行、两列布局)。
- 代码锚点:
  [internal/tui/ui/widget/dialog/box.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/dialog/box.go)
  [internal/tui/ui/widget/dialog/form.go:185](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/dialog/form.go:185)
  [internal/tui/ui/widget/dialog/form.go:50](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/dialog/form.go:50)
  [internal/tui/ui/component/styles_load.go:347](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/styles_load.go:347) (`case "formInput"`)
  [internal/tui/ui/widget/dialog/form_test.go:27](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/dialog/form_test.go:27) `TestFormDialogTwoColumnLayout`
  [internal/tui/ui/widget/dialog/form_test.go:48](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/dialog/form_test.go:48) `TestFormDialogSelectCollapsedWithMarker`
- 期望行为:
  1. 按钮行(以及 dialog 其他内部行)的背景色从 `│`(左 border)内一格到右 `│`前一格连续无断裂。
  2. `FormText` / `FormPath` / `FormInt` 字段的值区带下划线视觉提示,聚焦态保留原有加粗 + caret 强调。
  3. 长 label 行(label 实际长度 > `labelWidth`):label 完整可见(测试 `TestFormDialogTwoColumnLayout` 必须通过),value 在 value 列起点。
  4. `FormSelect` / `FormMultiSelect` 的 marker `▾` 与 value 同行单行(`TestFormDialogSelectCollapsedWithMarker` 通过)。
- 验收标准:
  1. `go test ./internal/tui/ui/widget/dialog/... ./internal/tui/ui/component/... -count=1` 全绿,包括 `TestFormDialogTwoColumnLayout` / `TestFormDialogSelectCollapsedWithMarker` / `TestLongPathCursorStaysVisible` / `TestFormRendersInSmallViewports`(后两个**待编写**)。
  2. 视觉:viewport 40x16 / 80x24 / 160x40 下,FormDialog 输出行宽恒 ≤ `dialogW`,无 42/162/322 溢出。
  3. 视觉:action bar 背景从 border 到 border 连续;输入值区有下划线;长 label 行 layout 完整;select marker 单行。
  4. 不引入 `github.com/fatih/color`(BR-000);不动 `config.Palette` schema 与 12 个编译期 hex 与 `style.Colors.BG`;不动主题 JSONC 颜色值。
  5. `go vet ./...` 与 `go test ./internal/... ./cmd/...` 全量回归通过。

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

- 状态: `partial` (2026-08-07 核查:键盘部分 N/Ctrl+N 已完成 commit d7e6a57;鼠标表头点击排序仍未实现)
- 优先级: `medium`
- 症状:
  - 当前表格排序由 `O`(`toggleSortColumn` 切换列)与 `Ctrl+O`(`toggleSortDirection` 切换升降序)承担。
  - 用户期望:`N` = 正排序(升序),`Ctrl+N` = 反排序(降序);此外,**鼠标点击表头**也能排序。
- 当前行为:
  - `internal/tui/keyboard/keyboard.go`:`KeyN` 调 `setSortDirection(true)`(升序),`KeyCtrlN` 调 `setSortDirection(false)`(降序),对当前排序列生效(选项 A 语义,commit d7e6a57)。`KeyO` / `KeyCtrlO` 的 `toggleSortColumn` / `toggleSortDirection` 逻辑保留。
  - 排序改动后 `Cursor` 与 `ViewOffset` 归零。
  - **没有**鼠标点击表头的处理;`mouse.go` 中只有 `clickListCursor` / `HitTest` / 滚轮事件,**没有**表头点击路由。`HitHeader` 常量存在于 `mouse.go` 但 `ApplyMouseClick` 仅处理 `HitPanel` 分支。
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

<a id="br-042"></a>

### BR-042 容器高级动作快捷键缺失(Rename/Top/Port 无绑定;TASK-019 默认键为 nil)— 需讨论与 action bar 的重复策略

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 容器页面部分高级功能**没有直接快捷键**:`Rename`(重命名)、`Top`(进程)、`Port`(端口映射)在 `keys/registry.go` 中完全无绑定。
  - TASK-019 的 6 个动作(`Update` / `Diff` / `Export` / `Commit` / `Wait` / `Copy`)在 registry.go:68-73 的默认键为 `nil`,仅能通过 `;` 打开的 action bar 触达。
  - 用户期望:一部分动作既能"直接快捷键"触达(方便熟练用户),也能在 action bar 中展示(不丢失高级功能入口)。需要讨论**哪些操作可以与 action 重复绑定**。
- 当前行为:
  - `registry.go` 仅注册了基础动作与 Detail(d 键,上下文仅 containers/images/volumes/networks)。`ActionContainerRename` / `ActionContainerTop` / `ActionContainerPort` 未出现在任何默认绑定中。
  - TASK-019 动作的默认键为 `nil`(registry.go:68-73),`actionbar` 组件(registry_test.go:17)断言它们出现在 `;` 弹出的 bar 中。
  - BR-033 负责 TASK-019 的 **handler 缺失**;本条目只讨论**键位绑定与重复策略**,两者互补。
- 期望行为:
  1. 明确"哪些动作同时拥有直接快捷键 + action bar 入口"的清单(讨论产出,记录到本条目)。
  2. 至少为 `Rename` 补一个默认键;TASK-019 动作在 BR-033 修复时一并重设非冲突默认键。
  3. 讨论结论需与 BR-033 / BR-032(排序键位)统一规划,避免再次冲突。
- 验收标准:
  1. 讨论结论写入本条目(状态可转为 `done` 或拆分出新 BR)。
  2. 容器页可直接按键触发 Rename 等动作;action bar 同时保留入口。

<a id="br-043"></a>

### BR-043 表格多选标记后,光标移动至标记行无视觉区分

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 表格多选(mark mode)后,标记行有底色;但**光标移动到标记行上看不出光标位置**——标记样式优先级高于选中样式,选中态被吞掉。
  - 用户指定修复方向原文:"当光标移动至多选列表时,文字颜色设置为光标的背景色"。
- 当前行为:
  - `rows.go:41-43`:`if marked` 分支优先于 `selected`,取 `RowStyleMarked`;两分支都走 `Width` + `Background` 铺满整行。
  - marked 样式来源 `styles_load.go:366-367`:`Marked.Background = theme.Data.MarkedBackground`、`Marked.Color = NameForeground`。
  - 因此光标所在的标记行与普通标记行视觉完全一致,无法区分当前行。
- 期望行为:
  1. 光标位于标记行时,该行**文字颜色设为光标背景色**(即"反色"提示:文字 = 光标背景色),与普通标记行/普通选中行均区分开。
  2. 光标离开后恢复 marked 样式。
- 验收标准:
  1. 多选 2+ 行后移动光标:光标所在标记行与其余标记行肉眼可区分。
  2. 未标记行的选中样式(现有)不被破坏;`go test ./internal/tui/ui/component/` 通过。

<a id="br-045"></a>

### BR-045 容器日志入口:区分运行中 / 已停止容器 — 需求讨论

- 状态: `open`
- 优先级: `low`
- 症状:
  - 容器日志视图对"运行中"与"已停止"容器无任何区分;已停止容器按 Enter 同样进入日志页。
  - 用户需求:在 UI 上区分运行中 / 已停止容器的日志入口(需求讨论项)。
- 当前行为:
  - `doLogAction`(container_action.go:185-193)对任意选中容器直接进入日志视图,**无 running/stopped 判断**。
  - podman 本身支持对已停止容器读取 logs(runtime 层可返回历史日志),因此功能上可用,仅缺 UI 区分。
- 期望行为(待讨论确认后落地):
  1. 日志页标题/页内标识容器当前状态(如 `● Running` / `■ Stopped`)。
  2. 入口处(如 action bar 或 Enter 行为)对已停止容器给出提示或差异化文案。
- 验收标准:
  1. 已停止容器进入日志页时,页面可见容器状态标识。
  2. 讨论结论记录到本条目。

<a id="br-046"></a>

### BR-046 Audit 详情页无法进入:Enter 被全局 ActionEnter 拦截、d 键上下文不含 audit;Trace ID 截断 16 字符

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 用户报告:Audit 页面信息太长存在截断;希望添加 enter 或 d 键进入详情页查看详情(类似容器页)。
  - 代码级定位:audit 面板按 Enter **完全无反应**;按 d 也无反应。`ModeAuditDetail` 目前**没有任何键盘路径可达**。
- 当前行为:
  - **Enter 被全局 ActionEnter 先拦截**:registry.go:47 `{ActionEnter, []string{KeyEnter}, main}`;keyboard.go:63-65 在 `handlePanelFallbacks`(line 67)之前执行;`doEnterAction`(delete_action.go:42-68)**无 `PanelAudit` 分支** → 返回 `m, nil`(静默无操作)。`handleAuditPanelKey` 的 Enter 分支(audit_keys.go:23-31)是**死代码**,永远收不到 Enter。
  - **d 键上下文不含 audit**:registry.go:53 `ActionDetail` 仅绑定 View containers/images/volumes/networks。
  - `ModeAuditDetail` 目前唯一赋值点是 audit_keys.go:29(死代码内),实际不可达;mouse.go:255 对 PanelAudit 只做 clickListCursor(移动光标),不进详情。
  - **截断**:audit view.go:99-102 将 Trace ID 截断为 16 字符 + `"..."`。
  - E 键循环过滤器(All→Errors→Warnings→All)已可用(audit_keys.go:32-44)。
- 期望行为:
  1. Audit 面板按 Enter(或 d)进入 `ModeAuditDetail` 详情视图;详情 Esc/Enter 返回列表(handleAuditDetailKey 已实现)。
  2. Trace ID 完整显示或提供展开方式(至少不在首屏截断关键信息)。
- 验收标准:
  1. 在 audit 面板按 Enter 可进入详情;详情 Esc 返回。
  2. Trace ID 可完整阅读(或可展开);`go test ./internal/tui/keyboard/ ./internal/tui/ui/pages/audit/` 通过。

<a id="br-048"></a>

### BR-048 鼠标点击选中表格行不精确(点击行与选中行错位)

- 状态: `partial`
- 优先级: `high`
- 症状:
  - 2026-08-09 用户反馈:"当前支持鼠标操作,但是鼠标不够精确,例如选中表项"。点击列表行选中位置与视觉点击行相差 1 到数行,滚动后误差更大。
  - 2026-08-09 用户确认方向:**修复点击精确性**(保留点击选行),否决 BR-030 取消方案。
- 修复记录 (2026-08-10):
  - `component.TableHitLayout` 共享布局契约:`RenderTable` 与 `mouse_lists.applyMouseListClick` 同时使用 `DataStartRow` / `HeaderRow` / `DataRowAt`,消除前置行 / spacing 漂移。
  - `mouse_click.applyMouseHeaderSort` 根据 `HeaderRow` 命中触发表头点击排序;Compose 右栏另由 `applyComposeSubviewClick` 处理。
  - 容器列默认排序改为 `ContainerSortByName`(修复 BR-048 反复出现的"看起来随机排序"问题根因之一)。
- **未修复**(转 BR-052 跟踪):
  - 各表行仍可能在 preview 与 table body 之间的边界出现 ±1 行偏差,详见 BR-052。
  - 表头点击排序只命中列名居中位置,其它空白处不触发,详见 BR-051。
- 当前行为(逻辑链梳理,详见 [bugfix-requirements.md BR-048](bugfix-requirements.md#br-048)):
  - 渲染侧 `RenderTable` 在 panel body 内先输出 SelectionInfo 预览(1+pad)、PageInfo(1)、Header(1) 后才到数据行;数据行 0 的真实偏移 ≈ 2~4 行(有 mark 时再加 2)。
  - `HitTest` 的 `row := y - bodyTop` 把 bodyTop 当作数据行 0,未扣除前置行 → 固定偏差。
  - `clickListCursor` 的 `*cursor = row` 未加 `ViewOffset`,滚动后点击错位;也未映射容器页多行行高。
- 代码锚点:
  [mouse.go:205](internal/tui/ui/app/mouse.go:205) `HitTest`
  [mouse.go:300](internal/tui/ui/app/mouse.go:300) `clickListCursor`
  [table.go:66](internal/tui/ui/component/table.go:66) `RenderTable`
  [helpers.go:29](internal/tui/ui/component/helpers.go:29) `CalcTableRowHeight`
  [view.go:149](internal/tui/ui/pages/containers/view.go:149) `selectedRowForCursor`
- 期望行为:
  1. 换算链 `cursor = viewOffset + (clickRow − 前置行数)`,前置行数按 ActivePanel/状态动态计算。
  2. 滚动、mark、ModeFilter、容器多行行高场景下均精确;表头/预览/页码区点击不移动 cursor。
- 验收标准:
  1. 各列表页未滚动点击精确;滚动后点击 item = offset + 可见行号。
  2. 表头/预览/页码点击不移动 cursor;`go test ./internal/tui/ui/app/` 通过。

---

<a id="br-049"></a>

### BR-049 容器页面 IP 列大部分为空,实际 inspect 可拿到 IP

- 状态: `open`
- 优先级: `high`
- 症状:
  - 用户反馈:容器页面 IP 字段几乎全部渲染为 `—`(占位),但同一个容器在 Docker `inspect` 输出中能拿到 IP 地址。
  - 截图/示例:列表中 ~90% 行 IP 列为 `—`;只有少数 bridge + 自定义网络上的容器能显示真实 IP。
- 当前行为(代码锚点):
  [internal/tui/ui/pages/containers/view.go:188](internal/tui/ui/pages/containers/view.go:188) `ip` 列使用 `c.IPs[0]`(`containers.CellValue`)；`runtimeapi.ContainerSummary.IPs` 由 docker / podman adapter 填充。
  [internal/data/runtime/docker/containers.go:64](internal/data/runtime/docker/containers.go:64) `IPs` 来自 `NetworkSettings.IPAddress`(旧版字段);`docker inspect` 返回的 `NetworkSettings.Networks[].IPAddress` 才是当前多网络的真实 IP。
  [internal/data/runtime/podman/mappers.go](internal/data/runtime/podman/mappers.go) Podman 路径同上,只读取 `NetworkSettings.IPAddress`,忽略多网络 map。
- 根因假设:
  1. **adapter 漏读多网络**:Docker/Podman 当前容器处于 `bridge` 默认网络,`NetworkSettings.IPAddress` 经常为空字符串,真实 IP 落在 `NetworkSettings.Networks[<name>].IPAddress`。`IPs` 切片始终为空 → 列表渲染 `—`。
  2. **运行时未刷新**:某些运行中容器首次 fetch 时还未拿到网络,后续 inspect 也未增量更新 IP。
  3. **字段定义缺失**:`runtimeapi.ContainerSummary` 没有 `Networks` map,只有 `IPs []string`,无法表达 `<network, ip>` 对应关系。
- 期望行为:
  1. `runtimeapi.ContainerSummary.IPs` 改为按 `bridge / custom / host / none` 顺序聚合所有网络的 IP(优先常用网络)。
  2. Docker/Podman adapter 同步把 `Networks.*.IPAddress` 写回 `IPs`,不再依赖单一 `NetworkSettings.IPAddress`。
  3. 列表渲染对单 IP 直接显示;多 IP 用 `,` 分隔(空间允许时);不可用时回退 `—`。
- 验收标准:
  1. 同一台主机运行 `nginx:alpine` 容器 → 列表 IP 列显示真实 IP(非 `—`)。
  2. 多网络容器显示所有网络 IP,顺序合理。
  3. `go test ./internal/data/runtime/... ./internal/tui/ui/pages/containers/` 通过。

---

<a id="br-050"></a>

### BR-050 Image 选中项展示与详情页 Registry 字段不完整;Volume 预览仅需 driver / mountpoint

- 状态: `open`
- 优先级: `medium`
- 症状:
  - Image 列表选中项展示(selection preview)当前为 `id  name  image-short`,但用户要求 `id` 字段后跟 **完整 Registry**(而非仅 name)+ name;详情页 Registry 字段当前拼接 `reg/name:tag`,若为空则显示 `<unnamed>`,与列表预览口径不一致。
  - Volume 选中项预览当前为 `name  driver  mountpoint`,用户确认 **只需 driver 与 mountpoint**(不需要 name,因为 name 已在列内)。
- 当前行为:
  [internal/tui/ui/pages/images/view.go:130-147](internal/tui/ui/pages/images/view.go:130-147) 选择 preview:`fmt.Sprintf("%s  %s  %s", short, name, image)`;name 为空时退化为 `short  <unnamed>  <image>`。
  [internal/tui/ui/pages/images/detail.go](internal/tui/ui/pages/images/detail.go) 详情页 `Registry` 字段拼接 `img.Registry + "/" + name + ":" + tag`,`img.Registry` 可能缺失。
  [internal/tui/ui/pages/volumes/view.go:76-85](internal/tui/ui/pages/volumes/view.go:76-85) 选择 preview:`v.Name + "  " + v.Driver + "  " + v.Mountpoint`;用户希望去掉 name。
- 期望行为:
  1. **Image 列表选中预览**:`<short-id>  <registry>/<name>:<tag>` 三段(若 registry 为空退化为 `docker.io/<name>:<tag>`)。
  2. **Image 详情 Registry** 字段同样展示完整 `<registry>/<repo>:<tag>`,与列表预览口径一致。
  3. **Volume 列表选中预览**:仅 `<driver>  <mountpoint>`(不显示 name,避免重复列)。
- 验收标准:
  1. Image 列表 preview 与详情页 Registry 字段口径一致,均为完整 ref。
  2. Volume 列表 preview 仅显示 driver 与 mountpoint;`go test ./internal/tui/ui/pages/images ./internal/tui/ui/pages/volumes` 通过。

---

<a id="br-051"></a>

### BR-051 表头点击排序:仅命中列名才切换排序

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 用户反馈:**只有点击列名(表头文字)才会触发排序切换**,点击列与列之间的空白处不触发。
  - 当前实现 `mouse_click.applyMouseHeaderSort` 调用 `component.ResolveColumnHit(rep.Panel.bodyLeft, x, cols, widths, gap)`,只在 X 落在列宽区间内时返回 true;空白处被丢弃。
- 当前行为(代码锚点):
  [internal/tui/ui/app/mouse_click.go:14-22](internal/tui/ui/app/mouse_click.go:14-22) `applyMouseHeaderSort` 调用 `component.ResolveColumnHit`,仅命中列内 X 才视为列点击。
  [internal/tui/ui/component/column_hit.go:14-26](internal/tui/ui/component/column_hit.go:14-26) `ResolveColumnHit` 按 widths 切片定位 X 所属列;gap 落在两列之间时落空。
- 期望行为:
  1. 表头区域内(HeaderRow)任意 X 都应解析到最近的列;gap 区段向左或向右归一化到最近列(例如靠近左列优先,靠近右列优先,平分时往左)。
  2. 不再要求用户精确点击列文字;Footer / 数据区点击行为不变。
- 验收标准:
  1. 表头任意位置(列内 + gap)点击都触发排序;相邻列不会误触发。
  2. 非表头行点击仍走 list 点击逻辑;`go test ./internal/tui/ui/app/ ./internal/tui/ui/component/` 通过。

---

<a id="br-052"></a>

### BR-052 表格鼠标点击仍存在偏移;渲染与命中合约未完全同步

- 状态: `open`
- 优先级: `high`
- 症状:
  - 用户反馈:虽然 BR-048 修复了一部分命中,但点击列表仍出现 1 行偏差,尤其在:
    - 容器列表有 mark banner(多 2 行)
    - 容器页多端口绑定(每个容器多行)
    - 镜像页 Selection preview 关闭后 1 行变化
- 当前行为(代码锚点):
  [internal/tui/ui/app/mouse_lists.go:54-77](internal/tui/ui/app/mouse_lists.go:54-77) `mouseListTargetFor` 内 `markedBannerLines(m, state.PanelContainers)` 写死返回 `1`;但 banner 实际是容器页 `MarkedItemsBanner` 加 `\n` 两行 + 数据行 0 之间的 preview pad,口径未统一。
  [internal/tui/ui/app/mouse_lists.go:71-78](internal/tui/ui/app/mouse_lists.go:71-78) 容器 rowCount 直接读 `len(containers.FormatPorts(items[i].PortBindings, false))`,与渲染端 `portList := FormatPorts(c.PortBindings, w < 80)` 的 `compact` 标志不一致。
  [internal/tui/ui/pages/containers/view.go:81](internal/tui/ui/pages/containers/view.go:81) `FormatPorts(c.PortBindings, w < 80)` 在窄屏使用 compact 模式,可能只占 1 行;鼠标侧用 `FormatPorts(..., false)` 计多行,导致点击命中错位。
  [internal/tui/ui/component/table.go:117-121](internal/tui/ui/component/table.go:117-121) `RenderTable` 现在硬编码 `sb.WriteString("\n")` 而非 `SelectionPreviewPadding()`,与 `TableHitLayout.DataStartRow()` 内 `1 + SelectionPreviewPad` 不一致。
- 期望行为:
  1. 渲染侧与命中侧 **完全共享** 同一份前置行公式:
     `DataStartRow = bannerLines + (topLabel?1:0) + (preview? 1+previewPad : 0) + (pageInfo?1:0) + 1`。
  2. 容器 rowCount 使用渲染时的 `FormatPorts(compact)` 标志,与渲染端一致。
  3. 渲染侧 preview 与 `TableHitLayout` 都从 `selectionInfo != ""` 与 `SelectionPreviewPadding()` 取值,不再写死。
- 验收标准:
  1. 容器页 mark 状态下点击第一项仍命中第一项,不再偏移。
  2. 容器多端口绑定下,点击第二端口行命中容器 0(已通过 BR-048 部分覆盖)。
  3. 镜像页 Selection preview 关闭前后命中行一致;`go test ./internal/tui/ui/app/ ./internal/tui/ui/pages/` 通过。

---

<a id="br-053"></a>

### BR-053 表格列宽排版不均匀 — Network ID 与 CREATED 间隙应均分剩余空间

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 用户反馈:Network 页 ID 列与 CREATED 列之间出现较大空隙;同样在 Image 宽屏下 registry / name / tag / id / created 之间间距不均衡,看上去列宽分配不够"均匀"。
  - 当前实现 `component.resolveTableLayout` 调用 `tables.ResolveContentLayout`,后者按 `colW / min / fill` 权重分配,**剩余空间只给 `fill>0` 的列**;`ID`/`CREATED` 标记为 `fixed` 不吸收剩余空间,如果中间列 fill 总权重小,中间会留出"看起来很大"的间隙。
- 当前行为(代码锚点):
  [internal/tui/tables/config.go:298-359](internal/tui/tables/config.go:298-359) `growColumnsTowardContent` + `fillColumns` 只在 `fill>0` 的列上分配剩余空间;fixed/min-only 列不会扩展。
  [internal/tui/tables/config.go:262-296](internal/tui/tables/config.go:262-296) `columnSizing.fluid` 仅当 `Fixed<=0` 时为 true;`Fixed > 0` 时列是 rigid,完全按 fixed 宽度。
  [internal/tui/ui/component/table_layout.go:21-36](internal/tui/ui/component/table_layout.go:21-36) `resolveTableLayout` 不再做二次重平衡。
- 期望行为:
  1. 表总宽度 - 列基础宽度 之差,在**所有流体列**(fluid = true)之间按"等分 / 接近等分"分配,而非仅 fill>0 列。
  2. fill 权重仍然生效(高 fill 列稍宽),但其余流体列也获得至少 1 cell 增量,避免相邻列中间出现大间隙。
  3. 固定列(fixed)保持原宽,不参与扩张。
- 验收标准:
  1. Network 页 ID 与 CREATED 之间无明显空隙;在 120 与 200 列宽下视觉均分。
  2. Image 宽屏下 registry / name / tag / id / created / size 列宽合理,无空列。
  3. 现有 `tables/config_test.go` 的列宽断言继续通过(必要时放宽 tolerance);`go test ./internal/tui/tables/ ./internal/tui/ui/component/` 通过。

---

---

## 维护说明

- 本文件由 `bugfix-requirements.md` 同步生成。任何条目状态升级为 `done` / `wontfix` 后,应:
  1. 在 `bugfix-requirements.md` 中更新对应条目状态与修复记录
  2. 从本文件中移除该条目
- 新 BUG 仍按"先在 `bugfix-requirements.md` 添加 → 若未完成再列入本文件"的流程处理。
- 优先级(`high` / `medium` / `low`)由业务影响决定,与状态无关。

## 关键事实:已 wontfix 的条目

- **BR-012** 鼠标点击行位置不对 — 原因 BR-030 决定取消鼠标点击行选中;2026-08-09 用户重新确认**修复点击精确性**(BR-048),本条目重新激活,不再按 wontfix 处理。
- **BR-030** 取消鼠标点击行选中,改用滚轮上下选行 — 2026-08-09 用户否决该方向,改为修复点击精确性(见 BR-048)。
