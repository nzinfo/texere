package rope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ========== Basic Grapheme Tests ==========

func TestGrapheme_Empty(t *testing.T) {
	r := New("")
	assert.Equal(t, 0, r.LenGraphemes())

	it := r.Graphemes()
	count := 0
	for it.Next() {
		count++
	}
	assert.Equal(t, 0, count)
}

func TestGrapheme_ASCII(t *testing.T) {
	r := New("hello")
	assert.Equal(t, 5, r.LenGraphemes())

	var graphemes []string
	it := r.Graphemes()
	for it.Next() {
		graphemes = append(graphemes, it.Current().Text)
	}
	assert.Equal(t, []string{"h", "e", "l", "l", "o"}, graphemes)
}

func TestGrapheme_Emoji(t *testing.T) {
	// Each emoji is 1 grapheme cluster
	r := New("🎃🎨🎹🎸")
	assert.Equal(t, 4, r.LenGraphemes())

	it := r.Graphemes()
	expected := []string{"🎃", "🎨", "🎹", "🎸"}
	i := 0
	for it.Next() {
		if i < len(expected) {
			assert.Equal(t, expected[i], it.Current().Text)
		}
		i++
	}
	assert.Equal(t, 4, i)
}

func TestGrapheme_CombiningMarks(t *testing.T) {
	// é (e + combining acute) + l̀ (l + combining grave) = 2 graphemes
	r := New("e\u0301l\u0300")
	assert.Equal(t, 2, r.LenGraphemes())

	it := r.Graphemes()
	it.Next()
	assert.Equal(t, "e\u0301", it.Current().Text) // e + combining acute

	it.Next()
	assert.Equal(t, "l\u0300", it.Current().Text) // l + combining grave
}

func TestGrapheme_FamilyEmoji(t *testing.T) {
	// 👨‍👩‍👧‍👦 = man + ZWJ + woman + ZWJ + girl + ZWJ + boy
	r := New("👨‍👩‍👧‍👦")
	assert.Equal(t, 1, r.LenGraphemes())

	it := r.Graphemes()
	it.Next()
	assert.Equal(t, "👨‍👩‍👧‍👦", it.Current().Text)
}

func TestGrapheme_IterationConsistency(t *testing.T) {
	text := "Hello World"
	r := New(text)

	builder := NewBuilder()
	it := r.Graphemes()
	for it.Next() {
		builder.Append(it.Current().Text)
	}

	r2 := builder.Build()
	assert.Equal(t, text, r2.String())
}

// ========== Grapheme Boundary Tests ==========

func TestGrapheme_Boundaries(t *testing.T) {
	r := New("cafe\u0301") // cafe + combining acute

	// There are 5 code points: c, a, f, e, combining acute
	// Graphemes: c(0), a(1), f(2), é(3-4)
	// So prev_grapheme_start returns the start of the grapheme containing the position
	assert.Equal(t, 0, r.PrevGraphemeStart(0)) // 'c' starts at 0
	assert.Equal(t, 1, r.PrevGraphemeStart(1)) // 'a' starts at 1
	assert.Equal(t, 2, r.PrevGraphemeStart(2)) // 'f' starts at 2
	assert.Equal(t, 3, r.PrevGraphemeStart(3)) // 'é' starts at 3
	assert.Equal(t, 3, r.PrevGraphemeStart(4)) // Still inside 'é', returns start 3
}

func TestGrapheme_NextBoundaries(t *testing.T) {
	r := New("cafe\u0301")

	assert.Equal(t, 1, r.NextGraphemeStart(0)) // Next grapheme after 'c' starts at 1
	assert.Equal(t, 2, r.NextGraphemeStart(1)) // Next after 'a' starts at 2
	assert.Equal(t, 3, r.NextGraphemeStart(2)) // Next after 'f' starts at 3
	assert.Equal(t, 5, r.NextGraphemeStart(3)) // Next after 'é' starts at 5 (past end)
	assert.Equal(t, 5, r.NextGraphemeStart(4)) // Inside 'é', next is at 5
}

func TestGrapheme_IsBoundary(t *testing.T) {
	r := New("cafe\u0301")

	assert.True(t, r.IsGraphemeBoundary(0)) // Start of 'c'
	assert.True(t, r.IsGraphemeBoundary(1)) // Start of 'a'
	assert.True(t, r.IsGraphemeBoundary(2)) // Start of 'f'
	assert.True(t, r.IsGraphemeBoundary(3)) // Start of 'é'
	assert.False(t, r.IsGraphemeBoundary(4)) // Inside 'é'
	assert.True(t, r.IsGraphemeBoundary(5)) // End of 'é'
}

