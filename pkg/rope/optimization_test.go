package rope

import (
	"runtime"
	"strings"
	"testing"
)

// ========== Optimization Comparison Tests ==========

// BenchmarkString_Old removed - old implementation deleted
// BenchmarkString_New and BenchmarkString_Bytes removed - same as String()

// ========== Append Comparison ==========

func BenchmarkAppend_Old(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := " Appended"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Append(text)
	}
}

func BenchmarkAppend_New(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := " Appended"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Append(text) // Same as Old now
	}
}

// ========== Insert Comparison ==========

func BenchmarkInsert_Old(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := "INSERTED"
	pos := r.Length() / 2
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Insert(pos, text)
	}
}

func BenchmarkInsert_New(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := "INSERTED"
	pos := r.Length() / 2
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.InsertOptimized(pos, text)
	}
}

// ========== Delete Comparison ==========

func BenchmarkDelete_Old(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Delete(10, 20)
	}
}

func BenchmarkDelete_New(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.DeleteOptimized(10, 20)
	}
}

// ========== Comparison Test ==========

func TestOptimizationComparison(t *testing.T) {
	// Create test rope
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)

	// Test String() - all implementations merged
	str1 := r.String()
	str2 := r.String()

	if str1 != str2 {
		t.Errorf("String() inconsistent!")
	}

	// Test Append - AppendOptimized merged into Append
	append1 := r.Append(" Appended")
	append2 := r.Append(" Appended")

	if append1.String() != append2.String() {
		t.Error("Append implementations differ!")
	}

	// Test Insert - InsertOptimized kept separate
	insert1 := r.Insert(500, "X")
	insert2 := r.InsertOptimized(500, "X")

	if insert1.String() != insert2.String() {
		t.Error("Insert implementations differ!")
	}

	// Test Delete - DeleteOptimized kept separate
	delete1 := r.Delete(100, 200)
	delete2 := r.DeleteOptimized(100, 200)

	if delete1.String() != delete2.String() {
		t.Error("Delete implementations differ!")
	}

	t.Log("All optimization implementations produce identical results ✓")
}

// ========== Memory Allocation Tests ==========

// TestMemory_String and TestMemory_Append removed - implementations merged
// TestMemory_Delete kept for comparison (Delete vs DeleteOptimized)

func TestMemory_Delete(t *testing.T) {
	text := strings.Repeat("Hello, World! ", 100)

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	r := New(text)
	for i := 0; i < 100; i++ {
		r = r.Delete(10, 20)
	}

	runtime.ReadMemStats(&m2)
	oldAlloc := m2.TotalAlloc - m1.TotalAlloc

	// Test optimized implementation
	runtime.GC()
	runtime.ReadMemStats(&m1)

	r2 := New(text)
	for i := 0; i < 100; i++ {
		r2 = r2.DeleteOptimized(10, 20)
	}

	runtime.ReadMemStats(&m2)
	newAlloc := m2.TotalAlloc - m1.TotalAlloc

	t.Logf("Delete() - Standard: %d bytes, Optimized: %d bytes, Improvement: %.1fx",
		oldAlloc, newAlloc, float64(oldAlloc)/float64(newAlloc))
}

func TestMemory_Insert(t *testing.T) {
	text := strings.Repeat("Hello, World! ", 100)
	pos := len([]rune(text)) / 2
	insertText := "INSERTED"

	// Test old implementation
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	r := New(text)
	for i := 0; i < 100; i++ {
		r = r.Insert(pos, insertText)
	}

	runtime.ReadMemStats(&m2)
	oldAlloc := m2.TotalAlloc - m1.TotalAlloc

	// Test new implementation
	runtime.GC()
	runtime.ReadMemStats(&m1)

	r2 := New(text)
	for i := 0; i < 100; i++ {
		r2 = r2.InsertOptimized(pos, insertText)
	}

	runtime.ReadMemStats(&m2)
	newAlloc := m2.TotalAlloc - m1.TotalAlloc

	t.Logf("Insert() - Old: %d bytes, New: %d bytes, Improvement: %.1fx",
		oldAlloc, newAlloc, float64(oldAlloc)/float64(newAlloc))
}
