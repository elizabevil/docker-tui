package state

import "testing"

func TestFeedbackStateOwnsToastLifecycle(t *testing.T) {
	var feedback FeedbackState
	feedback.ShowToast("failed", NotificationError, 2)
	if feedback.ToastGeneration != 1 {
		t.Fatalf("toast generation = %d, want 1", feedback.ToastGeneration)
	}
	if !feedback.TickToast() || feedback.ToastMessage != "failed" || feedback.ToastTimer != 1 {
		t.Fatalf("first tick = %#v", feedback)
	}
	if !feedback.TickToast() || feedback.ToastMessage != "" || feedback.ToastTimer != 0 {
		t.Fatalf("expired toast = %#v", feedback)
	}
	feedback.ShowToast("next", NotificationInfo, 1)
	if feedback.ToastGeneration != 2 {
		t.Fatalf("next toast generation = %d, want 2", feedback.ToastGeneration)
	}
}

func TestFeedbackStateOwnsErrorLifecycle(t *testing.T) {
	var feedback FeedbackState
	feedback.RecordError("first")
	feedback.RecordError("second")
	if feedback.ErrorMessage != "second" || feedback.ErrorCount != 2 {
		t.Fatalf("recorded errors = %#v", feedback)
	}
	feedback.ClearError()
	if feedback.ErrorMessage != "" || feedback.ErrorCount != 0 {
		t.Fatalf("cleared errors = %#v", feedback)
	}
}
