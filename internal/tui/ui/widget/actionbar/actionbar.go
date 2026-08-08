package actionbar

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	actionmodel "github.com/elizabevil/docker-tui/internal/tui/actionbar"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/dialog"
	"github.com/elizabevil/docker-tui/internal/utils"
)

const maxVisibleActions = 12

func RenderBar(m *state.AppModel, content string, body dialog.PanelBody) string {
	if m == nil || body.Width <= 0 || body.Rows <= 0 {
		return content
	}
	box := renderBox(actionmodel.VisibleItems(m), m.Navigation.ActionBar.Selected,
		m.Navigation.ActionBar.Filtering, m.Navigation.ActionBar.Filter, body, !m.CursorBlinkHidden)
	return dialog.PlaceDialogInPanel(content, box, body, dialog.LoadDialogConfig())
}

func renderBox(items []actionmodel.ActionItem, selected int, filtering bool, filter string, body dialog.PanelBody, cursorVisible ...bool) string {
	// Width is 1/2 of the active panel body so the action bar shares
	// visual weight with the panel underneath (rather than dominating it).
	// Height is 2/3 of the panel body, enforced by limiting the visible
	// action rows so the dialog stays inside the panel even when many
	// actions are available.
	boxWidth := max(body.Width/2, 28)
	if boxWidth > 72 {
		boxWidth = 72
	}
	if maxW := body.Width - 2; maxW > 0 && boxWidth > maxW {
		boxWidth = maxW
	}
	innerWidth := max(4, boxWidth-4)

	actionVisibleRows := max(body.Rows*2/3, 4)
	// Reserve budget for title + filter + blanks + buttons + borders.
	actionVisibleRows -= 5
	if actionVisibleRows < 3 {
		actionVisibleRows = 3
	}

	// Title row occupies the box's full width so the rounded border sits
	// flush with the sides. The box background is intentionally transparent
	// so the panel below shows through, matching the message / query rails
	// and avoiding a "second background" that doesn't agree with the theme.
	barStyle := component.GetStyle(component.StyleActionBar)
	title := barStyle.Foreground(component.GetStyle(component.StylePanelTitle).GetForeground()).
		Width(innerWidth).Render("Action Bar")

	filterRow := renderFilterRow(filtering, filter, innerWidth, cursorVisible...)

	actionLines := renderActionLines(items, selected, innerWidth, actionVisibleRows)

	buttonRow := renderActionBarButtons()

	// The blank separator between action list and buttons gives the footer
	// visual room without forcing a hard border.
	parts := []string{title, filterRow, "", actionLines, "", buttonRow}
	joined := strings.Join(parts, "\n")

	// Outer wrapper: rounded border + 1-cell horizontal padding. No
	// background so the box integrates with the panel it overlays.
	borderFg := component.GetStyle(component.StyleActionBarBorder).GetForeground()
	return barStyle.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderFg).
		Padding(0, 1).
		Width(boxWidth).
		Render(joined)
}

// renderFilterRow builds the / … filter input. The "/" leads with a hint
// colour so it reads as "type-to-filter", and the buffer cursor is the
// existing narrow caret glyph. A real background is applied so the
// input looks like a discrete field instead of a single inline string.
func renderFilterRow(filtering bool, filter string, innerWidth int, cursorVisible ...bool) string {
	barStyle := component.GetStyle(component.StyleActionBar)
	if !filtering {
		hint := "/ filter   1-9 jump   Esc close"
		return barStyle.Foreground(component.GetStyle(component.StyleDim).GetForeground()).
			Width(innerWidth).Render(hint)
	}
	cursor := component.NarrowCursor
	if len(cursorVisible) > 0 && !cursorVisible[0] {
		cursor = " "
	}
	prompt := component.GetStyle(component.StyleDim).Render("/")
	buffer := component.GetStyle(component.StyleHeaderBar).Render(filter + cursor)
	// Use PadVisible to fill the inner width so the row's background
	// (StyleActionBar background) extends to the trailing edge.
	filled := prompt + " " + buffer
	pad := utils.PadVisible(filled, innerWidth)
	return barStyle.Width(innerWidth).Render(pad)
}

// renderActionLines returns the joined action list (one line per action)
// padded to innerWidth so the StyleSelectedRow background fills the
// selected row across the full column. maxRows caps the visible rows so
// the dialog stays within the 2/3-of-panel-row budget the caller's
// layout enforced.
func renderActionLines(items []actionmodel.ActionItem, selected int, innerWidth int, maxRows int) string {
	if len(items) == 0 {
		return component.GetStyle(component.StyleDim).Render("No matching actions")
	}
	selected = max(0, min(selected, len(items)-1))
	availableRows := max(1, min(maxVisibleActions, maxRows))
	start := max(0, selected-availableRows+1)
	end := min(len(items), start+availableRows)
	if end-start < availableRows {
		start = max(0, end-availableRows)
	}
	barStyle := component.GetStyle(component.StyleActionBar)
	lines := make([]string, 0, end-start)
	for index := start; index < end; index++ {
		item := items[index]
		marker := "  "
		if index == selected {
			marker = "> "
		}
		keyWidth := min(14, max(7, innerWidth/4))
		key := component.TruncateVisible(item.Key, keyWidth)
		labelWidth := max(1, innerWidth-keyWidth-3)
		line := fmt.Sprintf("%s%-*s %s", marker, keyWidth, key, component.TruncateVisible(item.Label, labelWidth))
		styled := barStyle.Width(innerWidth).Render(line)
		if item.Disabled {
			styled = component.GetStyle(component.StyleDim).Render(line)
		} else if index == selected {
			styled = component.GetStyle(component.StyleSelectedRow).Width(innerWidth).Render(line)
		}
		lines = append(lines, styled)
	}
	if len(items) > availableRows {
		lines = append(lines, component.GetStyle(component.StyleDim).Render(fmt.Sprintf("%d-%d / %d", start+1, end, len(items))))
	}
	return strings.Join(lines, "\n")
}

// renderActionBarButtons renders the Confirm / Cancel hint row at the
// bottom of the Action Bar. The Confirm button mirrors the dialog's
// StyleDialogConfirm palette so the user can see the same "press Enter to
// commit" affordance already used elsewhere. Cancel uses StyleDim so it
// recedes relative to Confirm.
func renderActionBarButtons() string {
	enterKey := i18n.T("key.sym_enter")
	escKey := i18n.T("key.sym_esc")
	confirmLabel := i18n.T("key.confirm")
	cancelLabel := i18n.T("key.cancel")
	barStyle := component.GetStyle(component.StyleActionBar)
	confirm := barStyle.
		Foreground(component.GetStyle(component.StyleDialogConfirm).GetForeground()).
		Render(enterKey + " " + component.ButtonIndicator + " " + confirmLabel)
	cancel := barStyle.
		Foreground(component.GetStyle(component.StyleDim).GetForeground()).
		Render(escKey + " " + cancelLabel)
	return cancel + "    " + confirm
}
