package podman

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	podmandriver "github.com/elizabevil/docker-tui/internal/driver/podman"
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

// newComposeService builds a PodmanComposeService wired to a mock HTTP
// transport. The configured APIVersion avoids version negotiation so every
// request targets /v5.2.0/libpod/... deterministically.
func newComposeService(t *testing.T, handler roundTripFunc) PodmanComposeService {
	t.Helper()
	rest, err := podmandriver.NewRESTClient(podmandriver.RESTConfig{
		Endpoint:   "http://podman.test",
		APIVersion: "5.2.0",
		HTTPClient: &http.Client{Transport: handler},
	})
	if err != nil {
		t.Fatalf("NewRESTClient() error = %v", err)
	}
	return PodmanComposeService{Client: &podmandriver.Client{REST: rest}}
}

const (
	composeListPath = "/v5.2.0/libpod/containers/json"
	composeCreate   = "/v5.2.0/libpod/containers/create"
	composeStart    = "/v5.2.0/libpod/containers/new123/start"
	composeWait     = "/v5.2.0/libpod/containers/new123/wait"
	composeRemove   = "/v5.2.0/libpod/containers/new123"
)

// composeListBody mirrors a Podman /containers/json response for project
// "demo": a web container (number 2, with config hash) and a db container
// (number 5) that must not affect web's container-number calculation.
const composeListBody = `[
  {"Id":"web123","Image":"nginx:latest","Names":["demo_web_1"],
   "Labels":{"com.docker.compose.project":"demo","com.docker.compose.service":"web",
             "com.docker.compose.container-number":"2","com.docker.compose.config-hash":"hash123"},
   "State":"running","Created":"2024-01-01T00:00:00Z"},
  {"Id":"db123","Image":"postgres:16","Names":["demo_db_1"],
   "Labels":{"com.docker.compose.project":"demo","com.docker.compose.service":"db",
             "com.docker.compose.container-number":"5"},
   "State":"exited","Created":"2024-01-01T00:00:00Z"}
]`

func TestComposeRunCreatesOneoffAndRemovesAfterWait(t *testing.T) {
	var createBody dto.SpecGenerator
	var requests []string
	svc := newComposeService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.Method+" "+req.URL.Path)
		switch {
		case req.Method == http.MethodGet && req.URL.Path == composeListPath:
			return response(http.StatusOK, composeListBody), nil
		case req.Method == http.MethodPost && req.URL.Path == composeCreate:
			_ = json.NewDecoder(req.Body).Decode(&createBody)
			return response(http.StatusCreated, `{"Id":"new123","Warnings":[]}`), nil
		case req.Method == http.MethodPost && req.URL.Path == composeStart:
			return response(http.StatusNoContent, ""), nil
		case req.Method == http.MethodPost && req.URL.Path == composeWait:
			return response(http.StatusOK, `{"StatusCode":0}`), nil
		case req.Method == http.MethodDelete && req.URL.Path == composeRemove:
			if got := req.URL.Query().Get("force"); got != "true" {
				t.Errorf("remove force = %q, want true", got)
			}
			return response(http.StatusNoContent, ""), nil
		default:
			return response(http.StatusNotFound, `{}`), nil
		}
	}))

	err := svc.Run(context.Background(), "demo", "web", []string{"echo", "hi"}, runtimeapi.RunOptions{RemoveAfter: true})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	want := []string{
		"GET " + composeListPath,
		"POST " + composeCreate,
		"POST " + composeStart,
		"POST " + composeWait,
		"DELETE " + composeRemove,
	}
	if strings.Join(requests, ",") != strings.Join(want, ",") {
		t.Fatalf("requests = %v, want %v", requests, want)
	}

	if createBody.Name != "demo_web_oneoff" {
		t.Errorf("create name = %q, want demo_web_oneoff", createBody.Name)
	}
	if createBody.Image != "nginx:latest" {
		t.Errorf("create image = %q, want nginx:latest", createBody.Image)
	}
	if createBody.Remove {
		t.Error("create remove = true, want false for attached run")
	}
	if len(createBody.Command) != 2 || createBody.Command[0] != "echo" || createBody.Command[1] != "hi" {
		t.Errorf("create command = %v, want [echo hi]", createBody.Command)
	}
	if createBody.Labels["com.docker.compose.project"] != "demo" {
		t.Errorf("project label = %q", createBody.Labels["com.docker.compose.project"])
	}
	if createBody.Labels["com.docker.compose.service"] != "web" {
		t.Errorf("service label = %q", createBody.Labels["com.docker.compose.service"])
	}
	if createBody.Labels["com.docker.compose.oneoff"] != "true" {
		t.Errorf("oneoff label = %q, want true", createBody.Labels["com.docker.compose.oneoff"])
	}
	if createBody.Labels["com.docker.compose.container-number"] != "3" {
		t.Errorf("container-number label = %q, want 3 (max web number 2 + 1)", createBody.Labels["com.docker.compose.container-number"])
	}
	if createBody.Labels["com.docker.compose.config-hash"] != "hash123" {
		t.Errorf("config-hash label = %q, want hash123", createBody.Labels["com.docker.compose.config-hash"])
	}
}

