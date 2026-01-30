package rope

import "testing"

// TestRange_NewRange tests creating new ranges
func TestRange_NewRange(t *testing.T) {
	// Test cursor (zero-width range)
	cursor := NewRange(5, 5)
	if cursor.From() != 5 || cursor.To() != 5 {
		t.Errorf("Expected cursor at 5, got %d-%d", cursor.From(), cursor.To())
	}
	if !cursor.IsCursor() {
		t.Error("Expected IsCursor() to return true")
	}

	// Test forward selection
	forward := NewRange(0, 5)
	if forward.From() != 0 || forward.To() != 5 {
		t.Errorf("Expected 0-5, got %d-%d", forward.From(), forward.To())
	}
	if !forward.IsForward() {
		t.Error("Expected IsForward() to return true")
	}

	// Test backward selection
	backward := NewRange(5, 0)
	if backward.From() != 0 || backward.To() != 5 {
		t.Errorf("Expected 0-5, got %d-%d", backward.From(), backward.To())
	}
	if !backward.IsBackward() {
		t.Error("Expected IsBackward() to return true")
	}
}

// TestRange_Point tests creating a point (cursor)
func TestRange_Point(t *testing.T) {
	cursor := Point(10)
	if cursor.Anchor != 10 || cursor.Head != 10 {
		t.Errorf("Expected cursor at 10, got anchor=%d, head=%d", cursor.Anchor, cursor.Head)
	}
	if !cursor.IsCursor() {
		t.Error("Point should create a cursor")
	}
}

// TestRange_Contains tests the Contains method
func TestRange_Contains(t *testing.T) {
	r := NewRange(5, 10)

	// Test positions
	testCases := []struct {
		pos    int
		expect bool
	}{
		{0, false},
		{5, true},
		{7, true},
		{9, true},
		{10, false},
		{15, false},
	}

	for _, tc := range testCases {
		result := r.Contains(tc.pos)
		if result != tc.expect {
			t.Errorf("Contains(%d): expected %v, got %v", tc.pos, tc.expect, result)
		}
	}
}

// TestSelection_NewSelection tests creating a new selection
func TestSelection_NewSelection(t *testing.T) {
	// Single cursor
	sel := NewSelection(Point(5))
	if sel.Len() != 1 {
		t.Errorf("Expected length 1, got %d", sel.Len())
	}
	if sel.PrimaryIndex() != 0 {
		t.Errorf("Expected primary index 0, got %d", sel.PrimaryIndex())
	}

	// Multiple ranges
	ranges := []Range{
		NewRange(0, 5),
		NewRange(10, 15),
		Point(20),
	}
	sel = NewSelection(ranges...)
	if sel.Len() != 3 {
		t.Errorf("Expected length 3, got %d", sel.Len())
	}

	// Test primary
	primary := sel.Primary()
	if primary.From() != 0 || primary.To() != 5 {
		t.Errorf("Expected primary range 0-5, got %d-%d", primary.From(), primary.To())
	}
}

// TestSelection_NewSelectionWithPrimary tests creating selection with specific primary
func TestSelection_NewSelectionWithPrimary(t *testing.T) {
	ranges := []Range{
		NewRange(0, 5),
		NewRange(10, 15),
		Point(20),
	}

	// Primary at index 1
	sel := NewSelectionWithPrimary(ranges, 1)
	if sel.PrimaryIndex() != 1 {
		t.Errorf("Expected primary index 1, got %d", sel.PrimaryIndex())
	}
	primary := sel.Primary()
	if primary.From() != 10 || primary.To() != 15 {
		t.Errorf("Expected primary range 10-15, got %d-%d", primary.From(), primary.To())
	}

	// Invalid primary index should default to 0
	sel = NewSelectionWithPrimary(ranges, 5)
	if sel.PrimaryIndex() != 0 {
		t.Errorf("Expected primary index 0 (out of bounds), got %d", sel.PrimaryIndex())
	}
}

// TestTransaction_WithSelection tests adding selection to transaction
func TestTransaction_WithSelection(t *testing.T) {
	doc := New("hello world")
	cs := NewChangeSet(doc.Length()).Retain(5).Insert(" beautiful")
	tx := NewTransaction(cs)

	// Add selection
	sel := NewSelection(NewRange(5, 10))
	txWithSel := tx.WithSelection(sel)

	if txWithSel.Selection() == nil {
		t.Error("Expected selection to be set")
	}
	if txWithSel.Selection().Len() != 1 {
		t.Errorf("Expected selection length 1, got %d", txWithSel.Selection().Len())
	}
}

