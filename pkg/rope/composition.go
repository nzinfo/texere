package rope

// import "fmt"

// Compose composes this changeset with another, producing a changeset that
// represents applying this changeset followed by the other.
// This implementation follows Helix editor's approach with OT-based composition.
func (cs *ChangeSet) Compose(other *ChangeSet) *ChangeSet {
	// Handle empty changesets
	if cs.IsEmpty() {
		if other == nil || other.IsEmpty() {
			return NewChangeSet(cs.lenBefore)
		}
		result := NewChangeSet(other.lenBefore)
		result.operations = make([]Operation, len(other.operations))
		copy(result.operations, other.operations)
		result.lenAfter = other.lenAfter
		return result
	}

	if other == nil || other.IsEmpty() {
		result := NewChangeSet(cs.lenBefore)
		result.operations = make([]Operation, len(cs.operations))
		copy(result.operations, cs.operations)
		result.lenAfter = cs.lenAfter
		return result
	}

	// Finalize both changesets to ensure they cover entire document length
	// This follows Helix's approach where changesets must account for all characters
	csFinal := cs.clone().finalize()
	otherFinal := other.clone().finalize()

	// KEY INVARIANT: csFinal.lenAfter must equal otherFinal.lenBefore
	// This is the fundamental requirement for composition
	if csFinal.lenAfter != otherFinal.lenBefore {
		// Cannot compose - length mismatch
		// Return a changeset that just applies cs (fallback)
		result := NewChangeSet(csFinal.lenBefore)
		result.operations = make([]Operation, len(csFinal.operations))
		copy(result.operations, csFinal.operations)
		result.lenAfter = csFinal.lenAfter
		return result
	}

	result := NewChangeSet(csFinal.lenBefore)
	i, j := 0, 0
	firstOps := csFinal.operations
	secondOps := otherFinal.operations

	// Debug: track composition steps
	step := 0

	// Use two-pointer iteration algorithm (like merge sort)
	for i < len(firstOps) || j < len(secondOps) {
		// Debug: track composition steps
		if i < len(firstOps) && j < len(secondOps) {
			step++
			// fmt.Printf("[Step %d] Compose: firstOps[%d]=%+v, secondOps[%d]=%+v\n",
			// 	step, i, firstOps[i], j, secondOps[j])
		}
		if i >= len(firstOps) {
			// Only second operations remaining
			result.addOperation(secondOps[j])
			j++
			continue
		}

		if j >= len(secondOps) {
			// Only first operations remaining
			result.addOperation(firstOps[i])
			i++
			continue
		}

		// Both operations available - need to compose them
		firstOp := firstOps[i]
		secondOp := secondOps[j]

		composed := composeOperations(firstOp, secondOp, &i, &j, firstOps, secondOps)
		if composed != nil {
			result.addOperation(*composed)
		}
	}

	result.recalculateLenAfter()
	result.fuse()
	return result
}

// clone creates a deep copy of this changeset.
func (cs *ChangeSet) clone() *ChangeSet {
	clone := NewChangeSet(cs.lenBefore)
	clone.operations = make([]Operation, len(cs.operations))
	copy(clone.operations, cs.operations)
	clone.lenAfter = cs.lenAfter
	return clone
}

