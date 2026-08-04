package component

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

func TestResolveBorderUsesTypedKinds(t *testing.T) {
	tests := []struct {
		kind config.BorderKind
		top  string
		left string
	}{
		{config.BorderRounded, borderLineHorizontal, borderLineVertical},
		{config.BorderSingle, borderLineHorizontal, borderLineVertical},
		{config.BorderDouble, borderDoubleHorizontal, borderDoubleVertical},
		{config.BorderThick, borderThickHorizontal, borderThickVertical},
		{config.BorderHidden, borderHiddenGlyph, borderHiddenGlyph},
	}
	for _, test := range tests {
		border := ResolveBorder(test.kind)
		if border.Top != test.top || border.Left != test.left {
			t.Errorf("ResolveBorder(%q) = top %q left %q", test.kind, border.Top, border.Left)
		}
	}
}

func TestResolveBorderUnknownKindUsesCompiledFallback(t *testing.T) {
	border := ResolveBorder(config.BorderKind("unknown"))
	if border.TopLeft != borderRoundedTopLeft || border.BottomRight != borderRoundedBottomRight {
		t.Fatalf("unknown border kind did not use rounded fallback: %#v", border)
	}
}
