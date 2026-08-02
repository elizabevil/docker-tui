# BR-034 Runtime 分析(只读,不改代码)

> 范围: Docker / Podman ImageHistory API 对接 + 统一 ImageService.History 接口设计
> 输出: 适配层修改清单 + 测试清单
> 状态: 2026-08-01 小模型只读分析
> 配套: BR-034 UI-State 分析 / BR-033 分析

## 1. 当前实现盘点

### 1.1 `internal/data/runtime/image.go` (共享 model)

- **`ImageHistoryLayer` 已定义**(line 57-63):
  ```go
  type ImageHistoryLayer struct {
      ID        string
      Created   int64
      CreatedBy string
      Size      int64
      Comment   string
  }
  ```
- **`ImageDetail` 已含 `History` 字段 + `HistorySource` + `HistoryError`**(line 105-107)。`ImageHistorySource` 枚举含 `ImageHistoryLayerAPI` / `ImageHistoryManifest` / `ImageHistoryPending`。
- **`ImageService` 接口只有 2 个方法**(line 118-121):
  ```go
  type ImageService interface {
      List(context.Context, ImageListOptions) ([]ImageSummary, error)
      Inspect(context.Context, ImageSummary) (*ImageDetail, error)
  }
  ```
  **没有 `History` 方法**。

### 1.2 Docker adapter (`internal/data/runtime/docker/`)

- `InspectImageDetailContext`(images.go:107-173)调 `c.cli.ImageInspect(...)` 填充 `detail`。
- **关键发现**(images.go:170-172):
  ```go
  // History is intentionally excluded from detail loading. Restore it as a
  // separate, on-demand request when the dedicated history view is added.
  return detail, nil
  ```
  这是**预存在的、有意为之的决策**——detail 不带 history,留给独立的 on-demand 请求。
- `MapImageInspect(detail, history)`(images.go:299+) 接受 history 作为第二个参数。**当前调用站点(service_image.go:24)没传 history**——因为没有。
- **`cli.ImageHistory` 端点未使用**——需要新增。

### 1.3 Podman adapter (`internal/data/runtime/podman/`)

- `PodmanImageService.Inspect`(service_image.go:33-41)显式排除 history:
  ```go
  // History is intentionally excluded from detail loading. Restore it as a
  // separate, on-demand request when the dedicated history view is added.
  return MapImageInspect(*inspect, nil), nil
  ```
- **`RESTClient.ImageHistory(ctx, id)` 已存在**(images_read.go:22-30):
  ```go
  func (c *RESTClient) ImageHistory(ctx context.Context, id string) ([]dto.LayerHistoryEntry, error) {
      var raw []dto.LayerHistoryEntry
      path := "/libpod/images/" + escape(id) + "/history"
      // ...
  }
  ```
  **这是现成的端点,只是没接到 ImageService**。
- `dto.LayerHistoryEntry` 结构已定义(dto/history.go)。`MapImageInspect(inspect, history)` 接受 history 参数。

### 1.4 现状总评

- **Domain type 已就绪**:`ImageHistoryLayer` + `ImageHistorySource` + `ImageHistoryError` 都已设计好。
- **Podman 端点已实现**:`RESTClient.ImageHistory` 可直接调。
- **Docker 端点已可用**:`cli.ImageHistory` 是标准 client API。
- **需要**:
  1. 在 `ImageService` 接口加 `History(ctx, ImageSummary)` 方法(签名已简化,见 §2)
  2. Docker / Podman adapter 各加 `History` 实现
  3. Podman 端 `Inspect` 的 history=nil 注释应保留(与 detail 解耦),`History` 走独立路径
  4. Docker 端 `Inspect` 的 history=nil 注释同样保留

## 2. 简化后的 `ImageService.History` 接口设计(主模型拍板)

```go
// In internal/data/runtime/image.go
type ImageService interface {
    List(context.Context, ImageListOptions) ([]ImageSummary, error)
    Inspect(context.Context, ImageSummary) (*ImageDetail, error)
    History(context.Context, ImageSummary) ([]ImageHistoryLayer, error)
}
```

