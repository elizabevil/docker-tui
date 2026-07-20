# dtui 页面与性能优化设计

> 建立日期: 2026-07-20
> 作用: 收敛页面布局优化、页面模板统一和渲染性能优化方向。

本文不替代 [current-design.md](current-design.md)。

- `current-design.md` 记录当前实现事实。
- 本文记录下一阶段页面结构优化与性能收敛方案。

## 设计目标

- 让主界面在模式切换、通知出现、搜索激活时保持稳定，不出现明显布局跳变。
- 让列表页、双栏页、详情页、日志页形成统一页面模板，而不是各页各自演化。
- 让 Footer / Help / 页面标题摘要基于同一动作与上下文系统派生。
- 控制高频渲染路径中的重复字符串计算、宽度计算和无效重排。
- 为后续 BR-001 ~ BR-005 的实现提供稳定 UI 容器。

## 审批约束

- 审批日期: `2026-07-20`
- 已确认结论:
  1. 固定页面骨架是当前页面与性能优化第一优先级。
  2. 后续页面渲染体系必须建立在固定骨架之上。
  3. 本文所有优化方向已获通过，但实现时必须与 [bugfix-design.md](bugfix-design.md) 中既有 BR 方案协同。
- 协同要求:
  1. Query / Message / Footer / Help 的渲染和数据来源必须兼容 `BR-002 / BR-003 / BR-004 / BR-005`。
  2. Detail page 优化必须兼容 `BR-001` 的 section model 与 builder 方案。
  3. 不允许为了页面优化重新引入新的静态快捷键说明源、独立消息通道或新的详情文本反解析依赖。

## 已验证的现状问题

### 1. 页面主体高度会随顶部/底部内容变化

- `RenderApp()` 当前会先渲染 header / toast / footer，再按实际行数回算中间面板高度。
- `searchBar` 进入后直接插入中部内容上方，会进一步压缩主内容区。
- 结果是:
  - toast 出现时 panel 高度变化
  - `/` 或 `:` 激活时 panel 高度变化
  - footer 行数变化时 panel 高度变化

代码锚点:
- [internal/tui/ui/app/layout.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/layout.go:289)
- [internal/tui/ui/app/layout.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/layout.go:317)

### 2. Help 与 Footer 仍是两套快捷键来源

- Footer 直接由 `footer.go` 内部 provider 生成。
- Help 仍从 `shortcuts.jsonc` 静态读取。
- 页面说明与实际行为没有统一信息源。

代码锚点:
- [internal/tui/ui/widget/footer/footer.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/footer/footer.go:39)
- [internal/tui/ui/pages/help/view.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/help/view.go:35)

### 3. 页面类型缺少统一模板

- 容器/镜像/卷/网络页都属于资源列表页，但顶部摘要、选中项预览、页内 footer 责任并不完全一致。
- Compose 已形成双栏结构，但这套结构还没有沉淀成通用模板。
- Detail 与 Logs 是信息密度很高的只读页，但两者仍各自维护自己的 header / footer 反馈方式。

代码锚点:
- [internal/tui/ui/pages/containers/view.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/containers/view.go:19)
- [internal/tui/ui/pages/images/view.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/images/view.go:18)
- [internal/tui/ui/pages/compose/view.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/compose/view.go:67)
- [internal/tui/ui/pages/detail/view.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/detail/view.go:39)
- [internal/tui/ui/pages/logs/view.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/logs/view.go:49)

### 4. 表格渲染路径仍有重复计算

- `RenderTable()` 每次渲染都重新扫描 header 与所有 row，计算可见宽度。
- `VisibleLen()` 被多次重复调用，且字符串样式渲染与宽度计算耦合较深。
- 这在大列表、高频刷新、容器 stats 开启时会放大成本。

代码锚点:
- [internal/tui/ui/component/table.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/table.go:68)
- [internal/tui/ui/component/table.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/table.go:214)
- [internal/tui/utils/ansi.go](/home/debi/IdeaProjects/docker-tui/internal/tui/utils/ansi.go:40)

### 5. 日志页宽度计算不基于真实 panel content width

- 日志换行宽度当前直接使用 `m.Width - 10`。
- 这不是 panel body 的真实可用宽度，会导致日志换行与视觉布局不一致。

代码锚点:
- [internal/tui/ui/pages/logs/view.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/logs/view.go:99)

## 优化原则

### 1. 先稳定骨架，再优化单页

- 优先消除 layout jitter。
- 页面内部的信息密度、排版和说明系统建立在稳定骨架之上。

### 2. 页面模板优先于页面个例

