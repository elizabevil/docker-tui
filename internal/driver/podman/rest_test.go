package podman

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestRESTClientNegotiatesVersionAndUsesVersionedPath(t *testing.T) {
	requests := make([]string, 0, 2)
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests = append(requests, request.URL.RequestURI())
		switch request.URL.Path {
		case PathSystemVersion:
			return response(http.StatusOK, `{"ApiVersion":"5.2.0"}`), nil
		case "/v5.2.0/libpod/images/json":
			return response(http.StatusOK, `[{"Id":"sha256:123"}]`), nil
		default:
			return response(http.StatusNotFound, `{}`), nil
		}
	})}

	client, err := NewRESTClient(RESTConfig{Endpoint: "http://podman.test", HTTPClient: httpClient})
	if err != nil {
		t.Fatalf("NewRESTClient() error = %v", err)
	}
	var images []struct {
		ID string `json:"Id"`
	}
	if err := client.Get(context.Background(), "/libpod/images/json", nil, &images); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if len(images) != 1 || images[0].ID != "sha256:123" {
		t.Fatalf("unexpected images: %#v", images)
	}
	if len(requests) != 2 || requests[1] != "/v5.2.0/libpod/images/json" {
		t.Fatalf("unexpected requests: %v", requests)
	}
}

func TestRESTClientUsesConfiguredVersionWithoutNegotiation(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/v4.9.0/libpod/info" {
			t.Fatalf("unexpected path %q", request.URL.Path)
		}
		return response(http.StatusOK, `{}`), nil
	})}
	client, err := NewRESTClient(RESTConfig{Endpoint: "http://podman.test", APIVersion: "v4.9.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatalf("NewRESTClient() error = %v", err)
	}
	if err := client.Get(context.Background(), "/libpod/info", nil, &struct{}{}); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
}

func TestRESTClientMapsHTTPError(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return response(http.StatusConflict, `{"message":"container is running"}`), nil
	})}
	client, err := NewRESTClient(RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.0.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatalf("NewRESTClient() error = %v", err)
	}
	err = client.Get(context.Background(), "/libpod/containers/id", nil, &struct{}{})
	if !IsPodmanError(err, KindInvalid) {
		t.Fatalf("expected conflict, got %v", err)
	}
	var pe *Error
	if !errors.As(err, &pe) || pe.StatusCode != http.StatusConflict || pe.Err == nil {
		t.Fatalf("unexpected podman error: %#v", pe)
	}
}

func TestRESTClientStreamMapsHTTPError(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return response(http.StatusNotFound, `{"message":"missing image"}`), nil
	})}
	client, err := NewRESTClient(RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.0.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.StreamGet(context.Background(), "/libpod/images/export", nil)
	if !IsPodmanError(err, KindNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestRESTClientStreamUsesContextCancellation(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		return nil, request.Context().Err()
	})}
	client, err := NewRESTClient(RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.0.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.StreamPost(ctx, "/libpod/images/abc/push", nil, nil, "")
	if !IsPodmanError(err, KindConnection) {
		t.Fatalf("expected canceled, got %v", err)
	}
}

func TestRESTClientDeleteUsesCorrectMethodAndPath(t *testing.T) {
	var receivedMethod string
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		receivedMethod = request.Method
		if request.URL.Path != "/v5.2.0/libpod/volumes/test-vol" {
			t.Fatalf("unexpected path %q", request.URL.Path)
		}
		return response(http.StatusNoContent, ""), nil
	})}
	client, err := NewRESTClient(RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.2.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatalf("NewRESTClient() error = %v", err)
	}
	if err := client.Delete(context.Background(), "/libpod/volumes/test-vol"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if receivedMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", receivedMethod)
	}
}