// composeOperations composes two individual operations.
// Returns the composed operation, or nil if operations should be processed separately.
// Updates indices i and j as needed.
func composeOperations(firstOp, secondOp Operation, i, j *int, firstOps, secondOps []Operation) *Operation {
	switch firstOp.OpType {
	case OpDelete:
		// Delete in first operation wins - second operation can't affect deleted content
		*i++
		return &firstOp

	case OpInsert:
		// First operation inserts text
		insertText := firstOp.Text
		insertLen := len([]rune(insertText))

		switch secondOp.OpType {
		case OpDelete:
			// Delete in second operation
			deleteLen := secondOp.Length

			if insertLen < deleteLen {
				// Insertion is shorter than deletion - all inserted text is deleted, plus more
				*i++
				*j++
				return &secondOp
			} else if insertLen == deleteLen {
				// Exact match - both operations cancel out
				*i++
				*j++
				return nil // Skip both
			} else {
				// Insertion is longer than deletion - part of insertion is deleted
				remainingInsert := insertText[insertLen:] // After deleteLen characters
				// Keep remaining part of insertion
				*i++
				*j++
				return &Operation{OpType: OpInsert, Text: remainingInsert}
			}

		case OpRetain:
			// Second operation retains
			retainLen := secondOp.Length

			if insertLen < retainLen {
				// Insertion is shorter - all of it is retained, then retain more
				result := Operation{OpType: OpInsert, Text: insertText}
				*i++
				// Don't increment j - keep secondOp for next iteration
				// But reduce its length by insertLen
				secondOps[*j] = Operation{OpType: OpRetain, Length: retainLen - insertLen}
				return &result
			} else if insertLen == retainLen {
				// Exact match - insert then retain cancels out
				*i++
				*j++
				return nil // Skip both
			} else {
				// Insertion is longer - split the insertion
				// Part of insertion is retained, part is kept as insert
				splitPoint := retainLen
				beforePart := insertText[:splitPoint]
				afterPart := insertText[splitPoint:]

				result := Operation{OpType: OpInsert, Text: beforePart}
				*i++
				*j++

				// Keep after part as a new Insert operation
				// But first, we need to handle the current secondOp
				if len(afterPart) > 0 {
					// Add the after part after processing current secondOp
					// For now, just return the first part
					return &result
				}
				return &result
			}

		case OpInsert:
			// Second operation also inserts - just combine both inserts
			combinedText := insertText + secondOp.Text
			*i++
			*j++
			return &Operation{OpType: OpInsert, Text: combinedText}
		}

	case OpRetain:
		// First operation retains some characters
		retainLen := firstOp.Length

		switch secondOp.OpType {
		case OpDelete:
			// Delete in second operation
			deleteLen := secondOp.Length

			if retainLen < deleteLen {
				// Retain less than delete - delete some, then delete more
				result := Operation{OpType: OpDelete, Length: retainLen}
				*i++
				// Update second operation to delete remaining
				*j++
				return &result
			} else if retainLen == deleteLen {
				// Exact match - delete the retained content
				result := Operation{OpType: OpDelete, Length: deleteLen}
				*i++
				*j++
				return &result
			} else {
				// Retain more than delete - retain some, then delete remaining
				result := Operation{OpType: OpRetain, Length: deleteLen}
				*i++
				// Keep first operation's remaining retain
				remainingRetain := retainLen - deleteLen
				if remainingRetain > 0 {
					// Add remaining retain as a new operation
					return &result
				}
				*j++
				return &result
			}

		case OpRetain:
			// Both retain - take minimum
			minLen := retainLen
			if secondOp.Length < minLen {
				minLen = secondOp.Length
			}

			result := Operation{OpType: OpRetain, Length: minLen}
			*i++
			*j++

			// Handle remaining parts by NOT incrementing the index
			// that still has more to retain
			firstRemaining := retainLen - minLen
			secondRemaining := secondOp.Length - minLen

			if firstRemaining > 0 && secondRemaining > 0 {
				// Both have remaining - need another iteration
				// This shouldn't happen with min, but handle it
				*i-- // Put back first operation
				*j-- // Put back second operation
				return &result
			} else if firstRemaining > 0 {
				// First has remaining - put it back with updated length
				*i--
				firstOps[*i] = Operation{OpType: OpRetain, Length: firstRemaining}
				return &result
			} else if secondRemaining > 0 {
				// Second has remaining - put it back with updated length
				*j--
				secondOps[*j] = Operation{OpType: OpRetain, Length: secondRemaining}
				return &result
			}

			return &result

		case OpInsert:
			// Second operation inserts after retained range
			// Process the retain first, then insert will be handled
			result := Operation{OpType: OpRetain, Length: retainLen}
			*i++
			return &result
		}
	}

	// Should not reach here
	return &firstOp
}

// addOperation adds an operation to the changeset with fusion.
func (cs *ChangeSet) addOperation(op Operation) {
	// Try to fuse with last operation
	if len(cs.operations) > 0 {
		last := &cs.operations[len(cs.operations)-1]

		if last.OpType == op.OpType {
			switch op.OpType {
			case OpRetain:
				last.Length += op.Length
				return
			case OpDelete:
				last.Length += op.Length
				return
			case OpInsert:
				last.Text += op.Text
				return
			}
		}
	}

	cs.operations = append(cs.operations, op)
}

