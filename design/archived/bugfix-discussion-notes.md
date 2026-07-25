# 历史 Bug 修复讨论纪要

> 记录日期: 2026-07-20
> 作用: 单独保存 BR-001 与 BR-002 的阶段性讨论结论，便于后续继续展开细节设计。
> 说明: 本文是讨论纪要，不替代 [bugfix-design.md](bugfix-design.md) 和 [current-design.md](current-design.md)。

## 使用方式

- 本文只记录“已经讨论达成的方向性结论”。
- 如果进入正式设计决策，应回写到 [bugfix-design.md](bugfix-design.md)。
- 如果进入实现验收，应更新 [../docs/bugfix-requirements.md](../docs/bugfix-requirements.md)。

---

## BR-001 镜像详情 `History` 区块

### 已确认的问题本质

- 当前问题不是运行时能力缺失，而是数据层没有把镜像 history 接入详情页。
- Docker SDK 自带 `ImageHistory()` 能力。
- Podman 的兼容接口也支持 image history 路径。
- 现状更接近“功能未接线”，不是“引擎拿不到数据”。

### 已达成的宏观方向

- `History` 保留为镜像详情中的父标题。
- 具体语义由子标题区分，而不是让所有镜像共用同一种 history 解释。
- 普通镜像与 manifest list 允许使用不同内容模型和不同排版。
- 当前阶段优先考虑架构正确性，不以最小改动为目标。
- 可以接受对镜像详情链路做定向重构，为后续需求铺路。

### 语义结论

#### 普通镜像

- `History` 指向镜像内容本身的 layer/build history。
- 不表示:
  - 容器使用历史
  - dtui 内部操作历史
  - tag 改名历史

#### Manifest List

- `History` 不应伪装成单一 layer history。
- 该场景下更合理的语义是“平台变体结构”或“manifest variants”。
- 也就是说:
  - 父标题仍可保留 `History`
  - 子标题和正文模型允许与普通镜像不同

### 页面能力定位

- `History` 不是附属字段，而是镜像详情中的一级区段能力。
- 第一阶段定位为只读结构视图，不承担深入交互能力。
- 当前阶段不要求它变成调试中心、审计页或供应链分析页。

### 详情页架构方向

- 不建议继续沿用“数据层拼大字符串，UI 层再反解析”的单一路线。
- BR-001 应作为镜像详情从纯文本模型向结构化 section model 演进的起点。
- 推荐的长期结构应分三层:
  1. 运行时数据获取层
  2. 领域归一化层
  3. 详情渲染层

### 当前未决定的问题

- `History` 子标题的最终命名
- 普通镜像 history 的最终字段集合
- manifest list 区段的最终字段集合
- history 拉取失败后的降级语义与文案
- 是否在本轮就引入完整结构化详情模型

### 当前阶段原则

- 先定语义和架构，不抢细节实现。
- 先把镜像详情链路定型，不顺手重构所有资源详情页。

---

## BR-002 自定义键位配置接线

### 已确认的问题本质

- BR-002 不是单纯“把 `config.Keymap` 接进 `HandleKeyPress()`”。
- 当前至少存在多套并行来源:
  - 运行时行为映射
  - 特殊页面硬编码分支
  - 帮助页快捷键说明
  - footer 快捷键提示
- 因此它本质上是“输入系统统一化”问题，而不是单点接线问题。

### 已达成的宏观方向

- 键位系统必须改成“动作驱动”，不是“按键驱动”。
- 系统需要一个单一真相源来描述:
  - 动作是什么
  - 默认绑定是什么
  - 在哪些上下文可用
- 用户配置应覆盖这套统一注册表，而不是覆盖散乱硬编码分支。

### 动作域边界

#### 应进入动作域的内容

- 有业务语义的用户意图
- 会改变应用状态的操作
- 资源操作
- 面板 / 模式 / 子视图切换
- 会出现在帮助页或 footer 中的能力

#### 不应进入动作域的内容

- 输入框内普通字符录入
- 行编辑行为本身
- exec passthrough 的原始终端输入
- 低层输入事件本身

### 输入系统分层结论

#### 业务动作域

- 面向“用户想做什么”
- 例如:
  - `resource.image.pull`
  - `resource.container.stop`
  - `nav.forward`
  - `view.sort_next`

#### 编辑命令域

- 面向“用户如何编辑输入框”
- 与业务动作域分离
- 例如:
  - `Ctrl+A`
  - `Ctrl+E`
  - `Ctrl+W`
  - `Backspace`
  - `Delete`

#### 终端透传域

- 面向 exec / tty 透传
- 不应被业务动作层截获

