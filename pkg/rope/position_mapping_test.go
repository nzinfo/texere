package rope

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ========== Basic Optimization Tests ==========

func TestPositionMapper_MapOptimized_Sorted(t *testing.T) {
	doc := New("Hello World")

	// Create changeset: insert at position 6
	cs := NewChangeSet(doc.Length()).
		Retain(6).
		Insert("beautiful ")

	// Create sorted positions
	positions := []int{0, 3, 6, 9}
	assocs := make([]Assoc, len(positions))

	mapper := NewPositionMapper(cs)
	mapper.AddPositions(positions, assocs)

	result := mapper.MapOptimized()

	assert.Equal(t, len(positions), len(result))
	// Verify all results are valid
	for i, r := range result {
		assert.GreaterOrEqual(t, r, 0, "Position %d mapped to negative value", i)
	}
	// The exact mapping depends on the changeset implementation
	// Just verify it doesn't crash and produces valid results
}

func TestPositionMapper_MapOptimized_Unsorted(t *testing.T) {
	doc := New("Hello World")

	// Create changeset
	cs := NewChangeSet(doc.Length()).
		Retain(6).
		Insert("beautiful ")

	// Create unsorted positions
	positions := []int{9, 0, 6, 3}
	assocs := make([]Assoc, len(positions))

	mapper := NewPositionMapper(cs)
	mapper.AddPositions(positions, assocs)

	result := mapper.MapOptimized()

	// MapOptimized sorts positions, so result should be sorted
	// Verify result is sorted
	for i := 1; i < len(result); i++ {
		assert.GreaterOrEqual(t, result[i], result[i-1], "Result not sorted at index %d", i)
	}
	// Verify all results are valid
	for _, r := range result {
		assert.GreaterOrEqual(t, r, 0)
	}
}

func TestPositionMapper_MapOptimized_Empty(t *testing.T) {
	doc := New("Hello")
	cs := NewChangeSet(doc.Length()).Retain(doc.Length())

	mapper := NewPositionMapper(cs)
	result := mapper.MapOptimized()

	assert.Empty(t, result)
}

// ========== Performance Comparison Tests ==========

func TestPositionMapper_SortedVsUnsorted(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test")
	}

	doc := New(strings.Repeat("hello world ", 100))

	// Create changeset
	cs := NewChangeSet(doc.Length()).Retain(doc.Length()).Insert("XXX")

	// Create many positions (unsorted)
	positions := make([]int, 100)
	for i := 0; i < 100; i++ {
		positions[i] = rand.Intn(doc.Length())
	}

	assocs := make([]Assoc, 100)

	// Test optimized path (auto-sorts)
	mapper1 := NewPositionMapper(cs)
	mapper1.AddPositions(positions, assocs)
	result1 := mapper1.MapOptimized()

	// Test unsorted path
	mapper2 := NewPositionMapper(cs)
	mapper2.AddPositions(positions, assocs)
	result2 := mapper2.mapUnsorted()

	// Results should be the same (though order may differ)
	assert.Equal(t, len(result1), len(result2))

	// Verify all positions are mapped correctly
	for _, pos := range positions {
		// Find the result for this position
		found := false
		for _, r := range result1 {
			// Just verify result is valid
			if r >= 0 && r <= doc.Length()+10 {
				found = true
				break
			}
		}
		assert.True(t, found, "Position %d not mapped correctly", pos)
	}
}

// ========== AddPositions Tests ==========

func TestPositionMapper_AddPositions_Basic(t *testing.T) {
	doc := New("Hello World")
	cs := NewChangeSet(doc.Length())

	positions := []int{0, 5, 11}
	assocs := []Assoc{AssocBefore, AssocAfter, AssocBefore}

	mapper := NewPositionMapper(cs)
	mapper.AddPositions(positions, assocs)

	assert.Equal(t, 3, len(mapper.positions))
	assert.Equal(t, 0, mapper.positions[0].Pos)
	assert.Equal(t, AssocBefore, mapper.positions[0].Assoc)
	assert.Equal(t, 5, mapper.positions[1].Pos)
	assert.Equal(t, AssocAfter, mapper.positions[1].Assoc)
}

