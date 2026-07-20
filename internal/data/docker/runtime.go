package docker

// ImageLister provides image listing with full metadata.
// Different runtimes (Docker SDK vs Podman SDK) implement this interface
// to provide runtime-specific metadata (architecture, manifest info, etc.).
type ImageLister interface {
	ListImages() ([]ImageSummary, error)
}

// dockerImageLister uses the Docker SDK to list images.
// Compatible with both Docker and Podman's Docker API endpoint,
// but may lack runtime-specific metadata (Arch, IsManifestList).
type dockerImageLister struct {
	client *Client
}

func (l *dockerImageLister) ListImages() ([]ImageSummary, error) {
	return l.client.listImagesDocker()
}

// podmanImageLister uses the Podman Go SDK's native bindings.
// Provides full metadata including Architecture and IsManifestList.
type podmanImageLister struct {
	client *Client
}

func (l *podmanImageLister) ListImages() ([]ImageSummary, error) {
	return l.client.listImagesPodman()
}
