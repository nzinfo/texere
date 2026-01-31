package rope

// Range represents a single selection range in the text.
//
// A range consists of an "anchor" and "head" position. The head is the part
// that moves when directly extending a selection. The anchor and head can be
// in any order, or even share the same position (for a cursor).
//
// Positions use gap indexing, meaning they represent the gaps between chars
// rather than the chars themselves. For example, position 1 is between the
// first and second character.
//
// Examples:
//   - anchor=0, head=3: [hel]lo (selects "hel")
//   - anchor=3, head=0: ]hel[lo (reverse selection of "hel")
//   - anchor=1, head=1: h[]ello (cursor at position 1)
//
// Ranges are inclusive on the left and exclusive on the right, regardless of
// anchor-head ordering.
type Range struct {
	Anchor int // The side that doesn't move when extending
	Head   int // The side that moves when extending
}

// NewRange creates a new Range with the given anchor and head positions.
func NewRange(anchor, head int) Range {
	return Range{
		Anchor: anchor,
		Head:   head,
	}
}

// Point creates a zero-width Range (cursor) at the given position.
func Point(pos int) Range {
	return Range{
		Anchor: pos,
		Head:   pos,
	}
}

// From returns the start of the range (minimum of anchor and head).
func (r Range) From() int {
	if r.Anchor < r.Head {
		return r.Anchor
	}
	return r.Head
}

// To returns the end of the range (maximum of anchor and head).
func (r Range) To() int {
	if r.Anchor > r.Head {
		return r.Anchor
	}
	return r.Head
}

// Len returns the length of the range.
func (r Range) Len() int {
	return r.To() - r.From()
}

// IsCursor returns true if the range is zero-width (a cursor).
func (r Range) IsCursor() bool {
	return r.Anchor == r.Head
}

// Contains returns true if pos is within the range.
func (r Range) Contains(pos int) bool {
	return pos >= r.From() && pos < r.To()
}

// Selection represents a collection of selection ranges.
// It always contains at least one range.
type Selection struct {
	ranges        []Range
	primaryIndex  int
}

// NewSelection creates a new Selection with a single Range.
func NewSelection(ranges ...Range) *Selection {
	if len(ranges) == 0 {
		// A selection must have at least one range
		ranges = []Range{Point(0)}
	}
	return &Selection{
		ranges:       ranges,
		primaryIndex: 0,
	}
}

// NewSelectionWithPrimary creates a new Selection with the specified primary index.
func NewSelectionWithPrimary(ranges []Range, primaryIndex int) *Selection {
	if len(ranges) == 0 {
		ranges = []Range{Point(0)}
	}
	if primaryIndex < 0 || primaryIndex >= len(ranges) {
		primaryIndex = 0
	}
	return &Selection{
		ranges:       ranges,
		primaryIndex: primaryIndex,
	}
}

// Primary returns the primary (active) selection range.
func (s *Selection) Primary() Range {
	if s.primaryIndex >= 0 && s.primaryIndex < len(s.ranges) {
		return s.ranges[s.primaryIndex]
	}
	if len(s.ranges) > 0 {
		return s.ranges[0]
	}
	return Point(0)
}

// PrimaryIndex returns the index of the primary selection.
func (s *Selection) PrimaryIndex() int {
	return s.primaryIndex
}

// Len returns the number of ranges in the selection.
func (s *Selection) Len() int {
	return len(s.ranges)
}

// Iter returns an iterator over the selection ranges.
func (s *Selection) Iter() []Range {
	return s.ranges
}

// Add adds a range to the selection.
func (s *Selection) Add(r Range) {
	s.ranges = append(s.ranges, r)
}

// SetPrimary sets the primary selection index.
func (s *Selection) SetPrimary(index int) {
	if index >= 0 && index < len(s.ranges) {
		s.primaryIndex = index
	}
}

// WithDirection creates a new Range with a specific direction.
// If forward is true, anchor < head; otherwise anchor > head.
func (r Range) WithDirection(forward bool) Range {
	if forward {
		if r.Anchor <= r.Head {
			return r
		}
		return Range{Anchor: r.Head, Head: r.Anchor}
	} else {
		if r.Anchor >= r.Head {
			return r
		}
		return Range{Anchor: r.Head, Head: r.Anchor}
	}
}

