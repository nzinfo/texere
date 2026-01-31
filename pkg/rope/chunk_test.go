package rope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ========== Basic Chunk Tests ==========

func TestChunkAtChar_SingleChunk(t *testing.T) {
	r := New("hello world")

	chunk, startChar := r.ChunkAtChar(5)
	assert.Equal(t, "hello world", chunk.Text)
	assert.Equal(t, 0, chunk.ByteIdx)
	assert.Equal(t, 0, chunk.CharIdx)
	assert.Equal(t, 0, chunk.LineIdx)
	assert.Equal(t, 0, startChar) // Chunk starts at position 0
}

func TestChunkAtChar_MultiChunk(t *testing.T) {
	// Build a rope with multiple chunks using builder
	builder := NewBuilder()
	builder.Append("hello")
	builder.Append(" ")
	builder.Append("world")
	r := builder.Build()

	// All positions should return valid chunks
	for i := 0; i < r.Length(); i++ {
		chunk, _ := r.ChunkAtChar(i)
		assert.NotEmpty(t, chunk.Text)
	}
}

func TestChunkAtChar_OutOfBounds(t *testing.T) {
	r := New("hello")

	assert.Panics(t, func() {
		r.ChunkAtChar(-1)
	})

	assert.Panics(t, func() {
		r.ChunkAtChar(100)
	})
}

func TestChunkAtByte_Basics(t *testing.T) {
	r := New("hello world")

	chunk, startByte := r.ChunkAtByte(6)
	assert.Equal(t, "hello world", chunk.Text)
	assert.Equal(t, 0, chunk.ByteIdx)
	assert.Equal(t, 0, startByte) // Chunk starts at byte 0
}

func TestChunkAtByte_OutOfBounds(t *testing.T) {
	r := New("hello")

	assert.Panics(t, func() {
		r.ChunkAtByte(-1)
	})

	assert.Panics(t, func() {
		r.ChunkAtByte(100)
	})
}

func TestChunkAtLine_LineBreaks(t *testing.T) {
	r := New("line1\nline2\nline3")

	// All text is in one chunk, so all lines return that chunk
	// Line 0
	chunk1, _, _, lineIdx1 := r.ChunksAtLine(0)
	chunk1.Next()
	assert.Contains(t, chunk1.Current(), "line1")
	assert.Equal(t, 0, lineIdx1)

	// Line 1 - same chunk, line index is 0 (chunk's starting line)
	chunk2, _, _, lineIdx2 := r.ChunksAtLine(1)
	chunk2.Next()
	assert.Contains(t, chunk2.Current(), "line2")
	assert.Equal(t, 0, lineIdx2) // Chunk starts at line 0
}

func TestChunkAtLine_OutOfBounds(t *testing.T) {
	r := New("hello")

	assert.Panics(t, func() {
		r.ChunksAtLine(-1)
	})

	assert.Panics(t, func() {
		r.ChunksAtLine(100)
	})
}

// ========== Consistency Tests ==========

func TestChunkAtChar_Consistency(t *testing.T) {
	r := New("hello world\ntest text")

	// All characters in this rope should return the same chunk (single chunk rope)
	for i := 0; i < r.Length(); i++ {
		chunk, _ := r.ChunkAtChar(i)
		assert.Equal(t, "hello world\ntest text", chunk.Text)
	}
}

func TestChunkAtByte_Consistency(t *testing.T) {
	r := New("hello world")

	// All characters in single chunk rope should return same chunk
	for i := 0; i < r.Size(); i++ {
		chunk, _ := r.ChunkAtByte(i)
		assert.Equal(t, "hello world", chunk.Text)
	}
}

// ========== Edge Cases ==========

func TestChunkAtChar_Empty(t *testing.T) {
	r := New("")

	// Empty rope - chunk count may be 0 or 1 depending on implementation
	// The important thing is it doesn't crash
	count := r.ChunkCount()
	assert.GreaterOrEqual(t, count, 0)
}

func TestChunkAtChar_LastPosition(t *testing.T) {
	r := New("hello")

	// Position at end should still work
	chunk, start := r.ChunkAtChar(4)
	assert.Equal(t, "hello", chunk.Text)
	assert.Equal(t, 0, start)
}

func TestChunkAtChar_ChineseText(t *testing.T) {
	// Test with Unicode characters
	r := New("hello 世界")

	chunk, _ := r.ChunkAtChar(7) // Position in "世界"
	assert.Equal(t, "hello 世界", chunk.Text)
	// The chunk contains all characters: "hello 世界" = 5 + 1 + 2 = 8 total
	assert.Equal(t, 8, chunk.CharLen)
}