func TestRESTClientDeletePreservesEscapedSegmentAndQuery(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.EscapedPath() != "/v5.2.0/libpod/volumes/team%2Fcache" {
			t.Fatalf("escaped path = %q", request.URL.EscapedPath())
		}
		if request.URL.Query().Get("force") != "true" {
			t.Fatalf("query = %q", request.URL.RawQuery)
		}
		return response(http.StatusNoContent, ""), nil
	})}
	client, err := NewRESTClient(RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.2.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatalf("NewRESTClient() error = %v", err)
	}
	query := make(url.Values)
	query.Set("force", "true")
	if err := client.DeleteWithQuery(context.Background(), "/libpod/volumes/team%2Fcache", query); err != nil {
		t.Fatalf("DeleteWithQuery() error = %v", err)
	}
}

func TestRESTClientPostEncodesJSONAndQuery(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost || request.URL.Path != "/v5.2.0/libpod/volumes/create" {
			t.Fatalf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Content-Type") != "application/json" || request.URL.Query().Get("dryrun") != "false" {
			t.Fatalf("headers/query = %#v %q", request.Header, request.URL.RawQuery)
		}
		body, err := io.ReadAll(request.Body)
		if err != nil || string(body) != `{"Name":"cache"}` {
			t.Fatalf("body = %q, error = %v", body, err)
		}
		return response(http.StatusCreated, `{"Name":"cache"}`), nil
	})}
	client, err := NewRESTClient(RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.2.0", HTTPClient: httpClient})
	if err != nil {
		t.Fatalf("NewRESTClient() error = %v", err)
	}
	query := url.Values{"dryrun": {"false"}}
	var output struct {
		Name string `json:"Name"`
	}
	if err := client.Post(context.Background(), "/libpod/volumes/create", query, struct {
		Name string `json:"Name"`
	}{Name: "cache"}, &output); err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if output.Name != "cache" {
		t.Fatalf("output = %#v", output)
	}
}

func TestPostLongRunningDisablesClientTimeout(t *testing.T) {
	var hadDeadline bool
	httpClient := &http.Client{
		Timeout: time.Millisecond,
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			_, hadDeadline = request.Context().Deadline()
			return response(http.StatusOK, `{"StatusCode":0}`), nil
		}),
	}
	client, err := NewRESTClient(RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.4.2", HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	var output dto.ContainerWaitResponse
	if err := client.postLongRunning(context.Background(), "/libpod/containers/id/wait", nil, nil, &output); err != nil {
		t.Fatal(err)
	}
	if hadDeadline {
		t.Fatal("long-running request inherited http.Client.Timeout")
	}
}

// TestAPIVersionFallsBackOnNonNotFound verifies that when the unversioned
// /libpod/version endpoint returns a non-404 error (e.g. 400 from Podman
// 5.x), APIVersion still tries the versioned /v4.0.0/libpod/version path
// before giving up.
func TestAPIVersionFallsBackOnNonNotFound(t *testing.T) {
	var requestedPaths []string
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requestedPaths = append(requestedPaths, request.URL.Path)
		switch request.URL.Path {
		case PathSystemVersion:
			return response(http.StatusBadRequest, `{"message":"versioned path required"}`), nil
		case "/v" + minPodmanAPIVersion + PathSystemVersion:
			return response(http.StatusOK, `{"Components":[{"Name":"Podman Engine","Details":{"APIVersion":"5.4.2"}}]}`), nil
		default:
			return response(http.StatusNotFound, `{}`), nil
		}
	})}
	client, err := NewRESTClient(RESTConfig{Endpoint: "http://podman.test", HTTPClient: httpClient})
	if err != nil {
		t.Fatalf("NewRESTClient() error = %v", err)
	}
	got, err := client.APIVersion(context.Background())
	if err != nil {
		t.Fatalf("APIVersion() error = %v", err)
	}
	if got != "5.4.2" {
		t.Fatalf("APIVersion() = %q, want 5.4.2", got)
	}
	if len(requestedPaths) != 2 {
		t.Fatalf("expected two endpoint probes, got %v", requestedPaths)
	}
	if requestedPaths[0] != PathSystemVersion {
		t.Fatalf("first request should target unversioned, got %q", requestedPaths[0])
	}
	if requestedPaths[1] != "/v"+minPodmanAPIVersion+PathSystemVersion {
		t.Fatalf("second request should target versioned, got %q", requestedPaths[1])
	}
}
