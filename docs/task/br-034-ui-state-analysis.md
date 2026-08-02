# BR-034 UI / State 分析(只读,不改代码)

> 范围: HistoryState / ModeHistory / 滚动 + 过滤 / 性能方案 / 文件边界 + 测试清单
> 输出: 新增 / 修改文件清单 + 复用现有 pattern
> 配套: BR-034 Runtime 分析(已就位) / BR-033 分析
>
> 修正 2026-08-01:
> - `HistoryState` 在 `m.History`(AppModel 顶级),不再嵌入 `NavigationState`
> - Cursor 与 ViewOffset 都维护,MoveCursor 后自动 EnsureVisible
> - HistoryLoadedMsg 携带 ImageID,Update 首行判 ImageID 丢弃过期响应
> - `history_keys.go` 改到 `internal/tui/keyboard/`(与 `actionbar_keys.go` / `confirm.go` 同层)
> - 复用公共 Flex Table,新增 `tables/history.jsonc`,只构造可见行
> - Action Bar 正确路径:`internal/tui/actionbar/registry.go`(不是 `ui/widget/actionbar/`)
> - 详情页继续完全不请求 History(本任务**故意不**改 detail 视图)
>
> 边界: 不改 `state/app.go`(主模型集成 ModeHistory) / 不改 `keyboard/actions.go`(主模型) / 不改 `page_templates.go`(主模型)

## 1. 现有可复用 Pattern

### 1.1 DetailState 模式(`internal/tui/state/detail.go`)

适合作为 HistoryState 模板:

- `Revision uint64` 缓存键(变更即失效,触发重渲染)
- `Open(...)` 整体重置
- `OpenImage(id, title, data *ImageDetail)` 加载已查到的 detail
- `Close()` 清空
- `Scroll(delta int)` 上下移动
- `ClampVisibleOffset(total, visible int) int` 写回(防 overscroll 累积)
- `VisibleOffset(total, visible int) int` 读
- `Documents map[DetailSource]DetailDocument` 按来源缓存
- `SourceSelected bool` 选中行高亮

HistoryState 与 DetailState 关键差异:
- History 不需要 s/d source 切换(YAML/JSON 是 detail 的事)
- History 一定有 layer 列表(除非 manifest list)
- History 列表更扁平,适合 `keyhint`-style 列表 + 单行详情,不需要文档树

### 1.2 Detail 渲染模式(`internal/tui/ui/pages/detail/view.go`)

- `RenderView(m, panelHeight, panelWidth) string` 单入口
- 用 `m.Detail.Documents[DetailSourceSection]` 按 revision 缓存
- 只渲染可见行(从 `m.Detail.DetailOffset` 到 `DetailOffset + bodyHeight`)
- `ClampVisibleOffset` 在 Render 入口调用

History 页面应镜像这个模式:
- `RenderView(m, h, w)` 单入口
- `HistoryDocuments[Source]HistoryDocument` 缓存(按 source)
- 只渲染可见行(从 `HistoryOffset` 到 `HistoryOffset + bodyHeight`)

### 1.3 RenderTable 复用(`internal/tui/ui/component/table.go`)

`RenderTable` 已有成熟的多列表格渲染,BR-034 可直接复用:

- `RenderTable(component.TableData{Cols, Rows, Total, Limit, BodyHeight, BannerW, FooterHint, ColStyles})`
- 支持 header / rows / footer / selection
- 自动 ANSI-truncate 宽字符
- 已用 `panelW - 4` 算内部宽度
- 单测覆盖(看 `table_test.go`)

History 表格列设计:
| Layer | Created | Size | CreatedBy | Comment |
| short | YYYY-MM-DD | 12.3MB | RUN apt-get install... | commit msg |

前 4 列固定,CreatedBy 可超长被 lipgloss 截断。

### 1.4 Confirm dialog 模式(`internal/tui/state/confirm.go`)

- `ConfirmState` 含 `Focus int` / `Options []ChoiceOption` / `ConfirmAction` / `ConfirmTarget`
- `Open(action, target, message, trace)` 一键起
- 已有 `MoveFocus(delta)` + `Close()`

