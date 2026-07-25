# 历史 Bug 修复需求

> 建立日期: 2026-07-20
> 目的: 独立记录“历史遗留问题”的修复需求，不与产品路线图和实现现状文档混写。

本文是修复需求台账，不是实现说明。实现现状仍以 [README.md](README.md)、[architecture.md](architecture.md) 和代码为准。

## 使用规则

- 只记录已经复现、已从代码验证、或已有明确用户反馈的问题。
- 每条需求必须包含: 症状、当前行为、代码锚点、期望行为、验收标准。
- 需求编号固定为 `BR-xxx`,便于在设计文档、提交记录和 PR 说明中交叉引用。
- 如果问题只是文档漂移,不进入本文;应直接修正文档。

## 颜色库选型(架构约束)

- 项目所有颜色 / 样式 / 布局输出统一走 `charm.land/lipgloss/v2`(配合 `charm.land/bubbletea/v2`),`internal/tui/ui/style/style.go` 与 `internal/tui/ui/component/styles*.go` 均基于 lipgloss 的 `lipgloss.Color` 与 `lipgloss.NewStyle()` 构建。
- 已确认**不引入** `github.com/fatih/color` 作为通用颜色库:lipgloss 与 fatih/color 是不同设计哲学(后者只支持前景色、不带 background / padding / width / 对齐,且 ANSI 调色板与 lipgloss 的 palette 解耦),强行混用会导致:
  1. 同一"红色"在 TUI 文字与日志文字里色值不一致;
  2. `color.NoColor` 与 `lipgloss.SetColorProfile` 在 TTY / 非 TTY 检测上不同步;
  3. bubbletea v2 的事件循环 / 渲染协议依赖 lipgloss 输出,移除 lipgloss = 重写整个 TUI 渲染层;
  4. 无法解决 BR-013(头部信息重复)、BR-014(footer 截断)、BR-015(选中行背景未覆盖全)等问题,换库治标不治本。
- 后续如需在 CLI 非交互输出(--version、--help、错误日志)做着色,优先使用 lipgloss 的 `lipgloss.NewStyle().Render` 保持调色板一致;若确实需要 CLI 风格的简化着色,**也只能新增一个独立的 CLI 输出辅助函数**,而不是把 fatih/color 引入到 TUI 渲染层。
- 本约束视为 BR-000,任何"用 fatih/color 替换 lipgloss"的提案都按 wontfix 处理。

## 状态定义

| 状态 | 含义 |
|---|---|
| `open` | 已确认问题，尚未开始修复 |
| `designing` | 正在补设计或拆方案 |
| `implementing` | 已开始代码修改 |
| `verifying` | 已实现，正在补测试或回归验证 |
| `done` | 已合并或已确认完成 |
| `wontfix` | 当前明确不修，需写原因 |

## 当前修复清单

### BR-001 镜像详情 `History` 区块没有真实数据

- 状态: `done`
- 优先级: `high`
- 症状:
  镜像详情页已经渲染出 `History` 分区，但当前始终显示占位文案，无法查看 layer history。
- 当前行为:
  `InspectImageDetail()` 为普通镜像调用 runtime layer history API，为 manifest list 投影平台变体；详情页根据显式来源状态渲染真实记录、空数据、加载中或获取失败。
- 代码锚点:
  [internal/data/docker/images.go](/home/debi/IdeaProjects/docker-tui/internal/data/docker/images.go:277)
  [internal/tui/ui/pages/detail/image.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/detail/image.go:14)
- 期望行为:
  镜像详情中的 `History` 应作为一级区段保留；普通镜像与 manifest list 允许使用不同子语义和排版模型。镜像引擎可提供历史数据时，普通镜像至少展示 `CreatedBy`、`Size`、`Comment` 等核心字段；manifest list 则展示其平台变体结构。确实拿不到时，后续再定义降级语义。
- 验收标准:
  1. 普通镜像详情页不再默认落到“暂不可用”占位文案。
  2. 至少支持一条有效历史记录渲染。
  3. 无历史数据时，错误或降级文案来源明确，可区分能力缺失与实现缺失。

### BR-002 自定义键位配置未生效

- 状态: `done`
- 优先级: `high`
- 症状:
  配置文件里已经存在 `keymap.*` 结构，但运行时不会读取用户自定义键位。
- 当前行为:
  `config.Keymap` 已经通过统一动作注册表编译为运行时绑定；未配置动作使用注册表默认值，已配置动作覆盖默认绑定。
