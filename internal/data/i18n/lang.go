package i18n

import (
	"embed"
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/tidwall/jsonc"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

//go:embed lang
var translationFS embed.FS

var (
	printers  = map[language.Tag]*message.Printer{}
	globalTag = language.English
)

const (
	LanguageEnglish  = "en"
	LanguageChinese  = "zh"
	LanguageJapanese = "ja"
)

func init() {
	SetLang(LanguageEnglish)
}

// loadMessages 加载语言文件
func loadMessages(tag language.Tag) {

	cat := catalog.NewBuilder()
	for _, ext := range []string{"jsonc", "json"} {
		file := fmt.Sprintf("lang/%s.%s", tag.String(), ext)
		data, err := translationFS.ReadFile(file)
		if err != nil {
			continue
		}
		var messages map[string]string
		if err := sonic.Unmarshal(jsonc.ToJSON(data), &messages); err != nil {
			panic(err)
		}
		for key, value := range messages {
			_ = cat.SetString(tag, key, value)
		}
	}
	printers[tag] = message.NewPrinter(tag, message.Catalog(cat))
}

// SetLang 设置当前语言
func SetLang(code string) {
	globalTag = parseTag(code)
	loadMessages(globalTag)
}

// parseTag 解析语言代码
func parseTag(code string) language.Tag {
	code = strings.ToLower(strings.TrimSpace(code))
	switch code {
	case "zh", "zh-cn", "zh_cn", "zh-hans", "zh_hans":
		return language.Chinese
	case "ja", "jp", "ja-jp", "ja_jp":
		return language.Japanese
	case "en", "en-us", "en_us", "en-gb", "en_gb":
		return language.English
	default:
		if tag, err := language.Parse(code); err == nil {
			return tag
		}
		return language.English
	}
}

// T 翻译文本
func T(key string, args ...any) string {
	if p, ok := printers[globalTag]; ok {
		return p.Sprintf(key, args...)
	}
	return printers[language.English].Sprintf(key, args...)
}

// Current 获取当前语言代码
func Current() string {
	return globalTag.String()
}

// MustGet 获取翻译，如果不存在则 panic
func MustGet(key string, args ...any) string {
	result := T(key, args...)
	if result == key {
		panic(fmt.Sprintf("translation key not found: %s", key))
	}
	return result
}
