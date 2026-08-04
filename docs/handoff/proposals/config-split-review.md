# BR-043 强类型配置与主题层叠方案

> 日期: 2026-08-04
> 状态: 已实施，待最终环境验证
> 关联需求: BR-043 §3.4
> 设计约束: 不兼容旧配置，只保留目标架构

## 1. 设计结论

配置系统采用“固定作用域 + 固定属性 + 强类型 token + 显式 patch”的模型。

核心规则：

1. 用户配置优先于应用主题。
2. 应用主题优先于应用内嵌静态视觉配置。
3. 应用内嵌 JSONC 优先于 Go 编译期 fallback。
4. 主题只能修改 UI 视觉属性，不能修改行为配置。
5. 每项配置都必须映射到明确的 Go struct 字段。
6. 禁止使用 `map[string]string`、`map[string]any` 或任意字符串 selector 表达配置作用域。
7. 禁止使用字符串数组表达复合样式，例如 `["green", "bold"]`。
8. slice 仅用于领域本身确实是可变集合的数据，例如远程 runtime connections。
9. 用户 patch 中每个叶子字段必须能区分“未声明”和“显式零值”。
10. 最终运行时配置必须是完整、无指针、已验证的 resolved struct。
11. 生产代码中的语义字符串不得以内联字符串面量出现，必须定义为命名常量或强类型常量。

## 2. CSS 参考模型

本方案借鉴 CSS 的四个概念，但不实现 CSS 语法或开放 selector：

```text
Scope       固定组件作用域，例如 header、main、dialog
Property    固定视觉属性，例如 foreground、background、bold
Token       可复用颜色变量，例如 palette.cyan
Cascade     低优先级值被高优先级显式属性覆盖
```

配置作用域由 Go 类型定义，不接受运行时创建的任意作用域：

```text
theme.palette
theme.root
theme.header
theme.main
theme.footer
theme.dialog
theme.toast
theme.text
```

这意味着主题文件即使包含 `runtime` 或 `keymap` 字段，也会在严格解析阶段失败，而不是加载后再通过白名单过滤。

## 3. 独立优先级链

应用行为配置、主题选择和主题内容使用三条独立链，不能表达成对所有字段生效的一条通用链。

### 3.1 应用配置

从低到高：

```text
Go AppConfig fallback
  -> embedded defaults/*.jsonc
  -> user config.yml
  -> CLI 显式行为参数
```

适用字段：runtime、docker、logs、keymap、commands、layout、UI 行为等。

### 3.2 主题选择

从低到高：

```text
embedded app default theme name
  -> user config.yml appearance.theme
  -> CLI --theme
```

`--theme` 只选择基础主题，不修改主题的具体视觉属性。

### 3.3 主题内容

从低到高：

```text
Go Theme fallback
  -> embedded themes/default.jsonc
  -> selected embedded or user theme
  -> user config.yml appearance.overrides
```

用户视觉 override 始终高于所选主题。即使 CLI 选择了另一个基础主题，用户显式覆盖的颜色仍然生效。

## 4. 文件布局

目标目录：

```text
internal/data/config/
  defaults/
    general.jsonc
    ui.jsonc
    docker.jsonc
    runtime.jsonc
    logs.jsonc
    layout.jsonc
    commands.jsonc
    keymap.jsonc
  themes/
    default.jsonc
    dark.jsonc
    light.jsonc
    nord.jsonc
    dracula.jsonc
    solarized.jsonc

~/.config/docker-tui/
  config.yml
  themes/
    <custom-theme>.jsonc
```

规则：

- `defaults/*.jsonc` 只包含应用配置 patch。
- `themes/*.jsonc` 只包含 `ThemeDocument`。
- 用户应用配置只使用一个 `config.yml`，不再提供同 schema 的用户 JSONC 分片。
- 用户自定义主题使用 JSONC，与内嵌主题采用同一 schema。
- 删除语义不准确的 `styles/` 配置目录。
- 删除与 `AppConfig` 或 `Theme` 竞争的组件私有配置。仅内部使用的结构化静态资源（例如表格列定义、边框字符集）可以保留 JSONC，但必须严格解码到固定 struct，不得包含任意 map key，也不得作为用户覆盖入口。

## 5. 顶层数据模型

加载输入、最终应用配置和最终主题是不同类型。

