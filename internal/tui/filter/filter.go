// Package filter centralizes the table / compose filter state machine.
//
// The Controller wraps all filter operations behind a single API so
// the rest of the codebase no longer needs to know about per-panel
// table filter fields, the FilterInput buffer, the ModeFilter mode,
// or the cursor/ViewOffset coupling.
//
// Dependency direction:
//
//	state  <-  no deps
//	utils  <-  state
//	filter <-  state, utils
//	keyboard (and other tui consumers) ->  filter, state, utils
//
// The filter package deliberately does NOT import keyboard/update/etc.
// to avoid circular dependencies.
package filter

import (
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/utils"
)

// Controller owns the filter state for a given AppModel. Construct one
// per keystroke via New; the Controller is cheap and stateless across
// methods.
type Controller struct {
	m *state.AppModel
}

// New returns a Controller bound to the given AppModel. m may be nil
// (all methods become no-ops).
func New(m *state.AppModel) *Controller {
	return &Controller{m: m}
}

// IsInputMode reports whether the filter input bar is open.
func (c *Controller) IsInputMode() bool {
	return c.m != nil && c.m.Navigation.Mode == state.ModeFilter
}

// Active returns the current filter text for the active panel, or "".
func (c *Controller) Active() string {
	if c.m == nil {
		return ""
	}
	if f := utils.ActiveTableFilter(c.m); f != nil {
		return f.FilterText()
	}
	if c.m.Navigation.ActivePanel == state.PanelCompose {
		if c.m.Compose.ComposeFocus == 1 {
			return c.m.Compose.ComposeServiceFilter
		}
		return c.m.Compose.ComposeProjectFilter
	}
	return ""
}

// HasActive reports whether any panel currently has a non-empty filter.
func (c *Controller) HasActive() bool {
	return c.Active() != ""
}

// Open enters ModeFilter and seeds FilterInput with the current filter
// text of the active panel (so the user can edit instead of retype).
func (c *Controller) Open() {
	if c.m == nil {
		return
	}
	c.m.Navigation.Mode = state.ModeFilter
	switch c.m.Navigation.ActivePanel {
	case state.PanelCompose:
		if c.m.Compose.ComposeFocus == 1 {
			c.m.Navigation.FilterInput.Set(c.m.Compose.ComposeServiceFilter)
		} else {
			c.m.Navigation.FilterInput.Set(c.m.Compose.ComposeProjectFilter)
		}
	default:
		if f := utils.ActiveTableFilter(c.m); f != nil {
			c.m.Navigation.FilterInput.Set(f.FilterText())
		} else {
			c.m.Navigation.FilterInput.Reset()
		}
	}
	c.m.Navigation.ClearFilterExit()
}

// Apply sets the active panel's filter to text (no-op if the panel
// does not support filtering or if c.m is nil).
func (c *Controller) Apply(text string) {
	if c.m == nil {
		return
	}
	if c.m.Navigation.ActivePanel == state.PanelCompose {
		if c.m.Compose.ComposeFocus == 1 {
			c.m.Compose.ComposeServiceFilter = text
		} else {
			c.m.Compose.ComposeProjectFilter = text
		}
		return
	}
	if f := utils.ActiveTableFilter(c.m); f != nil {
		f.SetFilter(text)
		c.clampCursor()
	}
}

// ApplyCurrent applies the current FilterInput.Text to the active panel.
// Used by the keystroke handler when the user types in the filter bar.
func (c *Controller) ApplyCurrent() {
	if c.m == nil {
		return
	}
	c.Apply(c.m.Navigation.FilterInput.Text)
}

// ClampCursor resets the active panel's Cursor/ViewOffset to 0 so the
// selection cannot point past the (now potentially shorter) filtered
// list.
func (c *Controller) ClampCursor() {
	c.clampCursor()
}

func (c *Controller) clampCursor() {
	if c.m == nil {
		return
	}
	switch c.m.Navigation.ActivePanel {
	case state.PanelContainers:
		c.m.Resources.Containers.Cursor = 0
		c.m.Resources.Containers.ViewOffset = 0
	case state.PanelImages:
		c.m.Resources.Images.Cursor = 0
		c.m.Resources.Images.ViewOffset = 0
	case state.PanelVolumes:
		c.m.Resources.Volumes.Cursor = 0
		c.m.Resources.Volumes.ViewOffset = 0
	case state.PanelNetworks:
		c.m.Resources.Networks.Cursor = 0
		c.m.Resources.Networks.ViewOffset = 0
	case state.PanelAudit:
		c.m.Audit.Cursor = 0
	}
}

// Clear removes the filter on the active panel (stays in current mode).
// Used when the user wants to drop the filter but keep the input state.
func (c *Controller) Clear() {
	if c.m == nil {
		return
	}
	c.m.Navigation.FilterInput.Reset()
	c.Apply("")
	c.clampCursor()
}

// Close clears the filter and exits filter input mode (or, if called
// from a non-input mode, simply clears and leaves the mode alone after
// resetting the input buffer). Also clears the FilterExitPending flag
// that gates the double-Esc-to-exit flow.
func (c *Controller) Close() {
	if c.m == nil {
		return
	}
	c.Clear()
	c.m.Navigation.CancelFilterExit()
	if c.m.Navigation.Mode == state.ModeFilter {
		c.m.Navigation.Mode = state.ModeNormal
	}
}
