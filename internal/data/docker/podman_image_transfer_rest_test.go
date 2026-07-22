package docker

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	podmanapi "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

func TestSplitTagTargetPreservesRegistryPort(t *testing.T) {
	repository, tag := splitTagTarget("registry.example:5000/team/api:v2")
	if repository != "registry.example:5000/team/api" || tag != "v2" {
		t.Fatalf("target = %q, %q", repository, tag)
	}
}

func TestPodmanImageLoadStreamsArchive(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "image.tar")
	if err := os.WriteFile(archive, []byte("archive"), 0o600); err != nil {
		t.Fatal(err)
	}
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		if request.Method != http.MethodPost || request.URL.Path != "/v5.0.0/libpod/images/load" || request.Header.Get("Content-Type") != "application/x-tar" || !bytes.Equal(body, []byte("archive")) {
			t.Fatalf("request = %s %s content-type=%q body=%q", request.Method, request.URL.Path, request.Header.Get("Content-Type"), body)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(`{"Names":["example/api:v1"]}`)), Header: make(http.Header), Request: request}, nil
	})}
	rest, err := podmanapi.NewRESTClient(podmanapi.RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.0.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	result, err := (imageTransferService{client: &Client{RuntimeType: RuntimePodman, podmanREST: rest}}).executePodman(context.Background(), runtimeapi.ImageTransferRequest{Operation: runtimeapi.ImageTransferLoad, Path: archive}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.References) != 1 || result.References[0] != "example/api:v1" {
		t.Fatalf("references = %#v", result.References)
	}
}

func TestPodmanImageSaveRemovesPartialFile(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "partial.tar")
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(io.MultiReader(bytes.NewBufferString("partial"), errorReader{})), Header: make(http.Header), Request: request}, nil
	})}
	rest, err := podmanapi.NewRESTClient(podmanapi.RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.0.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	_, err = (imageTransferService{client: &Client{RuntimeType: RuntimePodman, podmanREST: rest}}).executePodman(context.Background(), runtimeapi.ImageTransferRequest{Operation: runtimeapi.ImageTransferSave, Source: "abc", Path: destination}, nil)
	if err == nil {
		t.Fatal("expected save error")
	}
	if _, statErr := os.Stat(destination); !os.IsNotExist(statErr) {
		t.Fatalf("partial file remains: %v", statErr)
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestPodmanImageTagUsesLibpodEndpoint(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost || request.URL.Path != "/v5.0.0/libpod/images/sha256:id/tag" || request.URL.Query().Get("repo") != "example/api" || request.URL.Query().Get("tag") != "v2" {
			t.Fatalf("request = %s %s?%s", request.Method, request.URL.Path, request.URL.RawQuery)
		}
		return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(&emptyReader{}), Header: make(http.Header), Request: request}, nil
	})}
	rest, err := podmanapi.NewRESTClient(podmanapi.RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.0.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	service := imageTransferService{client: &Client{RuntimeType: RuntimePodman, podmanREST: rest}}
	events, err := service.Run(context.Background(), runtimeapi.ImageTransferRequest{Operation: runtimeapi.ImageTransferTag, Source: "sha256:id", Destination: "example/api:v2"})
	if err != nil {
		t.Fatal(err)
	}
	event := <-events
	if !event.Done || event.Error != nil {
		t.Fatalf("terminal event = %#v", event)
	}
}
