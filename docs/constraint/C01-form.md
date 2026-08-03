# C01 Form 横切约束

## 适用范围

任何需要参数输入的对话框(容器 Copy / Update / Export / Commit、镜像 Save / Load / Import、Registry Login 等)。涉及 FormState / FormField / FormPopupState 与 `internal/tui/ui/widget/dialog/form.go` + `form_popup.go` 渲染。

## 来源文档

- [../task/br-041-unified-form-path-completion.md](../task/br-041-unified-form-path-completion.md)
- [../task/br-041-form-navigation-followup-prompt.md](../task/br-041-form-navigation-followup-prompt.md)
- BR-042 运行时刷新缺陷已合并入本文档;原 BR-042 归档删除
- [../task/br-043-form-action-bar-redesign.md](../task/br-043-form-action-bar-redesign.md)

## 规则

### 字段类型

- `Text` 自由文本
- `Int` 仅数字
- `Path` 文件路径(分 Local / Container)
- `Bool` 复选框(Space 切换)
- `Select` 单选弹层(Enter 打开)
- `MultiSelect` 多选弹层(Space 切换工作副本,Enter 提交)

### 焦点模型 — 当前实现

- 默认焦点 = Cancel(渲染在底部按钮组左位)
- 字段间移动:`Up` / `Down` 跨字段与按钮区
- 按钮区:`Left` / `Right` 切换 Cancel / Confirm,`Up` 返回最后一字段,`Down` 保持
- `Tab` / `Shift+Tab` 在所有非 Path 字段上为 no-op(不切焦点);Path 字段 Tab / Shift+Tab 用于补全
- 弹层打开时焦点跳层优先级最高(Esc 先关弹层)

### 焦点模型 — 目标规范(BR-043)

- Confirm 与 options 共用焦点环(不再是独立 OnConfirm 槽位)
- 默认焦点 = Cancel
- `Up` / `Down` 跨字段 + Confirm/Cancel 连续编号
- 按钮区 `Down` 保持不循环不执行
- 详见 [../requirement/R03-form-action/R03-03-action-dialog-layout.md](../requirement/R03-form-action/R03-03-action-dialog-layout.md)

### Enter / Space / Tab / Arrow 语义

- `Enter` 普通字段下一字段;Select / MultiSelect 打开弹层;Path 无候选下一字段,有候选应用候选;按钮执行
- `Space` Bool 切换;Text / Path 输入空格;Select / MultiSelect 不直接切换(弹层打开后 Space 在弹层内有效)
- `Tab` Path 单候选直接补全 / 多候选补公共前缀+弹层 / 无候选 toast
- 方向键 参见焦点模型

### 条件字段显示

- 字段可在 `FormField` 上声明依赖(`DependsOn` + `DependsEq`)
- 提交前与 toggle 后调用 `RecomputeVisibility` 重新计算
- 隐藏字段跳过 `MoveField` / `Field()` / 渲染
- 当前实现未启用,目标规范 [R03-01](../requirement/R03-form-action/R03-01-form-pattern.md) 落实

### 错误提示位置

- 字段错误显示在下一行,与 Value 列左边界对齐
- 必填字段为空:Form 内 toast,Form 保持打开
- 提交校验失败:Form 不关闭

### 文本光标规则

- `Left` / `Right` / `Home` / `End` / `Backspace` / `Delete` 走 `editQueryInput`
- 长输入光标可见:水平滚动保留光标位置

### 表单确认 / 取消

- Cancel/Esc 关闭 Form,不清空运行时状态
- Confirm 提交,验证失败保留 Form,验证成功清空并发出 Cmd
- 覆盖类动作(PathSaveFile 已存在):统一 Confirm 弹层,默认 Cancel + Force 选项

### 路径字段 — 当前实现

- Local 路径字段提交前自动展开 `~` / `$VAR` / 相对路径转绝对
- 渲染层 `renderEditableValue` 直接显示 `f.Input.Text`(若用户输入相对路径,显示相对)
- 路径字段每次 keystroke 后**不**自动展开为绝对(目标规范要求)

### 路径字段 — 目标规范(BR-043)

- 路径字段每次 keystroke 后自动展开 `~` / `$VAR` / 相对路径转绝对(详见 [C05-path](./C05-path.md))
- 渲染层显示绝对路径形式
- 详见 [../requirement/R03-form-action/R03-01-form-pattern.md](../requirement/R03-form-action/R03-01-form-pattern.md)

## 与需求文档的引用

- [../requirement/R01-container/R01-01-advanced-ops.md](../requirement/R01-container/R01-01-advanced-ops.md): Copy / Update / Export / Commit 表单
- [../requirement/R01-container/R01-03-stats-top.md](../requirement/R01-container/R01-03-stats-top.md): Wait 运行时
- [../requirement/R02-image/R02-02-import-tarball.md](../requirement/R02-image/R02-02-import-tarball.md): Import 表单
- [../requirement/R03-form-action/R03-01-form-pattern.md](../requirement/R03-form-action/R03-01-form-pattern.md): 公共 Form 模式汇总
- [../requirement/R03-form-action/R03-03-action-dialog-layout.md](../requirement/R03-form-action/R03-03-action-dialog-layout.md): 布局与尺寸

## 待确认项

1. `DependsEq` 仅支持 bool,是否需要支持枚举(如 `PathMode`)。
2. 条件字段触发 Recompute 的时机:Open 时一次性 vs 每次 toggle 都重算。
3. 弹层 Select / MultiSelect 在 Path 字段是否共用 FormPopupState,是否需要拆分为 PathPopup / OptionPopup。
4. 覆盖确认(FORCE)是 Form 内嵌字段还是 ConfirmState 模式(当前是后者)。
5. 路径字段每次 keystroke 后展开为绝对的 cursor 位置调整策略。
6. Update / Commit Form 中其他字段(非 Path)的 Tab 行为:当前为 no-op,目标规范是否需要切焦点(默认不动)。