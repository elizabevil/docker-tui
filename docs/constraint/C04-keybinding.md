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

Action Bar 是**操作的发现机制**(discoverability),不是唯一入口。两条
触发路径并存:

- **快捷键**:熟练用户的最快路径,在任何 Mode 下都生效(全局快捷键)
- **Action Bar**:探索场景,`;` 打开,看 description 后 Enter 执行

#### 该进 Action Bar 的(发现价值高)

- **无快捷键的操作**:Rename / Top / Port / Diff / Copy / Update / Export /
  Commit / Wait / imageHistory / volumeRemove / networkRemove /
  volumePrune / networkPrune / imagePrune — 这些不通过 Action Bar 就触发不了
- **破坏性操作即使有快捷键**:Prune / Remove — description 让用户在按下 Enter 前看清后果,
  KeyP / Ctrl+D 自己按的失误成本太高

#### 不该进 Action Bar 的(自解释 + 简单)

- **简单 + 已有快捷键 + 后果不严重**:Create — `C` 就够了,弹一个 input,
  没快捷键用户也能在 command palette 里输入 `:create`
- schema 通过 `OperationSpec.InActionBar: false` 显式排除。loader 默认
  `true`(行为不变);只在"想排除"的少数 op 上写 `false`

#### 不该进 Action Bar 的(全局命令)

- 全部全局命令走 `:` command palette 或直接快捷键
- Action Bar 不接管 Refresh / Switch Runtime / Show Help 等

#### 重复是允许的

- 快捷键和 Action Bar 入口**可以共存**(快捷键给熟练用户,Action Bar 给探索场景)
- 不再要求"两者只能选其一" — 这是 R06-09 修订,反映 schema 演进

#### 在哪修改

- `internal/data/config/defaults/operations/scopes/<scope>.jsonc` —
  加 `inActionBar: false` 字段排除
- 字段语义与默认值见 `internal/data/config/operations.go: OperationSpec.InActionBar`
- actionbar 渲染侧过滤:`internal/tui/actionbar/registry.go: buildItemsForScope`
  跳过 `!spec.InActionBar` 的项

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