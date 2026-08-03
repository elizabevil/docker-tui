# R02-01 镜像 History 顶层页

## 元信息

- 状态: implementing
- 优先级: high
- 来源: [../../task/br-034-image-history-top-page.md](../../task/br-034-image-history-top-page.md)、[../../task/br-034-implementation-completed.md](../../task/br-034-implementation-completed.md)
- 关联任务:
  - [../../task/br-034-runtime-analysis.md](../../task/br-034-runtime-analysis.md)
  - [../../task/br-034-ui-state-analysis.md](../../task/br-034-ui-state-analysis.md)
- 关联约束:
  - [../../constraint/C03-table.md](../../constraint/C03-table.md)
  - [../../constraint/C06-i18n.md](../../constraint/C06-i18n.md)

## 目标

镜像页提供独立 History 顶层页,浏览当前选中镜像的 layer 历史;解决大镜像 > 50 layer 的滚动性能问题。

## 用户流程

镜像页选中镜像 → `H` 键 或 Action Bar "Image History" → 进入 `ModeHistory` 顶层页 → 浏览 layer 列表 → `Esc` 返回镜像页。

## UI/UX

- 复用公共 `tables` 配置(新增 `tables/history.jsonc`)
- 列:Layer / CreatedBy / Size / Comment
- 长列通过 [C03-table](../../constraint/C03-table.md) 截断 + tooltip
- 仅渲染 `ViewOffset..ViewOffset+BodyHeight` 可见行,`EnsureVisible` 保证 cursor 可见

## 功能规则

- 入口:H 键(Images 页)或 Action Bar
- 仅对非 manifest 镜像加载 history(manifest 走 Inspect 详情)
- `HistoryLoadedMsg` 携带 `ImageID`,update 时丢弃过期响应
- 大 layer 数滚动平滑,首屏渲染 < 100ms
- 顶部状态栏显示镜像 ID + 总 layer 数

## 实现设计

- 接口:`internal/data/runtime/image.go:ImageService.History(ctx, ImageSummary)`
- Docker adapter:调 `cli.ImageHistory` 转换 `image.HistoryResponseItem`
- Podman adapter:调 `s.Client.REST.ImageHistory` 转换 `dto.LayerHistoryEntry`
- 状态:`internal/tui/state/history.go` 新增 `HistoryState`(`AppModel` 顶级字段,非嵌入 NavigationState)
- 视图:`internal/tui/ui/pages/history/view.go` 单入口,只构造可见行
- 表格 profile:`internal/tui/tables/history.jsonc`
- Action Bar 注册:`Image History` 项 + H 键绑定

## 验收标准

- `H` 键从 Images 页打开 History
- 大镜像(>50 layer)滚动无明显卡顿
- 镜像切换后旧 history 响应被丢弃
- `Esc` 返回 Images 页,资源状态保留

## 非目标

- manifest 镜像的 layer 浏览
- 多镜像并发 history

## 迁移记录

- 旧文档:[task/br-034-image-history-top-page.md](../../task/br-034-image-history-top-page.md) 等 4 个 br-034 文件
- 保留信息:H 入口、HistoryState 顶级字段设计、EnsureVisible 滚动模型
- 待确认状态:`implementing` 表示核心完成,但需双 runtime 端到端验证;主模型确认后调整。