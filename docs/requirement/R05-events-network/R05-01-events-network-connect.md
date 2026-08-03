# R05-01 Events 独立面板 + Network Connect/Disconnect

## 元信息

- 状态: implementing
- 优先级: medium
- 来源: [../../task/br-035-events-panel-and-network-connect.md](../../task/br-035-events-panel-and-network-connect.md)
- 关联任务: 无
- 关联约束:
  - [../../constraint/C03-table.md](../../constraint/C03-table.md)
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)

## 目标

1. **Events 独立面板**:`F3` 全局进入 `ModeEvents`,浏览 runtime events 流。
2. **Network Connect / Disconnect**:网络页与容器详情页加入容器接入 / 脱离网络动作。

## 用户流程

### Events

任意 ModeNormal 页面 → `F3` → `ModeEvents` → 浏览事件列表(保留最近 1000) → Filter / Pause / Clear → `Esc` 返回原页面。

### Network Connect / Disconnect

网络页选中网络 → Action Bar "Connect Container" → 选容器 → Submit。或容器详情页 → Action Bar "Network" → 选网络 → Submit。Disconnect 反向。

## UI/UX

### Events

- 复用 [C03-table](../../constraint/C03-table.md) 表格布局
- 顶部状态:Filter / Paused 状态指示
- Filter 输入走 input mode(`/` 进入)
- 鼠标滚轮支持(可选)
- Clear 二次确认

### Network

- 接入 / 脱离走 Action Bar
- 容器列表灰显已连接的网络,网络列表灰显已接入的容器

## 功能规则

### Events

- 默认保留最近 1000 条,超限丢最早
- Pause 期间继续订阅到内存但停止渲染
- Filter 按 type / actor / action 匹配
- 切 runtime 时清空旧 events,新 events 用新 runtime 类型

### Network

- NetworkService.Connect(ctx, networkID, containerID, opts)
- NetworkService.Disconnect(ctx, networkID, containerID, force)
- 双 runtime 行为一致

## 实现设计

### Events

- 状态:`internal/tui/state/events.go` `EventsState`
- 订阅:`internal/tui/update/update_events.go:28` `subscribeEventsCmd`
- 键盘:`internal/tui/keyboard/events_keys.go`
- 视图:`internal/tui/ui/pages/events/view.go`
- i18n:`events.title` / `events.filter` / `events.paused` / `events.clear.confirm`

### Network

- 接口:`internal/data/runtime/network.go:NetworkService.Connect/Disconnect`
- 适配:docker / podman adapter 各自实现
- 键盘:`internal/tui/keyboard/network_action.go` 已有 `ActionNetworkRemove`,新增 Connect / Disconnect
- Action Bar 注册:网络页 + 容器详情页

## 验收标准

### Events

- `F3` 进入 / 退出正确
- Pause / Filter / Clear 行为正确
- 大事件量(>500/分钟)不卡顿
- 切 runtime 不串事件

### Network

- 接入 / 脱离双 runtime 测试通过
- Action Bar 灰显已连接 / 已接入项
- 失败 toast 区分 runtime

## 非目标

- 事件持久化
- 容器迁移时自动重连网络

## 迁移记录

- 旧文档:[task/br-035-events-panel-and-network-connect.md](../../task/br-035-events-panel-and-network-connect.md)
- 保留信息:Events 状态、订阅机制
- 待确认状态:`implementing`,Events 部分完成;Network Connect / Disconnect 待实施;主模型确认后调整。