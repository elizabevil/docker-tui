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

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
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
// path like /v4.0.0/libpod/version. We try the unversioned endpoint first
// and fall back to the minimum API version on any failure — not just 404 —
// because some Podman 5.x builds reject the unversioned path with 400 or
// 405 instead of a clean NotFound.
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
	var lastErr error
	candidates := []string{
		PathSystemVersion,
		"/v" + minPodmanAPIVersion + PathSystemVersion,
	}
	for _, path := range candidates {
		u := *c.baseURL
		u.Path = path
		err := c.do(ctx, http.MethodGet, &u, nil, &version)
		if err == nil {
			libpodVersion := extractLibpodVersion(version.Components)
			if libpodVersion == "" {
				libpodVersion = version.APIVersion
			}
			if libpodVersion != "" {
				c.apiVersion = normalizeVersion(libpodVersion)
				return c.apiVersion, nil
			}
			// Response lacked a usable version — try the next candidate.
			lastErr = newPodmanError(KindInvalid, "system.version",
				fmt.Errorf("Podman response omitted ApiVersion"))
			continue
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = newPodmanError(KindInvalid, "system.version",
			fmt.Errorf("no Podman version endpoint responded"))
	}
	return "", lastErr
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

// url builds a full request URL from baseURL + version + path + query.
// path may contain pre-escaped segments (e.g. team%2Fcache); the
// RawPath field preserves them on the wire.
func (c *RESTClient) url(path string, query url.Values) *url.URL {
	version := c.apiVersion
	if version == "" {
		version = c.configured
	}
	var fullPath string
	if version != "" {
		fullPath = "/v" + version + path
	} else {
		fullPath = path
	}
	u := *c.baseURL
	parsed, _ := url.Parse(fullPath) //nolint:errcheck // path comes from a configured versioned prefix; always parses.
	u.Path = parsed.Path
	if parsed.RawPath != "" {
		u.RawPath = parsed.RawPath
	} else if strings.Contains(fullPath, "%") {
		u.RawPath = fullPath
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	return &u
}

// Get performs a versioned GET request and decodes the response into output.
func (c *RESTClient) Get(ctx context.Context, path string, query url.Values, output any) error {
	if err := c.ensureVersion(ctx); err != nil {
		return err
	}
	return c.do(ctx, http.MethodGet, c.url(path, query), nil, output)
}

// StreamGet opens a versioned Libpod GET stream. The returned body belongs to
// the caller and remains active until it is closed or ctx is cancelled.
func (c *RESTClient) StreamGet(ctx context.Context, path string, query url.Values) (io.ReadCloser, error) {
	if err := c.ensureVersion(ctx); err != nil {
		return nil, err
	}
	return c.stream(ctx, http.MethodGet, c.url(path, query), nil, "")
}

// StreamPost opens a versioned Libpod POST stream with a caller-owned body.
func (c *RESTClient) StreamPost(ctx context.Context, path string, query url.Values, body io.Reader, contentType string) (io.ReadCloser, error) {
	if err := c.ensureVersion(ctx); err != nil {
		return nil, err
	}
	return c.stream(ctx, http.MethodPost, c.url(path, query), body, contentType)
}

// UpgradePost opens a versioned bidirectional stream using HTTP Upgrade.
// Exec and attach sessions use the returned stream for concurrent reads
// and writes; the caller owns and must close it.
func (c *RESTClient) UpgradePost(ctx context.Context, path string, body io.Reader, contentType string) (io.ReadWriteCloser, error) {
	if err := c.ensureVersion(ctx); err != nil {
		return nil, err
	}
	u := c.url(path, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), body)
	if err != nil {
		return nil, newPodmanError(KindInvalid, "", err)
	}
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tcp")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	client := *c.client
	client.Timeout = 0
	response, err := client.Do(req)
	if err != nil {
		kind := classifyContextError(err)
		if kind == KindInternal {
			kind = KindConnection
		}
		return nil, newPodmanError(kind, "", err)
	}
	if response.StatusCode != http.StatusSwitchingProtocols {
		defer func() { _ = response.Body.Close() }() //nolint:errcheck // error body drained by decodeResponseError.
		return nil, c.decodeResponseError(response)
	}
	stream, ok := response.Body.(io.ReadWriteCloser)
	if !ok {
		_ = response.Body.Close() //nolint:errcheck // upgrade fallback; body discarded.
		return nil, newPodmanError(KindInternal, "", fmt.Errorf("upgrade response is not bidirectional"))
	}
	return stream, nil
}

func (c *RESTClient) stream(ctx context.Context, method string, u *url.URL, body io.Reader, contentType string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, newPodmanError(KindInvalid, "", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	client := *c.client
	client.Timeout = 0
	response, err := client.Do(req)
	if err != nil {
		kind := classifyContextError(err)
		if kind == KindInternal {
			kind = KindConnection
		}
		return nil, newPodmanError(kind, "", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		defer func() { _ = response.Body.Close() }() //nolint:errcheck // error body drained by decodeResponseError.
		return nil, c.decodeResponseError(response)
	}
	return response.Body, nil
}

func (c *RESTClient) decodeResponseError(response *http.Response) error {
	var payload dto.EngineErrorPayload
	_ = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload) //nolint:errcheck // best-effort error decode; fall back to status text.
	message := strings.TrimSpace(payload.Message)
	if message == "" {
		message = strings.TrimSpace(payload.Cause)
	}
	if message == "" {
		message = response.Status
	}
	kind := classifyHTTPStatus(response.StatusCode)
	return &Error{Kind: kind, Operation: "", StatusCode: response.StatusCode, Err: fmt.Errorf("%s", message)}
}

// Delete performs a versioned Libpod DELETE request.
func (c *RESTClient) Delete(ctx context.Context, path string) error {
	return c.DeleteWithQuery(ctx, path, nil)
}

// Post performs a versioned POST request with an optional JSON body.
func (c *RESTClient) Post(ctx context.Context, path string, query url.Values, input, output any) error {
	if err := c.ensureVersion(ctx); err != nil {
		return err
	}
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return newPodmanError(KindInvalid, "", err)
		}
		body = bytes.NewReader(encoded)
	}
	return c.do(ctx, http.MethodPost, c.url(path, query), body, output)
}

// DeleteWithQuery performs a versioned DELETE while preserving operation-specific
// options such as force.
func (c *RESTClient) DeleteWithQuery(ctx context.Context, path string, query url.Values) error {
	if err := c.ensureVersion(ctx); err != nil {
		return err
	}
	return c.do(ctx, http.MethodDelete, c.url(path, query), nil, nil)
}

// ensureVersion negotiates the API version if not already known.
func (c *RESTClient) ensureVersion(ctx context.Context) error {
	_, err := c.APIVersion(ctx)
	return err
}

// do executes an HTTP request with the given URL. Callers build the full URL
// via c.url() before calling do.
func (c *RESTClient) do(ctx context.Context, method string, u *url.URL, body io.Reader, output any) error {
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return newPodmanError(KindInvalid, "", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := c.client.Do(req)
	if err != nil {
		kind := classifyContextError(err)
		if kind == KindInternal {
			kind = KindConnection
		}
		return newPodmanError(kind, "", err)
	}
	defer func() { _ = response.Body.Close() }() //nolint:errcheck // body fully consumed by Decode or error path.
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return c.decodeResponseError(response)
	}
	if output == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(output); err != nil {
		return newPodmanError(KindInvalid, "", err)
	}
	return nil
}

func normalizeVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}
