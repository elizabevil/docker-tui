# Draft: UI 组件 + 样式体系设计总纲

## 项目分析结论

### 技术栈

- **框架**: Bubbletea (MVU) + Lipgloss (样式)
- **语言**: Go
- **项目**: Docker/Podman TUI 管理工具 (k9s 风格)

### 现有组件包结构

```
internal/tui/ui/
├── component/          ← 核心可复用组件库 (15 文件)
│   ├── table.go         RenderTable 通用表格
│   ├── rows.go          RowRenderer 行渲染
│   ├── config.go        JSONC 配置加载 (列宽/断点)
│   ├── dialog.go        确认/Shell 对话框
│   ├── helpers.go       工具函数
│   ├── stateicon.go     状态颜色
│   ├── styles_load.go   延迟样式解析
│   ├── viewhelper.go    BuildRows 泛型数据转行
│   ├── viewport.go      EnsureVisible 滚动视口
│   └── constants.go     常量
│
├── header/             顶部信息栏
├── footer/             底部状态栏+快捷键
├── dialog/             覆盖层对话框 (selection/exec/export)
├── detail/             详情视图
├── containers/         容器表格
├── images/             镜像表格
├── volumes/            卷表格
├── networks/           网络表格
├── compose/            Compose 面板
├── logs/               日志视图
├── help/               帮助视图
├── style/              样式注册表
└── layout.go           主布局渲染
```

### 现有样式体系结构

```
TOML 主题(6套)
    │
    ▼
tui/styles.go → ApplyTheme()
    ├── 12 色调色板全局变量 (ColorGreen, ColorCyan...)
    ├── style.Colors map (供 component.GetStyle 使用)
    └── 26 个全局 Style 变量 (ActiveBorderStyle, TitleStyle...)
            │
            ▼
    component/styles_load.go → GetStyle()
            │
            ▼
    各视图包渲染
```

---

## Part A: 6 个新增/重构组件

### A1. SearchFilter — 搜索过滤组件

**从 layout.go 提取**: `renderSearchBar()`, `autocompleteHint()`
**新建**: `component/filter.go`, `component/filter.jsonc`
**配置**: `config.jsonc` 新增 `"search"` 节

### A2. Breadcrumb — 面包屑导航

**从 layout.go 提取**: `breadcrumb()`
**新建**: `component/breadcrumb.go`
**配置**: `config.jsonc` 新增 `"breadcrumb"` 节

### A3. Toast — 通知队列系统

**从 layout.go 提取**: `renderToast()`
**新建**: `component/toast.go`, `component/toast.jsonc`
**新增**: ToastQueue (Push/Tick/Active)

### A4. Dialog 统一入口

**合并**: `component/dialog.go` + `dialog/` 包
**新增**: `component.RenderDialog(DialogConfig)` 统一 API

### A5. KeyHint — 快捷键提示

**从 footer.go 提取**: 面板感知快捷键定义
**新建**: `component/keyhint.go`

### A6. TableFooter — 独立页脚

**扩展**: `component/table.go` 新增 `RenderTableFooter()`

---

## 代码强制规则

### Rule 1: 按键一律使用 key 常量

```go
// 来源: internal/tui/key/keys.go + letters.go + display.go

// ✅ 正确用法
key.KeyS          // 小写 "s"
key.KeySUpper     // 大写 "S"
key.KeyCtrlD      // "ctrl+d"
key.KDown         // 显示符号 "↓"
key.KTab          // 显示 "Tab"

// ❌ 禁止
string(key.KeyS)  // KeyS 已是 string，无需转换
key.KeyS_up       // 已废弃，用 KeySUpper
"s"               // 裸字符串
```

**Key 包完整结构**:

| 文件           | 内容           | 示例                                                |
|--------------|--------------|---------------------------------------------------|
| `keys.go`    | 功能键常+大写兼容    | KeySpace, KeyEnter, KeyEsc, KeyCtrlD, KeyS_up(废弃) |
| `letters.go` | 字母 a-z + A-Z | KeyA..KeyZ, KeyAUpper..KeyZUpper                  |
| `display.go` | 显示符号         | KUp(↑), KDown(↓), KTab, KCtrlD, KJDown(j↓)        |
| `action.go`  | 动作类型         | ActionContainerStart, ActionContainerStop         |

### Rule 2: 所有面向用户的文字使用 i18n

```go
// ✅ 正确
i18n.T("key.down")          // 从 en.jsonc / zh.jsonc 获取翻译
i18n.T("toast.started", name)  // 带参数: "Started nginx"
i18n.T("table.id")          // 表头: "ID"

// ❌ 禁止
"Down"                // 硬编码英文
fmt.Sprintf("过滤: %s", text)  // 硬编码中文
```

**i18n key 命名规范**:

