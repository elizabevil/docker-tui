package detail

import (
	"fmt"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func imageFullRef(detail *dockerclient.ImageDetail) string {
	registry := detail.Registry
	if registry == "" {
		registry = "docker.io"
	}
	name, tag := detail.Name, detail.Tag
	if tag == "" {
		tag = "latest"
	}
	return fmt.Sprintf("%s/%s:%s", registry, name, tag)
}