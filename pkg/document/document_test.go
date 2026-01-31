package document

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// This file contains tests for the Document interface.
// Since Document is an interface, we test it through the StringDocument implementation.

// StringDocument is a simple string-based implementation of Document for testing.
type StringDocument struct {
	content string
}

// NewStringDocument creates a new StringDocument.
func NewStringDocument(content string) *StringDocument {
	return &StringDocument{content: content}
}

// Length returns the number of characters in the document.
func (d *StringDocument) Length() int {
	return len([]rune(d.content))
}

// Slice returns a substring from start to end (exclusive).
func (d *StringDocument) Slice(start, end int) string {
	runes := []rune(d.content)
	return string(runes[start:end])
}

// String returns the complete document content as a string.
func (d *StringDocument) String() string {
	return d.content
}

// Bytes returns the complete document content as a byte slice.
func (d *StringDocument) Bytes() []byte {
	return []byte(d.content)
}

// Clone creates a copy of the document.
func (d *StringDocument) Clone() Document {
	return &StringDocument{content: d.content}
}

// ========== Document Interface Tests ==========

func TestDocument_Interface_Length(t *testing.T) {
	doc := NewStringDocument("Hello World")
	assert.Equal(t, 11, doc.Length())
}

func TestDocument_Interface_Slice(t *testing.T) {
	doc := NewStringDocument("Hello World")
	assert.Equal(t, "Hello", doc.Slice(0, 5))
	assert.Equal(t, "World", doc.Slice(6, 11))
}

func TestDocument_Interface_String(t *testing.T) {
	doc := NewStringDocument("Hello World")
	assert.Equal(t, "Hello World", doc.String())
}

func TestDocument_Interface_Bytes(t *testing.T) {
	doc := NewStringDocument("Hello World")
	assert.Equal(t, []byte("Hello World"), doc.Bytes())
}

func TestDocument_Interface_Clone(t *testing.T) {
	doc := NewStringDocument("Hello World")
	doc2 := doc.Clone()

	assert.Equal(t, doc.String(), doc2.String())
	assert.NotSame(t, doc, doc2)
}

func TestDocument_Interface_UTF8(t *testing.T) {
	doc := NewStringDocument("Hello 世界")
	assert.Equal(t, 8, doc.Length()) // 5 + 1 + 2 Chinese chars
	assert.Equal(t, "Hello 世界", doc.String())
}

func TestDocument_Interface_Empty(t *testing.T) {
	doc := NewStringDocument("")
	assert.Equal(t, 0, doc.Length())
	assert.Equal(t, "", doc.String())
}

func TestDocument_Interface_Slice_UTF8(t *testing.T) {
	doc := NewStringDocument("你好世界")

	assert.Equal(t, "你好", doc.Slice(0, 2))
	assert.Equal(t, "世界", doc.Slice(2, 4))
}

// ========== Document Interface Compliance ==========

func TestDocument_Interface_Compliance(t *testing.T) {
	// This test ensures that StringDocument implements the Document interface
	var _ Document = (*StringDocument)(nil)

	doc := NewStringDocument("Test")
	assert.Implements(t, (*Document)(nil), doc)
}