func TestComposeRunDetachedDelegatesRemoveToEngine(t *testing.T) {
	var createBody dto.SpecGenerator
	var requests []string
	svc := newComposeService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.Method+" "+req.URL.Path)
		switch {
		case req.Method == http.MethodGet && req.URL.Path == composeListPath:
			return response(http.StatusOK, composeListBody), nil
		case req.Method == http.MethodPost && req.URL.Path == composeCreate:
			_ = json.NewDecoder(req.Body).Decode(&createBody)
			return response(http.StatusCreated, `{"Id":"new123","Warnings":[]}`), nil
		case req.Method == http.MethodPost && req.URL.Path == composeStart:
			return response(http.StatusNoContent, ""), nil
		default:
			return response(http.StatusNotFound, `{}`), nil
		}
	}))

	err := svc.Run(context.Background(), "demo", "web", nil, runtimeapi.RunOptions{
		Detach:             true,
		RemoveAfter:        true,
		EntrypointOverride: "/bin/sh",
		EnvOverrides:       map[string]string{"FOO": "bar"},
		User:               "1000:1000",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	want := []string{
		"GET " + composeListPath,
		"POST " + composeCreate,
		"POST " + composeStart,
	}
	if strings.Join(requests, ",") != strings.Join(want, ",") {
		t.Fatalf("requests = %v, want %v (no wait/remove for detached run)", requests, want)
	}
	if !createBody.Remove {
		t.Error("create remove = false, want true for detached --rm run")
	}
	if len(createBody.Entrypoint) != 1 || createBody.Entrypoint[0] != "/bin/sh" {
		t.Errorf("entrypoint = %v, want [/bin/sh]", createBody.Entrypoint)
	}
	if createBody.Env["FOO"] != "bar" {
		t.Errorf("env = %v, want FOO=bar", createBody.Env)
	}
	if createBody.User != "1000:1000" {
		t.Errorf("user = %q, want 1000:1000", createBody.User)
	}
}

func TestComposeRunServiceNotFound(t *testing.T) {
	svc := newComposeService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet || req.URL.Path != composeListPath {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.Path)
		}
		// Only a db container exists in the project; requesting web must fail
		// before any create request is issued.
		return response(http.StatusOK, `[
			{"Id":"db123","Image":"postgres:16","Names":["demo_db_1"],
			 "Labels":{"com.docker.compose.project":"demo","com.docker.compose.service":"db"},
			 "State":"exited","Created":"2024-01-01T00:00:00Z"}
		]`), nil
	}))

	err := svc.Run(context.Background(), "demo", "web", nil, runtimeapi.RunOptions{})
	if err == nil {
		t.Fatal("Run() error = nil, want service-not-found error")
	}
	if !strings.Contains(err.Error(), `service "web" not found in project "demo"`) {
		t.Fatalf("Run() error = %v, want service-not-found message", err)
	}
}

