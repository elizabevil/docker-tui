# Image Table View — 镜像表格 (v2)

## 状态

- [x] 已实现

## 列定义 (7列)

| 列 | 宽度 | 排序 | 搜索 | 说明 |
|---|---|---|---|---|
| **→** | 2 | — | — | 展开指示器 (有容器时 `▸`) |
| **REGISTRY** | 16 | — | — | 仓库域名，超长截断 `swr.cn-nort...` |
| **NAME** | 22 | ✅ | ✅ | `library/nginx` 或 `ddn-k8s/.../postgres` |
| **TAG** | 12 | — | ✅ | `latest`, `18.4-alpine` |
| **IMAGE ID** | 16 | ✅ | ✅ | `sha256:aaa111bb` |
| **ARCH** | 6 | — | — | `amd64`, `arm64` (来自 Inspect) |
| **CREATED** | 16 | ✅ | — | `01-15 10:00` 月-日 时:分 |
| **SIZE** | 9 | ✅ | — | `187.0 MB` |

## REGISTRY 提取规则

```
完整 RepoTag                                → REGISTRY         NAME                  TAG
────────────────────────────────────────────────────────────────────────────────────────
nginx:latest                                → docker.io        library/nginx         latest
redis:7-alpine                              → docker.io        library/redis         7-alpine
swr.cn-north-4.myhuaweicloud.com/ddn-k8s/   → swr.cn-nort...   ddn-k8s/.../postgres  18.4-alpine
  docker.io/postgres:18.4-alpine
<none>:<none>                               → —                <none>                <none>
```

> REGISTRY 超过 16 字符截断为 13 + `...`

## 布局

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│ Images                                                                           │
│                                                                                  │
│   REGISTRY      NAME            TAG         IMAGE ID        ARCH  CREATED   SIZE │
│   docker.io     library/nginx   latest      sha256:aaa111bb amd64 01-15 10  187MB│
│ ▸ docker.io     library/redis   7-alpine    sha256:bbb222cc amd64 01-14 08   32MB│
│   │ CONTAINER ID  NAME         STATE   CREATED     PORTS                         │
│   │ abc123def     redis-cache  running 01-14 08:30 6379→6379                     │
│   │ def456abc     redis-queue  running 01-10 06:00 6380→6379                     │
│   │ ─────────────────────────────────────────────────────                        │
│   │ 2 containers │ s:Start S:Stop l:Logs Ctrl+D:Del                              │
│   docker.io     library/postgres 16       sha256:ccc333dd amd64 01-16 14  420MB  │
│   swr.cn-nort.. ddn-k8s/.../p.. 18.4-alp sha256:ddd444ee amd64 01-10 06  7.2MB  │
│                                                                                  │
│  1-4/4 │ 4 images │ Sort: name ▲                                                 │
└──────────────────────────────────────────────────────────────────────────────────┘
```

## 内联展开 (USED BY)

| 按键 | 操作 |
|---|---|
| `→` / `Enter` | 展开当前行，显示容器子表 |
| `←` / `Esc` | 折叠子表 |
| 子表内 `↑↓` | 在容器行间移动光标 |
| 子表内 `s`/`S`/`l`/`Ctrl+D` | 操作选中容器 |

无需额外 API，直接从已加载的容器列表过滤匹配。

## 排序

| 按键 | 排序 |
|---|---|
| `o1` | NAME 字母序 |
| `o2` | CREATED 时间 |
| `o3` | SIZE 数值 |
| `O` | 反转方向 |

## 搜索

`/` 匹配: NAME, TAG, IMAGE ID

## 键盘交互

| 按键 | 操作 |
|---|---|
| `→`/`Enter` | 展开 USED BY 子表 |
| `←`/`Esc` | 折叠 / 取消搜索 |
| `P` | 拉取镜像 |
| `p` | 清理悬空镜像 |
| `d` | **镜像详情** (全屏 inspect) |
| `o`/`O` | 排序 / 反转 |
| `Ctrl+D` | 删除 (确认弹窗) |
| `Space` | 标记 |
| `/` | 搜索 |

## v1 → v2 变更

| 列 | v1 | v2 |
|---|---|---|
| DIGEST | 显示(19列宽) | **移除** |
| REGISTRY | — | **新增** (仓库来源) |
| ARCH | — | **新增** (CPU架构) |
| USED BY | Enter→全屏页 | **→ 内联展开** |
