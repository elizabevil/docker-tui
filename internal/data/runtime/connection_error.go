package runtime

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/fs"
	"net"
	"strings"
)

// ConnectionErrorKind classifies the root cause of a connection failure.
type ConnectionErrorKind string

// Connection error kinds. These drive TLS state display, security prompts, and
// error projection in the TUI without exposing transport-specific error chains.
const (
	ConnectionErrorUnknown    ConnectionErrorKind = "unknown"
	ConnectionErrorCA         ConnectionErrorKind = "ca"
	ConnectionErrorClientCert ConnectionErrorKind = "client_certificate"
	ConnectionErrorHostname   ConnectionErrorKind = "hostname"
	ConnectionErrorHandshake  ConnectionErrorKind = "handshake"
	ConnectionErrorNetwork    ConnectionErrorKind = "network"
)

// ConnectionFailure carries the classified kind of a connection error.
type ConnectionFailure struct {
	Kind ConnectionErrorKind
}

type classifiedConnectionError struct {
	kind  ConnectionErrorKind
	cause error
}

func (e *classifiedConnectionError) Error() string { return e.cause.Error() }
func (e *classifiedConnectionError) Unwrap() error { return e.cause }

// ClassifyConnectionError inspects the error chain and returns the most
// specific connection failure kind. It unwraps TLS, x509, and net errors
// to map them to the appropriate ConnectionErrorKind.
func ClassifyConnectionError(err error) ConnectionFailure {
	if err == nil {
		return ConnectionFailure{}
	}
	var classified *classifiedConnectionError
	if errors.As(err, &classified) {
		return ConnectionFailure{Kind: classified.kind}
	}
	var hostname x509.HostnameError
	if errors.As(err, &hostname) {
		return ConnectionFailure{Kind: ConnectionErrorHostname}
	}
	var unknownAuthority x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthority) {
		return ConnectionFailure{Kind: ConnectionErrorCA}
	}
	var certificateInvalid x509.CertificateInvalidError
	if errors.As(err, &certificateInvalid) {
		return ConnectionFailure{Kind: ConnectionErrorCA}
	}
	var verification *tls.CertificateVerificationError
	if errors.As(err, &verification) {
		return ClassifyConnectionError(verification.Err)
	}
	var recordHeader tls.RecordHeaderError
	if errors.As(err, &recordHeader) || strings.Contains(strings.ToLower(err.Error()), "tls handshake") {
		return ConnectionFailure{Kind: ConnectionErrorHandshake}
	}
	// fs.PathError satisfies net.Error (both Timeout()/Temporary() exist but
	// return false), so it must be checked before the broader net.Error match.
	// These are file-system errors from TLS config loading, not network errors.
	var pathError *fs.PathError
	if errors.As(err, &pathError) {
		return ConnectionFailure{Kind: ConnectionErrorUnknown}
	}
	var netError net.Error
	if errors.As(err, &netError) {
		return ConnectionFailure{Kind: ConnectionErrorNetwork}
	}
	return ConnectionFailure{Kind: ConnectionErrorUnknown}
}
