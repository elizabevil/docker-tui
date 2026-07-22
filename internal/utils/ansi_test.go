package utils

import "testing"

func TestDisplayWidthUsesTerminalCells(t *testing.T) {
	if got := DisplayWidth("中文"); got != 4 {
		t.Fatalf("DisplayWidth(Chinese)=%d, want 4", got)
	}
	if got := DisplayWidth("a\u0301"); got != 1 {
		t.Fatalf("DisplayWidth(combining)=%d, want 1", got)
	}
	if got := DisplayWidth("\033[31m中文\033[0m"); got != 4 {
		t.Fatalf("DisplayWidth(ANSI Chinese)=%d, want 4", got)
	}
}

func TestTruncateVisibleUsesCellWidth(t *testing.T) {
	if got := TruncateVisible("中文测试", 5); got != "中..." {
		t.Fatalf("TruncateVisible=%q, want %q", got, "中...")
	}
	if got := PadVisible("中文", 6); got != "中文  " {
		t.Fatalf("PadVisible=%q, want %q", got, "中文  ")
	}
}
