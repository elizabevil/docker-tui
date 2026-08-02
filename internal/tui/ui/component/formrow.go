package component

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/utils"
)

// FormLabelColumnWidth computes the label column width for a set of form
// labels. Width accounts for visible (wide-run) characters and is capped by
// the given max ratio of the total available width (BR-041 §9: 外到内).
func FormLabelColumnWidth(labels []string, total int) int {
	if total <= 0 {
		return 0
	}
	maxLabel := 0
	for _, l := range labels {
		w := utils.VisibleLen(l)
		if w > maxLabel {
			maxLabel = w
		}
	}
	maxByRatio := total * 2 / 5
	if maxLabel > maxByRatio {
		maxLabel = maxByRatio
	}
	if maxLabel >= total {
		return total - 1
	}
	return maxLabel
}

// FormRow assembles one form row: a right-aligned label column plus a value
// column. Label is padded to labelWidth; value is truncated to the remaining
// space then padded to maintain cell width. All values share one left edge.
func FormRow(label string, labelWidth, valueWidth int, value string) string {
	if labelWidth < 0 {
		labelWidth = 0
	}
	if valueWidth <= 0 {
		return utils.PadVisible(label, labelWidth)
	}
	label = utils.TruncateVisible(label, labelWidth)
	lbl := strings.Repeat(" ", max(0, labelWidth-utils.DisplayWidth(label))) + label
	val := utils.FitVisible(value, valueWidth)
	return lbl + " " + val
}
