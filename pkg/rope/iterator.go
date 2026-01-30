package rope

import "unicode/utf8"

// Iterator provides efficient forward and backward iteration through a Rope.
//
// The iterator maintains a position in the rope and allows seeking,
// which makes it efficient for operations that need to access multiple
// positions sequentially.
//
// Example usage:
//
//	r := rope.New("Hello World")
//	it := r.NewIterator()
//
//	for it.Next() {
//		ch := it.Current()
//		// Process character
//	}
type Iterator struct {
	rope      *Rope
	position  int    // Current character position
	runePos   int    // Position within the current leaf (-1 = not initialized)
	bytePos   int    // Byte position within current leaf (avoids []rune conversion)
	stack     []frame // Stack for tree traversal
	current   string  // Current leaf text
	exhausted bool    // True when iteration is complete
}

// frame represents a position in the tree traversal stack.
type frame struct {
	node     RopeNode
	position int // Position within this node
}

// NewIterator creates a new iterator starting at the beginning of the rope.
// The iterator starts positioned before the first character.
// Call Next() to advance to the first character.
func (r *Rope) NewIterator() *Iterator {
	return &Iterator{
		rope:      r,
		position:  0,
		runePos:   -1, // Not initialized - Next() will initialize
		stack:     make([]frame, 0, 16),
		exhausted: (r == nil || r.Length() == 0),
	}
}

// IteratorAt creates a new iterator starting at the given character position.
// The iterator is positioned at `pos`, so Current() will return the character at that position.
// Call Next() to advance to the next character.
func (r *Rope) IteratorAt(pos int) *Iterator {
	it := &Iterator{
		rope:      r,
		position:  pos,
		stack:     make([]frame, 0, 16),
		exhausted: false,
	}
	it.seekTo(pos)
	return it
}

// moveToFirst positions the iterator at the first character.
func (it *Iterator) moveToFirst() {
	if it.rope == nil || it.rope.Length() == 0 {
		it.exhausted = true
		return
	}

	it.stack = it.stack[:0]
	it.pushLeft(it.rope.root)
	it.loadCurrentLeaf()
}

// pushLeft pushes the leftmost path from node to the stack.
func (it *Iterator) pushLeft(node RopeNode) {
	it.stack = append(it.stack, frame{node: node, position: 0})

	for !node.IsLeaf() {
		internal := node.(*InternalNode)
		node = internal.left
		it.stack = append(it.stack, frame{node: node, position: 0})
	}
}

// loadCurrentLeaf loads the current leaf text from the stack.
func (it *Iterator) loadCurrentLeaf() {
	if len(it.stack) == 0 {
		it.exhausted = true
		return
	}

	frame := it.stack[len(it.stack)-1]
	if frame.node.IsLeaf() {
		leaf := frame.node.(*LeafNode)
		it.current = leaf.text
		it.runePos = frame.position

		// Calculate byte position for this rune position
		it.bytePos = 0
		for i := 0; i < frame.position; i++ {
			_, size := utf8.DecodeRuneInString(it.current[it.bytePos:])
			it.bytePos += size
		}

		it.exhausted = false
	} else {
		it.exhausted = true
	}
}

// Next advances the iterator to the next character.
// Returns false if there are no more characters.
func (it *Iterator) Next() bool {
	if it.exhausted {
		return false
	}

	// First call - initialize iterator to first character
	if it.runePos == -1 {
		it.moveToFirst()
		if it.exhausted {
			return false
		}
		// moveToFirst() positioned us at first character
		it.position = 1
		return true
	}

	// After IteratorAt/Seek, the first Next() should only advance position,
	// not the character (because IteratorAt already positioned us on a character).
	// We detect this by checking if position matches what we set in IteratorAt/Seek.
	if it.position == it.runePos {
		it.position++
		return !it.exhausted
	}

	// Move to next rune in current leaf
	// Get size of current rune to advance byte position
	_, size := utf8.DecodeRuneInString(it.current[it.bytePos:])
	it.bytePos += size
	it.runePos++
	it.position++

	// Check if we've exhausted the current leaf
	if it.bytePos >= len(it.current) {
		it.advanceToNextLeaf()
	}

	return !it.exhausted
}

// advanceToNextLeaf moves to the next leaf in the tree.
func (it *Iterator) advanceToNextLeaf() {
	for len(it.stack) > 0 {
		frame := it.stack[len(it.stack)-1]
		it.stack = it.stack[:len(it.stack)-1]

		if !frame.node.IsLeaf() {
			internal := frame.node.(*InternalNode)
			// Move to right subtree
			it.pushLeft(internal.right)
			it.loadCurrentLeaf()
			return
		}
	}

	it.exhausted = true
	it.current = ""
}

// Previous moves the iterator to the previous character.
// Returns false if there are no more characters.
func (it *Iterator) Previous() bool {
	if it.position <= 0 {
		return false
	}

	// Simple implementation: seek to previous position
	it.position--
	it.seekTo(it.position)
	return true
}

