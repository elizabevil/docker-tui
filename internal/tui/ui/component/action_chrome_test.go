package component

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

// TestWrapActionWindowRendersChrome is the R06-05 acceptance check: the
// wrapper actually emits ANSI escape sequences for the border + the
// resolved background colour. If a future change accidentally bypasses
// the StyleName lookup (e.g. by hard-coding transparent), this test
// fails.
func TestWrapActionWindowRendersChrome(t *testing.T) {
	resolved, err := config.LoadResolved(config.LoadOptions{ThemeName: config.ThemeName("light")})
	if err != nil {
		t.Fatalf("LoadResolved light: %v", err)
	}
	ApplyThemeStyles(resolved.Theme)

	got := WrapActionWindow("hello", config.OperationScopeContainer, 40)
	if got == "hello" {
		t.Fatalf("WrapActionWindow returned the raw content unchanged; chrome not applied")
	}
	// The wrapper emits border glyphs (RoundedBorder uses ─ │ ┌ ┐ └ ┘).
	if !strings.Contains(got, "─") {
		t.Errorf("expected horizontal border glyph in wrapped output, got %q", got)
	}
	// And the light-theme hex for action.container.window.border (#3875d7).
	if !strings.Contains(got, "56;117;215") {
		t.Errorf("expected light-theme window.border RGB in wrapped output, got %q", got)
	}
}

// TestWrapActionWindowImageScopeUsesImageTheme verifies the chrome
// honours the scope parameter — image-scope pages must NOT pull the
// container-scope border colour.
func TestWrapActionWindowImageScopeUsesImageTheme(t *testing.T) {
	resolved, err := config.LoadResolved(config.LoadOptions{ThemeName: config.ThemeName("light")})
	if err != nil {
		t.Fatalf("LoadResolved light: %v", err)
	}
	ApplyThemeStyles(resolved.Theme)

	container := WrapActionWindow("c", config.OperationScopeContainer, 40)
	image := WrapActionWindow("i", config.OperationScopeImage, 40)
	if !strings.Contains(container, "56;117;215") {
		t.Errorf("container scope chrome missing light-theme border colour: %q", container)
	}
	if strings.Contains(image, "56;117;215") {
		t.Errorf("image scope chrome should not pull container.border (light-theme #3875d7): %q", image)
	}
}

// TestWrapActionWindowUnknownScopeReturnsPlainContent guards the
// zero-value fallback: an unrecognised scope must NOT panic and must
// NOT inject chrome (otherwise callers could be tricked into using
// the wrong theme contract).
func TestWrapActionWindowUnknownScopeReturnsPlainContent(t *testing.T) {
	got := WrapActionWindow("plain", config.OperationScope("bogus"), 40)
	if got != "plain" {
		t.Errorf("unknown scope should bypass chrome, got %q", got)
	}
}

// TestWaitIndicatorRendersShortID is the R06-05 Async-mode acceptance
// check: the indicator truncates the full container ID to the short
// form so the message rail line stays single-line.
func TestWaitIndicatorRendersShortID(t *testing.T) {
	full := "cbfdc75a10a1c89ec7ea76a9c4f6dd00b8f81b83f8e812ef374acb1a8a51d648"
	got := NewWaitIndicator(full, 1).Render(40)
	if !strings.Contains(got, "cbfdc75a10a1") {
		t.Errorf("indicator must show truncated id, got %q", got)
	}
	if strings.Contains(got, full) {
		t.Errorf("indicator must NOT show full id, got %q", got)
	}
	if !strings.HasPrefix(got, "⏳") {
		t.Errorf("indicator must lead with hourglass glyph, got %q", got)
	}
}

// TestWaitIndicatorEmptyID covers the corner case where the wait
// dispatch fires without a target (e.g. a future Operation that has
// no container yet). The indicator must still produce output.
func TestWaitIndicatorEmptyID(t *testing.T) {
	got := NewWaitIndicator("", 0).Render(40)
	if got == "" {
		t.Errorf("empty-id indicator should still produce a non-empty string")
	}
}