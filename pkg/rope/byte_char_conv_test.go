package rope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestByteCharConversion_BasicConversion tests basic byte to char conversion
func TestByteCharConversion_BasicConversion(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		byteIdx     int
		expectedChar int
	}{
		{
			name:        "ASCII text",
			text:        "Hello",
			byteIdx:     2,
			expectedChar: 2,
		},
		{
			name:        "Unicode text",
			text:        "Hello世界",
			byteIdx:     7, // After "Hello" (5 bytes) + first byte of "世"
			expectedChar: 5,
		},
		{
			name:        "Mixed ASCII and emoji",
			text:        "Hi🌍World",
			byteIdx:     3, // After "Hi" + first byte of emoji
			expectedChar: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			charIdx := r.ByteToChar(tt.byteIdx)
			assert.Equal(t, tt.expectedChar, charIdx)
		})
	}
}

// TestByteCharConversion_CharToByte tests character to byte conversion
func TestByteCharConversion_CharToByte(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		charIdx     int
		expectedByte int
	}{
		{
			name:        "ASCII text",
			text:        "Hello",
			charIdx:     2,
			expectedByte: 2,
		},
		{
			name:        "Unicode text",
			text:        "Hello世界",
			charIdx:     6, // "Hello" + "世"
			expectedByte: 8, // 5 + 3 bytes for "世"
		},
		{
			name:        "Emoji",
			text:        "Hi🌍",
			charIdx:     2,
			expectedByte: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			byteIdx := r.CharToByte(tt.charIdx)
			assert.Equal(t, tt.expectedByte, byteIdx)
		})
	}
}

// TestByteCharConversion_RoundTrip tests bidirectional conversion
func TestByteCharConversion_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		charIdx int
	}{
		{
			name:    "ASCII",
			text:    "Hello World",
			charIdx: 5,
		},
		{
			name:    "Unicode Chinese",
			text:    "你好世界",
			charIdx: 2,
		},
		{
			name:    "Emoji",
			text:    "Hello🌍🌎🌏",
			charIdx: 7,
		},
		{
			name:    "Mixed",
			text:    "A你好B🌍C",
			charIdx: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			byteIdx := r.CharToByte(tt.charIdx)
			roundTripChar := r.ByteToChar(byteIdx)
			assert.Equal(t, tt.charIdx, roundTripChar)
		})
	}
}

// TestByteCharConversion_BoundaryConditions tests edge cases
func TestByteCharConversion_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		byteIdx     int
		expectedChar int
	}{
		{
			name:        "Empty rope",
			text:        "",
			byteIdx:     0,
			expectedChar: 0,
		},
		{
			name:        "Single character ASCII",
			text:        "A",
			byteIdx:     0,
			expectedChar: 0,
		},
		{
			name:        "Single character Unicode",
			text:        "世",
			byteIdx:     0,
			expectedChar: 0,
		},
		{
			name:        "Beyond end",
			text:        "Hello",
			byteIdx:     100,
			expectedChar: 5,
		},
		{
			name:        "Negative index",
			text:        "Hello",
			byteIdx:     -1,
			expectedChar: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			charIdx := r.ByteToChar(tt.byteIdx)
			assert.Equal(t, tt.expectedChar, charIdx)
		})
	}
}

// TestByteCharConversion_CharToByteBoundary tests CharToByte edge cases
func TestByteCharConversion_CharToByteBoundary(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		charIdx     int
		expectedByte int
	}{
		{
			name:        "Empty rope",
			text:        "",
			charIdx:     0,
			expectedByte: 0,
		},
		{
			name:        "Single character",
			text:        "A",
			charIdx:     0,
			expectedByte: 0,
		},
		{
			name:        "Beyond end",
			text:        "Hello",
			charIdx:     100,
			expectedByte: 5,
		},
		{
			name:        "Negative index",
			text:        "Hello",
			charIdx:     -1,
			expectedByte: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			byteIdx := r.CharToByte(tt.charIdx)
			assert.Equal(t, tt.expectedByte, byteIdx)
		})
	}
}

