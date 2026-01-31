package rope

// SimpleCompose composes two changesets by applying them sequentially
// and creating a new changeset that represents the combined effect.
// This is simpler and more reliable than complex composition algorithms.
func SimpleCompose(first, second *ChangeSet, original *Rope) *ChangeSet {
	if first == nil || first.IsEmpty() {
		if second == nil || second.IsEmpty() {
			return NewChangeSet(original.Length())
		}
		result := NewChangeSet(second.lenBefore)
		result.operations = make([]Operation, len(second.operations))
		copy(result.operations, second.operations)
		result.lenAfter = second.lenAfter
		return result
	}

	if second == nil || second.IsEmpty() {
		result := NewChangeSet(first.lenBefore)
		result.operations = make([]Operation, len(first.operations))
		copy(result.operations, first.operations)
		result.lenAfter = first.lenAfter
		return result
	}

	// Apply first changeset
	middleDoc := first.Apply(original)

	// Calculate what the second changeset should do on the middle document
	// This requires us to understand what the second changeset is trying to do
	// and adjust it based on what the first changeset did

	// For now, let's implement a specific case: first inserts, second deletes
	// We need to track where characters ended up after first operation

	// Build a map of where characters moved after first operation
	charMap := make([]int, first.lenAfter+1) // Maps old position to new position
	oldPos := 0
	newPos := 0

	for _, op := range first.operations {
		switch op.OpType {
		case OpRetain:
			// Characters stay in same relative position
			for i := 0; i < op.Length; i++ {
				charMap[oldPos+i+1] = newPos + i + 1
			}
			oldPos += op.Length
			newPos += op.Length

		case OpDelete:
			// Characters are removed
			for i := 0; i < op.Length; i++ {
				charMap[oldPos+i+1] = newPos // Deleted chars map to current pos
			}
			oldPos += op.Length

		case OpInsert:
			// New characters are inserted
			insertLen := len([]rune(op.Text))
			// Inserted chars don't exist in original, so we don't map them
			// But they affect subsequent positions
			newPos += insertLen
		}
	}

	// Now apply second changeset to middleDoc and record what happened
	// This is the ground truth
	finalDoc := second.Apply(middleDoc)

	// Now create a changeset that goes from original to final directly
	// We do this by comparing original and final
	result := NewChangeSet(original.Length())

	origIter := original.NewIterator()
	finalIter := finalDoc.NewIterator()
	pos := 0

	// Find common prefix
	for origIter.Next() && finalIter.Next() {
		if origIter.Current() == finalIter.Current() {
			pos++
		} else {
			break
		}
	}

	if pos > 0 {
		result.Retain(pos)
	}

	// Now find differences
	// Reset iterators
	origIter = original.IteratorAt(pos)
	finalIter = finalDoc.IteratorAt(pos)

	// Track insertions and deletions
	for origIter.HasNext() || finalIter.HasNext() {
		origCh := rune(0)
		finalCh := rune(0)

		hasOrig := origIter.Next()
		hasFinal := finalIter.Next()

		if hasOrig {
			origCh = origIter.Current()
		}

		if hasFinal {
			finalCh = finalIter.Current()
		}

		if hasOrig && hasFinal {
			if origCh == finalCh {
				// Same character, retain
				result.Retain(1)
			} else {
				// Different character - delete old, insert new
				// Delete all remaining orig chars
				result.Delete(1)
				for origIter.Next() {
					// Skip all
				}
				// Insert all remaining final chars
				result.Insert(string(finalCh))
				for finalIter.Next() {
					result.Insert(string(finalIter.Current()))
				}
				break
			}
		} else if hasOrig && !hasFinal {
			// orig has more chars - delete them
			result.Delete(1)
		} else if !hasOrig && hasFinal {
			// final has more chars - insert them
			inserted := string(finalCh)
			for finalIter.Next() {
				inserted += string(finalIter.Current())
			}
			result.Insert(inserted)
			break
		}
	}

	result.lenAfter = finalDoc.Length()
	return result
}

