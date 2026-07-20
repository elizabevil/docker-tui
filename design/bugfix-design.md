# dtui 历史 Bug 修复设计

> 建立日期: 2026-07-20
> 作用: 为历史 bug 修复建立统一设计约束、分层策略和变更记录方式。

本文不替代 [current-design.md](current-design.md)。

- `current-design.md` 负责当前 UI / 交互 / 视觉设计事实。
- 本文负责历史 bug 修复时的设计边界、实现路径和验收关注点。

## 设计目标

- 让“历史问题修复”成为一条可持续维护的工作流，而不是零散改动。
- 修复时优先保证行为一致性，再处理文档、帮助页和默认配置漂移。
- 避免把“新功能规划”伪装成“bug 修复”。

## 基本原则

### 1. 代码优先验证

- 任何修复需求都必须能落到具体代码锚点。
- 如果文档与代码冲突，以代码真实行为为起点做设计。
- 如果问题来自“旧设计已失效”，先收敛现状，再决定是否回滚到旧设计。

### 2. 分层修复

修复方案按这四层拆解，避免一次改动把所有层混在一起：

1. 数据层: `internal/data/*`
2. 状态与交互层: `internal/tui/state/*`、`internal/tui/keyboard/*`、`internal/tui/update*.go`
3. 渲染层: `internal/tui/ui/*`
4. 文档与默认配置: `docs/*`、`design/*`、`internal/data/config/default.jsonc`

### 3. 小步闭环

- 每个 bug 优先争取“一次提交解决一个编号”。
- 若问题跨度过大，拆成 `BR-xxx-A`、`BR-xxx-B` 这样的实现子任务，但外部仍维持一个主编号。
- 每个修复至少补一种验证:
  单测、手动复现步骤、或最小回归说明三者之一。

### 4. 默认值与帮助信息必须一起改

历史问题里，很多偏差不在主逻辑本身，而在这些外围面：

- 默认配置
- 帮助页
- footer 快捷键提示
- README / 设计文档

只改业务逻辑、不改外围说明，会继续制造新的历史 bug。

## 修复流程

1. 在 [../docs/bugfix-requirements.md](../docs/bugfix-requirements.md) 新增或更新问题条目。
2. 判断问题属于哪一层，先写最小设计决策。
3. 实施代码修复。
4. 补验证和文档同步。
5. 将需求状态从 `open` 推进到 `done`。

## 当前设计决策

### BR-001 镜像详情历史数据接入

- 目标:
  让详情页 `History` 分区展示真实数据，并成为镜像详情中的一级能力，而不是长期占位文本。
- 设计边界:
  当前阶段优先确定宏观架构与功能方向，不提前锁死最终字段、交互或失败降级文案。
- 现状拆解:
  1. 数据层当前只有 `InspectImage(id) string` 这一条纯文本输出链路。
  2. 详情页当前通过 `buildImageDetailSections(content)` 反向解析文本，再组装区段。
  3. `History` 区段已经预留，但没有任何真实数据来源，最终固定回退到 `inspect.no_history`。
  4. 现有镜像列表侧已经具备部分 manifest 基础数据，例如 `ImageSummary.IsManifest` 与 `ImageSummary.Manifests`，但详情链路尚未消费。
- 已确认的方向:
  1. `History` 作为父标题保留，具体语义由子标题区分。
  2. 普通镜像与 manifest list 允许使用不同内容模型，不强行共享同一套 history 语义。
  3. 普通镜像的 `History` 面向 layer/build history；manifest list 的 `History` 面向平台变体结构，而不是伪造 layer history。
  4. 镜像详情页后续应从“字符串拼接 + 文本反解析”逐步演进为“结构化 section model + 按类型渲染”。
  5. BR-001 可以接受针对镜像详情链路的重构，不以最小改动为首要目标；目标是给后续需求铺出更稳的架构。
  6. 第一阶段 `History` 定位为只读结构视图，不承担深入交互能力。
- 宏观架构方向:
  1. 运行时数据获取层: 吸收 Docker / Podman 的原始历史与 manifest 数据差异。
  2. 领域归一化层: 生成镜像详情统一模型，并显式区分普通镜像与 manifest list 的 history section。
  3. 渲染层: 基于 section type 进行排版，而不是依赖文本 parser 猜测区段含义。
- Section Model 草案:
  1. 详情页顶层模型建议从单一 `string` 演进为:
     - `ImageDetailViewModel`
     - `[]ImageDetailSection`
  2. `ImageDetailViewModel` 第一版职责:
     - 承载镜像基础元信息
     - 承载结构化 section 列表
     - 标记当前镜像是否为 manifest list
     - 承载 history 数据来源状态
  3. `ImageDetailSection` 第一版建议字段:
     - `key`
     - `title`
     - `subtitle`
     - `kind`
     - `items`
     - `empty_state`
  4. `kind` 第一版建议枚举:
     - `kv_list`
     - `bullet_list`
     - `history_layers`
     - `manifest_variants`
     - `notice`
  5. `items` 建议使用统一条目模型承接，而不是直接拼文本:
     - `label`
     - `value`
     - `dimmed`
     - `children`
  6. `empty_state` 只表达当前区段为何无内容，例如:
     - `none`
     - `not_available`
     - `not_requested`
     - `unsupported`
