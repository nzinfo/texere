package rope

import (
	"fmt"
	"strings"
	"testing"
)

// ========== Insert Benchmarks ==========

// BenchmarkInsert_Standard for comparison.
func BenchmarkInsert_Standard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)
	pos := r.Length() / 2
	insertText := "INSERTED"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Insert(pos, insertText)
	}
}

// BenchmarkInsert_Optimized for comparison.
func BenchmarkInsert_Optimized(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)
	pos := r.Length() / 2
	insertText := "INSERTED"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.InsertOptimized(pos, insertText)
	}
}

// ========== Delete Benchmarks ==========

func BenchmarkDelete_Standard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Delete(10, 20)
	}
}

func BenchmarkDelete_Optimized(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.DeleteOptimized(10, 20)
	}
}

// ========== Mixed Operation Benchmarks ==========

func BenchmarkMixedOps_Standard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = r.Append(" X")
		r = r.Insert(r.Length()/2, " Y")
		r = r.Delete(0, 1)
	}
}

// ========== Sequential Operations ==========

func BenchmarkSequentialInserts_Standard(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := New("")
		for j := 0; j < 100; j++ {
			r = r.Append(fmt.Sprintf("Item %d ", j))
		}
	}
}

// ========== Large Text Benchmarks ==========

func BenchmarkInsert_Large_Standard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 1000) // ~26KB
	r := New(text)
	pos := r.Length() / 2
	insertText := strings.Repeat("X", 1000) // 1KB insert

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Insert(pos, insertText)
	}
}

// ========== Copy-on-Write Benchmarks ==========

func BenchmarkCowRope_Insert(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := NewCowRope(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Insert(r.Length()/2, "X")
	}
}

func BenchmarkCowRope_Delete(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := NewCowRope(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Delete(10, 20)
	}
}

func BenchmarkCowRope_ShareAndMutate(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := NewCowRope(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r1 := r.Insert(10, "X")
		r2 := r.Insert(20, "Y")
		r3 := r.Insert(30, "Z")
		_ = r1
		_ = r2
		_ = r3
	}
}

// ========== Rebalancing Benchmarks ==========

func BenchmarkRebalance_Balanced(b *testing.B) {
	// Create a balanced rope
	ropes := make([]*Rope, 100)
	for i := 0; i < 100; i++ {
		ropes[i] = New(fmt.Sprintf("Chunk %d ", i))
	}
	r := Concat(ropes...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Balance()
	}
}

func BenchmarkRebalance_Unbalanced(b *testing.B) {
	// Create an unbalanced rope (left-skewed)
	r := New("")
	for i := 0; i < 100; i++ {
		r = r.Append(fmt.Sprintf("Chunk %d ", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Balance()
	}
}

// ========== Memory Allocation Benchmarks ==========

func BenchmarkAllocations_InsertStandard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	r := New(text)
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = r.Insert(r.Length()/2, "X")
	}
}

func BenchmarkAllocations_DeleteStandard(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := New(text)
		r = r.Delete(10, 20)
	}
}

// ========== Comparison Tests ==========

func TestCompareImplementations(t *testing.T) {
	text := strings.Repeat("Hello, World! ", 50)

	// Test Insert
	r1 := New(text).Insert(50, "INSERTED")
	r2 := New(text).InsertOptimized(50, "INSERTED")

	if r1.String() != r2.String() {
		t.Error("Insert implementations differ")
	}

	// Test Delete
	r1 = New(text).Delete(10, 20)
	r2 = New(text).DeleteOptimized(10, 20)

	if r1.String() != r2.String() {
		t.Error("Delete implementations differ")
	}

	// Test Append
	r1 = New(text).Append("APPENDED")
	result := r1.String()

	if !strings.HasSuffix(result, "APPENDED") {
		t.Error("Append failed to add text at end")
	}

	t.Log("All implementations produce identical results ✓")
}

// ========== Stress Tests ==========

func TestStress_ManyInserts(t *testing.T) {
	r := New("")
	for i := 0; i < 1000; i++ {
		r = r.Insert(i, fmt.Sprintf("%d", i%10))
	}
	expectedLen := 1000
	if r.Length() != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, r.Length())
	}
}

func TestStress_ManyDeletes(t *testing.T) {
	// Build rope
	r := New("")
	for i := 0; i < 1000; i++ {
		r = r.Append(fmt.Sprintf("%d", i%10))
	}

	// Delete from beginning
	for i := 0; i < 500; i++ {
		r = r.Delete(0, 1)
	}

	if r.Length() != 500 {
		t.Errorf("Expected length 500, got %d", r.Length())
	}
}
