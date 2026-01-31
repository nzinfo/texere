package rope

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/graphemes"
)

// ========== Grapheme Support ==========

// Grapheme represents a user-perceived character (grapheme cluster).
// A grapheme can be:
// - A single ASCII character (e.g., 'a')
// - A single Unicode code point (e.g., 'é' as U+00E9)
// - Multiple code points (e.g., 'e' + combining acute, emoji families, etc.)
type Grapheme struct {
	Text      string  // The grapheme cluster text
	StartPos  int     // Character position in rope (where this grapheme starts)
	byteLen   int     // Length in bytes (private)
	CharLen   int     // Length in characters (code points)
}

// GraphemeIterator iterates over grapheme clusters in a rope.
type GraphemeIterator struct {
	rope       *Rope
	graphemes  []Grapheme
	index      int
	exhausted  bool
}

// Graphemes returns an iterator over grapheme clusters in the rope.
// This is essential for proper Unicode handling in text editors.
func (r *Rope) Graphemes() *GraphemeIterator {
	if r == nil || r.Length() == 0 {
		return &GraphemeIterator{rope: r, exhausted: true}
	}

	content := r.String()
	segments := graphemes.SegmentAllString(content)

	graphemes := make([]Grapheme, len(segments))
	charPos := 0
	for i, seg := range segments {
		byteLen := len(seg)
		charLen := utf8.RuneCountInString(seg)
		graphemes[i] = Grapheme{
			Text:     seg,
			StartPos: charPos,
			byteLen:  byteLen,
			CharLen:  charLen,
		}
		charPos += charLen
	}

	return &GraphemeIterator{
		rope:      r,
		graphemes: graphemes,
		index:     -1,
		exhausted: len(graphemes) == 0,
	}
}

// Next advances to the next grapheme cluster and returns true if there are more.
func (it *GraphemeIterator) Next() bool {
	if it.exhausted {
		return false
	}

	it.index++
	if it.index >= len(it.graphemes) {
		it.exhausted = true
		return false
	}

	return true
}

// Current returns the current grapheme cluster.
func (it *GraphemeIterator) Current() Grapheme {
	if it.exhausted || it.index < 0 || it.index >= len(it.graphemes) {
		return Grapheme{}
	}

	return it.graphemes[it.index]
}

// Position returns the character position of the current grapheme.
func (it *GraphemeIterator) Position() int {
	if it.exhausted {
		return it.rope.LenGraphemes()
	}
	return it.Current().StartPos
}

// Reset resets the iterator to the beginning of the rope.
func (it *GraphemeIterator) Reset() {
	if it.rope == nil || it.rope.Length() == 0 {
		it.exhausted = true
		return
	}

	newIt := it.rope.Graphemes()
	it.graphemes = newIt.graphemes
	it.index = -1
	it.exhausted = len(it.graphemes) == 0
}

// Collect collects all graphemes into a slice.
func (it *GraphemeIterator) Collect() []Grapheme {
	var graphemes []Grapheme
	for it.Next() {
		graphemes = append(graphemes, it.Current())
	}
	return graphemes
}

// ToSlice is an alias for Collect.
func (it *GraphemeIterator) ToSlice() []Grapheme {
	return it.Collect()
}

// HasNext returns true if there are more graphemes to iterate.
func (it *GraphemeIterator) HasNext() bool {
	if it.exhausted {
		return false
	}
	return it.index+1 < len(it.graphemes)
}

// LenGraphemes returns the total number of grapheme clusters in the rope.
// This is O(n) where n is the byte length of the rope.
func (r *Rope) LenGraphemes() int {
	if r == nil || r.Length() == 0 {
		return 0
	}

	count := 0
	it := r.Graphemes()
	for it.Next() {
		count++
	}
	return count
}

