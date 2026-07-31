# UI Design System

> Extracted from `design/current-design.md` for developer reference.
> The full authoritative design document remains at `design/current-design.md`.

---

## Layout Architecture

### Box Model

dtui uses a nested Box Model, shrinking width from outer to inner layers:

```
Terminal (m.Width x m.Height)
  └─ marginTop
     └─ usableW = m.Width * contentWidthPct / 100
        ├─ Header (natural height, ~4 lines)
        ├─ Toast (optional, max 3 lines)
        ├─ Search Bar (optional, 1 line)
        ├─ Panel (remaining height)
        │  ├─ border (2 lines)
        │  ├─ title line (1 line)
        │  └─ content body (bodyH = midH - 3)
        ├─ Footer Shortcuts (2-3 lines)
        └─ Status Bar (1 line)
```

### Flex Height Allocation

Top/middle/bottom sections use weight-based allocation (default `2:7:1`), then re-calculate by natural height at render time (`layout.go:sectionHeights`):

| Section | Config | Natural Height | Notes |
|---------|--------|---------------|-------|
| Header | `header.jsonc` | 4 lines | host/connection/keystroke/logo |
| Toast | — | 0-3 lines | Shows on message, truncated at 3 |
| Search Bar | — | 1 line | Appears in `/` or `:` mode |
| Panel | — | `midH` | Border(2) + title(1) + content(bodyH) |
| Shortcuts | — | 1-3 lines | Global + page + operation log |
| Status Bar | — | 1 line | Engine ID / connection / load |

### App Window Config

File: `internal/tui/ui/app/app.jsonc` (embedded in binary)

```jsonc
{
  "marginTopPct": 5,
  "marginBottomPct": 5,
  "contentWidthPct": 90
}
```

Background config lives in user `config.yml` under `layout.background` / `layout.sectionWeights`.

---

## Width Propagation Chain

Width shrinks from outer to inner, each layer subtracting its own border/padding:

```
Terminal Width (m.Width)
  │ contentWidthPct mapping
  │ usableW = m.Width * contentWidthPct / 100
  │ padH = (m.Width - usableW) / 2
  ▼
usableW ← all sub-sections render width base
  │
  ├─ Header:   usableW - 4 (internal column spacing)
  ├─ Search:   usableW - 4 (border 2 + padding 2)
  ├─ Toast:    usableW (direct render, MaxHeight=3)
  │
  ├─ Panel:    panelW = usableW (layout.go:391)
  │  ├─ border (2):      tui.ActiveBorderStyle.Width(panelW - 2)
  │  ├─ title line:      lineW = panelW - 6
  │  └─ content area:    contentW = panelW - 4
  │     ├─ Table:        containerW = contentW
  │     │  ├─ prefix:    rowPrefix (2), deducted once
  │     │  ├─ col width: stable profile basis/min/max tracks inside contentW - prefix
  │     │  ├─ fitting:   satisfy visible content first, then share surplus across flexible columns
  │     │  ├─ gap:       fixed 2 cells between columns
  │     │  └─ summary:   top-right, bodyH = panelHeight - overhead
  │     ├─ Log:          bodyH x contentW
  │     └─ Detail:       bodyH x contentW
  │
  ├─ Compose (dual column):
  │  totalW = panel contentWidth
  │  ratioL : ratioR = 4 : 6 (compose.jsonc configurable)
  │  service title = Compose > project > Services
  │  project detail = read-only Service/Image/Status/Containers table
  │
  └─ Footer:  usableW (direct render)
```

**Key code locations:**

| Layer | File | Line | Expression |
|-------|------|------|------------|
| usableW | `layout.go` | 212 | `m.Width * contentWidthPct / 100` |
| padH | `layout.go` | 216 | `(m.Width - usableW) / 2` |
| panelW | `layout.go` | 391 | `panel.Panel{Width: usableW, ...}` |
| border deduction | `panel.go` | 43 | `p.Width - 2` |
| contentW | `layout.go` | 363 | `panelW - 4` |
| col widths | `tables/config.go` | `ResolveLayout` | shared Excel-like stable grid allocation |
| col gap | `component/table.jsonc` | `table.columnSpacing` | fixed two-cell gap |

---

## Section Config System

Each widget has an independent JSONC config (embedded in binary), loaded via `ConfigLoader[T]`:

```
All configs use: component.ConfigLoader[T]{
    RawData:  embedded JSONC bytes
    Fallback: code-level fallback struct
    Normalize: optional post-processing function
}.Load()
```

### Config Files Overview

