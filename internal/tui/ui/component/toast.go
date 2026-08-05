package component

// ToastLevel represents the severity/type of a toast notification.
type ToastLevel int

const (
	ToastSuccess ToastLevel = iota
	ToastError
	ToastInfo
	ToastWarning
)

// defaultDuration returns tick count based on level.
// Error 显示 5s, Warning 4s, Info 3s, Success 2s
func (l ToastLevel) defaultDuration() int {
	switch l {
	case ToastError:
		return 50 // 5s
	case ToastWarning:
		return 40 // 4s
	case ToastInfo:
		return 30 // 3s
	case ToastSuccess:
		return 20 // 2s
	default:
		return 30
	}
}

// ToastMessage represents a single toast notification.
type ToastMessage struct {
	Text     string
	Level    ToastLevel
	Duration int // tick count before auto-dismiss
}

// ToastQueue manages a FIFO queue of toast messages with auto-dismiss.
// 保留历史记录供追溯查看。
type ToastQueue struct {
	messages []ToastMessage
	active   *ToastMessage
	timer    int
	maxQueue int
	history  []ToastMessage // 已消费的消息历史，最多 50 条
}

// NewToastQueue creates a toast queue with default settings.
func NewToastQueue() *ToastQueue {
	return &ToastQueue{maxQueue: 5}
}

// Push adds a message to the queue. If nothing is active, it shows immediately.
func (q *ToastQueue) Push(text string, level ToastLevel) {
	if q == nil {
		return
	}
	duration := level.defaultDuration()
	msg := ToastMessage{Text: text, Level: level, Duration: duration}
	if q.active == nil {
		q.active = &msg
		q.timer = msg.Duration
		return
	}
	if len(q.messages) < q.maxQueue {
		q.messages = append(q.messages, msg)
	}
}

// Tick decrements the active message timer. Called every 100ms.
// When timer reaches 0, move to history and show next queued message.
func (q *ToastQueue) Tick() {
	if q == nil || q.active == nil {
		return
	}
	q.timer--
	if q.timer <= 0 {
		// 归档到历史
		q.archiveActive()
		// 播放下一条
		if len(q.messages) > 0 {
			next := q.messages[0]
			q.messages = q.messages[1:]
			q.active = &next
			q.timer = next.Duration
		} else {
			q.active = nil
		}
	}
}

// archiveActive 将当前消息移至历史，最多保留 50 条。
func (q *ToastQueue) archiveActive() {
	if q == nil || q.active == nil {
		return
	}
	q.history = append(q.history, *q.active)
	if len(q.history) > 50 {
		q.history = q.history[len(q.history)-50:]
	}
}

// Active returns the currently displayed toast, or empty string if none.
func (q *ToastQueue) Active() string {
	if q == nil || q.active == nil {
		return ""
	}
	return q.active.Text
}

// ActiveLevel returns the level of the currently displayed toast.
func (q *ToastQueue) ActiveLevel() ToastLevel {
	if q == nil || q.active == nil {
		return ToastInfo
	}
	return q.active.Level
}

// History 返回最近 N 条已消费的 toast 消息（从旧到新）。
func (q *ToastQueue) History(n int) []ToastMessage {
	if q == nil || len(q.history) == 0 {
		return nil
	}
	if n <= 0 || n > len(q.history) {
		n = len(q.history)
	}
	return q.history[len(q.history)-n:]
}

// styleName maps ToastLevel to a GetStyle name.
func (l ToastLevel) styleName() StyleName {
	switch l {
	case ToastSuccess:
		return StyleToastSuccess
	case ToastError:
		return StyleToastError
	case ToastInfo:
		return StyleToastInfo
	case ToastWarning:
		return StyleToastWarning
	default:
		return StyleToastInfo
	}
}

// icon returns the prefix symbol for the toast level.
func (l ToastLevel) icon() string {
	switch l {
	case ToastSuccess:
		return "✓"
	case ToastError:
		return "✕"
	case ToastInfo:
		return "ℹ"
	case ToastWarning:
		return "⚠"
	default:
		return ""
	}
}

// RenderToast renders the active toast message with appropriate color and icon.
// Returns empty string if no toast is active.
func RenderToast(msg *ToastMessage) string {
	if msg == nil || msg.Text == "" {
		return ""
	}
	level := msg.Level
	return GetStyle(level.styleName()).Render(level.icon() + " " + msg.Text)
}
