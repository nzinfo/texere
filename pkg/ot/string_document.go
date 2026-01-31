package ot

import (
	"fmt"
	"unicode/utf8"
)

// StringDocument is a simple string-based implementation of the Document interface.
//
// This is a basic implementation suitable for small documents. For large documents,
// consider using a RopeDocument from the concordia package.
//
// Example:
//
//	doc := &StringDocument{content: "Hello World"}
//	op := ot.NewBuilder().Retain(6).Insert("Go ").Build()
//	newDoc, _ := op.ApplyToDocument(doc)
type StringDocument struct {
	content string
	history []*Operation // Simple history stack for undo
	historyPos int       // Current position in history (-1 means no history)
}

// NewStringDocument creates a new StringDocument with the given content.
//
// Parameters:
//   - content: the initial document content
//
// Returns:
//   - a new StringDocument
//
// Example:
//
//	doc := NewStringDocument("Hello World")
func NewStringDocument(content string) *StringDocument {
	return &StringDocument{
		content: content,
		history: make([]*Operation, 0, 50),
		historyPos: -1,
	}
}

// Length returns the length of the document in UTF-16 code units.
// This matches JavaScript's string.length behavior.
func (d *StringDocument) Length() int {
	// Count UTF-16 code units (not runes, not bytes)
	// This matches JavaScript's string.length
	count := 0
	for _, r := range d.content {
		if r >= 0x10000 {
			// Characters outside BMP need 2 UTF-16 code units (surrogate pair)
			count += 2
		} else {
			count += 1
		}
	}
	return count
}

// LengthBytes returns the length of the document in bytes.
// This is an alias for Length() for explicit intent.
func (d *StringDocument) LengthBytes() int {
	return d.Length()
}

// LengthChars returns the length of the document in characters (code points).
func (d *StringDocument) LengthChars() int {
	return utf8.RuneCountInString(d.content)
}

// String returns the document content as a string.
func (d *StringDocument) String() string {
	return d.content
}

// Slice returns a substring of the document.
//
// Parameters:
//   - start: starting byte position (inclusive)
//   - end: ending byte position (exclusive)
//
// Returns:
//   - the substring from start to end
func (d *StringDocument) Slice(start, end int) string {
	return d.content[start:end]
}

// Bytes returns the document content as a byte slice.
func (d *StringDocument) Bytes() []byte {
	return []byte(d.content)
}

// Clone creates a deep copy of the document.
func (d *StringDocument) Clone() Document {
	return &StringDocument{content: d.content}
}

// ========== UndoableDocument Interface ==========

// Undo reverses the last operation.
func (d *StringDocument) Undo() error {
	if !d.CanUndo() {
		return fmt.Errorf("cannot undo")
	}

	// Get the operation to undo (invert it)
	op := d.history[d.historyPos]
	inverse := op.Invert(d.content)

	// Apply inverse operation
	newContent, err := inverse.Apply(d.content)
	if err != nil {
		return err
	}

	d.content = newContent
	d.historyPos--
	return nil
}

// Redo reapplies the most recently undone operation.
func (d *StringDocument) Redo() error {
	if !d.CanRedo() {
		return fmt.Errorf("cannot redo")
	}

	// Move forward in history
	d.historyPos++
	op := d.history[d.historyPos]

	// Apply the operation
	newContent, err := op.Apply(d.content)
	if err != nil {
		return err
	}

	d.content = newContent
	return nil
}

// CanUndo returns true if undo is possible.
func (d *StringDocument) CanUndo() bool {
	return d.historyPos >= 0
}

// CanRedo returns true if redo is possible.
func (d *StringDocument) CanRedo() bool {
	return d.historyPos < len(d.history)-1
}

// ApplyOperationWithHistory applies an operation and records it to history.
func (d *StringDocument) ApplyOperationWithHistory(op *Operation) (UndoableDocument, error) {
	// Apply operation
	newContent, err := op.Apply(d.content)
	if err != nil {
		return nil, err
	}

	// Update content
	d.content = newContent

	// Add to history (remove any redo history first)
	if d.historyPos < len(d.history)-1 {
		d.history = d.history[:d.historyPos+1]
	}
	d.history = append(d.history, op)
	d.historyPos++

	return d, nil
}
