# Exec Shell — 容器 Shell

## 状态

- [ ] 待实现

## 触发

- 容器面板选中容器，按 **`e`** 键
- 直接启动子进程，不弹对话框

## 行为

```
e → Docker: docker exec -it <id> sh
  → Podman: podman exec -it <id> sh
  → 检测容器内是否有 bash/sh
  → 优先 bash，备用 sh
```

## Engine 适配

| 引擎     | 默认 Shell                  | 降级                       |
|--------|---------------------------|--------------------------|
| Docker | `docker exec -it <id> sh` | `ash`, `bash`, `/bin/sh` |
| Podman | `podman exec -it <id> sh` | 同上                       |

## 实现方式

Bubbletea `tea.ExecCommand` → 在终端中直接启动子进程:

```go
cmd := exec.Command("docker", "exec", "-it", containerID, "sh")
return tea.ExecProcess(cmd, func(err error) tea.Msg {
    if err != nil { return ExecError{Error: err} }
    return ExecDone{}
})
```

## 退出

- 容器 Shell 内输入 `exit` 或 `Ctrl+D`
- 自动返回 dtui 界面

## 降级策略

如果终端不支持 `tea.ExecCommand`:

- Toast: `Run: docker exec -it <id> sh`
- 用户手动粘贴到另一个终端

## 消息类型

```go
type ExecDone struct{}
type ExecError struct {
    ContainerID string
    Error       error
}
```
