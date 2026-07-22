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

// NativeFilters returns a copy suitable for driver encoding. Multiple values
// for one field require AND semantics and are rejected until that field has an
// equivalent native representation or a lossless post-filter.
func (o ImageListOptions) NativeFilters() (map[string][]string, error) {
	filters := make(map[string][]string, len(o.Filters))
	for field, values := range o.Filters {
		if !isImageFilter(field) {
			return nil, NewError(ErrorInvalid, "image.list.filter", field, fmt.Errorf("unknown image filter"))
		}
		if len(values) > 1 {
			return nil, UnsupportedError("image.list.filter." + field)
		}
		if len(values) == 1 {
			filters[field] = append([]string(nil), values...)
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