```go
type Resolved struct {
    App   AppConfig
    Theme Theme
}

type LoadOptions struct {
    ConfigPath string
    ThemeName  string
    Host       string
    Podman     bool
    Lang       string
}

func LoadResolved(options LoadOptions) (*Resolved, error)
```

只保留 `LoadResolved()` 作为公开加载入口。删除 `Load()`、`LoadSplit()` 等会返回部分结果或在 main 中继续拼装优先级的入口。

### 5.1 最终应用配置

```go
type AppConfig struct {
    Version  int
    General  GeneralConfig
    UI       UIConfig
    Docker   DockerConfig
    Runtime  RuntimeConfig
    Keymap   KeymapConfig
    Logs     LogsConfig
    Layout   LayoutConfig
    Commands CommandConfig
}
```

`AppConfig` 不包含颜色、主题对象或 theme override。业务代码可以直接消费该对象中的每一个字段。

### 5.2 用户配置文档

```go
type UserConfig struct {
    Version    int
    App        AppPatch
    Appearance AppearanceConfig
}

type AppearanceConfig struct {
    Theme     string
    Overrides ThemePatch
}
```

推荐 YAML：

```yaml
version: 1

app:
  general:
    lang: zh
  runtime:
    health:
      interval: 5s

appearance:
  theme: nord
  overrides:
    palette:
      cyan: "#00ffff"
    dialog:
      border:
        token: cyan
      overlay:
        value: "#101114cc"
      overlayOpacity: 80
```

用户文件中的 `app` 与 `appearance` 是互斥职责的顶层作用域。视觉属性不能出现在 `app` 中，行为属性不能出现在 `appearance` 中。

## 6. 应用配置类型

所有固定配置项使用命名 struct，不使用自由 key map。

### 6.1 General

```go
type GeneralConfig struct {
    ScrollHeight int
    Reporting    ReportingMode
    Language     Language
    SizeFormat   SizeFormat
}
```

`ReportingMode`、`Language`、`SizeFormat` 使用命名 enum 类型并执行值域校验。

所有 enum 值必须定义为强类型常量：

```go
type Language string

const (
    LanguageEnglish Language = "en"
    LanguageChinese Language = "zh"
)

type SizeFormat string

const (
    SizeFormatBinary SizeFormat = "binary"
    SizeFormatSI     SizeFormat = "si"
)
```

### 6.2 Runtime

```go
type RuntimeConfig struct {
    DefaultConnection string
    Discovery         RuntimeDiscoveryConfig
    Health            RuntimeHealthConfig
    Connections       []RuntimeConnection
}

type RuntimeDiscoveryConfig struct {
    LocalDocker bool
    LocalPodman bool
}

type RuntimeHealthConfig struct {
    Interval         time.Duration
    Timeout          time.Duration
    FailureThreshold int
}

type RuntimeConnection struct {
    Name       string
    Driver     RuntimeDriver
    Endpoint   string
    APIVersion string
    TLS        RuntimeTLSConfig
}
```

`Connections` 保留 slice，因为它是用户可定义数量的真实实体集合。元素必须是强类型 `RuntimeConnection`，并按 connection name 做唯一性校验；不使用 `map[string]any`。

### 6.3 Commands

删除：

```go
map[string]string
```

改为固定命令作用域：

```go
type CommandConfig struct {
    DockerCompose DockerComposeCommand
}

type DockerComposeCommand struct {
    Executable string
    Subcommand string
}
```

`Executable` 例如 `docker`，`Subcommand` 例如 `compose`。如未来新增命令，先为该领域新增专用 struct 和明确字段并同步 schema，而不是增加通用参数数组或允许用户注入任意命令 key。

### 6.4 Keymap

删除每个 action 的 `[]string`，改成固定 binding：

```go
type KeyBinding struct {
    Primary   Key
    Secondary Key
}

type KeymapConfig struct {
    Global    GlobalKeymap
    Container ContainerKeymap
    Image     ImageKeymap
    Volume    VolumeKeymap
    Network   NetworkKeymap
    Navigation NavigationKeymap
}

type GlobalKeymap struct {
    Quit      KeyBinding
    ActionBar KeyBinding
    Help      KeyBinding
    Filter    KeyBinding
    Refresh   KeyBinding
}
```

其他领域按相同方式定义明确字段。`Key` 是经过解析和校验的值对象，不在业务层继续传递任意字符串。

### 6.5 UI 与 Dialog 行为