- `History` 区段草案:
  1. 父标题固定为 `History`。
  2. 普通镜像:
     - `subtitle` 建议语义为 `Layers` 或 `Build History`
     - `kind = history_layers`
  3. manifest list:
     - `subtitle` 建议语义为 `Variants` 或 `Platform Variants`
     - `kind = manifest_variants`
  4. 不再要求普通镜像与 manifest list 共享同一条目结构。
- 普通镜像 `History` 第一版字段范围:
  1. 每条 history 记录至少支持:
     - `created_by`
     - `size`
     - `comment`
  2. 可选字段:
     - `created`
     - `empty_layer`
     - `tags`
  3. UI 第一版按“单层一条”垂直渲染，不引入折叠或树形交互。
  4. 若 `created_by` 过长，允许 UI 截断显示，但数据模型中保留完整值。
- manifest `History` 第一版字段范围:
  1. 每条变体记录至少支持:
     - `digest`
     - `os`
     - `architecture`
     - `variant`
     - `size`
     - `available`
  2. 第一版可直接复用现有 `ImageManifestEntry` 归一化后投影到 section item。
  3. manifest 场景不伪造 `created_by` 或 layer 序列。
- 数据归一化草案:
  1. 建议新增镜像详情专用领域模型，而不是继续复用 `InspectImage() string`。
  2. 第一阶段可以保留旧文本详情作为 fallback，但 `History` 区段应优先走结构化模型。
  3. 列表页已有的 `ImageSummary.IsManifest / Manifests` 可作为 manifest `History` 的第一版输入。
  4. 普通镜像 history 需要新增单独数据获取路径，不应从 `inspect` 文本反解析。
- Go Struct 草案:
  1. `ImageDetailViewModel`
     ```go
     type ImageDetailViewModel struct {
         ImageID         string
         RepoTags        []string
         RepoDigests     []string
         Registry        string
         Name            string
         Tag             string
         Created         string
         SizeBytes       int64
         Architecture    string
         OS              string
         OSVersion       string
         IsManifest      bool
         HistorySource   ImageHistorySource
         Sections        []ImageDetailSection
     }
     ```
  2. `ImageDetailSection`
     ```go
     type ImageDetailSection struct {
         Key        string
         Title      string
         Subtitle   string
         Kind       ImageDetailSectionKind
         Items      []ImageDetailItem
         EmptyState SectionEmptyState
     }
     ```
  3. `ImageDetailSectionKind`
     ```go
     type ImageDetailSectionKind string

     const (
         SectionKVList          ImageDetailSectionKind = "kv_list"
         SectionBulletList      ImageDetailSectionKind = "bullet_list"
         SectionHistoryLayers   ImageDetailSectionKind = "history_layers"
         SectionManifestVariant ImageDetailSectionKind = "manifest_variants"
         SectionNotice          ImageDetailSectionKind = "notice"
     )
     ```
  4. `SectionEmptyState`
     ```go
     type SectionEmptyState string

     const (
         SectionEmptyNone         SectionEmptyState = "none"
         SectionEmptyNotAvailable SectionEmptyState = "not_available"
         SectionEmptyNotRequested SectionEmptyState = "not_requested"
         SectionEmptyUnsupported  SectionEmptyState = "unsupported"
     )
     ```
  5. `ImageHistorySource`
     ```go
     type ImageHistorySource string

     const (
         HistorySourceNone      ImageHistorySource = "none"
         HistorySourceLayerAPI  ImageHistorySource = "layer_api"
         HistorySourceManifest  ImageHistorySource = "manifest"
         HistorySourceFallback  ImageHistorySource = "fallback"
     )
     ```
  6. `ImageDetailItem`
     ```go
     type ImageDetailItem struct {
         Label    string
         Value    string
         Dimmed   bool
         Children []ImageDetailItem
     }
     ```
  7. 普通镜像 history 条目:
     ```go
     type ImageHistoryLayerItem struct {
         Created   string
         CreatedBy string
         SizeBytes int64
         Comment   string
         EmptyLayer bool
         Tags      []string
     }
     ```
  8. manifest variant 条目:
     ```go
     type ImageManifestVariantItem struct {
         Digest       string
         OS           string
         Architecture string
         Variant      string
         SizeBytes    int64
         Available    bool
     }
     ```
  9. 原始详情聚合体:
     ```go
     type ImageDetailData struct {
         Inspect   ImageInspectData
         History   []ImageHistoryLayerItem
         Variants  []ImageManifestVariantItem
     }
     ```
  10. `ImageInspectData`
      ```go
      type ImageInspectData struct {
          ID           string
          RepoTags     []string
          RepoDigests  []string
          Created      string
          SizeBytes    int64
          Architecture string
          OS           string
          OSVersion    string
          Author       string
          Comment      string
          WorkingDir   string
          User         string
          StopSignal   string
          Entrypoint   []string
          Cmd          []string
          Shell        []string
          OnBuild      []string
          ExposedPorts []string
          Env          []string
          Volumes      []string
          Healthcheck  []string
          StorageDriver string
          LayerCount   int
          Labels       map[string]string
      }
      ```
