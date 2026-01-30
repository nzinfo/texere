package rope

import (
	"time"
)

// Assoc represents cursor association behavior for operations.
// This determines how the cursor position should be adjusted after edits.
type Assoc int

const (
	// AssocBefore places cursor before the inserted/deleted text
	AssocBefore Assoc = iota

	// AssocAfter places cursor after the inserted/deleted text
	AssocAfter

	// AssocBeforeWord places cursor at the start of the word before the position
	AssocBeforeWord

	// AssocAfterWord places cursor at the start of the word after the position
	AssocAfterWord

	// AssocBeforeSticky keeps cursor at the same relative offset in exact-size replacements
	AssocBeforeSticky

	// AssocAfterSticky keeps cursor at the same relative offset in exact-size replacements
	AssocAfterSticky
)

// String returns the string representation of Assoc
func (a Assoc) String() string {
	switch a {
	case AssocBefore:
		return "Before"
	case AssocAfter:
		return "After"
	case AssocBeforeWord:
		return "BeforeWord"
	case AssocAfterWord:
		return "AfterWord"
	case AssocBeforeSticky:
		return "BeforeSticky"
	case AssocAfterSticky:
		return "AfterSticky"
	default:
		return "Unknown"
	}
}

// Position represents a position in the document with association information.
type Position struct {
	Pos    int    // Position in the document
	Assoc  Assoc  // How to adjust this position after edits
	Offset int    // Offset from the position (for sticky positioning)
}

// NewPosition creates a new position with the given position and association.
func NewPosition(pos int, assoc Assoc) *Position {
	return &Position{
		Pos:   pos,
		Assoc: assoc,
	}
}

// NewPositionWithOffset creates a new position with offset for sticky positioning.
func NewPositionWithOffset(pos int, assoc Assoc, offset int) *Position {
	return &Position{
		Pos:    pos,
		Assoc:  assoc,
		Offset: offset,
	}
}

// PositionMapper maps positions through a changeset.
// This handles cursor position updates after edits.
type PositionMapper struct {
	changeset     *ChangeSet
	positions     []*Position
	document      *Rope     // Optional: document for word boundary detection
	wordBoundary  *WordBoundary
}

// NewPositionMapper creates a new position mapper for the given changeset.
func NewPositionMapper(cs *ChangeSet) *PositionMapper {
	return &PositionMapper{
		changeset: cs,
		positions: make([]*Position, 0),
	}
}

// NewPositionMapperWithDoc creates a new position mapper with document for word boundaries.
func NewPositionMapperWithDoc(cs *ChangeSet, doc *Rope) *PositionMapper {
	return &PositionMapper{
		changeset:    cs,
		positions:    make([]*Position, 0),
		document:     doc,
		wordBoundary: NewWordBoundary(doc),
	}
}

// AddPosition adds a position to be mapped.
func (pm *PositionMapper) AddPosition(pos int, assoc Assoc) *PositionMapper {
	position := &Position{
		Pos:   pos,
		Assoc: assoc,
	}
	pm.positions = append(pm.positions, position)
	return pm
}

// AddPositionWithOffset adds a position with offset for sticky positioning.
func (pm *PositionMapper) AddPositionWithOffset(pos int, assoc Assoc, offset int) *PositionMapper {
	position := &Position{
		Pos:    pos,
		Assoc:  assoc,
		Offset: offset,
	}
	pm.positions = append(pm.positions, position)
	return pm
}

// Map maps all positions through the changeset and returns the new positions.
// This is optimized for sorted positions - O(N+M) where N is changeset length
// and M is number of positions. For unsorted positions, it falls back to O(M*N).
func (pm *PositionMapper) Map() []int {
	if len(pm.positions) == 0 {
		return []int{}
	}

	// Check if positions are already sorted
	sorted := pm.isSorted()

	if sorted {
		return pm.mapSorted()
	}

	return pm.mapUnsorted()
}

// isSorted checks if positions are sorted in ascending order.
func (pm *PositionMapper) isSorted() bool {
	for i := 1; i < len(pm.positions); i++ {
		if pm.positions[i].Pos < pm.positions[i-1].Pos {
			return false
		}
	}
	return true
}

