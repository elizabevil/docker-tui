package detail

import "testing"

func TestClassifyImageHeader(t *testing.T) {
	cases := []struct {
		line   string
		want   imageDetailSection
		wantOk bool
	}{
		{imgHeaderSystem, imgSectionSystem, true},
		{imgHeaderRuntime, imgSectionConfig, true},
		{imgHeaderEntrypoint, imgSectionConfig, true},
		{imgHeaderVolumes, imgSectionVolumes, true},
		{imgHeaderHealth, imgSectionHealth, true},
		{imgHeaderStorage, imgSectionStorage, true},
		{imgHeaderLabels, imgSectionLabels, true},
		{"not a header", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, ok := classifyImageHeader(tc.line)
		if ok != tc.wantOk {
			t.Errorf("classifyImageHeader(%q) ok = %v, want %v", tc.line, ok, tc.wantOk)
		}
		if ok && got != tc.want {
			t.Errorf("classifyImageHeader(%q) = %q, want %q", tc.line, got, tc.want)
		}
	}
}

func TestImageEnvironmentHeaderPrefixSuffix(t *testing.T) {
	if imgEnvironmentHeaderPrefix != "── Environment (" {
		t.Errorf("prefix drift: %q", imgEnvironmentHeaderPrefix)
	}
	if imgEnvironmentHeaderSuffix != " vars) ──" {
		t.Errorf("suffix drift: %q", imgEnvironmentHeaderSuffix)
	}
}
