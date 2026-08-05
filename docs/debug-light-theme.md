> ⚠️ **本 debug 文档针对的 #656565 bug 已被 [V3 颜色透明方案](color-transparency-plan.md) 消除**(2026-08-05)。
>
> V3 通过将半透明色在加载期预混合到 `Palette.Foreground`/`Palette.Background` 解决了原 101,101,101 混合产物;本文件保留作为历史 debug 方法论参考。

# Light 主题背景色 Debug 验证方案

> 用途: 一键复制所有 debug 命令,定位 "header 右侧 / panel / table 仍是黑色" 的根因。
> 运行顺序: Level 1 → Level 2 → Level 3 → Level 4,每级失败就定位到具体层级。

---

## Level 1: 终端是否支持 24-bit 背景色

**这一步决定问题在终端还是代码**。`$TERM=xterm-256color` 只保证 256 色,不保证 24-bit(`\x1b[48;2;...`)。

```bash
# Test 1A: 24-bit 背景(关键)
printf '\x1b[48;2;254;254;254m test \x1b[0m|\n'
# 期望: 'test' 背景是浅色,光标 '|' 在浅色之后
# 如果 'test' 背景是终端默认色(深)→ 终端不支持 24-bit 背景

# Test 1B: 256 色背景(兜底)
printf '\x1b[48;5;231m test \x1b[0m|\n'
# 期望: 'test' 背景是浅色(231 = 255,255,255 灰阶)
# 如果深 → 终端连 256 色背景都不支持

# Test 1C: 16 色背景
printf '\x1b[47m test \x1b[0m|\n'
# 期望: 'test' 背景是浅色
# 如果深 → 终端完全不支持 ANSI 背景,问题不在我们
```

**根据 Test 1A 结果分支**:
- **1A 浅色**: 终端 OK → 跳到 Level 2 排查代码
- **1A 深色 + 1B 浅色**: 终端不支持 24-bit 背景但支持 256 色 → 改 `renderAppBackground` 用 256 色
- **1A + 1B 都深 + 1C 浅色**: 终端只支持 16 色 → 改用 16 色
- **全部深**: 终端问题,检查 `$TERM` 和终端设置

---

## Level 2: `renderAppBackground` 实际输出字节

**确认函数输出**是否含 `bgAnsi` 序列 + 每行 padding 长度。

**已在 `renderAppBackground` 加了 print 全部 `result` 的版本**。运行:

```bash
go build -o dist/dtui-debug ./cmd/docker-tui
./dist/dtui-debug --theme light 2>/tmp/light.stderr
# 退出后(随便按个键退):
cat /tmp/light.stderr | head -100
```

**预期 stderr 格式**(每帧一次):
```
DEBUG renderAppBackground: bgHex="#fefefe" viewport=WxH
DEBUG renderAppBackground: input lines=N
DEBUG line 0: visW=?? pad=?? (line preview: ...)
...
DEBUG line 0 RESULT (last 200 chars): "..." ← 实际 ANSI 字节
```

**关键看**:
- `bgHex` 是否 `#fefefe`(light 主题应如此)
- `visW` 是否 ≈ `viewport.Width`(差几是 padding)
- `RESULT` 里**是否含 `\x1b[48;2;254;254;254m`**(`bgAnsi` 的字节序列)
- 每行最后一个字符前**是否有 `bgAnsi`**

如果 `bgAnsi` 不在 result 里 → `renderAppBackground` 没正确 prepend/append。
如果 `bgAnsi` 在但终端还是黑 → 终端不识别(回 Level 1)。

---

## Level 3: 二进制是否最新

**防止跑了旧二进制,所有 fix 都看不到效果**。

```bash
# 看源码最新修改时间
stat -c "%Y %n" internal/tui/ui/app/layout.go

# 看二进制最新构建时间
stat -c "%Y %n" dist/dtui-debug

# 二进制应该 >= 源码
# 如果二进制 < 源码,需要重建:
go build -o dist/dtui-debug ./cmd/docker-tui
```

**确认用 `dist/dtui-debug` 跑,不要用 just / go run**(可能跑不同的二进制)。

---

## Level 4: 主题 token 是否解析到 `#fefefe`

**确认 `Palette.Background` 在加载后是 `#fefefe`,不是被某处覆盖**。

