package ot

// UndoableDocument extends Document with undo/redo capabilities.
type UndoableDocument interface {
	Document

	// Undo reverses the last operation.
	Undo() error

	// Redo reapplies the most recently undone operation.
	Redo() error

	// CanUndo returns true if undo is possible.
	CanUndo() bool

	// CanRedo returns true if redo is possible.
	CanRedo() bool

	// ApplyOperationWithHistory applies an operation and records it to history.
	ApplyOperationWithHistory(op *Operation) (UndoableDocument, error)
}
