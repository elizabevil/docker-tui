package docker

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	podmanapi "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

func TestExecuteContainerPodmanRESTMapsActions(t *testing.T) {
	tests := []struct {
		name       string
		action     runtimeapi.Action
		options    runtimeapi.ActionOptions
		wantMethod string
		wantPath   string
		wantQuery  string
	}{
		{name: "start", action: runtimeapi.ActionStart, wantMethod: http.MethodPost, wantPath: "/v5.0.0/libpod/containers/abc/start"},
		{name: "stop", action: runtimeapi.ActionStop, options: runtimeapi.ActionOptions{Timeout: 5 * time.Second}, wantMethod: http.MethodPost, wantPath: "/v5.0.0/libpod/containers/abc/stop", wantQuery: "timeout=5"},
		{name: "rename", action: runtimeapi.ActionRename, options: runtimeapi.ActionOptions{Name: "api"}, wantMethod: http.MethodPost, wantPath: "/v5.0.0/libpod/containers/abc/rename", wantQuery: "name=api"},
		{name: "remove", action: runtimeapi.ActionRemove, options: runtimeapi.ActionOptions{Force: true}, wantMethod: http.MethodDelete, wantPath: "/v5.0.0/libpod/containers/abc", wantQuery: "force=true"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if request.Method != test.wantMethod || request.URL.Path != test.wantPath || request.URL.RawQuery != test.wantQuery {
					t.Fatalf("request = %s %s?%s", request.Method, request.URL.Path, request.URL.RawQuery)
				}
				return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(&emptyReader{}), Header: make(http.Header), Request: request}, nil
			})}
			rest, err := podmanapi.NewRESTClient(podmanapi.RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.0.0", HTTPClient: httpClient})
			if err != nil {
				t.Fatal(err)
			}
			client := &Client{podmanREST: rest}
			if err := client.executeContainerPodmanREST(context.Background(), "abc", test.action, test.options); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type emptyReader struct{}

func (*emptyReader) Read([]byte) (int, error) { return 0, io.EOF }
