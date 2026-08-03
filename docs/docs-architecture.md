# dtui 文档真理源整理方案

> 建立日期: 2026-08-02
> 状态: 待执行
> 执行对象: 小模型
> 范围: 仅整理 `docs/` 文档结构,不修改代码

## 1. 核心结论

dtui 文档应按“需求树”组织,不是按一次次 BR/task 堆叠。

最终结构:

```text
总需求清单 -> 大需求 -> 小需求 -> 具体实现任务
```

UI 设计与需求不是完全分离关系。每个需求文档必须包含该需求绑定的 UI/UX 行为;只有多个需求共享的 UI/交互规则才抽取到 `constraint/`。

因此:

- `requirement/` 是需求真理源。
- `constraint/` 是横切 UI/交互/工程约束真理源。
- `task/` 是单次实施卡,不是最终真理源。
- `bugfix-requirements.md` 与 `pending-bugs.md` 是问题台账,修复后必须回写到对应 requirement。
- `architecture.md`、`project-structure.md` 记录代码现状,不承载产品需求。

## 2. 新目录结构

```text
docs/
├── README.md                         # 文档入口
├── requirements.md                   # 总需求清单:索引/状态/优先级
├── architecture.md                   # 当前代码架构
├── project-structure.md              # 当前目录/包职责
├── navigation.md                     # 当前快捷键与导航
├── ui-design.md                      # UI 设计历史与视觉基线
├── i18n.md                           # i18n 规则
├── feature-design.md                 # 历史功能设计总览,后续逐步收敛到 requirement/
├── feature-todo-list.md              # 历史功能待办,后续逐步收敛到 requirement/
├── bugfix-requirements.md            # bug 全量台账
├── pending-bugs.md                   # 当前未关闭 bug 快照
├── ai-prompts.md                     # AI 协作 prompt
├── docs-architecture.md              # 本方案
│
├── requirement/                      # 需求树真理源
│   ├── README.md                     # 需求树索引
│   ├── R01-container/
│   │   ├── README.md                 # 容器大需求总览
│   │   ├── R01-01-advanced-ops.md    # Copy/Update/Diff/Export/Commit/Wait
│   │   ├── R01-02-exec-shell.md      # Exec shell
│   │   └── R01-03-stats-top.md       # Stats/Top/Wait 刷新体验
│   ├── R02-image/
│   │   ├── README.md
│   │   ├── R02-01-history.md
│   │   ├── R02-02-import-tarball.md
│   │   └── R02-03-registry-login.md
│   ├── R03-form-action/
│   │   ├── README.md
│   │   ├── R03-01-form-pattern.md
│   │   ├── R03-02-action-bar.md
│   │   └── R03-03-action-dialog-layout.md
│   ├── R04-runtime/
│   │   ├── README.md
│   │   └── R04-01-docker-podman-capabilities.md
│   └── R05-events-network/
│       ├── README.md
│       └── R05-01-events-network-connect.md
│
├── constraint/                       # 横切约束
│   ├── README.md
│   ├── C01-form.md                   # FORM 字段、焦点、校验、条件显示
│   ├── C02-dialog.md                 # Dialog 尺寸、panel 居中、透明区域
│   ├── C03-table.md                  # 表格列布局、选中态、统计位置
│   ├── C04-keybinding.md             # 快捷键分层、冲突规则
│   ├── C05-path.md                   # Local/Container path、绝对路径、补全
│   └── C06-i18n.md                   # i18n key 与终端宽度约束
│
├── task/                             # 单次实施卡,保留
│   ├── README.md
│   └── br-*.md / task-*.md
│
└── design/                           # 历史设计档案,保留
```

## 3. 文档角色定义

