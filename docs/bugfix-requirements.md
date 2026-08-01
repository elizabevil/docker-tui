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

- 状态: `done`
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
- 修复记录:
  - `doInspectAction` 已改为先切入详情页再异步发起 `tea.Cmd` inspect,避免 `Update` 阶段阻塞 socket I/O。
  - 详情结果回调会把成功 inspect 写回 `DetailState`,失败则直接在详情页显示 `Inspect failed: ...` 错误占位。
  - `SetContainerDetail` 继续负责写入结构化 raw JSON,源视图可以正常切换。

### BR-007 镜像详情页不显示 yaml / json 原始数据

- 状态: `done`
- 优先级: `high`
- 症状: 镜像页面按 D 进入详情页后,只能看到 `image.go` 渲染的分区内容;页面上下滚动时仍有轻微卡顿,按 s 切到 yaml / json 视图时输出空白,看不到原始 inspect 数据。
- 当前行为:
  `doImageDetail` 调用 `m.Detail.OpenImage(id, title, data)`。`OpenImage` 只设置了 `ImageDetailID / DetailTitle / ImageDetailData`,**没有把 inspect 原始字节写入 `DetailRawJSON`**。`renderSourceView` 直接对 `DetailRawJSON` 做 `sonic.Unmarshal`,空字节切片 unmarshal 失败,`text` 落到空字符串兜底,只渲染出滚动 footer。这是与容器/卷/网络的 `SetContainerDetail / SetVolumeDetail` 行为不一致造成的缺失。detail 渲染还会在可见区变化时重建整段 section,滚动时体感会比列表更重。
- 代码锚点:
  [internal/tui/state/detail.go:58](/home/debi/IdeaProjects/docker-tui/internal/tui/state/detail.go:58) `OpenImage` (未写 `DetailRawJSON / DetailResourceType`)
  [internal/tui/keyboard/image_action.go:99](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/image_action.go:99) `doImageDetail`
  [internal/tui/ui/pages/detail/view.go:133](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/detail/view.go:133) `renderSections`
  [internal/tui/ui/pages/detail/view.go:167](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/detail/view.go:167) `renderSourceView`
- 期望行为:
  1. 镜像详情在成功 inspect 后必须把原始字节(最好是 inspect API 返回的 JSON 或等价结构体 marshal)写入 `DetailRawJSON`,并设置 `DetailResourceType = ResourceImage`。
  2. 提供与容器/卷一致的 `SetImageDetail` 辅助函数。
  3. yaml / json 视图对镜像也要可用,渲染时支持 manifest list 与普通 image 两种来源。
- 验收标准:
  1. 镜像详情页按 s 后能看到完整 yaml 文本。
  2. yaml / json 视图在镜像为空数据时给出与卷/网络一致的错误占位,不再静默空白。
  3. `DetailRawJSON` / `DetailResourceType` 在镜像路径下与其他资源保持一致写入时序。
- 修复记录:
  - `OpenImage` / `ApplyImage` 已写入规范化 `ImageDetail` JSON，并设置镜像资源类型。
  - section、YAML、JSON 按详情 revision 分别缓存；滚动只截取和渲染可见行。
  - raw source 缺失或不可解析时显示明确占位，不再输出空白页。

### BR-016 容器详情页上下滚动卡顿

- 状态: `done`
- 优先级: `high`
- 症状: 容器详情页在上下滚动时有明显卡顿,重复多次后更明显。
- 当前行为:
  详情页滚动状态没有稳定回写到可见边界,滚动到底部后继续下滚会积累不可见偏移,再向上滚时要先“还债”才能看到变化。当前渲染路径还会在每次刷新时重新构造完整内容,放大了卡顿感。
- 代码锚点:
  [internal/tui/state/detail.go:68](/home/debi/IdeaProjects/docker-tui/internal/tui/state/detail.go:68) `Scroll`
  [internal/tui/state/detail.go:120](/home/debi/IdeaProjects/docker-tui/internal/tui/state/detail.go:120) `ClampVisibleOffset`
  [internal/tui/ui/pages/detail/view.go:100](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/detail/view.go:100) `renderSections`
  [internal/tui/keyboard/detail.go:13](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/detail.go:13) `handleDetailKeys`
- 期望行为:
  1. 容器详情滚动状态应回写到可见边界,不能积累不可见 overscroll。
  2. 向上 / 向下滚动在底部与顶部都应即时生效,不需要先回滚“欠账”。
  3. 结构化详情视图不应在每次滚动时重建不必要的数据结构。
- 验收标准:
  1. 上下滚动容器详情页 10 次以上后,继续反向滚动能立即看到位置变化。
  2. 连续滚动不会出现“看起来卡住了几步”的现象。
  3. 详情页 footer 的可见行号与当前内容范围一致。
- 修复记录:
  - 详情 offset 会回写到当前文档的合法可见边界，反向滚动不再偿还不可见 overscroll。
  - 结构化详情和源码视图使用 revision/source 文档缓存，滚动不再重新构造完整 section 或格式化 YAML/JSON。
  - Toast timer 改为按 generation 启动和终止，空闲页面不再由永久 100ms tick 触发整屏重绘。
  - 500 个环境变量和 500 个标签的缓存滚动 benchmark 已加入回归测试。

### BR-017 Compose 详情页不应把快捷键写进正文

- 状态: `done`
- 优先级: `medium`
- 症状: Compose 页面左侧 project 栏按 `d` 进入详情页后,详情正文里显示快捷操作说明;按 `s` 之后页面内容变空,用户以为详情丢了数据。
- 当前行为:
  `BuildProjectDetail()` 直接把“快捷操作”写进项目详情正文,详情页自身又会切换 `s` 到 YAML / JSON 源码视图,导致正文与底部快捷提示混在一起。Compose 是双栏布局,左栏和右栏的可用快捷键不同,固定正文里的单套提示会误导用户。
- 代码锚点:
  [internal/tui/ui/pages/compose/view.go:304](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/compose/view.go:304) `BuildProjectDetail`
  [internal/tui/ui/action/registry.go:85](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/action/registry.go:85) `Context`
  [internal/tui/ui/action/registry.go:98](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/action/registry.go:98) `Compose`
  [internal/tui/ui/widget/footer/footer.go:12](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/footer/footer.go:12) `Shortcuts`
- 期望行为:
  1. Compose 详情页正文只显示项目概览数据,不再写快捷键说明。
  2. 快捷键说明统一放到下方状态栏 / footer。
  3. footer 在 compose 场景下要按当前焦点输出对应快捷键,左栏项目与右栏服务不能共用一套固定提示。
- 验收标准:
  1. 按 `d` 进入 compose 详情页后,正文不再出现“快捷操作”段落。
  2. 按 `s` 后显示的是合法源码视图或空值占位,不是被正文快捷说明污染后的空白页。
  3. footer 的快捷键随 compose 左右栏焦点切换而变化。
- 修复记录:
  - Compose 项目详情正文已移除快捷操作段落。
  - footer 在普通 Compose 双栏模式下按 `ComposeFocus` 投影项目或服务动作。
  - 详情 footer 只在 raw source 存在时显示 `s Source`；没有 raw source 的 Compose 项目详情按 `s` 保持 section 视图。
  - 详情标题统一为 `Compose Detail: <project>`，breadcrumb 使用与镜像、容器详情相同的 `Compose > detail` 结构。
  - 项目正文不再使用独立的 `Project: <name>` 文本格式。
  - Compose 服务栏标题统一为 `Compose > <project> > Services`，不再使用 `Services: <project>`。
  - Compose 项目详情改为只读表格，按 `Service / Image / Status / Containers` 展示并复用详情滚动状态。
  - 表格列宽统一由 `tables.ResolveLayout` 计算；页面不再预计算或传递 `Widths`，`RenderTable` 不再按当前行内容二次收缩列宽。

### BR-008 H 键应进 Help,F1 行为复用;H 键同时承担镜像 History 入口

- 状态: `open`
- 优先级: `medium`
- 症状:
  1. 在任何页面按 H 当前是 `Viewport.ToggleHeader()`(隐藏/显示 header),用户期望 H 直接进入 Help(与 F1 / `?` 一致)。
  2. 镜像详情页里 `History` 曾是一个分区,用户希望 History 提到镜像页顶层,**详情页不再显示 History**。
- 当前行为:
  `keys.KeyH` 在 `keyboard.go:215` 直接执行 `m.Viewport.ToggleHeader()`,没有走 `ActionHelp`;镜像详情中的 `buildImageDetailDataSections` 和文本回退路径已不再渲染 History 段,Docker / Podman 的详情 inspect 也暂时不再请求 History endpoint。详情滚动会回写合法偏移且只渲染当前可见行,避免底部越界偏移累积和全量样式重绘。
