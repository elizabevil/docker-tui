package state

import "strings"

const MaxLogLines = 2000

type LogState struct {
	LogContainerID string
	LogContent     []string
	LogViewOffset  int
	LogSearchText  string
	LogSearchIdx   int
	LogSearchMatch int
	LogWrapEnabled bool
}

func (s *LogState) Open(containerID string) {
	s.LogContainerID = containerID
	s.LogContent = s.LogContent[:0]
	s.LogViewOffset = 0
	s.LogSearchText = ""
	s.LogSearchIdx = 0
	s.LogSearchMatch = 0
}

func (s *LogState) Close() {
	*s = LogState{}
}

func (s *LogState) Append(lines []string) {
	s.LogContent = append(s.LogContent, lines...)
	if len(s.LogContent) > MaxLogLines {
		s.LogContent = append([]string(nil), s.LogContent[len(s.LogContent)-MaxLogLines:]...)
	}
}

func (s *LogState) Scroll(delta int) {
	s.LogViewOffset = max(0, s.LogViewOffset+delta)
}

func (s *LogState) ToTop() { s.LogViewOffset = 0 }

func (s *LogState) ToBottom() { s.LogViewOffset = len(s.LogContent) }

func (s *LogState) ToggleWrap() { s.LogWrapEnabled = !s.LogWrapEnabled }

func (s *LogState) ApplySearch(query string) int {
	s.LogSearchText = query
	s.LogSearchMatch = 0
	count := s.MatchCount()
	if count > 0 {
		s.scrollToMatch()
	}
	return count
}

func (s *LogState) MoveMatch(delta int) {
	if s.LogSearchText == "" {
		return
	}
	s.LogSearchMatch = max(0, s.LogSearchMatch+delta)
	s.scrollToMatch()
}

func (s *LogState) MatchCount() int {
	if s.LogSearchText == "" {
		return 0
	}
	count := 0
	for _, line := range s.LogContent {
		if strings.Contains(strings.ToLower(line), strings.ToLower(s.LogSearchText)) {
			count++
		}
	}
	return count
}

func (s *LogState) scrollToMatch() {
	count := 0
	for i, line := range s.LogContent {
		if strings.Contains(strings.ToLower(line), strings.ToLower(s.LogSearchText)) {
			if count == s.LogSearchMatch {
				s.LogViewOffset = i
				return
			}
			count++
		}
	}
}

func (s *LogState) VisibleOffset(total, visible int) int {
	return min(max(0, s.LogViewOffset), max(0, total-visible))
}

// ComposeLogBatchReceived is the multi-source variant of LogBatchReceived
// for compose-aggregated log views (R08-06). It carries per-source
// stream errors so the Update loop can surface them without aborting
// the whole view.
type ComposeLogBatchReceived struct {
	ContainerID string
	Lines       []string
	StreamErrs  []LogStreamError
}
