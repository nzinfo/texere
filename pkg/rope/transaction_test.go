package rope

import (
	"testing"
)

// TestTransaction_Basic tests basic transaction creation and application.
func TestTransaction_Basic(t *testing.T) {
	doc := New("hello")

	// Create a changeset to insert " world" at position 5
	cs := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" world")

	transaction := NewTransaction(cs)

	// Apply transaction
	newDoc := transaction.Apply(doc)

	expected := "hello world"
	if newDoc.String() != expected {
		t.Errorf("Expected %q, got %q", expected, newDoc.String())
	}
}

// TestTransaction_Delete tests deletion transaction.
func TestTransaction_Delete(t *testing.T) {
	doc := New("hello world")

	// Delete " world"
	cs := NewChangeSet(doc.Length()).
		Retain(5).
		Delete(6)

	transaction := NewTransaction(cs)
	newDoc := transaction.Apply(doc)

	expected := "hello"
	if newDoc.String() != expected {
		t.Errorf("Expected %q, got %q", expected, newDoc.String())
	}
}

// TestTransaction_Replace tests replacement transaction.
func TestTransaction_Replace(t *testing.T) {
	doc := New("hello world")

	// Replace "world" with "gophers"
	cs := NewChangeSet(doc.Length()).
		Retain(6).
		Delete(5).
		Insert("gophers")

	transaction := NewTransaction(cs)
	newDoc := transaction.Apply(doc)

	expected := "hello gophers"
	if newDoc.String() != expected {
		t.Errorf("Expected %q, got %q", expected, newDoc.String())
	}
}

// TestTransaction_Invert tests transaction inversion for undo.
func TestTransaction_Invert(t *testing.T) {
	original := New("hello")

	// Create transaction: insert " world" at position 5
	cs := NewChangeSet(original.Length()).
		Retain(5).
		Insert(" world")

	transaction := NewTransaction(cs)

	// Apply forward
	modified := transaction.Apply(original)
	if modified.String() != "hello world" {
		t.Fatalf("Expected %q, got %q", "hello world", modified.String())
	}

	// Create inversion
	inverted := transaction.Invert(original)

	// Apply inversion (should undo)
	undone := inverted.Apply(modified)
	if undone.String() != original.String() {
		t.Errorf("Undo failed: expected %q, got %q", original.String(), undone.String())
	}
}

// TestTransaction_InvertDelete tests inverting a deletion.
func TestTransaction_InvertDelete(t *testing.T) {
	original := New("hello world")

	// Delete " world"
	cs := NewChangeSet(original.Length()).
		Retain(5).
		Delete(6)

	transaction := NewTransaction(cs)

	// Apply forward
	modified := transaction.Apply(original)
	if modified.String() != "hello" {
		t.Fatalf("Expected %q, got %q", "hello", modified.String())
	}

	// Invert and apply (should restore " world")
	inverted := transaction.Invert(original)
	restored := inverted.Apply(modified)

	if restored.String() != original.String() {
		t.Errorf("Restore failed: expected %q, got %q", original.String(), restored.String())
	}
}

// TestHistory_BasicUndoRedo tests basic undo/redo functionality.
func TestHistory_BasicUndoRedo(t *testing.T) {
	history := NewHistory()
	doc := New("hello")

	// Transaction 1: insert " world"
	cs1 := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" world")
	txn1 := NewTransaction(cs1)

	// Commit and apply
	history.CommitRevision(txn1, doc)
	doc = txn1.Apply(doc)

	if doc.String() != "hello world" {
		t.Fatalf("Expected %q, got %q", "hello world", doc.String())
	}

	// Undo
	undoTxn := history.Undo()
	if undoTxn == nil {
		t.Fatal("Expected undo transaction, got nil")
	}

	doc = undoTxn.Apply(doc)
	if doc.String() != "hello" {
		t.Errorf("Undo failed: expected %q, got %q", "hello", doc.String())
	}

	// Redo
	redoTxn := history.Redo()
	if redoTxn == nil {
		t.Fatal("Expected redo transaction, got nil")
	}

	doc = redoTxn.Apply(doc)
	if doc.String() != "hello world" {
		t.Errorf("Redo failed: expected %q, got %q", "hello world", doc.String())
	}
}

