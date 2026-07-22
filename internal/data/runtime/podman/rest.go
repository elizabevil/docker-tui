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
)

const defaultRequestTimeout = 10 * time.Second

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
	var version struct {
		APIVersion string `json:"ApiVersion"`
	}
	if err := c.do(ctx, "system.version", http.MethodGet, "/libpod/version", nil, &version); err != nil {
		return "", err
	}
	if version.APIVersion == "" {
		return "", runtimeapi.NewError(runtimeapi.ErrorInvalid, "system.version", "", fmt.Errorf("Podman response omitted ApiVersion"))
	}
	c.apiVersion = normalizeVersion(version.APIVersion)
	return c.apiVersion, nil
}

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

func (c *RESTClient) Delete(ctx context.Context, operation, path string) error {
	return c.DeleteWithQuery(ctx, operation, path, nil)
}

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
		var payload struct {
			Message string `json:"message"`
			Cause   string `json:"cause"`
		}
		_ = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload)
		message := strings.TrimSpace(payload.Message)
		if message == "" {
			message = strings.TrimSpace(payload.Cause)
		}
		if message == "" {
			message = response.Status
		}
		runtimeErr := &runtimeapi.Error{
			Kind:       runtimeapi.ClassifyHTTPStatus(response.StatusCode),
			Operation:  operation,
			Driver:     runtimeapi.Podman,
			Retryable:  runtimeapi.IsRetryableKind(runtimeapi.ClassifyHTTPStatus(response.StatusCode)),
			StatusCode: response.StatusCode,
			Err:        fmt.Errorf("%s", message),
		}
		return runtimeErr
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
