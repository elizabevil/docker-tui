package docker

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/image"
	dockerapi "github.com/docker/docker/client"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func TestMapDockerHistoryEmpty(t *testing.T) {
	if got := mapDockerHistory(nil); len(got) != 0 {
		t.Errorf("nil input = %d layers, want 0", len(got))
	}
	if got := mapDockerHistory([]image.HistoryResponseItem{}); len(got) != 0 {
		t.Errorf("empty input = %d layers, want 0", len(got))
	}
}

func TestMapDockerHistoryFieldPassthrough(t *testing.T) {
	raw := []image.HistoryResponseItem{
		{ID: "abc", Created: 1700000000, CreatedBy: "RUN apt-get install vim", Size: 12345, Comment: "first layer"},
		{ID: "def", Created: 1700001000, CreatedBy: `CMD ["/bin/sh"]`, Size: 0, Comment: ""},
	}
	got := mapDockerHistory(raw)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != "abc" || got[0].Created != 1700000000 || got[0].Size != 12345 || got[0].Comment != "first layer" {
		t.Errorf("got[0] = %+v", got[0])
	}
	if got[0].CreatedBy != "RUN apt-get install vim" {
		t.Errorf("got[0].CreatedBy = %q", got[0].CreatedBy)
	}
	if got[1].ID != "def" || got[1].Size != 0 || got[1].Comment != "" {
		t.Errorf("got[1] = %+v", got[1])
	}
}

func TestMapDockerHistoryReturnsRuntimeType(t *testing.T) {
	raw := []image.HistoryResponseItem{{ID: "x", Created: 1, Size: 2, Comment: "c"}}
	got := mapDockerHistory(raw)
	var _ = got
}

func TestDockerHistoryForwardsSummaryID(t *testing.T) {
	var requestedPath string
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requestedPath = r.URL.Path
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`[{"Id":"layer-one","CreatedBy":"FROM scratch"}]`)),
			Request:    r,
		}, nil
	})}

	cli, err := dockerapi.NewClientWithOpts(dockerapi.WithHost("http://engine.test"), dockerapi.WithHTTPClient(httpClient), dockerapi.WithVersion("1.44"))
	if err != nil {
		t.Fatalf("create SDK client: %v", err)
	}
	defer cli.Close()
	svc := (&Client{cli: cli}).Images()
	summary := runtimeapi.ImageSummary{ID: "summary-id-passthrough", RepoTags: []string{"alpine:latest"}}
	layers, err := svc.History(context.Background(), summary)
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if !strings.Contains(requestedPath, "/images/summary-id-passthrough/history") {
		t.Fatalf("request path = %q", requestedPath)
	}
	if len(layers) != 1 || layers[0].ID != "layer-one" {
		t.Fatalf("layers = %#v", layers)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
