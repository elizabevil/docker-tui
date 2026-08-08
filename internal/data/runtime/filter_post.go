package runtime

import "slices"

import "strings"

// FilterSlice returns the items for which match returns true. If match returns
// an error, FilterSlice aborts and propagates it. Adapters reuse this helper
// to implement their resource-specific post-filters.
func FilterSlice[T any](items []T, match func(T) (bool, error)) ([]T, error) {
	filtered := make([]T, 0, len(items))
	for _, item := range items {
		matched, err := match(item)
		if err != nil {
			return nil, err
		}
		if matched {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

// MatchesLabel evaluates a label expression of the form "key" (key exists) or
// "key=value" (key matches the given value).
func MatchesLabel(labels map[string]string, expression string) bool {
	key, expected, hasValue := strings.Cut(expression, "=")
	actual, ok := labels[key]
	return ok && (!hasValue || actual == expected)
}

// ContainsString reports whether values contains expected.
func ContainsString(values []string, expected string) bool {
	return slices.Contains(values, expected)
}

// ContainsReference reports whether any element of references is equal to or
// contains expected as a substring. It is used for image reference matching
// where a partial repo/name/tag match is acceptable.
func ContainsReference(references []string, expected string) bool {
	for _, reference := range references {
		if reference == expected || strings.Contains(reference, expected) {
			return true
		}
	}
	return false
}

// IsDanglingImage reports whether the image has no repo tags or only the
// placeholder "<none>:<none>" tag.
func IsDanglingImage(repoTags []string) bool {
	return len(repoTags) == 0 || (len(repoTags) == 1 && repoTags[0] == "<none>:<none>")
}