```go
type UIConfig struct {
    WrapMainPanel     bool
    ReturnImmediately bool
    ShowHelp          bool
    HintTimeout       time.Duration
    EnableMouse       bool
    Dialog            DialogLayoutConfig
}

type DialogLayoutConfig struct {
    Width    ResponsiveSize
    Height   ResponsiveSize
    Position Position
}

type ResponsiveSize struct {
    Percent int
    Min     int
    Max     int
}

type Position struct {
    Horizontal HorizontalAlignment
    Vertical   VerticalAlignment
    OffsetX    int
    OffsetY    int
}
```

Dialog 的确认键和取消键进入 `KeymapConfig`。Dialog 的颜色和透明度进入 `Theme.Dialog`。删除组件私有 `dialog.jsonc`。

## 7. 主题类型

### 7.1 Theme document

```go
type ThemeDocument struct {
    Meta  ThemeMetadata
    Theme ThemePatch
}

type ThemeMetadata struct {
    Name        string
    Description string
}
```

主题文件本身是 patch：`default.jsonc` 应完整定义产品默认视觉；其他主题可以只声明与 default 不同的属性。

### 7.2 Resolved Theme

```go
type Theme struct {
    Palette Palette
    Border  BorderStyles
    Header  HeaderStyles
    Main    MainStyles
    Footer  FooterStyles
    Dialog  DialogStyles
    Toast   ToastStyles
    Text    TextStyles
}
```

作用域固定，不支持字符串 selector、任意组件名或 `map[string]Style`。

### 7.3 Palette

```go
type Palette struct {
    Green      Color
    Cyan       Color
    Blue       Color
    Red        Color
    Yellow     Color
    Orange     Color
    Purple     Color
    White      Color
    Gray       Color
    Dark       Color
    Surface    Color
    Background Color
}
```

调色板 token 是封闭集合。新增 token 必须修改 struct、schema、default theme 和测试。

### 7.4 基础视觉属性

```go
type ColorRef struct {
    Token ColorToken
    Value Color
}

```

`ColorRef` 必须恰好使用 token 或具体颜色值之一。配置写法分别为 `{ "token": "cyan" }` 和 `{ "value": "#00ffff" }`。`Color` 在解析时验证十六进制格式，业务层不处理未验证字符串。

### 7.5 组件视觉作用域

```go
type BorderStyles struct {
    Active   ColorRef
    Inactive ColorRef
    Focused  ColorRef
    Kind     BorderKind
}

type HeaderStyles struct {
    Background ColorRef
    Label      ColorRef
    Value      ColorRef
    Logo       ColorRef
}

type MainStyles struct {
    BorderActive   ColorRef
    BorderInactive ColorRef
    Title          ColorRef
    TableHeader    ColorRef
    RowSelected    ColorRef
    RowText        ColorRef
    Footer         ColorRef
}

type DialogStyles struct {
    Border         ColorRef
    Title          ColorRef
    Body           ColorRef
    OptionActive   ColorRef
    OptionInactive ColorRef
    Overlay        ColorRef
    OverlayOpacity uint8
}
```

Footer、Toast 和通用文本语义按相同原则定义专用 struct。即使字段结构相似，也保留语义明确的组件类型，不使用任意 selector 或 `map[string]Style`。

## 8. 字符串常量策略

配置系统生产代码中的字符串分为四类，均不得以无名称的语义面量散落在实现中。

### 8.1 领域枚举

运行时驱动、语言、尺寸格式、边框类型、颜色 token、对齐方式、reporting mode 和背景类型都使用命名类型与常量：

```go
type RuntimeDriver string

const (
    RuntimeDriverDocker RuntimeDriver = "docker"
    RuntimeDriverPodman RuntimeDriver = "podman"
)

type BorderKind string

const (
    BorderKindNone    BorderKind = "none"
    BorderKindNormal  BorderKind = "normal"
    BorderKindRounded BorderKind = "rounded"
    BorderKindDouble  BorderKind = "double"
)

type ColorToken string

const (
    ColorTokenGreen      ColorToken = "green"
    ColorTokenCyan       ColorToken = "cyan"
    ColorTokenBlue       ColorToken = "blue"
    ColorTokenRed        ColorToken = "red"
    ColorTokenYellow     ColorToken = "yellow"
    ColorTokenOrange     ColorToken = "orange"
    ColorTokenPurple     ColorToken = "purple"
    ColorTokenWhite      ColorToken = "white"
    ColorTokenGray       ColorToken = "gray"
    ColorTokenDark       ColorToken = "dark"
    ColorTokenSurface    ColorToken = "surface"
    ColorTokenBackground ColorToken = "background"
)
```