- 代码锚点:
  [internal/data/config/types.go](/home/debi/IdeaProjects/docker-tui/internal/data/config/types.go:27)
  [internal/tui/keyboard/keyboard.go](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/keyboard.go:185)
  [internal/tui/keys/mapping.go](/home/debi/IdeaProjects/docker-tui/internal/tui/keys/mapping.go:1)
- 期望行为:
  程序启动后应将配置中的 `keymap.*` 与默认映射合并，未配置的动作保留默认值，已配置动作按用户配置覆盖。
- 验收标准:
  1. 修改 `config.yml` 后，至少一组自定义键位可以实际触发目标动作。
  2. 未配置的动作不回归、不失效。
  3. 帮助页和 footer 展示与实际生效键位保持一致，或至少明确标注“当前显示默认键位”。

### BR-003 默认键位示例与实际按键体系不一致

- 状态: `done`
- 优先级: `medium`
- 症状:
  默认配置里的 `keymap` 示例仍使用旧的大写键位写法，如 `S/K/P`，而运行时已经转向 `Ctrl+S`、`Ctrl+K`、`Ctrl+P`。
- 当前行为:
  默认配置、运行时解析、Help 和 Footer 已统一使用动作注册表及其有效绑定，不再将大写单键作为 stop / kill / pull 的默认说明。
- 代码锚点:
  [internal/data/config/default.jsonc](/home/debi/IdeaProjects/docker-tui/internal/data/config/default.jsonc:37)
  [internal/tui/keyboard/keyboard.go](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/keyboard.go:13)
  [internal/tui/keys/mapping.go](/home/debi/IdeaProjects/docker-tui/internal/tui/keys/mapping.go:6)
- 期望行为:
  默认配置、帮助信息、运行时行为应统一为同一套键位语义；如果保留兼容层，也要明确写出兼容规则。
- 验收标准:
  1. 默认配置示例与实际默认映射一致。
  2. 帮助页、footer、README 中不再出现与运行时冲突的默认键位说明。
  3. 如保留兼容输入，需有最小测试或代码注释说明。

### BR-004 过滤模式的死配置与现行为不一致

- 状态: `done`
- 优先级: `medium`
- 症状:
  过滤交互已经变为“资源表输入即过滤、日志页按 Enter 应用搜索”，但 `SearchTimer` / `SearchTick` / `StartSearchDebounce()` 仍保留旧的防抖模型。
- 当前行为:
  资源 Filter 与日志 Search 已使用独立模式和输入状态；资源输入即时生效，日志搜索按 Enter 应用。旧防抖状态、消息和 `ui.searchDebounceMs` 已删除。
- 代码锚点:
  [internal/tui/keyboard/keyboard.go](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/keyboard.go:198)
  [internal/tui/update_tick.go](/home/debi/IdeaProjects/docker-tui/internal/tui/update_tick.go:121)
  [internal/tui/keyboard/navigate.go](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/navigate.go:103)
  [internal/tui/state/app.go](/home/debi/IdeaProjects/docker-tui/internal/tui/state/app.go:30)
- 期望行为:
  资源页与日志页应拆分为不同查询语义: 结构化资源页使用 `Filter`，日志页使用 `Search`。资源页过滤默认即时生效，并采用双 `Esc` 退出筛选功能的交互；日志页搜索不改变数据集，只负责匹配与跳转。
- 验收标准:
  1. `Filter` 与 `Search` 拥有明确分离的模式语义。
  2. 资源页过滤行为、退出行为与配置含义一致。
  3. 不再保留无效的旧搜索防抖路径。
  4. 相关文档与帮助说明同步更新。

### BR-005 消息 / 日志管理缺少统一审计模型

- 状态: `done`
- 优先级: `high`
- 症状:
  当前消息通知、底部操作日志、状态提示、正文日志和潜在落盘日志彼此分散，缺少统一事件模型和追溯链路。
- 当前行为:
  用户显式触发的资源操作已统一生成 requested / started / terminal 审计记录；终态记录投影到顶部通知和 Footer，并按日写入 JSONL。普通查询、导航和非审计提示只进入 UI 消息投影，不污染审计历史。
- 代码锚点:
  [internal/tui/state/app.go](/home/debi/IdeaProjects/docker-tui/internal/tui/state/app.go:58)
  [internal/tui/ui/component/toast.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/toast.go:1)
  [internal/tui/ui/widget/footer/oplog.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/footer/oplog.go:1)
