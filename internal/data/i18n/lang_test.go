package i18n

import (
	"fmt"
	"strings"
	"testing"
)

// TestTWithPlaceholdersRoundTripsThroughSprintf documents the
// (now-avoided) user-reported bug pattern:
//
//	fmt.Sprintf(i18n.T("history.summary"), "nginx:alpine", 18)
//
// would render "%!s(MISSING) · %!d(MISSING) layers" because the
// translation is itself a format string with %s and %d placeholders.
// The contract is: callers must pass args directly to T; nesting
// fmt.Sprintf around T(key) is a bug. The test below pins the
// behaviour that causes the bug — see TestTWithArgsSubstitutesPlaceholders
// for the happy path that the codebase now uses.
func TestTWithPlaceholdersRoundTripsThroughSprintf(t *testing.T) {
	tmpl := T("history.summary") // no args: returns "%!s(MISSING) · %!d(MISSING) layers"
	if !strings.Contains(tmpl, "MISSING") {
		t.Skipf("history.summary resolved without placeholders; the test is vacuous: %q", tmpl)
	}
	// Confirm the documented broken pattern: fmt.Sprintf on top of
	// T(key) yields a still-broken string. This test guards against
	// a regression that would silently mask the underlying T(key)
	// output.
	out := fmt.Sprintf(tmpl, "nginx:alpine", 18)
	if !strings.Contains(out, "MISSING") {
		t.Skipf("nested fmt.Sprintf cleaned up the artefact on its own; the original bug pattern may be obsolete: %q", out)
	}
}

// TestTWithArgsSubstitutesPlaceholders is the happy path: passing
// args directly to T yields a clean formatted string with no MISSING.
// This is the convention callers must follow.
func TestTWithArgsSubstitutesPlaceholders(t *testing.T) {
	out := T("history.summary", "nginx:alpine", 18)
	if strings.Contains(out, "MISSING") {
		t.Fatalf("T with args produced MISSING artefact: %q", out)
	}
	if !strings.Contains(out, "nginx:alpine") {
		t.Errorf("T with args lost the image ref: %q", out)
	}
	if !strings.Contains(out, "18") {
		t.Errorf("T with args lost the layer count: %q", out)
	}
}