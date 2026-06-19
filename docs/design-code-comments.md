# dtui 代码注释规范

## 风格

Go 标准 Godoc 风格:

```go
// Package docker provides a wrapper around the Docker SDK client.
//
// It handles connection detection (Docker/Podman), socket auto-start,
// and simplified APIs for container, image, volume, and network operations.
package docker

// Client wraps the Docker SDK client with engine type detection.
//
// Use NewClient to create an instance. The client automatically detects
// whether the target is Docker or Podman. Close() must be called to
// release the underlying SDK client and cancel the background context.
type Client struct {
    cli        *client.Client   // Underlying Docker SDK client
    ctx        context.Context  // Background context for API calls
    cancel     context.CancelFunc
    EngineType EngineType       // Detected engine (docker/podman)
    Host       string           // Connected socket/host URI
}

// ListContainers returns containers matching the given options.
//
// Network IPs are extracted from NetworkSettings.Networks.
// Mount counts are taken from the Mounts array.
func (c *Client) ListContainers(opts ContainerListOptions) ([]ContainerSummary, error) {
```

## 规则

| 规则        | 强制 | 说明                          |
|-----------|----|-----------------------------|
| 导出符号必须注释  | ✅  | 类型、函数、常量                    |
| 注释以符号名开头  | ✅  | `// ListContainers returns` |
| 参数说明      | ✅  | opts, args 等                |
| 不注释显而易见的事 | ✅  | getter 不注释                  |

## 覆盖范围

### Phase 1: `internal/docker/` (数据层)

```go
// types.go
// ContainerListOptions holds parameters for listing containers.
// ContainerSummary is a simplified container representation.
// ImageSummary is a simplified image representation.
// VolumeItem is a simplified volume representation.
// NetworkItem is a simplified network representation.

// containers.go
// ListContainers returns containers matching the given options.
// ContainerStart starts a container by ID.
// ContainerStop stops a running container.
// ContainerRestart restarts a container.
// ContainerKill kills a running container.
// ContainerRemove removes a container.
// ContainerLogs streams logs from a container.
// ContainerStats subscribes to stats stream for a container.

// images.go
// ListImages returns all available images.
// RemoveImage removes an image by ID.
// PullImage pulls an image from a registry.
// PruneImages removes unused images.
// InspectImage returns detailed image information.
// SplitImageRef splits a RepoTag into registry, name, tag.

// host.go
// ReadHostStats returns CPU%, MEM%, and disk usage for the host machine.

// events.go
// EventsMsg represents a Docker event or error.
// Events returns a channel that streams Docker events.

// stats.go
// ParseStats decodes a Docker stats JSON stream and returns CPU/memory metrics.

// volumes.go
// ListVolumes returns all volumes.
// RemoveVolume removes a volume by name.

// networks.go
// ListNetworks returns all networks.
// RemoveNetwork removes a network by ID.

// inspect.go
// InspectContainer returns formatted container details.
```

### Phase 2: `internal/tui/` (逻辑层)

```go
// styles.go
// ApplyTheme initializes all styles from a theme configuration.
// ContainerStateColor returns a color based on container state.
// 以及 22 个样式变量的简短注释

// update.go
// Update handles all messages and returns the updated model and command.

// cmd_fetch.go
// FetchContainers creates a command to load the container list.
// FetchImages creates a command to load the image list.
// FetchVolumes creates a command to load the volume list.
// FetchNetworks creates a command to load the network list.
// FetchAll creates commands to fetch all resource types.

// keys.go
// KeyAction represents a named action bound to keyboard keys.
// KeyMapping maps key sequences to actions.
// DefaultKeyMapping returns the default key bindings.
// HelpEntries returns all available key bindings.
// KeyFromMsg extracts a key string from a key message.

// handler_container.go
// containerStartCmd creates a command to start a container.
// containerStopCmd creates a command to stop a container.
// containerRestartCmd creates a command to restart a container.
// containerKillCmd creates a command to kill a container.
// containerRemoveCmd creates a command to remove a container.
// doContainerAction handles container actions.
// doContainerRemove handles container removal.
// doLogAction opens the log view for a container.
// doStatsAction toggles stats polling.
// doExecAction shows exec hint.
// doInspectAction shows container inspect details.
// doDeleteAction handles delete across all panels.
// doEnterAction handles the Enter key.
```

### Phase 3: `internal/tui/view/` (视图层)

```go
// layout.go
// RenderApp renders the full application view.

// containers.go
// renderContainerList renders the container table.

// helpers.go
// shortID returns a shortened container/image ID.
// truncate truncates a string with ellipsis.
// formatSize formats bytes as a human-readable string.
```

## 总计

| Phase       | 文件数    | 注释数     |
|-------------|--------|---------|
| 1 — docker/ | 9      | ~35     |
| 2 — tui/    | 8      | ~30     |
| 3 — view/   | 4      | ~10     |
| **总计**      | **21** | **~75** |