- 代码锚点:
  [internal/tui/keyboard/keyboard.go:215](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/keyboard.go:215) `case keys.KeyH: m.Viewport.ToggleHeader()`
  [internal/tui/keyboard/actions.go:21](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/actions.go:21) `case keys.ActionHelp: ToHelp(m)`
  [internal/tui/ui/pages/detail/image.go:14](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/detail/image.go:14) 镜像分区构造
- 期望行为:
  1. H 键在所有模式下都映射到 `ActionHelp`,与 F1 / `?` 等价;`ToggleHeader` 行为改由其他键(或显式 `Ctrl+H`)承担(可选,若决定取消 Toggle 入口则记录)。
  2. 新增 `ModeHistory / PanelHistory`,以及 `ActionImageHistory`。在镜像页按下 H(或绑定到具体单字母键)进入 History 页,展示该镜像的 layer history(数据来源复用 `InspectImageDetail()` 的 history 输出)。
  3. 镜像详情页已移除 History 段,仅保留其它字段。
  4. Help 页 / Footer / 默认键位示例同步说明 H = Help、H 在镜像页 = History(单一键位根据上下文路由)。
- 验收标准:
  1. 全局按 H 进入 Help,与 F1 行为一致。
  2. 镜像页内 H 进入镜像 History(可与 Help 区分:Help 是全局提示,History 是当前镜像数据)。
  3. 镜像详情页不再渲染 History 分区;键位说明与帮助文案同步更新。

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

- 状态: `done`
- 优先级: `high`
- 症状: 网络页面按 D 进入详情页后,长时间卡在加载占位或空白,看不到 inspect 数据;按 s 切换 yaml / json 也得不到原始数据。
- 当前行为:
  `doNetworkInspect` 已切换为异步 `tea.Cmd`，不再在 `Update` 路径上直接阻塞 `Engine.Networks().Inspect(ctx, id)`；`SetNetworkDetail` 仍负责写入 `DetailRawJSON`，因此只要 inspect 成功，YAML/JSON 视图就能拿到 raw 数据。
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
- 修复记录:
  - 网络详情加载已和容器 / 卷一致改成异步详情消息流。
  - 失败态会在详情页直接显示错误内容,不再依赖后台 toast。

### BR-011 镜像页面超一页时光标下划页面不滚动

- 状态: `done`
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
- 修复记录:
  - 新增表格真实数据行计算，扣除选中预览、页码、空行、表头、footer 与行间距。
  - 镜像列表直接将 `EnsureVisible` 结果回写 `ImageListModel.ViewOffset`，不再只修改当前渲染帧的局部变量。
  - 已覆盖 30 个镜像在 12 行 panel 中向下跨页和向上回滚的测试；Network 已迁移为直接回写 `NetworkListModel.ViewOffset`，容器与卷仍按后续独立问题推进。

### BR-012 所有表格鼠标点击选中的行位置不对