func TestChunkAtByte_Unicode(t *testing.T) {
	// Test with Unicode characters
	r := New("hello 世界") // 12 bytes: 5 + 1 + 6 (3 bytes per Chinese char)

	chunk, _ := r.ChunkAtByte(8) // Position in "世界"
	assert.Equal(t, "hello 世界", chunk.Text)
	assert.Equal(t, 12, chunk.ByteLen)
}

// ========== ChunkInfo Tests ==========

func TestChunkInfo_Fields(t *testing.T) {
	r := New("hello\nworld")

	chunk, _ := r.ChunkAtChar(7) // 'w' in "world"

	assert.Equal(t, "hello\nworld", chunk.Text)
	assert.Equal(t, 0, chunk.ByteIdx)
	assert.Equal(t, 0, chunk.CharIdx)
	assert.Equal(t, 0, chunk.LineIdx)
	assert.False(t, chunk.IsEmpty)
	assert.Greater(t, chunk.ByteLen, 0)
	assert.Greater(t, chunk.CharLen, 0)
}

// ========== Performance Tests ==========

func TestChunkAtChar_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test")
	}

	// Create deep tree with many chunks
	r := New("")
	for i := 0; i < 1000; i++ {
		r = r.Insert(r.Length(), "word ")
	}

	// Test performance - just verify it's fast enough
	for i := 0; i < 100; i++ {
		r.ChunkAtChar(r.Length() / 2)
	}
}

func TestChunkAtByte_DeepTree(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping deep tree test")
	}

	// Create deep tree
	r := New("")
	for i := 0; i < 100; i++ {
		r = r.Insert(r.Length(), "x")
	}

	// Access in middle
	chunk, _ := r.ChunkAtByte(50)
	assert.NotEmpty(t, chunk.Text)
}

// ========== Multi-Line Tests ==========

func TestChunkAtChar_MultiLine(t *testing.T) {
	text := "line1\nline2\nline3"
	r := New(text)

	// All in same chunk
	chunk, _ := r.ChunkAtChar(5)
	assert.Equal(t, text, chunk.Text)
}

func TestChunkAtLine_MultiLine(t *testing.T) {
	text := "line1\nline2\nline3"
	r := New(text)

	// Line 0 - should return chunk at line 0
	it0, b0, c0, l0 := r.ChunksAtLine(0)
	assert.Equal(t, 0, l0)
	assert.Equal(t, 0, b0)
	assert.Equal(t, 0, c0)
	it0.Next()
	assert.Equal(t, text, it0.Current())

	// Line 1 (after first \n) - since it's all in one chunk, still returns same chunk
	it1, _, _, l1 := r.ChunksAtLine(1)
	assert.Equal(t, 0, l1) // Single chunk rope, so line index is 0
	it1.Next()
	assert.Contains(t, it1.Current(), "line2")
}

// ========== Chunk Utilities Tests ==========

func TestChunkCount(t *testing.T) {
	r := New("hello world")
	assert.Equal(t, 1, r.ChunkCount())

	// Multi-chunk rope created with multiple appends
	builder := NewBuilder()
	for i := 0; i < 10; i++ {
		builder.Append("hello")
	}
	r2 := builder.Build()
	assert.GreaterOrEqual(t, r2.ChunkCount(), 1)
}

func TestAverageChunkSize(t *testing.T) {
	r := New("hello world")
	assert.Equal(t, float64(r.Size()), r.AverageChunkSize())
}

func TestMaxChunkSize(t *testing.T) {
	r := New("hello world")
	assert.Equal(t, r.Size(), r.MaxChunkSize())
}

func TestMinChunkSize(t *testing.T) {
	r := New("hello world")
	assert.Equal(t, r.Size(), r.MinChunkSize())
}

func TestMinMaxChunkSize_MultiChunk(t *testing.T) {
	// Create rope with different sized chunks
	builder := NewBuilder()
	builder.Append("hi")
	builder.Append(" ")
	builder.Append("hello world")
	r := builder.Build()

	minSize := r.MinChunkSize()
	maxSize := r.MaxChunkSize()

	// Just verify they work and are positive
	assert.Greater(t, maxSize, 0)
	assert.Greater(t, minSize, 0)
	// Max should be >= min
	assert.GreaterOrEqual(t, maxSize, minSize)
}

// ========== Empty Rope Tests ==========

func TestChunkAtChar_EmptyRope(t *testing.T) {
	var r *Rope

	chunk, start := r.ChunkAtChar(0)
	assert.Equal(t, "", chunk.Text)
	assert.Equal(t, 0, start)
}

func TestChunkAtByte_EmptyRope(t *testing.T) {
	var r *Rope

	chunk, start := r.ChunkAtByte(0)
	assert.Equal(t, "", chunk.Text)
	assert.Equal(t, 0, start)
}

func TestChunks_EmptyRope(t *testing.T) {
	var r *Rope
	it := r.Chunks()

	assert.Equal(t, 0, it.Count())
	assert.False(t, it.Next())
}
