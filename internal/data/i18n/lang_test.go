package i18n

import (
	"fmt"
	"testing"
)

func TestSetLangLoadsTranslations(t *testing.T) {
	SetLang(LanguageChinese)
	if Current() != LanguageChinese {
		t.Fatalf("current language=%q", Current())
	}
	if got := T("panel.containers"); got != "容器" {
		t.Fatalf("Chinese translation=%q", got)
	}

	SetLang(LanguageJapanese)
	fmt.Println(T("panel.containers"))
}