- Builder 责任划分:
  1. `ImageDetailLoader`
     - 只负责从 runtime 拉取原始详情数据
     - 聚合 inspect、history、manifest 变体
     - 不做 UI 文案拼接
  2. `ImageDetailNormalizer`
     - 把 runtime 原始数据归一化为 `ImageDetailData`
     - 屏蔽 Docker / Podman 字段差异
     - 负责“缺字段”“无能力”“无数据”状态判定
  3. `ImageDetailSectionBuilder`
     - 从 `ImageDetailData` 生成 `ImageDetailViewModel`
     - 组装 `basic/system/config/storage/tags/metadata/history` 各 section
     - 决定 `History` 使用 `history_layers` 还是 `manifest_variants`
  4. `ImageDetailRenderer`
     - 只消费 `ImageDetailViewModel`
     - 按 `section.kind` 选择渲染分支
     - 不再反向解析文本
- Go 接口草案:
  1. `ImageDetailLoader`
     ```go
     type ImageDetailLoader interface {
         LoadImageDetail(context.Context, string) (ImageDetailData, error)
     }
     ```
  2. `ImageDetailSectionBuilder`
     ```go
     type ImageDetailSectionBuilder interface {
         Build(ImageDetailData) ImageDetailViewModel
     }
     ```
  3. 可选拆分接口:
     ```go
     type ImageHistorySectionBuilder interface {
         BuildHistorySection(ImageDetailData) ImageDetailSection
     }

     type ImageBaseSectionBuilder interface {
         BuildBaseSections(ImageDetailData) []ImageDetailSection
     }
     ```
- Builder 输出边界:
  1. Loader 输出的是领域原始详情，不包含 i18n 标题。
  2. Section Builder 才负责 section 标题、subtitle、empty state 与 item 组织方式。
  3. Renderer 只负责样式和排版，不负责字段选择。
  4. 若后续要支持 JSON 导出或测试快照，优先复用 `ImageDetailViewModel`，而不是渲染后的字符串。
- Builder 拆分策略:
  1. 第一阶段可以先用一个总 builder:
     - `DefaultImageDetailSectionBuilder`
  2. 其内部再拆为:
     - `buildBasicSection`
     - `buildSystemSection`
     - `buildConfigSection`
     - `buildStorageSection`
     - `buildTagsSection`
     - `buildMetadataSection`
     - `buildHistorySection`
  3. `buildHistorySection` 内部再按:
     - 普通镜像 -> `buildLayerHistorySection`
     - manifest list -> `buildManifestVariantSection`
  4. 先拆责任，再决定是否抽成公开接口，避免第一版过度设计。
- 渲染层草案:
  1. `detail/view.go` 继续负责通用 section 渲染骨架。
  2. `detail/image.go` 从“文本解析器”演进为“结构化 section 组装器”。
  3. `history_layers` 与 `manifest_variants` 允许使用独立渲染分支，但保持同一父区段视觉层级。
  4. 当结构化模型可用时，优先走结构化 section；仅在旧资源类型上保留文本 section fallback。
- 建议路径:
  1. 为镜像详情新增结构化领域模型与 section builder。
  2. 将当前基础区段 `basic/system/config/storage/tags/metadata` 逐步改为结构化输出。
  3. 为普通镜像接入真实 history 数据源。
  4. 为 manifest list 接入平台变体 section，并复用现有 manifest 数据。
  5. 将 `History` 区段从固定占位文案改为真实 section 投影。
  6. 最后再清理仅为图片详情服务的文本反解析路径。
- 后续实现清单:
  1. 定义 `ImageDetailViewModel` 与 `ImageDetailSection` 草案结构。
  2. 定义 `ImageDetailLoader` 与默认 loader 实现。
  3. 定义 `DefaultImageDetailSectionBuilder` 与 `buildHistorySection` 分支。
  4. 为镜像详情新增结构化 builder，而不是只返回 `string`。
  5. 为普通镜像接入 history 数据获取与归一化。
  6. 为 manifest list 接入 `manifest_variants` section builder。
  7. 将 `detail/image.go` 改为优先消费结构化模型。
  8. 为 `History` 空状态区分“无数据”“能力缺失”“实现未接线”三类来源。
  9. 补最小回归验证: 普通镜像至少一条 history、manifest list 至少一条 variant、无数据场景一条明确空状态。
- 后续考虑清单:
  1. `subtitle` 最终是否使用 `Layers / Build History / Variants`，还是统一走 i18n 资源键。
  2. 是否把镜像详情的所有区段都迁移为 typed section item，而不仅是 `History`。
  3. 是否为长 `created_by`、digest、comment 增加 detail 内部换行或摘要显示策略。
  4. 是否需要在详情页显式显示 `History` 数据来源，例如 engine / runtime 能力差异。
  5. `ImageInspectData` 是否继续保持扁平 struct，还是后续再按 `config/storage/labels` 子 struct 收敛。
- 当前不做的决定:
  1. 暂不锁定具体 section 字段清单。
  2. 暂不锁定 history 拉取失败后的降级文案和失败策略。
  3. 暂不展开具体代码改造路径和函数签名。
- 风险点:
  1. 若继续沿用纯文本 detail 链路，后续普通镜像 / manifest 差异化排版会越来越脆弱。
  2. 若一次性把所有资源详情页都纳入重构，范围会明显失控；本轮应收敛在镜像详情链路。
  3. 若仍尝试从 `InspectImage()` 文本反推 history，字段稳定性和跨 runtime 兼容性都会很差。