| 位置 | 角色 | 是否真理源 | 说明 |
|---|---|---|---|
| `docs/requirements.md` | 总需求清单 | 是 | 只放需求树索引、状态、优先级、整体路线 |
| `docs/requirement/Rxx-*` | 大需求/小需求 | 是 | 产品目标、用户流程、绑定 UI/UX、验收标准 |
| `docs/constraint/*.md` | 横切约束 | 是 | 多个需求共享的 FORM/Dialog/Table/Key 等规则 |
| `docs/task/*.md` | 实施任务卡 | 否 | 一次执行的上下文、代码索引、prompt、review 记录 |
| `docs/bugfix-requirements.md` | bug 台账 | 否 | 问题来源与修复记录;修复后回写 requirement |
| `docs/pending-bugs.md` | bug 快照 | 否 | 当前未关闭问题列表 |
| `docs/architecture.md` | 实现架构 | 是 | 当前代码事实,但不表达需求 |
| `docs/ui-design.md` | UI 历史/基线 | 半真理源 | 可引用,但新规则优先收敛到 `constraint/` |
| `docs/feature-design.md` | 历史功能设计 | 迁移来源 | 后续逐步拆进 `requirement/` |
| `docs/feature-todo-list.md` | 历史功能待办 | 迁移来源 | 后续逐步拆进 `requirements.md` 与 `requirement/` |

## 4. 单个需求文档模板

每个小需求文档必须使用以下结构。

```markdown
# Rxx-yy 需求标题

## 元信息

- 状态: planned | implementing | implemented | blocked | cancelled
- 优先级: high | medium | low
- 来源: BR-xxx / TASK-xxx / 用户反馈
- 关联任务: ../../task/xxx.md
- 关联约束:
  - ../../constraint/C01-form.md
  - ../../constraint/C02-dialog.md

## 目标

说明用户为什么需要这个能力。

## 用户流程

从哪个页面进入,触发什么动作,看到什么反馈,如何退出。

## UI/UX

只写本需求特有的界面行为。
通用 FORM/Dialog/Table/快捷键规则不要复制,只引用 constraint。

## 功能规则

输入、状态、边界条件、错误处理、运行时差异。

## 实现设计

涉及模块、调用链、关键状态、runtime API。

## 验收标准

可验证的最终行为,必须能转化为测试或人工检查。

## 非目标

明确本需求不解决的内容。

## 迁移记录

从哪些旧文档合并而来,哪些信息被保留/废弃。
```

## 5. 大需求目录规则

每个 `Rxx-*` 目录必须包含 `README.md`。

`README.md` 只做索引和状态汇总,不堆实现细节。

模板:

```markdown
# Rxx 大需求标题

## 范围

一句话定义大需求边界。

## 子需求

| 编号 | 标题 | 状态 | 优先级 | 来源 |
|---|---|---|---|---|
| Rxx-01 | xxx | implemented | high | BR-xxx |

## 关联约束

- ../../constraint/Cxx-xxx.md

## 关联任务卡

- ../../task/br-xxx.md
```

## 6. 横切约束文档规则

约束文档只写跨需求通用规则,不写单个功能的业务流程。

### C01 Form

应包含:

- 字段类型: Text / Int / Path / Bool / Select / MultiSelect。
- 焦点模型。
- Enter / Space / Tab / Arrow 语义。
- 条件字段显示规则。
- 错误提示位置。
- 文本光标规则。
- 表单确认/取消规则。

来源:

- `task/br-041-unified-form-path-completion.md`
- `task/br-041-form-navigation-followup-prompt.md`
- BR-042 已合并归档;原文档删除
- `task/br-043-form-action-bar-redesign.md`

### C02 Dialog

应包含:

- Dialog 在当前 panel 内居中。
- 四周透明。
- 默认尺寸规则。
- Action 展示框统一为 panel 等高、宽度 3/4。
- 小窗口降级规则。

来源:

- BR-040 已合并归档;原文档删除
- `task/br-040-implementation-decisions.md`
- `task/br-043-form-action-bar-redesign.md`

### C03 Table

应包含:

