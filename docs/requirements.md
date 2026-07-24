# 需求规格

本文是产品范围与路线图文档，不是当前实现清单。实现现状请看 [README.md](README.md) 和 [architecture.md](architecture.md)。

## 当前范围与状态

| 能力 | 状态 | 代码验证备注 |
|---|---|---|
| 容器列表、启停、重启、Kill、删除 | 已实现 | `internal/tui/keyboard/container_action.go` |
| 实时日志 | 已实现 | 当前为按需拉取/显示，不是启动即常驻事件流 |
| 容器 stats | 已实现 | 定时轮询当前容器列表，不是只轮询单个聚焦容器 |
| 镜像列表、Pull、Prune、Tag、Push、Save、Load、详情 | 已实现 | 镜像传输支持进度、取消和错误展示；Debug 保留命令预览；详情页按普通镜像 layer history 或 manifest 平台变体分区渲染；Docker 与 Podman 都提供结构化 `ImageDetail`（架构、OS、Driver、LayerCount、Runtime config、History） |
| Volume / Network 创建、列表、清理、删除、详情 | 已实现 | `c` 创建、`p` 清理；清理报告保留逐资源部分失败 |
| Compose 项目视图 | 已实现 | 基于 `com.docker.compose.*` labels 聚合，不解析 `compose.yaml` |
| 主题、i18n、帮助页、命令模式、过滤 | 已实现 | `zh` / `en` 已接入 |
| 运行时切换 | 已实现 | 连接池通过 `runtimeapi.BuildConnections` 组合本地 Docker / Podman 候选与配置连接；`F2` 打开 `runtime/runtimeSelector` 切换面板；按 `5s` 周期刷新延迟/状态；配置连接和远程 Docker / Podman 都经 TLS 配置校验后接入 |
| Docker / Podman Events 实时同步 | 已实现 | 绑定活动连接，支持退避重连、事件合并、局部刷新和轮询降级；独立事件面板仍属后续增强 |
| 自定义快捷键 | 已实现 | `keymap.*` 覆盖会编译为运行时绑定，Help / Footer 投影当前有效键位 |
| 用户操作审计 | 已实现 | 资源操作共享 trace，终态投影到通知与 Footer，并按日写入 JSONL |
| Bulk 批量操作 | 部分实现 | 已有 mark 模式，但覆盖范围仍有限 |
| 鼠标支持 | 部分实现 | 当前主要用于日志/详情滚轮滚动 |

## 当前已验证差距

这些点适合作为后续 bug / 需求跟踪入口，均已按代码核对：

- 镜像详情使用结构化数据链路；普通镜像 History 来自 runtime layer API，manifest list History 展示平台变体，并区分加载、空数据与获取失败。
- `config.Keymap` 已接入动作注册表和上下文解析器；当前开放动作级默认绑定覆盖，用户级上下文规则仍属于后续能力。
- 默认配置、运行时、Help 和 Footer 已统一为 `Ctrl+S`、`Ctrl+K`、`Ctrl+P` 等默认键位语义。
- Filter 与 Search 已拆分：资源列表输入即时过滤并使用双 `Esc` 清除退出，日志搜索按 Enter 应用且不改变原始数据集。
- 用户业务操作已接入统一审计模型；非审计 UI 提示不会写入审计文件。

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

### P1：强化资源联动

- 补齐 Compose / 镜像 / 容器联动反馈

### P2：完善操作面

- Volume / Network 多字段创建表单（driver、labels、options）
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