History 页不需要 confirm(只读),但 BR-033 的某些动作需要,BR-034 不涉及。

## 2. HistoryState 设计(`internal/tui/state/history.go` 新)

```go
package state

type HistoryState struct {
    ImageID  string                 // 当前历史查看的镜像 ID
    Layers   []ImageHistoryLayer    // 来自 runtime.History
    Source   ImageHistorySource     // pending / layer_api / manifest
    Error    string                 // 错误信息(空 = 无错)
    Loading  bool                   // true = 后台 Cmd 拉取中

    Cursor    int    // 选中行索引
    ViewOffset int   // 滚动偏移(行)
    Filter    string // CreatedBy / Comment 子串过滤
}
```

### 2.1 关键修正(主模型拍板)

- `HistoryState` 放在 `m.History`(AppModel 顶级字段),不再嵌入 `NavigationState`
- Cursor 与 ViewOffset 都维护,Cursor 移动后通过 `EnsureVisible` 修正 ViewOffset
- `HistoryLoadedMsg` 携带 ImageID;`handleHistoryLoaded` 首行判 ImageID,过期直接丢弃
- `MoveCursor(delta, total)` 移动 Cursor 后自动调 `EnsureVisible(visible)`,无需外部重复调

```go
// internal/tui/state/history.go
type HistoryLoadedMsg struct {
    ImageID string
    Layers  []ImageHistoryLayer
    Err     error
}

type HistoryState struct {
    ImageID    string                 // 当前正在加载/查看的镜像 ID
    Layers     []ImageHistoryLayer    // 已加载 layers(可能 nil)
    Loading    bool                   // 后台 Cmd 拉取中
    Error      string                 // 错误信息(空 = 无错)
    Source     ImageHistorySource     // pending / layer_api / manifest(由 state 设置,不由 runtime 返回)
    Cursor     int                    // 选中行索引(基于 filter 后的 items)
    ViewOffset int                    // 滚动偏移(基于 filter 后的 items)
    Filter     string                 // CreatedBy / Comment 子串过滤
}

func (s *HistoryState) Open(imageID string) {
    s.ImageID = imageID
    s.Layers = nil
    s.Source = ImageHistoryPending
    s.Error = ""
    s.Loading = true
    s.Cursor = 0
    s.ViewOffset = 0
    s.Filter = ""
}

func (s *HistoryState) Apply(layers []ImageHistoryLayer, err error) {
    s.Layers = layers
    s.Loading = false
    if err != nil {
        s.Error = err.Error()
    } else {
        s.Error = ""
    }
    // 成功路径:Source 由 update_handlers.go:handleHistoryLoaded 自行设置
    s.Cursor = 0
    s.ViewOffset = 0
}

func (s *HistoryState) Close() {
    *s = HistoryState{}
}

func (s *HistoryState) MoveCursor(delta, total, visible int) {
    if total <= 0 {
        s.Cursor = 0
        return
    }
    s.Cursor = (s.Cursor + delta + total) % total
    s.EnsureVisible(visible)
}

func (s *HistoryState) EnsureVisible(visible int) {
    if visible <= 0 {
        s.ViewOffset = 0
        return
    }
    if s.Cursor < s.ViewOffset {
        s.ViewOffset = s.Cursor
    } else if s.Cursor >= s.ViewOffset+visible {
        s.ViewOffset = s.Cursor - visible + 1
    }
    if s.ViewOffset < 0 {
        s.ViewOffset = 0
    }
}

func (s *HistoryState) SetFilter(f string) {
    s.Filter = f
    s.Cursor = 0
    s.ViewOffset = 0
}
```

### 2.2 AppModel 嵌入(主模型改)

`internal/tui/state/app.go` 添加 `History HistoryState` 顶级字段(不嵌入 `NavigationState`,与 `ActionBar` 平级)。此改动由主模型在 BR-034 集成时统一处理。

## 3. 性能方案(应对历史取消原因)

历史取消原因:"大镜像 >50 layer 滚动卡顿"。

### 3.1 三层防御