解析边界负责把用户字符串转换成领域类型。转换完成后，业务代码只能和强类型常量比较，禁止出现 `driver == "docker"` 之类的判断。

### 8.2 文件名与路径段

配置文件名、目录名、扩展名和内嵌资源名定义为包内常量：

```go
const (
    configDirName       = "docker-tui"
    configFileName      = "config.yml"
    userThemesDirName   = "themes"
    defaultThemeName    = "default"
    jsoncExtension      = ".jsonc"
    embeddedDefaultsDir = "defaults"
    embeddedThemesDir   = "themes"
)

const (
    defaultGeneralFile  = "general.jsonc"
    defaultUIFile       = "ui.jsonc"
    defaultDockerFile   = "docker.jsonc"
    defaultRuntimeFile  = "runtime.jsonc"
    defaultLogsFile     = "logs.jsonc"
    defaultLayoutFile   = "layout.jsonc"
    defaultCommandsFile = "commands.jsonc"
    defaultKeymapFile   = "keymap.jsonc"
)
```

路径通过 `filepath.Join()` 和这些常量构建，不在 loader 中重复书写路径字符串。

### 8.3 默认值与固定命令

Go fallback 中出现的主题名、颜色值、默认 endpoint、默认 connection name 和固定命令片段全部定义为常量。相关常量靠近其领域类型，不建立包含所有字符串的全局 constants 文件。

```go
const (
    fallbackThemeName       = "Default"
    fallbackColorBackground = "#0d1117"
    localDockerConnection   = "local-docker"
    localPodmanConnection   = "local-podman"
    dockerExecutable        = "docker"
    dockerComposeSubcommand = "compose"
)
```

产品正常默认值仍由 embedded JSONC 提供；这些 Go 常量只用于 fallback。JSONC/YAML 中的字符串属于配置数据，不要求复制成 Go 常量。

### 8.4 错误与日志标识

可被程序判断、测试断言或外部消费的错误码和日志事件名必须定义为常量。面向用户的格式化错误模板也定义为局部常量，避免同一语义出现多种文本：

```go
const (
    errorCodeUnknownField = "config.unknown_field"
    errorCodeInvalidTheme = "config.invalid_theme"
    eventConfigLoaded     = "config.loaded"
)

const errFormatUnknownField = "unknown configuration field %q in %s"
```

错误中的用户输入、文件路径和字段值是运行时数据，不属于字符串面量。

### 8.5 必要例外

以下位置由 Go 语法或序列化机制要求必须使用字符串字面量，不能替换成常量：

- struct tag，例如 ``json:"theme" yaml:"theme"``。
- `//go:embed` pattern。
- 编译器要求为常量表达式但不接受标识符替换的语法位置。

测试 fixture、JSONC/YAML 示例和用户配置内容属于输入数据，可以使用字符串字面量。除此之外，配置模块生产代码不保留裸语义字符串。

## 9. Patch 类型

不能直接把用户配置解码到 resolved struct 后通过零值猜测字段是否出现。每个可覆盖叶子属性都要使用指针表示 presence：

```go
type TextStylePatch struct {
    Foreground *ColorRef
    Background *ColorRef
    Bold       *bool
    Italic     *bool
    Underline  *bool
    Dim        *bool
}

type BorderStylePatch struct {
    Color *ColorRef
    Kind  *BorderKind
}

type DialogStylePatch struct {
    Overlay   *OverlayStylePatch
    Container *BoxStylePatch
    Title     *TextStylePatch
    Body      *TextStylePatch
    Option    *DialogOptionStylePatch
}

type ThemePatch struct {
    Palette *PalettePatch
    Root    *RootStylePatch
    Header  *HeaderStylePatch
    Main    *MainStylePatch
    Footer  *FooterStylePatch
    Dialog  *DialogStylePatch
    Toast   *ToastStylePatch
    Text    *TextStylesPatch
}
```

应用配置使用同样模式定义 `AppPatch`：

```go
type AppPatch struct {
    General  *GeneralPatch
    UI       *UIPatch
    Docker   *DockerPatch
    Runtime  *RuntimePatch
    Keymap   *KeymapPatch
    Logs     *LogsPatch
    Layout   *LayoutPatch
    Commands *CommandPatch
}
```

