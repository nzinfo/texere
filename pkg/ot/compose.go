package ot

import (
	"errors"
)

// Compose combines two consecutive operations into a single operation.
//
// For operations A and B where A is applied before B, Compose creates a
// new operation C such that:
//
//	apply(apply(S, A), B) = apply(S, C)
//
// This is useful for combining sequences of operations, reducing overhead,
// and managing operation history.
//
// Parameters:
//   - operation1: the first operation (applied first)
//   - operation2: the second operation (applied after operation1)
//
// Returns:
//   - a composed operation that has the same effect as applying both
//   - an error if the operations cannot be composed
//
// Example:
//
//	op1 := NewBuilder().Insert("Hello ").Build()
//	op2 := NewBuilder().Retain(6).Insert("World").Build()
//	composed, _ := Compose(op1, op2)
//	// composed is equivalent to Insert("Hello World")
func Compose(operation1, operation2 *Operation) (*Operation, error) {
	if operation1.targetLength != operation2.baseLength {
		return nil, errors.New("the base length of the second operation has to be the target length of the first operation")
	}

	// IMPORTANT: The Insert-Delete swap rule in Builder.Insert() ensures
	// proper normalization, so we can use the regular builder with optimization.
	// This merges adjacent operations of the same type for canonical form.
	operation := NewBuilder()
	ops1 := operation1.ops
	ops2 := operation2.ops

	i1 := 0
	i2 := 0

	var op1 Op
	var op2 Op

	// Get initial ops
	if i1 < len(ops1) {
		op1 = ops1[i1]
		i1++
	}
	if i2 < len(ops2) {
		op2 = ops2[i2]
		i2++
	}

	for {
		if op1 == nil && op2 == nil {
			// End condition: both operations have been processed
			break
		}

		// Handle delete from op1 first (deletions happen before anything in op2)
		if op1 != nil && IsDelete(op1) {
			operation.Delete(op1.Length())
			if i1 < len(ops1) {
				op1 = ops1[i1]
				i1++
			} else {
				op1 = nil
			}
			continue
		}

		// Handle insert from op2 next (insertions happen before anything in op1)
		if op2 != nil && IsInsert(op2) {
			operation.Insert(string(op2.(InsertOp)))
			if i2 < len(ops2) {
				op2 = ops2[i2]
				i2++
			} else {
				op2 = nil
			}
			continue
		}

		// Check for missing operations
		if op1 == nil {
			return nil, errors.New("first operation is too short")
		}
		if op2 == nil {
			return nil, errors.New("second operation is too long")
		}

		// At this point, we have:
		// - op1 is not delete (so it's retain or insert)
		// - op2 is not insert (so it's retain or delete)

		if IsRetain(op1) && IsRetain(op2) {
			// Both retain - retain the minimum
			minl := min(op1.Length(), op2.Length())
			operation.Retain(minl)

			if op1.Length() > op2.Length() {
				op1 = RetainOp(op1.Length() - op2.Length())
				if i2 < len(ops2) {
					op2 = ops2[i2]
					i2++
				} else {
					op2 = nil
				}
			} else if op1.Length() < op2.Length() {
				op2 = RetainOp(op2.Length() - op1.Length())
				if i1 < len(ops1) {
					op1 = ops1[i1]
					i1++
				} else {
					op1 = nil
				}
			} else {
				// Equal lengths
				if i1 < len(ops1) {
					op1 = ops1[i1]
					i1++
				} else {
					op1 = nil
				}
				if i2 < len(ops2) {
					op2 = ops2[i2]
					i2++
				} else {
					op2 = nil
				}
			}
		} else if IsInsert(op1) && IsDelete(op2) {
			// Insert and delete - cancel each other out
			if op1.Length() > op2.Length() {
				op1 = InsertOp(string(op1.(InsertOp))[op2.Length():])
				if i2 < len(ops2) {
					op2 = ops2[i2]
					i2++
				} else {
					op2 = nil
				}
			} else if op1.Length() < op2.Length() {
				op2 = DeleteOp(-(op2.Length() - op1.Length()))
				if i1 < len(ops1) {
					op1 = ops1[i1]
					i1++
				} else {
					op1 = nil
				}
			} else {
				// Equal lengths - they cancel completely
				if i1 < len(ops1) {
					op1 = ops1[i1]
					i1++
				} else {
					op1 = nil
				}
				if i2 < len(ops2) {
					op2 = ops2[i2]
					i2++
				} else {
					op2 = nil
				}
			}
		} else if IsInsert(op1) && IsRetain(op2) {
			// Insert and retain - insert part of the string
			if op1.Length() > op2.Length() {
				// Insert the first part
				str := string(op1.(InsertOp))
				operation.Insert(str[:op2.Length()])
				op1 = InsertOp(str[op2.Length():])
				if i2 < len(ops2) {
					op2 = ops2[i2]
					i2++
				} else {
					op2 = nil
				}
			} else if op1.Length() < op2.Length() {
				// Insert the whole string
				operation.Insert(string(op1.(InsertOp)))
				op2 = RetainOp(op2.Length() - op1.Length())
				if i1 < len(ops1) {
					op1 = ops1[i1]
					i1++
				} else {
					op1 = nil
				}
			} else {
				// Equal lengths
				operation.Insert(string(op1.(InsertOp)))
				if i1 < len(ops1) {
					op1 = ops1[i1]
					i1++
				} else {
					op1 = nil
				}
				if i2 < len(ops2) {
					op2 = ops2[i2]
					i2++
				} else {
					op2 = nil
				}
			}
		} else if IsRetain(op1) && IsDelete(op2) {
			// Retain and delete - delete from the document
			if op1.Length() > op2.Length() {
				operation.Delete(op2.Length())
				op1 = RetainOp(op1.Length() - op2.Length())
				if i2 < len(ops2) {
					op2 = ops2[i2]
					i2++
				} else {
					op2 = nil
				}
			} else if op1.Length() < op2.Length() {
				operation.Delete(op1.Length())
				op2 = DeleteOp(-(op2.Length() - op1.Length()))
				if i1 < len(ops1) {
					op1 = ops1[i1]
					i1++
				} else {
					op1 = nil
				}
			} else {
				// Equal lengths
				operation.Delete(op1.Length())
				if i1 < len(ops1) {
					op1 = ops1[i1]
					i1++
				} else {
					op1 = nil
				}
				if i2 < len(ops2) {
					op2 = ops2[i2]
					i2++
				} else {
					op2 = nil
				}
			}
		} else {
			return nil, errors.New("invalid operation combination")
		}
	}

	return operation.Build(), nil
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