// Cursor returns the block cursor position for this range.
// By convention, the cursor is positioned one grapheme inward from the edge.
// For a forward selection (anchor < head), the cursor is at head - 1.
// For a reverse selection (anchor > head), the cursor is at head.
// For a cursor (anchor == head), the cursor is at that position.
func (r Range) Cursor() int {
	if r.IsCursor() {
		return r.Head
	}
	if r.Head > r.Anchor {
		// Forward selection: cursor is at the end
		return r.Head
	}
	// Reverse selection: cursor is at the start
	return r.Head
}

// IsForward returns true if the range is a forward selection (anchor <= head).
func (r Range) IsForward() bool {
	return r.Anchor <= r.Head
}

// IsBackward returns true if the range is a backward selection (anchor > head).
func (r Range) IsBackward() bool {
	return r.Anchor > r.Head
}

// Slice returns the range as a (from, to) tuple.
func (r Range) Slice() (int, int) {
	return r.From(), r.To()
}

// Map maps this range through a changeset with the given association.
// Returns the mapped range with the anchor and head positions transformed.
func (r Range) Map(cs *ChangeSet, assoc Assoc) Range {
	// Use a single mapper to map both positions
	mapper := NewPositionMapper(cs)
	mapper.AddPosition(r.Anchor, AssocBefore)
	mapper.AddPosition(r.Head, assoc)

	mapped := mapper.Map()
	if len(mapped) < 2 {
		return r
	}

	return Range{
		Anchor: mapped[0],
		Head:   mapped[1],
	}
}

// Merge merges this range with another, producing a range that covers both.
func (r Range) Merge(other Range) Range {
	from := r.From()
	to := r.To()

	if other.From() < from {
		from = other.From()
	}
	if other.To() > to {
		to = other.To()
	}

	return Range{Anchor: from, Head: to}
}

// Intersect returns the intersection of this range with another.
func (r Range) Intersect(other Range) Range {
	from := r.From()
	to := r.To()

	otherFrom := other.From()
	otherTo := other.To()

	if from < otherFrom {
		from = otherFrom
	}
	if to > otherTo {
		to = otherTo
	}

	if from >= to {
		return Point(from)
	}

	return Range{Anchor: from, Head: to}
}

// ContainsRange returns true if this range fully contains another range.
func (r Range) ContainsRange(other Range) bool {
	return r.From() <= other.From() && r.To() >= other.To()
}

// Overlaps returns true if this range overlaps with another.
func (r Range) Overlaps(other Range) bool {
	return r.From() < other.To() && r.To() > other.From()
}

// ========== Position Mapping Integration ==========

// MapPositions maps all cursor positions in the selection through a changeset.
// Returns a new selection with all positions mapped.
func (s *Selection) MapPositions(cs *ChangeSet) *Selection {
	if s == nil || len(s.ranges) == 0 {
		return s
	}

	positions := s.GetPositions()
	assocs := s.GetAssociations()

	mapped := MapPositionsOptimized(cs, positions, assocs)

	return s.FromPositions(mapped)
}

// GetPositions returns all cursor positions from the selection ranges.
// Uses the Cursor() method for each range, which represents the actual cursor position.
func (s *Selection) GetPositions() []int {
	positions := make([]int, len(s.ranges))
	for i, r := range s.ranges {
		positions[i] = r.Cursor()
	}
	return positions
}

// GetAssociations returns default associations for all positions.
// Currently returns AssocBefore for all positions.
func (s *Selection) GetAssociations() []Assoc {
	assocs := make([]Assoc, len(s.ranges))
	for i := range assocs {
		assocs[i] = AssocBefore
	}
	return assocs
}

// FromPositions creates a new selection from a slice of positions.
// Each position becomes a single-point cursor (Range with Anchor == Head).
// Preserves the primary index from the original selection.
func (s *Selection) FromPositions(positions []int) *Selection {
	if len(positions) == 0 {
		return NewSelection(Point(0))
	}

	newRanges := make([]Range, len(positions))
	for i, pos := range positions {
		newRanges[i] = Point(pos)
	}

	primaryIdx := s.primaryIndex
	if primaryIdx >= len(newRanges) {
		primaryIdx = 0
	}

	return &Selection{
		ranges:       newRanges,
		primaryIndex: primaryIdx,
	}
}

