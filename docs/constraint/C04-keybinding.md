# C04 Keybinding 横切约束

## 适用范围

全局快捷键、面板快捷键、对话框快捷键、弹层快捷键。

## 来源文档

- [../task/br-039-action-bar-replace-q.md](../task/br-039-action-bar-replace-q.md)
- [../task/br-039-a-implementation-design.md](../task/br-039-a-implementation-design.md)
- [../task/br-043-form-action-bar-redesign.md](../task/br-043-form-action-bar-redesign.md)

## 规则

### 分层 — 当前实现

- 全局快捷键:任意 Mode 都生效,只有 `Esc` 双重 press 等少数例外
- 面板快捷键:仅在 `ModeNormal` + 目标 panel 激活时生效
- 对话框快捷键:仅在对应 Mode(如 `ModeContainerForm`、`ModeConfirm`)激活时生效
- 弹层快捷键:在 Form Popup 打开期间覆盖对话框层

### 分层 — 目标规范

- 与当前实现一致
- Action Bar 单独成层(`ModeActionBar`)

### Action Bar 角色

- 只承载复杂操作(多个 action 选项)
- 不承载全局命令(全局命令走 `:` command palette 或直接快捷键)
- 直接快捷键与 Action Bar 项不重复

### 链接型快捷键(R / C) — 待主模型决定

- R(刷新)、C(连接信息)等同表格/详情页"链接型"快捷键,待主模型决议其归宿:
  - 保留为全局快捷键
  - 仅面板内生效
  - 完全移除(由 Action Bar 承担)
- 决议前保持现状不改动

### 配置与覆盖

- 用户可通过 `keymap.*` YAML 覆盖
- 解析后写入 `effectiveBindings`,运行期不重读
- 帮助页 / Footer 投影当前有效键位

## 与需求文档的引用

- [../requirement/R03-form-action/R03-02-action-bar.md](../requirement/R03-form-action/R03-02-action-bar.md)
- [../requirement/R05-events-network/R05-01-events-network-connect.md](../requirement/R05-events-network/R05-01-events-network-connect.md): F3 Events 入口

## 待确认项

1. R / C 等表格链接键的最终归宿。
2. Action Bar 内是否允许二级弹层(例:多步 wizard)。
3. 帮助页 Action Bar 章节是否分页(每页独立)还是统一一页。