- 新需求优先接入页面模板，不再让单个页面继续长出专属布局语言。
- 同类页面应共享结构，而不是只共享配色。

### 3. 事实与投影分离

- 动作、模式、选中对象、状态提示属于事实层。
- Footer、Help、标题摘要、通知属于投影层。
- 投影层可以长得不同，但不能各自拥有独立真相源。

### 4. 性能优化优先针对高频路径

- 先优化每帧都会走的字符串拼接、宽度计算、换行、样式组合。
- 不优先优化低频页面或只在 resize 时发生的重计算。

## 布局优化方案

### 1. 固定主骨架轨道

建议将页面整体骨架固定为五条轨道:

1. `Header Rail`
2. `Message Rail`
3. `Query Rail`
4. `Panel Rail`
5. `Footer Rail`

第一版目标高度:

- `Header Rail`: 固定 4 行
- `Message Rail`: 固定 1 行
- `Query Rail`: 固定 1 行
- `Panel Rail`: 吃掉剩余高度
- `Footer Rail`: 固定 3 行

结果约束:

- toast 是否为空，不改变 panel 高度
- `/`、`:`、filter/search 是否激活，不改变 panel 高度
- footer 内容变化，不改变 panel 高度

### 2. Query Rail 固定占位

- 不再把 search / filter / command 直接插入 panel 之上。
- Query Rail 空闲时显示弱提示或空行。
- 激活时按 mode 渲染:
  - `Filter`
  - `Search`
  - `Command`
  - 后续可扩展 `Confirm Inline`

### 3. Message Rail 统一承接提示

- `Toast`
- Filter 第一次 `Esc` 待退出提示
- Search 匹配提示
- 轻量错误 / 成功提示

全部进入同一 message rail，不再额外挤占 panel。

### 4. Footer Rail 固定三行

固定结构:

1. 全局动作行
2. 当前上下文动作行
3. 状态 / 操作日志行

约束:

- 模式切换后，第二行内容可变，但轨道高度不变
- 第三行继续固定占位，避免日志出现时跳高

## 页面模板方案

### 1. `List Page`

适用:

- `containers`
- `images`
- `volumes`
- `networks`

结构:

1. Panel Title
2. Dataset Summary
3. Selection Preview
4. Table Body
5. Table Footer

责任边界:

- 标题只表达页面名、runtime、局部上下文
- Summary 只表达当前数据集统计
- Selection Preview 只表达当前焦点对象摘要
- Table Footer 只表达表格范围和表内 hint
- 全局动作不进入表格 footer

### 2. `Split Page`

适用:

- `compose`
- `images.containers_subview`
- 后续所有“主集合 -> 子集合 / 子详情”页面

结构:

1. 左栏: 集合列表
2. 右栏: 子集合 / 子详情
3. 顶部焦点标识
4. 固定 `Left / Right / Enter / Esc` 语义

统一语义:

- `Left`: 返回 / 切回左栏 / 收起
- `Right`: 进入 / 展开 / 切到右栏
- `Enter`: 当前焦点主动作
- `Esc`: 退出子层级

### 3. `Detail Page`

适用:

- `detail`
- 后续结构化 inspect / metadata 页

结构:

1. 标题区
2. Section 流
3. 页面底部范围状态

约束:

- 不再依赖纯文本反解析决定布局
- 统一 section 标题、subtitle、empty state 呈现
- 高密度区块允许独立 renderer

### 4. `Log Page`

适用:

- `logs`

结构:

1. 上下文摘要
2. 日志正文 viewport
3. 页内状态行

约束:

- Search 不改变数据集，只负责匹配与定位
- Wrap 宽度基于真实 panel content width
- 页内 footer 只表达滚动范围和日志局部动作

### 5. `Help Page`

适用:

- `help`

结构:

1. 左栏: 全局动作、模式动作
2. 右栏: 当前页面动作
3. 顶部: 当前 context 标识

约束:

- 动作来源统一来自动作注册表
- Help 是动作系统投影，不是静态说明页

## 页面状态投影方案

### 1. 标题投影

Panel Title 统一拆为三部分:

1. `Primary Title`
2. `Context Badge`
3. `Breadcrumb`

示例:

- `Containers | docker`
- `Images | podman`
- `Compose | project > service`

### 2. Summary 投影

资源页统一使用“数据集摘要”而不是散落 hint:

- `total`
- `filtered`
- `selected`
- 局部状态统计

例如容器页:

- `42 total | 18 running | 20 exited | 4 created`

### 3. Selection Preview 投影

统一要求:

- 只显示当前焦点对象的 1 行摘要
- 可为空
- 不承担长文本详情

### 4. Message 投影

统一从通知系统派生:

