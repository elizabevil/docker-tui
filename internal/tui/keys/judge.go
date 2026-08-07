package keys

// IsEnter returns true if the key is the Enter/Return key.
func IsEnter(key string) bool {
	return key == KeyEnter
}

// IsEsc returns true if the key is the Escape key.
func IsEsc(key string) bool {
	return key == KeyEsc
}

// IsSpace returns true if the key is the Space key.
func IsSpace(key string) bool {
	return key == KeySpace || key == KeySpaceName
}

// IsTab returns true if the key is the Tab key.
func IsTab(key string) bool {
	return key == KeyTab
}

// IsBackspace returns true if the key is the Backspace key.
func IsBackspace(key string) bool {
	return key == KeyBackspace
}

// IsNavigation returns true if the key is a navigation key (arrows, j/k, pgup/pgdn).
func IsNavigation(key string) bool {
	switch key {
	case KeyUp, KeyDown, KeyLeft, KeyRight,
		KeyJ, KeyK,
		KeyPgUp, KeyPgDn:
		return true
	}
	return false
}

// IsConfirm returns true if the key confirms an action (Enter, y, Y).
func IsConfirm(key string) bool {
	return key == KeyEnter || key == KeyY || key == KeyYUpper
}

// IsCancel returns true if the key cancels an action (Esc, n, N).
func IsCancel(key string) bool {
	return key == KeyEsc || key == KeyN || key == KeyNUpper
}

// IsQuit returns true if the key is the Quit key (q).
func IsQuit(key string) bool {
	return key == KeyQ
}

// IsDelete returns true if the key is the Delete combination (ctrl+d).
func IsDelete(key string) bool {
	return key == KeyCtrlD
}

// IsFilter returns true if the key triggers a filter/search.
func IsFilter(key string) bool {
	return key == KeySlash || key == KeyQmark
}
