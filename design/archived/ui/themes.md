# Themes — 主题系统

## 状态

- [x] 已实现 (6 套主题，`go:embed` 编译嵌入)

## 主题清单

| 主题名 | 描述 | 色调 |
|---|---|---|
| `default` | k9s 风格深色 | 青色调 |
| `dark` | 高对比深色 | 亮青/亮绿 |
| `light` | 清爽浅色 | 蓝灰调 |
| `nord` | 北极蓝调 | 冰蓝/雪白 |
| `dracula` | 德古拉紫调 | 紫色/粉色 |
| `solarized` | 精准色系 | 青绿/暖黄 |

## 使用方式

```bash
dtui --theme nord              # CLI 参数
dtui -t dracula                # 短参数

# 配置文件 config.yml
theme: solarized

# 列出所有主题
dtui --list-themes
```

## 主题文件结构

每个主题 YAML 定义:

```yaml
name: "Nord"
description: "Arctic, north-bluish color palette"

colors:           # 12 色调色板
  green:   "#a3be8c"
  cyan:    "#88c0d0"
  blue:    "#81a1c1"
  red:     "#bf616a"
  yellow:  "#ebcb8b"
  orange:  "#d08770"
  purple:  "#b48ead"
  white:   "#eceff4"
  gray:    "#4c566a"
  dark:    "#2e3440"
  surface: "#3b4252"
  bg:      "#242933"

border:           # 边框颜色
  active:   "cyan"
  inactive: "gray"
  focused:  "green"
  style:    "rounded"

text:             # 文本角色颜色
  title:    "cyan"
  header:   "blue"
  selected: "blue"
  status:   "green"
  info:     "cyan"
  error:    "red"
  success:  "green"
  warning:  "yellow"
  dim:      "gray"
  help_key: "cyan"
  help_desc: "white"
```

## 应用流程

```
main.go → config.LoadTheme(name)
        → theme.go parseTheme() 解析 YAML
        → ApplyTheme(theme) 填充 styles.go 全局变量
        → 所有 view 组件使用全局 Style 变量渲染
```

## 自定义主题

用户可创建自定义 `~/.config/dtui/themes/custom.yml`，然后:

```bash
dtui --theme ~/.config/dtui/themes/custom.yml
```

`LoadTheme` 先查编译嵌入的主题，找不到则从文件系统路径加载。
