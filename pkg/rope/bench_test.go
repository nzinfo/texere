package rope

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// ========== Memory Benchmarks ==========

// BenchmarkNew_String benchmarks creating a rope from a string.
func BenchmarkNew_String(b *testing.B) {
	text := strings.Repeat("Hello, World! ", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New(text)
	}
}

// BenchmarkString_Conversion benchmarks converting rope to string.
func BenchmarkString_Conversion(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.String()
	}
}

// ========== Read Operation Benchmarks ==========

func BenchmarkLength_Small(b *testing.B) {
	r := New("Hello, World!")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Length()
	}
}

func BenchmarkLength_Large(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Length()
	}
}

func BenchmarkCharAt_Random(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	pos := r.Length() / 2
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.CharAt(pos)
	}
}

func BenchmarkSlice_Middle(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	start := r.Length() / 4
	end := r.Length() * 3 / 4
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Slice(start, end)
	}
}

// ========== Iteration Benchmarks ==========

func BenchmarkIterator_Forward(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it := r.NewIterator()
		for it.Next() {
			_ = it.Current()
		}
	}
}

func BenchmarkForEach_Char(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ForEach(func(ch rune) {
			_ = ch
		})
	}
}

func BenchmarkForEachByte(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ForEachByte(func(b byte) bool {
			_ = b
			return true
		})
	}
}

// ========== Write Operation Benchmarks ==========

func BenchmarkInsert_Small_Beginning(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := "INSERTED"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Insert(0, text)
	}
}

func BenchmarkInsert_Small_End(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := "INSERTED"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Insert(r.Length(), text)
	}
}

func BenchmarkInsert_Small_Middle(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := "INSERTED"
	pos := r.Length() / 2
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Insert(pos, text)
	}
}

func BenchmarkDelete_Small(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Delete(10, 20)
	}
}

func BenchmarkReplace_Small(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	replacement := "REPLACED"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Replace(10, 20, replacement)
	}
}

// ========== Concatenation Benchmarks ==========

func BenchmarkAppend_Small(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := " Appended"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Append(text)
	}
}

func BenchmarkAppendRope_Small(b *testing.B) {
	r1 := New(strings.Repeat("Hello, World! ", 100))
	r2 := New(" Appended")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r1.AppendRope(r2)
	}
}

func BenchmarkConcat_Two(b *testing.B) {
	r1 := New(strings.Repeat("Hello, World! ", 100))
	r2 := New(strings.Repeat("Hello, World! ", 100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Concat(r1, r2)
	}
}

func BenchmarkConcat_Multiple(b *testing.B) {
	ropes := []*Rope{
		New(strings.Repeat("Hello, World! ", 25)),
		New(strings.Repeat("Hello, World! ", 25)),
		New(strings.Repeat("Hello, World! ", 25)),
		New(strings.Repeat("Hello, World! ", 25)),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Concat(ropes...)
	}
}

// ========== Builder Benchmarks ==========

func BenchmarkBuilder_Append(b *testing.B) {
	text := "Hello, World! "
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder()
		for j := 0; j < 100; j++ {
			builder.Append(text)
		}
		_ = builder.Build()
	}
}

func BenchmarkBuilder_Append_vs_Insert(b *testing.B) {
	text := "Hello, World! "

	b.Run("Append", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder := NewBuilder()
			for j := 0; j < 100; j++ {
				builder.Append(text)
			}
			_ = builder.Build()
		}
	})

	b.Run("Insert", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r := Empty()
			for j := 0; j < 100; j++ {
				r = r.Insert(r.Length(), text)
			}
		}
	})
}

// ========== Memory Allocation Benchmarks ==========

func BenchmarkAllocations_Insert(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 100))
	text := "INSERTED"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Insert(r.Length()/2, text)
	}
}

func BenchmarkAllocations_String(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.String()
	}
}

func BenchmarkAllocations_Clone(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Clone()
	}
}

// ========== Chunk Operation Benchmarks ==========

func BenchmarkChunks_Iteration(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it := r.Chunks()
		for it.Next() {
			_ = it.Current()
		}
	}
}

func BenchmarkChunkCount(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.ChunkCount()
	}
}

// ========== Line Operation Benchmarks ==========

func BenchmarkLineCount(b *testing.B) {
	r := New(strings.Repeat("Hello, World!\n", 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.LineCount()
	}
}

func BenchmarkLineAtChar(b *testing.B) {
	r := New(strings.Repeat("Hello, World!\n", 1000))
	pos := r.Length() / 2
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.LineAtChar(pos)
	}
}

// ========== Comparison with String ==========

func BenchmarkComparison_Rope_vs_String_Append(b *testing.B) {
	text := "Hello, World!"

	b.Run("Rope_Builder", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder := NewBuilder()
			for j := 0; j < 100; j++ {
				builder.Append(text)
			}
			_ = builder.Build()
		}
	})

	b.Run("String_Append", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var s string
			for j := 0; j < 100; j++ {
				s += text
			}
			_ = s
		}
	})

	b.Run("StringsBuilder", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			for j := 0; j < 100; j++ {
				sb.WriteString(text)
			}
			_ = sb.String()
		}
	})
}

// ========== Stress Tests ==========

func BenchmarkStress_MillionChars(b *testing.B) {
	// Only run this benchmark explicitly
	b.Skip("Stress test - run with -bench=Stress")

	large := strings.Repeat("a", 1000000)
	r := New(large)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Length()
		_ = r.CharAt(500000)
		_ = r.Slice(100000, 900000)
	}
}

// ========== Parallel Benchmarks ==========

func BenchmarkParallel_String(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = r.String()
		}
	})
}

func BenchmarkParallel_Length(b *testing.B) {
	r := New(strings.Repeat("Hello, World! ", 1000))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = r.Length()
		}
	})
}

// ========== Custom Benchmark Helpers ==========

// benchmarkMemoryUsage creates a custom memory usage benchmark.
func benchmarkMemoryUsage(b *testing.B, fn func()) {
	var m runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&m)
	beforeAlloc := m.TotalAlloc
	beforeMallocs := m.Mallocs

	for i := 0; i < b.N; i++ {
		fn()
	}

	runtime.ReadMemStats(&m)
	b.ReportMetric(float64(m.TotalAlloc-beforeAlloc), "alloc/bytes")
	b.ReportMetric(float64(m.Mallocs-beforeMallocs), "mallocs")
}

// ========== Benchmark Runner ==========

// RunFullBenchmarkSuite runs all benchmarks and prints results.
func RunFullBenchmarkSuite() {
	tests := []struct {
		name string
		fn   func(*testing.B)
	}{
		{"New_String", BenchmarkNew_String},
		{"String_Conversion", BenchmarkString_Conversion},
		{"Insert_Middle", BenchmarkInsert_Small_Middle},
		{"Append_Small", BenchmarkAppend_Small},
		{"Iterator_Forward", BenchmarkIterator_Forward},
		{"Chunks_Iteration", BenchmarkChunks_Iteration},
		{"Builder_Append", BenchmarkBuilder_Append},
	}

	fmt.Println("=== Running Full Benchmark Suite ===")
	fmt.Println()

	for _, test := range tests {
		fmt.Printf("Running %s...\n", test.name)
		result := testing.Benchmark(test.fn)

		// Parse and display results
		fmt.Printf("  %s\n", result.String())
		fmt.Printf("  ns/op: %d\n", result.NsPerOp())
		fmt.Printf("  B/op: %d\n", result.AllocedBytesPerOp())
		fmt.Printf("  allocs/op: %d\n", result.AllocsPerOp())
		fmt.Println()
	}
}
