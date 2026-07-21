package docker

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/docker/docker/client"
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
	Host    string
	TLS     TLSConfig
	Timeout time.Duration
	Runtime string // "docker", "podman", or "" for auto-detect
}

type TLSConfig struct {
	Enabled  bool
	Verify   bool
	CAFile   string
	CertFile string
	KeyFile  string
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
		if cfg.Runtime == string(RuntimeDocker) {
			runtimeType = RuntimeDocker
		} else if cfg.Runtime == string(RuntimePodman) {
			runtimeType = RuntimePodman
		}
		return cfg.Host, runtimeType
	}
	// Runtime without an endpoint selects the corresponding local socket.
	if cfg.Runtime != "" {
		switch cfg.Runtime {
		case "docker":
			return fmt.Sprintf("unix://%s", "/var/run/docker.sock"), RuntimeDocker
		case "podman":
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

	opts := []client.Opt{
		client.WithHost(host),
		client.WithAPIVersionNegotiation(),
	}

	if cfg.TLS.Enabled {
		if !cfg.TLS.Verify {
			return nil, fmt.Errorf("TLS certificate verification is required")
		}
		opts = append(opts, client.WithTLSClientConfig(cfg.TLS.CAFile, cfg.TLS.CertFile, cfg.TLS.KeyFile))
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
	ctx, cancel := context.WithTimeout(c.ctx, 2*time.Second)
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
