package docker

import runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"

var _ runtimeapi.Engine = (*Client)(nil)
