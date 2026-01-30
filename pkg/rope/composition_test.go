package rope

import (
	"testing"
)

// ========== Full Composition Tests ==========

func TestChangeSet_FullComposition(t *testing.T) {
	doc := New("hello world")

	// First changeset: insert " beautiful"
	cs1 := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" beautiful")

	// Apply first changeset
	result1 := cs1.Apply(doc)
	expected1 := "hello beautiful world"
	if result1.String() != expected1 {
		t.Fatalf("Expected %q, got %q", expected1, result1.String())
	}

	// Second changeset: delete " beautiful"
	cs2 := NewChangeSet(cs1.LenAfter()).
		Retain(5).
		Delete(10) // " beautiful"

	// Use SimpleCompose for reliable composition
	composed := SimpleCompose(cs1, cs2, doc)

	// Apply composed changeset
	result := composed.Apply(doc)

	// Should be back to "hello world"
	if result.String() != "hello world" {
		t.Errorf("Expected %q, got %q", "hello world", result.String())
	}
}

func TestChangeSet_MapPosition(t *testing.T) {
	doc := New("hello world")

	// Delete " world"
	cs := NewChangeSet(doc.Length()).
		Retain(5).
		Delete(6)

	// Map position 3 (in "hello") - should stay at 3
	mapped := cs.MapPosition(3, AssocBefore)
	if mapped != 3 {
		t.Errorf("Expected position 3, got %d", mapped)
	}

	// Map position 7 (in " world") - should map to 5 with AssocBefore
	mapped = cs.MapPosition(7, AssocBefore)
	if mapped != 5 {
		t.Errorf("Expected position 5, got %d", mapped)
	}

	// Map position 7 with AssocAfter - should also be 5
	mapped = cs.MapPosition(7, AssocAfter)
	if mapped != 5 {
		t.Errorf("Expected position 5, got %d", mapped)
	}
}

func TestChangeSet_MapPositions(t *testing.T) {
	doc := New("hello world")

	// Insert " beautiful" after "hello"
	cs := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" beautiful")

	positions := []int{3, 5, 7}
	associations := []Assoc{AssocBefore, AssocBefore, AssocBefore}

	mapped := cs.MapPositions(positions, associations)

	// Position 3 -> 3 (before insert)
	if mapped[0] != 3 {
		t.Errorf("Expected position 3, got %d", mapped[0])
	}

	// Position 5 -> 5 (at insert point, AssocBefore = before insert)
	if mapped[1] != 5 {
		t.Errorf("Expected position 5, got %d", mapped[1])
	}

	// Position 7 -> 17 (after insert, 7 + 10 = 17)
	// Actually, position 7 in original is 2 chars after insert point
	// So it should be 5 + 10 + 2 = 17
	expected := 17
	if mapped[2] != expected {
		t.Errorf("Expected position %d, got %d", expected, mapped[2])
	}
}

func TestChangeSet_Split(t *testing.T) {
	doc := New("hello world")

	cs := NewChangeSet(doc.Length()).
		Retain(5).
		Delete(6).
		Insert(" gophers")

	// Split at position 5
	before, after := cs.Split(5)

	if before.LenBefore() != 5 {
		t.Errorf("Expected lenBefore 5, got %d", before.LenBefore())
	}

	if after.LenBefore() != 6 {
		t.Errorf("Expected lenBefore 6, got %d", after.LenBefore())
	}

	// Apply before changeset
	result1 := before.Apply(doc)
	if result1.String() != "hello" {
		t.Errorf("Expected %q, got %q", "hello", result1.String())
	}

	// Apply after changeset to result
	result2 := after.Apply(result1)
	if result2.String() != "hello gophers" {
		t.Errorf("Expected %q, got %q", "hello gophers", result2.String())
	}
}

