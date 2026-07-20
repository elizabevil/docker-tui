# dtui 设计系统规范 v3

> 更新日期: 2026-06-15
> 对应代码重构: 组件提取 + 样式体系改造完成

---

## 一、UI 四层架构

```
Layer 0: 应用骨架 (app/layout.go)
├── Header (widget/header/header.go)
├── Toast (component/toast.go)
├── SearchBar (component/filter.go)
├── Main Panel
│   ├── Resource Table (pages/*/view.go)
│   ├── Detail View (pages/detail/view.go)
│   ├── Log View (pages/logs/view.go)
│   ├── Compose Panel (pages/compose/view.go)
│   └── Help View (pages/help/view.go)
├── Dialog Overlay
│   ├── Confirm Dialog (component/dialog.go)
│   └── Selection Dialog (widget/dialog/)
└── Footer (widget/footer/footer.go)
    ├── Shortcuts Bar
    └── Status Bar

Layer 1: 基础组件 (component/)
├── table.go            RenderTable + ColStyles (列级 TextStyle)
├── rows.go             RowRenderer (按列独立样式)
├── filter.go           SearchFilter
├── breadcrumb.go       Breadcrumb
├── toast.go            ToastQueue
├── keyhint.go          KeyHint
├── dialog.go           RenderDialog 统一 API
├── viewport.go         EnsureVisible
└── stateicon.go        StateColor

Layer 2: 配置系统 (JSONC)
├── config.jsonc        列宽/断点/FlexTable
├── styles.jsonc        语义 Token
├── table.jsonc         列级样式定义
├── toast.jsonc         通知配置
├── filter.jsonc        搜索样式
└── dialog.jsonc        对话框配置

Layer 3: 样式系统 (style/ + TOML)
├── Palette (12色)       → style.Colors struct
├── Semantic Tokens       → styles.jsonc
├── Component Overrides   → {comp}.jsonc
└── 6 TOML 主题           → ApplyTheme 入口
```

---

## 二、组件树

```
App (package view, app/layout.go)
├── Header (widget/header/header.go)
│   ├── Col 1: Host dynamic info (CPU/MEM/Disk)
│   ├── Col 2: Connection + App config
│   ├── Col 3: Keystroke display (spinner box)
│   └── Col 4: Logo + version
├── Toast (component/renderToast) — optional
├── SearchBar (component/renderSearchBar) — optional
├── Main Panel
│   ├── Normal Mode
│   │   ├── Title (PanelLabel + InfoMessage)
│   │   ├── FlexTable (pages/*/view.go)
│   │   └── Table Footer (N-N/N | hint)
│   ├── Detail Mode (pages/detail/view.go)
│   │   └── Collapsible sections (1-9)
│   ├── Log Mode (pages/logs/view.go)
│   │   └── Timestamped log lines
│   ├── Compose Mode (pages/compose/view.go)
│   │   ├── Left: Project List
│   │   └── Right: Service List
│   └── Help Mode (pages/help/view.go)
├── Dialog Overlay
│   ├── Confirm Dialog (component/dialog.go)
│   ├── Shell Dialog (component/dialog.go)
│   ├── Selection Dialog (widget/dialog/selection.go)
│   └── Exec Dialog (widget/dialog/exec.go)
└── Footer (widget/footer/footer.go)
    ├── Shortcuts Bar (面板感知快捷键)
    └── Status Bar (连接/容器/错误状态)
```

---

## 三、样式系统

### 3.1 配置链

```
TOML 主题 (6套)
    ↓
tui/styles.go → ApplyTheme()
    ├── 更新 style.Colors 12色 struct
    └── 重建 3 个 border 全局样式
    ↓
component/GetStyle("name")
    ├── ① 组件私有 {comp}.jsonc
    ├── ② 全局 styles.jsonc
    ├── ③ palette 色名
    └── ④ SafeFallback.Normal
    ↓
pages/*/view.go 渲染
```

### 3.2 调色板 (12色)

```go
var Colors = Palette{
Green:   "#2ecc71", Cyan:    "#00bcd4",
Blue:    "#42a5f5", Red:     "#ef5350",
Yellow:  "#ffca28",  Orange:  "#ff9800",
Purple:  "#ab47bc", White:   "#ffffff",
Gray:    "#546e7a", Dark:    "#1a1a2e",
Surface: "#16213e", BG:      "#0d1117",
}
```

### 3.3 语义 Token (styles.jsonc)

```jsonc
"searchBar":     { "color": "white" },
"toastSuccess":  { "color": "green" },
"toastError":    { "color": "red" },
"breadcrumb":    { "color": "gray", "faint": true },
"hintKey":       { "color": "white", "bold": true },
"dialogTitle":   { "color": "red", "bold": true }
```

### 3.4 列级样式 (table.jsonc, 类似 Compose TextStyle)

```jsonc
{
  "key": "name",  "width": 14,
  "style": { "color": "white", "bold": true }   // fontColor + fontWeight
},
{
  "key": "state", "width": 16,
  "style": { "color": "green" }                  // 默认运行色
}
```

### 3.5 安全回退 SafeFallback

```go
SafeFallback.Normal → 白色文字
SafeFallback.Bold   → 白色加粗
SafeFallback.Dim    → 灰色弱化
SafeFallback.Accent → 青色强调
SafeFallback.Error  → 红色错误
```

---

## 四、键盘处理架构

```
tea.KeyPressMsg
    ↓
update.go → keyboard.HandleKeyPress(msg, m)
    ↓
keyboard.go dispatcher:
├── ModeHandlers (ModeHelp/Filter/Detail/LogView/Confirm)
├── PanelHandlers (images/compose/volumes)
└── GlobalKeys (Space=mark, Ctrl+D=delete, H=header)
    └── ActionMapping (navigate/compose_action/container_action)
```

---

## 五、数据流

```
Docker Events API           Keyboard Input
       │                          │
       ▼                          ▼
   docker/events.go          tea.KeyPressMsg
       │                          │
       └──────┬───────────────────┘
              ▼
      Update() 消息路由
         switch msg.type
              │
              ▼
      View() → RenderApp(m)
              │
              ▼
        终端输出
```

---

## 六、响应式断点配置

列宽通过 `config.jsonc` + `table.jsonc` 配置：

| 页面 | 窄屏列 | 中屏列 | 宽屏列 | 断点阈值               |
|----|-----|-----|-----|--------------------|
| 容器 | 6   | 8   | 9   | wide:130, more:85  |
| 镜像 | 4   | 6   | 7   | wide:140, more:108 |
| 卷  | 3   | 5   | 6   | wide:110, extra:75 |
| 网络 | 4   | 6   | 7   | wide:110, extra:75 |