### 输入框编辑模型结论

- 用户输入框需要提供类似 shell / vi 的快捷编辑能力。
- 当前阶段建议:
  - 先把 `shell-like` 作为默认编辑模型
  - `vi-like` 作为后续可扩展 profile
- “进入输入模式”属于动作。
- “输入模式内部编辑”不属于业务动作，而属于编辑命令。

### 导航语义结论

- 项目希望采用类似 k9s 的快捷键体验，但不机械照抄。
- 左右箭头允许在不同页面承担不同具体功能。
- 但这些差异必须服从统一方向语义，而不是页面随意定义。

建议的统一语义:

- `Left`:
  - 返回
  - 收起
  - 回到上一级
  - 切到前一个栏
- `Right`:
  - 进入
  - 展开
  - 前进
  - 切到下一个栏
- `Enter`:
  - 当前焦点对象的主动作

结论:

- 同一个物理键可以在不同上下文触发不同动作。
- 但动作之间必须保持稳定语义。
- 不能继续依赖散乱硬编码来维护这种差异。

### 新增需求的接入原则

后续新增需求不应再走“多处手改”的旧路径，而应统一走:

1. 新增动作定义
2. 声明动作元数据
3. 绑定默认按键
4. 声明可用上下文
5. 由帮助页与 footer 自动派生展示

也就是说，新增需求应接入统一输入架构，而不是继续堆新分支。

### 当前建议的核心对象

- `ActionSpec`
- `Binding`
- `Context`
- `InputProfile`
- `Resolver`

### 当前未决定的问题

- 动作上下文模型的精确划分
- 动作命名空间最终形式
- 多绑定和冲突处理细则
- 帮助页和 footer 是否完全动态生成
- 编辑命令域是否独立成通用输入组件

### 当前阶段原则

- BR-002 可以接受输入层重构。
- 目标是建立可扩展输入基础设施，而不是只让一两个配置项生效。

### BR-002 上下文模型设计草案

#### 设计目标

- 为统一动作系统提供稳定的作用域模型。
- 让同一动作在不同页面结构下保持统一语义，同时允许局部差异化行为。
- 为 Footer / Help / 用户自定义键位提供一致的“当前可用动作集合”来源。

#### 四层上下文模型

建议将输入上下文固定为四层：

1. `App Context`
2. `Surface Context`
3. `View Context`
4. `Mode Context`

优先级顺序：

`Mode > View > Surface > App`

#### 1. App Context

定位：

- 全局应用级动作集合
- 不依赖具体资源页
- 仅在未被强输入模式屏蔽时生效

适合承载的动作：

- `app.help`
- `app.quit`
- `app.refresh`
- `app.filter`
- `app.command`
- `nav.panel_next`
- `nav.panel_prev`

#### 2. Surface Context

定位：

- 当前主表面 / 主页面
- 对应大面板或覆盖层级页面
- 决定资源域默认动作集合

当前项目建议的 `Surface Context` 清单：

- `containers`
- `images`
- `volumes`
- `networks`
- `compose`
- `logs`
- `detail`
- `help`

说明：

- `logs/detail/help` 虽然表现为覆盖层，但仍建议作为独立 surface 保留。
- 后续新增资源页时，优先新增 surface，而不是直接写新的按键分支。

#### 3. View Context

定位：

- Surface 内部的交互结构层
- 用于承接双栏、子视图、钻取视图、焦点区域切换
- 是解释 `Left/Right/Enter/Esc` 语义变化的关键层

建立独立 `View Context` 的判定条件：

- 焦点对象变了
- `Left/Right/Enter/Esc` 的语义变了
- 可用动作集合明显变了
- Footer / Help 需要展示不同动作集

当前项目建议的 `View Context` 草案：

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

不建议仅因数据状态不同而拆 view，例如：

- `containers.list.running`
- `containers.list.stopped`

因为这类变化属于数据内容变化，不属于交互结构变化。

#### 4. Mode Context

定位：

- 输入模式层
- 拥有最高优先级
- 用于临时覆盖普通页面动作

当前项目建议的 `Mode Context` 草案：

- `mode.normal`
- `mode.filter`
- `mode.search`
- `mode.command`
- `mode.confirm`
- `mode.mark`
- `mode.exec_dialog`
- `mode.exec_passthrough`
- `mode.detail_view`
- `mode.log_view`

说明：

- `mode.filter/search/command` 共享输入组件，但语义不同。
- `mode.exec_passthrough` 应视为强占输入模式，几乎屏蔽全部应用级动作。

#### 动作解析规则草案

