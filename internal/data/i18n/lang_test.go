package i18n

import "testing"

func TestSetLangLoadsTranslations(t *testing.T) {
	Init(LanguageChinese)
	if Current() != LanguageChinese {
		t.Fatalf("current language=%q", Current())
	}
	if got := T("panel.containers"); got != "容器" {
		t.Fatalf("Chinese translation=%q", got)
	}
	Init(LanguageEnglish)
	if got := T("panel.containers"); got != "Containers" {
		t.Fatalf("English translation=%q", got)
	}
}

func TestSetLangAcceptsChineseLocale(t *testing.T) {
	Init("zh-CN")
	if Current() != LanguageChinese || T("panel.images") != "镜像" {
		t.Fatalf("locale was not normalized: current=%q text=%q", Current(), T("panel.images"))
	}
}
