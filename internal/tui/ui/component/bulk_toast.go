package component

import (
	"fmt"
	"strings"
)

// BulkActionToast summarizes the aggregate outcome of a batch operation
// (e.g. "3 succeeded, 1 failed"). It is business-agnostic: counts are
// supplied by the caller and formatted with the given i18n labels.
type BulkActionToast struct {
	Success int
	Skipped int
	Failed  int
	Level   ToastLevel
}

// NewBulkActionToast builds a bulk summary toast from the result counts.
func NewBulkActionToast(success, skipped, failed int) BulkActionToast {
	level := ToastSuccess
	switch {
	case failed > 0 && success > 0:
		level = ToastWarning
	case failed > 0:
		level = ToastError
	case success == 0 && skipped > 0:
		level = ToastWarning
	}
	return BulkActionToast{Success: success, Skipped: skipped, Failed: failed, Level: level}
}

// Text renders the summary line using the provided i18n labels.
// Labels are plain words (e.g. "succeeded") rendered as "label count".
func (t BulkActionToast) Text(successLabel, skippedLabel, failedLabel string) string {
	if successLabel == "" {
		successLabel = "succeeded"
	}
	if skippedLabel == "" {
		skippedLabel = "skipped"
	}
	if failedLabel == "" {
		failedLabel = "failed"
	}
	var parts []string
	if t.Success > 0 {
		parts = append(parts, fmt.Sprintf("%s %d", successLabel, t.Success))
	}
	if t.Skipped > 0 {
		parts = append(parts, fmt.Sprintf("%s %d", skippedLabel, t.Skipped))
	}
	if t.Failed > 0 {
		parts = append(parts, fmt.Sprintf("%s %d", failedLabel, t.Failed))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ", ")
}
