package audit

import (
	"strings"
	"testing"
	"time"

	auditdata "github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func TestRenderListUsesSharedFlexLayout(t *testing.T) {
	auditState := &state.AuditState{
		Records: []auditdata.Record{{
			Time:    time.Date(2026, 7, 31, 12, 34, 56, 0, time.UTC),
			Action:  "container.restart",
			Result:  auditdata.ResultSucceeded,
			Level:   auditdata.LevelInfo,
			Message: "restart completed without errors and retained the full audit context",
			Target:  auditdata.TargetDTO{Name: "dtui-test-redis"},
		}},
	}

	const pageWidth = 140
	rendered := component.StripANSI(RenderList(auditState, pageWidth, 12))
	if !strings.Contains(rendered, "restart completed") {
		t.Fatalf("message column was not rendered: %q", rendered)
	}

	wantRowWidth := pageWidth
	for line := range strings.SplitSeq(rendered, "\n") {
		if strings.Contains(line, "container.restart") {
			if got := component.VisibleLen(line); got != wantRowWidth {
				t.Fatalf("audit row width = %d, want %d: %q", got, wantRowWidth, line)
			}
			return
		}
	}
	t.Fatal("audit row was not rendered")
}

func TestRenderListKeepsAuditCursorInsideVisibleRows(t *testing.T) {
	auditState := &state.AuditState{Records: make([]auditdata.Record, 30), Cursor: 20}
	for i := range auditState.Records {
		auditState.Records[i] = auditdata.Record{
			Time:   time.Date(2026, 7, 31, 12, 34, i, 0, time.UTC),
			Action: "container.inspect",
			Result: auditdata.ResultSucceeded,
			Level:  auditdata.LevelInfo,
			Target: auditdata.TargetDTO{Name: "container"},
		}
	}

	RenderList(auditState, 140, 12)
	rowHeight := component.CalcTableRowHeight(12, false)
	wantOffset := auditState.Cursor - rowHeight + 1
	if auditState.ViewOffset != wantOffset {
		t.Fatalf("ViewOffset = %d, want %d", auditState.ViewOffset, wantOffset)
	}
}
