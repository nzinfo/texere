package rope

import (
	"fmt"
	"strings"
	"testing"
)

// ========== Fast Path Benchmarks ==========

func BenchmarkInsertFast_SingleLeaf(b *testing.B) {
	// Single leaf rope (fast path)
	r := New("Hello, World!")
	text := " INSERTED"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.InsertFast(7, text)
	}
}

func BenchmarkInsertFast_Beginning(b *testing.B) {
	r := New("Hello, World!")
	text := "Prepended "

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.InsertFast(0, text)
	}
}

func BenchmarkInsertFast_End(b *testing.B) {
	r := New("Hello, World!")
	text := " Appended"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.InsertFast(r.Length(), text)
	}
}

func BenchmarkInsertFast_EmptyText(b *testing.B) {
	r := New("Hello, World!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.InsertFast(5, "")
	}
}

func BenchmarkDeleteFast_SingleLeaf(b *testing.B) {
	r := New("Hello, World!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.DeleteFast(7, 12)
	}
}

func BenchmarkDeleteFast_Beginning(b *testing.B) {
	r := New("Hello, World!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.DeleteFast(0, 7)
	}
}

func BenchmarkDeleteFast_End(b *testing.B) {
	r := New("Hello, World!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.DeleteFast(7, r.Length())
	}
}

func BenchmarkDeleteFast_All(b *testing.B) {
	r := New("Hello, World!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.DeleteFast(0, r.Length())
	}
}

func BenchmarkSliceFast_Full(b *testing.B) {
	r := New("Hello, World!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.SliceFast(0, r.Length())
	}
}

func BenchmarkSliceFast_SingleLeaf(b *testing.B) {
	r := New("Hello, World!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.SliceFast(7, 12)
	}
}

// ========== Comparison: Fast vs Standard ==========

func BenchmarkCompare_InsertFast_vs_Standard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 10)
	r := New(text)
	pos := r.Length() / 2
	insertText := "X"

	b.Run("Fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = r.InsertFast(pos, insertText)
		}
	})

	b.Run("Standard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = r.Insert(pos, insertText)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = r.InsertOptimized(pos, insertText)
		}
	})
}

func BenchmarkCompare_DeleteFast_vs_Standard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 10)
	r := New(text)

	b.Run("Fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = r.DeleteFast(10, 20)
		}
	})

	b.Run("Standard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = r.Delete(10, 20)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = r.DeleteOptimized(10, 20)
		}
	})
}

// ========== ASCII Fast Path Benchmarks ==========

func BenchmarkFindBytePos_ASCII(b *testing.B) {
	text := "Hello, World! This is ASCII text."
	pos := 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = findBytePosInString(text, pos)
	}
}

func BenchmarkFindBytePos_UTF8(b *testing.B) {
	text := "Hello, 世界! This has UTF-8."
	pos := 5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = findBytePosInString(text, pos)
	}
}

func BenchmarkIsASCII_True(b *testing.B) {
	text := "Hello, World! Pure ASCII text here."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsASCII(text)
	}
}

func BenchmarkIsASCII_False(b *testing.B) {
	text := "Hello, 世界! UTF-8 text here."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsASCII(text)
	}
}

// ========== Batch Operation Benchmarks ==========

func BenchmarkBatchInsert_Small(b *testing.B) {
	r := New("Hello, World!")
	inserts := []Insertion{
		{Pos: 0, Text: "Start"},
		{Pos: 5, Text: "Mid"},
		{Pos: 10, Text: "End"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.BatchInsert(inserts)
	}
}

func BenchmarkBatchInsert_vs_Sequential(b *testing.B) {
	r := New("Hello, World!")
	inserts := []Insertion{
		{Pos: 0, Text: "A"},
		{Pos: 5, Text: "B"},
		{Pos: 10, Text: "C"},
	}

	b.Run("Batch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = r.BatchInsert(inserts)
		}
	})

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := r
			for _, ins := range inserts {
				result = result.InsertFast(ins.Pos, ins.Text)
			}
			_ = result
		}
	})
}

func BenchmarkBatchDelete_Small(b *testing.B) {
	r := New("Hello, World! This is a test string.")
	ranges := []Range{
		NewRange(0, 5),
		NewRange(10, 15),
		NewRange(20, 25),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.BatchDelete(ranges)
	}
}

// ========== String Benchmarks ==========

func BenchmarkString_Small(b *testing.B) {
	r := New("Hello, World!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.String()
	}
}

func BenchmarkString_Medium(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.String()
	}
}

func BenchmarkString_Large(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.String()
	}
}

// ========== Micro-Operation Benchmarks ==========

func BenchmarkAppend_ASCII(b *testing.B) {
	r := New("Hello")
	text := " World"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Append(text)
	}
}

func BenchmarkPrepend_ASCII(b *testing.B) {
	r := New("World")
	text := "Hello "

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Prepend(text)
	}
}

// ========== Cache Efficiency Benchmarks ==========

