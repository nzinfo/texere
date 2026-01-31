package rope

import (
	"sort"
	"unicode/utf8"
)

// ========== Micro-Optimizations for Peak Performance ==========

// InsertFast is the fastest insertion implementation with fast paths.
func (r *Rope) InsertFast(pos int, text string) *Rope {
	// Fast path 1: Empty text
	if text == "" {
		return r
	}

	// Fast path 2: Nil or empty rope
	if r == nil || r.length == 0 {
		return New(text)
	}

	// Fast path 3: Insert at beginning
	if pos == 0 {
		return r.Prepend(text)
	}

	// Fast path 4: Insert at end
	if pos == r.length {
		return r.Append(text)
	}

	// Fast path 5: Small rope with single leaf (avoid tree traversal)
	if r.root.IsLeaf() {
		return insertIntoSingleLeaf(r, pos, text)
	}

	// Standard path for complex cases
	return r.InsertOptimized(pos, text)
}

// DeleteFast is the fastest deletion implementation with fast paths.
func (r *Rope) DeleteFast(start, end int) *Rope {
	// Fast path 1: Nil or empty
	if r == nil || r.length == 0 {
		return r
	}

	// Fast path 2: Empty range
	if start == end {
		return r
	}

	// Fast path 3: Delete all
	if start == 0 && end == r.length {
		return Empty()
	}

	// Fast path 4: Small rope with single leaf
	if r.root.IsLeaf() {
		return deleteFromSingleLeaf(r, start, end)
	}

	// Fast path 5: Delete from beginning
	if start == 0 {
		return r.SliceToRope(end, r.length)
	}

	// Fast path 6: Delete from end
	if end == r.length {
		return r.SliceToRope(0, start)
	}

	// Standard path
	return r.DeleteOptimized(start, end)
}

// AppendFast and PrependFast removed - they were slower than standard Append/Prepend.
// Use Append() and Prepend() directly instead.

// SliceFast is the fastest slice implementation with optimizations.
func (r *Rope) SliceFast(start, end int) string {
	// Fast path 1: Full slice
	if start == 0 && end == r.length {
		return r.String()
	}

	// Fast path 2: Empty slice
	if start == end {
		return ""
	}

	// Fast path 3: Single leaf
	if r.root.IsLeaf() {
		leaf := r.root.(*LeafNode)
		return sliceSingleLeaf(leaf, start, end)
	}

	// Standard path
	return r.Slice(start, end)
}

// SliceToRope returns a slice as a new Rope.
func (r *Rope) SliceToRope(start, end int) *Rope {
	return New(r.SliceFast(start, end))
}

// ========== Single Leaf Optimizations ==========

// insertIntoSingleLeaf optimizes insertion into a single leaf.
// This avoids tree traversal overhead.
func insertIntoSingleLeaf(r *Rope, pos int, text string) *Rope {
	leaf := r.root.(*LeafNode)

	// Fast path for small insertions at string boundaries
	if pos == 0 {
		// Prepend to leaf
		newLeaf := AcquireLeaf()
		newLeaf.text = text + leaf.text
		return &Rope{
			root:   newLeaf,
			length: r.length + utf8.RuneCountInString(text),
			size:   r.size + len(text),
		}
	}

	if pos == r.length {
		// Append to leaf
		newLeaf := AcquireLeaf()
		newLeaf.text = leaf.text + text
		return &Rope{
			root:   newLeaf,
			length: r.length + utf8.RuneCountInString(text),
			size:   r.size + len(text),
		}
	}

	// Standard insertion in middle
	return r.InsertOptimized(pos, text)
}

// deleteFromSingleLeaf optimizes deletion from a single leaf.
func deleteFromSingleLeaf(r *Rope, start, end int) *Rope {
	leaf := r.root.(*LeafNode)

	// Fast path: Delete entire content
	if start == 0 && end == r.length {
		return Empty()
	}

	// Fast path: Delete from beginning
	if start == 0 {
		newLeaf := AcquireLeaf()
		// Find byte position
		endByte := findBytePosInString(leaf.text, end)
		newLeaf.text = leaf.text[endByte:]
		return &Rope{
			root:   newLeaf,
			length: r.length - utf8.RuneCountInString(leaf.text[:endByte]),
			size:   r.size - endByte,
		}
	}

	// Fast path: Delete from end
	if end == r.length {
		newLeaf := AcquireLeaf()
		// Find byte position
		startByte := findBytePosInString(leaf.text, start)
		newLeaf.text = leaf.text[:startByte]
		return &Rope{
			root:   newLeaf,
			length: start,
			size:   startByte,
		}
	}

	// Standard deletion
	return r.DeleteOptimized(start, end)
}

