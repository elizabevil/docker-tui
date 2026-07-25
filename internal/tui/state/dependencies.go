package state

import (
	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
)

type Dependencies struct {
	Config     *config.Config
	Theme      *config.Theme
	AppVersion string
	Audit      *audit.Service
}