- 状态: `wontfix`
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
- 不修复原因:
  - 用户在 [BR-030](#br-030-取消鼠标点击表格行选中改用滚轮上下选行) 中明确决定**取消鼠标点击行选中**,改为只用滚轮上下选行。`LayoutReport` 按 panel 分列几何这一修复已不再需要。
  - `clickListCursor` 与 `HitTest` 中与点击相关的分支可以移除,降低 standard 模式下多列布局的几何计算复杂度。
- 替代方案: 见 BR-030。

### BR-013 头部 Connection 列 Engine / Runtime 信息重复

- 状态: `done`
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

- 状态: `done`
- 优先级: `medium`
- 症状: 终端宽度有限时,应用最下方的 status bar(包括连接状态、host、operation log)在一行内拼接,中间或尾部操作日志内容被裁掉,看上去"信息显示不全,到中间就截断"。
- 当前行为:
  footer 现在只保留两行快捷键提示;统一提示信息上移到 header 下方的两行 message rail,错误 / 操作 / 提示不再挤在底部一行中。
- 代码锚点:
  [internal/tui/ui/widget/footer/footer.go:14](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/footer/footer.go:14) `StatusBar`
  [internal/tui/ui/widget/footer/footer.go:56](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/footer/footer.go:56) `Render`
  [internal/tui/ui/widget/footer/oplog.go:29](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/footer/oplog.go:29) `operationLogStatus`
  [internal/tui/ui/app/rails.go:13](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/rails.go:13) `footerRailHeight = 2` (硬编码 2 行)
- 期望行为:
  1. `StatusBar` 应按信息优先级与可见宽度分段布局:连接状态(`● local-podman`)放固定左侧,`hostStr` 紧跟其后,op log / error 占右侧并按剩余宽度截断。
  2. 当操作日志过长时,要么展开成两行(突破 `footerRailHeight = 2` 限制),要么使用滚动 / 折叠;不能因为宽度问题直接砍掉关键错误信息。
  3. 终端宽度变窄时,优先级:连接状态 > host > op log;op log 应能完整显示至少尾部 `…` 提示截断,而非无声消失。
- 验收标准:
  1. 80-120 宽终端下,统一提示在 header 下方可见且不会被 footer 截断。
  2. 错误信息不再依赖底部状态栏展示。
  3. `footerRailHeight` 已不再承载状态栏内容。
- 修复记录:
  - 底部 footer 已缩减为两行快捷键提示。
  - `ErrorMessage` / `AuditOperationMessage` / `InfoMessage` 统一上移到 header 与 panel 之间的两行消息 rail。
  - 长错误按终端可见宽度换行;超过两行时在第二行显示省略标记。

### BR-015 表格选中行背景色未覆盖整行

- 状态: `done`
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

### BR-018 Network 与 Audit 表格列布局不一致

- 状态: `done`
- 优先级: `medium`
- 症状: Network 在部分宽度下名称列被限制后仍出现异常留白；Audit 使用手工固定列宽和字符串截断，未遵循其他资源页的共享表格、右上角统计和整行选中样式。
- 期望行为:
  1. 两页都使用共享 `RenderTable` 与 `ResolveLayout`，表头和数据行使用相同列轨道。
  2. Network 名称与子网列使用稳定首选宽度，空间不足时收缩，空间充足时填满表格 viewport。
  3. Audit 的时间、结果、级别保持固定宽度，操作、目标、消息使用稳定内容宽度；统计和筛选信息统一显示在右上角。
  4. 两页滚动后都将实际 offset 回写页面状态，选中项始终在可见窗口内。
- 修复记录:
  - 新增 `tables/audit.jsonc`，Audit 列表删除手工 header/row 拼接并迁移到共享表格渲染。
  - 调整 `tables/networks.jsonc` 的名称列上限；Network 直接回写 `NetworkListModel.ViewOffset`。
  - 增加宽窄终端行宽和多页滚动回归测试。

### BR-019 表格列布局未遵循稳定 Grid 约束

- 状态: `done`
- 优先级: `high`
- 症状: 页面重复扣减 content width，列宽从 0 分配且所有列封顶后将剩余空间强制塞入最后一个 Flex 列；右侧空间、列比例和列间距的视觉结果不稳定。
- 修复记录:
  - 初版为 `ColumnDef` 增加 `basis/min/max/grow/shrink`，按 CSS Flex 语义执行 grow/shrink，删除突破 `max` 的 spillover。
  - 列 gap 固定为 2 个终端单元，并作为布局缓存 key 的一部分。
  - 初版将全部表格 profile 迁移到 grow/fill 语义。
  - 顶层资源页直接使用 panel content width，不再重复扣减 8 列；Compose 和 Processes 不再重复扣减 4 列。
  - 补齐镜像和卷子表的 `BannerW`，确保子表与主表使用同一布局契约。
  - 后续宽屏回归确认单一 grow/fill 列会把剩余宽度变成中间单元格的大段空白，因此移除页面级 `grow` 权重，改由共享解析器管理扩展。
  - 最终规则为：先按可见内容需求扩展被截断列，再仅由显式 `fill` 主列按权重分配剩余宽度；固定列不扩展，稀疏列不吸收无意义空白，末列右边界与 viewport 对齐。
  - `ResolveContentLayout` 显式接收每列内容需求；增加长镜像名回归测试，验证宽屏下 Registry、Tag 完整显示、ID 保持 12 位且不存在未使用的行尾空间。
  - Images 平台列统一为 `OS/ARCH`；Network 增加基础 short ID；Containers 使用紧凑 IMAGE，PORTS/IP 仅按真实内容扩展，MOUNTS/CONTAINERS 计数表头不再被截断。
  - 配置来源收敛为两层：`component/table.jsonc` 唯一管理公共布局与视觉，`tables/*.jsonc` 管理资源数据 schema/profile；删除旧 `component/config.jsonc`、`column_widths.go` 和 `tables/_global.jsonc`。

### BR-020 容器端口与详情长字段、源码复制不完整

- 状态: `done`
- 优先级: `high`
- 症状: Podman 容器列表忽略未发布的暴露端口；镜像 Registry 超过详情面板宽度后被裁掉；YAML/JSON 源码页无法使用 `Ctrl+A`、`Ctrl+C` 完整复制。
- 修复记录:
  - Podman 列表映射合并 `ExposedPorts` 和已发布端口，并按容器端口与协议去重；未发布端口以 `6379/tcp` 显示。
  - 详情渲染接收实际面板宽度，按终端单元宽度折行并按宽度缓存文档，Registry 等长字段不再用省略号截断。
  - YAML/JSON 源码视图支持 `Ctrl+A` 全选高亮、`Ctrl+C` 复制完整格式化源码；Linux、macOS、Windows 分别复用可用的系统剪贴板命令。

### BR-021 资源操作失败未进入统一错误提示区

- 状态: `done`
- 优先级: `high`
- 症状: 容器启动失败只在底部操作日志显示 `resource.container.start: ✕ failed: started`，Header 下方错误提示区为空，且实际运行时错误原因丢失。
- 修复记录:
  - 容器、镜像和通用资源 action handler 在失败时统一调用 `Feedback.RecordError`。
  - 错误提示包含动作和运行时返回的具体原因，持久显示在 Header 下方两行提示区并可用 `Esc` 清除。

### BR-022 确认窗口默认选项与容器批量停止

- 状态: `done`
- 优先级: `medium`
- 症状: 确认窗口默认聚焦确认操作，批量删除选项混合了多个确认枚举；容器批量停止没有独立的强制停止选择。
- 修复记录:
  - 所有确认窗口统一为左侧 `Cancel` 默认，确认操作位于右侧。
  - 批量删除使用 `Cancel / Force`；网络删除使用 `Cancel / Delete`。
  - 容器页面标记多个容器后使用批量停止快捷键，确认窗口提供 `Cancel / Force`，Force 通过 kill 立即停止容器。

### BR-022 首次布局尺寸不稳定且宽屏表格列过度拉伸

- 状态: `done`
- 优先级: `high`
- 症状: 应用首次显示时 Header/Footer 可能被裁切，刷新后恢复；超宽窗口把剩余空间全部分配给 NAME、SUBNET 等 `fill` 列，形成大段单元格空白。
- 修复记录:
  - 应用初始化主动发送 `RequestWindowSize`，确保首屏使用终端当前尺寸计算全部 rail。
  - 内容感知表格只扩展存在截断需求的列，不再把无内容的剩余宽度灌入单元格；剩余宽度均匀分配到列间 gutter，使最后一列对齐右边界。
  - 普通行、斑马线、选中行和标记行继续按 viewport 宽度完整铺底，列轨道与整行视觉边界解耦。
  - 删除负收益的 `TableLayoutCache`：缓存 key 仍需扫描并哈希全部可见单元格；改为直接解析布局并预计算 gutter 字符串。
  - 终端宽度测量改为无分配 ANSI 扫描；50 行基准由约 `7.8 µs/156 allocs` 降至约 `5.8 µs/5 allocs`。

### BR-023 F2 唯一提供 runtime 连接选择器；其它页面不得用 C 刷新 Conn

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
  [internal/tui/keyboard/runtime_selector.go:13](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/runtime_selector.go:13) `openRuntimeSelector`
  [internal/tui/ui/app/layout.go:293](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/layout.go:293) `renderRuntimeSelector`
  [internal/tui/keys/registry.go:79](/home/debi/IdeaProjects/docker-tui/internal/tui/keys/registry.go:79) `ActionVolumeCreate` ← `KeyC`
  [internal/tui/keys/registry.go:82](/home/debi/IdeaProjects/docker-tui/internal/tui/keys/registry.go:82) `ActionNetworkCreate` ← `KeyC`
  [internal/tui/keys/registry.go:85](/home/debi/IdeaProjects/docker-tui/internal/tui/keys/registry.go:85) `ActionRefreshConnections` ← `KeyF12, KeyR`
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

### BR-024 详情源码视图 Ctrl+C 复制不完整 / 一次性失效

- 状态: `open`
- 优先级: `medium`
- 症状:
  - YAML / JSON 源码视图下,`Ctrl+A` 可以"全选"高亮,但 `Ctrl+C` 后续复制不可重复执行或复制内容不完整,无法满足"多段、多次、批量"的复制需求。
- 当前行为:
  - `keyboard/detail.go:41-52` 中 `Ctrl+A` 在 `DetailSourceType != Section` 且 `HasRawSource()` 时把 `SourceSelected = true`;`Ctrl+C` 仅在 `SourceSelected == true` 时复制 `SourceText()`(完整格式化 YAML / JSON)到剪贴板。
  - `SourceSelected` 设置后**从不重置**,但应用没有给用户可见的"已选中"高亮区分,且 `Ctrl+C` 缺乏撤销 / 重选控制;若用户在终端里尝试鼠标拖选或多次 `Ctrl+C`,界面没有任何提示当前到底是"全选态"还是"待重新选择"。
- 代码锚点:
  [internal/tui/keyboard/detail.go:41](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/detail.go:41) `Ctrl+A` 分支
  [internal/tui/keyboard/detail.go:45](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/detail.go:45) `Ctrl+C` 分支
  [internal/tui/state/detail.go:57](/home/debi/IdeaProjects/docker-tui/internal/tui/state/detail.go:57) `SourceSelected` 字段
- 期望行为:
  1. `Ctrl+C` 应当能从源码视图复制**完整**格式化文本(`SourceText()` 全量输出),支持多次连续执行。
  2. 应用应在源码视图给出"已选中 / 已复制"的可视提示,例如选中行有可视高亮、复制后 toast 显示已复制字节数。
  3. 复制行为不应阻止终端自身的鼠标拖选 / 系统剪贴板路径;当应用处于源码视图时,用户可以用 `Ctrl+C` 触发应用复制,也可以用终端自身选区复制。
- 验收标准:
  1. 在 YAML / JSON 视图,`Ctrl+A` 后多次 `Ctrl+C` 都能成功复制,toast 反馈每次复制字节数。
  2. 复制内容长度等于 `SourceText()` 完整输出,不出现截断 / 单行。
  3. 切换回 section 视图后,`SourceSelected` 重置为 false,避免误复制。

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
  [internal/tui/keyboard/container_action.go:409](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/container_action.go:409) `title := i18n.T("detail.title.container", ctr.Name, ctr.ID)`
  [internal/data/i18n/lang.go:80](/home/debi/IdeaProjects/docker-tui/internal/data/i18n/lang.go:80) `T(key, args...)`
  [internal/data/i18n/lang/en.jsonc:275](/home/debi/IdeaProjects/docker-tui/internal/data/i18n/lang/en.jsonc:275) `"detail.title.container": "Container Detail: {0} ({1})"`
  [internal/tui/ui/app/page_templates.go:73](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/page_templates.go:73) `view.title = m.Detail.DetailTitle`
- 期望行为:
  容器详情页标题正确替换为 `Container Detail: <name> (<id>)`;同样适用于 `detail.title.container_yaml` / `detail.title.container_json`(切到 YAML / JSON 视图时也要正确替换);`detail.title.image*` 等其它含占位符的标题一并按相同规则处理。
- 验收标准:
  1. 进入任一容器详情页,标题栏显示完整替换后的文本,不出现字面 `{0}` `{1}`。
  2. 切到 YAML / JSON 视图时标题正确替换。
  3. 三种语言(en/zh/ja)都要正确替换。

### BR-026 详情页快捷键集合需要明确:支持 j/k/PgUP/PgDn/Filter,不支持 Space/R

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 详情页当前把 `Space` 当作 PageDown(向下滚 20 行),用户期望**不提供 Space 滚动**;`R` 在详情页当前不绑定,但用户要求**明确不提供 R 刷新**,且 Help / Footer 必须按页面对应的快捷键集合展示。
- 当前行为:
  - `keyboard/detail.go:33` `case keys.KeySpace, keys.KeyPgDn: m.Detail.Scroll(20)` —— `Space` 与 `PgDn` 共用 20 行滚动分支。
  - `R` 在 `ModeDetail` 下不绑定任何动作(`handleDetailKeys` 未匹配 `KeyR` 时落入 `return true, nil`,实际吞掉按键但不报错),但 Help / Footer 不会主动标出 "R 在此处不可用"。
- 代码锚点:
  [internal/tui/keyboard/detail.go:33](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/detail.go:33) `case keys.KeySpace, keys.KeyPgDn`
  [internal/tui/ui/action/registry.go:235](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/action/registry.go:235) Detail footer/shortcuts 投影
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

### BR-027 容器日志页滚到底后向上划卡顿

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 容器日志页用鼠标滚轮滑到最下面,再向上划会卡一下。多次正反向滚动后卡顿更明显。
- 当前行为:
  - 与 BR-016(详情页滚动卡顿)同类根因:`m.Log.Scroll(delta)` 只增减 offset,不立即把 offset 钳制到合法可见范围;反向滚动时 offset 先回到 0 再开始正向,体感是先"还债"再移动。
  - 长日志下 `ui/pages/logs/view.go` 在每帧重建完整可见行,放大卡顿。
- 代码锚点:
  [internal/tui/state/log.go](/home/debi/IdeaProjects/docker-tui/internal/tui/state/log.go) `Scroll` / `VisibleOffset`
  [internal/tui/ui/pages/logs/view.go:79](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/logs/view.go:79) 渲染循环起点使用 `m.Log.VisibleOffset(...)`
  [internal/tui/keyboard/log.go:22](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/log.go:22) `m.Log.Scroll(±1)`
  [internal/tui/keyboard/log.go:31](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/log.go:31) `Scroll(±20)`
- 期望行为:
  1. 日志页滚动状态回写到当前可见边界,不能积累不可见 overscroll。
  2. 向上 / 向下滚动在底部与顶部都应即时生效,不需要先回滚"欠账"。
  3. 长日志(`>=1000` 行)下连续正反向滚动 10+ 次,体感无卡顿;每次渲染只截取可见行,不重建完整内容。
- 验收标准:
  1. 日志页滚到底后立即向上滚 1 行,视图即时变化。
  2. 连续滚动 10 次后,反向滚动立即响应。
  3. 1000+ 行日志下反向滚动无卡顿。

### BR-029 筛选框 Enter/Esc 双层退出:Enter 应用并保留框,首次 Esc 回退到框,二次 Esc 退出筛选模式

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 在资源页按 `/` 进入筛选模式后,按 Enter 当前会触发"应用 + 退出"语义,但用户期望 **Enter 仅应用过滤,筛选框保留可见**。
  - 在资源页筛选模式中按 Esc 当前直接退出筛选模式 (`actions.go:138-140` 直接调用 `BackFromFilter`),用户期望 **首次 Esc 仅把焦点退回到筛选输入框**(`ModeFilter` 不退出,只是焦点复位 / 输入框继续可见可编辑),**第二次 Esc 才退出筛选模式**。
- 当前行为:
  - `keyboard/actions.go:25-26`: `ActionFilter` 调用 `ToFilter(m)` 进入筛选模式。
  - `keyboard/actions.go:138-140`: 在 `ModeFilter` 下按 Esc,直接调用 `BackFromFilter(m)` 退出整个筛选模式,单 Esc 即退出。
  - 资源列表过滤已实现"输入即时过滤"(BR-004 修复记录),与 Enter 应用语义等价,因此 Enter 不必"额外触发退出"。
- 代码锚点:
  [internal/tui/keyboard/actions.go:25](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/actions.go:25) `ActionFilter` 入口
  [internal/tui/keyboard/actions.go:138](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/actions.go:138) `ModeFilter` 下 Esc 单次退出
  [internal/tui/state/filter.go](/home/debi/IdeaProjects/docker-tui/internal/tui/state/filter.go) (或对应 model) `ToFilter` / `BackFromFilter` / 焦点状态
- 期望行为:
  1. **筛选模式下的按键语义明确分层**:
     - `Enter`:把当前输入应用到过滤(`FilterInput` 已自动实时生效时,Enter 仅"确认提交",不退出模式);筛选输入框**继续可见**,用户可继续编辑。
     - 首次 `Esc`:把焦点从其它位置退回到筛选输入框(若焦点本就在输入框内则 no-op),**不退出 ModeFilter**,输入框继续可见可编辑。
     - 第二次 `Esc` (在首次 Esc 之后,焦点已回到输入框):退出整个筛选模式,清空输入并恢复 ModeNormal。
  2. 资源过滤仍然"输入即时过滤"(沿用 BR-004 行为),Enter 与 Esc 不参与实时过滤逻辑,只控制焦点与模式。
  3. 与日志页的 `Search`(`ModeSearch`)保持一致的双 Esc 退出语义(若适用)。
- 验收标准:
  1. 在容器 / 镜像 / 卷 / 网络任一页面按 `/`,筛选框出现并获得焦点。
  2. 输入过滤词后立即过滤;按 Enter,过滤仍然生效,筛选框继续可见,焦点保留在输入框。
  3. 按 Esc:若焦点不在输入框,焦点回到输入框;再按 Esc,ModeFilter 退出,输入框消失,ModeNormal 恢复。
  4. 在筛选输入框中按 Esc 同样计数为"已回到输入框"的一次,下一次 Esc 退出整个筛选模式。

### BR-030 取消鼠标点击表格行选中,改用滚轮上下选行

- 状态: `open`
- 优先级: `high`
- 症状:
  - 镜像 / 卷 / 网络 / 容器 / 审计列表上,鼠标点击行尝试选中时位置错位(BR-012)。
  - 用户决定**取消鼠标点击行选中**,改为只用鼠标**滚轮**驱动上下选行;终端鼠标键盘边界因此更清晰,无需处理 standard 模式下多列布局的几何对齐。
- 当前行为:
  - `internal/tui/ui/app/mouse.go:174` `HitTest` 把鼠标 `(x,y)` 映射到 panel 行号;`mouse.go:254` `clickListCursor` 写入选中行;镜像 panel 在 standard 模式下与整条 rail 几何不一致,导致点击命中错位。
  - 滚轮已经在日志/详情页支持,但**资源列表页**未把滚轮事件路由到 cursor 上下移动。
- 代码锚点:
  [internal/tui/ui/app/mouse.go:174](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/mouse.go:174) `HitTest`
  [internal/tui/ui/app/mouse.go:223](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/mouse.go:223) `clickListCursor` 调用点
  [internal/tui/ui/app/mouse.go:254](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/mouse.go:254) `clickListCursor`
  [internal/tui/update/update_events.go](/home/debi/IdeaProjects/docker-tui/internal/tui/update/update_events.go) (或等价) 鼠标滚轮事件分发
- 期望行为:
  1. **取消**鼠标点击行选中:BR-012 已 wontfix,`clickListCursor` 路径移除或退化为 no-op;HitTest 对 panel body 的点击不再触发 cursor 移动。
  2. **新增**滚轮上下选行:在资源列表(容器 / 镜像 / 卷 / 网络 / 审计)的 panel body 上,鼠标滚轮向下滚动一格 → cursor 向下 1 行,向上滚动一格 → cursor 向上 1 行;与键盘 `j` / `k` 行为完全一致,同步刷新 `ViewOffset`。
  3. 鼠标滚轮的事件分发应该由 `ActivePanel` 决定路由,不必再依赖 panel 的 body geometry 计算。
  4. 终端原生复制 / 选区行为不被应用拦截;只在 panel body 区域生效,panel 之外(头部 / 状态栏 / footer)不抢滚轮事件。
- 验收标准:
  1. 资源列表的鼠标点击不再移动 cursor;HitTest 对 panel 点击不调用 `clickListCursor`。
  2. 在容器列表 panel 内滚动鼠标滚轮,选中行实时上下移动,与键盘 `j` / `k` 行为一致;`ViewOffset` 同步推进。
  3. 在头部 / 消息栏 / footer 区域滚动滚轮,不触发列表 cursor 变化。
  4. 终端原生选区复制(鼠标拖选)在 panel 内仍可用。

### BR-031 容器页容器资源占用(CPU / 内存)未实时刷新

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 在容器页按 `m` 开启 stats 后,选中容器的 CPU / 内存数据没有按预期实时刷新;视图上的数值停留在初始值或上一次刷新时的快照。
- 当前行为:
  - `internal/tui/update/update_tick.go:32` `handleStatsTick` 已实现:容器页激活时 `StatsActive=true`,按 `Docker.StatsPollSec`(默认 3 秒) 周期为**当前过滤后**的容器列表派发 `FetchStats` 命令;`update_tick.go:17` `handleStatsReceived` 把数据写入 `m.Resources.Containers.Stats[msg.ContainerID]`。
  - `internal/tui/ui/pages/containers/view.go:90` 渲染时通过 `cm.Stats[c.ID]` 读取最新值。
  - 但实测中数据没有"看起来"实时刷新:可能原因包括 (a) `tea.Tick` 被 `Feedback.ToastGeneration` 或 detail revision 等其它 ticker 阻塞,导致 stats tick 实际频率低于 `StatsPollSec`;(b) 长轮询或 socket I/O 在 podman no-cgo 模式下阻塞过久,(c) 数据写入路径触发了不必要的全屏重绘,(d) `FetchStats` 出错被静默忽略。
- 代码锚点:
  [internal/tui/update/update_tick.go:17](/home/debi/IdeaProjects/docker-tui/internal/tui/update/update_tick.go:17) `handleStatsReceived`
  [internal/tui/update/update_tick.go:32](/home/debi/IdeaProjects/docker-tui/internal/tui/update/update_tick.go:32) `handleStatsTick`
  [internal/tui/update/update.go:129](/home/debi/IdeaProjects/docker-tui/internal/tui/update/update.go:129) `StatsTick` 分发
  [internal/tui/ui/pages/containers/view.go:90](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/containers/view.go:90) stats 渲染读取
  [internal/tui/state/containers.go](/home/debi/IdeaProjects/docker-tui/internal/tui/state/containers.go) `Stats` map 与 `StatsActive` 字段
- 期望行为:
  1. 容器页开启 stats 后,选中的(以及当前过滤后可见的)容器 CPU / 内存数据按 `StatsPollSec` 周期稳定刷新;视觉上能看到数值跳动。
  2. stats tick 与其它 ticker(toast / runtime selector / detail revision)独立,任一 ticker 卡顿不影响 stats。
  3. stats 拉取失败(socket 超时 / podman REST 错误)有明确 toast 或状态栏提示,不再静默忽略。
  4. stats 数据更新路径避免触发全屏重绘,只刷新 stats 列;`BenchmarkRenderCachedContainerDetail` 等现有滚动 benchmark 不应回退。
- 验收标准:
  1. 容器页按 `m` 开启 stats,等待 3 秒以上,选中的(以及当前过滤后可见的)容器的 CPU% / 内存% 至少有一次刷新;数值变化可观察。
  2. 在容器页停留 30 秒以上,stats 数据按 `StatsPollSec` 持续刷新,不被其它 ticker 阻塞。
  3. 拔掉容器运行时后,UI 显示明确的 stats 拉取失败提示(类似 toast 或错误占位),不再显示陈旧值。
  4. stats 更新不影响滚动 / 选中行高亮 / 详情页 footer。

### BR-032 表格排序:N / Ctrl+N 正反排序,鼠标点击表头排序

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 当前表格排序由 `O`(`toggleSortColumn` 切换列)与 `Ctrl+O`(`toggleSortDirection` 切换升降序)承担。
  - 用户期望:`N` = 正排序(升序或切换到下一个排序列的升序),`Ctrl+N` = 反排序(降序或切换到下一个排序列的降序);此外,**鼠标点击表头**也能排序。
- 当前行为:
  - `internal/tui/keyboard/keyboard.go:221-234`:`KeyO` 调 `toggleSortColumn`,`KeyCtrlO` 调 `toggleSortDirection`。
  - `toggleSortColumn` 在容器 / 镜像 / 网络 panel 间循环排序字段;`toggleSortDirection` 翻转 `SortAsc` 标志。
  - **没有** `KeyN` / `KeyCtrlN` 在正常模式下的排序绑定;`KeyN` 只在 dialog 中被当作"关闭"(`keyboard.go:145`)。
  - **没有**鼠标点击表头的处理;`mouse.go` 中只有 `clickListCursor` / `HitTest` / 滚轮事件,**没有**表头点击路由。
- 代码锚点:
  [internal/tui/keyboard/keyboard.go:221](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/keyboard.go:221) `KeyO` 入口
  [internal/tui/keyboard/keyboard.go:228](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/keyboard.go:228) `KeyCtrlO` 入口
  [internal/tui/keyboard/keyboard.go:248](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/keyboard.go:248) `toggleSortColumn`
  [internal/tui/keyboard/keyboard.go:268](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/keyboard.go:268) `toggleSortDirection`
  [internal/tui/ui/app/mouse.go:154](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/mouse.go:154) `HitTest` 缺表头分支
  [internal/tui/ui/component/table.go:54](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/table.go:54) `SortColKey` / `SortAsc` 数据结构已存在
- 期望行为:
  1. `N` 键 = 正排序;`Ctrl+N` 键 = 反排序。具体语义二选一,与用户确认:
     - 选项 A:`N` 把当前排序列设为升序,`Ctrl+N` 把当前排序列设为降序(不切换列)。
     - 选项 B:`N` 切换到下一个排序列(升序),`Ctrl+N` 切换到下一个排序列(降序)。
     - 推荐选项 A,与"过滤式"交互一致。
  2. 鼠标点击表头:
     - 在升序列上点击 → 切到降序。
     - 在降序列上点击 → 切到升序。
     - 在非当前排序列上点击 → 切到该列,默认升序。
     - 表头必须显示当前排序状态(箭头 / 高亮)。
  3. 现有的 `O` / `Ctrl+O` 是否保留由用户决定:可以并存(`O` 切列,`N` 改方向)或废弃(只保留 `N` / `Ctrl+N` + 鼠标)。
  4. 排序改动后 `Cursor` 与 `ViewOffset` 应归零或保持一致;Help / Footer 显示当前可用的排序快捷键。
- 验收标准:
  1. 按 `N` 后,排序列标记为升序,排序箭头在表头显示,行顺序按升序排列。
  2. 按 `Ctrl+N` 后,排序列标记为降序,排序箭头反向,行顺序按降序排列。
  3. 鼠标点击表头某列:若非当前列,切到该列升序;若当前列升序,切到降序;若当前列降序,切到升序。
  4. 排序时 `Cursor` 不丢失选中行;`ViewOffset` 跟随 cursor 调整。
  5. Help / Footer 准确显示当前可用的排序快捷键。

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
  - `internal/data/runtime/podman/service_action.go:63-110` 已经实现了 Update / Diff / Export / Commit / Wait 的 Podman runtime 调用。`runtimeapi.ActionCopy` 等也已存在,但 Docker 端是否同等实现需要进一步核对(`internal/data/runtime/docker/containers.go` / `service_action.go`)。
- 代码锚点:
  [internal/tui/keys/registry.go:66](/home/debi/IdeaProjects/docker-tui/internal/tui/keys/registry.go:66) TASK-019 默认键位注册
  [internal/tui/keyboard/actions.go:16](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/actions.go:16) `handleAction` switch 缺 TASK-019 case
  [internal/tui/keys/action.go:23](/home/debi/IdeaProjects/docker-tui/internal/tui/keys/action.go:23) TASK-019 Action 定义
  [internal/tui/keys/keys.go:34](/home/debi/IdeaProjects/docker-tui/internal/tui/keys/keys.go:34) `KeymapConfig` TASK-019 字段
  [internal/data/runtime/podman/service_action.go:63](/home/debi/IdeaProjects/docker-tui/internal/data/runtime/podman/service_action.go:63) Podman Update/Diff/Export/Commit/Wait/Copy 分发
  [internal/data/runtime/docker/containers.go](/home/debi/IdeaProjects/docker-tui/internal/data/runtime/docker/containers.go) (待核对 Docker 端 Copy/Update 等是否实现)
- 期望行为:
  1. **取消 docs/requirements.md 的"已实现"声明**:这些动作实际未实现,文档需改回"待实现"。
  2. **接入 handler**:为每个 TASK-019 动作补一个 `case keys.ActionContainerCopy` 等分支,根据操作打开对应表单(dialog)或直接执行:
     - `Copy`:打开"源路径 / 目标路径"对话框,执行 `Engine.Containers().CopyToContainer` / `CopyFromContainer`。
     - `Update`:打开"内存 / CPU / 重启策略"表单,执行 `Engine.Containers().Update`。
     - `Diff`:打开 diff 视图(类似详情但只显示 filesystem diff)。
     - `Export`:打开"目标 tar 路径"对话框,执行 `Engine.Containers().Export`。
     - `Commit`:打开"repository:tag"对话框,执行 `Engine.Containers().Commit`。
     - `Wait`:打开"阻塞条件(运行 / 停止 / 删除)"对话框,执行 `Engine.Containers().Wait`。
  3. **解决键位冲突**:`F2` 与 `ActionSwitchRuntime` / `ActionContainerDiff` 冲突;`Ctrl+K` 与 `ActionContainerKill` / `ActionContainerCommit` 冲突;`Ctrl+O` 与排序 `toggleSortColumn` 冲突。建议:
     - TASK-019 动作默认绑定由用户重设,不沿用当前 `F2 / Ctrl+K / Ctrl+O` 的冲突值。
     - 提供与 BR-032 一致的按键规则,确保所有动作的键位唯一且不与现有功能冲突。
  4. **Docker runtime 支持**:核对 Docker engine 是否提供 Copy/Update/Commit/Wait 的等价 API,缺失时在 helper 中给出 "Docker 引擎不支持" 的明确 toast,而不是静默无响应。
- 验收标准:
  1. 在容器页按 TASK-019 的默认键位,要么弹出对应表单(dialog),要么执行动作并反馈结果,**不再静默吞键**。
  2. 文档 `requirements.md` 中"镜像列表、Pull、Prune、Tag、Push、Save、Load、详情"等声称"已实现"的能力与代码保持一致;TASK-019 这一类"已实现"但实际 handler 缺失的条目必须修订。
  3. TASK-019 动作的键位与现有 `O` / `Ctrl+O` / `F2` / `Ctrl+K` 不冲突;Help / Footer 显示当前生效键位。
  4. 至少 Podman runtime 下 `Copy` (容器 → 主机) 与 `Export` (tar 导出) 完整跑通;Docker runtime 下给出明确"未支持"提示或同等实现。
  5. 不与 BR-032 的排序键位 (`N` / `Ctrl+N` / 鼠标点击表头) 冲突;两个 BUG 修复时统一规划键位分配。

### BR-034 镜像 History 顶层页 (H 键) — 历史因性能取消,需重新设计

- 状态: `open`
- 优先级: `high`
- 症状:
  - 在镜像页按 H 当前是 `Viewport.ToggleHeader()`(`keyboard.go:215`),不是进入镜像 History。
  - 镜像详情页里曾渲染完整 History 段,大镜像(>50 layer)下每次滚动都重建整段 section,因**性能问题**取消。
  - 用户期望:H 键在**镜像页**进入**镜像 History 顶层页**(浏览 layer history),详情页不再渲染 History 段。
- 当前行为:
  - `keys.KeyH` 在 `keyboard.go:215` 直接执行 `m.Viewport.ToggleHeader()`,没有走 `ActionHelp`;镜像详情中的 `buildImageDetailDataSections` 与文本回退路径已不再渲染 History 段,Docker / Podman 的详情 inspect 也暂时不再请求 History endpoint。详情滚动已用 BR-016 / BR-007 优化(按 revision 缓存 + ClampVisibleOffset),但仍未重新引入 History 顶层页。
- 代码锚点:
  [internal/tui/keyboard/keyboard.go:215](internal/tui/keyboard/keyboard.go:215) `KeyH` 当前 ToggleHeader
  [internal/tui/ui/pages/detail/image.go:14](internal/tui/ui/pages/detail/image.go:14) 镜像分区构造(已不渲染 History)
  [internal/data/runtime/image.go:57](internal/data/runtime/image.go:57) `ImageHistoryLayer` 数据结构已定义
  [internal/data/runtime/docker/images.go:107](internal/data/runtime/docker/images.go:107) `InspectImageDetailContext` history 字段
  [internal/data/runtime/podman/service_image.go](internal/data/runtime/podman/service_image.go) Podman 等价 inspect
- 期望行为:
  1. 在**镜像页**(`ModeImages` / `PanelImages`),按 H 键进入 `ModeHistory`(独立顶层页,不是详情页内分区)。
  2. History 顶层页只展示镜像 layer 列表:
     - 每层一行:Created (相对时间) / Size / CreatedBy(单行摘要,可展开)/ Comment
     - 支持 j/k 上下选行、PgUP/PgDn 翻页、`/` Filter 按 CreatedBy 过滤。
     - 顶部固定显示镜像 Repository:Tag 与总 layer 数。
  3. 数据源复用 `InspectImageDetail()` 的 history 输出;无需新加 engine API。
  4. **解决历史性能问题**:
     - 复用 BR-007 的 `Documents map[DetailSource]DetailDocument` 按 revision/source 缓存机制,只渲染当前可见行。
     - 大镜像(>50 layer)滚动时,验证 benchmark 不退化(`BenchmarkRenderCachedContainerDetail` 不回退,新增镜像 history 等价 benchmark)。
     - history 数据通过 `InspectImageDetail` 一次拿全,不每行拉取。
- 验收标准:
  1. 镜像页按 H 键,弹出 History 顶层页,标题为 `<image>:<tag> history`;行数等于镜像 layer 数。
  2. 选中镜像 > 50 layer 时,上下滚动 / PgDn / j 翻页均无卡顿,体感与 BR-016 修复后一致。
  3. History 页与镜像详情页的 `s` Source(YAML / JSON)互不干扰。
  4. 详情页不再渲染 History 分区(沿用 BR-008 决策);`History` 段在 detail 渲染路径上完全移除。
- 设计参考: [docs/feature-design.md §5.1](docs/feature-design.md#51-镜像-history-顶层页-h-键)

### BR-035 Events 独立面板 (F3 键) + Network Connect/Disconnect

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 当前 `Engine.Events().Subscribe(...)` 已在 `update/update_events.go:28-31` 实现后台订阅,合并到资源更新路径,但**没有独立的 Events 浏览面板**;用户无法浏览 runtime events 历史。
  - `NetworkService` 接口 (`docker/service_network.go` / `podman/service_network.go`) 只有 List/Create/Remove/Prune,**没有** Connect / Disconnect;无法动态调整容器的网络接入。
- 当前行为:
  - F3 当前未绑定任何动作(`KeyF3` 在 `keys/keys.go` 已定义,但 `registry.go` 没有 `KeyF3` 注册条目)。
  - 容器启动后只能使用启动时配置的网络,无法在运行时补接 / 脱离。
- 代码锚点:
  [internal/tui/update/update_events.go:28](internal/tui/update/update_events.go:28) `subscribeEventsCmd`
  [internal/tui/update/update_events.go:31](internal/tui/update/update_events.go:31) `subscribeEventsCmd` 定义
  [internal/data/runtime/streaming.go:64](internal/data/runtime/streaming.go:64) `EventService` 接口
  [internal/data/runtime/docker/service_event.go:15](internal/data/runtime/docker/service_event.go:15) Docker `Events()` 实现
  [internal/data/runtime/podman/service_event.go](internal/data/runtime/podman/service_event.go) Podman `Events()` 实现
  [internal/data/runtime/docker/service_network.go](internal/data/runtime/docker/service_network.go) Docker `NetworkService` 缺 Connect/Disconnect
  [internal/data/runtime/podman/service_network.go](internal/data/runtime/podman/service_network.go) Podman 同上
  [internal/tui/keys/registry.go](internal/tui/keys/registry.go) `KeyF3` 未注册
- 期望行为:
  1. **Events 独立面板**:
     - 全局按 `F3` 进入 `ModeEvents` / `PanelEvents`(`registry.go` 注册 `{ActionEvents, []string{KeyF3}, app}`)。
     - UI:类似日志页的实时流 + 列表(条目含 `Time / Type / Action / Resource` 四列)。
     - 数据源:`Engine.Events().Subscribe(...)`,在后台订阅并在独立 buffer(默认保留最近 1000 条)。
     - 支持 `Filter`:按 type (container/image/network/volume) / action (start/stop/create/destroy/pull/push) / 时间范围。
     - 支持 `Pause / Resume`:`space` 暂停累积,新事件进入缓冲,Resume 时一次性冲入。
     - `Esc` 返回;`Ctrl+D` 清空 buffer(高危动作,需 confirm)。
  2. **Network Connect / Disconnect**:
     - 在网络页:`ModeNetworks` 增加动作 `Connect Container` / `Disconnect Container`,触发表单"选容器"或从已选网络弹出"挂接容器"菜单。
     - 在容器详情页:显示"已接入网络列表" + 每行 `[Disconnect]` 按钮或快捷键。
     - runtime 层新增 `NetworkService.Connect(ctx, networkID, containerID, opts)` / `Disconnect(ctx, networkID, containerID, force bool)`:
       - Docker adapter:`s.client.cli.NetworkConnect(ctx, networkID, containerID, config)` / `NetworkDisconnect(ctx, networkID, containerID, force)`.
       - Podman adapter:等价 REST 调用。
     - 已有 `NetworkService.List/Create/Remove/Prune` 接口扩展,**不破坏现有实现**。
     - 权限 / 引擎能力检测:`Engine.Capabilities()` 检查是否支持 connect;Docker 默认支持,Podman 需要 rootless 配置。
- 验收标准:
  1. 全局按 F3,进入 Events 面板;Docker / Podman events 流持续滚动;按 space 暂停、再按恢复。
  2. 容器启动 / 停止 / 镜像 pull 等事件在 Events 面板实时出现,带时间戳与动作类型。
  3. 在网络页 / 容器详情页能把运行中容器接入新网络;`docker exec ctr ping new-net-peer` 验证网络可达。
  4. 把容器从网络 disconnect 后,容器在该网络的接口消失;不影响其它网络。
  5. Events 面板与 Network Connect/Disconnect 都按 [docs/feature-design.md](docs/feature-design.md) 中的设计目标实现,不破坏现有功能。
- 设计参考: [docs/feature-design.md §5.7](docs/feature-design.md#57-events-流独立面板-f3)、[§5.3](docs/feature-design.md#53-network-connect--disconnect)

### BR-036 镜像 tarball 导入功能 (Image Import)

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 镜像页缺少 tarball 导入动作;用户拿到 `docker export` 的扁平 tar 或第三方 tar 后,无法在 dtui 内导入为新镜像。
  - 当前 ImageService (`docker/service_image.go` / `podman/service_image.go`) 只有 List / Inspect,无 Import。
- 当前行为:
  - 镜像页 Action Bar / 命令面板无 "Import" 入口。
  - `ActionImageLoad` (`Ctrl+L`) 是 `docker load`(从 docker save 多 layer tar 恢复),与 `docker import`(从扁平 tar 创建单层新镜像)语义不同,不能复用。
- 代码锚点:
  [internal/data/runtime/image.go](internal/data/runtime/image.go) `ImageService` 接口(无 Import)
  [internal/data/runtime/docker/service_image.go](internal/data/runtime/docker/service_image.go) Docker ImageService
  [internal/data/runtime/podman/service_image.go](internal/data/runtime/podman/service_image.go) Podman ImageService
  [internal/tui/keys/action.go:35](internal/tui/keys/action.go:35) `ActionImageLoad` (现有 load,与 import 不同)
  [internal/tui/keys/registry.go:78](internal/tui/keys/registry.go:78) `ActionImageLoad` ← `Ctrl+L`
- 期望行为:
  1. 镜像页 Action Bar 中加 "Import tarball" 动作(走 BR-039 Action Bar,不直接占用键位)。
  2. 触发表单:
     - 源文件路径(`/path/to/file.tar` 或 URL)。
     - 目标 repository:tag(可选,如 `myapp:custom`)。
     - 提交消息 / 作者(可选)。
  3. 提交后调 `Engine.Images().Import(ctx, source, ref, msg)`(runtime 需新增):
     - Docker:`s.client.cli.ImageImport(ctx, source, ref, msg)`.
     - Podman:等价 REST `images/import`.
  4. 进度反馈:复用 `ImageTransfer` 框架的进度 / toast / 取消机制。
- 验收标准:
  1. 镜像页 Action Bar 中 "Import tarball" 可点击;触发表单。
  2. 提交后,新镜像出现在镜像列表,inspect 可见但只有 1 层(history 为空)。
  3. `docker save` 出来的多 layer tarball 走 `Load` (`Ctrl+L`),`docker export` 出来的扁平 tarball 走 Import,两条路径独立、不混淆。
  4. Podman 端行为一致。
- 设计参考: [docs/feature-design.md §5.10](docs/feature-design.md#510-image-import-br-036)

### BR-037 docker / podman Registry Login(私库认证)

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 当前无 `docker login` 等价入口;用户访问私有 registry 时必须在外部 CLI 登录,dtui 内无法完成。
- 当前行为:
  - `ImageService` 没有 `Login` 方法;`keys.Action*` 无 `ActionLogin` 注册。
  - 实际上 docker / podman 凭证存储在 `~/.docker/config.json`,引擎自身处理凭证;dtui 只需调用 engine 完成登录即可,无需自己存凭证。
- 代码锚点:
  [internal/data/runtime/image.go](internal/data/runtime/image.go) `ImageService` 接口(无 Login)
  [internal/data/runtime/docker/service_image.go](internal/data/runtime/docker/service_image.go) Docker ImageService
  [internal/data/runtime/podman/service_image.go](internal/data/runtime/podman/service_image.go) Podman ImageService
  [internal/tui/keys/action.go](internal/tui/keys/action.go) 无 `ActionRegistryLogin` 定义
- 期望行为:
  1. Action Bar 中加 "Registry Login" 动作(走 BR-039 Action Bar,不直接占用键位)。
  2. 触发表单:
     - 服务器 URL(默认 `docker.io` 或 `https://index.docker.io/v1/`)。
     - 用户名。
     - 密码(掩码输入)。
     - 高级:跳过 TLS 校验 / 标识身份。
  3. 提交后调 `Engine.Login(ctx, server, user, password, opts)`(runtime 需新增):
     - Docker:`s.client.cli.RegistryLogin(ctx, auth)`.
     - Podman:等价 REST `auth`.
  4. **不存储密码到 dtui 配置**;依赖引擎自身的 `~/.docker/config.json`。
  5. Podman 与 docker 凭证共用 `~/.docker/config.json`,可直接复用。
- 验收标准:
  1. Action Bar 中 "Registry Login" 可点击,表单可填写。
  2. 登录成功后,dtui 内 `ImagePull` (`Ctrl+P`) 私有仓库镜像成功,无需外部 `docker login`。
  3. 错误凭证返回明确 toast(`Unauthorized` 等),不泄露密码。
  4. Podman 端等价行为。
  5. dtui 进程退出后,引擎凭证仍存在(由引擎自身管理)。
- 设计参考: [docs/feature-design.md §5.9](docs/feature-design.md#59-registry-login-br-037)

### BR-038 容器 Exec 页面 UI 不符合 shell 终端

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 容器页按 `e` 进入 Exec 后,Exec 页面 UI 不像 shell 终端:
    - 缺少典型的 shell 提示符、路径前缀、颜色。
    - 历史命令 / Tab 补全 / 方向键(↑↓)历史回看 / Ctrl+L 清屏等 shell 特性未提供。
    - 当前实现更像"裸 stdin/stdout 转发",不是真正的 shell 体验。
- 当前行为:
  - `internal/tui/ui/widget/dialog/exec.go` 提供 `RenderExecDialog` / `RenderTextInput` 风格的 exec 对话框(原 BR-033 Task-019 之前已存在),输出是单行输入框 + 输出区,不是全屏 terminal。
  - 用户期望:Exec 页面提供**真 shell 体验**:全屏 terminal view / 真实 PTY 透传 / ANSI 颜色 / shell 行编辑(↑↓ 历史 / Tab 补全 / Ctrl+L 清屏)。
- 代码锚点:
  [internal/tui/ui/widget/dialog/exec.go](internal/tui/ui/widget/dialog/exec.go) `RenderExecDialog` / `RenderTextInput`
  [internal/tui/keyboard/container_action.go](internal/tui/keyboard/container_action.go) `doAutoExecAction`
  [internal/data/runtime/docker/service_exec.go](internal/data/runtime/docker/service_exec.go) Docker exec PTY 透传
  [internal/data/runtime/podman/service_exec.go](internal/data/runtime/podman/service_exec.go) Podman exec PTY 透传
- 期望行为:
  1. Exec 进入新模式 `ModeExec`,**全屏渲染**(类似 logs 页),不再是 floating dialog。
  2. 真正的 PTY 透传:`Engine.Exec().Start(ctx, containerID, []string{"/bin/sh"}, TTY=true)`(Docker / Podman 都已支持)。
  3. 输入处理:
     - 方向键 ↑↓ 翻历史命令。
     - Tab 触发补全(走 shell 自身)。
     - Ctrl+L 清屏。
     - Ctrl+C 转发 SIGINT 到前台进程,不退出 dtui。
     - Ctrl+D 退出前台进程(类似真实终端 EOF)。
  4. 输出:ANSI 颜色 / 宽字符 / 进度条正确渲染。
  5. `Esc` 退出 exec mode,**不发信号给前台进程**;确认窗口(类 vim Ctrl+P Ctrl+Q)脱离但不终止。
- 验收标准:
  1. 容器页按 `e` 进入 exec,页面**全屏 shell**,提示符正确(`root@container-id:/path#` 之类)。
  2. 在 shell 内执行 `ls /` / `cat /etc/hostname` / `top` 等命令,输出正确(包括 ANSI 颜色)。
  3. ↑↓ 翻历史、Tab 补全、Ctrl+L 清屏、Ctrl+C 发 SIGINT(能被前台进程接收)均工作。
  4. Ctrl+D 退出前台 shell 但容器继续运行(对应 `detach` 语义)。
  5. Esc 退出 dtui 的 exec mode,**不**杀掉容器前台进程。
- 设计参考: [docs/feature-design.md §5.4](docs/feature-design.md#54-container-attach-vs-exec)

### BR-039 取消 Q 键退出,改为每页多功能 Action Bar

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 当前 `Q` 键直接退出应用(`keys/registry.go:38` `{ActionQuit, []string{KeyQ, KeyCtrlC}, app}`),用户希望**取消** Q 退出功能。
  - 当前各 table 页面(容器 / 镜像 / 卷 / 网络)的可用操作分散在多个直接键位 / 命令面板,无法集中发现。
- 当前行为:
  - `Q` 与 `Ctrl+C` 都绑定 `ActionQuit`,按了立刻退出。
  - 容器 / 镜像 / 卷 / 网络页可用操作:直键 (`s`/`Ctrl+S`/`p`/`Ctrl+D` 等) + 命令面板(`:` + `rename`/`top`/`port` 等)。
  - 没有集中可发现的多功能面板。
- 代码锚点:
  [internal/tui/keys/registry.go:38](internal/tui/keys/registry.go:38) `ActionQuit` ← `KeyQ, KeyCtrlC`
  [internal/tui/keys/actions.go:18](internal/tui/keys/actions.go:18) `ActionQuit` handler (`tea.Quit`)
  [internal/tui/keys/commands.go](internal/tui/keys/commands.go) 命令面板列表
  [internal/tui/keyboard/command.go](internal/tui/keyboard/command.go) `executeCommand`
- 期望行为:
  1. **取消 Q 键退出**:`registry.go` 删除 `KeyQ` 绑定,只保留 `Ctrl+C` 作为硬退出(`Ctrl+C` 在 dtui 中通常需要二段确认,沿用现有 `EscPending` 机制)。
  2. **新增 Action Bar**:每页触发后弹出浮层,显示当前页可用动作清单。
  3. 触发键(候选):
     - 方案 A:`;` (vim 风格)作为 Action Bar 入口,与现有 `:` 命令面板并列。
     - 方案 B:扩展 `:` 命令面板,加入非命令类动作(普通按钮),并提供可点击列表。
     - 优先方案 A(更符合 vim 习惯),如用户偏好可改方案 B。
  4. Action Bar 内容(按当前 panel 动态生成):
     - **镜像页**: Pull / Tag / Push / Save / Load / **Import** (BR-036) / Prune / Remove / Inspect / Detail / **History** (BR-034) / Refresh / Filter。
     - **容器页**: Start / Stop / Restart / Pause / Kill / Remove / Logs / Stats / Exec / Top / Port / Rename / Inspect / Detail / **Update** / **Diff** / **Export** / **Commit** / **Wait** / **Copy** (BR-033) / Refresh / Filter。
     - **卷页**: Create / Prune / Remove / Detail / Inspect / Refresh / Filter。
     - **网络页**: Create / Prune / Remove / Detail / Inspect / **Connect Container** / **Disconnect Container** (BR-035) / Refresh / Filter。
     - **通用(所有页)**: Login (BR-037) / Switch Runtime (F2) / Refresh Connections (R/F12) / Help (F1/?) / Events (F3, BR-035) / Quit (Ctrl+C)。
     - **detail / logs / help / 其它无新动作的页**: Action Bar 显示"该页面无额外动作",或直接不打开(空页面不弹)。
  5. 每个条目显示:**动作名 + 当前键位**(`Tag (Ctrl+T)` / `Save (Ctrl+E)`)。
  6. **已有命令保留**:`:rename` / `:top` / `:port` / `:images` / `:containers` 等命令面板入口**不取消**,与 Action Bar 并存。
- 验收标准:
  1. 按 Q 不再退出应用(可改为显示 Help 或 toast "Q 已重映射到 Action Bar",具体看用户偏好)。
  2. 在镜像页按 `;`,弹出 Action Bar,列出该页所有可用动作 + 当前键位。
  3. Action Bar 选中某动作后,与原直键效果一致;选中"无键位"动作(如 BR-033 暂未分配键位的动作),也能执行。
  4. detail / logs 页不强行弹 Action Bar(空列表或直接 no-op)。
  5. 命令面板 `:` 仍可用,所有原命令保留。
  6. Help / Footer 准确反映当前键位(移除 `Q`)。
- 设计参考: [docs/feature-design.md §5.6](docs/feature-design.md#56-q-键退出--多功能-action-bar)

### BR-040 Dialog 风格统一:四周透明 + panel 居中

- 状态: `open`
- 优先级: `medium`
- 症状:
  - 当前 dialog(confirm / shell / filter / exec / notification / selection / 自定义表单)风格不统一:
    - 有的占满整屏 / 有的浮在中间 / 有的位置在 panel 外。
    - 边框是否透明、是否有 padding、位置算法各不相同。
  - 用户明确要求:**所有 dialog 均为"四周透明 + 中间窗口"**,位置是**当前 panel 的居中**(因为作用域是 panel)。
- 当前行为:
  - `internal/tui/ui/widget/dialog/` 下各 dialog(confirm / notification / selection / exec / view)的渲染位置和边框样式各自实现,无统一约束。
  - `widget/dialog/overlay.go` 提供 overlay 工具但调用方各自选择尺寸 / 位置。
  - 不同 dialog 在 panel 居中 / 全屏 / 顶部等位置混用。
- 代码锚点:
  [internal/tui/ui/widget/dialog/overlay.go](internal/tui/ui/widget/dialog/overlay.go) `PlaceOverlay`
  [internal/tui/ui/widget/dialog/view.go](internal/tui/ui/widget/dialog/view.go) 各 dialog 渲染
  [internal/tui/ui/widget/dialog/confirm.go](internal/tui/ui/widget/dialog/confirm.go) confirm dialog
  [internal/tui/ui/widget/dialog/notification.go](internal/tui/ui/widget/dialog/notification.go) notification
  [internal/tui/ui/widget/dialog/selection.go](internal/tui/ui/widget/dialog/selection.go) selection
  [internal/tui/ui/widget/dialog/exec.go](internal/tui/ui/widget/dialog/exec.go) exec dialog
  [internal/tui/ui/app/layout.go](internal/tui/ui/app/layout.go) 各 dialog 调用方
- 期望行为:
  1. **统一接口**:新增 `widget/dialog/centered.go`,提供 `CenterOnPanel(panelRect Rect, content string, w, h int) string`:
     - 取 panel 的几何(从 `LayoutReport.Panel` / 视图层传入的 rect)。
     - 计算内容居中位置 = `panelRect.bodyTop + (panelRect.bodyRows - h) / 2`,`x = (panelW - w) / 2`。
     - 渲染时,panel 内除 dialog 区域外的部分**透明**(不画背景填充)。
  2. **对话框渲染规范**:
     - 边框:lipgloss `Border` 样式,**不**带 `Background` 填充。
     - 内边距:统一 `Padding(1, 2)` 之类。
     - 标题:统一在顶部,带可选 icon / 颜色。
     - 内容区:背景透明或半透明。
  3. **应用范围**:
     - confirm / notification / selection / exec / filter / 自定义表单(Copy / Update / Commit / Wait / Import / Login / Events Filter)。
     - 渲染日志 / 详情 / 主列表 / footer / header 仍按各自语义,但**所有 dialog(浮层)** 走 `CenterOnPanel`。
  4. **不破坏现有 BR-040 之前已经稳定的 dialog 行为**(如 selection 的 tab 切换、confirm 的默认焦点)。
- 验收标准:
  1. 所有 dialog 在 panel 内居中显示,四周透明,不覆盖 header / footer / message rail。
  2. 同类 dialog(例如两个 confirm)风格一致:边框 / 内边距 / 标题位置相同。
  3. panel 缩窄 / 拉宽时,dialog 跟着居中且不溢出 panel。
  4. dialog 内的按钮 / 输入框 / 文本行版式统一。
- 设计参考: [docs/feature-design.md §5.8](docs/feature-design.md#58-dialog-风格统一四周透明--panel-居中)

### BR-028 (TBD - 待用户补充)

- 状态: `pending`
- 优先级: -
- 症状: 用户第 6 条 BUG 内容未提供。
- 期望行为: 用户明确补充第 6 条 BUG 描述后,按 `新增条目模板` 拆解为症状、当前行为、代码锚点、期望行为、验收标准。

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
