package dto

// Version is the Podman Libpod /libpod/version response.
type Version struct {
	APIVersion string             `json:"ApiVersion"`
	Components []ComponentVersion `json:"Components"`
}

// ComponentVersion is a single component entry inside Version.Components.
type ComponentVersion struct {
	Name    string           `json:"Name"`
	Details ComponentDetails `json:"Details"`
}

// ComponentDetails is the nested details payload of ComponentVersion.
type ComponentDetails struct {
	APIVersion string `json:"APIVersion"`
}

// EngineErrorPayload is the Libpod error JSON body returned by the engine.
type EngineErrorPayload struct {
	Message string `json:"message"`
	Cause   string `json:"cause"`
}
