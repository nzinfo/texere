package concordia

import (
	"github.com/coreseekdev/texere/pkg/ot"
	"github.com/coreseekdev/texere/pkg/rope"
)

// ========== OT Operation Helpers (Recommended API) ==========

// OperationFromChanges creates an ot.Operation from a set of edit operations.
// This is the recommended way to create operations for document editing.
//
// Example:
//
//	changes := []rope.EditOperation{
//	    {From: 5, To: 10, Text: "World"},
//	    {From: 15, To: 20, Text: ""},
//	}
//	op := OperationFromChanges(doc, changes)
func OperationFromChanges(doc *rope.Rope, changes []rope.EditOperation) *ot.Operation {
	if doc == nil {
		return ot.NewOperation()
	}

	length := doc.Length()
	builder := ot.NewBuilder()

	last := 0
	for _, ch := range changes {
		// Verify ranges are ordered
		if ch.From < last {
			// Skip overlapping or out-of-order changes
			continue
		}
		if ch.From > ch.To {
			// Invalid range
			continue
		}

		// Retain from last "to" to current "from"
		if ch.From > last {
			builder.Retain(ch.From - last)
		}

		span := ch.To - ch.From
		if ch.Text != "" {
			// Replace: delete then insert
			builder.Delete(span)
			builder.Insert(ch.Text)
		} else {
			// Just delete
			builder.Delete(span)
		}

		last = ch.To
	}

	// Retain remaining characters
	if length > last {
		builder.Retain(length - last)
	}

	return builder.Build()
}

// OperationFromDeletions creates an ot.Operation from a set of deletions.
// Deletions can be overlapping - they will be merged.
func OperationFromDeletions(doc *rope.Rope, deletions []rope.Deletion) *ot.Operation {
	if doc == nil {
		return ot.NewOperation()
	}

	length := doc.Length()
	builder := ot.NewBuilder()

	// Sort and merge deletions
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
			builder.Retain(from - last)
		}

		// Delete the range
		builder.Delete(to - from)
		last = to
	}

	// Retain remaining characters
	if length > last {
		builder.Retain(length - last)
	}

	return builder.Build()
}

// InsertOperation creates an insert operation at the specified position.
func InsertOperation(doc *rope.Rope, pos int, text string) *ot.Operation {
	if doc == nil {
		builder := ot.NewBuilder()
		builder.Insert(text)
		return builder.Build()
	}

	length := doc.Length()
	builder := ot.NewBuilder()

	if pos > 0 {
		builder.Retain(pos)
	}
	builder.Insert(text)
	if length > pos {
		builder.Retain(length - pos)
	}

	return builder.Build()
}

// DeleteOperation creates a delete operation for the specified range.
func DeleteOperation(doc *rope.Rope, from, to int) *ot.Operation {
	if doc == nil {
		return ot.NewOperation()
	}

	length := doc.Length()
	builder := ot.NewBuilder()

	if from > 0 {
		builder.Retain(from)
	}
	if to > from {
		builder.Delete(to - from)
	}
	if length > to {
		builder.Retain(length - to)
	}

	return builder.Build()
}

// ========== Rope OT Integration ==========

// ApplyOperation applies an OT operation to the rope and returns a new Rope.
// This is an adapter function that bridges rope with the ot package.
//
// This function is intentionally separated from core rope functionality
// to allow rope to remain independent from ot in future iterations.
//
// Example:
//
//	op := ot.NewBuilder().Retain(5).Insert("World").Build()
//	newRope, err := ApplyOperation(doc, op)
func ApplyOperation(r *rope.Rope, op *ot.Operation) (*rope.Rope, error) {
	if r == nil || op == nil {
		return r, nil
	}

	// Validate operation length matches rope length
	if op.BaseLength() != r.Length() {
		return nil, ot.ErrInvalidBaseLength
	}

	// Create RopeDocument to apply operation
	doc := NewRopeDocumentFromRope(r)
	result, err := op.ApplyToDocument(doc)
	if err != nil {
		return nil, err
	}

	// Convert back to Rope
	if ropeDoc, ok := result.(*RopeDocument); ok {
		return ropeDoc.rope, nil
	}

	// Fallback: create rope from string
	return rope.New(result.String()), nil
}