- 期望行为:
  建立面向用户业务操作的统一审计日志模型。通知与操作日志应从同一操作链派生，并通过 `trace_id` 串联；应用应支持落盘审计日志，记录用户对资源或数据源上下文的实际操作，而不是混入程序内部运行细节。
- 验收标准:
  1. 用户业务操作拥有统一事件模型与 `trace_id`。
  2. 审计日志包含被操作对象的最小统一字段 `type/id/name`。
  3. 资源对象扩展字段使用强类型 `struct + interface` 约束，而不是松散 map。
  4. 落盘日志与界面通知/操作历史基于同一事件来源派生。
  5. 容器、镜像、卷、网络、Compose、运行时切换和 Exec 操作均携带完整 trace。

### BR-006 容器详情页 D 键后只显示“加载中”

- 状态: `open`
- 优先级: `high`
- 症状: 本地 podman (no-cgo 模式) 容器页面按 D 进入详情页后,长时间停留在"加载中"占位文案;按 s 切换到 YAML / JSON 也不显示原始数据。
- 当前行为:
  `doInspectAction` 在 `Update` 消息处理路径上**同步**调用 `Engine.Containers().Inspect(ctx, id)`,同步 `SetContainerDetail(detail)` 后再 `ToDetail(...)` 跳转。Podman REST 路径在 no-cgo 模式下连接 unix socket 阻塞时间较长,会导致 Update 阶段被阻塞,事件循环被卡住,UI 无法进入正常渲染。即便 inspect 成功返回,`renderSourceView` 也只在 `DetailSourceType == yaml|json` 时进入,而 `SetContainerDetail` 没有把 `DetailSourceType` 显式置空,加上 `Inspect` 抛错时 `RecordError` 没有把错误状态写入 `DetailState`,错误会被静默吞掉,UI 永远停在 `hint.loading_detail` 占位。
- 代码锚点:
  [internal/tui/keyboard/container_action.go:361](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/container_action.go:361) `doInspectAction` (同步 inspect, 无 cmd)
  [internal/tui/keyboard/delete_action.go:74](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/delete_action.go:74) `doDetailAction`
  [internal/tui/state/detail.go:60](/home/debi/IdeaProjects/docker-tui/internal/tui/state/detail.go:60) `SetContainerDetail` (未写 `DetailSourceType`)
  [internal/tui/ui/pages/detail/view.go:53](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/detail/view.go:53) `RenderView`
- 期望行为:
  1. 容器 inspect 必须走 `tea.Cmd` 异步路径,`Update` 中只发请求、收到 `inspectResultMsg` 后再写状态与跳转,事件循环不能被 socket I/O 阻塞。
  2. inspect 失败时在 detail 页给出明确错误占位(参考镜像详情的错误文案),不能静默吞掉。
  3. 进入详情时强制把 `DetailSourceType` 重置为 `DetailSourceSection`,避免旧值残留导致 yaml/json 视图永远空。
- 验收标准:
  1. podman no-cgo 启动后,D 键在事件循环无明显阻塞,加载占位 <= 2s 后被替换为真实数据或显式错误。
  2. inspect 失败时,详情页出现可读错误而非 “加载中”。
  3. 按 s 切换 yaml / json 时能显示完整原始数据(前提是 inspect 已成功)。

### BR-007 镜像详情页不显示 yaml / json 原始数据

- 状态: `open`
- 优先级: `high`
- 症状: 镜像页面按 D 进入详情页后,只能看到 `image.go` 渲染的分区内容;按 s 切到 yaml / json 视图,输出空白,看不到原始 inspect 数据。
- 当前行为:
  `doImageDetail` 调用 `m.Detail.OpenImage(id, title, data)`。`OpenImage` 只设置了 `ImageDetailID / DetailTitle / ImageDetailData`,**没有把 inspect 原始字节写入 `DetailRawJSON`**。`renderSourceView` 直接对 `DetailRawJSON` 做 `sonic.Unmarshal`,空字节切片 unmarshal 失败,`text` 落到空字符串兜底,只渲染出滚动 footer。这是与容器/卷/网络的 `SetContainerDetail / SetVolumeDetail` 行为不一致造成的缺失。
- 代码锚点:
  [internal/tui/state/detail.go:35](/home/debi/IdeaProjects/docker-tui/internal/tui/state/detail.go:35) `OpenImage` (未写 `DetailRawJSON / DetailResourceType`)
  [internal/tui/keyboard/image_action.go:99](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/image_action.go:99) `doImageDetail`
  [internal/tui/ui/pages/detail/view.go:138](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/detail/view.go:138) `renderSourceView`
