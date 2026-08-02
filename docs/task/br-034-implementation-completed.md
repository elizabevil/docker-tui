# BR-034 实施完成报告(小模型 A + B,落盘时间 2026-08-01)

> 范围: A (runtime) + B (state + tables + view) 两个并发子任务的代码与文件变更
> 状态: 全部 7 个新文件已写、build 通过、测试通过
> 落盘清单:本文档 + 4 个分析文件更新 + 任务卡更新
> 主模型集成:已于 2026-08-01 完成并复核

## 最终集成结果

- 已接入 `ModeHistory`、Action Bar、异步加载、过期响应丢弃和顶层页面模板。
- 已实现 j/k、方向键、PgUp/PgDn、g/G、鼠标滚轮、Esc 返回及内联过滤。
- 已增加镜像引用与 layer 总数摘要、History 面包屑、footer/help 提示和三语言文案。
- 已补齐入口、过滤、鼠标、过期响应及 Docker 请求 ID 测试。
- `go test -short ./...`、`git diff --check` 通过；100 层 benchmark 约 `0.33 ms/op`。

## 1. 实施总览(主模型决策已采纳)

按最新主模型拍板:
- 入口:Images Action Bar 的 "Image History" 项触发(无键位)
- `HistoryState` 在 `AppModel.History`(顶级字段,不嵌 NavigationState)
- 运行时接口简化:`History(ctx, ImageSummary) ([]ImageHistoryLayer, error)` — manifest 由调用方 `summary.IsManifest` 判断,**adapter 不返回 `ImageHistorySource`**
- `HistoryLoadedMsg` 携带 `ImageID`;`handleHistoryLoaded` 首行判 ImageID,过期直接丢弃
- Cursor + ViewOffset 都维护,`MoveCursor` 后自动 `EnsureVisible`
- 复用公共 `component.RenderTable` + 新增 `tables/history.jsonc`,**只构造可见行**
- `history_keys.go` 改到 `internal/tui/keyboard/`(主模型在主路径做)
- Action Bar 路径:`internal/tui/actionbar/`(主模型在主路径做)

## 2. 改动文件清单(已落盘,build OK)

### A 小模型(runtime)

| 文件 | 改动 |
|---|---|
| `internal/data/runtime/image.go` | `ImageService` 接口加 `History(ctx, ImageSummary) ([]ImageHistoryLayer, error)` |
| `internal/data/runtime/docker/service_image.go` | 实现 `imageService.History` + 提取 `mapDockerHistory` 辅助函数 |
| `internal/data/runtime/docker/service_image_history_test.go` | 新建 3 个测试(`Empty` / `FieldPassthrough` / `ReturnsRuntimeType`) |
| `internal/data/runtime/podman/service_image.go` | 实现 `PodmanImageService.History` + 提取 `mapPodmanHistory` 辅助函数 |
| `internal/data/runtime/podman/service_image_history_test.go` | 新建 2 个测试(`Empty` / `FieldPassthrough`) |

### B 小模型(state + tables + view)

| 文件 | 改动 |
|---|---|
| `internal/tui/state/app.go` | `AppModel` 加 `History HistoryState` 顶级字段 |
| `internal/tui/state/history.go` | 新建 — `HistoryState` + `HistoryLoadedMsg` + Open/Apply/Close/MoveCursor/EnsureVisible/SetFilter |
| `internal/tui/state/history_test.go` | 新建 9 个测试(Open/Apply success/Apply error/MoveCursor wrap/EnsureVisible/SetFilter/Close 等) |
| `internal/tui/tables/history.jsonc` | 新建 — 5 列 profile(Layer / Created / Size / CreatedBy / Comment),CreatedBy 优先扩展 |
| `internal/tui/ui/pages/history/view.go` | 新建 — `RenderView(m, h, w) string` 走 `component.RenderTable` 公共 Flex Table |
| `internal/tui/ui/pages/history/view_test.go` | 新建 11 个测试(各状态 + 过滤 + 滚动 clamp 等) |

## 3. 关键实现决策

### 3.1 运行时接口简化

`ImageService.History(ctx, ImageSummary) ([]ImageHistoryLayer, error)` —
- 入参 `ImageSummary` 而非 `string`,让 adapter 直接拿 `summary.ID`,避免 runtime 内 splitImageRef 重复
- **不返回 `ImageHistorySource`**——manifest 在调用方判 `summary.IsManifest` 决定走 manifest 变体视图还是 history

### 3.2 helper 函数抽离

Docker 与 Podman adapter 都把 `mapXxxHistory` 抽离为可单测的纯函数(类似 Podman `MapImageSummaries` 模式),避免 mock `s.client.cli.ImageHistory` / `s.Client.REST.ImageHistory` 这种麻烦。

### 3.3 `HistoryState.Source` 字段保留

虽然 runtime 不再返回 `ImageHistorySource`,但 `state.HistoryState` 仍保留该字段(由 `handleHistoryLoaded` 自行设置 `Source = ImageHistoryLayerAPI`),UI 端 `renderStatus` 据此判断显示 manifest 变体提示或 layer 表。

