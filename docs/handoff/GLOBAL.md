# GLOBAL — 全局信息(决策日志 + 待确认问题)

> 上次更新: 2026-08-04
> 范围: append-only,时间倒序排列

## 决策日志(主模型 → 全体)

### 2026-08-04 — R12 增量清理:Legacy 别名与 caller 全部移除

- **决策**:用户反馈"明显不符合要求"指出 R12 残留 legacy 别名(`ColorGreen = "success"` 等)与 caller(stateicon.go / dialog.go)间接通过别名引用旧色名。决定**完全删除 legacy 别名块**,所有 caller 直接用新语义名。
- **理由**:R12 哲学是"语义名完整覆盖",不应保留 `ColorGreen = "success"` 这种隐式映射——caller 既然已迁移到新名,legacy 别名已无 caller,删除它们让语义完整。
- **改动**(3 文件增量):`internal/constants/color.go` 删除 12 个 legacy 别名块;`internal/tui/ui/component/dialog.go` 4 处改用 `ColorWarning/Primary`;`internal/tui/ui/component/stateicon.go` 7 处改用 `ColorSuccess/Warning/ForegroundMuted/Info/Accent`
- **破坏性**:`ParseColor("white")`、`ParseColor("grey")` 等历史名不再工作——与 R12 §3 一致,解析路径走语义名
- **验证**:`go test ./...` 35 包全过、`go vet ./...` 干净、旧色名常量引用扫描 0 处遗留

### 2026-08-04 — R12 style.Colors 字段重命名为语义名

- **决策**:用户指出"token 命名是颜色名不是作用范围名"的哲学同样适用于 `style.Colors` 启动兜底层。12 字段从颜色名(`Green/Cyan/Blue/Red/Yellow/Orange/Purple/White/Gray/Dark/Surface`)改为语义名(`Success/Primary/Info/Danger/Warning/Accent/AccentSecondary/Foreground/ForegroundMuted/BackgroundSubtle/BackgroundDeep`),`BG` 保留。`internal/constants/color.go` 新增 12 个语义常量,12 个 legacy 别名(`ColorGreen/...`)指语义名(`"success"/"primary"/...`)而非原始字符串(`"green"/"cyan"/...`)。
- **理由**:`style.Colors.Green` 实际存的是 `theme.Palette.Success`(成功语义色),字段名误导。R11 想消除的双标问题不应在启动兜底层残留。
- **改动**(6 文件):`internal/constants/color.go`(12 语义常量 + 12 legacy 别名)+ `internal/tui/ui/style/style.go`(Palette struct 11 字段 + Colors 默认值 + SyncPalette + Color() fallback)+ `internal/tui/styles.go`(ApplyTheme 11 行注入映射)+ `internal/tui/ui/widget/dialog/box.go`(1 处)+ `internal/tui/ui/widget/header/header.go`(3 处 cpuLoadColor)+ `internal/utils/color_test.go`(1 测试改用新名)
- **关键转折**:`constants.ColorGreen = "success"`(语义别名)而非 `"green"`(原始字符串)——保持 `ParseColor` 接受新名同时兼容旧名
- **影响**:所有 widget 通过 `style.Colors.Success/Danger/Warning/...` 读取语义色;legacy 组件(stateicon.go/dialog.go)继续用 `constants.ColorGreen/...` 指向语义槽位;`ParseColor("white"/"grey"/...)` 不再工作(语义哲学完整),但 `ParseColor("foreground"/"foregroundMuted"/...)` 工作
- **验证**:`go test ./...` 35 包全过(含 TestParseColorPaletteAndANSI 改用新名)、`go vet ./...` 干净、旧 `style.Colors.*` 引用扫描 0 处遗留

### 2026-08-04 — R11 主题 token 重命名为作用范围名(方案 C 破坏性修改)

