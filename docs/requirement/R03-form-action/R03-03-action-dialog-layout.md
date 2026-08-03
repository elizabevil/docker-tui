# R03-03 Action 展示框布局

## 元信息

- 状态: planned
- 优先级: high
- 来源: 原 BR-040 已合并入本文档 + C02,BR-043 保留作为问题清单原始记录
- 关联任务:
  - [../../task/br-040-code-location-analysis.md](../../task/br-040-code-location-analysis.md)
  - [../../task/br-040-implementation-decisions.md](../../task/br-040-implementation-decisions.md)
- 关联约束:
  - [../../constraint/C02-dialog.md](../../constraint/C02-dialog.md)
  - [../../constraint/C01-form.md](../../constraint/C01-form.md)
  - [../../constraint/C06-i18n.md](../../constraint/C06-i18n.md)

## 目标

统一 Action 展示框(Confirm / Form / Action Bar)在 panel 之上的视觉与交互,确保:
- 尺寸:宽 3/4,等高
- 布局:flex 上下均匀分布
- 焦点:Confirm 与 options 共享同一焦点环
- 内容:标题 + 当前操作对象 + option list + Confirm + 快捷键提示

## 用户流程

用户进入任意 Action 展示框 → 默认焦点在 Cancel/最末位 → Up / Down 跨字段与按钮 → Enter 提交 → Esc 取消。

## UI/UX

按 flex 上下均匀分布:

1. **标题**(左对齐)
2. **当前操作对象**(例 `Container: web (abc123)`)— 来源 `m.Form.TargetName + (TargetID[:12])`
3. **Option list**(Fields)— 主交互区
4. **Confirm box**(与 options 共用焦点,不在底部独立槽位)
5. **快捷键提示**(底部 dim 文本)

## 功能规则

- 默认焦点在 Cancel 槽位
- Up / Down 跨字段与按钮连续编号
- Left / Right 在按钮区切换 Cancel / Confirm
- Enter 在 Confirm 行执行,在 Cancel 行关闭
- 弹层(Esc / Confirm / Cancel)走 [C01-form](../../constraint/C01-form.md)

## 实现设计

- 焦点重构:`FormState.OnConfirm` + `FieldFocus == -1` 双语义合并为单一行号
- 方案 A:`FormFieldKind` 新增 `FormConfirm` / `FormCancel` 两种不可编辑行
- 方案 B:`FormSpec` 加 `ConfirmLabel` / `CancelLabel` 字段,渲染为列表末两行
- 渲染:`FormDialog` 改为 `lipgloss.JoinVertical(lipgloss.Center, ...)`
- 尺寸:`dialogWidth/Height` 在 panel 上下文直接用 `body.Width * 3/4` 与 `body.Rows`

## 验收标准

- 所有 Action 展示框宽 3/4 + 等高
- Confirm 与 options 共用焦点环,Enter / Up / Down 一致
- 小窗口(40x16)下不重叠
- 中日韩标签对齐一致

## 非目标

- 自适应宽度(只固定 3/4)
- 多列布局

## 迁移记录

- 旧文档:BR-040 已合并入本文档 + C02,BR-040 归档删除;[task/br-043-form-action-bar-redesign.md](../../task/br-043-form-action-bar-redesign.md) 保留作为 6 项问题原始面谈纪要
- 保留信息:BR-040 居中 + 透明规则;BR-043 flex 布局 + Confirm 共享焦点方向
- 待确认状态:`planned`,方案 A/B 待主模型决策;主模型确认后启动实施。