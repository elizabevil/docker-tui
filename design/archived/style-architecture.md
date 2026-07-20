# Draft: 样式设计 — 安全回退与主题切换

## 一、安全回退：Compose 默认值模式

### Compose 怎么做

```kotlin
// 1. 定义带默认值的 CompositionLocal
val LocalTuiColors = staticCompositionLocalOf {
    TuiColors(       // ← 这是默认值，不是 nil
        border  = Color.Cyan,
        success = Color.Green,
        error   = Color.Red,
        // ...每个字段都有安全值
    )
}

// 2. 子组件读取
@Composable
fun TuiToast() {
    val successColor = LocalTuiColors.current.success  // ← 永远有值，不会 null
}
```

关键点：**不是返回"空"，而是返回"合理的默认值"**

### 当前 Go TUI 问题

```go
// 当前: 找不到就返回空 Style
func GetStyle(name string) lipgloss.Style {
    ref, ok := rawStyles[name]
    if !ok {
        return lipgloss.NewStyle()  // ← 空 Style，渲染出来是透明的
    }
    // ...
}
```

### 改造方案：DefStyle + StyleFunc

```go
// component/styles_load.go

// DefStyle 内置 5 个安全默认 Style，永不返回空
var DefStyle = struct {
    Normal  lipgloss.Style
    Bold    lipgloss.Style
    Dim     lipgloss.Style
    Accent  lipgloss.Style
    Error   lipgloss.Style
}{
    Normal: lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")),
    Bold:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true),
    Dim:    lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Faint(true),
    Accent: lipgloss.NewStyle().Foreground(lipgloss.Color("#00bcd4")),
    Error:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ef5350")),
}

// GetStyle 增强版: 多层查找 + 类型化回退
//  1. 先查组件私有 {} styles
//  2. 再查全局 styles.jsonc
//  3. 再查 palette 色名 (green → #2ecc71)
//  4. 回退到 DefStyle.Normal (不会 panic，不会透明)
func GetStyle(name string) lipgloss.Style {
    // ... 1/2/3 ...
    return DefStyle.Normal  // ← 最终安全回退
}

// GetStateStyle 专为状态颜色设计
//  不存在的状态名返回 DefStyle.Normal
func GetStateStyle(state string) lipgloss.Style {
    switch state {
    case "running":    return GetStyle("stateRunning")
    case "exited":     return GetStyle("stateStopped")
    case "paused":     return GetStyle("statePaused")
    case "created":    return GetStyle("stateCreated")
    default:           return DefStyle.Normal
    }
}
```

### Compose vs Go TUI 回退链对比

```
Compose:         LocalColor.current.primary
                         ↓
                 MaterialTheme.colorScheme.primary
                         ↓
                 darkColorScheme().primary    ← Compose 内置默认值
                         ↓
                 Color(0xFF...)              ← 硬件默认(永远不会 null)

Go TUI (改造后): GetStyle("rowSelected")
                         ↓
                 filterStyles["rowSelected"]   ← 组件私有
                         ↓
                 rawStyles["rowSelected"]      ← styles.jsonc 全局
                         ↓
                 style.Color("white")          ← palette 色名
                         ↓
                 DefStyle.Normal               ← 安全默认
                         ↓
                 lipgloss.NewStyle()           ← lipgloss 底层(永不 panic)
```

---

## 二、主题切换方案

### 切换的本质

```
Vue:    data-theme="nord"  →  CSS 变量层叠覆盖  →  所有 var(--tui-xxx) 自动更新
Compose: CompositionLocalProvider  →  theme对象替换  →  所有 .current 读取自动更新
Go TUI: color.Colors 更新  →  GetStyle() 下次读取新值  →  触发 tea.WindowSizeMsg 重绘
```

### 当前实现分析

