// Package podman provides the Podman transport and optional CGO driver
// used by the runtime/podman adapter.
//
// The package keeps Podman wire types, REST transport, error payloads,
// and CGO bindings behind a narrow boundary so the runtime adapter can
// focus on runtime.Engine semantics, domain mapping, and capability
// reporting.
package podman
