package tui

import "os"

// SupportsTruecolor reports whether the active terminal renders truecolor
// SGR escape sequences (24-bit RGB). It is used to decide whether alpha
// channels in #RRGGBBAA hex values produce visible transparency.
//
// Detection rules (any match returns true):
//   - COLORTERM is "truecolor" or "24bit"
//   - TERM contains "256color" or "truecolor"
//   - TERM_PROGRAM is "iTerm.app", "WezTerm", "ghostty", or "Alacritty"
//
// Non-truecolor terminals receive alpha-stripped output so dialog overlays
// remain visible as a solid scrim rather than disappearing.
func SupportsTruecolor() bool {
	if v := os.Getenv("COLORTERM"); v == "truecolor" || v == "24bit" {
		return true
	}
	if v := os.Getenv("TERM"); v != "" {
		if containsCI(v, "256color") || containsCI(v, "truecolor") {
			return true
		}
	}
	switch os.Getenv("TERM_PROGRAM") {
	case "iTerm.app", "WezTerm", "ghostty", "Alacritty":
		return true
	}
	return false
}

func containsCI(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i+len(substr) <= len(s); i++ {
		if equalFold(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}