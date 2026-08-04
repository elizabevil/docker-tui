package keys

// ── Display key symbols ────────────────────────────────────────
const (
	KUp    = "\u2191"
	KDown  = "\u2193"
	KLeft  = "\u2190"
	KRight = "\u2192"
	KTab   = "Tab"
	KSpace = "Space"
	KEsc   = "Esc"
	KEnter = "Enter"
	KCtrlD = "Ctrl+D"
	KCtrlK = "Ctrl+K"
	KCtrlS = "Ctrl+S"
	KCtrlE = "Ctrl+E"
	KCtrlP = "Ctrl+P"
	KCtrlR = "Ctrl+R"
	KCtrlB = "Ctrl+B"
	KCtrlG = "Ctrl+G"
	KCtrlN = "Ctrl+N"
	KCtrlO = "Ctrl+O"
	KPgUp  = "PgUp"
	KPgDn  = "PgDn"
	KF1    = "F1"
	KF2    = "F2"

	KJDown   = "j\u2193"
	KKUp     = "k\u2191"
	KJSlashK = "j/k"
	KSShiftS = "s/S"
	KUpDown  = "\u2191\u2193"

	KRightLeft = "\u2192/\u2190"
	KEscEnter  = "Esc/Enter"
	KPgUpDn    = "PgUp/Dn"
	KGG        = "g/G"
)

// ── Display symbols for special keys (shown in keystroke animation) ──
const (
	// DisplaySpace is the U+2423 OPEN BOX used to render the Space key.
	DisplaySpace = "\u2423"
	// DisplayEnter is the U+21B5 DOWNWARDS ARROW WITH CORNER LEFTWARDS used for Enter.
	DisplayEnter = "\u21b5"
	// DisplayTab is the U+21E5 RIGHTWARDS ARROW TO BAR used for Tab.
	DisplayTab = "\u21e5"
	// DisplayBackspace is the U+232B ERASE TO THE LEFT used for Backspace.
	DisplayBackspace = "\u232b"
	// DisplayCtrl is the caret prefix used for Ctrl-modified keys.
	DisplayCtrl = "^"
)

// ── Keystroke action labels (displayed in header bar animation) ──
const (
	ActionLabelStart      = "Start"
	ActionLabelStop       = "Stop"
	ActionLabelRestart    = "Restart"
	ActionLabelDetail     = "Detail"
	ActionLabelDelete     = "Delete"
	ActionLabelLogs       = "Logs"
	ActionLabelExec       = "Exec"
	ActionLabelProjects   = "Projects"
	ActionLabelServices   = "Services"
	ActionLabelContainers = "Containers"
	ActionLabelDown       = "Down"
	ActionLabelSort       = "Sort"
	ActionLabelToggle     = "Toggle"
	ActionLabelMark       = "Mark"
	ActionLabelSwitch     = "Switch"
	ActionLabelHeader     = "Header"
	ActionLabelHelp       = "Help"
	ActionLabelFilter     = "Filter"
	ActionLabelQuit       = "Quit"

	// Additional labels used by KeyStrokeActionLabel.
	ActionLabelRefresh = "Refresh"
	ActionLabelKill    = "Kill"
	ActionLabelInspect = "Inspect"
	ActionLabelStats   = "Stats"
	ActionLabelPull    = "Pull"
	ActionLabelPrune   = "Prune"
	ActionLabelHistory = "History"
	ActionLabelConn    = "Conn"
	ActionLabelOpen    = "Open"
	ActionLabelBack    = "Back"
	ActionLabelPanel   = "Panel"
	ActionLabelUp      = "Up"
	ActionLabelCmd     = "Cmd"
)

// ── Focus indicator colors (ANSI 256-color codes) ───────────────
const (
	FocusActive   = "63"  // 青色：焦点面板边框
	FocusInactive = "240" // 灰色：非焦点面板边框
)