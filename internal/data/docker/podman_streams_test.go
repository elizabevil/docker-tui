package docker

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/docker/docker/pkg/stdcopy"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	podmanapi "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

func TestPodmanEventServiceMapsLibpodStream(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := io.NopCloser(bytes.NewBufferString(`{"Type":"container","Status":"start","ID":"abc","Name":"api","TimeNano":2000000000}` + "\n"))
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: make(http.Header), Request: request}, nil
	})}
	rest, err := podmanapi.NewRESTClient(podmanapi.RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.0.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	items, err := (podmanEventService{client: &Client{podmanREST: rest}}).Subscribe(context.Background(), runtimeapi.EventOptions{})
	if err != nil {
		t.Fatal(err)
	}
	item := <-items
	if item.Error != nil || item.Event.ResourceType != "container" || item.Event.Action != "start" || item.Event.ActorID != "abc" || item.Event.Attributes["name"] != "api" {
		t.Fatalf("event item = %#v", item)
	}
}

func TestPodmanContainerLogsDemultiplexesStream(t *testing.T) {
	var stream bytes.Buffer
	stdout := stdcopy.NewStdWriter(&stream, stdcopy.Stdout)
	_, _ = stdout.Write([]byte("hello\n"))
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(stream.Bytes())), Header: make(http.Header), Request: request}, nil
	})}
	rest, err := podmanapi.NewRESTClient(podmanapi.RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.0.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	reader, err := (podmanContainerService{client: &Client{podmanREST: rest}}).Logs(context.Background(), "abc", runtimeapi.ContainerLogOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var output bytes.Buffer
	_, _ = output.ReadFrom(reader)
	if output.String() != "hello\n" {
		t.Fatalf("logs = %q", output.String())
	}
}
