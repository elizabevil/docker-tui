package docker

import (
	"errors"
	"testing"

	"github.com/docker/docker/errdefs"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type codedTestError struct{ status int }

func (e codedTestError) Error() string { return "runtime request failed" }
func (e codedTestError) Code() int     { return e.status }

func TestMapRuntimeErrorClassifiesDriverErrors(t *testing.T) {
	ref := runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: "container-id"}
	tests := []struct {
		name string
		err  error
		kind runtimeapi.ErrorKind
	}{
		{name: "docker not found", err: errdefs.NotFound(errors.New("missing")), kind: runtimeapi.ErrorNotFound},
		{name: "podman conflict", err: codedTestError{status: 409}, kind: runtimeapi.ErrorConflict},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := mapRuntimeError(test.err, "container.start", ref, RuntimePodman)
			var runtimeErr *runtimeapi.Error
			if !errors.As(err, &runtimeErr) {
				t.Fatalf("error type = %T, want *runtime.Error", err)
			}
			if runtimeErr.Kind != test.kind || runtimeErr.Resource != ref.ID || runtimeErr.Driver != runtimeapi.Podman {
				t.Fatalf("unexpected mapped error: %#v", runtimeErr)
			}
		})
	}
}

func TestMapRuntimeErrorDoesNotMutateExistingError(t *testing.T) {
	original := &runtimeapi.Error{Kind: runtimeapi.ErrorUnsupported, Operation: "container.pause"}
	mapped := mapRuntimeError(original, "container.pause", runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: "id"}, RuntimeDocker)
	if original.Resource != "" || original.Driver != "" {
		t.Fatalf("input error was mutated: %#v", original)
	}
	var runtimeErr *runtimeapi.Error
	if !errors.As(mapped, &runtimeErr) || runtimeErr.Resource != "id" || runtimeErr.Driver != runtimeapi.Docker {
		t.Fatalf("unexpected mapped error: %#v", mapped)
	}
}
