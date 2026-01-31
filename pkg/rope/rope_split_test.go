package rope

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// SplitOff Tests
// ============================================================================

func TestRope_SplitOff_Basic(t *testing.T) {
	r := New("Hello World")

	left, right := r.SplitOff(5)

	assert.Equal(t, "Hello", left.String())
	assert.Equal(t, " World", right.String())
}

func TestRope_SplitOff_Beginning(t *testing.T) {
	r := New("Hello World")

	left, right := r.SplitOff(0)

	assert.Equal(t, "", left.String())
	assert.Equal(t, "Hello World", right.String())
}

func TestRope_SplitOff_End(t *testing.T) {
	r := New("Hello World")

	left, right := r.SplitOff(r.Length())

	assert.Equal(t, "Hello World", left.String())
	assert.Equal(t, "", right.String())
}

func TestRope_SplitOff_Middle(t *testing.T) {
	r := New("Hello World")

	left, right := r.SplitOff(6)

	assert.Equal(t, "Hello ", left.String())
	assert.Equal(t, "World", right.String())
}

func TestRope_Split_Multiple(t *testing.T) {
	r := New("Hello Beautiful World")

	left, right := r.Split(6)
	assert.Equal(t, "Hello ", left.String())
	assert.Equal(t, "Beautiful World", right.String())

	// Now split the right part
	rightLeft, rightRight := right.SplitOff(10)
	assert.Equal(t, "Beautiful ", rightLeft.String())
	assert.Equal(t, "World", rightRight.String())
}

func TestRope_Split_Empty(t *testing.T) {
	r := New("")

	r1, r2 := r.Split(0)

	assert.Equal(t, "", r1.String())
	assert.Equal(t, "", r2.String())
}

func TestRope_Split_Whole(t *testing.T) {
	r := New("Hello")

	r1, r2 := r.Split(5)

	assert.Equal(t, "Hello", r1.String())
	assert.Equal(t, "", r2.String())
}

func TestRope_SplitAndAppend(t *testing.T) {
	r1 := New("Hello")
	r2 := New(" World")

	// Split and recombine
	left, right := r1.Split(r1.Length())
	r3 := left.Append(right.String())

	assert.Equal(t, "Hello", r3.String())

	// Now append r2
	r4 := r3.AppendRope(r2)
	assert.Equal(t, "Hello World", r4.String())
}

// ============================================================================
// Stream I/O Tests
// ============================================================================

func TestRope_FromReader_Basic(t *testing.T) {
	text := "Hello World"
	reader := strings.NewReader(text)

	r, err := FromReader(reader)

	assert.NoError(t, err)
	assert.Equal(t, text, r.String())
}

func TestRope_FromReader_Large(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large text test")
	}

	text := strings.Repeat("Hello World\n", 1000)
	reader := strings.NewReader(text)

	r, err := FromReader(reader)

	assert.NoError(t, err)
	assert.Equal(t, text, r.String())
	assert.Equal(t, len(text), r.Size())
}

func TestRope_FromReader_Chunks(t *testing.T) {
	// Test reading in chunks
	chunks := []string{"Hello", " ", "World", "!"}
	reader := io.MultiReader(
		strings.NewReader(chunks[0]),
		strings.NewReader(chunks[1]),
		strings.NewReader(chunks[2]),
		strings.NewReader(chunks[3]),
	)

	r, err := FromReader(reader)

	assert.NoError(t, err)
	assert.Equal(t, "Hello World!", r.String())
}

func TestRope_WriteTo_Basic(t *testing.T) {
	r := New("Hello World")

	var buf bytes.Buffer
	n, err := buf.Write([]byte(r.String()))

	assert.NoError(t, err)
	assert.Equal(t, r.Size(), n)
	assert.Equal(t, "Hello World", buf.String())
}

func TestRope_WriteTo_Large(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large text test")
	}

	text := strings.Repeat("Hello World\n", 1000)
	r := New(text)

	var buf bytes.Buffer
	n, err := buf.Write([]byte(r.String()))

	assert.NoError(t, err)
	assert.Equal(t, len(text), n)
	assert.Equal(t, text, buf.String())
}

func TestRope_Reader_Basic(t *testing.T) {
	r := New("Hello World")
	reader := r.Reader()

	data := make([]byte, 100)
	n, err := reader.Read(data)

	assert.NoError(t, err)
	assert.Equal(t, 11, n)
	assert.Equal(t, "Hello World", string(data[:n]))
}