patch struct 虽然字段较多，但它使覆盖语义完全由类型表达，避免反射式 deep merge、`map[string]any` 和零值歧义。

## 10. Patch 应用规则

每种 patch 提供显式应用函数：

```go
func (p AppPatch) Apply(target *AppConfig)
func (p ThemePatch) Apply(target *Theme)
func (p DialogStylePatch) Apply(target *DialogStyle)
```

规则：

1. patch 字段为 `nil`：继承目标当前值。
2. patch 字段非 `nil`：使用其值，包括 `false`、`0` 和空字符串。
3. 嵌套 patch：调用对应类型的 `Apply()`。
4. 固定 struct：逐字段覆盖，不使用反射和 map。
5. 真实集合：整体替换，不隐式追加或按索引合并。
6. 用户配置中的 `null` 一律拒绝，不将其解释为删除或继承。
7. 每层解析时拒绝未知字段。
8. 所有层应用完成后只执行一次最终业务校验。

`Apply()` 建议由简单、可审查的手写代码实现。若字段数量增长明显，可以使用 `go generate` 根据 resolved/patch 类型生成，但生成产物仍必须是静态字段访问，不能在运行时使用反射。

## 11. 加载算法

### 11.1 应用配置

```text
1. app = FallbackAppConfig()
2. 按固定清单读取 embedded defaults/*.jsonc
3. 每个文件严格解析为 AppPatch
4. 依次 patch.Apply(&app)
5. 严格解析用户 config.yml 为 UserConfig
6. user.App.Apply(&app)
7. 将 CLI 显式行为参数转换为 AppPatch 并应用
8. ValidateApp(app)
```

内嵌默认文件使用代码中的固定顺序，不依赖文件名字母排序：

```go
var embeddedDefaultFiles = [...]string{
    "general.jsonc",
    "ui.jsonc",
    "docker.jsonc",
    "runtime.jsonc",
    "logs.jsonc",
    "layout.jsonc",
    "commands.jsonc",
    "keymap.jsonc",
}
```

原则上每个字段只能由一个内嵌默认文件拥有。测试检查内嵌文件之间不存在重复字段。

### 11.2 主题选择

```text
1. themeName = embedded application default
2. 如果 user.appearance.theme 非空，覆盖 themeName
3. 如果 CLI --theme 显式出现，覆盖 themeName
```

主题查找顺序：

```text
用户 ~/.config/docker-tui/themes/<name>.jsonc
  > 内嵌 themes/<name>.jsonc
```

用户主题文件与内嵌主题同名时，用户文件胜出。

### 11.3 主题内容

```text
1. theme = FallbackTheme()
2. 读取 embedded themes/default.jsonc 并应用
3. 解析 themeName 对应的主题 patch 并应用
4. user.Appearance.Overrides.Apply(&theme)
5. ValidateTheme(theme)
```

第 3 步的精确规则：

- `themeName == "default"` 且不存在用户 `default.jsonc`：不重复应用内嵌 default。
- `themeName == "default"` 且存在用户 `default.jsonc`：在内嵌 default 之上应用用户 default patch。
- `themeName != "default"`：优先应用同名用户主题，否则应用同名内嵌主题。
- 找不到选中主题：返回错误。

### 11.4 错误策略

- 用户配置不存在：使用默认值启动。
- 用户配置存在但无法读取、解析或校验：返回错误并终止启动。
- 用户显式选择的主题不存在或无效：返回错误并终止启动。
- 内嵌 default JSONC 无效：自动化测试必须失败；运行时记录 warning 并保留对应 Go fallback。
- 显式选中的内嵌主题无效：返回错误并终止启动。
- 未知字段：返回包含文件路径和字段路径的错误。
- 不允许静默忽略用户拼写错误，也不允许用户错误自动回退主题。

## 12. 严格解析

JSONC 和 YAML 都必须启用未知字段检查。JSONC 解析流程应是：

```text
remove comments
  -> reject null
  -> strict decode into concrete patch struct
  -> reject trailing document content
```

禁止先解码到 `map[string]any` 再手工查找字段。若当前 JSONC 库不支持严格字段检查，应使用标准 `json.Decoder.DisallowUnknownFields()` 解析去除注释后的 JSON。

YAML 使用 `KnownFields(true)`，并检查只包含一个 YAML document。

