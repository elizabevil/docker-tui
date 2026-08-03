# GLOBAL — 全局信息(决策日志 + 待确认问题)

> 上次更新: 2026-08-03
> 范围: append-only,时间倒序排列

## 决策日志(主模型 → 全体)

### 2026-08-03 — BR-043 §3.2 高度修订 + Confirm/Cancel 同行渲染

- **决策**:选择框高度从 panel 等高修订为 panel * 3/4(与宽度对称);Confirm/Cancel 从方案 B 的列表末尾两行改为同行左右排版
- **理由**:用户 2026-08-03 反馈——BR-043 §8.2 原话"等高+宽为 3/4"应解读为"宽 3/4、高 3/4"两轴对称;Confirm/Cancel 同行布局更符合 flex 视觉对齐
- **改动**:
  - `panelDialogHeight`: `bodyH` → `bodyH * 3 / 4`(仍 clamp cfg.MinHeight/MaxHeight)
  - `FormDialog` 渲染:移除两个独立 `renderBtnRow` 调用,合并为单行左 Cancel + 右 Confirm,焦点模型不变(Confirm=n、Cancel=n+1 线性)
- **影响**:所有 `Render*InPanel` 调用(FormDialog/ExecDialog/SelectionDialog)高度变为 panel * 3/4;小窗口(40x16)下 height=12,需确认无挤压
- **验证**:form/dialog/state/keyboard/keys 包测试全过

### 2026-08-03 — BR-043 批次 A-D 实施完成

- **决策**:R2 批次按 GLOBAL.md § 2026-08-03 BR-043 六个问题决策实施
- **交付**:4 个 feat commits
  - `1731789` 批次 A:路径绝对化(blur-time) + 面板尺寸统一(3/4 宽 + 等高)
  - `8ec0e4d` 批次 B:Commit form 条件字段(DependsOn/DependsEq/Hidden)
  - `9e0078f` 批次 C:Action 页面布局重构(方案 B + 共享焦点 + 列表行渲染)
  - `29d8cfe` 批次 D:C 键连接信息迁 `:conn-info` command palette
- **范围**:12 个文件改动(+/- ~440 行)
- **未实施 followup**:3.4 配置拆分、per-form DefaultFocus 配置、DependsValue 枚举扩展
- **验证**:`go vet` + `go test` state/keyboard/keys/dialog 全过;`git diff --check` 通过

### 2026-08-03 — BR-043 六个问题决策

**范围**:`docs/task/br-043-form-action-bar-redesign.md` §7 五项待确认项;3.4 配置拆分不在本轮,后续另起任务卡。

**逐项决策**:

1. **3.1 路径绝对化触发时机 → on field-exit (blur)**,不在每次 keystroke。
   - 理由:per-keystroke 会干扰 Tab/Ctrl+Space 路径补全(`~/` 候选立即被改写为 `/home/...` 丢失补全前缀),且打断光标位置计算。
   - 实现位置:`internal/tui/keyboard/container_form.go` `handleContainerFormKey` 中,Tab/ShiftTab/Down/Up/Enter 离开 FormPath 字段时调用 `state.Absolute(text, m.Form.CWD)` 写回并调整 cursor;提交时 `normalizeSaveDestination` 仍是兜底。
   - 状态:仅读 `internal/tui/state/path.go`,不改。

2. **3.3 Action 页面重构方案 → 方案 B**(FormSpec 加 ConfirmLabel/CancelLabel 字段,渲染为末尾两行)。
   - 理由:方案 A 新增 `FormFieldKind FormConfirm/FormCancel` 会污染 Fields 迭代(需跳过 Confirm/Cancel 数据收集);方案 B 将 Confirm/Cancel 作为末尾两行,与 options 共用同一焦点 `0..n+1`,改动最小,符合用户设计"Confirm box 与 options 共用焦点"。
   - 同步变更:删除 `FormState.OnConfirm` 双语义,焦点统一为 row index;`MoveField`/`MoveButton` 合并为单一线性 Move。
   - 影响文件:`internal/tui/state/form.go` + `internal/tui/ui/widget/dialog/form.go` + `internal/tui/keyboard/container_form.go` + 键盘测试。

3. **3.3 默认焦点 → 保持 Cancel**,per-form 配置留作 follow-up。
   - 理由:安全默认,避免误触 Confirm;per-form `DefaultFocus` 可后续追加(本轮不实施)。

4. **3.5 DependsEq 类型 → bool only**。
   - 理由:当前唯一依赖关系 `archivePath` 依赖 `exportTar`(bool);无枚举依赖需求。
   - 实现:`FormField.DependsOn string` + `FormField.DependsEq bool` + `FormState.RecomputeVisibility()`。
   - 枚举/字符串比较作为 future work 留待 `DependsValue any` 扩展。