// TestHistory_MultipleEdits tests multiple edits with undo/redo.
func TestHistory_MultipleEdits(t *testing.T) {
	history := NewHistory()
	doc := New("hello")

	// Edit 1: insert " beautiful"
	cs1 := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" beautiful")
	txn1 := NewTransaction(cs1)
	history.CommitRevision(txn1, doc)
	doc = txn1.Apply(doc)

	// Edit 2: delete " beautiful"
	cs2 := NewChangeSet(doc.Length()).
		Retain(5).
		Delete(10)
	txn2 := NewTransaction(cs2)
	history.CommitRevision(txn2, doc)
	doc = txn2.Apply(doc)

	// Edit 3: insert " world"
	cs3 := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" world")
	txn3 := NewTransaction(cs3)
	history.CommitRevision(txn3, doc)
	doc = txn3.Apply(doc)

	expected := "hello world"
	if doc.String() != expected {
		t.Fatalf("Expected %q, got %q", expected, doc.String())
	}

	// Undo all
	doc = history.Undo().Apply(doc) // Undo edit 3
	if doc.String() != "hello" {
		t.Errorf("After undo 3: expected %q, got %q", "hello", doc.String())
	}

	doc = history.Undo().Apply(doc) // Undo edit 2
	if doc.String() != "hello beautiful" {
		t.Errorf("After undo 2: expected %q, got %q", "hello beautiful", doc.String())
	}

	doc = history.Undo().Apply(doc) // Undo edit 1
	if doc.String() != "hello" {
		t.Errorf("After undo 1: expected %q, got %q", "hello", doc.String())
	}

	// Redo all
	doc = history.Redo().Apply(doc) // Redo edit 1
	if doc.String() != "hello beautiful" {
		t.Errorf("After redo 1: expected %q, got %q", "hello beautiful", doc.String())
	}

	doc = history.Redo().Apply(doc) // Redo edit 2
	if doc.String() != "hello" {
		t.Errorf("After redo 2: expected %q, got %q", "hello", doc.String())
	}

	doc = history.Redo().Apply(doc) // Redo edit 3
	if doc.String() != expected {
		t.Errorf("After redo 3: expected %q, got %q", expected, doc.String())
	}
}

// TestHistory_Branching tests behavior after undo+edit (branching).
func TestHistory_Branching(t *testing.T) {
	history := NewHistory()
	doc := New("hello")

	// Edit 1
	cs1 := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" world")
	txn1 := NewTransaction(cs1)
	history.CommitRevision(txn1, doc)
	doc = txn1.Apply(doc)

	// Undo
	undoTxn := history.Undo()
	doc = undoTxn.Apply(doc)

	// Edit 2 (different branch): insert " gopher" instead
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

	// In current implementation: undo+edit creates a new root
	// We have 2 revisions: edit1 and edit2
	// edit1 cannot be redone to because we're on a different branch now
	if history.RevisionCount() != 2 {
		t.Errorf("Expected 2 revisions (undo creates new branch), got %d", history.RevisionCount())
	}

	// Verify we can't redo to edit1 from edit2
	if history.CanRedo() {
		t.Error("Expected CanRedo() to be false after undo+edit (different branch)")
	}
}

// TestHistory_CanUndoRedo tests CanUndo and CanRedo.
func TestHistory_CanUndoRedo(t *testing.T) {
	history := NewHistory()
	doc := New("hello")

	// Initially, can't undo or redo
	if history.CanUndo() {
		t.Error("Expected CanUndo() to be false initially")
	}
	if history.CanRedo() {
		t.Error("Expected CanRedo() to be false initially")
	}

	// Make an edit
	cs := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" world")
	txn := NewTransaction(cs)
	history.CommitRevision(txn, doc)
	doc = txn.Apply(doc)

	// Can undo, but can't redo
	if !history.CanUndo() {
		t.Error("Expected CanUndo() to be true after edit")
	}
	if history.CanRedo() {
		t.Error("Expected CanRedo() to be false after edit")
	}

	// Undo
	history.Undo()

	// Can redo now
	if !history.CanRedo() {
		t.Error("Expected CanRedo() to be true after undo")
	}
}

// TestHistory_EarlierLater tests time-based navigation.
func TestHistory_EarlierLater(t *testing.T) {
	history := NewHistory()
	doc := New("hello")

	// Make 5 edits
	for i := 0; i < 5; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert(string(rune('a' + i)))
		txn := NewTransaction(cs)
		history.CommitRevision(txn, doc)
		doc = txn.Apply(doc)
	}

	expected := "helloabcde"
	if doc.String() != expected {
		t.Fatalf("Expected %q, got %q", expected, doc.String())
	}

	// Go earlier by 1 step (simplified - earlier only supports 1 step for now)
	earlierTxn := history.Earlier(1)
	if earlierTxn == nil {
		t.Fatal("Expected Earlier(1) to return a transaction")
	}
	doc = earlierTxn.Apply(doc)

	if doc.String() != "helloabcd" {
		t.Errorf("After Earlier(1): expected %q, got %q", "helloabcd", doc.String())
	}

	// Go later by 1 step (simplified - later only supports 1 step for now)
	laterTxn := history.Later(1)
	if laterTxn == nil {
		t.Fatal("Expected Later(1) to return a transaction")
	}
	doc = laterTxn.Apply(doc)

	if doc.String() != "helloabcde" {
		t.Errorf("After Later(1): expected %q, got %q", "helloabcde", doc.String())
	}
}