## 13. 校验边界

### AppConfig

- duration 必须为正数或处于字段允许范围。
- runtime connection name 唯一。
- default connection 必须存在或是允许的本地发现连接。
- TLS 组合合法。
- KeyBinding 的 primary 必须存在，同一作用域内检查冲突。
- Dialog min/max/percent 和 position 合法。
- enum 值必须属于定义域。

### Theme

- 每个必需视觉属性均已解析。
- `ColorRef` 只能包含 token 或 value 之一。
- token 必须存在于固定 `ColorToken` enum。
- 十六进制颜色格式合法。
- overlay opacity 为 `0..100`。
- border kind 属于固定 enum。
- theme patch 不可能出现应用配置字段。

## 14. 代码边界

建议文件结构：

```text
internal/data/config/
  constants_domain.go
  constants_files.go
  model_app.go
  model_theme.go
  patch_app.go
  patch_theme.go
  apply_app.go
  apply_theme.go
  decode_jsonc.go
  decode_yaml.go
  validate_app.go
  validate_theme.go
  loader.go
  embed.go
  defaults/*.jsonc
  themes/*.jsonc
```

职责：

- `constants_*`：领域枚举值、资源名、路径段和 fallback 字符串常量。
- `model_*`：完整 resolved 类型和 enum。
- `patch_*`：presence-aware 输入类型。
- `apply_*`：静态逐字段层叠逻辑。
- `decode_*`：严格格式解析，不处理优先级。
- `validate_*`：最终领域校验。
- `loader.go`：唯一的优先级编排入口。
- `cmd/docker-tui/main.go`：只构造 `LoadOptions` 并消费 `Resolved`。

TUI 组件不得从私有 JSONC 加载用户可配置行为或主题视觉属性；这些值必须通过 `Dependencies.Config` 和 `Dependencies.Theme` 注入。组件可加载不可由用户覆盖的结构化静态资源，但 schema 必须是固定 struct 并启用未知字段检查。

## 15. 删除清单

实施时直接删除，不提供兼容层：

- `Config.Theme string` 旧顶层字段。
- `UIConfig.Theme ThemeConfig`。
- `ThemeConfig` 类型。
- `UIConfig.BorderStyle`，改由 `Theme` 唯一负责。
- `UIConfig.DialogOverlayColor`，改由 `Theme.Dialog.Overlay` 负责。
- `map[string]string CommandTemplates`。
- 样式中的 `[]string` 表达。
- Keymap action 的 `[]string` 表达。
- `internal/data/config/styles/` 路径。
- `~/.config/docker-tui/styles/` 用户路径。
- `internal/tui/ui/widget/dialog/dialog.jsonc`。
- 其他组件内用于用户可配置值的私有 JSONC。
- `LoadSplit()`、旧 `Load()` 及其兼容 wrapper。
- v1 到 v2 migration 和旧字段映射代码。

## 16. 实施阶段

### P0：建立新模型

1. 定义 `AppConfig`、`Theme`、enum 和基础值对象。
2. 定义全部领域字符串、文件名、路径段和 fallback 常量。
3. 定义完整的 `AppPatch`、`ThemePatch`。
4. 实现静态 `Apply()`。
5. 为显式 `false`、`0`、空字符串和未声明字段增加测试。

### P1：迁移内嵌默认值

1. 新建 `defaults/*.jsonc`。
2. 将行为默认值迁移到明确 section。
3. 将视觉默认值迁移到 `themes/default.jsonc`。
4. 其他主题改成相对 default 的 patch。
5. 删除重复或无归属的默认值。

### P2：统一加载器

1. 实现严格 JSONC/YAML decoder。
2. 实现 `LoadResolved()`。
3. 实现用户主题查找。
4. 将 CLI 显式参数转换为 typed patch。
5. main 改为只消费 resolved 结果。

### P3：迁移消费者

1. TUI 行为读取 `Resolved.App`。
2. 所有颜色和样式读取 `Resolved.Theme`。
3. Dialog 布局迁移到 `AppConfig.UI.Dialog`。
4. Dialog 和其他组件视觉配置迁移到固定 Theme scope。
5. Keymap 消费 `KeyBinding` 值对象。

### P4：删除旧系统

1. 执行 §15 删除清单。
2. 删除无引用 loader、类型和测试。
3. 更新示例配置和用户文档。
4. 全仓搜索旧字段名，确保没有双轨读取。
5. 扫描配置模块生产代码，确保不存在允许范围外的裸字符串面量。

