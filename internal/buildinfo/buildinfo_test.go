package buildinfo

import (
	"encoding/json"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestReadFillsStaticFields(t *testing.T) {
	info := Read("dtui")
	if info.Name != "dtui" {
		t.Errorf("Name = %q, want %q", info.Name, "dtui")
	}
	if info.Version != version {
		t.Errorf("Version = %q, want %q", info.Version, version)
	}
	if info.GoVersion == "" {
		t.Error("GoVersion empty")
	}
	wantPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if info.Platform != wantPlatform {
		t.Errorf("Platform = %q, want %q", info.Platform, wantPlatform)
	}
}

func TestReadJSONRoundTrip(t *testing.T) {
	info := Info{
		Name:      "dtui",
		Version:   "0.2.0",
		Commit:    "abc1234",
		BuildTime: "2026-07-24T15:00:00Z",
		GoVersion: "go1.24.0",
		Platform:  "linux/amd64",
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Info
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(got, info) {
		t.Errorf("round trip mismatch:\n got  %+v\n want %+v", got, info)
	}
}

func TestReadOMitsEmptyVCSFields(t *testing.T) {
	info := Info{Name: "dtui", Version: "0.2.0", GoVersion: "x", Platform: "x"}
	data, _ := json.Marshal(info)
	s := string(data)
	if !strings.Contains(s, `"name":"dtui"`) {
		t.Errorf("expected name field, got %s", s)
	}
	if strings.Contains(s, "commit") || strings.Contains(s, "buildTime") {
		t.Errorf("expected empty commit/buildTime to be omitted, got %s", s)
	}
}
