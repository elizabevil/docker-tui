# 颜色配置与透明度方案 (Color Configuration & Transparency)

> 建立日期: 2026-08-05
> 状态: V3 设计确认(用户已拍板三项技术决定),待实施
> 涉及文件: `internal/utils/color.go`, `internal/utils/ansi.go`,
> `internal/data/config/config.go`, `internal/tui/ui/component/styles_load.go`,
> `internal/tui/ui/app/layout.go`, `internal/tui/ui/style/style.go`,
> `internal/tui/styles.go`, `design/current-design.md` §3.2

## 1. 背景与终端能力约束

### 1.1 终端能力的硬约束(2026-08-05 实测确认)

- **24-bit RGB** 颜色(`\033[38;2;R;G;Bm` / `\033[48;2;R;G;Bm`)广泛支持(iTerm2 / kitty / Alacritty / WezTerm / Windows Terminal / GNOME Terminal 新版)
- **alpha 通道 SGR 不存在**:ECMA-48 / ISO 8613-6 SGR 没有 alpha 参数;没有任何标准转义序列支持半透明叠加
- **OSC 11 设背景色** 只支持实色(Konsole / iTerm2),不支持 alpha
- **kitty graphics / Sixel 图片协议** 支持图片本身的 alpha,但**不作用于文本 cell 的背景**

结论:终端文本层没有"半透明混合"概念;所有 alpha 必须在到达 lipgloss 之前**预计算为实色**。

### 1.2 lipgloss v2 的 alpha 行为(Phase 0 实测证据,见 [`test/diagnostics/lipgloss-transparent-probe/main.go`](../test/diagnostics/lipgloss-transparent-probe/main.go))

| 输入 | lipgloss 输出 | 解读 |
|---|---|---|
| `NRGBA{R:0x00, G:0x00, B:0x00, A:0x00}` | `\x1b[48;2;0;0;0mx\x1b[m` (17 字节) | 画**黑色**,**非**透明 |
| `NRGBA{R:0xff, G:0x00, B:0x00, A:0x00}` | `\x1b[48;2;0;0;0mx\x1b[m` (17 字节) | 红被吞成**黑色** |
| `nil` Background | `x` (1 字节) | **唯一**安全 sentinel |
| `A:0` + Width(8) padding | 含 `48;2;0;0;0` 的 40 字节 | padding 也画黑色 |

机制:Go 标准库 `NRGBA.RGBA()` 方法预乘 alpha,A=0 时 RGB 通道全部乘 0;lipgloss 拿到 `(0,0,0,0)` 当成"不透明黑色"渲染,不会走 nil-skip 分支。

## 2. 设计决策(用户三轮对话确认)

### 2.1 V3 双路径设计

| 用户配置 | α 值 | 终端表现 | 实现路径 |
|---|---|---|---|
| `"transparent"` | 隐式 α=0 | **不画**(透到下层) | ParseColor 返回 `NRGBA{A:0}`;消费者侧 α==0 → 跳过 `lipgloss.Foreground/Background()` |
| `"#ff0000"` / `"red"` / `"rgb(255,0,0)"` | α=255 | 画原色 | ParseColor 返回 `NRGBA{R,G,B,0xff}`;消费者直通 |
| `"rgba(255,0,0,0.5)"` / `"#ff000080"` | 0<α<255 | 画**混合后的实色** | ParseColor 保留 α;消费者预混合到基底色 → 实色 hex → 传给 lipgloss |

### 2.2 三态共用一条 `color.Color` 通道

- 单条 `color.Color` 路径,靠 `α` 分流:`α==0` 跳过 / `0<α<255` 预混合 / `α==255` 直通
- **不**引入新类型;**不**新增 `IsTransparent()` 函数;消费者直接读 `.A` 字段判断
- `transparent` 关键字在 ParseColor 中**返回** `NRGBA{R:0, G:0, B:0, A:0}`(用户第二轮确认);消费者读 α==0 触发跳过分支

### 2.3 混合基底(用户第二轮确认)

| 通道 | 半透明混合基底 |
|---|---|
| Background 通道(0<α<255) | `theme.Palette.Background` |
| Foreground 通道(0<α<255) | `theme.Palette.Foreground` |

混合公式(沿用现有 `BlendColors`):

```
result = base + (top - base) * α / 255
```

### 2.4 混合时机:加载期一次性算好(用户第三轮确认)

- 混合计算放在 `ApplyThemeStyles` 加载路径,缓存为实色 hex/rgb
- 运行时(render 阶段)不再算混合,只读取已缓存的实色
- 与现有 `flattenThemeBackground` 模式一致:该函数本来就是加载期算,沿用即可,**不**删除

### 2.5 校验层:放行半透明,拒绝嵌套 transparent

- `isValidColor`:放行所有合法格式(包括 `0<α<255`);**不**拒绝半透明(因为半透明现在有合法用途)
- `ValidateTheme`:增加**嵌套 transparent 拒绝**——若 `Palette.Background == "transparent"` 或 `Palette.Foreground == "transparent"`,报"无基底色可用于半透明混合"

