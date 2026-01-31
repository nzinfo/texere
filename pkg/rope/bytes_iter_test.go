package rope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBytesIterator_NewIterator tests creating a new bytes iterator
func TestBytesIterator_NewIterator(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		expectEmpty bool
	}{
		{
			name:        "Non-empty rope",
			text:        "Hello World",
			expectEmpty: false,
		},
		{
			name:        "Empty rope",
			text:        "",
			expectEmpty: true,
		},
		{
			name:        "Unicode text",
			text:        "Hello世界",
			expectEmpty: false,
		},
		{
			name:        "Emoji text",
			text:        "Hi🌍🌎",
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			it := r.NewBytesIterator()

			assert.Equal(t, tt.expectEmpty, it.IsExhausted())
			if !tt.expectEmpty {
				assert.Equal(t, -1, it.Position())
			}
		})
	}
}

// TestBytesIterator_NilRope tests iterator behavior with nil rope
func TestBytesIterator_NilRope(t *testing.T) {
	var r *Rope
	it := r.NewBytesIterator()

	assert.True(t, it.IsExhausted())
	assert.False(t, it.Next())
	assert.Panics(t, func() { it.Current() })
}

// TestBytesIterator_BasicIteration tests basic iteration
func TestBytesIterator_BasicIteration(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		bytes []byte
	}{
		{
			name:  "ASCII text",
			text:  "Hello",
			bytes: []byte{72, 101, 108, 108, 111},
		},
		{
			name:  "Empty string",
			text:  "",
			bytes: []byte{},
		},
		{
			name:  "Single character",
			text:  "A",
			bytes: []byte{65},
		},
		{
			name:  "Unicode Chinese",
			text:  "你好",
			bytes: []byte{228, 189, 160, 229, 165, 189},
		},
		{
			name:  "Emoji",
			text:  "🌍",
			bytes: []byte{240, 159, 140, 141},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			it := r.NewBytesIterator()

			collected := []byte{}
			for it.Next() {
				collected = append(collected, it.Current())
			}

			assert.Equal(t, tt.bytes, collected)
			assert.True(t, it.IsExhausted())
		})
	}
}

// TestBytesIterator_CurrentPanic tests that Current() panics appropriately
func TestBytesIterator_CurrentPanic(t *testing.T) {
	t.Run("Before first Next", func(t *testing.T) {
		r := New("Hello")
		it := r.NewBytesIterator()

		assert.Panics(t, func() { it.Current() })
	})

	t.Run("After exhaustion", func(t *testing.T) {
		r := New("Hi")
		it := r.NewBytesIterator()
		it.Next()
		it.Next()
		it.Next() // Now exhausted

		assert.Panics(t, func() { it.Current() })
	})
}

// TestBytesIterator_Position tests position tracking
func TestBytesIterator_Position(t *testing.T) {
	r := New("Hello")
	it := r.NewBytesIterator()

	assert.Equal(t, -1, it.Position())

	it.Next()
	assert.Equal(t, 0, it.Position())

	it.Next()
	assert.Equal(t, 1, it.Position())

	it.Next()
	it.Next()
	it.Next()
	assert.Equal(t, 4, it.Position())
}

// TestBytesIterator_BytePosition tests BytePosition alias
func TestBytesIterator_BytePosition(t *testing.T) {
	r := New("World")
	it := r.NewBytesIterator()

	it.Next()
	it.Next()

	assert.Equal(t, it.Position(), it.BytePosition())
	assert.Equal(t, 1, it.BytePosition())
}

// TestBytesIterator_HasNext tests HasNext method
func TestBytesIterator_HasNext(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		nextCount     int
		expectHasNext bool
	}{
		{
			name:          "Has next after first",
			text:          "Hello",
			nextCount:     1,
			expectHasNext: true,
		},
		{
			name:          "No next at end",
			text:          "Hi",
			nextCount:     2,
			expectHasNext: false,
		},
		{
			name:          "Empty rope",
			text:          "",
			nextCount:     0,
			expectHasNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			it := r.NewBytesIterator()

			for i := 0; i < tt.nextCount; i++ {
				it.Next()
			}

			assert.Equal(t, tt.expectHasNext, it.HasNext())
		})
	}
}

