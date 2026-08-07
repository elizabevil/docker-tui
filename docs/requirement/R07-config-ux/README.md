# R07 配置系统便利化（config-ux）

## 范围

BR-043 强类型配置落地后引入大量配置项（8 个 defaults 分片、主题 6 个、operations scopes 4 组），用户缺少便捷的配置入口。本大需求通过 CLI 子命令与首次运行提示，降低配置的发现与使用成本，不改变现有配置加载语义。

## 子需求

| 编号 | 标题 | 状态 | 优先级 | 来源 |
|---|---|---|---|---|
| R07-01 | `dtui info` — 配置路径与主题列表 | planned | high | BR-043 follow-up |
| R07-02 | `dtui config init` — 生成配置文件模板 | planned | high | BR-043 follow-up |
| R07-03 | `dtui config validate` — 校验配置文件 | planned | high | BR-043 follow-up |
| R07-04 | 首次运行提示（toast） | planned | medium | BR-043 follow-up |

## 关联约束

- [../../constraint/C06-i18n.md](../../constraint/C06-i18n.md)（toast 文案三语）

## 关联任务卡

- [../../handoff/proposals/config-split-review.md](../../handoff/proposals/config-split-review.md)（BR-043，§19 配置热加载为明确非目标）
- 实施计划：[../../.omo/plans/config-ux.md](../../.omo/plans/config-ux.md)（T1-T11）

## 待确认

- shell 补全（`dtui config` Tab 补全）列为未来需求，不在本大需求范围。
