# 镜像 History 顶层页

> 任务卡 / BR-034 / 历史页
> 修正 2026-08-01:
> - 最终入口为 **Images 页 H 键或 Action Bar 触发**；H 不再切换 Header
> - `HistoryState` 改为 `AppModel.History`(顶级字段),不再嵌入 `NavigationState`
> - `history_keys.go` 改到 `internal/tui/keyboard/`(与现有 `actionbar_keys.go` / `confirm.go` 同层)
> - 复用公共 `tables` 配置(新增 `tables/history.jsonc`)+ `EnsureVisible` 滚动模型
> - HistoryLoadedMsg 携带 ImageID,Update 时丢弃过期响应

## 元信息

- **关联编号**:BR-034
- **优先级**:high
- **状态**:`completed`
- **依赖**:BR-007(revision 缓存)、BR-016(ClampVisibleOffset)、BR-039(Action Bar 入口)
- **关联设计**:[docs/feature-design.md §5.1](../feature-design.md)、[bugfix-requirements.md BR-034](../bugfix-requirements.md)
- **历史**:曾因性能取消;现已通过可见行切片解决性能问题,H 在 Images 页恢复为 History

## 目标

在**镜像页**通过 H 键或 Action Bar 的 "Image History" 项触发,进入**独立顶层页** `ModeHistory`,浏览当前选中镜像的 layer history。解决历史性能问题(大镜像 >50 layer 滚动卡顿)。

## 代码结构索引

### 必须读懂的文件

| 文件 | 作用 |
|---|---|
| `internal/data/runtime/image.go` | `ImageHistoryLayer` 数据结构 + `ImageService` 接口当前只有 List/Inspect,**需加 History(ctx, ImageSummary)** |
| `internal/data/runtime/docker/images.go:107-173` | `InspectImageDetailContext` history 显式不拉,需新增 `ImageHistory(ctx, id)` |
| `internal/data/runtime/podman/service_image.go` | Podman adapter 显式传 `history=nil`;需新增 `History` 方法 |
| `internal/driver/podman/images_read.go:22` | `RESTClient.ImageHistory(ctx, id)` 已实现,可直接调 |
| `internal/tui/state/detail.go` | 现有 `DetailState`(`Revision` / `ClampVisibleOffset` / `Documents map` 等)作为 HistoryState 设计参考 |
| `internal/tui/ui/component/table.go` | `RenderTable` 公共 Flex Table 组件,直接复用 |
| `internal/tui/tables/` | 公共表格 profile 存放位置;新增 `tables/history.jsonc` |
| `internal/tui/actionbar/registry.go` | Action Bar 主路径(已由 BR-039 落地);镜像页 Action Bar 加 "Image History" 项 |
| `internal/tui/keyboard/actionbar_keys.go` | 同包 — `history_keys.go` 放在此处,处理 ModeHistory 内按键 |
| `internal/data/i18n/lang/*.jsonc` | 加 `history.title` / `history.layer` / `history.created_by` / `history.size` / `history.comment` |

### 必须修改的文件

| 文件 | 改动 |
|---|---|
| `internal/data/runtime/image.go` | `ImageService` 接口加 `History(ctx, ImageSummary) ([]ImageHistoryLayer, error)`(无 ImageHistorySource 返回,manifest 由 ImageSummary.IsManifest 在调用前判断) |
| `internal/data/runtime/docker/service_image.go` | 加 `History` 方法实现 — 调 `s.client.cli.ImageHistory(ctx, summary.ID)` 转换 `image.HistoryResponseItem` → `runtimeapi.ImageHistoryLayer` |
| `internal/data/runtime/podman/service_image.go` | 加 `History` 方法实现 — 调 `s.Client.REST.ImageHistory(ctx, summary.ID)` 转换 `dto.LayerHistoryEntry` → `runtimeapi.ImageHistoryLayer` |
| `internal/tui/state/app.go` | `AppModel` 加 `History HistoryState` 顶级字段(不嵌 NavigationState) |
| `internal/tui/state/history.go`(新) | `HistoryState{ImageID, Layers, Loading, Error, Cursor, ViewOffset, Filter, Source}` + `Open(id)` / `Close()` / `EnsureVisible(visible)` / `MoveCursor(delta)` / `SetFilter()` |
| `internal/tui/tables/history.jsonc`(新) | History 表格 profile(列宽 / 列名 i18n / 边界) |
| `internal/tui/ui/pages/history/view.go`(新) | `RenderView` 单入口;用 `tables/history.jsonc` 拉列定义;只构造 `ViewOffset..ViewOffset+BodyHeight` 可见行;调用 `m.History.EnsureVisible(visible)` |
| `internal/tui/actionbar/registry.go` | 镜像页 Action Bar 加 `ActionImageHistory / Label: "Image History"`;键盘注册 H |
| `internal/tui/ui/action/registry.go` | 加 ActionImageHistory 的 Help / Footer 文案 |

