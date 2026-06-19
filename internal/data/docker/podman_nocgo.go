//go:build !cgo

package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

func (c *Client) listImagesPodman() ([]ImageSummary, error) {
	sockPath := c.hostSocketPath()
	if sockPath == "" {
		return nil, fmt.Errorf("cannot determine socket from %s", c.Host)
	}
	dial := func(_, _ string) (net.Conn, error) {
		return net.DialTimeout("unix", sockPath, 5*time.Second)
	}
	hc := &http.Client{
		Transport: &http.Transport{Dial: dial},
		Timeout:   10 * time.Second,
	}
	u := &url.URL{Scheme: "http", Host: "localhost", Path: "/v5.0.0/libpod/images/json"}
	resp, err := hc.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("podman list: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	var raw []struct {
		ID             string            `json:"Id"`
		RepoTags       []string          `json:"RepoTags"`
		Created        int64             `json:"Created"`
		Size           int64             `json:"Size"`
		Labels         map[string]string `json:"Labels"`
		Architecture   string            `json:"Architecture,omitempty"`
		IsManifestList *bool             `json:"IsManifestList,omitempty"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	out := make([]ImageSummary, 0, len(raw))
	for _, p := range raw {
		s := ImageSummary{
			ID: p.ID, RepoTags: p.RepoTags, Created: p.Created,
			Size: p.Size, Labels: p.Labels,
		}
		if p.Architecture != "" {
			s.Arch = p.Architecture
		} else {
			s.Arch = "\u2014"
		}
		if p.IsManifestList != nil && *p.IsManifestList {
			s.IsManifest = true
		}
		s.Registry, _, _ = splitImageRef(s.RepoTags)
		out = append(out, s)
	}
	return out, nil
}

func (c *Client) hostSocketPath() string {
	if len(c.Host) > 7 && c.Host[:7] == "unix://" {
		return c.Host[7:]
	}
	return ""
}
