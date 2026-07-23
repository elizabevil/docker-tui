package docker

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/utils"
)

func TestClassifyConnectionError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want runtime.ConnectionErrorKind
	}{
		{name: "ca", err: x509.UnknownAuthorityError{}, want: runtime.ConnectionErrorCA},
		{name: "hostname", err: x509.HostnameError{Certificate: &x509.Certificate{}, Host: "runtime.example"}, want: runtime.ConnectionErrorHostname},
		{name: "handshake", err: tls.RecordHeaderError{}, want: runtime.ConnectionErrorHandshake},
		{name: "network", err: &net.DNSError{Err: "timeout", Name: "runtime.example", IsTimeout: true}, want: runtime.ConnectionErrorNetwork},
		{name: "unknown", err: errors.New("unexpected response"), want: runtime.ConnectionErrorUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runtime.ClassifyConnectionError(tt.err).Kind; got != tt.want {
				t.Fatalf("kind = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTLSConfigurationErrorsHaveSafeCategories(t *testing.T) {
	tests := []struct {
		name string
		cfg  utils.TLSConfig
		want runtime.ConnectionErrorKind
	}{
		{name: "missing ca", cfg: utils.TLSConfig{Enabled: true, Verify: true, CAFile: "/does/not/exist/ca.pem"}, want: runtime.ConnectionErrorUnknown},
		{name: "missing client key", cfg: utils.TLSConfig{Enabled: true, Verify: true, CertFile: "cert.pem"}, want: runtime.ConnectionErrorUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tlsConfig(tt.cfg, "tcp://runtime.example:2376")
			if err == nil {
				t.Fatal("expected TLS configuration error")
			}
			if got := runtime.ClassifyConnectionError(err).Kind; got != tt.want {
				t.Fatalf("kind = %q, want %q", got, tt.want)
			}
		})
	}
}