##### 规则 A：Mode 先截获

- 只要进入特殊模式，先按 `Mode Context` 处理输入。
- 例如：
  - `mode.filter` 优先处理字符输入和 `Esc`
  - `mode.confirm` 优先处理 `y/n/esc`
  - `mode.exec_passthrough` 优先处理原始终端输入

##### 规则 B：View 决定当前结构动作

- 若 mode 未截获，则交给当前 `View Context`。
- View 负责解释：
  - 当前焦点对象是谁
  - `Left/Right/Enter/Esc` 的具体行为
  - 当前结构特有动作

例如：

- `compose.projects` 中 `nav.forward` 表示切换到服务栏
- `images.containers_subview` 中 `nav.back` 表示退出容器子视图

##### 规则 C：Surface 提供页面级默认动作

- 若 view 未单独处理，则回退到 `Surface Context`。
- Surface 负责提供资源页默认动作集合。

例如：

- `images.*` 默认允许 `image.pull`、`image.prune`、`image.copy_ref`
- `containers.*` 默认允许 `container.start`、`container.stop`、`container.logs`

##### 规则 D：App 提供全局兜底动作

- 最后回退到 `App Context`
- 仅提供稳定的全局动作
- 受 mode 屏蔽规则约束

##### 规则 E：动作解析优先于按键意义

- 不直接写“右键在 compose 中做什么”
- 应先解析为语义动作，例如 `nav.forward`
- 再由当前上下文解释该动作作用到哪个对象

##### 规则 F：View 可屏蔽 Surface，Mode 可屏蔽全部

- `View Context` 可以屏蔽不适用于当前结构的 surface 动作
- `Mode Context` 可以临时屏蔽 view/surface/app 动作

例如：

- `images.containers_subview` 可屏蔽 `image.pull`
- `mode.filter` 可屏蔽普通导航
- `mode.exec_passthrough` 可屏蔽除退出外的大部分应用动作

#### 当前项目 View Context 与动作集合草案

##### `containers.list`

- 全局/导航：
  - `nav.up`
  - `nav.down`
  - `nav.enter`
  - `nav.back`
- 页面扩展：
  - `resource.container.start`
  - `resource.container.stop`
  - `resource.container.restart`
  - `resource.container.kill`
  - `resource.container.logs`
  - `resource.container.exec`
  - `resource.container.inspect`
  - `resource.container.stats_toggle`
  - `resource.container.delete`
  - `view.mark.toggle`
  - `view.sort.next`
  - `view.sort.order_toggle`

##### `images.list`

- 全局/导航：
  - `nav.up`
  - `nav.down`
  - `nav.forward`
  - `nav.enter`
  - `nav.back`
- 页面扩展：
  - `resource.image.pull`
  - `resource.image.prune`
  - `resource.image.delete`
  - `resource.image.detail`
  - `resource.image.debug`
  - `resource.image.export`
  - `resource.image.copy_ref`
  - `view.mark.toggle`
  - `view.sort.next`
  - `view.sort.order_toggle`

##### `images.containers_subview`

- 全局/导航：
  - `nav.up`
  - `nav.down`
  - `nav.back`
- 页面扩展：
  - `resource.container.start`
  - `resource.container.stop`
  - `resource.container.restart`
  - `resource.container.logs`
  - `resource.container.exec`

说明：

- 该 view 中不再暴露 `image.pull/prune/debug/export`

##### `volumes.list`

- 全局/导航：
  - `nav.up`
  - `nav.down`
  - `nav.enter`
  - `nav.back`
- 页面扩展：
  - `resource.volume.detail`
  - `resource.volume.delete`
  - `view.mark.toggle`

##### `volumes.detail_subview`

- 全局/导航：
  - `nav.up`
  - `nav.down`
  - `nav.back`
- 页面扩展：
  - 保留最小动作集，避免和资源列表语义混淆

##### `networks.list`

- 全局/导航：
  - `nav.up`
  - `nav.down`
  - `nav.enter`
  - `nav.back`
- 页面扩展：
  - `resource.network.detail`
  - `resource.network.delete`
  - `view.sort.next`
  - `view.sort.order_toggle`
  - `view.mark.toggle`

##### `compose.projects`

- 全局/导航：
  - `nav.up`
  - `nav.down`
  - `nav.forward`
  - `nav.enter`
  - `nav.back`
- 页面扩展：
  - `resource.compose.start`
  - `resource.compose.stop`
  - `resource.compose.logs`
  - `resource.compose.down`
  - `resource.compose.detail`

说明：

- `nav.forward` 与 `nav.enter` 主要作用为切换到 `compose.services`

