# Image Containers View — 使用该镜像的容器 (v2)

## 状态

- [x] 已实现

## v2 变更

| 方面 | v1 (旧) | v2 (新) |
|---|---|---|
| 触发 | `Enter` 跳转全屏页 | `→` / `Enter` 内联展开 |
| 显示位置 | 独立全屏页面 | 镜像表格行下方子表 |
| 上下文 | 离开镜像表格 | 保持在镜像表格中 |
| 容器操作 | 完整支持 | 完整支持 (`s`/`S`/`l`/`Ctrl+D`) |

## 内联子表布局

```
 ▸ docker.io  library/redis  7-alpine  sha256:bbb222  amd64  01-14 08:30  32 MB
   │ CONTAINER ID  NAME          STATE    CREATED      PORTS
   │ abc123def     redis-cache   running  2024-01-14   6379→6379
   │ def456abc     redis-queue   running  2024-01-10   6380→6379
   │ ─────────────────────────────────────────────────────────
   │ 2 containers │ s:Start S:Stop l:Logs Ctrl+D:Del
```

## 快捷键

| 按键 | 操作 |
|---|---|
| `→` / `Enter` | 展开子表 |
| `←` / `Esc` | 折叠子表 |
| `↑↓` | 在容器行间移动光标 |
| `s` / `S` | Start / Stop 选中容器 |
| `l` | 查看容器日志 |
| `Ctrl+D` | 删除容器 |

## 数据获取

客户端过滤: 从已加载的 `m.Containers.Items` 中匹配 Image 字段，无需额外 API 调用。