- 期望行为:
  1. 镜像详情在成功 inspect 后必须把原始字节(最好是 inspect API 返回的 JSON 或等价结构体 marshal)写入 `DetailRawJSON`,并设置 `DetailResourceType = ResourceImage`。
  2. 提供与容器/卷一致的 `SetImageDetail` 辅助函数。
  3. yaml / json 视图对镜像也要可用,渲染时支持 manifest list 与普通 image 两种来源。
- 验收标准:
  1. 镜像详情页按 s 后能看到完整 yaml 文本。
  2. yaml / json 视图在镜像为空数据时给出与卷/网络一致的错误占位,不再静默空白。
  3. `DetailRawJSON` / `DetailResourceType` 在镜像路径下与其他资源保持一致写入时序。

### BR-008 H 键应进 Help,F1 行为复用;H 键同时承担镜像 History 入口

- 状态: `open`
- 优先级: `medium`
- 症状:
  1. 在任何页面按 H 当前是 `Viewport.ToggleHeader()`(隐藏/显示 header),用户期望 H 直接进入 Help(与 F1 / `?` 一致)。
  2. 镜像详情页里 `History` 是一个分区,用户希望 History 提到镜像页顶层,**详情页不再显示 History**。
- 当前行为:
  `keys.KeyH` 在 `keyboard.go:215` 直接执行 `m.Viewport.ToggleHeader()`,没有走 `ActionHelp`;镜像详情中的 `buildImageDetailDataSections` 把 History 作为最后一段分区渲染。
- 代码锚点:
  [internal/tui/keyboard/keyboard.go:215](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/keyboard.go:215) `case keys.KeyH: m.Viewport.ToggleHeader()`
  [internal/tui/keyboard/actions.go:21](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/actions.go:21) `case keys.ActionHelp: ToHelp(m)`
  [internal/tui/ui/pages/detail/image.go:14](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/detail/image.go:14) 镜像分区构造
- 期望行为:
  1. H 键在所有模式下都映射到 `ActionHelp`,与 F1 / `?` 等价;`ToggleHeader` 行为改由其他键(或显式 `Ctrl+H`)承担(可选,若决定取消 Toggle 入口则记录)。
  2. 新增 `ModeHistory / PanelHistory`,以及 `ActionImageHistory`。在镜像页按下 H(或绑定到具体单字母键)进入 History 页,展示该镜像的 layer history(数据来源复用 `InspectImageDetail()` 的 history 输出)。
  3. 镜像详情页 `buildImageDetailDataSections` 移除 History 段,保留其它字段。
  4. Help 页 / Footer / 默认键位示例同步说明 H = Help、H 在镜像页 = History(单一键位根据上下文路由)。
- 验收标准:
  1. 全局按 H 进入 Help,与 F1 行为一致。
  2. 镜像页内 H 进入镜像 History(可与 Help 区分:Help 是全局提示,History 是当前镜像数据)。
  3. 镜像详情页不再渲染 History 分区;键位说明与帮助文案同步更新。

### BR-009 卷详情页面无法加载数据;按 enter 进入容器子视图后无法上下选择

- 状态: `open`
- 优先级: `high`
- 症状:
  1. 卷页面按 D 进入详情页后,长时间卡在加载占位或空白,看不到卷的 inspect 数据。
  2. 按 enter 进入卷的容器子视图(由 `Volumes.DetailName` 触发)后,显示容器列表但**无法用上下键选中行**。
- 当前行为:
  1. `doVolumeInspect` 同样是同步 inspect + `SetVolumeDetail` + `ToDetail`,在 podman no-cgo 模式下 socket I/O 阻塞事件循环;错误被 `RecordError` 静默,详情页停占位。
  2. `volumes.RenderList → renderContainers` 渲染容器子表时,**只用了 `cm.Items` 与 `rowLimit`,忽略了 `cm.Cursor` 与 `cm.ViewOffset`**,`RenderTable` 的 `Selected` / `Offset` 字段未传入;此外 `Volumes.DetailName` 进入子视图时 `m.Resources.Volumes.Cursor` 重置为 0,但渲染走的却是 `cm.Cursor`,从未联动,选中态无法显示也无法被上下键驱动。
- 代码锚点:
  [internal/tui/keyboard/volume_action.go:24](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/volume_action.go:24) `doVolumeInspect` (同步 inspect, 无 cmd)
  [internal/tui/ui/pages/volumes/view.go:139](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/volumes/view.go:139) `renderContainers`
  [internal/tui/keyboard/navigate.go:29](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/navigate.go:29) `ToVolumeDetail`
