package component

// ── Common display strings ─────────────────────────────────────
const (
	StrLoading     = "Loading..."
	StrNoContainer = "No containers using this image."
	StrDash        = "\u2014"
	StrNone        = "<none>"

	// BlockCursor is the solid block character used as a text-input cursor.
	BlockCursor = "\u2588"
	// TriangleUp is a solid upward triangle used as a sort indicator.
	TriangleUp = "\u25b2"
	// TriangleDown is a solid downward triangle used as a sort indicator.
	TriangleDown = "\u25bc"
	// ButtonIndicator is the right-pointing triangle used as a button / option marker
	// (e.g. exec, choice, selection, form dialogs).
	ButtonIndicator = "\u25b6"

	// LinkUp is the U+25CF BLACK CIRCLE used to indicate an active link
	// (rendered with the Success color in the header).
	LinkUp = "\u25cf"

	// MarkCheck is the U+2713 CHECK MARK used inside boolean form fields
	// when the field is toggled on.
	MarkCheck = "\u2713"

	// NarrowCursor is the U+258F LEFT ONE QUARTER BLOCK used as the in-text
	// caret for editable form fields; rendered between / after the runes
	// to mark the focused position.
	NarrowCursor = "\u258f"

	// TriangleDownSmall is the U+25BE SMALL DOWN-POINTING TRIANGLE used as
	// the dropdown indicator on select / multiselect form cells.
	TriangleDownSmall = "\u25be"

	// NetInternal is the U+26B2 NEUTER SIGN appended to internal Docker
	// network names so users can tell them apart from bridge networks.
	NetInternal = "\u26b2"

	// BulletEmpty is the U+25CB WHITE CIRCLE used by the header link
	// indicator to denote a Danger state (contrast with LinkUp).
	BulletEmpty = "\u25cb"

	// BoxHorizontal is the U+2500 BOX DRAWINGS LIGHT HORIZONTAL used as
	// the section delimiter in detail page section headers.
	BoxHorizontal = "\u2500"
)
