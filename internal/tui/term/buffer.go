// Package term provides a simple virtual terminal buffer for rendering
// Docker exec TTY output. It handles carriage-return line overwrites,
// backspace, newlines, and ANSI escape stripping, maintaining a
// scrollback buffer of output lines.
package term

import (
	"strconv"
	"strings"
	"unicode"
)

// Buffer is a scrollable terminal output buffer.
// It tracks a virtual cursor and accumulates lines from raw TTY data.
type Buffer struct {
	lines          []string
	maxLines       int  // scrollback limit (rows)
	row            int  // current write row index
	col            int  // current write column
	lastWasNewline bool // true if the last processed char was \n (collapses consecutive newlines)
}

// NewBuffer creates a terminal buffer with the given scrollback capacity.
func NewBuffer(maxLines int) *Buffer {
	return &Buffer{
		lines:    make([]string, 0, 128),
		maxLines: maxLines,
	}
}

// Write ingests raw TTY data, updating the line buffer.
// It handles:
//   - \r (carriage return → overwrite current line from col 0)
//   - \n (line feed → advance to next row)
//   - \b (backspace → delete before cursor)
//   - CSI sequences for cursor movement, erase, insert, delete
//   - Strip other ANSI escape sequences
func (b *Buffer) Write(data string) {
	for len(data) > 0 {
		if data[0] == '\x1b' {
			consumed := b.parseANSI(data)
			data = data[consumed:]
			continue
		}

		printable := b.extractPrintable(data)
		if len(printable) > 0 {
			b.writeString(printable)
			data = data[len(printable):]
			continue
		}

		r := rune(data[0])
		data = data[1:]

		switch {
		case r == '\r':
			b.col = 0
			b.lastWasNewline = false
		case r == '\n':
			if !b.lastWasNewline {
				b.lastWasNewline = true
				b.col = 0
				b.row++
			}
		case r == '\b' || r == 0x7f:
			b.handleBackspace()
		case r < 0x20:
			// Skip other control characters
		}
	}
}

func (b *Buffer) extractPrintable(data string) string {
	end := 0
	for end < len(data) {
		c := data[end]
		if c == '\x1b' || c == '\r' || c == '\n' || c == '\b' || c == 0x7f || c < 0x20 {
			break
		}
		end++
	}
	return data[:end]
}

func (b *Buffer) writeString(s string) {
	if len(s) == 0 {
		return
	}
	b.lastWasNewline = false
	b.ensureRow()

	runes := []rune(b.lines[b.row])
	sRunes := []rune(s)

	if b.col >= len(runes) {
		padding := make([]rune, b.col-len(runes)+len(sRunes))
		for i := range padding {
			padding[i] = ' '
		}
		runes = append(runes, padding...)
	} else {
		insertRunes := make([]rune, len(runes)+len(sRunes))
		copy(insertRunes, runes[:b.col])
		copy(insertRunes[b.col:], sRunes)
		copy(insertRunes[b.col+len(sRunes):], runes[b.col:])
		runes = insertRunes
	}
	b.lines[b.row] = string(runes)
	b.col += len(sRunes)
}

func (b *Buffer) handleBackspace() {
	b.lastWasNewline = false
	if b.col > 0 {
		b.col--
		line := []rune(b.currentLine())
		if b.col < len(line) {
			b.setLine(string(line[:b.col]) + string(line[b.col+1:]))
		}
	}
}

