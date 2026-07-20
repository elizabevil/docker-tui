# 需求规格

本文是产品范围与路线图文档，不是当前实现清单。实现现状请看 [README.md](README.md) 和 [architecture.md](architecture.md)。

## 当前范围与状态

| 能力 | 状态 | 代码验证备注 |
|---|---|---|
| 容器列表、启停、重启、Kill、删除 | 已实现 | `internal/tui/keyboard/container_action.go` |
| 实时日志 | 已实现 | 当前为按需拉取/显示，不是启动即常驻事件流 |
| 容器 stats | 已实现 | 定时轮询当前容器列表，不是只轮询单个聚焦容器 |
| 镜像列表、Pull、Prune、详情 | 已实现 | Export / Debug 目前提供命令预览对话框；详情页已分区渲染，但 History 区块仍是占位提示 |
| Volume / Network 列表、删除、详情 | 已实现 | 当前没有“创建”入口 |
| Compose 项目视图 | 已实现 | 基于 `com.docker.compose.*` labels 聚合，不解析 `compose.yaml` |
| 主题、i18n、帮助页、命令模式、过滤 | 已实现 | `zh` / `en` 已接入 |
| 运行时切换 | 部分实现 | 连接池和 `F2` 切换已接入，默认工作流偏本地 Docker / Podman |
| Docker Events 实时订阅 UI | 规划中 | `Client.Events()` 已封装，但主循环尚未接线 |
| 自定义快捷键 | 已实现 | `keymap.*` 覆盖会编译为运行时绑定，Help / Footer 投影当前有效键位 |
| Bulk 批量操作 | 部分实现 | 已有 mark 模式，但覆盖范围仍有限 |
| 鼠标支持 | 部分实现 | 当前主要用于日志/详情滚轮滚动 |

## 当前已验证差距

这些点适合作为后续 bug / 需求跟踪入口，均已按代码核对：

- 镜像详情页的 `History` 区块已经预留，但 `internal/data/docker/InspectImage()` 还没有接入镜像 layer history 数据，当前只显示“暂不可用”占位文案。
- `config.Keymap` 已接入动作注册表和上下文解析器；当前开放动作级默认绑定覆盖，用户级上下文规则仍属于后续能力。
- 默认配置、运行时、Help 和 Footer 已统一为 `Ctrl+S`、`Ctrl+K`、`Ctrl+P` 等默认键位语义。
- 过滤交互当前是“资源列表输入即过滤、日志搜索按 Enter 应用”，不再是旧文档描述的“1 秒防抖后自动退出”。

## 竞品参考

| 项目 | TUI | 优势 | 借鉴点 |
|---|---|---|---|
| **lazydocker** | gocui | 功能最全、51k⭐ | 功能清单、Compose 集成 |
| **Cruise** | Bubbletea | 架构现代 | 面板布局、stats 展示 |
| **EasyDocker** | Bubbletea | 小屏适配 | Metrics 面板 |
| **oxker** | Ratatui(Rust) | 性能优先 | Events 驱动刷新 |
| **dtui** | Bubbletea | Docker / Podman 双运行时、终端体验完整 | 单 binary 分发、布局与模式分层 |

## 差异化机会

1. **多面板同时展示** — 类似 k9s 的 table + detail + logs 三栏视图
2. **内嵌 Docker Events** — 实时事件流面板（现有竞品薄弱）
3. **键盘体系** — 完整的 vim 风格 + 可自定义键位
4. **主题系统** — Lip Gloss 驱动的可换肤

## 后续路线

### P0：补齐实现与文档漂移

- 把帮助页、README、构建脚本中的命名和路径说明统一
- 继续扩展可配置动作范围，并评估用户级上下文覆盖与冲突诊断
- 明确 `dtui` / `docker-tui` 的对外命名策略

### P1：强化实时性与资源联动

- 把 `internal/data/docker/events.go` 接入主循环
- 用事件流替代部分全量刷新
- 补齐 Compose / 镜像 / 容器联动反馈

### P2：完善操作面

- Volume / Network 创建入口
- 更完整的批量操作
- 更清晰的多连接 / 多主机工作流

### P3：体验打磨

- 更完整的鼠标支持
- 发布流程和分发整理
- 预览系统与实际 TUI 的文档同步约束

## 用户场景

### 场景 1：日常开发调试
开发者运行多个容器，需要快速查看日志、状态、资源占用。
→ 打开 `dtui`（或仓库构建出的 `docker-tui` 可执行文件），`j/k` 切换容器，`l` 看日志，`m` 开关 stats

### 场景 2：生产环境排查
SRE SSH 到跳板机，没有 GUI 工具，需要排查容器问题。
→ 在 `tmux` 中运行，使用日志、详情、stats 和资源切换做排查；事件流面板仍属于后续增强项

### 场景 3：Compose 项目管理
多服务 Compose 项目，需要统一启停和查看聚合日志。
→ 进入 Compose 环境对应主机后，通过容器 labels 聚合出的 Compose 视图按 project / service 维度管理
