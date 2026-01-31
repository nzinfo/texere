package rope

import (
	"unicode"
)

// WordBoundary finds word boundaries in the document.
type WordBoundary struct {
	rope *Rope
}

// NewWordBoundary creates a new word boundary finder.
func NewWordBoundary(rope *Rope) *WordBoundary {
	return &WordBoundary{rope: rope}
}

// IsWordChar returns true if the rune is a word character.
// Word characters are: letters, digits, and underscore.
func (wb *WordBoundary) IsWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// IsWhitespace returns true if the rune is whitespace.
func (wb *WordBoundary) IsWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

// PrevWordStart finds the start of the word before the given position.
// Returns the position of the first character of the previous word.
func (wb *WordBoundary) PrevWordStart(pos int) int {
	if pos <= 0 {
		return 0
	}

	if pos > wb.rope.Length() {
		pos = wb.rope.Length()
	}

	// Create iterator at position
	iter := wb.rope.IteratorAt(pos)
	if iter.HasPrevious() {
		iter.Previous()
	}

	// Skip whitespace backwards
	for iter.HasPrevious() {
		r := iter.Current()
		if !wb.IsWhitespace(r) {
			break
		}
		iter.Previous()
	}

	// Now find the start of the word
	wordStart := iter.Position()
	for iter.HasPrevious() {
		r := iter.Current()
		if wb.IsWhitespace(r) {
			break
		}
		iter.Previous()
		wordStart = iter.Position()
	}

	return wordStart
}

// NextWordStart finds the start of the word after the given position.
// Returns the position of the first character of the next word.
func (wb *WordBoundary) NextWordStart(pos int) int {
	if pos < 0 {
		return 0
	}

	if pos >= wb.rope.Length() {
		return wb.rope.Length()
	}

	// Create iterator at position
	iter := wb.rope.IteratorAt(pos)

	// Skip whitespace forwards
	for iter.Next() {
		r := iter.Current()
		if !wb.IsWhitespace(r) {
			break
		}
	}

	// Now at start of word, find end of word
	wordStart := iter.Position()
	for iter.Next() {
		r := iter.Current()
		if wb.IsWhitespace(r) {
			break
		}
	}

	return wordStart
}

// PrevWordEnd finds the end of the word before the given position.
// Returns the position after the last character of the previous word.
func (wb *WordBoundary) PrevWordEnd(pos int) int {
	if pos <= 0 {
		return 0
	}

	if pos > wb.rope.Length() {
		pos = wb.rope.Length()
	}

	// Skip whitespace backwards
	iter := wb.rope.IteratorAt(pos)
	currentPos := pos

	for iter.Previous() {
		r := iter.Current()
		if !wb.IsWhitespace(r) {
			break
		}
		currentPos--
	}

	// Now find the end of the word (which is currentPos)
	return currentPos
}

// NextWordEnd finds the end of the word after the given position.
// Returns the position after the last character of the next word.
func (wb *WordBoundary) NextWordEnd(pos int) int {
	if pos < 0 {
		return 0
	}

	if pos >= wb.rope.Length() {
		return wb.rope.Length()
	}

	// Skip whitespace forwards
	iter := wb.rope.IteratorAt(pos)
	currentPos := pos

	for iter.Next() {
		r := iter.Current()
		if !wb.IsWhitespace(r) {
			break
		}
		currentPos++
	}

	// Now find the end of the word
	for iter.Next() {
		r := iter.Current()
		if wb.IsWhitespace(r) {
			break
		}
		currentPos++
	}

	return currentPos
}

// CurrentWordStart finds the start of the word at the given position.
// If the position is on whitespace, returns the position.
func (wb *WordBoundary) CurrentWordStart(pos int) int {
	if pos < 0 {
		return 0
	}

	if pos >= wb.rope.Length() {
		return wb.rope.Length()
	}

	iter := wb.rope.IteratorAt(pos)

	// Check if current position is on a word character
	if iter.Previous() {
		r := iter.Current()
		if !wb.IsWordChar(r) && !wb.IsWhitespace(r) {
			// On a non-word, non-space character
			iter.Next()
			return pos
		}
		iter.Next()
	} else {
		// At start of document
		return 0
	}

	// Check if we're on whitespace
	iter.Seek(pos)
	r := iter.Current()
	if wb.IsWhitespace(r) {
		return pos
	}

	// We're on a word, find its start
	wordStart := pos
	for iter.Previous() {
		r := iter.Current()
		if !wb.IsWordChar(r) {
			break
		}
		wordStart--
	}

	return wordStart
}

// CurrentWordEnd finds the end of the word at the given position.
// If the position is on whitespace, returns the position.
func (wb *WordBoundary) CurrentWordEnd(pos int) int {
	if pos < 0 {
		return 0
	}

	if pos >= wb.rope.Length() {
		return wb.rope.Length()
	}

	iter := wb.rope.IteratorAt(pos)
	r := iter.Current()

	// Check if current position is on whitespace
	if wb.IsWhitespace(r) {
		return pos
	}

	// Check if current position is on a non-word character
	if !wb.IsWordChar(r) {
		return pos + 1
	}

	// We're on a word, find its end
	wordEnd := pos
	for iter.Next() {
		r := iter.Current()
		if !wb.IsWordChar(r) {
			break
		}
		wordEnd++
	}

	return wordEnd
}

// WordAt returns the word at the given position, along with its start and end positions.
// If the position is not on a word, returns empty string.
func (wb *WordBoundary) WordAt(pos int) (string, int, int) {
	start := wb.CurrentWordStart(pos)
	end := wb.CurrentWordEnd(pos)

	if start >= end {
		return "", start, end
	}

	word := wb.rope.Slice(start, end)
	return word, start, end
}

