package ot

import (
	"errors"
)

// Transform transforms two concurrent operations against each other.
//
// This is the core OT algorithm. Given two operations that were applied
// concurrently to the same document state, Transform produces two new
// operations such that:
//
//	apply(apply(S, A), B') = apply(apply(S, B), A')
//
// where (A', B') = Transform(A, B).
//
// This ensures that concurrent operations converge to the same final state
// regardless of the order in which they are applied.
//
// Parameters:
//   - operation1: the first operation
//   - operation2: the second operation
//
// Returns:
//   - operation1': the transformed version of operation1
//   - operation2': the transformed version of operation2
//   - an error if the operations are incompatible
//
// Example:
//
//	// Two users concurrently edit at position 0
//	op1 := NewBuilder().Insert("Hello").Build()
//	op2 := NewBuilder().Insert("Hi").Build()
//	op1Prime, op2Prime := Transform(op1, op2)
//	// Now op1' and op2' can be applied in any order
func Transform(operation1, operation2 *Operation) (*Operation, *Operation, error) {
	if operation1.baseLength != operation2.baseLength {
		return nil, nil, errors.New("both operations must have the same base length")
	}

	operation1Prime := NewBuilder()
	operation2Prime := NewBuilder()

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
		// End condition: both ops1 and ops2 have been processed
		if op1 == nil && op2 == nil {
			break
		}

		// Handle insert operations first (they don't conflict)
		if op1 != nil && IsInsert(op1) {
			operation1Prime.Insert(string(op1.(InsertOp)))
			operation2Prime.Retain(op1.Length())
			if i1 < len(ops1) {
				op1 = ops1[i1]
				i1++
			} else {
				op1 = nil
			}
			continue
		}

		if op2 != nil && IsInsert(op2) {
			operation1Prime.Retain(op2.Length())
			operation2Prime.Insert(string(op2.(InsertOp)))
			if i2 < len(ops2) {
				op2 = ops2[i2]
				i2++
			} else {
				op2 = nil
			}
			continue
		}

		// At this point, both ops must be non-nil and not inserts
		if op1 == nil {
			return nil, nil, errors.New("first operation is too short")
		}
		if op2 == nil {
			return nil, nil, errors.New("second operation is too short")
		}

		// Handle retain/retain
		if IsRetain(op1) && IsRetain(op2) {
			minl := min(op1.Length(), op2.Length())

			operation1Prime.Retain(minl)
			operation2Prime.Retain(minl)

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
		} else if IsDelete(op1) && IsDelete(op2) {
			// Both delete - just skip, they don't conflict
			if op1.Length() > op2.Length() {
				op1 = DeleteOp(-(op1.Length() - op2.Length()))
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
		} else if IsDelete(op1) && IsRetain(op2) {
			// Delete and retain
			minl := min(op1.Length(), op2.Length())

			operation1Prime.Delete(minl)

			if op1.Length() > op2.Length() {
				op1 = DeleteOp(-(op1.Length() - op2.Length()))
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
		} else if IsRetain(op1) && IsDelete(op2) {
			// Retain and delete
			minl := min(op1.Length(), op2.Length())

			operation2Prime.Delete(minl)

			if op1.Length() > op2.Length() {
				op1 = RetainOp(op1.Length() - op2.Length())
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
		} else {
			return nil, nil, errors.New("incompatible operation types")
		}
	}

	return operation1Prime.Build(), operation2Prime.Build(), nil
}