- **决策**:用户讨论后选定方案 C:Q1 重命名不缩写+允许破坏性修改+不遗留兼容;Q2 调色板槽位=语义名;Q3 保留 12 调色板槽位;Q4 不需要独立 semantic 层。12 调色板槽位从颜色名(`green/cyan/blue/red/yellow/orange/purple/white/gray/dark/surface/background`)改为作用范围名(`primary/success/danger/warning/info/accent/accentSecondary/foreground/foregroundMuted/background/backgroundSubtle/backgroundDeep`),`transparent` 保留。
- **理由**:`token: green` 表达的是"颜色=绿"但用户期望"成功语义"——dracula 主题想让成功用紫时改不了,因为硬编码锁死在调色板槽位名上。改为`token: success` 后用户写语义名,dracula 可让`palette.success = "#bd93f9"`(紫)。
- **改动**(~13 文件):constants_domain.go/theme.go/config.go/styles.go/table_config.go + 3 测试文件 + 6 主题 JSONC
- **影响**:JSONC 用户直接看到"success/danger/warning"语义名而非颜色名;主题作者可为每个语义槽位选任意 hex;dracula 可让 success=紫;6 主题 JSONC palette 字段名+所有 token 引用全部重命名
- **破坏性**:6 主题 JSONC 一次性重写,无兼容层
- **保留**:`style.Colors` 12 字段名(`Green/Cyan/...`)作为启动兜底独立命名空间;`FallbackColor*` 12 编译兜底 hex;`FallbackColorTransparent` 字面量"transparent"
- **验证**:`go test ./...` 全过、`go vet ./...` 干净、JSONC 6 主题 47 键 1:1 镜像保持、旧 token 引用扫描 0 处遗留

### 2026-08-04 — 主题透明架构实施(R10,Q1-Q4 全 a)

- **决策**:实施用户讨论确定的 4 点设计:
  - Q1 transparent token:新增 `ColorTokenTransparent` 调色板 token,`Theme.ResolveColor` 返回字面量 `"transparent"`,`isColorRef` validator 接受;新增 `FallbackColorTransparent = "transparent"` 命名常量
  - Q2 弹框主体背景:新增 `theme.Dialog.BodyBackground` 槽位(token + value);`DefaultTheme()` 默认 `TokenRef(ColorTokenBackground)`;6 主题 JSONC 覆盖 `value: <主背景 +5% 加亮 hex>`
  - Q3 终端行为统一:新增 `internal/tui/terminal.go` `SupportsTruecolor()`(读 `COLORTERM`/`TERM`/`TERM_PROGRAM`),`tui.ApplyTheme` 接入 alpha 降级作为未来扩展
  - Q4 transparent 默认范围:6 主题 JSONC 改 4 个主背景槽位为 `token: transparent`
- **改动**(7 文件):constants_domain.go / theme.go / config.go / styles_load.go / box.go(6 改)+ terminal.go(新)+ 6 主题 JSONC
- **理由**:用户明确"绝大多数组件透明/共用背景色"+"弹框/选择框要区分图层";"transparent" 字面量避免空字符串歧义;alpha 降级保证非 truecolor 终端遮罩仍可见
- **影响**:
  - 切主题时 header/footer/toast/panel body 等非弹框组件**真正透明**——只有 layer 1 主背景 + layer 3 强调块 + layer 4 弹框层
  - 弹框打开时遮罩 + 弹框主体 + 弹框标题三层清晰区分(layer 4a/4b/4c)
  - light 主题下 dialog 主体 `#fefefe` 主背景 → `#ffffff` 加亮 5%,视觉上 dialog 与背景有细微分层
- **验证**:`go test ./...` 全过(含 3 个新增测试)、`go vet ./...` 干净、JSONC 6 主题 47 键 1:1 镜像保持
- **未实施**(留待后续):alpha 降级算法在 `tui.ApplyTheme` 中实际应用——已留 `SupportsTruecolor()` 接口,等用户决定降级策略后再接入
- **下次注意**:新增主背景用途槽位时直接用 `token: transparent`;新增强调块独立背景槽位时用具体 hex 或 `token: dark/surface`

### 2026-08-04 — C 类启动兜底保留 + 透明/RGBA JSONC 支持设计留置

- **决策**:用户讨论后明确:"jsonc 配置也应该支持透明，用户根据需要配置 颜色（例如 RGBA）"。本次 R9 决策:**不修改 C 类代码**(24 处启动兜底 hex 保留),将"透明/RGBA JSONC 支持"作为**设计留置**进 proposal §12,待未来实施。
- **理由**:
  - C 类 12 处 `style.Colors` 字面量(`internal/tui/ui/style/style.go:35-48`)是 `init()` 之前 `utils.ParseColor("green")` 解析调色板名的依赖——移除会让 `tui.ApplyTheme` 之前的渲染调用崩溃
  - C 类 12 处 `FallbackColor*` 常量(`internal/data/config/constants_domain.go:139-150`)是 `DefaultTheme()` 的 palette 兜底——移除会让 `theme.Palette` 在极端启动场景下为空
  - 这 24 处兜底与 R8 架构"透明/共用背景"**不冲突**——它们只在启动最早期生效(毫秒级),用户视觉看不到