// SelectWord selects the word at the given position.
// Returns the start and end positions of the word.
func (wb *WordBoundary) SelectWord(pos int) (int, int) {
	return wb.CurrentWordStart(pos), wb.CurrentWordEnd(pos)
}

// FindBoundary finds a boundary in the specified direction.
// Direction: -1 for backward, 1 for forward
func (wb *WordBoundary) FindBoundary(pos int, direction int) int {
	if direction < 0 {
		return wb.PrevWordStart(pos)
	}
	return wb.NextWordEnd(pos)
}

// MoveToWordBoundary moves the cursor to a word boundary.
// Assoc determines which boundary to move to.
func (wb *WordBoundary) MoveToWordBoundary(pos int, assoc Assoc) int {
	switch assoc {
	case AssocBeforeWord:
		return wb.PrevWordStart(pos)
	case AssocAfterWord:
		return wb.NextWordStart(pos)
	default:
		return pos
	}
}

// BigWordStart finds the start of the "big word" before the given position.
// Big words are separated by whitespace only.
func (wb *WordBoundary) BigWordStart(pos int) int {
	if pos <= 0 {
		return 0
	}

	if pos > wb.rope.Length() {
		pos = wb.rope.Length()
	}

	iter := wb.rope.IteratorAt(pos)
	currentPos := pos

	// Skip whitespace backwards
	for iter.Previous() {
		r := iter.Current()
		if !wb.IsWhitespace(r) {
			break
		}
		currentPos--
	}

	// Now find the start of the big word
	wordStart := currentPos
	for iter.Previous() {
		r := iter.Current()
		if wb.IsWhitespace(r) {
			break
		}
		wordStart--
	}

	return wordStart
}

// BigWordEnd finds the end of the "big word" after the given position.
// Big words are separated by whitespace only.
func (wb *WordBoundary) BigWordEnd(pos int) int {
	if pos < 0 {
		return 0
	}

	if pos >= wb.rope.Length() {
		return wb.rope.Length()
	}

	iter := wb.rope.IteratorAt(pos)
	currentPos := pos

	// Skip whitespace forwards
	for iter.Next() {
		r := iter.Current()
		if !wb.IsWhitespace(r) {
			break
		}
		currentPos++
	}

	// Now find the end of the big word
	for iter.Next() {
		r := iter.Current()
		if wb.IsWhitespace(r) {
			break
		}
		currentPos++
	}

	return currentPos
}

// ParagraphStart finds the start of the paragraph before the given position.
// Paragraphs are separated by one or more newline characters.
func (wb *WordBoundary) ParagraphStart(pos int) int {
	if pos <= 0 {
		return 0
	}

	if pos > wb.rope.Length() {
		pos = wb.rope.Length()
	}

	iter := wb.rope.IteratorAt(pos)
	currentPos := pos

	// Find previous newline or start of document
	for iter.Previous() {
		r := iter.Current()
		currentPos--
		if r == '\n' {
			// Skip the newline and check if there are more
			for iter.Previous() {
				r := iter.Current()
				if r != '\n' {
					break
				}
				currentPos--
			}
			// Return position after the newlines
			return currentPos + 1
		}
	}

	return 0
}

// ParagraphEnd finds the end of the paragraph after the given position.
// Paragraphs are separated by one or more newline characters.
func (wb *WordBoundary) ParagraphEnd(pos int) int {
	if pos < 0 {
		return 0
	}

	if pos >= wb.rope.Length() {
		return wb.rope.Length()
	}

	iter := wb.rope.IteratorAt(pos)
	currentPos := pos

	// Find next newline or end of document
	for iter.Next() {
		r := iter.Current()
		currentPos++
		if r == '\n' {
			return currentPos
		}
	}

	return wb.rope.Length()
}

// LineStart finds the start of the line at the given position.
func (wb *WordBoundary) LineStart(pos int) int {
	if pos <= 0 {
		return 0
	}

	if pos > wb.rope.Length() {
		pos = wb.rope.Length()
	}

	iter := wb.rope.IteratorAt(pos)
	currentPos := pos

	// Find previous newline or start of document
	for iter.Previous() {
		r := iter.Current()
		currentPos--
		if r == '\n' {
			return currentPos + 1
		}
	}

	return 0
}

// LineEnd finds the end of the line at the given position.
func (wb *WordBoundary) LineEnd(pos int) int {
	if pos < 0 {
		return 0
	}

	if pos >= wb.rope.Length() {
		return wb.rope.Length()
	}

	iter := wb.rope.IteratorAt(pos)
	currentPos := pos

	// Find next newline or end of document
	for iter.Next() {
		r := iter.Current()
		currentPos++
		if r == '\n' {
			return currentPos
		}
	}

	return wb.rope.Length()
}

// ScreenLineStart finds the start of the screen line at the given position.
// This considers soft wrapping (not yet implemented).
func (wb *WordBoundary) ScreenLineStart(pos int) int {
	// For now, same as LineStart
	// A full implementation would consider soft wrapping and display width
	return wb.LineStart(pos)
}

// ScreenLineEnd finds the end of the screen line at the given position.
// This considers soft wrapping (not yet implemented).
func (wb *WordBoundary) ScreenLineEnd(pos int) int {
	// For now, same as LineEnd
	// A full implementation would consider soft wrapping and display width
	return wb.LineEnd(pos)
}