- 成功 / 失败 / 警告
- Filter 退出提示
- Search 命中提示
- Runtime 切换提示

## 性能优化方案

### 1. 布局层优化

- 将 header / message / query / footer 改为固定轨道高度，减少整页重排。
- `RenderApp()` 不再依赖先 render 再 `strings.Count("\n")` 回推主区高度。
- query rail 空时也保持固定高度。

### 2. 表格宽度计算缓存

目标:

- 避免 `RenderTable()` 每次全量扫描 header + rows 计算 `VisibleLen()`

建议策略:

1. 引入 `TableLayoutCache`
2. 缓存维度:
   - page
   - profile
   - container width
   - columns schema
3. 行内容变化但 schema 不变时，优先复用宽度结果
4. 只有 profile 切换、terminal resize、列集合变化时才重算

### 3. 字符串可见宽度缓存

目标:

- 减少高频 `VisibleLen()` 重复计算

建议策略:

1. 对 header、固定标签、常见状态词建立短期缓存
2. 对表头、footer 模板、breadcrumb 结果建立按帧或按 state 的 memo
3. 不对高频变化的大块日志正文做全局缓存，避免收益过低

### 4. Render Model 分层

将页面渲染拆为:

1. `State -> ViewModel`
2. `ViewModel -> StyledLines`
3. `StyledLines -> FinalJoin`

收益:

- 字段选择与排版分离
- 更容易做增量缓存
- 更容易写测试快照

### 5. 日志页性能优化

- wrap 宽度改为真实 panel body width
- 仅对当前 viewport 做换行与高亮
- 搜索命中定位与高亮分离
- 避免每帧对全部日志重新 wrap

### 6. 背景图路径约束

`imageColorCache` 已存在，但仍应明确约束:

- 背景图色带只在 resize、换图、配置变化时重算
- 普通状态更新不触发背景重采样
- 若后续加入动画背景，应与主内容渲染完全解耦

## 建议的实现分层

### 1. Layout Shell

建议新增:

- `LayoutShell`
- `RailHeights`
- `RailRenderer`

职责:

- 只决定五条轨道高度与拼接顺序
- 不决定具体页面内容

### 2. Page Template

建议新增:

- `ListPageRenderer`
- `SplitPageRenderer`
- `DetailPageRenderer`
- `LogPageRenderer`
- `HelpPageRenderer`

职责:

- 统一页面类型内部结构
- 页面只提供 view model，不再直接拼整个 body 字符串

### 3. Projection Layer

建议新增:

- `TitleProjector`
- `SummaryProjector`
- `SelectionProjector`
- `MessageProjector`
- `ShortcutProjector`

职责:

- 将事实层状态投影为界面文案与短摘要

### 4. Table Layout Cache

建议新增:

- `TableLayoutKey`
- `TableLayoutCache`
- `ColumnWidthResolver`

职责:

- 承接列表类页面的宽度与 header 计算缓存

## 第一阶段落地范围

1. 固定五轨道骨架，消除 query / toast / footer 引发的 panel 跳变
2. 抽离 Query Rail 与 Message Rail
3. 让 Footer 固定为三行结构
4. 修正日志页 wrap 宽度来源
5. 为表格列宽计算建立第一版缓存
6. 为 Help / Footer 切换到统一动作投影预留接口

## 后续实现清单

1. 重构 `RenderApp()` 为固定 rail 模式
2. 为 `message` 与 `query` 引入统一 rail renderer
3. 抽象 `List Page` 与 `Split Page` 模板
4. 为 `RenderTable()` 引入宽度缓存接口
5. 将日志页换行宽度切到 panel body width
6. 将 Help 页从静态 `shortcuts.jsonc` 迁移到动作投影层
7. 收敛 panel title / summary / selection preview 的统一接口

## 后续考虑清单

1. 是否在超小终端下退化为简化骨架，例如隐藏 message rail 或压缩 footer 行数
2. 是否允许用户通过配置调整 rail 高度
3. 是否对高频 stats 页做帧率节流或脏区刷新
4. 表格宽度缓存是否需要按 runtime / locale 参与 key
5. Help 页是否完全动态生成，还是保留分组模板 + 动态动作填充

## 风险点

1. 若直接在现有页面上继续局部打补丁，页面模板差异会继续扩大。
2. 若先做视觉样式微调而不先固定骨架，布局跳变问题不会消失。
3. 若表格缓存做得过深、过早，可能抬高状态同步复杂度；第一阶段应先缓存宽度与表头，不缓存整表字符串。
4. 若 Help / Footer 不切到同一动作系统，页面优化完成后仍会保留说明漂移问题。