// parseANSI consumes and interprets an ANSI escape sequence from data.
// Returns the number of bytes consumed.
func (b *Buffer) parseANSI(data string) int {
	if len(data) < 2 {
		return 1
	}

	// Two-character sequences: \x1b + letter
	if data[1] >= 0x40 && data[1] <= 0x5F {
		switch data[1] {
		case 'D': // \x1bD - Index (scroll down)
			b.col = 0
			b.row++
			return 2
		case 'M': // \x1bM - Reverse index (scroll up)
			if b.row > 0 {
				b.row--
			}
			return 2
		case 'H': // \x1bH - Set tab stop (ignore)
			return 2
		case '7': // \x1b7 - Save cursor (ignore)
			return 2
		case '8': // \x1b8 - Restore cursor (ignore)
			return 2
		}
	}

	// CSI sequences: \x1b[ ... <letter>
	if len(data) >= 3 && data[1] == '[' {
		end := 2
		for end < len(data) {
			c := data[end]
			if (c >= 0x40 && c <= 0x7E) || c == 0x1b {
				break
			}
			end++
		}
		if end >= len(data) {
			return len(data) // Incomplete, consume everything
		}
		term := data[end]
		paramsStr := data[2:end]
		consumed := end + 1

		if term >= 0x40 && term <= 0x7E {
			b.handleCSI(paramsStr, term)
		}
		return consumed
	}

	// Unrecognized: strip it
	// Find the end (next letter)
	end := 1
	for end < len(data) {
		c := data[end]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			end++
			break
		}
		if c == '\x1b' {
			break
		}
		end++
	}
	if end > len(data) {
		end = len(data)
	}
	return end
}

// handleCSI processes a parsed CSI sequence.
func (b *Buffer) handleCSI(params string, term byte) {
	args := parseCSIArgs(params)

	switch term {
	case 'A': // CUU - Cursor Up
		n := argOr(args, 0, 1)
		b.row -= n
		if b.row < 0 {
			b.row = 0
		}

	case 'B': // CUD - Cursor Down
		n := argOr(args, 0, 1)
		b.row += n

	case 'C': // CUF - Cursor Forward
		n := argOr(args, 0, 1)
		b.col += n

	case 'D': // CUB - Cursor Back
		n := argOr(args, 0, 1)
		b.col -= n
		if b.col < 0 {
			b.col = 0
		}

	case 'G': // CHA - Cursor Horizontal Absolute
		n := argOr(args, 0, 1)
		b.col = max(n-1, 0)

	case 'H', 'f': // CUP / HVP - Cursor Position
		row := argOr(args, 0, 1) - 1
		col := argOr(args, 1, 1) - 1
		if row < 0 {
			row = 0
		}
		if col < 0 {
			col = 0
		}
		b.row = row
		b.col = col

	case 'J': // ED - Erase in Display
		n := argOr(args, 0, 0)
		switch n {
		case 0: // Erase below
			line := b.currentLine()
			if b.col < len([]rune(line)) {
				b.setLine(string([]rune(line)[:b.col]))
			}
		case 1: // Erase above (ignore)
		case 2: // Erase entire screen
			b.lines = b.lines[:0]
			b.row = 0
			b.col = 0
		}

	case 'K': // EL - Erase in Line
		n := argOr(args, 0, 0)
		line := b.currentLine()
		switch n {
		case 0: // Erase to end of line
			if b.col < len([]rune(line)) {
				b.setLine(string([]rune(line)[:b.col]))
			}
		case 1: // Erase from beginning of line
			if b.col <= len([]rune(line)) {
				padding := make([]rune, b.col)
				for i := range padding {
					padding[i] = ' '
				}
				b.setLine(string(padding) + string([]rune(line)[b.col:]))
			}
		case 2: // Erase entire line
			b.setLine("")
			b.col = 0
		}

	case 'L': // IL - Insert Line
		b.row++

	case 'M': // DL - Delete Line
		if b.row < len(b.lines) {
			b.lines = append(b.lines[:b.row], b.lines[b.row+1:]...)
		}

	case 'P': // DCH - Delete Character
		n := argOr(args, 0, 1)
		runes := []rune(b.currentLine())
		if b.col < len(runes) {
			end := min(b.col+n, len(runes))
			b.setLine(string(runes[:b.col]) + string(runes[end:]))
		}

	case '@': // ICH - Insert Character
		n := argOr(args, 0, 1)
		runes := []rune(b.currentLine())
		spaces := make([]rune, n)
		for i := range spaces {
			spaces[i] = ' '
		}
		if b.col <= len(runes) {
			b.setLine(string(runes[:b.col]) + string(spaces) + string(runes[b.col:]))
		}

	case 'd': // VPA - Vertical Position Absolute
		n := max(argOr(args, 0, 1)-1, 0)
		b.row = n

	case 's': // Save cursor (ignore)

	case 'u': // Restore cursor (ignore)

	case 'S': // SU - Scroll Up
		n := argOr(args, 0, 1)
		b.row += n

	case 'T': // SD - Scroll Down
		n := argOr(args, 0, 1)
		if b.row >= n {
			b.row -= n
		} else {
			b.row = 0
		}
	}
}

