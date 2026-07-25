package component

// Spinner 提供终端进度指示器，使用字符轮转动画。
// 由 tea.Tick 驱动，每 tick 调用 Tick() 更新帧。
type Spinner struct {
	frames  []string
	current int
	active  bool
}

// NewSpinner 创建默认 spinner，帧序列为 |/-\ 循环。
func NewSpinner() *Spinner {
	return &Spinner{
		frames: []string{"|", "/", "-", "\\"},
	}
}

// NewDots 创建圆点等待 spinner：●○○ ○●○ ○○●
func NewDots() *Spinner {
	return &Spinner{
		frames: []string{"●○○", "○●○", "○○●"},
	}
}

// Tick 推进到下一帧。每 100ms 调用一次可获得流畅动画。
func (s *Spinner) Tick() {
	if s == nil || !s.active {
		return
	}
	s.current = (s.current + 1) % len(s.frames)
}

// String 返回当前帧字符。
func (s *Spinner) String() string {
	if s == nil || !s.active || len(s.frames) == 0 {
		return ""
	}
	return s.frames[s.current]
}

// Start 激活 spinner。
func (s *Spinner) Start() {
	if s == nil {
		return
	}
	s.active = true
	s.current = 0
}

// Stop 停用 spinner，重置到初始帧。
func (s *Spinner) Stop() {
	if s == nil {
		return
	}
	s.active = false
	s.current = 0
}

// Active 返回 spinner 是否在运行。
func (s *Spinner) Active() bool {
	return s != nil && s.active
}

// Render 返回带样式的 spinner 文字。
// 非激活时返回空字符串。
func (s *Spinner) Render() string {
	if s == nil || !s.active {
		return ""
	}
	return GetStyle("toastInfo").Render(" " + s.String())
}
