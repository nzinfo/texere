package rope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestReverseIterator_NewIterator tests creating a new reverse iterator
func TestReverseIterator_NewIterator(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		expectEmpty bool
	}{
		{
			name:        "Non-empty rope",
			text:        "Hello",
			expectEmpty: false,
		},
		{
			name:        "Empty rope",
			text:        "",
			expectEmpty: true,
		},
		{
			name:        "Unicode text",
			text:        "世界",
			expectEmpty: false,
		},
		{
			name:        "Emoji text",
			text:        "🌍🌎",
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			it := r.NewReverseIterator()

			assert.Equal(t, tt.expectEmpty, it.IsExhausted())
			if !tt.expectEmpty {
				assert.Equal(t, -1, it.Position())
			}
		})
	}
}

// TestReverseIterator_NilRope tests reverse iterator with nil rope
func TestReverseIterator_NilRope(t *testing.T) {
	var r *Rope
	it := r.NewReverseIterator()

	assert.True(t, it.IsExhausted())
	assert.False(t, it.Next())
	assert.Panics(t, func() { it.Current() })
}

// TestReverseIterator_BasicIteration tests basic reverse iteration
func TestReverseIterator_BasicIteration(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []rune
	}{
		{
			name:     "ASCII text",
			text:     "Hello",
			expected: []rune{'o', 'l', 'l', 'e', 'H'},
		},
		{
			name:     "Single character",
			text:     "A",
			expected: []rune{'A'},
		},
		{
			name:     "Chinese characters",
			text:     "你好",
			expected: []rune{'好', '你'},
		},
		{
			name:     "Emoji",
			text:     "🌍🌎",
			expected: []rune{'🌎', '🌍'},
		},
		{
			name:     "Mixed",
			text:     "Hi🌍",
			expected: []rune{'🌍', 'i', 'H'},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			it := r.NewReverseIterator()

			collected := []rune{}
			for it.Next() {
				collected = append(collected, it.Current())
			}

			assert.Equal(t, tt.expected, collected)
			assert.True(t, it.IsExhausted())
		})
	}
}

// TestReverseIterator_CurrentPanic tests that Current() panics appropriately
func TestReverseIterator_CurrentPanic(t *testing.T) {
	t.Run("Before first Next", func(t *testing.T) {
		r := New("Hello")
		it := r.NewReverseIterator()

		assert.Panics(t, func() { it.Current() })
	})

	t.Run("After exhaustion", func(t *testing.T) {
		r := New("Hi")
		it := r.NewReverseIterator()
		it.Next()
		it.Next()
		it.Next() // Now exhausted

		assert.Panics(t, func() { it.Current() })
	})
}

// TestReverseIterator_Position tests position tracking
func TestReverseIterator_Position(t *testing.T) {
	r := New("Hello")
	it := r.NewReverseIterator()

	assert.Equal(t, -1, it.Position())

	it.Next() // Position 0 = last char 'o'
	assert.Equal(t, 0, it.Position())

	it.Next() // Position 1 = second-to-last 'l'
	assert.Equal(t, 1, it.Position())
}

// TestReverseIterator_PositionFromStart tests PositionFromStart method
func TestReverseIterator_PositionFromStart(t *testing.T) {
	r := New("Hello")
	it := r.NewReverseIterator()

	it.Next()
	assert.Equal(t, 4, it.PositionFromStart())

	it.Next()
	assert.Equal(t, 3, it.PositionFromStart())
}

// TestReverseIterator_HasNext tests HasNext method
func TestReverseIterator_HasNext(t *testing.T) {
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
			it := r.NewReverseIterator()

			for i := 0; i < tt.nextCount; i++ {
				it.Next()
			}

			assert.Equal(t, tt.expectHasNext, it.HasNext())
		})
	}
}

// TestReverseIterator_Reset tests iterator reset
func TestReverseIterator_Reset(t *testing.T) {
	r := New("Hello")
	it := r.NewReverseIterator()

	// Consume some characters
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
	assert.Equal(t, rune('o'), it.Current())
}