- **透明/RGBA JSONC 设计要点**(留待未来实施):
  - `#RRGGBBAA` 9 位 hex 已支持(truecolor 终端 alpha 生效,普通终端降级)
  - 新增 `ColorTokenTransparent = "transparent"` token,`Theme.ResolveColor` 返回空字符串表示"不画背景"
  - JSONC 配置示例:`{"token": "transparent"}`、`{"value": "#18191bcc"}`(alpha 80%)
  - 默认哪些槽位透明、哪些实色——按 R8 原则,主背景槽位建议透明,强调块保留实色
- **实施清单**(如果未来启动,见 proposal §12.2.3):
  1. `constants_domain.go` 新增 `ColorTokenTransparent`
  2. `theme.go` `ResolveColor` 增加 transparent 分支
  3. `DefaultTheme()` 给 4 个透明槽位默认 `TokenRef(ColorTokenTransparent)`
  4. JSONC 6 主题的 4 个槽位改 `token: transparent`
  5. 测试加 transparent 断言
- **不动**:本次无代码改动,保留所有 24 处启动兜底。
- **下次注意**:如未来决定实施透明 token,需先确认 lipgloss v2 对 `.Background("")` 的行为(等价于 NoColor 还是 fall back to default)。

### 2026-08-04 — 架构级重设计:"仅主背景"架构(用户截图反馈)

- **反馈**:用户截图显示 R4-R7 后 light 主题下 header 和 footer 仍画独立背景色,看起来"仍是 dark 主题"。明确指引:"一般情况下都不需要单独指定主题色,直接透明 共用背景色即可。只有部分组件需要用于单独设置背景色"。
- **决策**:修改 6 主题 JSONC,4 个主背景槽位从 `token: dark/surface` 改为 `token: background`:
  - `theme.Header.Background` → `token: background`
  - `theme.Footer.ShortcutBackground` → `token: background`
  - `theme.Footer.StatusBackground` → `token: background`
  - `theme.Toast.Background` → `token: background`
- **理由**:绝大多数组件"透明"——让外层 `palette.background` 透过,header/footer/toast 与主背景同色,视觉一致;只有 dialog、selectedRow、keyBadge 等强调块保留独立背景。
- **影响**:light 主题下 header、footer、toast 全部显示 `#fefefe` 主背景色(与 panel body 一致);dark 主题下显示 `#181919b`;其它主题跟随。
- **不动**:palette.dark/surface 槽位(保留为"通用浅深背景色"语义);dialog.overlay(selectedRow、markedRow、keyBadge、detailSelection、logHighlight 等强调块独立背景保留)。
- **验证**:`go test ./...` 全过、`go vet ./...` 干净、JSONC 6 主题 47 键 1:1 镜像保持。
- **下次注意**:新增"主背景"用途的槽位时,直接用 `token: background`;新增"强调背景"用途的槽位(用于 selectedRow/markedRow/toast 之类)才用 `token: dark/surface` 或 `value: hex`。

### 2026-08-04 — Light 主题 B 类全量重构:dialog 30 处显式化

- **决策**:R6 修了"对比度差"13 处;剩 21 处 `style.Colors.{Cyan,Green,Red,Yellow,Orange,White,Purple}` 在 dialog 中直接绑调色板——它们**本身跟主题走**(`tui.ApplyTheme:19-30` 把 `theme.Palette.X` 注入 `style.Colors.X`),不构成"颜色泄露",但代码上**隐式依赖注入绑定**,用户自定义主题时不易追踪。R7 决定**显式化**:把 30 处 `style.Colors.X` 全部替换为 `component.GetStyle(name).GetForeground()`,新增 3 个 token 槽位(`DialogConfirm/DialogError/DialogWarning`)覆盖所有"动作色"。
- **改动**(8 文件,30 处):
  - `internal/tui/ui/component/styles_load.go` —— 加 `DialogConfirm/DialogError/DialogWarning` 字段 + 3 lookup case + 3 投影
  - `internal/tui/ui/widget/dialog/{exec,form,form_popup,choice,selection,notification,view}.go` —— 30 处 `style.Colors.X` 替换 + 删 style import(view.go 增 component import)
- **理由**:消除隐式绑定;用户自定义主题时修改 `theme.Text.Success/Error/Warning` 立刻可见,无需再追 `style.Colors.Green` 这种间接绑定;3 个新槽位用现有 `theme.Text.*` token,不破 JSONC schema。
- **影响**:dialog 包内 `style.Colors.*` 使用数从 30 处 → 0 处;JSONC 6 主题仍 47 键 1:1 镜像。
- **验证**:`go test ./...` 全过、`go vet ./...` 干净、JSONC 镜像保持。
- **下次注意**:如有新增 dialog 类型,继续走 `component.GetStyle(name)` 模式,不引入新的 `style.Colors.X` 绑定。

