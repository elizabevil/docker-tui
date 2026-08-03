# BR-043: Form / Action Bar 现有问题整理与设计方向

> 日期: 2026-08-02
> 状态: `pending` — 文档归档,未实现
> 类型: 允许修改代码(暂未实施,等主模型决策)
> 提交规则: 不 commit、不 push
> 范围: 仅归档;本任务卡本身不写代码

## 1. 背景

BR-041 第一阶段完成后,在实际使用与代码走读中发现以下六个待整理的问题,涉及路径语义、布局尺寸、Action 页面设计、配置拆分、表单条件字段、以及 table 页面快捷键。本任务卡归档这些发现并提出待决策的设计方向,**不实施**。

来源:用户 2026-08-02 反馈,经主模型当面讨论后记录。

## 2. 现状索引

按问题分别列出相关代码位置,作为后续决策与实施的参考。

| 项 | 文件 | 关键位置 |
|---|---|---|
| 1. 路径绝对化 | `internal/tui/keyboard/container_form.go` | `workingDir()` L472,default 分支 |
| 1. 路径绝对化 | `internal/tui/state/path.go` | `Absolute()` L156,`ExpandPath()` L142 |
| 1. 路径绝对化 | `internal/tui/ui/widget/dialog/form.go` | `renderEditableValue()` L142-168 |
| 2. 尺寸统一 | `internal/tui/ui/widget/dialog/view.go` | `dialogWidth()` L111,`dialogHeight()` L127,`RenderXxxOverlayInPanel` L300+ |
| 2. 尺寸统一 | `internal/tui/ui/widget/dialog/dialog.jsonc` | 默认 `WidthPercent=25`、`HeightPercent=30` |
| 3. Action 重构 | `internal/tui/ui/widget/dialog/form.go` | `FormDialog()` L19,`renderFormField()` L94 |
| 3. Action 重构 | `internal/tui/keyboard/container_form.go` | `handleContainerFormKey()` L142,`MoveField/MoveButton` |
| 4. 配置拆分 | `internal/data/config/default.jsonc` | 单文件 embed 默认值 |
| 4. 配置拆分 | `internal/tui/ui/widget/dialog/dialog.jsonc` | 样式单文件 |
| 5. 条件字段 | `internal/tui/keyboard/container_form.go` | `openContainerCommitForm()` L217-256 |
| 5. 条件字段 | `internal/tui/state/form.go` | `FormField` 缺 `DependsOn`/`Hidden` |
| 6. 链接键 | `internal/tui/keyboard/keyboard.go` | `handleShortcuts()` L243 |
| 6. 链接键 | `internal/tui/keys/registry.go` | `ActionRefresh`、`ActionConnInfo` 等动作注册 |

## 3. 六个待整理问题

### 3.1 FORM 路径语义(中等风险)

**问题**:
- `workingDir()` 已使用 `os.Getwd()`(用户 shell CWD,非 binary 所在目录),提交前 `normalizeSaveDestination()` 会把 `~`/`$VAR`/相对路径展开成绝对路径。**CWD 部分已正确**。
- 但渲染层 `renderEditableValue` 直接显示 `f.Input.Text`,若用户手工输入相对路径(如 `output.tar`),界面就显示相对形式而非绝对形式。

**验收方向**(主模型确认后实施):
- 在 path 字段每次 `editQueryInput` 之后,若输入文本不是绝对路径,运行 `state.Absolute(text, m.Form.CWD)` 写回 `f.Input.Text` 并相应调整 `Cursor`。
- 提交前的 `normalizeSaveDestination` 仍是兜底。
- 涉及文件:`internal/tui/keyboard/container_form.go` + `internal/tui/state/path.go`(只读,不改)。

### 3.2 Action 显示框尺寸统一(低风险)

**问题**:`dialogWidth/Height` 使用百分比配置(默认 25% 宽 / 30% 高),不同 action 类型尺寸不一致。

