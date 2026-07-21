package state

import "testing"

func TestFeedbackStateOwnsToastLifecycle(t *testing.T) {
	var feedback FeedbackState
	feedback.ShowToast("failed", NotificationError, 2)
	if !feedback.TickToast() || feedback.ToastMessage != "failed" || feedback.ToastTimer != 1 {
		t.Fatalf("first tick = %#v", feedback)
	}
	if !feedback.TickToast() || feedback.ToastMessage != "" || feedback.ToastTimer != 0 {
		t.Fatalf("expired toast = %#v", feedback)
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
