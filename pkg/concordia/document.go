// Package concordia provides document abstractions for OT operations.
//
// The concordia package defines interfaces for text documents that can be
// manipulated by OT operations. This allows different underlying representations
// (string, rope, piece table, etc.) to be used interchangeably.
package concordia

// Document represents a text document that can be manipulated by OT operations.
//
// The Document interface provides the minimal set of operations needed for
// OT algorithms to work with different document representations.
//
// Examples:
//
//	// String document (simple, efficient for small documents)
//	doc := NewStringDocument("Hello World")
//
//	// Rope document (efficient for large documents with frequent edits)
//	// doc := NewRopeDocument(largeContent)
type Document interface {
	// Length returns the length of the document in characters.
	Length() int

	// Slice returns a substring of the document from start to end.
	// The result is implementation-specific but should be efficient.
	//
	// Parameters:
	//   - start: starting index (inclusive)
	//   - end: ending index (exclusive)
	//
	// Returns:
	//   - the substring from start to end
	Slice(start, end int) string

	// String returns the entire document as a string.
	// This may be inefficient for large documents; consider using Slice instead.
	String() string

	// Bytes returns the entire document as a byte slice.
	// This is useful for serialization or I/O operations.
	Bytes() []byte

	// Clone creates a deep copy of the concordia.
	// The returned document is independent and can be modified without
	// affecting the original.
	Clone() Document
}
