# Container Table View — 容器表格

## 状态

- [x] 已实现

## 设计原则

- **表格即主视图**: 容器信息全部在表格中呈现，不再需要右侧详情面板
- **按 `d` 查看完整详情**: 所有表格统一 `d` = Detail (全屏详情页)
- **按 `s`/`S` 快速操作**: 单键启停，高频操作零延迟

## 列定义

| 列 | 宽度 | 数据来源 | 排序 | 搜索 | 说明 |
|---|---|---|---|---|---|
| **ID** | 12 | `ContainerSummary.ID` | ✅ | ✅ | 容器 ID 前 12 字符 |
| **NAME** | 20 | `ContainerSummary.Name` | ✅ | ✅ | 容器名称 |
| **IMAGE** | 18 | `ContainerSummary.Image` | ✅ | ✅ | 镜像名:标签 |
| **STATE** | 8 | `ContainerSummary.State` | ✅ | ✅ | running/exited/paused/created |
| **CREATED** | 14 | `ContainerSummary.Created` | ✅ | — | `01-15 10:00` 月-日 时:分 |
| **IP** | 15 | `NetworkSettings.Networks[].IPAddress` | — | — | 主网络 IP |
| **PORTS** | 18 | `ContainerSummary.Ports` | — | ✅ | 端口映射 |
| **MOUNTS** | 6 | `Mounts[].count` | — | — | 挂载数量 `3vol` |
| **CPU** | 6 | `ContainerListModel.Stats` | ✅ | — | 实时 CPU% (m 键开启) |
| **MEM** | 6 | `ContainerListModel.Stats` | ✅ | — | 实时 MEM% (m 键开启) |

## 布局

### 正常模式

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ Containers                                                                  [filter]  │
│                                                                                      │
│ ID        NAME           IMAGE          STATE    CREATED       IP             PORTS        M CPU MEM│
│ ▸abc123de web-app        nginx:latest   running  01-15 10:00   172.17.0.2     0.0.0.0:80→80  2v 2.1 45│
│  def456ab redis-cache    redis:7-alpine running  01-14 08:30   172.17.0.3     6379→6379      1v 1.5 32│
│  ghi789jk postgres-db    postgres:16    running  01-16 14:20   172.17.0.4     5432→5432      3v 3.2 60│
│  jkl012mn stopped-app    alpine:3.19    exited   01-10 06:00   —              —             —  —  —│
│  mno345pq worker-1       python:3.12    running  01-12 11:00   172.17.0.5     —             0v —  —│
│  pqr678st builder        golang:1.22    created  —              —             —  —  —│
│                                                                                      │
│  1-6/6 │ 4 running, 1 exited, 1 created │ Sort: name ▲                                │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

### Stats 开启时 (按 `m`)

```
... CPU MEM ...
... 2.1 45 ...
... 1.5 32 ...
... 3.2 60 ...
...  —   — ...
```

> CPU/MEM 列仅在 Stats 模式时显示，节省横向空间。关闭时隐藏这两列。

### Footer 信息

```
1-6/6 │ 4 running, 1 exited, 1 created │ Sort: name ▲ │ 3 marked
```

## 状态颜色

| 状态 | 颜色 | 图标 |
|---|---|---|
| `running` | 绿色 | `●` |
| `exited` | 红色 | `◼` |
| `dead` | 红色 | `✕` |
| `paused` | 黄色 | `⏸` |
| `created` | 蓝色 | `○` |
| `restarting` | 橙色 | `↻` |

## 排序

| 按键 | 排序字段 | 说明 |
|---|---|---|
| `o1` | NAME | 字母序 |
| `o2` | STATE | running → exited → paused → created |
| `o3` | IMAGE | 镜像名 |
| `o4` | CPU (desc) | 高负载在前 |
| `o5` | MEM (desc) | 高内存在前 |
| `O` | 反转方向 | asc ↔ desc |

## 搜索/过滤 (`/`)

匹配范围: NAME, IMAGE, ID, STATE, PORTS

> 搜索后 `Enter` 确认 → 光标聚焦到第一个匹配项

## 键盘交互

### 导航

| 按键 | 操作 |
|---|---|
| `↑`/`↓`/`j`/`k` | 移动光标 |
| `Tab` | 切换到下一个面板 (镜像) |
| `PgUp`/`PgDn` | 翻页 |
| `g` | 跳到第一个 |
| `G` | 跳到最后一个 |

### 快速操作 (单键直达)

| 按键 | 操作 | 说明 |
|---|---|---|
| `s` | **Start** | 启动已停止的容器 |
| `S` | **Stop** | 停止运行中的容器 |
| `R` | **Restart** | 重启容器 |
| `K` | **Kill** | 强制杀死 |
| `Ctrl+D` | **Delete** | 确认弹窗 (支持强制/删卷) |

### 查看详情

| 按键 | 操作 |
|---|---|
| `d` | **容器详情** (全屏 inspect) |
| `l` | **查看日志** (全屏日志视图) |
| `m` | **切换 Stats** (显示 CPU/MEM 列) |
| `Enter` | **进入容器 Shell** (exec 提示) |

### 标记与过滤

| 按键 | 操作 |
|---|---|
| `Space` | 标记/取消标记 (批量操作) |
| `/` | 搜索/过滤 |
| `r` | 刷新列表 |

## 数据增强

### ContainerSummary 扩展字段

```go
type ContainerSummary struct {
    ID      string
    Name    string
    Image   string
    State   string
    Status  string
    Created int64
    Ports   string
    Labels  map[string]string
    
    // 新增字段
    IPs     []string  // 从 NetworkSettings.Networks 提取
    Mounts  int       // Mounts 数量
    Size    string    // 容器磁盘占用
    
    ComposeProject string
    ComposeService string
}
```

### IP 获取

Docker `ContainerList` API 返回的 `types.Container` 已包含 `NetworkSettings`:
```json
{
  "NetworkSettings": {
    "Networks": {
      "bridge": { "IPAddress": "172.17.0.2" }
    }
  }
}
```

无需额外 API 调用，直接从列表响应提取。

### Mounts 获取

`types.Container` 包含 `Mounts` 数组:
```json
{
  "Mounts": [
    { "Source": "/data", "Destination": "/var/lib/postgresql/data" },
    { "Source": "/config", "Destination": "/etc/config" }
  ]
}
```

表格中显示数量 (如 `3v` = 3 个挂载)，详情中列出全部。
