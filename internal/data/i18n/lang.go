package i18n

import (
	"embed"
	"fmt"
	"strings"

	"github.com/bytedance/sonic"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

//go:embed *.jsonc
var translationFS embed.FS

type lang struct {
	code string
	data map[string]string
}

const (
	LanguageEnglish = "en"
	LanguageChinese = "zh"
)

var (
	global *lang
	enData map[string]string
	zhData map[string]string
)

func loadJSONC(name string) map[string]string {
	data, err := translationFS.ReadFile(name)
	if err != nil {
		return nil
	}
	clean := config.StripJSONComments(data)
	var m map[string]string
	if err := sonic.Unmarshal(clean, &m); err != nil {
		return nil
	}
	return m
}

func Init(code string) {
	if m := loadJSONC("en.jsonc"); m != nil {
		enData = m
	} else if m := loadJSONC("en.json"); m != nil {
		enData = m
	} else {
		enData = map[string]string{}
	}
	if m := loadJSONC("zh.jsonc"); m != nil {
		zhData = m
	} else if m := loadJSONC("zh.json"); m != nil {
		zhData = m
	} else {
		zhData = map[string]string{}
	}
	SetLang(code)
}

func SetLang(code string) {
	code = strings.ToLower(strings.TrimSpace(code))
	switch code {
	case LanguageChinese, "zh-cn", "zh_cn":
		global = &lang{code: LanguageChinese, data: zhData}
	default:
		global = &lang{code: LanguageEnglish, data: enData}
	}
}

func Current() string {
	if global == nil {
		return LanguageEnglish
	}
	return global.code
}

func T(key string, args ...any) string {
	if global == nil {
		Init(LanguageEnglish)
	}
	if msg, ok := global.data[key]; ok {
		return interpolate(msg, args)
	}
	if global.code != LanguageEnglish {
		if msg, ok := enData[key]; ok {
			return interpolate(msg, args)
		}
	}
	return key
}

func interpolate(msg string, args []any) string {
	for i, arg := range args {
		msg = strings.ReplaceAll(msg, fmt.Sprintf("{%d}", i), fmt.Sprint(arg))
	}
	return msg
}

func mustRead(name string) []byte {
	data, err := translationFS.ReadFile(name)
	if err != nil {
		return []byte("{}")
	}
	return data
}