### 2026-08-04 — Light 主题 B 类修复:对比度差 13 处

- **决策**:R5 完成后用户确认 A 类已修,但切 light 仍能看到"颜色泄露"。进一步盘点发现 B 类问题——虽然 `style.Colors.X` 已在 `tui.ApplyTheme:19-30` 注入主题调色板,但**在 light 主题下,某些 token 解析后的 hex 与浅色背景对比度差**,导致视觉"不可见/看不清"。按"精确修复"原则只改 13 处对比度差位置,不动其它 21 处跟着调色板的 token。
- **改动**(5 文件,13 处):
  - `header.go:163` keyBadge 背景:`style.Colors.Blue` → `theme.Header.Background` 解析值(浅主题对比稳定)
  - `header.go:178` keystroke border 前景:`style.Colors.Blue` → `theme.Main.Title` 解析值
  - `form_popup.go:175,336` highlight 背景:`style.Colors.Surface` → `theme.Header.Background` 解析值(**最关键**——之前 Surface=`#f0f0f0` 与 body `#fefefe` 同色,高亮不可见)
  - `dialog/{choice,exec,selection,form}.go` 7 处 unselected 按钮前景:`style.Colors.Gray` → `theme.Text.Dim` 解析值(从 3.2:1 升到 5.8:1 对比度)
- **影响**:light 主题下 keyBadge 高亮可识别;keystroke border 可见;form popup 高亮行不再与背景同色;dialog 按钮 unselected 文字达到 WCAG AA 4.5:1。
- **未修**(按精确修复原则):21 处 `style.Colors.Cyan/Green/Red/Yellow/Orange/White/Blue/Purple` 跟着调色板走,light 下对比度 OK,不动。
- **验证**:`go test ./...` 全过、`go vet ./...` 干净。

### 2026-08-04 — Light 主题背景一致性 R5:A 类全容器补背景

- **决策**:R4 修复后用户仍报告"light 主题背景仍有黑色泄露"。进一步盘点出 3 类未跟随主题的源头:A 类(主容器无背景)11 处、B 类(组件细节直接绑 `style.Colors.X`)48 处、C 类(theme-hardcoded-migration §6 已知保留项)12 处。用户选"当前先补充 A 类"。一次性补 11 处 A 类未画背景容器,各加 `.Background(style.Colors.BG)`。
- **理由**:A 类是视觉上最显眼的"主容器",B/C 类影响较小或已被 R4 主题化覆盖。先修 A,B/C 留后续。
- **改动**(6 文件,11 处):
  - `internal/tui/ui/app/layout.go`:runtimeSelectorUI 容器 + renderTruncatedLines 行尾容器
  - `internal/tui/ui/widget/panel/panel.go`:Panel.Render body 容器
  - `internal/tui/ui/pages/compose/view.go`:topBar JoinHorizontal 3 列
  - `internal/tui/ui/pages/help/view.go`:help leftBox + rightBox
  - `internal/tui/ui/pages/logs/view.go`:logs body
  - `internal/tui/ui/pages/detail/view.go`:detail body
- **影响**:切 light 主题时上述 11 个主容器全部显示 `#fefefe` 浅色底;切 dark 显示 `#181919b` 深色底;其它主题跟随调色板。
- **未修**(留后续):B 类 48 处 `style.Colors.X` 直接绑、Compose leftBar/topBar 内的 bar 函数、R5 列表外的 dialog 按钮/keystroke 列/query rail/logo 列等。
- **验证**:`go test ./...` 全过、`go vet ./...` 干净、JSONC 6 主题 47 键 1:1 镜像保持。

### 2026-08-04 — Light 主题背景一致性修复

- **决策**:在 light 主题下,action bar / panel / border 容器**不显示任何背景色**——它们没有显式调用 `.Background()`,落在用户的终端默认黑色背景上,造成"亮前景文字 + 黑色容器 + 浅色背景方块"的视觉错位。修复:给所有主容器加 `.Background(style.Colors.BG)`(`BG` = `theme.Palette.Background` 在 styles.go:30 的别名)。
- **理由**:`theme.Palette.Background` 是主题定义好的"主背景色",dark=#18191b、light=#fefefe、nord=#242933。给主容器加这个背景后,切主题时**整屏视觉一致**。与已有的 `layout.Background.Fallthrough` 行为兼容(已在 styles.go:53 处理)。
- **改动**:
  - `internal/tui/styles.go` `ApplyTheme`:3 个 BorderStyle 各加 `.Background(style.Colors.BG)`
  - `internal/tui/ui/widget/actionbar/actionbar.go` `renderBox`:返回 style 加 `.Background(style.Colors.BG)` + import style 包
