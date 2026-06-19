# 项目结构

```
docker-tui/
├── cmd/
│   └── docker-tui/
│       └── main.go                 # CLI 入口 + Cobra 命令 + Bubbletea 主循环
│
├── internal/
│   ├── config/
│   │   ├── config.go               # YAML 配置加载/保存
│   │   └── types.go                # Config 结构体定义 + 默认值
│   │
│   ├── docker/
│   │   └── client.go               # Docker SDK 封装（全部资源类型）
│   │
│   └── tui/
│       ├── model/
│       │   ├── app.go              # 主 Model + 面板/模式枚举
│       │   └── containers.go       # 各资源 Model + 异步消息类型
│       ├── view/
│       │   └── layout.go           # 所有渲染函数 + 布局管理
│       ├── update.go               # 消息处理路由 + 异步命令工厂
│       ├── styles.go               # Lip Gloss 主题定义
│       └── keys.go                 # 键盘绑定映射 + 帮助文本
│
├── docs/                           # 本文档目录
│   ├── README.md
│   ├── requirements.md
│   ├── architecture.md
│   ├── project-structure.md
│   └── ai-prompts.md
│
├── go.mod
└── go.sum
```

## 包职责

### `cmd/docker-tui/`

TUI 程序的入口点。职责：
1. Cobra CLI 参数解析（`--config`, `--host`, `--version`）
2. 配置文件加载
3. Docker client 初始化
4. Bubbletea Program 创建与启动
5. `mainModel` 结构体实现 `tea.Model` 接口：
   - `Init()` — 首次加载全部资源
   - `Update()` — 委托给 `tui.Update()`
   - `View()` — 委托给 `view.RenderApp()`

### `internal/config/`

配置管理。职责：
- 定义完整的配置结构体 `Config`，含默认值
- 从 YAML 文件加载配置，缺失字段用默认值填补
- 保存配置到文件
- `ConfigDir()` / `ConfigFile()` 计算 XDG 兼容路径 (`~/.config/docker-tui/`)

### `internal/docker/`

Docker 引擎交互层。职责：
- 封装 `github.com/docker/docker/client` SDK
- 为每种资源类型提供简化接口
- **容器**: List/Start/Stop/Restart/Kill/Remove/Logs/Stats
- **镜像**: List/Remove
- **Volume**: List/Remove
- **Network**: List/Remove
- **Events**: 流式事件通道

关键模式：所有流式 API（Logs/Stats/Events）返回 `io.ReadCloser` 或 `<-chan`，消费方在 goroutine 中处理。

### `internal/tui/model/`

Bubbletea Model 定义。职责：
- `AppModel` — 顶层状态，持有所有子 Model 和配置引用
- 各资源 `*ListModel` — 数据项 + 光标 + 加载 + 过滤状态
- 所有 `type XxxMsg struct` 消息类型定义

### `internal/tui/view/`

纯渲染层，无副作用。职责：
- `RenderApp()` — 顶层布局入口
- 各 `renderXxxList()` — 资源表格渲染
- `renderStatusBar()` — 底部状态栏
- `renderHelpView()` — 帮助面板
- `renderLogView()` — 日志预览

### `internal/tui/update.go`

消息处理路由。职责：
- 接收所有 `tea.Msg` 类型并路由
- `tea.WindowSizeMsg` → 更新宽/高
- `tea.KeyPressMsg` → 导航/操作/命令
- 各 `XxxLoaded` → 数据填充
- `FetchXxx()` — 工厂函数，返回 `tea.Cmd` 执行异步操作

### `internal/tui/styles.go`

视觉主题。职责：
- 颜色调色板定义
- 各 Style 变量声明
- 工具函数（如 `ContainerStateColor()`）

### `internal/tui/keys.go`

输入绑定。职责：
- `KeyAction` 枚举 + 默认 `KeyMapping`
- `HelpEntry` 列表 + `HelpEntries()` 给帮助面板
- `KeyFromMsg()` 工具函数

## 扩展指南

### 新增资源类型（如 Compose）

1. `internal/docker/client.go` → 添加 List/CRUD 方法
2. `internal/tui/model/containers.go` → 添加 `*ListModel` + 消息类型
3. `internal/tui/model/app.go` → 添加 `PanelType` 枚举值 + 子 Model 引用
4. `internal/tui/update.go` → 添加 `FetchXxx()` 命令 + 消息处理分支
5. `internal/tui/view/layout.go` → 添加 `renderXxxList()` 函数

### 新增操作（如 Image Prune）

1. `internal/docker/client.go` → 添加 SDK 方法
2. `internal/tui/model/containers.go` → 添加消息类型
3. `internal/tui/update.go` → 添加 `tea.Cmd` 工厂 + Update 分支
4. `internal/tui/keys.go` → 绑定快捷键
5. `internal/tui/view/layout.go` → 按需添加 UI 反馈
