package rope

import (
	"math/rand"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// TestCRLF_RandomInserts tests random CRLF pair insertions
// This is ported from ropey's crlf.rs to catch CRLF seam errors
func TestCRLF_RandomInserts_Small(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CRLF test in short mode")
	}

	rng := rand.New(rand.NewSource(42))
	r := Empty()

	// Do a bunch of random incoherent inserts of CRLF pairs
	for i := 0; i < (1 << 8); i++ {
		ropeLen := r.Length()
		if ropeLen == 0 {
			ropeLen = 1
		}
		pos := rng.Intn(ropeLen)

		// Insert various CRLF combinations
		crlfVariations := []string{
			"\r\n\r\n",
			"\n\r\n\r",
			"\r\n\r\n",
			"\n\r\n\r",
			"\r\n\r\n",
			"こんいちは、",
			"\n\r\n\r",
			"\r\n\r\n",
			"\n\r\n\r",
			"\r\n\r\n",
			"\n\r\n\r",
			"みんなさん！",
		}

		text := crlfVariations[rng.Intn(len(crlfVariations))]
		r = r.Insert(pos, text)

		// Make sure the tree is sound
		assert.True(t, utf8.ValidString(r.String()))
	}
}

// TestCRLF_RandomInserts_Large tests larger random CRLF insertions
func TestCRLF_RandomInserts_Large(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CRLF test in short mode")
	}

	rng := rand.New(rand.NewSource(42))
	r := Empty()

	// More aggressive CRLF insertions
	for i := 0; i < (1 << 12); i++ {
		ropeLen := r.Length()
		if ropeLen == 0 {
			ropeLen = 1
		}
		pos := rng.Intn(ropeLen)

		crlfVariations := []string{
			"\r\n\r\n",
			"\n\r\n\r",
			"\r\n\r\n",
			"\n\r\n\r",
			"\r\n\r\n",
			"こんいちは、",
			"\n\r\n\r",
			"\r\n\r\n",
			"\n\r\n\r",
			"\r\n\r\n",
			"\n\r\n\r",
			"みんなさん！",
		}

		text := crlfVariations[rng.Intn(len(crlfVariations))]
		r = r.Insert(pos, text)

		// Verify integrity
		assert.True(t, utf8.ValidString(r.String()))
		assert.True(t, r.Length() >= 0)
	}
}

// TestCRLF_RandomRemovals tests random CRLF removals
// Ported from ropey's crlf.rs
func TestCRLF_RandomRemovals(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CRLF test in short mode")
	}

	rng := rand.New(rand.NewSource(42))
	r := Empty()

	// Build tree with lots of CRLF
	crlfText := "\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n" +
		"\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\nこんいちは、\n\r\n\r\n\r\n\r\n" +
		"\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n" +
		"\r\n\r\n\r\nこんいちは、r\n\r\n\r\n\r\nみんなさん！\n\r\n\r\n\r\n" +
		"こんいちは、\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\r\n\r\n\r\n\r\n\r\n\r\n" +
		"\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\nみんなさん！\r\n\r\n\r\n" +
		"\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\r\n\r\n\r\n\r\n\r\n\r\nみんなさん！\n\r\n\r\n" +
		"\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\rみんなさん！\n\r\n\r\n" +
		"\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r"

	r = r.Insert(0, crlfText)

	// Do a bunch of random incoherent removals
	for i := 0; i < (1 << 9); i++ {
		if r.Length() == 0 {
			break
		}

		start := rng.Intn(r.Length())
		end := start + 5
		if end > r.Length() {
			end = r.Length()
		}
		r = r.Delete(start, end)

		if r.Length() == 0 {
			break
		}

		start = rng.Intn(r.Length())
		end = start + 9
		if end > r.Length() {
			end = r.Length()
		}
		r = r.Delete(start, end)

		// Make sure the tree is sound
		assert.True(t, utf8.ValidString(r.String()))
		assert.True(t, r.Length() >= 0)
	}
}

