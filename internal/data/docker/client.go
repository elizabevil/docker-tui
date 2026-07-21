package docker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/client"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type RuntimeType string

const (
	RuntimeDocker RuntimeType = "docker"
	RuntimePodman RuntimeType = "podman"
)

type Client struct {
	cli           *client.Client
	ctx           context.Context
	cancel        context.CancelFunc
	RuntimeType   RuntimeType
	Host          string
	EngineVersion string
	imageLister   ImageLister // nil = use Docker SDK default
}

type ClientConfig struct {
	Host       string
	APIVersion string
	TLS        TLSConfig
	Timeout    time.Duration
	Runtime    RuntimeType
}

type TLSConfig struct {
	Enabled            bool
	Verify             bool
	InsecureSkipVerify bool
	CAFile             string
	CertFile           string
	KeyFile            string
	ServerName         string
}

var knownSockets = []struct {
	path string
	rt   RuntimeType
}{
	{"/var/run/docker.sock", RuntimeDocker},
	{filepath.Join("/", "run", "podman", "podman.sock"), RuntimePodman},
}

func runtimeDir() string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir != "" {
		return dir
	}
	return filepath.Join("/", "run", "user", fmt.Sprintf("%d", os.Getuid()))
}

func podmanUserSockPath() string {
	return filepath.Join(runtimeDir(), "podman", "podman.sock")
}

func isSocketLive(path string) bool {
	conn, err := net.DialTimeout("unix", path, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func resolvePodmanUserSocket() string {
	uid := os.Getuid()
	if uid == 0 {
		return ""
	}
	sockPath := podmanUserSockPath()
	if isSocketLive(sockPath) {
		return sockPath
	}
	return ""
}

func detectHost(cfg ClientConfig) (string, RuntimeType) {
	if cfg.Host != "" {
		runtimeType := detectRuntimeType(cfg.Host)
		if cfg.Runtime == RuntimeDocker {
			runtimeType = RuntimeDocker
		} else if cfg.Runtime == RuntimePodman {
			runtimeType = RuntimePodman
		}
		return cfg.Host, runtimeType
	}
	// Runtime without an endpoint selects the corresponding local socket.
	if cfg.Runtime != "" {
		switch cfg.Runtime {
		case RuntimeDocker:
			return fmt.Sprintf("unix://%s", "/var/run/docker.sock"), RuntimeDocker
		case RuntimePodman:
			if sock := firstLivePodmanSocket(); sock != "" {
				return fmt.Sprintf("unix://%s", sock), RuntimePodman
			}
		}
	}
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		return host, detectRuntimeType(host)
	}
	for _, s := range knownSockets {
		if isSocketLive(s.path) {
			return fmt.Sprintf("unix://%s", s.path), s.rt
		}
	}
	if sock := firstLivePodmanSocket(); sock != "" {
		return fmt.Sprintf("unix://%s", sock), RuntimePodman
	}
	if sock := autoStartPodmanSocket(); sock != "" {
		return sock, RuntimePodman
	}
	return fmt.Sprintf("unix://%s", "/var/run/docker.sock"), RuntimeDocker
}

// firstLivePodmanSocket checks multiple Podman socket locations and returns the first live one.
func firstLivePodmanSocket() string {
	if sock := resolvePodmanUserSocket(); sock != "" {
		return sock
	}
	for _, s := range knownSockets {
		if s.rt == RuntimePodman && isSocketLive(s.path) {
			return s.path
		}
	}
	return ""
}

func autoStartPodmanSocket() string {
	uid := os.Getuid()
	if uid == 0 {
		return ""
	}
	podmanBin, err := exec.LookPath("podman")
	if err != nil {
		return ""
	}
	sockDir := filepath.Join(runtimeDir(), "podman")
	sockPath := podmanUserSockPath()

	if err := os.MkdirAll(sockDir, 0o700); err != nil {
		return ""
	}

	sockURI := fmt.Sprintf("unix://%s", sockPath)
	cmd := exec.Command(podmanBin, "system", "service", "--time=0", sockURI)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return ""
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if isSocketLive(sockPath) {
			return sockURI
		}
		time.Sleep(100 * time.Millisecond)
	}
	return ""
}

func detectRuntimeType(host string) RuntimeType {
	if host == "" {
		return RuntimeDocker
	}
	if filepath.Base(host) == "podman.sock" ||
		host == fmt.Sprintf("unix://%s", filepath.Join("/", "run", "podman", "podman.sock")) {
		return RuntimePodman
	}
	return RuntimeDocker
}

