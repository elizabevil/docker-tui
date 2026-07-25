# dtui UI Preview

`docs/ui-preview` 是一个独立的前端预览工程，用来快速演示 TUI 页面布局和主题效果。当前实现不是“零依赖静态 HTML”，而是 `Vue 3 + Vue Router + TypeScript + Vite`。

## 快速开始

```bash
cd docs/ui-preview
npm install
npm run dev
```

常用命令：

```bash
npm run build
npm run preview
```

## 当前技术栈

- `vue`
- `vue-router`
- `typescript`
- `vite`

## 目录结构

```text
docs/ui-preview/
├── src/
│   ├── App.vue
│   ├── main.ts
│   ├── components/
│   ├── pages/
│   └── styles/
├── public/data/
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## 页面路由

当前路由定义在 `src/main.ts`：

- `/containers`
- `/images`
- `/volumes`
- `/networks`
- `/detail`
- `/logs`
- `/help`
- `/catalog`

## 主题

当前侧边栏可切换的主题来自 `src/App.vue`：

- `default`
- `nord`
- `dracula`

样式入口：

- `src/styles/terminal.css`
- `src/styles/themes/nord.css`
- `src/styles/themes/dracula.css`

## 这套预览的作用边界

- 它用于验证页面结构、颜色、终端氛围和组件组合。
- 它不是主程序的运行时 UI。
- 它和 Go TUI 不共享组件代码，只共享设计方向。

涉及布局规范时，仍以 [../../design/current-design.md](../../design/current-design.md) 为准。
