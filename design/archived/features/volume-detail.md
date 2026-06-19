# Volume & Network Detail — 卷/网络详情

## 状态

- [ ] 待实现

## 触发

- 卷/网络面板选中行，按 **`d`** 键
- 全屏详情展示，与容器/镜像详情风格统一

## 卷详情

```
┌──────────────────────────────────────────────────────┐
│ Volume Detail: app-data                               │
│                                                      │
│ ── 基本信息 ──                                        │
│ Name:        app-data                                 │
│ Driver:      local                                    │
│ Mountpoint:  /var/lib/docker/volumes/app-data/_data   │
│ Scope:       local                                    │
│ Created:     2024-01-15 10:00:00 UTC                  │
│                                                      │
│ ── 使用情况 ──                                        │
│ Used by:     2 containers                             │
│  web-app     /usr/share/nginx/html (rw)               │
│  redis-cache /data (rw)                               │
│                                                      │
│ ── 标签 ──                                            │
│  com.docker.volume.description=Application data       │
│                                                      │
│ Esc: back                                              │
└──────────────────────────────────────────────────────┘
```

## 网络详情

```
┌──────────────────────────────────────────────────────┐
│ Network Detail: my-network                            │
│                                                      │
│ ── 基本信息 ──                                        │
│ Name:        my-network                               │
│ Driver:      bridge                                   │
│ Scope:       local                                    │
│ ID:          abc123def456...                          │
│                                                      │
│ ── 网络配置 ──                                        │
│ Subnet:      172.18.0.0/16                           │
│ Gateway:     172.18.0.1                              │
│ IP Range:    172.18.0.0/16                           │
│                                                      │
│ ── 容器列表 ──                                        │
│ Containers:  3                                        │
│  web-app     172.18.0.2                               │
│  redis       172.18.0.3                               │
│  postgres    172.18.0.4                               │
│                                                      │
│ ── 标签 ──                                            │
│  com.docker.network.driver.mtu=1500                   │
│                                                      │
│ Esc: back                                              │
└──────────────────────────────────────────────────────┘
```

## 数据来源

- 卷: `docker volume inspect <name>`
- 网络: `docker network inspect <id>`
- 使用该卷/网络的容器: 从 `m.Containers.Items` 客户端过滤
