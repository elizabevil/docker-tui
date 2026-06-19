package component

// TableProfileSelector 抽象“宽度 -> 渲染配置(profile)”的选择逻辑。
// 页面只声明规则，不再各自维护分支判断。
type TableProfileSelector interface {
	Select(width int) string
}

type ProfileRule struct {
	Name     string
	MinWidth int
}

// BreakpointProfileSelector 按规则顺序匹配 profile。
// 第一个满足 width >= MinWidth 的规则生效，否则返回 Default。
type BreakpointProfileSelector struct {
	Default string
	Rules   []ProfileRule
}

func (s BreakpointProfileSelector) Select(width int) string {
	for _, r := range s.Rules {
		if width >= r.MinWidth {
			return r.Name
		}
	}
	if s.Default == "" {
		return "default"
	}
	return s.Default
}

// ContainerProfileSelector 统一容器列表在 more/stats 场景下的 profile 选择。
type ContainerProfileSelector struct {
	Default       string
	DefaultStats  string
	More          string
	MoreStats     string
	MoreMinWidth  int
	StatsEnabled  bool
	StatsMinWidth int
}

func (s ContainerProfileSelector) Select(width int) string {
	showMore := width >= s.MoreMinWidth
	showStats := s.StatsEnabled && width >= s.StatsMinWidth
	if showMore && showStats {
		if s.MoreStats != "" {
			return s.MoreStats
		}
		return "more_stats"
	}
	if showMore {
		if s.More != "" {
			return s.More
		}
		return "more"
	}
	if showStats {
		if s.DefaultStats != "" {
			return s.DefaultStats
		}
		return "default_stats"
	}
	if s.Default == "" {
		return "default"
	}
	return s.Default
}
