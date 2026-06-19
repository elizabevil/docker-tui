# Handoff: UI Preview Flex + I18n

## 当前状态

Vite 6 + Vue 3 + TypeScript + Vue Router，8 个页面全部构建通过。

```
docs/ui-preview/
├── src/
│   ├── components/
│   │   ├── TuiFrame.vue        — 终端窗口框架
│   │   ├── TuiPanel.vue        — 面板容器（title + content + footer slot）
│   │   ├── TuiTable.vue        — 表格组件（columns/rows/selectedIdx）
│   │   └── TuiShortcuts.vue    — 快捷键栏
│   ├── pages/
│   │   ├── ContainersPage.vue  容器列表（53行）
│   │   ├── ImagesPage.vue      镜像列表（80行）
│   │   ├── VolumesPage.vue     卷列表（51行）
│   │   ├── NetworksPage.vue    网络列表（51行）
│   │   ├── DetailPage.vue      详情（87行，可折叠）
│   │   ├── LogViewPage.vue     日志（60行）
│   │   ├── HelpPage.vue        帮助（75行）
│   │   └── CatalogPage.vue     目录（35行）
│   ├── App.vue                 侧栏 + router-view
│   ├── main.ts                 路由 + 入口
│   └── styles/
│       ├── terminal.css         基础样式（115个CSS变量）
│       └── themes/
│           ├── nord.css         [data-theme="nord"]
│           └── dracula.css      [data-theme="dracula"]
├── public/data/                 静态JSON数据
├── index.html
├── vite.config.ts
├── tsconfig.json
└── package.json
```

## 任务一：Flex 布局优化

当前问题：组件使用固定宽度表格，未充分利用 flex 弹性布局。

### 需要修改

**1. TuiTable.vue** — 改用 flex 实现列宽自适应

```vue
<!-- 当前：<table> 标签 -->
<!-- 目标：flex 容器 + 每列 flex-grow 根据 width 比例 -->

// columns 扩展支持 flex 权重
interface FlexCol {
  key: string
  label: string
  flex?: number    // flex-grow 权重，默认1
  minWidth?: number // 最小宽度 px
}
```

**2. 各 Page — 响应式断点**

参考 Go 实现的 `config.jsonc` 列配置：

```json
"container": {
  "pct": [8, 16, 14, 18, 0, 14],
  "more": [8, 14, 12, 16, 0, 5, 8, 12],
}
```

Vue 中实现：

- 使用 `window.matchMedia` 或 `ResizeObserver` 监听容器宽度
- 根据宽度选择列配置（narrow/more/wide）
- 窄屏隐藏低优先级列

**3. TuiFrame.vue** — iframe 内自适应

```css
.tui-frame-body {
  display: flex;
  flex-direction: column;
  flex: 1;
}
```

## 任务二：I18n 国际化

### 推荐方案：vue-i18n

```bash
npm install vue-i18n
```

### 目录结构

```
src/
├── i18n/
│   ├── index.ts          // createI18n 配置
│   ├── locales/
│   │   ├── en.json       // 英文
│   │   └── zh.json       // 中文
```

### 示例 en.json

```json
{
  "nav.containers": "Containers",
  "nav.images": "Images",
  "nav.volumes": "Volumes",
  "nav.networks": "Networks",
  "table.id": "ID",
  "table.name": "NAME",
  "table.image": "IMAGE",
  "table.state": "STATE",
  "table.status": "STATUS",
  "table.ports": "PORTS",
  "table.created": "CREATED",
  "table.registry": "REGISTRY",
  "table.tag": "TAG",
  "table.arch": "ARCH",
  "table.size": "SIZE",
  "table.driver": "DRIVER",
  "table.scope": "SCOPE",
  "table.subnet": "SUBNET",
  "hint.compose_detail": "Enter:detail",
  "key.start": "Start",
  "key.stop": "Stop",
  "key.restart": "Restart",
  "key.logs": "Logs",
  "key.detail": "Detail",
  "key.delete": "Delete",
  "key.filter": "Filter",
  "key.help": "Help",
  "key.quit": "Quit",
  "key.mark": "Mark",
  "key.back": "Back",
  "key.scroll": "Scroll",
  "key.page": "Page",
  "key.top": "Top",
  "key.bottom": "Bottom"
}
```

### 使用方式

```vue
<script setup>
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>
<template>
  <div class="tui-panel-title">{{ t('nav.containers') }}</div>
</template>
```

### 中文 zh.json

复制 en.json 的 key，value 翻译为中文。

### 语言切换

App.vue 添加语言选择器：

```vue
<select v-model="$i18n.locale">
  <option value="en">English</option>
  <option value="zh">中文</option>
</select>
```

## 执行顺序

1. `npm install vue-i18n`
2. 创建 `src/i18n/index.ts` + `locales/en.json` + `locales/zh.json`
3. 更新 `main.ts` 注册 i18n
4. Flex 改造 TuiTable.vue（flex-grow 列）
5. 各 Page 添加响应式断点（narrow / more / wide）
6. App.vue 添加语言选择器
7. 逐步替换硬编码字符串为 `t()`

## 验证

```bash
npm run build   # 确保编译通过
npm run dev     # 浏览器预览
```
