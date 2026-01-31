package rope

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// ============================================================================
// SplitOff Performance Baseline
// ============================================================================

func BenchmarkSplitOff_Small(b *testing.B) {
	r := New(strings.Repeat("Hello World ", 10))
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = r.SplitOff(r.Length() / 2)
	}
}

func BenchmarkSplitOff_Medium(b *testing.B) {
	r := New(strings.Repeat("Hello World ", 100))
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = r.SplitOff(r.Length() / 2)
	}
}

func BenchmarkSplitOff_Large(b *testing.B) {
	r := New(strings.Repeat("Hello World ", 1000))
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = r.SplitOff(r.Length() / 2)
	}
}

// ============================================================================
// Stream I/O Performance Baseline
// ============================================================================

func BenchmarkFromReader_Small(b *testing.B) {
	text := strings.Repeat("Hello World\n", 10)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(text)
		_, _ = FromReader(reader)
	}
}

func BenchmarkFromReader_Medium(b *testing.B) {
	text := strings.Repeat("Hello World\n", 100)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(text)
		_, _ = FromReader(reader)
	}
}

func BenchmarkFromReader_Large(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark in short mode")
	}

	text := strings.Repeat("Hello World\n", 10000)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(text)
		_, _ = FromReader(reader)
	}
}

func BenchmarkWriteTo_Small(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 10))
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = r.WriteTo(&buf)
	}
}

func BenchmarkWriteTo_Medium(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 100))
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = r.WriteTo(&buf)
	}
}

func BenchmarkWriteTo_Large(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark in short mode")
	}

	r := New(strings.Repeat("Hello World\n", 10000))
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_, _ = r.WriteTo(&buf)
	}
}

func BenchmarkRopeReader_Small(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 10))
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := r.Reader()
		_, _ = io.ReadAll(reader)
	}
}

func BenchmarkRopeReader_Medium(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 100))
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := r.Reader()
		_, _ = io.ReadAll(reader)
	}
}

func BenchmarkRopeReader_Large(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark in short mode")
	}

	r := New(strings.Repeat("Hello World\n", 10000))
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := r.Reader()
		_, _ = io.ReadAll(reader)
	}
}

// ============================================================================
// History Hooks Performance Baseline
// ============================================================================

func BenchmarkHookManager_Trigger(b *testing.B) {
	hm := NewHookManager()
	r := New("Hello World")

	// Register 5 hooks
	for i := 0; i < 5; i++ {
		hm.Register(HookBeforeEdit, "hook", 10, func(ctx *HookContext) error {
			return nil
		})
	}

	edit := &EditInfo{Operation: "insert", Length: 5}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = hm.TriggerBeforeEdit(r, edit)
	}
}

func BenchmarkHookManager_TriggerWithPriority(b *testing.B) {
	hm := NewHookManager()
	r := New("Hello World")

	// Register hooks with different priorities
	hm.Register(HookBeforeEdit, "low", 1, func(ctx *HookContext) error {
		return nil
	})
	hm.Register(HookBeforeEdit, "high", 10, func(ctx *HookContext) error {
		return nil
	})
	hm.Register(HookBeforeEdit, "medium", 5, func(ctx *HookContext) error {
		return nil
	})

	edit := &EditInfo{Operation: "insert", Length: 5}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = hm.TriggerBeforeEdit(r, edit)
	}
}

func BenchmarkBuiltinHook_TrackMetrics(b *testing.B) {
	hm := NewHookManager()
	r := New("Hello World")
	builtin := DefaultBuiltinHooks()
	metrics := &EditMetrics{}

	hm.Register(HookAfterEdit, "metrics", 10, builtin.TrackMetrics(metrics))

	edit := &EditInfo{Operation: "insert", Length: 5}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		hm.TriggerAfterEdit(r, edit)
	}
}

func BenchmarkHashCode32(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 100))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = r.HashCode32()
	}
}

func BenchmarkHashCode64(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 100))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = r.HashCode64()
	}
}