**验收方向**:
- 在 `RenderContainerFormOverlayInPanel`、`RenderOverlayInPanel`、`RenderExecOverlayInPanel` 调用处,直接用 `body.Width * 3 / 4` 作为 dialog 宽,`body.Rows` 作为 dialog 高。
- 保留 `dialogConfig` 作为默认值兜底,但 UI 行为固化到 panel。
- `dialogWidth/Height` 仅用于无 panel 上下文的 dialog(如有遗留)。

涉及文件:`internal/tui/ui/widget/dialog/view.go` + `internal/tui/ui/widget/dialog/dialog.jsonc`。

### 3.3 Action 页面布局重构(中高风险,**待主模型决策**)

**当前布局**(`FormDialog`):
- 标题
- 空行
- Fields
- 空行
- Buttons(Confirm/Cancel,独立焦点槽位 `OnConfirm` + `FieldFocus == -1`)
- 空行
- Hint

**用户设计**(flex 上下均匀分布,Confirm 与 options 共用焦点):
- 标题
- 当前操作对象(例 `Container: web (abc123)`)— 取自 `m.Form.TargetName + (TargetID[:12])`
- Option list(Fields)— 主交互区
- Confirm box(**与 options 共用焦点**,不再是独立底部按钮)
- 快捷键提示

**关键设计变更**(需主模型决策):
- 删除独立的 Confirm/Cancel 槽位(`FormState.OnConfirm` 标志、`FieldFocus == -1` 双语义)。
- 选项:方案 A `FormFieldKind` 新增 `FormConfirm`/`FormCancel` 两种不可编辑行;方案 B 在 `FormSpec` 加 `ConfirmLabel`/`CancelLabel` 字段,渲染为列表末两行。
- 焦点跳转变连续:取消"`-1` 是按钮"语义,改为 row index 0..n。
- Enter 在 Confirm 行执行、在 Cancel 行关闭、在普通字段下一行。
- 默认焦点仍是 Cancel(可配)。

**当前不做**:先出设计稿 + 焦点伪代码,主模型决策后再实施。

涉及文件:`internal/tui/state/form.go` + `internal/tui/ui/widget/dialog/form.go` + `internal/tui/keyboard/container_form.go` + 大量测试。

### 3.4 配置文件拆分(后续设计,**本轮不实施**)

**当前问题**:样式配置全部集中在 `internal/tui/ui/widget/dialog/dialog.jsonc` 与 `internal/data/config/default.jsonc`,功能增加后单文件过长,维护与覆盖策略不清晰。

**设计方向**(后续另起任务卡):
- 拆分为 `internal/ui/styles/*.jsonc` 多文件:`dialog.jsonc`/`table.jsonc`/`form.jsonc`/`colors.jsonc` 等。
- 用户配置目录 `~/.config/dtui/styles/` 按相同文件名覆盖。
- 加载优先级:embed 默认 < 用户文件 < CLI 标志。
- 本任务卡只记录意图,不写代码。

### 3.5 FORM 条件字段(中等风险)

**当前状态**:`openContainerCommitForm` 在 `internal/tui/keyboard/container_form.go:233-241` 创建 7 个 fields:`repository/tag/author/comment/pause/exportTar/archivePath`。**`archivePath` 当前无条件显示**,应只在 `exportTar == true` 时显示。

**验收方向**:
- `state.FormField` 新增 `DependsOn string` + `DependsEq bool`(默认 bool 期望值)。
- `state.FormState` 新增 `RecomputeVisibility()`:遍历 Fields,根据依赖字段当前值设置内部 `Hidden`。
- `FormState.Open` 调用 `RecomputeVisibility` 初始化;Bool 字段 toggle 时重新计算。
- `MoveField` / `Field()` / 渲染层跳过 `Hidden` 字段。
- Commit form 中 `archivePath` 配置 `DependsOn: fieldExportTar, DependsEq: true`。

涉及文件:`internal/tui/state/form.go` + `internal/tui/state/form_test.go` + `internal/tui/keyboard/container_form.go` + `internal/tui/ui/widget/dialog/form.go`。

