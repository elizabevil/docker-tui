# Events 独立面板 (F3) + Network Connect/Disconnect

> 任务卡 / BR-035

## 元信息

- **关联编号**:BR-035
- **优先级**:medium
- **状态**:`partial` - Events 面板已完成；Network Connect / Disconnect 待实现
- **依赖**:TASK-008(Docker / Podman Events 已接入主循环)
- **关联设计**:[docs/feature-design.md §5.7 / §5.3](../feature-design.md)

## 目标

1. **Events 独立面板**:`F3` 全局进入 `ModeEvents`,浏览 runtime events 流,默认保留最近 1000 条,支持 Filter / Pause / Clear。
2. **Network Connect / Disconnect**:网络页与容器详情页加入容器接入 / 脱离网络动作,补齐 `NetworkService` 接口。

## 当前实现进度（2026-08-02）

### Events

- [x] F3 全局入口、独立 `ModeEvents`、面包屑与返回原页面。
- [x] 复用事件订阅，保留最近 1000 条，Pause / Resume 使用合并容量受限的缓冲区。
- [x] Filter、滚动、鼠标滚轮、Clear 二次确认。
- [x] 公共 Flex Table、i18n、底部状态栏快捷键提示。
- [x] Events 页面内部只显示状态与过滤条件，不重复显示快捷键。

### Network Connect / Disconnect

- [ ] 扩展 `NetworkService` 与 Docker / Podman adapter。
- [ ] 增加网络连接参数表单及 Action Bar 动作。
- [ ] 增加结果反馈、audit、刷新与双 runtime 测试。

## 代码结构索引

### Events 面板

| 文件 | 作用 |
|---|---|
| `internal/data/runtime/streaming.go:64` | `EventService.Subscribe` 接口(已实现) |
| `internal/data/runtime/docker/service_event.go` | Docker events 适配器(已实现) |
| `internal/data/runtime/podman/service_event.go` | Podman events 适配器(已实现) |
| `internal/tui/update/update_events.go:28` | `subscribeEventsCmd`(已实现,后台订阅) |
| `internal/tui/state/events.go` | 需新增:`EventsState` (Events []EventItem / Cursor / ViewOffset / Paused / Filter) |
| `internal/tui/keyboard/event_keys.go` | 需新增:`handleEventKeys` (Filter/Pause/Clear/Esc) |
| `internal/tui/ui/pages/events/view.go` | 需新增:Events 列表渲染(类似 logs 页) |
| `internal/tui/ui/app/layout.go` | 注册 `ModeEvents` → 渲染 events 页面 |
| `internal/tui/keys/registry.go` | 加 `{ActionEvents, []string{KeyF3}, app}` |
| `internal/tui/keys/action.go` | 加 `ActionEvents KeyAction = "events"` |
| `internal/data/i18n/lang/*.jsonc` | 加 `events.title` / `events.filter` / `events.paused` 等 |

### Network Connect / Disconnect

| 文件 | 作用 |
|---|---|
| `internal/data/runtime/networks.go` | 需新增 `NetworkService.Connect(ctx, networkID, containerID, opts)` / `Disconnect(ctx, networkID, containerID, force)` 接口 |
| `internal/data/runtime/docker/service_network.go` | Docker `NetworkConnect` / `NetworkDisconnect` 实现(参考 `cli.NetworkConnect`) |
| `internal/data/runtime/podman/service_network.go` | Podman REST `NetworkConnect` / `NetworkDisconnect` 实现 |
| `internal/tui/keys/action.go` | 加 `ActionNetworkConnect` / `ActionNetworkDisconnect` |
| `internal/tui/keys/registry.go` | 注册键位(建议 `+` / `-`,或走 Action Bar) |
| `internal/tui/keyboard/network_action.go` | 加 `doNetworkConnect` / `doNetworkDisconnect` + 表单弹层 |
| `internal/tui/ui/widget/dialog/network_connect.go` | 新增:容器 / IP / Alias 表单 |

### 必须新增文件

- `internal/tui/state/events.go`
- `internal/tui/keyboard/event_keys.go`
- `internal/tui/ui/pages/events/view.go`
- `internal/tui/ui/pages/events/view_test.go`
- `internal/tui/ui/widget/dialog/network_connect.go`

## 操作链路

### Events 面板

