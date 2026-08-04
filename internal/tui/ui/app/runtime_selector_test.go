package view

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/utils"
)

func TestRuntimeSelectorShowsSafeTLSStateAndFailure(t *testing.T) {
	i18n.SetLang(i18n.LanguageEnglish)
	pool := dockerclient.NewPool(nil)
	pool.AddHost(dockerclient.HostEntry{
		Name: "remote",
		Host: "tcp://runtime.example:2376",
		TLS:  utils.TLSConfig{Enabled: true, Verify: true},
	})
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Connection.Pool = pool
	m.Connection.RuntimeSelectorError = map[string]dockerclient.ConnectionFailure{
		"remote": {Kind: dockerclient.ConnectionErrorClientCert},
	}

	got := renderRuntimeSelector(m)
	if !strings.Contains(got, "TLS configured") || !strings.Contains(got, "TLS client") || !strings.Contains(got, "certificate or key is invalid") {
		t.Fatalf("selector did not expose safe TLS state: %q", got)
	}
	for _, sensitive := range []string{"PRIVATE KEY", "BEGIN CERTIFICATE", "secret"} {
		if strings.Contains(got, sensitive) {
			t.Fatalf("selector exposed sensitive value %q", sensitive)
		}
	}
}
