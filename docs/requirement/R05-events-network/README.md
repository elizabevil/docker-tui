# R05 Events & Network(事件流与网络)

> 范围: Runtime events 独立面板、网络连接/脱离

## 子需求

| 编号 | 标题 | 状态 | 优先级 | 来源 |
|---|---|---|---|---|
| [R05-01](./R05-01-events-network-connect.md) | Events 独立面板 (F3) + Network Connect/Disconnect | implementing | medium | [task/br-035-events-panel-and-network-connect.md](../../task/br-035-events-panel-and-network-connect.md) |

## 关联约束

- [constraint/C03-table.md](../../constraint/C03-table.md) — Events 列表布局
- [constraint/C04-keybinding.md](../../constraint/C04-keybinding.md) — F3 / F2 分层

## 关联任务卡

- [task/br-035-events-panel-and-network-connect.md](../../task/br-035-events-panel-and-network-connect.md)

## 待确认

- R05-01 拆分:Events 面板已实现,Network Connect/Disconnect 未实现,是否需要拆为 R05-01 events 与 R05-02 network-connect。
- Events 面板是否还需要 Filter / Pause / Clear 等额外验收项。