- 期望行为:
  1. 卷 inspect 走异步 `tea.Cmd` 路径,与 BR-006 容器详情保持一致;失败时给出明确错误占位。
  2. 卷 → 容器子视图要正确联动 `cm.Cursor` 与 `cm.ViewOffset`,渲染时传入 `Selected / Offset / Limit`,上下键能正常切换选中行。
  3. 进入子视图时仍把 `Volumes.Cursor` 归零以避免脏状态,但渲染与交互要基于 `cm`,行为参考容器列表自身的 cursor 模型。
  4. 在子视图里选中具体容器后,按 enter / l 等键应可进一步进入该容器的日志或详情(可选,但要避免目前“能进不能选”的死路)。
- 验收标准:
  1. podman no-cgo 启动后,卷 D 键加载可见,失败时显式错误。
  2. 卷 → 容器子视图中,上下键能逐行选中,Selected 行高亮。
  3. 在容器子视图里选中行后能正常退出回卷列表(BEsc 退到上一层,数据正确清理)。

### BR-010 网络详情页无法加载数据,无法获取 yaml / json 原始数据

- 状态: `open`
- 优先级: `high`
- 症状: 网络页面按 D 进入详情页后,长时间卡在加载占位或空白,看不到 inspect 数据;按 s 切换 yaml / json 也得不到原始数据。
- 当前行为:
  `doNetworkInspect` 同样在 `Update` 路径上**同步**调 `Engine.Networks().Inspect(ctx, id)`,与 BR-006 / BR-009 容器/卷详情同一根因 —— podman no-cgo 模式下 socket I/O 阻塞事件循环;错误被 `RecordError` 静默,详情页停占位。`SetNetworkDetail` 本身**确实**写入 `DetailRawJSON`,因此只要 inspect 成功,YAML/JSON 视图能拿到 raw,问题主要在同步阻塞而非 raw 缺失。
- 代码锚点:
  [internal/tui/keyboard/network_action.go:46](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/network_action.go:46) `doNetworkInspect` (同步 inspect, 无 cmd)
  [internal/tui/state/detail.go:72](/home/debi/IdeaProjects/docker-tui/internal/tui/state/detail.go:72) `SetNetworkDetail`
  [internal/tui/keyboard/delete_action.go:74](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/delete_action.go:74) `doDetailAction`
- 期望行为:
  1. 网络 inspect 走异步 `tea.Cmd` 路径,与 BR-006 / BR-009 容器/卷保持一致;失败时给出明确错误占位。
  2. 进入详情时重置 `DetailSourceType`,避免上一资源模式残留。
  3. 错误信息应能区分能力缺失(API 不支持)与运行时错误(socket / 404)。
- 验收标准:
  1. podman no-cgo 启动后,网络 D 键加载可见,失败时显式错误。
  2. yaml / json 视图对网络详情可显示完整原始 inspect 数据。
  3. inspect 错误时详情页给出明确错误占位,不再静默吞掉。

### BR-011 镜像页面超一页时光标下划页面不滚动

- 状态: `open`
- 优先级: `high`
- 症状: 镜像列表条目超过一页后,光标下移到第二页,渲染窗口未同步跟随,选中行看不见;再次按 enter / s 之前,视图位置与光标位置不一致。
- 当前行为:
  `images.RenderList` 第 49-50 行做了:
  ```go
  viewOffset := im.ViewOffset
  component.EnsureVisible(&viewOffset, im.Cursor, rowHeight, total)
  ```
  `EnsureVisible` 通过指针**只修改局部 `viewOffset`**,本次帧渲染 `Rows = items[viewOffset..viewOffset+rowHeight]` 因此窗口位置正确。但 **`im.ViewOffset` 永远没有被回写**,所有持久状态(滚动条 / footer offset 指示 / 键盘翻页 / 进入子视图后回到父视图)都依赖字段值;字段值不前进 → 跨页行为表面看"页面没跟随光标"。同时镜像 panel 与其它 panel 在 standard 模式下共用同一 rail,而 `panelHeight = rails.panel` 用的是整条 rail 的高度,**不是镜像 panel 自己的 bodyRows**,`CalcRowHeight(panelHeight)` 给出的 `rowHeight` 可能大于镜像专属可见行,造成视觉上"滚动过头但实际未动"。
