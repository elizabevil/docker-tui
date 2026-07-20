# 历史 Bug 修复需求

> 建立日期: 2026-07-20
> 目的: 独立记录“历史遗留问题”的修复需求，不与产品路线图和实现现状文档混写。

本文是修复需求台账，不是实现说明。实现现状仍以 [README.md](README.md)、[architecture.md](architecture.md) 和代码为准。

## 使用规则

- 只记录已经复现、已从代码验证、或已有明确用户反馈的问题。
- 每条需求必须包含: 症状、当前行为、代码锚点、期望行为、验收标准。
- 需求编号固定为 `BR-xxx`，便于在设计文档、提交记录和 PR 说明中交叉引用。
- 如果问题只是文档漂移，不进入本文；应直接修正文档。

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

- 状态: `designing`
- 优先级: `high`
- 症状:
  镜像详情页已经渲染出 `History` 分区，但当前始终显示占位文案，无法查看 layer history。
- 当前行为:
  `detail/image.go` 会构造 `inspect.section_history` 区块，但当内容为空时回退到 `inspect.no_history`。
- 代码锚点:
  [internal/data/docker/images.go](/home/debi/IdeaProjects/docker-tui/internal/data/docker/images.go:141)
  [internal/tui/ui/pages/detail/image.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/detail/image.go:22)
- 期望行为:
  镜像详情中的 `History` 应作为一级区段保留；普通镜像与 manifest list 允许使用不同子语义和排版模型。镜像引擎可提供历史数据时，普通镜像至少展示 `CreatedBy`、`Size`、`Comment` 等核心字段；manifest list 则展示其平台变体结构。确实拿不到时，后续再定义降级语义。
- 验收标准:
  1. 普通镜像详情页不再默认落到“暂不可用”占位文案。
  2. 至少支持一条有效历史记录渲染。
  3. 无历史数据时，错误或降级文案来源明确，可区分能力缺失与实现缺失。

### BR-002 自定义键位配置未生效

- 状态: `designing`
- 优先级: `high`
- 症状:
  配置文件里已经存在 `keymap.*` 结构，但运行时不会读取用户自定义键位。
- 当前行为:
  `HandleKeyPress()` 直接使用 `keys.DefaultKeyMapping()`，未将 `config.Keymap` 编译为实际映射。
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

- 状态: `designing`
- 优先级: `medium`
- 症状:
  默认配置里的 `keymap` 示例仍使用旧的大写键位写法，如 `S/K/P`，而运行时已经转向 `Ctrl+S`、`Ctrl+K`、`Ctrl+P`。
- 当前行为:
  文档和默认配置会误导用户，以为大写单键仍可直接触发 stop / kill / pull。
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

- 状态: `designing`
- 优先级: `medium`
- 症状:
  过滤交互已经变为“资源表输入即过滤、日志页按 Enter 应用搜索”，但 `SearchTimer` / `SearchTick` / `StartSearchDebounce()` 仍保留旧的防抖模型。
- 当前行为:
  运行时与配置项 `ui.searchDebounceMs` 的含义不一致，维护者容易误判真实行为。
- 代码锚点:
  [internal/tui/keyboard/keyboard.go](/home/debi/IdeaProjects/docker-tui/internal/tui/keyboard/keyboard.go:198)
  [internal/tui/update_tick.go](/home/debi/IdeaProjects/docker-tui/internal/tui/update_tick.go:121)
  [internal/tui/composable/filter.go](/home/debi/IdeaProjects/docker-tui/internal/tui/composable/filter.go:1)
  [internal/data/config/types.go](/home/debi/IdeaProjects/docker-tui/internal/data/config/types.go:52)
- 期望行为:
  资源页与日志页应拆分为不同查询语义: 结构化资源页使用 `Filter`，日志页使用 `Search`。资源页过滤默认即时生效，并采用双 `Esc` 退出筛选功能的交互；日志页搜索不改变数据集，只负责匹配与跳转。
- 验收标准:
  1. `Filter` 与 `Search` 拥有明确分离的模式语义。
  2. 资源页过滤行为、退出行为与配置含义一致。
  3. 不再保留无效的旧搜索防抖路径。
  4. 相关文档与帮助说明同步更新。

### BR-005 消息 / 日志管理缺少统一审计模型

- 状态: `designing`
- 优先级: `high`
- 症状:
  当前消息通知、底部操作日志、状态提示、正文日志和潜在落盘日志彼此分散，缺少统一事件模型和追溯链路。
- 当前行为:
  顶部通知主要依赖 `ToastMessage/ToastTimer`，底部操作行主要依赖 `InfoMessage/ErrorMessage`，应用本身尚未形成以用户操作为中心的审计日志方案。
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
