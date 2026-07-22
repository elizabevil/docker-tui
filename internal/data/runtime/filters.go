package runtime

// FilterSet contains runtime-neutral list filters. Resource services define
// their supported field names and adapters translate them to native filters.
type FilterSet map[string][]string

// Add appends a value to the filter field, ignoring empty strings.
func (f FilterSet) Add(field, value string) {
	if field == "" || value == "" {
		return
	}
	f[field] = append(f[field], value)
}

// Values returns a copy of the values for the given filter field.
func (f FilterSet) Values(field string) []string {
	return append([]string(nil), f[field]...)
}

// Clone returns a deep copy of the filter set.
func (f FilterSet) Clone() FilterSet {
	if f == nil {
		return nil
	}
	clone := make(FilterSet, len(f))
	for field, values := range f {
		clone[field] = append([]string(nil), values...)
	}
	return clone
}