- 入参 `ImageSummary`(而非 `string`):让 adapter 直接拿 `summary.ID / summary.RepoTags` 解析,避免在 runtime 层重复 `splitImageRef` 逻辑。
- 返回 `([]ImageHistoryLayer, error)`:
  - **`ImageHistorySource` 不再由 adapter 返回**——改为**调用方在调 `History` 前**根据 `summary.IsManifest` 判断,manifest list 走 manifest 变体视图(已由 `ImageDetail.ManifestVariants` 承担),根本不发 `History` 请求。
  - `error`:网络 / 解析 / 引擎错误(adapter 内不做 source 区分)。

### 2.1 调用方使用模式(主模型负责)

```go
// 例:keyboard/actions.go 中 ActionImageHistory handler
func openHistoryPage(m *state.AppModel) (*state.AppModel, tea.Cmd) {
    ctr := m.Resources.Images.Selected()
    if ctr == nil {
        ShowToastWarn(m, "no image selected")
        return m, nil
    }
    if ctr.IsManifest {
        // Manifest list: 不调 History,改弹 manifest 变体视图
        // (由 mainPageTemplate 中 m.Detail.ManifestVariants 渲染)
        ShowToastInfo(m, "image is a manifest list; use detail view")
        return m, nil
    }
    m.History.Open(ctr.ID)
    m.Navigation.Mode = state.ModeHistory
    return m, historyFetchCmd(m.Connection.Engine, ctr)
}
```

### 2.2 ImageHistorySource 字段的归宿

`ImageHistorySource` 在 `state.HistoryState` 中**仍然存在**(用于 UI 状态显示),但**不再由 runtime 返回**。State 层接收 `HistoryLoadedMsg{ImageID, Layers, Err}`,根据 `Err == nil` 自行设 `Source = ImageHistoryLayerAPI`(成功的非 manifest 路径)。

## 3. 适配层修改清单

### 3.1 `internal/data/runtime/image.go`

```go
// 在 ImageService 接口 line 118-121 添加:
History(context.Context, ImageSummary) ([]ImageHistoryLayer, error)
```

### 3.2 `internal/data/runtime/docker/service_image.go`

```go
func (s imageService) History(ctx context.Context, summary runtimeapi.ImageSummary) ([]runtimeapi.ImageHistoryLayer, error) {
    raw, err := s.client.cli.ImageHistory(ctx, summary.ID)
    if err != nil {
        return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceImage, "history"),
            runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage, ID: summary.ID}, runtimeapi.Docker)
    }
    layers := make([]runtimeapi.ImageHistoryLayer, 0, len(raw))
    for _, h := range raw {
        layers = append(layers, runtimeapi.ImageHistoryLayer{
            ID:        h.ID,
            Created:   h.Created,
            CreatedBy: h.CreatedBy,
            Size:      h.Size,
            Comment:   h.Comment,
        })
    }
    return layers, nil
}
```

### 3.3 `internal/data/runtime/podman/service_image.go`

```go
func (s PodmanImageService) History(ctx context.Context, summary runtimeapi.ImageSummary) ([]runtimeapi.ImageHistoryLayer, error) {
    raw, err := s.Client.REST.ImageHistory(ctx, summary.ID)
    if err != nil {
        return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceImage, "history"),
            runtimeapi.ResourceRef{Type: runtimeapi.ResourceImage, ID: summary.ID}, runtimeapi.Podman)
    }
    layers := make([]runtimeapi.ImageHistoryLayer, 0, len(raw))
    for _, h := range raw {
        created := h.Created.Unix() // dto.LayerHistoryEntry.Created 是 time.Time, 转为 epoch
        layers = append(layers, runtimeapi.ImageHistoryLayer{
            ID:        h.ID,
            Created:   created,
            CreatedBy: h.CreatedBy,
            Size:      h.Size,
            Comment:   h.Comment,
        })
    }
    return layers, nil
}
```

> **疑问需确认**:`dto.LayerHistoryEntry.Created` 类型是 `time.Time` 还是 `int64`?需读 `internal/driver/podman/dto/history.go` 确认。如果小模型实现时发现是 `int64` 类型,删 `.Unix()` 调用即可。

## 4. mock 测试更新清单

### 4.1 `internal/data/runtime/mockengine/mockengine.go`

```go
// 在 mock engine.Images() 返回值类型上加 History:
func (e *Engine) History(ctx context.Context, summary runtimeapi.ImageSummary) ([]runtimeapi.ImageHistoryLayer, error) {
    return e.Called.(func(context.Context, runtimeapi.ImageSummary) ([]runtimeapi.ImageHistoryLayer, error))(ctx, summary)
}
```

