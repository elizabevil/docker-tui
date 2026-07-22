package podman

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
)

// ExecuteImageActionREST performs a lifecycle action on an image via the
// Podman Libpod REST API.
func ExecuteImageActionREST(ctx context.Context, client *Client, id string, action runtimeapi.Action, options runtimeapi.ActionOptions, result *runtimeapi.ActionResult) error {
	if client.REST == nil {
		return runtimeapi.NewError(runtimeapi.ErrorUnavailable, "image."+string(action), id, errPodmanRESTNotReady)
	}
	switch action {
	case runtimeapi.ActionRemove:
		query := url.Values{"force": {strconv.FormatBool(options.Force)}}
		return client.REST.DeleteWithQuery(ctx, "image.remove", fmt.Sprintf(PathImageRemove, url.PathEscape(id)), query)
	case runtimeapi.ActionPull:
		reader, err := client.REST.StreamPost(ctx, "image.pull", PathImagePull, url.Values{"reference": {id}}, nil, "")
		if err != nil {
			return err
		}
		defer reader.Close()
		return DecodeImageProgress(ctx, reader, nil)
	case runtimeapi.ActionPrune:
		var reports []dto.ImagePruneReportItem
		if err := client.REST.Post(ctx, "image.prune", PathImagePrune, nil, nil, &reports); err != nil {
			return err
		}
		for _, report := range reports {
			result.SpaceReclaimed += report.Size
			if report.Err != "" {
				return fmt.Errorf("prune image %s: %s", report.ID, report.Err)
			}
		}
		return nil
	default:
		return runtimeapi.UnsupportedError("image." + string(action))
	}
}
