package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// execService implements runtime.ExecService for the Docker/Podman adapter.
type execService struct{ client *Client }

// Exec returns the exec service facade.
func (c *Client) Exec() runtimeapi.ExecService { return execService{client: c} }

// Open creates an exec session inside a running container and returns an
// interactive I/O stream.
func (s execService) Open(ctx context.Context, containerID string, options runtimeapi.ExecOptions) (runtimeapi.ExecSession, error) {
	created, err := s.client.cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd: options.Command, Env: options.Environment, WorkingDir: options.WorkingDir, User: options.User,
		Tty: options.TTY, AttachStdin: options.AttachStdin, AttachStdout: options.AttachStdout, AttachStderr: options.AttachStderr,
	})
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "exec.create"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: containerID}, runtimeapi.Docker)
	}
	response, err := s.client.cli.ContainerExecAttach(ctx, created.ID, container.ExecAttachOptions{Tty: options.TTY})
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "exec.attach"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: containerID}, runtimeapi.Docker)
	}
	return &dockerExecSession{client: s.client, id: created.ID, containerID: containerID, response: response}, nil
}

type dockerExecSession struct {
	client      *Client
	id          string
	containerID string
	response    types.HijackedResponse
}

func (s *dockerExecSession) ID() string { return s.id }

func (s *dockerExecSession) Read(buffer []byte) (int, error) {
	return s.response.Reader.Read(buffer)
}

func (s *dockerExecSession) Write(buffer []byte) (int, error) {
	return s.response.Conn.Write(buffer)
}

func (s *dockerExecSession) Close() error {
	s.response.Close()
	return nil
}

func (s *dockerExecSession) Resize(ctx context.Context, size runtimeapi.TerminalSize) error {
	err := s.client.cli.ContainerExecResize(ctx, s.id, container.ResizeOptions{Width: size.Width, Height: size.Height})
	if err != nil {
		return runtimeapi.MapRuntimeError(err, runtimeapi.Operation(runtimeapi.ResourceContainer, "exec.resize"), runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: s.containerID}, runtimeapi.Docker)
	}
	return nil
}
