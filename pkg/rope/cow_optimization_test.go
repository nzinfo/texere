package rope

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// ============================================================================
// WriteTo Comparison Benchmarks
// ============================================================================

func BenchmarkWriteTo_Standard_Small(b *testing.B) {
	r := New("Hello World")
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		_, _ = r.WriteTo(&buf)
	}
}

func BenchmarkWriteTo_Chunked_Small(b *testing.B) {
	r := New("Hello World")
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		_, _ = r.WriteToChunked(&buf, 4096)
	}
}

func BenchmarkWriteTo_Standard_Medium(b *testing.B) {
	text := strings.Repeat("Hello World\n", 100)
	r := New(text)
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		_, _ = r.WriteTo(&buf)
	}
}

func BenchmarkWriteTo_Chunked_Medium(b *testing.B) {
	text := strings.Repeat("Hello World\n", 100)
	r := New(text)
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		_, _ = r.WriteToChunked(&buf, 4096)
	}
}

func BenchmarkWriteTo_Standard_Large(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark")
	}

	text := strings.Repeat("Hello World\n", 10000)
	r := New(text)
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		_, _ = r.WriteTo(&buf)
	}
}

func BenchmarkWriteTo_Chunked_Large(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark")
	}

	text := strings.Repeat("Hello World\n", 10000)
	r := New(text)
	var buf bytes.Buffer

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		_, _ = r.WriteToChunked(&buf, 4096)
	}
}

// ============================================================================
// Reader Comparison Benchmarks
// ============================================================================

func BenchmarkReader_Standard_Small(b *testing.B) {
	r := New("Hello World")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := r.Reader()
		buf := make([]byte, r.Size())
		_, _ = io.ReadFull(reader, buf)
	}
}

func BenchmarkReader_Standard_Medium(b *testing.B) {
	text := strings.Repeat("Hello World\n", 100)
	r := New(text)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := r.Reader()
		buf := make([]byte, r.Size())
		_, _ = io.ReadFull(reader, buf)
	}
}

func BenchmarkReader_Standard_Large(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark")
	}

	text := strings.Repeat("Hello World\n", 10000)
	r := New(text)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := r.Reader()
		buf := make([]byte, r.Size())
		_, _ = io.ReadFull(reader, buf)
	}
}

// ============================================================================
// FromReader Comparison Benchmarks
// ============================================================================

func BenchmarkFromReader_Standard_Small(b *testing.B) {
	text := "Hello World"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(text)
		_, _ = FromReader(reader)
	}
}

func BenchmarkFromReader_Optimized_Small(b *testing.B) {
	text := "Hello World"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(text)
		_, _ = FromReader(reader) // Now uses optimized implementation
	}
}

func BenchmarkFromReader_Standard_Medium(b *testing.B) {
	text := strings.Repeat("Hello World\n", 100)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(text)
		_, _ = FromReader(reader)
	}
}

func BenchmarkFromReader_Optimized_Medium(b *testing.B) {
	text := strings.Repeat("Hello World\n", 100)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(text)
		_, _ = FromReader(reader) // Now uses optimized implementation
	}
}

func BenchmarkFromReader_Standard_Large(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark")
	}

	text := strings.Repeat("Hello World\n", 10000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(text)
		_, _ = FromReader(reader)
	}
}

func BenchmarkFromReader_Optimized_Large(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark")
	}

	text := strings.Repeat("Hello World\n", 10000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(text)
		_, _ = FromReader(reader) // Now uses optimized implementation
	}
}

// ============================================================================
// ============================================================================

func BenchmarkBatchInsertOptimized_Small(b *testing.B) {
	r := New("Hello World\n")
	inserts := make([]Insertion, 5)
	for i := range inserts {
		inserts[i] = Insertion{Pos: i * 2, Text: "X"}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = r.BatchInsert(inserts)
	}
}

func BenchmarkBatchInsertOptimized_Medium(b *testing.B) {
	r := New(strings.Repeat("Line\n", 100))
	inserts := make([]Insertion, 20)
	for i := range inserts {
		inserts[i] = Insertion{Pos: i * 10, Text: "X"}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = r.BatchInsert(inserts)
	}
}

func BenchmarkBatchInsertOptimized_Large(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark")
	}

	r := New(strings.Repeat("Line\n", 1000))
	inserts := make([]Insertion, 100)
	for i := range inserts {
		inserts[i] = Insertion{Pos: i * 10, Text: "X"}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = r.BatchInsert(inserts)
	}
}

func BenchmarkBatchDeleteOptimized_Small(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 10))
	ranges := make([]Range, 5)
	for i := range ranges {
		ranges[i] = Range{Anchor: i * 10, Head: i*10 + 5}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = r.BatchDelete(ranges)
	}
}

func BenchmarkBatchDeleteOptimized_Medium(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 100))
	ranges := make([]Range, 20)
	for i := range ranges {
		ranges[i] = Range{Anchor: i * 10, Head: i*10 + 5}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = r.BatchDelete(ranges)
	}
}

func BenchmarkBatchDeleteOptimized_Large(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark")
	}

	r := New(strings.Repeat("Hello World\n", 1000))
	ranges := make([]Range, 100)
	for i := range ranges {
		ranges[i] = Range{Anchor: i * 10, Head: i*10 + 5}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = r.BatchDelete(ranges)
	}
}

// ============================================================================
// Batch vs Sequential Operations Comparison
// ============================================================================

func BenchmarkSequentialInsert_Small(b *testing.B) {
	r := New("Hello World\n")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rope := r
		for j := 0; j < 5; j++ {
			rope = rope.InsertFast(j*2, "X")
		}
	}
}

func BenchmarkSequentialInsert_Medium(b *testing.B) {
	r := New(strings.Repeat("Line\n", 100))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rope := r
		for j := 0; j < 20; j++ {
			rope = rope.InsertFast(j*10, "X")
		}
	}
}

func BenchmarkSequentialDelete_Small(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 10))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rope := r
		for j := 0; j < 5; j++ {
			rope = rope.DeleteFast(j*10, j*10+5)
		}
	}
}

func BenchmarkSequentialDelete_Medium(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 100))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rope := r
		for j := 0; j < 20; j++ {
			rope = rope.DeleteFast(j*10, j*10+5)
		}
	}
}
