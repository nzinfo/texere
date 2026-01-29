package concordia

// OperationBuilder builds operations with automatic optimization.
//
// The builder pattern allows for efficient construction of operations by:
//   - Merging adjacent operations of the same type (retain, insert, delete)
//   - Removing no-op operations (retain(0), insert(""), delete(0))
//   - Providing a fluent API for chaining operations
//
// Example usage:
//
//	op := NewBuilder().
//	    Retain(5).
//	    Insert("Hello").
//	    Retain(3).
//	    Delete(2).
//	    Build()
//
// The builder automatically optimizes the operation sequence during construction.
// For example, retain(5).retain(3) is automatically merged to retain(8).
type OperationBuilder struct {
	ops            []Op
	baseLength     int
	targetLength   int
	optimizeEnabled bool
}

// NewBuilder creates a new operation builder with optimization enabled by default.
//
// Returns:
//   - a new OperationBuilder ready to build operations
//
// Example:
//
//	builder := NewBuilder()
//	op := builder.Retain(5).Insert("Hello").Build()
func NewBuilder() *OperationBuilder {
	return &OperationBuilder{
		ops:            make([]Op, 0, 16), // Pre-allocate for efficiency
		optimizeEnabled: true,
	}
}

// Retain appends a retain operation to the builder.
//
// Retain operations skip over characters without modifying them.
// Adjacent retain operations are automatically merged for efficiency.
//
// Parameters:
//   - n: number of characters to retain
//
// Returns:
//   - the builder for method chaining
//
// Example:
//
//	builder.Retain(5).Retain(3) // Merges to retain(8)
func (b *OperationBuilder) Retain(n int) *OperationBuilder {
	if n == 0 {
		return b // Skip no-op
	}

	// Optimization: merge with previous retain if possible
	if b.optimizeEnabled && len(b.ops) > 0 {
		if lastRetain, ok := b.ops[len(b.ops)-1].(RetainOp); ok {
			b.ops[len(b.ops)-1] = lastRetain + RetainOp(n)
			b.baseLength += n
			b.targetLength += n
			return b
		}
	}

	b.ops = append(b.ops, RetainOp(n))
	b.baseLength += n
	b.targetLength += n
	return b
}

// Insert appends an insert operation to the builder.
//
// Insert operations add new text at the current position.
// Adjacent insert operations are automatically merged for efficiency.
//
// Parameters:
//   - str: the string to insert
//
// Returns:
//   - the builder for method chaining
//
// Example:
//
//	builder.Insert("Hello").Insert(" World") // Merges to insert("Hello World")
func (b *OperationBuilder) Insert(str string) *OperationBuilder {
	if str == "" {
		return b // Skip no-op
	}

	// Optimization: merge with previous insert if possible
	if b.optimizeEnabled && len(b.ops) > 0 {
		if lastInsert, ok := b.ops[len(b.ops)-1].(InsertOp); ok {
			b.ops[len(b.ops)-1] = lastInsert + InsertOp(str)
			b.targetLength += len(str)
			return b
		}
	}

	b.ops = append(b.ops, InsertOp(str))
	b.targetLength += len(str)
	return b
}

// Delete appends a delete operation to the builder.
//
// Delete operations remove characters from the current position.
// Adjacent delete operations are automatically merged for efficiency.
//
// Parameters:
//   - n: number of characters to delete
//
// Returns:
//   - the builder for method chaining
//
// Example:
//
//	builder.Delete(2).Delete(3) // Merges to delete(5)
func (b *OperationBuilder) Delete(n int) *OperationBuilder {
	if n == 0 {
		return b // Skip no-op
	}

	// Optimization: merge with previous delete if possible
	if b.optimizeEnabled && len(b.ops) > 0 {
		if lastDelete, ok := b.ops[len(b.ops)-1].(DeleteOp); ok {
			b.ops[len(b.ops)-1] = lastDelete + DeleteOp(-n)
			b.baseLength += n // FIX: baseLength should increase (consume chars)
			return b
		}
	}

	b.ops = append(b.ops, DeleteOp(-n))
	b.baseLength += n // FIX: baseLength should increase (consume chars)
	return b
}

// BaseLength returns the current baseLength of the operation being built.
//
// This is useful for validation and testing purposes.
//
// Returns:
//   - the current baseLength
//
// Example:
//
//	builder := NewBuilder()
//	builder.Retain(5)
//	builder.BaseLength() // returns 5
func (b *OperationBuilder) BaseLength() int {
	return b.baseLength
}

// Build constructs an immutable operation from the builder.
//
// This method performs final optimization and creates an operation
// that is safe to use concurrently. The builder can be reused after
// calling Build().
//
// Returns:
//   - an immutable Operation containing the built operation sequence
//
// Example:
//
//	op := NewBuilder().
//	    Retain(5).
//	    Insert("Hello").
//	    Build()
func (b *OperationBuilder) Build() *Operation {
	if b.optimizeEnabled {
		b.optimize()
	}

	// Copy ops to ensure immutability
	opsCopy := make([]Op, len(b.ops))
	copy(opsCopy, b.ops)

	return &Operation{
		ops:          opsCopy,
		baseLength:   b.baseLength,
		targetLength: b.targetLength,
	}
}

// optimize performs final optimization on the operation sequence.
//
// This removes no-ops and ensures the operation sequence is in canonical form.
func (b *OperationBuilder) optimize() {
	optimized := make([]Op, 0, len(b.ops))

	for _, op := range b.ops {
		// Skip no-ops
		if op.Length() == 0 {
			continue
		}

		// Merge adjacent operations of the same type
		if len(optimized) > 0 && sameType(optimized[len(optimized)-1], op) {
			optimized[len(optimized)-1] = mergeOps(optimized[len(optimized)-1], op)
		} else {
			optimized = append(optimized, op)
		}
	}

	b.ops = optimized
}

// sameType checks if two operations are of the same concrete type.
func sameType(a, b Op) bool {
	switch a.(type) {
	case RetainOp:
		_, ok := b.(RetainOp)
		return ok
	case InsertOp:
		_, ok := b.(InsertOp)
		return ok
	case DeleteOp:
		_, ok := b.(DeleteOp)
		return ok
	}
	return false
}

// mergeOps merges two adjacent operations of the same type.
func mergeOps(a, b Op) Op {
	switch a := a.(type) {
	case RetainOp:
		return a + b.(RetainOp)
	case InsertOp:
		return a + b.(InsertOp)
	case DeleteOp:
		return a + b.(DeleteOp)
	default:
		return b
	}
}
