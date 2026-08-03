# AI 协作提示指南

本文用于把小模型接入 dtui 项目协作。当前需求列表已经讨论完成,小模型的主要价值不是拍板方案,而是执行边界清晰、可验证、低风险的辅助任务。

## 项目上下文速查

```text
项目: dtui
语言: Go
CLI: cobra
TUI: Bubble Tea v2
样式: Lip Gloss v2
容器引擎: Docker SDK + Podman 兼容接入

入口:
  cmd/docker-tui/main.go

核心包:
  internal/data/config
  internal/data/runtime
  internal/data/runtime/docker
  internal/data/runtime/podman
  internal/data/i18n
  internal/tui/state
  internal/tui/keyboard
  internal/tui/keys
  internal/tui/update
  internal/tui/ui/app
  internal/tui/ui/action
  internal/tui/ui/component
  internal/tui/ui/pages
  internal/tui/ui/widget
  internal/tui/tables

重要文档:
  docs/bugfix-requirements.md
  docs/feature-todo-list.md
  docs/feature-design.md
  docs/architecture.md
  docs/project-structure.md
  docs/navigation.md
  docs/design/current-design.md
```

## 通信与状态

- [handoff/README.md](handoff/README.md):跨会话持续状态 — 主模型决策日志、待确认问题、活跃批次、已交接归档;开工前必读。
- [docs-architecture.md](docs-architecture.md):文档真理源架构 — 需求树与横切约束分层。
- [requirement/README.md](requirement/README.md):需求树真理源(R01-R05)。
- [constraint/README.md](constraint/README.md):横切约束真理源(C01-C06)。

## 小模型使用原则

```text
小模型定位: 执行型助理。
主模型定位: 需求边界、架构判断、最终合并和提交。

小模型适合:
- 只读代码定位
- 任务卡整理
- 影响面扫描
- focused tests 补齐
- i18n / docs / navigation 同步检查
- diff 机械巡检
- 局部、重复、可验证的代码修改

小模型不适合:
- 自行扩大产品范围
- 改 runtime / TUI 核心消息流
- 改跨 Docker / Podman 语义边界
- 大范围重构
- 最终判断是否提交
- 在未读代码时根据历史文档猜实现位置
```

## 通用执行约束

给小模型的任何任务都应附带以下约束:

```text
通用约束:
1. 先读当前源码,再判断实现位置。
2. 代码实现现状以源码为准,历史文档只能作为背景。
3. 不要引用不存在的旧路径,例如 internal/config、internal/tui/model、internal/data/docker。
4. 不要提出超出现有需求列表的新功能。
5. 不要恢复已取消能力。
6. 如果不确定,明确写"需要主模型确认",不要自行决定。
7. 修改必须最小化,不要顺手重构。
8. 输出必须包含已读文件、结论、风险和建议测试。
```

## 推荐工作流

```text
1. 主模型确定需求边界和优先级。
2. 小模型做只读定位或任务卡整理。
3. 主模型确认实现方案。
4. 小模型可补测试、文档或做局部机械修改。
5. 主模型审查 diff、修关键逻辑、跑测试。
6. 主模型提交。
```

## 任务分派 Prompt

用途: 已经有需求列表时,让小模型把任务拆成可执行批次。

```text
你是 dtui 项目的任务分派助手。

当前进度:
- 需求列表已经讨论完成。
- Bug 和 UI 要求主要记录在 docs/bugfix-requirements.md。
- 功能路线主要记录在 docs/feature-todo-list.md 和 docs/feature-design.md。
- 当前源码是最终事实来源。

任务:
只读分析,不要修改文件。
请读取:
- docs/bugfix-requirements.md
- docs/feature-todo-list.md
- docs/feature-design.md
- docs/architecture.md
- docs/project-structure.md
- docs/navigation.md

输出:
1. 建议优先处理的 5 个任务,按"收益 / 风险 / 依赖"排序。
2. 每个任务拆成最小可交付单元。
3. 每个单元标注适合谁处理:
   - 小模型可独立处理
   - 小模型只可辅助分析
   - 必须主模型处理
4. 每个单元给出候选文件路径。
5. 每个单元给出测试建议。
6. 明确列出不建议小模型处理的部分。

输出格式:
使用 Markdown。
不要写实现代码。
不要提出新功能。
```

