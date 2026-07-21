package view

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestRuntimeSelectorShowsSafeTLSStateAndFailure(t *testing.T) {
	i18n.Init(i18n.LanguageEnglish)
	pool := dockerclient.NewPool()
	pool.AddHost(dockerclient.HostEntry{
		Name: "remote",
		Host: "tcp://runtime.example:2376",
		TLS:  dockerclient.TLSConfig{Enabled: true, Verify: true},
	})
	m := state.NewAppModel(config.DefaultConfig(), nil, "test")
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
