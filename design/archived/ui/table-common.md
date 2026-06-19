# Table View — 主表格视图

## 状态

- [x] 已实现 (容器/镜像/卷/网络表格)

## 适用面板

- PanelContainers
- PanelImages  
- PanelVolumes
- PanelNetworks

## 容器列表 (Container)

```
┌─────────────────────────────────────────────────────────┐
│  Containers                                             │
│                                                         │
│  ID            NAME              IMAGE         STATUS   │
│  ▸abc123def456  web-app           nginx:latest  running  │
│   def456abc789  redis-cache       redis:7       running  │
│   ghi789jkl012  postgres-db       postgres:16   running  │
│   jkl012mno345  stopped-app       alpine:3      exited   │
│                                                         │
│  1-4/4                                                  │
└─────────────────────────────────────────────────────────┘
```

### 列定义

| 列 | 宽度 | 说明 |
|---|---|---|
| ID | 20 | 容器 ID 前20字符 |
| NAME | 25 | 容器名称，超长截断 |
| IMAGE | 20 | 镜像名:标签 |
| STATUS | 15 | 状态 (running/exited/paused) 带颜色 |
| CPU/MEM | — | 实时 stats (m 键开启后显示) |

### 状态颜色

| 状态 | 颜色 |
|---|---|
| running | 绿色 |
| exited, dead | 红色 |
| paused | 黄色 |
| created | 蓝色 |
| restarting | 橙色 |

### 交互

| 按键 | 操作 |
|---|---|
| `j`/`↓` | 下移光标 |
| `k`/`↑` | 上移光标 |
| `Ctrl+D` | 删除 (确认弹窗，支持强制/卷/链接选项) |
| `Space` | 标记/取消标记 |
| `s` | 启动 |
| `S` | 停止 |
| `R` | 重启 |
| `l` | 查看日志 |
| `m` | 切换 Stats 统计 |
| `Enter` | 进入详情 (图片) |

## 镜像列表 (Image)

> **详细设计见**: [image-table.md](image-table.md)

### 列定义 (更新后)

| 列 | 宽度 | 说明 |
|---|---|---|
| NAME | 25 | 镜像名称 (不含 tag) |
| TAG | 15 | 版本标签 |
| IMAGE ID | 20 | 镜像 ID 前 20 字符 |
| DIGEST | 19 | RepoDigests 摘要 (可为空) |
| CREATED | 19 | 创建时间 |
| SIZE | 10 | 格式化大小 |

### 排序

| 按键 | 排序方式 |
|---|---|
| `o1` | NAME 字母序 |
| `o2` | 创建时间 |
| `o3` | 文件大小 |
| `O` | 反转排序方向 |

### 交互

| 按键 | 操作 |
|---|---|
| `P` | 拉取镜像 |
| `p` | 清理悬空镜像 |
| `Enter` | **查看使用该镜像的容器列表** |
| `d` | **查看镜像详细信息** |
| `Ctrl+D` | 删除镜像 (确认弹窗) |
| `o` | 切换排序列 |
| `O` | 反转排序方向 |
| `Space` | 标记 |
| `/` | 搜索/过滤 |

## 卷列表 (Volume)

| 列 | 说明 |
|---|---|
| NAME | 卷名称 |
| DRIVER | 存储驱动 |
| MOUNTPOINT | 挂载路径 |

## 网络列表 (Network)

| 列 | 说明 |
|---|---|
| NAME | 网络名称 |
| DRIVER | 网络驱动 |
| SCOPE | 作用域 |
| SUBNET | 子网地址 |

## Viewport 滚动

- `ViewOffset` 跟踪可见区域起始行
- `ensureCursorVisible()` 自动滚动跟随光标
- Footer 显示 `起始-结束/总数`
- 超出可见区域的行不渲染，节省性能
