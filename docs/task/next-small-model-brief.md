# 小模型下一步执行简报

> 更新日期: 2026-08-01
> 当前状态: BR-039 与 BR-040 已完成。

## 下一任务

优先分析 `BR-034 镜像 History 顶层页`。Action Bar 已提供无冲突动作入口,因此本轮不要预设新的单字母快捷键。

读取:

- `docs/task/br-034-image-history-top-page.md`
- `docs/task/br-039-action-bar-replace-q.md`
- `internal/tui/state/`
- `internal/tui/keyboard/image_action.go`
- `internal/tui/keyboard/actions.go`
- `internal/tui/ui/pages/images/`
- `internal/tui/actionbar/registry.go`
- Docker 与 Podman 的 image inspect/history runtime 接口

## 输出要求

只读分析,不要修改代码、不要提交、不要 push。报告写入 `docs/task/br-034-code-location-analysis.md`,内容必须包括:

1. Docker 与 Podman 当前是否已有 History 能力,返回数据结构是否统一。
2. History 应作为 images panel 子视图、独立 mode 还是独立 panel 的最小方案。
3. 从 Action Bar 的 `ActionImageHistory` 到 runtime 请求与页面渲染的完整链路。
4. 需要新增或修改的 state、keyboard、runtime、UI 与测试文件。
5. 大量 layer 数据下的滚动、截断和性能风险。
6. 与详情页“不再请求 History”决策的边界。

## 禁止范围

- 不恢复镜像详情页 History 请求或 History section。
- 不占用 `H`,避免与 Help/Header 相关任务冲突。
- 不修改 BR-039 Action Bar 的状态机与布局算法。
- 不顺带实现 BR-033、BR-035、BR-036 或 BR-037。
