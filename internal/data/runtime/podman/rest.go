package podman

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
)

const (
	defaultRequestTimeout = 10 * time.Second
	// minPodmanAPIVersion is the minimum Libpod API version supported by
	// Podman ≥ 5.x. Used as a fallback when the unversioned /libpod/version
	// endpoint returns 404.
	minPodmanAPIVersion = "4.0.0"
)

// RESTConfig configures a Podman Libpod REST client.
type RESTConfig struct {
	Endpoint   string
	APIVersion string
	TLS        bool
	HTTPClient *http.Client
}

// RESTClient is the shared non-CGO transport for the Libpod API.
type RESTClient struct {
	client     *http.Client
	baseURL    *url.URL
	configured string

	mu         sync.Mutex
	apiVersion string
}

// NewRESTClient creates an HTTP transport for the Podman Libpod API. It
// supports unix, tcp, http, and https endpoint schemes and configures TLS
// where applicable.
func NewRESTClient(config RESTConfig) (*RESTClient, error) {
	endpoint, err := url.Parse(config.Endpoint)
	if err != nil || endpoint.Scheme == "" {
		return nil, fmt.Errorf("invalid Podman endpoint %q", config.Endpoint)
	}

	client := config.HTTPClient
	baseURL := &url.URL{}
	switch endpoint.Scheme {
	case "unix":
		if config.TLS {
			return nil, fmt.Errorf("TLS cannot be used with a unix socket")
		}
		socketPath := endpoint.Path
		transport := &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var dialer net.Dialer
				return dialer.DialContext(ctx, "unix", socketPath)
			},
		}
		client = &http.Client{Transport: transport, Timeout: defaultRequestTimeout}
		baseURL = &url.URL{Scheme: "http", Host: "localhost"}
	case "tcp":
		baseURL = &url.URL{Scheme: "http", Host: endpoint.Host}
		if config.TLS {
			baseURL.Scheme = "https"
		}
	case "http", "https":
		baseURL = &url.URL{Scheme: endpoint.Scheme, Host: endpoint.Host}
	default:
		return nil, fmt.Errorf("unsupported Podman endpoint scheme %q", endpoint.Scheme)
	}
	if client == nil {
		client = &http.Client{Timeout: defaultRequestTimeout}
	} else if client.Timeout == 0 {
		clientCopy := *client
		clientCopy.Timeout = defaultRequestTimeout
		client = &clientCopy
	}
	return &RESTClient{client: client, baseURL: baseURL, configured: normalizeVersion(config.APIVersion)}, nil
}

// APIVersion returns the negotiated Libpod API version, querying the server
// if needed. The result is cached after the first successful call.
//
// Podman ≤ 4.x serves /libpod/version; Podman ≥ 5.x requires a versioned
// path like /v4.0.0/libpod/version. We try unversioned first and fall back
// to the minimum API version on 404.
//
// The top-level ApiVersion field returns the Docker-compatible version
// (e.g. "1.41"), but Libpod endpoints need the Podman engine version
// from Components[0].Details.APIVersion (e.g. "5.4.2").
func (c *RESTClient) APIVersion(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.apiVersion != "" {
		return c.apiVersion, nil
	}
	if c.configured != "" {
		c.apiVersion = c.configured
		return c.apiVersion, nil
	}
	var version dto.Version
	err := c.do(ctx, "system.version", http.MethodGet, "/libpod/version", nil, &version)
	if err != nil {
		if !runtimeapi.IsErrorKind(err, runtimeapi.ErrorNotFound) {
			return "", err
		}
		// Podman ≥ 5.x: use minimum supported version as the fallback.
		err = c.do(ctx, "system.version", http.MethodGet, "/v"+minPodmanAPIVersion+"/libpod/version", nil, &version)
		if err != nil {
			return "", err
		}
	}
	// Prefer the Podman engine version from Components for Libpod endpoints.
	libpodVersion := extractLibpodVersion(version.Components)
	if libpodVersion == "" {
		libpodVersion = version.APIVersion
	}
	if libpodVersion == "" {
		return "", runtimeapi.NewError(runtimeapi.ErrorInvalid, "system.version", "", fmt.Errorf("Podman response omitted ApiVersion"))
	}
	c.apiVersion = normalizeVersion(libpodVersion)
	return c.apiVersion, nil
}