### BR-002 自定义键位配置接线

- 目标:
  让 `config.Keymap` 成为真实运行时输入，而不是静态结构。
- 设计边界:
  第一阶段允许重构输入层，但范围收敛在“统一动作模型 + 上下文模型 + 配置接线”，不追求一次做完整多绑定冲突检测器或完整用户级上下文覆盖能力。
- 已确认的方向:
  1. 键位系统改为“动作驱动”，不是“按键驱动”。
  2. 键位绑定默认属于动作本身，而不是页面本身。
  3. 页面只声明“当前上下文有哪些动作可用”，不拥有独立键位定义权。
  4. Help / Footer / 默认配置都应从同一动作与绑定系统派生。
  5. 输入框编辑命令域独立于业务动作域，`mode` 作为上下文存在，不作为顶层动作命名空间。
- 上下文模型方向:
  1. 输入上下文固定为四层:
     - `App Context`
     - `Surface Context`
     - `View Context`
     - `Mode Context`
  2. 解析优先级固定为:
     `Mode > View > Surface > App`
  3. `Mode Context` 负责特殊输入模式截获，例如:
     - `filter`
     - `search`
     - `command`
     - `confirm`
     - `mark`
     - `exec_dialog`
     - `exec_passthrough`
  4. `View Context` 负责承接交互结构变化，是双栏、子视图、钻取结构的关键层。
  5. `Surface Context` 提供页面级默认动作集合。
  6. `App Context` 只提供全局兜底动作。
- 当前建议的 `Surface Context` 清单:
  - `containers`
  - `images`
  - `volumes`
  - `networks`
  - `compose`
  - `logs`
  - `detail`
  - `help`
- 当前建议的 `View Context` 清单:
  - `containers.list`
  - `images.list`
  - `images.containers_subview`
  - `volumes.list`
  - `volumes.detail_subview`
  - `networks.list`
  - `compose.projects`
  - `compose.services`
  - `compose.containers_subview`
  - `logs.main`
  - `detail.main`
  - `help.main`
- 动作解析规则:
  1. `Mode Context` 优先截获输入。
  2. `View Context` 决定当前焦点结构下的具体行为。
  3. `Surface Context` 提供页面默认动作集合。
  4. `App Context` 提供全局动作兜底。
  5. 动作解析优先于按键意义，内部优先使用语义动作而不是直接依赖物理方向键。
  6. `View` 可屏蔽 `Surface`，`Mode` 可屏蔽全部。
- 动作命名空间方向:
  1. 优先格式:
     - `domain.object.verb`
  2. 没有明确对象时:
     - `domain.verb`
  3. 顶层 domain 固定为:
     - `app`
     - `nav`
     - `panel`
     - `resource`
     - `logs`
     - `detail`
  4. `resource.*` 成为所有业务资源操作的统一归属。
  5. `nav.forward/back` 逐步取代直接依赖 `left/right` 的语义表达。
- 当前建议的动作簇:
  1. `app.*`
     - `app.help.open`
     - `app.help.close`
     - `app.quit`
     - `app.refresh`
     - `app.filter.open`
     - `app.command.open`
  2. `nav.*`
     - `nav.up`
     - `nav.down`
     - `nav.forward`
     - `nav.back`
     - `nav.enter`
     - `nav.page_up`
     - `nav.page_down`
     - `nav.top`
     - `nav.bottom`
     - `nav.panel_next`
     - `nav.panel_prev`
  3. `panel.*`
     - `panel.header.toggle`
     - `panel.connection.show`
     - `panel.runtime.switch`
     - `panel.sort.next`
     - `panel.sort.order_toggle`
     - `panel.mark.enter`
     - `panel.mark.toggle`
     - `panel.mark.clear`
  4. `resource.container.*`
     - `resource.container.start`
     - `resource.container.stop`
     - `resource.container.restart`
     - `resource.container.kill`
     - `resource.container.delete`
     - `resource.container.logs`
     - `resource.container.exec`
     - `resource.container.inspect`
     - `resource.container.stats_toggle`
  5. `resource.image.*`
     - `resource.image.pull`
     - `resource.image.prune`
     - `resource.image.delete`
     - `resource.image.detail`
     - `resource.image.debug`
     - `resource.image.export`
     - `resource.image.copy_ref`
     - `resource.image.expand_containers`
  6. `resource.volume.*`
     - `resource.volume.detail`
     - `resource.volume.delete`
     - `resource.volume.expand`
  7. `resource.network.*`
     - `resource.network.detail`
     - `resource.network.delete`
  8. `resource.compose_project.*`
     - `resource.compose_project.start`
     - `resource.compose_project.stop`
     - `resource.compose_project.logs`
     - `resource.compose_project.down`
     - `resource.compose_project.detail`
  9. `logs.*`
     - `logs.search.open`
     - `logs.search.next`
     - `logs.search.prev`
     - `logs.wrap.toggle`
  10. `detail.*`
      - `detail.top`
      - `detail.page_up`
      - `detail.page_down`
