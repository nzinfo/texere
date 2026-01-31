package rope

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
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
// Query Comparison Benchmarks
// ============================================================================

func setupQueryManager() *EnhancedSavePointManager {
	manager := NewEnhancedSavePointManager()

	// Create 100 savepoints with various metadata
	for i := 0; i < 100; i++ {
		rope := New(strings.Repeat("Content ", i%10+1))
		metadata := SavePointMetadata{
			UserID:      "user-" + string(rune('A'+i%5)),
			Description: "Savepoint " + string(rune('0'+i%10)),
		}

		// Add tags
		if i%3 == 0 {
			metadata.Tags = append(metadata.Tags, "important")
		}
		if i%5 == 0 {
			metadata.Tags = append(metadata.Tags, "backup")
		}

		manager.Create(rope, i, metadata)
	}

	return manager
}

func BenchmarkQuery_Standard(b *testing.B) {
	manager := setupQueryManager()
	userID := "user-A"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = manager.Query(SavePointQuery{
			UserID: &userID,
			Limit:  10,
		})
	}
}

func BenchmarkQuery_Optimized(b *testing.B) {
	manager := setupQueryManager()
	userID := "user-A"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = manager.QueryOptimized(SavePointQuery{
			UserID: &userID,
			Limit:  10,
		})
	}
}

func BenchmarkQuery_Preallocated(b *testing.B) {
	manager := setupQueryManager()
	userID := "user-A"
	results := make([]SavePointResult, 0, 16)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		results = manager.QueryPreallocated(SavePointQuery{
			UserID: &userID,
			Limit:  10,
		}, results)
	}
}

func BenchmarkQuery_Concurrent_Standard(b *testing.B) {
	manager := setupQueryManager()
	userID := "user-A"

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = manager.Query(SavePointQuery{
				UserID: &userID,
				Limit:  10,
			})
		}
	})
}

func BenchmarkQuery_Concurrent_Optimized(b *testing.B) {
	manager := setupQueryManager()
	userID := "user-A"

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = manager.QueryOptimized(SavePointQuery{
				UserID: &userID,
				Limit:  10,
			})
		}
	})
}

func BenchmarkQuery_Concurrent_Preallocated(b *testing.B) {
	manager := setupQueryManager()
	userID := "user-A"

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		results := make([]SavePointResult, 0, 16)
		for pb.Next() {
			results = manager.QueryPreallocated(SavePointQuery{
				UserID: &userID,
				Limit:  10,
			}, results)
		}
	})
}

// ============================================================================
// Query with Time Filter Comparison
// ============================================================================

func BenchmarkQueryByTime_Standard(b *testing.B) {
	manager := setupQueryManager()
	now := time.Now()
	startTime := now.Add(-time.Hour)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = manager.ByTime(startTime, now, 10)
	}
}

func BenchmarkQueryByTime_Optimized(b *testing.B) {
	manager := setupQueryManager()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		now := time.Now()
		startTime := now.Add(-time.Hour)
		_ = manager.QueryOptimized(SavePointQuery{
			StartTime: &startTime,
			EndTime:   &now,
			Limit:     10,
		})
	}
}

// ============================================================================
// Query with Tag Filter Comparison
// ============================================================================

func BenchmarkQueryByTag_Standard(b *testing.B) {
	manager := setupQueryManager()
	tag := "important"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = manager.ByTag(tag, 10)
	}
}

func BenchmarkQueryByTag_Optimized(b *testing.B) {
	manager := setupQueryManager()
	tag := "important"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = manager.QueryOptimized(SavePointQuery{
			Tag:   &tag,
			Limit: 10,
		})
	}
}

// ============================================================================
// Batch Operations Comparison (with efficient sorting)
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

