package rope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHash_Consistency verifies that ropes with same content
// but different chunk boundaries produce the same hash
// This is ported from ropey's hash.rs
func TestHash_Consistency_Small(t *testing.T) {
	// Build two ropes with the same contents but different chunk boundaries
	r1 := New("")
	b1 := NewBuilder()
	b1.Append("Hello w")
	b1.Append("orld")
	r1 = b1.Build()

	r2 := New("")
	b2 := NewBuilder()
	b2.Append("Hell")
	b2.Append("o world")
	r2 = b2.Build()

	// Should have same hash
	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.Equal(t, hash1, hash2)
	assert.Equal(t, r1.String(), r2.String())
}

// TestHash_Consistency_Medium tests hash consistency with larger text
func TestHash_Consistency_Medium(t *testing.T) {
	text := "Hello World! This is a test string for hashing. " +
		"It should produce the same hash regardless of chunk boundaries. " +
		"The quick brown fox jumps over the lazy dog. " +
		"こんにちは世界 🌍"

	// Build rope with 5-byte chunks
	r1 := New("")
	b1 := NewBuilder()
	for i := 0; i < len(text); i += 5 {
		end := i + 5
		if end > len(text) {
			end = len(text)
		}
		b1.Append(text[i:end])
	}
	r1 = b1.Build()

	// Build rope with 7-byte chunks
	r2 := New("")
	b2 := NewBuilder()
	for i := 0; i < len(text); i += 7 {
		end := i + 7
		if end > len(text) {
			end = len(text)
		}
		b2.Append(text[i:end])
	}
	r2 = b2.Build()

	// Should have same hash
	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.Equal(t, hash1, hash2)
	assert.Equal(t, r1.String(), r2.String())
}

// TestHash_Consistency_Large tests hash consistency with large text
func TestHash_Consistency_Large(t *testing.T) {
	text := ""
	for i := 0; i < 100; i++ {
		text += "Hello World! " +
			"The quick brown fox jumps over the lazy dog. " +
			"こんにちは世界 🌍🌎🌏\n"
	}

	// Build rope with 521-byte chunks
	r1 := New("")
	b1 := NewBuilder()
	for i := 0; i < len(text); i += 521 {
		end := i + 521
		if end > len(text) {
			end = len(text)
		}
		b1.Append(text[i:end])
	}
	r1 = b1.Build()

	// Build rope with 547-byte chunks
	r2 := New("")
	b2 := NewBuilder()
	for i := 0; i < len(text); i += 547 {
		end := i + 547
		if end > len(text) {
			end = len(text)
		}
		b2.Append(text[i:end])
	}
	r2 = b2.Build()

	// Should have same hash
	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.Equal(t, hash1, hash2)
	assert.Equal(t, r1.String(), r2.String())
}

// TestHash_DifferentContent produces different hashes
func TestHash_DifferentContent(t *testing.T) {
	r1 := New("Hello World")
	r2 := New("Hello World!")

	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.NotEqual(t, hash1, hash2)
}

// TestHash_EmptyRope produces consistent hash for empty rope
func TestHash_EmptyRope(t *testing.T) {
	r1 := Empty()
	r2 := New("")

	// Empty ropes should have same hash
	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.Equal(t, hash1, hash2)
}

// TestHash_HashCode32 produces consistent 32-bit hash
func TestHash_HashCode32(t *testing.T) {
	r1 := New("Hello World")
	r2 := New("Hello World")

	hash1 := r1.HashCode32()
	hash2 := r2.HashCode32()

	assert.Equal(t, hash1, hash2)
}

// TestHash_HashEquals verifies HashEquals method
func TestHash_HashEquals(t *testing.T) {
	r1 := New("Hello World")
	r2 := New("Hello World")
	r3 := New("Hello World!")

	assert.True(t, r1.HashEquals(r2))
	assert.False(t, r1.HashEquals(r3))
}

// TestHash_SingleInsert verifies hash changes after insert
func TestHash_SingleInsert(t *testing.T) {
	r1 := New("Hello World")
	hash1 := r1.HashCode64()

	r2 := r1.Insert(5, "XXX")
	hash2 := r2.HashCode64()

	assert.NotEqual(t, hash1, hash2)
}

// TestHash_Delete verifies hash changes after delete
func TestHash_Delete(t *testing.T) {
	r1 := New("Hello World")
	hash1 := r1.HashCode64()

	r2 := r1.Delete(5, 6)
	hash2 := r2.HashCode64()

	assert.NotEqual(t, hash1, hash2)
}

// TestHash_SplitMerge verifies hash consistency after split/merge
func TestHash_SplitMerge(t *testing.T) {
	text := "Hello World Test String"
	r := New(text)
	hash1 := r.HashCode64()

	left, right := r.Split(6)
	merged := left.AppendRope(right)
	hash2 := merged.HashCode64()

	assert.Equal(t, hash1, hash2)
	assert.Equal(t, text, merged.String())
}

