package exec

import (
	"context"
	"fmt"
	"os/exec"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type CLIType string

const (
	CLIDocker CLIType = "docker"
	CLIPodman CLIType = "podman"
)

type DetectionResult struct {
	CLIType    CLIType
	CLIPath    string
	Endpoint   string
	MatchError error
}

type CLIDetector interface {
	Detect(ctx context.Context, engineType string, endpoint string) *DetectionResult
}

type binaryDetector struct{}

func NewBinaryDetector() CLIDetector {
	return &binaryDetector{}
}

func (d *binaryDetector) Detect(ctx context.Context, engineType string, endpoint string) *DetectionResult {
	switch runtimeapi.RuntimeType(engineType) {
	case runtimeapi.RuntimeDocker:
		return d.detectDocker(ctx, endpoint)
	case runtimeapi.RuntimePodman:
		return d.detectPodman(ctx, endpoint)
	default:
		return &DetectionResult{
			MatchError: fmt.Errorf("unsupported runtime type: %s", engineType),
		}
	}
}

func (d *binaryDetector) detectDocker(ctx context.Context, endpoint string) *DetectionResult {
	path, err := exec.LookPath("docker")
	if err != nil {
		return &DetectionResult{
			CLIType:    CLIDocker,
			MatchError: fmt.Errorf("docker CLI not found in PATH"),
		}
	}

	if !d.matchEndpoint(endpoint, "docker") {
		return &DetectionResult{
			CLIType:    CLIDocker,
			CLIPath:    path,
			Endpoint:   endpoint,
			MatchError: fmt.Errorf("docker endpoint does not match current engine"),
		}
	}

	return &DetectionResult{
		CLIType:  CLIDocker,
		CLIPath:  path,
		Endpoint: endpoint,
	}
}

func (d *binaryDetector) detectPodman(ctx context.Context, endpoint string) *DetectionResult {
	path, err := exec.LookPath("podman")
	if err != nil {
		return &DetectionResult{
			CLIType:    CLIPodman,
			MatchError: fmt.Errorf("podman CLI not found in PATH"),
		}
	}

	if !d.matchEndpoint(endpoint, "podman") {
		return &DetectionResult{
			CLIType:    CLIPodman,
			CLIPath:    path,
			Endpoint:   endpoint,
			MatchError: fmt.Errorf("podman endpoint does not match current engine"),
		}
	}

	return &DetectionResult{
		CLIType:  CLIPodman,
		CLIPath:  path,
		Endpoint: endpoint,
	}
}

func (d *binaryDetector) matchEndpoint(endpoint, cliName string) bool {
	if endpoint == "" {
		return true
	}

	switch cliName {
	case "docker":
		return endpoint == "unix:///var/run/docker.sock" ||
			endpoint == "unix:///run/docker.sock"
	case "podman":
		return true
	}

	return false
}
