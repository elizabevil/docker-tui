package config

// Template 渲染相关常量：键名、中文说明与格式串。
// 全部为包级常量以满足 string_literals_test 的豁免规则。

const (
	newline          = "\n"
	indentUnit       = "  "
	bindingSeparator = "/"
	emptyListLiteral = "[]"
	emptyMapLiteral  = "{}"
)

const (
	scalarCommentFmt    = "%s# %s — %s（默认: %s）\n"
	containerCommentFmt = "%s# %s — %s\n"
	sectionHeaderFmt    = "%s# ===== %s =====\n"
	kvFmt               = "%s%s: %s\n"
	keyOnlyFmt          = "%s%s:\n"
)

const templateHeader = `# =========================================================
# docker-tui 配置模板
#
# 本文件由 dtui config init 生成。所有键均为可选，
# 省略的键在加载时回退到内嵌默认值。修改后请使用
# dtui config validate 校验配置。
# =========================================================`

const (
	keyVersion    = "version"
	keyApp        = "app"
	keyAppearance = "appearance"
	keyTheme      = "theme"
	keyOverrides  = "overrides"
)

const (
	descVersion   = "配置版本号"
	descApp       = "应用配置"
	descTheme     = "界面主题"
	descOverrides = "主题覆盖设置"
)

const (
	keyGeneral      = "general"
	keyScrollHeight = "scrollHeight"
	keyReporting    = "reporting"
	keyLang         = "lang"
	keySizeFormat   = "sizeFormat"
)

const (
	descScrollHeight = "滚动保留的行数"
	descReporting    = "上报统计信息"
	descLang         = "界面语言"
	descSizeFormat   = "文件大小显示格式"
)

const (
	keyUI                  = "ui"
	keyWrapMainPanel       = "wrapMainPanel"
	keyReturnImmediately   = "returnImmediately"
	keyShowHelp            = "showHelp"
	keyHintTimeout         = "hintTimeout"
	keyEnableMouse         = "enableMouse"
	keyDialog              = "dialog"
	keyWindow              = "window"
	keyHeader              = "header"
	keyTable               = "table"
	keyWidth               = "width"
	keyHeight              = "height"
	keyPosition            = "position"
	keyPanelSize           = "panelSize"
	keyPercent             = "percent"
	keyMin                 = "min"
	keyMax                 = "max"
	keyHorizontal          = "horizontal"
	keyVertical            = "vertical"
	keyOffsetX             = "offsetX"
	keyOffsetY             = "offsetY"
	keyWidthPercent        = "widthPercent"
	keyHeightPercent       = "heightPercent"
	keyMaxWidth            = "maxWidth"
	keyMaxHeight           = "maxHeight"
	keyMarginTopPercent    = "marginTopPercent"
	keyMarginBottomPercent = "marginBottomPercent"
	keyContentWidthPercent = "contentWidthPercent"
	keyColumns             = "columns"
	keyKeystrokeContentRatio = "keystrokeContentRatio"
	keyHost                = "host"
	keyConnection          = "connection"
	keyKeystroke           = "keystroke"
	keyLogo                = "logo"
	keyColumnSpacing       = "columnSpacing"
	keyRowPrefix           = "rowPrefix"
	keyRowPrefixSelected   = "rowPrefixSelected"
	keyRowSpacing          = "rowSpacing"
	keySelectionInfo       = "selectionInfo"
	keyEnabled             = "enabled"
	keyPadLines            = "padLines"
)

const (
	descWrapMainPanel       = "主面板内容自动换行"
	descReturnImmediately   = "操作后立即返回主界面"
	descShowHelp            = "显示帮助提示"
	descHintTimeout         = "帮助提示持续时间（秒）"
	descEnableMouse         = "启用鼠标支持"
	descDialog              = "对话框布局"
	descDialogWidth         = "对话框宽度"
	descDialogHeight        = "对话框高度"
	descDialogPosition      = "对话框位置"
	descPanelSize           = "对话框面板尺寸"
	descPercent             = "占屏幕百分比"
	descMin                 = "最小尺寸"
	descMax                 = "最大尺寸"
	descHorizontal          = "水平位置"
	descVertical            = "垂直位置"
	descOffsetX             = "水平偏移"
	descOffsetY             = "垂直偏移"
	descWidthPercent        = "宽度百分比"
	descHeightPercent       = "高度百分比"
	descMaxWidth            = "最大宽度"
	descMaxHeight           = "最大高度"
	descWindow              = "窗口布局"
	descMarginTopPercent    = "顶部边距百分比"
	descMarginBottomPercent = "底部边距百分比"
	descContentWidthPercent = "内容宽度百分比"
	descHeader              = "头部布局"
	descColumns             = "列权重"
	descKeystrokeContentRatio = "按键与内容比例"
	descHost                = "主机列权重"
	descConnection          = "连接列权重"
	descKeystroke           = "按键列权重"
	descLogo                = "徽标列权重"
	descTable               = "表格布局"
	descColumnSpacing       = "列间距"
	descRowPrefix           = "普通行前缀"
	descRowPrefixSelected   = "选中行前缀"
	descRowSpacing          = "行间距"
	descSelectionInfo       = "选中信息"
	descEnabled             = "是否启用"
	descPadLines            = "上下填充行数"
)