- 命名收敛规则:
  1. 动作名表达语义，不表达按键。
  2. 动作名必须带 domain。
  3. 资源动作必须带资源类型。
  4. 统一动词词表，优先使用:
     - `open`
     - `close`
     - `toggle`
     - `show`
     - `switch`
     - `next`
     - `prev`
     - `forward`
     - `back`
     - `enter`
     - `expand`
     - `collapse`
     - `start`
     - `stop`
     - `restart`
     - `kill`
     - `delete`
     - `pull`
     - `prune`
     - `inspect`
     - `detail`
     - `logs`
     - `exec`
     - `copy_ref`
     - `down`
  5. 避免继续保留扁平旧名，例如:
     - `detail`
     - `delete`
     - `back`
     - `refresh`
- 建议路径:
  1. 新增“从 `config.Keymap` 构建 KeyMapping”的转换层。
  2. 引入上下文模型和动作注册表，逐步替换散落的按键分支。
  3. 保留默认映射作为 fallback，但默认映射本身应迁移到统一动作注册表。
  4. 将 Help 与 Footer 的展示逻辑切换到同一份有效动作/绑定来源。
- UI 投影与展示约束:
  1. Footer 展示“当前上下文全部可用动作”，不再做手工筛选。
  2. Footer 第一组固定展示全局动作，其余动作按当前页面上下文平铺展示。
  3. Help、Footer、默认配置导出可以使用不同展示形式，但事实来源必须是同一动作注册表与有效绑定集。
  4. “全局动作”按跨页面稳定语义定义，例如统一的上下移动、切换、滚动、返回、进入；不能按某个页面是否恰好复用同一物理按键来定义。
- 后续实现清单:
  1. 定义 `ActionSpec`、`Binding`、`Context`、`Resolver` 等核心对象。
  2. 将默认键位从散落常量迁移到统一动作注册表。
  3. 新增 `config.Keymap` 到运行时绑定覆盖的编译层。
  4. 为输入系统实现四层上下文解析器: `App / Surface / View / Mode`。
  5. 将输入框编辑命令从业务动作解析链中独立出来。
  6. 为 `Left / Right / Enter / Esc` 等高频导航键接入语义动作解析，而不是继续页面硬编码。
  7. 让 Help 与 Footer 基于“当前有效动作集合”动态投影。
  8. 第一阶段只开放“默认绑定覆盖”，不立即开放用户自定义上下文覆盖规则。
- 后续考虑清单:
  1. 是否支持一个动作绑定多个按键，以及冲突检测与告警策略。
  2. 是否支持 `shell-like` 之外的 `vi-like` 输入编辑 profile。
  3. 是否支持用户按上下文覆盖键位，而不仅是覆盖默认绑定。
  4. 是否导出“当前生效键位表”供帮助页、调试和配置回写使用。
  5. 是否保留旧动作名或旧默认键位的兼容别名，以及兼容期限。
  6. 是否将帮助页完全改为动态生成，并如何处理 i18n 描述文本。
- 风险点:
  1. 若只做配置接线而不做上下文/动作收敛，最终仍会保留“部分可配置、部分硬编码”的分裂状态。
  2. 展示层目前大量使用静态 key 文案，接线后需要同步重构帮助页与 Footer 数据源。
  3. 若过早把上下文覆盖能力开放给用户配置，会显著抬高输入系统复杂度。

### BR-003 默认键位示例统一

- 目标:
  消除默认配置、帮助页、运行时行为三者之间的冲突。
- 设计边界:
  这是修复一致性问题，不是新增快捷键功能。
- 已确认的方向:
  1. `BR-003` 的本质不是“帮助页改文案”，而是建立键位说明的统一信息源。
  2. 必须统一的事实只有三类:
     - 动作是什么
     - 默认绑定是什么
     - 在哪些上下文可用
  3. Help、Footer、README、默认配置可以有不同展示形式，但不能各自维护独立键位真相。
  4. 页面只声明“当前上下文有哪些动作可用”，不应再单独拥有自己的键位事实源。
  5. 默认配置的 `keymap` 只是统一动作注册表默认绑定的导出视图，而不是独立事实源。
- 展示约束:
  1. Footer 展示当前上下文完整动作集，第一组固定为全局动作。
  2. Help 可以比 Footer 更详细、更分组，但动作与键位必须从同一注册表派生。
  3. README 和其他摘要文档只允许引用导出后的默认键位摘要，不允许再手写一套默认键位表。
  4. “全局动作”按跨页面稳定语义定义，而不是按某个按键是否恰好在所有页面复用定义。
- 建议路径:
  1. 以统一动作注册表和运行时默认绑定为唯一基准。
  2. 将 `default.jsonc` 改为从该基准导出的默认配置视图。
  3. 将 Help / Footer 的键位展示切换为同源派生，而不是静态维护。
  4. 回写 README 和相关文档，清除旧大写单键说明。
  5. 如果仍保留旧大写兼容，必须明确兼容范围、优先级和淘汰策略。
- 后续实现清单:
  1. 建立“动作定义 + 默认绑定 + 上下文可用性”的统一注册表。
  2. 让默认配置生成逻辑从统一注册表导出。
  3. 让 Help 数据源改为动作注册表与当前有效绑定集。
  4. 让 Footer 数据源改为当前上下文可用动作集合。
  5. 清理 README、导航文档、默认配置中的旧键位示例。
  6. 为旧兼容键位保留最小测试或代码注释说明。
