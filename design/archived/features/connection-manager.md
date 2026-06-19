# Connection Manager — 环境选择器

## 状态

- [ ] 待实现

## 触发

- 全局快捷键 **`c`** (connect)
- 弹出环境选择对话框

## 数据来源

1. **自动检测**: 当前运行时检测到的 Docker/Podman socket
2. **配置文件**: `~/.config/dtui/config.yml` 中的 `docker.hosts` 列表

## 配置文件扩展

```yaml
# ~/.config/dtui/config.yml
docker:
  host: ""                # 默认自动检测
  hosts:                  # 预定义环境列表
    - name: "local-docker"
      host: "unix:///var/run/docker.sock"
    - name: "local-podman"
      host: "unix:///run/user/1000/podman/podman.sock"
    - name: "staging"
      host: "tcp://192.168.1.100:2375"
      tls: false
    - name: "prod-us"
      host: "tcp://10.0.1.50:2376"
      tls: true
      tlsCertPath: "~/.docker/certs/prod-us"
    - name: "dev-server"
      host: "ssh://devbox.example.com"
```

## 对话框布局

```
┌──────────────────────────────────────────────────┐
│  🔌 Select Environment                           │
│                                                  │
│  Current: ● podman (local)                       │
│                                                  │
│  ┌─ Local ──────────────────────────────────┐    │
│  │ ▸ Docker  (unix:///var/run/docker.sock)  │    │
│  │   Podman  (unix:///run/user/.../podman)  │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌─ Remote ─────────────────────────────────┐    │
│  │   staging    (tcp://192.168.1.100:2375)   │    │
│  │   prod-us    (tcp://10.0.1.50:2376)  🔒  │    │
│  │   dev-server (ssh://devbox.example.com)   │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  [Enter] Connect  [n] New  [d] Delete  [Esc] Cancel│
└──────────────────────────────────────────────────┘
```

## 键盘交互

| 按键 | 操作 |
|---|---|
| `↑`/`↓`/`j`/`k` | 选择环境 |
| `Enter` | 连接到选中环境 |
| `n` | 新建环境配置 (进入编辑模式) |
| `d` | 删除选中环境配置 |
| `Esc` / `c` | 关闭对话框 |

## 连接流程

```
按 c → 弹出选择器
  → 选中目标 → Enter
  → dockerclient.NewClient(selectedHost)
  → 成功: Toast "✓ Connected to staging (tcp://192.168.1.100:2375)"
  → 失败: Toast "✕ staging: connection refused"
  → 刷新所有资源列表
  → 状态栏更新: ● staging │ 5 containers
```

## 全局访问

- 所有面板均可按 `c` 打开
- 不局限于某个资源面板
- 在日志/详情模式下也可使用

## 快捷键栏

```
... │ c Connect │ r Refresh │ / Filter │ ? Help │ q Quit
```
