# Compose Panel — 双栏项目+服务面板

> 最后更新: 2026-06-15
> 实现文件: `internal/tui/ui/pages/compose/view.go`
> 配置文件: `internal/tui/tables/compose.jsonc`

## 状态

- [x] 已实现 (双栏布局 + 顶部焦点色条 + 项目概览)

## 布局架构

```
┌──────────────────────────────────────────────┐  ← panel.Render() 统一外层边框
│  Compose                                      │  ← 面板标题
│  ─────────────────────────                    │  ← 顶部色条（绿色=左栏焦点）
│  项目列表            │ 服务列表                │  ← 竖线分隔（首行）
│  ▸myapp   running   │ nginx  ● 2/2            │
│   blog    partial   │ redis  ○ 0/1            │
│   test    stopped   │ api    ● 1/1            │
│  1-3/3              │ s/S 启停                │
└──────────────────────────────────────────────┘  ← 底部边框始终可见
```

所有页面共享统一外层边框组件 `widget/panel/panel.go`。

## 焦点指示器

| 设计   | 说明                           |
|------|------------------------------|
| 顶部色条 | 聚焦面板上方显示 `─` 横线，聚焦=绿色，非聚焦=灰色 |
| 无全边框 | 左右子面板**不**使用独立圆角边框，避免高度挤压    |
| 竖线分隔 | 首行 `│` 标识两栏边界，不额外占高度         |

焦点指示器颜色通过 `compose.jsonc` 配置：

```jsonc
"layout": {
  "focus": {
    "activeColor": "63",    // 聚焦面板色 (ANSI 256)
    "inactiveColor": "240"  // 非聚焦面板色
  }
}
```

## 高度计算

```
midH  = 中间段总高度
panel.Render(Height: midH):
  contentH = midH - 2     ← 预留 2 行给上下边框
  inner = MaxHeight(contentH).Render(...)  ← 内容截断在 contentH 内
  总输出 = contentH + 2 = midH            ← 边框始终完整可见
```

## 交互

| 操作           | 结果                 |
|--------------|--------------------|
| `→`          | 切换到右栏 (服务列表)       |
| `←`          | 切换到左栏 (项目列表)       |
| `Enter` (左栏) | 切换到右栏（同 →）         |
| `Enter` (右栏) | 跳转到容器面板，过滤显示该服务的容器 |
| `d` (左栏)     | 项目概览 (ModeDetail)  |
| `s` / `S`    | 启动/停止项目所有容器        |
| `l`          | 服务日志               |
| `Ctrl+D`     | Compose Down       |

## 层级钻取

```
Compose Panel (L1 项目 + L2 服务)
    │ Enter (右栏)
    ▼
Container Panel (L3 容器, 自动过滤)
    │ d/l/e
    ▼
Detail/Logs/Exec (L4)
```

## 数据来源

从容器列表 (`m.Containers.Items`) 聚合：

1. 遍历所有容器，提取 `ComposeProject` 和 `ComposeService` 标签
2. 按项目名分组 → `composeProj`（内含服务列表 `svcs`）
3. 按服务名分组 → `composeSvc`

## 列定义

### 项目列表

| 列        | 字段                          |
|----------|-----------------------------|
| PROJECT  | 项目名                         |
| STATUS   | running / partial / stopped |
| SERVICES | 服务数                         |
| PODS     | 容器总数                        |

### 服务列表

| 列       | 字段  |
|---------|-----|
| SERVICE | 服务名 |
| IMAGE   | 镜像名 |
| PODS    | 容器数 |