- 表格列布局必须占满可用宽度。
- 列从外到内计算。
- 选中行整行高亮。
- 同级统计信息放右上角。
- 斑马线、颜色、列样式配置来源。

来源:

- 表格相关历史对话要求。
- `internal/tui/tables/*.jsonc` 当前配置。
- `internal/tui/ui/component/table*.go` 当前实现。

### C04 Keybinding

应包含:

- 全局快捷键、panel 快捷键、dialog 快捷键、popup 快捷键分层。
- Action Bar 只承载复杂操作,不承载全局命令。
- 直接快捷键与 Action Bar 的关系。
- R / C 等 table 页面链接型快捷键的待决策状态。

来源:

- `task/br-039-action-bar-replace-q.md`
- `task/br-039-a-implementation-design.md`
- `task/br-043-form-action-bar-redesign.md`

### C05 Path

应包含:

- Local path 与 Container path 的区别。
- 默认目录是 dtui 启动时 shell CWD,不是 binary 目录。
- 保存类路径最终必须绝对化。
- Container path 的 Tab/Arrow/Enter 语义。
- Local destination 的默认文件名生成规则。

来源:

- `task/br-041-unified-form-path-completion.md`
- `task/br-043-form-action-bar-redesign.md`

### C06 i18n

应包含:

- i18n key 命名空间。
- 缺失 key 回退。
- 终端宽度与中日韩字符宽度。
- FORM/Dialog/ActionBar 文案约束。

来源:

- `docs/i18n.md`
- `internal/data/i18n/lang/*.jsonc`

## 7. 旧文档迁移映射

第一阶段不要 `git mv` 大量文件。先新建真理源文档并引用旧文档。

| 旧文档 | 新归属 | 处理方式 |
|---|---|---|
| `task/br-033-task019-advanced-container-actions.md` | `requirement/R01-container/R01-01-advanced-ops.md` | 摘要合并,保留旧文档链接 |
| `task/br-033-advanced-container-ops-analysis.md` | `R01-01` 实现设计 | 摘要合并 |
| `task/br-033-small-model-implementation-prompt.md` | `task/` | 保留 |
| `task/br-033-035-review-fixes.md` | `task/` | 保留 |
| `task/br-034-image-history-top-page.md` | `requirement/R02-image/R02-01-history.md` | 摘要合并 |
| `task/br-034-runtime-analysis.md` | `R02-01` 实现设计 | 摘要合并 |
| `task/br-034-ui-state-analysis.md` | `R02-01` UI/状态设计 | 摘要合并 |
| `task/br-034-implementation-completed.md` | `R02-image/README.md` | 状态汇总 |
| `task/br-035-events-panel-and-network-connect.md` | `requirement/R05-events-network/R05-01-events-network-connect.md` | 摘要合并 |
| `task/br-039-action-bar-replace-q.md` | `requirement/R03-form-action/R03-02-action-bar.md` | 摘要合并 |
| `task/br-039-a-implementation-design.md` | `R03-02` 实现设计 | 摘要合并 |
| `task/br-039-code-location-analysis.md` | `R03-02` 代码索引 | 摘要合并 |
| `task/br-039-main-model-feedback.md` | `task/` | 保留 |
| `task/br-040-code-location-analysis.md` | `R03-03` 代码索引 | 摘要合并 |
| `task/br-040-implementation-decisions.md` | `C02` | 摘要合并 |
| `task/br-041-unified-form-path-completion.md` | `requirement/R03-form-action/R03-01-form-pattern.md` + `C01` + `C05` | 拆分摘要 |
| `task/br-041-form-navigation-followup-prompt.md` | `C01` + `C05` | 摘要合并 |
| `task/br-043-form-action-bar-redesign.md` | `R03-03` + `C01` + `C02` + `C04` + `C05` | 拆分摘要 |
| `task/task-020-podman-capabilities-evaluation.md` | `requirement/R04-runtime/R04-01-docker-podman-capabilities.md` | 摘要合并 |
| `task/task-020-capabilities-review.md` | `R04-01` | 摘要合并 |

