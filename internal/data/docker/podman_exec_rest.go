package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	runtimepodman "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

type podmanExecService struct{ client *Client }

func (s podmanExecService) Open(ctx context.Context, containerID string, options runtimeapi.ExecOptions) (runtimeapi.ExecSession, error) {
	if s.client.podmanREST == nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorUnavailable, "container.exec.create", containerID, errPodmanRESTNotReady)
	}
	request := struct {
		AttachStdin  bool     `json:"AttachStdin"`
		AttachStdout bool     `json:"AttachStdout"`
		AttachStderr bool     `json:"AttachStderr"`
		TTY          bool     `json:"Tty"`
		Command      []string `json:"Cmd"`
		Environment  []string `json:"Env,omitempty"`
		WorkingDir   string   `json:"WorkingDir,omitempty"`
		User         string   `json:"User,omitempty"`
	}{options.AttachStdin, options.AttachStdout, options.AttachStderr, options.TTY, options.Command, options.Environment, options.WorkingDir, options.User}
	var created struct {
		ID string `json:"Id"`
	}
	path := runtimepodman.ContainerPath(containerID, "/exec")
	if err := s.client.podmanREST.Post(ctx, "container.exec.create", path, nil, request, &created); err != nil {
		return nil, mapRuntimeError(err, "container.exec.create", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: containerID}, RuntimePodman)
	}
	if created.ID == "" {
		return nil, runtimeapi.NewError(runtimeapi.ErrorInvalid, "container.exec.create", containerID, fmt.Errorf("Podman response omitted exec ID"))
	}
	startBody, err := json.Marshal(struct {
		Detach bool `json:"Detach"`
		TTY    bool `json:"Tty"`
	}{TTY: options.TTY})
	if err != nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorInvalid, "container.exec.start", created.ID, err)
	}
	stream, err := s.client.podmanREST.UpgradePost(ctx, "container.exec.start", runtimepodman.ExecPath(created.ID, "/start"), bytes.NewReader(startBody), "application/json")
	if err != nil {
		return nil, mapRuntimeError(err, "container.exec.start", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: containerID}, RuntimePodman)
	}
	return &podmanExecSession{client: s.client, id: created.ID, containerID: containerID, stream: stream}, nil
}

type podmanExecSession struct {
	client      *Client
	id          string
	containerID string
	stream      io.ReadWriteCloser
}

func (s *podmanExecSession) ID() string                       { return s.id }
func (s *podmanExecSession) Read(buffer []byte) (int, error)  { return s.stream.Read(buffer) }
func (s *podmanExecSession) Write(buffer []byte) (int, error) { return s.stream.Write(buffer) }
func (s *podmanExecSession) Close() error                     { return s.stream.Close() }

func (s *podmanExecSession) Resize(ctx context.Context, size runtimeapi.TerminalSize) error {
	query := url.Values{
		"h": {strconv.FormatUint(uint64(size.Height), 10)},
		"w": {strconv.FormatUint(uint64(size.Width), 10)},
	}
	err := s.client.podmanREST.Post(ctx, "container.exec.resize", runtimepodman.ExecPath(s.id, "/resize"), query, nil, nil)
	return mapRuntimeError(err, "container.exec.resize", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: s.containerID}, RuntimePodman)
}
