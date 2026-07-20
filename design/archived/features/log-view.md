# Log View — 全屏日志视图

## 状态

- [x] 已实现

## 触发

- 容器面板按 `l`
- 替换主内容区域，全屏显示

## 布局

```
┌──────────────────────────────────────────────────────────────────┐
│  Logs: abc123def456                                              │
│  200 lines (auto-refresh every 2s)                               │
│                                                                  │
│    1 2024-01-15T10:00:00.123Z stdout  GET /index.html            │
│    2 2024-01-15T10:00:01.456Z stdout  200 OK                     │
│    3 2024-01-15T10:00:02.789Z stderr  Error: connection refused   │
│    4 2024-01-15T10:00:03.012Z stdout  POST /api/data             │
│    ...                                                           │
│                                                                  │
│  25-50/200 │ j/k scroll │ Esc back                               │
└──────────────────────────────────────────────────────────────────┘
```

## 样式

| 元素 | 样式 |
|---|---|
| 行号 (4位右对齐) | DimStyle (灰色) |
| 时间戳 (30字符) | 青色 (cyan) |
| stdout 文本 | 白色 |
| stderr 文本 | 红色 |
| Footer | DimStyle |

## 时间戳解析

```
Docker 格式:  2024-01-15T10:00:00.123456789Z → 提取前 30 字符
Podman 格式:  (同上，兼容)
无时间戳:     整行显示为文本
```

## 快捷键

| 按键 | 操作 |
|---|---|
| `j`/`↓` | 下滚 1 行 |
| `k`/`↑` | 上滚 1 行 |
| `PgDn` | 下滚 20 行 |
| `PgUp` | 上滚 20 行 |
| `g` | 跳到顶部 |
| `G` | 跳到底部 |
| `Esc` | 返回容器列表 |

## 数据流

```
容器面板按 l
  → doLogAction()
  → fetchLogBatch (since:1h, tail:200, follow:true)
  → LogBatchReceived → 追加到 LogContent (最多 500 行)
  → LogTick 每 2s 自动刷新
```