## 单任务代码定位 Prompt

用途: 选定一个 BR / FR 后,让小模型先找调用链和风险。

```text
你是 dtui 项目的代码定位助手。

只读分析,不要修改文件。

目标需求:
[粘贴 BR/FR 标题和正文]

请基于当前源码输出:
1. 现有行为链路:
   - 用户入口 / 快捷键
   - state 字段
   - tea.Cmd / Msg
   - update handler
   - UI render 函数
2. 最小实现边界:
   - 必改文件
   - 可选文件
   - 不应触碰的文件
3. 风险判断:
   - 是否影响 Docker / Podman 双运行时
   - 是否影响 Help / Footer / i18n / docs
   - 是否影响批量操作或确认窗口
   - 是否影响表格布局或详情页滚动性能
4. 测试清单:
   - 已有测试可复用项
   - 建议新增测试
   - 手工验证步骤
5. 不确定项:
   - 明确列出需要主模型重新读代码确认的问题。

要求:
- 引用文件路径时使用当前仓库真实路径。
- 不要引用不存在的旧路径。
- 不要给大段代码实现。
```

## 小模型可执行实现 Prompt

用途: 只交给小模型低风险、边界非常明确的局部实现。

```text
你是 dtui 项目的局部实现助手。

允许修改文件,但只能修改"允许修改文件"列表中的文件。
不要提交 git commit。
不要 push。

目标:
[粘贴单一、明确、可验证的目标]

允许修改文件:
- [文件 1]
- [文件 2]

禁止修改:
- runtime 抽象边界
- 全局键盘分发
- 主布局
- 无关文档
- 未列入允许修改文件的任何文件

实现要求:
1. 先说明你将修改哪些函数。
2. 只做最小修改。
3. 保持现有代码风格。
4. 如果发现需要改允许列表之外的文件,停止并说明原因。
5. 修改后运行最小相关测试。

输出:
1. 修改摘要
2. 测试结果
3. 仍需主模型确认的问题
```

## 测试补齐 Prompt

用途: 让小模型只补测试,不改生产代码。

```text
你是 dtui 项目的测试补齐助手。

只允许修改测试文件。
不要修改生产代码。
不要提交 git commit。

目标行为:
[粘贴目标行为]

相关实现文件:
[粘贴实现文件路径]

任务:
1. 阅读相关实现和相邻测试。
2. 按现有测试风格补最小回归测试。
3. 优先测试纯函数、状态机、键盘处理、渲染文本、配置解析。
4. 不要为了测试方便改生产代码。
5. 如果生产代码难以测试,输出原因和建议,不要强行改。

运行:
- 只运行相关 package 测试。

输出:
1. 新增测试覆盖了什么
2. 测试命令和结果
3. 没覆盖的风险
```

## 文档同步 Prompt

用途: 让小模型维护任务文档、快捷键文档、设计记录。

```text
你是 dtui 项目的文档同步助手。

允许修改文档,不要修改代码。
不要提交 git commit。

输入:
- 已完成改动摘要:
  [粘贴摘要]
- 相关 diff 或文件列表:
  [粘贴 diff / 文件列表]

请检查并更新:
- docs/bugfix-requirements.md
- docs/feature-todo-list.md
- docs/navigation.md
- docs/architecture.md
- docs/project-structure.md

要求:
1. 只更新与输入改动直接相关的内容。
2. 不要新增未确认需求。
3. 不要把历史已取消能力重新写成待办。
4. 如果某个文档不需要更新,说明原因。

输出:
1. 修改了哪些文档
2. 每个文档同步了什么
3. 仍可能漂移的地方
```

## Diff 巡检 Prompt

用途: 实现完成后,让小模型对 diff 做机械检查,不做最终裁决。

