# Texere-Rope Implementation Plan
**Priority:** High - Critical Features for Text Editor
**Timeline:** 2-3 weeks
**Date:** 2025-01-31

---

## Overview

This plan details the implementation of 4 critical features from ropey/helix that are missing in texere-rope:

1. **Grapheme Support** - Unicode-aware text operations
2. **Chunk_at Methods** - Low-level chunk access for performance
3. **Position Mapping Optimization** - Multi-cursor performance O(N+M)
4. **Time-based Undo** - Natural undo/redo navigation

---

## Feature 1: Grapheme Support (CRITICAL)

### Why This Matters

Grapheme clusters are essential for proper Unicode handling:
- Emoji like "👨‍👩‍👧‍👦" should be **1 character**, not 8 code points
- Cursor movement must respect grapheme boundaries
- Text selection should not split combining characters
- Deletion operations must be grapheme-aware

### Reference: Ropey Implementation

**Sources:**
- [ropey Graphemes API](https://docs.rs/ropey/latest/ropey/struct.Rope.html#method.graphemes)
- [Unicode UAX #29](https://unicode.org/reports/tr29/)

**Ropey API:**
```rust
pub fn graphemes(&self) -> Graphemes<'_>  // Iterator
pub fn len_graphemes(&self) -> usize
pub fn try_graphemes(&self) -> TryGraphemes<'_>

// Grapheme boundary queries
pub fn grapheme_at(&self, char_idx: usize) -> Grapheme
pub fn prev_grapheme_boundary(&self, char_idx: usize) -> usize
pub fn next_grapheme_boundary(&self, char_idx: usize) -> usize
```

### Implementation Plan

**File:** `pkg/rope/graphemes.go` (new)
**Test File:** `pkg/rope/grapheme_test.go` (new)

#### API Design

```go
package rope

import (
    "unicode/utf8"
    "golang.org/x/text/unicode/segment"
)

// GraphemeIterator iterates over grapheme clusters in a rope
type GraphemeIterator struct {
    rope       *Rope
    seg        *segment.GraphemeScanner
    currentPos int    // Byte position in rope
    exhausted  bool
}

// Grapheme represents a user-perceived character (grapheme cluster)
type Grapheme struct {
    Text      string  // The grapheme cluster text
    StartPos  int     // Character position in rope
    ByteLen   int     // Length in bytes
    CharLen   int     // Length in chars (code points)
}

// Graphemes returns an iterator over grapheme clusters
func (r *Rope) Graphemes() *GraphemeIterator

// LenGraphemes returns the total number of grapheme clusters
func (r *Rope) LenGraphemes() int

// GraphemeAt returns the grapheme at the given character position
func (r *Rope) GraphemeAt(charIdx int) Grapheme

// PrevGraphemeStart returns the character position of the start
// of the grapheme cluster containing the given position
func (r *Rope) PrevGraphemeStart(charIdx int) int

// NextGraphemeStart returns the character position of the start
// of the grapheme cluster after the given position
func (r *Rope) NextGraphemeStart(charIdx int) int

// IsGraphemeBoundary returns true if the given position is at
// a grapheme cluster boundary
func (r *Rope) IsGraphemeBoundary(charIdx int) bool

// GraphemeSlice returns a rope slice from start to end (in graphemes)
func (r *Rope) GraphemeSlice(start, end int) *Rope
```

#### Implementation Details

```go
// Graphemes creates a new grapheme iterator
func (r *Rope) Graphemes() *GraphemeIterator {
    if r == nil || r.Length() == 0 {
        return &GraphemeIterator{rope: r, exhausted: true}
    }

    content := r.String()

    return &GraphemeIterator{
        rope:      r,
        seg:       segment.NewGraphemeScannerString(content),
        currentPos: 0,
        exhausted:  false,
    }
}

// Next advances to the next grapheme cluster
func (it *GraphemeIterator) Next() bool {
    if it.exhausted {
        return false
    }

    it.exhausted = !it.seg.Next()
    return !it.exhausted
}

// Current returns the current grapheme cluster
func (it *GraphemeIterator) Current() Grapheme {
    if it.exhausted {
        return Grapheme{}
    }

    start, end := it.seg.Indices()
    text := it.seg.Str()

    // Calculate character position
    charPos := utf8.RuneCountInString(it.rope.String()[:start])

    return Grapheme{
        Text:     text,
        StartPos: charPos,
        ByteLen:  end - start,
        CharLen:  utf8.RuneCountInString(text),
    }
}

// Position returns the character position of the current grapheme
func (it *GraphemeIterator) Position() int {
    if it.exhausted {
        return it.rope.LenGraphemes()
    }
    return it.Current().StartPos
}
```

### Test Cases to Migrate

Based on ropey's grapheme tests and Unicode standards:

```go
// pkg/rope/grapheme_test.go

package rope

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

// TestGrapheme_Empty tests empty rope
func TestGrapheme_Empty(t *testing.T) {
    r := New("")
    assert.Equal(t, 0, r.LenGraphemes())

    it := r.Graphemes()
    count := 0
    for it.Next() {
        count++
    }
    assert.Equal(t, 0, count)
}

// TestGrapheme_ASCII tests ASCII text (1:1 mapping)
func TestGrapheme_ASCII(t *testing.T) {
    r := New("hello")
    assert.Equal(t, 5, r.LenGraphemes())

    graphemes := []string{}
    it := r.Graphemes()
    for it.Next() {
        graphemes = append(graphemes, it.Current().Text)
    }

    assert.Equal(t, []string{"h", "e", "l", "l", "o"}, graphemes)
}

// TestGrapheme_CombiningDiacritics tests combining marks
func TestGrapheme_CombiningDiacritics(t *testing.T) {
    // é can be 1 code point (U+00E9) or 2 (e + combining acute)
    // Both should be 1 grapheme

    // Single code point
    r1 := New("é")  // U+00E9
    assert.Equal(t, 1, r1.LenGraphemes())

    // Combining characters
    r2 := New("e\u0301")  // e + combining acute
    assert.Equal(t, 1, r2.LenGraphemes())

    // Mixed
    r3 := New("cafe\u0301")  // cafe + combining acute
    assert.Equal(t, 4, r3.LenGraphemes())  // c,a,f,é
}

// TestGrapheme_Emoji tests emoji (multi-codepoint)
func TestGrapheme_Emoji(t *testing.T) {
    // Family emoji: 👨‍👩‍👧‍👦 (man + ZWJ + woman + ZWJ + girl + ZWJ + boy)
    r := New("👨‍👩‍👧‍👦")
    assert.Equal(t, 1, r.LenGraphemes())

    it := r.Graphemes()
    it.Next()
    assert.Equal(t, "👨‍👩‍👧‍👦", it.Current().Text)
}

// TestGrapheme_EmojiWithSkinTone tests emoji with skin tone modifier
func TestGrapheme_EmojiWithSkinTone(t *testing.T) {
    // 👋 + skin tone modifier (U+1F3FB)
    r := New("👋🏻")
    assert.Equal(t, 1, r.LenGraphemes())
}

// TestGrapheme_RegionalIndicator tests flag emojis
func TestGrapheme_RegionalIndicator(t *testing.T) {
    // 🇺🇸 = 🇺 (U+1F1FA) + 🇸 (U+1F1F8)
    r := New("🇺🇸")
    assert.Equal(t, 1, r.LenGraphemes())
}

// TestGrapheme_Keycaps tests keycap emojis
func TestGrapheme_Keycaps(t *testing.T) {
    // 🔟 (keycap 10) = 🔞 + zero-width join + ️⃣
    r := New("🔟")
    assert.Equal(t, 1, r.LenGraphemes())
}

// TestGrapheme_PrevBoundary tests backward navigation
func TestGrapheme_PrevBoundary(t *testing.T) {
    r := New("cafe\u0301")  // c,a,f,é (4 graphemes)

    assert.Equal(t, 0, r.PrevGraphemeStart(0))
    assert.Equal(t, 0, r.PrevGraphemeStart(1))
    assert.Equal(t, 0, r.PrevGraphemeStart(2))
    assert.Equal(t, 3, r.PrevGraphemeStart(3))  // Start of é
    assert.Equal(t, 3, r.PrevGraphemeStart(4))  // Start of last grapheme
}

// TestGrapheme_NextBoundary tests forward navigation
func TestGrapheme_NextBoundary(t *testing.T) {
    r := New("cafe\u0301")  // c,a,f,é (4 graphemes)

    assert.Equal(t, 1, r.NextGraphemeStart(0))
    assert.Equal(t, 2, r.NextGraphemeStart(1))
    assert.Equal(t, 3, r.NextGraphemeStart(2))
    assert.Equal(t, 4, r.NextGraphemeStart(3))  // End
}

// TestGrapheme_At tests GraphemeAt method
func TestGrapheme_At(t *testing.T) {
    r := New("👨‍👩‍👧‍👦 cafe")

    g0 := r.GraphemeAt(0)
    assert.Equal(t, "👨‍👩‍👧‍👦", g0.Text)
    assert.Equal(t, 0, g0.StartPos)

    g1 := r.GraphemeAt(1)
    assert.Equal(t, " ", g1.Text)
    assert.Equal(t, 1, g1.StartPos)

    g2 := r.GraphemeAt(2)
    assert.Equal(t, "c", g2.Text)
}

// TestGrapheme_IsBoundary tests boundary detection
func TestGrapheme_IsBoundary(t *testing.T) {
    r := New("écafé")  // 4 graphemes: é,c,a,é

    // Positions: 0=between, 1=after é, 2=after c, 3=after a, 4=after é
    assert.True(t, r.IsGraphemeBoundary(0))
    assert.True(t, r.IsGraphemeBoundary(1))
    assert.True(t, r.IsGraphemeBoundary(2))
    assert.True(t, r.IsGraphemeBoundary(3))
    assert.True(t, r.IsGraphemeBoundary(4))
}

// TestGrapheme_IterationRoundtrip tests consistent iteration
func TestGrapheme_IterationRoundtrip(t *testing.T) {
    text := "Hello 🌍 café 👨‍👩‍👧‍👦"
    r := New(text)

    // Collect graphemes
    var graphemes []string
    it := r.Graphemes()
    for it.Next() {
        graphemes = append(graphemes, it.Current().Text)
    }

    // Reconstruct and verify
    reconstructed := ""
    for _, g := range graphemes {
        reconstructed += g
    }

    assert.Equal(t, text, reconstructed)
}

// TestGrapheme_Slice tests slicing by grapheme
func TestGrapheme_Slice(t *testing.T) {
    r := New("Hello World")

    // Get first 5 graphemes
    slice := r.GraphemeSlice(0, 5)
    assert.Equal(t, "Hello", slice.String())
}

// TestGrapheme_ComplexUnicode tests complex Unicode text
func TestGrapheme_ComplexUnicode(t *testing.T) {
    // Mix of ASCII, combining marks, emoji, ZWJ sequences
    text := "Hello 🌍\ncafé\n👨‍👩‍👧‍👦 family\n"
    r := New(text)

    count := r.LenGraphemes()
    it := r.Graphemes()

    actualCount := 0
    for it.Next() {
        actualCount++
    }

    assert.Equal(t, count, actualCount)
}

// TestGrapheme_Performance tests performance on large text
func TestGrapheme_Performance(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test")
    }

    // Large text with many graphemes
    text := strings.Repeat("Hello World! ", 1000)
    r := New(text)

    // Should be fast
    b := testing.Benchmark(func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            r.LenGraphemes()
        }
    })

    // Verify correctness
    assert.Equal(t, 12000, r.LenGraphemes())  // 13 * 1000 - 1 (last space missing)
}

// TestGrapheme_EdgeCases tests edge cases
func TestGrapheme_EdgeCases(t *testing.T) {
    // Empty string
    r := New("")
    assert.Equal(t, 0, r.LenGraphemes())

    // Single character
    r = New("a")
    assert.Equal(t, 1, r.LenGraphemes())

    // Single emoji
    r = New("🌍")
    assert.Equal(t, 1, r.LenGraphemes())

    // Only combining marks
    r = New("\u0301\u0302")
    assert.Equal(t, 2, r.LenGraphemes())  // Each is separate
}

// TestGrapheme_CursorMovement tests cursor-like operations
func TestGrapheme_CursorMovement(t *testing.T) {
    r := New("cafe\u0301")

    // Simulate cursor movement by grapheme
    pos := 0
    positions := []int{pos}

    // Move right by graphemes
    for i := 0; i < r.LenGraphemes(); i++ {
        pos = r.NextGraphemeStart(pos)
        positions = append(positions, pos)
    }

    expected := []int{0, 1, 2, 3, 4}
    assert.Equal(t, expected, positions)
}
```

### Dependencies

```go
import (
    "golang.org/x/text/unicode/segment"  // Grapheme segmentation
)
```

### Implementation Estimate

- **Coding:** 3-4 days
- **Testing:** 1 day
- **Documentation:** 0.5 day
- **Total:** 4.5-5.5 days

### Risk Assessment

- **Complexity:** Medium (Unicode is complex)
- **Risk:** Low-Medium (well-defined problem, standard libraries available)
- **Mitigation:** Use `golang.org/x/text/segment` (battle-tested)

---

## Feature 2: Chunk_at Methods

### Why This Matters

Direct chunk access enables:
- High-performance custom operations
- Zero-copy string manipulation
- Building specialized iterators
- Advanced text processing algorithms

### Reference: Ropey Implementation

**Sources:**
- [ropey Low-level APIs](https://docs.rs/ropey/latest/ropey/#low-level-apis)
- [ropey Chunk Methods](https://docs.rs/ropey/latest/ropey/struct.Rope.html#method.chunk_at_char)

**Ropey API:**
```rust
pub fn chunk_at_char(&self, char_idx: usize) -> (&str, usize, usize, usize)
pub fn chunk_at_byte(&self, byte_idx: usize) -> (&str, usize, usize, usize)
pub fn chunk_at_line(&self, line_idx: usize) -> (&str, usize, usize, usize)

// Returns: (chunk_str, chunk_byte_idx, chunk_char_idx, chunk_line_idx)
```

### Implementation Plan

**File:** `pkg/rope/chunk_at.go` (extend existing)
**Test File:** `pkg/rope/chunk_at_test.go` (new)

#### API Design

```go
// ChunkInfo contains information about a chunk
type ChunkInfo struct {
    Text       string  // The chunk's text content
    ByteIndex  int     // Starting byte index in rope
    CharIndex  int     // Starting character index in rope
    LineIndex  int     // Starting line index in rope
    NodeDepth  int     // Depth of node containing this chunk
}

// ChunkAtChar returns the chunk containing the given character position
// Returns ChunkInfo and error if position is out of bounds
func (r *Rope) ChunkAtChar(charIdx int) (ChunkInfo, error)

// ChunkAtByte returns the chunk containing the given byte position
func (r *Rope) ChunkAtByte(byteIdx int) (ChunkInfo, error)

// ChunkAtLine returns the chunk containing the given line
func (r *Rope) ChunkAtLine(lineIdx int) (ChunkInfo, error)

// ChunkAtPosition returns chunk by NodePosition (from ChunksIterator)
func (r *Rope) ChunkAtPosition(pos NodePosition) ChunkInfo
```

#### Implementation Details

```go
// ChunkAtChar finds and returns the chunk containing charIdx
func (r *Rope) ChunkAtChar(charIdx int) (ChunkInfo, error) {
    if charIdx < 0 || charIdx >= r.length {
        return ChunkInfo{}, fmt.Errorf("character index %d out of bounds", charIdx)
    }

    // Traverse tree to find the leaf containing charIdx
    byteIdx := r.bytePosForCharPos(0, charIdx, r.root)

    return r.chunkAtByte(byteIdx)
}

// ChunkAtByte finds and returns the chunk containing byteIdx
func (r *Rope) ChunkAtByte(byteIdx int) (ChunkInfo, error) {
    if byteIdx < 0 || byteIdx >= r.size {
        return ChunkInfo{}, fmt.Errorf("byte index %d out of bounds", byteIdx)
    }

    // Traverse tree to find leaf
    node, depth, offset := r.findNodeAtByte(byteIdx)

    // Calculate accumulated indices
    byteIdx := offset
    charIdx := r.charPosForBytePos(0, byteIdx, node)
    lineIdx := r.linePosForBytePos(0, byteIdx, node)

    return ChunkInfo{
        Text:      node.Text(),
        ByteIndex: byteIdx,
        CharIndex: charIdx,
        LineIndex: lineIdx,
        NodeDepth: depth,
    }, nil
}

// Helper: findNodeAtByte traverses tree to find node at byte position
func (r *Rope) findNodeAtByte(byteIdx int) (LeafNode, int, int) {
    node := r.root
    depth := 0
    offset := 0

    for {
        switch n := node.(type) {
        case *LeafNode:
            return n, depth, offset
        case *InternalNode:
            leftSize := n.Left.Size()
            if byteIdx < offset+leftSize {
                // In left subtree
                node = n.Left
                depth++
            } else if byteIdx < offset+leftSize+n.Right.Size() {
                // In right subtree
                offset += leftSize
                node = n.Right
                depth++
            } else {
                // Past this node
                return LeafNode{}, depth, offset
            }
        }
    }
}
```

### Test Cases to Migrate

Based on ropey's chunk tests:

```go
// pkg/rope/chunk_at_test.go

package rope

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

// TestChunkAtChar_SingleChunk tests single chunk rope
func TestChunkAtChar_SingleChunk(t *testing.T) {
    r := New("Hello World")

    chunk, err := r.ChunkAtChar(5)
    assert.NoError(t, err)
    assert.Equal(t, "Hello World", chunk.Text)
    assert.Equal(t, 0, chunk.ByteIndex)
    assert.Equal(t, 0, chunk.CharIndex)
    assert.Equal(t, 0, chunk.LineIndex)
}

// TestChunkAtChar_MultiChunk tests multi-chunk rope
func TestChunkAtChar_MultiChunk(t *testing.T) {
    r := New("Hello")
    r = r.Insert(5, " ")
    r = r.Insert(6, "World")

    // Now we have at least 2 chunks
    chunk1, _ := r.ChunkAtChar(0)
    assert.Equal(t, "Hello ", chunk1.Text)

    chunk2, _ := r.ChunkAtChar(6)
    assert.Equal(t, "World", chunk2.Text)
}

// TestChunkAtChar_OutOfBounds tests error handling
func TestChunkAtChar_OutOfBounds(t *testing.T) {
    r := New("Hello")

    _, err := r.ChunkAtChar(-1)
    assert.Error(t, err)

    _, err = r.ChunkAtChar(5)
    assert.Error(t, err)
}

// TestChunkAtChar_Positions tests correct position tracking
func TestChunkAtChar_Positions(t *testing.T) {
    // Create multi-chunk rope
    var r *Rope
    for i := 0; i < 100; i++ {
        r = r.Insert(r.Length(), "word ")
    }

    // Test position at various points
    for i := 0; i < r.Length(); i += 10 {
        chunk, err := r.ChunkAtChar(i)
        assert.NoError(t, err)

        // Verify position is within chunk
        assert.True(t, chunk.CharIndex <= i)
        assert.True(t, i < chunk.CharIndex + utf8.RuneCountInString(chunk.Text))
    }
}

// TestChunkAtByte_Basics tests byte-based chunk access
func TestChunkAtByte_Basics(t *testing.T) {
    r := New("Hello World")

    chunk, err := r.ChunkAtByte(0)
    assert.NoError(t, err)
    assert.Equal(t, "Hello World", chunk.Text)
    assert.Equal(t, 0, chunk.ByteIndex)
}

// TestChunkAtByte_Unicode tests Unicode chunk boundaries
func TestChunkAtByte_Unicode(t *testing.T) {
    r := New("Hello 世界")

    chunk, err := r.ChunkAtByte(0)
    assert.NoError(t, err)

    // Verify chunk contains both ASCII and Unicode
    assert.Contains(t, chunk.Text, "Hello")
    assert.Contains(t, chunk.Text, "世界")
}

// TestChunkAtLine_LineBreaks tests line-based chunk access
func TestChunkAtLine_LineBreaks(t *testing.T) {
    r := New("Line 1\nLine 2\nLine 3")

    // Line 0 is in first chunk
    chunk1, _ := r.ChunkAtLine(0)
    assert.Contains(t, chunk1.Text, "Line 1")

    // Line 1 might be in same or different chunk
    chunk2, _ := r.ChunkAtLine(1)
    assert.Contains(t, chunk2.Text, "Line 2")
}

// TestChunkAtChar_Performance tests performance on deep tree
func TestChunkAtChar_Performance(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test")
    }

    // Create deep tree with many chunks
    var r *Rope
    for i := 0; i < 1000; i++ {
        r = r.Insert(r.Length(), "x")
    }

    b := testing.Benchmark(func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            r.ChunkAtChar(r.Length() / 2)
        }
    })

    // Should be very fast (< 1μs per call)
    assert.True(t, b.Elapsed() < 100*time.Millisecond)
}

// TestChunkAtChar_Consistency tests consistency with iterator
func TestChunkAtChar_Consistency(t *testing.T) {
    r := New("Hello World\nThis is a test")

    it := r.Chunks()
    chunks := []string{}
    for it.Next() {
        chunks = append(chunks, it.Current())
    }

    // Verify ChunkAtChar finds same chunks
    for i, expectedChunk := range chunks {
        charIdx := r.charPosForBytePos(0, len(expectedChunk)-1, r.root)
        if charIdx < r.Length() {
            chunk, _ := r.ChunkAtChar(charIdx)
            assert.Equal(t, expectedChunk, chunk.Text)
        }
    }
}

// TestChunkAtByte_AccumulatedIndices tests index calculation
func TestChunkAtByte_AccumulatedIndices(t *testing.T) {
    // Create rope with known structure
    r := New("Hello")
    r = r.Insert(5, " ")
    r = r.Insert(6, "World")
    r = r.Insert(11, "\n")
    r = r.Insert(12, "Second line")

    // Test at specific positions
    tests := []struct {
        byteIdx    int
        expectText string
        expectChar int
        expectLine int
    }{
        {0, "Hello ", 0, 0},
        {7, "World\n", 6, 0},
        {13, "Second line", 12, 1},
    }

    for _, tt := range tests {
        chunk, err := r.ChunkAtByte(tt.byteIdx)
        assert.NoError(t, err)
        assert.Equal(t, tt.expectText, chunk.Text)
        assert.Equal(t, tt.expectChar, chunk.CharIndex)
        assert.Equal(t, tt.expectLine, chunk.LineIndex)
    }
}

// TestChunkInfo_NodeDepth tests depth tracking
func TestChunkInfo_NodeDepth(t *testing.T) {
    r := New("A")

    // Build a deeper tree
    for i := 0; i < 10; i++ {
        r = r.Insert(r.Length(), "B")
    }

    chunk, _ := r.ChunkAtChar(0)
    assert.True(t, chunk.NodeDepth >= 1)  // Should have some depth

    chunk, _ = r.ChunkAtChar(r.Length() - 1)
    assert.True(t, chunk.NodeDepth >= 1)
}
```

### Implementation Estimate

- **Coding:** 2-3 days
- **Testing:** 1 day
- **Documentation:** 0.5 day
- **Total:** 3.5-4.5 days

### Risk Assessment

- **Complexity:** Low (tree traversal is well-understood)
- **Risk:** Low (straightforward implementation)
- **Mitigation:** Leverage existing tree traversal code

---

## Feature 3: Position Mapping Optimization

### Why This Matters

For multi-cursor editing with N cursors and M operations:
- **Current:** O(M*N) when positions are unsorted
- **Optimized:** O(M log M + N + M) = O(M log M + N)

For 100 cursors and 1000 operations:
- Unsorted: 100 * 1000 = 100,000 operations
- Sorted: 100 * log(100) + 1000 = ~6,645 operations
- **Speedup: ~15x**

### Current Implementation

```go
// pkg/rope/transaction_advanced.go

func (pm *PositionMapper) mapSorted() []int {
    // O(N+M) when positions are sorted
    ...
}

func (pm *PositionMapper) mapUnsorted() []int {
    // O(M*N) when positions are unsorted
    ...
}
```

**Problem:** Unsorted positions fall back to slow path

### Implementation Plan

**File:** `pkg/rope/transaction_advanced.go` (extend existing)

#### Optimization: Auto-sorting

```go
// MapOptimized always uses the fast path by auto-sorting
func (pm *PositionMapper) MapOptimized() []int {
    if len(pm.positions) == 0 {
        return []int{}
    }

    // Auto-sort if not already sorted
    if !pm.isSorted() {
        pm.sortPositions()
    }

    return pm.mapSorted()
}

// sortPositions sorts positions along with their associations
func (pm *PositionMapper) sortPositions() {
    // Use stable sort to maintain order of equal positions
    sort.SliceStable(pm.positions, func(i, j int) bool {
        return pm.positions[i].Pos < pm.positions[j].Pos
    })
}
```

#### Batch Position Creation

```go
// AddPositions adds multiple positions at once
func (pm *PositionMapper) AddPositions(positions []int, assocs []Assoc) *PositionMapper {
    for i, pos := range positions {
        assoc := AssocBefore
        if i < len(assocs) {
            assoc = assocs[i]
        }
        pm.positions = append(pm.positions, &Position{
            Pos:   pos,
            Assoc: assoc,
        })
    }
    return pm
}

// MapPositionsOptimized is a convenience function for batch position mapping
func MapPositionsOptimized(cs *ChangeSet, positions []int, assocs []Assoc) []int {
    mapper := NewPositionMapper(cs)
    mapper.AddPositions(positions, assocs)
    return mapper.MapOptimized()
}
```

#### Selection Integration

```go
// MapPositions maps all cursor positions in a selection
func (s *Selection) MapPositions(cs *ChangeSet) *Selection {
    positions := s.GetPositions()
    assocs := s.GetAssociations()

    mapped := MapPositionsOptimized(cs, positions, assocs)

    return s.FromPositions(mapped)
}

// GetPositions returns all cursor positions in the selection
func (s *Selection) GetPositions() []int {
    positions := make([]int, len(s.ranges))
    for i, r := range s.ranges {
        positions[i] = r.Head
    }
    return positions
}

// GetAssociations returns associations for all positions
func (s *Selection) GetAssociations() []Assoc {
    assocs := make([]Assoc, len(s.ranges))
    for i := range assocs {
        assocs[i] = AssocBefore
    }
    return assocs
}
```

### Test Cases

```go
// pkg/rope/position_mapping_test.go

package rope

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

// TestPositionMapper_SortedVsUnsorted tests performance difference
func TestPositionMapper_SortedVsUnsorted(t *testing.T) {
    doc := New(strings.Repeat("hello world ", 100))

    // Create changeset
    cs := NewChangeSet(doc.Length()).Retain(doc.Length()).Insert("XXX")

    // Create many positions (unsorted)
    positions := make([]int, 100)
    for i := 0; i < 100; i++ {
        positions[i] = rand.Intn(doc.Length())
    }

    // Test unsorted path
    mapper1 := NewPositionMapper(cs)
    for _, pos := range positions {
        mapper1.AddPosition(pos, AssocBefore)
    }
    result1 := mapper1.mapUnsorted()

    // Test optimized path
    mapper2 := NewPositionMapper(cs)
    mapper2.AddPositions(positions, make([]Assoc, 100))
    result2 := mapper2.MapOptimized()

    // Results should be the same (though possibly in different order)
    assert.ElementsMatch(t, result1, result2)
}

// TestPositionMapper_AddPositions tests batch position addition
func TestPositionMapper_AddPositions(t *testing.T) {
    doc := New("Hello World")
    cs := NewChangeSet(doc.Length()).Retain(6).Insert("beautiful ")

    positions := []int{0, 6, 15}
    assocs := []Assoc{AssocBefore, AssocBefore, AssocAfter}

    mapped := MapPositionsOptimized(cs, positions, assocs)

    expected := []int{0, 6, 22}  // After inserting "beautiful " (7 chars)
    assert.Equal(t, expected, mapped)
}

// TestPositionMapper_Performance tests optimization effectiveness
func TestPositionMapper_Performance(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test")
    }

    doc := New(strings.Repeat("a", 10000))

    // Insert in middle
    cs := NewChangeSet(doc.Length()).Retain(5000).Insert("XXX")

    // Many random positions
    positions := make([]int, 1000)
    for i := 0; i < 1000; i++ {
        positions[i] = rand.Intn(doc.Length())
    }

    // Benchmark unsorted
    mapper := NewPositionMapper(cs)
    mapper.AddPositions(positions, make([]Assoc, 1000))

    b := testing.Benchmark(func(b *testing.B) {
        b.Run("Unsorted", func(b *testing.B) {
            mapper.mapUnsorted()
        })

        b.Run("Optimized", func(b *testing.B) {
            mapper.MapOptimized()
        })
    })

    // Optimized should be significantly faster
    unsortedTime := b.Result("Unsorted").MemBytes()
    optimizedTime := b.Result("Optimized").MemBytes()

    assert.True(t, optimizedTime < unsortedTime)
}

// TestSelection_MapPositions tests selection-level position mapping
func TestSelection_MapPositions(t *testing.T) {
    doc := New("Hello World\nLine 2\nLine 3")

    // Create multi-cursor selection
    ranges := []Range{
        NewRange(0, 5),    // "Hello"
        NewRange(6, 11),   // "World"
        NewRange(12, 17),  // "Line 2"
    }
    sel := NewSelection(ranges...)

    // Insert at position 6
    cs := NewChangeSet(doc.Length()).Retain(6).Insert("beautiful ")

    mappedSel := sel.MapPositions(cs)

    // Selections should be shifted
    newRanges := mappedSel.Iter()
    assert.Equal(t, 3, len(newRanges))

    assert.Equal(t, 0, newRanges[0].From())
    assert.Equal(t, 5, newRanges[0].To())

    assert.Equal(t, 6, newRanges[1].From())
    assert.Equal(t, 11, newRanges[1].To())

    assert.Equal(t, 12, newRanges[2].From())
    assert.Equal(t, 17, newRanges[2].To())
}
```

### Implementation Estimate

- **Coding:** 2-3 days
- **Testing:** 1 day
- **Benchmarking:** 0.5 day
- **Documentation:** 0.5 day
- **Total:** 4-5 days

### Risk Assessment

- **Complexity:** Low-Medium (optimization, not new feature)
- **Risk:** Low (infrastructure exists)
- **Mitigation:** Extensive benchmarking to verify improvement

---

## Feature 4: Time-based Undo

### Why This Matters

Natural undo navigation:
- "Undo 5 minutes ago" instead of "Undo 47 operations"
- Better mental model for users
- Easier to navigate complex edit histories

### Current Status

✅ **Infrastructure exists:**
- `UndoRequest` with time duration
- `History.Earlier()` and `History.Later()` placeholders
- Timestamp tracking in `Transaction`

❌ **Missing:**
- Functional time navigation
- Duration parsing
- Integration with history

### Reference: Helix Implementation

**Sources:**
- [Helix Architecture - Transactions](https://helix-editor.vercel.app/contributing/architecture/)

### Implementation Plan

**File:** `pkg/rope/history.go` (extend existing)
**Test File:** `pkg/rope/history_time_test.go` (new)

#### API Enhancement

```go
// EarlierByDuration moves back in history by time duration
func (h *History) EarlierByDuration(duration time.Duration) *History {
    if h.IsEmpty() {
        return h
    }

    targetTime := time.Now().Add(-duration)

    // Walk back through history
    for i := h.currentIndex; i >= 0; i-- {
        if h.transactions[i].Timestamp.Before(targetTime) ||
           h.transactions[i].Timestamp.Equal(targetTime) {
            // Found state at or before target time
            return &History{
                root:        h.root,
                transactions: h.transactions[:i+1],
                currentIndex: i,
                pool:        h.pool,
            }
        }
    }

    // If not found, return root state
    return h.AtRoot()
}

// LaterByDuration moves forward in history by time duration
func (h *History) LaterByDuration(duration time.Duration) *History {
    if h.IsEmpty() {
        return h
    }

    targetTime := h.CurrentTransaction().Timestamp.Add(duration)

    // Walk forward through history
    for i := h.currentIndex + 1; i < len(h.transactions); i++ {
        if h.transactions[i].Timestamp.After(targetTime) ||
           h.transactions[i].Timestamp.Equal(targetTime) {
            // Found state at or after target time
            return &History{
                root:        h.root,
                transactions: h.transactions,
                currentIndex: i - 1,
                pool:        h.pool,
            }
        }
    }

    // If not found, return tip state
    return h.AtTip()
}

// EarlierBySteps moves back N steps (convenience wrapper)
func (h *History) EarlierBySteps(steps int) *History {
    return h.Earlier(NewUndoSteps(steps))
}

// LaterBySteps moves forward N steps (convenience wrapper)
func (h *History) LaterBySteps(steps int) *History {
    return h.Later(NewUndoSteps(steps))
}

// TimeAt returns the timestamp of the current state
func (h *History) TimeAt() time.Time {
    if h.IsEmpty() {
        return time.Time{}
    }
    return h.CurrentTransaction().Timestamp
}

// DurationFromRoot returns time elapsed since root
func (h *History) DurationFromRoot() time.Duration {
    if h.IsEmpty() {
        return 0
    }
    return time.Since(h.transactions[0].Timestamp)
}

// DurationToTip returns time from current to tip
func (h *History) DurationToTip() time.Duration {
    if h.IsEmpty() {
        return 0
    }
    tipIdx := len(h.transactions) - 1
    return h.transactions[tipIdx].Timestamp.Sub(h.CurrentTransaction().Timestamp)
}
```

#### Duration Parsing

```go
// pkg/rope/duration.go (new file)

package rope

import (
    "errors"
    "regexp"
    "strconv"
    "strings"
    "time"
)

var durationPattern = regexp.MustCompile(`(?i)^(\d+)\s*(s|sec|second|second|m|min|minute|minutes|h|hour|hours|d|day|days)?`)

// ParseDuration parses human-friendly duration strings
// Examples: "5m", "30min", "2 hours", "60s"
func ParseDuration(s string) (time.Duration, error) {
    s = strings.TrimSpace(s)

    match := durationPattern.FindStringSubmatch(s)
    if match == nil {
        return 0, errors.New("invalid duration format")
    }

    value, err := strconv.Atoi(match[1])
    if err != nil {
        return 0, err
    }

    unit := strings.ToLower(strings.TrimSpace(match[2]))
    if unit == "" || unit == "s" || unit == "sec" || unit == "second" || unit == "seconds" {
        return time.Duration(value) * time.Second, nil
    } else if unit == "m" || unit == "min" || unit == "minute" || unit == "minutes" {
        return time.Duration(value) * time.Minute, nil
    } else if unit == "h" || unit == "hour" || unit == "hours" {
        return time.Duration(value) * time.Hour, nil
    } else if unit == "d" || unit == "day" || unit == "days" {
        return time.Duration(value) * 24 * time.Hour, nil
    }

    return 0, errors.New("unknown time unit: " + unit)
}
```

### Test Cases

```go
// pkg/rope/history_time_test.go

package rope

import (
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
)

// TestHistory_EarlierByDuration tests time-based undo
func TestHistory_EarlierByDuration(t *testing.T) {
    doc := New("Hello")
    h := NewHistory(doc)

    // Create transactions with known timestamps
    baseTime := time.Now().Add(-10 * time.Minute)

    tx1 := NewTransaction(NewChangeSet(doc.Length()))
    tx1.timestamp = baseTime
    h = h.Append(tx1)

    doc = tx1.Apply(doc)
    tx2 := NewTransaction(NewChangeSet(doc.Length()).Insert(" World")
    tx2.timestamp = baseTime.Add(1 * time.Minute)
    h = h.Append(tx2)

    doc = tx2.Apply(doc)
    tx3 := NewTransaction(NewChangeSet(doc.Length()).Insert("!")
    tx3.timestamp = baseTime.Add(2 * time.Minute)
    h = h.Append(tx3)

    // Undo to 5 minutes ago (should undo tx3)
    h5m := h.EarlierByDuration(5 * time.Minute)
    assert.Equal(t, 2, h5m.currentIndex)  // At tx2
    assert.Equal(t, "Hello World", h5m.CurrentDocument().String())
}

// TestHistory_LaterByDuration tests time-based redo
func TestHistory_LaterByDuration(t *testing.T) {
    doc := New("Hello")
    h := NewHistory(doc)

    baseTime := time.Now().Add(-10 * time.Minute)

    // Create 3 transactions
    for i := 0; i < 3; i++ {
        tx := NewTransaction(NewChangeSet(doc.Length()).Insert("X")
        tx.timestamp = baseTime.Add(time.Duration(i) * time.Minute)
        h = h.Append(tx)
        doc = tx.Apply(doc)
    }

    // Undo all
    h = h.EarlierByDuration(100 * time.Hour)
    assert.Equal(t, "Hello", h.CurrentDocument().String())

    // Redo 1 minute
    h1m := h.LaterByDuration(1 * time.Minute)
    assert.Equal(t, "HelloX", h1m.CurrentDocument().String())

    // Redo to end
    hEnd := h.LaterByDuration(100 * time.Hour)
    assert.Equal(t, "HelloXXX", hEnd.CurrentDocument().String())
}

// TestParseDuration tests duration parsing
func TestParseDuration(t *testing.T) {
    tests := []struct {
        input    string
        expected time.Duration
    }{
        {"5s", 5 * time.Second},
        {"30 sec", 30 * time.Second},
        {"10min", 10 * time.Minute},
        {"2 hours", 2 * time.Hour},
        {"1d", 24 * time.Hour},
        {"60", 60 * time.Second},  // Default is seconds
    }

    for _, tt := range tests {
        d, err := ParseDuration(tt.input)
        assert.NoError(t, err)
        assert.Equal(t, tt.expected, d)
    }
}

// TestHistory_TimeAt tests timestamp queries
func TestHistory_TimeAt(t *testing.T) {
    doc := New("Hello")
    h := NewHistory(doc)

    now := time.Now()
    tx := NewTransaction(NewChangeSet(doc.Length()).Insert(" World"))
    tx.timestamp = now
    h = h.Append(tx)

    assert.Equal(t, now, h.TimeAt())
}

// TestHistory_DurationQueries tests duration helpers
func TestHistory_DurationQueries(t *testing.T) {
    doc := New("Hello")
    h := NewHistory(doc)

    baseTime := time.Now().Add(-10 * time.Minute)

    for i := 0; i < 3; i++ {
        tx := NewTransaction(NewChangeSet(doc.Length()).Insert("X")
        tx.timestamp = baseTime.Add(time.Duration(i) * time.Minute)
        h = h.Append(tx)
        doc = tx.Apply(doc)
    }

    // From root to tip
    assert.Equal(t, 2*time.Minute, h.DurationFromRoot())
    assert.Equal(t, 0, h.DurationToTip())  // At tip
}

// TestHistory_RoundTrip tests undo/redo roundtrip with time
func TestHistory_RoundTrip(t *testing.T) {
    doc := New("Original")
    h := NewHistory(doc)

    // Create history
    for i := 0; i < 5; i++ {
        tx := NewTransaction(NewChangeSet(doc.Length()).Insert("X")
        h = h.Append(tx)
        doc = tx.Apply(doc)
    }

    // Undo 3 minutes
    h1 := h.EarlierByDuration(3 * time.Minute)

    // Redo 3 minutes
    h2 := h1.LaterByDuration(3 * time.Minute)

    // Should be back at same state
    assert.Equal(t, h.CurrentDocument().String(), h2.CurrentDocument().String())
}

// TestHistory_EdgeCases tests edge cases
func TestHistory_EdgeCases(t *testing.T) {
    doc := New("Hello")
    h := NewHistory(doc)

    // Earlier by zero duration
    h0 := h.EarlierByDuration(0)
    assert.Equal(t, h.currentIndex, h0.currentIndex)

    // Later by zero duration
    h1 := h.LaterByDuration(0)
    assert.Equal(t, h.currentIndex, h1.currentIndex)

    // Earlier by very large duration
    hBig := h.EarlierByDuration(1000 * time.Hour)
    assert.Equal(t, 0, hBig.currentIndex)  // At root

    // Later by very large duration
    hBig2 := h.LaterByDuration(1000 * time.Hour)
    assert.Equal(t, len(h.transactions)-1, hBig2.currentIndex)  // At tip

    // Empty history
    hEmpty := NewHistory(nil)
    hEmpty2 := hEmpty.EarlierByDuration(1 * time.Minute)
    assert.Equal(t, hEmpty, hEmpty2)
}
```

### Implementation Estimate

- **Coding:** 2 days
- **Testing:** 1 day
- **Documentation:** 0.5 day
- **Total:** 3.5 days

### Risk Assessment

- **Complexity:** Low (infrastructure exists)
- **Risk:** Low
- **Mitigation:** Leverage existing history infrastructure

---

## Summary

### Total Timeline

| Feature | Estimate | Priority |
|---------|----------|----------|
| Grapheme Support | 5 days | HIGH |
| Chunk_at Methods | 4 days | HIGH |
| Position Mapping | 4 days | HIGH |
| Time-based Undo | 3.5 days | MEDIUM |
| **Total** | **16.5 days (~3.5 weeks)** | |

### Implementation Order

1. **Week 1:** Grapheme Support (foundation for Unicode)
2. **Week 2:** Chunk_at Methods + Position Mapping (performance)
3. **Week 3:** Time-based Undo (UX improvements)

### Success Criteria

For each feature:
- ✅ All unit tests pass
- ✅ Property tests pass
- ✅ Performance benchmarks meet expectations
- ✅ Documentation complete
- ✅ Backward compatibility maintained

### Next Steps

1. Review and approve this plan
2. Set up feature branch
3. Begin with Feature 1: Grapheme Support
4. Implement tests first (TDD approach)
5. Regular integration testing

---

**Prepared by:** Claude (AI Assistant)
**Date:** 2025-01-31
**Version:** 1.0