##### `compose.services`

- 全局/导航：
  - `nav.up`
  - `nav.down`
  - `nav.forward`
  - `nav.back`
  - `nav.enter`
- 页面扩展：
  - `resource.compose.start`
  - `resource.compose.stop`
  - `resource.compose.logs`
  - `resource.compose.down`

说明：

- `nav.enter` 可进入 `compose.containers_subview`
- `nav.back` 可返回 `compose.projects`

##### `compose.containers_subview`

- 全局/导航：
  - `nav.up`
  - `nav.down`
  - `nav.back`
- 页面扩展：
  - 当前阶段保持最小动作集

##### `logs.main`

- 全局/导航：
  - `nav.up`
  - `nav.down`
  - `nav.page_up`
  - `nav.page_down`
  - `nav.back`
- 页面扩展：
  - `mode.search.open`
  - `logs.search.next`
  - `logs.search.prev`
  - `logs.wrap.toggle`
  - `logs.top`
  - `logs.bottom`

##### `detail.main`

- 全局/导航：
  - `nav.up`
  - `nav.down`
  - `nav.page_up`
  - `nav.page_down`
  - `nav.back`
  - `nav.enter`
- 页面扩展：
  - `detail.top`

##### `help.main`

- 页面扩展：
  - `app.help.close`
  - `nav.back`

#### 当前未决定的问题

- 每个动作的最终命名空间
- `Mode Context` 与 `Surface/View` 的代码承载位置
- Footer / Help 如何从上下文动作集合中自动生成展示
- 用户配置是否在后续阶段暴露上下文级覆盖能力

### BR-002 动作命名空间收敛草案

#### 设计目标

- 为统一动作系统建立长期稳定的命名规则。
- 让动作名表达语义而不是物理按键。
- 为 Keymap、Help、Footer、审计日志等后续系统提供统一引用标识。

#### 总体命名规则

优先格式：

- `domain.object.verb`

没有明确对象时：

- `domain.verb`

命名原则：

- 动作名表达用户意图，不表达具体按键
- 动作名必须带 domain
- 资源动作必须带资源类型
- 相同语义跨页面复用同一动作名

#### 顶层 Domain 草案

建议固定以下顶层域：

- `app`
- `nav`
- `panel`
- `resource`
- `logs`
- `detail`

说明：

- `mode` 作为上下文存在，不作为顶层动作命名空间
- 输入态应通过上下文和编辑命令域表达，不纳入统一业务动作命名空间

#### 1. `app.*`

定位：

- 全局应用级能力
- 不依赖具体资源对象

建议动作：

- `app.help.open`
- `app.help.close`
- `app.quit`
- `app.refresh`
- `app.filter.open`
- `app.command.open`

#### 2. `nav.*`

定位：

- 统一承接方向、进入、返回、分页、面板切换等结构导航语义

建议动作：

- `nav.up`
- `nav.down`
- `nav.left`
- `nav.right`
- `nav.forward`
- `nav.back`
- `nav.enter`
- `nav.page_up`
- `nav.page_down`
- `nav.top`
- `nav.bottom`
- `nav.panel_next`
- `nav.panel_prev`

当前建议：

- 内部语义动作优先使用 `nav.forward` / `nav.back`
- 物理按键 `Left/Right` 只是默认绑定，不应成为动作身份本身

#### 3. `panel.*`

定位：

- 不直接修改资源状态，但影响当前界面能力或当前页面状态

建议动作：

- `panel.header.toggle`
- `panel.connection.show`
- `panel.runtime.switch`
- `panel.sort.next`
- `panel.sort.order_toggle`
- `panel.mark.enter`
- `panel.mark.toggle`
- `panel.mark.clear`

#### 4. `resource.*`

定位：

- 所有真正的业务资源操作

统一格式：

- `resource.<type>.<verb>`

建议资源对象：

- `container`
- `image`
- `volume`
- `network`
- `compose_project`
- `compose_service`
- `runtime`
- `exec_session`

##### `resource.container.*`

- `resource.container.start`
- `resource.container.stop`
- `resource.container.restart`
- `resource.container.kill`
- `resource.container.delete`
- `resource.container.logs`
- `resource.container.exec`
- `resource.container.inspect`
- `resource.container.stats_toggle`

##### `resource.image.*`

- `resource.image.pull`
- `resource.image.prune`
- `resource.image.delete`
- `resource.image.detail`
- `resource.image.debug`
- `resource.image.export`
- `resource.image.copy_ref`
- `resource.image.expand_containers`

##### `resource.volume.*`

