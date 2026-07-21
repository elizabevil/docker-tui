package i18n

func ConnectionFailureMessage(kind string) string {
	key := "connection.error.unknown"
	switch kind {
	case "ca":
		key = "connection.error.ca"
	case "client_certificate":
		key = "connection.error.client_certificate"
	case "hostname":
		key = "connection.error.hostname"
	case "handshake":
		key = "connection.error.handshake"
	case "network":
		key = "connection.error.network"
	}
	return T(key)
}
