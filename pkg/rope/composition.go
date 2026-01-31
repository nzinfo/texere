package rope

// Compose composes this changeset with another, producing a changeset that
// represents applying this changeset followed by the other.
// This implementation follows Helix editor's approach with OT-based composition.
// See: https://github.com/helix-editor/helix/blob/master/helix-core/src/transaction.rs#L163
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
	firstOps := csFinal.operations
	secondOps := otherFinal.operations
	i, j := 0, 0

	// Helix-style composition loop
	for i < len(firstOps) || j < len(secondOps) {
		// Get current operations (use placeholder if exhausted)
		var firstOp, secondOp *Operation
		if i < len(firstOps) {
			firstOp = &firstOps[i]
		}
		if j < len(secondOps) {
			secondOp = &secondOps[j]
		}

		// Rule 1: Deletion in first (A) has highest priority
		if firstOp != nil && firstOp.OpType == OpDelete {
			// Check if secondOp is also Delete
			if secondOp != nil && secondOp.OpType == OpDelete {
				// Delete(A) + Delete(B): merge them
				// Delete(B) wants to delete deleteLen chars, Delete(A) deletes deleteALen chars
				// These are deleting the same content, so output Delete(deleteALen)
				// and reduce Delete(B) by deleteALen
				deleteALen := firstOp.Length
				deleteBLen := secondOp.Length

				if deleteALen < deleteBLen {
					// Delete(A) is smaller - output Delete(deleteALen), reduce Delete(B)
					result.addOperation(*firstOp)
					i++
					// Put back reduced Delete(B)
					secondOps[j] = Operation{OpType: OpDelete, Length: deleteBLen - deleteALen}
					continue
				} else if deleteALen == deleteBLen {
					// Delete(A) == Delete(B) - output Delete(deleteALen), consume both
					result.addOperation(*firstOp)
					i++
					j++
					continue
				} else {
					// Delete(A) is larger - output Delete(deleteBLen), reduce Delete(A)
					result.addOperation(Operation{OpType: OpDelete, Length: deleteBLen})
					j++
					// Put back reduced Delete(A)
					i++
					firstOps[i-1] = Operation{OpType: OpDelete, Length: deleteALen - deleteBLen}
					continue
				}
			} else {
				// Delete(A) with non-Delete(B): output Delete(A) as-is
				result.addOperation(*firstOp)
				i++
				// Don't increment j - keep second operation for next iteration
				continue
			}
		}

		// Rule 2: Insertion in second (B) has highest priority
		// But NOT when first is Delete (Delete won't be skipped)
		if secondOp != nil && secondOp.OpType == OpInsert {
			if firstOp == nil || firstOp.OpType != OpDelete {
				result.addOperation(*secondOp)
				j++
				// Don't increment i - keep first operation for next iteration
				continue
			}
			// If first is Delete, fall through - Delete(A) already handled this
		}

		// If we get here, both operations are present and neither is Delete(A) nor Insert(B)
		if firstOp == nil {
			// Only second operations remaining
			result.addOperation(*secondOp)
			j++
			continue
		}

		if secondOp == nil {
			// Only first operations remaining
			result.addOperation(*firstOp)
			i++
			continue
		}

		// Compose the two operations
		composed := composeOperations(*firstOp, *secondOp, &i, &j, firstOps, secondOps)
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
// Returns the composed operation, or nil if operations should be skipped.
// Updates indices i and j as needed.
// This implements the remaining cases after handling Delete(A) and Insert(B) priority.
func composeOperations(firstOp, secondOp Operation, i, j *int, firstOps, secondOps []Operation) *Operation {
	switch firstOp.OpType {
	case OpInsert:
		// First operation inserts text
		insertText := firstOp.Text
		insertLen := len([]rune(insertText))

		switch secondOp.OpType {
		case OpDelete:
			// Delete in second operation
			deleteLen := secondOp.Length

			if insertLen < deleteLen {
				// Insert is shorter than delete - both operations cancel out (insert is deleted)
				*i++
				*j++
				// Put back remaining delete
				remainingDelete := deleteLen - insertLen
				if remainingDelete > 0 {
					*j--
					secondOps[*j] = Operation{OpType: OpDelete, Length: remainingDelete}
				}
				return nil // Skip the Insert (it's deleted)
			} else if insertLen == deleteLen {
				// Exact match - both operations cancel out
				*i++
				*j++
				return nil // Skip both
			} else {
				// Insert is longer than delete - part of insert is deleted
				*i++
				*j++
				// Put back the remaining insert
				remainingInsert := string([]rune(insertText)[deleteLen:])
				*i--
				firstOps[*i] = Operation{OpType: OpInsert, Text: remainingInsert}
				return nil // Skip the Delete (consumed by insert)
			}

		case OpRetain:
			// Second operation retains
			// Insert doesn't consume characters, Retain does
			// So we should output Insert, and the Retain will be processed later
			result := Operation{OpType: OpInsert, Text: insertText}
			*i++
			// Don't increment j - the Retain will consume characters from Retain(A) or Delete(A) operations
			return &result

		case OpInsert:
			// This shouldn't happen - Insert(B) is handled with priority
			// But if we get here, just return the second insert
			*j++
			return &secondOp
		}

	case OpRetain:
		// First operation retains some characters
		retainLen := firstOp.Length

		switch secondOp.OpType {
		case OpDelete:
			// Delete in second operation
			deleteLen := secondOp.Length

			if retainLen < deleteLen {
				// Retain less than delete - delete the retained content
				result := Operation{OpType: OpDelete, Length: retainLen}
				*i++
				// Put back delete with reduced length
				*j++
				remainingDelete := deleteLen - retainLen
				if remainingDelete > 0 {
					*j--
					secondOps[*j] = Operation{OpType: OpDelete, Length: remainingDelete}
				}
				return &result
			} else if retainLen == deleteLen {
				// Exact match - delete the retained content
				result := Operation{OpType: OpDelete, Length: deleteLen}
				*i++
				*j++
				return &result
			} else {
				// Retain more than delete - delete some, retain the rest
				result := Operation{OpType: OpDelete, Length: deleteLen}
				*i++
				*j++

				// Put back remaining retain
				remainingRetain := retainLen - deleteLen
				if remainingRetain > 0 {
					*i--
					firstOps[*i] = Operation{OpType: OpRetain, Length: remainingRetain}
				}
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

			// Handle remaining parts
			firstRemaining := retainLen - minLen
			secondRemaining := secondOp.Length - minLen

			if firstRemaining > 0 && secondRemaining > 0 {
				// Both have remaining - put back both
				*i--
				*j--
				firstOps[*i] = Operation{OpType: OpRetain, Length: firstRemaining}
				secondOps[*j] = Operation{OpType: OpRetain, Length: secondRemaining}
				return &result
			} else if firstRemaining > 0 {
				// First has remaining - put it back
				*i--
				firstOps[*i] = Operation{OpType: OpRetain, Length: firstRemaining}
				return &result
			} else if secondRemaining > 0 {
				// Second has remaining - put it back
				*j--
				secondOps[*j] = Operation{OpType: OpRetain, Length: secondRemaining}
				return &result
			}

			return &result

		case OpInsert:
			// This shouldn't happen - Insert(B) is handled with priority
			// But if we get here, just return the insert
			return &secondOp
		}
	}

	// Should not reach here
	return nil
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
// Both changesets can be applied to the original document independently.
func (cs *ChangeSet) Split(pos int) (*ChangeSet, *ChangeSet) {
	if pos <= 0 {
		return NewChangeSet(cs.lenBefore), cs.clone()
	}
	if pos >= cs.lenBefore {
		return cs.clone(), NewChangeSet(cs.lenBefore)
	}

	// Both changesets have the same lenBefore as original
	before := NewChangeSet(cs.lenBefore)
	after := NewChangeSet(cs.lenBefore)

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
			// Inserts at current position
			if currentPos < pos {
				before.Insert(op.Text)
			} else {
				after.Insert(op.Text)
			}
		}
	}

	// Finalize both changesets to ensure they cover entire document
	// before needs to retain the rest of the document unchanged
	// after needs to retain the prefix of the document unchanged
	before.retainRemaining(pos)
	after.retainPrefix(pos)

	before.recalculateLenAfter()
	after.recalculateLenAfter()

	return before, after
}

// retainRemaining retains all characters from current position to end
func (cs *ChangeSet) retainRemaining(fromPos int) {
	processed := 0
	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain, OpDelete:
			processed += op.Length
		case OpInsert:
			// Inserts don't consume original characters
		}
	}
	remaining := cs.lenBefore - processed
	if remaining > 0 {
		cs.Retain(remaining)
	}
}

