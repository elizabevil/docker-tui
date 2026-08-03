# R03 Form & Action(表单与动作展示)

> 范围: 跨页面通用的 Form 模式、Action Bar、Action 展示框布局

## 子需求

| 编号 | 标题 | 状态 | 优先级 | 来源 |
|---|---|---|---|---|
| [R03-01](./R03-01-form-pattern.md) | 公共 Form 字段模式、状态、提交 | implementing | high | [task/br-041-unified-form-path-completion.md](../../task/br-041-unified-form-path-completion.md) + [task/br-041-form-navigation-followup-prompt.md](../../task/br-041-form-navigation-followup-prompt.md) |
| [R03-02](./R03-02-action-bar.md) | Action Bar 替换 Q,承载多动作入口 | implementing | high | [task/br-039-action-bar-replace-q.md](../../task/br-039-action-bar-replace-q.md) |
| [R03-03](./R03-03-action-dialog-layout.md) | Action 展示框尺寸 + flex 布局 + 共享焦点 | planned | high | 原 BR-040 已归档;[task/br-043-form-action-bar-redesign.md](../../task/br-043-form-action-bar-redesign.md) 保留 |
| [R03-04](./R03-04-path-popup.md) | FormPath 补全 popup:eza-l 风格 + BrowseMode + drill-down | implementing | high | [task/br-041-unified-form-path-completion.md](../../task/br-041-unified-form-path-completion.md) + BR-043 §3.1 |

## 关联约束

- [constraint/C01-form.md](../../constraint/C01-form.md) — Form 字段、焦点、校验、条件显示
- [constraint/C02-dialog.md](../../constraint/C02-dialog.md) — Dialog 尺寸、居中、透明
- [constraint/C03-table.md](../../constraint/C03-table.md) — Action Bar 表格列布局
- [constraint/C04-keybinding.md](../../constraint/C04-keybinding.md) — 快捷键分层
- [constraint/C05-path.md](../../constraint/C05-path.md) — Form Path 字段语义
- [constraint/C06-i18n.md](../../constraint/C06-i18n.md) — 文案

## 关联任务卡

- [task/br-041-unified-form-path-completion.md](../../task/br-041-unified-form-path-completion.md)
- [task/br-041-form-navigation-followup-prompt.md](../../task/br-041-form-navigation-followup-prompt.md)
- [task/br-039-main-model-feedback.md](../../task/br-039-main-model-feedback.md)
- [task/br-043-form-action-bar-redesign.md](../../task/br-043-form-action-bar-redesign.md) — 6 项未实施问题原始清单(R03-03 保留)
- BR-042 / BR-040 已合并归档:内容见 R03-01 / R03-03 与 [constraint/C01-form.md](../../constraint/C01-form.md) / [constraint/C02-dialog.md](../../constraint/C02-dialog.md)

## 待确认

- R03-01 的状态:`implementing` 还是 `implemented`?旧 br-041 task 文档说"第一阶段完成",但 [task/br-043-form-action-bar-redesign.md](../../task/br-043-form-action-bar-redesign.md) 列出 6 项未实施项,实际并未完整 done。
- R03-03 当前是设计草稿,主模型确认后才进入实施。
- 表单条件字段(DependsOn)是否纳入 R03-01 还是另立 R03-04。