- 代码锚点:
  [internal/tui/ui/pages/images/view.go:49](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/images/view.go:49) `viewOffset := im.ViewOffset` + `EnsureVisible(&viewOffset, ...)` (未回写)
  [internal/tui/ui/pages/images/view.go:89](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/images/view.go:89) 行循环起点使用局部 `viewOffset`
  [internal/tui/ui/component/viewport.go:8](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/viewport.go:8) `EnsureVisible` (只改指针,无回写协议)
  [internal/tui/ui/pages/containers/view.go:51](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/containers/view.go:51) 容器列表同模式 (未回写 `ViewOffset`)
  [internal/tui/ui/pages/volumes/view.go:76](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/volumes/view.go:76) 卷列表同模式 (未回写)
  [internal/tui/ui/pages/networks/view.go:73](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/networks/view.go:73) 网络列表同模式 (未回写)
- 期望行为:
  1. `EnsureVisible` 改成"返回新 offset 并由调用方显式回写",或在原地调用改为写入字段,使 `ViewOffset` 在每帧之后与渲染窗口一致;这是列表通用契约问题,容器/卷/网络同理修复。
  2. 镜像 panel 在 standard 模式下要走"按 panel 子区域计算 bodyRows"的逻辑,而非整条 rail 高度;`CalcRowHeight` 接收实际可见行数。
  3. 键盘翻页(PgUp / PgDn / ctrl+d / ctrl+u)按 `rowHeight` 跳 cursor,同时滚动 `ViewOffset`,行为一致。
- 验收标准:
  1. 镜像列表条目超过一页后,光标下移过窗口底部时,渲染窗口同步前进,选中行始终可见。
  2. `im.ViewOffset` 在每帧渲染后等于本次实际使用的 `viewOffset`;键盘翻页后状态正确。
  3. 容器/卷/网络列表同样修复,行为统一。
  4. 镜像 panel 在 standard 多 panel 布局下,滚动行数与该 panel 实际可见行数一致,不再受整条 rail 高度干扰。

### BR-012 所有表格鼠标点击选中的行位置不对

- 状态: `open`
- 优先级: `high`
- 症状: 在镜像 / 卷 / 网络列表上用鼠标点击某一行,选中行落在与点击位置相差一行或几行的位置,无法精准定位。
- 当前行为:
  `LayoutReport` 只持有**单条 panel rail** 的 `Panel.bodyTop / bodyRows`,不区分该 rail 在 standard 模式下是单列还是多列,也不区分 ActivePanel 在 6 个资源里是哪个。HitTest 用 `rep.Panel.bodyTop` 与 `y` 相减得到行号,该行号是相对整条 rail;`clickListCursor` 再用 `rep.Panel.bodyRows`(同样是整条 rail)做翻页边界估算。镜像 panel 在 standard 模式下只占整条 rail 的某一列(右侧窄列),`bodyTop / bodyRows` 都与镜像专属区域不一致,行号换算失真,因此点击命中行错位。
- 代码锚点:
  [internal/tui/ui/app/mouse.go:53](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/mouse.go:53) `LayoutReport.Panel` (单一 panelRect)
  [internal/tui/ui/app/mouse.go:65](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/mouse.go:65) `ResolveLayout`
  [internal/tui/ui/app/mouse.go:104](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/mouse.go:104) `resolveStandardLayout` (没有按 panel 分列几何)
  [internal/tui/ui/app/mouse.go:174](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/mouse.go:174) `HitTest` 用 `rep.Panel.bodyTop`
  [internal/tui/ui/app/mouse.go:254](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/mouse.go:254) `clickListCursor` 用 `rep.Panel.bodyRows`
- 期望行为:
  1. `LayoutReport` 需要按 ActivePanel(或按下点击列归属)分别给出每个 panel 子区域的几何,或者在 HitTest / clickListCursor 阶段根据 `m.Navigation.ActivePanel` 重新计算对应 panel 的 `bodyTop / bodyRows`。
  2. `clickListCursor` 用镜像 / 卷 / 网络 panel 自己的 bodyRows 做翻页估算,而不是整条 rail 高度。
  3. 鼠标点击坐标应能精准落在用户视觉上看到的行号上,误差 ≤ 0。
- 验收标准:
  1. 镜像 / 卷 / 网络列表用鼠标点击任意一行,选中行与点击行完全一致。
  2. standard 模式下多列布局(镜像在右列,容器在左列),点击右侧命中镜像专属区域、点击左侧命中容器专属区域,不串列。
  3. compact 模式下点击行为不变(单 panel)。

