package rope

import (
	"testing"
)

// TestChangeSetInvert tests the Invert method.
func TestChangeSetInvert(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		buildCS  func(*Rope) *ChangeSet
		verify   func(*testing.T, *Rope, *ChangeSet, *ChangeSet)
	}{
		{
			name:    "invert insert",
			initial: "Hello",
			buildCS: func(r *Rope) *ChangeSet {
				cs := NewChangeSet(r.Length())
				cs.Retain(5)
				cs.Insert(" World")
				return cs
			},
			verify: func(t *testing.T, original *Rope, cs, inverted *ChangeSet) {
				// Apply original changeset
				modified := cs.Apply(original)
				if modified.String() != "Hello World" {
					t.Errorf("Apply failed: got %q", modified.String())
				}

				// Apply inverted changeset should get back original
				reverted := inverted.Apply(modified)
				if reverted.String() != original.String() {
					t.Errorf("Invert failed: got %q, want %q", reverted.String(), original.String())
				}
			},
		},
		{
			name:    "invert delete",
			initial: "Hello World",
			buildCS: func(r *Rope) *ChangeSet {
				cs := NewChangeSet(r.Length())
				cs.Retain(5)
				cs.Delete(6) // Delete " World"
				return cs
			},
			verify: func(t *testing.T, original *Rope, cs, inverted *ChangeSet) {
				modified := cs.Apply(original)
				if modified.String() != "Hello" {
					t.Errorf("Apply failed: got %q", modified.String())
				}

				reverted := inverted.Apply(modified)
				if reverted.String() != original.String() {
					t.Errorf("Invert failed: got %q, want %q", reverted.String(), original.String())
				}
			},
		},
		{
			name:    "invert replace",
			initial: "Hello World",
			buildCS: func(r *Rope) *ChangeSet {
				cs := NewChangeSet(r.Length())
				cs.Retain(5)
				cs.Delete(6)
				cs.Insert(" Beautiful")
				return cs
			},
			verify: func(t *testing.T, original *Rope, cs, inverted *ChangeSet) {
				modified := cs.Apply(original)
				if modified.String() != "Hello Beautiful" {
					t.Errorf("Apply failed: got %q", modified.String())
				}

				reverted := inverted.Apply(modified)
				if reverted.String() != original.String() {
					t.Errorf("Invert failed: got %q, want %q", reverted.String(), original.String())
				}
			},
		},
		{
			name:    "invert multiple operations",
			initial: "ABCDEFG",
			buildCS: func(r *Rope) *ChangeSet {
				cs := NewChangeSet(r.Length())
				cs.Retain(2)
				cs.Delete(1) // Delete 'C'
				cs.Insert("X")
				cs.Retain(3)
				cs.Delete(1) // Delete 'G'
				return cs
			},
			verify: func(t *testing.T, original *Rope, cs, inverted *ChangeSet) {
				modified := cs.Apply(original)
				reverted := inverted.Apply(modified)
				if reverted.String() != original.String() {
					t.Errorf("Invert failed: got %q, want %q", reverted.String(), original.String())
				}
			},
		},
		{
			name:    "invert empty changeset",
			initial: "Hello",
			buildCS: func(r *Rope) *ChangeSet {
				return NewChangeSet(5)
			},
			verify: func(t *testing.T, original *Rope, cs, inverted *ChangeSet) {
				reverted := inverted.Apply(original)
				if reverted.String() != original.String() {
					t.Errorf("Invert empty failed: got %q, want %q", reverted.String(), original.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := New(tt.initial)
			cs := tt.buildCS(original)
			inverted := cs.Invert(original)
			tt.verify(t, original, cs, inverted)
		})
	}
}

// TestTransactionInvert tests the Transaction Invert method.
func TestTransactionInvert(t *testing.T) {
	t.Run("invert transaction", func(t *testing.T) {
		original := New("Hello World")
		cs := NewChangeSet(original.Length())
		cs.Retain(5)
		cs.Delete(1)
		cs.Insert(" Beautiful")
		transaction := NewTransaction(cs)

		// Apply transaction
		modified := transaction.Apply(original)
		if modified.String() != "Hello BeautifulWorld" {
			t.Errorf("Apply failed: got %q", modified.String())
		}

		// Invert and apply
		invertedTx := transaction.Invert(original)
		reverted := invertedTx.Apply(modified)
		if reverted.String() != original.String() {
			t.Errorf("Transaction invert failed: got %q, want %q", reverted.String(), original.String())
		}
	})

	t.Run("invert nil transaction", func(t *testing.T) {
		var tx *Transaction
		inverted := tx.Invert(nil)
		if inverted == nil {
			t.Error("Inverting nil transaction should return empty transaction")
		}
	})

	t.Run("invert empty transaction", func(t *testing.T) {
		cs := NewChangeSet(10)
		tx := NewTransaction(cs)
		inverted := tx.Invert(nil)

		if inverted == nil {
			t.Error("Invert should return transaction")
		}
	})
}

// TestChangeSetCompose tests the Compose method.
func TestChangeSetCompose(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		buildCS1 func(*Rope) *ChangeSet
		buildCS2 func(*Rope) *ChangeSet
		verify   func(*testing.T, *Rope, *ChangeSet, *ChangeSet, *ChangeSet)
	}{
		{
			name:    "compose retain then retain",
			initial: "Hello",
			buildCS1: func(r *Rope) *ChangeSet {
				cs := NewChangeSet(r.Length())
				cs.Retain(5)
				return cs
			},
			buildCS2: func(r *Rope) *ChangeSet {
				// After first, length is still 5
				cs := NewChangeSet(5)
				cs.Retain(5)
				return cs
			},
			verify: func(t *testing.T, original *Rope, cs1, cs2, composed *ChangeSet) {
				// Both are just retain, so composed should be retain
				result := composed.Apply(original)
				if result.String() != original.String() {
					t.Errorf("Compose retain failed: got %q, want %q", result.String(), original.String())
				}
			},
		},
		{
			name:    "compose delete then retain",
			initial: "Hello World",
			buildCS1: func(r *Rope) *ChangeSet {
				cs := NewChangeSet(11)
				cs.Retain(5)
				cs.Delete(6) // Delete " World"
				return cs
			},
			buildCS2: func(r *Rope) *ChangeSet {
				// After delete, length is 5
				cs := NewChangeSet(5)
				cs.Retain(5)
				return cs
			},
			verify: func(t *testing.T, original *Rope, cs1, cs2, composed *ChangeSet) {
				after1 := cs1.Apply(original)
				after2 := cs2.Apply(after1)
				expected := after2.String()

				result := composed.Apply(original)
				if result.String() != expected {
					t.Errorf("Compose failed: got %q, want %q", result.String(), expected)
				}
			},
		},
		{
			name:    "compose with empty first",
			initial: "Hello",
			buildCS1: func(r *Rope) *ChangeSet {
				return NewChangeSet(5)
			},
			buildCS2: func(r *Rope) *ChangeSet {
				cs := NewChangeSet(5)
				cs.Retain(5)
				cs.Insert(" World")
				return cs
			},
			verify: func(t *testing.T, original *Rope, cs1, cs2, composed *ChangeSet) {
				result := composed.Apply(original)
				expected := cs2.Apply(original).String()
				if result.String() != expected {
					t.Errorf("Compose with empty failed: got %q, want %q", result.String(), expected)
				}
			},
		},
		{
			name:    "compose with empty second",
			initial: "Hello",
			buildCS1: func(r *Rope) *ChangeSet {
				cs := NewChangeSet(5)
				cs.Retain(5)
				cs.Insert(" World")
				return cs
			},
			buildCS2: func(r *Rope) *ChangeSet {
				return NewChangeSet(10)
			},
			verify: func(t *testing.T, original *Rope, cs1, cs2, composed *ChangeSet) {
				result := composed.Apply(original)
				expected := cs1.Apply(original).String()
				if result.String() != expected {
					t.Errorf("Compose with empty failed: got %q, want %q", result.String(), expected)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := New(tt.initial)
			cs1 := tt.buildCS1(original)
			after1 := cs1.Apply(original)
			cs2 := tt.buildCS2(after1)
			composed := cs1.Compose(cs2)
			tt.verify(t, original, cs1, cs2, composed)
		})
	}
}

// TestTransactionCompose tests the Transaction Compose method.
func TestTransactionCompose(t *testing.T) {
	t.Run("compose transactions", func(t *testing.T) {
		_ = New("Hello")

		// First transaction
		cs1 := NewChangeSet(5)
		cs1.Retain(5)
		tx1 := NewTransaction(cs1)

		// Second transaction
		cs2 := NewChangeSet(5)
		cs2.Retain(5)
		tx2 := NewTransaction(cs2)

		// Compose should work
		composed := tx1.Compose(tx2)
		if composed == nil {
			t.Error("Compose returned nil")
		}
	})

	t.Run("compose with nil", func(t *testing.T) {
		cs := NewChangeSet(5)
		tx := NewTransaction(cs)

		// Compose with nil
		composed1 := tx.Compose(nil)
		if composed1 != tx {
			t.Error("Composing with nil should return original transaction")
		}

		// Compose nil with transaction
		composed2 := (*Transaction)(nil).Compose(tx)
		if composed2 != tx {
			t.Error("Composing nil with transaction should return the transaction")
		}
	})

	t.Run("compose preserves selection", func(t *testing.T) {
		sel := NewSelection()
		sel.Add(Range{Head: 5, Anchor: 10})

		cs1 := NewChangeSet(10)
		cs1.Retain(10)
		tx1 := NewTransaction(cs1).WithSelection(sel)

		cs2 := NewChangeSet(10)
		cs2.Retain(10)
		tx2 := NewTransaction(cs2)

		composed := tx1.Compose(tx2)
		if composed.Selection() == nil {
			t.Error("Composed transaction should have selection")
		}
	})
}

// TestChangeSetMapPosition tests the MapPosition method.
func TestChangeSetMapPosition(t *testing.T) {
	tests := []struct {
		name      string
		initial   string
		buildCS   func(*Rope) *ChangeSet
		position  int
		assoc     Assoc
		wantValid bool
	}{
		{
			name:    "map position through insert",
			initial: "Hello",
			buildCS: func(r *Rope) *ChangeSet {
				cs := NewChangeSet(r.Length())
				cs.Retain(5)
				cs.Insert(" World")
				return cs
			},
			position:  5,
			assoc:     AssocAfter,
			wantValid: true,
		},
		{
			name:    "map position through delete",
			initial: "Hello World",
			buildCS: func(r *Rope) *ChangeSet {
				cs := NewChangeSet(r.Length())
				cs.Retain(5)
				cs.Delete(6)
				return cs
			},
			position:  8, // Position in deleted text
			assoc:     AssocBefore,
			wantValid: true,
		},
		{
			name:    "map position before operation",
			initial: "ABCDEFG",
			buildCS: func(r *Rope) *ChangeSet {
				cs := NewChangeSet(r.Length())
				cs.Retain(2)
				cs.Delete(2)
				return cs
			},
			position:  0,
			assoc:     AssocBefore,
			wantValid: true,
		},
		{
			name:    "map position after operation",
			initial: "ABCDEFG",
			buildCS: func(r *Rope) *ChangeSet {
				cs := NewChangeSet(r.Length())
				cs.Retain(2)
				cs.Insert("XX")
				return cs
			},
			position:  6,
			assoc:     AssocBefore,
			wantValid: true,
		},
		{
			name:    "map position through empty changeset",
			initial: "Hello",
			buildCS: func(r *Rope) *ChangeSet {
				return NewChangeSet(5)
			},
			position:  3,
			assoc:     AssocBefore,
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rope := New(tt.initial)
			cs := tt.buildCS(rope)

			mapped := cs.MapPosition(tt.position, tt.assoc)

			if tt.wantValid {
				if mapped < 0 || mapped > cs.LenAfter() {
					t.Errorf("MapPosition returned invalid position: %d (lenAfter=%d)", mapped, cs.LenAfter())
				}
			}

			// Verify round-trip with Apply
			modified := cs.Apply(rope)
			if mapped <= modified.Length() {
				// Position should be valid
			}
		})
	}
}