// TestBytesIterator_Reset tests iterator reset
func TestBytesIterator_Reset(t *testing.T) {
	r := New("Hello")
	it := r.NewBytesIterator()

	// Consume some bytes
	it.Next()
	it.Next()
	assert.Equal(t, 1, it.Position())

	// Reset
	it.Reset()
	assert.Equal(t, -1, it.Position())
	assert.False(t, it.IsExhausted())

	// Iterate again
	it.Next()
	assert.Equal(t, 0, it.Position())
	assert.Equal(t, byte('H'), it.Current())
}

// TestBytesIterator_Collect tests Collect method
func TestBytesIterator_Collect(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		bytes []byte
	}{
		{
			name:  "ASCII",
			text:  "Hello",
			bytes: []byte("Hello"),
		},
		{
			name:  "Unicode",
			text:  "Hi世界",
			bytes: []byte("Hi世界"),
		},
		{
			name:  "Empty",
			text:  "",
			bytes: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			it := r.NewBytesIterator()

			collected := it.Collect()
			assert.Equal(t, tt.bytes, collected)
		})
	}
}

// TestBytesIterator_ToBytes tests ToBytes alias
func TestBytesIterator_ToBytes(t *testing.T) {
	r := New("Test")
	it := r.NewBytesIterator()

	assert.Equal(t, it.Collect(), it.ToBytes())
}

// TestBytesIterator_Skip tests Skip method
func TestBytesIterator_Skip(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		skipCount     int
		expectedByte  byte
		expectedHasNext bool
	}{
		{
			name:          "Skip 2 bytes",
			text:          "Hello",
			skipCount:     2,
			expectedByte:  'l',
			expectedHasNext: true,
		},
		{
			name:          "Skip all",
			text:          "Hi",
			skipCount:     2,
			expectedByte:  0,
			expectedHasNext: false,
		},
		{
			name:          "Skip zero",
			text:          "Test",
			skipCount:     0,
			expectedByte:  'T',
			expectedHasNext: true,
		},
		{
			name:          "Skip negative",
			text:          "Hello",
			skipCount:     -1,
			expectedByte:  'H',
			expectedHasNext: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			it := r.NewBytesIterator()

			it.Skip(tt.skipCount)

			if tt.expectedHasNext {
				assert.True(t, it.Next())
				assert.Equal(t, tt.expectedByte, it.Current())
			} else {
				assert.False(t, it.Next())
			}
		})
	}
}

// TestBytesIterator_Peek tests Peek method
func TestBytesIterator_Peek(t *testing.T) {
	r := New("Hello")
	it := r.NewBytesIterator()

	it.Next() // Position at 'H' (byte 0)

	// Peek at 'e' (byte 1)
	nextByte := it.Peek()
	assert.Equal(t, byte('e'), nextByte)

	// Position should not have changed
	it.Next()
	assert.Equal(t, byte('e'), it.Current())

	// Peek at last byte - advance to 'o'
	for i := 0; i < 3; i++ {
		it.Next()
	}
	assert.Equal(t, byte('o'), it.Current())

	// Peek beyond should panic
	assert.Panics(t, func() { it.Peek() })
}

// TestBytesIterator_HasPeek tests HasPeek method
func TestBytesIterator_HasPeek(t *testing.T) {
	r := New("Hi")
	it := r.NewBytesIterator()

	it.Next()
	assert.True(t, it.HasPeek())

	it.Next()
	assert.False(t, it.HasPeek())
}

