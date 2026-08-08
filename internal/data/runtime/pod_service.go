package runtime

import (
	"context"
	"time"
)

// Well-known filter field names for pod list operations.
const (
	PodFilterName  = "name"
	PodFilterLabel = "label"
)

// PodListOptions holds the filter set for pod list requests.
type PodListOptions struct{ Filters FilterSet }

// NativeFilters converts the filter set to the native Docker/Podman
// representation. Adapters post-filter all values using AND semantics.
func (o PodListOptions) NativeFilters() (map[string][]string, error) {
	return nativeSingleValueFilters("pod.list.filter", o.Filters, map[string]struct{}{
		PodFilterName: {}, PodFilterLabel: {},
	})
}

// PodService 是 Pod 一等公民视图的统一接口(per R08-15 Q6 + R08-14)。
//
// docker 端:daemon 无 /pods 端点,driver 层自实现 wrap(读 inspect.NetworkMode/PidMode/IpcMode
// 推 Group,per R08-14 §F2 算法)。Source = docker_network_mode_service 等。
//
// podman 端:daemon 提供 /libpod/pods 一等公民,driver 层直包装;podman-compose 默认 1 项目 1 pod,
// Source = podman_default_pod。
//
// 两端各自实现,但接口形态统一 — runtime 层永远拿同一形状 PodSummary / PodDetail。
type PodService interface {
	ListPods(ctx context.Context, opts PodListOptions) ([]PodSummary, error)
	InspectPod(ctx context.Context, name string) (PodDetail, error)
}

type PodSummary struct {
	Name     string
	Source   string
	Running  int
	Stopped  int
	Created  time.Time
}

type PodDetail struct {
	Name       string
	Source     string
	Containers []PodContainer
	Labels     map[string]string
}

type PodContainer struct {
	ContainerID string
	Service     string
	Number      string
}