### BR-013 头部 Connection 列 Engine / Runtime 信息重复

- 状态: `verifying`
- 优先级: `medium`
- 症状: 头部第二列同时显示 `Engine <name>` 与 `Runtime <name>`,两份内容几乎一致,占用有限宽度、降低可读性。
- 当前行为:
  `header.go` 中 `eng` (118-130) 与 `runtimeName` (160-179) 都优先从 `app.Connection.Pool.ActiveName()` 取值,只是在二者都为空时分别降级到 `RuntimeType` / `ConnectionTarget`。连接正常时两者**一定相等**,最终在 `colConn` (180-185) 里被分别渲染成两行:
  ```
  Engine    local-podman
  Runtime   local-podman
  ```
  其中 `runtimeName` 还会附加 TLS 状态后缀,但与 `eng` 共享同一名称。这是早期两份独立信息源合并到统一 ConnectionState 之后没有清理的死代码。
- 代码锚点:
  [internal/tui/ui/widget/header/header.go:118](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/header/header.go:118) `eng` 计算
  [internal/tui/ui/widget/header/header.go:160](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/header/header.go:160) `runtimeName` 计算
  [internal/tui/ui/widget/header/header.go:180](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/header/header.go:180) `colConn` 渲染
- 期望行为:
  1. 头部 Connection 列只显示一份名称 + TLS 状态 + Engine 版本;不要出现 `Engine` / `Runtime` 双行重复。
  2. Engine 版本号 / socket / language 信息可保留,各自独立显示。
  3. 若要保留 "Engine / Runtime" 区分,需要给两个字段不同语义(例如 Engine = 协议层,Docker/Podman 二选一;Runtime = 当前激活端点 + 版本);但当前实现两者等价,合并即可。
- 验收标准:
  1. 头部不再出现相同名称的 Engine / Runtime 双行。
  2. TLS 状态后缀仍能正确显示。
  3. 头部列宽度不被无意义信息占据。
- 修复记录:
  - `header.go:160-179` 删去 `runtimeName` 冗余计算与 `colConn` 中的 `Runtime` 行;TLS 后缀改为追加到 `eng`,`colConn` 改为 3 行(`Engine` + `Socket` + `Language`),其他列宽与帮助文案已无需变动。

### BR-014 应用底部状态栏 / 操作日志在窄终端下中间部分被截断

- 状态: `open`
- 优先级: `medium`
- 症状: 终端宽度有限时,应用最下方的 status bar(包括连接状态、host、operation log)在一行内拼接,中间或尾部操作日志内容被裁掉,看上去"信息显示不全,到中间就截断"。
- 当前行为:
  footer `Render` (footer.go:56-67) 把 `Shortcuts`(2 行)+ `StatusBar`(1 行)合并为固定 3 行;第 64 行 `TruncateVisible(rows[i], width)` 把每行按**终端宽度**整体截断。`StatusBar` (footer.go:14-46) 把
  - `engineLabel`(`RuntimeType` 或 `ConnectionTarget`)
  - `hostStr`(`Engine.Identity().Endpoint`)
  - `operationLogStatus`(`ErrorMessage` / `AuditOperationMessage` / `InfoMessage`,oplog.go:29-53)
  用 `│` 连接,**全部挤在同一行**。当 `engineLabel + "│" + hostStr` 已经接近终端宽度时,`│ op` 后面的 op log 就会被尾部截断;若 `ErrorMessage` 很长,前面也可能溢出。`operationLogStatus` 自身又把 `ErrorMessage` 截到 40 字符,但仍然可能与其他部分叠加超出。整行只有一个 `│` 分隔,无法分行展示。
- 代码锚点:
  [internal/tui/ui/widget/footer/footer.go:14](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/footer/footer.go:14) `StatusBar`
  [internal/tui/ui/widget/footer/footer.go:56](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/footer/footer.go:56) `Render`
  [internal/tui/ui/widget/footer/oplog.go:29](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/footer/oplog.go:29) `operationLogStatus`
  [internal/tui/ui/app/rails.go:13](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/rails.go:13) `footerRailHeight = 3` (硬编码 3 行)
- 期望行为:
  1. `StatusBar` 应按信息优先级与可见宽度分段布局:连接状态(`● local-podman`)放固定左侧,`hostStr` 紧跟其后,op log / error 占右侧并按剩余宽度截断。
  2. 当操作日志过长时,要么展开成两行(突破 `footerRailHeight = 3` 限制),要么使用滚动 / 折叠;不能因为宽度问题直接砍掉关键错误信息。
  3. 终端宽度变窄时,优先级:连接状态 > host > op log;op log 应能完整显示至少尾部 `…` 提示截断,而非无声消失。
