package i18n

import (
	"testing"
)

func TestConnectionFailureMessageFollowsLanguage(t *testing.T) {
	t.Cleanup(func() { SetLang(LanguageEnglish) })
	Init(LanguageEnglish)
	if got := ConnectionFailureMessage("hostname"); got != "TLS certificate hostname mismatch" {
		t.Fatalf("english message = %q", got)
	}
	SetLang(LanguageChinese)
	if got := ConnectionFailureMessage("hostname"); got != "TLS 证书主机名不匹配" {
		t.Fatalf("chinese message = %q", got)
	}
}
