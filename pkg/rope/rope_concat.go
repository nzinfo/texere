package rope

import (
	"unicode/utf8"
)

// ========== Rope Concatenation ==========

// AppendRope appends another rope to the end of this rope.
// Returns a new Rope, leaving both original ropes unchanged.
// This is more efficient than converting the rope to a string and appending.
func (r *Rope) AppendRope(other *Rope) *Rope {
	if r == nil || r.Length() == 0 {
		return other.Clone()
	}
	if other == nil || other.Length() == 0 {
		return r.Clone()
	}

	// Create a new internal node that joins both ropes
	return &Rope{
		root: &InternalNode{
			left:   r.root,
			right:  other.root,
			length: r.Length(),
			size:   r.Size(),
		},
		length: r.Length() + other.Length(),
		size:   r.Size() + other.Size(),
	}
}

// PrependRope prepends another rope to the beginning of this rope.
// Returns a new Rope, leaving both original ropes unchanged.
func (r *Rope) PrependRope(other *Rope) *Rope {
	if r == nil || r.Length() == 0 {
		return other.Clone()
	}
	if other == nil || other.Length() == 0 {
		return r.Clone()
	}

	// Create a new internal node with other as left child
	return &Rope{
		root: &InternalNode{
			left:   other.root,
			right:  r.root,
			length: other.Length(),
			size:   other.Size(),
		},
		length: other.Length() + r.Length(),
		size:   other.Size() + r.Size(),
	}
}

// Concat concatenates multiple ropes together.
// Returns a new Rope, leaving all original ropes unchanged.
func Concat(ropes ...*Rope) *Rope {
	if len(ropes) == 0 {
		return Empty()
	}
	if len(ropes) == 1 {
		return ropes[0].Clone()
	}

	// Filter out empty ropes
	nonEmpty := make([]*Rope, 0, len(ropes))
	for _, r := range ropes {
		if r != nil && r.Length() > 0 {
			nonEmpty = append(nonEmpty, r)
		}
	}

	if len(nonEmpty) == 0 {
		return Empty()
	}
	if len(nonEmpty) == 1 {
		return nonEmpty[0].Clone()
	}

	// Build balanced tree of ropes
	return concatBalanced(nonEmpty, 0, len(nonEmpty))
}

// concatBalanced recursively builds a balanced tree of ropes.
func concatBalanced(ropes []*Rope, start, end int) *Rope {
	count := end - start
	if count == 0 {
		return Empty()
	}
	if count == 1 {
		return ropes[start].Clone()
	}
	if count == 2 {
		return ropes[start].AppendRope(ropes[start+1])
	}

	mid := start + count/2
	left := concatBalanced(ropes, start, mid)
	right := concatBalanced(ropes, mid, end)

	return left.AppendRope(right)
}

// Join joins multiple ropes with a separator between them.
// Returns a new Rope, leaving all original ropes unchanged.
func (r *Rope) Join(ropes []*Rope, separator string) *Rope {
	if len(ropes) == 0 {
		return Empty()
	}
	if len(ropes) == 1 {
		return ropes[0].Clone()
	}

	sep := New(separator)
	result := ropes[0].Clone()

	for i := 1; i < len(ropes); i++ {
		result = result.AppendRope(sep)
		result = result.AppendRope(ropes[i])
	}

	return result
}

// ========== String Append/Prepend ==========

// AppendStr appends a string to the end of the rope.
// Optimized version that directly creates nodes instead of using Insert().
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) AppendStr(text string) *Rope {
	if r == nil {
		return New(text)
	}
	if text == "" {
		return r
	}
	if r.length == 0 {
		return New(text)
	}

	// Create rope from text and append it directly
	textRope := New(text)

	return &Rope{
		root: &InternalNode{
			left:   r.root,
			right:  textRope.root,
			length: r.Length(),
			size:   r.Size(),
		},
		length: r.length + utf8.RuneCountInString(text),
		size:   r.size + len(text),
	}
}

// PrependStr prepends a string to the beginning of the rope.
// Uses optimized implementation that directly creates a node instead of Insert().
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) PrependStr(text string) *Rope {
	if r == nil {
		return New(text)
	}
	if text == "" {
		return r
	}
	if r.length == 0 {
		return New(text)
	}

	// Optimized: Create rope from text and prepend it directly
	// This is faster than Insert(0, text) which needs to traverse the tree
	textRope := New(text)

	return &Rope{
		root: &InternalNode{
			left:   textRope.root,
			right:  r.root,
			length: textRope.Length(),
			size:   textRope.Size(),
		},
		length: r.length + utf8.RuneCountInString(text),
		size:   r.size + len(text),
	}
}

// Append appends a string to the end of the rope.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) Append(text string) *Rope {
	return r.AppendStr(text)
}

// Prepend prepends a string to the beginning of the rope.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) Prepend(text string) *Rope {
	return r.PrependStr(text)
}

// ========== Builder Integration ==========

// AppendFromBuilder appends the contents of a builder to the rope.
func (r *Rope) AppendFromBuilder(b *RopeBuilder) *Rope {
	return r.AppendRope(b.Build())
}

// PrependFromBuilder prepends the contents of a builder to the rope.
func (r *Rope) PrependFromBuilder(b *RopeBuilder) *Rope {
	return r.PrependRope(b.Build())
}

// ========== Optimization Checks ==========

// CanAppendWithoutRebalance checks if appending would be efficient
// (i.e., won't require significant rebalancing).
func (r *Rope) CanAppendWithoutRebalancing(other *Rope) bool {
	if r == nil || other == nil {
		return true
	}
	// Simple heuristic: if the right side is small, appending is efficient
	// This could be made more sophisticated
	return r.Depth() <= 20
}

// CanPrependWithoutRebalance checks if prepending would be efficient.
func (r *Rope) CanPrependWithoutRebalancing(other *Rope) bool {
	if r == nil || other == nil {
		return true
	}
	// Similar heuristic for prepend
	return r.Depth() <= 20
}

// ========== Concatenation Operators ==========

// Add is an alias for AppendRope for convenience.
func (r *Rope) Add(other *Rope) *Rope {
	return r.AppendRope(other)
}

// Plus is an alias for AppendRope for convenience.
func (r *Rope) Plus(other *Rope) *Rope {
	return r.AppendRope(other)
}

// ========== Multi-Concatenation ==========

// AppendAll appends multiple ropes to this rope.
func (r *Rope) AppendAll(others ...*Rope) *Rope {
	result := r.Clone()
	for _, other := range others {
		if other != nil && other.Length() > 0 {
			result = result.AppendRope(other)
		}
	}
	return result
}

// PrependAll prepends multiple ropes to this rope.
func (r *Rope) PrependAll(others ...*Rope) *Rope {
	result := r.Clone()
	// Prepend in reverse order to maintain order
	for i := len(others) - 1; i >= 0; i-- {
		other := others[i]
		if other != nil && other.Length() > 0 {
			result = result.PrependRope(other)
		}
	}
	return result
}

// ConcatWithSeparator joins ropes with a separator rope.
func ConcatWithSeparator(ropes []*Rope, separator *Rope) *Rope {
	if len(ropes) == 0 {
		return Empty()
	}
	if len(ropes) == 1 {
		return ropes[0].Clone()
	}

	result := ropes[0].Clone()
	for i := 1; i < len(ropes); i++ {
		if separator != nil && separator.Length() > 0 {
			result = result.AppendRope(separator)
		}
		result = result.AppendRope(ropes[i])
	}

	return result
}
