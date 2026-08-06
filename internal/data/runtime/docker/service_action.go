package docker

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// resourceActionService implements runtime.ResourceActionService for the
// Docker/Podman adapter. It translates generic actions into SDK-specific calls.
type resourceActionService struct{ client *Client }

// Actions returns the resource action service facade.
func (c *Client) Actions() runtimeapi.ResourceActionService {
	return resourceActionService{client: c}
}

// Execute performs a lifecycle action on a resource. The action is routed to
// the appropriate SDK call based on the resource type and action kind.
func (s resourceActionService) Execute(ctx context.Context, ref runtimeapi.ResourceRef, action runtimeapi.Action, options runtimeapi.ActionOptions) (runtimeapi.ActionResult, error) {
	result := runtimeapi.ActionResult{Resource: ref, Action: action}
	err := s.execute(ctx, ref, action, options, &result)
	if err != nil {
		return runtimeapi.ActionResult{}, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(ref.Type, string(action)), ref, runtimeapi.Docker)
	}
	return result, nil
}

func (s resourceActionService) execute(ctx context.Context, ref runtimeapi.ResourceRef, action runtimeapi.Action, options runtimeapi.ActionOptions, result *runtimeapi.ActionResult) error {
	switch ref.Type {
	case runtimeapi.ResourceContainer:
		return s.executeContainer(ctx, ref.ID, action, options)
	case runtimeapi.ResourceImage:
		return s.executeImage(ctx, ref.ID, action, options, result)
	case runtimeapi.ResourceVolume:
		if action != runtimeapi.ActionRemove {
			return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceVolume, string(action)))
		}
		return s.client.cli.VolumeRemove(ctx, ref.ID, options.Lifecycle.Force)
	case runtimeapi.ResourceNetwork:
		if action != runtimeapi.ActionRemove {
			return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceNetwork, string(action)))
		}
		return s.client.cli.NetworkRemove(ctx, ref.ID)
	default:
		return runtimeapi.NewError(runtimeapi.ErrorInvalid, "resource.action", ref.ID,
			fmt.Errorf("unknown resource type %q", ref.Type))
	}
}

func (s resourceActionService) executeContainer(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions) error {
	lc := options.Lifecycle
	switch action {
	case runtimeapi.ActionStart:
		return s.client.cli.ContainerStart(ctx, id, container.StartOptions{})
	case runtimeapi.ActionStop:
		return s.client.cli.ContainerStop(ctx, id, container.StopOptions{Timeout: durationSeconds(lc.Timeout)})
	case runtimeapi.ActionRestart:
		return s.client.cli.ContainerRestart(ctx, id, container.StopOptions{Timeout: durationSeconds(lc.Timeout)})
	case runtimeapi.ActionKill:
		return s.client.cli.ContainerKill(ctx, id, lc.Signal)
	case runtimeapi.ActionPause:
		return s.client.cli.ContainerPause(ctx, id)
	case runtimeapi.ActionUnpause:
		return s.client.cli.ContainerUnpause(ctx, id)
	case runtimeapi.ActionRename:
		if lc.Name == "" {
			return invalidContainerAction(action, id, "name is required")
		}
		return s.client.cli.ContainerRename(ctx, id, lc.Name)
	case runtimeapi.ActionRemove:
		return s.client.cli.ContainerRemove(ctx, id, container.RemoveOptions{
			Force:         lc.Force,
			RemoveVolumes: lc.RemoveVolumes,
			RemoveLinks:   lc.RemoveLinks,
		})

	// TASK-019 advanced container actions. Each reads exactly one of
	// the typed sub-payloads from the ActionOptions union. The compiler
	// enforces that the call site filled the right field for the right
	// action.
	case runtimeapi.ActionUpdate, runtimeapi.ActionDiff, runtimeapi.ActionExport,
		runtimeapi.ActionCommit, runtimeapi.ActionWait, runtimeapi.ActionCopy:
		return s.dispatchAdvanced(ctx, id, action, options)

	default:
		return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceContainer, string(action)))
	}
}

