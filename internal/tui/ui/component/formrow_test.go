package component

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/utils"
)

func TestFormLabelColumnWidthUsesVisibleWidth(t *testing.T) {
	// "容器路径" 是 4 个双宽字符（8 cells），"Restart" 是 7 cells。
	labels := []string{"Restart policy", "容器路径"}
	got := FormLabelColumnWidth(labels, 40)
	if got < 8 {
		t.Fatalf("label width = %d, want at least max visible width 8", got)
	}
	if got > 16 { // 40 * 2/5 = 16 上限
		t.Fatalf("label width = %d, want capped by 40*2/5", got)
	}
}

func TestFormLabelColumnWidthCapsByRatio(t *testing.T) {
	labels := []string{"A very very long label that exceeds the cap"}
	wantCap := 50 * 2 / 5
	if got := FormLabelColumnWidth(labels, 50); got > wantCap {
		t.Fatalf("label width = %d, want cap %d", got, wantCap)
	}
}

func TestFormRowLabelRightAlignedAndValueLeftAligned(t *testing.T) {
	rows := []string{
		FormRow("A", 4, 20, "value1"),
		FormRow("AB", 4, 20, "value2"),
	}
	valueEdge := -1
	for _, r := range rows {
		idx := strings.Index(utils.StripANSI(r), "v")
		if idx >= 0 {
			if valueEdge == -1 {
				valueEdge = idx
			} else if idx != valueEdge {
				t.Fatalf("value left edge misaligned: %q vs edge %d", r, valueEdge)
			}
		}
	}
	if clean := utils.StripANSI(rows[0]); !strings.HasPrefix(clean, "   A") {
		t.Fatalf("short label is not right aligned: %q", clean)
	}
}

func TestFormRowTruncatesLongValue(t *testing.T) {
	out := FormRow("L", 1, 5, "abcdefghij")
	clean := utils.StripANSI(out)
	if utils.DisplayWidth(clean) > 7 { // 1 label + 1 space + 5 value
		t.Fatalf("row too wide: %q (%d cells)", clean, utils.DisplayWidth(clean))
	}
}
