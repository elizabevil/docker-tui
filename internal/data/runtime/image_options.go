package runtime

import "fmt"

// Image filter constants map 1:1 to Docker/Podman native filter keys.
// Each constant is whitelisted in isImageFilter and validated by NativeFilters.
const (
	ImageFilterReference = "reference"
	ImageFilterLabel     = "label"
	ImageFilterDangling  = "dangling"
	ImageFilterBefore    = "before"
	ImageFilterSince     = "since"
	ImageFilterUntil     = "until"
)

// NativeFilters returns a safe subset for driver encoding. The first value is
// pushed down; adapters apply the complete filter set afterward to preserve
// same-field AND semantics.
func (o ImageListOptions) NativeFilters() (map[string][]string, error) {
	filters := make(map[string][]string, len(o.Filters))
	for field, values := range o.Filters {
		if !isImageFilter(field) {
			return nil, NewError(ErrorInvalid, "image.list.filter", field, fmt.Errorf("unknown image filter"))
		}
		if len(values) > 0 {
			filters[field] = []string{values[0]}
		}
	}
	return filters, nil
}

func isImageFilter(field string) bool {
	switch field {
	case ImageFilterReference, ImageFilterLabel, ImageFilterDangling,
		ImageFilterBefore, ImageFilterSince, ImageFilterUntil:
		return true
	default:
		return false
	}
}
