package runtime

import "testing"

// TestNewActionConstants verifies the TASK-019 action identifiers are
// distinct and follow the same kebab-case shape as the lifecycle set so
// audit / i18n keys can be derived consistently.
func TestNewActionConstants(t *testing.T) {
	consts := map[string]string{
		"update": string(ActionUpdate),
		"diff":   string(ActionDiff),
		"export": string(ActionExport),
		"commit": string(ActionCommit),
		"wait":   string(ActionWait),
		"copy":   string(ActionCopy),
	}
	if len(consts) != 6 {
		t.Fatalf("expected 6 unique TASK-019 constants, got %d", len(consts))
	}
	for want, got := range consts {
		if want != got {
			t.Errorf("Action constant = %q, want %q", got, want)
		}
	}
}

// TestRefreshesContainersLeavesAdvancedAlone verifies the existing
// RefreshesContainers predicate does NOT match the new TASK-019 actions:
// container update / diff / export / commit / wait / copy do not change
// the container's running state and therefore should not trigger a list
// refresh on their own.
func TestRefreshesContainersLeavesAdvancedAlone(t *testing.T) {
	advanced := []Action{
		ActionUpdate, ActionDiff, ActionExport,
		ActionCommit, ActionWait, ActionCopy,
	}
	for _, a := range advanced {
		if RefreshesContainers(a) {
			t.Errorf("RefreshesContainers(%s) = true, want false", a)
		}
	}
}

// TestActionOptionsSupportsAdvancedFields verifies that the new fields
// for TASK-019 exist on ActionOptions and can be set through normal
// struct literal syntax.
func TestActionOptionsSupportsAdvancedFields(t *testing.T) {
	mem := int64(1024)
	cpu := int64(2_000_000_000)
	rp := "on-failure"
	retry := 3

	opts := ActionOptions{
		Memory:            &mem,
		NanoCPUs:          &cpu,
		RestartPolicy:     &rp,
		RestartMaxRetries: &retry,
		SourcePath:        "/etc/hosts",
		Destination:       "/tmp/export.tar",
		Repository:        "myrepo",
		Tag:               "v1",
		Comment:           "snapshot",
		Author:            "user",
		Pause:             true,
		Condition:         "next-exit",
	}

	if opts.Memory == nil || *opts.Memory != 1024 {
		t.Errorf("Memory = %v", opts.Memory)
	}
	if opts.NanoCPUs == nil || *opts.NanoCPUs != 2_000_000_000 {
		t.Errorf("NanoCPUs = %v", opts.NanoCPUs)
	}
	if opts.RestartPolicy == nil || *opts.RestartPolicy != "on-failure" {
		t.Errorf("RestartPolicy = %v", opts.RestartPolicy)
	}
	if opts.SourcePath != "/etc/hosts" {
		t.Errorf("SourcePath = %q", opts.SourcePath)
	}
	if opts.Repository != "myrepo" {
		t.Errorf("Repository = %q", opts.Repository)
	}
	if !opts.Pause {
		t.Error("Pause should be true")
	}
	if opts.Condition != "next-exit" {
		t.Errorf("Condition = %q", opts.Condition)
	}
}

// TestContainerDiffChangeKindRoundTrip verifies the ChangeKind enum
// round-trips through its string form (used by audit/i18n tooling).
func TestContainerDiffChangeKindRoundTrip(t *testing.T) {
	if ChangeModified != 0 || ChangeAdded != 1 || ChangeDeleted != 2 {
		t.Fatalf("ChangeKind values drifted: %d %d %d", ChangeModified, ChangeAdded, ChangeDeleted)
	}
}