- **影响**:切 light 主题时所有主容器(panel / action bar / borders)显示 `#fefefe` 浅色底;切 dark 时显示 `#18191b` 深色底;其它主题跟随各自调色板。
- **验证**:`go test ./...` 全过、`go vet ./...` 干净、JSONC 6 主题 47 键 1:1 镜像保持。
- **未修复的容器**(留待后续):dialog 内的 body parts 各 part 自带背景,dialog 框被 `DialogBox` 整体背景覆盖所以视觉一致;`header.go:80` pct 函数、`compose/view.go` 多处无背景但落在 panel 容器内(panel 已有背景)。
- **下次注意**:如果用户报告"还有黑底漏出"现象,搜索 `lipgloss.NewStyle()...Render(` 但无 `.Background(` 的位置即可定位。

### 2026-08-04 — 主题硬编码迁移 Follow-up(shortcutBar 投影 + light 表色)

- **决策**:完成 [proposals/theme-hardcoded-migration.md §7](./proposals/theme-hardcoded-migration.md#7-未来扩展点) 列出的高优先级 follow-up:
  - **真 bug 修复**:`footer.go:39` 与 `keyhint.go:25` 调 `GetStyle("shortcutBar")`,但 `globalStyleRefs.lookup` 未注册——永远走 SafeFallback。在 `internal/tui/ui/component/styles_load.go` 补 3 处:`globalStyleRefs.ShortcutBar` 字段、`ApplyThemeStyles` 投影(`Background: theme.Footer.ShortcutBackground`)、`lookup` case。
  - **light 主题表色调整**:`light.jsonc` 的 `theme.table.columnForeground` 从 `#ECEFF1` 改为 `#1f2328`(`#fefefe` 背景对比度 15.5:1,AAA);`nameForeground` 从 `#FAF0E6` 改为 `#0d1117`(对比度 18.9:1,AAA)。其余 5 主题保留绝对 hex(深背景已足够对比)。
  - **Footer.StatusBackground 调查**:无任何调用点,保持预留槽位。
  - **测试**:`styles_load_test.go` `TestApplyThemeStylesProjectsFixedScopes` 增 1 项 ShortcutBar 背景断言。
- **理由**:switch 主题时 footer keybar 之前永远白色无背景——是真 bug。light 主题表格之前几乎不可读,影响实际使用。
- **影响**:footer keybar 现在显示 `theme.Footer.ShortcutBackground` 决定的深色底;light 主题表格恢复可读性。JSONC 嵌套键仍 1:1 镜像(6 主题均 47 键)。
- **验证**:`go test ./...` 全过、`go vet ./...` 干净。
- **下次注意**:`theme-hardcoded-migration.md §7.4` (新增主题接入指南) 仍未做;若需新增 monokai/gruvbox/tokyo-night 等主题,请参考 `theme-hardcoded-followups.md §1` 引用模式。

### 2026-08-04 — 主题硬编码色全面迁移至配置(13 处)

- **决策**:把组件渲染层 13 处硬编码十六进制字面量(3 处 table_config、5 处 SafeFallback、4 处 dialog overlay、1 处 layout background)下沉为主题配置。新增 2 个作用域 `theme.table.*` (3 键)与 `theme.safeFallback.*` (5 键),dialog overlay 与 layout 改走已有 `theme.dialog.overlay` + `theme.palette.background`。
- **理由**:BR-043 §3.4 主题层叠方案升级后只补了 `default.jsonc` 51 键,5 个非默认主题 + 13 处硬编码仍绕过主题系统,导致切 dark/dracula/solarized 等主题时表格色、被标记行、找不到名字的兜底样式、对话框遮罩、image fallthrough 背景全部不变。
- **改动**(16 文件,详 [proposals/theme-hardcoded-migration.md](./proposals/theme-hardcoded-migration.md)):
  - schema:`theme.go` 加 `TableStyles` / `SafeFallbackStyles` + patch,`constants_domain.go` 加 3 个 `FallbackColorTable*` 命名常量
  - 投影:`component/styles_load.go` 重写 `SafeFallback` 为 `ApplySafeFallback(theme)` 函数,`ApplyThemeStyles` 末尾链式调 `ApplyTableTheme` + `ApplySafeFallback`;`table_config.go` 删 3 个 const,`defaultColumnStyles` 改接受参数
  - 主题文件:6 份 JSONC 各加 8 个键,1:1 镜像 default
  - 硬编码替换:`dialog/{form,notification,selection,exec}.go` 4 处 `overlayColor = "#0d1117cc"` 改 `OverlayColor(nil)`;`app/layout.go:143` 改 `blendColors(string(theme.Palette.Background), c, op)`
  - 测试:`table_test.go:46` 改用 `config.FallbackColorTableNameForeground` 引用;`styles_load_test.go` 增 5 项 SafeFallback / Table 投影断言
- **影响**:
  - 切换任意主题时,表格 / 标记行 / Name 列 / SafeFallback / 对话框遮罩 / image fallthrough 全部跟随调色板
  - `light` 主题的表色仍是绝对 `#ECEFF1`(深背景前景),与浅色背景对比度较弱——**有意保留**,避免 light 主题表格几乎不可读
  - 未来扩展点:在 `light.jsonc` 单独覆盖 `theme.table.*` 3 键即可,无需改代码
- **验证**:`go test ./...` 全过、`go vet ./...` 干净、JSONC 6 主题 47 键 1:1 镜像、13 处字面量扫描全清
- **下次注意**:`footer.go:39` 与 `keyhint.go:25` 调 `GetStyle("shortcutBar")`,但 `globalStyleRefs.lookup` 没注册 `shortcutBar`,**永远走 SafeFallback**——是本次范围外的真 bug,下次修

### 2026-08-03 — BR-043 §3.2 高度修订 + Confirm/Cancel 同行渲染

- **决策**:选择框高度从 panel 等高修订为 panel * 3/4(与宽度对称);Confirm/Cancel 从方案 B 的列表末尾两行改为同行左右排版
- **理由**:用户 2026-08-03 反馈——BR-043 §8.2 原话"等高+宽为 3/4"应解读为"宽 3/4、高 3/4"两轴对称;Confirm/Cancel 同行布局更符合 flex 视觉对齐
- **改动**:
  - `panelDialogHeight`: `bodyH` → `bodyH * 3 / 4`(仍 clamp cfg.MinHeight/MaxHeight)
  - `FormDialog` 渲染:移除两个独立 `renderBtnRow` 调用,合并为单行左 Cancel + 右 Confirm,焦点模型不变(Confirm=n、Cancel=n+1 线性)
- **影响**:所有 `Render*InPanel` 调用(FormDialog/ExecDialog/SelectionDialog)高度变为 panel * 3/4;小窗口(40x16)下 height=12,需确认无挤压
- **验证**:form/dialog/state/keyboard/keys 包测试全过

### 2026-08-03 — BR-043 §3.4 配置拆分实施

- **决策**:把 `internal/data/config/default.jsonc`(189 行,9 个 section)拆分为
  `internal/data/config/styles/*.jsonc` 多文件 + 用户目录覆盖机制
- **拆分结构**(每个分片包一个顶层 section):
  - `general.jsonc` / `ui.jsonc` / `docker.jsonc` / `runtime.jsonc`
  - `logs.jsonc` / `layout.jsonc` / `commandTemplates.jsonc` / `keymap.jsonc`
- **加载机制** (`internal/data/config/loader.go`):
  - `//go:embed styles/*.jsonc` → `StylesFS() fs.FS`
  - `LoadSplit(userDir)` 按字母序加载嵌入分片,合并为 `map[string]json.RawMessage`,
    再应用 `userDir/*.jsonc` 用户覆盖,最终 unmarshal 到 `Config`
  - 加载顺序:嵌入分片 < 用户目录 < CLI YAML(`config.go Load`)
- **路径选择**: `internal/data/config/styles/`(而非原计划 `internal/ui/styles/`),
  与 `config.go` 同包,loader 直接调用 `DefaultConfigJSON`/`StylesFS` 紧耦合
- **向后兼容**:`default.jsonc` 保留未删,`DefaultConfig()` 仍走旧路径;
  `LoadSplit` 与 `DefaultConfigJSON` 等价(regression test 覆盖)
- **用户目录**:`~/.config/docker-tui/styles/*.jsonc`(自动创建,不报错)
- **影响范围**:`internal/data/config/{embed,config,loader,config_test}.go`
  + 8 个 `styles/*.jsonc`
- **验证**:5 个 `TestLoadSplit*` 测试全过 + 全包测试无回归

### 2026-08-03 — BR-043 批次 A-D 实施完成

- **决策**:R2 批次按 GLOBAL.md § 2026-08-03 BR-043 六个问题决策实施
- **交付**:4 个 feat commits
  - `1731789` 批次 A:路径绝对化(blur-time) + 面板尺寸统一(3/4 宽 + 等高)
  - `8ec0e4d` 批次 B:Commit form 条件字段(DependsOn/DependsEq/Hidden)
  - `9e0078f` 批次 C:Action 页面布局重构(方案 B + 共享焦点 + 列表行渲染)
  - `29d8cfe` 批次 D:C 键连接信息迁 `:conn-info` command palette
- **范围**:12 个文件改动(+/- ~440 行)
- **未实施 followup**:3.4 配置拆分、per-form DefaultFocus 配置、DependsValue 枚举扩展
- **验证**:`go vet` + `go test` state/keyboard/keys/dialog 全过;`git diff --check` 通过

### 2026-08-03 — BR-043 六个问题决策

**范围**:`docs/task/br-043-form-action-bar-redesign.md` §7 五项待确认项;3.4 配置拆分不在本轮,后续另起任务卡。

**逐项决策**:

1. **3.1 路径绝对化触发时机 → on field-exit (blur)**,不在每次 keystroke。
   - 理由:per-keystroke 会干扰 Tab/Ctrl+Space 路径补全(`~/` 候选立即被改写为 `/home/...` 丢失补全前缀),且打断光标位置计算。
   - 实现位置:`internal/tui/keyboard/container_form.go` `handleContainerFormKey` 中,Tab/ShiftTab/Down/Up/Enter 离开 FormPath 字段时调用 `state.Absolute(text, m.Form.CWD)` 写回并调整 cursor;提交时 `normalizeSaveDestination` 仍是兜底。
   - 状态:仅读 `internal/tui/state/path.go`,不改。

2. **3.3 Action 页面重构方案 → 方案 B**(FormSpec 加 ConfirmLabel/CancelLabel 字段,渲染为末尾两行)。
   - 理由:方案 A 新增 `FormFieldKind FormConfirm/FormCancel` 会污染 Fields 迭代(需跳过 Confirm/Cancel 数据收集);方案 B 将 Confirm/Cancel 作为末尾两行,与 options 共用同一焦点 `0..n+1`,改动最小,符合用户设计"Confirm box 与 options 共用焦点"。
   - 同步变更:删除 `FormState.OnConfirm` 双语义,焦点统一为 row index;`MoveField`/`MoveButton` 合并为单一线性 Move。
   - 影响文件:`internal/tui/state/form.go` + `internal/tui/ui/widget/dialog/form.go` + `internal/tui/keyboard/container_form.go` + 键盘测试。

3. **3.3 默认焦点 → 保持 Cancel**,per-form 配置留作 follow-up。
   - 理由:安全默认,避免误触 Confirm;per-form `DefaultFocus` 可后续追加(本轮不实施)。

4. **3.5 DependsEq 类型 → bool only**。
   - 理由:当前唯一依赖关系 `archivePath` 依赖 `exportTar`(bool);无枚举依赖需求。
   - 实现:`FormField.DependsOn string` + `FormField.DependsEq bool` + `FormState.RecomputeVisibility()`。
   - 枚举/字符串比较作为 future work 留待 `DependsValue any` 扩展。

5. **3.6 R/C 键归宿**:
   - **R 保留为全局 refresh**(`ActionRefresh`)。理由:R 是标准操作且 README 已文档化(`r Refresh all resources`),非"链接型"快捷键;移除会破坏既有契约。
   - **C 取消连接信息 fallback**:从 `handleShortcuts` 移除 `KeyC → showConnectionInfo`;连接信息迁移到 `:` command palette(`conn-info` 命令)。理由:C 在 volumes/networks 上下文是 CREATE(非 link 型,保留);containers/images/compose 上的 C=连接信息才是 link 型,按用户"取消 table 链接键"要求迁出快捷键层。
   - 影响文件:`internal/tui/keyboard/keyboard.go` + command palette 注册表 + README/docs 同步。

**实施批次**(本轮范围):
| 批次 | 内容 | 风险 |
|---|---|---|
| A | 3.1 路径绝对化 + 3.2 尺寸统一(panel 3/4 宽 + 等高) | 低 |
| B | 3.5 Commit form 条件字段(DependsOn/DependsEq/Hidden) | 中 |
| C | 3.3 Action 页面布局重构(方案 B + 共享焦点) | 中高 |
| D | 3.6 C 键连接信息迁 command palette | 低 |

**不在本轮**:3.4 配置拆分(后续另起任务卡);per-form DefaultFocus 配置;DependsValue 枚举扩展。

### 2026-08-03 — R1 批次主模型审查与 commit

- 决策:R1 文档批次审查通过并 commit;审查中修复 5 处悬空 task 引用 + 13 处跨组断链
- 理由:子模型自报"残留引用 0"与事实不符 — 实查 `grep -rn` 发现 `docs/task/README.md`、br-039-a、br-039-code-location、br-040-code-location、br-041 共 5 处指向已删 task 的链接;另修复 requirement 文档跨组链接缺 `R03-form-action/` 前缀、`docs/feature-design.md` 冗余前缀、R04 相对深度错误
- 验证:`python3` 全仓 md 链接校验 479 个内部链接全部可解析(`internal/*` 代码锚点 109 个为 bugfix 文档约定,非断链);`git diff --check` 通过
- 影响范围:`docs/` 14 个文件链接修复 + handoff 状态更新

### 2026-08-03 — handoff 目录建立

- 决策:新建 `docs/handoff/` 目录,采用 GLOBAL/STATE/COMPLETED 三文件 + README 索引
- 理由:`ai-prompts.md` 是会话内粘贴的模板,跨会话持续状态/决策/问题需要独立文档
- 影响范围:仅新建,不动其他文件

### 2026-08-03 — 删除 6 个冗余 task 文件

- 决策:删除 `next-small-model-brief.md` + `br-042-form-runtime-refresh-corrections.md` + `br-036-image-import-tarball.md` + `br-037-registry-login.md` + `br-038-exec-shell-ui.md` + `br-040-dialog-style-center-on-panel.md`
- 理由:内容已合并入 `requirement/R##-##/` 与 `constraint/C##.md`
- 验证:`grep -rn <删文件名> docs/` 残留引用为 0;`git diff --check` 通过
- 影响范围:13 个反向引用文件已修复链接

### 2026-08-02 — 文档真理源重组

- 决策:按 [docs-architecture.md](../docs-architecture.md) 建立需求树 `requirement/R##/` + 横切约束 `constraint/C##/`
- 理由:旧 task/ 散乱 BR 文档无法支持"按大需求点查询";requirement/ + constraint/ 提供清晰的真理源分层
- 影响范围:新建 25 个文档,更新 3 个入口索引,保留全部旧 task/ 作为迁移来源(后删除冗余)

### 2026-08-01 — BR-034 Image History 顶层页

- 决策:Images 页 H 键 / Action Bar 触发独立 History 顶层页
- 理由:大镜像 >50 layer 滚动卡顿;独立页配 EnsureVisible + 可见行切片解决
- 状态:implementing(详见 [requirement/R02-image/R02-01-history.md](../requirement/R02-image/R02-01-history.md))

## 待确认问题(子模型 → 主模型)

> 状态: open(子模型提的) / answered(主模型答完)

### 2026-08-03 — BR-043 §3.4 配置拆分方案审查

详见 [`proposals/config-split-review.md`](./proposals/config-split-review.md)。

3 项决策待主模型裁定：
1. **P1**：字段级深合并（替换当前浅合并）
2. **P2**：Dialog 配置搬到 `internal/data/config/styles/dialog.jsonc`
3. **N/A**：YAML+JSONC 双格式 / 主题独立 / 配置版本 — 确认保持现状

主模型响应后子模型实施，无需主模型参与代码。

### 2026-08-03 — handoff 命名 (状态: open)

- 问题:子模型问:目录名 `handoff/` 是否合适?或用 `coordination/` / `comm/`?
- 答复:见 [README.md 待确认 § 命名](./README.md#待确认)
- 影响范围:仅命名,不动结构

### 2026-08-02 — BR-041 状态标记 (状态: answered)

- 问题:子模型问:`requirement/R03-form-action/R03-01-form-pattern.md` 状态写 `implementing` 还是 `implemented`?旧 br-041 task 文档说"第一阶段完成",但 [task/br-043-form-action-bar-redesign.md](../task/br-043-form-action-bar-redesign.md) 列出 6 项未实施项
- 答复:状态保持 `implementing`(第一阶段完成 + 仍有 followup);R03-01 README 待确认项已说明此差异
- 影响范围:无