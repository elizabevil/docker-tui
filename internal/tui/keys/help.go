package keys

import (
	tea "charm.land/bubbletea/v2"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
)

// HelpEntry describes a key binding for the help view.
type HelpEntry struct {
	Key         string
	Description string
}

// HelpEntries returns all available key bindings.
func HelpEntries() []HelpEntry {
	return []HelpEntry{
		{Key: KUpDown + " or j/k", Description: i18n.T("key.down") + "/" + i18n.T("key.up")},
		{Key: KTab + " / Shift+Tab", Description: i18n.T("key.next") + "/" + i18n.T("key.back")},
		{Key: KEnter, Description: i18n.T("key.expand")},
		{Key: KEsc, Description: i18n.T("key.back")},
		{Key: KRight + " / " + KLeft, Description: i18n.T("key.expand") + " / " + i18n.T("key.collapse")},
		{Key: KeyS, Description: i18n.T("key.start")},
		{Key: KeyS_up, Description: i18n.T("key.stop")},
		{Key: KeyR_up, Description: i18n.T("key.restart")},
		{Key: KeyK_up, Description: i18n.T("key.kill")},
		{Key: KeyD, Description: i18n.T("key.detail")},
		{Key: KeyL, Description: i18n.T("key.logs")},
		{Key: KeyM, Description: i18n.T("key.stats")},
		{Key: KeyP_up, Description: i18n.T("key.pull")},
		{Key: KeyP, Description: i18n.T("key.prune")},
		{Key: KeyD_up, Description: i18n.T("key.debug")},
		{Key: KeyE_up, Description: i18n.T("key.export")},
		{Key: KCtrlD, Description: i18n.T("key.delete")},
		{Key: KeySpace_pretty, Description: i18n.T("key.mark")},
		{Key: KeyH, Description: i18n.T("key.toggle")},
		{Key: KeySlash, Description: i18n.T("key.filter")},
		{Key: KeyR, Description: i18n.T("key.restart")},
		{Key: KeyQmark, Description: i18n.T("key.help")},
		{Key: KeyQ, Description: i18n.T("key.quit")},
	}
}

// KeyFromMsg extracts a key string from a Bubbletea key message.
func KeyFromMsg(msg tea.KeyPressMsg) string {
	return msg.String()
}