// TestBytesIterator_AtPosition tests BytesIteratorAt
func TestBytesIterator_AtPosition(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		startPos     int
		expectPanic  bool
		expectedByte byte
	}{
		{
			name:         "Start at position 2",
			text:         "Hello",
			startPos:     2,
			expectPanic:  false,
			expectedByte: 'l',
		},
		{
			name:         "Start at beginning",
			text:         "Hello",
			startPos:     0,
			expectPanic:  false,
			expectedByte: 'H',
		},
		{
			name:         "Start at end",
			text:         "Hello",
			startPos:     5,
			expectPanic:  false,
		},
		{
			name:         "Negative position",
			text:         "Hello",
			startPos:     -1,
			expectPanic:  true,
		},
		{
			name:         "Beyond end",
			text:         "Hello",
			startPos:     10,
			expectPanic:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)

			if tt.expectPanic {
				assert.Panics(t, func() { r.BytesIteratorAt(tt.startPos) })
			} else {
				it := r.BytesIteratorAt(tt.startPos)
				if tt.startPos < r.Size() {
					it.Next()
					assert.Equal(t, tt.expectedByte, it.Current())
				} else {
					assert.True(t, it.IsExhausted())
				}
			}
		})
	}
}

// TestBytesIterator_IterBytesAt tests IterBytesAt alias
func TestBytesIterator_IterBytesAt(t *testing.T) {
	r := New("Hello")
	it1 := r.BytesIteratorAt(2)
	it2 := r.IterBytesAt(2)

	it1.Next()
	it2.Next()

	assert.Equal(t, it1.Current(), it2.Current())
}

// TestBytesIterator_Seek tests Seek method
func TestBytesIterator_Seek(t *testing.T) {
	r := New("Hello World")
	it := r.NewBytesIterator()

	// Seek to position 6 ('W')
	success := it.Seek(6)
	assert.True(t, success)
	assert.False(t, it.IsExhausted())

	it.Next()
	assert.Equal(t, byte('W'), it.Current())
	assert.Equal(t, 6, it.Position())

	// Seek to invalid position
	success = it.Seek(100)
	assert.False(t, success)
	assert.True(t, it.IsExhausted())

	// Seek to negative
	success = it.Seek(-1)
	assert.False(t, success)
}

// TestBytesIterator_ForEachByte tests ForEachByte
func TestBytesIterator_ForEachByte(t *testing.T) {
	r := New("Hello")

	count := 0
	r.ForEachByte(func(b byte) bool {
		count++
		return true
	})

	assert.Equal(t, 5, count)
}

// TestBytesIterator_ForEachByteEarlyExit tests early exit from ForEachByte
func TestBytesIterator_ForEachByteEarlyExit(t *testing.T) {
	r := New("Hello")

	count := 0
	result := r.ForEachByte(func(b byte) bool {
		count++
		return count < 3
	})

	assert.False(t, result)
	assert.Equal(t, 3, count)
}

// TestBytesIterator_ForEachByteWithIndex tests ForEachByteWithIndex
func TestBytesIterator_ForEachByteWithIndex(t *testing.T) {
	r := New("ABCD")

	indices := []int{}
	r.ForEachByteWithIndex(func(i int, b byte) bool {
		indices = append(indices, i)
		return true
	})

	assert.Equal(t, []int{0, 1, 2, 3}, indices)
}

// TestBytesIterator_MapBytes tests MapBytes
func TestBytesIterator_MapBytes(t *testing.T) {
	r := New("Hello")

	result := r.MapBytes(func(b byte) byte {
		if b >= 'a' && b <= 'z' {
			return b - 32
		}
		return b
	})

	assert.Equal(t, []byte("HELLO"), result)
}

// TestBytesIterator_FilterBytes tests FilterBytes
func TestBytesIterator_FilterBytes(t *testing.T) {
	r := New("Hello")

	result := r.FilterBytes(func(b byte) bool {
		return b >= 'a' && b <= 'z'
	})

	assert.Equal(t, []byte("ello"), result)
}

// TestBytesIterator_FindByte tests FindByte
func TestBytesIterator_FindByte(t *testing.T) {
	r := New("Hello")

	// Find first vowel
	pos, found := r.FindByte(func(b byte) bool {
		return b == 'e' || b == 'o'
	})

	assert.True(t, found)
	assert.Equal(t, 1, pos)
}

