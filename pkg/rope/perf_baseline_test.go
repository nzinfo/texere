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
// Enhanced SavePoint Performance Baseline
// ============================================================================

func BenchmarkEnhancedSavePoint_Create(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 100))
	metadata := SavePointMetadata{
		UserID:      "user1",
		ViewID:      "view1",
		Tags:        []string{"checkpoint", "important"},
		Description: "Test savepoint",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = NewEnhancedSavePoint(r, i, metadata)
	}
}

func BenchmarkEnhancedSavePoint_HasTag(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 100))
	metadata := SavePointMetadata{
		Tags: []string{"tag1", "tag2", "tag3", "tag4", "tag5"},
	}
	sp := NewEnhancedSavePoint(r, 1, metadata)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = sp.HasTag("tag3")
	}
}

func BenchmarkEnhancedSavePoint_Metadata(b *testing.B) {
	r := New(strings.Repeat("Hello World\n", 100))
	metadata := SavePointMetadata{
		UserID:      "user1",
		ViewID:      "view1",
		Tags:        []string{"tag1", "tag2", "tag3"},
		Description: "Test savepoint",
	}
	sp := NewEnhancedSavePoint(r, 1, metadata)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = sp.Metadata()
	}
}

func BenchmarkEnhancedSavePointManager_Create(b *testing.B) {
	sm := NewEnhancedSavePointManager()
	r := New(strings.Repeat("Hello World\n", 100))
	metadata := SavePointMetadata{
		UserID: "user1",
		Tags:   []string{"checkpoint"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sm.Create(r, i, metadata)
	}
}

func BenchmarkEnhancedSavePointManager_Query(b *testing.B) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow)

	r := New(strings.Repeat("Hello World\n", 100))

	// Create 100 savepoints
	for i := 0; i < 100; i++ {
		metadata := SavePointMetadata{
			UserID: "user1",
			Tags:   []string{"checkpoint", "important"},
		}
		sm.Create(r, i, metadata)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = sm.ByUser("user1", 10)
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

// ============================================================================
// Hash Performance Baseline
// ============================================================================

func BenchmarkHashToString(b *testing.B) {
	hash := uint32(0x12345678)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = HashToString(hash)
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