- `resource.volume.detail`
- `resource.volume.delete`
- `resource.volume.expand`

##### `resource.network.*`

- `resource.network.detail`
- `resource.network.delete`

##### `resource.compose_project.*`

- `resource.compose_project.start`
- `resource.compose_project.stop`
- `resource.compose_project.logs`
- `resource.compose_project.down`
- `resource.compose_project.detail`

##### `resource.compose_service.*`

当前先保留为后续扩展域，必要时再补：

- `resource.compose_service.logs`
- `resource.compose_service.expand`

#### 5. `logs.*`

定位：

- 日志页专属能力

建议动作：

- `logs.search.open`
- `logs.search.next`
- `logs.search.prev`
- `logs.wrap.toggle`

#### 6. `detail.*`

定位：

- 详情页专属能力

建议动作：

- `detail.top`
- `detail.page_up`
- `detail.page_down`

后续若支持区段折叠，可再扩展：

- `detail.section.expand`
- `detail.section.collapse`
- `detail.section.toggle`

#### 动词统一词表草案

建议优先使用以下动词：

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

#### 需要避免的命名问题

##### 避免动作名直接表达按键

不应出现：

- `ctrl_p_pull`
- `right_expand`
- `f2_switch`

##### 避免缺少作用域的扁平名字

不应继续长期保留：

- `detail`
- `delete`
- `back`
- `refresh`

##### 避免同义词混用

建议统一：

- `delete` 优先于 `remove`
- `detail` 优先于 `info`
- `switch` 优先于 `change`

#### 旧动作名到新动作名映射草案

##### 旧 `KeyAction` / 逻辑名

- `quit` -> `app.quit`
- `help` -> `app.help.open`
- `filter` -> `app.filter.open`
- `refresh` -> `app.refresh`
- `tabNext` -> `nav.panel_next`
- `tabPrev` -> `nav.panel_prev`
- `up` -> `nav.up`
- `down` -> `nav.down`
- `enter` -> `nav.enter`
- `back` -> `nav.back`
- `command` -> `app.command.open`

##### 旧资源动作

- `containerStart` -> `resource.container.start`
- `containerStop` -> `resource.container.stop`
- `containerRestart` -> `resource.container.restart`
- `containerKill` -> `resource.container.kill`
- `containerRemove` -> `resource.container.delete`
- `containerLogs` -> `resource.container.logs`
- `containerExec` -> `resource.container.exec`
- `containerInspect` -> `resource.container.inspect`
- `containerStats` -> `resource.container.stats_toggle`

- `imagePull` -> `resource.image.pull`
- `imageRemove` -> `resource.image.delete`
- `imagePrune` -> `resource.image.prune`

- `volumeRemove` -> `resource.volume.delete`
- `networkRemove` -> `resource.network.delete`

##### 旧页面/界面动作

- `switchRuntime` -> `panel.runtime.switch`
- `showConnectionInfo` -> `panel.connection.show`
- `toggleHeader` -> `panel.header.toggle`
- `sort` -> `panel.sort.next`
- `sortOrderToggle` -> `panel.sort.order_toggle`
- `mark` -> `panel.mark.enter`
- `toggleMark` -> `panel.mark.toggle`

##### 旧日志/详情动作

- `logSearchNext` -> `logs.search.next`
- `logSearchPrev` -> `logs.search.prev`
- `logWrapToggle` -> `logs.wrap.toggle`
- `detailTop` -> `detail.top`

#### 当前阶段结论

- 动作命名空间应以语义域为中心，而不是以当前代码文件为中心
- `mode` 是上下文，不是顶层动作域
- `resource.*` 必须成为所有业务操作的统一归属
- `nav.forward/back` 应逐步取代直接依赖 `left/right` 的语义表达
- 后续新增动作必须先纳入命名空间，再接入绑定和展示系统

---

## 后续讨论入口

后续建议继续按这个顺序展开:

1. BR-002 上下文模型划分
2. BR-002 动作命名空间与注册表结构
3. BR-001 结构化详情 section model
4. BR-001 普通镜像与 manifest list 的区段排版

---

## BR-003 默认键位、帮助页、Footer 的统一信息源

### 已确认的问题本质

- BR-003 不是“帮助页文案修一下”。
- 核心问题是键位说明没有统一信息源。
- 当前至少有多套并行来源:
  - 默认配置
  - 运行时默认映射
  - 帮助页说明
  - Footer 展示
  - README / 文档摘要

### 已达成的宏观方向

- 动作定义、默认绑定、上下文可用性只维护一份。
- 所有展示层都从同一套动作/绑定系统派生。
- 事实统一，展示形式可以不同。

