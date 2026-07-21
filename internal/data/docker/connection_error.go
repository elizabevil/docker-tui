package docker

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"strings"
)

type ConnectionErrorKind string

const (
	ConnectionErrorUnknown    ConnectionErrorKind = "unknown"
	ConnectionErrorCA         ConnectionErrorKind = "ca"
	ConnectionErrorClientCert ConnectionErrorKind = "client_certificate"
	ConnectionErrorHostname   ConnectionErrorKind = "hostname"
	ConnectionErrorHandshake  ConnectionErrorKind = "handshake"
	ConnectionErrorNetwork    ConnectionErrorKind = "network"
)

type ConnectionFailure struct {
	Kind ConnectionErrorKind
}

type classifiedConnectionError struct {
	kind  ConnectionErrorKind
	cause error
}

func (e *classifiedConnectionError) Error() string { return e.cause.Error() }
func (e *classifiedConnectionError) Unwrap() error { return e.cause }

func connectionError(kind ConnectionErrorKind, cause error) error {
	return &classifiedConnectionError{kind: kind, cause: cause}
}

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
	var netError net.Error
	if errors.As(err, &netError) {
		return ConnectionFailure{Kind: ConnectionErrorNetwork}
	}
	return ConnectionFailure{Kind: ConnectionErrorUnknown}
}
