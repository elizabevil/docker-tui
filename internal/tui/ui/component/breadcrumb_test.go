package component

import (
	"testing"
)

func TestRenderBreadcrumb(t *testing.T) {
	items := []BreadcrumbItem{
		{Label: "Images", ID: "images"},
		{Label: "containers", ID: "containers"},
		{Label: "detail", ID: "detail"},
	}

	// Full rendering with enough width
	got := RenderBreadcrumb(items, " > ", 50)
	if got == "" {
		t.Error("RenderBreadcrumb returned empty string")
	}

	// Empty items
	got = RenderBreadcrumb(nil, " > ", 50)
	if got != "" {
		t.Errorf("RenderBreadcrumb(nil) = %q, want empty", got)
	}

	// Zero width
	got = RenderBreadcrumb(items, " > ", 0)
	if got != "" {
		t.Errorf("RenderBreadcrumb(items, 0) = %q, want empty", got)
	}

	// Single item
	single := []BreadcrumbItem{{Label: "Containers", ID: "containers"}}
	got = RenderBreadcrumb(single, " > ", 50)
	if got == "" {
		t.Error("RenderBreadcrumb(single) returned empty")
	}
}

func TestToastQueue(t *testing.T) {
	q := NewToastQueue()

	// Empty queue
	if q.Active() != "" {
		t.Errorf("New queue Active() = %q, want empty", q.Active())
	}

	// Push first message
	q.Push("nginx started", ToastSuccess)
	if q.Active() != "nginx started" {
		t.Errorf("After push Active() = %q, want 'nginx started'", q.Active())
	}

	// Tick until dismissed (30 ticks = 3s at 100ms)
	for i := 0; i < 30; i++ {
		q.Tick()
	}
	if q.Active() != "" {
		t.Errorf("After 30 ticks Active() = %q, want empty", q.Active())
	}
}

func TestToastLevels(t *testing.T) {
	tests := []struct {
		level ToastLevel
		want  string
	}{
		{ToastSuccess, "toastSuccess"},
		{ToastError, "toastError"},
		{ToastInfo, "toastInfo"},
		{ToastWarning, "toastWarning"},
	}
	for _, tt := range tests {
		got := tt.level.styleName()
		if got != tt.want {
			t.Errorf("ToastLevel(%d).styleName() = %q, want %q", tt.level, got, tt.want)
		}
	}
}