### 信息源统一原则

必须统一的事实:

- 动作是什么
- 默认绑定是什么
- 在什么上下文可用

不要求完全统一的内容:

- 帮助页怎么分组
- Footer 怎么排版
- README 是否只展示摘要

结论:

- 展示层可以有不同布局
- 但不能再自己维护独立键位真相

### 关于动作与页面的关系

- 键位绑定默认属于动作本身，而不是页面本身。
- 页面只声明“哪些动作在当前上下文可用”。
- 页面不应单独拥有一套独立键位定义。

### 全局动作的定义

全局动作不按“是否出现在所有页面”定义，而按“功能语义是否跨页面稳定存在”定义。

例如这些可以视为全局动作:

- `up` / `down`
- `left` / `right`
- `enter`
- `esc`
- `tab` / `shift+tab`
- `/`
- `?`

判断原则:

- 只要一个动作在多数页面都承担同类功能，它就是全局动作。
- 页面差异只允许改变作用对象，不改变动作语义。

### 动作集合模型

动作集合分两层:

1. 全局动作集
2. 页面扩展动作集

页面只做这件事:

- 声明当前支持哪些全局动作
- 补充当前页面专属动作

结论:

- 页面绑定的是动作能力，不是具体按键。
- 具体按键优先从全局动作定义继承。

### Footer 展示规则

已确认方向:

- Footer 展示当前页面所有快捷键操作
- 不做人工筛选
- “全局”快捷键固定放在第一列/第一组
- 其余页面动作直接平铺

这意味着:

- Footer 不再是“重点摘要”
- Footer 是“当前上下文完整动作面板”

### Help 展示规则

- Help 页与 Footer 一样，应从统一动作系统派生。
- Help 可以比 Footer 更分组、更详细。
- 但 Help 不应继续手写具体按键字符串作为事实源。

### 默认配置的定位

- 默认配置不应再作为独立键位事实源。
- 它应被视为统一动作注册表默认绑定的导出视图。
- 如果以后支持生成默认配置，生成源应来自同一套动作定义。

### 新增需求的接入原则

后续新增需求统一走:

1. 新增动作
2. 绑定默认按键
3. 声明可用上下文
4. 自动派生 Help / Footer / 默认配置展示

禁止继续走:

- 改运行时一处
- 手写帮助页一处
- 手写 Footer 一处
- 再手写默认配置一处

### 当前未决定的问题

- 全局动作集的最终边界清单
- 页面扩展动作如何声明
- Footer 的最终排版格式
- Help 是否完全动态生成还是“分组静态、键位动态”

---

## BR-004 Filter / Search 模式收敛

### 已确认的问题本质

- 当前资源过滤与日志搜索共享 `/` 入口和同一输入条 UI。
- 但两者实际行为已经不同:
  - 资源页输入时即时过滤数据集
  - 日志页只是编辑搜索词，按 `Enter` 后应用匹配
- `SearchTimer` / `SearchTick` / `StartSearchDebounce()` 仍保留旧防抖路径，但已经不驱动真实交互。

### 已达成的宏观方向

- 资源过滤与日志搜索不再视为同一种模式。
- 它们可以共享输入组件，但不共享语义控制器。
- `/` 保留为统一查询入口键，但进入哪种模式由当前页面决定。

### 模式边界

#### Filter

- 面向结构化资源数据
- 典型页面: containers / images / volumes / networks / compose
- 输入即生效
- 作用是缩小当前资源集合

#### Search

- 面向连续文本内容
- 典型页面: logs
- 不改变原始数据集
- 作用是匹配、定位、跳转

#### Command

- 继续保持独立
- 不归入 Filter / Search 语义

### Filter 退出语义

已确认方向:

- 第一次 `Esc`:
  - 不清空筛选词
  - 不恢复默认列表
  - 进入 5 秒待退出窗口
  - 搜索框仍可继续输入
- 5 秒内第二次 `Esc`:
  - 清空筛选条件
  - 恢复默认资源列表
  - 退出筛选功能本身
- 超过 5 秒未再次按 `Esc`:
  - 本次待退出失效
  - 当前筛选继续保持
- 退出筛选后，只有再次按 `/` 才会重新进入筛选状态

### 可见反馈

- Filter 的退出提示不另起新展示体系。
- 相关提示统一集成到消息通知体系中。

### 当前未决定的问题

- `Filter` / `Search` 的最终状态结构体设计
- 旧 `searchDebounceMs` 配置如何迁移或删除
- 输入组件与模式控制器的最终拆分方式

