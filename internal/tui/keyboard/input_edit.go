package keyboard

import (
	"unicode"

	"github.com/elizabevil/docker-tui/internal/tui/keys"
)

// editTextInput handles shell-style editing independently from business actions.
// It returns whether the key was handled and whether the text changed.
func editTextInput(key string, text *string, cursor *int) (bool, bool) {
	if text == nil || cursor == nil {
		return false, false
	}
	runes := []rune(*text)
	if *cursor < 0 {
		*cursor = 0
	}
	if *cursor > len(runes) {
		*cursor = len(runes)
	}

	switch key {
	case keys.KeyLeft, "ctrl+b":
		if *cursor > 0 {
			*cursor--
		}
		return true, false
	case keys.KeyRight, "ctrl+f":
		if *cursor < len(runes) {
			*cursor++
		}
		return true, false
	case "alt+b":
		*cursor = previousWordStart(runes, *cursor)
		return true, false
	case "alt+f":
		*cursor = nextWordEnd(runes, *cursor)
		return true, false
	case "ctrl+a", keys.KeyHome:
		*cursor = 0
		return true, false
	case "ctrl+e", keys.KeyEnd:
		*cursor = len(runes)
		return true, false
	case keys.KeyBackspace, "ctrl+h":
		if *cursor == 0 {
			return true, false
		}
		index := *cursor
		*text = string(append(runes[:index-1], runes[index:]...))
		*cursor--
		return true, true
	case keys.KeyDelete:
		if *cursor >= len(runes) {
			return true, false
		}
		index := *cursor
		*text = string(append(runes[:index], runes[index+1:]...))
		return true, true
	case "ctrl+u":
		if *cursor == 0 {
			return true, false
		}
		*text = string(runes[*cursor:])
		*cursor = 0
		return true, true
	case "ctrl+k":
		if *cursor >= len(runes) {
			return true, false
		}
		*text = string(runes[:*cursor])
		return true, true
	case "ctrl+w":
		start := previousWordStart(runes, *cursor)
		if start == *cursor {
			return true, false
		}
		*text = string(append(runes[:start], runes[*cursor:]...))
		*cursor = start
		return true, true
	default:
		if len([]rune(key)) != 1 {
			return false, false
		}
		insert := []rune(key)
		index := *cursor
		*text = string(append(append(runes[:index:index], insert...), runes[index:]...))
		*cursor += len(insert)
		return true, true
	}
}

func previousWordStart(text []rune, cursor int) int {
	for cursor > 0 && unicode.IsSpace(text[cursor-1]) {
		cursor--
	}
	for cursor > 0 && !unicode.IsSpace(text[cursor-1]) {
		cursor--
	}
	return cursor
}

func nextWordEnd(text []rune, cursor int) int {
	for cursor < len(text) && unicode.IsSpace(text[cursor]) {
		cursor++
	}
	for cursor < len(text) && !unicode.IsSpace(text[cursor]) {
		cursor++
	}
	return cursor
}