func TestComposeTopReturnsProcesses(t *testing.T) {
	svc := newComposeService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == composeListPath:
			return response(http.StatusOK, composeListBody), nil
		case req.Method == http.MethodGet && req.URL.Path == "/v5.2.0/libpod/containers/web123/top":
			return response(http.StatusOK, `{"Titles":["UID","PID","CMD"],"Processes":[["1000","1","nginx"]]}`), nil
		default:
			return response(http.StatusNotFound, `{}`), nil
		}
	}))

	got, err := svc.Top(context.Background(), "demo", "web")
	if err != nil {
		t.Fatalf("Top() error = %v", err)
	}
	if len(got.Titles) != 3 || got.Titles[0] != "UID" || got.Titles[2] != "CMD" {
		t.Errorf("Titles = %v", got.Titles)
	}
	if len(got.Processes) != 1 || len(got.Processes[0]) != 3 || got.Processes[0][2] != "nginx" {
		t.Errorf("Processes = %v", got.Processes)
	}
}

func TestComposePortReturnsHostBinding(t *testing.T) {
	svc := newComposeService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == composeListPath:
			return response(http.StatusOK, composeListBody), nil
		case req.Method == http.MethodGet && req.URL.Path == "/v5.2.0/libpod/containers/web123/json":
			return response(http.StatusOK, `{"NetworkSettings":{"Ports":{"8080/tcp":[{"HostIp":"127.0.0.1","HostPort":"18080"}]}}}`), nil
		default:
			return response(http.StatusNotFound, `{}`), nil
		}
	}))

	got, err := svc.Port(context.Background(), "demo", "web", 8080)
	if err != nil {
		t.Fatalf("Port() error = %v", err)
	}
	if got != "127.0.0.1:18080" {
		t.Errorf("Port() = %q, want 127.0.0.1:18080", got)
	}
}

func TestComposePortDefaultsHostIP(t *testing.T) {
	svc := newComposeService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == composeListPath:
			return response(http.StatusOK, composeListBody), nil
		case req.Method == http.MethodGet && req.URL.Path == "/v5.2.0/libpod/containers/web123/json":
			return response(http.StatusOK, `{"NetworkSettings":{"Ports":{"8080/tcp":[{"HostIp":"","HostPort":"18080"}]}}}`), nil
		default:
			return response(http.StatusNotFound, `{}`), nil
		}
	}))

	got, err := svc.Port(context.Background(), "demo", "web", 8080)
	if err != nil {
		t.Fatalf("Port() error = %v", err)
	}
	if got != "0.0.0.0:18080" {
		t.Errorf("Port() = %q, want 0.0.0.0:18080 when HostIp is empty", got)
	}
}

func TestComposePortNotExposed(t *testing.T) {
	svc := newComposeService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == composeListPath:
			return response(http.StatusOK, composeListBody), nil
		case req.Method == http.MethodGet && req.URL.Path == "/v5.2.0/libpod/containers/web123/json":
			return response(http.StatusOK, `{"NetworkSettings":{"Ports":{"443/tcp":[{"HostIp":"127.0.0.1","HostPort":"443"}]}}}`), nil
		default:
			return response(http.StatusNotFound, `{}`), nil
		}
	}))

	_, err := svc.Port(context.Background(), "demo", "web", 8080)
	if err == nil {
		t.Fatal("Port() error = nil, want not-exposed error")
	}
	if !strings.Contains(err.Error(), `port 8080 not exposed for service "web"`) {
		t.Fatalf("Port() error = %v, want not-exposed message", err)
	}
}

func TestComposeStatsMapsValues(t *testing.T) {
	svc := newComposeService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && req.URL.Path == composeListPath:
			return response(http.StatusOK, composeListBody), nil
		case req.Method == http.MethodGet && req.URL.Path == "/v5.2.0/libpod/containers/web123/stats":
			if got := req.URL.Query().Get("stream"); got != "false" {
				t.Errorf("stats stream = %q, want false", got)
			}
			return response(http.StatusOK, `{
				"cpu_stats":{"cpu_usage":{"total_usage":100},"system_cpu_usage":1000,"online_cpus":4},
				"precpu_stats":{"cpu_usage":{"total_usage":50},"system_cpu_usage":900,"online_cpus":4},
				"memory_stats":{"usage":1048576,"limit":2097152},
				"networks":{"eth0":{"rx_bytes":100,"tx_bytes":200}}
			}`), nil
		default:
			return response(http.StatusNotFound, `{}`), nil
		}
	}))

	got, err := svc.Stats(context.Background(), "demo", "web")
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if got.MemoryUsage != 1048576 || got.MemoryLimit != 2097152 {
		t.Errorf("memory = %f/%f, want 1048576/2097152", got.MemoryUsage, got.MemoryLimit)
	}
	if got.MemoryPercent != 50 {
		t.Errorf("MemoryPercent = %f, want 50", got.MemoryPercent)
	}
	if got.NetworkRx != 100 || got.NetworkTx != 200 {
		t.Errorf("network = rx:%f tx:%f, want rx:100 tx:200", got.NetworkRx, got.NetworkTx)
	}
}

