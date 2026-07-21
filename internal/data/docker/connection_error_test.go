package docker

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"testing"
)

func TestClassifyConnectionError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ConnectionErrorKind
	}{
		{name: "ca", err: x509.UnknownAuthorityError{}, want: ConnectionErrorCA},
		{name: "hostname", err: x509.HostnameError{Certificate: &x509.Certificate{}, Host: "runtime.example"}, want: ConnectionErrorHostname},
		{name: "handshake", err: tls.RecordHeaderError{}, want: ConnectionErrorHandshake},
		{name: "network", err: &net.DNSError{Err: "timeout", Name: "runtime.example", IsTimeout: true}, want: ConnectionErrorNetwork},
		{name: "unknown", err: errors.New("unexpected response"), want: ConnectionErrorUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClassifyConnectionError(tt.err).Kind; got != tt.want {
				t.Fatalf("kind = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTLSConfigurationErrorsHaveSafeCategories(t *testing.T) {
	tests := []struct {
		name string
		cfg  TLSConfig
		want ConnectionErrorKind
	}{
		{name: "missing ca", cfg: TLSConfig{Enabled: true, Verify: true, CAFile: "/does/not/exist/ca.pem"}, want: ConnectionErrorCA},
		{name: "missing client key", cfg: TLSConfig{Enabled: true, Verify: true, CertFile: "cert.pem"}, want: ConnectionErrorClientCert},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tlsConfig(tt.cfg, "tcp://runtime.example:2376")
			if err == nil {
				t.Fatal("expected TLS configuration error")
			}
			if got := ClassifyConnectionError(err).Kind; got != tt.want {
				t.Fatalf("kind = %q, want %q", got, tt.want)
			}
		})
	}
}