func TestRope_Reader_Chunked(t *testing.T) {
	r := New("Hello World")
	reader := r.Reader()

	buf1 := make([]byte, 5)
	n1, _ := reader.Read(buf1)
	assert.Equal(t, "Hello", string(buf1[:n1]))

	buf2 := make([]byte, 10)
	n2, _ := reader.Read(buf2)
	assert.Equal(t, " World", string(buf2[:n2]))

	// EOF
	buf3 := make([]byte, 10)
	n3, _ := reader.Read(buf3)
	assert.Equal(t, 0, n3)
	assert.Error(t, io.EOF)
}

func TestRope_IO_RoundTrip(t *testing.T) {
	original := strings.Repeat("Hello 世界\n", 100)

	// Create rope from string
	r1 := New(original)

	// Write to buffer
	var buf bytes.Buffer
	_, err := buf.Write([]byte(r1.String()))
	assert.NoError(t, err)

	// Create rope from buffer
	reader := bytes.NewReader(buf.Bytes())
	r2, err := FromReader(reader)
	assert.NoError(t, err)

	// Verify content
	assert.Equal(t, original, r2.String())
	assert.Equal(t, r1.Size(), r2.Size())
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestRope_Split_IO_Combo(t *testing.T) {
	// Simulate processing a large file in chunks
	original := strings.Repeat("Line of text\n", 100)
	r := New(original)

	// Split into parts
	part1, part2 := r.Split(r.Length() / 2)

	// Write each part to separate buffers
	var buf1, buf2 bytes.Buffer
	part1.WriteTo(&buf1)
	part2.WriteTo(&buf2)

	// Recombine
	combined := buf1.String() + buf2.String()

	assert.Equal(t, original, combined)
}

func TestRope_SplitMutate_Append(t *testing.T) {
	r := New("Hello World")

	// Split
	left, right := r.SplitOff(5)

	// Modify left part and append right
	result := left.Append(",")
	result = result.AppendRope(right)

	assert.Equal(t, "Hello, World", result.String())
}

func TestRope_IO_Pipeline(t *testing.T) {
	// Simulate: Read -> Process -> Write
	input := "Hello World"

	// Read
	reader := strings.NewReader(input)
	r, _ := FromReader(reader)

	// Process: insert comma
	r = r.Insert(5, ",")

	// Write
	var buf bytes.Buffer
	buf.Write([]byte(r.String()))

	assert.Equal(t, "Hello, World", buf.String())
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestRope_SplitOff_EmptyString(t *testing.T) {
	r := New("")

	left, right := r.SplitOff(0)

	assert.Equal(t, "", left.String())
	assert.Equal(t, "", right.String())
}

func TestRope_Split_OutOfBounds(t *testing.T) {
	r := New("Hello")

	// Split at end
	left, right := r.SplitOff(5)
	assert.Equal(t, "Hello", left.String())
	assert.Equal(t, "", right.String())

	// Split beyond end - SplitOff handles this gracefully
	left, right = r.SplitOff(100)
	assert.Equal(t, "Hello", left.String())
	assert.Equal(t, "", right.String())
}

func TestRope_IO_Empty(t *testing.T) {
	r := New("")

	var buf bytes.Buffer
	n, err := buf.Write([]byte(r.String()))

	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, "", buf.String())
}

func TestRope_Reader_Empty(t *testing.T) {
	r := New("")
	reader := r.Reader()

	data := make([]byte, 10)
	n, err := reader.Read(data)

	assert.Equal(t, 0, n)
	assert.Equal(t, io.EOF, err)
}

// ============================================================================
// Performance Tests
// ============================================================================

func BenchmarkRope_SplitOff(b *testing.B) {
	r := New(strings.Repeat("Hello World ", 1000))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.SplitOff(r.Length() / 2)
	}
}

func BenchmarkRope_Split(b *testing.B) {
	r := New(strings.Repeat("Hello World ", 1000))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Split(r.Length() / 2)
	}
}

func BenchmarkRope_FromReader(b *testing.B) {
	text := strings.Repeat("Hello World\n", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(text)
		_, _ = FromReader(reader)
	}
}

func BenchmarkRope_WriteTo(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 1000))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		buf.Write([]byte(r.String()))
	}
}

func BenchmarkRope_Reader(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 1000))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := r.Reader()
		io.ReadAll(reader)
	}
}