// sliceSingleLeaf optimizes slicing a single leaf.
func sliceSingleLeaf(leaf *LeafNode, start, end int) string {
	// Find byte positions
	startByte := findBytePosInString(leaf.text, start)
	endByte := findBytePosInString(leaf.text, end)
	return leaf.text[startByte:endByte]
}

// ========== Fast Byte Position Finding ==========

// findBytePosInString finds byte position without allocations.
// Optimized for ASCII (fast path) and UTF-8 (slow path).
func findBytePosInString(s string, charPos int) int {
	// Fast path: Try ASCII first (most common case)
	bytePos := 0
	asciiCount := 0

	for bytePos < len(s) && asciiCount < charPos {
		b := s[bytePos]
		if b < utf8.RuneSelf {
			// ASCII character
			bytePos++
			asciiCount++
		} else {
			// UTF-8 character, fall back to DecodeRune
			break
		}
	}

	// If we found all characters in ASCII mode, we're done
	if asciiCount == charPos {
		return bytePos
	}

	// Slow path: Handle UTF-8 characters
	remaining := charPos - asciiCount
	for i := 0; i < remaining; i++ {
		_, size := utf8.DecodeRuneInString(s[bytePos:])
		bytePos += size
	}

	return bytePos
}

// ========== Batch Operations ==========

// BatchInsert performs multiple insertions efficiently.
// Positions are relative to the original rope (not updated after each insertion).
func (r *Rope) BatchInsert(inserts []Insertion) *Rope {
	if len(inserts) == 0 {
		return r
	}

	// Fast path for single insertion
	if len(inserts) == 1 {
		return r.InsertFast(inserts[0].Pos, inserts[0].Text)
	}

	// Sort inserts by position (descending to avoid position recalculation)
	sortedInserts := make([]Insertion, len(inserts))
	copy(sortedInserts, inserts)

	// Use efficient sorting for larger arrays
	sort.Slice(sortedInserts, func(i, j int) bool {
		return sortedInserts[i].Pos > sortedInserts[j].Pos // Descending order
	})

	// Apply inserts from right to left (positions stay valid)
	result := r
	for _, ins := range sortedInserts {
		result = result.InsertFast(ins.Pos, ins.Text)
	}

	return result
}

// BatchDelete performs multiple deletions efficiently.
// Ranges are relative to the original rope.
func (r *Rope) BatchDelete(ranges []Range) *Rope {
	if len(ranges) == 0 {
		return r
	}

	// Fast path for single deletion
	if len(ranges) == 1 {
		return r.DeleteFast(ranges[0].From(), ranges[0].To())
	}

	// Sort ranges by start position (descending)
	sortedRanges := make([]Range, len(ranges))
	copy(sortedRanges, ranges)

	// Use efficient sorting for larger arrays
	sort.Slice(sortedRanges, func(i, j int) bool {
		return sortedRanges[i].From() > sortedRanges[j].From() // Descending order
	})

	// Apply deletions from right to left (positions stay valid)
	result := r
	for _, rng := range sortedRanges {
		result = result.DeleteFast(rng.From(), rng.To())
	}

	return result
}

// Insertion represents a single insertion operation.
type Insertion struct {
	Pos  int
	Text string
}

// Range represents a range [Start, End).
type ByteRange struct {
	Start int
	End   int
}

// ========== Inlined Helpers ==========

// IsASCII fast-checks if a string is pure ASCII.
func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

// ASCIIStringLength returns length if ASCII, -1 otherwise.
// Fast path for common ASCII strings.
func ASCIIStringLength(s string) int {
	if IsASCII(s) {
		return len(s)
	}
	return -1
}

// RuneCountInStringFast is optimized for ASCII strings.
func RuneCountInStringFast(s string) int {
	// Fast path: ASCII
	if IsASCII(s) {
		return len(s)
	}

	// Slow path: UTF-8
	return utf8.RuneCountInString(s)
}

// ========== Branch Prediction Hints ==========

// likely hints that a condition is likely true.
//go:nosplit
func likely(b bool) bool {
	return b
}

// unlikely hints that a condition is likely false.
//go:nosplit
func unlikely(b bool) bool {
	return b
}
