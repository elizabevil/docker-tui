package component

import (
	"encoding/binary"
	"hash/fnv"
	"sync"

	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/utils"
)

const defaultTableLayoutCacheSize = 256

type tableLayout struct {
	headers       []string
	widths        []int
	gap           int
	contentWidth  int
	trailingWidth int
}

// TableLayoutCache caches viewport-level header and column width calculations.
// The key includes every input that can affect layout, so resize, locale,
// profile, schema, sorting, and visible row changes invalidate naturally.
type TableLayoutCache struct {
	mu      sync.RWMutex
	entries map[uint64]tableLayout
	maxSize int
	hits    uint64
	misses  uint64
}

func NewTableLayoutCache(maxSize int) *TableLayoutCache {
	if maxSize <= 0 {
		maxSize = defaultTableLayoutCacheSize
	}
	return &TableLayoutCache{entries: make(map[uint64]tableLayout), maxSize: maxSize}
}

func (c *TableLayoutCache) resolve(data TableData, headers []string, containerWidth, gap int) tableLayout {
	desiredWidths := tableContentWidths(data, headers)
	key := tableLayoutKey(data, headers, desiredWidths, containerWidth, gap)
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if ok {
		c.mu.Lock()
		c.hits++
		c.mu.Unlock()
		return cloneTableLayout(entry)
	}

	resolved := tables.ResolveContentLayout(data.Cols, desiredWidths, containerWidth, gap)
	entry = tableLayout{
		headers: append([]string(nil), headers...),
		widths:  resolved.Widths, gap: resolved.Gap,
		contentWidth: resolved.ContentWidth, trailingWidth: resolved.TrailingWidth,
	}
	c.mu.Lock()
	if len(c.entries) >= c.maxSize {
		clear(c.entries)
	}
	c.entries[key] = cloneTableLayout(entry)
	c.misses++
	c.mu.Unlock()
	return entry
}

func cloneTableLayout(layout tableLayout) tableLayout {
	return tableLayout{
		headers: append([]string(nil), layout.headers...),
		widths:  append([]int(nil), layout.widths...), gap: layout.gap,
		contentWidth: layout.contentWidth, trailingWidth: layout.trailingWidth,
	}
}

func tableLayoutKey(data TableData, headers []string, desiredWidths []int, containerWidth, gap int) uint64 {
	hash := fnv.New64a()
	writeString := func(value string) {
		_, _ = hash.Write([]byte(value))
		_, _ = hash.Write([]byte{0})
	}
	writeInt := func(value int) {
		var buffer [8]byte
		binary.LittleEndian.PutUint64(buffer[:], uint64(value))
		_, _ = hash.Write(buffer[:])
	}

	writeInt(containerWidth)
	writeInt(gap)
	writeString(data.SortColKey)
	if data.SortAsc {
		writeInt(1)
	}
	for _, column := range data.Cols {
		writeString(column.Key)
		writeString(column.Header)
		writeInt(column.Fixed)
		writeInt(column.Basis)
		writeInt(column.Min)
		writeInt(column.Max)
		writeInt(column.Shrink)
		writeInt(column.Fill)
	}
	for _, header := range headers {
		writeString(header)
	}
	for _, width := range desiredWidths {
		writeInt(width)
	}
	return hash.Sum64()
}

func tableContentWidths(data TableData, headers []string) []int {
	widths := make([]int, len(data.Cols))
	for i := range widths {
		if i < len(headers) {
			widths[i] = utils.VisibleLen(headers[i])
		}
	}
	for _, row := range data.Rows {
		for i, cell := range row {
			if i < len(widths) {
				widths[i] = max(widths[i], utils.VisibleLen(cell))
			}
		}
	}
	return widths
}

var defaultTableLayoutCache = NewTableLayoutCache(defaultTableLayoutCacheSize)