func BenchmarkCacheEfficiency_SequentialAccess(b *testing.B) {
	// Create rope with many small leaves
	ropes := make([]*Rope, 100)
	for i := 0; i < 100; i++ {
		ropes[i] = New(fmt.Sprintf("Chunk %d ", i))
	}
	r := Concat(ropes...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Sequential access (cache-friendly)
		// Rope length is ~800 chars (100 * "Chunk xx "), so use 80 char slices
		for j := 0; j < 10; j++ {
			_ = r.Slice(j*80, (j+1)*80)
		}
	}
}

func BenchmarkCacheEfficiency_RandomAccess(b *testing.B) {
	// Create rope with many small leaves
	ropes := make([]*Rope, 100)
	for i := 0; i < 100; i++ {
		ropes[i] = New(fmt.Sprintf("Chunk %d ", i))
	}
	r := Concat(ropes...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Random access (cache-unfriendly)
		// Rope length is ~800 chars (100 * "Chunk xx "), stay within bounds
		_ = r.Slice(100, 200)
		_ = r.Slice(500, 600)
		_ = r.Slice(700, 750)
		_ = r.Slice(100, 200)
		_ = r.Slice(600, 700)
	}
}

// ========== Correctness Tests ==========

func TestFastPaths_Correctness(t *testing.T) {
	text := "Hello, World!"
	r := New(text)

	// Test InsertFast
	result := r.InsertFast(7, "INSERTED")
	expected := "Hello, INSERTEDWorld!"
	if result.String() != expected {
		t.Errorf("InsertFast failed: got %q, want %q", result.String(), expected)
	}

	// Test DeleteFast - delete "World" (positions 7-12)
	result = r.DeleteFast(7, 12)
	expectedStr := "Hello, !"
	if result.String() != expectedStr {
		t.Errorf("DeleteFast failed: got %q, want %q", result.String(), expectedStr)
	}

	// Test SliceFast
	sliceResult := r.SliceFast(7, 12)
	expectedSlice := "World"
	if sliceResult != expectedSlice {
		t.Errorf("SliceFast failed: got %q, want %q", sliceResult, expectedSlice)
	}

	// Test Append
	r2 := New("Hello, World!")
	result = r2.Append("!")
	expected = "Hello, World!!"
	if result.String() != expected {
		t.Errorf("Append failed: got %q, want %q", result.String(), expected)
	}

	// Test Prepend
	result = r2.Prepend("Say: ")
	expected = "Say: Hello, World!"
	if result.String() != expected {
		t.Errorf("Prepend failed: got %q, want %q", result.String(), expected)
	}

	t.Log("All fast path operations are correct ✓")
}

func TestBatchOperations_Correctness(t *testing.T) {
	r := New("Hello, World!")

	// Test BatchInsert
	// Note: positions are relative to original rope
	inserts := []Insertion{
		{Pos: 0, Text: "Start-"},
		{Pos: 7, Text: "-Middle-"},
	}
	result := r.BatchInsert(inserts)
	expected := "Start-Hello, -Middle-World!"
	if result.String() != expected {
		t.Errorf("BatchInsert failed: got %q, want %q", result.String(), expected)
	}

	// Test BatchDelete
	// "ABCDEFGH" -> delete positions 0-2 (AB) and 4-6 (EF) -> "CDGH"
	r2 := New("ABCDEFGH")
	ranges := []Range{
		NewRange(0, 2),
		NewRange(4, 6),
	}
	result = r2.BatchDelete(ranges)
	expected = "CDGH"
	if result.String() != expected {
		t.Errorf("BatchDelete failed: got %q, want %q", result.String(), expected)
	}

	t.Log("All batch operations are correct ✓")
}

func TestFindBytePosInString_Correctness(t *testing.T) {
	// ASCII
	pos := findBytePosInString("Hello", 2)
	if pos != 2 {
		t.Errorf("ASCII failed: got %d, want %d", pos, 2)
	}

	// UTF-8
	pos = findBytePosInString("Hello, 世界", 7)
	if pos != 7 {
		t.Errorf("UTF-8 failed: got %d, want %d", pos, 7)
	}

	pos = findBytePosInString("Hello, 世界", 8)
	if pos != 10 { // '世' is 3 bytes
		t.Errorf("UTF-8 char position failed: got %d, want %d", pos, 10)
	}

	t.Log("findBytePosInString is correct ✓")
}

// ========== Stress Tests ==========

func TestStress_FastPaths_ManyOperations(t *testing.T) {
	r := New("")

	// Many fast path appends
	for i := 0; i < 100; i++ {
		r = r.Append(fmt.Sprintf("%d", i%10))
	}

	if r.Length() != 100 {
		t.Errorf("Expected length 100, got %d", r.Length())
	}

	// Many fast path deletes
	for i := 0; i < 50; i++ {
		r = r.DeleteFast(0, 1)
	}

	if r.Length() != 50 {
		t.Errorf("Expected length 50, got %d", r.Length())
	}

	t.Log("Stress test passed ✓")
}
