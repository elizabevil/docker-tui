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
	"encoding/json"
	"fmt"
	"io"

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

// MatchCount returns a "(filtered/total)" string for the active panel so
// the user sees the match result in real time while typing in the
// filter bar. Returns "" when the active panel is not filter-capable
// or has no items.
//
// Direct field access on the concrete list models is used (instead of
// widening the TableFilter interface) because each model returns a
// different concrete slice type ([]ContainerSummary, []ImageSummary, ...)
// so a generic []T in the interface would erase that information.
func (c *Controller) MatchCount() string {
	if c.m == nil {
		return ""
	}
	switch c.m.Navigation.ActivePanel {
	case state.PanelContainers:
		total := c.m.Resources.Containers.Len()
		if total <= 0 {
			return ""
		}
		return fmt.Sprintf("(%d/%d)", len(c.m.Resources.Containers.FilteredItems()), total)
	case state.PanelImages:
		total := c.m.Resources.Images.Len()
		if total <= 0 {
			return ""
		}
		return fmt.Sprintf("(%d/%d)", len(c.m.Resources.Images.FilteredItems()), total)
	case state.PanelVolumes:
		total := c.m.Resources.Volumes.Len()
		if total <= 0 {
			return ""
		}
		return fmt.Sprintf("(%d/%d)", len(c.m.Resources.Volumes.FilteredItems()), total)
	case state.PanelNetworks:
		total := c.m.Resources.Networks.Len()
		if total <= 0 {
			return ""
		}
		return fmt.Sprintf("(%d/%d)", len(c.m.Resources.Networks.FilteredItems()), total)
	case state.PanelAudit:
		if c.m.Audit.FilterText() == "" {
			return ""
		}
		// AuditState doesn't expose the unfiltered count, so we show
		// just the filtered count without the ratio.
		return fmt.Sprintf("(%d)", len(c.m.Audit.AuditListModel()))
	}
	return ""
}

// BannerPrefixFor returns a banner prefix string for a panel whose
// TableFilter may or may not be active. Format: "filter: <text> | ".
// Returns "" when f is nil or has no active filter, so callers can
// concatenate safely:
//
//	banner := filter.BannerPrefixFor(cm) + selectionLabel
//
// This is the package-level form (vs. Controller.BannerPrefix) so view
// functions that receive only the list model — not the full AppModel —
// can use it without threading the model through.
func BannerPrefixFor(f state.TableFilter) string {
	if f == nil {
		return ""
	}
	text := f.FilterText()
	if text == "" {
		return ""
	}
	return "filter: " + text + " | "
}

// filterData is the on-disk shape for Save/Load. JSON-marshaled to
// <configDir>/filters.json so the user's per-panel filters survive
// app restarts.
type filterData struct {
	Containers     string `json:"containers"`
	Images         string `json:"images"`
	Volumes        string `json:"volumes"`
	Networks       string `json:"networks"`
	Audit          string `json:"audit"`
	ComposeService string `json:"compose_service,omitempty"`
	ComposeProject string `json:"compose_project,omitempty"`
}

// Save writes the current per-panel filter text to w as JSON.
// Safe on a nil receiver (no-op).
func (c *Controller) Save(w io.Writer) error {
	if c == nil || c.m == nil {
		return nil
	}
	data := filterData{
		Containers:     c.m.Resources.Containers.Filter,
		Images:         c.m.Resources.Images.Filter,
		Volumes:        c.m.Resources.Volumes.Filter,
		Networks:       c.m.Resources.Networks.Filter,
		Audit:          c.m.Audit.FilterText(),
		ComposeService: c.m.Compose.ComposeServiceFilter,
		ComposeProject: c.m.Compose.ComposeProjectFilter,
	}
	return json.NewEncoder(w).Encode(data)
}

// Load reads JSON from r and applies the per-panel filter text to m.
// Safe on a nil m (no-op). An empty input (io.EOF) is also a no-op
// — callers can just always call Load on the opened file.
func Load(m *state.AppModel, r io.Reader) error {
	if m == nil {
		return nil
	}
	var data filterData
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	m.Resources.Containers.SetFilter(data.Containers)
	m.Resources.Images.SetFilter(data.Images)
	m.Resources.Volumes.SetFilter(data.Volumes)
	m.Resources.Networks.SetFilter(data.Networks)
	m.Audit.SetFilter(data.Audit)
	m.Compose.ComposeServiceFilter = data.ComposeService
	m.Compose.ComposeProjectFilter = data.ComposeProject
	return nil
}

// BannerPrefixForCount is like BannerPrefixFor but also includes the
// filtered/total match count, e.g. "filter: foo (3/42) | ". Pass
// filtered=0 or total<=0 to suppress the count.
func BannerPrefixForCount(f state.TableFilter, filtered, total int) string {
	if f == nil {
		return ""
	}
	text := f.FilterText()
	if text == "" {
		return ""
	}
	if total > 0 {
		return fmt.Sprintf("filter: %s (%d/%d) | ", text, filtered, total)
	}
	return "filter: " + text + " | "
}

// BannerPrefix returns a banner prefix string to be prepended to a panel's
// banner when a filter is active. Format: "filter: <text> | ".
// Returns "" when no filter is active, so callers can concatenate safely:
//
//	banner := c.BannerPrefix() + selectionLabel
func (c *Controller) BannerPrefix() string {
	return BannerPrefixFor(c.tableFilter())
}

// tableFilter returns the active panel's TableFilter (no-AppModel form).
func (c *Controller) tableFilter() state.TableFilter {
	if c.m == nil {
		return nil
	}
	return utils.ActiveTableFilter(c.m)
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

// ClearAll wipes the filter on every panel (containers, images,
// volumes, networks, audit, compose) and resets each panel's cursor /
// view-offset. Bound to ActionClearFilters; useful as a one-key
// "reset" when the user has accumulated several narrow filters.
func (c *Controller) ClearAll() {
	if c.m == nil {
		return
	}
	c.m.Navigation.FilterInput.Reset()
	c.m.Resources.Containers.SetFilter("")
	c.m.Resources.Images.SetFilter("")
	c.m.Resources.Volumes.SetFilter("")
	c.m.Resources.Networks.SetFilter("")
	c.m.Audit.SetFilter("")
	c.m.Compose.ComposeServiceFilter = ""
	c.m.Compose.ComposeProjectFilter = ""
	c.m.Resources.Containers.Cursor = 0
	c.m.Resources.Containers.ViewOffset = 0
	c.m.Resources.Images.Cursor = 0
	c.m.Resources.Images.ViewOffset = 0
	c.m.Resources.Volumes.Cursor = 0
	c.m.Resources.Volumes.ViewOffset = 0
	c.m.Resources.Networks.Cursor = 0
	c.m.Resources.Networks.ViewOffset = 0
	c.m.Audit.Cursor = 0
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
