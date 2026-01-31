package rope

import (
	"github.com/texere-rope/pkg/document"
)

// RopeDocument adapts a Rope to implement the Document interface.
//
// This allows Rope to be used interchangeably with other Document implementations
// (e.g., StringDocument) in the OT (Operational Transformation) layer.
//
// Since Rope is immutable, Clone() returns the same instance without copying.
type RopeDocument struct {
	rope *Rope
}

// NewRopeDocument creates a new RopeDocument from the given text.
func NewRopeDocument(text string) *RopeDocument {
	return &RopeDocument{
		rope: New(text),
	}
}

// NewRopeDocumentFromRope creates a RopeDocument from an existing Rope.
func NewRopeDocumentFromRope(r *Rope) *RopeDocument {
	return &RopeDocument{
		rope: r,
	}
}

// Length returns the number of characters (Unicode code points) in the document.
func (d *RopeDocument) Length() int {
	if d == nil || d.rope == nil {
		return 0
	}
	return d.rope.Length()
}

// Slice returns a substring from start to end (exclusive).
// The indices are character positions (not byte positions).
func (d *RopeDocument) Slice(start, end int) string {
	if d == nil || d.rope == nil {
		return ""
	}
	return d.rope.Slice(start, end)
}

// String returns the complete document content as a string.
func (d *RopeDocument) String() string {
	if d == nil || d.rope == nil {
		return ""
	}
	return d.rope.String()
}

// Bytes returns the complete document content as a byte slice.
func (d *RopeDocument) Bytes() []byte {
	if d == nil || d.rope == nil {
		return []byte{}
	}
	return d.rope.Bytes()
}

// Clone creates a copy of the document.
// Since Rope is immutable, this returns the same instance without copying.
// The returned value can be safely used as an independent Document.
func (d *RopeDocument) Clone() document.Document {
	if d == nil {
		return &RopeDocument{rope: Empty()}
	}
	return &RopeDocument{rope: d.rope.Clone()}
}

// Rope returns the underlying Rope for direct access.
// This allows using Rope-specific operations when needed.
func (d *RopeDocument) Rope() *Rope {
	if d == nil {
		return nil
	}
	return d.rope
}

// ========== Document-specific Operations ==========

// Insert returns a new RopeDocument with text inserted at the given position.
// This is a convenience method that wraps Rope.Insert().
func (d *RopeDocument) Insert(pos int, text string) *RopeDocument {
	if d == nil {
		return &RopeDocument{rope: New(text)}
	}
	return &RopeDocument{
		rope: d.rope.Insert(pos, text),
	}
}

// Delete returns a new RopeDocument with characters removed from start to end.
// This is a convenience method that wraps Rope.Delete().
func (d *RopeDocument) Delete(start, end int) *RopeDocument {
	if d == nil {
		return &RopeDocument{rope: Empty()}
	}
	return &RopeDocument{
		rope: d.rope.Delete(start, end),
	}
}

// Replace returns a new RopeDocument with characters replaced.
// This is a convenience method that wraps Rope.Replace().
func (d *RopeDocument) Replace(start, end int, text string) *RopeDocument {
	if d == nil {
		return &RopeDocument{rope: New(text)}
	}
	return &RopeDocument{
		rope: d.rope.Replace(start, end, text),
	}
}

// Concat returns a new RopeDocument with another document appended.
func (d *RopeDocument) Concat(other document.Document) *RopeDocument {
	if d == nil {
		if other == nil {
			return &RopeDocument{rope: Empty()}
		}
		return &RopeDocument{rope: New(other.String())}
	}

	if other == nil {
		return &RopeDocument{rope: d.rope.Clone()}
	}

	// Try to optimize if the other document is also a RopeDocument
	if otherDoc, ok := other.(*RopeDocument); ok {
		return &RopeDocument{
			rope: d.rope.Concat(otherDoc.rope),
		}
	}

	// Fall back to string-based concatenation
	return &RopeDocument{
		rope: d.rope.Concat(New(other.String())),
	}
}

