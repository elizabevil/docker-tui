package state

import "testing"

func TestLogStateLifecycleAndLimit(t *testing.T) {
	var logs LogState
	logs.Open("container")
	lines := make([]string, MaxLogLines+2)
	for i := range lines {
		lines[i] = "line"
	}
	logs.Append(lines)
	if logs.LogContainerID != "container" || len(logs.LogContent) != MaxLogLines {
		t.Fatalf("logs = id %q, lines %d", logs.LogContainerID, len(logs.LogContent))
	}
	logs.Close()
	if logs.LogContainerID != "" || logs.LogContent != nil {
		t.Fatalf("closed logs = %#v", logs)
	}
}

func TestLogStateSearchOwnsViewport(t *testing.T) {
	logs := LogState{LogContent: []string{"ready", "first failed", "ok", "second failed"}}
	if count := logs.ApplySearch("failed"); count != 2 || logs.LogViewOffset != 1 {
		t.Fatalf("search = count %d, state %#v", count, logs)
	}
	logs.MoveMatch(1)
	if logs.LogViewOffset != 3 {
		t.Fatalf("next match offset = %d", logs.LogViewOffset)
	}
}

func TestLogVisibleOffsetDoesNotMutateState(t *testing.T) {
	logs := LogState{LogViewOffset: 99}
	if got := logs.VisibleOffset(10, 3); got != 7 || logs.LogViewOffset != 99 {
		t.Fatalf("visible offset = %d, state = %d", got, logs.LogViewOffset)
	}
}