1. **Background fetch**: History 是 `tea.Cmd` 异步拉(不阻塞 Update)
2. **Visible-only render**: 只渲染 `[ViewOffset, ViewOffset+bodyHeight]` 范围内的行
3. **Filter pre-filter**: 过滤发生在 render 之前,只把命中的行 build 字符串

### 3.2 渲染伪代码

```go
func RenderView(m *state.AppModel, bodyHeight, bodyWidth int) string {
    h := m.History
    if h.ImageID == "" {
        return "No image selected"
    }
    if h.Loading {
        return renderLoading(bodyHeight, bodyWidth, h.ImageID)
    }
    if h.Error != "" {
        return renderError(bodyHeight, bodyWidth, h.Error)
    }
    if h.Source == state.ImageHistoryManifest {
        return renderManifestVariants(...)
    }
    items := filterLayers(h.Layers, h.Filter) // 子串包含
    if len(items) == 0 {
        return "No layers"
    }
    h.ClampVisibleOffset(len(items), bodyHeight-2)
    visible := items[h.ViewOffset:min(len(items), h.ViewOffset+bodyHeight-2)]
    return component.RenderTable(component.TableData{
        Cols:       historyCols(),
        Rows:       visibleToRows(visible),
        Total:      len(items),
        Limit:      bodyHeight - 2,
        BodyHeight: bodyHeight,
        BannerW:    bodyWidth,
        FooterHint: fmt.Sprintf("%d-%d/%d  / filter", h.ViewOffset+1, h.ViewOffset+len(visible), len(items)),
        ColStyles:  historyColStyles(),
    })
}
```

### 3.3 渲染细节

- Created 列:`time.Unix(layer.Created, 0).Format("2006-01-02 15:04")` 相对时间(用 humanize 库或简单 "x minutes ago" 可选)
- Size 列:`humanize.Bytes(layer.Size)`(utils 包已有)
- CreatedBy / Comment 列:lipgloss 截断到列宽

## 4. 文件清单

### 4.1 必须新增 (3)