// Split splits the document at the given position.
func (d *RopeDocument) Split(pos int) (*RopeDocument, *RopeDocument) {
	if d == nil {
		return &RopeDocument{rope: Empty()}, &RopeDocument{rope: Empty()}
	}

	left, right := d.rope.Split(pos)
	return &RopeDocument{rope: left}, &RopeDocument{rope: right}
}

// ========== Type Conversion ==========

// AsRopeDocument attempts to convert any Document to a RopeDocument.
// If the document is already a RopeDocument, it returns it directly.
// Otherwise, it creates a new RopeDocument from the document's content.
func AsRopeDocument(doc document.Document) *RopeDocument {
	if doc == nil {
		return &RopeDocument{rope: Empty()}
	}

	if ropeDoc, ok := doc.(*RopeDocument); ok {
		return ropeDoc
	}

	// Convert from other Document implementations
	return NewRopeDocument(doc.String())
}

// IsRopeDocument returns true if the document is a RopeDocument.
func IsRopeDocument(doc document.Document) bool {
	_, ok := doc.(*RopeDocument)
	return ok
}

// ========== Document Metrics ==========

// Size returns the size in bytes of the document.
func (d *RopeDocument) Size() int {
	if d == nil || d.rope == nil {
		return 0
	}
	return d.rope.Size()
}

// Depth returns the depth of the underlying rope tree.
func (d *RopeDocument) Depth() int {
	if d == nil || d.rope == nil {
		return 0
	}
	return d.rope.Depth()
}

// IsBalanced returns true if the underlying rope is balanced.
func (d *RopeDocument) IsBalanced() bool {
	if d == nil || d.rope == nil {
		return true
	}
	return d.rope.IsBalanced()
}

// Stats returns statistics about the document's rope structure.
func (d *RopeDocument) Stats() *TreeStats {
	if d == nil || d.rope == nil {
		return &TreeStats{}
	}
	return d.rope.Stats()
}

// Balance balances the underlying rope and returns a new document.
func (d *RopeDocument) Balance() *RopeDocument {
	if d == nil || d.rope == nil {
		return &RopeDocument{rope: Empty()}
	}
	return &RopeDocument{
		rope: d.rope.Balance(),
	}
}

// Optimize optimizes the underlying rope and returns a new document.
func (d *RopeDocument) Optimize() *RopeDocument {
	if d == nil || d.rope == nil {
		return &RopeDocument{rope: Empty()}
	}
	return &RopeDocument{
		rope: d.rope.Optimize(),
	}
}

// Validate checks the integrity of the document's rope structure.
func (d *RopeDocument) Validate() error {
	if d == nil || d.rope == nil {
		return nil
	}
	return d.rope.Validate()
}

// ========== Document Comparison ==========

// Equals returns true if two documents have identical content.
func (d *RopeDocument) Equals(other document.Document) bool {
	if d == nil && other == nil {
		return true
	}
	if d == nil || other == nil {
		return false
	}
	return d.String() == other.String()
}

// Compare compares two documents lexicographically.
// Returns -1 if d < other, 0 if d == other, 1 if d > other.
func (d *RopeDocument) Compare(other document.Document) int {
	if d == nil && other == nil {
		return 0
	}
	if d == nil {
		return -1
	}
	if other == nil {
		return 1
	}

	if otherDoc, ok := other.(*RopeDocument); ok {
		return d.rope.Compare(otherDoc.rope)
	}

	// Fall back to string comparison
	dStr := d.String()
	otherStr := other.String()

	if dStr < otherStr {
		return -1
	} else if dStr > otherStr {
		return 1
	}
	return 0
}

// Contains returns true if the document contains the given substring.
func (d *RopeDocument) Contains(substring string) bool {
	if d == nil || d.rope == nil {
		return false
	}
	return d.rope.Contains(substring)
}

