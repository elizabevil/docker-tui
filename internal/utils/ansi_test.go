package utils

import (
	"reflect"
	"testing"
)

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

func TestWrapCellsUsesTerminalWidth(t *testing.T) {
	got := WrapCells("ab中文cd", 4)
	want := []string{"ab中", "文cd"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("WrapCells = %#v, want %#v", got, want)
	}
}

func TestFitVisiblePreservesANSIAndExactWidth(t *testing.T) {
	got := FitVisible("\x1b[31mabcdefgh\x1b[0m", 5)
	if DisplayWidth(got) != 5 || StripANSI(got) != "abcde" {
		t.Fatalf("FitVisible = %q, width=%d", got, DisplayWidth(got))
	}
}
