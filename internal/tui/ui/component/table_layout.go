package component

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/utils"
)

type tableLayout struct {
	headers       []string
	widths        []int
	gapWidths     []int
	gapStrings    []string
	trailingWidth int
}

// resolveTableLayout is intentionally uncached. Computing the cache key
// requires the same visible-cell scan as layout resolution and costs more
// allocations than resolving these small column sets directly.
func resolveTableLayout(data TableData, headers []string, containerWidth, baseGap int) tableLayout {
	desiredWidths := tableContentWidths(data, headers)
	resolved := tables.ResolveContentLayout(data.Cols, desiredWidths, containerWidth, baseGap)
	gapWidths := adaptiveGapWidths(resolved.Gap, max(0, len(resolved.Widths)-1), resolved.TrailingWidth)
	gapStrings := make([]string, len(gapWidths))
	for i, width := range gapWidths {
		gapStrings[i] = strings.Repeat(" ", width)
	}
	return tableLayout{
		headers:       headers,
		widths:        resolved.Widths,
		gapWidths:     gapWidths,
		gapStrings:    gapStrings,
		trailingWidth: resolved.TrailingWidth,
	}
}

func tableContentWidths(data TableData, headers []string) []int {
	widths := make([]int, len(data.Cols))
	for i := range widths {
		if i < len(headers) {
			widths[i] = utils.VisibleLen(headers[i])
		}
	}
	for _, row := range data.Rows {
		for i, cell := range row {
			if i < len(widths) {
				widths[i] = max(widths[i], utils.VisibleLen(cell))
			}
		}
	}
	return widths
}