## 17. 测试矩阵

### 层叠

| 场景 | 预期 |
|---|---|
| 无用户配置 | 内嵌默认覆盖 Go fallback |
| 用户只改一个嵌套属性 | 其他属性继承 |
| 用户显式 `false` 或 `0` | 精确保留并由 validator 判断是否合法 |
| CLI 与用户修改同一行为字段 | CLI 胜出 |
| 用户 theme override 与主题冲突 | 用户 override 胜出 |
| CLI 更换主题 | 更换基础主题，用户视觉 override 继续生效 |
| 两个内嵌默认文件拥有同一字段 | 测试失败 |

### 强类型边界

| 场景 | 预期 |
|---|---|
| theme 文件包含 runtime | 未知字段错误 |
| config.yml 的 appearance 包含 keymap | 未知字段错误 |
| 任意 selector 名 | 未知字段错误 |
| 未定义 palette token | 校验失败 |
| 样式使用数组 | 类型错误 |
| command 使用未知 key | 未知字段错误 |
| keymap 使用未定义 action | 未知字段错误 |

### 集合

| 场景 | 预期 |
|---|---|
| 用户声明 connections | 整体替换默认集合 |
| connection name 重复 | 校验失败 |
| 空 connections | 按 runtime discovery/default 规则校验 |
| KeyBinding secondary 为空 | 合法 |
| 同一作用域 primary 冲突 | 校验失败 |

### 错误处理

| 场景 | 预期 |
|---|---|
| config.yml 不存在 | 正常启动 |
| config.yml 语法错误 | 启动失败并显示路径 |
| 用户主题不存在 | 启动失败并显示主题名 |
| 用户主题存在但非法 | 启动失败并显示字段路径 |
| 内嵌默认非法 | 测试失败，运行时仅使用 Go fallback |

## 18. 验收标准

1. 仓库中不存在用于配置作用域的 `map[string]string` 或 `map[string]any`。
2. 样式模型中不存在 `[]string`。
3. Keymap action 不再使用 `[]string`。
4. 每个用户可配置字段都属于明确的 named struct。
5. 每个主题属性都属于固定组件 scope 和固定 property。
6. 用户 patch 的每个叶子字段都能区分未声明与显式零值。
7. 主题在类型层面无法表达 runtime、keymap、logs 或 command 配置。
8. 应用配置中不存在重复主题字段。
9. 所有用户可配置行为和视觉属性都通过 resolved 配置依赖获取；组件 JSONC 仅承载强类型内部静态资源。
10. 优先级链有端到端测试。
11. `go test ./...` 通过。
12. 全仓搜索确认旧配置模型和兼容代码已删除。
13. 配置模块生产代码中的语义字符串均为命名常量或强类型常量。
14. 裸字符串扫描只允许 struct tag、`go:embed` pattern 和明确登记的语法例外。

## 19. 非目标

- CSS parser 或 CSS selector。
- 任意用户自定义配置作用域。
- 配置热加载。
- 远程下载主题。
- 插件动态注册配置字段。
- 旧配置迁移或兼容提示。
- 运行时反射式 deep merge。

## 20. 最终决策表

| 决策项 | 结论 |
|---|---|
| 配置模型 | 固定作用域的强类型 struct |
| CSS 借鉴范围 | scope、property、token、cascade |
| 任意 map | 禁止 |
| 样式数组 | 禁止 |
| Keymap 数组 | 改成 `KeyBinding` |
| 真实实体集合 | 允许强类型 slice，整体替换 |
| 用户配置格式 | 单一 `config.yml` |
| 用户主题格式 | 严格 JSONC `ThemeDocument` |
| 用户视觉覆盖 | `appearance.overrides` typed `ThemePatch` |
| 主题继承 | Go fallback -> embedded default -> selected -> user override |
| Dialog | 行为进入 AppConfig，视觉进入 Theme |
| 组件私有配置 | 删除行为/视觉双轨；允许强类型内部静态资源 |
| 兼容策略 | 不兼容、不迁移、不保留 wrapper |
| 加载入口 | 仅 `LoadResolved()` |
| 合并实现 | 静态 typed `Apply()`，不使用通用 deep merge |
| 字符串面量 | 生产代码语义字符串全部命名常量化，仅保留语言强制例外 |