| Widget | Config File | Load Point | Key Fields |
|--------|------------|------------|------------|
| App Window | `app/app.jsonc` | `layout.go:34-55` | `marginTopPct`, `marginBottomPct`, `contentWidthPct` |
| Header | `widget/header/header.jsonc` | `header.go:21-88` | `columns[].weight`, `keystroke.contentRatio` |
| Footer | `widget/footer/footer.jsonc` | `footer.go:16-40` | `markSymbol` |
| Panel Border | `component/borders.jsonc` | `border.go` | `borders.*` (rounded/double/thick/single/hidden) |
| Shared Table | `component/table.jsonc` | `table_config.go` | `table`, `rowStyles`, `stateStyles`, `columnStyles` |
| Table Data | `tables/*.jsonc` | `tables/config.go` | `columns.*.basis/min/max/shrink`, `show`, resource options |
| Component Styles | `component/styles.jsonc` | `styles_load.go` | `styles.*` (search/breadcrumb/toast/hints/dialog) |
| Dialog | `component/dialog.jsonc` | `dialog.go` | Dialog layout params |
| Filter | `component/filter.jsonc` | `filter.go` | Search filter params |
| Toast | `component/toast.jsonc` | `toast.go` | Notification queue params |

### Style Resolution Chain

```
GetStyle("stateRunning")
  → ① Component-private {comp}.jsonc (e.g. table.jsonc stateStyles)
  → ② Global styles.jsonc (component-level shared styles)
  → ③ Palette color name → #499c54 (style.Color() lookup)
  → ④ SafeFallback.Normal (safe fallback: white normal)
```

### Responsive Column Width Profiles

Each page defines semantic column sizing profiles, selected automatically by `ContainerProfileSelector` / `BreakpointProfileSelector`:

```jsonc
{
  "key": "name",
  "basis": 40,
  "min": 12,
  "shrink": 1,
  "header": "table.name"
}
```

Selection logic (`profile_selector.go:21-31`): match from wide to narrow, first profile where `width >= MinWidth` wins.

The layout contract is:

1. `contentWidth` is calculated by the panel layout and passed unchanged to the page and `RenderTable`.
2. The row prefix is deducted exactly once by `RenderTable`.
3. Adjacent columns always use the configured fixed two-cell gap; free width never changes the gap.
4. `basis` is the preferred track width shared by the header and every row. Columns shrink, weighted by `shrink`, without crossing `min`.
5. On wider viewports, columns with truncated visible content expand toward their measured content width first. Remaining width is then shared evenly by flexible columns, respecting positive `max` values.
6. The final column reaches the viewport's right edge whenever at least one flexible column can still grow. Fixed columns never absorb surplus width.
7. Selected and marked row backgrounds span the same full-width grid.
8. When minimum widths no longer fit, the page must select a narrower profile instead of violating column constraints.

---

## Panel Internal Layout

### Panel Structure

```go
type Panel struct {
    Title       string  // panel title
    Info        string  // additional info (selected item)
    Content     string  // pre-rendered content
    Breadcrumb  string  // breadcrumb navigation (right side)
    SearchText  string  // compatibility field
    BorderLabel string  // top border label (search content)
    Width       int
    Height      int
}
```

### Render Flow

```
Panel.Render()
  ├─ titleLine = renderTitle(p)
  │  ├─ title + " │ " + info → styled(panelTitle)
  │  └─ breadcrumb → JustifyBetween(titleLine, breadcrumb, lineW)
  │
  ├─ innerH = Height - 2       ← remove top/bottom border
  ├─ contentH = innerH - 1     ← remove title line
  ├─ body = MaxHeight(contentH).Render(p.Content)
  ├─ inner = JoinVertical(titleLine, body)
  ├─ boxed = ActiveBorderStyle.Width(Width - 2).Render(inner)
  │
  └─ BorderLabel != "" → replace first line with ╭─ Search: xxx ─╮
```

### Panel Height Composition

```
midH  (sectionHeights allocation)
  ├─ border top:    1 line  ← ╭──────────╮
  ├─ title line:    1 line  ← panel name │ breadcrumb
  ├─ content body: bodyH   ← table/log/detail
  ├─ border bottom: 1 line  ← ╰──────────╯
  └─ table pagination / peer summary stays inside the content top-right
```

### Table Internal Layout

```
[SelectionInfo]        ← selected item preview (optional, centered)
[page / summary]       ← right aligned, e.g. "1-6/20 │ 4 running"
[empty line]           ← 1 line gap between summary and header
[header row]           ← column names + sort arrows (↑/↓)
[data row 0]           ← RowRenderer.RenderRow()
[data row 1]
  ...
[empty fill]           ← fill when bodyLines < bodyHeight
```

Pagination and peer-level statistics have one location: the table's upper-right
summary line. Tables must not duplicate this information at the lower-left.
Application shortcuts remain in the global two-row footer, not in table content.

Row renderer (`rows.go`):
- `buildStyle(ref).Render(content)` wraps entire row
- Selected row prefix: `┃` (rowPrefixSelected), normal: `  ` (rowPrefix)
- Selected/marked background doesn't reset in-cell ANSI

