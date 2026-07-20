package component

import (
	"image/color"
)

// StateColor 返回容器状态的调色板颜色。
func StateColor(state string) color.Color {
	switch state {
	case "running":
		return GetColor("green")
	case "stopping":
		return GetColor("yellow")
	case "stopped", "exited", "dead":
		return GetColor("gray")
	case "paused":
		return GetColor("yellow")
	case "created":
		return GetColor("blue")
	case "restarting":
		return GetColor("orange")
	default:
		return GetColor("gray")
	}
}

// RenderStateText 返回带状态色的文字（不含圆点）。
// 用于表格单元格，颜色从 table.jsonc 的 stateStyles 读取。
func RenderStateText(state string) string {
	ref := GetStateStyle(state)
	return buildStyle(ref).Render(state)
}