// extractLibpodVersion finds the Podman Engine component version.
func extractLibpodVersion(components []dto.ComponentVersion) string {
	for _, c := range components {
		if c.Name == "Podman Engine" && c.Details.APIVersion != "" {
			return c.Details.APIVersion
		}
	}
	return ""
}

// Get performs a versioned Libpod GET request and decodes the response into
// output. The operation string is used for error classification.
func (c *RESTClient) Get(ctx context.Context, operation, path string, query url.Values, output any) error {
	version, err := c.APIVersion(ctx)
	if err != nil {
		return err
	}
	versionedPath := "/v" + version + "/libpod/" + strings.TrimPrefix(path, "/")
	if len(query) > 0 {
		versionedPath += "?" + query.Encode()
	}
	return c.do(ctx, operation, http.MethodGet, versionedPath, nil, output)
}

// StreamGet opens a versioned Libpod GET stream. The returned body belongs to
// the caller and remains active until it is closed or ctx is cancelled.
func (c *RESTClient) StreamGet(ctx context.Context, operation, path string, query url.Values) (io.ReadCloser, error) {
	return c.stream(ctx, operation, http.MethodGet, path, query, nil, "")
}

// StreamPost opens a versioned Libpod POST stream with a caller-owned body.
func (c *RESTClient) StreamPost(ctx context.Context, operation, path string, query url.Values, body io.Reader, contentType string) (io.ReadCloser, error) {
	return c.stream(ctx, operation, http.MethodPost, path, query, body, contentType)
}

// UpgradePost opens a versioned Libpod bidirectional stream using HTTP
// Upgrade. Exec and attach sessions use the returned stream for concurrent
// reads and writes; the caller owns and must close it.
func (c *RESTClient) UpgradePost(ctx context.Context, operation, path string, body io.Reader, contentType string) (io.ReadWriteCloser, error) {
	version, err := c.APIVersion(ctx)
	if err != nil {
		return nil, err
	}
	requestURL := *c.baseURL
	requestURL.Path = "/v" + version + "/libpod/" + strings.TrimPrefix(path, "/")
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), body)
	if err != nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorInvalid, operation, "", err)
	}
	request.Header.Set("Connection", "Upgrade")
	request.Header.Set("Upgrade", "tcp")
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	client := *c.client
	client.Timeout = 0
	response, err := client.Do(request)
	if err != nil {
		kind := runtimeapi.ClassifyContextError(err)
		if kind == runtimeapi.ErrorInternal {
			kind = runtimeapi.ErrorConnection
		}
		runtimeErr := runtimeapi.NewError(kind, operation, "", err).(*runtimeapi.Error)
		runtimeErr.Driver = runtimeapi.Podman
		return nil, runtimeErr
	}
	if response.StatusCode != http.StatusSwitchingProtocols {
		defer response.Body.Close()
		return nil, c.decodeResponseError(operation, response)
	}
	stream, ok := response.Body.(io.ReadWriteCloser)
	if !ok {
		response.Body.Close()
		return nil, runtimeapi.NewError(runtimeapi.ErrorInternal, operation, "", fmt.Errorf("upgrade response is not bidirectional"))
	}
	return stream, nil
}

func (c *RESTClient) stream(ctx context.Context, operation, method, path string, query url.Values, body io.Reader, contentType string) (io.ReadCloser, error) {
	version, err := c.APIVersion(ctx)
	if err != nil {
		return nil, err
	}
	versionedPath := "/v" + version + "/libpod/" + strings.TrimPrefix(path, "/")
	if len(query) > 0 {
		versionedPath += "?" + query.Encode()
	}
	requestURL := *c.baseURL
	parsed, err := url.Parse(versionedPath)
	if err != nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorInvalid, operation, "", err)
	}
	requestURL.Path, requestURL.RawPath, requestURL.RawQuery = parsed.Path, parsed.RawPath, parsed.RawQuery
	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), body)
	if err != nil {
		return nil, runtimeapi.NewError(runtimeapi.ErrorInvalid, operation, "", err)
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	client := *c.client
	client.Timeout = 0
	response, err := client.Do(request)
	if err != nil {
		kind := runtimeapi.ClassifyContextError(err)
		if kind == runtimeapi.ErrorInternal {
			kind = runtimeapi.ErrorConnection
		}
		runtimeErr := runtimeapi.NewError(kind, operation, "", err).(*runtimeapi.Error)
		runtimeErr.Driver = runtimeapi.Podman
		return nil, runtimeErr
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		defer response.Body.Close()
		return nil, c.decodeResponseError(operation, response)
	}
	return response.Body, nil
}