- 验收标准:
  1. 80-120 宽终端下,StatusBar 三段信息均可见或带截断提示。
  2. op log 出现新错误时,最近一条至少能完整看到(或带省略号)。
  3. `footerRailHeight` 应与内容解耦,不再硬编码 3。

### BR-015 表格选中行背景色未覆盖整行

- 状态: `verifying`
- 优先级: `medium`
- 症状: 表格(容器 / 镜像 / 卷 / 网络 / 审计)选中行的高亮背景色只覆盖到部分 cell,行末或某些列尾部能看到未着色的"尾巴",视觉上不像选中态。
- 当前行为:
  `RowRenderer.RenderRow` (rows.go:23-55) 对选中/标记行:
  ```go
  content := prefix + mark + line
  visLen := utils.VisibleLen(content)
  if r.rowWidth > visLen {
      content += strings.Repeat(" ", r.rowWidth-visLen)
  }
  ref := GetRowStyle("selected")
  return buildStyle(ref).Render(content)
  ```
  填充部分对 **空 cell 或 ANSI 残留 cell** 处理脆弱:`joinRow` 中 `useInline=true` 走 `cellStyleANSI`,只输出 `\033[38;2;...m` (前景) 而**不输出背景**也不**重置 SGR**;随后 `buildStyle(ref).Render(content)` 用 lipgloss 前缀包 `\033[48;2;bgm` 与末尾 `\033[0m`。整行理论上都被背景覆盖,但**当某个 cell 的 `display` 经 `utils.PadVisible(utils.TruncateVisible(cell, colW[i]), colW[i])` 后宽度刚好填满 cell** —— 此时 `visLen` 计算时若 cell 文本含 wide chars(中文 / box drawing)与 `runewidth.StringWidth` 计算口径不一致,会出现 `r.rowWidth < visLen`,**填充分支被跳过**,行末就裸露;反之,若 cell 的 ANSI 被错误累加 / `rowPrefix` 宽度算错,也会让 padding 偏少。同时 lipgloss 在嵌套 ANSI 时对 background 的解析可能截断一段 background-active 范围,使尾段被外部 reset。
- 代码锚点:
  [internal/tui/ui/component/rows.go:23](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/rows.go:23) `RenderRow` 填充逻辑
  [internal/tui/ui/component/rows.go:103](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/rows.go:103) `cellStyleANSI` (不闭合 SGR)
  [internal/tui/ui/component/rows.go:121](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/rows.go:121) `RenderHeader`
  [internal/tui/ui/component/table.go:114](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/table.go:114) `RowRenderer` 构造
  [internal/tui/ui/component/table.jsonc:19](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/table.jsonc:19) `rowStyles.selected.background = "#37676f"`
- 期望行为:
  1. 选中行整行宽度(从 `rowPrefixSel` 到行末右边界)都被 `rowStyles.selected.background` 覆盖。
  2. wide chars / 中英文混排 / ANSI 转义残留 / 列 padding 错算等情形都不能让背景"穿帮"。
  3. 至少在 80-200 宽终端、宽字符 / box-drawing 字符混排、宽列宽窄列宽同时出现时,选中行视觉完整。
- 验收标准:
  1. 选中行背景从行首到行末连续、无裸露。
  2. 列内文本过长被截断 / 含 wide chars / 含 ANSI 颜色时,选中行背景仍连续。
  3. `marked` 与 `selected` 背景同时存在时(同一行既是当前选中又被标记),高亮一致或可区分,且不出现背景色块被切的现象。
- 修复记录:
  - `rows.go:23-55` 重写 `RenderRow` 的选中/标记分支:不再用 `cellStyleANSI` 内嵌前景 ANSI + 手填空格,改为 `joinRow(cells, false)` 走 lipgloss 完整 style(自带 SGR reset),再用 `buildStyle(ref).Width(r.rowWidth).Render(prefix + line)`。`Width` 让 lipgloss 自动填满行尾空格,`Background` 由 `ref` 提供,跨列 SGR 不再互相截断背景。

## 新增条目模板

```md
### BR-xxx 问题标题

- 状态: `open`
- 优先级: `high|medium|low`
- 症状:
- 当前行为:
- 代码锚点:
- 期望行为:
- 验收标准:
  1.
  2.
```
