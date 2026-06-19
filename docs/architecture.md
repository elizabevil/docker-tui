# 架构设计

## 技术选型

| 维度 | 选择 | 理由 |
|---|---|---|
| 语言 | Go | Docker SDK 同语言、单 binary 分发、跨平台 |
| TUI 框架 | **Bubbletea v2** | Elm MVU 模式，天然适合状态复杂的 TUI |
| 样式 | Lip Gloss v2 | 声明式样式、自动亮/暗色适配 |
| Docker | `github.com/docker/docker/client` | 官方 SDK，流式 API 原生支持 |
| CLI | Cobra | 业界标准，k9s 也用 |
| 配置 | YAML | 同 k9s 风格，用户无学习成本 |

### 为什么选 Bubbletea 不是 tview

| 维度 | Bubbletea v2 | tview/tcell | gocui |
|---|---|---|---|
| 架构 | Elm MVU | Retained widget | Callback |
| 异步 | 原生 Cmd/Msg | 自行管理 goroutine | 自行管理 |
| 组件 | Bubbles 官方组件库 | 内置丰富 | 几乎无 |
| 样式 | Lip Gloss 声明式 | 属性式 API | 无 |
| 社区 | 42.8k⭐，Charm 商业化 | 13.9k⭐，稳定 | 停滞 |
| Docker TUI 参考 | Cruise, EasyDocker, docker-tui | k9s | lazydocker |

## 分层架构

```
┌─────────────────────────────────────────────────────┐
│                  CLI (cobra)                         │
│  docker-tui [--config] [--host] [--compose-file]     │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│              Program (Bubbletea)                     │
│  tea.NewProgram(model, ...)                          │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│  internal/                                           │
│                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │   tui/       │  │   docker/    │  │   config/  │ │
│  │              │  │              │  │            │ │
│  │  model/      │  │  client.go   │  │  config.go │ │
│  │  view/       │  │  containers  │  │  types.go  │ │
│  │  update.go   │  │  images      │  │            │ │
│  │  styles.go   │  │  volumes     │  │            │ │
│  │  keys.go     │  │  networks    │  │            │ │
│  │              │  │  events      │  │            │ │
│  └──────────────┘  └──────────────┘  └────────────┘ │
└──────────────────────────────────────────────────────┘
```

## 数据流（核心模式）

```
Docker Events API                    Keyboard Input
      │                                    │
      ▼                                    ▼
  docker/events.go                   tea.KeyPressMsg
      │                                    │
      │  EventsMsg                         │
      ▼                                    ▼
  ┌──────────────────────────────────────────┐
  │          Update() 消息路由                │
  │                                          │
  │  switch msg := msg.(type) {              │
  │  case EventsMsg:     → 增量更新          │
  │  case KeyPressMsg:   → 导航/操作         │
  │  case ContainersLoaded: → 列表填充        │
  │  }                                       │
  └────────────────┬─────────────────────────┘
                   │ 返回更新后的 Model
                   ▼
  ┌──────────────────────────────────────────┐
  │          View() 渲染                      │
  │  RenderApp(m) → lipgloss 布局 → tea.View│
  └────────────────┬─────────────────────────┘
                   │
                   ▼
              终端输出
```

## 三种数据获取策略

| 策略 | 适用场景 | 实现 |
|---|---|---|
| **事件驱动** | 容器状态变更、镜像拉取完成 | Docker Events API → EventsMsg → 增量更新 Model |
| **定时轮询** | CPU/内存 stats | 聚焦容器 1s 轮询 stats API（同 k9s） |
| **按需加载** | inspect 详情、镜像层信息 | 用户选中时一次性请求 → 展示在详情面板 |

## 与 k9s 架构对照

```
k9s                           docker-tui
─────                         ──────────
tview + tcell                 Bubbletea v2 + Lip Gloss
internal/model/               internal/tui/model/（Bubbletea Models）
internal/ui/Table             lipgloss 表格渲染 + 自定义组件
k8s Watch API → 事件驱动      Docker Events API → tea.Msg
Flex/Grid 布局                lipgloss.JoinVertical/Horizontal
皮肤系统 (skins/)              Lip Gloss 主题定义
```

## 关键设计决策

### ADR-001: 使用 Docker Events API 而非轮询
- 避免定时 `docker ps` 带来的开销
- Events API 提供增量更新，仅在有变化时刷新
- 实现: `docker/events.go` → goroutine 消费事件流 → channel → EventsMsg

### ADR-002: 离线优先启动
- Docker daemon 不可达时仍启动 TUI，显示"Disconnected"状态
- 允许用户配置后再重试连接
- 实现: `NewClient()` 失败返回 nil，`mainModel` 检查 `m.model.Docker == nil`

### ADR-003: Stats 只在聚焦容器时轮询
- 参考 `docker stats` 的 1s 间隔
- 只对当前选中的 container 拉取，避免 N 个容器全量轮训
- 非聚焦时停止 goroutine
