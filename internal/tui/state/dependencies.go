package state

import (
	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
)

type Dependencies struct {
	Config     *config.AppConfig
	Theme      *config.Theme
	AppVersion string
	Audit      *audit.Service
}
