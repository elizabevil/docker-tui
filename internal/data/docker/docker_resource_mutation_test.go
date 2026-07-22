package docker

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	engineclient "github.com/docker/docker/client"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func TestDockerVolumeAndNetworkCreatePrune(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := ""
		switch {
		case strings.HasSuffix(request.URL.Path, "/volumes/create"):
			body = `{"Name":"cache","Driver":"local"}`
		case strings.HasSuffix(request.URL.Path, "/volumes/prune"):
			body = `{"VolumesDeleted":["old-cache"],"SpaceReclaimed":128}`
		case strings.HasSuffix(request.URL.Path, "/networks/create"):
			body = `{"Id":"network-id","Warning":""}`
		case strings.HasSuffix(request.URL.Path, "/networks/prune"):
			body = `{"NetworksDeleted":["old-network"]}`
		default:
			t.Fatalf("unexpected path %q", request.URL.Path)
		}
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})}
	raw, err := engineclient.NewClientWithOpts(engineclient.WithHost("http://runtime.test"), engineclient.WithHTTPClient(httpClient), engineclient.WithVersion("1.45"))
	if err != nil {
		t.Fatalf("NewClientWithOpts() error = %v", err)
	}
	client := &Client{cli: raw, ctx: context.Background(), RuntimeType: RuntimeDocker}
	volume, err := client.Volumes().Create(context.Background(), runtimeapi.VolumeCreateOptions{Name: "cache"})
	if err != nil || volume.Name != "cache" {
		t.Fatalf("volume create = %#v, %v", volume, err)
	}
	volumePrune, err := client.Volumes().Prune(context.Background(), runtimeapi.PruneOptions{})
	if err != nil || volumePrune.SpaceReclaimed != 128 || len(volumePrune.Resources) != 1 {
		t.Fatalf("volume prune = %#v, %v", volumePrune, err)
	}
	network, err := client.Networks().Create(context.Background(), runtimeapi.NetworkCreateOptions{Name: "frontend"})
	if err != nil || network.ID != "network-id" {
		t.Fatalf("network create = %#v, %v", network, err)
	}
	networkPrune, err := client.Networks().Prune(context.Background(), runtimeapi.PruneOptions{})
	if err != nil || len(networkPrune.Resources) != 1 || networkPrune.Resources[0].ID != "old-network" {
		t.Fatalf("network prune = %#v, %v", networkPrune, err)
	}
}
