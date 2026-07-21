package runtime

// FilterSet contains runtime-neutral list filters. Resource services define
// their supported field names and adapters translate them to native filters.
type FilterSet map[string][]string

func (f FilterSet) Add(field, value string) {
	if field == "" || value == "" {
		return
	}
	f[field] = append(f[field], value)
}

func (f FilterSet) Values(field string) []string {
	return append([]string(nil), f[field]...)
}

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