- 后续考虑清单:
  1. Help 是否完全动态生成，还是保留“静态分组 + 动态键位填充”的折中方案。
  2. 是否提供“导出生效键位表”能力，方便调试与回写配置。
  3. 旧键位兼容期是否需要版本标记与迁移提示。
  4. i18n 文案如何与动作注册表描述字段协同维护。
- 风险点:
  1. 若只回写 `default.jsonc` 而不改 Help / Footer 数据源，文档漂移会继续重复出现。
  2. 若展示层继续手写键位字符串，自定义键位接线后会出现“显示与实际不一致”的新问题。
  3. 若过早把 README 当作键位权威来源，后续变更会反复产生人工同步成本。

### BR-004 过滤模式收敛

- 目标:
  让过滤行为、配置项、实现路径三者一致。
- 设计边界:
  先做“行为定义收敛”，再决定保留防抖还是删除防抖。
- 已确认的方向:
  1. 资源页 `Filter` 与日志页 `Search` 不再被视为同一种模式。
  2. 两者可以共享输入组件，但不能继续共享同一语义控制器。
  3. `/` 保留为统一查询入口键，但进入 `Filter` 还是 `Search` 由当前页面决定。
  4. 结构化资源页使用 `Filter`，连续文本内容页使用 `Search`。
- 模式边界:
  1. `Filter`
     - 面向结构化资源列表
     - 输入即生效
     - 作用是缩小当前资源集合
  2. `Search`
     - 面向日志等连续文本
     - 不改变原始数据集
     - 作用是匹配、定位、跳转
  3. `Command`
     - 保持独立，不并入 `Filter / Search`
- `Filter` 状态模型:
  1. 进入方式:
     - 仅通过 `/` 进入 `Filter` 模式
  2. 活跃过滤态:
     - 输入即更新筛选结果
     - 当前筛选词与筛选结果同步生效
  3. 第一次 `Esc`:
     - 不清空筛选词
     - 不恢复默认资源列表
     - 进入 5 秒待退出窗口
     - 输入框仍可继续输入
  4. 5 秒内第二次 `Esc`:
     - 清空筛选条件
     - 恢复默认资源列表
     - 退出 `Filter` 功能本身
  5. 超时:
     - 超过 5 秒未再次按 `Esc`，待退出状态失效
     - 当前筛选继续保持
  6. 退出后:
     - 只有再次按 `/` 才重新进入 `Filter` 模式
- 可见反馈约束:
  1. `Filter` 第一次 `Esc` 的待退出提示进入统一消息通知系统。
  2. 不为 `Filter` 退出提示新增独立展示通道。
  3. `Search` 的匹配状态与跳转反馈也应复用统一消息体系，而不是散落到临时字段。
- 建议路径:
  1. 拆分 `Filter` 与 `Search` 的模式状态和控制器。
  2. 保留共享输入组件，但将“输入编辑”与“模式语义”分层。
  3. 资源页保留即时过滤，并删除 `SearchTimer / SearchTick / StartSearchDebounce()` 等旧防抖路径。
  4. 日志页搜索保留“编辑后应用匹配”的语义，不参与数据集过滤。
  5. 将第一次 `Esc` 的待退出窗口接入统一消息通知系统。
  6. 重新定义或删除 `ui.searchDebounceMs`，避免保留死配置。
- 后续实现清单:
  1. 为 `Filter` 与 `Search` 建立独立模式状态结构。
  2. 新增 `Filter` 的待退出窗口状态与 5 秒计时逻辑。
  3. 将资源页即时过滤逻辑从旧防抖链路中剥离。
  4. 将日志页搜索保留为匹配/跳转控制器。
  5. 为 `Esc` 第一次提示接入统一通知事件。
  6. 清理无效 `SearchTick` 路径与相关配置说明。
  7. 更新帮助页、文档和状态提示文案，明确 `Filter / Search` 语义差异。
- 后续考虑清单:
  1. `ui.searchDebounceMs` 是彻底移除、兼容迁移，还是重新定义为仅对某类输入生效。
  2. `Filter` 是否需要支持大小写、正则、字段过滤等高级能力。
  3. `Search` 是否需要支持上一个/下一个匹配、命中计数、循环跳转。
  4. `Filter` 与 `Search` 是否需要在状态栏中提供更显式的模式标记。
- 风险点:
  1. 该问题同时影响交互体验、配置兼容和文档说明，容易出现局部修好、整体继续漂移。
  2. 若继续复用同一模式状态，后续 `Esc` 状态机和日志搜索语义会持续互相污染。
  3. 若删除旧防抖路径但不处理配置迁移，维护者仍会被 `ui.searchDebounceMs` 误导。

### BR-005 消息 / 审计日志管理

- 目标:
  建立统一的用户业务操作审计模型，让通知、操作历史和落盘日志共享同一事件来源，并可通过 `trace_id` 追溯完整操作链。
- 设计边界:
  当前阶段优先把“用户操作审计”设计正确，不把程序内部运行细节混入审计日志；`runtime/debug` 日志可后续单独完善。
- 已确认的方向:
  1. 当前落盘主日志优先定位为审计日志，而不是程序运行日志。
  2. `Notification` 与 `Operation Log` 不是平行系统，而是同一业务操作链的不同视图。
  3. `trace_id` 只服务用户业务操作链，不覆盖普通查询、搜索、导航等轻交互。
  4. 资源对象最小统一字段固定为 `type/id/name`。
  5. 资源对象扩展字段使用 `struct + interface` 强类型约束，不使用松散 map。
  6. 应用必须支持落盘审计日志，便于操作历史查看与追溯。