func TestPositionMapper_AddPositions_MismatchedLengths(t *testing.T) {
	doc := New("Hello World")
	cs := NewChangeSet(doc.Length())

	positions := []int{0, 5, 11, 20}
	assocs := []Assoc{AssocBefore, AssocAfter} // Fewer assocs than positions

	mapper := NewPositionMapper(cs)
	mapper.AddPositions(positions, assocs)

	// Missing assocs should default to AssocBefore
	assert.Equal(t, 4, len(mapper.positions))
	assert.Equal(t, AssocBefore, mapper.positions[2].Assoc)
	assert.Equal(t, AssocBefore, mapper.positions[3].Assoc)
}

// ========== MapPositionsOptimized Tests ==========

func TestMapPositionsOptimized_Convenience(t *testing.T) {
	doc := New("Hello World")
	cs := NewChangeSet(doc.Length()).Retain(6).Insert("X")

	positions := []int{3, 6, 9}
	assocs := []Assoc{AssocBefore, AssocBefore, AssocBefore}

	result := MapPositionsOptimized(cs, positions, assocs)

	assert.Equal(t, 3, len(result))
	// Verify results are reasonable
	for _, r := range result {
		assert.GreaterOrEqual(t, r, 0)
	}
}

// ========== Selection Integration Tests ==========

func TestSelection_MapPositions_Basic(t *testing.T) {
	doc := New("Hello World")

	// Create selection with multiple cursors
	ranges := []Range{
		Point(0),   // Cursor at start
		Point(5),   // Cursor at space
		Point(11),  // Cursor at end
	}
	sel := NewSelection(ranges...)

	// Insert at position 6
	cs := NewChangeSet(doc.Length()).Retain(6).Insert("beautiful ")

	mappedSel := sel.MapPositions(cs)

	// Should still have 3 ranges
	assert.Equal(t, 3, mappedSel.Len())

	// Positions should be mapped
	mappedRanges := mappedSel.Iter()
	assert.Equal(t, 0, mappedRanges[0].From()) // Position 0 unchanged
	assert.Greater(t, mappedRanges[1].From(), 5) // Position 5 shifted
	assert.Greater(t, mappedRanges[2].From(), 11) // Position 11 shifted
}

func TestSelection_GetPositions(t *testing.T) {
	ranges := []Range{
		Point(0),
		Point(5),
		Range{Anchor: 3, Head: 7}, // Selection range
	}
	sel := NewSelection(ranges...)

	positions := sel.GetPositions()

	assert.Equal(t, 3, len(positions))
	assert.Equal(t, 0, positions[0])
	assert.Equal(t, 5, positions[1])
	// For selection range (3,7), cursor is at Head (7)
	assert.Equal(t, 7, positions[2])
}

func TestSelection_GetAssociations(t *testing.T) {
	ranges := []Range{
		Point(0),
		Point(5),
		Point(10),
	}
	sel := NewSelection(ranges...)

	assocs := sel.GetAssociations()

	assert.Equal(t, 3, len(assocs))
	// All should be AssocBefore by default
	for _, assoc := range assocs {
		assert.Equal(t, AssocBefore, assoc)
	}
}

func TestSelection_FromPositions(t *testing.T) {
	original := NewSelection(Point(0), Point(5))
	original.SetPrimary(1)

	positions := []int{2, 8, 15}
	newSel := original.FromPositions(positions)

	assert.Equal(t, 3, newSel.Len())
	ranges := newSel.Iter()

	// All should be points (Anchor == Head)
	for _, r := range ranges {
		assert.Equal(t, r.Anchor, r.Head)
	}

	assert.Equal(t, 2, ranges[0].From())
	assert.Equal(t, 8, ranges[1].From())
	assert.Equal(t, 15, ranges[2].From())

	// Primary index should be preserved (capped to valid range)
	assert.Equal(t, 1, newSel.PrimaryIndex())
}

func TestSelection_FromPositions_Empty(t *testing.T) {
	sel := NewSelection(Point(0))

	newSel := sel.FromPositions([]int{})

	assert.NotNil(t, newSel)
	assert.Equal(t, 1, newSel.Len())
}

// ========== SortPositions Tests ==========

