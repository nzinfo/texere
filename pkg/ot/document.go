package ot

// Document represents an editable document that can be transformed by OT operations.
//
// This interface abstracts the underlying document implementation, allowing
// OT operations to work with different document types (string, rope, etc.).
//
// Example implementations:
//   - StringDocument: Simple string-based implementation
//   - RopeDocument: Efficient rope-based implementation for large documents
type Document interface {
	// Length returns the length of the document in bytes.
	Length() int

	// String returns the document content as a string.
	String() string

	// Slice returns a substring of the document.
	// Parameters:
	//   - start: starting byte position (inclusive)
	//   - end: ending byte position (exclusive)
	// Returns:
	//   - the substring from start to end
	Slice(start, end int) string

	// Bytes returns the document content as a byte slice.
	Bytes() []byte

	// Clone creates a deep copy of the document.
	Clone() Document
}
