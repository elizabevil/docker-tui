# dtui — 全局设计总纲

> 最后更新: 2026-06-15
> 代码结构: `internal/tui/ui/` 已重构为 app/pages/widget/component

## 项目定位

终端容器管理工具，k9s 风格，支持 Docker + Podman，键盘驱动、实时刷新。

## 设计原则

| 原则           | 说明                                  |
|--------------|-------------------------------------|
| **键盘驱动**     | 所有操作通过键盘完成，鼠标仅辅助                    |
| **渐进式信息**    | 表格→详情(d键)→操作，层层深入                   |
| **统一键位**     | 同义操作跨面板一致 (d=Detail, Ctrl+D=Delete) |
| **k9s 紧凑风格** | 最小间距、最大信息密度                         |
| **响应式**      | 列数随终端宽度自动适配                         |
| **离线降级**     | 连接断开时保留最后数据，状态栏显示 disconnected      |
| **搜索防抖**     | 即时过滤延迟 1s 防抖                        |

## 代码规则

| 规则 | 内容          |
|----|-------------|
| R1 | 按键使用 key 常量 |
| R2 | 用户文字使用 i18n |
| R3 | 禁止匿名 struct |
| R4 | 注释使用中文      |

---

## Ⅰ. Flex 布局架构

### 屏幕分层

```
┌──────────────────────────────────────────────┐
│ Header Bar     (H 切换)  三栏 4:4:4          │ ← widget/header/
├──────────────────────────────────────────────┤
│ Toast 通知     (3秒自动消失)                   │ ← component/toast.go
├──────────────────────────────────────────────┤
│ Search Bar     (/ 激活)                       │ ← component/filter.go
├──────────────────────────────────────────────┤
│                                              │
│  Main Content Area (flex: 占满剩余高度)       │ ← panel.Render()
│  ┌────────────────────────────────────────┐  │    widget/panel/panel.go
│  │  [标题]  Containers                    │  │    统一外层边框
│  │  ID  NAME  IMAGE  STATE               │  │
│  │  ▸abc nginx ● running                 │  │    panel.Height = midH
│  │  1-6/6 │ 4 running                     │  │    内容 = panel.Height - 2
│  └────────────────────────────────────────┘  │    边框保证可见
│                                              │
├──────────────────────────────────────────────┤
│ Status Bar   ● podman │ 5 containers          │ ← widget/footer/
├──────────────────────────────────────────────┤
│ Shortcuts Bar  j↓ k↑ Tab s/S d Detail ...      │ ← widget/footer/
└──────────────────────────────────────────────┘
```

### 高度分配 (flex 权重)

| 区域                           | 权重 | 说明               |
|------------------------------|----|------------------|
| 顶部 (header + toast + search) | 2  | 变高内容，无法精准预测行数    |
| 中部 (main content)            | 7  | flex-grow，占满剩余空间 |
| 底部 (shortcuts + statusBar)   | 1  | 固定 2 行           |

```
midH = 终端高度 × 7/10 - 顶部实际高度
panel.Height = midH
panel 内容 = midH - 2  ← 确保底部边框不截断
```

### 统一面板组件 (`widget/panel/panel.go`)

所有页面通过 `panel.Render()` 获得一致的外框样式：

| 方法                 | 用例                 | 说明                     |
|--------------------|--------------------|------------------------|
| `Render()`         | 容器/镜像/卷/网络/Compose | 青色圆角边框 + 标题 + 内容 + 面包屑 |
| `RenderNoBorder()` | 全屏覆盖模式             | 无边框，仅标题 + 内容           |

panel 内部高度管理：

```go
contentH = p.Height - 2          // 留 2 行给边框
inner = MaxHeight(contentH).Render(parts)
border.Render(inner)             // 总高 = contentH + 2 = p.Height
```

---

## Ⅱ. 全局键位

### 导航

| 按键                  | 功能             |
|---------------------|----------------|
| `j`/`k` / `↑`/`↓`   | 上下导航           |
| `Tab` / `Shift+Tab` | 切换面板           |
| `Esc`               | 返回/取消 (连按两次退出) |
| `Enter`             | 确认/钻取          |
| `/`                 | 搜索过滤 (1s 防抖)   |
| `H`                 | 切换 Header Bar  |

### 容器操作

| 按键        | 功能           |
|-----------|--------------|
| `s` / `S` | Start / Stop |
| `R`       | Restart      |
| `K`       | Kill         |
| `l`       | 日志           |
| `m`       | Stats 切换     |
| `d`       | 详情           |
| `e`       | Exec Shell   |
| `Space`   | 标记           |
| `Ctrl+D`  | 删除 (确认弹窗)    |

### 镜像操作

| 按键        | 功能               |
|-----------|------------------|
| `P` / `p` | Pull / Prune     |
| `d`       | 镜像详情             |
| `D`       | Debug (run --rm) |
| `E`       | Export (save -o) |
| `o` / `O` | 排序/反转            |
| `Enter`   | 查看使用该镜像的容器       |

### Compose 操作

