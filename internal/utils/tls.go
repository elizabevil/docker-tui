package utils

// TLSConfig configures TLS for the engine connection.
type TLSConfig struct {
	Enabled            bool
	Verify             bool
	InsecureSkipVerify bool
	CAFile             string
	CertFile           string
	KeyFile            string
	ServerName         string
}
