package docker

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	podmanapi "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

func TestPodmanVolumeCreateAndPruneREST(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := ""
		switch request.URL.Path {
		case "/v5.0.0/libpod/volumes/create":
			body = `{"Name":"cache","Driver":"local"}`
		case "/v5.0.0/libpod/volumes/prune":
			if !strings.Contains(request.URL.Query().Get("filters"), "label") {
				t.Fatalf("filters = %q", request.URL.Query().Get("filters"))
			}
			body = `[{"Id":"deleted","Size":64,"Err":null},{"Id":"busy","Err":"in use"}]`
		default:
			t.Fatalf("unexpected path %q", request.URL.Path)
		}
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})}
	rest, err := podmanapi.NewRESTClient(podmanapi.RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.0.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatalf("NewRESTClient() error = %v", err)
	}
	client := &Client{podmanREST: rest}
	created, err := client.createVolumePodmanREST(context.Background(), runtimeapi.VolumeCreateOptions{Name: "cache"})
	if err != nil || created.Name != "cache" {
		t.Fatalf("create = %#v, %v", created, err)
	}
	result, err := client.pruneVolumesPodmanREST(context.Background(), runtimeapi.PruneOptions{Filters: runtimeapi.FilterSet{"label": {"temporary=true"}}})
	if err != nil {
		t.Fatalf("prune error = %v", err)
	}
	succeeded, failed := result.Counts()
	if succeeded != 1 || failed != 1 || result.SpaceReclaimed != 64 {
		t.Fatalf("prune result = %#v", result)
	}
}