// dispatchAdvanced routes the TASK-019 actions to their dedicated
// ContainerService implementations. It reads the typed sub-payload from
// the union and unwraps it into the dedicated ContainerService parameter
// type.
func (s resourceActionService) dispatchAdvanced(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions) error {
	svc := s.client.Containers()
	switch action {
	case runtimeapi.ActionUpdate:
		if options.Update == nil {
			return invalidContainerAction(action, id, "update options required")
		}
		_, err := svc.Update(ctx, id, runtimeapi.ContainerUpdateOptions{
			Memory:            options.Update.Memory,
			NanoCPUs:          options.Update.NanoCPUs,
			RestartPolicy:     options.Update.RestartPolicy,
			RestartMaxRetries: options.Update.RestartMaxRetries,
		})
		return err
	case runtimeapi.ActionDiff:
		_, err := svc.Diff(ctx, id)
		return err
	case runtimeapi.ActionExport:
		rc, err := svc.Export(ctx, id)
		if err != nil {
			return err
		}
		return rc.Close()
	case runtimeapi.ActionCommit:
		if options.Commit == nil {
			return invalidContainerAction(action, id, "commit options required")
		}
		_, err := svc.Commit(ctx, id, runtimeapi.ContainerCommitOptions{
			Repository: options.Commit.Repository,
			Tag:        options.Commit.Tag,
			Comment:    options.Commit.Comment,
			Author:     options.Commit.Author,
			Pause:      options.Commit.Pause,
		})
		return err
	case runtimeapi.ActionWait:
		var cond string
		if options.Wait != nil {
			cond = options.Wait.Condition
		}
		_, err := svc.Wait(ctx, id, cond)
		return err
	case runtimeapi.ActionCopy:
		if options.Copy == nil {
			return invalidContainerAction(action, id, "copy options required")
		}
		rc, err := svc.CopyFromContainer(ctx, id, options.Copy.SourcePath)
		if err != nil {
			return err
		}
		return rc.Close()
	}
	return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceContainer, string(action)))
}

// invalidContainerAction produces the standard "missing option" error used
// by advanced container actions that require typed sub-payloads.
func invalidContainerAction(action runtimeapi.Action, id, message string) error {
	return runtimeapi.NewError(runtimeapi.ErrorInvalid,
		runtimeapi.Operation(runtimeapi.ResourceContainer, string(action)), id, fmt.Errorf("%s", message))
}

func (s resourceActionService) executeImage(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions, result *runtimeapi.ActionResult) error {
	lc := options.Lifecycle
	switch action {
	case runtimeapi.ActionRemove:
		platforms := parsePlatforms(lc.Platforms)
		_, err := s.client.cli.ImageRemove(ctx, id, image.RemoveOptions{
			Force:         lc.Force,
			PruneChildren: lc.PruneChildren,
			Platforms:     platforms,
		})
		return err
	case runtimeapi.ActionPull:
		reader, err := s.client.cli.ImagePull(ctx, id, image.PullOptions{})
		if err != nil {
			return err
		}
		defer func() { _ = reader.Close() }() //nolint:errcheck // pull stream exhausted.
		_, err = io.Copy(io.Discard, reader)
		return err
	case runtimeapi.ActionPrune:
		report, err := s.client.cli.ImagesPrune(ctx, filters.NewArgs())
		result.SpaceReclaimed = report.SpaceReclaimed
		return err
	default:
		return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceImage, string(action)))
	}
}

func durationSeconds(timeout time.Duration) *int {
	if timeout <= 0 {
		return nil
	}
	seconds := int(timeout.Seconds())
	return &seconds
}

// parsePlatforms converts "os/arch" strings (e.g. "linux/amd64") into the
// ocispec.Platform slice expected by the Docker SDK. Empty / unparsable
// entries are skipped so a typo never produces a request that the daemon
// rejects for the wrong reason.
func parsePlatforms(values []string) []ocispec.Platform {
	if len(values) == 0 {
		return nil
	}
	out := make([]ocispec.Platform, 0, len(values))
	for _, raw := range values {
		os, arch, ok := strings.Cut(strings.TrimSpace(raw), "/")
		if !ok || os == "" || arch == "" {
			continue
		}
		out = append(out, ocispec.Platform{OS: os, Architecture: arch})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