- 事件分层方向:
  1. `Audit Event`: 用户业务操作事件，进入 trace 链，支持落盘审计。
  2. `UI Message Event`: 界面提示事件，只服务通知与界面反馈，不进入审计日志。
- 审计记录方向:
  1. `Action` 使用统一动作命名空间，例如 `resource.container.stop`。
  2. `Result` 表示当前记录所属事件的阶段状态，而不是资源最终业务状态。
  3. 当前建议的 `Result` 枚举为:
     - `requested`
     - `started`
     - `succeeded`
     - `failed`
     - `cancelled`
  4. 一次业务操作可以产生多条相同 `Action`、相同 `TraceID`、但 `Result` 不同的记录。
- 建议的数据模型:
  1. `AuditRecord`
     - `time`
     - `trace_id`
     - `event_id`
     - `action`
     - `result`
     - `level`
     - `message`
     - `runtime`
     - `ui`
     - `target`
     - `details`
  2. `RuntimeContext`
     - `type`
     - `name`
     - `host`
  3. `UIContext`
     - `surface`
     - `view`
     - `mode`
  4. `AuditTarget`
     - 统一接口，输出 `AuditTargetDTO`
  5. `AuditTargetDTO`
     - `type`
     - `id`
     - `name`
     - `meta`
  6. `Details`
     - 仅承载补充信息，例如 `duration_ms`、`error`、`shell`、`exit_code`
- 当前已收敛的字段清单:
  1. `AuditRecord` 第一版固定字段为:
     - `time`
     - `trace_id`
     - `event_id`
     - `action`
     - `result`
     - `level`
     - `message`
     - `runtime`
     - `ui`
     - `target`
     - `details`
  2. 当前阶段不加入:
     - `actor`
     - `session_id`
     - `app_version`
     - `hostname`
     - 其他外围审计增强字段
  3. `RuntimeContext` 第一版固定字段为:
     - `type`
     - `name`
     - `host`
  4. `UIContext` 第一版固定字段为:
     - `surface`
     - `view`
     - `mode`
  5. `AuditTargetDTO` 第一版固定字段为:
     - `type`
     - `id`
     - `name`
     - `meta`
  6. `Details` 第一版固定字段为:
     - `duration_ms`
     - `error`
     - `shell`
     - `exit_code`
- 审计目标方向:
  1. 强类型接口建议为:
     `TargetType() / TargetID() / TargetName() / ToDTO()`
  2. `meta` 来源于各资源自己的强类型 struct 序列化结果。
  3. 没有天然独立 ID 的对象，第一阶段允许 `id == name` 作为回退。
- Go 结构体 / 接口草案:
  1. `AuditRecord`
     ```go
     type AuditRecord struct {
         Time    time.Time      `json:"time"`
         TraceID string         `json:"trace_id"`
         EventID string         `json:"event_id"`
         Action  string         `json:"action"`
         Result  AuditResult    `json:"result"`
         Level   AuditLevel     `json:"level"`
         Message string         `json:"message"`
         Runtime RuntimeContext `json:"runtime"`
         UI      UIContext      `json:"ui"`
         Target  AuditTargetDTO `json:"target"`
         Details AuditDetails   `json:"details,omitempty"`
     }
     ```
  2. `AuditResult`
     ```go
     type AuditResult string

     const (
         AuditRequested AuditResult = "requested"
         AuditStarted   AuditResult = "started"
         AuditSucceeded AuditResult = "succeeded"
         AuditFailed    AuditResult = "failed"
         AuditCancelled AuditResult = "cancelled"
     )
     ```
  3. `AuditLevel`
     ```go
     type AuditLevel string

     const (
         AuditInfo  AuditLevel = "info"
         AuditWarn  AuditLevel = "warn"
         AuditError AuditLevel = "error"
     )
     ```
  4. `RuntimeContext`
     ```go
     type RuntimeContext struct {
         Type string `json:"type"`
         Name string `json:"name"`
         Host string `json:"host"`
     }
     ```
  5. `UIContext`
     ```go
     type UIContext struct {
         Surface string `json:"surface,omitempty"`
         View    string `json:"view,omitempty"`
         Mode    string `json:"mode,omitempty"`
     }
     ```
  6. `AuditTarget`
     ```go
     type AuditTarget interface {
         TargetType() string
         TargetID() string
         TargetName() string
         ToDTO() AuditTargetDTO
     }
     ```
  7. `AuditTargetDTO`
     ```go
     type AuditTargetDTO struct {
         Type string          `json:"type"`
         ID   string          `json:"id"`
         Name string          `json:"name"`
         Meta json.RawMessage `json:"meta,omitempty"`
     }
     ```
  8. `AuditDetails`
     ```go
     type AuditDetails struct {
         DurationMs int64  `json:"duration_ms,omitempty"`
         Error      string `json:"error,omitempty"`
         Shell      string `json:"shell,omitempty"`
         ExitCode   *int   `json:"exit_code,omitempty"`
     }
     ```
  9. `AuditSink`
     ```go
     type AuditSink interface {
         WriteAudit(context.Context, AuditRecord) error
     }
     ```
