# 容器 Exec 页面 shell UI

> 任务卡 / BR-038

## 元信息

- **关联编号**:BR-038
- **优先级**:medium
- **状态**:`open`
- **依赖**:TASK-017(`Exec` 动作已实现)
- **关联设计**:[docs/feature-design.md §5.4](../feature-design.md)

## 目标

把当前容器页 Exec 实现从"dialog 风格"升级为**全屏 shell 终端**,提供真实的 PTY 透传、shell 行编辑(↑↓ 历史 / Tab 补全 / Ctrl+L 清屏 / Ctrl+C 转发 SIGINT / Ctrl+D EOF 脱离),符合用户对"shell 页面"的预期。

## 代码结构索引

### 必须读懂的文件

| 文件 | 作用 |
|---|---|
| `internal/tui/ui/widget/dialog/exec.go` | 当前 Exec 对话框(需被替换或扩展为全屏模式) |
| `internal/tui/keyboard/container_action.go:227` | `doAutoExecAction` 入口 |
| `internal/data/runtime/streaming.go` | `ExecService` / `ExecSession` 接口(已有) |
| `internal/data/runtime/docker/service_exec.go` | Docker exec PTY 实现(已有) |
| `internal/data/runtime/podman/service_exec.go` | Podman exec PTY 实现(已有) |
| `internal/tui/term/buffer.go` | 终端缓冲渲染(可能复用) |
| `internal/tui/ui/widget/dialog/exec.go` | `RenderExecDialog` 模板 |
| `internal/tui/state/log.go`(参考) | 日志页 state 模式,Exec 状态类似 |

### 必须修改的文件

| 文件 | 改动 |
|---|---|
| `internal/tui/state/exec.go`(或新建) | `ExecState`:`ContainerID` / `Session runtimeapi.ExecSession` / `Buffer` / `Mode(ModeExec)` |
| `internal/tui/state/navigation.go` | 加 `ModeExec` 枚举 |
| `internal/tui/keyboard/container_action.go:227` | `doAutoExecAction` 改为进入 `ModeExec`,不再走 dialog |
| `internal/tui/keyboard/exec_keys.go`(新) | `handleExecKeys`:↑↓ 历史 / Tab 补全 / Ctrl+L 清屏 / Ctrl+C SIGINT / Ctrl+D EOF / Esc 脱离 |
| `internal/tui/ui/pages/exec/view.go`(新) | 全屏 exec 渲染:shell 提示符 + 输出 + 输入框 |
| `internal/tui/ui/app/layout.go` | 注册 `ModeExec` → 渲染 exec 页面 |

### 必须新增的文件

- `internal/tui/state/exec.go`
- `internal/tui/keyboard/exec_keys.go`
- `internal/tui/ui/pages/exec/view.go`
- `internal/tui/ui/pages/exec/view_test.go`

## 操作链路

```text
User 在容器页按 e (ActionContainerExec)
  → keyboard/actions.go:case ActionContainerExec
  → doAutoExecAction(m)
      检查 ActivePanel == PanelContainers, Selected != nil
      构造 ExecOptions{Command: []string{"/bin/sh"}, TTY: true, AttachStdin/Stdout/Stderr: true}
      cmd := Engine.Exec().Open(ctx, containerID, opts)
        Docker: cli.ContainerExecCreate + ContainerExecAttach
        Podman: REST exec start
      返回 ExecSession(可读可写)
  → state.Exec.Open(containerID, session)
      m.Navigation.Mode = state.ModeExec
  → ui/pages/exec/view.go:RenderView(m, h, w)
      全屏渲染 shell 提示符(`containerID:/#`)+ PTY 输出 + 输入框
  → keyboard/exec_keys.go:handleExecKeys
      普通字符 → Write to session (PTY stdin)
      Enter → 写入 `\n`
      ↑↓ → 翻本地命令历史(input_edit.go 类似)
      Tab → 写入 `\t`(由 shell 自身补全)
      Ctrl+L → 写入 `\x0c`(清屏)
      Ctrl+C → 写入 `\x03`(SIGINT)
      Ctrl+D → EOF,若 shell 退出则自动脱离 session
      Esc → 脱离,关闭 session, ModeExec → ModeNormal
  → 后台 goroutine:session.Read → 写入 state.Exec.Buffer
  → 视图:从 Buffer 渲染 ANSI / 宽字符
