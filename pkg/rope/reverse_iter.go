package rope

// ========== Reverse Iterator ==========

// ReverseIterator iterates over the rope's characters in reverse order.
type ReverseIterator struct {
	rope      *Rope
	position  int // Current character position (from end)
	totalLen  int
	exhausted bool
}

// NewReverseIterator creates a new reverse iterator.
func (r *Rope) NewReverseIterator() *ReverseIterator {
	if r == nil || r.Length() == 0 {
		return &ReverseIterator{
			rope:      r,
			position:  -1,
			totalLen:  0,
			exhausted: true,
		}
	}
	return &ReverseIterator{
		rope:      r,
		position:  -1, // Start before last character
		totalLen:  r.Length(),
		exhausted: false,
	}
}

// IterReverse creates a reverse iterator.
func (r *Rope) IterReverse() *ReverseIterator {
	return r.NewReverseIterator()
}

// CharsAtReverse creates a reverse iterator starting from the character at charIdx.
// The first call to Next() will position the iterator at charIdx (from start).
func (r *Rope) CharsAtReverse(charIdx int) *ReverseIterator {
	if r == nil || r.Length() == 0 {
		return &ReverseIterator{rope: r, exhausted: true}
	}

	if charIdx < 0 || charIdx > r.Length() {
		panic("character index out of bounds")
	}

	if charIdx == r.Length() {
		// Start from end
		return r.NewReverseIterator()
	}

	// Set position so that Next() moves to charIdx
	// We want: totalLen - 1 - (position + 1) = charIdx
	// Therefore: position = totalLen - 2 - charIdx
	return &ReverseIterator{
		rope:      r,
		position:  r.Length() - 2 - charIdx,
		totalLen:  r.Length(),
		exhausted: false,
	}
}

// Next advances to the previous character and returns true if there are more.
func (it *ReverseIterator) Next() bool {
	if it.exhausted {
		return false
	}

	it.position++
	if it.position >= it.totalLen {
		it.exhausted = true
		return false
	}

	return true
}

// Current returns the current character (from the end).
func (it *ReverseIterator) Current() rune {
	if it.position < 0 || it.position >= it.totalLen {
		panic("iterator out of bounds")
	}
	// Position from start = totalLen - 1 - position
	posFromStart := it.totalLen - 1 - it.position
	return it.rope.CharAt(posFromStart)
}

// Position returns the current position from the end.
// 0 means last character, 1 means second-to-last, etc.
func (it *ReverseIterator) Position() int {
	return it.position
}

// PositionFromStart returns the current position from the start of the rope.
func (it *ReverseIterator) PositionFromStart() int {
	if it.position < 0 || it.position >= it.totalLen {
		return -1
	}
	return it.totalLen - 1 - it.position
}

// HasNext returns true if there are more characters to iterate.
func (it *ReverseIterator) HasNext() bool {
	return !it.exhausted && (it.position+1) < it.totalLen
}

// Reset resets the iterator to the beginning (end of rope).
func (it *ReverseIterator) Reset() {
	it.position = -1
	it.exhausted = (it.rope == nil || it.rope.Length() == 0)
	it.totalLen = 0
	if it.rope != nil {
		it.totalLen = it.rope.Length()
	}
}

// IsExhausted returns true if the iterator has been exhausted.
func (it *ReverseIterator) IsExhausted() bool {
	return it.exhausted
}

// Peek returns the next character without advancing the iterator.
func (it *ReverseIterator) Peek() rune {
	if it.exhausted || !it.HasPeek() {
		panic("no next character")
	}
	nextPos := it.totalLen - 1 - (it.position + 1)
	return it.rope.CharAt(nextPos)
}

// HasPeek returns true if there is a next character to peek.
func (it *ReverseIterator) HasPeek() bool {
	return it.position+1 < it.totalLen
}