> 现有 mock engine 不在 `ImageService` 接口路径上,而是包了一层 `Engine.Images()`。需读 `mockengine.go` 核对 — `Engine.Images()` 内部是直接转给底层 mock 的 imageService 还是单独 mock。ImageService 接口加方法必然需要 mock 同步更新。

### 4.2 新增测试(各 adapter 各 1 个)

- `internal/data/runtime/docker/service_image_history_test.go`:
  - `TestDockerImageService_History_OK` — mock `cli.ImageHistory`,验证 mapping 字段
  - `TestDockerImageService_History_Error` — 模拟 client 错误
- `internal/data/runtime/podman/service_image_history_test.go`:
  - `TestPodmanImageService_History_OK` — 基于 Podman REST fixture(参考 `images_json.json` testdata)
  - `TestPodmanImageService_History_Error` — 模拟 REST 错误

## 5. 关键风险

| 风险 | 说明 | 缓解 |
|---|---|---|
| **Inspect 与 History 数据一致性** | 如果 History 与 Inspect 异步拉取,层 hash 与 detail ID 可能不匹配 | UI 层用 `ImageDetailID` 缓存两者对齐;Task 卡片已说 BR-007 缓存机制可复用 |
| **Manifest list 误调 History** | manifest list 调 History 浪费 + 返空 | **调用方在调 `History` 前**判 `summary.IsManifest`,manifest 走 manifest 变体视图(已有 `ImageDetail.ManifestVariants`) |
| **HistoryLoadedMsg 过期** | 用户在拉取中切到别的镜像,后到响应被错误 Apply | `handleHistoryLoaded` 首行 `if m.History.ImageID != msg.ImageID: return`,**此为 state 层职责,非 runtime** |
| **History 端点性能** | 大镜像(>50 layer)拉取可能慢 | 单独 `tea.Cmd` 异步拉;UI 加 loading 状态 |
| **Podman 5.x / 4.x 端点差异** | REST 端点版本协商 | 沿用 `RESTClient.ensureVersion` 机制 |
| **Docker API 限流** | `cli.ImageHistory` 在 daemon 端可能慢 | UI 层加超时 |

## 6. 不应触碰的文件

- `internal/data/runtime/containers.go`(容器 API,与 BR-034 无关)
- `internal/tui/state/*`(BR-034 UI 层,主模型/小模型 B 改)
- `internal/tui/keyboard/*`(键盘路由,主模型/小模型 B 改)
- `internal/tui/ui/app/page_templates.go`(共享集成文件,主模型改)
- `internal/tui/actionbar/registry.go`(Action Bar 项,主模型改 — **注意路径已纠正:不再用 `internal/tui/ui/widget/actionbar/`**)

## 7. 输出给 BR-034 UI-State 分析的输入

- 新的 `ImageService.History(ctx, ImageSummary) ([]ImageHistoryLayer, error)` 接口已就绪
- `ImageHistorySource` 由 **调用方(state 层)**根据 `summary.IsManifest` 决定,不进 runtime 返回值
- UI 需按 source 决定渲染:
  - `manifest`:不调 `History`,改弹 manifest 变体视图(已有 model,详见 `image.go:14-25`)
  - `layer_api`:**调用方在调 `History` 前**已确保 `!summary.IsManifest`;UI 收到后 `m.History.Source = ImageHistoryLayerAPI`
  - `pending`:渲染 loading 态

## 8. 输出给主模型(实施时)

1. **接口签名已定(主模型拍板)**: `History(ctx, ImageSummary) ([]ImageHistoryLayer, error)` —— 不含 `ImageHistorySource` 返回
2. **Manifest 决策在调用方**: `keyboard/actions.go:openHistoryPage` 入口先判 `summary.IsManifest`,manifest 走变体视图
3. Podman 端 Inspect 仍传 `history=nil`(`MapImageInspect(*inspect, nil)`),不动 Inspect 语义
4. Docker 端 Inspect 仍不拉 history(已显式注释),不动 Inspect 语义
5. 两个 adapter 的 `History` 方法共享 model 转换逻辑(可考虑提到 runtime 包,本次分析先保留 adapter 内)
6. **Action Bar 正确路径**:`internal/tui/actionbar/registry.go`(不是 `internal/tui/ui/widget/actionbar/`)

报告完成。**未修改任何代码**。