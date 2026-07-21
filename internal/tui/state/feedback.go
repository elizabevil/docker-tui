package state

// NotificationLevel is semantic state; UI packages decide how to render it.
type NotificationLevel int

const (
	NotificationSuccess NotificationLevel = iota
	NotificationInfo
	NotificationWarning
	NotificationError
)

// FeedbackState owns transient messages, operation summaries and key feedback.
type FeedbackState struct {
	ErrorMessage          string
	ErrorCount            int
	InfoMessage           string
	AuditOperationMessage string
	KeyHint               string
	KeyHintTimer          int
	KeyStrokeBuffer       []KeyStrokeEvent
	KeyStrokeTimer        int
	KeyStrokeAnim         int
	LastKeyStroke         []KeyStrokeEvent
	KeyStrokeDispTimer    int
	KeyStrokeDuration     int
	KeyStrokeAnimFrames   int
	ToastMessage          string
	ToastLevel            NotificationLevel
	ToastTimer            int
}

func NewFeedbackState() FeedbackState {
	return FeedbackState{KeyStrokeDuration: 30, KeyStrokeAnimFrames: 5}
}

func (s *FeedbackState) ShowToast(message string, level NotificationLevel, ticks int) {
	s.ToastMessage = message
	s.ToastLevel = level
	s.ToastTimer = max(0, ticks)
}

func (s *FeedbackState) TickToast() bool {
	if s.ToastTimer <= 0 {
		return false
	}
	s.ToastTimer--
	if s.ToastTimer == 0 {
		s.ToastMessage = ""
		s.ToastLevel = NotificationSuccess
	}
	return true
}

func (s *FeedbackState) RecordError(message string) {
	s.ErrorMessage = message
	s.ErrorCount++
}

func (s *FeedbackState) ClearError() {
	s.ErrorMessage = ""
	s.ErrorCount = 0
}

func (s *FeedbackState) SetKeyHint(message string, ticks int) {
	s.KeyHint = message
	s.KeyHintTimer = max(0, ticks)
}

// KeyStrokeEvent represents a single key press for the keystroke display column.
type KeyStrokeEvent struct {
	Key    string
	Action string
}