// GraphemeAt returns the grapheme at the given character position.
// Panics if position is out of bounds.
func (r *Rope) GraphemeAt(charIdx int) Grapheme {
	if charIdx < 0 || charIdx >= r.LenGraphemes() {
		panic("grapheme index out of bounds")
	}

	it := r.Graphemes()
	for i := 0; i <= charIdx; i++ {
		it.Next()
	}

	return it.Current()
}

// PrevGraphemeStart returns the character position of the start
// of the grapheme cluster containing the given position.
// Panics if position is out of bounds.
func (r *Rope) PrevGraphemeStart(charIdx int) int {
	if charIdx < 0 || charIdx > r.LenGraphemes() {
		panic("character position out of bounds")
	}

	if charIdx == 0 {
		return 0
	}

	// Scan forward to find grapheme containing charIdx
	currentGraphemeStart := 0

	it := r.Graphemes()
	for it.Next() {
		g := it.Current()
		nextPos := g.StartPos + g.CharLen

		if charIdx < nextPos {
			// Found it - charIdx is in current grapheme
			return currentGraphemeStart
		}

		currentGraphemeStart = nextPos
	}

	// Should not reach here
	return currentGraphemeStart
}

// NextGraphemeStart returns the character position of the start
// of the grapheme cluster after the given position.
// Returns the position past the end if position is at or past the last grapheme.
func (r *Rope) NextGraphemeStart(charIdx int) int {
	if charIdx < 0 || charIdx >= r.Length() {
		return r.Length()
	}

	// Scan forward to find grapheme after charIdx
	it := r.Graphemes()
	for it.Next() {
		g := it.Current()
		nextPos := g.StartPos + g.CharLen

		if charIdx < g.StartPos {
			// We've passed it
			return g.StartPos
		}

		if charIdx < nextPos {
			// charIdx is in this grapheme, return next
			return nextPos
		}
	}

	// At or past end
	return r.Length()
}

// IsGraphemeBoundary returns true if the given position is at
// a grapheme cluster boundary.
func (r *Rope) IsGraphemeBoundary(charIdx int) bool {
	if charIdx < 0 || charIdx > r.Length() {
		return false
	}

	if charIdx == 0 {
		return true
	}

	if charIdx == r.Length() {
		return true
	}

	// Check if position is at the start of a grapheme
	it := r.Graphemes()
	for it.Next() {
		g := it.Current()
		if g.StartPos == charIdx {
			return true
		}
		if g.StartPos > charIdx {
			return false
		}
	}

	return false
}

// GraphemeSlice returns a new rope containing graphemes from start to end (in grapheme indices).
// Panics if indices are out of bounds.
func (r *Rope) GraphemeSlice(start, end int) *Rope {
	if start < 0 || end > r.LenGraphemes() || start > end {
		panic("grapheme slice indices out of bounds")
	}

	it := r.Graphemes()
	builder := NewBuilder()

	currentGrapheme := 0
	for it.Next() {
		if currentGrapheme >= start && currentGrapheme < end {
			builder.Append(it.Current().Text)
		}
		currentGrapheme++
	}

	return builder.Build()
}

// ========== Helper ==========

// String returns a string representation of the grapheme.
func (g Grapheme) String() string {
	return g.Text
}

// Bytes returns the grapheme as a byte slice.
func (g Grapheme) Bytes() []byte {
	return []byte(g.Text)
}

// Runes returns the grapheme as a rune slice.
func (g Grapheme) Runes() []rune {
	return []rune(g.Text)
}

// Len returns the character length of the grapheme.
func (g Grapheme) Len() int {
	return g.CharLen
}

// ByteLen returns the byte length of the grapheme.
func (g Grapheme) ByteLen() int {
	return g.byteLen
}

// IsSingleRune returns true if the grapheme is a single rune.
func (g Grapheme) IsSingleRune() bool {
	return g.CharLen == 1
}

// IsASCII returns true if the grapheme contains only ASCII characters.
func (g Grapheme) IsASCII() bool {
	for _, r := range g.Runes() {
		if r > 127 {
			return false
		}
	}
	return true
}

// ========== Rope Methods ==========

