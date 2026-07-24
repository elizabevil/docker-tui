// Package buildinfo centralizes the runtime/build metadata for the dtui
// binary. It is consumed by `dtui version` and by the runtime state model
// for display in the TUI header.
package buildinfo

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// Info is the JSON-serializable record exposed by `dtui version`. Fields are
// tagged with omitempty for fields that are not always populated (commit /
// build time are only set when the binary is built with VCS metadata, e.g.
// via `go build` from a git checkout).
type Info struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Commit    string `json:"commit,omitempty"`
	BuildTime string `json:"buildTime,omitempty"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// Read returns the build info for the current binary. The name and version
// are injected by the caller (typically the main package's hard-coded
// version constant); the rest is read from the Go runtime and the embedded
// build info populated by the go toolchain.
func Read(name, version string) Info {
	info := Info{
		Name:      name,
		Version:   version,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				info.Commit = s.Value
			case "vcs.time":
				info.BuildTime = s.Value
			}
		}
	}
	return info
}