// TestHash_ChunkHashes returns hashes of all chunks
func TestHash_ChunkHashes(t *testing.T) {
	r1 := New("Hello")
	r2 := r1.Append(" World")

	hashes := r2.ChunkHashes()

	// Should have at least 2 chunks
	assert.True(t, len(hashes) >= 2)

	// Hashes should be non-zero
	for _, h := range hashes {
		assert.NotEqual(t, uint32(0), h)
	}
}

// TestHash_CombinedChunkHash returns combined hash
func TestHash_CombinedChunkHash(t *testing.T) {
	r := New("Hello World")
	r = r.Append(" Test")

	hash := r.CombinedChunkHash()

	assert.NotEqual(t, uint32(0), hash)
}

// TestHash_Unicode produces consistent hash for unicode
func TestHash_Unicode(t *testing.T) {
	text := "Hello 世界 🌍"

	r1 := New(text)
	r2 := New(text)

	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.Equal(t, hash1, hash2)
}

// TestHash_CRLF produces consistent hash with CRLF
func TestHash_CRLF(t *testing.T) {
	text := "Line 1\r\nLine 2\r\nLine 3"

	r1 := New(text)
	r2 := New(text)

	hash1 := r1.HashCode64()
	hash2 := r2.HashCode64()

	assert.Equal(t, hash1, hash2)
}

// TestHash_Integrity verifies hash doesn't change for same rope
func TestHash_Integrity(t *testing.T) {
	r := New("Hello World Test")

	hash1 := r.HashCode64()
	hash2 := r.HashCode64()
	hash3 := r.HashCode64()

	// Hash should be stable
	assert.Equal(t, hash1, hash2)
	assert.Equal(t, hash2, hash3)
}

// TestHash_HashString tests HashString method
func TestHash_HashString(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		notEmpty bool
	}{
		{
			name:     "ASCII text",
			text:     "Hello World",
			notEmpty: true,
		},
		{
			name:     "Empty rope",
			text:     "",
			notEmpty: false,
		},
		{
			name:     "Unicode text",
			text:     "你好世界",
			notEmpty: true,
		},
		{
			name:     "Emoji",
			text:     "Hello🌍",
			notEmpty: true,
		},
		{
			name:     "Numbers",
			text:     "12345",
			notEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			hashStr := r.HashString()

			if tt.notEmpty {
				assert.NotEqual(t, "0", hashStr)
				assert.NotEqual(t, "", hashStr)
				// Should be 8 characters (hex string)
				assert.Equal(t, 8, len(hashStr))
			} else {
				assert.Equal(t, "0", hashStr)
			}
		})
	}
}

// TestHash_HashCode32 tests 32-bit hash code
func TestHash_HashCode32_Detailed(t *testing.T) {
	tests := []struct {
		name     string
		text     string
	}{
		{
			name: "Simple ASCII",
			text: "Hello",
		},
		{
			name: "With spaces",
			text: "Hello World",
		},
		{
			name: "Unicode",
			text: "Hello世界",
		},
		{
			name: "Empty",
			text: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			hash := r.HashCode32()

			if tt.text == "" {
				assert.Equal(t, uint32(0), hash)
			} else {
				assert.NotEqual(t, uint32(0), hash)
			}
		})
	}
}

