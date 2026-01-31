package ot

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
	return &StringDocument{content: content}
}

// Length returns the length of the document in bytes.
func (d *StringDocument) Length() int {
	return len(d.content)
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
