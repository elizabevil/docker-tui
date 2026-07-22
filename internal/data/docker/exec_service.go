package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type execService struct{ client *Client }

func (c *Client) Exec() runtimeapi.ExecService { return execService{client: c} }

func (s execService) Open(ctx context.Context, containerID string, options runtimeapi.ExecOptions) (runtimeapi.ExecSession, error) {
	created, err := s.client.cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd: options.Command, Env: options.Environment, WorkingDir: options.WorkingDir, User: options.User,
		Tty: options.TTY, AttachStdin: options.AttachStdin, AttachStdout: options.AttachStdout, AttachStderr: options.AttachStderr,
	})
	if err != nil {
		return nil, mapRuntimeError(err, "container.exec.create", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: containerID}, s.client.RuntimeType)
	}
	response, err := s.client.cli.ContainerExecAttach(ctx, created.ID, container.ExecAttachOptions{Tty: options.TTY})
	if err != nil {
		return nil, mapRuntimeError(err, "container.exec.attach", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: containerID}, s.client.RuntimeType)
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
		return mapRuntimeError(err, "container.exec.resize", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: s.containerID}, s.client.RuntimeType)
	}
	return nil
}