// TestByteCharConversion_NilRope tests nil rope handling
func TestByteCharConversion_NilRope(t *testing.T) {
	var r *Rope

	assert.Equal(t, 0, r.ByteToChar(0))
	assert.Equal(t, 0, r.ByteToChar(5))
	assert.Equal(t, 0, r.CharToByte(0))
	assert.Equal(t, 0, r.CharToByte(5))
}

// TestLenRune tests the lenRune helper function
func TestLenRune(t *testing.T) {
	tests := []struct {
		name     string
		r        rune
		expected int
	}{
		{
			name:     "ASCII - single byte",
			r:        'A',
			expected: 1,
		},
		{
			name:     "ASCII - digit",
			r:        '0',
			expected: 1,
		},
		{
			name:     "Latin extended - 2 bytes",
			r:        0x7F,
			expected: 1,
		},
		{
			name:     "2-byte Unicode",
			r:        0x80,
			expected: 2,
		},
		{
			name:     "2-byte Unicode boundary",
			r:        0x7FF,
			expected: 2,
		},
		{
			name:     "3-byte Unicode",
			r:        0x800,
			expected: 3,
		},
		{
			name:     "3-byte Unicode boundary",
			r:        0xFFFF,
			expected: 3,
		},
		{
			name:     "4-byte Unicode",
			r:        0x10000,
			expected: 4,
		},
		{
			name:     "Emoji - 4 bytes",
			r:        '🌍',
			expected: 4,
		},
		{
			name:     "Max valid rune",
			r:        0x10FFFF,
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lenRune(tt.r)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestByteCharConversion_ComplexText tests real-world text
func TestByteCharConversion_ComplexText(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "Japanese",
			text: "こんにちは世界",
		},
		{
			name: "Korean",
			text: "안녕하세요 세계",
		},
		{
			name: "Arabic",
			text: "مرحبا بالعالم",
		},
		{
			name: "Russian",
			text: "Привет мир",
		},
		{
			name: "Mixed scripts",
			text: "Hello你好안녕하세요🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			length := r.Length()

			// Test all character positions
			for i := 0; i <= length; i++ {
				byteIdx := r.CharToByte(i)
				roundTripChar := r.ByteToChar(byteIdx)
				assert.Equal(t, i, roundTripChar,
					"Round trip failed at position %d for text: %s", i, tt.text)
			}
		})
	}
}

// TestByteCharConversion_CRLF tests CRLF handling
func TestByteCharConversion_CRLF(t *testing.T) {
	text := "Line1\r\nLine2\r\nLine3"
	r := New(text)

	// Test conversion at CRLF boundaries
	tests := []struct {
		charIdx     int
		expectedByte int
	}{
		{5, 5},  // Before first \r
		{6, 6},  // At \r
		{7, 7},  // At \n
		{8, 8},  // After CRLF
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			byteIdx := r.CharToByte(tt.charIdx)
			assert.Equal(t, tt.expectedByte, byteIdx)
		})
	}
}

// TestByteCharConversion_MultiByteSequence tests multi-byte character sequences
func TestByteCharConversion_MultiByteSequence(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		indices []struct {
			charIdx int
			byteIdx int
		}
	}{
		{
			name: "Chinese characters",
			text: "你好世界",
			indices: []struct {
				charIdx int
				byteIdx int
			}{
				{0, 0},  // Before "你"
				{1, 3},  // After "你" (3 bytes)
				{2, 6},  // After "你好"
				{3, 9},  // After "你好世"
				{4, 12}, // After all
			},
		},
		{
			name: "Emoji sequence",
			text: "🌍🌎🌏",
			indices: []struct {
				charIdx int
				byteIdx int
			}{
				{0, 0},  // Before first emoji
				{1, 4},  // After first emoji
				{2, 8},  // After second emoji
				{3, 12}, // After all
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)

			for _, idx := range tt.indices {
				// Test CharToByte
				byteIdx := r.CharToByte(idx.charIdx)
				assert.Equal(t, idx.byteIdx, byteIdx,
					"CharToByte failed at char index %d", idx.charIdx)

				// Test ByteToChar
				charIdx := r.ByteToChar(idx.byteIdx)
				assert.Equal(t, idx.charIdx, charIdx,
					"ByteToChar failed at byte index %d", idx.byteIdx)
			}
		})
	}
}
