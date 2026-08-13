package exec

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type CLIExecutor interface {
	Execute(ctx context.Context, cliPath, containerID, shell string) error
}

type binaryExecutor struct{}

func NewBinaryExecutor() CLIExecutor {
	return &binaryExecutor{}
}

func (e *binaryExecutor) Execute(ctx context.Context, cliPath, containerID, shell string) error {
	args := []string{"exec", "-it", containerID, shell}

	cmd := exec.CommandContext(ctx, cliPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

type sessionAdapter struct {
	cmd         *exec.Cmd
	cliPath     string
	containerID string
	shell       string
}

func (s *sessionAdapter) ID() string {
	return fmt.Sprintf("cli-%s-%s", s.cliPath, s.containerID)
}

func (s *sessionAdapter) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("CLI session does not support read")
}

func (s *sessionAdapter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("CLI session does not support write")
}

func (s *sessionAdapter) Close() error {
	return nil
}

func (s *sessionAdapter) Resize(ctx context.Context, size runtimeapi.TerminalSize) error {
	return nil
}
