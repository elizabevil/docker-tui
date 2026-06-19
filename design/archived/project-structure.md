# dtui 项目结构图与实现状态

## 一、代码结构

```
internal/
├── data/                          # 数据层
│   ├── config/                     配置加载 (TOML/YAML/JSONC)
│   ├── docker/                     Docker SDK 封装 + 连接池
│   └── i18n/                       国际化 (en/zh JSONC)
│
├── tui/                           # 视图模型层
│   ├── key/                        按键常量 + 动作类型 + 键映射
│   │   ├── keys.go                   功能键常量        [已实现]
│   │   ├── letters.go                字母键常量        [已实现]
│   │   ├── display.go                显示符号          [已实现]
│   │   ├── action.go                 动作类型          [已实现]
│   │   ├── mapping.go                默认键映射        [已实现]
│   │   └── help.go                   帮助条目结构      [已实现]
│   │
│   ├── keyboard/                    按键分发处理器
│   │   ├── keyboard.go               主分发器          [已实现]
│   │   ├── navigate.go               导航              [已实现]
│   │   ├── container_action.go       容器操作          [已实现]
│   │   ├── image_action.go           镜像操作          [已实现]
│   │   ├── volume_action.go          卷操作            [已实现]
│   │   ├── network_action.go         网络操作          [已实现]
│   │   ├── compose_action.go         Compose 操作      [已实现]
│   │   ├── delete_action.go          删除操作          [已实现]
│   │   ├── mark_action.go            标记操作          [已实现]
│   │   ├── confirm.go                确认模式          [已实现]
│   │   ├── exec_dialog.go            执行对话框        [已实现]
│   │   └── log.go                    日志操作          [已实现]
│   │
│   ├── state/                       应用状态模型
│   │   ├── app.go                     AppModel 主状态   [已实现]
│   │   ├── containers.go              容器状态/消息    [已实现]
│   │   ├── images.go                  镜像状态/消息    [已实现]
│   │   ├── volumes.go                 卷状态/消息      [已实现]
│   │   └── networks.go                网络状态/消息    [已实现]
│   │
│   ├── tables/                       表格列定义配置
│   │   └── config.go                  列宽/列定义       [已实现]
│   │
│   ├── ui/                           视图渲染层
│   │   ├── app/                        应用入口
│   │   │   └── layout.go                主布局 (RenderApp) [已实现]
│   │   │
│   │   ├── component/                  可复用组件库 (22 文件)
│   │   │   ├── styles_load.go           样式加载+SafeFallback [✅重构]
│   │   │   ├── config.go                JSONC 配置加载       [已实现]
│   │   │   ├── config.jsonc             列宽/断点/搜索/面包屑 [✅扩展]
│   │   │   ├── styles.jsonc             语义 Token 定义      [✅扩展16个]
│   │   │   ├── table.go                 通用表格渲染          [已实现]
│   │   │   ├── rows.go                  行渲染器              [已实现]
│   │   │   ├── dialog.go                对话框统一API         [✅重构]
│   │   │   ├── filter.go                搜索过滤 [NEW]        [✅已实现]
│   │   │   ├── breadcrumb.go            面包屑导航 [NEW]      [✅已实现]
│   │   │   ├── toast.go                 通知队列 [NEW]        [✅已实现]
│   │   │   ├── keyhint.go               快捷键提示 [NEW]      [✅已实现]
│   │   │   ├── viewport.go              滚动视口              [已实现]
│   │   │   ├── stateicon.go             状态颜色              [已实现]
│   │   │   ├── helpers.go               工具函数              [已实现]
│   │   │   └── *_test.go                测试 (4)              [✅新增]
│   │   │   └── *.jsonc                  配置 (5)              [✅新增]
│   │   │
│   │   ├── pages/                       页面级视图 (10 文件)
│   │   │   ├── containers/view.go        容器表格             [已实现]
│   │   │   ├── images/view.go            镜像表格             [已实现]
│   │   │   ├── volumes/view.go           卷表格               [已实现]
│   │   │   ├── networks/view.go          网络表格             [已实现]
│   │   │   ├── compose/view.go           Compose 面板         [已实现]
│   │   │   ├── detail/view.go            容器详情             [已实现]
│   │   │   ├── detail/image.go           镜像详情             [已实现]
│   │   │   ├── logs/view.go              日志视图             [已实现]
│   │   │   └── help/view.go              帮助页面             [已实现]
│   │   │
│   │   ├── widget/                       界面部件 (8 文件)
│   │   │   ├── header/header.go          顶部信息栏           [已实现]
│   │   │   ├── footer/footer.go          底部状态栏+快捷键    [✅i18n清理]
│   │   │   └── dialog/                   覆盖层对话框         [已实现]
│   │   │       ├── view.go               统一渲染入口         [已实现]
│   │   │       ├── box.go                对话框容器           [已实现]
│   │   │       ├── selection.go          选择对话框           [已实现]
│   │   │       ├── exec.go               执行对话框           [已实现]
│   │   │       └── notification.go       通知对话框           [已实现]
│   │   │
│   │   └── style/                        样式工具 (1 文件)
│   │       └── style.go                  调色板 Palette       [✅重构]
│   │
│   ├── composable/                     可组合工具
│   │   └── filter.go                   搜索过滤          [已实现]
│   ├── term/                           终端工具
│   │   └── buffer.go                   环形缓冲区        [已实现]
│   ├── utils/                          通用工具
│   ├── update.go                       消息路由          [已实现]
│   ├── styles.go                       ApplyTheme 入口   [✅精简]
│   └── logo_embed.go                   Logo 嵌入         [已实现]
│
└── test/
    └── perf/                           性能测试          [基准测试]
```

