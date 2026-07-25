package component

import (
	"encoding/binary"
	"hash/fnv"
	"sync"
)

const defaultTableLayoutCacheSize = 256

type tableLayout struct {
	headers []string
	widths  []int
	gap     int
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

func (c *TableLayoutCache) resolve(data TableData, headers []string, containerWidth int) tableLayout {
	key := tableLayoutKey(data, headers, containerWidth)
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if ok {
		c.mu.Lock()
		c.hits++
		c.mu.Unlock()
		return cloneTableLayout(entry)
	}

	widths := computeContentWidths(data, headers)
	entry = tableLayout{headers: append([]string(nil), headers...), widths: widths, gap: computeGap(widths, containerWidth)}
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
		widths:  append([]int(nil), layout.widths...),
		gap:     layout.gap,
	}
}

func tableLayoutKey(data TableData, headers []string, containerWidth int) uint64 {
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
	writeString(data.SortColKey)
	if data.SortAsc {
		writeInt(1)
	}
	for index, column := range data.Cols {
		writeString(column.Key)
		writeString(column.Header)
		writeInt(column.Fixed)
		writeInt(column.Flex)
		writeInt(column.Max)
		if index < len(data.Widths) {
			writeInt(data.Widths[index])
		}
	}
	for _, header := range headers {
		writeString(header)
	}
	for _, row := range data.Rows {
		writeInt(len(row))
		for _, cell := range row {
			writeString(cell)
		}
	}
	return hash.Sum64()
}

var defaultTableLayoutCache = NewTableLayoutCache(defaultTableLayoutCacheSize)