```go
// tui/styles.go
func ApplyTheme(theme *config.Theme) {
    // 1. 更新 12 色调色板
    ColorGreen = lipgloss.Color(c.Green)
    ColorCyan = lipgloss.Color(c.Cyan)
    // ...

    // 2. 同步 style.Colors map (component.GetStyle 依赖此)
    style.Colors = map[string]color.Color{
        "green": ColorGreen, "cyan": ColorCyan,
        // ...
    }

    // 3. 重建 26 个全局 Style 变量 ← 脆弱点：每新增一个都要在此加一行
    ActiveBorderStyle = lipgloss.NewStyle()...Foreground(resolveColor(m.BorderActive))
    TitleStyle = lipgloss.NewStyle()...Foreground(resolveColor(m.Title))
    // ...23行
}
```

### 问题

1. **26 个全局变量手动维护** — 新增组件样式需要同时改 styles.jsonc + styles.go
2. **不级联** — style.Colors 更新后，已缓存的 GetStyle() 会自动刷新（因为每次调用都重新 resolve），但全局变量不会
3. **ApplyTheme 过于集中** — 每次改主题都要全部重建

### 改进方案：Colors 作为单一数据源

```go
// style/style.go — 调色板是唯一的可变状态

var Colors = struct {
    Green   color.Color
    Cyan    color.Color
    Blue    color.Color
    Red     color.Color
    Yellow  color.Color
    Orange  color.Color
    Purple  color.Color
    White   color.Color
    Gray    color.Color
    Dark    color.Color
    Surface color.Color
    BG      color.Color
}{
    // 编译时默认值 (对应 default 主题)
    Green:   lipgloss.Color("#2ecc71"),
    Cyan:    lipgloss.Color("#00bcd4"),
    // ...
}
```

```go
// ApplyTheme 精简版 — 只更新调色板
func ApplyTheme(theme *config.Theme) {
    c := theme.Colors
    style.Colors.Green  = lipgloss.Color(c.Green)
    style.Colors.Cyan   = lipgloss.Color(c.Cyan)
    style.Colors.Blue   = lipgloss.Color(c.Blue)
    // ...
    
    // ★ 不再重建 26 个全局 Style
    // ★ component.GetStyle() 读取 style.Colors 时自动取到新值
}
```

```go
// styles.go 全局变量 → 函数化，惰性读取 Colors
// 原来: var TitleStyle lipgloss.Style  (在 ApplyTheme 中赋值的)
// 现在: 
func TitleStyle() lipgloss.Style {
    return lipgloss.NewStyle().Foreground(style.Colors.Cyan)
}
// ★ 每次调用读当前 Colors，自动跟随主题切换
```

```go
// 或者保留全局变量但改为在 View 阶段刷新
// 在 layout.go RenderApp() 开头调用
func refreshThemeStyles() {
    // 用当前 style.Colors 重建几个核心全局 Style
    // 这些是高频调用的，函数化有性能开销
    ActiveBorderStyle = lipgloss.NewStyle().Border(...).Foreground(style.Colors.Cyan)
}
```

### 切换触发流程

```
用户选主题(dtui --theme nord)
        │
        ▼
config.LoadTheme("nord")    ← 加载 TOML
        │
        ▼
ApplyTheme(nordTheme)       ← 更新 style.Colors 全部 12 色
        │
        ▼
发送 tea.WindowSizeMsg      ← 触发全量重绘
        │
        ▼
layout.go RenderApp()       ← 每个组件调用 GetStyle()
        │                       → GetStyle("stateRunning") → style.Colors.Green
        │                       → 此时 Green 已指向 nord 的 #a3be8c
        ▼
界面刷新为新主题             ← JSONC 配置不动，调色板变了
```

### 可视化：主题切换时什么变、什么不变