func TestComposeEventsStreamsProjectEvents(t *testing.T) {
	eventBody := `{"Type":"container","Action":"start","ID":"web123","Name":"demo_web_1","TimeNano":1700000000000000000,"Attributes":{"com.docker.compose.project":"demo","com.docker.compose.service":"web"}}` + "\n"
	svc := newComposeService(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/v5.2.0/libpod/events" {
			t.Fatalf("unexpected path %q", req.URL.Path)
		}
		if got := req.URL.Query().Get("stream"); got != "true" {
			t.Errorf("events stream = %q, want true", got)
		}
		// /libpod/events filter spec is generic in podman swagger and
		// podman rejects "label" / "type" filter keys on this endpoint.
		// We subscribe to all container events and filter client-side.
		if got := req.URL.Query().Get("filters"); got != "" {
			t.Errorf("events filters = %q, want empty (server-side disabled)", got)
		}
		return response(http.StatusOK, eventBody), nil
	}))

	ctx := t.Context()
	out, err := svc.Events(ctx, "demo")
	if err != nil {
		t.Fatalf("Events() error = %v", err)
	}
	select {
	case ev, ok := <-out:
		if !ok {
			t.Fatal("events channel closed before first event")
		}
		if ev.Type != "container" || ev.Project != "demo" || ev.Service != "web" || ev.Action != "start" {
			t.Errorf("event = %#v, want container/demo/web/start", ev)
		}
		if want := time.Unix(0, 1700000000000000000); !ev.Timestamp.Equal(want) {
			t.Errorf("event timestamp = %v, want %v", ev.Timestamp, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for compose event")
	}
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected events channel to close after stream EOF")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for events channel close")
	}
}

func TestMaxContainerNumber(t *testing.T) {
	containers := []runtimeapi.ContainerSummary{
		{ComposeService: "web", ContainerNumber: "2"},
		{ComposeService: "db", ContainerNumber: "9"},
		{ComposeService: "web", ContainerNumber: "7"},
		{ComposeService: "web", ContainerNumber: "not-a-number"},
	}
	if got := maxContainerNumber(containers, "web"); got != 7 {
		t.Errorf("maxContainerNumber(web) = %d, want 7", got)
	}
	if got := maxContainerNumber(containers, "redis"); got != 0 {
		t.Errorf("maxContainerNumber(redis) = %d, want 0", got)
	}
}

func TestComposeServiceRequiresREST(t *testing.T) {
	svc := PodmanComposeService{}
	if err := svc.Run(context.Background(), "demo", "web", nil, runtimeapi.RunOptions{}); err != podmandriver.ErrPodmanRESTNotReady {
		t.Errorf("Run() error = %v, want ErrPodmanRESTNotReady", err)
	}
	if _, err := svc.Top(context.Background(), "demo", "web"); err != podmandriver.ErrPodmanRESTNotReady {
		t.Errorf("Top() error = %v, want ErrPodmanRESTNotReady", err)
	}
	if _, err := svc.Port(context.Background(), "demo", "web", 80); err != podmandriver.ErrPodmanRESTNotReady {
		t.Errorf("Port() error = %v, want ErrPodmanRESTNotReady", err)
	}
	if _, err := svc.Stats(context.Background(), "demo", "web"); err != podmandriver.ErrPodmanRESTNotReady {
		t.Errorf("Stats() error = %v, want ErrPodmanRESTNotReady", err)
	}
}
