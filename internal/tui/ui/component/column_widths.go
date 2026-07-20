package component

// column_widths.go — 列宽计算、响应式断点判断。
// 从 config.go 拆分，独立管理列百分比换算逻辑。

// ColPct 返回某页的默认百分比列宽。
func ColPct(page string) []int {
	p, ok := active.Cols[page]
	if !ok {
		return nil
	}
	return p.Pct
}

// ColMore 返回某页的中屏扩展列宽。
func ColMore(page string) []int {
	p, ok := active.Cols[page]
	if !ok {
		return nil
	}
	return p.More
}

// ColWide 返回某页的宽屏完整列宽。
func ColWide(page string) []int {
	p, ok := active.Cols[page]
	if !ok {
		return nil
	}
	return p.Wide
}

// ColCompact 返回某页的窄屏紧凑列宽。
func ColCompact(page string) []int {
	p, ok := active.Cols[page]
	if !ok {
		return nil
	}
	return p.Compact
}

// CalcColWidths 将百分比列宽转换为绝对字符宽度。
//
//	正数 = 百分比；0 = 填充剩余空间；-1 = 隐藏。
func CalcColWidths(pct []int, totalWidth int) []int {
	n := len(pct)
	if n == 0 {
		return nil
	}
	widths := make([]int, n)
	fixed := n - 1
	fillIdx := -1
	for i, v := range pct {
		if v == -1 {
			widths[i] = 0
			fixed--
		} else if v == 0 {
			fillIdx = i
		} else {
			widths[i] = totalWidth * v / 100
			fixed += widths[i]
		}
	}
	if fillIdx >= 0 {
		remain := totalWidth - fixed
		if remain < 4 {
			remain = 4
		}
		widths[fillIdx] = remain
	}
	return widths
}

// ShowBreak 返回某页某断点的最小宽度阈值。
func ShowBreak(page, name string) int {
	p, ok := active.Cols[page]
	if !ok || p.Show == nil {
		return 0
	}
	v, ok := p.Show[name]
	if !ok {
		return 0
	}
	return v
}
