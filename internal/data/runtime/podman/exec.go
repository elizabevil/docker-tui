package podman

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
)

// ExecService implements runtime.ExecService for the Podman adapter.
type ExecService struct {
	Client *Client
}

// Open creates an exec session inside a running container and returns an
// interactive I/O stream.
func (s ExecService) Open(ctx context.Context, containerID string, options runtimeapi.ExecOptions) (runtimeapi.ExecSession, error) {
	if s.Client.REST == nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorUnavailable, "container.exec.create", containerID, errPodmanRESTNotReady)
	}
	request := dto.ExecCreateRequest{
		AttachStdin: options.AttachStdin, AttachStdout: options.AttachStdout,
		AttachStderr: options.AttachStderr, TTY: options.TTY, Command: options.Command,
		Environment: options.Environment, WorkingDir: options.WorkingDir, User: options.User,
	}
	var created dto.ExecCreateResponse
	path := fmt.Sprintf(PathContainerExec, containerID)
	if err := s.Client.REST.Post(ctx, "container.exec.create", path, nil, request, &created); err != nil {
		return nil, mapRuntimeError(err, "container.exec.create", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: containerID}, runtimeapi.Podman)
	}
	if created.ID == "" {
		return nil, runtimeapi.NewError(runtimeapi.ErrorInvalid, "container.exec.create", containerID, fmt.Errorf("Podman response omitted exec ID"))
	}
	startBody, err := json.Marshal(dto.ExecStartRequest{TTY: options.TTY})
	if err != nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorInvalid, "container.exec.start", created.ID, err)
	}
	stream, err := s.Client.REST.UpgradePost(ctx, "container.exec.start", fmt.Sprintf(PathExecStart, created.ID), bytes.NewReader(startBody), "application/json")
	if err != nil {
		return nil, mapRuntimeError(err, "container.exec.start", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: containerID}, runtimeapi.Podman)
	}
	return &execSession{client: s.Client, id: created.ID, containerID: containerID, stream: stream}, nil
}

type execSession struct {
	client      *Client
	id          string
	containerID string
	stream      io.ReadWriteCloser
}

func (s *execSession) ID() string                       { return s.id }
func (s *execSession) Read(buffer []byte) (int, error)  { return s.stream.Read(buffer) }
func (s *execSession) Write(buffer []byte) (int, error) { return s.stream.Write(buffer) }
func (s *execSession) Close() error                     { return s.stream.Close() }

func (s *execSession) Resize(ctx context.Context, size runtimeapi.TerminalSize) error {
	query := url.Values{
		"h": {strconv.FormatUint(uint64(size.Height), 10)},
		"w": {strconv.FormatUint(uint64(size.Width), 10)},
	}
	err := s.client.REST.Post(ctx, "container.exec.resize", fmt.Sprintf(PathExecResize, s.id), query, nil, nil)
	return mapRuntimeError(err, "container.exec.resize", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: s.containerID}, runtimeapi.Podman)
}