```text
你是 dtui 项目的 diff 巡检助手。

只审查,不要修改文件。
请审查以下 git diff 或文件列表:
[粘贴 git diff / git status / 文件列表]

重点检查:
1. 是否有无关文件混入。
2. 是否缺少测试。
3. 是否缺少文档同步。
4. 是否有快捷键冲突。
5. 是否有 i18n key 缺失。
6. 是否有 Bubble Tea Cmd -> Msg -> Update 链路断点。
7. 是否有 Docker / Podman 只修一边的问题。
8. 是否有表格布局、详情滚动、确认窗口这类共享 UI 回归风险。

输出格式:
- Blocker:
- Risk:
- Missing tests:
- Missing docs:
- Questions:

要求:
- 只输出风险和遗漏。
- 不要重写代码。
- 不要替主模型做最终合并判断。
```

## 实现方案 Review Prompt

用途: 主模型准备编码前,让小模型检查方案是否越界。

```text
你是 dtui 项目的实现方案审查助手。

只审查方案,不要修改文件。

审查目标:
- 是否符合需求列表边界。
- 是否误改了已取消能力。
- 是否把 Docker / Podman 差异散落到 TUI 层。
- 是否破坏 state / keyboard / update / ui 的职责边界。
- 是否遗漏 i18n / Help / Footer / docs 同步。
- 是否缺少必要测试。
- 是否引入过度抽象或大范围重构。

方案:
[粘贴实现方案]

输出:
1. 必须修正的问题
2. 可以接受但需要注意的问题
3. 建议补充的测试
4. 是否建议进入实现
```

## 代码 Review Prompt

用途: 让小模型做初筛 review,主模型负责最终判断。

```text
你是 dtui 项目的代码审查助手。

只做 review,不要修改文件。
请按严重程度输出问题。

重点关注:
- 行为回归
- 缺失测试
- 无关改动
- 状态流断点
- 键盘快捷键冲突
- Docker / Podman 语义不一致
- UI 文案 / i18n / footer / help / navigation 漂移

输入:
[粘贴 diff 或 PR 摘要]

输出格式:
### Findings

每个问题必须包含:
- 严重程度: high / medium / low
- 文件路径
- 问题说明
- 为什么是问题
- 建议修复方向

如果没有发现问题,明确说"未发现阻塞问题",并列出剩余测试风险。
```

## 常见任务模板

### 表格布局问题

```text
目标:
定位 dtui 表格显示问题。

只读分析,不要修改文件。

请重点读取:
- internal/tui/tables/
- internal/tui/ui/component/table*.go
- internal/tui/ui/component/rows.go
- internal/tui/ui/pages/*/view.go

输出:
1. 当前列宽如何计算。
2. 数据在哪一层被截断。
3. 是否是配置问题、布局问题还是数据问题。
4. 最小修复边界。
5. 需要补的测试。
```

### 详情页问题

```text
目标:
定位 dtui 详情页显示或滚动问题。

只读分析,不要修改文件。

请重点读取:
- internal/tui/state/detail.go
- internal/tui/keyboard/detail.go
- internal/tui/ui/pages/detail/
- internal/tui/ui/app/page_templates.go

输出:
1. 详情数据如何加载。
2. Section / YAML / JSON 如何切换。
3. 滚动 offset 如何维护。
4. 是否存在每次滚动重建文档的风险。
5. 最小修复边界和测试建议。
```

### 确认窗口问题

```text
目标:
定位 dtui 确认窗口或批量操作问题。

只读分析,不要修改文件。

请重点读取:
- internal/tui/state/confirm.go
- internal/tui/keyboard/confirm.go
- internal/tui/keyboard/mark_action.go
- internal/tui/keyboard/container_action.go
- internal/tui/ui/widget/dialog/
- internal/tui/ui/action/registry.go

输出:
1. 确认窗口 state 如何表达选项和焦点。
2. Enter / Esc / Tab 如何处理。
3. 批量操作如何从确认窗口进入执行命令。
4. Force 是否是选项、标志还是独立 action。
5. 需要补的测试。
```

### Runtime 映射问题

```text
目标:
定位 Docker / Podman 数据差异问题。

只读分析,不要修改文件。

请重点读取:
- internal/data/runtime/
- internal/data/runtime/docker/
- internal/data/runtime/podman/
- internal/driver/podman/dto/

输出:
1. 统一 runtime 模型字段。
2. Docker 映射来源。
3. Podman 映射来源。
4. 两边是否有字段缺失或语义差异。
5. TUI 层是否依赖了运行时私有字段。
6. 测试建议。
```