// recalculateLenAfter recalculates lenAfter based on operations.
func (cs *ChangeSet) recalculateLenAfter() {
	lenAfter := cs.lenBefore

	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain:
			// No change to length
		case OpDelete:
			lenAfter -= op.Length
		case OpInsert:
			lenAfter += len([]rune(op.Text))
		}
	}

	cs.lenAfter = lenAfter
}

// InvertAt creates an inverted changeset that undoes this changeset at a position.
func (cs *ChangeSet) InvertAt(original *Rope, pos int) *ChangeSet {
	if original == nil {
		return NewChangeSet(cs.lenAfter)
	}

	inverted := NewChangeSet(cs.lenAfter)
	currentPos := pos

	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain:
			currentPos += op.Length

		case OpDelete:
			// Re-insert the deleted text
			if currentPos+op.Length <= original.Length() {
				deletedText := original.Slice(currentPos, currentPos+op.Length)
				inverted.Insert(deletedText)
			}
			currentPos += op.Length

		case OpInsert:
			// Delete the inserted text
			inverted.Delete(len([]rune(op.Text)))
		}
	}

	// Fuse operations
	inverted.fuse()

	return inverted
}

// CanApplyAt checks if this changeset can be applied at the given position.
func (cs *ChangeSet) CanApplyAt(pos int) bool {
	return pos >= 0 && pos <= cs.lenBefore
}

// Optimized returns an optimized version of this changeset with fused operations.
func (cs *ChangeSet) Optimized() *ChangeSet {
	optimized := NewChangeSet(cs.lenBefore)
	optimized.operations = make([]Operation, len(cs.operations))
	copy(optimized.operations, cs.operations)
	optimized.fuse()
	return optimized
}

// Split splits this changeset at the given position.
// Returns two changesets: before and after the position.
func (cs *ChangeSet) Split(pos int) (*ChangeSet, *ChangeSet) {
	if pos <= 0 {
		return NewChangeSet(cs.lenBefore), cs
	}
	if pos >= cs.lenBefore {
		return cs, NewChangeSet(cs.lenAfter)
	}

	before := NewChangeSet(pos)
	after := NewChangeSet(cs.lenBefore - pos)

	currentPos := 0

	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain:
			if currentPos+op.Length <= pos {
				// Entire operation is before split point
				before.Retain(op.Length)
				currentPos += op.Length
			} else if currentPos >= pos {
				// Entire operation is after split point
				after.Retain(op.Length)
			} else {
				// Split the retain operation
				beforeLen := pos - currentPos
				afterLen := op.Length - beforeLen
				before.Retain(beforeLen)
				after.Retain(afterLen)
				currentPos += beforeLen
			}

		case OpDelete:
			if currentPos+op.Length <= pos {
				// Entire delete is before split point
				before.Delete(op.Length)
				currentPos += op.Length
			} else if currentPos >= pos {
				// Entire delete is after split point
				after.Delete(op.Length)
			} else {
				// Split the delete operation
				beforeLen := pos - currentPos
				afterLen := op.Length - beforeLen
				before.Delete(beforeLen)
				after.Delete(afterLen)
				currentPos += beforeLen
			}

		case OpInsert:
			// Inserts always happen at the current position
			if currentPos < pos {
				before.Insert(op.Text)
			} else {
				after.Insert(op.Text)
			}
		}
	}

	before.recalculateLenAfter()
	after.recalculateLenAfter()

	return before, after
}

// Merge merges this changeset with another at the same position.
// This is useful for combining concurrent edits.
func (cs *ChangeSet) Merge(other *ChangeSet) *ChangeSet {
	if cs.IsEmpty() {
		return other
	}
	if other.IsEmpty() {
		return cs
	}

	// Simply concatenate operations and fuse
	result := NewChangeSet(cs.lenBefore)
	result.operations = append(result.operations, cs.operations...)
	result.operations = append(result.operations, other.operations...)
	result.fuse()
	result.recalculateLenAfter()

	return result
}
