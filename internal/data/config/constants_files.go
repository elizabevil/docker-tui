package config

const (
	configDirName      = "docker-tui"
	userConfigRootName = ".config"
	configFileName     = "config.yml"
	userThemesDirName  = "themes"
	jsoncExtension     = ".jsonc"
	themeDefaultName   = "default"
	defaultsFSRoot     = "defaults"
	themesFSRoot       = "themes"
	themeSourcePrefix  = "embedded:"
	currentFSDir       = "."
)

const (
	themeNameDash       = '-'
	themeNameUnderscore = '_'
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

const (
	errMultipleYAMLDocuments = "multiple YAML documents are not allowed"
	errMultipleJSONDocuments = "multiple JSON documents are not allowed"
	errNullNotAllowed        = "null is not allowed"
	yamlNullTag              = "!!null"
)