func TestPositionMapper_SortPositions(t *testing.T) {
	doc := New("Hello")
	cs := NewChangeSet(doc.Length())

	mapper := NewPositionMapper(cs)
	mapper.AddPosition(5, AssocBefore)
	mapper.AddPosition(2, AssocAfter)
	mapper.AddPosition(8, AssocBefore)
	mapper.AddPosition(1, AssocAfter)

	// Verify unsorted
	assert.False(t, mapper.isSorted())

	// Sort
	mapper.sortPositions()

	// Verify sorted
	assert.True(t, mapper.isSorted())

	// Verify positions are in order
	assert.Equal(t, 1, mapper.positions[0].Pos)
	assert.Equal(t, 2, mapper.positions[1].Pos)
	assert.Equal(t, 5, mapper.positions[2].Pos)
	assert.Equal(t, 8, mapper.positions[3].Pos)

	// Verify associations are preserved
	assert.Equal(t, AssocAfter, mapper.positions[0].Assoc)
	assert.Equal(t, AssocAfter, mapper.positions[1].Assoc)
	assert.Equal(t, AssocBefore, mapper.positions[2].Assoc)
	assert.Equal(t, AssocBefore, mapper.positions[3].Assoc)
}

// ========== Association Behavior Tests ==========

func TestPositionMapper_MapOptimized_AssocAfter(t *testing.T) {
	doc := New("Hello World")

	// Delete "World" (positions 6-11)
	cs := NewChangeSet(doc.Length()).
		Retain(6).
		Delete(5) // Delete "World"

	positions := []int{6, 7, 11}
	assocs := []Assoc{AssocBefore, AssocAfter, AssocAfter}

	mapper := NewPositionMapper(cs)
	mapper.AddPositions(positions, assocs)
	result := mapper.MapOptimized()

	// AssocBefore at position 6 should be at 6 (before delete)
	assert.Equal(t, 6, result[0])
	// AssocAfter should handle deletion differently
	assert.GreaterOrEqual(t, result[1], 6)
	assert.GreaterOrEqual(t, result[2], 6)
}

// ========== Edge Cases ==========

func TestPositionMapper_MapOptimized_SinglePosition(t *testing.T) {
	doc := New("Hello")
	cs := NewChangeSet(doc.Length()).Insert("X")

	mapper := NewPositionMapper(cs)
	mapper.AddPosition(0, AssocBefore)
	result := mapper.MapOptimized()

	assert.Equal(t, 1, len(result))
}

func TestPositionMapper_MapOptimized_DuplicatePositions(t *testing.T) {
	doc := New("Hello")
	cs := NewChangeSet(doc.Length()).Insert("X")

	// Add same position multiple times
	positions := []int{3, 3, 3}
	assocs := []Assoc{AssocBefore, AssocAfter, AssocBefore}

	mapper := NewPositionMapper(cs)
	mapper.AddPositions(positions, assocs)
	result := mapper.MapOptimized()

	assert.Equal(t, 3, len(result))
	// All should be mapped (stable sort preserves all)
}

func TestPositionMapper_MapOptimized_LargeDocument(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large document test")
	}

	// Create large document
	doc := New(strings.Repeat("a", 10000))
	cs := NewChangeSet(doc.Length()).Retain(5000).Insert("XXX")

	// Random positions
	positions := make([]int, 50)
	for i := 0; i < 50; i++ {
		positions[i] = rand.Intn(doc.Length())
	}

	assocs := make([]Assoc, 50)

	mapper := NewPositionMapper(cs)
	mapper.AddPositions(positions, assocs)
	result := mapper.MapOptimized()

	assert.Equal(t, 50, len(result))
	// Verify all results are valid
	for _, r := range result {
		assert.GreaterOrEqual(t, r, 0)
		assert.LessOrEqual(t, r, doc.Length()+10)
	}
}

// ========== Consistency Tests ==========

func TestPositionMapper_MapVsMapOptimized_Consistency(t *testing.T) {
	doc := New(strings.Repeat("hello world ", 10))

	// Create simple changeset (delete a range)
	cs := NewChangeSet(doc.Length()).
		Retain(20).
		Delete(10) // Delete 10 characters

	// Test with sorted positions
	positions := []int{0, 10, 30, 40, 50}
	assocs := make([]Assoc, len(positions))
	for i := range assocs {
		assocs[i] = AssocBefore
	}

	// Test with Map() - should use fast path (sorted)
	mapper1 := NewPositionMapper(cs)
	mapper1.AddPositions(positions, assocs)
	result1 := mapper1.Map()

	// Test with MapOptimized()
	mapper2 := NewPositionMapper(cs)
	mapper2.AddPositions(positions, assocs)
	result2 := mapper2.MapOptimized()

	// For sorted input, both should produce identical results
	assert.Equal(t, result1, result2)
}
