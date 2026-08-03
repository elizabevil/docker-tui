# constraint/ — 横切约束真理源

> 建立日期: 2026-08-02
> 状态: 形成中(按 [docs-architecture.md §8](../docs-architecture.md) 第一阶段)
> 范围: 仅整理,不修改代码

## 目录结构

```text
constraint/
├── README.md                       # 本文件
├── C01-form.md                     # FORM 字段、焦点、校验、条件显示
├── C02-dialog.md                   # Dialog 尺寸、居中、透明
├── C03-table.md                    # 表格列布局、选中态、统计位置
├── C04-keybinding.md               # 快捷键分层、冲突规则
├── C05-path.md                     # Local/Container path、绝对路径、补全
└── C06-i18n.md                     # i18n key 与终端宽度约束
```

## 角色

只写跨需求通用规则,单个功能的业务流程放在 [../requirement/](../requirement/)。
需求文档需要引用此处约束,而不复制通用段落。

## 现有约束与出处

每个约束文档必须列出:适用范围、规则、来源文档、待确认项。

| 约束 | 出处 |
|---|---|
| [C01-form.md](./C01-form.md) | task/br-041/042/043 |
| [C02-dialog.md](./C02-dialog.md) | task/br-040/043 |
| [C03-table.md](./C03-table.md) | 历史对话要求 + tables/*.jsonc |
| [C04-keybinding.md](./C04-keybinding.md) | task/br-039/043 |
| [C05-path.md](./C05-path.md) | task/br-041/043 |
| [C06-i18n.md](./C06-i18n.md) | docs/i18n.md + lang/*.jsonc |

## 关联

- 需求索引:[../requirement/](../requirement/)
- 文档架构:[../docs-architecture.md](../docs-architecture.md)

## 待确认

- C03-table 与 C01-form 都涉及"列宽/对齐",是否合并或保持独立边界。
- C04 是否纳入 Action Bar 与全局快捷键冲突的最终决议(目前待主模型决策)。