// TestTransaction_Change tests creating transaction from changes
func TestTransaction_Change(t *testing.T) {
	doc := New("hello world")

	// Replace "world" with "gophers"
	changes := []EditOperation{
		{From: 6, To: 11, Text: "gophers"},
	}

	tx := Change(doc, changes)
	result := tx.Apply(doc)

	expected := "hello gophers"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

// TestTransaction_DeleteFromDeletions tests creating transaction from deletions
func TestTransaction_DeleteFromDeletions(t *testing.T) {
	doc := New("hello world")

	// Delete "world "
	deletions := []Deletion{
		{From: 5, To: 11},
	}

	tx := Delete(doc, deletions)
	result := tx.Apply(doc)

	expected := "hello"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

// TestTransaction_InsertAtEOF tests inserting at end of document
func TestTransaction_InsertAtEOF(t *testing.T) {
	doc := New("hello")

	tx := NewTransaction(NewChangeSet(doc.Length()))
	tx = tx.InsertAtEOF(" world")

	result := tx.Apply(doc)

	expected := "hello world"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

// TestTransaction_Insert tests inserting at all cursor positions
func TestTransaction_Insert(t *testing.T) {
	doc := New("abc")

	// Create selection with two cursors
	sel := NewSelection(Point(1), Point(2))

	tx := Insert(doc, sel, "X")
	result := tx.Apply(doc)

	expected := "aXbXc"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

// TestTransaction_Compose tests composing transactions with selections
func TestTransaction_Compose(t *testing.T) {
	doc := New("hello")

	cs1 := NewChangeSet(doc.Length()).Retain(5).Insert(" world")
	tx1 := NewTransaction(cs1)

	sel := NewSelection(Point(10))
	tx1 = tx1.WithSelection(sel)

	// Compose with another transaction that deletes the inserted " world"
	// Need to Retain(5) then Delete(6) to delete " world" (6 chars starting at position 5)
	cs2 := NewChangeSet(tx1.Changeset().LenAfter()).Retain(5).Delete(6)
	tx2 := NewTransaction(cs2)

	composed := tx1.Compose(tx2)
	result := composed.Apply(doc)

	expected := "hello"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

// TestChangeIterator_Basic tests the ChangeIterator
func TestChangeIterator_Basic(t *testing.T) {
	doc := New("hello world")
	cs := NewChangeSet(doc.Length()).
		Retain(5).
		Insert(" beautiful").
		Retain(1).
		Delete(5)

	iter := NewChangeIterator(cs)

	// First operation: Retain(5) at position 0
	info := iter.Next()
	if info == nil {
		t.Fatal("Expected first operation")
	}
	if info.Operation.OpType != OpRetain {
		t.Errorf("Expected Retain, got %v", info.Operation.OpType)
	}
	if info.Position != 0 {
		t.Errorf("Expected position 0, got %d", info.Position)
	}

	// Second operation: Insert at position 5
	info = iter.Next()
	if info == nil {
		t.Fatal("Expected second operation")
	}
	if info.Operation.OpType != OpInsert {
		t.Errorf("Expected Insert, got %v", info.Operation.OpType)
	}
	if info.Position != 5 {
		t.Errorf("Expected position 5, got %d", info.Position)
	}

	// Check position after insert
	if iter.Position() != 15 { // 5 + len(" beautiful") = 15
		t.Errorf("Expected position 15, got %d", iter.Position())
	}

	// Third operation: Retain(1) at position 15
	info = iter.Next()
	if info.Position != 15 {
		t.Errorf("Expected position 15, got %d", info.Position)
	}

	// Fourth operation: Delete(5) at position 16
	info = iter.Next()
	if info == nil {
		t.Fatal("Expected fourth operation")
	}
	if info.Operation.OpType != OpDelete {
		t.Errorf("Expected Delete, got %v", info.Operation.OpType)
	}
	if info.Position != 16 {
		t.Errorf("Expected position 16, got %d", info.Position)
	}

	// No more operations
	info = iter.Next()
	if info != nil {
		t.Error("Expected nil, got operation")
	}

	if iter.HasMore() {
		t.Error("Expected HasMore() to return false")
	}
}

// TestChangeIterator_Reset tests resetting the iterator
func TestChangeIterator_Reset(t *testing.T) {
	doc := New("hello")
	cs := NewChangeSet(doc.Length()).Retain(5).Insert(" world")

	iter := NewChangeIterator(cs)

	// Consume first operation
	iter.Next()
	if iter.index != 1 {
		t.Errorf("Expected index 1, got %d", iter.index)
	}

	// Reset
	iter.Reset()
	if iter.index != 0 {
		t.Errorf("Expected index 0 after reset, got %d", iter.index)
	}
	if iter.Position() != 0 {
		t.Errorf("Expected position 0 after reset, got %d", iter.Position())
	}

	// Should be able to iterate again
	info := iter.Next()
	if info == nil {
		t.Fatal("Expected operation after reset")
	}
	if info.Position != 0 {
		t.Errorf("Expected position 0, got %d", info.Position)
	}
}

// TestRange_Map tests mapping a range through a changeset
func TestRange_Map(t *testing.T) {
	doc := New("hello world")

	// Replace "world" with "gophers"
	cs := NewChangeSet(doc.Length()).
		Retain(6).
		Delete(5).
		Insert("gophers")

	// Create a range that covers "world"
	r := NewRange(6, 11)

	// Map the range through the changeset
	mapped := r.Map(cs, AssocBefore)

	// After the change, "world" is replaced with "gophers" (7 chars)
	// So the range should now cover "gophers"
	if mapped.From() != 6 {
		t.Errorf("Expected from 6, got %d", mapped.From())
	}
	// The length changed from 5 to 7, so to should be 6 + 7 = 13
	if mapped.To() != 13 {
		t.Errorf("Expected to 13, got %d", mapped.To())
	}
}
