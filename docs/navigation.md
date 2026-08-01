# 交互与导航

本文记录当前默认行为，基于 `internal/tui/keyboard/*`、`internal/tui/keys/*` 与页面渲染代码整理。表中的可配置动作均可通过 `config.yml` 的 `keymap.*` 覆盖；Help 和 Footer 会显示当前有效绑定。

## 面板与模式

### 主面板

- `containers`
- `images`
- `volumes`
- `networks`
- `compose`

### 覆盖层 / 模式

- `help`
- `detail`
- `logs`
- `filter`
- `command`
- `exec`
- `exec passthrough`
- `confirm`
- `mark`
- `rename`
- `top`

## 全局默认按键

| 按键 | 行为 |
|---|---|
| `Tab` / `Shift+Tab` | 在五个主面板之间切换 |
| `j` / `k` / `↑` / `↓` | 移动光标 |
| `Enter` | 进入当前面板的主动作 |
| `Esc` | 返回上一级；主界面双击 `Esc` 退出 |
| `/` | 打开过滤模式 |
| `:` | 打开命令模式 |
| `?` / `F1` | 打开帮助 |
| `r` | 刷新全部资源 |
| `H` | 显示/隐藏 Header |
| `F2` | 切换当前运行时连接 |
| `c` | 显示连接信息 |
| `Space` | 进入或操作标记模式 |
| `Ctrl+D` | 删除当前资源 |

按键名称统一做大小写归一化。Stop、Kill、Pull 等与小写单键动作冲突的操作使用 `Ctrl+S`、`Ctrl+K`、`Ctrl+P` 组合键。

## 面板级行为

### Containers

| 按键 | 行为 |
|---|---|
| `s` | Start |
| `Ctrl+S` | Stop |
| `Ctrl+R` | Restart |
| `Ctrl+K` | Kill |
| `l` | 打开日志视图 |
| `m` | 开关容器 stats |
| `p` | 运行中容器 Pause；已暂停容器 Unpause |
| `d` | 详情视图 |
| `i` | Inspect |
| `e` | 自动探测可用 shell 并进入容器；探测失败时打开 shell 对话框 |
| `o` / `Ctrl+O` | 切换排序列 / 切换升降序 |

### Images

| 按键 | 行为 |
|---|---|
| `Enter` / `→` | 展开“使用该镜像的容器”子视图 |
| `Ctrl+P` | Pull |
| `p` | Prune |
| `Ctrl+B` | Debug 入口 |
| `Ctrl+E` | Export 入口 |
| `y` | 复制镜像引用 |
| `d` | 详情 |
| `o` / `Ctrl+O` | 排序 |

镜像子视图中可对映射出的容器执行 `s`、`Ctrl+S`、`Ctrl+R`、`l`、`e` 等操作。

### Volumes

| 按键 | 行为 |
|---|---|
| `Enter` | 展开卷详情/关联容器子视图 |
| `d` | 详情 |
| `Ctrl+D` | 删除 |

### Networks

| 按键 | 行为 |
|---|---|
| `d` | 详情 |
| `Ctrl+D` | 删除 |
| `o` / `Ctrl+O` | 排序 |

### Compose

Compose 面板是双栏：

- 左栏：项目
- 右栏：服务

常用按键：

| 按键 | 行为 |
|---|---|
| `Tab` | 主面板切换，不在 Compose 左右栏之间切换 |
| `←` / `→` | 在 Compose 内部左右栏或子视图间切换 |
| `Enter` | 进入服务容器子视图或详情 |
| `s` | Compose start |
| `Ctrl+S` | Compose stop |
| `l` | Compose logs |
| `Ctrl+D` | Compose down |
| `d` | 项目详情 |

## 过滤与命令模式

### 过滤模式

- `/` 进入
- 输入即时应用过滤，`Enter` 保留当前过滤并退出编辑
- 第一次 `Esc` 进入 5 秒待退出窗口，第二次 `Esc` 清空过滤并退出
- 支持左右移动、`Ctrl+B` / `Ctrl+F`、`Alt+B` / `Alt+F`、`Home` / `End`、`Ctrl+A` / `Ctrl+E`
- 支持 `Backspace` / `Delete`、`Ctrl+H`、`Ctrl+W`、`Ctrl+U`、`Ctrl+K`
- 在日志视图中，`/` 进入独立搜索模式；`Enter` 应用并跳转到首个匹配，`Esc` 取消草稿

### 命令模式

由 `:` 进入，当前支持的命令由 `internal/tui/keyboard/command.go` 定义：

- `compose`
- `images`
- `containers`
- `volumes`
- `networks`
- `logs`
- `rename`：打开当前容器重命名输入框
- `top`：打开运行中容器的独立进程页面
- `port`：打开当前容器的结构化端口详情
- `help`

支持 `Tab` 自动补全。

命令输入与过滤输入共享上述 shell 风格光标移动和删除操作。

## 覆盖层行为

### Help

- 使用当前 Help 绑定或 Back 绑定关闭，并返回之前的面板

### Detail

- `Esc` / `Enter` 返回
- `j` / `k` / 鼠标滚轮滚动
- `Space` / `PgDn` 向下翻页
- `PgUp` 向上翻页
- `g` 回到顶部

### Top

- 仅运行中的容器可以进入
- `j` / `k` / `↑` / `↓` 移动进程行
- `r` 重新读取进程列表
- `Esc` 返回容器列表

### Logs

- `Esc` 返回
- `j` / `k` / `PgUp` / `PgDn` / 鼠标滚轮滚动
- `/` 进入搜索，`Enter` 应用
- `n` / `Ctrl+N` 跳到下一个 / 上一个匹配
- `g` 回到顶部，`Ctrl+G` 跳到底部
- `w` 切换自动换行

### Exec

分两段：

1. `ModeExec`：自动探测失败后，可选择 `/bin/sh`、`/bin/bash`、`/bin/ash` 或输入自定义 shell
2. `ModeExecPassthrough`：进入真实容器终端透传

在 exec 对话框中：

- `Tab` / `Shift+Tab` 切换焦点
- `Enter` 执行或确认
- `Esc` 取消

在 exec 透传中：

- 普通按键直接写入容器会话
- `Esc` 关闭当前 exec 会话并返回
