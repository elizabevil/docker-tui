package runtimeinit

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	dockeradapter "github.com/elizabevil/docker-tui/internal/data/runtime/docker"
	podmanadapter "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
	"github.com/elizabevil/docker-tui/internal/driver/podman"
	"github.com/elizabevil/docker-tui/internal/utils"
)

var knownSockets = []struct {
	path string
	rt   runtimeapi.RuntimeType
}{
	{runtimeapi.DefaultDockerSocket, runtimeapi.RuntimeDocker},
	{podmanadapter.DefaultPodmanSocket, runtimeapi.RuntimePodman},
}

// NewEngineFactory returns the default EngineFactory that dispatches to the
// docker or podman adapter based on ClientConfig.Runtime. Wire it into the
// ConnectionPool at startup:
//
//	pool := runtimeapi.NewPool(runtimeinit.NewEngineFactory())
//
// The factory also performs host auto-detection (DOCKER_HOST env, default
// sockets, podman auto-start) when an entry omits an explicit Host.
func NewEngineFactory() runtimeapi.EngineFactory {
	return newEngine
}

func isSocketLive(path string) bool {
	conn, err := net.DialTimeout("unix", path, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func detectRuntimeType(host string) runtimeapi.RuntimeType {
	if host == "" {
		return runtimeapi.RuntimeDocker
	}
	parsed, err := url.Parse(host)
	if err != nil {
		return runtimeapi.RuntimeDocker
	}
	if strings.HasSuffix(parsed.Path, "podman.sock") {
		return runtimeapi.RuntimePodman
	}
	return runtimeapi.RuntimeDocker
}

func detectHost(cfg runtimeapi.ClientConfig) (string, runtimeapi.RuntimeType) {
	if cfg.Host != "" {
		rt := detectRuntimeType(cfg.Host)
		if cfg.Runtime == runtimeapi.RuntimeDocker {
			rt = runtimeapi.RuntimeDocker
		} else if cfg.Runtime == runtimeapi.RuntimePodman {
			rt = runtimeapi.RuntimePodman
		}
		return cfg.Host, rt
	}
	if cfg.Runtime != "" {
		switch cfg.Runtime {
		case runtimeapi.RuntimeDocker:
			return utils.SocketURI(runtimeapi.DefaultDockerSocket), runtimeapi.RuntimeDocker
		case runtimeapi.RuntimePodman:
			if sock := firstLivePodmanSocket(); sock != "" {
				return utils.SocketURI(sock), runtimeapi.RuntimePodman
			}
		}
	}
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		return host, detectRuntimeType(host)
	}
	for _, s := range knownSockets {
		if isSocketLive(s.path) {
			return utils.SocketURI(s.path), s.rt
		}
	}
	if sock := firstLivePodmanSocket(); sock != "" {
		return utils.SocketURI(sock), runtimeapi.RuntimePodman
	}
	if sock := autoStartPodmanSocket(); sock != "" {
		return sock, runtimeapi.RuntimePodman
	}
	return utils.SocketURI(runtimeapi.DefaultDockerSocket), runtimeapi.RuntimeDocker
}

func firstLivePodmanSocket() string {
	sock := podmanadapter.PodmanUserEndpoint(os.Getuid())
	if sock != "" {
		if u, err := url.Parse(sock); err == nil && u.Scheme == "unix" {
			if isSocketLive(u.Path) {
				return u.Path
			}
		}
	}
	if isSocketLive(podmanadapter.DefaultPodmanSocket) {
		return podmanadapter.DefaultPodmanSocket
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

	sockDir := fmt.Sprintf("/run/user/%d/podman", uid)
	sockPath := sockDir + "/podman.sock"

	if err := os.MkdirAll(sockDir, 0o700); err != nil {
		return ""
	}

	sockURI := utils.SocketURI(sockPath)
	cmd := exec.Command(podmanBin, "system", "service", "--time=0", sockURI)
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

func newEngine(config runtimeapi.ClientConfig) (runtimeapi.Engine, error) {
	host, rt := detectHost(config)

	switch rt {
	case runtimeapi.RuntimePodman:
		return sanitizeEngine(newPodmanEngine(host, config))
	default:
		return sanitizeEngine(newDockerEngine(host, config))
	}
}

// sanitizeEngine collapses the classic typed-nil interface trap: when a
// concrete adapter returns (nilPtr, err), assigning to the Engine interface
// yields a non-nil interface whose dynamic value is a nil pointer. Callers
// that compare against nil pass the check, then panic when they invoke
// methods. Returning an explicit nil interface here keeps the pool safe.
func sanitizeEngine(e runtimeapi.Engine, err error) (runtimeapi.Engine, error) {
	if e == nil {
		return nil, err
	}
	v := reflect.ValueOf(e)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil, err
	}
	return e, err
}

func newPodmanEngine(host string, config runtimeapi.ClientConfig) (runtimeapi.Engine, error) {
	rest, err := podman.NewRESTClient(podman.RESTConfig{
		Endpoint:   host,
		APIVersion: config.APIVersion,
		TLS:        config.TLS.Enabled,
	})
	if err != nil {
		return nil, fmt.Errorf("create Podman transport: %w", err)
	}
	return podmanadapter.NewPodmanEngine(podmanadapter.PodmanEngineConfig{
		REST: rest,
		Host: host,
		TLS:  config.TLS.Enabled,
	})
}

func newDockerEngine(host string, config runtimeapi.ClientConfig) (runtimeapi.Engine, error) {
	dockerCfg := runtimeapi.ClientConfig{
		Host:       host,
		APIVersion: config.APIVersion,
		TLS:        config.TLS,
		Timeout:    config.Timeout,
	}
	return dockeradapter.NewClient(dockerCfg)
}