- UI 投影层方向:
  1. `NotificationProjector`
     - 顶部短时通知
     - 偏结果型反馈
     - 可消费 `Audit Event` 与 `UI Message Event`
  2. `OperationLogProjector`
     - Footer / 后续操作历史面板
     - 偏过程型回看
     - 主要消费 `Audit Event`
  3. `StatusProjector`
     - 状态栏 / 页面状态摘要
     - 面向当前状态，不面向历史链路
- Go 接口草案:
  1. `NotificationProjector`
     ```go
     type NotificationProjector interface {
         OnAuditRecord(AuditRecord)
         OnUIMessage(UIMessageEvent)
     }
     ```
  2. `OperationLogProjector`
     ```go
     type OperationLogProjector interface {
         OnAuditRecord(AuditRecord)
         Recent() []AuditRecord
         Current() *AuditRecord
     }
     ```
  3. `StatusProjector`
     ```go
     type StatusProjector interface {
         CurrentStatus() string
     }
     ```
- 当前已收敛的投影规则:
  1. `NotificationProjector`
     - 接收:
       - `Audit Event` 中的 `succeeded` / `failed` / `cancelled`
       - `UI Message Event` 中的 warning / info 提示
     - 默认忽略:
       - `requested`
       - `started`
     - 默认直接使用 `message` 作为通知主文案
  2. `OperationLogProjector`
     - 内存历史接收全部 `AuditRecord`
     - Footer 当前行只显示最近一条高价值事件
     - Footer 优先级:
       - `failed`
       - `cancelled`
       - `succeeded`
       - `started`
       - `requested`
     - 第一版不做 trace 聚合折叠
     - 建议保留最近 `100` 条历史
  3. `StatusProjector`
     - 当前阶段继续聚焦“当前状态摘要”
     - 不承担事件历史展示职责
  4. `NotificationProjector` 第一版目标:
     - 只维护一条 active notification 和有界队列
     - 不做 trace 级聚合
     - 不做复杂模板系统
  5. `OperationLogProjector` 第一版目标:
     - 内存历史保留完整事件流
     - Footer 只投影最近一条高价值事件
     - 不做折叠、不做分页、不做文件回放
- 落盘策略方向:
  1. 审计日志按 `类型-日期` 命名:
     - `audit-2026-07-20.jsonl`
  2. 审计日志只记录用户显式触发的业务动作，不混入:
     - 打开文件
     - 日志轮转
     - 后台刷新
     - UI 渲染
     - 自动重连
  3. `runtime` 日志可后续单独设计为 `类型-日期-级别`
- 第一版落地范围:
  1. `resource.container.start`
  2. `resource.container.stop`
  3. `resource.container.restart`
  4. `resource.container.kill`
  5. `resource.container.delete`
  6. `resource.image.pull`
  7. `resource.image.prune`
  8. `resource.image.delete`
  9. `resource.compose_project.start`
  10. `resource.compose_project.stop`
  11. `resource.compose_project.down`
  12. `panel.runtime.switch`
  13. `resource.container.exec` 的 session start / finish / fail
- 后续实现清单:
  1. 新增 `AuditRecord`、`AuditTarget`、`AuditTargetDTO`、`RuntimeContext`、`UIContext`、`Details` 结构定义
  2. 新增 `AuditResult` 与 `AuditLevel` 枚举
  3. 为 container / image / volume / network / compose / runtime / exec_session 建立强类型 target struct
  4. 新增统一 `trace_id` / `event_id` 生成器
  5. 新增 `AuditSink` 接口与 `FileAuditSink`
  6. 新增审计日志文件写入与按日命名策略
  7. 新增 `NotificationProjector`
  8. 新增 `OperationLogProjector`
  9. 将顶部通知从 `ToastMessage/ToastTimer` 逐步切换到投影结果
  10. 将底部操作历史从 `InfoMessage/ErrorMessage` 逐步切换到投影结果
  11. 为第一版落地动作接入 audit trace 写入
  12. 为非审计型提示补 `UI Message Event` 通道
- 后续考虑清单:
  1. 是否增加 `actor`、`session_id`、`app_version` 等增强审计字段
  2. 是否引入 `runtime` 日志的独立结构和落盘策略
  3. 是否在 UI 中增加完整操作历史面板，而不仅是 Footer 单行
  4. 是否对同一 `trace_id` 做聚合展示或折叠展示
  5. 是否支持审计日志筛选、导出、回放
  6. 是否支持审计日志轮转、压缩与保留策略
  7. `exec` 是否提供可选 debug I/O 记录，并如何处理敏感信息
  8. `AuditTargetDTO.meta` 是否继续保持 `json.RawMessage`，还是未来演进为更明确的 tagged union
  9. 哪些非资源变更动作应例外纳入 trace
- 风险点:
  1. 若继续让 `InfoMessage/ErrorMessage` 直接承担通知与操作历史，后续 trace 和落盘模型会继续割裂。
  2. 若把 `UI Message Event` 与 `Audit Event` 混写入审计日志，会削弱审计价值并污染追溯链路。
  3. 若过早把所有内部运行事件并入审计模型，范围会膨胀并模糊“用户操作日志”的定位。

## 单条设计记录模板

```md
### BR-xxx

- 目标:
- 影响层:
- 不做什么:
- 方案:
  1.
  2.
- 风险:
- 验证:
```