### 必须新增的文件

| 文件 | 作用 |
|---|---|
| `internal/tui/state/history.go` | `HistoryState` + 方法 + `HistoryLoadedMsg{ImageID, Layers, Err}` |
| `internal/tui/state/history_test.go` | 状态机 + EnsureVisible 边界测试 |
| `internal/tui/ui/pages/history/view.go` | 顶层页渲染(复用 RenderTable + 可见行切片) |
| `internal/tui/ui/pages/history/view_test.go` | 渲染快照 / 可见行切片 / 滚动 |
| `internal/tui/keyboard/history_keys.go` | `handleHistoryKeys`(ModeHistory 内 j/k/PgUp/PgDn/g/G//Filter/Esc) |
| `internal/tui/keyboard/history_keys_test.go` | 按键分支测试 |
| `internal/tui/tables/history.jsonc` | 列 profile |
| `test/benchmarks/history_bench_test.go` | 大镜像(>50 layer)滚动 benchmark(对齐 BR-016) |

### 不应触碰的文件

- `internal/data/runtime/containers.go`(与 history 无关)
- `internal/tui/ui/pages/detail/`(详情页继续完全不请求 history,本任务**故意不**改 detail 视图)
- 共享集成文件(由主模型在 BR-034 实施时统一处理):`internal/tui/state/app.go`(`ModeHistory` 枚举加这里)、`internal/tui/keyboard/actions.go`、`internal/tui/keyboard/keyboard.go`、`internal/tui/ui/app/page_templates.go`、`internal/tui/ui/app/mouse.go`、`internal/tui/ui/widget/footer/footer.go`

## 操作链路(经 Action Bar)

```text
User 在镜像页 → H 键或 Action Bar 触发 "Image History" 项
  → keyboard/actions.go:case ActionImageHistory: openHistoryPage(m)
      校验 m.Navigation.ActivePanel == PanelImages
      校验 ctr := m.Resources.Images.Selected()(无选 → toast)
      校验 !ctr.IsManifest(manifest list 改弹 manifest 变体视图而非 history)
      m.History.Open(ctr.ID)
          ImageID = ctr.ID
          Layers = nil
          Loading = true
          Error = ""
          Cursor = 0
          ViewOffset = 0
          Filter = ""
          Source = ImageHistoryPending
      m.Navigation.Mode = state.ModeHistory
  → tea.Cmd: historyFetchCmd(m.Connection.Engine, ctr) → HistoryLoadedMsg{ImageID, Layers, Err}
  → update.go: handleHistoryLoaded(m, msg)
      IF m.History.ImageID != msg.ImageID:  // 过期响应,丢弃
          return m, nil
      m.History.Loading = false
      m.History.Layers = msg.Layers
      m.History.Error = ""
      IF msg.Err != nil: m.History.Error = msg.Err.Error()
      m.History.Cursor = 0
      m.History.ViewOffset = 0
      m.History.EnsureVisible(bodyHeight - 2)  // clamp cursor + viewoffset
  → ui/pages/history/view.go:RenderView(m, h, w)
      1. 校验 m.History.Loading → 渲染 loading 占位
      2. 校验 m.History.Error != "" → 渲染错误占位
      3. 校验 m.History.ImageID == "" → 渲染 "No image"
      4. items := filterLayers(m.History.Layers, m.History.Filter)
      5. m.History.EnsureVisible(bodyHeight - 2)
      6. visible := items[m.History.ViewOffset : min(len(items), m.History.ViewOffset+bodyHeight-2)]
      7. rows := layersToRows(visible)
      8. component.RenderTable(tables.HistoryData{Cols, Rows, Total: len(items), Limit, BodyHeight, BannerW: w, FooterHint, ColStyles})
  → keyboard/history_keys.go:handleHistoryKeys(rawKey, m)
      j / down     → MoveCursor(+1, len(filtered)) + EnsureVisible(visible)
      k / up       → MoveCursor(-1, len(filtered)) + EnsureVisible(visible)
      PgDn / PgUp  → MoveCursor(±pageStep) + EnsureVisible
      g / G        → MoveCursor to top/bottom
      /            → ToHistoryFilter(m)(可考虑打开 ModeHistoryFilter 子模式,本轮先简化为 input 累积到 Filter)
      Esc          → Close History + ModeHistory → ModeImages
```

## State 设计(关键差异 vs 旧版)

```go
// internal/tui/state/history.go
type HistoryState struct {
    ImageID    string                 // 当前正在加载/查看的镜像 ID
    Layers     []ImageHistoryLayer    // 已加载 layers(可能 nil)
    Loading    bool                   // 后台 Cmd 拉取中
    Error      string                 // 错误信息(空 = 无错)

    Source     ImageHistorySource     // pending / layer_api / manifest
    Cursor     int                    // 选中行索引(基于 filter 后的 items)
    ViewOffset int                    // 滚动偏移(基于 filter 后的 items)
    Filter     string                 // CreatedBy / Comment 子串过滤
}
```

注意:
- `Source` 状态在 `Apply` 时被设置(pending → layer_api / manifest),不进 History API 返回值
- `Cursor` 和 `ViewOffset` 各自维护,均基于 filter 后的 `items` 索引(不是原始 `Layers`)
- `MoveCursor` 后必须调 `EnsureVisible` 修正 ViewOffset

## 验收标准

- [x] 镜像页 H 键及 Action Bar 均可进入 History 顶层页;H 不再隐藏 Header
- [x] 行数等于镜像 layer 数;manifest list 禁用 History 入口
- [x] 每行展示:`Created`(相对时间)/ `Size`(人类可读)/ `CreatedBy`(摘要)/ `Comment`
- [x] 选中镜像 > 50 layer,上下滚动 / PgDn / j 翻页无卡顿
- [x] 顶部固定显示镜像 `Repository:Tag` 与总 layer 数
- [x] `/` 触发内联 Filter,输入字符串后 CreatedBy / Comment 模糊匹配,行数实时减少
- [x] Cursor 移动后 ViewOffset 自动调整,Cursor 始终可见
- [x] `Esc` 返回镜像页,`m.History` reset,`m.History.ImageID == ""`
- [x] 详情页**完全不**请求 History(与 BR-008 决策一致)
- [x] 大镜像 benchmark:`BenchmarkRenderCachedHistory` 通过
- [x] 镜像不存在 / history 失败:显示明确错误占位
- [x] HistoryLoadedMsg 携带 ImageID,过期响应被丢弃
- [x] manifest 列表不调 History API,Action Bar 项禁用
- [x] Help 页增加 "Action Bar → Image History" 提示

## 风险

| 风险 | 说明 | 缓解 |
|---|---|---|
| **H 键历史性能问题复现** | 大镜像 >50 layer 滚动卡顿 | `EnsureVisible` 钳制 + `RenderTable` 只构造可见行(对齐 BR-016) |
| **HistoryLoadedMsg 过期** | 用户在拉取中切到别的镜像,后到响应会被错误 Apply | `handleHistoryLoaded` 首行 `if m.History.ImageID != msg.ImageID: return` |
| **manifest 误拉 history** | manifest list 拉 history 浪费 + 显示空 | 调用前 `if ctr.IsManifest: 走 manifest 变体视图, 不调 History` |
| **Filter 越界** | filter 缩小 items 后 Cursor 越界 | `SetFilter` 强制 Cursor = ViewOffset = 0;`MoveCursor` 后 `EnsureVisible` |
| **与 detail 子模式混淆** | 详情页(DetailState) vs History 页(HistoryState) | 完全独立 state,共享 `ImageDetail` 数据但独立 `HistoryState`;两 mode 互不切换 |
| **i18n 漏 key** | `history.title` / `history.layer` / `history.created_by` 等 | i18n key 集中加在 `internal/data/i18n/lang/{en,zh,ja}.jsonc` |

## 建议任务分解

按主模型最新划分:

- **小模型 A**(`runtime` + adapter 测试):
  1. `internal/data/runtime/image.go` — 加 `History(ctx, ImageSummary)` 接口
  2. `internal/data/runtime/docker/service_image.go` — 实现 History
  3. `internal/data/runtime/podman/service_image.go` — 实现 History
  4. `internal/data/runtime/mockengine/` — 加 mock
  5. 2 个 adapter 测试(`*_history_test.go`)
- **小模型 B**(`state` + `ui/pages/history` + `tables` + `keyboard`):
  1. `internal/tui/state/history.go` — `HistoryState` + `HistoryLoadedMsg`
  2. `internal/tui/state/history_test.go`
  3. `internal/tui/tables/history.jsonc`
  4. `internal/tui/ui/pages/history/view.go` — `RenderView` 复用 RenderTable + 可见行切片
  5. `internal/tui/ui/pages/history/view_test.go` — 渲染快照
  6. `internal/tui/keyboard/history_keys.go` — `handleHistoryKeys`
  7. `internal/tui/keyboard/history_keys_test.go`
- **主模型**(集成):
  1. `internal/tui/state/app.go` — 加 `ModeHistory` 枚举
  2. `internal/tui/state/navigation.go` — 不动(History 在 AppModel 顶层)
  3. `internal/tui/keyboard/actions.go` — `case ActionImageHistory: openHistoryPage(m)`
  4. `internal/tui/keyboard/keyboard.go` — `dispatchByMode(ModeHistory)` 分支
  5. `internal/tui/ui/app/page_templates.go` — `ModeHistory → historyPageTemplate`
  6. `internal/tui/ui/app/mouse.go` — 加 `mode = ModeHistory` 时 Get() panel body(若 History 占整 panel)
  7. `internal/tui/ui/widget/footer/footer.go` — 改 Footer 模式判定加 `ModeHistory`
  8. `internal/tui/actionbar/registry.go` — 镜像页 Action Bar 加 "Image History" 项(主模型 B 协作)
  9. `internal/tui/ui/action/registry.go` — Help / Footer 文案
  10. `internal/data/i18n/lang/{en,zh,ja}.jsonc` — `history.*` key

**A 与 B 文件无交叉**(`runtime` 与 `state/ui` 隔离),可安全并发。

## 待确认项

- [ ] **manifest 列表的处理**:在 Action Bar 触发时直接走 manifest 变体视图,还是弹一个"切换到 manifest 变体"的二级菜单?建议:在 Action Bar 里把"Image History"项对 manifest 标 `Disabled=true`,加 tooltip 说明(变体视图由 manifest 详情页承担)。
- [ ] **`/` Filter 模式**:独立子模式(`ModeHistoryFilter`)还是 inline 累积到 `Filter` 字段?建议:inline(简单,BR-039 Action Bar 的 `Filtering bool` 模式可参考)
- [ ] **History 与 Detail 同开**:用户在 History 页能否按 `s` 切到 source / yaml / json?建议:History 页**不**有 source 切换(没有 raw source 数据)
- [ ] **HistoryLoadedMsg 消息类型**:放 `internal/tui/state/history.go` 还是 `update.go`?建议:放 `history.go`,与 `HistoryState` 同包,Update 引用

## 参考

- [docs/feature-design.md §5.1 镜像 History 顶层页](../feature-design.md)
- [docs/bugfix-requirements.md BR-008](../bugfix-requirements.md)(H 键路由历史决策)
- [docs/bugfix-requirements.md BR-016](../bugfix-requirements.md)(ClampVisibleOffset 机制)
- [docs/bugfix-requirements.md BR-007](../bugfix-requirements.md)(按 revision 缓存文档)
- BR-039 实施 — Action Bar 入口已就位