## 3. 现状盘点(已校正的事实)

| 环节 | 状态 | 位置 |
|---|---|---|
| `ParseColor` 支持 8 位 hex / `rgba()`(保留 α) | ✅ 已支持 | `internal/utils/color.go:27-46` |
| `ParseColor` 不识别 `"transparent"` 关键字 | ❌ 缺失 | `color.go:30-45` |
| `isValidColor` 放行半透明(不会拒绝 α=128) | ✅ 已正确 | `config.go:317-319` |
| `flattenThemeBackground` 已实现 bg 混合(加载期缓存) | ✅ 已支持 | `styles_load.go:235-264` |
| Foreground 通道半透明混合 | ❌ 缺失 | 应新增平行函数 |
| `buildStyle` 直塞 `ParseColor` 结果(α 透传 lipgloss) | ⚠️ 需改为三态分流 | `styles_load.go:435-454` |
| `renderAppBackground` 用 `HexToRGB`(只认 6 位 hex) | ⚠️ 受限 | `layout.go:260` |
| `renderAppBackground` 字面量 `"transparent"` 特判 | ⚠️ 散落 1 处 | `layout.go:257` |
| `styles_load.go:237` 用 `FallbackColorTransparent` 常量 | ⚠️ 散落 1 处(常量比较) | `styles_load.go:237` |
| `theme.go` 内 transparent 特判 | ❌ 无(grep 验证 0 命中) | — |
| `RenderBackgroundLayer` 通过 `color.Color.RGBA()` 接口支持任意格式 | ✅ 已正确 | `styles_load.go:311-323` |
| `ParseColor("transparent")` 返回 `NRGBA{A:0}` | ❌ 未实现 | Phase 1 |
| 嵌套 transparent 校验 | ❌ 未实现 | Phase 2 |

## 4. 实施方案

### Phase 1: `internal/utils/color.go` — 加 `transparent` 关键字

- `ParseColor("transparent")` → 返回 `color.NRGBA{R:0, G:0, B:0, A:0}`
- 在调色板查找、CSS 命名、hex、rgb/rgba、ANSI 索引 **之前** 短路
- 解析层保持完整(8 位 hex / `rgba()` 仍解析出中间 α),中间 α 由消费者侧处理

### Phase 2: `internal/data/config/config.go` — 校验层设界

- `isValidColor`(行 317-319)维持放行所有合法格式;**不**拒绝半透明
- `ValidateTheme`(行 257-295)新增:`palette.Background == "transparent"` 或 `palette.Foreground == "transparent"` → 报错 "无基底可用于半透明混合"
- 保留 `ColorTokenTransparent` token 机制(给 theme/main 用,palette 色直接写 `"transparent"` 也合法)

### Phase 3: `internal/tui/ui/component/styles_load.go` — bg/fg 双路径混合

- 保留 `flattenThemeBackground`(行 235-264),**不**删除;它是加载期混合实现
- 新增 `flattenThemeForeground(theme, ref)` 平行函数:基底换成 `theme.Palette.Foreground`
- `ApplyThemeStyles`(行 172-233)中所有 `Background:` 字段继续走 `flattenThemeBackground`;所有 `Color:` 字段改走 `flattenThemeForeground`(替换原来的 `resolve()`)
- `FormInput` 旁路(行 219)的 `string(theme.Palette.Background)` 保留(它本身是背景通道,逻辑一致)

### Phase 4: `internal/tui/ui/component/styles_load.go` — `buildStyle` 三态分流

`buildStyle`(行 435-454)改为:

```go
applyChannel := func(value string, set func(color.Color)) {
    if value == "" {
        return
    }
    parsed, ok := utils.ParseColor(value)
    if !ok {
        return // 解析失败 = 不设置(保留 lipgloss 默认)
    }
    nrgba, ok := color.NRGBAModel.Convert(parsed).(color.NRGBA)
    if !ok {
        set(parsed)
        return
    }
    switch {
    case nrgba.A == 0:
        return // transparent: 跳过该通道,不调用 set
    case nrgba.A == 0xff:
        set(nrgba) // 不透明: 直通
    default:
        // 半透明: ParseColor 调用方已 flatten 过,这里是 fallback
        set(nrgba)
    }
}
```

实际上 `ApplyThemeStyles` 在加载期已通过 `flattenThemeBackground/flattenThemeForeground` 把所有 `0<α<255` 算成实色,所以 `buildStyle` 实际只会收到 α=0 或 α=255;`default` 分支是防御性的。

### Phase 5: `internal/tui/ui/app/layout.go` — 根背景判断

- `renderAppBackground`(行 252-281)的 `bgHex == "transparent"` 字面量特判**保留**(语义与 V3 一致:transparent 不画)
- 由于 `Palette.Background` 校验已拒绝嵌套 transparent(Phase 2),此处 `bgHex == "transparent"` 实际上是死路径,保留只是兜底
- RGB 取用改走 `ParseColor` 替换 `HexToRGB`(ansi.go:160-172)的局限,以支持 `rgba()` / 命名色 / 8 位 hex(用户输入新格式时不至于渲染成黑色)

