package docker

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	runtimepodman "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

func (c *Client) executeImagePodmanREST(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions, result *runtimeapi.ActionResult) error {
	if c.podmanREST == nil {
		return runtimeapi.NewError(runtimeapi.ErrorUnavailable, "image."+string(action), id, errPodmanRESTNotReady)
	}
	switch action {
	case runtimeapi.ActionRemove:
		query := url.Values{"force": {strconv.FormatBool(options.Force)}}
		return c.podmanREST.DeleteWithQuery(ctx, "image.remove", runtimepodman.ImagePath(id, ""), query)
	case runtimeapi.ActionPull:
		reader, err := c.podmanREST.StreamPost(ctx, "image.pull", "/images/pull", url.Values{"reference": {id}}, nil, "")
		if err != nil {
			return err
		}
		defer reader.Close()
		return decodeImageProgress(ctx, reader, nil)
	case runtimeapi.ActionPrune:
		var reports []struct {
			ID   string `json:"Id"`
			Size uint64 `json:"Size"`
			Err  string `json:"Err"`
		}
		if err := c.podmanREST.Post(ctx, "image.prune", "/images/prune", nil, nil, &reports); err != nil {
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
