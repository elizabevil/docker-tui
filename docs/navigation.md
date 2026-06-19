# Page Navigation & Component Architecture

## Window Naming Convention

The UI is divided into three named windows (percentage-split):

```
┌─ HUD (20%) ──────────────────────────────────┐
│ eng docker v24.0.7  host ...    │  █████╗   │  ← Header Bar
│ tz Asia/Shanghai  lang en      │  ██╔══██╗  │     (connection info,
│ cpu 25%  mem 60%               │  ██║  ██║  │      logo, version)
│ dsk 120G  ctr 12/15            │  v0.2.0   │
├─ Panel (70%) ────────────────────────────────┤
│  images > containers          │              │  ← Breadcrumb (top-left)
│  ┌─────────────────────────────────────────┐ │
│  │ docker.io/library/nginx:latest          │ │  ← SelectionBanner
│  │ ID       NAME    IMAGE    STATE  PORTS  │ │  ← Table
│  │ abc123   nginx   nginx    ●up    80:80  │ │
│  │ ...                                     │ │
│  │ 1-6/6                                   │ │
│  └─────────────────────────────────────────┘ │
├─ Bar (10%) ──────────────────────────────────┤
│ j/k Down  Tab Panel  l Logs  / Filter  ? Help│  ← ShortcutsBar
│ ● local-podman │ unix:///run/... │ 4 ctr      │  ← StatusBar
└──────────────────────────────────────────────┘
```

| Window    | Section      | Content                                              |
|-----------|--------------|------------------------------------------------------|
| **HUD**   | Top (20%)    | HeaderBar (connection, logo, version)                |
| **Panel** | Middle (70%) | Bordered content (breadcrumb, banner, table, detail) |
| **Bar**   | Bottom (10%) | ShortcutsBar + StatusBar                             |

## Breadcrumb

The breadcrumb sits at the **top-left corner of the Panel window** (inside the border, as the title line). It replaces
the panel label when a sub-view is active.

Format: `panel > sub-view > detail`

Implemented in `layout.go:breadcrumb()`:

| State                         | Breadcrumb                                            |
|-------------------------------|-------------------------------------------------------|
| Images list                   | `images`                                              |
| Images → Containers sub-view  | `images > containers`                                 |
| Images → Detail               | `images > detail` (or `images > containers > detail`) |
| Volumes list                  | `volumes`                                             |
| Volumes → Containers sub-view | `volumes > containers`                                |
| Compose list                  | `compose`                                             |
| Compose → Services detail     | `compose > services`                                  |
| Compose → Detail              | `compose > services > detail`                         |

The breadcrumb is derived from model state (panel type + sub-view flags). No manual breadcrumb push/pop is needed — it's
computed each render.

## Navigation Flow

```
Tab:      Cycle panels (Containers ↔ Images ↔ Volumes ↔ Networks ↔ Compose)
Enter/→:  Expand to sub-view (image containers, volume containers, compose detail)
Esc/←:    Back from sub-view
l:        Log view (container logs)
d:        Detail/Inspect
o:        Sort (when supported by panel)
```

### Sub-View Navigation

| Parent Panel   | Action  | Sub-View                   | Panel Focus | Back      |
|----------------|---------|----------------------------|-------------|-----------|
| Images         | Enter/→ | Containers using image     | —           | Esc/←     |
| Volumes        | Enter   | Containers using volume    | —           | Esc       |
| Compose        | Enter/→ | Compose detail (services)  | —           | Esc/←     |
| Compose Detail | Enter   | Switch to Containers panel | Containers  | —         |
| (any)          | d       | Detail/Inspect (overlay)   | current     | Esc/Enter |
| (container)    | l       | Log view (overlay)         | —           | Esc       |

## Page Component Architecture

Each page is a self-contained component with a `RenderList` function as its entry point.

### Interface Convention

```go
// Component signature pattern:
func RenderList(model *Model, [relatedModel *Model,] width int, panelHeight int) string
func RenderView(app *AppModel, panelHeight int) string  // full-panel views
func renderXxx(...) string                                // private sub-views
```

### Component Registry

| Component          | File                 | Type     | Description                      |
|--------------------|----------------------|----------|----------------------------------|
| `ContainerList`    | `containers/view.go` | Page     | Container table with stats       |
| `ImageList`        | `images/view.go`     | Page     | Image table                      |
| `ImageContainers`  | `images/view.go`     | Sub-view | Containers using selected image  |
| `VolumeList`       | `volumes/view.go`    | Page     | Volume table                     |
| `VolumeContainers` | `volumes/view.go`    | Sub-view | Containers using selected volume |
| `NetworkList`      | `networks/view.go`   | Page     | Network table                    |
| `ComposeList`      | `compose/view.go`    | Page     | Compose project list             |
| `ComposeDetail`    | `compose/view.go`    | Sub-view | Compose services + pods          |
| `LogView`          | `logs/view.go`       | Overlay  | Container logs with search       |
| `DetailView`       | `detail/view.go`     | Overlay  | Resource inspect detail          |

### Shared Components

| Component            | File                     | Purpose                            |
|----------------------|--------------------------|------------------------------------|
| `SelectionBanner`    | `component/helpers.go`   | Centered current-item banner       |
| `FlexTable`          | `component/flextable.go` | Responsive table with pagination   |
| `StateDot`           | `component/stateicon.go` | Plain state indicator              |
| `StateColoredDot`    | `component/stateicon.go` | ANSI-colored state dot             |
| `Breadcrumb`         | `layout.go`              | Nav path in Panel title bar        |
| `RenderContentLayer` | `layout.go`              | Background image + content overlay |

### Layout Components

| Component   | File               | Purpose                                    |
|-------------|--------------------|--------------------------------------------|
| `AppLayout` | `layout.go`        | Three-section renderer (HUD/Panel/Bar)     |
| `HeaderBar` | `header/header.go` | HUD window content                         |
| `FooterBar` | `footer/footer.go` | Bar window content (Shortcuts + StatusBar) |

## Sub-Page Sizing

All sub-views conform to the percentage-based height system:

```
panelHeight = Panel_window_height - border(2) - title(1) - breadcrumb(1?)
Where Panel_window_height = terminal_height * 70%
```

Sub-views use `CalcRowHeight(panelHeight)` for row limiting. Nested sub-views (ComposeDetail) split `panelHeight`
internally (e.g., 60/40).
