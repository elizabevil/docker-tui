# dtui — 当前设计文档

> 最后更新: 2026-06-19
> 此文档为唯一权威设计参考，旧版 design/*.md 已归档。

---

## 一、布局架构

### 1.1 渲染管线的 Box Model 层次

dtui 采用类 HTML 嵌套盒模型（Box Model），从外到内逐层缩进宽度，每层减去各自的边框/间距：

```
┌──────────────────────────────────────────────────┐
│  Terminal (m.Width × m.Height)                   │
│  ┌────────────────────────────────────────────┐  │ ← marginTop (空行)
│  │  ┌──────────────────────────────────────┐  │  │
│  │  │  usableW = m.Width * contentWidthPct │  │  │
│  │  │        ──────────────────            │  │  │
│  │  │        100                            │  │  │
│  │  │  padH = (m.Width - usableW) / 2       │  │  │
│  │  │                                        │  │  │
│  │  │  ┌─ Header (自然高度 4行) ──────────┐  │  │  │ ← widget/header/
│  │  │  │ host(3) | conn(5) | key(7) | logo│  │  │  │
│  │  │  └───────────────────────────────────┘  │  │  │
│  │  │  ┌─ Toast (可选, ≤3行) ───────────────┐  │  │  │ ← component/toast.go
│  │  │  │  通知消息 (3秒自动消失)               │  │  │  │
│  │  │  └───────────────────────────────────┘  │  │  │
│  │  │  ┌─ Search Bar (可选) ────────────────┐  │  │  │ ← layout:renderSearchBar
│  │  │  │  / 搜索 或 : 命令输入框              │  │  │  │
│  │  │  └───────────────────────────────────┘  │  │  │
│  │  │  ┌─ Panel (剩余全部高度) ─────────────┐  │  │  │ ← widget/panel/
│  │  │  │ ╭─ panelW = usableW ────────────╮  │  │  │  │
│  │  │  │ │ Title: 面板名 │ 面包屑      │  │  │  │  │
│  │  │  │ │ Content: contentW = panelW-4  │  │  │  │  │
│  │  │  │ │   ┌── 表格 / 日志 / 详情 ──┐  │  │  │  │  │
│  │  │  │ │   │ 表头行                  │  │  │  │  │  │
│  │  │  │ │   │ 数据行 (colW计算后填充) │  │  │  │  │  │
│  │  │  │ │   │ 页脚: 1-6/6 │ pagination│  │  │  │  │  │
│  │  │  │ │   └─────────────────────────┘  │  │  │  │  │
│  │  │  │ ╰──────────────────────────────╯  │  │  │  │
│  │  │  └───────────────────────────────────┘  │  │  │
│  │  │  ┌─ Footer Shortcuts (双行) ──────────┐  │  │  │ ← widget/footer/
│  │  │  │  j↓ k↑ Tab / : ? H C q            │  │  │  │
│  │  │  │  s S r K l m d e                   │  │  │  │
│  │  │  └───────────────────────────────────┘  │  │  │
│  │  │  ┌─ Status Bar (1行) ─────────────────┐  │  │  │ ← widget/footer/
│  │  │  │  ● docker │ /var/run/docker.sock   │  │  │  │
│  │  │  └───────────────────────────────────┘  │  │  │
│  │  └──────────────────────────────────────┘  │  │
│  └────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────┘
```

### 1.2 Flex 高度分配 (layout.go:sectionHeights)

顶部/中部/底部三段按权重分配，默认 `2:7:1`，但实际运行时**按自然高度重算**：

```
usableH = m.Height - marginTop - marginBot
topH, midH, botH = sectionHeights(m, usableH)
  → 权重分配: topH = usableH * 2/10
              botH = usableH * 1/10
              midH = usableH - topH - botH

实际渲染时重算（layout.go:296-308）:
  headerH = header.Render() 的实际行数 (+ toast)
  footerH = footer.Render() 的实际行数
  midH    = usableH - headerH - footerH   ← 剩余全部给面板
```

| 区域 | 配置 | 自然高度 | 说明 |
|------|------|---------|------|
| Header | `header.jsonc` | 4 行 | 四栏: host/connection/keystroke/logo |
| Toast | — | 0~3 行 | 有消息时显示，超 3 行截断 |
| Search Bar | — | 1 行 | `/` 或 `:` 模式出现 |
| Panel | 无 | `midH` | 边框(2) + 标题(1) + 内容(bodyH = midH-3) |
| Shortcuts | — | 1~3 行 | 全局行 + 页面行 + 操作日志行 |
| Status Bar | — | 1 行 | 引擎标识/连接状态/负载 |

### 1.3 App Window 配置 (app.jsonc)

```
internal/tui/ui/app/
├── app.jsonc        ← 应用窗口占比配置 (嵌入 Go 二进制)
├── layout.go        ← 主布局入口 RenderApp()
└── bgcache.go       ← 图片背景缓存
```

```jsonc
{
  "marginTopPct": 5,
  "marginBottomPct": 5,
  "contentWidthPct": 90
}
```

代码兜底 (`layout.go:45-55`):
```go
Fallback: appWindowConfig{
    MarginTopPct:    5,
    MarginBottomPct: 5,
    ContentWidthPct: 90,
}
```

---

## 二、宽度传播链 (Width Propagation Chain)

宽度遵循**从外到内逐层缩减**的传递链，每层减去自己的边框/间距开销：

```
Terminal Width (m.Width)
    │
    ├── contentWidthPct 映射
    │   usableW = m.Width * contentWidthPct / 100
    │   padH   = (m.Width - usableW) / 2  ← 水平填充（左右空白）
    │
    ▼
usableW  ← 所有子区域的渲染宽度基准
    │
    ├── Header:   usableW - 4 (内部列间距) → 各列按权重分配
    ├── Search:   usableW - 4 (searchBar 边框 2 + padding 2) → innerW
    ├── Toast:    usableW (直接渲染, MaxHeight=3)
    │
    ├── Panel:    panelW = usableW  ← layout.go:391 传入 panel.Panel{}
    │   │
    │   ├── 边框 (2):  `tui.ActiveBorderStyle.Width(panelW - 2)`
    │   │
    │   ├── 标题行:     lineW = panelW - 6 (距边框间距)
    │   │
    │   └── 内容区:     contentW = panelW - 4 ← layout.go:363
    │       │                    (panelW 传入后减左右边框)
    │       │
    │       ├── 表格: containerW = contentW
    │       │   ├── 前缀: rowPrefix (2) + 间隙
    │       │   ├── 列宽: colW[i] = max(表头宽, 内容宽), 受 Widths[i] 限制
    │       │   ├── 间距: gap = (containerW - ΣcolW) / (n-1)
    │       │   └── 分页: 页脚占 1 行, bodyH = panelHeight - overhead
    │       │
    │       ├── 日志: 直接使用 bodyH × contentW
    │       │
    │       └── 详情: 使用 bodyH × contentW
    │
    ├── Compose (双栏):
    │   │   totalW = panelWidth - 4
    │   │   ratioL : ratioR = 4 : 6 (compose.jsonc 可配置)
    │   │   leftW  = totalW * ratioL / (ratioL + ratioR)
    │   │   rightW = totalW - leftW - 1  ← 分隔线占 1
    │   │
    │   └── 各子表格: 各自按 width - 前缀等再缩减
    │
    └── Footer:  usableW (直接渲染)
```

**关键代码跟踪:**

| 层次 | 文件 | 行 | 表达式 |
|------|------|---|--------|
| usableW | `layout.go` | 212 | `m.Width * contentWidthPct / 100` |
| padH | `layout.go` | 216 | `(m.Width - usableW) / 2` |
| panelW | `layout.go` | 391 | `panel.Panel{Width: usableW, ...}` |
| 边框扣减 | `panel.go` | 43 | `p.Width - 2` (传给 lipgloss) |
| contentW | `layout.go` | 363 | `panelW - 4` |
| 表列宽 | `table.go` | 78 | `computeContentWidths(d, headers)` |
| 列间距 | `table.go` | 259 | `(containerWidth - total) / (n-1)` |
| Compose 分栏 | `view.go` | 98-104 | `totalW * ratioL / (ratioL + ratioR)` |

---

## 三、Section 配置体系

每个界面部件均有独立 JSONC 配置文件（嵌入 Go 二进制），通过 `ConfigLoader[T]` 加载：

```
所有 config 均使用: component.ConfigLoader[T]{
    RawData:  embed 的 JSONC 字节
    Fallback: 代码内兜底结构体
    Normalize: 可选后处理函数
}.Load()
```

### 3.1 各部件配置一览

| 部件 | 配置文件 | 加载点 | 关键键 |
|------|---------|--------|--------|
| App Window | `internal/tui/ui/app/app.jsonc` | `layout.go:34-55` | `marginTopPct`, `marginBottomPct`, `contentWidthPct` |
| Header | `internal/tui/ui/widget/header/header.jsonc` | `header.go:21-88` | `columns[].weight`, `keystroke.contentRatio` |
| Footer | `internal/tui/ui/widget/footer/footer.jsonc` | `footer.go:16-40` | `markSymbol` |
| Panel Border | `internal/tui/ui/component/borders.jsonc` | `border.go` | `borders.*` (5 种: rounded/double/thick/single/hidden) |
| Table | `internal/tui/ui/component/table.jsonc` | `table_config.go` | `rowStyles`, `stateStyles`, `pages.*.columns` |
| Table Cols | `internal/tui/ui/component/config.jsonc` | `config.go` | `cols.*.pct/more/wide/compact/show` |
| Component Styles | `internal/tui/ui/component/styles.jsonc` | `styles_load.go` | `styles.*` (搜索/面包屑/Toast/快捷键/对话框等) |
| Dialog | `internal/tui/ui/component/dialog.jsonc` | `dialog.go` | 对话框布局参数 |
| Filter | `internal/tui/ui/component/filter.jsonc` | `filter.go` | 搜索过滤参数 |
| Toast | `internal/tui/ui/component/toast.jsonc` | `toast.go` | 通知队列参数 |

### 3.2 Style 解析链

```
GetStyle("stateRunning")
  → ① 组件私有 {comp}.jsonc  ← (如 table.jsonc 的 stateStyles)
  → ② 全局 styles.jsonc       ← component 级共享样式
  → ③ palette 色名 → #499c54  ← style.Color() 调色板查找
  → ④ SafeFallback.Normal    ← 安全兜底 (白色普通)
```

### 3.3 响应式列宽断点

每页定义多套列宽配置，通过 `ContainerProfileSelector` / `BreakpointProfileSelector` 按可用宽度自动选择：

```jsonc
"container": {
  "wide":  [8, 14, 12, 16, 10, 5, 8, 12, 0],  // ≥130 列
  "more":  [8, 14, 12, 16, 0,  5, 8, 12],      // ≥85 列
  "pct":   [8, 16, 14, 18, 0, 14],              // 默认 (百分比)
  "show":  { "wide": 130, "more": 85, "stats": 110 }
}
```

选择逻辑 (`profile_selector.go:21-31`): 从宽到窄匹配，第一个满足 `width >= MinWidth` 的 profile 生效。

---

## 四、Panel 内部布局

### 4.1 Panel 结构 (widget/panel/panel.go)

```go
type Panel struct {
    Title       string  // 面板标题
    Info        string  // 附加信息 (选中项详情)
    Content     string  // 预渲染内容
    Breadcrumb  string  // 面包屑导航 (右侧)
    SearchText  string  // 兼容字段
    BorderLabel string  // 上边框标签 (搜索内容)
    Width       int
    Height      int
}
```

### 4.2 渲染流程

```
Panel.Render()
    │
    ├── titleLine = renderTitle(p)
    │   ├── title + " │ " + info → styled(panelTitle)
    │   └── 面包屑 → JustifyBetween(titleLine, breadcrumb, lineW)
    │
    ├── innerH = Height - 2       ← 去掉上下边框
    ├── contentH = innerH - 1     ← 去掉标题行
    ├── body = MaxHeight(contentH).Render(p.Content)
    ├── inner = JoinVertical(titleLine, body)
    ├── boxed = ActiveBorderStyle.Width(Width - 2).Render(inner)
    │
    └── BorderLabel != "" → 替换首行为 ╭─ Search: xxx ─╮
```

### 4.3 Panel 高度构成

```
midH  (sectionHeights 分配)
  │
  ├── border top:    1 行  ← ╭──────────╮
  ├── title line:    1 行  ← 面板名 │ 面包屑
  ├── content body: bodyH  ← 表格/日志/详情
  ├── border bottom: 1 行  ← ╰──────────╯
  └── (panel footer 在 content 内部由 RenderTable 处理)
```

### 4.4 表格内部布局 (component/table.go:RenderTable)

```
[SelectionInfo]            ← 选中项预览 (可选, 居中)
[表头行]                   ← 列名 + 排序箭头 (↑/↓)
[空行]                     ← 表头与数据间 1 行间距

[数据行 0]                 ← RowRenderer.RenderRow()
[数据行 1]
  ...
[空行填充]                 ← bodyLines 不足 bodyHeight 时的填充

[页脚]                     ← "1-6/6 │ 4 running"
```

数据行渲染器 (`rows.go`):
- `buildStyle(ref).Render(content)` 整行包裹
- 选中行前置 `┃` (rowPrefixSelected), 普通行前置 `  ` (rowPrefix)
- 选中/标记背景色不重置单元格内 ANSI

---

## 五、组件架构

```
internal/tui/ui/
├── app/
│   ├── layout.go             主布局入口 (RenderApp)
│   ├── app.jsonc             应用窗口占比配置
│   └── bgcache.go            图片背景缓存
│
├── component/                可复用组件库
│   ├── table.go              RenderTable + SelectionInfoProvider 接口
│   ├── table_config.go       table.jsonc 加载
│   ├── table.jsonc           列样式/行间距/选中项配置
│   ├── config.go             全局 Config 加载 + ColProfile
│   ├── config.jsonc          列宽百分比/搜索/面包屑参数
│   ├── column_widths.go      列宽计算 + 响应式断点判断
│   ├── profile_selector.go   容器/断点 Profile 选择器
│   ├── rows.go               行渲染 + 列独立样式 + 选中/标记背景
│   ├── border.go             边框样式选择 + ActiveBorderStyle
│   ├── borders.jsonc         5 种预定义边框字符集
│   ├── filter.go             搜索过滤组件
│   ├── filter.jsonc          过滤参数
│   ├── breadcrumb.go         面包屑导航
│   ├── toast.go              通知队列 (4 级时长 + 历史 50 条)
│   ├── toast.jsonc           通知参数
│   ├── keyhint.go            快捷键提示
│   ├── spinner.go            加载动画
│   ├── dialog.go             对话框统一 API
│   ├── dialog.jsonc          对话框参数
│   ├── styles_load.go        SafeFallback + getStyleChain
│   ├── styles.jsonc          通用组件样式
│   ├── config_loader.go      ConfigLoader[T] 统一加载器
│   ├── helpers.go            验证辅助 (Coerce/WrapStyle/FirstNonZero)
│   ├── selection_info.go     SelectionInfo 选中项预览
│   ├── stateicon.go          状态颜色
│   ├── viewport.go           滚动视口
│   ├── viewhelper.go         视图辅助
│   ├── str.go                字符串工具 (VisibleLen/Truncate)
│   ├── layout_line.go        布局行工具
│   └── constants.go          常量
│
├── pages/                    页面级视图
│   ├── containers/view.go    容器表格
│   ├── images/view.go        镜像表格
│   ├── volumes/view.go       卷表格
│   ├── networks/view.go      网络表格
│   ├── compose/view.go       Compose 双栏 + 容器子视图
│   ├── detail/                详情视图
│   ├── logs/view.go          日志视图
│   └── help/view.go          帮助页面
│
├── widget/                   界面部件
│   ├── panel/panel.go        统一面板 (边框/标题/内容)
│   ├── header/header.go      顶部信息栏 (四栏: host/conn/keystroke/logo)
│   ├── footer/footer.go      状态栏 + 快捷键 (双行)
│   └── dialog/*              覆盖层对话框
│
└── style/style.go            调色板 Palette (6 套主题)
```

---

## 六、表格组件

### 6.1 TableData

```go
type TableData struct {
    Cols, Widths, Rows, Selected int
    Total, Offset, Limit         int
    BannerW    int                 // 宽度参考
    FooterHint string
    MarkedRows map[int]bool
    ColStyles []ColumnStyle        // 列级样式
    SelectionProvider SelectionInfoProvider
}
```

### 6.2 SelectionInfoProvider 接口

```go
type SelectionInfoProvider interface { SelectionInfo() string }
type FuncInfoProvider struct { Fn func() string }
```

| 页面 | 预览内容 (表头上方居中) |
|------|----------------------|
| 容器 | `id  image` |
| 镜像 | `registry/name:tag` |
| 卷 | `name` |
| 网络 | `name  driver` |

### 6.3 行样式 (table.jsonc)

```jsonc
"selected": { "color": "white", "background": "#37676f", "bold": true },
"normal":   { "color": "white" },
"alt":      { "faint": true, "background": "#1e1f22" },
"marked":   { "color": "white", "background": "#463f16", "bold": true }
```

| 状态 | 前置 | 背景 | 文字 |
|------|------|------|------|
| selected | ┃ | 青色灰 #37676f | 白色加粗 |
| marked | ☑ | 黄 50% #463f16 | 白色加粗 |
| alt | — | 暗灰 #1e1f22 | 弱化 |

选中/标记行使用 `buildStyle(ref).Render(content)` 整行包裹，单元格 ANSI 内嵌不重置背景。

### 6.4 列样式 (table.jsonc pages)

| 列类型 | 颜色 | 示例 |
|--------|------|------|
| name | white+bold | container name, image name |
| id | white+faint | container id, image id |
| state | green (动态) | running/exited |
| tag/ports/subnet | cyan | nginx:latest, 80:80 |
| created | white+faint | 2024-01-15 |
| driver/arch | white | overlay2, amd64 |

### 6.5 行间距

```jsonc
"rowSpacing": 0    // 行间额外空行 (table.jsonc.table.rowSpacing)
```

---

## 七、搜索与命令

### 7.1 搜索 (/ 键)

```
/ 按下 → ModeFilter + FilterText=""
键入   → handleFilterInput 累加字符 + 重置 SearchTimer
停止   → SearchTimer 到 0 → ApplyFilter + ModeNormal (自动过滤)
Enter → 手动确认过滤
Esc   → 取消
```

- 防抖: `searchDebounceMs` 配置 (默认 1000ms, 见 config.jsonc)
- 搜索输入显示在面板标题行的 BorderLabel 中: `╭─ Search: xxx ─╮`
- 支持 Backspace 删除

### 7.2 命令 (: 键)

```
: 按下 → ModeCommand + FilterText=""
键入   → 累加字符 (前缀 ": " 绿色)
Tab   → 自动补全 (compose/images/containers/...)
Enter → 执行命令 (跳转面板)
Esc   → 取消
```

支持命令: `compose`, `images`, `containers`, `volumes`, `networks`, `logs`, `help`

---

## 八、键盘快捷键

### 8.1 全局 (所有面板)

```
j/k Tab 导航  / 搜索  : 命令  ? 帮助  H 顶栏  C 连接  q 退出
```

### 8.2 页面特有

| 面板 | 快捷键 |
|------|--------|
| 容器 | s 启动, Ctrl+S 停止, r 重启, Ctrl+K 杀死, l 日志, m 统计, d 详情 |
| 镜像 | p Prune, Ctrl+P Pull, d 详情, Ctrl+B 调试, Ctrl+E 导出, o 排序 |
| 卷/网络 | Enter 扩展, Ctrl+D 删除 |
| Compose | ←→ 切换焦点, Enter 钻取, d 详情, s 启动, Ctrl+S 停止 |

### 8.3 Footer 双行

```
第一行: j↓ k↑ Tab / : ? H C q      ← 全局
第二行: s S r K l m d e            ← 容器面板特有
第三行: 操作日志 (短提示, 始终占位)    ← OperationLogLine
```

---

## 九、样式与配置

### 9.1 主题 (6 套 JetBrains Rider 风格)

| 主题 | 主色调 |
|------|--------|
| default | Rider Darcula 蓝/青/橙 |
| dark | 高对比 Rider 亮白/深黑 |
| light | Rider Light 浅灰/蓝 |
| nord | 北极蓝 Rider 冰蓝 |
| dracula | 紫色 Rider 紫/粉 |
| solarized | 暖色 Rider 青绿/暖黄 |

### 9.2 调色板 (default 主题)

```jsonc
"green": "#499c54", "cyan": "#56b4c2", "blue": "#589df6",
"red": "#db5a5a", "yellow": "#c8a35e", "orange": "#cc7832",
"purple": "#a962b5", "white": "#c9d1d9", "gray": "#5a6270",
"dark": "#1e1f22", "surface": "#2b2d30", "background": "#18191b"
```

### 9.3 样式解析链

```
GetStyle("stateRunning")
  → ① 组件私有 {comp}.jsonc
  → ② 全局 styles.jsonc
  → ③ palette 色名 → #499c54
  → ④ SafeFallback.Normal (安全兜底)
```

### 9.4 配置项快速参考

| 文件 | 配置项 |
|------|--------|
| `app.jsonc` | `marginTopPct: 5`, `marginBottomPct: 5`, `contentWidthPct: 90` |
| `config.jsonc` | `search.debounceMs: 1000`, `cols.*.pct/more/wide/compact/show` |
| `table.jsonc` | `rowStyles`, `stateStyles`, `pages.*.columns`, `selectionInfo`, `rowSpacing` |
| `header.jsonc` | `columns[].weight`, `keystroke.displayDuration: 30`, `keystroke.animDuration: 5` |
| `footer.jsonc` | `markSymbol: ☑` |
| `borders.jsonc` | `borders.rounded/double/thick/single/hidden` |
| `styles.jsonc` | `searchBar`, `panelTitle`, `toast*`, `hint*`, `dialog*`, `detail*`, `log*` |
| `dialog.jsonc` | 对话框布局/按钮配置 |
| `filter.jsonc` | 过滤输入参数 |
| `toast.jsonc` | 通知队列/超时参数 |

---

## 十、代码规则

| 规则 | 说明 |
|------|------|
| 不区分大小写 | 字母按键统一小写匹配，大写功能移至 Ctrl+ |
| i18n | 用户文字通过 `i18n.T()` |
| key 常量 | 按键用 `key.KeyXxx`，显示用 `key.KXxx` |
| 命名类型 | 禁止匿名 struct |
| 中文注释 | Go 注释使用中文 |

(End of file - total 620 lines)
