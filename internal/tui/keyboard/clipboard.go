package keyboard

import (
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func clipboardCmd(text string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		for _, candidate := range []struct {
			name string
			args []string
		}{
			{name: "wl-copy"},
			{name: "xclip", args: []string{"-selection", "clipboard"}},
			{name: "pbcopy"},
			{name: "clip"},
		} {
			if _, err := exec.LookPath(candidate.name); err == nil {
				cmd = exec.Command(candidate.name, candidate.args...)
				break
			}
		}
		if cmd == nil {
			_ = os.WriteFile("/tmp/dtui-clipboard.txt", []byte(text), 0644) //nolint:errcheck // fallback clipboard is best-effort.
			return nil
		}
		cmd.Stdin = strings.NewReader(text)
		_ = cmd.Run() //nolint:errcheck // clipboard integration is best-effort.
		return nil
	}
}