---

## BR-005 消息 / 日志管理

### 已确认的问题本质

- 当前至少存在多套并行机制:
  - 顶部通知 (`ToastMessage/ToastTimer`)
  - 底部操作行 (`InfoMessage/ErrorMessage`)
  - 状态栏摘要
  - Header 提示 (`KeyHint`)
  - 正文日志 (`LogContent`)
- 这些机制尚未形成统一事件模型，也没有清晰的操作追溯链。

### 已达成的宏观方向

- 当前阶段落盘日志应优先定位为审计日志，而不是程序内部运行日志。
- `Notification` 与 `Operation Log` 不是平行系统，而是同一业务操作链的不同视图。
- 重要用户操作应通过统一 `trace_id` 串联。
- 应用本身应支持落盘日志，方便操作历史查看与追溯。

### Trace 边界

#### 应生成 `trace_id` 的动作

- 会改变已有资源状态的业务动作
- 会改变数据源上下文的动作
- 高价值副作用动作

典型包括:

- 容器 `start/stop/restart/kill/delete`
- 镜像 `pull/remove/prune`
- volume / network 删除
- compose `start/stop/down`
- runtime switch
- exec session start / finish / fail

#### 不进入完整 trace 的动作

- 普通查询
- 普通过滤 / 搜索
- 普通导航
- 查看详情
- 高频输入事件

这些动作最多做 simple event 记录，或完全不记。

### Notification / Operation Log / File Log 的关系

- Notification:
  即时反馈视图，面向“用户现在需要知道什么”
- Operation Log:
  过程视图，面向“系统最近做了什么”
- File Log:
  长期留存视图，面向审计追溯

它们都应由统一事件模型派生，而不是各自生成独立内容。

### 落盘日志定位

- 用户操作审计日志只记录用户显式触发的业务动作。
- 不混入系统内部运行细节，例如:
  - 打开日志文件
  - 文件轮转
  - 后台刷新
  - 内部自动重连
  - UI 渲染事件

建议的逻辑分类:

- `audit`
- `runtime`

当前阶段优先先把 `audit` 设计正确。

### 日志命名方向

- 审计日志优先按 `类型-日期` 命名:
  - `audit-2026-07-20.jsonl`
- 运行日志可按 `类型-日期-级别` 命名:
  - `runtime-2026-07-20-error.log`
  - `runtime-2026-07-20-info.log`

### 审计目标模型

已确认最小统一字段:

- `type`
- `id`
- `name`

已确认方向:

- 资源对象扩展字段不用 `map[string]string`
- 使用 `struct + interface` 做强类型约束
- 没有天然独立 ID 的对象，第一阶段允许 `id == name` 作为回退

建议接口方向:

```go
type AuditTarget interface {
    TargetType() string
    TargetID() string
    TargetName() string
}
```

### 资源对象扩展字段方向

#### 必须扩展

- `container`
- `image`
- `runtime`
- `exec_session`
- `compose_project`
- `compose_service`

#### 建议扩展

- `volume`
- `network`

原则:

- 默认统一 `type/id/name`
- 只有在审计追溯价值明确时再加资源专属字段
- 扩展字段通过各资源独立 struct 表达

### 当前未决定的问题

- 审计记录本身的统一 struct
- `audit` 与 `runtime` 的实际文件落盘策略
- `trace_id` / `event_id` 的最终生成规则
- 资源对象强类型 struct 的最终字段清单

### BR-005 审计记录与 UI 投影细化方案

#### `AuditRecord.Result` 的语义

已确认方向：

- `Result` 表示当前这条 `AuditRecord` 对应事件的阶段状态 / 结果。
- `Result` 不表示整条操作链的最终总结。
- `Result` 也不表示被操作资源对象的最终业务状态。

换句话说：

- `Action` 回答：这次在做什么
- `Result` 回答：这条事件在说这件事走到了哪一步
- 资源最终状态应放在 `Target` 扩展字段或 `Details`

例如一次容器停止操作：

1. `action=resource.container.stop`, `result=requested`
2. `action=resource.container.stop`, `result=started`
3. `action=resource.container.stop`, `result=succeeded`

此时容器最终可能进入：

- `target.meta.state=exited`

结论：

- 一次业务操作可以产生多条相同 `Action`、相同 `TraceID`、但 `Result` 不同的记录。

#### `Result` 与“最终状态”的边界

应放在 `Result` 的内容：

- 是否已请求
- 是否已开始
- 是否成功
- 是否失败
- 是否取消

不应放在 `Result` 的内容：

