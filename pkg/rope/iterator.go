package rope

// ChangeIterator provides an iterator over the operations in a ChangeSet.
// It tracks the current position through the document as operations are applied.
type ChangeIterator struct {
	cs       *ChangeSet
	index    int
	position int // Current position in the document (chars processed)
}

// NewChangeIterator creates a new iterator for the given changeset.
func NewChangeIterator(cs *ChangeSet) *ChangeIterator {
	return &ChangeIterator{
		cs:       cs,
		index:    0,
		position: 0,
	}
}

// Next returns the next operation in the changeset, along with the document position
// where it applies. Returns nil when there are no more operations.
func (it *ChangeIterator) Next() *OperationInfo {
	if it.cs == nil || it.index >= len(it.cs.operations) {
		return nil
	}

	op := &it.cs.operations[it.index]
	info := &OperationInfo{
		Operation: op,
		Position:  it.position,
	}

	// Update position based on operation type
	switch op.OpType {
	case OpRetain:
		it.position += op.Length
	case OpDelete:
		// Delete removes content, so position stays the same
	case OpInsert:
		// Insert adds content, so position advances
		it.position += len([]rune(op.Text))
	}

	it.index++
	return info
}

// Reset resets the iterator to the beginning.
func (it *ChangeIterator) Reset() {
	it.index = 0
	it.position = 0
}

// Position returns the current document position.
func (it *ChangeIterator) Position() int {
	return it.position
}

// HasMore returns true if there are more operations to iterate.
func (it *ChangeIterator) HasMore() bool {
	return it.cs != nil && it.index < len(it.cs.operations)
}

// OperationInfo wraps an Operation with additional context.
type OperationInfo struct {
	Operation *Operation
	Position  int // Document position where this operation applies
}