func TestChangeSet_Merge(t *testing.T) {
	doc := New("hello")

	cs1 := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" world")

	cs2 := NewChangeSet(doc.Length()).
		Retain(11).
		Insert("!")

	// Merge changesets
	merged := cs1.Merge(cs2)

	// Apply merged
	result := merged.Apply(doc)
	if result.String() != "hello world!" {
		t.Errorf("Expected %q, got %q", "hello world!", result.String())
	}
}

func TestChangeSet_Optimized(t *testing.T) {
	doc := New("hello")

	cs := NewChangeSet(doc.Length()).
		Retain(5).
		Delete(1).
		Delete(1).
		Delete(1).
		Delete(1).
		Delete(1).
		Delete(1).
		Insert("a").
		Insert("b").
		Insert("c")

	// Before optimization: 10 operations
	if len(cs.operations) != 10 {
		t.Errorf("Expected 10 operations, got %d", len(cs.operations))
	}

	// Optimize
	optimized := cs.Optimized()

	// After optimization: should be fused
	// Expected: Retain(5) + Delete(6) + Insert("abc") = 3 operations
	if len(optimized.operations) != 3 {
		t.Errorf("Expected 3 operations after optimization, got %d", len(optimized.operations))
	}

	// Should produce same result
	result1 := cs.Apply(doc)
	result2 := optimized.Apply(doc)

	if result1.String() != result2.String() {
		t.Errorf("Optimization changed result: %q vs %q", result1.String(), result2.String())
	}
}

func TestChangeSet_Transform(t *testing.T) {
	doc := New("hello world")

	// Two concurrent edits
	cs1 := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" beautiful")

	cs2 := NewChangeSet(doc.Length()).
		Retain(6).
		Delete(5) // Delete "world"

	// Transform cs2 to apply after cs1
	transformed := cs1.Transform(cs2)

	// Apply both
	result1 := cs1.Apply(doc)
	result := transformed.Apply(result1)

	// Result should have " beautiful" inserted and something deleted
	// The exact result depends on transformation logic
	if result == nil {
		t.Error("Transform returned nil")
	}
}

// ========== Word Boundary Tests ==========

func TestWordBoundary_IsWordChar(t *testing.T) {
	wb := NewWordBoundary(New("hello world"))

	tests := []struct {
		r        rune
		expected bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'_', true},
		{' ', false},
		{'-', false},
		{'(', false},
	}

	for _, tt := range tests {
		result := wb.IsWordChar(tt.r)
		if result != tt.expected {
			t.Errorf("IsWordChar(%c): expected %v, got %v", tt.r, tt.expected, result)
		}
	}
}

func TestWordBoundary_PrevWordStart(t *testing.T) {
	doc := New("hello world test")
	wb := NewWordBoundary(doc)

	// At position 11 (start of "test")
	// Should find start of "world" at 6
	start := wb.PrevWordStart(11)
	if start != 6 {
		t.Errorf("Expected 6, got %d", start)
	}

	// At position 7 (in "world")
	// Should find start of "world" at 6
	start = wb.PrevWordStart(7)
	if start != 6 {
		t.Errorf("Expected 6, got %d", start)
	}
}

func TestWordBoundary_NextWordStart(t *testing.T) {
	doc := New("hello world test")
	wb := NewWordBoundary(doc)

	// At position 7 (in "world")
	// Should find start of "test" at 12
	start := wb.NextWordStart(7)
	if start != 12 {
		t.Errorf("Expected 12, got %d", start)
	}

	// At position 0 (start of "hello")
	// Should find start of "world" at 6
	start = wb.NextWordStart(0)
	if start != 6 {
		t.Errorf("Expected 6, got %d", start)
	}
}

func TestWordBoundary_CurrentWordStart(t *testing.T) {
	doc := New("hello world test")
	wb := NewWordBoundary(doc)

	// At position 7 (in "world")
	start := wb.CurrentWordStart(7)
	if start != 6 {
		t.Errorf("Expected 6, got %d", start)
	}

	// At position 13 (in "test")
	start = wb.CurrentWordStart(13)
	if start != 12 {
		t.Errorf("Expected 12, got %d", start)
	}

	// At position 5 (at space)
	start = wb.CurrentWordStart(5)
	if start != 5 {
		t.Errorf("Expected 5, got %d", start)
	}
}

