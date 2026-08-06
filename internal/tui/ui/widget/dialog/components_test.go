package dialog

import (
	"strings"
	"testing"
)

func TestFormSectionHeaderRender(t *testing.T) {
	t.Run("renders title followed by rule", func(t *testing.T) {
		got := NewFormSectionHeader("Image Options").Render(30)
		if !strings.Contains(got, "Image Options") {
			t.Fatalf("header must contain title: %q", got)
		}
		if !strings.Contains(got, "\u2500") {
			t.Fatalf("header must contain horizontal rule: %q", got)
		}
	})

	t.Run("empty title renders pure rule", func(t *testing.T) {
		got := NewFormSectionHeader("").Render(10)
		if !strings.Contains(got, "\u2500") {
			t.Fatalf("empty header should be a rule: %q", got)
		}
	})
}

func TestFormDisplayRender(t *testing.T) {
	t.Run("label plus value", func(t *testing.T) {
		got := NewFormDisplay("Target ID", "abc123").Render(20)
		if !strings.Contains(got, "Target ID") || !strings.Contains(got, "abc123") {
			t.Fatalf("display must include label and value: %q", got)
		}
	})

	t.Run("value only", func(t *testing.T) {
		got := NewFormDisplay("", "some value").Render(20)
		if !strings.Contains(got, "some value") {
			t.Fatalf("display must include value: %q", got)
		}
	})
}

func TestFormInfoRender(t *testing.T) {
	t.Run("info banner has icon and text", func(t *testing.T) {
		got := NewFormInfo("registry will be contacted").Render(30)
		if !strings.Contains(got, "registry will be contacted") {
			t.Fatalf("info banner must include text: %q", got)
		}
	})

	t.Run("warning banner distinct", func(t *testing.T) {
		got := NewFormWarning("this will delete data").Render(30)
		if !strings.Contains(got, "this will delete data") {
			t.Fatalf("warning banner must include text: %q", got)
		}
	})
}

func TestFormListPaneRender(t *testing.T) {
	t.Run("all items without pagination", func(t *testing.T) {
		pane := NewFormListPane([]string{"a", "b", "c"})
		got := pane.Render(20)
		for _, item := range []string{"a", "b", "c"} {
			if !strings.Contains(got, item) {
				t.Fatalf("pane missing item %q: %q", item, got)
			}
		}
	})

	t.Run("paginates and shows page indicator", func(t *testing.T) {
		pane := NewFormListPane([]string{"a", "b", "c", "d", "e"})
		pane.PageSize = 2
		pane.Page = 1
		if pane.TotalRows() != 2 {
			t.Fatalf("page 1 should show 2 rows, got %d", pane.TotalRows())
		}
		if pane.PageCount() != 3 {
			t.Fatalf("5 items at page size 2 should be 3 pages, got %d", pane.PageCount())
		}
		got := pane.Render(20)
		if !strings.Contains(got, "2/3") {
			t.Fatalf("pane must show page indicator: %q", got)
		}
		if strings.Contains(got, "\n  a\n") {
			t.Fatalf("page 1 should not contain item a: %q", got)
		}
	})

	t.Run("title renders above items", func(t *testing.T) {
		pane := NewFormListPane([]string{"x"})
		pane.Title = "Affected"
		got := pane.Render(20)
		if !strings.Contains(got, "Affected") {
			t.Fatalf("pane must include title: %q", got)
		}
	})
}