// TestReverseIterator_Collect tests Collect method
func TestReverseIterator_Collect(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []rune
	}{
		{
			name:     "ASCII",
			text:     "Hello",
			expected: []rune{'o', 'l', 'l', 'e', 'H'},
		},
		{
			name:     "Unicode",
			text:     "Hi世界",
			expected: []rune{'界', '世', 'i', 'H'},
		},
		{
			name:     "Empty",
			text:     "",
			expected: []rune{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			it := r.NewReverseIterator()

			collected := it.Collect()
			assert.Equal(t, tt.expected, collected)
		})
	}
}

// TestReverseIterator_ToSlice tests ToSlice alias
func TestReverseIterator_ToSlice(t *testing.T) {
	r := New("Test")
	it := r.NewReverseIterator()

	assert.Equal(t, it.Collect(), it.ToSlice())
}

// TestReverseIterator_ToRunes tests ToRunes method
func TestReverseIterator_ToRunes(t *testing.T) {
	r := New("Test")
	it := r.NewReverseIterator()

	assert.Equal(t, it.Collect(), it.ToRunes())
}

// TestReverseIterator_Skip tests Skip method
func TestReverseIterator_Skip(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		skipCount     int
		expectedRune  rune
		expectedHasNext bool
	}{
		{
			name:          "Skip 2 characters",
			text:          "Hello",
			skipCount:     2,
			expectedRune:  'l',
			expectedHasNext: true,
		},
		{
			name:          "Skip all",
			text:          "Hi",
			skipCount:     2,
			expectedRune:  0,
			expectedHasNext: false,
		},
		{
			name:          "Skip zero",
			text:          "Test",
			skipCount:     0,
			expectedRune:  't',
			expectedHasNext: true,
		},
		{
			name:          "Skip negative",
			text:          "Hello",
			skipCount:     -1,
			expectedRune:  'o',
			expectedHasNext: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			it := r.NewReverseIterator()

			it.Skip(tt.skipCount)

			if tt.expectedHasNext {
				assert.True(t, it.Next())
				assert.Equal(t, tt.expectedRune, it.Current())
			} else {
				assert.False(t, it.Next())
			}
		})
	}
}

// TestReverseIterator_Peek tests Peek method
func TestReverseIterator_Peek(t *testing.T) {
	r := New("Hello")
	it := r.NewReverseIterator()

	it.Next() // Position at 'o'

	// Peek at 'l'
	nextRune := it.Peek()
	assert.Equal(t, rune('l'), nextRune)

	// Position should not have changed
	it.Next()
	assert.Equal(t, rune('l'), it.Current())

	// Peek at first character
	it.Next()
	it.Next()
	assert.Equal(t, rune('e'), it.Current())

	// Peek should return 'H' at this point
	nextRune = it.Peek()
	assert.Equal(t, rune('H'), nextRune)

	// Move to 'H'
	it.Next()
	assert.Equal(t, rune('H'), it.Current())

	// Peek beyond should panic
	assert.Panics(t, func() { it.Peek() })
}

// TestReverseIterator_HasPeek tests HasPeek method
func TestReverseIterator_HasPeek(t *testing.T) {
	r := New("Hi")
	it := r.NewReverseIterator()

	it.Next()
	assert.True(t, it.HasPeek())

	it.Next()
	assert.False(t, it.HasPeek())
}

// TestReverseIterator_CharsAtReverse tests CharsAtReverse method
func TestReverseIterator_CharsAtReverse(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		startPos     int
		expectPanic  bool
		expectedRune rune
	}{
		{
			name:         "Start at position 1",
			text:         "Hello",
			startPos:     1,
			expectPanic:  false,
			expectedRune: 'e',
		},
		{
			name:         "Start at beginning",
			text:         "Hello",
			startPos:     0,
			expectPanic:  false,
			expectedRune: 'H',
		},
		{
			name:         "Start at end",
			text:         "Hello",
			startPos:     5,
			expectPanic:  false,
			expectedRune: 'o',
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
				assert.Panics(t, func() { r.CharsAtReverse(tt.startPos) })
			} else {
				it := r.CharsAtReverse(tt.startPos)
				it.Next()
				assert.Equal(t, tt.expectedRune, it.Current())
			}
		})
	}
}