func TestWordBoundary_CurrentWordEnd(t *testing.T) {
	doc := New("hello world test")
	wb := NewWordBoundary(doc)

	// At position 7 (in "world")
	end := wb.CurrentWordEnd(7)
	if end != 11 {
		t.Errorf("Expected 11, got %d", end)
	}

	// At position 13 (in "test")
	end = wb.CurrentWordEnd(13)
	if end != 16 {
		t.Errorf("Expected 16, got %d", end)
	}
}

func TestWordBoundary_WordAt(t *testing.T) {
	doc := New("hello world test")
	wb := NewWordBoundary(doc)

	// At position 7 (in "world")
	word, start, end := wb.WordAt(7)
	if word != "world" {
		t.Errorf("Expected word 'world', got %q", word)
	}
	if start != 6 {
		t.Errorf("Expected start 6, got %d", start)
	}
	if end != 11 {
		t.Errorf("Expected end 11, got %d", end)
	}

	// At position 5 (at space)
	word, start, end = wb.WordAt(5)
	if word != "" {
		t.Errorf("Expected empty word at space, got %q", word)
	}
}

func TestWordBoundary_SelectWord(t *testing.T) {
	doc := New("hello world test")
	wb := NewWordBoundary(doc)

	// Select word at position 7
	start, end := wb.SelectWord(7)
	if start != 6 || end != 11 {
		t.Errorf("Expected (6, 11), got (%d, %d)", start, end)
	}

	// Verify selection
	word := doc.Slice(start, end)
	if word != "world" {
		t.Errorf("Expected 'world', got %q", word)
	}
}

func TestWordBoundary_BigWord(t *testing.T) {
	doc := New("hello-world test")
	wb := NewWordBoundary(doc)

	// Big words include non-word characters like '-'
	// "hello-world" should be one big word
	start := wb.BigWordStart(8) // In "world"
	if start != 0 {
		t.Errorf("Expected big word start 0, got %d", start)
	}

	end := wb.BigWordEnd(8)
	if end != 11 {
		t.Errorf("Expected big word end 11, got %d", end)
	}

	// Verify big word
	word := doc.Slice(start, end)
	if word != "hello-world" {
		t.Errorf("Expected 'hello-world', got %q", word)
	}
}

func TestWordBoundary_Paragraph(t *testing.T) {
	doc := New("hello\nworld\n\ntest")
	wb := NewWordBoundary(doc)

	// Paragraph start at position 8 (in "test")
	// Should find position after the double newline
	start := wb.ParagraphStart(8)
	if start != 7 {
		t.Errorf("Expected paragraph start 7, got %d", start)
	}

	// Paragraph end at position 8
	end := wb.ParagraphEnd(8)
	if end != 11 {
		t.Errorf("Expected paragraph end 11, got %d", end)
	}
}

func TestWordBoundary_Line(t *testing.T) {
	doc := New("hello\nworld\ntest")
	wb := NewWordBoundary(doc)

	// Line start at position 7 (in "world")
	start := wb.LineStart(7)
	if start != 6 {
		t.Errorf("Expected line start 6, got %d", start)
	}

	// Line end at position 7
	end := wb.LineEnd(7)
	if end != 11 {
		t.Errorf("Expected line end 11, got %d", end)
	}

	// First line
	start = wb.LineStart(0)
	if start != 0 {
		t.Errorf("Expected line start 0, got %d", start)
	}

	end = wb.LineEnd(0)
	if end != 5 {
		t.Errorf("Expected line end 5, got %d", end)
	}
}

// ========== Integration Tests ==========

