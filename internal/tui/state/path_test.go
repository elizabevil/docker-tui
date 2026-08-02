package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExpandPathTilde(t *testing.T) {
	got, err := ExpandPath("~/dev/dtui", "/home/alice")
	if err != nil {
		t.Fatalf("ExpandPath returned err: %v", err)
	}
	want := "/home/alice/dev/dtui"
	if got != want {
		t.Errorf("ExpandPath(~) = %q, want %q", got, want)
	}
}

func TestExpandPathEnv(t *testing.T) {
	t.Setenv("MYDIR", "work/common")
	got, err := ExpandPath("$MYDIR/app", "")
	if err != nil {
		t.Fatalf("ExpandPath returned err: %v", err)
	}
	if got != "work/common/app" {
		t.Errorf("ExpandPath($MYDIR) = %q, want %q", got, "work/common/app")
	}
	gotBrace, _ := ExpandPath("${MYDIR}/x", "")
	if gotBrace != "work/common/x" {
		t.Errorf("ExpandPath(${MYDIR}) = %q, want %q", gotBrace, "work/common/x")
	}
}

func TestAbsolute(t *testing.T) {
	if got := Absolute("/etc/nginx", "/w"); got != "/etc/nginx" {
		t.Errorf("Absolute(abs) = %q, want /etc/nginx", got)
	}
	if got := Absolute("rel/x", "/w"); got != filepath.Clean("/w/rel/x") {
		t.Errorf("Absolute(rel) = %q, want %q", got, filepath.Clean("/w/rel/x"))
	}
	if got := Absolute("", "/w"); got != "" {
		t.Errorf("Absolute(empty) = %q, want empty", got)
	}
}

func TestSanitizeName(t *testing.T) {
	cases := []struct{ in, fallback, want string }{
		{"/web/nginx", "sh", "web-nginx"},
		{"a:b|*?", "sh", "a-b---"},
		{"", "sh", "sh"},
	}
	for _, c := range cases {
		if got := SanitizeName(c.in, c.fallback); got != c.want {
			t.Errorf("SanitizeName(%q,%q) = %q, want %q", c.in, c.fallback, got, c.want)
		}
	}
}

func TestPathBase(t *testing.T) {
	cases := []struct{ in, want string }{
		{"/var/log/app.conf", "app.conf"},
		{"/", "rootfs"},
		{"", "rootfs"},
		{"/etc", "etc"},
	}
	for _, c := range cases {
		if got := PathBase(c.in); got != c.want {
			t.Errorf("PathBase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDefaultCopyName(t *testing.T) {
	fixed := time.Date(2026, 8, 2, 15, 30, 12, 0, time.UTC)
	got := DefaultCopyName("/home/user", "web/nginx", "/etc/nginx.conf", "abc", fixed)
	want := filepath.Join("/home/user", "web-nginx-nginx.conf-20260802-153012.tar")
	if got != want {
		t.Errorf("DefaultCopyName = %q, want %q", got, want)
	}
}

func TestDefaultExportName(t *testing.T) {
	fixed := time.Date(2026, 8, 2, 15, 30, 12, 0, time.UTC)
	got := DefaultExportName("/out", "web", "", fixed)
	want := filepath.Join("/out", "web-filesystem-20260802-153012.tar")
	if got != want {
		t.Errorf("DefaultExportName = %q, want %q", got, want)
	}
}

func TestDefaultImageSaveName(t *testing.T) {
	fixed := time.Date(2026, 8, 2, 15, 30, 12, 0, time.UTC)
	got := DefaultImageSaveName("/out", "registry.example/app:v1", "abc", fixed)
	want := filepath.Join("/out", "app-v1-20260802-153012.tar")
	if got != want {
		t.Errorf("DefaultImageSaveName = %q, want %q", got, want)
	}
}

func TestLocalPathProviderReadsDir(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "conf.d"), []byte{}, 0o644)
	os.Mkdir(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, ".hidden"), []byte{}, 0o644)

	p := LocalPathProvider{CWD: dir}
	entries, err := p.Complete(PathCompletionRequest{Path: dir + string(filepath.Separator), Mode: PathAny})
	if err != nil {
		t.Fatalf("Complete returned err: %v", err)
	}
	names := map[string]PathEntry{}
	for _, e := range entries {
		names[e.Name] = e
	}
	if _, ok := names["conf.d"]; !ok {
		t.Errorf("expected conf.d in candidates, got %v", names)
	}
	if e, ok := names["sub"]; !ok || !e.IsDir {
		t.Errorf("expected dir sub.IsDir=true, got %v", e)
	}
	if _, ok := names[".hidden"]; ok {
		t.Errorf("hidden file must be excluded unless prefix starts with '.', got %v", names)
	}
}

func TestLocalPathProviderModeFileIncludesDirsForTraversal(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "d"), 0o755)
	os.WriteFile(filepath.Join(dir, "f"), []byte{}, 0o644)
	p := LocalPathProvider{CWD: dir}
	entries, err := p.Complete(PathCompletionRequest{Path: dir + "/", Mode: PathFile})
	if err != nil {
		t.Fatalf("Complete returned err: %v", err)
	}
	foundDir, foundFile := false, false
	for _, e := range entries {
		foundDir = foundDir || e.Name == "d" && e.IsDir
		foundFile = foundFile || e.Name == "f" && !e.IsDir
	}
	if !foundDir || !foundFile {
		t.Errorf("PathFile must include directories for traversal and files for selection: %v", entries)
	}
}

func TestLocalPathProviderMissingDirReturnsError(t *testing.T) {
	p := LocalPathProvider{CWD: t.TempDir()}
	entries, err := p.Complete(PathCompletionRequest{Path: "/nonexistent-xyz/path", Mode: PathAny})
	if err == nil {
		t.Fatal("Complete on missing dir must return an error")
	}
	if len(entries) != 0 {
		t.Errorf("expected no candidates, got %v", entries)
	}
}
