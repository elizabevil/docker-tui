package component

import (
	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/data/config"
)

const (
	borderLineHorizontal     = "─"
	borderLineVertical       = "│"
	borderRoundedTopLeft     = "╭"
	borderRoundedTopRight    = "╮"
	borderRoundedBottomLeft  = "╰"
	borderRoundedBottomRight = "╯"
	borderDoubleHorizontal   = "═"
	borderDoubleVertical     = "║"
	borderDoubleTopLeft      = "╔"
	borderDoubleTopRight     = "╗"
	borderDoubleBottomLeft   = "╚"
	borderDoubleBottomRight  = "╝"
	borderThickHorizontal    = "━"
	borderThickVertical      = "┃"
	borderThickTopLeft       = "┏"
	borderThickTopRight      = "┓"
	borderThickBottomLeft    = "┗"
	borderThickBottomRight   = "┛"
	borderSingleTopLeft      = "┌"
	borderSingleTopRight     = "┐"
	borderSingleBottomLeft   = "└"
	borderSingleBottomRight  = "┘"
	borderHiddenGlyph        = " "
)

// BorderDef defines the 8 characters that make up a complete box-drawing border.
type BorderDef struct {
	Top         string `json:"top"`
	Bottom      string `json:"bottom"`
	Left        string `json:"left"`
	Right       string `json:"right"`
	TopLeft     string `json:"topLeft"`
	TopRight    string `json:"topRight"`
	BottomLeft  string `json:"bottomLeft"`
	BottomRight string `json:"bottomRight"`
}

type BorderDefinitions struct {
	Rounded BorderDef `json:"rounded"`
	Double  BorderDef `json:"double"`
	Thick   BorderDef `json:"thick"`
	Single  BorderDef `json:"single"`
	Hidden  BorderDef `json:"hidden"`
}

// ResolveBorder resolves a border style name (e.g. "rounded", "double") to a
// lipgloss.Border. Returns the rounded border as fallback if the style is
// unknown or empty. This is used by the top-level theme to select the outer
// window border character set.
func ResolveBorder(kind config.BorderKind) lipgloss.Border {
	borders := defaultBorderDefs()
	var def BorderDef
	switch kind {
	case config.BorderDouble:
		def = borders.Double
	case config.BorderThick:
		def = borders.Thick
	case config.BorderSingle:
		def = borders.Single
	case config.BorderHidden:
		def = borders.Hidden
	case config.BorderRounded:
		def = borders.Rounded
	default:
		def = defaultBorderDefs().Rounded
	}
	return lipgloss.Border{
		Top:         def.Top,
		Bottom:      def.Bottom,
		Left:        def.Left,
		Right:       def.Right,
		TopLeft:     def.TopLeft,
		TopRight:    def.TopRight,
		BottomLeft:  def.BottomLeft,
		BottomRight: def.BottomRight,
	}
}

func defaultBorderDefs() BorderDefinitions {
	return BorderDefinitions{
		Rounded: BorderDef{
			Top: borderLineHorizontal, Bottom: borderLineHorizontal, Left: borderLineVertical, Right: borderLineVertical,
			TopLeft: borderRoundedTopLeft, TopRight: borderRoundedTopRight, BottomLeft: borderRoundedBottomLeft, BottomRight: borderRoundedBottomRight,
		},
		Double: BorderDef{
			Top: borderDoubleHorizontal, Bottom: borderDoubleHorizontal, Left: borderDoubleVertical, Right: borderDoubleVertical,
			TopLeft: borderDoubleTopLeft, TopRight: borderDoubleTopRight, BottomLeft: borderDoubleBottomLeft, BottomRight: borderDoubleBottomRight,
		},
		Thick: BorderDef{
			Top: borderThickHorizontal, Bottom: borderThickHorizontal, Left: borderThickVertical, Right: borderThickVertical,
			TopLeft: borderThickTopLeft, TopRight: borderThickTopRight, BottomLeft: borderThickBottomLeft, BottomRight: borderThickBottomRight,
		},
		Single: BorderDef{
			Top: borderLineHorizontal, Bottom: borderLineHorizontal, Left: borderLineVertical, Right: borderLineVertical,
			TopLeft: borderSingleTopLeft, TopRight: borderSingleTopRight, BottomLeft: borderSingleBottomLeft, BottomRight: borderSingleBottomRight,
		},
		Hidden: BorderDef{
			Top: borderHiddenGlyph, Bottom: borderHiddenGlyph, Left: borderHiddenGlyph, Right: borderHiddenGlyph,
			TopLeft: borderHiddenGlyph, TopRight: borderHiddenGlyph, BottomLeft: borderHiddenGlyph, BottomRight: borderHiddenGlyph,
		},
	}
}