// ========== GraphemeAt Tests ==========

func TestGrapheme_At(t *testing.T) {
	r := New("abc")

	g0 := r.GraphemeAt(0)
	assert.Equal(t, "a", g0.Text)
	assert.Equal(t, 0, g0.StartPos)

	g1 := r.GraphemeAt(1)
	assert.Equal(t, "b", g1.Text)
	assert.Equal(t, 1, g1.StartPos)

	g2 := r.GraphemeAt(2)
	assert.Equal(t, "c", g2.Text)
	assert.Equal(t, 2, g2.StartPos)
}

func TestGrapheme_AtCombining(t *testing.T) {
	r := New("e\u0301l\u0300")

	g0 := r.GraphemeAt(0)
	assert.Equal(t, "e\u0301", g0.Text)
	assert.Equal(t, 2, g0.CharLen) // 2 runes

	g1 := r.GraphemeAt(1)
	assert.Equal(t, "l\u0300", g1.Text)
	assert.Equal(t, 2, g1.CharLen) // 2 runes
}

// ========== GraphemeSlice Tests ==========

func TestGrapheme_Slice(t *testing.T) {
	r := New("Hello World")

	result := r.GraphemeSlice(0, 5)
	assert.Equal(t, "Hello", result.String())

	result = r.GraphemeSlice(6, 11)
	assert.Equal(t, "World", result.String())
}

func TestGrapheme_SliceCombining(t *testing.T) {
	r := New("e\u0301l\u0300o")

	result := r.GraphemeSlice(0, 2)
	assert.Equal(t, "e\u0301l\u0300", result.String())
}

// ========== Helper Method Tests ==========

func TestGrapheme_ForEach(t *testing.T) {
	r := New("hello")
	count := 0

	r.ForEachGrapheme(func(g Grapheme) {
		count++
	})

	assert.Equal(t, 5, count)
}

func TestGrapheme_Map(t *testing.T) {
	r := New("hello")

	result := r.MapGraphemes(func(g Grapheme) string {
		return g.Text + g.Text
	})

	assert.Equal(t, "hheelllloo", result.String())
}

func TestGrapheme_Filter(t *testing.T) {
	r := New("hello")

	result := r.FilterGraphemes(func(g Grapheme) bool {
		return g.Text == "l"
	})

	assert.Equal(t, "ll", result.String())
}

func TestGrapheme_Contains(t *testing.T) {
	r := New("hello")

	assert.True(t, r.ContainsGrapheme("h"))
	assert.True(t, r.ContainsGrapheme("e"))
	assert.False(t, r.ContainsGrapheme("x"))
}

func TestGrapheme_Index(t *testing.T) {
	r := New("hello")

	assert.Equal(t, 0, r.IndexGrapheme("h"))
	assert.Equal(t, 1, r.IndexGrapheme("e"))
	assert.Equal(t, 4, r.IndexGrapheme("o"))
	assert.Equal(t, -1, r.IndexGrapheme("x"))
}

func TestGrapheme_Count(t *testing.T) {
	r := New("hello")

	assert.Equal(t, 2, r.CountGrapheme("l"))
	assert.Equal(t, 1, r.CountGrapheme("h"))
	assert.Equal(t, 0, r.CountGrapheme("x"))
}

// ========== Grapheme Struct Tests ==========

func TestGraphemeStruct_Methods(t *testing.T) {
	r := New("e\u0301")
	it := r.Graphemes()
	it.Next()
	g := it.Current()

	assert.Equal(t, "e\u0301", g.String())
	assert.Equal(t, 3, g.ByteLen()) // 3 bytes: e + combining acute
	assert.Equal(t, 2, g.Len())     // 2 runes
	assert.True(t, g.IsSingleRune() == false)
	assert.False(t, g.IsASCII())
}

func TestGraphemeStruct_ASCII(t *testing.T) {
	r := New("a")
	it := r.Graphemes()
	it.Next()
	g := it.Current()

	assert.Equal(t, "a", g.String())
	assert.Equal(t, 1, g.ByteLen())
	assert.Equal(t, 1, g.Len())
	assert.True(t, g.IsSingleRune())
	assert.True(t, g.IsASCII())
}

// ========== Iterator Tests ==========

func TestGraphemeIterator_Reset(t *testing.T) {
	r := New("hello")
	it := r.Graphemes()

	// Consume all
	for it.Next() {
	}
	assert.False(t, it.HasNext())

	// Reset and consume again
	it.Reset()
	count := 0
	for it.Next() {
		count++
	}
	assert.Equal(t, 5, count)
}