// Seek seeks to a specific position from the end.
// pos 0 = last character, pos 1 = second-to-last, etc.
func (it *ReverseIterator) Seek(pos int) bool {
	if pos < 0 || pos >= it.totalLen {
		it.exhausted = true
		return false
	}

	it.position = pos - 1 // Next() will move to pos
	it.exhausted = false
	return true
}

// SeekFromStart seeks to a specific position from the start of the rope.
// After calling Next(), the iterator will be at the character at pos from start.
func (it *ReverseIterator) SeekFromStart(pos int) bool {
	if pos < 0 || pos >= it.totalLen {
		it.exhausted = true
		return false
	}

	// Convert to reverse position, accounting for Next() increment
	// We want: totalLen - 1 - (position + 1) = pos
	// Therefore: position = totalLen - 2 - pos
	it.position = it.totalLen - 2 - pos
	it.exhausted = false
	return true
}

// Collect collects all characters in reverse order into a slice.
func (it *ReverseIterator) Collect() []rune {
	runes := make([]rune, 0, it.totalLen)
	it.Reset()
	for it.Next() {
		runes = append(runes, it.Current())
	}
	return runes
}

// ToSlice is an alias for Collect.
func (it *ReverseIterator) ToSlice() []rune {
	return it.Collect()
}

// ToRunes collects characters in reverse order.
func (it *ReverseIterator) ToRunes() []rune {
	return it.Collect()
}

// Skip skips n characters in reverse.
func (it *ReverseIterator) Skip(n int) bool {
	if n < 0 {
		return false
	}
	for i := 0; i < n && it.Next(); i++ {
	}
	return it.HasNext() || it.position < it.totalLen-1
}

// String returns the remaining characters as a string in reverse order.
func (it *ReverseIterator) String() string {
	runes := make([]rune, 0)
	for it.Next() {
		runes = append(runes, it.Current())
	}
	// Return in reverse order (as collected)
	return string(runes)
}

// ========== Reverse Operations ==========

// ForEachReverse applies a function to each character in reverse order.
func (r *Rope) ForEachReverse(fn func(rune) bool) bool {
	it := r.IterReverse()
	for it.Next() {
		if !fn(it.Current()) {
			return false
		}
	}
	return true
}

// ForEachReverseWithIndex applies a function to each character with its index in reverse order.
func (r *Rope) ForEachReverseWithIndex(fn func(int, rune) bool) bool {
	it := r.IterReverse()
	for it.Next() {
		if !fn(it.PositionFromStart(), it.Current()) {
			return false
		}
	}
	return true
}

// MapReverse maps each character through a function in reverse order.
// Returns a new Rope with the mapped characters (in original order).
func (r *Rope) MapReverse(fn func(rune) rune) *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	// Collect in reverse, then reverse back
	it := r.IterReverse()
	runes := it.Collect()

	// Reverse to get original order
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	// Apply mapping
	for i, r := range runes {
		runes[i] = fn(r)
	}

	b := NewBuilder()
	for _, r := range runes {
		b.AppendRune(r)
	}
	return b.Build()
}

// FilterReverse filters characters by a predicate in reverse order.
// Returns a new Rope with characters that satisfy the predicate (in original order).
func (r *Rope) FilterReverse(fn func(rune) bool) *Rope {
	if r == nil || r.Length() == 0 {
		return Empty()
	}

	// Collect in reverse
	it := r.IterReverse()
	runes := it.Collect()

	// Reverse to get original order
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	// Filter
	b := NewBuilder()
	for _, r := range runes {
		if fn(r) {
			b.AppendRune(r)
		}
	}
	return b.Build()
}

// FindReverse finds the last character that satisfies the predicate.
// Returns the character position and true if found, -1 and false otherwise.
func (r *Rope) FindReverse(fn func(rune) bool) (int, bool) {
	it := r.IterReverse()
	for it.Next() {
		if fn(it.Current()) {
			return it.PositionFromStart(), true
		}
	}
	return -1, false
}

