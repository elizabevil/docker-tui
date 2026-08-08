package docker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/client"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// Client wraps the Docker SDK client and provides Docker engine operations.
type Client struct {
	cli           *client.Client
	ctx           context.Context
	cancel        context.CancelFunc
	Host          string
	APIVersion    string
	TLS           utils.TLSConfig
	EngineVersion string
}

// NewClient creates and pings a Docker client using the given configuration.
func NewClient(cfg runtimeapi.ClientConfig) (*Client, error) {
	host := cfg.Host
	if host == "" {
		host = defaultDockerHost()
	}

	opts := []client.Opt{client.WithHost(host)}
	if cfg.APIVersion != "" {
		opts = append(opts, client.WithVersion(cfg.APIVersion))
	} else {
		opts = append(opts, client.WithAPIVersionNegotiation())
	}

	if cfg.TLS.Enabled {
		if !cfg.TLS.Verify && !cfg.TLS.InsecureSkipVerify {
			return nil, fmt.Errorf("TLS certificate verification is required")
		}
		h, err := tlsHTTPClient(cfg.TLS, host)
		if err != nil {
			return nil, err
		}
		opts = append(opts, client.WithHTTPClient(h))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if _, err := cli.Ping(ctx); err != nil {
		_ = cli.Close() //nolint:errcheck // client failed to ping; closing best-effort.
		return nil, fmt.Errorf("ping failed (%s): %w", host, err)
	}

	c := &Client{
		cli:           cli,
		Host:          host,
		APIVersion:    cfg.APIVersion,
		TLS:           cfg.TLS,
		EngineVersion: fetchVersion(cli, 2*time.Second),
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	return c, nil
}

func defaultDockerHost() string {
	return "unix:///var/run/docker.sock"
}

func tlsHTTPClient(cfg utils.TLSConfig, host string) (*http.Client, error) {
	tlsCfg, err := tlsConfig(cfg, host)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
	}, nil
}

func tlsConfig(cfg utils.TLSConfig, host string) (*tls.Config, error) {
	tlsCfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}
	if cfg.ServerName != "" {
		tlsCfg.ServerName = cfg.ServerName
	} else if inferred := inferServerName(host); inferred != "" {
		tlsCfg.ServerName = inferred
	}
	if cfg.CAFile != "" {
		caPEM, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read CA file: %w", err)
		}
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(caPEM); !ok {
			return nil, fmt.Errorf("parse CA file: no certificates found")
		}
		tlsCfg.RootCAs = pool
	}
	if cfg.CertFile != "" || cfg.KeyFile != "" {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			return nil, fmt.Errorf("TLS certFile and keyFile must be configured together")
		}
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load TLS cert pair: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}
	return tlsCfg, nil
}

func inferServerName(host string) string {
	if host == "" {
		return ""
	}
	if parsed, err := url.Parse(host); err == nil && parsed.Host != "" {
		switch parsed.Scheme {
		case "tcp", "http", "https":
			if name := parsed.Hostname(); name != "" {
				return name
			}
		}
	}
	trimmed := strings.TrimSpace(host)
	for _, prefix := range []string{"tcp://", "https://", "http://"} {
		trimmed = strings.TrimPrefix(trimmed, prefix)
	}
	if trimmed == "" || strings.Contains(trimmed, "/") {
		return ""
	}
	if name, _, ok := strings.Cut(trimmed, ":"); ok {
		return name
	}
	if trimmed != "" {
		return trimmed
	}
	return ""
}

func fetchVersion(cli *client.Client, timeout time.Duration) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ver, err := cli.ServerVersion(ctx)
	if err != nil {
		return "?"
	}
	return ver.Version
}

// Close releases resources held by the client.
func (c *Client) Close() error {
	c.cancel()
	return c.cli.Close()
}

// Ping verifies the runtime connection with the default timeout.
func (c *Client) Ping() error {
	return c.PingTimeout(2 * time.Second)
}

// PingContext implements the runtime engine lifecycle contract.
func (c *Client) PingContext(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	return err
}

// Identity returns engine metadata without leaking Docker SDK types.
func (c *Client) Identity() runtimeapi.Identity {
	return runtimeapi.Identity{
		Type:     runtimeapi.Docker,
		Endpoint: c.Host,
		Version:  c.EngineVersion,
	}
}