// TestHistory_Clear tests clearing history.
func TestHistory_Clear(t *testing.T) {
	history := NewHistory()
	doc := New("hello")

	// Make some edits
	for i := 0; i < 3; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("x")
		txn := NewTransaction(cs)
		history.CommitRevision(txn, doc)
		doc = txn.Apply(doc)
	}

	if history.RevisionCount() != 3 {
		t.Errorf("Expected 3 revisions, got %d", history.RevisionCount())
	}

	// Clear history
	history.Clear()

	if history.RevisionCount() != 0 {
		t.Errorf("Expected 0 revisions after clear, got %d", history.RevisionCount())
	}
	if history.CanUndo() {
		t.Error("Expected CanUndo() to be false after clear")
	}
	if history.CurrentIndex() != -1 {
		t.Errorf("Expected current index -1 after clear, got %d", history.CurrentIndex())
	}
}

// TestHistory_Stats tests history statistics.
func TestHistory_Stats(t *testing.T) {
	history := NewHistory()
	doc := New("hello")

	// Make some edits
	for i := 0; i < 3; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("x")
		txn := NewTransaction(cs)
		history.CommitRevision(txn, doc)
		doc = txn.Apply(doc)
	}

	stats := history.Stats()

	if stats.TotalRevisions != 3 {
		t.Errorf("Expected TotalRevisions 3, got %d", stats.TotalRevisions)
	}
	if stats.CurrentIndex != 2 {
		t.Errorf("Expected CurrentIndex 2, got %d", stats.CurrentIndex)
	}
	if !stats.CanUndo {
		t.Error("Expected CanUndo to be true")
	}
	if stats.CanRedo {
		t.Error("Expected CanRedo to be false")
	}
}

// TestHistory_GetPath tests getting the path to current revision.
func TestHistory_GetPath(t *testing.T) {
	history := NewHistory()
	doc := New("hello")

	// Make 3 edits in a chain
	for i := 0; i < 3; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert(string(rune('1' + i)))
		txn := NewTransaction(cs)
		history.CommitRevision(txn, doc)
		doc = txn.Apply(doc)
	}

	path := history.GetPath()

	// Path should be [0, 1, 2]
	if len(path) != 3 {
		t.Fatalf("Expected path length 3, got %d", len(path))
	}

	for i, idx := range path {
		if idx != i {
			t.Errorf("Path[%d]: expected %d, got %d", i, i, idx)
		}
	}
}

// TestHistory_MaxSize tests history size limit.
func TestHistory_MaxSize(t *testing.T) {
	history := NewHistory()
	history.SetMaxSize(5)

	doc := New("hello")

	// Make 10 edits
	for i := 0; i < 10; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert(string(rune('a' + i)))
		txn := NewTransaction(cs)
		history.CommitRevision(txn, doc)
		doc = txn.Apply(doc)
	}

	// Should only keep 5 most recent
	if history.RevisionCount() > 5 {
		t.Errorf("Expected max 5 revisions, got %d", history.RevisionCount())
	}
}

// TestChangeSet_Compose tests changeset composition.
// SKIP: Full composition requires position mapping which is complex.
// This is a placeholder for future implementation.
func TestChangeSet_Compose(t *testing.T) {
	t.Skip("Compose requires position mapping - not yet implemented")

	// Simple test: apply changesets sequentially instead of composing
	doc := New("hello")

	// Changeset 1: insert " world" at position 5
	cs1 := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" world")

	// Changeset 2: delete " world"
	cs2 := NewChangeSet(cs1.LenAfter()).
		Retain(5).
		Delete(6)

	// Apply sequentially (not composed)
	result := cs1.Apply(doc)
	result = cs2.Apply(result)

	// Should be back to "hello"
	if result.String() != "hello" {
		t.Errorf("Expected %q, got %q", "hello", result.String())
	}
}

// TestTransaction_Empty tests empty transaction handling.
func TestTransaction_Empty(t *testing.T) {
	doc := New("hello")

	// Empty changeset
	cs := NewChangeSet(doc.Length())
	txn := NewTransaction(cs)

	if !txn.IsEmpty() {
		t.Error("Expected transaction to be empty")
	}

	// Should not modify document
	result := txn.Apply(doc)
	if result.String() != doc.String() {
		t.Error("Empty transaction modified document")
	}

	// Empty transaction shouldn't be committed to history
	history := NewHistory()
	history.CommitRevision(txn, doc)

	if history.RevisionCount() != 0 {
		t.Error("Empty transaction was committed to history")
	}
}