## 二、文档清单与实现状态

| 编号  | 文档路径                                     | 内容            | 状态    |
|-----|------------------------------------------|---------------|-------|
| D01 | `PROJECT.md`                             | 项目总览、架构、键位表   | ✅ 已审阅 |
| D02 | `design/master.md`                       | 设计总纲、7条原则、布局  | ✅ 已审阅 |
| D03 | `design/design-system.md`                | UI三层、组件树、样式系统 | ✅ 已审阅 |
| D04 | `design/ui/layout.md`                    | 屏幕分层、响应式      | ✅ 已审阅 |
| D05 | `design/ui/header-bar.md`                | 顶部信息栏设计       | ✅ 已审阅 |
| D06 | `design/ui/status-bar.md`                | 底部状态栏设计       | ✅ 已审阅 |
| D07 | `design/ui/shortcuts-bar.md`             | 快捷键栏设计        | ✅ 已审阅 |
| D08 | `design/ui/toast.md`                     | Toast 通知设计    | ✅ 已审阅 |
| D09 | `design/ui/confirm-dialog.md`            | 确认弹窗设计        | ✅ 已审阅 |
| D10 | `design/ui/help-view.md`                 | 帮助页面设计        | ✅ 已审阅 |
| D11 | `design/ui/table-common.md`              | 通用表格设计        | ✅ 已审阅 |
| D12 | `design/ui/themes.md`                    | 6套主题配色        | ✅ 已审阅 |
| D13 | `design/ui/preview-system.md`            | HTML 预览系统     | ✅ 已审阅 |
| D14 | `.omo/drafts/ui-component-design.md`     | 组件设计总纲        | ✅ 已实现 |
| D15 | `.omo/drafts/component-style-design.md`  | 样式设计(安全回退+主题) | ✅ 已实现 |
| D16 | `.omo/plans/ui-component-refactoring.md` | 重构工作计划        | ✅ 已完成 |

## 三、架构关键节点

### 3.1 数据流

```
Docker API → tea.Cmd → Msg struct → Update(m, msg) → View(m)
```

### 3.2 样式解析链

```
getStyleChain(name, component):
  第2层 → 组件私有 (RegisterComponentStyles)
  第1层 → 全局 styles.jsonc
  第0层 → 调色板名 (style.Colors)
  兜底 → SafeFallback.Normal (永不panic)
```

### 3.3 主题切换

```
ApplyTheme(theme) → style.Colors 更新 → GetStyle()自动取新值
JSONC配置不动，仅调色板变 → 语义色自动跟随
```

### 3.4 代码规则 (新增生效)

| 规则 | 内容          | 验证                     |
|----|-------------|------------------------|
| R1 | 按键使用 key 常量 | `string(key.Key)` → 0次 |
| R2 | 用户文字使用 i18n | `"Down"` 在 footer → 0次 |
| R3 | 禁止匿名 struct | 4处已修复为具名类型             |
| R4 | 注释使用中文      | 核心文件已翻译                |
