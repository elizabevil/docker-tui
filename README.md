# dtui

`dtui` 是一个使用 Go、Bubble Tea 和 Lip Gloss 构建的 Docker / Podman 终端管理工具。它将容器、镜像、卷、网络和 Compose 项目集中在同一个键盘驱动界面中，并提供资源详情、日志搜索、容器 Exec、可配置快捷键和用户操作审计。

> CLI 名称是 `dtui`，仓库默认构建产物目前为 `dist/docker-tui`。

## 功能

- 管理容器：列表、启动、停止、重启、Pause/Unpause、Rename、Top、Port、Kill、删除、日志、Stats、Inspect 和 Exec
- 管理镜像：列表、Pull、Prune、删除和结构化详情
- 创建、查看、清理和删除 Volume / Network，并查看关联资源
- 订阅运行时 Events，并按资源类型合并和局部刷新；断线时退避重连并降级轮询
- 基于容器 labels 聚合 Compose 项目与服务，支持 start、stop、down 和日志
- 在 Docker / Podman 本地连接之间切换
- 即时过滤资源列表，在日志页搜索并跳转匹配项
- 通过 YAML 覆盖动作快捷键，Help 和 Footer 显示当前有效绑定
- 为用户资源操作生成 trace，并将结果投影到 Toast、Footer 和按日 JSONL 审计日志

当前尚未完整实现的能力包括完整多主机工作流、镜像传输、更完整的批量操作和鼠标交互。详细状态见 [需求规格](docs/requirements.md)。

## 环境要求