const (
	keyDocker      = "docker"
	keyTimeout     = "timeout"
	keyStatsPollSec = "statsPollSec"
)

const (
	descTimeout     = "命令超时时间"
	descStatsPollSec = "统计轮询间隔（秒）"
)

const (
	keyRuntime         = "runtime"
	keyDefault         = "default"
	keyDiscovery       = "discovery"
	keyHealth          = "health"
	keyLocalDocker     = "localDocker"
	keyLocalPodman     = "localPodman"
	keyIntervalSec     = "intervalSec"
	keyTimeoutSec      = "timeoutSec"
	keyFailureThreshold = "failureThreshold"
	keyConnections     = "connections"
)

const (
	descDefault          = "默认运行时"
	descDiscovery        = "本地运行时自动发现"
	descHealth           = "运行时健康检查"
	descLocalDocker      = "发现本地 Docker"
	descLocalPodman      = "发现本地 Podman"
	descIntervalSec      = "检查间隔（秒）"
	descTimeoutSec       = "检查超时（秒）"
	descFailureThreshold = "连续失败阈值"
	descConnections      = "远程运行时连接列表"
)

const (
	keyKeymap      = "keymap"
	keyGlobal      = "global"
	keyContainer   = "container"
	keyImage       = "image"
	keyVolume      = "volume"
	keyNetwork     = "network"
	keyNavigation  = "navigation"
	keyQuit        = "quit"
	keyActionBar   = "actionBar"
	keyHelp        = "help"
	keyFilter      = "filter"
	keyRefresh     = "refresh"
	keyStart       = "start"
	keyStop        = "stop"
	keyRestart     = "restart"
	keyKill        = "kill"
	keyRemove      = "remove"
	keyLogs        = "logs"
	keyExec        = "exec"
	keyInspect     = "inspect"
	keyStats       = "stats"
	keyPause       = "pause"
	keyPull        = "pull"
	keyPrune       = "prune"
	keyTag         = "tag"
	keyPush        = "push"
	keySave        = "save"
	keyLoad        = "load"
	keyHistory     = "history"
	keyCreate      = "create"
	keyTabNext     = "tabNext"
	keyTabPrev     = "tabPrev"
	keyUp          = "up"
	keyDown        = "down"
	keyEnter       = "enter"
	keyBack        = "back"
	keyDelete      = "delete"
	keyConfirm     = "confirm"
	keyCancel      = "cancel"
	keyPrimary     = "primary"
	keySecondary   = "secondary"
)

const (
	descGlobal      = "全局按键"
	descContainer   = "容器按键"
	descImage       = "镜像按键"
	descVolume      = "卷按键"
	descNetwork     = "网络按键"
	descNavigation  = "导航按键"
	descKeymapDialog = "对话框按键"
	descQuit        = "退出程序"
	descActionBar   = "打开命令面板"
	descHelp        = "打开帮助"
	descFilter      = "过滤资源"
	descRefresh     = "刷新"
	descStart       = "启动"
	descStop        = "停止"
	descRestart     = "重启"
	descKill        = "强制终止"
	descRemove      = "删除"
	descLogs        = "查看日志"
	descExec        = "进入容器"
	descInspect     = "查看详情"
	descStats       = "查看统计"
	descPause       = "暂停"
	descPull        = "拉取镜像"
	descPrune       = "清理"
	descTag         = "打标签"
	descPush        = "推送镜像"
	descSave        = "保存镜像"
	descLoad        = "加载镜像"
	descHistory     = "查看历史"
	descCreate      = "创建"
	descTabNext     = "下一个标签页"
	descTabPrev     = "上一个标签页"
	descUp          = "上移"
	descDown        = "下移"
	descEnter       = "确认"
	descBack        = "返回"
	descDelete      = "删除"
	descConfirm     = "确认"
	descCancel      = "取消"
)

const (
	keySince       = "since"
	keyTail        = "tail"
	keyTimestamps  = "timestamps"
)

const (
	descSince      = "日志起始时间"
	descTail       = "日志末尾行数"
	descTimestamps = "显示时间戳"
)

const (
	keyLayout         = "layout"
	keySectionWeights = "sectionWeights"
	keyBackground     = "background"
	keyTop            = "top"
	keyContent        = "content"
	keyBottom         = "bottom"
	keyEnable         = "enable"
	keyFallthrough    = "fallthrough"
	keyType           = "type"
	keyColor          = "color"
	keySizing         = "sizing"
	keyOpacity        = "opacity"
)

const (
	descSectionWeights = "分区高度权重"
	descBackground     = "背景设置"
	descTop            = "顶部"
	descContent        = "内容"
	descBottom         = "底部"
	descBgEnable       = "启用背景"
	descBgFallthrough  = "背景点击穿透"
	descBgType         = "背景类型"
	descBgColor        = "背景颜色"
	descBgImage        = "背景图片"
	descBgSizing       = "图片缩放方式"
	descBgPosition     = "图片位置"
	descBgOpacity      = "图片不透明度"
)

const (
	keyCommands      = "commands"
	keyDockerCompose = "dockerCompose"
	keyExecutable    = "executable"
	keySubcommand    = "subcommand"
)

const (
	descDockerCompose = "docker compose 命令"
	descExecutable    = "可执行文件"
	descSubcommand    = "子命令"
)