```
主题切换 (default → nord):
┌────────────────────────────────────────────────────────────────┐
│  变: style.Colors 全部 12 色                                   │
│     Green:  #2ecc71  →  #a3be8c                                │
│     Cyan:   #00bcd4  →  #88c0d0                                │
│     Blue:   #42a5f5  →  #81a1c1                                │
│     ...                                                        │
├────────────────────────────────────────────────────────────────┤
│  不变: JSONC 配置文件                                           │
│     styles.jsonc → "stateRunning": { "color": "green" }        │
│     config.jsonc → 列宽/间距/断点                               │
│     filter.jsonc → 搜索栏私有样式                               │
├────────────────────────────────────────────────────────────────┤
│  自动跟随: 所有 GetStyle() 调用                                  │
│     "stateRunning"  →  green  →  现在是 nord 的 #a3be8c        │
│     "toastSuccess"  →  green  →  同上                           │
│     "rowAlt"        →  gray   →  现在是 nord 的 #4c566a         │
└────────────────────────────────────────────────────────────────┘
```

### 与 Vue / Compose 主题切换对比

| 步骤     | Vue CSS             | Compose                        | Go TUI (改造后)                   |
|--------|---------------------|--------------------------------|--------------------------------|
| **触发** | `data-theme="nord"` | `TuiTheme(nordColors)`         | `ApplyTheme(nordTheme)`        |
| **传播** | CSS 变量层叠            | `CompositionLocalProvider`     | `style.Colors` 全局更新            |
| **读取** | `var(--tui-green)`  | `LocalTuiColors.current.green` | `style.Colors.Green`           |
| **重绘** | 浏览器自动重排             | Compose 重组                     | `tea.WindowSizeMsg` 触发的 View() |
| **不变** | CSS class 定义        | Composable 代码                  | JSONC 配置                       |

---

## 三、综合后的新 GetStyle 设计

```go
// component/styles_load.go

// 1. 5 个安全默认值
var SafeFallback = struct {
    Normal  lipgloss.Style
    Bold    lipgloss.Style
    Dim     lipgloss.Style
    Accent  lipgloss.Style
    Error   lipgloss.Style
}{
    Normal: lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")),
    Bold:   lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true),
    Dim:    lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Faint(true),
    Accent: lipgloss.NewStyle().Foreground(lipgloss.Color("#00bcd4")),
    Error:  lipgloss.NewStyle().Foreground(lipgloss.Color("#ef5350")),
}

// 2. 组件私有样式存储 (每个组件 init 时注册)
var componentStyles = map[string]map[string]styleRef{}

// RegisterComponentStyles 由各组件的 init() 调用
func RegisterComponentStyles(component string, refs map[string]styleRef) {
    componentStyles[component] = refs
}

// 3. 多层查找 GetStyle
func GetStyle(name string) lipgloss.Style {
    return getStyleChain(name, "")
}

func GetComponentStyle(component, name string) lipgloss.Style {
    return getStyleChain(name, component)
}

func getStyleChain(name, component string) lipgloss.Style {
    // Layer 2: 组件私有
    if component != "" {
        if refs, ok := componentStyles[component]; ok {
            if ref, ok := refs[name]; ok {
                return buildStyle(ref)
            }
        }
    }
    // Layer 1: 全局 styles.jsonc
    if ref, ok := rawStyles[name]; ok {
        return buildStyle(ref)
    }
    // Layer 0: palette 色名
    if c := style.Color(name); c != nil {
        return lipgloss.NewStyle().Foreground(c)
    }
    // Fallback: 安全默认
    return SafeFallback.Normal
}

func buildStyle(ref styleRef) lipgloss.Style {
    s := lipgloss.NewStyle()
    if ref.Color != "" {
        if c := style.Color(ref.Color); c != nil {
            s = s.Foreground(c)
        }
    }
    if ref.Background != "" {
        if c := style.Color(ref.Background); c != nil {
            s = s.Background(c)
        }
    }
    if ref.Bold   { s = s.Bold(true) }
    if ref.Faint  { s = s.Faint(true) }
    return s
}
```