// TestBytesIterator_FindByteNotFound tests FindByte when not found
func TestBytesIterator_FindByteNotFound(t *testing.T) {
	r := New("Hello")

	pos, found := r.FindByte(func(b byte) bool {
		return b == 'z'
	})

	assert.False(t, found)
	assert.Equal(t, -1, pos)
}

// TestBytesIterator_FindByteFrom tests FindByteFrom
func TestBytesIterator_FindByteFrom(t *testing.T) {
	r := New("Hello World")

	// Find space starting from position 5 (the space itself)
	pos, found := r.FindByteFrom(5, func(b byte) bool {
		return b == ' '
	})

	assert.True(t, found)
	assert.Equal(t, 5, pos)
}

// TestBytesIterator_AllBytes tests AllBytes
func TestBytesIterator_AllBytes(t *testing.T) {
	r := New("Hello")

	// All ASCII
	result := r.AllBytes(func(b byte) bool {
		return b < 128
	})
	assert.True(t, result)

	// All uppercase
	result = r.AllBytes(func(b byte) bool {
		return b >= 'A' && b <= 'Z'
	})
	assert.False(t, result)
}

// TestBytesIterator_AnyByte tests AnyByte
func TestBytesIterator_AnyByte(t *testing.T) {
	r := New("Hello")

	// Any uppercase
	result := r.AnyByte(func(b byte) bool {
		return b >= 'A' && b <= 'Z'
	})
	assert.True(t, result)

	// Any digit
	result = r.AnyByte(func(b byte) bool {
		return b >= '0' && b <= '9'
	})
	assert.False(t, result)
}

// TestBytesIterator_CountBytes tests CountBytes
func TestBytesIterator_CountBytes(t *testing.T) {
	r := New("Hello World")

	// Count vowels
	count := r.CountBytes(func(b byte) bool {
		return b == 'e' || b == 'o'
	})

	assert.Equal(t, 3, count)
}

// TestBytesIterator_BytesEquals tests BytesEquals
func TestBytesIterator_BytesEquals(t *testing.T) {
	r := New("Hello")

	tests := []struct {
		name     string
		bytes    []byte
		expected bool
	}{
		{
			name:     "Equal bytes",
			bytes:    []byte("Hello"),
			expected: true,
		},
		{
			name:     "Different length",
			bytes:    []byte("Hi"),
			expected: false,
		},
		{
			name:     "Different content",
			bytes:    []byte("World"),
			expected: false,
		},
		{
			name:     "Empty slice",
			bytes:    []byte{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.BytesEquals(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBytesIterator_ToBytes tests Rope ToBytes method
func TestBytesIterator_ToBytesRope(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		bytes []byte
	}{
		{
			name:  "ASCII",
			text:  "Hello",
			bytes: []byte("Hello"),
		},
		{
			name:  "Unicode",
			text:  "你好",
			bytes: []byte("你好"),
		},
		{
			name:  "Empty",
			text:  "",
			bytes: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.ToBytes()
			assert.Equal(t, tt.bytes, result)
		})
	}
}

// TestBytesIterator_MultiLeafRope tests iteration across multiple leaf nodes
func TestBytesIterator_MultiLeafRope(t *testing.T) {
	// Create a rope by concatenating multiple pieces
	r1 := New("Hello")
	r2 := New(" ")
	r3 := New("World")
	r := r1.Concat(r2).Concat(r3)

	it := r.NewBytesIterator()
	collected := []byte{}

	for it.Next() {
		collected = append(collected, it.Current())
	}

	assert.Equal(t, []byte("Hello World"), collected)
}

// TestBytesIterator_UnicodeMultiByte tests iteration through multi-byte UTF-8
func TestBytesIterator_UnicodeMultiByte(t *testing.T) {
	text := "Hello世界🌍World"
	r := New(text)

	it := r.NewBytesIterator()
	expectedBytes := []byte(text)

	collected := it.Collect()
	assert.Equal(t, expectedBytes, collected)
}