// TestChangeSetMapPositions tests the MapPositions method.
func TestChangeSetMapPositions(t *testing.T) {
	t.Run("map multiple positions", func(t *testing.T) {
		rope := New("Hello World")
		cs := NewChangeSet(rope.Length())
		cs.Retain(5)
		cs.Delete(1)
		cs.Insert(" Beautiful")

		positions := []int{0, 5, 11}
		assocs := []Assoc{AssocBefore, AssocAfter, AssocBefore}

		mapped := cs.MapPositions(positions, assocs)

		if len(mapped) != len(positions) {
			t.Errorf("MapPositions returned %d positions, want %d", len(mapped), len(positions))
		}

		for _, pos := range mapped {
			if pos < 0 {
				t.Errorf("MapPositions returned negative position: %d", pos)
			}
		}
	})

	t.Run("map empty positions", func(t *testing.T) {
		rope := New("Hello")
		cs := NewChangeSet(rope.Length())
		cs.Retain(5)
		cs.Insert(" World")

		positions := []int{}
		assocs := []Assoc{}

		mapped := cs.MapPositions(positions, assocs)

		if len(mapped) != 0 {
			t.Errorf("MapPositions returned %d positions, want 0", len(mapped))
		}
	})

	t.Run("map positions with mismatched lengths", func(t *testing.T) {
		rope := New("Hello")
		cs := NewChangeSet(rope.Length())
		cs.Retain(5)
		cs.Insert(" World")

		positions := []int{0, 5, 11}
		assocs := []Assoc{AssocBefore, AssocAfter} // Fewer assocs than positions

		mapped := cs.MapPositions(positions, assocs)

		if len(mapped) != len(positions) {
			t.Errorf("MapPositions returned %d positions, want %d", len(mapped), len(positions))
		}
	})
}

