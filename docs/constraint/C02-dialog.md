# C02 Dialog 横切约束

## 适用范围

任何覆盖在 panel 之上的对话框(Confirm、Form、Select、Exec、ActionBar、Help、Toast)。

## 来源文档

- [../task/br-040-implementation-decisions.md](../task/br-040-implementation-decisions.md)
- BR-040 已合并入本文档 + R03-03,原 BR-040 归档删除
- [../task/br-043-form-action-bar-redesign.md](../task/br-043-form-action-bar-redesign.md)

## 规则

### 居中与透明 — 当前实现

- Dialog 必须相对当前激活的 panel 居中(`PlaceDialogInPanel`)
- Dialog 之外区域显示原 panel 内容(不做全局遮罩)
- Overlay 颜色:可配置,默认半透明(用户视觉保留上下文)

### 默认尺寸 — 当前实现

- 宽度 / 高度:`dialogConfig` 提供 `WidthPercent` / `HeightPercent`(默认 25 / 30)+ `MinWidth` / `MinHeight` 兜底
- 不同 action 类型尺寸不一(差异来源:`view.go:111` `dialogWidth` 与 `dialogHeight`)

### 默认尺寸 — 目标规范(BR-043)

- 宽度:panel 宽度的 3/4(`body.Width * 3 / 4`)
- 高度:panel 高度(等高)
- 小窗口降级:`MinWidth` / `MinHeight` 兜底仍保留
- 详见 [../requirement/R03-form-action/R03-03-action-dialog-layout.md](../requirement/R03-form-action/R03-03-action-dialog-layout.md)

### 标题与对齐 — 当前实现

- Dialog 标题左对齐(走 `panelTitle` 样式)
- 内容区域根据类型走不同对齐:Form Label 右对齐 + Value 左对齐;Confirm 行居中

### 标题与对齐 — 目标规范(BR-043)

- Action 展示框 flex 上下均匀分布:
  - 标题(左对齐)
  - 当前操作对象(例 `Container: web (abc123)`)
  - Option list(Fields,主交互区)
  - Confirm box(与 options 共享焦点)
  - 快捷键提示(底部 dim 文本)

### 按钮组 — 当前实现

- Confirm / Cancel 整组在底部居中;组内 Cancel 在左、Confirm 在右(已修正)
- 聚焦按钮加 `▶` 标记或背景高亮
- Confirm / Cancel 是独立 `OnConfirm` 槽位(`FieldFocus == -1` + `OnConfirm` 双语义)

### 按钮组 — 目标规范(BR-043)

- Confirm / Cancel 集成到 fields 列表(共享焦点)
- 删除 `OnConfirm` 独立槽位
- 详见 [C01-form](./C01-form.md) 的焦点模型目标规范

### 禁止重叠

- 不允许 UI 内容与按钮、提示、Dialog 边框重叠
- 小窗口(40x16)下保留显示完整性的优先级

## 与需求文档的引用

- [../requirement/R03-form-action/R03-02-action-bar.md](../requirement/R03-form-action/R03-02-action-bar.md)
- [../requirement/R03-form-action/R03-03-action-dialog-layout.md](../requirement/R03-form-action/R03-03-action-dialog-layout.md)

## 待确认项

1. Action 展示框布局中"Confirm 与 options 共享焦点"的具体实现路径(用 FormFieldKind 新增 FormConfirm/FormCancel,还是在 FormSpec 加单独字段)。
2. 默认尺寸是"宽 3/4 + 等高"还是仅 Action Bar / Form Dialog 适用,Confirm / Toast 是否沿用旧百分比。
3. Overlay 颜色是否需要按 dialog 类型区分(Confirm vs Form vs Toast)。
4. 当前 `dialogConfig` 的百分比字段是保留作 fallback 还是彻底移除。