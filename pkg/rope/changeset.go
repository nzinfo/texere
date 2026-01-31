package rope

// Operation represents a single edit operation for Rope's internal ChangeSet.
// This is different from ot.Operation - this is Rope's internal representation.
type Operation struct {
	OpType OpType
	Length int    // For Retain and Delete
	Text   string // For Insert
}

// OpType represents the type of operation.
type OpType int

const (
	OpRetain OpType = iota // Keep n characters
	OpDelete               // Delete n characters
	OpInsert               // Insert text
)

// ChangeSet represents a set of changes to transform one document state to another.
// It is composable and invertible, making it ideal for undo/redo.
//
// This is Rope's internal representation. For OT operations, use ot.Operation instead.
type ChangeSet struct {
	operations []Operation
	lenBefore  int // Document length before changes
	lenAfter   int // Document length after changes
}

// NewChangeSet creates a new empty ChangeSet.
func NewChangeSet(lenBefore int) *ChangeSet {
	return &ChangeSet{
		operations: make([]Operation, 0, 8),
		lenBefore:  lenBefore,
		lenAfter:   lenBefore,
	}
}

// Retain keeps n characters unchanged.
func (cs *ChangeSet) Retain(n int) *ChangeSet {
	cs.operations = append(cs.operations, Operation{OpType: OpRetain, Length: n})
	return cs
}

// Delete deletes n characters.
func (cs *ChangeSet) Delete(n int) *ChangeSet {
	cs.operations = append(cs.operations, Operation{OpType: OpDelete, Length: n})
	cs.lenAfter -= n
	return cs
}

// Insert inserts text.
func (cs *ChangeSet) Insert(text string) *ChangeSet {
	cs.operations = append(cs.operations, Operation{OpType: OpInsert, Text: text})
	cs.lenAfter += len([]rune(text))
	return cs
}

// LenBefore returns the document length before applying this changeset.
func (cs *ChangeSet) LenBefore() int {
	return cs.lenBefore
}

// LenAfter returns the document length after applying this changeset.
func (cs *ChangeSet) LenAfter() int {
	return cs.lenAfter
}

// IsEmpty returns true if the changeset has no operations.
func (cs *ChangeSet) IsEmpty() bool {
	return len(cs.operations) == 0
}

// finalize ensures the changeset covers the entire document by retaining
// any remaining characters. This follows Helix's approach where changesets
// must account for every character in the input document.
func (cs *ChangeSet) finalize() *ChangeSet {
	// Calculate how many characters have been processed
	processed := 0
	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain, OpDelete:
			processed += op.Length
		case OpInsert:
			// Inserts don't consume input characters
		}
	}

	// Retain remaining characters to reach lenBefore
	remaining := cs.lenBefore - processed
	if remaining > 0 {
		cs.Retain(remaining)
	}

	return cs
}

// fuse merges consecutive operations of the same type for optimization.
// This reduces the number of operations and improves performance.
// For example: Insert("a") + Insert("b") → Insert("ab")
func (cs *ChangeSet) fuse() {
	if len(cs.operations) <= 1 {
		return
	}

	fused := make([]Operation, 0, len(cs.operations))

	for _, op := range cs.operations {
		if len(fused) > 0 && fused[len(fused)-1].OpType == op.OpType {
			// Merge with previous operation of the same type
			prev := &fused[len(fused)-1]
			switch op.OpType {
			case OpRetain:
				prev.Length += op.Length
			case OpDelete:
				prev.Length += op.Length
			case OpInsert:
				prev.Text += op.Text
			}
		} else {
			fused = append(fused, op)
		}
	}

	cs.operations = fused
}

// Apply applies the changeset to a rope and returns the modified rope.
func (cs *ChangeSet) Apply(r *Rope) *Rope {
	if r == nil || cs.IsEmpty() {
		return r
	}

	// Check if document length matches changeset's expected input length
	if r.Length() != cs.lenBefore {
		// Length mismatch - cannot apply
		return r
	}

	// Make a copy to finalize (don't modify original)
	csCopy := NewChangeSet(cs.lenBefore)
	csCopy.operations = make([]Operation, len(cs.operations))
	copy(csCopy.operations, cs.operations)
	csCopy.lenAfter = cs.lenAfter
	csCopy.finalize()

	// Fuse operations for optimization (reduces number of rope mutations)
	csCopy.fuse()

	result := r
	pos := 0

	for _, op := range csCopy.operations {
		switch op.OpType {
		case OpRetain:
			pos += op.Length

		case OpDelete:
			result = result.Delete(pos, pos+op.Length)
			// Delete removes content, so pos stays the same

		case OpInsert:
			result = result.Insert(pos, op.Text)
			pos += len([]rune(op.Text))
		}
	}

	return result
}

// Invert creates an inverted changeset that undoes this changeset.
// The original rope state is needed to properly invert deletions.
func (cs *ChangeSet) Invert(original *Rope) *ChangeSet {
	if original == nil {
		return NewChangeSet(cs.lenAfter)
	}

	inverted := NewChangeSet(cs.lenAfter)
	pos := 0

	for _, op := range cs.operations {
		switch op.OpType {
		case OpRetain:
			inverted.Retain(op.Length)
			pos += op.Length

		case OpDelete:
			// Re-insert the deleted text
			deletedText := original.Slice(pos, pos+op.Length)
			inverted.Insert(deletedText)
			pos += op.Length

		case OpInsert:
			// Delete the inserted text
			inverted.Delete(len([]rune(op.Text)))
		}
	}

	// Fuse operations in the inverted changeset for optimization
	inverted.fuse()

	return inverted
}

// MapPosition maps a single position through this changeset with the given association.
func (cs *ChangeSet) MapPosition(pos int, assoc Assoc) int {
	mapper := NewPositionMapper(cs)
	mapper.AddPosition(pos, assoc)
	result := mapper.Map()
	if len(result) == 0 {
		return pos
	}
	return result[0]
}

// MapPositions maps multiple positions through this changeset with the given associations.
func (cs *ChangeSet) MapPositions(positions []int, associations []Assoc) []int {
	mapper := NewPositionMapper(cs)
	for i, pos := range positions {
		assoc := AssocBefore
		if i < len(associations) {
			assoc = associations[i]
		}
		mapper.AddPosition(pos, assoc)
	}
	return mapper.Map()
}

// Transform transforms this changeset to apply after another changeset.
// This is used for operational transformation in concurrent editing.
func (cs *ChangeSet) Transform(other *ChangeSet) *ChangeSet {
	if other == nil || other.IsEmpty() {
		result := NewChangeSet(cs.lenBefore)
		result.operations = make([]Operation, len(cs.operations))
		copy(result.operations, cs.operations)
		result.lenAfter = cs.lenAfter
		return result
	}

	if cs == nil || cs.IsEmpty() {
		result := NewChangeSet(other.lenBefore)
		result.operations = make([]Operation, len(other.operations))
		copy(result.operations, other.operations)
		result.lenAfter = other.lenAfter
		return result
	}

	// For now, use simple merge as placeholder
	// A full OT-based transform would require more complex logic
	result := NewChangeSet(cs.lenBefore)
	result.operations = append(result.operations, cs.operations...)
	result.lenAfter = cs.lenAfter
	return result
}

// ChangesIterator returns an iterator over the changeset's operations.
func (cs *ChangeSet) ChangesIterator() *ChangeIterator {
	return NewChangeIterator(cs)
}