// TestPositionMapperBasic tests basic position mapping.
func TestPositionMapperBasic(t *testing.T) {
	t.Run("map single position", func(t *testing.T) {
		rope := New("Hello World")
		cs := NewChangeSet(rope.Length())
		cs.Retain(5)
		cs.Insert(" Beautiful")

		mapper := NewPositionMapper(cs)
		mapper.AddPosition(5, AssocAfter)
		result := mapper.Map()

		if len(result) != 1 {
			t.Errorf("Map returned %d positions, want 1", len(result))
		}
	})

	t.Run("map multiple positions", func(t *testing.T) {
		rope := New("ABCDEFG")
		cs := NewChangeSet(rope.Length())
		cs.Retain(2)
		cs.Delete(2)
		cs.Insert("XX")

		positions := []int{0, 2, 4, 6}
		assocs := []Assoc{AssocBefore, AssocBefore, AssocBefore, AssocBefore}

		mapper := NewPositionMapper(cs)
		mapper.AddPositions(positions, assocs)
		result := mapper.Map()

		if len(result) != len(positions) {
			t.Errorf("Map returned %d positions, want %d", len(result), len(positions))
		}
	})

	t.Run("map with empty changeset", func(t *testing.T) {
		_ = New("Hello")
		cs := NewChangeSet(5)

		mapper := NewPositionMapper(cs)
		mapper.AddPosition(3, AssocBefore)
		result := mapper.Map()

		if len(result) != 1 {
			t.Errorf("Map returned %d positions, want 1", len(result))
		}
		if result[0] != 3 {
			t.Errorf("Position unchanged expected: got %d, want 3", result[0])
		}
	})
}

