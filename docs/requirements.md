# 需求规格

## 功能矩阵

```
资源层                 操作层                    体验层
─────────────────────────────────────────────────────────
容器列表               启停/重启/删除            键盘导航 (hjkl)
实时日志（streaming）  Exec/Attach Shell         搜索/过滤
实时 CPU/内存          Inspect/详情             面板布局切换
镜像列表               拉取/删除/Prune           主题/皮肤
Volume 列表            创建/删除/Inspect         鼠标支持
网络列表               创建/删除/Inspect         快捷键自定义
Compose 项目           启动/停止/聚合日志        配置持久化
Docker Events 事件流   Bulk 批量操作             状态栏/帮助
```

## 竞品参考

| 项目 | TUI | 优势 | 借鉴点 |
|---|---|---|---|
| **lazydocker** | gocui | 功能最全、51k⭐ | 功能清单、Compose 集成 |
| **Cruise** | Bubbletea | 架构现代 | 面板布局、stats 展示 |
| **EasyDocker** | Bubbletea | 小屏适配 | Metrics 面板 |
| **oxker** | Ratatui(Rust) | 性能优先 | Events 驱动刷新 |
| **docker-tui** | Bubbletea | MCP Server | 单 binary 分发 |

## 差异化机会

1. **多面板同时展示** — 类似 k9s 的 table + detail + logs 三栏视图
2. **内嵌 Docker Events** — 实时事件流面板（现有竞品薄弱）
3. **键盘体系** — 完整的 vim 风格 + 可自定义键位
4. **主题系统** — Lip Gloss 驱动的可换肤

## MVP 分期

### Phase 1 — 骨架（已完成）
- 项目脚手架 + Bubbletea 主循环
- Docker client 初始化 + 连接检测
- 容器列表（展示 + 启停删除）
- 键盘导航 + 面板切换
- 配置文件加载

### Phase 2 — 核心功能
- 实时日志流
- 实时 Stats (CPU/Mem)
- 镜像管理（拉取/删除/Prune）
- Exec/Attach Shell
- 搜索过滤

### Phase 3 — 全面覆盖
- Volume / Network 管理
- Compose 项目检测 + 管理
- Docker Events 实时流
- 主题系统 + 状态栏
- Bulk 批量操作

### Phase 4 — 打磨
- 鼠标支持
- 自定义快捷键
- 多主机切换（DOCKER_HOST）
- MCP Server
- goreleaser 发布管道

## 用户场景

### 场景 1：日常开发调试
开发者运行多个容器，需要快速查看日志、状态、资源占用。
→ 打开 docker-tui，j/k 切换容器，l 看日志，m 看 metrics

### 场景 2：生产环境排查
SRE SSH 到跳板机，没有 GUI 工具，需要排查容器问题。
→ docker-tui 在 tmux 中运行，全键盘操作，查看 Events + 日志 + Stats

### 场景 3：Compose 项目管理
多服务 Compose 项目，需要统一启停和查看聚合日志。
→ 进入项目目录运行 docker-tui，自动检测 compose.yaml，按 service 维度管理
