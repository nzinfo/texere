package concordia

import (
	"github.com/coreseekdev/texere/pkg/rope"
)

// EditOperation represents a single edit operation with (from, to, replacement).
// This is an alias for rope.EditOperation for use in OT operations.
type EditOperation = rope.EditOperation

// Deletion represents a deletion range (from, to).
// This is an alias for rope.Deletion for use in OT operations.
type Deletion = rope.Deletion
