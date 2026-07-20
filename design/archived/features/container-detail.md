# Container Detail View — 容器详细信息

## 状态

- [x] 已实现

## 触发

- 任意资源面板，选中行，按 **`d`** 键
- 所有表格统一: `d` = Detail 详细信息
- 替换主内容区域，全屏显示

## 设计原则

- **全屏沉浸**: 不再使用右侧分栏，直接全屏展示
- **分区折叠**: 信息按类别分区，可折叠/展开
- **可滚动**: 超出一屏时支持上下滚动
- **快捷跳转**: `s`/`S` 可直接在详情页操作容器

## 布局

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ Container Detail: nginx-container (abc123de)                              [Esc: back] │
│                                                                                      │
│  ── 基本信息 ────────────────────────────────────────────────────────────── [1 折叠] │
│  ID:            abc123def456789...                                                    │
│  Name:          /nginx-container                                                      │
│  Image:         nginx:latest                                                         │
│  State:         ● running                                                            │
│  Status:        Up 3 hours                                                           │
│  Created:       2024-01-15 10:00:00 UTC                                              │
│  Started:       2024-01-15 10:30:00 UTC                                              │
│  Finished:      —                                                                    │
│  RestartCount:  0                                                                    │
│  Platform:      linux/amd64                                                          │
│                                                                                      │
│  ── 网络信息 ────────────────────────────────────────────────────────────── [2 折叠] │
│  Network Mode:  bridge                                                               │
│  ┌──────────────┬──────────────┬──────────────────────┬────────────────────┐         │
│  │ Network      │ IP Address   │ Gateway              │ MAC Address        │         │
│  │ bridge       │ 172.17.0.2   │ 172.17.0.1           │ 02:42:ac:11:00:02  │         │
│  └──────────────┴──────────────┴──────────────────────┴────────────────────┘         │
│                                                                                      │
│  Ports:                                                                              │
│  ┌──────────────┬──────────────┬──────────┐                                          │
│  │ Host         │ Container    │ Type     │                                          │
│  │ 0.0.0.0:80   │ 80           │ tcp      │                                          │
│  │ 0.0.0.0:443  │ 443          │ tcp      │                                          │
│  └──────────────┴──────────────┴──────────┘                                          │
│                                                                                      │
│  DNS:            8.8.8.8, 1.1.1.1                                                    │
│  Hostname:       abc123def456                                                        │
│  DomainName:     —                                                                    │
│                                                                                      │
│  ── 资源使用 ────────────────────────────────────────────────────────────── [3 折叠] │
│  CPU:            2.1%                                                                 │
│  Memory:         45.2% (256MB / 512MB)                                                │
│  PIDs:           12                                                                   │
│                                                                                      │
│  Limits:                                                                              │
│  CPU Limit:      1.0 cores                                                            │
│  Memory Limit:   512 MB                                                               │
│                                                                                      │
│  ── 挂载卷 ──────────────────────────────────────────────────────────────── [4 折叠] │
│  ┌──────────────────────────┬──────────────────────────────────────┬──────┐           │
│  │ Source                   │ Destination                          │ Mode │           │
│  │ /data/nginx/html         │ /usr/share/nginx/html                │ rw   │           │
│  │ /data/nginx/conf         │ /etc/nginx/conf.d                    │ ro   │           │
│  │ nginx-logs               │ /var/log/nginx                       │ rw   │           │
│  └──────────────────────────┴──────────────────────────────────────┴──────┘           │
│                                                                                      │
│  ── 配置信息 ────────────────────────────────────────────────────────────── [5 折叠] │
│  Cmd:            ["nginx", "-g", "daemon off;"]                                       │
│  Entrypoint:     ["/docker-entrypoint.sh"]                                            │
│  WorkingDir:     /                                                                    │
│  User:           nginx                                                                │
│                                                                                      │
│  Env:                                                                                 │
│    NGINX_VERSION=1.25.3                                                               │
│    PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin                  │
│    NGINX_SETUP_PATH=/docker-entrypoint-initdb.d                                       │
│                                                                                      │
│  ExposedPorts:  80/tcp, 443/tcp                                                       │
│                                                                                      │
│  ── Compose ─────────────────────────────────────────────────────────────── [6 折叠] │
│  Project:       my-web-app                                                            │
│  Service:       nginx                                                                 │
│  Config:        docker-compose.yml                                                    │
│  WorkingDir:    /home/user/projects/my-web-app                                        │
│                                                                                      │
│  ── 标签 ────────────────────────────────────────────────────────────────── [7 折叠] │
│  com.docker.compose.project=my-web-app                                                │
│  com.docker.compose.service=nginx                                                     │
│  maintainer=devops@example.com                                                        │
│  environment=production                                                               │
│                                                                                      │
│  Press s:Start  S:Stop  R:Restart  l:Logs  Enter:Shell  Esc:back                     │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

## 信息分区

| 分区 | 快捷键 | 内容 | 默认 |
|---|---|---|---|
| **基本信息** | `1` | ID, Name, Image, State, Status, Created, Started, Finished, RestartCount, Platform | 展开 |
| **网络信息** | `2` | NetworkMode, IP 表格, Ports 表格, DNS, Hostname | 展开 |
| **资源使用** | `3` | CPU%, Memory%, PIDs, Limits | 展开 |
| **挂载卷** | `4` | Mounts 表格 (Source, Destination, Mode) | 折叠 |
| **配置信息** | `5` | Cmd, Entrypoint, WorkingDir, User, Env, ExposedPorts | 折叠 |
| **Compose** | `6` | Project, Service, Config, WorkingDir | 有则展开 |
| **标签** | `7` | Labels 列表 | 折叠 |

> Tab 在分区间跳转焦点，Space/Enter 折叠/展开

## 键盘交互

| 按键 | 操作 |
|---|---|
| `↑`/`↓`/`j`/`k` | 滚动内容 |
| `PgUp`/`PgDn` | 翻页 |
| `1`–`7` | 跳转到对应分区，切换折叠 |
| `Tab` | 分区焦点跳转 |
| `s` | **Start** 容器 (详情页直接操作!) |
| `S` | **Stop** 容器 |
| `R` | **Restart** 容器 |
| `l` | 查看日志 |
| `Enter` | Exec Shell 提示 |
| `Ctrl+D` | 删除容器 (确认弹窗) |
| `Esc` | 返回容器列表 |

## 数据来源

`docker.InspectContainer(id)` → `ContainerInspect()` API

所有字段从 `types.ContainerJSON` 提取，单次 API 调用覆盖全部信息。

## 样式

| 元素 | 样式 |
|---|---|
| 标题 | TitleStyle (青色) |
| 分区标题 | HeaderStyle (蓝色) |
| 字段名 | DimStyle (灰色) |
| 字段值 | 白色 |
| 状态值 | 彩色 (running=绿, exited=红, ...) |
| 表格边框 | 细线 |
| 快捷键提示 | ShortcutKeyStyle |