| 按键        | 功能                | 焦点 |
|-----------|-------------------|----|
| `←` / `→` | 左栏(项目) ↔ 右栏(服务)   | 任意 |
| `Enter`   | 左栏→右栏 / 右栏→容器面板   | 任意 |
| `d`       | 项目概览 (ModeDetail) | 左栏 |
| `s` / `S` | 启动/停止项目           | 右栏 |
| `l`       | 服务日志              | 右栏 |
| `Ctrl+D`  | Compose Down      | 右栏 |

---

## Ⅲ. 页面设计

### A. 容器表格

多列响应式，CPU/MEM 可选显示。

| 列       | 宽度%  | 颜色            |
|---------|------|---------------|
| ID      | 8%   | gray          |
| NAME    | 14%  | white bold    |
| IMAGE   | 12%  | white         |
| STATE   | 16%  | 状态色 (●运行/●停止) |
| CREATED | 10%  | gray          |
| IP      | 5%   | white         |
| PORTS   | 8%   | cyan          |
| CPU/MEM | fill | white         |

### B. 镜像表格

| 列        | 宽度%  | 颜色         |
|----------|------|------------|
| REGISTRY | 12%  | gray       |
| NAME     | 18%  | white bold |
| TAG      | 10%  | cyan       |
| IMAGE ID | 16%  | gray       |
| ARCH     | 6%   | white      |
| CREATED  | 14%  | gray       |
| SIZE     | fill | white      |

### C. Compose 双栏面板

```
┌──────────────────────────────────────────────┐
│  Compose                                      │
│  ────────────────                              │ ← 顶部色条(绿=左焦点)
│  项目列表       │ 服务列表                     │
│  ▸myapp running │ nginx ● 2/2                 │
│  1-3/3          │ s/S 启停                    │
└──────────────────────────────────────────────┘
```

详见 `design/features/compose-panel.md`

### D. 容器详情 (d 键)

7 分区可折叠：基本信息 → 网络 → 挂载 → 环境变量 → 标签 → 资源 → 配置

### E. 日志视图 (l 键)

```
   1 2024-01-15T10:00:00.123Z stdout  GET /index.html
   2 2024-01-15T10:00:01.456Z stdout  200 OK
   1-5/200 │ j/k scroll │ Esc back
```

### F. 帮助视图 (? 键)

dtui logo + 快捷键手册

---

## Ⅳ. 组件架构

### 目录结构

```
internal/tui/ui/
├── app/layout.go             主布局入口 (RenderApp)
├── component/                可复用组件库 (22 文件)
│   ├── styles_load.go         SafeFallback + getStyleChain
│   ├── table.go               RenderTable (通用表格)
│   ├── table_config.go        table.jsonc 加载
│   ├── rows.go                行渲染 + 按列独立样式
│   ├── filter.go              搜索过滤组件
│   ├── breadcrumb.go          面包屑导航
│   ├── toast.go               通知队列 (4级+历史缓冲)
│   ├── keyhint.go             快捷键提示
│   ├── spinner.go             加载动画
│   ├── dialog.go              对话框统一 API
│   └── *.jsonc                (5 配置文件)
├── pages/                    页面级视图
│   ├── containers/view.go     容器表格
│   ├── images/view.go         镜像表格
│   ├── volumes/view.go        卷表格
│   ├── networks/view.go       网络表格
│   ├── compose/view.go        Compose 双栏面板
│   ├── detail/*               详情视图
│   ├── logs/view.go           日志视图
│   └── help/view.go           帮助页面
├── widget/                   界面部件
│   ├── panel/panel.go         统一面板 (边框+标题) [NEW]
│   ├── header/header.go       顶部信息栏
│   ├── footer/footer.go       底部状态栏+快捷键
│   └── dialog/*               覆盖层对话框
└── style/style.go             调色板 Palette
```

### 样式解析链

```
GetStyle("stateRunning")
  → ① 组件私有 {comp}.jsonc
  → ② 全局 styles.jsonc → "green"
  → ③ palette 色名 "green" → #2ecc71
  → ④ SafeFallback.Normal (安全兜底)
```

### 主题切换

```
ApplyTheme(nordTheme) → style.Colors 12色更新
  → JSONC 配置不动
  → GetStyle() 下次调用自动取新值
```

---

## Ⅴ. 主题系统 (6套)

| 主题        | 色调           |
|-----------|--------------|
| default   | k9s 风格深色，青色调 |
| dark      | 高对比深色        |
| light     | 清爽浅色，蓝灰调     |
| nord      | 北极蓝调         |
| dracula   | 德古拉紫调        |
| solarized | 精准色系         |

---

## Ⅵ. 设计文件索引

| 编号      | 文件                             | 内容            |
|---------|--------------------------------|---------------|
| D01     | `master.md`                    | **本文档**       |
| D02     | `design-system.md`             | 设计系统规范        |
| D03     | `README.md`                    | design/ 目录导览  |
| D04     | `ui-component-architecture.md` | 组件架构设计        |
| D05     | `style-architecture.md`        | 样式体系设计        |
| D06     | `project-structure.md`         | 项目结构图         |
| D07-D16 | `ui/*`                         | UI 组件设计 (10份) |
| D17-D26 | `features/*`                   | 功能设计 (10份)    |
| D27     | `task/ui-review-report.md`     | UI 覆盘报告       |
| D28     | `task/change-plan.md`          | 变更计划          |