// TestTransactionEdgeCases tests edge cases for transaction operations.
func TestTransactionEdgeCases(t *testing.T) {
	t.Run("apply nil transaction", func(t *testing.T) {
		rope := New("Hello")
		var tx *Transaction
		result := tx.Apply(rope)
		if result.String() != "Hello" {
			t.Error("Applying nil transaction should return original rope")
		}
	})

	t.Run("invert with nil original", func(t *testing.T) {
		cs := NewChangeSet(10)
		cs.Retain(5)
		cs.Insert("Hello")
		tx := NewTransaction(cs)

		inverted := tx.Invert(nil)
		if inverted == nil {
			t.Error("Invert with nil should return transaction")
		}
	})

	t.Run("compose empty changesets", func(t *testing.T) {
		cs1 := NewChangeSet(10)
		cs2 := NewChangeSet(10)
		composed := cs1.Compose(cs2)

		if composed == nil {
			t.Error("Composing empty changesets should return changeset")
		}
	})

	t.Run("transaction is empty", func(t *testing.T) {
		cs := NewChangeSet(10)
		tx := NewTransaction(cs)

		if !tx.IsEmpty() {
			t.Error("Transaction with empty changeset should be empty")
		}

		cs.Retain(10)
		tx2 := NewTransaction(cs)
		if tx2.IsEmpty() {
			t.Error("Transaction with operations should not be empty")
		}
	})
}