// Seek moves the iterator to the given character position.
// Returns false if the position is out of bounds.
// After seeking, Current() will return the character at `pos`.
func (it *Iterator) Seek(pos int) bool {
	if pos < 0 || pos > it.rope.Length() {
		return false
	}

	it.position = pos
	it.seekTo(pos)
	return true
}

// seekTo positions the iterator at the given character position.
func (it *Iterator) seekTo(pos int) {
	it.stack = it.stack[:0]
	it.exhausted = false

	if pos == it.rope.Length() {
		// Seek to end (one past the last character)
		it.pushRight(it.rope.root)
		it.loadCurrentLeaf()
		if !it.exhausted {
			// Position at the end of the last leaf
			it.runePos = utf8.RuneCountInString(it.current)
			it.bytePos = len(it.current)
		}
		return
	}

	it.seekInNode(it.rope.root, pos)
	// loadCurrentLeaf() already positioned us correctly with both runePos and bytePos
}

// seekToForIteratorAt positions the iterator at the given character position
// for IteratorAt/Seek. Unlike seekTo(), this does NOT decrement runePos,
// so the first Next() call will return the character at the current position.
func (it *Iterator) seekToForIteratorAt(pos int) {
	it.stack = it.stack[:0]
	it.exhausted = false

	if pos == it.rope.Length() {
		// Seek to end
		it.pushRight(it.rope.root)
		it.loadCurrentLeaf()
		if !it.exhausted {
			it.runePos = utf8.RuneCountInString(it.current)
			it.bytePos = len(it.current)
		}
		return
	}

	it.seekInNode(it.rope.root, pos)
	// Note: we do NOT decrement runePos here, unlike seekTo()
	// This means the first Next() will return the character at 'pos'
}

// seekInNode seeks to a position within a node.
func (it *Iterator) seekInNode(node RopeNode, pos int) {
	if node.IsLeaf() {
		it.stack = append(it.stack, frame{node: node, position: pos})
		it.loadCurrentLeaf()
		return
	}

	internal := node.(*InternalNode)
	leftLen := internal.left.Length()

	if pos < leftLen {
		it.stack = append(it.stack, frame{node: node, position: 0})
		it.seekInNode(internal.left, pos)
	} else {
		it.stack = append(it.stack, frame{node: node, position: leftLen})
		it.seekInNode(internal.right, pos-leftLen)
	}
}

// pushRight pushes the rightmost path from node to the stack.
func (it *Iterator) pushRight(node RopeNode) {
	it.stack = append(it.stack, frame{node: node, position: node.Length()})

	for !node.IsLeaf() {
		internal := node.(*InternalNode)
		node = internal.right
		it.stack = append(it.stack, frame{node: node, position: node.Length()})
	}
}

// Current returns the character at the current position.
// Panics if the iterator is exhausted.
func (it *Iterator) Current() rune {
	if it.exhausted || it.runePos < 0 {
		panic("iterator is exhausted or not positioned")
	}

	if it.bytePos < 0 || it.bytePos >= len(it.current) {
		panic("iterator byte position out of bounds")
	}

	rune, _ := utf8.DecodeRuneInString(it.current[it.bytePos:])
	return rune
}

// Position returns the current character position.
func (it *Iterator) Position() int {
	return it.position
}

// HasNext returns true if there are more characters to iterate.
func (it *Iterator) HasNext() bool {
	return !it.exhausted && it.position < it.rope.Length()
}

// HasPrevious returns true if there are characters before the current position.
func (it *Iterator) HasPrevious() bool {
	return it.position > 0
}

// IsAtStart returns true if the iterator is at the start of the rope.
func (it *Iterator) IsAtStart() bool {
	return it.position == 0
}

// IsAtEnd returns true if the iterator is at the end of the rope.
func (it *Iterator) IsAtEnd() bool {
	return it.position == it.rope.Length()
}

// Remaining returns the number of characters remaining from the current position.
func (it *Iterator) Remaining() int {
	return it.rope.Length() - it.position
}

// Reset resets the iterator to the beginning of the rope.
func (it *Iterator) Reset() {
	it.position = 0
	it.moveToFirst()
}

// Slice returns a substring from the current position to the given end position.
// The end position is relative to the current position.
func (it *Iterator) Slice(length int) string {
	if length < 0 {
		panic("slice length cannot be negative")
	}
	if length == 0 {
		return ""
	}
	if it.position+length > it.rope.Length() {
		panic("slice exceeds rope bounds")
	}

	return it.rope.Slice(it.position, it.position+length)
}

// Peek returns the next character without advancing the iterator.
// Returns (0, false) if there are no more characters.
func (it *Iterator) Peek() (rune, bool) {
	if it.exhausted || it.position >= it.rope.Length() {
		return 0, false
	}

	// Initialize iterator if this is the first operation
	if it.runePos == -1 {
		it.moveToFirst()
		if it.exhausted {
			return 0, false
		}
		// Don't advance position - we're just peeking
		return it.Current(), true
	}

	// Save current state
	pos := it.position

	// Get current character
	ch := it.Current()

	// Restore state (in case we need to load a new leaf)
	if pos != it.position {
		it.seekTo(pos)
	}

	return ch, true
}

