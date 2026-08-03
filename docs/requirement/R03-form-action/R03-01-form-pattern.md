# R03-01 Form 公共模式

## 元信息

- 状态: implementing
- 优先级: high
- 来源: [../../task/br-041-unified-form-path-completion.md](../../task/br-041-unified-form-path-completion.md)、[../../task/br-041-form-navigation-followup-prompt.md](../../task/br-041-form-navigation-followup-prompt.md)
- 关联任务:
  - [../../task/br-043-form-action-bar-redesign.md](../../task/br-043-form-action-bar-redesign.md)
  - 运行时刷新缺陷已合并入 C01 + C05,原 BR-042 归档删除
- 关联约束:
  - [../../constraint/C01-form.md](../../constraint/C01-form.md)
  - [../../constraint/C02-dialog.md](../../constraint/C02-dialog.md)
  - [../../constraint/C05-path.md](../../constraint/C05-path.md)

## 目标

建立唯一的公共 Form 状态、渲染、补全来源,所有页面只声明字段配置 + 提交逻辑,不复用输入框或选择列表渲染。

## 用户流程

用户在任何需要参数输入的页面 → Action Bar → 打开 Form → 按 [C01-form](../../constraint/C01-form.md) 规则浏览 → Submit / Cancel。

## UI/UX

详见 [C01-form](../../constraint/C01-form.md) 与 [R03-03](./R03-03-action-dialog-layout.md)。

## 功能规则

- 字段类型:Text / Int / Path / Bool / Select / MultiSelect
- 焦点:Up / Down / Left / Right;Tab / Shift+Tab 仅 Path 补全
- Enter / Space 按 [C01-form](../../constraint/C01-form.md) 规则分发
- 条件字段:DependsOn / DependsEq + RecomputeVisibility
- 路径:Local / Container 区分,详见 [C05-path](../../constraint/C05-path.md)

## 实现设计

- 状态:`internal/tui/state/form.go` `FormState` / `FormField` / `FormPopupState`
- 字段扩展:`DependsOn string` + `DependsEq bool`(条件显示)
- 路径补全:`internal/tui/state/path.go` `PathProvider` / `LocalPathProvider`(第二阶段加 `ContainerPathProvider`)
- 表单渲染:`internal/tui/ui/widget/dialog/form.go` 两列表格外到内
- 弹层:`internal/tui/ui/widget/dialog/form_popup.go` Select / Multi / Path 三种
- 键盘:`internal/tui/keyboard/container_form.go` + `image_form.go` 等各页 form handler
- 第一阶段已覆盖基础;第二阶段需补 Image Save / Load / Import 的 FormPath 复用

## 验收标准

- 任何 Form 打开后默认焦点 Cancel
- Up / Down 跨字段与按钮区导航,Tab 不切焦点
- Select / Multi / Path 三种弹层均有 Home / End / PgUp / PgDn / Space / Enter / Esc 行为
- 长路径光标可见,水平滚动保留
- Path 提交前自动展开 `~` / `$VAR` / 相对路径

## 非目标

- 表单嵌套(主表单里再开子表单)
- 自定义控件(下拉 / 颜色选择器)

## 迁移记录

- 旧文档:br-041/042/043
- 保留信息:FormState / FormField / FormPopupState / LocalPathProvider 设计
- 废弃信息:无
- 待确认状态:`implementing`,br-041 first phase 完成;br-042/043 列出 6 项未实施(路径绝对化、尺寸统一、条件字段、Action 页面重构等),主模型确认后逐项推进。