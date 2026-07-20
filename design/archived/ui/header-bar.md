# Header Bar — 顶部全局信息栏

## 状态

- [x] 已实现

## 触发

| 操作 | 行为 |
|---|---|
| 按 `H` / `F1` | 切换显示/隐藏 |
| 首次启动 | 显示 3 秒后自动隐藏 |
| 有 Toast 通知 | 自动展开显示 |
| 终端高度 < 15 行 | 强制隐藏 |

## 布局

```
┌──────────────────────────────────────────────────────────────────────────┐
│  dtui v0.2.0 │ engine:podman │ user:debi │ CPU: 12% MEM: 45% DISK: 2.3G/10G  │
└──────────────────────────────────────────────────────────────────────────┘
```

## 字段定义

| 字段 | 数据来源 | 颜色规则 |
|---|---|---|
| `dtui v0.2.0` | 编译时 version 常量 | **蓝色** (主题色) |
| `engine:podman` | `client.EngineType` | **绿色** |
| `user:debi` | `os.Getenv("USER")` | 默认前景色 |
| `CPU: N%` | `/proc/stat` 采样 | <50% 绿, 50-80% 黄, >80% 红 |
| `MEM: N%` | `/proc/meminfo` | 同上 |
| `DISK: used/total` | Docker `system df` API | <80% 绿, ≥80% 黄 |

## CPU/MEM 采样

```
间隔 2s 读取 /proc/stat (cpu line) 计算增量百分比
间隔 2s 读取 /proc/meminfo (MemTotal / MemAvailable)
```

## DISK 计算

```
Docker:  docker system df → Images + Containers + Volumes 总和
Podman:  podman system df → 同上
宿主机:  syscall.Statfs("/") → Total / Available
```

显示格式: `引擎占用/宿主机总量`