func (c *RESTClient) decodeResponseError(operation string, response *http.Response) error {
	var payload dto.EngineErrorPayload
	_ = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload)
	message := strings.TrimSpace(payload.Message)
	if message == "" {
		message = strings.TrimSpace(payload.Cause)
	}
	if message == "" {
		message = response.Status
	}
	kind := runtimeapi.ClassifyHTTPStatus(response.StatusCode)
	return &runtimeapi.Error{Kind: kind, Operation: operation, Driver: runtimeapi.Podman, Retryable: runtimeapi.IsRetryableKind(kind), StatusCode: response.StatusCode, Err: fmt.Errorf("%s", message)}
}

// Delete performs a versioned Libpod DELETE request.
func (c *RESTClient) Delete(ctx context.Context, operation, path string) error {
	return c.DeleteWithQuery(ctx, operation, path, nil)
}

// Post performs a versioned Libpod POST request with an optional JSON body.
func (c *RESTClient) Post(ctx context.Context, operation, path string, query url.Values, input, output any) error {
	version, err := c.APIVersion(ctx)
	if err != nil {
		return err
	}
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return runtimeapi.NewError(runtimeapi.ErrorInvalid, operation, "", err)
		}
		body = bytes.NewReader(encoded)
	}
	versionedPath := "/v" + version + "/libpod/" + strings.TrimPrefix(path, "/")
	if len(query) > 0 {
		versionedPath += "?" + query.Encode()
	}
	return c.do(ctx, operation, http.MethodPost, versionedPath, body, output)
}

// DeleteWithQuery performs a versioned Libpod DELETE while preserving
// operation-specific options such as force. Keeping query encoding here
// prevents resource adapters from constructing versioned URLs themselves.
func (c *RESTClient) DeleteWithQuery(ctx context.Context, operation, path string, query url.Values) error {
	version, err := c.APIVersion(ctx)
	if err != nil {
		return err
	}
	versionedPath := "/v" + version + "/libpod/" + strings.TrimPrefix(path, "/")
	if len(query) > 0 {
		versionedPath += "?" + query.Encode()
	}
	return c.do(ctx, operation, http.MethodDelete, versionedPath, nil, nil)
}

func (c *RESTClient) do(ctx context.Context, operation, method, path string, body io.Reader, output any) error {
	requestURL := *c.baseURL
	requestURL.Path = path
	if parsed, err := url.Parse(path); err == nil {
		requestURL.Path = parsed.Path
		// Path stores the decoded form. RawPath is required so escaped resource
		// names such as "team%2Fcache" remain one URL segment on the wire.
		requestURL.RawPath = parsed.RawPath
		requestURL.RawQuery = parsed.RawQuery
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), body)
	if err != nil {
		return runtimeapi.NewError(runtimeapi.ErrorInvalid, operation, "", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := c.client.Do(req)
	if err != nil {
		kind := runtimeapi.ClassifyContextError(err)
		if kind == runtimeapi.ErrorInternal {
			kind = runtimeapi.ErrorConnection
		}
		runtimeErr := runtimeapi.NewError(kind, operation, "", err).(*runtimeapi.Error)
		runtimeErr.Driver = runtimeapi.Podman
		return runtimeErr
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return c.decodeResponseError(operation, response)
	}
	if output == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(output); err != nil {
		return runtimeapi.NewError(runtimeapi.ErrorInvalid, operation, "", err)
	}
	return nil
}

func normalizeVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}
