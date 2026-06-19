# UI Preview System — 终端模拟设计预览

## 目录结构

```
ui-preview/                    # ← 独立于 design/ 目录
├── index.html                 # 入口：组件目录 + 页面预览 + 主题切换
├── README.md                  # 使用说明
│
├── styles/
│   ├── terminal.css            # 终端模拟基础样式 (暗底、等宽、行高、边框)
│   └── themes/
│       ├── default.css         # 默认暗色主题 (CSS 变量)
│       ├── nord.css            # Nord 北极蓝调
│       └── dracula.css         # Dracula 紫调
│
├── components/                 # 可复用 TUI 组件 (HTML 片段)
│   ├── header-bar.html         # 顶部信息栏
│   ├── toast.html              # Toast 通知
│   ├── table.html              # 通用表格组件
│   ├── table-cell.html         # 表格单元格 (支持状态颜色)
│   ├── detail-section.html     # 详情分区 (可折叠)
│   ├── log-line.html           # 日志行 (时间戳高亮)
│   ├── confirm-dialog.html     # 确认弹窗
│   ├── status-bar.html         # 底部状态栏
│   └── shortcuts-bar.html      # 底部快捷键栏
│
├── pages/                      # 完整页面 (组件拼接)
│   ├── container-table.html    # 容器列表页面
│   ├── container-detail.html   # 容器详情页面
│   ├── image-table.html        # 镜像列表页面
│   ├── image-detail.html       # 镜像详情页面
│   ├── log-view.html           # 日志查看页面
│   └── help-view.html          # 帮助页面
│
├── data/                       # 模拟数据 (JSON)
│   ├── containers.json         # 容器列表数据
│   ├── images.json             # 镜像列表数据
│   └── volumes.json            # 卷列表数据
│
└── scripts/
    ├── loader.js               # 数据加载器 (fetch JSON + 注入)
    ├── theme-switcher.js       # 主题切换 (CSS 变量替换)
    └── router.js               # 页面导航 (hash-based)
```

## 使用方式

```bash
# 1. 直接在浏览器打开
open ui-preview/index.html

# 2. 或用 HTTP 服务 (支持 fetch JSON)
cd ui-preview && python3 -m http.server 8080
# 打开 http://localhost:8080

# 3. 切换主题
# 页面右上角下拉选择 default / nord / dracula

# 4. 切换页面
# 左侧导航栏点击容器表格 / 镜像详情 / 日志视图 ...

# 5. 修改模拟数据
# 编辑 data/containers.json → 刷新浏览器 → 即时看到变化

# 6. 修改组件样式
# 编辑 styles/terminal.css → 刷新浏览器 → 全局样式更新
```

## 组件设计

每个组件是一个 HTML 片段，通过 CSS class 控制样式：

```html
<!-- components/table-row.html -->
<tr class="tui-row">
  <td class="tui-cell-id">abc123de</td>
  <td class="tui-cell-name">nginx-container</td>
  <td class="tui-cell-state" data-state="running">● running</td>
  <td class="tui-cell-ip">172.17.0.2</td>
  <td class="tui-cell-ports">80:80</td>
  <td class="tui-cell-mounts">2v</td>
</tr>
```

## 页面组装

页面通过 `<div data-component="...">` 声明式引入组件：

```html
<!-- pages/container-table.html -->
<div class="tui-page">
  <!-- Header Bar -->
  <div data-component="header-bar" data-theme="default"></div>
  
  <!-- Toast -->
  <div data-component="toast" data-message="✓ Container started (0.3s)"></div>
  
  <!-- Table -->
  <div data-component="table" data-source="data/containers.json">
    <thead>
      <tr><th>ID</th><th>NAME</th><th>IMAGE</th><th>STATE</th><th>IP</th><th>PORTS</th></tr>
    </thead>
  </div>
  
  <!-- Status Bar -->
  <div data-component="status-bar" data-engine="podman" data-count="5"></div>
  
  <!-- Shortcuts Bar -->
  <div data-component="shortcuts-bar" data-panel="containers"></div>
</div>
```

JS loader 读取 `data-component` 属性，fetch 对应组件 HTML，注入页面。

## CSS 变量体系

```css
/* styles/terminal.css */
:root {
  /* 基础 */
  --tui-bg:           #0d1117;
  --tui-font:         'Cascadia Code', 'JetBrains Mono', 'Fira Code', monospace;
  --tui-font-size:    14px;
  --tui-line-height:  1.4;
  
  /* 调色板 (由主题覆盖) */
  --tui-green:        #2ecc71;
  --tui-cyan:         #00bcd4;
  --tui-blue:         #42a5f5;
  --tui-red:          #ef5350;
  --tui-yellow:       #ffca28;
  --tui-orange:       #ff9800;
  --tui-purple:       #ab47bc;
  --tui-white:        #ffffff;
  --tui-gray:         #546e7a;
  --tui-surface:      #16213e;
  --tui-dark:         #1a1a2e;
  
  /* 语义色 */
  --tui-border:       var(--tui-cyan);
  --tui-title-fg:     var(--tui-cyan);
  --tui-header-fg:    var(--tui-blue);
  --tui-selected-bg:  var(--tui-blue);
  --tui-status-fg:    var(--tui-green);
  --tui-dim-fg:       var(--tui-gray);
}
```

## 迭代流程

```
1. 修改组件 → 修改 components/xxx.html  → 刷新页面 → 全局生效
2. 修改数据 → 修改 data/xxx.json       → 刷新页面 → 数据变化
3. 修改主题 → 修改 themes/xxx.css      → 切换下拉 → 配色变化
4. 修改布局 → 修改 pages/xxx.html      → 切换导航 → 布局变化
5. 新增页面 → 创建 pages/xxx.html      → 注册到 index.html 导航
```

## 与 design/ 的关系

| 目录 | 用途 | 产出 |
|---|---|---|
| `design/` | **设计规范文档** | Markdown 文档，人阅读 |
| `ui-preview/` | **设计预览系统** | HTML/CSS/JS，浏览器查看 |

> `design/` 回答 "应该做成什么样"
> `ui-preview/` 回答 "做出来是什么样"

修改流程: `design/` 讨论确定 → `ui-preview/` 修改预览验证 → 代码实现