// Capabilities describes the normalized behavior exposed by this adapter.
func (c *Client) Capabilities() runtimeapi.CapabilitySet {
	return runtimeapi.CapabilitySet{
		runtimeapi.CapabilityContainers: {Support: runtimeapi.Available},
		runtimeapi.CapabilityImages:     {Support: runtimeapi.Available},
		runtimeapi.CapabilityVolumes:    {Support: runtimeapi.Available},
		runtimeapi.CapabilityNetworks:   {Support: runtimeapi.Available},
		runtimeapi.CapabilityEvents:     {Support: runtimeapi.Available},
		runtimeapi.CapabilityExec:       {Support: runtimeapi.Available},
		runtimeapi.CapabilityFiltering: {
			Support:    runtimeapi.Degraded,
			Reason:     "the first same-field value is native; remaining values use equivalent adapter post-filtering",
			ReasonCode: "same_field_and_post_filter",
		},
		runtimeapi.CapabilityContainerListFilter: {
			Support:    runtimeapi.Degraded,
			Reason:     "the first same-field value is native; remaining values use equivalent adapter post-filtering",
			ReasonCode: "same_field_and_post_filter",
		},
		runtimeapi.CapabilityImageListFilter: {
			Support:    runtimeapi.Degraded,
			Reason:     "same-field AND uses adapter post-filtering; relative before, since, and until combinations are rejected",
			ReasonCode: "partial_same_field_and_post_filter",
		},
		runtimeapi.CapabilityVolumeListFilter: {
			Support:    runtimeapi.Degraded,
			Reason:     "same-field AND uses adapter post-filtering; multiple dangling conditions are rejected",
			ReasonCode: "partial_same_field_and_post_filter",
		},
		runtimeapi.CapabilityNetworkListFilter: {
			Support:    runtimeapi.Degraded,
			Reason:     "same-field AND uses adapter post-filtering; multiple type conditions are rejected",
			ReasonCode: "partial_same_field_and_post_filter",
		},
		runtimeapi.CapabilityEventFilter: {Support: runtimeapi.Available},
		runtimeapi.CapabilityExecResize:  {Support: runtimeapi.Available},
		runtimeapi.CapabilityStatsStream: {Support: runtimeapi.Available},
		runtimeapi.CapabilityComposeAggregate:    {Support: runtimeapi.Available},
		runtimeapi.CapabilityComposeProjectLevel: {Support: runtimeapi.Degraded, Reason: "list/inspect wrap, no atomic project-level operation", ReasonCode: "wrap_no_atomic"},
		runtimeapi.CapabilityComposeUp:           {Support: runtimeapi.Available},
		runtimeapi.CapabilityComposeBuild:        {Support: runtimeapi.Unsupported, Reason: "daemon has no Dockerfile context path", ReasonCode: "no_dockerfile_context"},
		runtimeapi.CapabilityComposeRun:          {Support: runtimeapi.Available, Reason: "one-off run with --rm / detach semantics per R08-08", ReasonCode: "oneoff_create_start_wait_remove"},
		runtimeapi.CapabilityComposeConfig:       {Support: runtimeapi.Unsupported, Reason: "yaml merge endpoint absent", ReasonCode: "no_yaml_endpoint"},
		runtimeapi.CapabilityComposePull:         {Support: runtimeapi.Degraded, Reason: "per-image supported, batch is loop", ReasonCode: "loop_only"},
		runtimeapi.CapabilityComposePush:         {Support: runtimeapi.Degraded, Reason: "per-image supported, batch is loop", ReasonCode: "loop_only"},
		runtimeapi.CapabilityComposePodScope:     {Support: runtimeapi.Available, Reason: "DockerPodService self-implements wrap (Q6)", ReasonCode: "wrap_self_implemented"},
		runtimeapi.CapabilityComposeCoLocated:    {Support: runtimeapi.Degraded, Reason: "N+1 inspect with 5s TTL + events invalidation (R08-15 Q2)", ReasonCode: "n_plus_1_cached"},
	}
}

// PingTimeout verifies the runtime connection with a caller-selected deadline.
func (c *Client) PingTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(c.ctx, timeout)
	defer cancel()
	_, err := c.cli.Ping(ctx)
	return err
}

func (c *Client) Compose() runtimeapi.ComposeService { return DockerComposeService{Client: c} }

func (c *Client) Pods() runtimeapi.PodService { return DockerPodService{Client: c} }
