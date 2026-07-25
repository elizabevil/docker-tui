package component

// NewSelectionProviderFromFn wraps selection-preview creation with a unified
// enable check, reducing duplicated conditionals in page renderers.
func NewSelectionProviderFromFn(fn func() string) SelectionInfoProvider {
	if !SelectionInfoEnabled() || fn == nil {
		return nil
	}
	return &FuncInfoProvider{Fn: fn}
}
