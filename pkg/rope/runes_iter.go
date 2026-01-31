package rope

import "unicode/utf8"

// ========== Rune Iterator ==========

// Iterator iterates over runes in a rope.
//
// Semantics:
// - charPos: position of the last rune returned by Next() (-1 if none)
// - Position(): returns charPos + 1 (the next rune to be returned)
// - Current(): returns the rune most recently returned by Next()
type Iterator struct {
	rope         *Rope
	chunksIter   *ChunksIterator
	currentChunk string
	chunkPos     int  // Position within current chunk (in bytes)
	charPos      int  // Position of last returned rune (-1 if none)
	currentRune  rune // Current rune (valid after Next() returns true)
	exhausted    bool
}

// NewIterator creates a new iterator starting from the beginning of the rope.
func (r *Rope) NewIterator() *Iterator {
	if r == nil || r.Length() == 0 {
		return &Iterator{rope: r, exhausted: true}
	}

	it := &Iterator{
		rope:        r,
		chunksIter:  r.Chunks(),
		charPos:     -1, // No rune returned yet
		currentRune: 0,
		exhausted:   false,
	}
	return it
}

// IteratorAt creates a new iterator starting at the specified character position.
// The iterator is positioned so that the first Next() call will return the rune at position pos.
func (r *Rope) IteratorAt(pos int) *Iterator {
	if r == nil || r.Length() == 0 {
		return &Iterator{rope: r, exhausted: true}
	}

	if pos < 0 {
		pos = 0
	}
	if pos >= r.Length() {
		pos = r.Length() - 1
	}

	it := &Iterator{
		rope:        r,
		chunksIter:  r.Chunks(),
		charPos:     pos - 1, // Will become pos after first Next()
		currentRune: 0,
		exhausted:   false,
	}

	// Find and load the chunk containing the target position
	targetCharIdx := pos
	currentCharIdx := 0
	found := false

	for it.chunksIter.Next() {
		chunk := it.chunksIter.Current()
		chunkLen := utf8.RuneCountInString(chunk)

		if currentCharIdx+chunkLen > targetCharIdx {
			// This chunk contains the target position
			it.currentChunk = chunk
			// Calculate byte position within this chunk
			charsIntoChunk := targetCharIdx - currentCharIdx
			bytePos := 0
			for i := 0; i < charsIntoChunk; i++ {
				_, size := utf8.DecodeRuneInString(it.currentChunk[bytePos:])
				bytePos += size
			}
			it.chunkPos = bytePos
			found = true
			break
		}

		currentCharIdx += chunkLen
	}

	if !found {
		it.exhausted = true
	}

	return it
}

// Next advances to the next rune and returns true if there are more runes.
func (it *Iterator) Next() bool {
	if it.exhausted {
		return false
	}

	// If we don't have a current chunk, get the first one
	if it.currentChunk == "" {
		if !it.chunksIter.Next() {
			it.exhausted = true
			return false
		}
		it.currentChunk = it.chunksIter.Current()
		it.chunkPos = 0
	}

	// Check if we've exhausted the current chunk
	if it.chunkPos >= len(it.currentChunk) {
		// Move to next chunk
		it.currentChunk = ""
		it.chunkPos = 0
		return it.Next()
	}

	// Decode the next rune
	r, size := utf8.DecodeRuneInString(it.currentChunk[it.chunkPos:])
	it.currentRune = r
	it.chunkPos += size
	it.charPos++

	return true
}

// Current returns the current rune.
// Panics if Next() hasn't been called yet or the iterator is exhausted.
func (it *Iterator) Current() rune {
	if it.currentRune == 0 {
		panic("iterator not positioned on a rune")
	}
	return it.currentRune
}

// Position returns the character position of the next rune to be returned.
// Consistently returns charPos + 1.
func (it *Iterator) Position() int {
	return it.charPos + 1
}

// Reset resets the iterator to the beginning of the rope.
func (it *Iterator) Reset() {
	if it.rope == nil || it.rope.Length() == 0 {
		it.exhausted = true
		return
	}

	it.chunksIter = it.rope.Chunks()
	it.currentChunk = ""
	it.chunkPos = 0
	it.charPos = -1
	it.currentRune = 0
	it.exhausted = false
}

// Collect collects all remaining runes into a slice.
func (it *Iterator) Collect() []rune {
	var runes []rune
	for it.Next() {
		runes = append(runes, it.Current())
	}
	return runes
}

// ToSlice is an alias for Collect.
func (it *Iterator) ToSlice() []rune {
	return it.Collect()
}

// HasNext returns true if there are more runes to iterate.
func (it *Iterator) HasNext() bool {
	if it.exhausted {
		return false
	}

	if it.currentChunk != "" && it.chunkPos < len(it.currentChunk) {
		// Check if there's another rune in current chunk
		_, size := utf8.DecodeRuneInString(it.currentChunk[it.chunkPos:])
		return it.chunkPos+size <= len(it.currentChunk)
	}

	// Check if there are more chunks
	return it.chunksIter.Position()+1 < it.chunksIter.Count()
}

// Seek positions the iterator at the specified character position.
// Returns true if the position is valid.
func (it *Iterator) Seek(pos int) bool {
	if it.rope == nil || it.rope.Length() == 0 {
		it.exhausted = true
		return false
	}

	if pos < 0 {
		pos = 0
	}
	if pos >= it.rope.Length() {
		it.exhausted = true
		return false
	}

	// Create a new iterator at the target position
	newIt := it.rope.IteratorAt(pos)
	*it = *newIt
	return true
}

// HasPrevious returns true if there is a previous rune.
// Note: This is a limited implementation - for full backwards iteration,
// use ReverseIterator instead.
func (it *Iterator) HasPrevious() bool {
	return it.charPos > 0
}

// Previous moves to the previous rune.
// Note: This is a limited implementation that resets and iterates to position-1.
// For efficient backwards iteration, use ReverseIterator instead.
func (it *Iterator) Previous() bool {
	if !it.HasPrevious() {
		return false
	}

	// Reset to beginning and iterate to charPos - 1
	targetPos := it.charPos - 1
	newIt := it.rope.IteratorAt(targetPos)
	*it = *newIt
	return true
}

// Peek returns the next rune without advancing the iterator.
// Returns the rune and true if there is a next rune, or (0, false) if exhausted.
func (it *Iterator) Peek() (rune, bool) {
	if it.exhausted {
		return 0, false
	}

	// Save current state
	oldChunk := it.currentChunk
	oldChunkPos := it.chunkPos
	oldCharPos := it.charPos
	oldExhausted := it.exhausted
	oldCurrentRune := it.currentRune

	// Advance to next
	hasNext := it.Next()

	// Get current rune
	var r rune
	if hasNext && !it.exhausted {
		r = it.Current()
	}

	// Restore state
	it.currentChunk = oldChunk
	it.chunkPos = oldChunkPos
	it.charPos = oldCharPos
	it.exhausted = oldExhausted
	it.currentRune = oldCurrentRune

	return r, hasNext
}

// Skip advances the iterator by n runes.
// Returns the number of runes actually skipped.
func (it *Iterator) Skip(n int) int {
	if it.exhausted || n <= 0 {
		return 0
	}

	skipped := 0
	for i := 0; i < n && it.Next(); i++ {
		skipped++
	}
	return skipped
}