5. **3.6 R/C 键归宿**:
   - **R 保留为全局 refresh**(`ActionRefresh`)。理由:R 是标准操作且 README 已文档化(`r Refresh all resources`),非"链接型"快捷键;移除会破坏既有契约。
   - **C 取消连接信息 fallback**:从 `handleShortcuts` 移除 `KeyC → showConnectionInfo`;连接信息迁移到 `:` command palette(`conn-info` 命令)。理由:C 在 volumes/networks 上下文是 CREATE(非 link 型,保留);containers/images/compose 上的 C=连接信息才是 link 型,按用户"取消 table 链接键"要求迁出快捷键层。
   - 影响文件:`internal/tui/keyboard/keyboard.go` + command palette 注册表 + README/docs 同步。

**实施批次**(本轮范围):
| 批次 | 内容 | 风险 |
|---|---|---|
| A | 3.1 路径绝对化 + 3.2 尺寸统一(panel 3/4 宽 + 等高) | 低 |
| B | 3.5 Commit form 条件字段(DependsOn/DependsEq/Hidden) | 中 |
| C | 3.3 Action 页面布局重构(方案 B + 共享焦点) | 中高 |
| D | 3.6 C 键连接信息迁 command palette | 低 |

**不在本轮**:3.4 配置拆分(后续另起任务卡);per-form DefaultFocus 配置;DependsValue 枚举扩展。

### 2026-08-03 — R1 批次主模型审查与 commit

- 决策:R1 文档批次审查通过并 commit;审查中修复 5 处悬空 task 引用 + 13 处跨组断链
- 理由:子模型自报"残留引用 0"与事实不符 — 实查 `grep -rn` 发现 `docs/task/README.md`、br-039-a、br-039-code-location、br-040-code-location、br-041 共 5 处指向已删 task 的链接;另修复 requirement 文档跨组链接缺 `R03-form-action/` 前缀、`docs/feature-design.md` 冗余前缀、R04 相对深度错误
- 验证:`python3` 全仓 md 链接校验 479 个内部链接全部可解析(`internal/*` 代码锚点 109 个为 bugfix 文档约定,非断链);`git diff --check` 通过
- 影响范围:`docs/` 14 个文件链接修复 + handoff 状态更新

### 2026-08-03 — handoff 目录建立

- 决策:新建 `docs/handoff/` 目录,采用 GLOBAL/STATE/COMPLETED 三文件 + README 索引
- 理由:`ai-prompts.md` 是会话内粘贴的模板,跨会话持续状态/决策/问题需要独立文档
- 影响范围:仅新建,不动其他文件

### 2026-08-03 — 删除 6 个冗余 task 文件

- 决策:删除 `next-small-model-brief.md` + `br-042-form-runtime-refresh-corrections.md` + `br-036-image-import-tarball.md` + `br-037-registry-login.md` + `br-038-exec-shell-ui.md` + `br-040-dialog-style-center-on-panel.md`
- 理由:内容已合并入 `requirement/R##-##/` 与 `constraint/C##.md`
- 验证:`grep -rn <删文件名> docs/` 残留引用为 0;`git diff --check` 通过
- 影响范围:13 个反向引用文件已修复链接

### 2026-08-02 — 文档真理源重组

- 决策:按 [docs-architecture.md](../docs-architecture.md) 建立需求树 `requirement/R##/` + 横切约束 `constraint/C##/`
- 理由:旧 task/ 散乱 BR 文档无法支持"按大需求点查询";requirement/ + constraint/ 提供清晰的真理源分层
- 影响范围:新建 25 个文档,更新 3 个入口索引,保留全部旧 task/ 作为迁移来源(后删除冗余)

### 2026-08-01 — BR-034 Image History 顶层页

- 决策:Images 页 H 键 / Action Bar 触发独立 History 顶层页
- 理由:大镜像 >50 layer 滚动卡顿;独立页配 EnsureVisible + 可见行切片解决
- 状态:implementing(详见 [requirement/R02-image/R02-01-history.md](../requirement/R02-image/R02-01-history.md))

## 待确认问题(子模型 → 主模型)

> 状态: open(子模型提的) / answered(主模型答完)

### 2026-08-03 — handoff 命名 (状态: open)

- 问题:子模型问:目录名 `handoff/` 是否合适?或用 `coordination/` / `comm/`?
- 答复:见 [README.md 待确认 § 命名](./README.md#待确认)
- 影响范围:仅命名,不动结构

### 2026-08-02 — BR-041 状态标记 (状态: answered)

- 问题:子模型问:`requirement/R03-form-action/R03-01-form-pattern.md` 状态写 `implementing` 还是 `implemented`?旧 br-041 task 文档说"第一阶段完成",但 [task/br-043-form-action-bar-redesign.md](../task/br-043-form-action-bar-redesign.md) 列出 6 项未实施项
- 答复:状态保持 `implementing`(第一阶段完成 + 仍有 followup);R03-01 README 待确认项已说明此差异
- 影响范围:无