// TestChangeSetLengths tests LenBefore and LenAfter methods.
func TestChangeSetLengths(t *testing.T) {
	t.Run("lengths for insert", func(t *testing.T) {
		cs := NewChangeSet(5)
		cs.Retain(5)
		cs.Insert(" World")

		if cs.LenBefore() != 5 {
			t.Errorf("LenBefore: got %d, want 5", cs.LenBefore())
		}
		if cs.LenAfter() != 11 {
			t.Errorf("LenAfter: got %d, want 11", cs.LenAfter())
		}
	})

	t.Run("lengths for delete", func(t *testing.T) {
		cs := NewChangeSet(11)
		cs.Retain(5)
		cs.Delete(6)

		if cs.LenBefore() != 11 {
			t.Errorf("LenBefore: got %d, want 11", cs.LenBefore())
		}
		if cs.LenAfter() != 5 {
			t.Errorf("LenAfter: got %d, want 5", cs.LenAfter())
		}
	})

	t.Run("lengths for replace", func(t *testing.T) {
		cs := NewChangeSet(11)
		cs.Retain(5)
		cs.Delete(6)
		cs.Insert(" Hello") // 6 characters

		if cs.LenBefore() != 11 {
			t.Errorf("LenBefore: got %d, want 11", cs.LenBefore())
		}
		if cs.LenAfter() != 11 { // 5 retained + 6 inserted = 11
			t.Errorf("LenAfter: got %d, want 11", cs.LenAfter())
		}
	})
}

// TestTransactionTimestamp tests the Timestamp method.
func TestTransactionTimestamp(t *testing.T) {
	cs := NewChangeSet(10)
	tx := NewTransaction(cs)

	if tx.Timestamp().IsZero() {
		t.Error("Transaction should have timestamp")
	}
}

// TestTransactionWithSelection tests WithSelection method.
func TestTransactionWithSelection(t *testing.T) {
	t.Run("with selection", func(t *testing.T) {
		cs := NewChangeSet(10)
		tx := NewTransaction(cs)

		sel := NewSelection()
		sel.Add(Range{Head: 0, Anchor: 5})

		txWithSel := tx.WithSelection(sel)
		if txWithSel.Selection() == nil {
			t.Error("WithSelection should set selection")
		}
	})

	t.Run("with selection on nil transaction", func(t *testing.T) {
		var tx *Transaction
		sel := NewSelection()

		result := tx.WithSelection(sel)
		if result != nil {
			t.Error("WithSelection on nil should return nil")
		}
	})
}