```text
User 按 F3 (任意页)
  → keys/registry.go:ActionEvents 命中
  → keyboard/actions.go:handleAction case ActionEvents
  → doEventsOpen(m): m.Navigation.Mode = state.ModeEvents
      m.Events.Open()  // 重置 cursor / paused
  → ui/pages/events/view.go:RenderView(m, h, w)
      读 m.Events.Events / Cursor / ViewOffset
      渲染 Time / Type / Action / Resource 四列(类似 logs 页)
  → keyboard/event_keys.go:handleEventKeys
      / → 触发 Filter 弹层
      Space → Paused toggle(暂停累积,新事件进缓冲)
      Ctrl+D → Clear buffer(高危,需 confirm)
      Esc → BackFromEvents (ModeEvents → ModeNormal)
      j/k → Cursor ±1
```

### Network Connect

```text
User 在网络页选中网络 → 选动作 (Action Bar / 命令)
  → 弹表单:选容器 / IP / Alias
  → submit → doNetworkConnect(m, networkID, containerID, opts)
      cmd := Engine.Networks().Connect(ctx, networkID, containerID, opts)
      Docker: cli.NetworkConnect
      Podman: REST /libpod/.../connect
  → toast + audit
  → Refresh 资源列表(网络页 + 容器页)
```

### Network Disconnect

```text
User 在网络详情 / 容器详情页选 Disconnect
  → 选目标网络(若是"网络页")或选 action(若是"容器详情")
  → doNetworkDisconnect(m, networkID, containerID)
      cmd := Engine.Networks().Disconnect(ctx, networkID, containerID, force)
  → toast + audit
  → Refresh
```

## 验收标准

### Events

- [ ] 全局按 F3 进入 Events 面板,标题 `Events` 或 `Events (paused/resumed)`。
- [ ] Docker / Podman events 流持续滚动;每条含 Time / Type / Action / Resource 四列。
- [ ] `Space` 暂停 / 恢复;暂停时新事件进入缓冲队列,Resume 时冲入。
- [ ] `/` 触发 Filter,按 type (container/image/network/volume) / action 过滤。
- [ ] `Ctrl+D` 清空 buffer(需 confirm dialog)。
- [ ] `Esc` 返回原 page,ModeEvents → ModeNormal。
- [ ] 默认保留最近 1000 条;超出滚动掉头。
- [ ] Help 页 / Footer 显示 `F3 Events` 提示。

### Network Connect / Disconnect

- [ ] 网络页 / 容器详情页可触发 Connect / Disconnect。
- [ ] Docker 与 Podman 两端 runtime 都接好。
- [ ] 接入 / 脱离后,容器在该网络接口立即可见 / 不可见(用 `docker exec ctr ping ...` 验证)。
- [ ] 错误时显示明确 toast(`network not found` / `already connected` 等)。
- [ ] Audit 写入 `resource.network.connect/disconnect`。
- [ ] Help / Footer 显示当前可用动作。

## 风险

| 风险 | 说明 |
|---|---|
| **Events 流中断** | 后台订阅的 `<-chan` 关闭时必须清理并可重订阅;见 `subscribeEventsCmd` 的 generation 处理 |
| **网络 API 兼容性** | Docker NetworkConnect 需要 endpoint 配置;Podman REST 因版本差异字段不同(尤其 IPv4/IPv6) |
| **状态广播** | Connect 后容器是否自动出现在新网络的 bridge 中,需要 Engine 端独立刷新容器列表 |
| **Help/Footer 漂移** | F3 是新全局快捷键,必须加 Help 提示 |
| **批量模式** | Connect/Disconnect 是否支持批量?单容器设计更简单 |

## 建议任务分解

1. **TASK-BR035-A:Events 面板 UI + 滚动**
   - `state/events.go` + `pages/events/view.go` + `view_test.go`
   - **预估**:小模型独立
2. **TASK-BR035-B:Events 订阅 + Pause/Clear/Filter**
   - 状态 + 键盘 + 集成现有后台订阅
   - **预估**:小模型辅助,主模型确认后台订阅架构
3. **TASK-BR035-C:Network Connect / Disconnect runtime**
   - `NetworkService` 接口 + Docker + Podman 实现
   - **预估**:主模型实现
4. **TASK-BR035-D:Network Connect / Disconnect TUI**
   - TUI handler + 表单 + audit
   - **预估**:小模型独立
5. **TASK-BR035-E:F3 键路由 + i18n + docs**
   - `ActionEvents` + 注册 + i18n + navigation.md 同步
   - **预估**:小模型机械同步

## 待确认项

- [ ] Events 缓冲 1000 条上限是否合理(高流量场景可能丢早期事件)
- [ ] 是否需要按 runtime 过滤(Docker vs Podman 独立显示)
- [ ] Network Connect 表单是否需要 IPv4 / IPv6 字段(Podman 支持,Docker 默认 v4)
- [ ] 批量 Connect 多个网络到同一容器是否支持(API 支持但 UI 复杂度高)
- [ ] Disconnect 是否需要 force 选项(强制从用户定义网络断开)