```

## 验收标准

- [ ] 容器页按 `e` 进入 exec,页面**全屏 shell**(`Prompt` 正确显示 `containerID:/#`)。
- [ ] 在 shell 内执行 `ls /` / `cat /etc/hostname` / `top` 等命令,输出正确(包括 ANSI 颜色 / 进度条 / 宽字符)。
- [ ] ↑↓ 翻历史(最近 50 条命令)、Tab 补全、Ctrl+L 清屏、Ctrl+C 发 SIGINT 均工作。
- [ ] Ctrl+D 退出前台 shell 但容器继续运行(若 detach 后还在容器里,shell 会 fork 新进程)。
- [ ] Esc 退出 dtui 的 exec mode,**不**杀掉容器前台进程。
- [ ] 容器 exit(前台进程退出)后自动脱离 session。
- [ ] exec 失败 / 容器未运行 / shell 不存在时显示明确错误,不卡死。
- [ ] 多个容器 session 不能并行(ModeExec 互斥)。
- [ ] Help / Footer 显示 `e Exec` 提示;ModeExec 下 footer 显示 shell 快捷键。

## 风险

| 风险 | 说明 |
|---|---|
| **PTY 兼容** | Docker / Podman 的 PTY 行为差异(信号转发、EOF 处理) |
| **ANSI 处理** | shell 输出可能含颜色 / 控制序列,需要 `internal/utils` 的 ANSI 长度计算 |
| **输入缓冲** | 长命令输入与历史回看的交互 |
| **resize** | 终端窗口变宽 / 变高时,需要通过 `ExecSession.Resize(W, H)` 通知前端 PTY |
| **Ctrl+C 转发** | 误转发可能杀掉容器前台进程;需要 `ExitPropagation=false`(默认 Docker 配置) |
| **历史与 Resize 协同** | Resize 时 Buffer 重渲染性能(参考 BR-016 ClampVisibleOffset) |
| **Help / Footer 漂移** | exec 全屏模式覆盖整个 panel,原有的 footer 快捷键提示如何展示? |

## 建议任务分解

1. **TASK-BR038-A:state + session 管理**
   - `state/exec.go` + session open/close 生命周期
   - **预估**:主模型实现
2. **TASK-BR038-B:全屏渲染 + ANSI**
   - `pages/exec/view.go` + 终端 buffer 渲染
   - **预估**:小模型辅助,主模型审渲染
3. **TASK-BR038-C:键盘 + shell 行编辑**
   - `keyboard/exec_keys.go` + 历史 + 补全 + 信号
   - **预估**:主模型审交互
4. **TASK-BR038-D:Resize 集成 + i18n + 文档**
   - 窗口 resize 通知 PTY + Help 同步
   - **预估**:小模型机械

## 待确认项

- [ ] **detach 语义**:用户按 Esc 是 detach(保留容器)还是 exit(杀死前台进程)?默认 detach
- [ ] **shell 选择**:固定 `/bin/sh` 还是按容器镜像检测(`/bin/bash` / `/bin/sh` / `powershell`)?
- [ ] **历史作用域**:每个容器独立历史 vs 全局历史?
- [ ] **i18n**:shell 提示符是否本地化?(`容器:$/#`)
- [ ] **多 tab**:支持同时开多个 exec session(切换 tab)还是一次只能一个?
- [ ] **窗口 resize**:是否监听 SIGWINCH 并通知 PTY,还是只在启动时 resize?