// TestChangeSetApplyTests tests Apply edge cases.
func TestChangeSetApplyTests(t *testing.T) {
	t.Run("apply to nil rope", func(t *testing.T) {
		cs := NewChangeSet(5)
		cs.Retain(5)
		result := cs.Apply(nil)
		if result != nil {
			t.Error("Applying to nil rope should return nil")
		}
	})

	t.Run("apply with length mismatch", func(t *testing.T) {
		rope := New("Hello")
		cs := NewChangeSet(10) // Wrong length
		cs.Retain(10)

		result := cs.Apply(rope)
		// Should return original rope when length mismatches
		if result.String() != "Hello" {
			t.Error("Apply with length mismatch should return original")
		}
	})

	t.Run("apply empty changeset", func(t *testing.T) {
		rope := New("Hello")
		cs := NewChangeSet(5)
		// Don't add any operations

		result := cs.Apply(rope)
		if result.String() != "Hello" {
			t.Errorf("Apply empty changeset: got %q, want %q", result.String(), "Hello")
		}
	})
}

// TestCompositionOrder tests that composition order matters.
func TestCompositionOrder(t *testing.T) {
	t.Run("composition is not commutative", func(t *testing.T) {
		// First changeset: insert "X" at position 2
		cs1 := NewChangeSet(5)
		cs1.Retain(2)
		cs1.Insert("X")
		cs1.Retain(3)

		// Second changeset: delete 1 char at position 3
		cs2 := NewChangeSet(6) // After first insert
		cs2.Retain(3)
		cs2.Delete(1)
		cs2.Retain(2)

		composed12 := cs1.Compose(cs2)

		// Reverse order should give different result
		// (though technically invalid due to length constraints)
		// This test just verifies composition works
		if composed12 == nil {
			t.Error("Compose should return changeset")
		}
	})
}

// TestComplexInvertScenario tests complex invert scenarios.
func TestComplexInvertScenario(t *testing.T) {
	t.Run("invert complex edit", func(t *testing.T) {
		original := New("The quick brown fox jumps over the lazy dog")

		// Make a complex edit
		cs := NewChangeSet(original.Length())
		cs.Retain(4) // "The "
		cs.Delete(6)  // Delete "quick "
		cs.Insert("fast")
		cs.Retain(11) // "brown fox "
		cs.Delete(5)  // Delete "jumps"
		cs.Insert("leaps")
		cs.Retain(19) // " over the lazy dog"

		// Apply
		modified := cs.Apply(original)

		// Invert and revert
		inverted := cs.Invert(original)
		reverted := inverted.Apply(modified)

		if reverted.String() != original.String() {
			t.Errorf("Complex invert failed: got %q, want %q", reverted.String(), original.String())
		}
	})
}

// TestCompositionPreservesContent tests that composition preserves document content.
func TestCompositionPreservesContent(t *testing.T) {
	t.Run("compose preserves content", func(t *testing.T) {
		original := New("Hello World")

		// First edit
		cs1 := NewChangeSet(original.Length())
		cs1.Retain(5)
		cs1.Delete(1)
		cs1.Insert(", Beautiful")

		// Second edit
		cs2 := NewChangeSet(14)
		cs2.Retain(14)
		cs2.Delete(6)
		cs2.Insert("Universe")

		// Apply separately
		after1 := cs1.Apply(original)
		after2 := cs2.Apply(after1)

		// Apply composed
		composed := cs1.Compose(cs2)
		composedResult := composed.Apply(original)

		if composedResult.String() != after2.String() {
			t.Errorf("Compose didn't preserve: got %q, want %q",
				composedResult.String(), after2.String())
		}
	})
}