func NewClient(cfg ClientConfig) (*Client, error) {
	host, rt := detectHost(cfg)

	opts := []client.Opt{client.WithHost(host)}
	if cfg.APIVersion != "" {
		opts = append(opts, client.WithVersion(cfg.APIVersion))
	} else {
		opts = append(opts, client.WithAPIVersionNegotiation())
	}

	if cfg.TLS.Enabled {
		if !cfg.TLS.Verify && !cfg.TLS.InsecureSkipVerify {
			return nil, connectionError(ConnectionErrorCA, fmt.Errorf("TLS certificate verification is required"))
		}
		httpClient, err := tlsHTTPClient(cfg.TLS, host)
		if err != nil {
			return nil, err
		}
		opts = append(opts, client.WithHTTPClient(httpClient))
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
		cli.Close()
		if cfg.Host == "" && rt == RuntimePodman {
			if sock := autoStartPodmanSocket(); sock != "" {
				cli2, err2 := client.NewClientWithOpts(
					client.WithHost(sock),
					client.WithAPIVersionNegotiation(),
				)
				if err2 == nil {
					ctx2, cancel2 := context.WithTimeout(context.Background(), timeout)
					defer cancel2()
					if _, err2 = cli2.Ping(ctx2); err2 == nil {
						c := &Client{
							cli:           cli2,
							RuntimeType:   RuntimePodman,
							Host:          sock,
							EngineVersion: fetchVersion(cli2, 2*time.Second),
						}
						c.ctx, c.cancel = context.WithCancel(context.Background())
						c.initImageLister()
						return c, nil
					}
					cli2.Close()
				}
			}
		}
		return nil, fmt.Errorf("ping failed (%s): %w", host, err)
	}

	c := &Client{
		cli:           cli,
		RuntimeType:   rt,
		Host:          host,
		EngineVersion: fetchVersion(cli, 2*time.Second),
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.initImageLister()
	return c, nil
}

func tlsHTTPClient(cfg TLSConfig, host string) (*http.Client, error) {
	tlsCfg, err := tlsConfig(cfg, host)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
	}, nil
}

func tlsConfig(cfg TLSConfig, host string) (*tls.Config, error) {
	tlsCfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: cfg.InsecureSkipVerify, // Set only by explicit, validated connection config.
	}
	if cfg.ServerName != "" {
		tlsCfg.ServerName = cfg.ServerName
	} else if inferred := inferServerName(host); inferred != "" {
		tlsCfg.ServerName = inferred
	}
	if cfg.CAFile != "" {
		caPEM, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, connectionError(ConnectionErrorCA, fmt.Errorf("read CA file: %w", err))
		}
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(caPEM); !ok {
			return nil, connectionError(ConnectionErrorCA, fmt.Errorf("parse CA file: no certificates found"))
		}
		tlsCfg.RootCAs = pool
	}
	if cfg.CertFile != "" || cfg.KeyFile != "" {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			return nil, connectionError(ConnectionErrorClientCert, fmt.Errorf("TLS certFile and keyFile must be configured together"))
		}
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, connectionError(ConnectionErrorClientCert, fmt.Errorf("load TLS cert pair: %w", err))
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

func (c *Client) Close() error {
	c.cancel()
	return c.cli.Close()
}

func (c *Client) Raw() *client.Client {
	return c.cli
}

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
		Type:     runtimeapi.Type(c.RuntimeType),
		Endpoint: c.Host,
		Version:  c.EngineVersion,
	}
}

// Capabilities describes the behavior currently exposed by this facade.
func (c *Client) Capabilities() runtimeapi.CapabilitySet {
	support := runtimeapi.Available
	reason := ""
	if c.RuntimeType == RuntimePodman {
		support = runtimeapi.Degraded
		reason = "provided through the Podman Docker compatibility API"
	}
	return runtimeapi.CapabilitySet{
		runtimeapi.CapabilityContainers: {Support: support, Reason: reason},
		runtimeapi.CapabilityImages:     {Support: support, Reason: reason},
		runtimeapi.CapabilityVolumes:    {Support: support, Reason: reason},
		runtimeapi.CapabilityNetworks:   {Support: support, Reason: reason},
		runtimeapi.CapabilityEvents:     {Support: support, Reason: reason},
		runtimeapi.CapabilityExec:       {Support: support, Reason: reason},
		runtimeapi.CapabilityFiltering: {
			Support: runtimeapi.Degraded,
			Reason:  "native filtering is currently exposed only by container lists",
		},
		runtimeapi.CapabilityContainerListFilter: {Support: support, Reason: reason},
		runtimeapi.CapabilityImageListFilter: {
			Support:    runtimeapi.Unsupported,
			Reason:     "unified image list filters have not been migrated",
			ReasonCode: "adapter_not_migrated",
		},
		runtimeapi.CapabilityVolumeListFilter: {
			Support:    runtimeapi.Unsupported,
			Reason:     "unified volume list filters have not been migrated",
			ReasonCode: "adapter_not_migrated",
		},
		runtimeapi.CapabilityNetworkListFilter: {
			Support:    runtimeapi.Unsupported,
			Reason:     "unified network list filters have not been migrated",
			ReasonCode: "adapter_not_migrated",
		},
		runtimeapi.CapabilityEventFilter: {Support: support, Reason: reason},
		runtimeapi.CapabilityExecResize:  {Support: support, Reason: reason},
		runtimeapi.CapabilityStatsStream: {Support: support, Reason: reason},
	}
}

// PingTimeout verifies the runtime connection with a caller-selected deadline.
func (c *Client) PingTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(c.ctx, timeout)
	defer cancel()
	_, err := c.cli.Ping(ctx)
	return err
}

// initImageLister sets the image listing backend based on the engine type.
func (c *Client) initImageLister() {
	switch c.RuntimeType {
	case RuntimePodman:
		c.imageLister = &podmanImageLister{client: c}
	default:
		c.imageLister = &dockerImageLister{client: c}
	}
}