## 8. 第一阶段执行任务

小模型按以下顺序执行。

### 8.1 创建目录

创建:

```text
docs/requirement/
docs/requirement/R01-container/
docs/requirement/R02-image/
docs/requirement/R03-form-action/
docs/requirement/R04-runtime/
docs/requirement/R05-events-network/
docs/constraint/
```

### 8.2 创建索引

新增:

- `docs/requirement/README.md`
- `docs/requirement/R01-container/README.md`
- `docs/requirement/R02-image/README.md`
- `docs/requirement/R03-form-action/README.md`
- `docs/requirement/R04-runtime/README.md`
- `docs/requirement/R05-events-network/README.md`
- `docs/constraint/README.md`

### 8.3 创建约束文档

新增:

- `docs/constraint/C01-form.md`
- `docs/constraint/C02-dialog.md`
- `docs/constraint/C03-table.md`
- `docs/constraint/C04-keybinding.md`
- `docs/constraint/C05-path.md`
- `docs/constraint/C06-i18n.md`

要求:

- 只摘要迁移,不要复制整篇旧文档。
- 每个约束文档必须列出“适用范围”“规则”“来源文档”“待确认项”。
- 不要删除旧文档。

### 8.4 创建第一批需求文档

新增:

- `docs/requirement/R01-container/R01-01-advanced-ops.md`
- `docs/requirement/R01-container/R01-02-exec-shell.md`
- `docs/requirement/R01-container/R01-03-stats-top.md`
- `docs/requirement/R02-image/R02-01-history.md`
- `docs/requirement/R02-image/R02-02-import-tarball.md`
- `docs/requirement/R02-image/R02-03-registry-login.md`
- `docs/requirement/R03-form-action/R03-01-form-pattern.md`
- `docs/requirement/R03-form-action/R03-02-action-bar.md`
- `docs/requirement/R03-form-action/R03-03-action-dialog-layout.md`
- `docs/requirement/R04-runtime/R04-01-docker-podman-capabilities.md`
- `docs/requirement/R05-events-network/R05-01-events-network-connect.md`

要求:

- 需求文档使用第 4 节模板。
- 内容先做摘要,每篇控制在可读范围。
- 必须保留旧 task 链接。
- 状态必须从旧文档推断,不确定写 `planned` 或 `pending-review`,不要猜 `done`。

### 8.5 更新入口索引

修改:

- `docs/README.md`
- `docs/requirements.md`
- `docs/task/README.md`

要求:

- `docs/README.md` 增加 `requirement/` 与 `constraint/` 入口。
- `docs/requirements.md` 改为总需求清单,链接到 `requirement/README.md`。
- `docs/task/README.md` 明确说明 `task/` 是实施卡,不是需求真理源。
- 不删除原有历史说明,只增加迁移后的入口与解释。

## 9. 执行限制

小模型必须遵守:

- 不修改任何 Go 代码。
- 不移动旧 `docs/task/*.md` 文件。
- 不删除旧文档。
- 不创建 git commit。
- 不使用网络。
- 只做文档新增和索引更新。
- 如发现旧文档互相冲突,在新文档写“待确认”,不要自行裁决。

## 10. 验证

执行后运行:

```bash
git diff --check
```

可选:

```bash
rg -n "feature/" docs/docs-architecture.md docs/README.md docs/requirements.md docs/task/README.md
```

确认新方案中不再把 `feature/` 作为目标目录。

## 11. 主模型审查重点

主模型审查时重点检查:

1. `requirement/` 是否形成清晰树形结构。
2. 每个需求是否包含绑定 UI/UX,但没有复制通用约束。
3. `constraint/` 是否只放跨需求规则。
4. `task/` 是否仍可作为实施记录被追溯。
5. 旧文档链接是否完整。
6. 状态是否保守,没有把未验证内容标成 done。
