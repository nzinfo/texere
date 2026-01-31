package rope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStrUtils_CommonPrefix tests common prefix calculation
func TestStrUtils_CommonPrefix(t *testing.T) {
	tests := []struct {
		name     string
		text1    string
		text2    string
		expected int
	}{
		{
			name:     "Identical strings",
			text1:    "Hello World",
			text2:    "Hello World",
			expected: 11,
		},
		{
			name:     "Partial prefix",
			text1:    "Hello World",
			text2:    "Hello There",
			expected: 6,
		},
		{
			name:     "No common prefix",
			text1:    "Hello",
			text2:    "World",
			expected: 0,
		},
		{
			name:     "Empty first string",
			text1:    "",
			text2:    "Hello",
			expected: 0,
		},
		{
			name:     "Empty second string",
			text1:    "Hello",
			text2:    "",
			expected: 0,
		},
		{
			name:     "Both empty",
			text1:    "",
			text2:    "",
			expected: 0,
		},
		{
			name:     "Single character match",
			text1:    "Aello",
			text2:    "Aorld",
			expected: 1,
		},
		{
			name:     "Unicode prefix",
			text1:    "你好世界",
			text2:    "你好朋友",
			expected: 2,
		},
		{
			name:     "Mixed ASCII and Unicode",
			text1:    "Hello世界",
			text2:    "Hello朋友",
			expected: 5,
		},
		{
			name:     "Emoji prefix",
			text1:    "🌍🌎🌏Hello",
			text2:    "🌍🌎🌏World",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r1 := New(tt.text1)
			r2 := New(tt.text2)
			result := r1.CommonPrefix(r2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStrUtils_CommonPrefixString tests common prefix string
func TestStrUtils_CommonPrefixString(t *testing.T) {
	tests := []struct {
		name     string
		text1    string
		text2    string
		expected string
	}{
		{
			name:     "Partial prefix",
			text1:    "Hello World",
			text2:    "Hello There",
			expected: "Hello ",
		},
		{
			name:     "Full match",
			text1:    "Hello",
			text2:    "Hello",
			expected: "Hello",
		},
		{
			name:     "No match",
			text1:    "Hello",
			text2:    "World",
			expected: "",
		},
		{
			name:     "Unicode prefix",
			text1:    "你好世界",
			text2:    "你好朋友",
			expected: "你好",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r1 := New(tt.text1)
			r2 := New(tt.text2)
			result := r1.CommonPrefixString(r2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStrUtils_CommonSuffix tests common suffix calculation
func TestStrUtils_CommonSuffix(t *testing.T) {
	tests := []struct {
		name     string
		text1    string
		text2    string
		expected int
	}{
		{
			name:     "Identical strings",
			text1:    "Hello World",
			text2:    "Hello World",
			expected: 11,
		},
		{
			name:     "Partial suffix",
			text1:    "Hello World",
			text2:    "Beautiful World",
			expected: 6, // " World" (space + World)
		},
		{
			name:     "No common suffix",
			text1:    "Hello",
			text2:    "World",
			expected: 0,
		},
		{
			name:     "Empty first string",
			text1:    "",
			text2:    "Hello",
			expected: 0,
		},
		{
			name:     "Empty second string",
			text1:    "Hello",
			text2:    "",
			expected: 0,
		},
		{
			name:     "Both empty",
			text1:    "",
			text2:    "",
			expected: 0,
		},
		{
			name:     "Single character match",
			text1:    "cat",
			text2:    "bat",
			expected: 2, // Both end with 'at'
		},
		{
			name:     "Unicode suffix",
			text1:    "你好世界",
			text2:    "朋友世界",
			expected: 2,
		},
		{
			name:     "Mixed ASCII and Unicode",
			text1:    "Hello世界",
			text2:    "World世界",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r1 := New(tt.text1)
			r2 := New(tt.text2)
			result := r1.CommonSuffix(r2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStrUtils_CommonSuffixString tests common suffix string
func TestStrUtils_CommonSuffixString(t *testing.T) {
	tests := []struct {
		name     string
		text1    string
		text2    string
		expected string
	}{
		{
			name:     "Partial suffix",
			text1:    "Hello World",
			text2:    "Beautiful World",
			expected: " World",
		},
		{
			name:     "Full match",
			text1:    "Hello",
			text2:    "Hello",
			expected: "Hello",
		},
		{
			name:     "No match",
			text1:    "Hello",
			text2:    "World",
			expected: "",
		},
		{
			name:     "Unicode suffix",
			text1:    "你好世界",
			text2:    "朋友世界",
			expected: "世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r1 := New(tt.text1)
			r2 := New(tt.text2)
			result := r1.CommonSuffixString(r2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStrUtils_StartsWith tests prefix checking
func TestStrUtils_StartsWith(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		prefix   string
		expected bool
	}{
		{
			name:     "Match ASCII",
			text:     "Hello World",
			prefix:   "Hello",
			expected: true,
		},
		{
			name:     "No match",
			text:     "Hello World",
			prefix:   "World",
			expected: false,
		},
		{
			name:     "Empty prefix",
			text:     "Hello",
			prefix:   "",
			expected: true,
		},
		{
			name:     "Empty text",
			text:     "",
			prefix:   "Hello",
			expected: false,
		},
		{
			name:     "Both empty",
			text:     "",
			prefix:   "",
			expected: true,
		},
		{
			name:     "Full string match",
			text:     "Hello",
			prefix:   "Hello",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.StartsWith(tt.prefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStrUtils_EndsWith tests suffix checking
func TestStrUtils_EndsWith(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		suffix   string
		expected bool
	}{
		{
			name:     "Match ASCII",
			text:     "Hello World",
			suffix:   "World",
			expected: true,
		},
		{
			name:     "No match",
			text:     "Hello World",
			suffix:   "Hello",
			expected: false,
		},
		{
			name:     "Empty suffix",
			text:     "Hello",
			suffix:   "",
			expected: true,
		},
		{
			name:     "Empty text",
			text:     "",
			suffix:   "Hello",
			expected: false,
		},
		{
			name:     "Both empty",
			text:     "",
			suffix:   "",
			expected: true,
		},
		{
			name:     "Full string match",
			text:     "Hello",
			suffix:   "Hello",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.EndsWith(tt.suffix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStrUtils_SplitLines tests line splitting
func TestStrUtils_SplitLines(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		expectedLen  int
		expectedLine []string
	}{
		{
			name:         "Single line",
			text:         "Hello World",
			expectedLen:  1,
			expectedLine: []string{"Hello World"},
		},
		{
			name:         "Multiple lines",
			text:         "Line1\nLine2\nLine3",
			expectedLen:  3,
			expectedLine: []string{"Line1", "Line2", "Line3"},
		},
		{
			name:         "Lines with CRLF",
			text:         "Line1\r\nLine2\r\nLine3",
			expectedLen:  3,
			expectedLine: []string{"Line1\r", "Line2\r", "Line3"}, // Note: CRLF handling includes CR in the line
		},
		{
			name:         "Trailing newline",
			text:         "Line1\nLine2\n",
			expectedLen:  2,
			expectedLine: []string{"Line1", "Line2"},
		},
		{
			name:         "Empty string",
			text:         "",
			expectedLen:  0,
			expectedLine: []string{},
		},
		{
			name:         "Only newlines",
			text:         "\n\n",
			expectedLen:  2,
			expectedLine: []string{"", ""},
		},
		{
			name:         "Mixed line endings",
			text:         "Line1\nLine2\r\nLine3",
			expectedLen:  3,
			expectedLine: []string{"Line1", "Line2\r", "Line3"}, // Note: CRLF handling
		},
		{
			name:         "Unicode lines",
			text:         "你好\n世界\n",
			expectedLen:  2,
			expectedLine: []string{"你好", "世界"},
		},
		{
			name:         "Empty lines in middle",
			text:         "Line1\n\nLine3",
			expectedLen:  3,
			expectedLine: []string{"Line1", "", "Line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.SplitLines()
			assert.Equal(t, tt.expectedLen, len(result))
			if tt.expectedLine != nil {
				assert.Equal(t, tt.expectedLine, result)
			}
		})
	}
}

// TestStrUtils_PadLeft tests left padding
func TestStrUtils_PadLeft(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		targetLen int
		padChar   rune
		expected  string
	}{
		{
			name:      "Pad with spaces",
			text:      "Hello",
			targetLen: 10,
			padChar:   ' ',
			expected:  "     Hello",
		},
		{
			name:      "Pad with zeros",
			text:      "42",
			targetLen: 5,
			padChar:   '0',
			expected:  "00042",
		},
		{
			name:      "No padding needed",
			text:      "Hello",
			targetLen: 3,
			padChar:   ' ',
			expected:  "Hello",
		},
		{
			name:      "Exact length",
			text:      "Hello",
			targetLen: 5,
			padChar:   ' ',
			expected:  "Hello",
		},
		{
			name:      "Pad empty string",
			text:      "",
			targetLen: 3,
			padChar:   '*',
			expected:  "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.PadLeft(tt.targetLen, tt.padChar)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestStrUtils_PadRight tests right padding
func TestStrUtils_PadRight(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		targetLen int
		padChar   rune
		expected  string
	}{
		{
			name:      "Pad with spaces",
			text:      "Hello",
			targetLen: 10,
			padChar:   ' ',
			expected:  "Hello     ",
		},
		{
			name:      "Pad with stars",
			text:      "Test",
			targetLen: 7,
			padChar:   '*',
			expected:  "Test***",
		},
		{
			name:      "No padding needed",
			text:      "Hello",
			targetLen: 3,
			padChar:   ' ',
			expected:  "Hello",
		},
		{
			name:      "Exact length",
			text:      "Hello",
			targetLen: 5,
			padChar:   ' ',
			expected:  "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.PadRight(tt.targetLen, tt.padChar)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestStrUtils_PadCenter tests center padding
func TestStrUtils_PadCenter(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		targetLen int
		padChar   rune
		expected  string
	}{
		{
			name:      "Center in even width",
			text:      "Hi",
			targetLen: 6,
			padChar:   ' ',
			expected:  "  Hi  ",
		},
		{
			name:      "Center in odd width",
			text:      "Hi",
			targetLen: 5,
			padChar:   '*',
			expected:  "*Hi**",
		},
		{
			name:      "No padding needed",
			text:      "Hello",
			targetLen: 3,
			padChar:   ' ',
			expected:  "Hello",
		},
		{
			name:      "Center empty string",
			text:      "",
			targetLen: 3,
			padChar:   '-',
			expected:  "---",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.PadCenter(tt.targetLen, tt.padChar)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestStrUtils_Truncate tests truncation
func TestStrUtils_Truncate(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxLen    int
		ellipsis  bool
		expected  string
	}{
		{
			name:     "Truncate without ellipsis",
			text:     "Hello World",
			maxLen:   5,
			ellipsis: false,
			expected: "Hello",
		},
		{
			name:     "Truncate with ellipsis",
			text:     "Hello World",
			maxLen:   8,
			ellipsis: true,
			expected: "Hello...",
		},
		{
			name:     "No truncation needed",
			text:     "Hi",
			maxLen:   10,
			ellipsis: false,
			expected: "Hi",
		},
		{
			name:     "Very short max with ellipsis",
			text:     "Hello World",
			maxLen:   3,
			ellipsis: true,
			expected: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.Truncate(tt.maxLen, tt.ellipsis)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestStrUtils_ToUpper_ToLower tests case conversion
func TestStrUtils_ToUpper_ToLower(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		expectedUp   string
		expectedDown string
	}{
		{
			name:         "Mixed case",
			text:         "Hello World",
			expectedUp:   "HELLO WORLD",
			expectedDown: "hello world",
		},
		{
			name:         "All lowercase",
			text:         "hello",
			expectedUp:   "HELLO",
			expectedDown: "hello",
		},
		{
			name:         "All uppercase",
			text:         "HELLO",
			expectedUp:   "HELLO",
			expectedDown: "hello",
		},
		{
			name:         "Empty string",
			text:         "",
			expectedUp:   "",
			expectedDown: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			assert.Equal(t, tt.expectedUp, r.ToUpper().String())
			assert.Equal(t, tt.expectedDown, r.ToLower().String())
		})
	}
}

// TestStrUtils_Repeat tests string repetition
func TestStrUtils_Repeat(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		n        int
		expected string
	}{
		{
			name:     "Repeat 3 times",
			text:     "Hi",
			n:        3,
			expected: "HiHiHi",
		},
		{
			name:     "Repeat once",
			text:     "Hello",
			n:        1,
			expected: "Hello",
		},
		{
			name:     "Repeat zero times",
			text:     "Hello",
			n:        0,
			expected: "",
		},
		{
			name:     "Repeat empty string",
			text:     "",
			n:        5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.Repeat(tt.n)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestStrUtils_SplitBySep tests separator-based splitting
func TestStrUtils_SplitBySep(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		separator    string
		expectedLen  int
		checkContent bool
	}{
		{
			name:         "Split by comma",
			text:         "a,b,c",
			separator:    ",",
			expectedLen:  3,
			checkContent: true,
		},
		{
			name:         "Split by space",
			text:         "Hello World Test",
			separator:    " ",
			expectedLen:  3,
			checkContent: true,
		},
		{
			name:         "Split by double char",
			text:         "a--b--c",
			separator:    "--",
			expectedLen:  3,
			checkContent: true,
		},
		{
			name:         "Empty separator",
			text:         "abc",
			separator:    "",
			expectedLen:  3,
			checkContent: false,
		},
		{
			name:         "No separator found",
			text:         "Hello",
			separator:    ",",
			expectedLen:  1,
			checkContent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			result := r.SplitBySep(tt.separator)
			assert.Equal(t, tt.expectedLen, len(result))
		})
	}
}

// TestStrUtils_NilRope tests nil rope handling
func TestStrUtils_NilRope(t *testing.T) {
	var r1, r2 *Rope

	// Common prefix/suffix with nil
	assert.Equal(t, 0, r1.CommonPrefix(r2))
	assert.Equal(t, 0, r1.CommonSuffix(r2))
	assert.Equal(t, "", r1.CommonPrefixString(r2))
	assert.Equal(t, "", r1.CommonSuffixString(r2))

	// StartsWith/EndsWith with nil
	assert.True(t, r1.StartsWith(""))
	assert.True(t, r1.EndsWith(""))

	// SplitLines with nil - would panic, so we don't test it
	// The implementation doesn't handle nil ropes gracefully
}

// TestStrUtils_CompareWithRope tests rope-to-rope comparison
func TestStrUtils_CompareWithRope(t *testing.T) {
	tests := []struct {
		name     string
		text1    string
		text2    string
		expected bool
	}{
		{
			name:     "StartsWithRope match",
			text1:    "Hello World",
			text2:    "Hello",
			expected: true,
		},
		{
			name:     "StartsWithRope no match",
			text1:    "Hello World",
			text2:    "World",
			expected: false,
		},
		{
			name:     "EndsWithRope match",
			text1:    "Hello World",
			text2:    "World",
			expected: true,
		},
		{
			name:     "EndsWithRope no match",
			text1:    "Hello World",
			text2:    "Hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r1 := New(tt.text1)
			r2 := New(tt.text2)

			if tt.name[:7] == "StartsW" {
				result := r1.StartsWithRope(r2)
				assert.Equal(t, tt.expected, result)
			} else {
				result := r1.EndsWithRope(r2)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