// ForEachGrapheme calls the function for each grapheme in the rope.
func (r *Rope) ForEachGrapheme(f func(Grapheme)) {
	if r == nil || r.Length() == 0 {
		return
	}

	it := r.Graphemes()
	for it.Next() {
		f(it.Current())
	}
}

// MapGraphemes creates a new rope by applying the function to each grapheme.
func (r *Rope) MapGraphemes(f func(Grapheme) string) *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	builder := NewBuilder()
	it := r.Graphemes()
	for it.Next() {
		builder.Append(f(it.Current()))
	}

	return builder.Build()
}

// FilterGraphemes creates a new rope with graphemes that satisfy the predicate.
func (r *Rope) FilterGraphemes(pred func(Grapheme) bool) *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	builder := NewBuilder()
	it := r.Graphemes()
	for it.Next() {
		g := it.Current()
		if pred(g) {
			builder.Append(g.Text)
		}
	}

	return builder.Build()
}

// ContainsGrapheme returns true if the rope contains the given grapheme text.
func (r *Rope) ContainsGrapheme(text string) bool {
	if r == nil || r.Length() == 0 {
		return false
	}

	it := r.Graphemes()
	for it.Next() {
		if it.Current().Text == text {
			return true
		}
	}
	return false
}

// IndexGrapheme returns the index of the first occurrence of the grapheme,
// or -1 if not found.
func (r *Rope) IndexGrapheme(text string) int {
	if r == nil || r.Length() == 0 {
		return -1
	}

	it := r.Graphemes()
	idx := 0
	for it.Next() {
		if it.Current().Text == text {
			return idx
		}
		idx++
	}
	return -1
}

// CountGrapheme returns the number of occurrences of the given grapheme text.
func (r *Rope) CountGrapheme(text string) int {
	if r == nil || r.Length() == 0 {
		return 0
	}

	count := 0
	it := r.Graphemes()
	for it.Next() {
		if it.Current().Text == text {
			count++
		}
	}
	return count
}

// ========== Duration Parsing ==========

// ParseDuration parses human-friendly duration strings.
// Supported formats:
//   - "30s", "30sec", "30 seconds" -> 30 seconds
//   - "5m", "5min", "5 minutes" -> 5 minutes
//   - "2h", "2hour", "2 hours" -> 2 hours
//   - "1d", "1day", "1 days" -> 24 hours
//   - "60" -> 60 seconds (default is seconds)
func ParseDuration(s string) (time.Duration, error) {
	// Normalize: lowercase, trim spaces
	s = strings.ToLower(strings.TrimSpace(s))

	// Match: number + optional unit
	if len(s) == 0 {
		return 0, fmt.Errorf("empty duration")
	}

	// Extract number
	numEnd := 0
	for numEnd < len(s) && (s[numEnd] >= '0' && s[numEnd] <= '9') {
		numEnd++
	}

	if numEnd == 0 {
		return 0, fmt.Errorf("no number found in duration: %s", s)
	}

	value, err := strconv.Atoi(s[:numEnd])
	if err != nil {
		return 0, err
	}

	// Extract unit
	unit := strings.TrimSpace(s[numEnd:])

	switch unit {
	case "", "s", "sec", "second", "seconds":
		return time.Duration(value) * time.Second, nil
	case "m", "min", "minute", "minutes":
		return time.Duration(value) * time.Minute, nil
	case "h", "hour", "hours":
		return time.Duration(value) * time.Hour, nil
	case "d", "day", "days":
		return time.Duration(value) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown time unit: %s in duration: %s", unit, s)
	}
}

// FormatDuration formats a duration as a human-readable string.
// Examples: "5m", "30s", "2h"
func FormatDuration(d time.Duration) string {
	d = d.Truncate(time.Second)

	if d < time.Minute {
		return fmt.Sprintf("%ds", d/time.Second)
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", d/time.Minute)
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", d/time.Hour)
	} else {
		return fmt.Sprintf("%dd", d/(24*time.Hour))
	}
}
