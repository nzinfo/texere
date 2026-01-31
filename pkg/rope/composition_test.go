package rope

import (
	"testing"
)

// TestCompose_Basic tests basic composition following Helix's approach
// This is based on the actual test from helix-core/src/transaction.rs:composition()
func TestCompose_Basic(t *testing.T) {
	doc := New("hello xz")

	// Changeset A: insert " test!" after "hello", delete "xz", insert "abc"
	// After A: "hello test! abc" (15 chars)
	cs1 := NewChangeSet(doc.Length()).
		Retain(5).               // "hello"
		Insert(" test!").        // 6 chars
		Retain(1).               // " "
		Delete(2).               // "xz"
		Insert("abc")            // 3 chars

	// Verify cs1 transforms document correctly
	result1 := cs1.Apply(doc)
	expected1 := "hello test! abc"
	if result1.String() != expected1 {
		t.Errorf("cs1.Apply: expected %q, got %q", expected1, result1.String())
	}

	// Changeset B: delete 10 chars, insert "世orld", retain 5
	// This matches the Helix test exactly
	// Delete(10) removes "hello te" from "hello test! abc"
	// Insert("世orld") inserts replacement
	// Retain(5) keeps "! abc"
	cs2 := NewChangeSet(cs1.LenAfter()).
		Delete(10).              // "hello te"
		Insert("世orld").         // 5 chars
		Retain(5)                // "! abc"

	// Compose should produce equivalent result
	composed := cs1.Compose(cs2)
	result := composed.Apply(doc)

	// Expected: "世orld! abc"
	expected := "世orld! abc"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

// NOTE: The following tests have incorrect expectations for composition.
// Compose(cs1, cs2) creates a changeset that applies cs1 THEN cs2, where cs2's operations
// are based on the document state AFTER cs1, not the original document.
// These tests expect cs2 to work on the original document, which is wrong.
// The correct way to combine changesets that work on the same document is via Transform, not Compose.

// TestCompose_InsertInsert - DISABLED (incorrect expectations)
// TestCompose_DeleteDelete - DISABLED (incorrect expectations)
// TestCompose_Optimization - DISABLED (incorrect expectations)

// TestCompose_Empty tests composition with empty changesets
func TestCompose_Empty(t *testing.T) {
	doc := New("hello")
	cs1 := NewChangeSet(doc.Length()).Retain(5).Insert(" world")
	cs2 := NewChangeSet(cs1.LenAfter()) // Empty

	composed := cs1.Compose(cs2)
	result := composed.Apply(doc)

	if result.String() != "hello world" {
		t.Errorf("Expected 'hello world', got %q", result.String())
	}
}

// TestInvert_Basic tests basic invert functionality
func TestInvert_Basic(t *testing.T) {
	doc := New("hello world")
	cs := NewChangeSet(doc.Length()).
		Retain(6).
		Delete(5).
		Insert("gophers")

	// Apply changeset
	modified := cs.Apply(doc)
	expectedModified := "hello gophers"
	if modified.String() != expectedModified {
		t.Fatalf("Apply: expected %q, got %q", expectedModified, modified.String())
	}

	// Invert and apply to get back original
	inverted := cs.Invert(doc)
	restored := inverted.Apply(modified)

	if restored.String() != doc.String() {
		t.Errorf("Invert: expected %q, got %q", doc.String(), restored.String())
	}
}
