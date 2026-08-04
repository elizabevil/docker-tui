package config

import "time"

const CurrentConfigVersion = 1

const emptyValue = ""

type RuntimeDriver string

const (
	RuntimeDriverDocker RuntimeDriver = "docker"
	RuntimeDriverPodman RuntimeDriver = "podman"
)

type ReportingMode string

const ReportingOff ReportingMode = "off"

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

type HorizontalAlignment string

const (
	HorizontalLeft   HorizontalAlignment = "left"
	HorizontalCenter HorizontalAlignment = "center"
	HorizontalRight  HorizontalAlignment = "right"
)

type VerticalAlignment string

const (
	VerticalTop    VerticalAlignment = "top"
	VerticalCenter VerticalAlignment = "center"
	VerticalBottom VerticalAlignment = "bottom"
)

type BackgroundType string

const (
	BackgroundSolid     BackgroundType = "solid"
	BackgroundGradient  BackgroundType = "gradient"
	BackgroundImageType BackgroundType = "image"
)

type BackgroundSizing string

const (
	BackgroundSizingCover   BackgroundSizing = "cover"
	BackgroundSizingContain BackgroundSizing = "contain"
)

type BackgroundPosition string

const (
	BackgroundPositionTop    BackgroundPosition = "top"
	BackgroundPositionCenter BackgroundPosition = "center"
	BackgroundPositionBottom BackgroundPosition = "bottom"
)

const (
	DefaultConnectionLocalDocker = "local-docker"
	DefaultConnectionLocalPodman = "local-podman"
	EndpointSchemeUnix           = "unix"
	DefaultDockerExecutable      = "docker"
	DefaultDockerSubcommand      = "compose"
	DefaultLogsSince             = "1h"
	DefaultLogsTail              = "200"
	DefaultDockerTimeout         = 30 * time.Second
	DefaultTableRowPrefix        = "  "
	DefaultTableSelectedPrefix   = "┃ "
)

const (
	KeyEmpty     Key = ""
	KeyQuestion  Key = "?"
	KeySlash     Key = "/"
	KeySemicolon Key = ";"
	KeyTab       Key = "tab"
	KeyShiftTab  Key = "shift+tab"
	KeyUp        Key = "up"
	KeyDown      Key = "down"
	KeyEnter     Key = "enter"
	KeyEscape    Key = "esc"
	KeyF1        Key = "f1"
	KeyC         Key = "c"
	KeyE         Key = "e"
	KeyH         Key = "h"
	KeyI         Key = "i"
	KeyJ         Key = "j"
	KeyK         Key = "k"
	KeyL         Key = "l"
	KeyM         Key = "m"
	KeyP         Key = "p"
	KeyR         Key = "r"
	KeyS         Key = "s"
	KeyCtrlC     Key = "ctrl+c"
	KeyCtrlD     Key = "ctrl+d"
	KeyCtrlE     Key = "ctrl+e"
	KeyCtrlK     Key = "ctrl+k"
	KeyCtrlL     Key = "ctrl+l"
	KeyCtrlP     Key = "ctrl+p"
	KeyCtrlR     Key = "ctrl+r"
	KeyCtrlS     Key = "ctrl+s"
	KeyCtrlT     Key = "ctrl+t"
	KeyCtrlU     Key = "ctrl+u"
)

type ColorToken string

const (
	ColorTokenPrimary          ColorToken = "primary"
	ColorTokenSuccess          ColorToken = "success"
	ColorTokenWarning          ColorToken = "warning"
	ColorTokenDanger           ColorToken = "danger"
	ColorTokenInfo             ColorToken = "info"
	ColorTokenAccent           ColorToken = "accent"
	ColorTokenAccentSecondary  ColorToken = "accentSecondary"
	ColorTokenForeground       ColorToken = "foreground"
	ColorTokenForegroundMuted  ColorToken = "foregroundMuted"
	ColorTokenBackground       ColorToken = "background"
	ColorTokenBackgroundSubtle ColorToken = "backgroundSubtle"
	ColorTokenBackgroundDeep   ColorToken = "backgroundDeep"
	ColorTokenTransparent      ColorToken = "transparent"
)

const (
	FallbackColorPrimary          Color = "#42a5f5"
	FallbackColorSuccess          Color = "#2ecc71"
	FallbackColorWarning          Color = "#ffca28"
	FallbackColorDanger           Color = "#ef5350"
	FallbackColorInfo             Color = "#00bcd4"
	FallbackColorAccent           Color = "#ff9800"
	FallbackColorAccentSecondary  Color = "#ab47bc"
	FallbackColorForeground       Color = "#ffffff"
	FallbackColorForegroundMuted  Color = "#546e7a"
	FallbackColorBackground       Color = "#0d1117"
	FallbackColorBackgroundSubtle Color = "#1a1a2e"
	FallbackColorBackgroundDeep   Color = "#16213e"
	FallbackColorTableMarkedBackground Color = "#463f16"
	FallbackColorTableColumnForeground Color = "#ECEFF1"
	FallbackColorTableNameForeground   Color = "#FAF0E6"
	FallbackColorTransparent            Color = "transparent"
)
