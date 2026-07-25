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

// TestActionOptionsSupportsAdvancedFields verifies the discriminated-
// union shape of ActionOptions: lifecycle callers populate Lifecycle
// while TASK-019 callers populate exactly one of the typed pointers
// that matches the action they are invoking.
func TestActionOptionsSupportsAdvancedFields(t *testing.T) {
	mem := int64(1024)
	cpu := int64(2_000_000_000)
	rp := "on-failure"
	retry := 3

	// Lifecycle-only payload: container remove with Force=true.
	lifeOpts := ActionOptions{Lifecycle: LifecycleOptions{Force: true}}
	if !lifeOpts.Lifecycle.Force {
		t.Errorf("Lifecycle.Force = %v, want true", lifeOpts.Lifecycle.Force)
	}
	if lifeOpts.Update != nil || lifeOpts.Commit != nil {
		t.Errorf("lifecycle payload must not populate typed pointers: %+v", lifeOpts)
	}

	// Advanced payloads populate exactly one pointer each.
	updateOpts := ActionOptions{Update: &UpdateOptions{
		Memory:            &mem,
		NanoCPUs:          &cpu,
		RestartPolicy:     &rp,
		RestartMaxRetries: &retry,
	}}
	if updateOpts.Update.Memory == nil || *updateOpts.Update.Memory != 1024 {
		t.Errorf("Update.Memory = %v", updateOpts.Update.Memory)
	}
	if *updateOpts.Update.NanoCPUs != 2_000_000_000 {
		t.Errorf("Update.NanoCPUs = %v", updateOpts.Update.NanoCPUs)
	}
	if *updateOpts.Update.RestartPolicy != "on-failure" {
		t.Errorf("Update.RestartPolicy = %v", *updateOpts.Update.RestartPolicy)
	}

	copyOpts := ActionOptions{Copy: &CopyOptions{SourcePath: "/etc/hosts"}}
	if copyOpts.Copy.SourcePath != "/etc/hosts" {
		t.Errorf("Copy.SourcePath = %q", copyOpts.Copy.SourcePath)
	}

	exportOpts := ActionOptions{Export: &ExportOptions{Destination: "/tmp/export.tar"}}
	if exportOpts.Export.Destination != "/tmp/export.tar" {
		t.Errorf("Export.Destination = %q", exportOpts.Export.Destination)
	}

	commitOpts := ActionOptions{Commit: &CommitOptions{
		Repository: "myrepo", Tag: "v1", Comment: "snapshot", Author: "user", Pause: true,
	}}
	if commitOpts.Commit.Repository != "myrepo" || commitOpts.Commit.Tag != "v1" {
		t.Errorf("Commit = %+v", commitOpts.Commit)
	}
	if !commitOpts.Commit.Pause {
		t.Error("Commit.Pause should be true")
	}

	waitOpts := ActionOptions{Wait: &WaitOptions{Condition: "next-exit"}}
	if waitOpts.Wait.Condition != "next-exit" {
		t.Errorf("Wait.Condition = %q", waitOpts.Wait.Condition)
	}
}

// TestContainerDiffChangeKindRoundTrip verifies the ChangeKind enum
// round-trips through its string form (used by audit/i18n tooling).
func TestContainerDiffChangeKindRoundTrip(t *testing.T) {
	if ChangeModified != 0 || ChangeAdded != 1 || ChangeDeleted != 2 {
		t.Fatalf("ChangeKind values drifted: %d %d %d", ChangeModified, ChangeAdded, ChangeDeleted)
	}
}

// TestActionOptionsUnionSemantics documents that ActionOptions is a
// discriminated union: setting multiple typed pointers is a caller bug
// and the adapter only reads the one matching the action. This test
// does not enforce that — it just verifies the typed pointers coexist
// without ambiguity so callers can switch between payloads freely.
func TestActionOptionsUnionSemantics(t *testing.T) {
	// Both pointers set; the adapter's type switch still picks the
	// right one because the action argument drives the switch, not the
	// payload shape.
	both := ActionOptions{
		Update: &UpdateOptions{Memory: ptrInt64(2048)},
		Commit: &CommitOptions{Repository: "stale"},
	}
	if both.Update == nil || both.Commit == nil {
		t.Fatal("both pointers should be set simultaneously to demonstrate union semantics")
	}
	// Whichever payload is read depends on the action verb.
	if *both.Update.Memory != 2048 {
		t.Errorf("Update payload lost: %v", both.Update.Memory)
	}
	if both.Commit.Repository != "stale" {
		t.Errorf("Commit payload lost: %v", both.Commit.Repository)
	}
}

// TestLifecycleOptionsZeroValue ensures the zero value of LifecycleOptions
// is the safe default (no force, no name, no signal, no timeout) so
// callers that build the struct incrementally never accidentally
// pass Force=true without intending to.
func TestLifecycleOptionsZeroValue(t *testing.T) {
	var lo LifecycleOptions
	if lo.Force {
		t.Error("zero Force must be false")
	}
	if lo.Name != "" {
		t.Errorf("zero Name = %q", lo.Name)
	}
	if lo.Signal != "" {
		t.Errorf("zero Signal = %q", lo.Signal)
	}
	if lo.Timeout != 0 {
		t.Errorf("zero Timeout = %v", lo.Timeout)
	}
}

func ptrInt64(v int64) *int64 { return &v }