// mapSorted maps positions in O(N+M) time using single pass.
func (pm *PositionMapper) mapSorted() []int {
	result := make([]int, len(pm.positions))

	// Process each position independently
	for i, position := range pm.positions {
		targetPos := position.Pos

		// Reset state for each position
		oldPos := 0
		newPos := 0

		// Debug: print initial state for this position
		// fmt.Printf("[Position %d] target=%d, oldPos=%d, newPos=%d\n", i, targetPos, oldPos, newPos)

		// Process operations until we reach or pass targetPos
		for _, op := range pm.changeset.operations {
			if oldPos >= targetPos {
				// Already reached or passed target
				// fmt.Printf("  Already at target, breaking\n")
				break
			}

			// fmt.Printf("  Processing op: Type=%v, Len=%d, Text=%q\n", op.OpType, op.Length, op.Text)

			switch op.OpType {
			case OpRetain:
				if oldPos+op.Length >= targetPos {
					// Target is within this retain
					advance := targetPos - oldPos
					oldPos += advance
					newPos += advance
					// fmt.Printf("    Partial retain: advance=%d, oldPos=%d, newPos=%d\n", advance, oldPos, newPos)
					// Don't break yet - we might have more positions to process
					// Actually, we should break since we've reached the target
					break
				} else {
					// Entire retain is before target
					oldPos += op.Length
					newPos += op.Length
					// fmt.Printf("    Full retain: oldPos=%d, newPos=%d\n", oldPos, newPos)
				}

			case OpDelete:
				if oldPos+op.Length >= targetPos {
					// Target is within this delete
					// Delete it, but don't advance oldPos past target
					oldPos = targetPos
					// fmt.Printf("    Partial delete: oldPos=%d\n", oldPos)
					break
				} else {
					// Entire delete is before target
					oldPos += op.Length
					// fmt.Printf("    Full delete: oldPos=%d\n", oldPos)
				}

			case OpInsert:
				// Inserted content affects newPos but not oldPos
				insertLen := len([]rune(op.Text))
				newPos += insertLen
				// fmt.Printf("    Insert: insertLen=%d, newPos=%d\n", insertLen, newPos)
			}
		}

		// If we ran out of operations but haven't reached targetPos,
		// the remaining characters are retained (no more changes)
		if oldPos < targetPos {
			remaining := targetPos - oldPos
			newPos += remaining
			oldPos += remaining
			// fmt.Printf("  Remaining: remaining=%d, oldPos=%d, newPos=%d\n", remaining, oldPos, newPos)
		}

		// Apply association behavior
		result[i] = pm.applyAssociation(position, targetPos, newPos, oldPos)
		// fmt.Printf("  Result[%d] = %d\n\n", i, result[i])
	}

	return result
}

// applyAssociation applies the association behavior to determine final position.
func (pm *PositionMapper) applyAssociation(position *Position, oldPos, newPos, currentPos int) int {
	switch position.Assoc {
	case AssocBefore:
		// Position is before the edit
		return newPos

	case AssocAfter:
		// Position is after the edit, may need to skip inserts/deletes
		return pm.applyAfterAssociation(oldPos, newPos, currentPos)

	case AssocBeforeWord:
		// Move to start of word before position
		if pm.wordBoundary != nil {
			return pm.wordBoundary.PrevWordStart(newPos)
		}
		return newPos

	case AssocAfterWord:
		// Move to start of word after position
		if pm.wordBoundary != nil {
			return pm.wordBoundary.NextWordStart(newPos)
		}
		return newPos

	case AssocBeforeSticky:
		// Keep relative offset in exact-size replacements
		return newPos + position.Offset

	case AssocAfterSticky:
		// Keep relative offset in exact-size replacements
		return newPos + position.Offset

	default:
		return newPos
	}
}

// applyAfterAssociation handles AssocAfter behavior.
func (pm *PositionMapper) applyAfterAssociation(oldPos, newPos, currentPos int) int {
	// If we're exactly at the position, stay after any inserts/deletes
	return newPos
}

// mapUnsorted maps positions in O(M*N) time.
func (pm *PositionMapper) mapUnsorted() []int {
	result := make([]int, len(pm.positions))

	for i, position := range pm.positions {
		result[i] = pm.mapSinglePosition(position)
	}

	return result
}

// mapSinglePosition maps a single position through the changeset.
func (pm *PositionMapper) mapSinglePosition(position *Position) int {
	pos := 0
	newPos := 0
	oldPos := position.Pos

	for _, op := range pm.changeset.operations {
		switch op.OpType {
		case OpRetain:
			if pos+op.Length >= oldPos {
				// Position is within this retain
				newPos += (oldPos - pos)
				return pm.applyAssociation(position, oldPos, newPos, oldPos)
			}
			pos += op.Length
			newPos += op.Length

		case OpDelete:
			if pos+op.Length >= oldPos {
				// Position is within deleted range
				// Apply association to determine where to place cursor
				return pm.applyAssociation(position, oldPos, newPos, pos)
			}
			pos += op.Length

		case OpInsert:
			if pos >= oldPos {
				// Already past the position
				return pm.applyAssociation(position, oldPos, newPos, pos)
			}
			newPos += len([]rune(op.Text))
		}

		if pos >= oldPos {
			break
		}
	}

	return newPos
}

// MapPositions is a convenience function to map positions through a changeset.
func MapPositions(cs *ChangeSet, positions []int, assoc Assoc) []int {
	mapper := NewPositionMapper(cs)
	for _, pos := range positions {
		mapper.AddPosition(pos, assoc)
	}
	return mapper.Map()
}

// UndoKind specifies how to navigate through history (steps or time).
type UndoKind int

const (
	// UndoSteps navigates by a specific number of steps
	UndoSteps UndoKind = iota

	// UndoTimePeriod navigates by a time duration
	UndoTimePeriod
)

// UndoRequest represents a request to navigate through history.
type UndoRequest struct {
	Kind     UndoKind
	Steps    int
	Duration time.Duration
}

// NewUndoSteps creates a request to undo a specific number of steps.
func NewUndoSteps(steps int) *UndoRequest {
	return &UndoRequest{
		Kind:  UndoSteps,
		Steps: steps,
	}
}

// NewUndoTimePeriod creates a request to undo to a specific time ago.
func NewUndoTimePeriod(duration time.Duration) *UndoRequest {
	return &UndoRequest{
		Kind:     UndoTimePeriod,
		Duration: duration,
	}
}

// EarlierRequest is an alias for NewUndoTimePeriod for backward compatibility.
func EarlierRequest(duration time.Duration) *UndoRequest {
	return NewUndoTimePeriod(duration)
}

// LaterRequest is an alias for NewUndoTimePeriod for redo.
func LaterRequest(duration time.Duration) *UndoRequest {
	return NewUndoTimePeriod(duration)
}
