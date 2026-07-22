package state

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
)

// AuditFilterLevel defines the filter scope for audit records.
type AuditFilterLevel int

const (
	AuditFilterAll    AuditFilterLevel = iota
	AuditFilterErrors
	AuditFilterWarnings
)

// AuditState owns the audit history panel state.
type AuditState struct {
	Records      []audit.Record
	Cursor       int
	ViewOffset   int
	filterText   string
	FilterLevel  AuditFilterLevel
	SelectedID   string
	DetailRecord *audit.Record
}

// SetFilter implements TableFilter for the audit panel.
func (s *AuditState) SetFilter(query string) {
	s.filterText = query
	s.Cursor = 0
	s.ViewOffset = 0
}

// FilterText implements TableFilter for the audit panel.
func (s *AuditState) FilterText() string {
	return s.filterText
}

// AuditListModel returns the filtered and sorted audit records for display.
func (s *AuditState) AuditListModel() []audit.Record {
	var filtered []audit.Record
	for _, r := range s.Records {
		if !s.matchFilter(r) {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered
}

func (s *AuditState) matchFilter(r audit.Record) bool {
	if s.FilterLevel == AuditFilterErrors && r.Level != audit.LevelError {
		return false
	}
	if s.FilterLevel == AuditFilterWarnings && r.Level != audit.LevelWarn {
		return false
	}
	if s.filterText == "" {
		return true
	}
	needle := strings.ToLower(s.filterText)
	return strings.Contains(strings.ToLower(r.Action), needle) ||
		strings.Contains(strings.ToLower(r.Message), needle) ||
		strings.Contains(strings.ToLower(r.Target.Type), needle) ||
		strings.Contains(strings.ToLower(r.Target.ID), needle) ||
		strings.Contains(strings.ToLower(r.Target.Name), needle)
}

// Total returns the count of filtered records.
func (s *AuditState) Total() int {
	return len(s.AuditListModel())
}
