# Toast Notification — 顶部操作反馈

## 状态

- [x] 已实现

## 位置

始终在 Header Bar 下方，Main Content 上方。

## 布局

```
┌──────────────────────────────────────────────────────────────────┐
│  ✓ Container started nginx-container (0.3s)                     │  ← 绿色，成功
│  ✕ Container stop failed: permission denied                      │  ← 红色，失败
└──────────────────────────────────────────────────────────────────┘
```

## 样式

| 类型 | 颜色 | 前缀 |
|---|---|---|
| 成功 | 绿色 (InfoStyle) | `✓` |
| 失败 | 红色 (ErrorStyle) | `✕` |

## 触发时机

| 操作 | Toast 内容 |
|---|---|
| 启动容器 | `✓ Container started nginx-container` |
| 停止容器 | `✓ Container stopped redis-cache` |
| 删除容器 | `✓ Deleted abc123 (0.3s)` |
| 强制删除 | `✓ Force removed abc123 (0.5s)` |
| 拉取镜像 | `✓ Image pulled nginx:latest (12.3s)` |
| 清理镜像 | `✓ pruned 1.2GB bytes reclaimed` |
| 切换排序 | `✓ Sort by name (ascending)` |
| 批量删除 | `✓ Deleted 3 items` |
| 操作失败 | `✕ start failed: permission denied` |

## 自动消失

```
显示 → ToastTick 每 100ms 递减 Timer (初始 30 = 3秒)
     → Timer 归零 → ToastMessage 清空 → 消失
```

## 扩展能力 (待实现)

- 显示操作耗时: `✓ Started nginx (0.3s)`
- 失败时显示错误详情
- 多个 Toast 排队显示