// TestReverseIterator_Seek tests Seek method
func TestReverseIterator_Seek(t *testing.T) {
	r := New("Hello World")
	it := r.NewReverseIterator()

	// Seek to position 2 (3rd from end = 'r')
	// Position 0='d', 1='l', 2='r', 3='o', 4='W', ...
	success := it.Seek(2)
	assert.True(t, success)
	assert.False(t, it.IsExhausted())

	it.Next()
	assert.Equal(t, rune('r'), it.Current())
	assert.Equal(t, 2, it.Position())

	// Seek to invalid position
	success = it.Seek(100)
	assert.False(t, success)
	assert.True(t, it.IsExhausted())

	// Seek to negative
	success = it.Seek(-1)
	assert.False(t, success)
}

// TestReverseIterator_SeekFromStart tests SeekFromStart method
func TestReverseIterator_SeekFromStart(t *testing.T) {
	r := New("Hello World")
	it := r.NewReverseIterator()

	// Seek to position 6 from start ('W')
	// "Hello World" - H(0) e(1) l(2) l(3) o(4) ' '(5) W(6) o(7) r(8) l(9) d(10)
	success := it.SeekFromStart(6)
	assert.True(t, success)
	assert.False(t, it.IsExhausted())

	it.Next()
	assert.Equal(t, rune('W'), it.Current())

	// Seek to invalid position
	success = it.SeekFromStart(100)
	assert.False(t, success)
}

// TestReverseIterator_String tests String method
func TestReverseIterator_String(t *testing.T) {
	r := New("Hello")
	it := r.NewReverseIterator()

	result := it.String()
	assert.Equal(t, "olleH", result)
}

// TestReverseIterator_ForEachReverse tests ForEachReverse
func TestReverseIterator_ForEachReverse(t *testing.T) {
	r := New("Hello")

	count := 0
	r.ForEachReverse(func(r rune) bool {
		count++
		return true
	})

	assert.Equal(t, 5, count)
}

// TestReverseIterator_ForEachReverseEarlyExit tests early exit from ForEachReverse
func TestReverseIterator_ForEachReverseEarlyExit(t *testing.T) {
	r := New("Hello")

	count := 0
	result := r.ForEachReverse(func(r rune) bool {
		count++
		return count < 3
	})

	assert.False(t, result)
	assert.Equal(t, 3, count)
}

// TestReverseIterator_ForEachReverseWithIndex tests ForEachReverseWithIndex
func TestReverseIterator_ForEachReverseWithIndex(t *testing.T) {
	r := New("ABCD")

	indices := []int{}
	r.ForEachReverseWithIndex(func(i int, r rune) bool {
		indices = append(indices, i)
		return true
	})

	assert.Equal(t, []int{3, 2, 1, 0}, indices)
}

// TestReverseIterator_MapReverse tests MapReverse
func TestReverseIterator_MapReverse(t *testing.T) {
	r := New("Hello")

	result := r.MapReverse(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - 32
		}
		return r
	})

	assert.Equal(t, "HELLO", result.String())
}

// TestReverseIterator_FilterReverse tests FilterReverse
func TestReverseIterator_FilterReverse(t *testing.T) {
	r := New("Hello")

	result := r.FilterReverse(func(r rune) bool {
		return r >= 'a' && r <= 'z'
	})

	assert.Equal(t, "ello", result.String())
}

// TestReverseIterator_FindReverse tests FindReverse
func TestReverseIterator_FindReverse(t *testing.T) {
	r := New("Hello")

	// Find last vowel
	pos, found := r.FindReverse(func(r rune) bool {
		return r == 'e' || r == 'o'
	})

	assert.True(t, found)
	assert.Equal(t, 4, pos) // Position of 'o'
}

// TestReverseIterator_FindReverseNotFound tests FindReverse when not found
func TestReverseIterator_FindReverseNotFound(t *testing.T) {
	r := New("Hello")

	pos, found := r.FindReverse(func(r rune) bool {
		return r == 'z'
	})

	assert.False(t, found)
	assert.Equal(t, -1, pos)
}

// TestReverseIterator_FindReverseFrom tests FindReverseFrom
func TestReverseIterator_FindReverseFrom(t *testing.T) {
	r := New("Hello World")

	// Find last vowel before position 5
	pos, found := r.FindReverseFrom(5, func(r rune) bool {
		return r == 'e' || r == 'o'
	})

	assert.True(t, found)
	assert.Equal(t, 1, pos) // Position of 'e'
}

