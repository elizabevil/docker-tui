package docker

import "testing"

func TestMapContainerInspectBuildsRuntimeDetail(t *testing.T) {
	raw := []byte(`{
		"Id":"1234567890123456","Name":"/api","Created":"2026-07-21T00:00:00Z","Platform":"linux",
		"State":{"Status":"running","Pid":42,"StartedAt":"2026-07-21T00:01:00Z"},
		"Config":{"Image":"alpine","Cmd":["sleep","10"],"Env":["MODE=test"],"ExposedPorts":{"8080/tcp":{}},"Labels":{"app":"api"}},
		"HostConfig":{"Memory":1024,"NetworkMode":"bridge","RestartPolicy":{"Name":"always","MaximumRetryCount":3}},
		"NetworkSettings":{"Networks":{"bridge":{"IPAddress":"10.0.0.2","Gateway":"10.0.0.1","MacAddress":"aa:bb"}},"Ports":{"8080/tcp":[{"HostIp":"127.0.0.1","HostPort":"18080"}]}},
		"Mounts":[{"Source":"/src","Destination":"/dst","Mode":"z","RW":true}]
	}`)

	detail, err := mapContainerInspect(raw)
	if err != nil {
		t.Fatalf("mapContainerInspect() error = %v", err)
	}
	if detail.ID != "1234567890123456" || detail.State.PID != 42 || detail.Config.Command[0] != "sleep" {
		t.Fatalf("unexpected detail: %#v", detail)
	}
	if detail.Networks["bridge"].MACAddress != "aa:bb" || detail.Ports["8080/tcp"][0].HostPort != "18080" {
		t.Fatalf("unexpected network mapping: %#v %#v", detail.Networks, detail.Ports)
	}
	if len(detail.Mounts) != 1 || !detail.Mounts[0].ReadWrite {
		t.Fatalf("unexpected mounts: %#v", detail.Mounts)
	}
}
