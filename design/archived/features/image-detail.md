# Image Detail View — 镜像详细信息

## 状态

- [x] 已实现

## 触发

- 镜像面板选中某行，按 **`d`** 键
- 替换主内容区域，全屏显示

## 折叠按钮

每个分区标题行右侧显示折叠按钮 `[-]` 展开 / `[+]` 折叠：

```
── 基本信息 ──────────────────────────────────────────── [-]
── 系统信息 ──────────────────────────────────────────── [-]
── 配置信息 ──────────────────────────────────────────── [+]
── 存储信息 ──────────────────────────────────────────── [+]
── 镜像历史 ──────────────────────────────────────────── [+]
```

> 默认全部展开 `[-]`，按 `Tab` 跳转焦点，`Space`/`Enter` 切换折叠。

## 布局

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│ Image Detail: nginx:latest                                            [Esc: back] │
│                                                                                  │
│  ── 基本信息 ──────────────────────────────────────────────────────────── [-] ── │
│                                                                                  │
│  ID            sha256:aaa111bbb222ccc333ddd444eee555fff666ggg777hhh888iii999     │
│  Registry      docker.io                                                         │
│  Name          library/nginx                                                     │
│  Tag           latest                                                            │
│                                                                                  │
│  Full Tags     nginx:latest, nginx:stable, nginx:1.25, nginx:alpine              │
│  Digest        sha256:abc123def456789abc123def456789abc123def456789abc123def456   │
│                                                                                  │
│  Created       2024-01-15 10:00:00 UTC (3 months ago)                            │
│  Size          187.0 MB (187,000,000 bytes)                                      │
│  Virtual Size  210.5 MB (includes shared layers)                                 │
│  Shared Size   45.2 MB (shared with other images)                                │
│  Containers    3 containers using this image                                      │
│                                                                                  │
│  ── 系统信息 ──────────────────────────────────────────────────────────── [-] ── │
│                                                                                  │
│  Architecture  amd64                                                             │
│  OS            linux                                                             │
│  OS Version    Debian GNU/Linux 12 (bookworm)                                    │
│  OS Features   —                                                                 │
│  Variant       —                                                                 │
│                                                                                  │
│  ── 配置信息 ──────────────────────────────────────────────────────────── [-] ── │
│                                                                                  │
│  Author        NGINX Docker Maintainers <docker@nginx.com>                       │
│  Comment       Stable production release of nginx                                │
│                                                                                  │
│  ┌─ Runtime ─────────────────────────────────────────────────────────────────┐   │
│  │ WorkingDir     /                                                          │   │
│  │ User           nginx:nginx (uid=101, gid=101)                             │   │
│  │ StopSignal     SIGQUIT                                                    │   │
│  │ StopTimeout    10s                                                        │   │
│  │ Tty            false                                                      │   │
│  │ OpenStdin      false                                                      │   │
│  │ StdinOnce      false                                                      │   │
│  └───────────────────────────────────────────────────────────────────────────┘   │
│                                                                                  │
│  ┌─ Entrypoint / Cmd ────────────────────────────────────────────────────────┐   │
│  │ Entrypoint     ["/docker-entrypoint.sh"]                                  │   │
│  │ Cmd            ["nginx", "-g", "daemon off;"]                             │   │
│  │ Shell          ["/bin/sh", "-c"]                                          │   │
│  │ OnBuild        —                                                          │   │
│  └───────────────────────────────────────────────────────────────────────────┘   │
│                                                                                  │
│  ┌─ Environment ─────────────────────────────────────────────────────────────┐   │
│  │ #  KEY                    VALUE                                           │   │
│  │ 1  NGINX_VERSION          1.25.3                                          │   │
│  │ 2  PATH                   /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin│   │
│  │ 3  NGINX_SETUP_PATH       /docker-entrypoint-initdb.d                     │   │
│  │ 4  LANG                   C.UTF-8                                         │   │
│  │ 5  TZ                     UTC                                             │   │
│  └───────────────────────────────────────────────────────────────────────────┘   │
│                                                                                  │
│  ┌─ Exposed Ports ───────────────────────────────────────────────────────────┐   │
│  │  PORT/PROTO                                                                 │
│  │  80/tcp                                                                     │
│  │  443/tcp                                                                    │
│  └───────────────────────────────────────────────────────────────────────────┘   │
│                                                                                  │
│  ┌─ Volumes ────────────────────────────────────────────────────────────────┐   │
│  │  CONTAINER PATH               MODE                                        │   │
│  │  /var/cache/nginx             rw                                          │   │
│  │  /etc/nginx/conf.d            ro (config volume)                          │   │
│  │  /var/log/nginx               rw (logs volume)                            │   │
│  └───────────────────────────────────────────────────────────────────────────┘   │
│                                                                                  │
│  ┌─ Healthcheck ─────────────────────────────────────────────────────────────┐   │
│  │  Test        ["CMD", "curl", "-f", "http://localhost/", "||", "exit", "1"]│  │
│  │  Interval    30s                                                          │   │
│  │  Timeout     5s                                                           │   │
│  │  Retries     3                                                            │   │
│  │  StartPeriod 5s                                                           │   │
│  └───────────────────────────────────────────────────────────────────────────┘   │
│                                                                                  │
│  ── 存储信息 ──────────────────────────────────────────────────────────── [+] ── │
│                                                                                  │
│  (折叠中...)                                                                      │
│                                                                                  │
│  ── 标签 ──────────────────────────────────────────────────────────────── [-] ── │
│                                                                                  │
│  #  KEY                                    VALUE                                 │
│  1  maintainer                            NGINX Docker Maintainers              │
│  2  org.opencontainers.image.version      1.25.3                                 │
│  3  org.opencontainers.image.created      2024-01-15                             │
│  4  com.docker.extension.version          0.2.0                                  │
│                                                                                  │
│  ── 镜像历史 ────────────────────────────────────────────────────────── [+] ──   │
│                                                                                  │
│  (折叠中...)                                                                      │
│                                                                                  │
│  s:Start  S:Stop  R:Restart  l:Logs  Enter:Containers  Esc:back                 │
└──────────────────────────────────────────────────────────────────────────────────┘
```

## 信息分区 (7区)

| # | 分区 | 内容 | 默认 |
|---|---|---|---|
| 1 | **基本信息** | ID, Registry, Name, Tag, Tags, Digest, Created, Size, VirtualSize, SharedSize, Containers | 展开 |
| 2 | **系统信息** | Architecture, OS, OSVersion, OSFeatures, Variant | 展开 |
| 3 | **配置信息** | Runtime(WorkingDir/User/StopSignal), Entrypoint/Cmd/Shell/OnBuild, Env表格, ExposedPorts, Volumes表格, Healthcheck | 展开 |
| 4 | **存储信息** | Driver, GraphDriver数据, RootFS类型, Layers列表 | 折叠 |
| 5 | **标签** | Labels表格 (key=value) | 展开 |
| 6 | **元数据** | Parent, Container, DockerVersion, 创建容器命令 | 折叠 |
| 7 | **镜像历史** | History表格 (CreatedBy, Size, Comment) | 折叠 |

## 配置信息增强

### Env 表格

环境变量以带编号表格显示，便于阅读长变量值:

```
#  KEY                    VALUE
1  NGINX_VERSION          1.25.3
2  PATH                   /usr/local/sbin:/usr/local/bin:...
```

### Healthcheck 详表

```
Test        ["CMD", "curl", "-f", "http://localhost/", "||", "exit", "1"]
Interval    30s
Timeout     5s
Retries     3
StartPeriod 5s
```

### Volumes 表格

```
CONTAINER PATH          MODE
/var/cache/nginx        rw
/etc/nginx/conf.d       ro (config volume)
```

## 键盘交互

| 按键 | 操作 |
|---|---|
| `↑`/`↓`/`j`/`k` | 滚动 |
| `PgUp`/`PgDn` | 翻页 |
| `Tab` / `S-Tab` | 跳转分区焦点 |
| `Space` / `Enter` | 折叠/展开当前分区 |
| `1`–`7` | 快速跳转分区 |
| `Enter` | **查看使用该镜像的容器列表** |
| `Esc` | 返回镜像列表 |

## 数据来源

`docker.InspectImage(id)` → `ImageInspectWithRaw()` API

全部字段来自 `types.ImageInspect`，单次 API 调用。
