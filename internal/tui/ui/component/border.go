package component

import (
	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/data/config"
)

const (
	BorderLineHorizontal     = "─"
	BorderLineVertical       = "│"
	BorderRoundedTopLeft     = "╭"
	BorderRoundedTopRight    = "╮"
	BorderRoundedBottomLeft  = "╰"
	BorderRoundedBottomRight = "╯"
	BorderDoubleHorizontal   = "═"
	BorderDoubleVertical     = "║"
	BorderDoubleTopLeft      = "╔"
	BorderDoubleTopRight     = "╗"
	BorderDoubleBottomLeft   = "╚"
	BorderDoubleBottomRight  = "╝"
	BorderThickHorizontal    = "━"
	BorderThickVertical      = "┃"
	BorderThickTopLeft       = "┏"
	BorderThickTopRight      = "┓"
	BorderThickBottomLeft    = "┗"
	BorderThickBottomRight   = "┛"
	BorderSingleTopLeft      = "┌"
	BorderSingleTopRight     = "┐"
	BorderSingleBottomLeft   = "└"
	BorderSingleBottomRight  = "┘"
	BorderHiddenGlyph        = " "
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
			Top: BorderLineHorizontal, Bottom: BorderLineHorizontal, Left: BorderLineVertical, Right: BorderLineVertical,
			TopLeft: BorderRoundedTopLeft, TopRight: BorderRoundedTopRight, BottomLeft: BorderRoundedBottomLeft, BottomRight: BorderRoundedBottomRight,
		},
		Double: BorderDef{
			Top: BorderDoubleHorizontal, Bottom: BorderDoubleHorizontal, Left: BorderDoubleVertical, Right: BorderDoubleVertical,
			TopLeft: BorderDoubleTopLeft, TopRight: BorderDoubleTopRight, BottomLeft: BorderDoubleBottomLeft, BottomRight: BorderDoubleBottomRight,
		},
		Thick: BorderDef{
			Top: BorderThickHorizontal, Bottom: BorderThickHorizontal, Left: BorderThickVertical, Right: BorderThickVertical,
			TopLeft: BorderThickTopLeft, TopRight: BorderThickTopRight, BottomLeft: BorderThickBottomLeft, BottomRight: BorderThickBottomRight,
		},
		Single: BorderDef{
			Top: BorderLineHorizontal, Bottom: BorderLineHorizontal, Left: BorderLineVertical, Right: BorderLineVertical,
			TopLeft: BorderSingleTopLeft, TopRight: BorderSingleTopRight, BottomLeft: BorderSingleBottomLeft, BottomRight: BorderSingleBottomRight,
		},
		Hidden: BorderDef{
			Top: BorderHiddenGlyph, Bottom: BorderHiddenGlyph, Left: BorderHiddenGlyph, Right: BorderHiddenGlyph,
			TopLeft: BorderHiddenGlyph, TopRight: BorderHiddenGlyph, BottomLeft: BorderHiddenGlyph, BottomRight: BorderHiddenGlyph,
		},
	}
}
