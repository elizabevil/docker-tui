package docker

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	podmanapi "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

func TestPodmanExecServiceUsesLibpodUpgrade(t *testing.T) {
	stream := &readWriteCloseBuffer{Buffer: bytes.NewBufferString("output")}
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.Path {
		case "/v5.0.0/libpod/containers/abc/exec":
			body, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(body), `"Cmd":["sh"]`) || !strings.Contains(string(body), `"Tty":true`) {
				t.Fatalf("create body = %s", body)
			}
			return &http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(`{"Id":"exec-id"}`)), Header: make(http.Header), Request: request}, nil
		case "/v5.0.0/libpod/exec/exec-id/start":
			if request.Header.Get("Connection") != "Upgrade" || request.Header.Get("Upgrade") != "tcp" {
				t.Fatalf("upgrade headers = %#v", request.Header)
			}
			return &http.Response{StatusCode: http.StatusSwitchingProtocols, Body: stream, Header: make(http.Header), Request: request}, nil
		case "/v5.0.0/libpod/exec/exec-id/resize":
			if request.URL.Query().Get("h") != "24" || request.URL.Query().Get("w") != "80" {
				t.Fatalf("resize query = %q", request.URL.RawQuery)
			}
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header), Request: request}, nil
		default:
			t.Fatalf("unexpected request %s %s", request.Method, request.URL.Path)
			return nil, nil
		}
	})}
	rest, err := podmanapi.NewRESTClient(podmanapi.RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.0.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	session, err := (podmanExecService{client: &Client{podmanREST: rest}}).Open(context.Background(), "abc", runtimeapi.ExecOptions{Command: []string{"sh"}, TTY: true, AttachStdin: true, AttachStdout: true})
	if err != nil {
		t.Fatal(err)
	}
	if session.ID() != "exec-id" {
		t.Fatalf("session ID = %q", session.ID())
	}
	if err := session.Resize(context.Background(), runtimeapi.TerminalSize{Height: 24, Width: 80}); err != nil {
		t.Fatal(err)
	}
	if _, err := session.Write([]byte("input")); err != nil {
		t.Fatal(err)
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
}

type readWriteCloseBuffer struct {
	*bytes.Buffer
	closed bool
}

func (b *readWriteCloseBuffer) Close() error {
	b.closed = true
	return nil
}
