# dtui 架构重构方案 — 最终版

## 零：配置统一 JSON

**决策**：所有配置统一为 JSON 格式。

| 旧格式                     | 新格式             | 理由                       |
|-------------------------|-----------------|--------------------------|
| `config.yml`            | `config.json`   | 统一格式 + encoding/json 零依赖 |
| `themes/*.toml`         | `themes/*.json` | 去掉 TOML 依赖               |
| `component/config.json` | 拆入各组件目录         | 组件自包含                    |

**迁移**：

```
internal/data/config/
├── config.json          ← 原 config.yml
├── defaults.json        ← DefaultConfig()
├── themes/
│   ├── default.json     ← 原 default.toml
│   └── nord.json
└── embed.go             //go:embed *.json themes/*.json
```

---

## 一：消除重复代码

### 1.1 CalcRowHeight

出现 6 次 `panelHeight - 3`。提取到 `component/helpers.go`：

```go
func CalcRowHeight(panelHeight int) int {
h := panelHeight - 3
if h < 1 { return 1 }
return h
}
```

### 1.2 ensureCursorVisible

出现 6 次。提取到 `component/viewport.go`。

### 1.3 Loading 检查

出现 4 次。提取到 `component/helpers.go`：

```go
func RenderLoading[T any](items []T) string {
if items == nil { return "(empty)" }
return ""
}
```

### 1.4 错误包装 (22处)

提取到 `tui/result.go`：

```go
func Result(err error) (bool, error) { return err == nil, err }
```

---

## 二：UI 组件架构化

**规则**：每个组件 = **1 个 Go 文件 + 1 个 JSON 配置文件**。

**命名规则**：Go 文件全小写 + JSON 文件同名。

### 组件清单

```
internal/tui/ui/component/
├── flextable.go / flextable.json     FlexTable (已有 ✅)
├── stateicon.go / stateicon.json     StateIcon (已有 ✅)
├── searchbar.go / searchbar.json     搜索栏 (新建)
├── toast.go / toast.json             Toast (新建，从 layout.go 提取)
├── dialog.go / dialog.json           对话框 (新建，从 ui/dialog.go 提取)
├── viewport.go / viewport.json       视口滚动 (新建)
└── helpers.go                        辅助函数 (新建)
```

### JSON 与 Go 嵌入

```go
// component/flextable.go
//go:embed flextable.json
var flexTableJSON embed.FS
var flexCfg FlexTableConfig
func init() {
data, _ := flexTableJSON.ReadFile("flextable.json")
json.Unmarshal(data, &flexCfg)
}
```

---

## 三：Vue 风格 Style 方案

**三层结构**：

```
theme.json (Layer 1: Palette — 颜色值)
  ↓ 引用颜色名
component/{name}.json (Layer 2: Component Styles — 样式定义)
  ↓ 生成 lipgloss.Style
component/{name}.go → Render() (Layer 3: 渲染)
```

**组件 JSON 示例**：

```json
{
  "properties": {
    "columnSpacing": 1,
    "rowPrefix": "  "
  },
  "styles": {
    "header": {
      "color": "cyan",
      "bold": true
    },
    "selected": {
      "color": "blue",
      "bold": true
    },
    "footer": {
      "color": "gray",
      "faint": true
    }
  }
}
```

组件 styles 引用 `theme.json` 中的**颜色名**（"cyan"），不是 hex 值 → 更换主题自动生效。

---

## 四：按键处理拆分

`keyboard.go` 335行 → 拆为同包文件：

```
internal/tui/
├── keyboard.go          顶层分发 (~40行)
├── keyboard_detail.go   ModeDetail
├── keyboard_log.go      ModeLogView
├── keyboard_confirm.go  ModeConfirm
├── keyboard_compose.go  Compose 面板
├── keyboard_images.go   Images 面板
├── keyboard_volumes.go  Volumes/Networks
├── keyboard_global.go   Space, Ctrl+D, H, c
└── keyboard_actions.go  ActionXxx 分发
```

所有文件同属 `package tui`，直接拆分无需改 import。

---

## 五：Update 消息路由拆分

`update.go` 253行 → 按消息类型拆分：

```
internal/tui/
├── update.go            主路由 + WindowSize (~25行)
├── update_container.go  ContainersLoaded, ContainerActioned
├── update_image.go      ImagesLoaded, ImageActioned
├── update_log.go        LogBatchReceived, LogTick
├── update_stats.go      StatsReceived, StatsTick
├── update_timer.go      ToastTick, SearchTick, EscTimeout
└── update_host.go       HostStatsTick
```

---

## 六：执行路线图

```
Phase 1 (当前)
  ✅ 配置统一 JSON — config.json + themes/*.json
  ✅ 消除重复 — CalcRowHeight + ensureCursorVisible
  ✅ 组件 go+json 配对

Phase 2
  ⬜ keyboard.go → keyboard_*.go 拆分
  ⬜ update.go → update_*.go 拆分
  ⬜ handler_container.go → command/ + action/

Phase 3
  ⬜ View 函数参数化 (去掉 *state.AppModel)
  ⬜ Vue 风格 Style 三层结构
  ⬜ detail.go / compose.go 打散
```