---

## Component Architecture

```
internal/tui/ui/
├── app/
│   ├── layout.go             Main layout entry (RenderApp) + background / imageColorCache
│   └── app.jsonc             App window ratio config
│
├── component/                Reusable component library
│   ├── table.go              RenderTable + SelectionInfoProvider interface
│   ├── table_config.go       table.jsonc loading
│   ├── table.jsonc           Shared table layout, row/state/column styles
│   ├── config.go             Global component style bootstrap
│   ├── profile_selector.go   Container/breakpoint Profile selectors
│   ├── rows.go               Row rendering + per-column styles + selection/marked bg
│   ├── border.go             Border style selection + ActiveBorderStyle
│   ├── borders.jsonc         5 predefined border character sets
│   ├── filter.go             Search filter component
│   ├── filter.jsonc          Filter params
│   ├── breadcrumb.go         Breadcrumb navigation
│   ├── toast.go              Notification queue (4-level timeout + 50-item history)
│   ├── toast.jsonc           Notification params
│   ├── keyhint.go            Keyboard hint
│   ├── spinner.go            Loading animation
│   ├── dialog.go             Dialog unified API
│   ├── dialog.jsonc          Dialog params
│   ├── styles_load.go        SafeFallback + getStyleChain
│   ├── styles.jsonc          Shared component styles
│   ├── config_loader.go      ConfigLoader[T] unified loader
│   ├── helpers.go            Validation helpers (Coerce/WrapStyle/FirstNonZero)
│   ├── selection_info.go     SelectionInfo selected item preview
│   ├── stateicon.go          State colors
│   ├── viewport.go           Scroll viewport
│   ├── viewhelper.go         View helpers
│   ├── str.go                String utils (VisibleLen/Truncate)
│   ├── layout_line.go        Layout line utils
│   └── constants.go          Constants
│
├── pages/                    Page-level views
│   ├── containers/view.go    Container table
│   ├── images/view.go        Image table
│   ├── volumes/view.go       Volume table
│   ├── networks/view.go      Network table
│   ├── compose/view.go       Compose dual-column + container sub-view
│   ├── detail/               Detail view
│   ├── logs/view.go          Log view
│   └── help/view.go          Help page
│
├── widget/                   UI widgets
│   ├── panel/panel.go        Unified panel (border/title/content)
│   ├── header/header.go      Top info bar (4 columns: host/conn/keystroke/logo)
│   ├── footer/footer.go      Status bar + shortcuts (dual line)
│   └── dialog/*              Overlay dialogs
│
└── style/style.go            Palette (6 themes)
```

---

## Table Components

### TableData

```go
type TableData struct {
    Cols, Rows, Selected         // column definitions + visible data
    Total, Offset, Limit         int
    BannerW    int                // authoritative viewport width
    FooterHint string                 // legacy name; rendered in top-right summary
    MarkedRows map[int]bool
    ColStyles []ColumnStyle       // per-column styles
    SelectionProvider SelectionInfoProvider
}
```

`TableData` does not accept caller-computed column widths. `RenderTable` subtracts
the active row prefix from `BannerW` and calls `tables.ResolveLayout` exactly
once. Row content never changes layout allocation; it is truncated or padded
inside the resolved basis/min/max/shrink column. Visible content requirements
receive free width first, then flexible tracks share the remainder so the table
fills the viewport.

### SelectionInfoProvider Interface

```go
type SelectionInfoProvider interface { SelectionInfo() string }
type FuncInfoProvider struct { Fn func() string }
```

| Page | Preview (centered above header) |
|------|--------------------------------|
| Container | `id  image` |
| Image | `registry/name:tag` |
| Volume | `name` |
| Network | `name  driver` |

### Row Styles (table.jsonc)

```jsonc
"selected": { "color": "white", "background": "#37676f", "bold": true },
"normal":   { "color": "white" },
"alt":      { "faint": true, "background": "#1e1f22" },
"marked":   { "color": "white", "background": "#463f16", "bold": true }
```

| State | Prefix | Background | Text |
|-------|--------|------------|------|
| selected | ┃ | teal gray #37676f | white bold |
| marked | ☑ | yellow 50% #463f16 | white bold |
| alt | — | dark gray #1e1f22 | faint |

### Column Styles (table.jsonc pages)

| Column Type | Color | Example |
|------------|-------|---------|
| name | white+bold | container name, image name |
| id | white+faint | container id, image id |
| state | green (dynamic) | running/exited |
| tag/ports/subnet | cyan | nginx:latest, 80:80 |
| created | white+faint | 2024-01-15 |
| driver/arch | white | overlay2, amd64 |

---

## Search & Commands

### Search (/ key)

