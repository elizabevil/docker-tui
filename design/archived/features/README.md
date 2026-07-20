# dtui — Functional Design

功能规范、交互流程、键位映射。回答 "怎么操作"。

## 容器

| 文档 | 描述 |
|---|---|
| [container-table.md](container-table.md) | 容器表格 (9列+状态+IP+挂载+CPU/MEM) |
| [container-detail.md](container-detail.md) | 容器详情 (d键, 7分区折叠) |

## 镜像

| 文档 | 描述 |
|---|---|
| [image-table.md](image-table.md) | 镜像表格 (7列+REGISTRY+ARCH+内联展开) |
| [image-detail.md](image-detail.md) | 镜像详情 (d键, 6分区+镜像历史) |
| [image-containers.md](image-containers.md) | 内联容器子表 (→/Enter 展开) |

## 日志

| 文档 | 描述 |
|---|---|
| [log-view.md](log-view.md) | 全屏日志 (l键, 行号+时间戳+滚动) |

## 全局功能

| 文档 | 描述 |
|---|---|
| [connection-manager.md](connection-manager.md) | 环境选择器 (c键, 本地/远程切换) |

| 功能 | 分布在各文档中 |
|---|---|
| 搜索/过滤 | 容器表格、镜像表格 |
| 排序 | 容器表格(o)、镜像表格(o/O) |
| 批量操作 | 容器表格(Space标记+Ctrl+D) |
| 导出/调试 | 镜像表格(E导出/D调试) |
| 删除确认 | confirm-dialog (Ctrl+D, force选项) |
