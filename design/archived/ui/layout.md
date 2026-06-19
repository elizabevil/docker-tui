# dtui UI Design — 布局总览

## 屏幕分层 (从上到下)

```
┌──────────────────────────────────────────────────────────────────┐  ← Header Bar (toggle H/F1)
│  dtui v0.2.0 │ engine:podman │ user:debi │ CPU:12% MEM:45%       │     默认隐藏
├──────────────────────────────────────────────────────────────────┤
│  ✓ Container started (0.3s)                                      │  ← Toast 通知 (临时)
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  Containers                                                │  │  ← 主表格 (全宽)
│  │                                                            │  │     d = 全屏详情
│  │  ID    NAME    IMAGE    STATE  IP        PORTS   M CPU MEM │  │     l = 全屏日志
│  │  ▸abc  nginx   nginx   run    172.17.0.2 80:80   2v 2.1 45│  │     / = 搜索过滤
│  │   def  redis   redis   run    172.17.0.3 6379    1v 1.5 32│  │
│  │                                                            │  │
│  │  1-2/6 │ 4 running, 1 exited │ Sort: name ▲                │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤  ← Status Bar (常驻)
│  ● podman │ 5 containers │ 3 marked                               │
├──────────────────────────────────────────────────────────────────┤  ← Shortcuts Bar (常驻)
│  j↓ Down │ k↑ Up │ Space Mark │ s Start │ S Stop │ d Detail ...   │
└──────────────────────────────────────────────────────────────────┘
```

### 全屏详情 (按 `d`)

```
┌──────────────────────────────────────────────────────────────────┐
│ Container Detail: nginx-container (abc123de)         [Esc: back] │
│                                                                  │
│  ── 基本信息 ──                                          [1 折叠] │
│  ID: abc123... │ Name: /nginx │ Image: nginx:latest              │
│  State: ● running │ Status: Up 3h │ Created: 2024-01-15          │
│                                                                  │
│  ── 网络信息 ──                                          [2 折叠] │
│  IP: 172.17.0.2 │ Ports: 80:80, 443:443                         │
│                                                                  │
│  ── 挂载卷 ──────────────────────────────────────────── [4 折叠] │
│  /data/nginx/html → /usr/share/nginx/html (rw)                   │
│                                                                  │
│  s:Start S:Stop R:Restart l:Logs Enter:Shell Esc:back            │
└──────────────────────────────────────────────────────────────────┘
```

## 响应式断点

| 终端宽度 | 模式 | 说明 |
|---|---|---|
| ≥ 120 列 | Wide | 表格列全部展开 |
| 80–119 列 | Normal | CPU/MEM 列自动隐藏 (按 m 开启) |
| < 80 列 | Compact | 部分列截断，优先保留 NAME/STATE |
| < 15 行 | Mini | 隐藏 Header，缩小间距 |

> **移除右侧详情面板**: 旧版 Dual/Triple 模式的右侧分栏已废弃。所有详情通过 `d` 全屏展示，简化 UI 并增加内容区域。

## 页面清单

| 页面 | 文件 | 描述 |
|---|---|---|
| 布局总览 | `layout-overview.md` | 本文件 |
| 顶部信息栏 | `header-bar.md` | 全局系统信息 |
| 容器表格 | `container-table.md` | 10 列容器表格 + 状态 + IP + 挂载 + CPU/MEM |
| 容器详情 | `container-detail.md` | 全屏容器 inspect (d 键) |
| 镜像表格 | `image-table.md` | 6 列镜像表格 + 排序 + 搜索 |
| 镜像详情 | `image-detail.md` | 镜像 Inspect 全字段 (d 键) |
| 镜像容器 | `image-containers.md` | 使用该镜像的容器列表 (Enter 键) |
| 日志全屏 | `log-view.md` | k9s 风格日志浏览器 (l 键) |
| 帮助页面 | `help-view.md` | dtui logo + 快捷键手册 |
| 确认对话框 | `confirm-dialog.md` | Ctrl+D 删除确认弹窗 |
| Toast 通知 | `toast.md` | 顶部操作反馈横幅 |
| 状态栏 | `status-bar.md` | 底部连接/容器状态 |
| 快捷键栏 | `shortcuts-bar.md` | 底部快捷键提示 |
| 主题系统 | `themes.md` | 6 套皮肤的配色定义 |
| ~~右侧详情~~ | ~~`detail-panel.md`~~ | **已废弃** → 改为全屏详情 |
| 卷/网络表格 | `table-view.md` | 卷和网络列表 |