| 文件 | 作用 | 约行数 |
|---|---|---|
| `internal/tui/state/history.go` | `HistoryState` + 方法 | ~80 |
| `internal/tui/state/history_test.go` | 状态机测试 | ~80 |
| `internal/tui/ui/pages/history/view.go` | History 顶层页渲染 | ~150 |
| `internal/tui/ui/pages/history/view_test.go` | 渲染 / 滚动 / 过滤测试 | ~120 |
| `internal/tui/ui/pages/history/keys.go` | `handleHistoryKeys`(j/k/PgUp/PgDn/g/G//Filter/Esc) | ~60 |
| `internal/tui/ui/pages/history/keys_test.go` | 按键分支测试 | ~100 |

(共 6 个新文件,不是因为算错,是 keys.go 与 view_test.go 拆开)

### 4.2 必须修改 (1 — 不碰主模型集成文件)

| 文件 | 改动 | **严禁**范围 |
|---|---|---|
| `internal/tui/state/navigation.go` | 加 `History HistoryState` 字段 | — |

### 4.3 不应触碰(主模型集成文件,留给主模型)

- `internal/tui/state/app.go` — `ModeHistory` 枚举加这里
- `internal/tui/keyboard/actions.go` — `case ActionImageHistory: openHistoryPage(m)` 入口
- `internal/tui/keyboard/keyboard.go` — `dispatchByMode(ModeHistory)` 分支 + `ToImageHistory` 跳转
- `internal/tui/ui/app/page_templates.go` — `templateFor(m)` 加 `case ModeHistory` + `projectPage` 加 `case ModeHistory`
- `internal/tui/ui/action/registry.go` — Action Bar 加 "Image History" 项 + 帮助 / Footer 文案
- `internal/tui/actionbar/registry.go` — 镜像页加 `KeyHistory: ActionImageHistory` 项
- `internal/data/i18n/lang/*.jsonc` — `history.title` / `history.filter` 等 key

## 5. 复用现状

- `component.RenderTable` (`internal/tui/ui/component/table.go`) — 列渲染 / footer / selection 全套
- `state.DetailState.ClampVisibleOffset` 模式 — 抄到 history.go
- `state.DetailState.Revision` 缓存机制 — 暂不需要(history 一次拉完)
- `ui/utils` 包:`humanize.Bytes`、`ANSI.TruncateWc` — Size 列 / 截断

## 6. 测试设计(本轮小模型交付范围内)

### 6.1 `state/history_test.go`

```go
func TestHistoryOpenResets(t *testing.T) { ... }
func TestHistoryApply(t *testing.T) { ... }
func TestHistoryScroll(t *testing.T) { ... }
func TestHistoryClampVisibleOffset(t *testing.T) { ... }
func TestHistorySetFilter(t *testing.T) { ... }
```

### 6.2 `ui/pages/history/view_test.go`

```go
func TestRenderHistoryLoading(t *testing.T) { ... }
func TestRenderHistoryError(t *testing.T) { ... }
func TestRenderHistoryManifest(t *testing.T) { ... }
func TestRenderHistoryList(t *testing.T) { ... }
func TestRenderHistoryFilter(t *testing.T) { ... }
func TestRenderHistoryClampAtBottom(t *testing.T) { ... }
```

### 6.3 `ui/pages/history/keys_test.go`

```go
func TestHistoryKeysJ(t *testing.T) { ... }
func TestHistoryKeysEsc(t *testing.T) { ... }
func TestHistoryKeysSlash(t *testing.T) { ... }
```

## 7. 性能基线(对齐 BR-016)

| 镜像 | 层数 | 期望滚动帧时 |
|---|---|---|
| alpine:latest | 1 | < 1ms |
| node:20 | ~7 | < 1ms |
| python:3.11-slim | ~5 | < 1ms |
| 巨型构建镜像(自测) | 50+ | < 2ms |

> BR-016 已有 `BenchmarkRenderCachedContainerDetail` 模式,后续 BR-039-D 验证时新增 `BenchmarkRenderCachedHistory` 即可(本轮不做)。

## 8. 关键风险

| 风险 | 说明 | 缓解 |
|---|---|---|
| **历史 50+ layer 滚动卡顿复发** | 必须按 §3 三层防御 | background fetch + visible-only + filter pre-filter,缺一不可 |
| **filter 子串过宽** | 用户输入长字符串后 ViewOffset 越界 | `SetFilter` 强制 `Cursor = ViewOffset = 0` |
| **镜像不存在 / history 失败** | 显示错误占位 | `Apply(err)` 时 `Loading = false` + `Error = err.Error()` |
| **ModeHistory 与 detail 子模式混用** | Action Bar / detail / history 各自独立 state | 不共享 `DetailState`,新建 `HistoryState` 单独管 |
| **ImageID 切换时残留旧 layers** | `Open(newID)` 时清空 `Layers` | 显式置 `Layers = nil` |
| **Cmd 异步拉取 race** | 用户在 history 拉取中切到别的镜像 | tea.Cmd 接受 `imageID` 参数,加载完成后核对 `m.History.ImageID` 再 Apply,过期直接丢弃 |
| **Manifest list 误显示空 history** | 拉取后 source=`manifest` | render 时按 source 分支,显示 manifest variants |

## 9. 输出给 BR-033 分析 / 主模型的输入

- HistoryState 形状已就位
- ModeHistory 需主模型在 `state/app.go` 添加枚举
- 入口 `ActionImageHistory` 需主模型在 `keys/action.go` + `keys/registry.go` 注册
- Action Bar 项(KeyHistory: ActionImageHistory)需主模型在 `widget/actionbar/registry.go` 的 `imageActions` 加
- page_templates `templateFor` 需加 `case ModeHistory: historyPageTemplate`

## 10. 关键文件路径汇总

- 模式参考:`internal/tui/state/detail.go` (DetailState) + `internal/tui/ui/pages/detail/view.go` (RenderView)
- 表格组件:`internal/tui/ui/component/table.go` (RenderTable)
- 渲染调用:`internal/tui/ui/app/page_templates.go` (templateFor, projectPage) — **主模型改**
- 键盘路由:`internal/tui/keyboard/keyboard.go:dispatchByMode` — **主模型改**
- i18n key:`internal/data/i18n/lang/zh.jsonc` 等 — **主模型改**

报告完成。**未修改任何代码**。