# Confirm Dialog — 删除确认对话框

## 状态

- [x] 已实现
  /home/debi/IdeaProjects/docker-tui/design
## 设计原则

**删除操作统一使用 `Ctrl+D`**，涵盖所有资源类型（容器/镜像/卷/网络），弹出确认对话框，提供安全选项。

## 触发

| 场景 | 触发方式 | 对话框标题 |
|---|---|---|
| 删除单个容器 | `Ctrl+D` (光标在容器行) | `Delete Container` |
| 删除单个镜像 | `Ctrl+D` (光标在镜像行) | `Delete Image` |
| 删除卷 | `Ctrl+D` (光标在卷行) | `Delete Volume` |
| 删除网络 | `Ctrl+D` (光标在网络行) | `Delete Network` |
| 批量删除 | `Space` 标记多行后 `Ctrl+D` | `Bulk Delete` |

> 旧按键 `d` / `x` / `D` 全部废弃，统一为 `Ctrl+D`。

## 布局

### 容器删除

```
                    ┌──────────────────────────────────┐
                    │                                  │
                    │  ⚠ Delete Container              │
                    │                                  │
                    │  Name:  nginx-container           │
                    │  ID:    abc123def456              │
                    │  State: running                   │
                    │                                  │
                    │  Options:                         │
                    │  [✓] Force remove                 │
                    │  [ ] Remove volumes               │
                    │  [ ] Remove links                 │
                    │                                  │
                    │  [y] Confirm  [n] Cancel          │
                    └──────────────────────────────────┘
```

### 镜像删除

```
                    ┌──────────────────────────────────┐
                    │  ⚠ Delete Image                  │
                    │                                  │
                    │  Image: nginx:latest              │
                    │  ID:    sha256:aaa111bbb          │
                    │  Used by: 3 containers ⚡          │
                    │                                  │
                    │  Options:                         │
                    │  [✓] Force remove                 │
                    │  [ ] Prune unused parents         │
                    │                                  │
                    │  [y] Confirm  [n] Cancel          │
                    └──────────────────────────────────┘
```

### 批量删除 (已标记 3 项)

```
                    ┌──────────────────────────────────┐
                    │  ⚠ Bulk Delete (3 items)         │
                    │                                  │
                    │  ▸ nginx-container  (running)     │
                    │    redis-cache      (running)     │
                    │    stopped-app      (exited)      │
                    │                                  │
                    │  Options:                         │
                    │  [✓] Force remove (all)           │
                    │  [ ] Remove volumes               │
                    │                                  │
                    │  [y] Confirm  [n] Cancel          │
                    └──────────────────────────────────┘
```

## 删除选项

### 容器删除流程 (更新)

```
Ctrl+D → 检查容器状态
  ├── 已停止 (exited):  直接执行 docker rm
  ├── 运行中 (running): 先 docker stop → 成功后 docker rm（自动）
  └── 强制模式: docker rm -f（跳过 stop）

选项:
  [✓] Force remove (运行中容器先 stop + rm)
  [ ] Remove volumes (同时删除匿名卷)
```

### 镜像删除流程 (更新)

```
Ctrl+D → 检查是否有容器使用
  ├── 无: 直接执行 docker rmi
  └── 有: 显示警告 "Used by 3 containers: web-app, redis-cache, api-gw"
      提供 Force 选项 (docker rmi -f)
```

## 选项切换

| 按键 | 操作 |
|---|---|
| `Tab` / `↓` | 下一选项 |
| `↑` | 上一选项 |
| `Space` | 切换 `[✓]` ↔ `[ ]` |

## 确认/取消

| 按键 | 操作 |
|---|---|
| `y` / `Y` | 确认删除 |
| `n` / `N` / `Esc` | 取消 |
| `Ctrl+D` | 再次按取消 (防误触) |

## 确认后行为

1. 执行删除 API (带选项参数)，**异步后台执行** (不阻塞 UI)
2. 清除 `MarkedIDs`
3. Toast: `✓ Deleting nginx-container...` → 完成后更新为 `✓ Deleted nginx-container (1.2s)`
4. 刷新列表
5. 如果是当前查看日志/详情的容器 → 自动返回列表

> **后台删除**: 所有删除操作均通过 Bubbletea `tea.Cmd` 异步执行，删除期间 UI 保持响应，用户可继续浏览其他资源。删除完成后通过 Toast 通知结果。

## 样式

| 元素 | 样式 |
|---|---|
| 边框 | 红色圆角 (危险) |
| 标题 `⚠` | 红色 WarningStyle |
| 资源名 | 白色加粗 |
| 有容器使用的镜像 | 黄色 `⚡` 警告 |
| 选项标签 | DimStyle |
| 选中 `[✓]` | 青色 |
| 未选 `[ ]` | 灰色 |