func TestGraphemeIterator_Collect(t *testing.T) {
	r := New("hello")

	graphemes := r.Graphemes().Collect()
	assert.Equal(t, 5, len(graphemes))
	assert.Equal(t, "h", graphemes[0].Text)
	assert.Equal(t, "o", graphemes[4].Text)
}

func TestGraphemeIterator_HasNext(t *testing.T) {
	r := New("hello")
	it := r.Graphemes()

	assert.True(t, it.HasNext())
	it.Next()
	assert.True(t, it.HasNext())

	// Consume all
	for it.Next() {
	}
	assert.False(t, it.HasNext())
}

// ========== Edge Cases ==========

func TestGrapheme_NilRope(t *testing.T) {
	var r *Rope

	assert.Equal(t, 0, r.LenGraphemes())

	it := r.Graphemes()
	assert.False(t, it.Next())
	assert.True(t, it.exhausted)
}

func TestGrapheme_OutOfBounds(t *testing.T) {
	r := New("hello")

	assert.Panics(t, func() {
		r.GraphemeAt(-1)
	})

	assert.Panics(t, func() {
		r.GraphemeAt(100)
	})

	assert.Panics(t, func() {
		r.GraphemeSlice(-1, 3)
	})

	assert.Panics(t, func() {
		r.GraphemeSlice(0, 100)
	})
}

func TestGrapheme_ComplexUnicode(t *testing.T) {
	// Skin tone modifier + emoji
	r := New("👋🏾")
	assert.Equal(t, 1, r.LenGraphemes())

	it := r.Graphemes()
	it.Next()
	g := it.Current()
	assert.Equal(t, "👋🏾", g.Text)
}

func TestGrapheme_CRLF(t *testing.T) {
	// CRLF is treated as a single grapheme cluster according to Unicode
	r := New("hello\r\nworld")
	assert.Equal(t, 11, r.LenGraphemes()) // h,e,l,l,o,\r\n,w,o,r,l,d

	it := r.Graphemes()
	graphemes := it.Collect()
	assert.Equal(t, "\r\n", graphemes[5].Text) // CRLF is one grapheme
}

// ========== Duration Parsing Tests ==========

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // For comparison simplicity
		compare  bool   // Whether to compare formatted string (false = compare duration)
	}{
		{"seconds", "30s", "30s", true},
		{"minutes", "5m", "5m", true},
		{"hours", "2h", "2h", true},
		{"days", "1d", "1d", true},
		{"default", "60", "1m", false}, // 60 seconds = 1 minute when formatted
		{"with space", "30 s", "30s", true},
		{"plural", "5 minutes", "5m", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := ParseDuration(tt.input)
			assert.NoError(t, err)
			if tt.compare {
				assert.Equal(t, tt.expected, FormatDuration(d))
			} else {
				// For the "default" case, just verify it parses correctly
				assert.Equal(t, 60*1000000000, int(d.Nanoseconds()))
			}
		})
	}
}

func TestParseDuration_Errors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{input: ""},
		{input: "abc"},
		{input: "5x"},
		{input: "seconds"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseDuration(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"30s", "30s"},
		{"5m", "5m"},
		{"2h", "2h"},
		{"2d", "2d"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, _ := ParseDuration(tt.input)
			assert.Equal(t, tt.expected, FormatDuration(d))
		})
	}
}

// ========== Performance/Regression Tests ==========

func TestGrapheme_LargeText(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large text test in short mode")
	}

	// Create large text
	text := ""
	for i := 0; i < 1000; i++ {
		text += "hello "
	}

	r := New(text)
	count := r.LenGraphemes()
	assert.Equal(t, 6000, count) // 1000 * (5 + 1 space)
}

func TestGrapheme_MixedContent(t *testing.T) {
	// Mix of ASCII, emoji, combining marks
	r := New("Hello e\u0301 🎃👨‍👩‍👧‍👦 World")

	// Graphemes: H,e,l,l,o, ,é, ,🎃,👨‍👩‍👧‍👦, ,W,o,r,l,d
	assert.Equal(t, 16, r.LenGraphemes())

	it := r.Graphemes()
	graphemes := it.Collect()

	// Verify family emoji is single grapheme (should be at index 9)
	// Index 0-4: Hello
	// Index 5: space
	// Index 6: é
	// Index 7: space
	// Index 8: 🎃
	// Index 9: 👨‍👩‍👧‍👦
	assert.Equal(t, "👨‍👩‍👧‍👦", graphemes[9].Text)
}