// PeekNext returns the next character (after current) without advancing.
// Returns (0, false) if there is no next character.
func (it *Iterator) PeekNext() (rune, bool) {
	if it.position+1 >= it.rope.Length() {
		return 0, false
	}

	// Save state
	pos := it.position
	defer it.seekTo(pos)

	// Move to next and peek
	it.Next()
	ch := it.Current()
	return ch, true
}

// Skip moves the iterator forward by the given number of characters.
// Returns the number of characters actually skipped (may be less if near end).
func (it *Iterator) Skip(count int) int {
	if count <= 0 {
		return 0
	}

	maxSkip := it.rope.Length() - it.position
	if count > maxSkip {
		count = maxSkip
	}

	it.position += count
	it.seekTo(it.position)
	return count
}

// Collect collects all remaining characters into a string.
func (it *Iterator) Collect() string {
	if it.position >= it.rope.Length() {
		return ""
	}
	return it.rope.Slice(it.position, it.rope.Length())
}

// CollectToSlice collects all remaining characters into a rune slice.
func (it *Iterator) CollectToSlice() []rune {
	return []rune(it.Collect())
}

// FindNext searches for the given substring starting from the current position.
// Returns true if found and positions the iterator at the start of the match.
// Returns false if not found (iterator position unchanged).
func (it *Iterator) FindNext(substring string) bool {
	pos := it.rope.String()[it.bytePosition():]
	idx := indexOfSubstring(pos, substring)
	if idx < 0 {
		return false
	}

	// Convert byte index to character index
	bytePos := it.bytePosition() + idx
	charPos := it.rope.IndexFromByte(bytePos)
	it.Seek(charPos)
	return true
}

// FindNextRune searches for the given rune starting from the current position.
// Returns true if found and positions the iterator at the rune.
// Returns false if not found (iterator position unchanged).
func (it *Iterator) FindNextRune(r rune) bool {
	for it.Next() {
		if it.Current() == r {
			return true
		}
	}
	return false
}

// bytePosition returns the byte position corresponding to the current character position.
func (it *Iterator) bytePosition() int {
	if it.position == 0 {
		return 0
	}
	return len(it.rope.Slice(0, it.position))
}

// IndexFromByte converts a byte position to a character position.
func (r *Rope) IndexFromByte(bytePos int) int {
	if bytePos < 0 || bytePos > r.Size() {
		return -1
	}
	return utf8.RuneCountInString(r.String()[:bytePos])
}

// indexOfSubstring returns the byte index of substring in s, or -1 if not found.
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ========== Runes Iterator ==========

// Runes returns a slice of all runes in the rope.
func (r *Rope) Runes() []rune {
	return []rune(r.String())
}

// ForEach calls the given function for each character in the rope.
func (r *Rope) ForEach(fn func(rune)) {
	it := r.NewIterator()
	for it.Next() {
		fn(it.Current())
	}
}

// ForEachWithIndex calls the given function for each character with its position.
func (r *Rope) ForEachWithIndex(fn func(int, rune)) {
	it := r.NewIterator()
	for it.Next() {
		fn(it.Position()-1, it.Current())
	}
}

// Map creates a new rope by applying a function to each character.
func (r *Rope) Map(fn func(rune) rune) *Rope {
	builder := NewBuilder()
	r.ForEach(func(ch rune) {
		builder.AppendRune(fn(ch))
	})
	return builder.Build()
}

// Filter creates a new rope containing only characters that satisfy the predicate.
func (r *Rope) Filter(fn func(rune) bool) *Rope {
	builder := NewBuilder()
	r.ForEach(func(ch rune) {
		if fn(ch) {
			builder.AppendRune(ch)
		}
	})
	return builder.Build()
}

// Reduce reduces the rope to a single value using the given function.
func (r *Rope) Reduce(initial interface{}, fn func(accum interface{}, ch rune) interface{}) interface{} {
	accum := initial
	r.ForEach(func(ch rune) {
		accum = fn(accum, ch)
	})
	return accum
}

// Any returns true if any character satisfies the predicate.
func (r *Rope) Any(fn func(rune) bool) bool {
	result := false
	r.ForEach(func(ch rune) {
		if fn(ch) {
			result = true
		}
	})
	return result
}

// All returns true if all characters satisfy the predicate.
func (r *Rope) All(fn func(rune) bool) bool {
	result := true
	r.ForEach(func(ch rune) {
		if !fn(ch) {
			result = false
		}
	})
	return result
}

// Count returns the number of characters that satisfy the predicate.
func (r *Rope) Count(fn func(rune) bool) int {
	count := 0
	r.ForEach(func(ch rune) {
		if fn(ch) {
			count++
		}
	})
	return count
}
