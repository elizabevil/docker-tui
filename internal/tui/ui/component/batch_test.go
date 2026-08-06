package component

import (
	"strings"
	"testing"
)

func TestBatchProgressIndicatorRender(t *testing.T) {
	t.Run("zero total renders empty", func(t *testing.T) {
		p := NewBatchProgressIndicator("stopping", 0, 0)
		if got := p.Render(30); got != "" {
			t.Fatalf("zero-total progress should render empty, got %q", got)
		}
	})

	t.Run("mid progress shows fraction and action", func(t *testing.T) {
		p := NewBatchProgressIndicator("stopping", 2, 5)
		got := p.Render(30)
		if !strings.Contains(got, "2/5") {
			t.Fatalf("progress must contain 2/5: %q", got)
		}
		if !strings.Contains(got, "stopping") {
			t.Fatalf("progress must contain action: %q", got)
		}
		if !strings.Contains(got, "%") {
			t.Fatalf("progress must contain percentage: %q", got)
		}
		if p.Done() {
			t.Fatalf("2/5 must not be done")
		}
	})

	t.Run("complete marks done", func(t *testing.T) {
		p := NewBatchProgressIndicator("stopping", 5, 5)
		if !p.Done() {
			t.Fatalf("5/5 must be done")
		}
	})
}

func TestBulkActionToastText(t *testing.T) {
	t.Run("success only", func(t *testing.T) {
		toast := NewBulkActionToast(3, 0, 0)
		got := toast.Text("succeeded", "skipped", "failed")
		if !strings.Contains(got, "succeeded 3") {
			t.Fatalf("unexpected summary: %q", got)
		}
		if strings.Contains(got, "failed") {
			t.Fatalf("must not mention failed: %q", got)
		}
	})

	t.Run("mixed outcome", func(t *testing.T) {
		toast := NewBulkActionToast(3, 1, 1)
		got := toast.Text("succeeded", "skipped", "failed")
		if !strings.Contains(got, "succeeded 3") ||
			!strings.Contains(got, "skipped 1") ||
			!strings.Contains(got, "failed 1") {
			t.Fatalf("unexpected summary: %q", got)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		toast := NewBulkActionToast(0, 0, 0)
		if got := toast.Text("succeeded", "skipped", "failed"); got != "" {
			t.Fatalf("empty result should render empty, got %q", got)
		}
	})

	t.Run("default labels", func(t *testing.T) {
		toast := NewBulkActionToast(1, 0, 0)
		got := toast.Text("", "", "")
		if !strings.Contains(got, "succeeded 1") {
			t.Fatalf("default label missing: %q", got)
		}
	})

	t.Run("level reflects failure", func(t *testing.T) {
		if NewBulkActionToast(0, 0, 2).Level != ToastError {
			t.Fatalf("all-failed must be ToastError")
		}
		if NewBulkActionToast(2, 0, 1).Level != ToastWarning {
			t.Fatalf("partial must be ToastWarning")
		}
		if NewBulkActionToast(2, 0, 0).Level != ToastSuccess {
			t.Fatalf("all-success must be ToastSuccess")
		}
	})
}
