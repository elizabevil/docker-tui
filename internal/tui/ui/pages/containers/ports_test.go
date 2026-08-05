package containers

import (
	"reflect"
	"testing"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func TestFormatPortsPreservesBindings(t *testing.T) {
	bindings := []dockerclient.PortBinding{
		{ContainerPort: 80, Protocol: "tcp", HostIP: "0.0.0.0", HostPort: 8080},
		{ContainerPort: 80, Protocol: "tcp", HostIP: "::", HostPort: 8080},
		{ContainerPort: 53, Protocol: "udp"},
	}
	want := []string{"80/tcp -> :8080", "80/tcp -> [::]:8080", "53/udp"}
	if got := FormatPorts(bindings, false); !reflect.DeepEqual(got, want) {
		t.Fatalf("ports = %#v, want %#v", got, want)
	}
	wantCompact := []string{"8080:80/tcp", "8080:80/tcp", "53/udp"}
	if got := FormatPorts(bindings, true); !reflect.DeepEqual(got, wantCompact) {
		t.Fatalf("compact ports = %#v, want %#v", got, wantCompact)
	}
}

func TestFormatPortsEmpty(t *testing.T) {
	if got := FormatPorts(nil, false); len(got) != 1 || got[0] != component.StrDash {
		t.Fatalf("empty ports = %#v", got)
	}
}
