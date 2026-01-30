package rope

import "time"

// Operation represents a single edit operation.
type Operation struct {
	OpType OpType
	Length int       // For Retain and Delete
	Text   string    // For Insert
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

// Transaction represents an atomic edit operation with optional selection state.
type Transaction struct {
	changeset  *ChangeSet
	selection  *Selection  // Optional selection state
	timestamp  time.Time
}

// NewTransaction creates a new transaction from a changeset.
func NewTransaction(changeset *ChangeSet) *Transaction {
	return &Transaction{
		changeset: changeset,
		selection: nil,
		timestamp: time.Now(),
	}
}

// Changeset returns the transaction's changeset.
func (t *Transaction) Changeset() *ChangeSet {
	return t.changeset
}

// Timestamp returns when the transaction was created.
func (t *Transaction) Timestamp() time.Time {
	return t.timestamp
}

// Apply applies the transaction to a rope.
func (t *Transaction) Apply(r *Rope) *Rope {
	if t == nil || t.changeset == nil {
		return r
	}
	return t.changeset.Apply(r)
}

// Invert creates an inverted transaction for undo.
func (t *Transaction) Invert(original *Rope) *Transaction {
	if t == nil || t.changeset == nil {
		return NewTransaction(NewChangeSet(0))
	}
	return NewTransaction(t.changeset.Invert(original))
}

// IsEmpty returns true if the transaction has no changes.
func (t *Transaction) IsEmpty() bool {
	return t == nil || t.changeset == nil || t.changeset.IsEmpty()
}

// Selection returns the transaction's selection, if any.
func (t *Transaction) Selection() *Selection {
	return t.selection
}

// WithSelection returns a new transaction with the given selection.
func (t *Transaction) WithSelection(selection *Selection) *Transaction {
	if t == nil {
		return nil
	}
	return &Transaction{
		changeset: t.changeset,
		selection: selection,
		timestamp: t.timestamp,
	}
}

// Compose composes this transaction with another, combining their changesets.
// The selection from the other transaction takes precedence.
func (t *Transaction) Compose(other *Transaction) *Transaction {
	if t == nil {
		return other
	}
	if other == nil {
		return t
	}

	var composedCs *ChangeSet
	if t.changeset != nil && other.changeset != nil {
		composedCs = t.changeset.Compose(other.changeset)
	} else if other.changeset != nil {
		composedCs = other.changeset
	} else {
		composedCs = t.changeset
	}

	selection := other.selection
	if selection == nil {
		selection = t.selection
	}

	return &Transaction{
		changeset: composedCs,
		selection: selection,
		timestamp: time.Now(),
	}
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

// EditOperation represents a single edit operation with (from, to, replacement).
type EditOperation struct {
	From int
	To   int
	Text string // Empty string for deletion
}

// Deletion represents a deletion range (from, to).
type Deletion struct {
	From int
	To   int
}

// Change creates a transaction from a set of edit operations.
// Each edit is a tuple (from, to, replacementText) where:
// - from: start position of the change
// - to: end position of the change
// - replacementText: text to insert (empty for deletion)
// Changes must be ordered and non-overlapping.
func Change(doc *Rope, changes []EditOperation) *Transaction {
	if doc == nil {
		return NewTransaction(NewChangeSet(0))
	}

	len := doc.Length()
	cs := NewChangeSet(len)

	last := 0
	for _, ch := range changes {
		// Verify ranges are ordered
		if ch.From < last {
			// Overlapping or out of order - skip
			continue
		}
		if ch.From > ch.To {
			// Invalid range
			continue
		}

		// Retain from last "to" to current "from"
		if ch.From > last {
			cs.Retain(ch.From - last)
		}
		
		span := ch.To - ch.From
		if ch.Text != "" {
			cs.Insert(ch.Text)
			cs.Delete(span)
		} else {
			cs.Delete(span)
		}
		
		last = ch.To
	}

	// Retain remaining characters
	if len > last {
		cs.Retain(len - last)
	}

	return NewTransaction(cs)
}

// Delete creates a transaction from a set of deletions.
// Deletions can be overlapping - they will be merged.
func Delete(doc *Rope, deletions []Deletion) *Transaction {
	if doc == nil {
		return NewTransaction(NewChangeSet(0))
	}

	len := doc.Length()
	cs := NewChangeSet(len)

	// Sort and merge deletions
	// For now, we'll assume deletions are already sorted
	last := 0
	for _, del := range deletions {
		from := del.From
		to := del.To

		// Skip if this deletion is completely before last
		if to < last {
			continue
		}

		// Adjust from if it overlaps with last deletion
		if from < last {
			from = last
		}

		// Validate range
		if from > to {
			continue
		}

		// Retain from last to current from
		if from > last {
			cs.Retain(from - last)
		}

		// Delete the range
		cs.Delete(to - from)
		last = to
	}

	// Retain remaining characters
	if len > last {
		cs.Retain(len - last)
	}

	return NewTransaction(cs)
}

// InsertAtEOF inserts text at the end of the document.
// Returns a modified transaction with the insert operation added.
func (t *Transaction) InsertAtEOF(text string) *Transaction {
	if t == nil || t.changeset == nil {
		return NewTransaction(NewChangeSet(0).Insert(text))
	}
	
	newCs := t.changeset.Clone()
	newCs.Insert(text)
	
	return &Transaction{
		changeset: newCs,
		selection: t.selection,
		timestamp: time.Now(),
	}
}

// ChangeBySelection applies a change function to each range in the selection.
// The function receives a Range and returns an EditOperation to apply at that range.
func ChangeBySelection(doc *Rope, selection *Selection, f func(Range) EditOperation) *Transaction {
	if selection == nil || doc == nil {
		return NewTransaction(NewChangeSet(doc.Length()))
	}

	changes := make([]EditOperation, 0, selection.Len())
	for _, r := range selection.Iter() {
		change := f(r)
		changes = append(changes, change)
	}

	return Change(doc, changes)
}

// Insert inserts text at all cursor positions in the selection.
// For non-cursor ranges, the text is inserted at the head position.
func Insert(doc *Rope, selection *Selection, text string) *Transaction {
	if selection == nil || doc == nil {
		cs := NewChangeSet(doc.Length())
		cs.Insert(text)
		return NewTransaction(cs)
	}

	return ChangeBySelection(doc, selection, func(r Range) EditOperation {
		return EditOperation{
			From: r.Head,
			To:   r.Head,
			Text: text,
		}
	})
}

// DeleteBySelection applies a deletion function to each range in the selection.
// The function receives a Range and returns a Deletion (from, to) to apply.
func DeleteBySelection(doc *Rope, selection *Selection, f func(Range) Deletion) *Transaction {
	if selection == nil || doc == nil {
		return NewTransaction(NewChangeSet(doc.Length()))
	}

	deletions := make([]Deletion, 0, selection.Len())
	for _, r := range selection.Iter() {
		del := f(r)
		deletions = append(deletions, del)
	}

	return Delete(doc, deletions)
}

// ChangeBySelectionIgnoreOverlapping creates a transaction from potentially overlapping changes.
// Overlapping changes are ignored (only the first one is applied).
// Returns the transaction and the resulting (filtered) selection.
func ChangeBySelectionIgnoreOverlapping(
	doc *Rope,
	selection *Selection,
	changeRange func(Range) (int, int),
	createTendril func(int, int) string,
) (*Transaction, *Selection) {
	if doc == nil || selection == nil {
		return NewTransaction(NewChangeSet(0)), NewSelection()
	}

	type indexedRange struct {
		Index int
		Range Range
	}

	ranges := make([]indexedRange, 0, selection.Len())
	lastSelectionIdx := -1
	newPrimaryIdx := 0

	// Process ranges in order, tracking the primary selection
	for idx, r := range selection.Iter() {
		from, to := changeRange(r)
		
		// Skip if this range is completely before the last processed position
		if from < ranges[len(ranges)-1].Range.To {
			continue
		}
		
		// Add the range
		ranges = append(ranges, indexedRange{Index: idx, Range: r})
		
		// Track primary selection
		if idx == selection.PrimaryIndex() {
			newPrimaryIdx = len(ranges) - 1
		} else if newPrimaryIdx == -1 {
			if idx > selection.PrimaryIndex() {
				newPrimaryIdx = len(ranges) - 1
			}
		}
	}

	// Create changes from the filtered ranges
	changes := make([]EditOperation, 0, len(ranges))
	filteredRanges := make([]Range, 0, len(ranges))

	for _, ir := range ranges {
		text := createTendril(ir.Range.From, ir.Range.To)
		changes = append(changes, EditOperation{
			From: ir.Range.From,
			To:   ir.Range.To,
			Text: text,
		})
		filteredRanges = append(filteredRanges, ir.Range)
	}

	tx := Change(doc, changes)
	newSel := NewSelectionWithPrimary(filteredRanges, newPrimaryIdx)

	return tx, newSel
}

// ChangesIterator returns an iterator over the changeset's operations.
func (cs *ChangeSet) ChangesIterator() *ChangeIterator {
	return NewChangeIterator(cs)
}