// Index returns the first character position of the given substring.
// Returns -1 if not found.
func (d *RopeDocument) Index(substring string) int {
	if d == nil || d.rope == nil {
		return -1
	}
	return d.rope.Index(substring)
}

// LastIndex returns the last character position of the given substring.
// Returns -1 if not found.
func (d *RopeDocument) LastIndex(substring string) int {
	if d == nil || d.rope == nil {
		return -1
	}
	return d.rope.LastIndex(substring)
}

// ========== Document Utilities ==========

// Empty creates an empty RopeDocument.
func EmptyDocument() *RopeDocument {
	return &RopeDocument{rope: Empty()}
}

// FromDocument creates a RopeDocument from any Document implementation.
func FromDocument(doc document.Document) *RopeDocument {
	if doc == nil {
		return EmptyDocument()
	}
	return AsRopeDocument(doc)
}

// CloneDocument safely clones a document, returning a RopeDocument.
func CloneDocument(doc document.Document) *RopeDocument {
	if doc == nil {
		return EmptyDocument()
	}
	return FromDocument(doc.Clone())
}

// MergeDocuments merges multiple documents into one RopeDocument.
func MergeDocuments(docs ...document.Document) *RopeDocument {
	builder := NewBuilder()
	for _, doc := range docs {
		if doc != nil {
			builder.Append(doc.String())
		}
	}
	return &RopeDocument{rope: builder.Build()}
}

// JoinDocuments joins multiple documents with a separator.
func JoinDocuments(docs []document.Document, separator string) *RopeDocument {
	if len(docs) == 0 {
		return EmptyDocument()
	}

	builder := NewBuilder()
	for i, doc := range docs {
		if doc != nil {
			builder.Append(doc.String())
		}
		if i < len(docs)-1 {
			builder.Append(separator)
		}
	}
	return &RopeDocument{rope: builder.Build()}
}

// ========== Document Persistence ==========

// ToBytes returns the document content as bytes.
func (d *RopeDocument) ToBytes() []byte {
	return d.Bytes()
}

// FromBytes creates a RopeDocument from bytes.
func FromBytes(data []byte) *RopeDocument {
	return NewRopeDocument(string(data))
}

// ToRunes returns the document content as a rune slice.
func (d *RopeDocument) ToRunes() []rune {
	if d == nil || d.rope == nil {
		return []rune{}
	}
	return d.rope.Runes()
}

// FromRunes creates a RopeDocument from a rune slice.
func FromRunes(runes []rune) *RopeDocument {
	return NewRopeDocument(string(runes))
}

// ========== Document Builder ==========

// DocumentBuilder provides a convenient way to build a RopeDocument.
type DocumentBuilder struct {
	builder *RopeBuilder
}

// NewDocumentBuilder creates a new document builder.
func NewDocumentBuilder() *DocumentBuilder {
	return &DocumentBuilder{
		builder: NewBuilder(),
	}
}

// Append appends text to the document.
func (b *DocumentBuilder) Append(text string) *DocumentBuilder {
	b.builder.Append(text)
	return b
}

// AppendLine appends a line with a newline.
func (b *DocumentBuilder) AppendLine(line string) *DocumentBuilder {
	b.builder.AppendLine(line)
	return b
}

// Insert inserts text at the given position.
func (b *DocumentBuilder) Insert(pos int, text string) *DocumentBuilder {
	b.builder.Insert(pos, text)
	return b
}

// Delete deletes characters from start to end.
func (b *DocumentBuilder) Delete(start, end int) *DocumentBuilder {
	b.builder.Delete(start, end)
	return b
}

// Build builds the final RopeDocument.
func (b *DocumentBuilder) Build() *RopeDocument {
	return &RopeDocument{
		rope: b.builder.Build(),
	}
}

// Reset resets the builder for reuse.
func (b *DocumentBuilder) Reset() *DocumentBuilder {
	b.builder.Reset()
	return b
}
