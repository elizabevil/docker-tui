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
		if len([]rune(key)) != 1 {
			return false, false
		}
		return true, input.Insert(key)
	}
}
