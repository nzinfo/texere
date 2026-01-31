package concordia

import (
	"testing"

	"github.com/coreseekdev/texere/pkg/ot"
	"github.com/coreseekdev/texere/pkg/rope"
)

// ========== Cursor Association Tests ==========

func TestAssoc_String(t *testing.T) {
	tests := []struct {
		assoc    rope.Assoc
		expected string
	}{
		{rope.AssocBefore, "Before"},
		{rope.AssocAfter, "After"},
		{rope.AssocBeforeWord, "BeforeWord"},
		{rope.AssocAfterWord, "AfterWord"},
		{rope.AssocBeforeSticky, "BeforeSticky"},
		{rope.AssocAfterSticky, "AfterSticky"},
	}

	for _, tt := range tests {
		if tt.assoc.String() != tt.expected {
			t.Errorf("Expected %q, got %q", tt.expected, tt.assoc.String())
		}
	}
}

// ========== History Navigation Tests ==========

func TestHistory_EarlierMultipleSteps(t *testing.T) {
	history := NewHistory()
	doc := rope.New("hello")

	// Create 5 edits using ot.Operation
	for i := 0; i < 5; i++ {
		builder := ot.NewBuilder()
		builder.Retain(doc.Length())
		builder.Insert(string(rune('a' + i)))
		op := builder.Build()

		history.CommitRevision(op, doc)

		// Apply the operation to the document
		var err error
		doc, err = ApplyOperation(doc, op)
		if err != nil {
			t.Fatalf("Failed to apply operation: %v", err)
		}
	}

	expected := "helloabcde"
	if doc.String() != expected {
		t.Fatalf("Expected %q, got %q", expected, doc.String())
	}

	// Undo 3 steps using Lamport-based navigation
	for i := 0; i < 3; i++ {
		undoOp := history.Undo()
		if undoOp != nil {
			var err error
			doc, err = ApplyOperation(doc, undoOp)
			if err != nil {
				t.Fatalf("Failed to apply undo: %v", err)
			}
		}
	}

	// Should be at "helloab"
	if doc.String() != "helloab" {
		t.Errorf("After undoing 3 times: expected %q, got %q", "helloab", doc.String())
	}

	// Verify history state
	if history.CurrentIndex() != 1 {
		t.Errorf("Expected current index 1, got %d", history.CurrentIndex())
	}
}

func TestHistory_LaterMultipleSteps(t *testing.T) {
	history := NewHistory()
	doc := rope.New("hello")

	// Create 5 edits using ot.Operation
	for i := 0; i < 5; i++ {
		builder := ot.NewBuilder()
		builder.Retain(doc.Length())
		builder.Insert(string(rune('a' + i)))
		op := builder.Build()

		history.CommitRevision(op, doc)

		var err error
		doc, err = ApplyOperation(doc, op)
		if err != nil {
			t.Fatalf("Failed to apply operation: %v", err)
		}
	}

	// Undo 2 steps using sequential Undo calls
	for i := 0; i < 2; i++ {
		undoOp := history.Undo()
		if undoOp != nil {
			var err error
			doc, err = ApplyOperation(doc, undoOp)
			if err != nil {
				t.Fatalf("Failed to apply undo: %v", err)
			}
		}
	}

	// Redo 1 step using Later
	redoOp := history.Later(1)
	if redoOp == nil {
		t.Fatal("Expected Later(1) to return an operation")
	}

	var err error
	doc, err = ApplyOperation(doc, redoOp)
	if err != nil {
		t.Fatalf("Failed to apply redo: %v", err)
	}

	// Should have moved forward 1 step
	if history.CurrentIndex() != 3 {
		t.Errorf("Expected current index 3, got %d", history.CurrentIndex())
	}
}

func TestHistory_LamportTimestamps(t *testing.T) {
	history := NewHistory()
	doc := rope.New("hello")

	// Create edits - each should get an incremented Lamport time
	var lastLamport LamportTime
	for i := 0; i < 5; i++ {
		builder := ot.NewBuilder()
		builder.Retain(doc.Length())
		builder.Insert(string(rune('a' + i)))
		op := builder.Build()

		history.CommitRevision(op, doc)

		var err error
		doc, err = ApplyOperation(doc, op)
		if err != nil {
			t.Fatalf("Failed to apply operation: %v", err)
		}

		// Verify Lamport time is increasing
		currentLamport := history.LamportAt()
		if currentLamport <= lastLamport {
			t.Errorf("Expected Lamport time to increase, got %d after %d", currentLamport, lastLamport)
		}
		lastLamport = currentLamport
	}

	// Final Lamport time should be 5
	if lastLamport != 5 {
		t.Errorf("Expected final Lamport time 5, got %d", lastLamport)
	}
}

// ========== Operation Application Tests ==========

func TestOperation_ApplyInsert(t *testing.T) {
	doc := rope.New("hello")

	// Insert " world" at position 5
	builder := ot.NewBuilder()
	builder.Retain(5)
	builder.Insert(" world")
	op := builder.Build()

	result, err := ApplyOperation(doc, op)
	if err != nil {
		t.Fatalf("Failed to apply operation: %v", err)
	}

	expected := "hello world"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

func TestOperation_ApplyDelete(t *testing.T) {
	doc := rope.New("hello world")

	// Delete " world" (positions 5-11)
	builder := ot.NewBuilder()
	builder.Retain(5)
	builder.Delete(6)
	op := builder.Build()

	result, err := ApplyOperation(doc, op)
	if err != nil {
		t.Fatalf("Failed to apply operation: %v", err)
	}

	expected := "hello"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

func TestOperation_ApplyReplace(t *testing.T) {
	doc := rope.New("hello world")

	// Replace "world" with "gophers"
	builder := ot.NewBuilder()
	builder.Retain(6)        // "hello "
	builder.Delete(5)        // "world"
	builder.Insert("gophers") // replacement
	op := builder.Build()

	result, err := ApplyOperation(doc, op)
	if err != nil {
		t.Fatalf("Failed to apply operation: %v", err)
	}

	expected := "hello gophers"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

// ========== Operation Conversion Tests ==========

func TestOperationFromChanges_SingleInsert(t *testing.T) {
	doc := rope.New("hello")

	// Create edit operation
	changes := []EditOperation{
		{From: 5, To: 5, Text: " world"},
	}

	op := OperationFromChanges(doc, changes)
	if op == nil {
		t.Fatal("Expected operation to be created")
	}

	result, err := ApplyOperation(doc, op)
	if err != nil {
		t.Fatalf("Failed to apply operation: %v", err)
	}

	expected := "hello world"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

func TestOperationFromChanges_SingleDelete(t *testing.T) {
	doc := rope.New("hello world")

	// Create deletion
	changes := []EditOperation{
		{From: 5, To: 11}, // delete " world"
	}

	op := OperationFromChanges(doc, changes)
	if op == nil {
		t.Fatal("Expected operation to be created")
	}

	result, err := ApplyOperation(doc, op)
	if err != nil {
		t.Fatalf("Failed to apply operation: %v", err)
	}

	expected := "hello"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

func TestOperationFromChanges_MultipleEdits(t *testing.T) {
	doc := rope.New("hello world")

	// Multiple edits: delete "world", insert "gophers"
	changes := []EditOperation{
		{From: 6, To: 11, Text: "gophers"},
	}

	op := OperationFromChanges(doc, changes)
	if op == nil {
		t.Fatal("Expected operation to be created")
	}

	result, err := ApplyOperation(doc, op)
	if err != nil {
		t.Fatalf("Failed to apply operation: %v", err)
	}

	expected := "hello gophers"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}