// TestReverseIterator_CountReverse tests CountReverse
func TestReverseIterator_CountReverse(t *testing.T) {
	r := New("Hello World")

	// Count vowels in reverse
	count := r.CountReverse(func(r rune) bool {
		return r == 'e' || r == 'o'
	})

	assert.Equal(t, 3, count)
}

// TestReverseIterator_AllReverse tests AllReverse
func TestReverseIterator_AllReverse(t *testing.T) {
	r := New("Hello")

	// All letters
	result := r.AllReverse(func(r rune) bool {
		return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
	})
	assert.True(t, result)

	// All uppercase
	result = r.AllReverse(func(r rune) bool {
		return r >= 'A' && r <= 'Z'
	})
	assert.False(t, result)
}

// TestReverseIterator_AnyReverse tests AnyReverse
func TestReverseIterator_AnyReverse(t *testing.T) {
	r := New("Hello")

	// Any uppercase
	result := r.AnyReverse(func(r rune) bool {
		return r >= 'A' && r <= 'Z'
	})
	assert.True(t, result)

	// Any digit
	result = r.AnyReverse(func(r rune) bool {
		return r >= '0' && r <= '9'
	})
	assert.False(t, result)
}

// TestReverseIterator_Reverse tests Reverse method
func TestReverseIterator_Reverse(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "ASCII",
			text:     "Hello",
			expected: "olleH",
		},
		{
			name:     "Unicode",
			text:     "你好",
			expected: "好你",
		},
		{
			name:     "Single char",
			text:     "A",
			expected: "A",
		},
		{
			name:     "Empty",
			text:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.text)
			reversed := r.Reverse()
			assert.Equal(t, tt.expected, reversed.String())
		})
	}
}

// TestReverseIterator_LastIndexOf tests LastIndexOf
func TestReverseIterator_LastIndexOf(t *testing.T) {
	r := New("Hello Hello World")

	tests := []struct {
		name     string
		substr   string
		expected int
	}{
		{
			name:     "Find last 'l'",
			substr:   "l",
			expected: 15, // Last 'l' in "World" is at position 15
		},
		{
			name:     "Find last 'Hello'",
			substr:   "Hello",
			expected: 6, // Second "Hello" starts at position 6
		},
		{
			name:     "Not found",
			substr:   "xyz",
			expected: -1,
		},
		{
			name:     "Empty substring",
			substr:   "",
			expected: 17,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.LastIndexOf(tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestReverseIterator_LastIndexOfAny tests LastIndexOfAny
func TestReverseIterator_LastIndexOfAny(t *testing.T) {
	r := New("Hello World")

	result := r.LastIndexOfAny('l', 'o')
	assert.Equal(t, 9, result) // Last 'l' is at position 9 (in "World")

	result = r.LastIndexOfAny('x', 'y', 'z')
	assert.Equal(t, -1, result)
}

// TestReverseIterator_TrimEnd tests TrimEnd
func TestReverseIterator_TrimEnd(t *testing.T) {
	r := New("Hello!!!")

	result := r.TrimEnd(func(r rune) bool {
		return r == '!'
	})

	assert.Equal(t, "Hello", result.String())
}

// TestReverseIterator_TrimStart tests TrimStart
func TestReverseIterator_TrimStart(t *testing.T) {
	r := New("!!!Hello")

	result := r.TrimStart(func(r rune) bool {
		return r == '!'
	})

	assert.Equal(t, "Hello", result.String())
}

// TestReverseIterator_Trim tests Trim
func TestReverseIterator_Trim(t *testing.T) {
	r := New("!!!Hello!!!")

	result := r.Trim(func(r rune) bool {
		return r == '!'
	})

	assert.Equal(t, "Hello", result.String())
}

// TestReverseIterator_IterReverse tests IterReverse alias
func TestReverseIterator_IterReverse(t *testing.T) {
	r := New("Hello")
	it1 := r.NewReverseIterator()
	it2 := r.IterReverse()

	it1.Next()
	it2.Next()

	assert.Equal(t, it1.Current(), it2.Current())
}

// TestReverseIterator_EmptySubstring tests edge cases with empty/subsingle strings
func TestReverseIterator_EmptySubstring(t *testing.T) {
	r := New("A")

	result := r.LastIndexOf("")
	assert.Equal(t, 1, result)

	result = r.LastIndexOf("A")
	assert.Equal(t, 0, result)
}