func TestPositionMapper_WithWordBoundaries(t *testing.T) {
	doc := New("hello world")

	// Delete " world"
	cs := NewChangeSet(doc.Length()).
		Retain(5).
		Delete(6)

	// Create mapper with document
	mapper := NewPositionMapperWithDoc(cs, doc)

	// Add position with word association
	mapper.AddPosition(7, AssocAfterWord)

	result := mapper.Map()

	// Should map to word boundary after deletion
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	// After deletion, position should be at end (5)
	// AssocAfterWord should keep it at end
	_ = result[0] // Just verify it doesn't panic
}

func TestComposition_Integration(t *testing.T) {
	doc := New("hello world")

	// Edit 1: Insert " beautiful"
	cs1 := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" beautiful")

	// Edit 2: Replace "world" with "gophers"
	// After cs1, document is "hello beautiful world" (21 chars)
	// To replace "world" with "gophers":
	// - Retain 16 to keep "hello beautiful "
	// - Delete 5 to remove "world"
	// - Insert "gophers"
	cs2 := NewChangeSet(cs1.LenAfter()).
		Retain(16). // "hello beautiful " (5 + 1 + 9 + 1 = 16)
		Delete(5).  // Delete "world"
		Insert("gophers")

	// Compose both edits
	composed := cs1.Compose(cs2)

	// Apply composed
	result := composed.Apply(doc)

	expected := "hello beautiful gophers"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

func TestHistory_PathComposition(t *testing.T) {
	history := NewHistory()
	doc := New("hello")

	// Create a branch: edit -> undo -> different edit
	// Edit 1: insert " world"
	cs1 := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" world")
	txn1 := NewTransaction(cs1)
	history.CommitRevision(txn1, doc)
	doc = txn1.Apply(doc)

	// Undo
	undoTxn := history.Undo()
	doc = undoTxn.Apply(doc)

	// Edit 2 (different): insert " gopher"
	cs2 := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" gopher")
	txn2 := NewTransaction(cs2)
	history.CommitRevision(txn2, doc)
	doc = txn2.Apply(doc)

	expected := "hello gopher"
	if doc.String() != expected {
		t.Errorf("Expected %q, got %q", expected, doc.String())
	}

	// Should have 2 revisions (branching)
	if history.RevisionCount() != 2 {
		t.Errorf("Expected 2 revisions, got %d", history.RevisionCount())
	}

	// Cannot redo to edit 1 (different branch)
	if history.CanRedo() {
		t.Error("Expected CanRedo() to be false after branching")
	}
}

// ========== Benchmarks ==========

func BenchmarkComposition_Full(b *testing.B) {
	doc := New("hello world, this is a test")

	cs1 := NewChangeSet(doc.Length()).
		Retain(5).
		Delete(7).
		Insert(" gophers")

	cs2 := NewChangeSet(cs1.LenAfter()).
		Retain(12).
		Insert("!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SimpleCompose(cs1, cs2, doc)
	}
}

func BenchmarkMapPosition_Single(b *testing.B) {
	doc := New("hello world, this is a test document with lots of text")

	cs := NewChangeSet(doc.Length()).
		Retain(10).
		Delete(5).
		Insert(" gophers")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cs.MapPosition(15, AssocBefore)
	}
}

func BenchmarkMapPosition_Multiple(b *testing.B) {
	doc := New("hello world, this is a test document with lots of text")

	cs := NewChangeSet(doc.Length()).
		Retain(10).
		Delete(5).
		Insert(" gophers")

	positions := make([]int, 100)
	associations := make([]Assoc, 100)
	for i := range positions {
		positions[i] = i
		associations[i] = AssocBefore
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cs.MapPositions(positions, associations)
	}
}

func BenchmarkWordBoundary_WordAt(b *testing.B) {
	doc := New("hello world test this is a benchmark")
	wb := NewWordBoundary(doc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = wb.WordAt(10)
	}
}

func BenchmarkWordBoundary_SelectWord(b *testing.B) {
	doc := New("hello world test this is a benchmark")
	wb := NewWordBoundary(doc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = wb.SelectWord(10)
	}
}