**单元测试**(已写好,直接跑):
```bash
go test ./internal/data/config/ -run "TestLightThemeBackgroundLoaded|TestLightHeaderBackgroundResolvesToPalette|TestLightMainPanelBackgroundResolvesToPalette" -count=1 -v
```

**预期**:
```
=== RUN   TestLightThemeBackgroundLoaded
--- PASS: TestLightThemeBackgroundLoaded (0.00s)
=== RUN   TestLightHeaderBackgroundResolvesToPalette
--- PASS: TestLightHeaderBackgroundResolvesToPalette (0.00s)
=== RUN   TestLightMainPanelBackgroundResolvesToPalette
--- PASS: TestLightMainPanelBackgroundResolvesToPalette (0.00s)
PASS
```

如果失败 → 主题加载层有 bug,需要修 `light.jsonc` 或 `LoadResolved`。

---

## Level 5: 代码侧最后一招 — 确认 `result` 没被下游修改

**确认 `renderAppBackground` 输出后,没有其他函数在它上面再处理**。

在 `layout.go:RenderApp` 里 `result = renderAppBackground(m, result)` 之后**立刻**加打印(不是等 Level 2 全跑完):

```go
result = renderAppBackground(m, result)
fmt.Fprintf(os.Stderr, "DEBUG POST-renderAppBackground (len=%d): %.200q\n", len(result), result)
return result
```

跑:

```bash
go build -o dist/dtui-debug ./cmd/docker-tui
./dist/dtui-debug --theme light 2>/tmp/light6.stderr
head -30 /tmp/light6.stderr
```

**关键看**: `DEBUG POST-renderAppBackground` 这一行的字符串**最右 50 字符**是否含 `bgAnsi + 一个空格 + bgAnsi` 模式。

- 如果含 → `renderAppBackground` 输出正确,问题在终端/下游
- 如果不含 → 函数输出有 bug

---

## 完整一键脚本

**最全的验证** — 一次跑完,贴所有输出给我:

```bash
# 0. 重建
go build -o dist/dtui-debug ./cmd/docker-tui

# 1. 终端 24-bit 背景测试
printf '\x1b[48;2;254;254;254m test \x1b[0m|\n' > /tmp/term-test.out
cat /tmp/term-test.out

# 2. 跑 light TUI 抓 stderr
timeout 3 ./dist/dtui-debug --theme light 2>/tmp/light-full.stderr
# 退出会卡 timeout 终止

# 3. 主题 token 解析测试
go test ./internal/data/config/ -run "TestLight" -count=1 -v 2>&1 | head -20

# 4. 汇总输出
echo "===TERM TEST===" && cat /tmp/term-test.out
echo "===LIGHT STDERR (first 60 lines)===" && head -60 /tmp/light-full.stderr
echo "===LIGHT TOKEN TEST===" && go test ./internal/data/config/ -run "TestLight" -count=1 -v 2>&1 | tail -20
```

**把整个输出贴给我**,我能直接定位到:
- 终端不支持 24-bit 背景 → 改用 256 色
- 二进制是旧的 → 重建
- 主题 token 没解析对 → 修 `light.jsonc`
- `renderAppBackground` 输出有 bug → 修 `renderAppBackground`
- 输出对但被下游吃 → 修下游(哪个具体函数)

---

## 备用方案:如果 Level 1 确认终端不支持 24-bit 背景

**直接改用 256 色**(`renderAppBackground` 一行改动):

```go
// 24-bit 版本(当前):
bgAnsi := fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)

// 256 色版本(改这一行):
// #fefefe → 256 色灰阶 255 = #eeeeee,接近
bgAnsi := "\033[48;5;255m"
```

**16 色版本**(再降一级):
```go
bgAnsi := "\033[47m" // 47 = white background in 16-color ANSI
```

如果这些 fallback 让右边变成浅色,问题就是终端的 24-bit 背景不支持,跟代码无关。

---

## 现状文档化

完整的修复方案已写入 `docs/theme-background-fix.md`(150+ 行),包括:
- 5 Phase 实施
- 三方向 padding 兜底原理
- light 主题配色全浅验证
- 改动文件清单
- 用户验证清单

如果这次 debug 能定位问题,我会把发现更新到该文档。
