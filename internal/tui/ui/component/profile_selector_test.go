package component

import "testing"

func TestBreakpointProfileSelector_Select(t *testing.T) {
	s := BreakpointProfileSelector{
		Default: "default",
		Rules: []ProfileRule{
			{Name: "wide", MinWidth: 120},
			{Name: "more", MinWidth: 90},
			{Name: "compact", MinWidth: 70},
		},
	}

	tests := []struct {
		name  string
		width int
		want  string
	}{
		{name: "wide", width: 130, want: "wide"},
		{name: "more", width: 95, want: "more"},
		{name: "compact", width: 75, want: "compact"},
		{name: "default", width: 60, want: "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.Select(tt.width); got != tt.want {
				t.Fatalf("Select(%d) = %q, want %q", tt.width, got, tt.want)
			}
		})
	}
}

func TestContainerProfileSelector_Select(t *testing.T) {
	s := ContainerProfileSelector{
		Default:       "default",
		DefaultStats:  "default_stats",
		More:          "more",
		MoreStats:     "more_stats",
		MoreMinWidth:  90,
		StatsEnabled:  true,
		StatsMinWidth: 60,
	}

	tests := []struct {
		name  string
		width int
		want  string
	}{
		{name: "more_and_stats", width: 100, want: "more_stats"},
		{name: "more_only", width: 90, want: "more_stats"},
		{name: "stats_only", width: 70, want: "default_stats"},
		{name: "default", width: 50, want: "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.Select(tt.width); got != tt.want {
				t.Fatalf("Select(%d) = %q, want %q", tt.width, got, tt.want)
			}
		})
	}
}

func TestContainerProfileSelector_StatsDisabled(t *testing.T) {
	s := ContainerProfileSelector{
		Default:       "default",
		DefaultStats:  "default_stats",
		More:          "more",
		MoreStats:     "more_stats",
		MoreMinWidth:  90,
		StatsEnabled:  false,
		StatsMinWidth: 60,
	}

	if got := s.Select(70); got != "default" {
		t.Fatalf("Select(70) with stats disabled = %q, want default", got)
	}
	if got := s.Select(100); got != "more" {
		t.Fatalf("Select(100) with stats disabled = %q, want more", got)
	}
}
