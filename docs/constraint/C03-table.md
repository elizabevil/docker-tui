# C03 Table 横切约束

## 适用范围

所有数据表格页(Containers / Images / Volumes / Networks / Compose / Events / History)。

## 来源文档

- 历史对话要求
- `internal/tui/tables/*.jsonc` 当前配置
- `internal/tui/ui/component/table*.go` 当前实现
- `docs/architecture.md` 表格相关说明

## 规则

### 当前实现

- 列布局:必须占满可用宽度(不留大段空白)
- 列宽从外到内计算:先算可用宽,再按列内容可见宽度 + 比例上限分配
- 文本列默认左对齐,数字列右对齐
- 中日韩等宽字符按 `utils.DisplayWidth` 计可见宽度
- 选中行整行高亮(背景色)
- 滚动:`EnsureVisible(visibleRows)` 保证 cursor 始终可见
- 表格 profile 在 `internal/tui/tables/{page}.jsonc`
- 配色由 `style.Colors` 统一
- 列定义需 i18n(`key` 而非裸字符串)
- 斑马线 / 列样式由 jsonc 配置

### 目标规范

- 同上,无重大重构
- Action Bar / Form Dialog 表格列布局走 `dialog.jsonc` 而非 `tables/`,两块配置源待统一

## 与需求文档的引用

- [../requirement/R02-image/R02-01-history.md](../requirement/R02-image/R02-01-history.md): History 表格
- [../requirement/R05-events-network/R05-01-events-network-connect.md](../requirement/R05-events-network/R05-01-events-network-connect.md): Events 表格

## 待确认项

1. Action Bar / FormDialog 是否也走 `internal/tui/tables/` 配置(目前是 `dialog.jsonc`)。
2. 多列表头(分组列)是否需要支持。
3. 列排序是否在所有页统一支持,还是仅主页支持。