```
/ pressed → resource page enters ModeFilter, log page enters ModeSearch
Resource table → FilterInput applies filter on each keystroke
Enter → keep current filter and exit edit
Esc Esc → clear filter within 5s double-tap
Log page → SearchInput edits draft only, Enter applies and jumps to first match
Esc → cancel search draft, keep last applied search
```

- Filter and Search use independent modes and input states (no debounce timer)
- Active resource filter displays in panel BorderLabel: `╭─ Filter: xxx ─╮`
- Supports left/right movement, word movement, `Home`/`End`, `Ctrl+A`/`Ctrl+E`, and shell-style deletion

### Commands (: key)

```
: pressed → ModeCommand + FilterText=""
type → accumulate characters (": " prefix in green)
Tab → autocomplete (compose/images/containers/...)
Enter → execute command (jump to panel)
Esc → cancel
```

Supported commands: `compose`, `images`, `containers`, `volumes`, `networks`, `logs`, `help`

---

## Keyboard Shortcuts

### Global (all panels)

```
j/k/↑/↓ navigate  Tab switch panel  / search  : command  ?/F1 help  H header  F2 switch runtime  C connect  q quit
```

### Per-Page

| Panel | Keys |
|-------|------|
| Containers | `s` start · `Ctrl+S` stop · `Ctrl+R` restart · `Ctrl+K` kill · `l` logs · `m` stats · `d` detail · `i` inspect · `e` exec |
| Images | `Enter/→` enter container sub-view · `Ctrl+P` pull · `p` prune · `d` detail · `Ctrl+B` debug · `Ctrl+E` export · `y` copy ref · `o/Ctrl+O` sort |
| Volumes | `Enter` enter volume detail sub-view · `d` detail · `Ctrl+D` delete |
| Networks | `d` detail · `o/Ctrl+O` sort · `Ctrl+D` delete |
| Compose | `←/→` switch focus or sub-view · `Enter` drill down · `s` start · `Ctrl+S` stop · `l` logs · `Ctrl+D` down · `d` detail |

### Footer Structure

```
Line 1: j↓ k↑ Tab / ? H C q              ← global shortcuts
Line 2: current page or mode shortcuts    ← e.g. container shows Space / s / Ctrl+S / ...
Line 3: operation log (short, always)     ← OperationLogLine
```

- In `ModeMark`, `ModeConfirm`, `ModeLogView`, `ModeDetail`, `ModeHelp`, line 2 is overridden by mode-level hints.
- `:` command mode, `F2` runtime switch, etc. exist but aren't shown in Footer global line.

---

## Styles & Config

### Themes (6 JetBrains Rider styles)

| Theme | Primary Color |
|-------|--------------|
| default | Rider Darcula blue/cyan/orange |
| dark | High-contrast Rider white/deep black |
| light | Rider Light gray/blue |
| nord | Arctic blue Rider ice blue |
| dracula | Purple Rider purple/pink |
| solarized | Warm Rider cyan/warm yellow |

### Palette (default theme)

```jsonc
"green": "#499c54", "cyan": "#56b4c2", "blue": "#589df6",
"red": "#db5a5a", "yellow": "#c8a35e", "orange": "#cc7832",
"purple": "#a962b5", "white": "#c9d1d9", "gray": "#5a6270",
"dark": "#1e1f22", "surface": "#2b2d30", "background": "#18191b"
```

### Config Quick Reference

| File | Config Items |
|------|-------------|
| `app.jsonc` | `marginTopPct: 5`, `marginBottomPct: 5`, `contentWidthPct: 90` |
| `config.yml` | `layout.sectionWeights`, `layout.background`, `keymap.*` |
| `tables/*.jsonc` | Data schema: `columns.*.basis/min/max/shrink`, `show`, resource options |
| `table.jsonc` | Shared visuals: `table`, `rowStyles`, `stateStyles`, `columnStyles` |
| `header.jsonc` | `columns[].weight`, `keystroke.displayDuration: 30`, `keystroke.animDuration: 5` |
| `footer.jsonc` | `markSymbol: ☑` |
| `borders.jsonc` | `borders.rounded/double/thick/single/hidden` |
| `styles.jsonc` | `searchBar`, `panelTitle`, `toast*`, `hint*`, `dialog*`, `detail*`, `log*` |
| `dialog.jsonc` | Dialog layout/button config |
| `filter.jsonc` | Filter input params |
| `toast.jsonc` | Notification queue/timeout params |

---

## Code Rules

| Rule | Description |
|------|-------------|
| Case-insensitive | Letter keys unified lowercase match; uppercase functions moved to Ctrl+ |
| i18n | User-facing text via `i18n.T()` |
| Key constants | Keys use `key.KeyXxx`, display uses `key.KXxx` |
| Named types | No anonymous structs allowed |
| Chinese comments | Go comments use Chinese |
