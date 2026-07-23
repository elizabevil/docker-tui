package podman

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// ExecuteImageAction performs a lifecycle action on an image via the
// Podman Libpod REST API. Action is a podman-native action name.
//
// For "pull", a raw response stream is returned for the caller to decode
// progress events. For "prune", reclaimed space is set on spaceReclaimed.
func (c *RESTClient) ExecuteImageAction(ctx context.Context, id string, action string, force bool, spaceReclaimed *uint64) (stream io.ReadCloser, err error) {
	switch action {
	case ActionRemove:
		query := url.Values{"force": {strconv.FormatBool(force)}}
		return nil, c.DeleteWithQuery(ctx, ImageRemovePath(id), query)
	case ActionPull:
		return c.StreamPost(ctx, PathImagePull, url.Values{"reference": {id}}, nil, "")
	case ActionPrune:
		var reports []dto.ImagePruneReportItem
		if err := c.Post(ctx, PathImagePrune, nil, nil, &reports); err != nil {
			return nil, err
		}
		for _, report := range reports {
			if spaceReclaimed != nil {
				*spaceReclaimed += report.Size
			}
			if report.Err != "" {
				return nil, fmt.Errorf("prune image %s: %s", report.ID, report.Err)
			}
		}
		return nil, nil
	default:
		return nil, newPodmanError(KindInvalid, "image."+action, fmt.Errorf("unsupported action: %s", action))
	}
}
