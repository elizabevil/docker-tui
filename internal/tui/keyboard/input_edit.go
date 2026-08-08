package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// editQueryInput maps terminal keys to QueryInputState editing methods.
func editQueryInput(key string, input *state.QueryInputState) (bool, bool) {
	if input == nil {
		return false, false
	}
	input.Clamp()
	switch key {
	case keys.KeyLeft, "ctrl+b":
		input.Move(-1)
		return true, false
	case keys.KeyRight, "ctrl+f":
		input.Move(1)
		return true, false
	case "alt+b":
		input.MoveWordBackward()
		return true, false
	case "alt+f":
		input.MoveWordForward()
		return true, false
	case "ctrl+a", keys.KeyHome:
		input.MoveHome()
		return true, false
	case "ctrl+e", keys.KeyEnd:
		input.MoveEnd()
		return true, false
	case keys.KeyBackspace, "ctrl+h":
		return true, input.DeleteBackward()
	case keys.KeyDelete:
		return true, input.DeleteForward()
	case "ctrl+u":
		return true, input.DeleteToStart()
	case "ctrl+k":
		return true, input.DeleteToEnd()
	case "ctrl+w":
		return true, input.DeleteWordBackward()
	default:
		// Per the IME compose fix (see .omo/ime-compose-shortcuts.md):
		// during Chinese IME composing the terminal may send a multi-rune
		// key string for a single user gesture (e.g. "ab" while typing
		// pinyin). The previous strict len([]rune(key)) != 1 check
		// returned (false, false) for those, letting the key fall through
		// to HandleKeyPress' global action table where letter
		// shortcuts (j/k/i/...) misfired. The field owns the input;
		// consume the key and let input.Insert attempt to place the
		// runes. Returning handled=true is the contract that suppresses
		// shortcut detection upstream.
		if len([]rune(key)) == 0 {
			return false, false
		}
		return true, input.Insert(key)
	}
}