// retainPrefix retains characters from start to given position
func (cs *ChangeSet) retainPrefix(toPos int) {
	processed := 0
	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain, OpDelete:
			if processed+op.Length <= toPos {
				processed += op.Length
			} else {
				// Need to split or skip
				remaining := toPos - processed
				if remaining > 0 {
					cs.Retain(remaining)
				}
				processed = toPos
			}
		case OpInsert:
			// Inserts don't consume original characters
		}
	}
}

// Merge merges this changeset with another at the same position.
// This is useful for combining concurrent edits.
// Both changesets should be based on the same document state.
func (cs *ChangeSet) Merge(other *ChangeSet) *ChangeSet {
	if cs.IsEmpty() {
		return other
	}
	if other.IsEmpty() {
		return cs
	}

	// Check if both changesets are based on the same document
	if cs.lenBefore != other.lenBefore {
		// Cannot merge different base documents
		// Return cs as fallback
		return cs
	}

	// Apply first changeset to calculate the document state after cs
	tempLen := cs.lenAfter

	// Adjust second changeset's operations based on first changeset's effect
	result := NewChangeSet(cs.lenBefore)
	result.operations = make([]Operation, 0, len(cs.operations)+len(other.operations))

	// Add first changeset's operations
	result.operations = append(result.operations, cs.operations...)

	// Add second changeset's operations, adjusting positions
	// This is a simplified version - proper OT transform would be more complex
	for _, op := range other.operations {
		switch op.OpType {
		case OpRetain:
			// Check if the retain position is within bounds
			if tempLen >= op.Length {
				result.Retain(op.Length)
			}
			// Skip if out of bounds
		case OpDelete:
			if tempLen >= op.Length {
				result.Delete(op.Length)
				tempLen -= op.Length
			}
			// Skip if out of bounds
		case OpInsert:
			result.Insert(op.Text)
			tempLen += len([]rune(op.Text))
		}
	}

	result.fuse()
	result.recalculateLenAfter()

	return result
}
