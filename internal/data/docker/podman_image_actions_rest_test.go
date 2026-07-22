package docker

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	podmanapi "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

func TestExecuteImagePodmanRESTMapsActions(t *testing.T) {
	tests := []struct {
		name       string
		action     runtimeapi.Action
		options    runtimeapi.ActionOptions
		wantMethod string
		wantPath   string
		wantQuery  string
		response   string
	}{
		{name: "remove", action: runtimeapi.ActionRemove, options: runtimeapi.ActionOptions{Force: true}, wantMethod: http.MethodDelete, wantPath: "/v5.0.0/libpod/images/abc", wantQuery: "force=true"},
		{name: "pull", action: runtimeapi.ActionPull, wantMethod: http.MethodPost, wantPath: "/v5.0.0/libpod/images/pull", wantQuery: "reference=abc", response: "{\"status\":\"pulling\"}\n{\"status\":\"done\"}\n"},
		{name: "prune", action: runtimeapi.ActionPrune, wantMethod: http.MethodPost, wantPath: "/v5.0.0/libpod/images/prune", response: `[{"Id":"abc","Size":42}]`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if request.Method != test.wantMethod || request.URL.Path != test.wantPath || request.URL.RawQuery != test.wantQuery {
					t.Fatalf("request = %s %s?%s", request.Method, request.URL.Path, request.URL.RawQuery)
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(test.response)), Header: make(http.Header), Request: request}, nil
			})}
			rest, err := podmanapi.NewRESTClient(podmanapi.RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.0.0", HTTPClient: httpClient})
			if err != nil {
				t.Fatal(err)
			}
			result := runtimeapi.ActionResult{}
			if err := (&Client{podmanREST: rest}).executeImagePodmanREST(context.Background(), "abc", test.action, test.options, &result); err != nil {
				t.Fatal(err)
			}
			if test.action == runtimeapi.ActionPrune && result.SpaceReclaimed != 42 {
				t.Fatalf("space reclaimed = %d", result.SpaceReclaimed)
			}
		})
	}
}
