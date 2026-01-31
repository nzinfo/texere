package concordia

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// This file contains tests for the Document interface.
// We test it through the StringDocument implementation from string_document.go.

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
	// Length returns bytes, not characters
	// "Hello 世界" = 5 + 1 + 2*3 = 12 bytes (每个中文字符3字节)
	assert.Equal(t, 12, doc.Length())
	assert.Equal(t, "Hello 世界", doc.String())
}

func TestDocument_Interface_Empty(t *testing.T) {
	doc := NewStringDocument("")
	assert.Equal(t, 0, doc.Length())
	assert.Equal(t, "", doc.String())
}

func TestDocument_Interface_Slice_UTF8(t *testing.T) {
	doc := NewStringDocument("你好世界")
	// Slice uses byte positions, not character positions
	// "你好" = bytes 0-5, "世界" = bytes 6-11
	assert.Equal(t, "你好", doc.Slice(0, 6))
	assert.Equal(t, "世界", doc.Slice(6, 12))
}

// ========== Document Interface Compliance ==========

func TestDocument_Interface_Compliance(t *testing.T) {
	// This test ensures that StringDocument implements the Document interface
	var _ Document = (*StringDocument)(nil)

	doc := NewStringDocument("Test")
	assert.Implements(t, (*Document)(nil), doc)
}
