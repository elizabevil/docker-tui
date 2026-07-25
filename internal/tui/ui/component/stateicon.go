package component

import (
	"image/color"

	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// StateColor 返回容器状态的调色板颜色。
func StateColor(containerState string) color.Color {
	switch containerState {
	case state.ContainerStateRunning:
		return GetColor("green")
	case state.ContainerStateStopping:
		return GetColor("yellow")
	case state.ContainerStateStopped, state.ContainerStateExited, state.ContainerStateDead:
		return GetColor("gray")
	case state.ContainerStatePaused:
		return GetColor("yellow")
	case state.ContainerStateCreated:
		return GetColor("blue")
	case state.ContainerStateRestarting:
		return GetColor("orange")
	default:
		return GetColor("gray")
	}
}

// RenderStateText 返回带状态色的文字（不含圆点）。
// 用于表格单元格，颜色从 table.jsonc 的 stateStyles 读取。
func RenderStateText(containerState string) string {
	ref := GetStateStyle(containerState)
	return buildStyle(ref).Render(containerState)
}