### Phase 6: `internal/tui/ui/style/style.go` + `internal/tui/styles.go` — 直用路径对齐

- `style.Color(value)`(行 76-81)维持现状:解析成功返回 `color.Color`,失败兜底 `Colors.Foreground`
- 因为 `tui/styles.go:19-30` 把 12 个调色板色都过一遍 `style.Color(...)`,若某 palette 色是 `"transparent"` 会得到 `NRGBA{A:0}`;`SyncPalette`(行 58-72)注册到 `utils`;后续 `ParseColor` 调色板查找会原样返回 `NRGBA{A:0}`(消费者三态分流处理)
- 检查直用点(component/{config,rows,dialog,helpers}.go、widget/header/header.go、tui/styles.go:36-38)的 `lipgloss.NewStyle().Foreground(c).Background(c)` 调用,确认收到 `NRGBA{A:0}` 时靠 `buildStyle` 同样的三态逻辑保护

### Phase 7: `RenderBackgroundLayer` 与防御

- `styles_loadLayer`(行 311-323)维持现状:`background.RGBA()` 接口已支持任意颜色格式
- Phase 3 已保证 `GetBackground()` 返回的要么是不透明实色、要么是 `nil`(palette 色为 transparent 时),函数内**不**需要新加 α 防御

### Phase 8: 文档同步

- `design/current-design.md` §3.2(行 211-213):改写为 V3 语义(详细条款见 §5)
- `docs/color-transparency-review.md` 标"已被 V3 取代"
- `docs/README.md` 索引若引用旧评审,加 superseded 标注

## 5. 兼容性与验证

### 5.1 向后兼容

- 不透明色路径(`#RRGGBB` / `rgb()` / 命名色 / ANSI / 调色板名)完全不变
- 现有 5 个主题、GetStyle 重构、`styles_load_test.go:62` 断言全部零影响
- `ColorTokenTransparent` 在 theme/main 中的语义不变(`Token == "transparent"` 仍解析为 palette.Foreground 默认或对应 fallback)
- `palette.*` 字段新增 `"transparent"` 字面量写法,与 token 形式并存

### 5.2 新增测试

- `ParseColor("transparent")` 返回 `NRGBA{A:0}`,且 `nrgba.A == 0`
- `flattenThemeBackground` / `flattenThemeForeground` 对 `0<α<255` 算混合,α=0 返回空字符串(调用方 `rawStyles.X.Background == ""` 触发 buildStyle 跳过)
- `ValidateTheme` 拒绝 `Palette.Background == "transparent"` 或 `Palette.Foreground == "transparent"`
- 端到端:`palette.Foreground = "#ff000080"` 时,所有消费者最终传给 lipgloss 的颜色是 `#fefefe` 基底 + 红 α=0.5 混合后的实色 hex
- `test/diagnostics/lipgloss-transparent/`:Phase 0 现有 6 个子测试保留作 lipgloss 行为基线

### 5.3 PTY 验证

- `light` 主题下,所有 `Palette.Background` / `Palette.Foreground` 派生组件的 bg/fg SGR 必须是不透明实色(无半透明 SGR,无 `48;2;0;0;0` 黑画)
- 在 `internal/data/config/defaults/light.jsonc` 中实验性把 `PanelBackground` 改成 `"rgba(254,254,254,0.8)"`,确认渲染 SGR 是混合后的实色 hex

## 6. 涉及文件清单

| 文件 | 改动 |
|---|---|
| `internal/utils/color.go` | ParseColor 添加 `"transparent"` → `NRGBA{A:0}` 短路(Phase 1) |
| `internal/data/config/config.go` | `ValidateTheme` 增加嵌套 transparent 拒绝(Phase 2) |
| `internal/tui/ui/component/styles_load.go` | 新增 `flattenThemeForeground`;`ApplyThemeStyles` 改走双路径(Phase 3);`buildStyle` 三态分流(Phase 4) |
| `internal/tui/ui/app/layout.go` | `renderAppBackground` 改用 ParseColor 替代 `HexToRGB`(Phase 5) |
| `internal/tui/ui/style/style.go` | `Color()` 维持现状,验证直用点安全(Phase 6) |
| `internal/tui/styles.go` | `ApplyTheme` 维持现状,验证 12 色直通(Phase 6) |
| `design/current-design.md` §3.2 | 改写为 V3 颜色语义(Phase 8) |
| `docs/color-transparency-review.md` | 加 header "已被 V3 取代" |

## 7. 不在范围内

- ❌ 真·半透明叠加(终端不支持)
- ❌ 运行时混合(必须加载期算好)
- ❌ 新增 `IsTransparent` 函数(三态靠 α 字段判断已足够)
- ❌ 新增 Palette 字段(基底用现有 `Palette.Background` / `Palette.Foreground`)