- Go：使用与 [go.mod](go.mod) 一致的版本
- Docker 或 Podman daemon，并允许当前用户访问对应 socket
- 可选：[`just`](https://github.com/casey/just)，用于统一执行构建和测试命令

Linux 下常见 socket：

```text
Docker: unix:///var/run/docker.sock
Podman: unix:///run/user/1000/podman/podman.sock
```

## 构建与运行

使用 `just`：

```bash
just build
just run
```

直接使用 Go：

```bash
CGO_ENABLED=0 go build -o ./dist/docker-tui ./cmd/docker-tui
./dist/docker-tui
```

直接连接 Podman：

```bash
just run-podman
```

常用 CLI 参数：

```text
-f, --config PATH    指定配置文件
-t, --theme NAME     指定主题
-L, --lang zh|en     指定界面语言
-p, --podman         使用 Podman socket
    --list-themes    列出内置主题
-v, --version        显示版本
```

程序会始终注册本地 Docker 与 Podman 候选，默认优先连接 `local-docker`；Docker 不可用且本地 Podman 可用时自动连接 Podman，并显示切换提示。可通过 `runtime.default` 选择配置连接，或使用 `--podman` 明确选择本地 Podman。`runtime.connections` 支持独立 TLS 配置，远程连接需提供证书与校验参数。

## 基本使用

### 全局按键

| 按键 | 操作 |
|---|---|
| `Tab` / `Shift+Tab` | 切换主面板 |
| `j` / `k` / `↑` / `↓` | 移动光标 |
| `Enter` | 打开当前项目或执行主动作 |
| `Esc` | 返回；主界面双击退出 |
| `/` | 过滤资源；日志页中进入搜索 |
| `:` | 打开命令模式 |
| `?` / `F1` | 打开帮助 |
| `r` | 刷新全部资源 |
| `F2` | 切换运行时连接 |
| `Space` | 进入或操作标记模式 |
| `Ctrl+D` | 删除当前资源 |

### 资源操作

| 面板 | 常用按键 |
|---|---|
| Containers | `s` 启动、`Ctrl+S` 停止、`Ctrl+R` 重启、`p` Pause/Unpause、`Ctrl+K` Kill、`l` 日志、`e` Exec、`d` 详情、`m` Stats |
| Images | `Ctrl+P` Pull、`p` Prune、`d` 详情、`Ctrl+D` 删除 |
| Volumes | `Enter` 展开、`d` 详情、`Ctrl+D` 删除 |
| Networks | `d` 详情、`Ctrl+D` 删除 |
| Compose | `s` start、`Ctrl+S` stop、`Ctrl+D` down、`l` 日志、`d` 详情 |

资源过滤会随输入即时生效。`Enter` 保留过滤并退出编辑；5 秒内连续按两次 `Esc` 会清空过滤并退出。日志搜索在按下 `Enter` 后应用，使用 `n` / `Ctrl+N` 跳转到下一个 / 上一个匹配。

容器命令模式提供 `:rename`、`:top` 和 `:port`。Top 仅允许运行中的容器进入，支持 `j` / `k` 滚动和 `r` 刷新；Port 详情显示结构化的容器端口、协议、Host IP 与 Host Port。批量 Pause/Unpause 会跳过状态不适用的容器并汇总成功、跳过和失败数量。

完整按键和模式说明见 [交互与导航](docs/navigation.md)。运行时 Help 与 Footer 是当前有效快捷键的权威展示。

## 配置

默认配置路径：

```text
~/.config/docker-tui/config.yml
```

也可以通过 `--config` 指定其他 YAML 文件。配置会覆盖内嵌默认值，未设置字段继续使用默认配置。

最小示例：

```yaml
configVersion: 1

general:
  lang: zh
  sizeFormat: binary

runtime:
  default: local-docker
  discovery:
    localDocker: true
    localPodman: true
  health:
    intervalSec: 3
    timeoutSec: 2
    failureThreshold: 2
  connections:
    - name: remote-docker
      driver: docker
      endpoint: tcp://docker.example.com:2376
      apiVersion: "" # 留空时自动协商
      tls:
        enabled: true
        verify: true
        insecureSkipVerify: false
        caFile: /etc/docker/certs/ca.pem
        certFile: /etc/docker/certs/cert.pem
        keyFile: /etc/docker/certs/key.pem

logs:
  since: 1h
  tail: "200"
  timestamps: false

keymap:
  help: [f1]
  containerStart: [s]
  containerStop: [ctrl+s]
  containerPause: [p]
  imagePull: [ctrl+p]
```

连接配置只接受新的 `runtime` schema。旧的 `docker.host`、`docker.tlsVerify`、`docker.tlsCertPath` 和 `general.runtime` 字段不会迁移，加载时会直接返回配置错误。

可配置内容包括语言、运行时连接、Stats 轮询间隔、日志范围、布局、主题和动作快捷键。完整字段见 [默认配置](internal/data/config/default.jsonc) 与 [配置类型](internal/data/config/types.go)。

## 审计日志

容器、镜像、卷、网络、Compose、运行时切换和 Exec 等用户操作会生成统一 trace。每条审计记录包含操作、结果、运行时、UI 上下文和强类型资源目标。

默认落盘路径：

```text
~/.config/docker-tui/logs/audit-YYYY-MM-DD.jsonl
```

审计结果同时驱动顶部通知和 Footer 操作行。过滤、导航、后台刷新等非业务操作不会写入审计日志。

## 设计与架构

界面采用稳定的三段式布局：顶部状态区、中部资源工作区、底部查询/消息/Footer 区。资源页面与覆盖层共享顶层 `AppModel`，键盘动作经上下文解析后生成 Bubble Tea 命令，异步结果再回到统一 Update 循环。

```text
keyboard input
  -> action registry and context resolver
  -> tea.Cmd
  -> resource/action message
  -> state update and audit projection
  -> render
```

代码按职责分为：

```text
cmd/docker-tui/       CLI 与启动流程
internal/data/        配置、Docker/Podman、i18n、审计
internal/tui/state/   应用状态与消息
internal/tui/keys/    动作注册表和有效键位
internal/tui/keyboard/交互与业务命令
internal/tui/ui/      页面、组件和布局渲染
test/                 跨包测试、诊断、基准和集成测试
```

设计和实现资料：

- [当前 UI 设计](design/current-design.md)
- [架构说明](docs/architecture.md)
- [项目结构](docs/project-structure.md)
- [历史修复设计](design/bugfix-design.md)
- [历史修复需求](docs/bugfix-requirements.md)
- [后续需求与规划讨论](design/future-requirements-discussion.md)
- [后续需求实施任务清单](design/future-requirements-task-list.md)
- [文档索引](docs/README.md)

## 开发与测试

查看全部任务：

```bash
just --list
```

常用检查：

```bash
just check             # go vet + 全量 Go 测试
just test              # 全量 Go 测试
just test-unit         # cmd/ 与 internal/ 单元测试
just test-integration  # 构建并运行容器引擎集成测试
just bench             # 序列化基准测试
```

集成测试可能拉取镜像并创建带 `dtui-test-` 前缀的临时容器，测试退出时会执行清理。测试目录说明见 [test/README.md](test/README.md)。
