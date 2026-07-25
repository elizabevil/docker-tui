package podman

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/driver/podman"
	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

type PodmanExecService struct {
	Client *podman.Client
}

func (s PodmanExecService) Open(ctx context.Context, containerID string, options runtime.ExecOptions) (runtime.ExecSession, error) {
	if s.Client.REST == nil {
		return nil, runtime.NewError(runtime.ErrorUnavailable, "container.exec.create", containerID, podman.ErrPodmanRESTNotReady)
	}
	request := dto.ExecCreateRequest{
		AttachStdin:  options.AttachStdin,
		AttachStdout: options.AttachStdout,
		AttachStderr: options.AttachStderr,
		TTY:          options.TTY,
		Command:      options.Command,
		Environment:  options.Environment,
		WorkingDir:   options.WorkingDir,
		User:         options.User,
	}
	var created dto.ExecCreateResponse
	if err := s.Client.REST.Post(ctx, podman.ContainerExecPath(containerID), nil, request, &created); err != nil {
		return nil, mapPodmanContainerErr(err, "exec.create", containerID)
	}
	if created.ID == "" {
		return nil, runtime.NewError(runtime.ErrorInvalid, "container.exec.create", containerID, fmt.Errorf("Podman response omitted exec ID"))
	}
	startBody, err := json.Marshal(dto.ExecStartRequest{TTY: options.TTY})
	if err != nil {
		return nil, runtime.NewError(runtime.ErrorInvalid, "container.exec.start", created.ID, err)
	}
	stream, err := s.Client.REST.UpgradePost(ctx, podman.ExecStartPath(created.ID), bytes.NewReader(startBody), "application/json")
	if err != nil {
		return nil, mapPodmanContainerErr(err, "exec.start", containerID)
	}
	return &podmanExecSession{client: s.Client, id: created.ID, containerID: containerID, stream: stream}, nil
}

type podmanExecSession struct {
	client      *podman.Client
	id          string
	containerID string
	stream      io.ReadWriteCloser
}

func (s *podmanExecSession) ID() string                       { return s.id }
func (s *podmanExecSession) Read(buffer []byte) (int, error)  { return s.stream.Read(buffer) }
func (s *podmanExecSession) Write(buffer []byte) (int, error) { return s.stream.Write(buffer) }
func (s *podmanExecSession) Close() error                     { return s.stream.Close() }

func (s *podmanExecSession) Resize(ctx context.Context, size runtime.TerminalSize) error {
	query := url.Values{
		"h": {strconv.FormatUint(uint64(size.Height), 10)},
		"w": {strconv.FormatUint(uint64(size.Width), 10)},
	}
	err := s.client.REST.Post(ctx, podman.ExecResizePath(s.id), query, nil, nil)
	return mapPodmanContainerErr(err, "exec.resize", s.containerID)
}