// ensureRow ensures the current row exists in the buffer, adding lines
// with scrollback as needed.
func (b *Buffer) ensureRow() {
	for b.row >= len(b.lines) {
		b.lines = append(b.lines, "")
		if len(b.lines) > b.maxLines {
			b.lines = b.lines[1:]
			if b.row > 0 {
				b.row--
			}
		}
	}
}

// parseCSIArgs parses the parameter string of a CSI sequence.
// Returns a slice of ints. Missing/default values are 0.
func parseCSIArgs(params string) []int {
	if params == "" {
		return nil
	}
	parts := strings.Split(params, ";")
	args := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			n = 0
		}
		args = append(args, n)
	}
	return args
}

// argOr returns the i-th argument, or def if not available.
func argOr(args []int, i, def int) int {
	if i < len(args) && args[i] > 0 {
		return args[i]
	}
	return def
}

// Len returns the number of lines in the buffer.
func (b *Buffer) Len() int {
	return len(b.lines)
}

// Lines returns a copy of all lines.
func (b *Buffer) Lines() []string {
	out := make([]string, len(b.lines))
	copy(out, b.lines)
	return out
}

// View returns up to `height` lines from the buffer, showing the
// most recent content. If total lines < height, the result is padded
// with empty strings at the top. Scroll offset > 0 shows earlier content.
func (b *Buffer) View(height, scroll int) []string {
	total := len(b.lines)
	if total == 0 {
		out := make([]string, height)
		for i := range out {
			out[i] = "~"
		}
		return out
	}

	if scroll < 0 {
		scroll = 0
	}
	maxScroll := max(total-height, 0)
	if scroll > maxScroll {
		scroll = maxScroll
	}

	start := max(total-height-scroll, 0)
	end := min(start+height, total)

	out := make([]string, height)
	padTop := height - (end - start)
	for i := range padTop {
		out[i] = ""
	}
	for i := start; i < end; i++ {
		line := cleanLine(b.lines[i])
		out[i-start+padTop] = line
	}
	return out
}

func (b *Buffer) CursorCol() int { return b.col }

func (b *Buffer) CursorRow() int { return b.row }

func (b *Buffer) CursorVisible(height, scroll int) (visRow, visCol int) {
	total := len(b.lines)
	if total == 0 || b.row < 0 {
		return -1, -1
	}
	// The cursor can be on a new empty line beyond the last content (b.row == total).
	// Clamp it to the last line for visibility.
	cursorRow := b.row
	if cursorRow >= total {
		cursorRow = total - 1
	}
	maxScroll := max(total-height, 0)
	if scroll > maxScroll {
		scroll = maxScroll
	}
	start := max(total-height-scroll, 0)
	if cursorRow < start || cursorRow >= start+height {
		return -1, -1
	}
	return cursorRow - start, max(b.col, 0)
}

// Reset clears the buffer.
func (b *Buffer) Reset() {
	b.lines = b.lines[:0]
	b.row = 0
	b.col = 0
}

func (b *Buffer) currentLine() string {
	if b.row >= 0 && b.row < len(b.lines) {
		return b.lines[b.row]
	}
	return ""
}

func (b *Buffer) setLine(line string) {
	if b.row >= 0 && b.row < len(b.lines) {
		b.lines[b.row] = line
	}
}

// cleanLine replaces non-printable characters with spaces.
func cleanLine(line string) string {
	runes := []rune(line)
	for i, r := range runes {
		if !unicode.IsPrint(r) && r != '\t' {
			runes[i] = ' '
		}
	}
	return string(runes)
}