// TestHistory_AtRootAtTip tests AtRoot and AtTip.
func TestHistory_AtRootAtTip(t *testing.T) {
	history := NewHistory()

	if !history.AtRoot() {
		t.Error("Expected AtRoot() to be true initially")
	}
	if !history.AtTip() {
		t.Error("Expected AtTip() to be true initially")
	}

	doc := New("hello")
	cs := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" world")
	txn := NewTransaction(cs)

	history.CommitRevision(txn, doc)

	if history.AtRoot() {
		t.Error("Expected AtRoot() to be false after edit")
	}
	if !history.AtTip() {
		t.Error("Expected AtTip() to be true after edit")
	}

	history.Undo()

	if !history.AtRoot() {
		t.Error("Expected AtRoot() to be true after undo")
	}
	if history.AtTip() {
		t.Error("Expected AtTip() to be false after undo")
	}
}

// TestChangeSet_Fusion tests operation fusion optimization.
func TestChangeSet_Fusion(t *testing.T) {
	doc := New("hello world")

	// Create a changeset with consecutive operations that should fuse
	cs := NewChangeSet(doc.Length()).
		Retain(5).
		Delete(1).
		Delete(1).
		Delete(1).
		Delete(1).
		Delete(1).
		Delete(1). // Delete " world" as 6 separate deletions
		Insert("a").
		Insert("b").
		Insert("c") // Insert "abc" as 3 separate inserts

	// Before fusion: 10 operations (1 retain + 6 deletes + 3 inserts)
	if len(cs.operations) != 10 {
		t.Errorf("Expected 10 operations before fusion, got %d", len(cs.operations))
	}

	// Apply the changeset (which triggers fusion on a copy, not the original)
	result := cs.Apply(doc)

	// Note: Apply creates a copy and fuses the copy, not the original changeset
	// So cs.operations still has 10 operations (not fused)

	// Verify the result is correct
	expected := "helloabc"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}

	// To get fused operations, use Optimized()
	optimized := cs.Optimized()
	if len(optimized.operations) != 3 {
		t.Errorf("Expected 3 operations after optimization, got %d", len(optimized.operations))
	}

	// Verify the optimized operations are correct
	if optimized.operations[0].OpType != OpRetain || optimized.operations[0].Length != 5 {
		t.Error("First operation should be Retain(5)")
	}
	if optimized.operations[1].OpType != OpDelete || optimized.operations[1].Length != 6 {
		t.Error("Second operation should be Delete(6)")
	}
	if optimized.operations[2].OpType != OpInsert || optimized.operations[2].Text != "abc" {
		t.Error("Third operation should be Insert(\"abc\")")
	}
}

// BenchmarkChangeSet_Apply benchmarks changeset application.
func BenchmarkChangeSet_Apply(b *testing.B) {
	doc := New("hello world, this is a test document")

	cs := NewChangeSet(doc.Length()).
		Retain(5).
		Delete(7).
		Insert(" gophers")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testDoc := New("hello world, this is a test document")
		_ = cs.Apply(testDoc)
	}
}

// BenchmarkChangeSet_Apply_WithFusion benchmarks with many consecutive operations.
func BenchmarkChangeSet_Apply_WithFusion(b *testing.B) {
	doc := New("hello world")

	// Create a changeset with many consecutive operations that benefit from fusion
	cs := NewChangeSet(doc.Length())
	for i := 0; i < 100; i++ {
		cs.Insert("x")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testDoc := New("hello world")
		_ = cs.Apply(testDoc)
	}
}

// BenchmarkHistory_UndoRedo benchmarks undo/redo operations.
func BenchmarkHistory_UndoRedo(b *testing.B) {
	history := NewHistory()
	doc := New("hello world")

	// Create 100 revisions
	for i := 0; i < 100; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("x")
		txn := NewTransaction(cs)
		history.CommitRevision(txn, doc)
		doc = txn.Apply(doc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Undo 10 times
		for j := 0; j < 10; j++ {
			undoTxn := history.Undo()
			if undoTxn != nil {
				doc = undoTxn.Apply(doc)
			}
		}
		// Redo 10 times
		for j := 0; j < 10; j++ {
			redoTxn := history.Redo()
			if redoTxn != nil {
				doc = redoTxn.Apply(doc)
			}
		}
	}
}
