package keys

type Modifier string

const (
	ModCtrl  Modifier = "ctrl"
	ModShift Modifier = "shift"
	ModAlt   Modifier = "alt"
)

// Combo returns a combination string like "ctrl+d" or "shift+tab".
func Combo(mod Modifier, key string) string {
	return string(mod) + "+" + key
}