| 前缀                 | 用途     | 例子                                        |
|--------------------|--------|-------------------------------------------|
| `panel.`           | 面板标签   | `panel.containers`, `panel.images`        |
| `table.`           | 表头     | `table.id`, `table.name`, `table.state`   |
| `key.`             | 快捷键描述  | `key.down`, `key.start`, `key.delete`     |
| `key.sym_`         | 按键显示符号 | `key.sym_up(↑)`, `key.sym_jk(j/k)`        |
| `toast.`           | 操作结果   | `toast.started`, `toast.deleted`          |
| `dialog.`          | 对话框文字  | `dialog.delete_title`, `dialog.force`     |
| `status.`          | 状态栏    | `status.connected`, `status.disconnected` |
| `msg.`             | 消息     | `msg.loading`, `msg.no_containers`        |
| `help.`            | 帮助页章节  | `help.navigation`, `help.containers`      |
| `header.`          | 顶栏文字   | `header.shortcuts`                        |
| `inspect.`         | 详情字段   | `inspect.id`, `inspect.digest`            |
| `filter.`          | 搜索过滤   | `filter.prompt`                           |
| `container.state.` | 容器状态   | `container.state.running`                 |

### Rule 3: 新增 i18n key 必须同时提供中英文

```jsonc
// en.jsonc
"key.batch_start": "B.Start",

// zh.jsonc  
"key.batch_start": "批量启动",
```

---

## Part B: 样式体系重构

### 四层样式架构

```
Layer 0: Palette (12色调色板)
  TOML 主题 → style.Colors map
  对标 Vue CSS :root 变量  /  Compose Palette 对象

Layer 1: Semantic Tokens (语义色 → 调色板色)
  styles.jsonc: "stateRunning": { "color": "green" }
  对标 Vue var(--tui-border: var(--tui-cyan))

Layer 2: Component Scope (组件私有)
  各组件自己的 {name}.jsonc: filter.jsonc 的 "searchCursor"
  对标 Vue <style scoped>  /  Compose Component-level Modifier

Layer 3: Modifier (运行时条件)
  marked/selected/alt 状态切换
  对标 Vue :class 绑定  /  Compose Modifier.then()
```

### 安全回退 — Compose 默认值模式

```go
// 5 个安全默认值，永不返回空 Style
SafeFallback.Normal → 白色文字
SafeFallback.Bold   → 白色加粗
SafeFallback.Dim    → 灰色弱化
SafeFallback.Accent → 青色强调
SafeFallback.Error  → 红色错误

// GetStyle 多层查找链:
GetStyle("rowSelected")
  → ① 组件私有 {comp}.jsonc
  → ② 全局 styles.jsonc
  → ③ palette 色名 (green → #2ecc71)
  → ④ SafeFallback.Normal (永不 panic)
```

### 主题切换 — Colors 作为单一数据源

```
切换主题 (default → nord):
  ┌─ 变: style.Colors 12色 (ApplyTheme 更新)
  │    Green: #2ecc71 → #a3be8c
  │    Cyan:  #00bcd4 → #88c0d0
  ├─ 不变: JSONC 全部配置文件
  │    styles.jsonc / config.jsonc / filter.jsonc / toast.jsonc
  └─ 自动跟随: 所有 GetStyle() / buildStyle()
       "stateRunning" → green → 现在是 nord 的 #a3be8c
```

### styles.go 精简

```
当前: 26 个全局 Style 变量 (在 ApplyTheme 中手动重建)
改造后:
  保留 3 个 border 变量 (无法 JSONC 表达)
  其余 23 个 → 调用方改为 GetStyle("xxx")  + style.Colors 惰性读取
```

---

## Part C: 实施计划

### Phase 0 — 样式基础设施 ⚡ (必须最先完成)

| 任务                               | 文件                         | 依赖 |
|----------------------------------|----------------------------|----|
| SafeFallback 5默认值                | `component/styles_load.go` | 无  |
| getStyleChain 多层查找               | `component/styles_load.go` | 无  |
| RegisterComponentStyles          | `component/styles_load.go` | 无  |
| ApplyTheme 精简 (仅更新 style.Colors) | `tui/styles.go`            | 无  |
| styles.jsonc 补充语义 Token          | `component/styles.jsonc`   | 无  |

### Phase 1 — 组件提取 ⚡ (并行)

| 任务           | 新建文件                         | 依赖      |
|--------------|------------------------------|---------|
| SearchFilter | `filter.go` + `filter.jsonc` | Phase 0 |
| Breadcrumb   | `breadcrumb.go`              | Phase 0 |
| Toast 队列     | `toast.go` + `toast.jsonc`   | Phase 0 |

### Phase 2 — 组件统一 ⚡ (并行)

| 任务                    | 文件                        | 依赖      |
|-----------------------|---------------------------|---------|
| Dialog 统一 API         | `dialog.go` 扩展            | Phase 0 |
| KeyHint + TableFooter | `keyhint.go` + `table.go` | Phase 0 |

### Phase 3 — 验证

| 任务           | 方法            |
|--------------|---------------|
| go vet/build | 编译检查          |
| 功能回归         | 与原有行为一致       |
| 单元测试         | 新增组件测试        |
| 配置一致         | JSONC + 主题名正确 |