### 3.6 Table 页面链接快捷键(待主模型讨论)

**当前状态**:
- `KeyC` → `showConnectionInfo`(全局)。
- `KeyR` → `ActionRefresh`(全局,但 BR-039 之后 Q 已被 ActionBar 替换,R 与 refresh 是否仍由快捷键直接触发待评估)。
- 用户判断:table 页面里的"链接型"快捷键 R 与 C 应取消或重新定义。

**待主模型决定**:
- R(刷新)保留为全局 refresh,还是仅在 panel 内 refresh?
- C(连接信息)保留还是移除?
- 是否在 ActionBar(`:` command palette)里承载?
- 表格行内是否引入新链接语义(如 `Enter` 之外的双键操作)?

本任务卡**不实施**,等决策后单独立项。

## 4. 建议任务分解(供后续主模型决策)

按工作量与风险排序,仅在主模型确认方向后实施:

| 批次 | 内容 | 工作量 | 依赖 |
|---|---|---|---|
| A | 3.1 路径绝对化 | 小 | 无 |
| A | 3.2 尺寸统一 panel 3/4 宽 + 等高 | 小 | 无 |
| B | 3.5 Commit form 条件字段 | 中 | state FormField 扩展 |
| C | 3.3 Action 页面布局重构 | 大 | 主模型决策 |
| C | 3.4 配置拆分 | 中 | 设计稿 |
| D | 3.6 table 链接键讨论结果 | 待定 | 主模型决策 |

## 5. 验收标准

由于本任务卡不实施,无验收门槛;主模型决策后,各批次按现有 BR 文档惯例设独立验收。

## 6. 风险

- 3.1 路径绝对化:与现有 path completion / cursor 位置计算耦合,需回归覆盖 `Tab`/`Enter`/`Ctrl+Space` 等键盘路径。
- 3.2 尺寸统一:小窗口(40x16)下可能挤压字段;需回归 40/80/160 三档 viewport。
- 3.5 条件字段:`Hidden` 与 `FieldFocus` / 提交校验 / popup dispatch 都要同步跳过隐藏字段。
- 3.3 布局重构:焦点模型改变会触发大量键盘测试需要更新。
- 3.6 链接键:与 ActionBar(`:` palette)与全局快捷键冲突。

## 7. 待确认项(需主模型决定)

1. 3.1 路径绝对化的触发时机(每次 keystroke vs 提交前)。
2. 3.3 Action 页面重构用方案 A(新增 FormConfirm/FormCancel Kind)还是方案 B(FormSpec 加 Confirm/Cancel 字段)。
3. 3.3 默认焦点是否仍为 Cancel。
4. 3.5 `DependsEq` 仅支持 bool,还是需要支持枚举(如 `PathMode` 匹配)。
5. 3.6 R / C 的最终归宿。

## 8. 主模型面谈纪要(原文)

> 1。容器页面,action 中 FORM 表单所有路径均为绝对路径,默认目录当前执行目录而非应用本身所在目录
> 2. 由于功能不同 action 展示框 大小也不同,现在统一定义为 panel 页面的等高+宽为 3/4。
> 3. action 页面展示设计:标题,当前操作对象(例如某容器,某镜像),选项列表,确认框(注意与选项共用一个焦点),快捷键提示信息。按类似 flex 上下均匀分布 。
> 4,设计问题, 默认配置项均采用 jsonc+embed 嵌入方式,同时提供配置文件(由于样式配置太多,因此需要拆分文件,后续设计)
> 5.FORM 表单一些项依赖于其他选项,例如 Commit container as image 中 Export image as tar [ ],为 true 时才应该展示 Archive destination。 这类逻辑需要梳理。
> 5.取消 table 页面所有链接快捷键,例如 R 与 C,按键功能待讨论。

## 9. 完成情况

- [x] 归档现状分析
- [x] 标注代码索引
- [x] 提出待确认项
- [ ] 实施任何代码(等主模型决策)