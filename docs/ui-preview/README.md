# dtui UI Preview

浏览器端终端模拟设计预览系统。零依赖，直接打开 HTML 即可查看。

## 快速开始

```bash
# 方式 1: 直接打开 (fetch API 需要 HTTP 服务)
cd ui-preview && python3 -m http.server 8080
# 打开 http://localhost:8080

# 方式 2: 使用 just
just preview

# 方式 3: 任意静态文件服务
npx serve ui-preview
```

## 迭代流程

```
1. 编辑 data/containers.json     → 修改模拟数据
2. 编辑 styles/terminal.css       → 修改全局样式
3. 编辑 styles/themes/nord.css    → 修改主题配色
4. 编辑 pages/* 部分 (在 index.html 中) → 修改页面布局
5. 刷新浏览器                      → 即时预览
```

## 目录结构

```
ui-preview/
├── index.html                # 主入口 (所有页面内联)
├── styles/
│   ├── terminal.css           # 终端模拟 CSS
│   └── themes/
│       ├── nord.css           # Nord 主题
│       └── dracula.css        # Dracula 主题
├── data/
│   ├── containers.json        # 容器模拟数据
│   └── images.json            # 镜像模拟数据
└── scripts/
    └── theme-switcher.js      # 主题切换逻辑
```

## 可用页面

| 页面               | 导航                                             |
|------------------|------------------------------------------------|
| Container Table  | 容器列表 + Toast + Header + Status Bar + Shortcuts |
| Container Detail | 容器详情 (7 分区)                                    |
| Image Table      | 镜像列表 (6 列)                                     |
| Image Detail     | 镜像详情                                           |
| Log View         | 日志视图 (行号+时间戳)                                  |
| Help View        | dtui logo + 快捷键                                |
| Toast Demo       | 成功/失败 Toast                                    |
| Confirm Dialog   | 删除确认弹窗                                         |
| Status Bar       | 状态栏变体                                          |
| Shortcuts Bar    | 不同面板的快捷键栏                                      |