- 容器最终 `running/exited`
- runtime 最终切换到哪个 host
- image 最终 tag/digest 是什么

这些属于业务对象最终状态，应进入：

- `Target` 扩展字段
- 或 `Details`

#### `Result` 枚举方向

当前建议的结果值：

- `requested`
- `started`
- `succeeded`
- `failed`
- `cancelled`

适用场景说明：

##### `requested`

- 用户已明确触发该业务动作
- 系统尚未真正开始执行底层操作

##### `started`

- 系统开始调用 runtime 或开始执行实际副作用动作

##### `succeeded`

- 当前这条业务操作事件成功完成

##### `failed`

- 当前这条业务操作事件执行失败

##### `cancelled`

- 用户在确认前取消
- 或流程在进入执行前被主动中止

#### 审计记录建议结构

当前讨论方向上的建议结构：

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

#### Target 强类型方向

已确认：

- 运行时使用 `struct + interface`
- 不使用松散 `map[string]string`
- 最小统一字段为：
  - `type`
  - `id`
  - `name`

建议接口方向：

```go
type AuditTarget interface {
    TargetType() string
    TargetID() string
    TargetName() string
    ToDTO() AuditTargetDTO
}
```

建议统一落盘 DTO：

```go
type AuditTargetDTO struct {
    Type string          `json:"type"`
    ID   string          `json:"id"`
    Name string          `json:"name"`
    Meta json.RawMessage `json:"meta,omitempty"`
}
```

说明：

- `Meta` 的来源是各资源自己的强类型 struct 的序列化结果
- 不是任意 map

#### Runtime / UI 上下文字段方向

建议不要把这些上下文塞进 `Details`。

建议保留独立结构：

```go
type RuntimeContext struct {
    Type string `json:"type"`
    Name string `json:"name"`
    Host string `json:"host"`
}
```

```go
type UIContext struct {
    Surface string `json:"surface,omitempty"`
    View    string `json:"view,omitempty"`
    Mode    string `json:"mode,omitempty"`
}
```

#### `Details` 字段边界

`Details` 只放补充信息，不放核心身份信息。

建议承载：

- `duration_ms`
- `error`
- `shell`
- `exit_code`

不建议承载：

- 资源身份
- runtime 身份
- UI 上下文

#### UI 投影层定位

已确认方向：

- UI 投影层不是日志本身
- UI 投影层是“底层事件 -> 展示形态”的转换层

同一条底层事件在不同 UI 区域可以有不同呈现：

- 顶部通知
- Footer 操作历史
- 状态栏摘要
- 文件审计日志

#### 建议的投影器拆分

##### `NotificationProjector`

定位：

- 顶部短时通知
- 面向“用户现在必须知道什么”
- 偏结果型，不偏过程型

建议规则：

- 默认不显示高频 `requested`
- 默认不显示高频 `started`
- `succeeded/failed` 可进入通知
- 警告类 UI 提示可进入通知

##### `OperationLogProjector`

定位：

- Footer / 后续操作历史面板
- 面向“系统最近做了什么”
- 保留较完整的操作链视图

建议规则：

- 主要消费用户业务操作审计事件
- 比通知更完整
- 可保留最近 N 条

##### `StatusProjector`

定位：

- 状态栏 / 面板状态摘要
- 面向“当前处于什么状态”

说明：

- 它更接近状态投影，不是事件历史投影
- 可直接消费状态变更事件，或消费应用状态树

#### 审计事件与界面提示事件分层

已确认方向：

- 审计事件与界面提示事件不应混为一类

建议拆分为：

##### A. `Audit Event`

用于：

- 用户业务操作审计
- 进入 trace 链
- 支撑落盘 audit log

##### B. `UI Message Event`

用于：

- 界面提示
- 临时说明
- 非审计型反馈

例如：

- “再按一次 Esc 清除筛选”
- “无匹配项”
- “当前筛选仍保留”

说明：

- `NotificationProjector` 可以同时消费两类事件
- `OperationLogProjector` 主要消费 `Audit Event`
- 落盘审计日志只记录 `Audit Event`

#### 第一版落地范围建议

当前建议第一版先接这些业务动作：

- `resource.container.start`
- `resource.container.stop`
- `resource.container.restart`
- `resource.container.kill`
- `resource.container.delete`
- `resource.image.pull`
- `resource.image.prune`
- `resource.image.delete`
- `resource.compose_project.start`
- `resource.compose_project.stop`
- `resource.compose_project.down`
- `panel.runtime.switch`
- `resource.container.exec` 的 session start / finish / fail

原因：

- 这些动作最适合验证 trace、target、落盘和通知投影链路
