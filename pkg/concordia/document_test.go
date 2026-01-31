package concordia

import (
	"testing"

	"github.com/coreseekdev/texere/pkg/ot"
	"github.com/stretchr/testify/assert"
)

// This file contains tests for the ot.Document interface.
// We test it through the ot.StringDocument implementation from the ot package.

// ========== Document Interface Tests ==========

func TestDocument_Interface_Length(t *testing.T) {
	doc := ot.NewStringDocument("Hello World")
	assert.Equal(t, 11, doc.Length())
}

func TestDocument_Interface_Slice(t *testing.T) {
	doc := ot.NewStringDocument("Hello World")
	assert.Equal(t, "Hello", doc.Slice(0, 5))
	assert.Equal(t, "World", doc.Slice(6, 11))
}

func TestDocument_Interface_String(t *testing.T) {
	doc := ot.NewStringDocument("Hello World")
	assert.Equal(t, "Hello World", doc.String())
}

func TestDocument_Interface_Bytes(t *testing.T) {
	doc := ot.NewStringDocument("Hello World")
	assert.Equal(t, []byte("Hello World"), doc.Bytes())
}

func TestDocument_Interface_Clone(t *testing.T) {
	doc := ot.NewStringDocument("Hello World")
	doc2 := doc.Clone()

	assert.Equal(t, doc.String(), doc2.String())
	assert.NotSame(t, doc, doc2)
}

func TestDocument_Interface_UTF8(t *testing.T) {
	doc := ot.NewStringDocument("Hello 世界")
	// Length returns bytes, not characters
	// "Hello 世界" = 5 + 1 + 2*3 = 12 bytes (每个中文字符3字节)
	assert.Equal(t, 12, doc.Length())
	assert.Equal(t, "Hello 世界", doc.String())
}

func TestDocument_Interface_Empty(t *testing.T) {
	doc := ot.NewStringDocument("")
	assert.Equal(t, 0, doc.Length())
	assert.Equal(t, "", doc.String())
}

func TestDocument_Interface_Slice_UTF8(t *testing.T) {
	doc := ot.NewStringDocument("你好世界")
	// Slice uses byte positions, not character positions
	// "你好" = bytes 0-5, "世界" = bytes 6-11
	assert.Equal(t, "你好", doc.Slice(0, 6))
	assert.Equal(t, "世界", doc.Slice(6, 12))
}

// ========== Document Interface Compliance ==========

func TestDocument_Interface_Compliance(t *testing.T) {
	// This test ensures that ot.StringDocument implements the ot.Document interface
	var _ ot.Document = (*ot.StringDocument)(nil)

	doc := ot.NewStringDocument("Test")
	assert.Implements(t, (*ot.Document)(nil), doc)
}

// ========== RopeDocument Tests ==========

func TestRopeDocument_Basic(t *testing.T) {
	doc := NewRopeDocument("Hello World")

	assert.Equal(t, 11, doc.Length())
	assert.Equal(t, "Hello World", doc.String())
	assert.Equal(t, []byte("Hello World"), doc.Bytes())
}

func TestRopeDocument_Slice(t *testing.T) {
	doc := NewRopeDocument("Hello World")

	assert.Equal(t, "Hello", doc.Slice(0, 5))
	assert.Equal(t, "World", doc.Slice(6, 11))
}

func TestRopeDocument_Clone(t *testing.T) {
	doc := NewRopeDocument("Hello World")
	doc2 := doc.Clone().(*RopeDocument)

	assert.Equal(t, doc.String(), doc2.String())
}

func TestRopeDocument_Insert(t *testing.T) {
	doc := NewRopeDocument("World")
	doc2 := doc.Insert(0, "Hello ")

	assert.Equal(t, "Hello World", doc2.String())
	assert.Equal(t, "World", doc.String())
}

func TestRopeDocument_Delete(t *testing.T) {
	doc := NewRopeDocument("Hello World")
	doc2 := doc.Delete(5, 11)

	assert.Equal(t, "Hello", doc2.String())
}

func TestRopeDocument_Concat(t *testing.T) {
	doc1 := NewRopeDocument("Hello")
	doc2 := NewRopeDocument(" World")
	doc3 := doc1.Concat(doc2)

	assert.Equal(t, "Hello World", doc3.String())
}

func TestRopeDocument_Equals(t *testing.T) {
	doc1 := NewRopeDocument("Hello")
	doc2 := NewRopeDocument("Hello")
	doc3 := NewRopeDocument("World")

	assert.True(t, doc1.Equals(doc2))
	assert.False(t, doc1.Equals(doc3))
}

// ========== DocumentBuilder Tests ==========

func TestDocumentBuilder_Basic(t *testing.T) {
	b := NewDocumentBuilder()
	b.Append("Hello")
	b.Append(" ")
	b.Append("World")

	doc := b.Build()
	assert.Equal(t, "Hello World", doc.String())
}

func TestDocumentBuilder_AppendLine(t *testing.T) {
	b := NewDocumentBuilder()
	b.AppendLine("Line1")
	b.AppendLine("Line2")

	doc := b.Build()
	assert.Equal(t, "Line1\nLine2\n", doc.String())
}

func TestDocumentBuilder_Reset(t *testing.T) {
	b := NewDocumentBuilder()
	b.Append("Hello")
	b.Reset()
	b.Append("World")

	doc := b.Build()
	assert.Equal(t, "World", doc.String())
}
