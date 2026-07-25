# 项目结构

本文按当前目录结构整理，旧版 `internal/{config,docker}`、`internal/tui/{model,view}` 说法已失效。

## 顶层目录

```text
docker-tui/
├── cmd/docker-tui/            # CLI 入口
├── internal/data/             # 配置、引擎接入、i18n
├── internal/tui/              # 状态、交互、渲染
├── docs/                      # 当前维护中的说明文档
├── design/                    # 设计文档，current-design.md 为主
├── scripts/                   # 辅助脚本
└── test/                      # 集成/性能/实验性测试资源
```

## `cmd/docker-tui`

- `main.go`
  - Cobra CLI 入口
  - 加载配置、主题、语言
  - 初始化连接池
  - 创建 `state.AppModel`
  - 启动 Bubble Tea 程序

## `internal/data`

### `internal/data/config`

- `config.go`：配置加载/保存、默认配置路径
- `types.go`：完整配置结构
- `default.jsonc`：内嵌默认配置
- `theme.go`：主题加载与主题列表
- `themes/*.jsonc`：内置主题

### `internal/data/docker`

- `client.go`：Docker / Podman 客户端创建与自动探测
- `pool.go`：多连接池与运行时切换
- `containers.go` / `images.go` / `volumes.go` / `networks.go`：资源操作
- `inspect.go`：详情文本构建
- `stats.go`：容器 stats 解析
- `host.go`：宿主机 CPU / 内存 / 磁盘读取
- `events.go`：事件流封装

### `internal/data/i18n`

- `lang.go`：语言初始化与 `T()` 查词
- `en.jsonc` / `zh.jsonc`：文案

## `internal/tui`

### 状态与消息

- `state/app.go`：顶层 `AppModel`、面板与模式枚举
- `state/containers.go`、`images.go`、`volumes.go`、`networks.go`
  - 各资源列表状态
  - `Loaded` / `Actioned` / `Tick` 等消息

### 输入与业务动作

- `keyboard/keyboard.go`：统一键盘分发入口
- `keyboard/actions.go`：默认动作路由
- `keyboard/*_action.go`：各资源操作
- `keyboard/*_nav.go`：子视图导航
- `keyboard/fetch.go`：异步拉取命令工厂
- `keyboard/command.go`：命令模式
- `keyboard/exec_dialog.go`：容器 exec 对话框

### 键位描述

- `keys/`：键名常量、默认映射、帮助文案、展示标签

### 更新循环

`internal/tui/update/` 子包集中了 tea.Msg 的路由与处理：

- `update.go`：`Update` 总路由入口
- `update_resources.go`：列表装载结果处理
- `update_actions.go`：操作结果处理（含 `BatchActioned` 聚合）
- `update_tick.go`：stats、toast、host 监控等定时消息
- `update_log.go`：日志流处理
- `update_events.go`：runtime 事件流与联动刷新
- `update_image_transfer.go`：镜像传输进度与终态
- `handler_container.go`：容器命令与动作的胶水函数

### UI 渲染

- `ui/app/`：顶层布局与窗口配置
- `ui/pages/containers` / `images` / `volumes` / `networks` / `compose`
  - 各资源页渲染
- `ui/pages/logs` / `detail` / `help`
  - 覆盖层页面
- `ui/component/`
  - 表格、过滤框、toast、breadcrumb、样式加载、视口等复用组件
- `ui/widget/`
  - header / footer / panel / dialog 等组合部件
- `ui/style/`
  - 当前主题调色板的共享访问点

### 其他支撑

- `tables/*.jsonc`：表格列定义、Compose 布局参数
- `term/`：exec 透传终端缓冲
- `utils/`：ANSI 与格式化工具
- `composable/filter.go`：资源过滤逻辑

## `docs/ui-preview`

这是独立的浏览器预览工程，不参与主程序编译：

- `src/`：Vue 页面和组件
- `public/data/`：静态模拟数据
- `package.json`：Vite 命令

## 修改入口建议

### 加资源操作

1. 先改 `internal/data/docker/*`
2. 再补 `internal/tui/state/*` 消息和状态
3. 在 `internal/tui/keyboard/*` 接入动作
4. 在 `internal/tui/update/*` 处理结果
5. 最后在 `internal/tui/ui/pages/*` 渲染反馈

### 改布局或视觉

1. `design/current-design.md`
2. `internal/tui/ui/app/*`
3. `internal/tui/ui/component/*.jsonc`
4. `internal/tui/tables/*.jsonc`
5. 必要时同步 `docs/ui-preview`