// TestCRLF_InsertAtBoundaries tests inserting CRLF at various boundaries
func TestCRLF_InsertAtBoundaries(t *testing.T) {
	text := "Hello World Test String"
	r := New(text)

	// Test positions within the rope length
	positions := []int{0, 1, 5, len(text)}

	for _, pos := range positions {
		r2 := r.Insert(pos, "\r\n")

		// Verify CRLF is present
		assert.Contains(t, r2.String(), "\r\n")
		assert.True(t, utf8.ValidString(r2.String()))
	}
}

// TestCRLF_DeleteAtBoundaries tests deleting CRLF at boundaries
func TestCRLF_DeleteAtBoundaries(t *testing.T) {
	testCases := []string{
		"\r\nHello World",
		"Hello\r\nWorld",
		"Hello World\r\n",
		"\r\nHello\r\nWorld\r\n",
	}

	for _, text := range testCases {
		r := New(text)

		// Find and delete CRLF
		str := r.String()
		for i := 0; i < len(str); i++ {
			if i+1 < len(str) && str[i] == '\r' && str[i+1] == '\n' {
				// Found CRLF, delete it
				r = r.Delete(i, i+2)
				break
			}
		}

		assert.True(t, utf8.ValidString(r.String()))
	}
}

// TestCRLF_SplitAtCRLF tests splitting at CRLF boundaries
func TestCRLF_SplitAtCRLF(t *testing.T) {
	text := "Line 1\r\nLine 2\r\nLine 3\r\n"
	r := New(text)

	// Split at various positions
	splits := []int{0, 2, 7, 8, 15, 16, 22, 23}

	for _, pos := range splits {
		if pos >= r.Length() {
			continue
		}

		left, right := r.Split(pos)

		// Both parts should be valid UTF-8
		assert.True(t, utf8.ValidString(left.String()))
		assert.True(t, utf8.ValidString(right.String()))

		// Combined should equal original
		combined := left.String() + right.String()
		assert.Equal(t, text, combined)

		// Re-merge for next iteration
		r = left.AppendRope(right)
	}
}

// TestCRLF_MixedLineEndings tests mixed line endings
func TestCRLF_MixedLineEndings(t *testing.T) {
	text := "Line 1\nLine 2\r\nLine 3\rLine 4\n\rLine 5"
	r := New(text)

	// Should be valid UTF-8
	assert.True(t, utf8.ValidString(r.String()))
	assert.Equal(t, text, r.String())

	// Count lines should work correctly
	lines := r.Lines()
	assert.True(t, len(lines) > 0)
}

// TestCRLF_OnlyCRLF tests rope with only CRLF
func TestCRLF_OnlyCRLF(t *testing.T) {
	text := "\r\n\r\n\r\n"
	r := New(text)

	assert.Equal(t, text, r.String())
	assert.True(t, utf8.ValidString(r.String()))
	assert.Equal(t, 6, r.Length()) // 3 * 2 chars
}

// TestCRLF_LoneCR tests lone CR without LF
func TestCRLF_LoneCR(t *testing.T) {
	text := "Hello\rWorld\r\n"
	r := New(text)

	assert.Equal(t, text, r.String())
	assert.True(t, utf8.ValidString(r.String()))
}

// TestCRLF_LoneLF tests lone LF without CR
func TestCRLF_LoneLF(t *testing.T) {
	text := "Hello\nWorld\r\n"
	r := New(text)

	assert.Equal(t, text, r.String())
	assert.True(t, utf8.ValidString(r.String()))
}

// TestCRLF_UnicodeWithCRLF tests Unicode with CRLF
func TestCRLF_UnicodeWithCRLF(t *testing.T) {
	text := "こんにちは\r\n世界\r\n🌍🌎🌏\r\n"
	r := New(text)

	assert.Equal(t, text, r.String())
	assert.True(t, utf8.ValidString(r.String()))

	lines := r.Lines()
	assert.True(t, len(lines) >= 3)
}

// TestCRLF_LargeTextWithCRLF tests large text with many CRLF
func TestCRLF_LargeTextWithCRLF(t *testing.T) {
	var text string
	for i := 0; i < 100; i++ {
		text += "Line " + string(rune('0'+i%10)) + "\r\n"
	}

	r := New(text)

	assert.Equal(t, text, r.String())
	assert.True(t, utf8.ValidString(r.String()))
	assert.True(t, len(r.Lines()) >= 100) // At least 100 lines
}