// FindReverseFrom finds the last character before a given position that satisfies the predicate.
func (r *Rope) FindReverseFrom(beforePos int, fn func(rune) bool) (int, bool) {
	if beforePos <= 0 {
		return -1, false
	}

	it := r.IterReverse()
	it.SeekFromStart(beforePos - 1)
	it.Next() // Move to position beforePos, so subsequent Next() calls go backwards

	for it.Next() {
		if fn(it.Current()) {
			return it.PositionFromStart(), true
		}
	}
	return -1, false
}

// CountReverse counts characters that satisfy the predicate, iterating in reverse.
func (r *Rope) CountReverse(fn func(rune) bool) int {
	count := 0
	it := r.IterReverse()
	for it.Next() {
		if fn(it.Current()) {
			count++
		}
	}
	return count
}

// AllReverse checks if all characters satisfy the predicate (in reverse order).
func (r *Rope) AllReverse(fn func(rune) bool) bool {
	it := r.IterReverse()
	for it.Next() {
		if !fn(it.Current()) {
			return false
		}
	}
	return true
}

// AnyReverse checks if any character satisfies the predicate (in reverse order).
func (r *Rope) AnyReverse(fn func(rune) bool) bool {
	it := r.IterReverse()
	for it.Next() {
		if fn(it.Current()) {
			return true
		}
	}
	return false
}

// ========== Reverse Utilities ==========

// Reverse creates a new rope with characters in reverse order.
func (r *Rope) Reverse() *Rope {
	if r == nil || r.Length() <= 1 {
		return r
	}

	it := r.IterReverse()
	b := NewBuilder()
	runes := it.Collect()

	// runes are in reverse order, so append them as-is
	for _, r := range runes {
		b.AppendRune(r)
	}

	return b.Build()
}

// LastIndexOf finds the last position of a substring.
func (r *Rope) LastIndexOf(substring string) int {
	if substring == "" {
		return r.Length()
	}
	if len(substring) == 1 {
		return r.LastIndexOfChar([]rune(substring)[0])
	}

	// Simple approach: iterate in reverse and check
	substrRunes := []rune(substring)

	for i := len(substrRunes); i <= r.Length(); i++ {
		match := true
		for j := 0; j < len(substrRunes); j++ {
			// Check if substring matches at position (Length() - i)
			if r.CharAt(r.Length()-i+j) != substrRunes[j] {
				match = false
				break
			}
		}
		if match {
			return r.Length() - i
		}
	}

	return -1
}

// LastIndexOfAny finds the last position of any of the specified characters.
func (r *Rope) LastIndexOfAny(chars ...rune) int {
	if len(chars) == 0 {
		return -1
	}

	charSet := make(map[rune]bool)
	for _, ch := range chars {
		charSet[ch] = true
	}

	it := r.IterReverse()
	for it.Next() {
		if charSet[it.Current()] {
			return it.PositionFromStart()
		}
	}

	return -1
}

// TrimEnd removes trailing characters that satisfy the predicate.
func (r *Rope) TrimEnd(fn func(rune) bool) *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	it := r.IterReverse()
	end := r.Length()

	for it.Next() {
		if !fn(it.Current()) {
			break
		}
		end--
	}

	if end == r.Length() {
		return r
	}
	return New(r.Slice(0, end))
}

// TrimStart removes leading characters that satisfy the predicate.
// This is implemented with reverse iterator for demonstration.
func (r *Rope) TrimStart(fn func(rune) bool) *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	it := r.NewIterator()
	start := 0

	for it.Next() {
		if !fn(it.Current()) {
			break
		}
		start++
	}

	if start == 0 {
		return r
	}
	return New(r.Slice(start, r.Length()))
}

// Trim removes both leading and trailing characters that satisfy the predicate.
func (r *Rope) Trim(fn func(rune) bool) *Rope {
	return r.TrimStart(fn).TrimEnd(fn)
}
