package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type ListPodReport struct {
	Id         string                 `json:"Id"`
	Name       string                 `json:"Name"`
	Created    time.Time              `json:"Created"`
	Labels     map[string]string      `json:"Labels"`
	Status     string                 `json:"Status"`
	Cgroup     string                 `json:"Cgroup"`
	InfraId    string                 `json:"InfraId"`
	Namespace  string                 `json:"Namespace"`
	Networks   []string               `json:"Networks"`
	Containers []ListPodContainerItem `json:"Containers"`
}

type ListPodContainerItem struct {
	Names     string            `json:"Names"`
	Id        string            `json:"Id"`
	Status    string            `json:"Status"`
	Number    uint              `json:"Number"`
	StartedAt time.Time         `json:"StartedAt"`
	Labels    map[string]string `json:"Labels"`
}

type PodInspectReport struct {
	Id         string                 `json:"Id"`
	Name       string                 `json:"Name"`
	Created    time.Time              `json:"Created"`
	Labels     map[string]string      `json:"Labels"`
	Status     string                 `json:"Status"`
	Cgroup     string                 `json:"Cgroup"`
	InfraId    string                 `json:"InfraId"`
	Namespace  string                 `json:"Namespace"`
	Networks   []string               `json:"Networks"`
	Containers []ListPodContainerItem `json:"Containers"`
}

func (c *RESTClient) ListPods(ctx context.Context, filters map[string][]string) ([]ListPodReport, error) {
	query := make(url.Values)
	if len(filters) > 0 {
		encoded, err := json.Marshal(filters)
		if err != nil {
			return nil, fmt.Errorf("encode Podman pod filters: %w", err)
		}
		query.Set("filters", string(encoded))
	}
	var raw []ListPodReport
	if err := c.Get(ctx, PathPodList, query, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (c *RESTClient) InspectPod(ctx context.Context, name string) (*PodInspectReport, error) {
	var raw PodInspectReport
	if err := c.Get(ctx, PodInspectPath(name), nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}
