# R08-10 实时事件流(events)

## 元信息

- 状态: implementing
- 优先级: medium
- 来源: docker compose CLI `compose events` + 用户评审结果
- 关联任务: [.omo/compose-todo.md 附录 A.6 `events` / §3.6 / 补 F.1](../../../.omo/compose-todo.md)
- 关联约束:
  - R05 事件流([../../requirement/R05-events-network/R05-01-events-network-connect.md](../../requirement/R05-events-network/R05-01-events-network-connect.md))
  - [../../constraint/C03-table.md](../../constraint/C03-table.md)

## ⚠ 重复功能说明

| 能力 | 已实现层 | 本需求增量 |
|---|---|---|
| `docker events` 容器级事件 | R05 事件流(Docker socket events) | **复用引擎层** |
| `docker events --filter label=com.docker.compose.project=<p>` | R05 事件流(支持 label filter) | R08-10 增量:补**项目级预置 filter**,UI 简单点击即可 |
| `docker compose events` | 无 project 维度过滤的 UI 入口 | R08-10 增量:在 Compose 面板的 Action Bar 提供 "Events" 入口 |

> R08-10 不新建独立事件面板(已有 R05 事件面板)。本需求做的是 **"从 Compose 面板便捷跳到事件面板并预填 filter"**。

## 目标

补 Compose-级 `events` 入口,与 R05 事件流协同:

1. Compose 面板 → 选中 project → Action Bar → Events
2. 跳转 R05 事件面板 + 预填 `label=com.docker.compose.project=<project>` filter
3. R05 现有"filter clear / follow"机制不变

## 用户流程

1. Compose 面板左栏选中 project
2. Action Bar (`;`) → Events(`Ctrl+F3` 直达,待 R08-12 决策键位)
3. 切换到 R05 事件面板 + filter 自动填入:
   - `label=com.docker.compose.project=<project>`
   - 可选:`type=container`(默认)或保留所有类型
4. 用户在事件面板可继续手动调整 filter / 取消 filter
5. Esc 返回 — 回到 compose 面板,**不保留**事件 filter(下次进入是重新预填)

## UI/UX

仅本需求特有:

- 跳转过渡:不弹 Confirm,直接切换 ActivePanel(ModeEvents)并设 filter 输入
- R05 面板 footer 显示"由 <project> 跳转"标识,**navigate back 时清除**

完整版式沿用 R05,无新设计。

## 功能规则

### F1 入口类型

```go
// ComposeService.Events:
// 真实事件流由 Engine 层 / R05 提供,ComposeService 不重发事件,只描述订阅维度
type ComposeEventFilter struct {
    Project       string
    Services      []string // service name list,空 = 全部
    Type          []string // container / network / volume / image 等
    Since         time.Time
    Until         time.Time
}
```

### F2 跳转 + filter 写入

实现:nav.mode = ModeEvents,同时调用 `m.Events.Filter.Set(filter)`,让 R05 面板立即按 filter 显示。

### F3 filter 文案

`label=com.docker.compose.project=<project>` — 这是 daemon filter 实际需要的语法。

> ⚠ Podman events 的 filter 语法不同(`--filter label=…` 类似,但默认时间格式 yaml vs RFC3339)。本需求采纳 docker 标准,在 Podman 侧复用事件流显示(R05 已经做了 docker filter 兼容 → podman auto-adapter,本需求直接复用即可)。

### F4 与 R05 events-network 的关系

| 维度 | R05-01 | R08-10 |
|---|---|---|
| 真实事件流实现 | ✓(docker socket events) | 复用 R05 |
| 事件面板 UI | ✓(独立面板,F3) | 复用 R05 |
| filter 文案 / 解析 | ✓ | 复用 R05 |
| **从 Compose 面板跳入 + 预填** | × | ✓(R08-10 增量) |

## 实现设计

涉及模块:

- 修改:`internal/tui/keyboard/compose_action.go` 加 `doComposeEvents`
- 修改:`internal/tui/state/events.go`(`m.Events.Filter.SetFromCompose`)
- 修改:`internal/tui/keyboard/compose_nav.go` 注册 `Ctrl+F3`(R08-12 同步)
- 复用:`internal/tui/ui/pages/events/` 全部

调用链:

```text
doComposeEvents(project)
  -> m.Navigation.Mode = ModeEvents
  -> m.Events.SetPreFilter({label=com.docker.compose.project=<project>})
  -> Render ModeEvents (R05 现有)
```

## 验收标准

1. Compose 面板选中项目 → Events 入口 → 跳到事件面板 + filter 已显示
2. 事件面板接受 manual filter 改动
3. Esc 返回 Compose 面板,filter 状态归零(下次进入预填新值)
4. audit 不触发(只读事件)

## 非目标

- 不实现"compose 专属独立事件面板"(沿用 R05)
- 不解析 `compose.yaml` 的事件分类(无 yaml)

## 迁移记录

- 旧代码:无
- 复用:R05 事件面板全部
- 保留信息:R05 事件流的 docker / podman 适配
- 废弃信息:无
- 待确认:filter 用 R05 当前 filter input 文案(`label=…`)还是新 Compose 文案 — 倾向**沿用 R05**
