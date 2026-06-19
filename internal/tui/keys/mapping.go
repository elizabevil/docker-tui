package keys

// KeyMapping maps key sequences to actions.
type KeyMapping map[string]KeyAction

// DefaultKeyMapping returns the default key bindings.
func DefaultKeyMapping() KeyMapping {
	return KeyMapping{
		KeyQ:        ActionQuit,
		KeyCtrlC:    ActionQuit,
		KeyQmark:    ActionHelp,
		KeyF1:       ActionHelp,
		"F1":        ActionHelp,
		KeySlash:    ActionFilter,
		KeyR:        ActionRefresh,
		KeyS:        ActionContainerStart,
		KeyD:        ActionDetail,
		KeyL:        ActionContainerLogs,
		KeyI:        ActionContainerInspect,
		KeyE:        ActionContainerExec,
		KeyM:        ActionContainerStats,
		KeyP:        ActionImagePrune,
		KeyTab:      ActionTabNext,
		KeyShiftTab: ActionTabPrev,
		KeyUp:       ActionUp,
		KeyK:        ActionUp,
		KeyDown:     ActionDown,
		KeyJ:        ActionDown,
		KeyEnter:    ActionEnter,
		KeyEsc:      ActionBack,
		":":         ActionCommand,
		// Ctrl+ 组合键：替代原大写功能（已不区分大小写）
		KeyCtrlS: ActionContainerStop,
		KeyCtrlK: ActionContainerKill,
		KeyCtrlR: ActionContainerRestart,
		KeyCtrlP: ActionImagePull,
	}
}
