package concordia

// StringDocument is a simple string-based document implementation.
//
// This is the most straightforward implementation of the Document interface,
// using a Go string as the underlying storage. It's efficient for small to
// medium-sized documents and for documents that don't require frequent
// insertions/deletions in the middle.
//
// For large documents with frequent edits, consider using a Rope-based
// implementation instead.
type StringDocument struct {
	content string
}

// NewStringDocument creates a new StringDocument from a string.
//
// Parameters:
//   - content: the initial document content
//
// Returns:
//   - a new StringDocument containing the given content
//
// Example:
//
//	doc := NewStringDocument("Hello World")
//	fmt.Println(doc.Length()) // 11
func NewStringDocument(content string) *StringDocument {
	return &StringDocument{
		content: content,
	}
}

// Length returns the length of the document in characters.
//
// This is O(1) time complexity.
func (d *StringDocument) Length() int {
	return len(d.content)
}

// Slice returns a substring of the document from start to end.
//
// This uses Go's built-in string slicing, which is O(1) time complexity
// (it creates a new string header pointing to the same underlying bytes).
//
// Parameters:
//   - start: starting index (inclusive)
//   - end: ending index (exclusive)
//
// Returns:
//   - the substring from start to end
//
// Panics:
//   - if start or end are out of bounds, or if start > end
func (d *StringDocument) Slice(start, end int) string {
	return d.content[start:end]
}

// String returns the entire document as a string.
//
// This is O(1) time complexity (returns the underlying string directly).
func (d *StringDocument) String() string {
	return d.content
}

// Bytes returns the entire document as a byte slice.
//
// This is O(n) time complexity where n is the length of the document,
// as it creates a new byte slice and copies the content.
func (d *StringDocument) Bytes() []byte {
	return []byte(d.content)
}

// Clone creates a deep copy of the concordia.
//
// For StringDocument, this creates a copy of the underlying string.
// Since Go strings are immutable, this is effectively a shallow copy,
// but the returned document is independent in the sense that modifications
// to the original document (through new operations) won't affect the clone.
//
// This is O(1) time complexity.
func (d *StringDocument) Clone() Document {
	return &StringDocument{
		content: d.content,
	}
}
