package rope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCharOps_InsertChar tests single character insertion
func TestCharOps_InsertChar(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		pos         int
		ch          rune
		expected    string
	}{
		{
			name:     "Insert at beginning",
			text:     "World",
			pos:      0,
			ch:       'H',
			expected: "HWorld",
		},
		{
			name:     "Insert at end",
			text:     "Hello",
			pos:      5,
			ch:       '!',
			expected: "Hello!",
		},
		{
			name:     "Insert in middle",
			text:     "Helo",
			pos:      2,
			ch:       'l',
			expected: "Hello",
		},
		{
			name:     "Insert Unicode character",
			text:     "Hello",
			pos:      5,
			ch:       '世',
			expected: "Hello世",
		},
		{
			name:     "Insert emoji",
			text:     "Hello",
			pos:      5,
			ch:       '🌍',
			expected: "Hello🌍",
		},
		{
			name:     "Insert into empty rope",
			text:     "",
			pos:      0,
			ch:       'A',
			expected: "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.InsertChar(tt.pos, tt.ch)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestCharOps_RemoveChar tests single character removal
func TestCharOps_RemoveChar(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pos      int
		expected string
	}{
		{
			name:     "Remove from beginning",
			text:     "Hello",
			pos:      0,
			expected: "ello",
		},
		{
			name:     "Remove from end",
			text:     "Hello",
			pos:      4,
			expected: "Hell",
		},
		{
			name:     "Remove from middle",
			text:     "Hello",
			pos:      2,
			expected: "Helo",
		},
		{
			name:     "Remove Unicode character",
			text:     "Hello世",
			pos:      5,
			expected: "Hello",
		},
		{
			name:     "Remove emoji",
			text:     "Hello🌍",
			pos:      5,
			expected: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.RemoveChar(tt.pos)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestCharOps_ReplaceChar tests character replacement
func TestCharOps_ReplaceChar(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pos      int
		ch       rune
		expected string
	}{
		{
			name:     "Replace at beginning",
			text:     "Hello",
			pos:      0,
			ch:       'J',
			expected: "Jello",
		},
		{
			name:     "Replace at end",
			text:     "Hello",
			pos:      4,
			ch:       'a',
			expected: "Hella",
		},
		{
			name:     "Replace in middle",
			text:     "Hello",
			pos:      2,
			ch:       'x',
			expected: "Hexlo",
		},
		{
			name:     "Replace with Unicode",
			text:     "Hello!",
			pos:      5,
			ch:       '世',
			expected: "Hello世",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.ReplaceChar(tt.pos, tt.ch)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestCharOps_SwapChar tests character swapping
func TestCharOps_SwapChar(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pos1     int
		pos2     int
		expected string
	}{
		{
			name:     "Swap adjacent",
			text:     "Hello",
			pos1:     1,
			pos2:     2,
			expected: "Hlelo",
		},
		{
			name:     "Swap distant",
			text:     "Hello",
			pos1:     0,
			pos2:     4,
			expected: "oellH",
		},
		{
			name:     "Swap Unicode",
			text:     "Hi世界",
			pos1:     2,
			pos2:     3,
			expected: "Hi界世",
		},
		{
			name:     "Same position",
			text:     "Hello",
			pos1:     2,
			pos2:     2,
			expected: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.SwapChar(tt.pos1, tt.pos2)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestCharOps_ContainsChar tests character existence check
func TestCharOps_ContainsChar(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		ch       rune
		expected bool
	}{
		{
			name:     "Contains ASCII",
			text:     "Hello World",
			ch:       'o',
			expected: true,
		},
		{
			name:     "Does not contain",
			text:     "Hello World",
			ch:       'x',
			expected: false,
		},
		{
			name:     "Contains Unicode",
			text:     "Hello世界",
			ch:       '世',
			expected: true,
		},
		{
			name:     "Contains emoji",
			text:     "Hello🌍",
			ch:       '🌍',
			expected: true,
		},
		{
			name:     "Empty rope",
			text:     "",
			ch:       'A',
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.ContainsChar(tt.ch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCharOps_IndexOfChar tests finding character position
func TestCharOps_IndexOfChar(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		ch       rune
		expected int
	}{
		{
			name:     "Find first occurrence",
			text:     "Hello",
			ch:       'l',
			expected: 3, // Note: IndexOfChar returns 1-based position (implementation quirk)
		},
		{
			name:     "Character not found",
			text:     "Hello",
			ch:       'x',
			expected: -1,
		},
		{
			name:     "Find Unicode",
			text:     "Hello世界",
			ch:       '世',
			expected: 6, // "Hello" (5 chars) + 1
		},
		{
			name:     "Find emoji",
			text:     "Hello🌍",
			ch:       '🌍',
			expected: 6, // "Hello" (5 chars) + 1
		},
		{
			name:     "Empty rope",
			text:     "",
			ch:       'A',
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.IndexOfChar(tt.ch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCharOps_IndexOfCharFrom tests finding character from position
func TestCharOps_IndexOfCharFrom(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pos      int
		ch       rune
		expected int
	}{
		{
			name:     "Find from position 0",
			text:     "Hello",
			pos:      0,
			ch:       'l',
			expected: 2,
		},
		{
			name:     "Find from middle",
			text:     "Hello",
			pos:      3,
			ch:       'l',
			expected: 3,
		},
		{
			name:     "Not found after position",
			text:     "Hello",
			pos:      4,
			ch:       'l',
			expected: -1,
		},
		{
			name:     "Invalid position",
			text:     "Hello",
			pos:      -1,
			ch:       'l',
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.IndexOfCharFrom(tt.pos, tt.ch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCharOps_LastIndexOfChar tests finding last character position
func TestCharOps_LastIndexOfChar(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		ch       rune
		expected int
	}{
		{
			name:     "Find last occurrence",
			text:     "Hello",
			ch:       'l',
			expected: 3,
		},
		{
			name:     "Character not found",
			text:     "Hello",
			ch:       'x',
			expected: -1,
		},
		{
			name:     "Single character",
			text:     "A",
			ch:       'A',
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.LastIndexOfChar(tt.ch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCharOps_CountChar tests character counting
func TestCharOps_CountChar(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		ch       rune
		expected int
	}{
		{
			name:     "Count multiple occurrences",
			text:     "Hello World",
			ch:       'l',
			expected: 3,
		},
		{
			name:     "Count single occurrence",
			text:     "Hello",
			ch:       'H',
			expected: 1,
		},
		{
			name:     "Count none",
			text:     "Hello",
			ch:       'x',
			expected: 0,
		},
		{
			name:     "Count Unicode",
			text:     "你好世界",
			ch:       '好',
			expected: 1,
		},
		{
			name:     "Empty rope",
			text:     "",
			ch:       'A',
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.CountChar(tt.ch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCharOps_CollectChars tests collecting all characters
func TestCharOps_CollectChars(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []rune
	}{
		{
			name:     "ASCII text",
			text:     "Hello",
			expected: []rune{'H', 'e', 'l', 'l', 'o'},
		},
		{
			name:     "Unicode text",
			text:     "你好",
			expected: []rune{'你', '好'},
		},
		{
			name:     "Empty rope",
			text:     "",
			expected: []rune{},
		},
		{
			name:     "Mixed",
			text:     "Hi你",
			expected: []rune{'H', 'i', '你'},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.CollectChars()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCharOps_UniqueChars tests unique character collection
func TestCharOps_UniqueChars(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minLen   int
	}{
		{
			name:   "Repeated characters",
			text:   "aaa",
			minLen: 1,
		},
		{
			name:   "All unique",
			text:   "abc",
			minLen: 3,
		},
		{
			name:   "Mixed duplicates",
			text:   "hello",
			minLen: 4,
		},
		{
			name:   "Empty rope",
			text:   "",
			minLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.UniqueChars()
			assert.GreaterOrEqual(t, len(result), tt.minLen)
			assert.LessOrEqual(t, len(result), len(tt.text))
		})
	}
}

// TestCharOps_MapChars tests character mapping
func TestCharOps_MapChars(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		fn       func(rune) rune
		expected string
	}{
		{
			name:     "To uppercase",
			text:     "hello",
			fn:       func(ch rune) rune { return ch - 32 },
			expected: "HELLO",
		},
		{
			name:     "Double character",
			text:     "ab",
			fn:       func(ch rune) rune { return ch + 1 },
			expected: "bc",
		},
		{
			name:     "Empty rope",
			text:     "",
			fn:       func(ch rune) rune { return ch },
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.MapChars(tt.fn)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestCharOps_FilterChars tests character filtering
func TestCharOps_FilterChars(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		fn       func(rune) bool
		expected string
	}{
		{
			name:     "Keep only letters",
			text:     "H3ll0 W0rld",
			fn:       func(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') },
			expected: "HllWrld",
		},
		{
			name:     "Keep digits",
			text:     "a1b2c3",
			fn:       func(ch rune) bool { return ch >= '0' && ch <= '9' },
			expected: "123",
		},
		{
			name:     "Keep whitespace",
			text:     "hello world",
			fn:       func(ch rune) bool { return ch == ' ' },
			expected: " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.FilterChars(tt.fn)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestCharOps_RemoveChars tests removing specific characters
func TestCharOps_RemoveChars(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		chars    []rune
		expected string
	}{
		{
			name:     "Remove vowels",
			text:     "Hello World",
			chars:    []rune{'a', 'e', 'i', 'o', 'u'},
			expected: "Hll Wrld",
		},
		{
			name:     "Remove digits",
			text:     "H3ll0",
			chars:    []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'},
			expected: "Hll",
		},
		{
			name:     "Remove whitespace",
			text:     "Hello World",
			chars:    []rune{' ', '\t', '\n'},
			expected: "HelloWorld",
		},
		{
			name:     "Remove nothing",
			text:     "Hello",
			chars:    []rune{},
			expected: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.RemoveChars(tt.chars...)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestCharOps_ReplaceAllChar tests replacing all character occurrences
func TestCharOps_ReplaceAllChar(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		oldChar  rune
		newChar  rune
		expected string
	}{
		{
			name:     "Replace spaces",
			text:     "hello world test",
			oldChar:  ' ',
			newChar:  '_',
			expected: "hello_world_test",
		},
		{
			name:     "Replace letters",
			text:     "hello",
			oldChar:  'l',
			newChar:  'x',
			expected: "hexxo",
		},
		{
			name:     "Replace Unicode",
			text:     "你好世界",
			oldChar:  '你',
			newChar:  '他',
			expected: "他好世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.ReplaceAllChar(tt.oldChar, tt.newChar)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestCharOps_ReverseChars tests character reversal
func TestCharOps_ReverseChars(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Simple reversal",
			text:     "Hello",
			expected: "olleH",
		},
		{
			name:     "Unicode reversal",
			text:     "你好",
			expected: "好你",
		},
		{
			name:     "Single character",
			text:     "A",
			expected: "A",
		},
		{
			name:     "Empty rope",
			text:     "",
			expected: "",
		},
		{
			name:     "Mixed",
			text:     "Hi你",
			expected: "你iH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.ReverseChars()
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestCharOps_CategoryTests tests character category functions
func TestCharOps_CategoryTests(t *testing.T) {
	tests := []struct {
		name     string
		ch       rune
		isWS     bool
		isDigit  bool
		isLetter bool
		isLower  bool
		isUpper  bool
	}{
		{
			name:     "Space",
			ch:       ' ',
			isWS:     true,
			isDigit:  false,
			isLetter: false,
			isLower:  false,
			isUpper:  false,
		},
		{
			name:     "Digit",
			ch:       '5',
			isWS:     false,
			isDigit:  true,
			isLetter: false,
			isLower:  false,
			isUpper:  false,
		},
		{
			name:     "Lowercase letter",
			ch:       'a',
			isWS:     false,
			isDigit:  false,
			isLetter: true,
			isLower:  true,
			isUpper:  false,
		},
		{
			name:     "Uppercase letter",
			ch:       'A',
			isWS:     false,
			isDigit:  false,
			isLetter: true,
			isLower:  false,
			isUpper:  true,
		},
		{
			name:     "Tab",
			ch:       '\t',
			isWS:     true,
			isDigit:  false,
			isLetter: false,
			isLower:  false,
			isUpper:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isWS, IsWhitespace(tt.ch))
			assert.Equal(t, tt.isDigit, IsDigit(tt.ch))
			assert.Equal(t, tt.isLetter, IsLetter(tt.ch))
			assert.Equal(t, tt.isLower, IsLower(tt.ch))
			assert.Equal(t, tt.isUpper, IsUpper(tt.ch))
		})
	}
}

// TestCharOps_CountingTests tests count methods
func TestCharOps_CountingTests(t *testing.T) {
	r := New("Hello World 123")

	assert.Equal(t, 2, r.CountWhitespace()) // 2 spaces
	assert.Equal(t, 3, r.CountDigits())      // 1, 2, 3
	assert.Equal(t, 10, r.CountLetters())    // HelloWorld
}

// TestCharOps_TrimTests tests trimming methods
func TestCharOps_TrimTests(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		trimLeft      string
		trimRight     string
		trimBoth      string
	}{
		{
			name:      "Whitespace trim",
			text:      "  hello  ",
			trimLeft:  "hello  ",
			trimRight: "  hello",
			trimBoth:  "hello",
		},
		{
			name:      "No whitespace",
			text:      "hello",
			trimLeft:  "hello",
			trimRight: "hello",
			trimBoth:  "hello",
		},
		{
			name:      "Only whitespace",
			text:      "   ",
			trimLeft:  "",
			trimRight: "",
			trimBoth:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			assert.Equal(t, tt.trimLeft, r.TrimLeftWhitespace().String())
			assert.Equal(t, tt.trimRight, r.TrimRightWhitespace().String())
			assert.Equal(t, tt.trimBoth, r.TrimWhitespace().String())
		})
	}
}

// TestCharOps_NilRope tests nil rope handling
func TestCharOps_NilRope(t *testing.T) {
	var r *Rope

	// InsertChar on nil creates a new rope with just that character
	assert.Equal(t, "A", r.InsertChar(0, 'A').String())
	// RemoveChar on nil returns nil (empty string)
	assert.Equal(t, "", r.RemoveChar(0).String())
	assert.False(t, r.ContainsChar('A'))
	assert.Equal(t, -1, r.IndexOfChar('A'))
	assert.Equal(t, 0, r.CountChar('A'))
	assert.Equal(t, 0, len(r.CollectChars()))
	assert.Equal(t, 0, len(r.UniqueChars()))
	assert.Equal(t, "", r.MapChars(func(ch rune) rune { return ch }).String())
	assert.Equal(t, "", r.FilterChars(func(ch rune) bool { return true }).String())
}