// TestHash_CombineHash tests hash combining
func TestHash_CombineHash(t *testing.T) {
	tests := []struct {
		name     string
		codes    []uint32
		expected uint32
	}{
		{
			name:     "Empty list",
			codes:    []uint32{},
			expected: 0,
		},
		{
			name:     "Single value",
			codes:    []uint32{42},
			expected: 42,
		},
		{
			name:     "Multiple values",
			codes:    []uint32{1, 2, 3},
			expected: (1*31 + 2)*31 + 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CombineHash(tt.codes...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHash_CombineHash64 tests 64-bit hash combining
func TestHash_CombineHash64(t *testing.T) {
	tests := []struct {
		name     string
		codes    []uint64
		expected uint64
	}{
		{
			name:     "Empty list",
			codes:    []uint64{},
			expected: 0,
		},
		{
			name:     "Single value",
			codes:    []uint64{42},
			expected: 42,
		},
		{
			name:     "Multiple values",
			codes:    []uint64{1, 2, 3},
			expected: (1*31 + 2)*31 + 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CombineHash64(tt.codes...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHash_HashBytes tests HashBytes helper
func TestHash_HashBytes(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Simple data",
			data: []byte("Hello"),
		},
		{
			name: "Empty data",
			data: []byte{},
		},
		{
			name: "Binary data",
			data: []byte{0x00, 0x01, 0x02, 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashBytes(tt.data)
			// FNV hash always returns a non-zero value, even for empty input
			assert.NotEqual(t, uint32(0), hash)
		})
	}
}

// TestHash_HashStringFunc tests HashString helper function
func TestHash_HashStringFunc(t *testing.T) {
	tests := []struct {
		name string
		s    string
	}{
		{
			name: "ASCII string",
			s:    "Hello World",
		},
		{
			name: "Empty string",
			s:    "",
		},
		{
			name: "Unicode string",
			s:    "你好世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashString(tt.s)
			// FNV hash always returns a non-zero value, even for empty input
			assert.NotEqual(t, uint32(0), hash)
		})
	}
}

// TestHash_HashRunes tests HashRunes helper
func TestHash_HashRunes(t *testing.T) {
	tests := []struct {
		name  string
		runes []rune
	}{
		{
			name:  "ASCII runes",
			runes: []rune{'H', 'e', 'l', 'l', 'o'},
		},
		{
			name:  "Empty runes",
			runes: []rune{},
		},
		{
			name:  "Unicode runes",
			runes: []rune{'你', '好', '世', '界'},
		},
		{
			name:  "Mixed runes",
			runes: []rune{'H', 'i', '你'},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashRunes(tt.runes)
			// FNV hash always returns a non-zero value, even for empty input
			assert.NotEqual(t, uint32(0), hash)
		})
	}
}

// TestHash_LikelyEquals tests LikelyEquals method
func TestHash_LikelyEquals(t *testing.T) {
	tests := []struct {
		name     string
		text1    string
		text2    string
		expected bool
	}{
		{
			name:     "Same content",
			text1:    "Hello World",
			text2:    "Hello World",
			expected: true,
		},
		{
			name:     "Different content",
			text1:    "Hello World",
			text2:    "Hello World!",
			expected: false,
		},
		{
			name:     "Both empty",
			text1:    "",
			text2:    "",
			expected: true,
		},
		{
			name:     "Different length",
			text1:    "Hi",
			text2:    "Hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r1 := New(tt.text1)
			r2 := New(tt.text2)
			result := r1.LikelyEquals(r2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHash_HashKey tests HashKey method
func TestHash_HashKey(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "Simple text",
			text: "Hello",
		},
		{
			name: "Empty text",
			text: "",
		},
		{
			name: "Unicode text",
			text: "你好",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			key := r.HashKey()

			// HashKey should be same as HashCode
			assert.Equal(t, r.HashCode(), key)
		})
	}
}

// TestHash_HashSlice tests HashSlice helper
func TestHash_HashSlice(t *testing.T) {
	r1 := New("Hello")
	r2 := New("World")
	r3 := New("Test")

	slice := []*Rope{r1, r2, r3}
	hashes := HashSlice(slice)

	assert.Equal(t, 3, len(hashes))
	assert.NotEqual(t, uint32(0), hashes[0])
	assert.NotEqual(t, uint32(0), hashes[1])
	assert.NotEqual(t, uint32(0), hashes[2])
}

// TestHash_IncrementalHasher tests IncrementalHasher
func TestHash_IncrementalHasher(t *testing.T) {
	tests := []struct {
		name     string
		baseHash uint32
		strings  []string
	}{
		{
			name:     "Add multiple strings",
			baseHash: 100,
			strings:  []string{"Hello", "World", "Test"},
		},
		{
			name:     "Add to zero",
			baseHash: 0,
			strings:  []string{"A", "B"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ih := NewIncrementalHasher(tt.baseHash)

			for _, s := range tt.strings {
				ih.AddString(s)
			}

			result := ih.Current()
			assert.NotEqual(t, tt.baseHash, result)
		})
	}
}

// TestHash_IncrementalHasher_AddRope tests adding ropes
func TestHash_IncrementalHasher_AddRope(t *testing.T) {
	ih := NewIncrementalHasher(0)

	r1 := New("Hello")
	r2 := New("World")

	ih.AddRope(r1)
	hash1 := ih.Current()

	ih.AddRope(r2)
	hash2 := ih.Current()

	assert.NotEqual(t, uint32(0), hash1)
	assert.NotEqual(t, hash1, hash2)
}

// TestHash_IncrementalHasher_Reset tests resetting
func TestHash_IncrementalHasher_Reset(t *testing.T) {
	ih := NewIncrementalHasher(100)
	ih.AddString("Hello")

	assert.NotEqual(t, uint32(100), ih.Current())

	ih.Reset()
	assert.Equal(t, uint32(0), ih.Current())
}

// TestHash_RollingHasher tests RollingHasher
func TestHash_RollingHasher(t *testing.T) {
	r := New("Hello World")

	tests := []struct {
		name      string
		windowSize int
	}{
		{
			name:      "Small window",
			windowSize: 3,
		},
		{
			name:      "Full window",
			windowSize: r.Length(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rh := r.NewRollingHasher(tt.windowSize)
			hash := rh.Current()

			assert.NotEqual(t, uint32(0), hash)
		})
	}
}

// TestHash_NilRope tests nil rope handling
func TestHash_NilRope(t *testing.T) {
	var r *Rope

	assert.Equal(t, uint32(0), r.HashCode())
	assert.Equal(t, uint32(0), r.HashCode32())
	assert.Equal(t, uint64(0), r.HashCode64())
	assert.Equal(t, "0", r.HashString())
	assert.Equal(t, uint32(0), r.HashKey())
}