### 3.4 Cursor wrap 公式

`MoveCursor` 用 `((c+delta)%total + total) % total` 兼容负数 delta(Go 的 `%` 对负数返负数)。

## 4. 验证

- `go build ./...` 通过
- `go test ./internal/tui/state/ -run TestHistory` 通过(9 个测试)
- `go test ./internal/tui/ui/pages/history/` 通过(11 个测试)
- `go test ./internal/data/runtime/{docker,podman}/` 通过(5 个新测试)

## 5. 文档落盘(分析文件更新)

为避免下次小模型误用旧路径,同步修正了 5 个分析文件 + 1 个早期设计文件 + 1 个 feature-design.md:

- `docs/task/br-034-image-history-top-page.md` — 重写,采纳 8 项修正(216 行)
- `docs/task/br-034-runtime-analysis.md` — 修正接口签名 / manifest 决策 / Action Bar 路径(233 行)
- `docs/task/br-034-ui-state-analysis.md` — 修正 `HistoryState` 字段位置 + HistoryLoadedMsg + EnsureVisible + 路径(304 行)
- `docs/task/br-033-advanced-container-ops-analysis.md` — Action Bar 路径改 `internal/tui/actionbar/`(235 行)
- `docs/task/br-039-a-implementation-design.md` — 所有 `ui/widget/actionbar` 改 `internal/tui/actionbar`(1002 行)
- `docs/task/br-039-code-location-analysis.md` — 路径修正(393 行)
- `docs/feature-design.md` — 路径修正(338 行)

## 6. 主模型集成 TODO(留给主模型)

按 A/B 文件无交叉原则,以下文件主模型处理:

| 集成点 | 文件 | 主模型任务 |
|---|---|---|
| `ModeHistory` 枚举 | `internal/tui/state/app.go` | 加 `ModeHistory` 在 AppMode 列表中 |
| `handleHistoryLoaded` 处理器 | `internal/tui/update/update_actions.go` 或新建 `update_handlers.go` | 接收 `HistoryLoadedMsg`,判 ImageID 过期,设 `m.History.Apply(layers, err)`;`Source` 字段设置(`ImageHistoryLayerAPI` / `ImageHistoryManifest` 来自 `summary.IsManifest` 在 `openHistoryPage` 处) |
| `openHistoryPage` 入口 | `internal/tui/keyboard/actions.go` | 加 `case ActionImageHistory`;校验 `summary.IsManifest`;开 dialog;发 `historyFetchCmd` |
| `historyFetchCmd` factory | `internal/tui/keyboard/` 或 `update/` | 创建 tea.Cmd,异步调 `Engine.Images().History(ctx, summary)`,返回 `HistoryLoadedMsg{ImageID, Layers, Err}` |
| `ModeHistory` dispatch | `internal/tui/keyboard/keyboard.go` | `dispatchByMode(ModeHistory)` 加 case,调 `handleHistoryKeys` |
| `handleHistoryKeys` 文件 | `internal/tui/keyboard/history_keys.go` | 调 `m.History.MoveCursor(±1, len(filtered), visible)` + Enter 关闭 + Esc 退出 + `/` 触发 filter + j/k 等 |
| `ActionImageHistory` 注册 | `internal/tui/keys/action.go` + `internal/tui/keys/registry.go` | KeyAction + 默认键位(暂不绑键,留 nil 或新模式) |
| Action Bar 项 | `internal/tui/actionbar/registry.go` | `imageActions(m)` 加 `{Key: "—", Label: "Image History", Action: keys.ActionImageHistory, Description: "View per-layer build history"}` |
| Page template | `internal/tui/ui/app/page_templates.go` | `templateFor(m)` + `projectPage` 加 `case ModeHistory: historyPageTemplate` + `history.RenderView(m, h, w)` |
| i18n 文案 | `internal/data/i18n/lang/{en,zh,ja}.jsonc` | `history.title` / `history.filter` / `history.layer` / `history.created` / `history.size` / `history.created_by` / `history.comment` |
| ActionImageHistory 集成 | `internal/tui/actionbar/registry_test.go` | 加 `ActionImageHistory` 项测试 |
| 测试/benchmark | `test/benchmarks/` | `BenchmarkRenderCachedHistory` 对齐 BR-016 |

## 7. 文件交叉验证

- A:仅动 `internal/data/runtime/{image.go, docker/, podman/}` — 与 B 无交叉
- B:仅动 `internal/tui/state/{app.go, history.go}` + `internal/tui/tables/history.jsonc` + `internal/tui/ui/pages/history/` — 与 A 无交叉
- A 改 `service_image.go` 同时需 `image.HistoryResponseItem` 类型 import,本轮已加
- B 改 `app.go` 加 `History` 字段,需主模型接着在主路径加 `ModeHistory` 枚举

**所有改动已落盘到文件,build / test 通过,无未提交内容。**

报告完成。
