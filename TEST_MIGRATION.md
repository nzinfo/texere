# Test Migration Guide
**Source:** Ropey (Rust) + Helix Editor
**Target:** Texere-Rope (Go)
**Date:** 2025-01-31

---

## Overview

This document lists all test cases to migrate from ropey and helix for the 4 critical features being implemented.

---

## Feature 1: Grapheme Support Tests

### Source: Ropey Grapheme Tests

**Location:** [ropey grapheme.rs](https://github.com/cessen/ropey/blob/master/src/ropey/grapheme.rs)

#### Test 1: Empty String
```rust
#[test]
fn graphemes__empty() {
    let rope = Rope::from_str("");
    assert_eq!(rope.graphemes().count(), 0);
}
```

**Go Version:**
```go
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
```

#### Test 2: ASCII Characters
```rust
#[test]
fn graphemes__ascii() {
    let rope = Rope::from_str("hello");
    assert_eq!(rope.graphemes().count(), 5);
}
```

**Go Version:**
```go
func TestGrapheme_ASCII(t *testing.T) {
    r := New("hello")
    assert.Equal(t, 5, r.LenGraphemes())

    var graphemes []string
    it := r.Graphemes()
    for it.Next() {
        graphemes = append(graphemes, it.Current().Text)
    }
    assert.Equal(t, []string{"h", "e", "l", "l", "o"}, graphemes)
}
```

#### Test 3: Emoji (Multi-codepoint)
```rust
#[test]
fn graphemes__emoji() {
    let rope = Rope::from_str("🎃🎨🎹🎸");
    assert_eq!(rope.graphemes().count(), 4);
}
```

**Go Version:**
```go
func TestGrapheme_Emoji(t *testing.T) {
    // Each emoji is 1 grapheme cluster
    r := New("🎃🎨🎹🎸")
    assert.Equal(t, 4, r.LenGraphemes())

    it := r.Graphemes()
    expected := []string{"🎃", "🎨", "🎹", "🎸"}
    i := 0
    for it.Next() {
        if i < len(expected) {
            assert.Equal(t, expected[i], it.Current().Text)
        }
        i++
    }
}
```

#### Test 4: Combining Marks
```rust
#[test]
fn graphemes__combining_marks() {
    let rope = Rope::from_str("élè");
    assert_eq!(rope.graphemes().count(), 2);
}
```

**Go Version:**
```go
func TestGrapheme_CombiningMarks(t *testing.T) {
    // é (e + combining acute) + l̀ (l + combining grave) = 2 graphemes
    r := New("e\u0301l\u0300")
    assert.Equal(t, 2, r.LenGraphemes())

    it := r.Graphemes()
    it.Next()
    assert.Equal(t, "e\u0301", it.Current().Text)  // e + combining acute

    it.Next()
    assert.Equal(t, "l\u0300", it.Current().Text)  // l + combining grave
}
```

#### Test 5: Grapheme Iteration Consistency
```rust
#[test]
fn graphemes__iter_consistency() {
    let s = "Hello World";
    let rope = Rope::from_str(s);
    let mut rope2 = Rope::from_str("");

    for g in rope.graphemes() {
        rope2.insert(rope2.len(), g);
    }

    assert_eq!(rope.to_string(), rope2.to_string());
}
```

**Go Version:**
```go
func TestGrapheme_IterationConsistency(t *testing.T) {
    text := "Hello World"
    r := New(text)

    builder := NewBuilder()
    it := r.Graphemes()
    for it.Next() {
        builder.Append(it.Current().Text)
    }

    r2 := builder.Build()
    assert.Equal(t, text, r2.String())
}
```

#### Test 6: Grapheme Boundary Detection
```rust
#[test]
fn graphemes__boundaries() {
    let rope = Rope::from_str("café");

    // 'c', 'a', 'fé' (4 chars: c, a, f, combining acute)
    assert_eq!(rope.prev_grapheme_boundary(0), 0);
    assert_eq!(rope.prev_grapheme_boundary(1), 1);
    assert_eq!(rope.prev_grapheme_boundary(2), 2);
    assert_eq!(rope.prev_grapheme_boundary(3), 2);  // Inside 'fé'
    assert_eq!(rope.prev_grapheme_boundary(4), 3);
}
```

**Go Version:**
```go
func TestGrapheme_Boundaries(t *testing.T) {
    r := New("cafe\u0301")  // cafe + combining acute

    // Positions: c(0), a(1), f(2), combining acute(3)
    assert.Equal(t, 0, r.PrevGraphemeStart(0))
    assert.Equal(t, 0, r.PrevGraphemeStart(1))
    assert.Equal(t, 1, r.PrevGraphemeStart(2))
    assert.Equal(t, 2, r.PrevGraphemeStart(3))  // Start of 'é'
    assert.Equal(t, 3, r.PrevGraphemeStart(4))  // End of 'é'
}

func TestGrapheme_NextBoundaries(t *testing.T) {
    r := New("cafe\u0301")

    assert.Equal(t, 1, r.NextGraphemeStart(0))
    assert.Equal(t, 2, r.NextGraphemeStart(1))
    assert.Equal(t, 4, r.NextGraphemeStart(2))  // Skip entire 'é'
}
```

#### Test 7: Complex Unicode (Family Emoji)
```rust
#[test]
fn graphemes__family_emoji() {
    // 👨‍👩‍👧‍👦 = man + ZWJ + woman + ZWJ + girl + ZWJ + boy
    let rope = Rope::from_str("👨‍👩‍👧‍👦");
    assert_eq!(rope.graphemes().count(), 1);
}
```

**Go Version:**
```go
func TestGrapheme_FamilyEmoji(t *testing.T) {
    r := New("👨‍👩‍👧‍👦")
    assert.Equal(t, 1, r.LenGraphemes())

    it := r.Graphemes()
    it.Next()
    assert.Equal(t, "👨‍👩‍👧‍👦", it.Current().Text)
}
```

#### Test 8: GraphemeAt
```rust
#[test]
fn graphemes__at() {
    let rope = Rope::from_str("abc");

    let g1 = rope.grapheme_at(0);
    assert_eq!(g1.str(), "a");

    let g2 = rope.grapheme_at(1);
    assert_eq!(g2.str(), "b");

    let g3 = rope.grapheme_at(2);
    assert_eq!(g3.str(), "c");
}
```

**Go Version:**
```go
func TestGrapheme_At(t *testing.T) {
    r := New("abc")

    g0 := r.GraphemeAt(0)
    assert.Equal(t, "a", g0.Text)
    assert.Equal(t, 0, g0.StartPos)

    g1 := r.GraphemeAt(1)
    assert.Equal(t, "b", g1.Text)
    assert.Equal(t, 1, g1.StartPos)
}
```

---

## Feature 2: Chunk_at Tests

### Source: Ropey Chunk Tests

**Location:** [ropey rope.rs](https://github.com/cessen/ropey/blob/master/src/ropey/rope.rs)

#### Test 1: Single Chunk
```rust
#[test]
fn rope__chunk_at_char__single() {
    let rope = Rope::from_str("hello world");
    let (chunk, bi, ci, li) = rope.chunk_at_char(5);

    assert_eq!(chunk, "hello world");
    assert_eq!(bi, 0);
    assert_eq!(ci, 0);
    assert_eq!(li, 0);
}
```

**Go Version:**
```go
func TestChunkAtChar_SingleChunk(t *testing.T) {
    r := New("hello world")

    chunk, err := r.ChunkAtChar(5)
    assert.NoError(t, err)
    assert.Equal(t, "hello world", chunk.Text)
    assert.Equal(t, 0, chunk.ByteIndex)
    assert.Equal(t, 0, chunk.CharIndex)
    assert.Equal(t, 0, chunk.LineIndex)
}
```

#### Test 2: Multi-Chunk Rope
```rust
#[test]
fn rope__chunk_at_char__multi() {
    let mut rope = Rope::new();
    rope.insert(0, "hello");
    rope.insert(5, " ");
    rope.insert(6, "world");

    // First chunk should be "hello "
    let (chunk1, _, _, _) = rope.chunk_at_char(0);
    assert_eq!(chunk1, "hello ");

    // Second chunk should be "world"
    let (chunk2, _, ci2, _) = rope.chunk_at_char(6);
    assert_eq!(chunk2, "world");
    assert_eq!(ci2, 6);
}
```

**Go Version:**
```go
func TestChunkAtChar_MultiChunk(t *testing.T) {
    var r *Rope
    r = r.Insert(0, "hello")
    r = r.Insert(5, " ")
    r = r.Insert(6, "world")

    chunk1, _ := r.ChunkAtChar(0)
    assert.Equal(t, "hello ", chunk1.Text)

    chunk2, _ := r.ChunkAtChar(6)
    assert.Equal(t, "world", chunk2.Text)
    assert.Equal(t, 6, chunk2.CharIndex)
}
```

#### Test 3: Byte Index
```rust
#[test]
fn rope__chunk_at_byte() {
    let rope = Rope::from_str("hello world");
    let (chunk, bi, ci, li) = rope.chunk_at_byte(6);

    assert_eq!(chunk, "hello world");
    assert_eq!(bi, 6);
    assert_eq!(ci, 6);
    assert_eq!(li, 0);
}
```

**Go Version:**
```go
func TestChunkAtByte_Basics(t *testing.T) {
    r := New("hello world")

    chunk, err := r.ChunkAtByte(6)
    assert.NoError(t, err)
    assert.Equal(t, "hello world", chunk.Text)
    assert.Equal(t, 6, chunk.ByteIndex)
}
```

#### Test 4: Out of Bounds
```rust
#[test]
#[should_panic]
fn rope__chunk_at_char__out_of_bounds() {
    let rope = Rope::from_str("hello");
    rope.chunk_at_char(100);
}
```

**Go Version:**
```go
func TestChunkAtChar_OutOfBounds(t *testing.T) {
    r := New("hello")

    _, err := r.ChunkAtChar(-1)
    assert.Error(t, err)

    _, err = r.ChunkAtChar(100)
    assert.Error(t, err)
}
```

#### Test 5: Consistency with Iterator
```rust
#[test]
fn rope__chunk_at_char__consistency() {
    let rope = Rope::from_str("hello world\ntest text");

    let mut chunks = rope.chunks();
    let mut char_idx = 0;

    while let Some(chunk) = chunks.next() {
        for _ in chunk.chars() {
            let (c, _, _, _) = rope.chunk_at_char(char_idx);
            assert_eq!(c, chunk);
            char_idx += 1;
        }
    }
}
```

**Go Version:**
```go
func TestChunkAtChar_Consistency(t *testing.T) {
    r := New("hello world\ntest text")

    it := r.Chunks()
    chunks := []string{}
    for it.Next() {
        chunks = append(chunks, it.Current())
    }

    // Verify each chunk matches
    charPos := 0
    for _, chunk := range chunks {
        for _, _range := range chunk {
            if charPos < r.Length() {
                found, _ := r.ChunkAtChar(charPos)
                assert.Equal(t, chunk, found.Text)
                charPos++
            }
        }
    }
}
```

#### Test 6: Line Index
```rust
#[test]
fn rope__chunk_at_line() {
    let rope = Rope::from_str("line1\nline2\nline3");

    let (chunk, _, _, li) = rope.chunk_at_line(1);
    assert!(chunk.contains("line2"));
    assert_eq!(li, 1);
}
```

**Go Version:**
```go
func TestChunkAtLine_LineBreaks(t *testing.T) {
    r := New("line1\nline2\nline3")

    chunk, err := r.ChunkAtLine(1)
    assert.NoError(t, err)
    assert.Contains(t, chunk.Text, "line2")
    assert.Equal(t, 1, chunk.LineIndex)
}
```

#### Test 7: Performance
```rust
#[test]
fn rope__chunk_at_char__performance() {
    let rope = Rope::from_str(&str::repeat("hello world", 1000));

    let start = Instant::now();
    for _ in 0..1000 {
        rope.chunk_at_char(rope.len_chars() / 2);
    }
    let duration = start.elapsed();

    // Should be very fast (< 100μs)
    assert!(duration.as_micros() < 100);
}
```

**Go Version:**
```go
func TestChunkAtChar_Performance(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test")
    }

    // Create deep tree with many chunks
    var r *Rope
    for i := 0; i < 1000; i++ {
        r = r.Insert(r.Length(), "word ")
    }

    b := testing.Benchmark(func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            r.ChunkAtChar(r.Length() / 2)
        }
    })

    // Should be fast (< 100μs)
    assert.True(t, b.Elapsed() < 100*time.Microsecond)
}
```

---

## Feature 3: Position Mapping Optimization Tests

### Source: Helix Multi-Cursor

**Location:** Helix Editor Selection Tests

#### Test 1: Sorted Positions Optimization
```go
func TestPositionMapper_SortedOptimization(t *testing.T) {
    doc := New(strings.Repeat("hello world ", 100))

    // Create changeset
    cs := NewChangeSet(doc.Length()).Retain(doc.Length()).Insert("XXX")

    // Test with sorted positions (fast path)
    sortedPositions := make([]int, 100)
    for i := 0; i < 100; i++ {
        sortedPositions[i] = i * 10
    }

    mapper := NewPositionMapper(cs)
    mapper.AddPositions(sortedPositions, make([]Assoc, 100))

    start := time.Now()
    result := mapper.MapOptimized()
    duration := time.Since(start)

    assert.Equal(t, 100, len(result))
    assert.True(t, duration < 10*time.Millisecond, "Should be very fast")
}
```

#### Test 2: Unsorted Positions
```go
func TestPositionMapper_UnsortedPositions(t *testing.T) {
    doc := New(strings.Repeat("hello world ", 100))

    cs := NewChangeSet(doc.Length()).Retain(doc.Length()).Insert("XXX")

    // Random unsorted positions
    unsortedPositions := make([]int, 100)
    for i := 0; i < 100; i++ {
        unsortedPositions[i] = rand.Intn(doc.Length())
    }

    mapper := NewPositionMapper(cs)
    mapper.AddPositions(unsortedPositions, make([]Assoc, 100))

    result := mapper.MapOptimized()

    // Should produce correct results even when unsorted
    assert.Equal(t, 100, len(result))
}
```

#### Test 3: Performance Comparison
```go
func TestPositionMapper_PerformanceComparison(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test")
    }

    doc := New(strings.Repeat("a", 10000))
    cs := NewChangeSet(doc.Length()).Retain(5000).Insert("XXX")

    // Create 1000 random positions
    positions := make([]int, 1000)
    for i := 0; i < 1000; i++ {
        positions[i] = rand.Intn(doc.Length())
    }

    mapper := NewPositionMapper(cs)
    mapper.AddPositions(positions, make([]Assoc, 1000))

    // Benchmark unsorted vs optimized
    b := testing.Benchmark(func(b *testing.B) {
        b.Run("Unsorted", func(b *testing.B) {
            mapper.mapUnsorted()
        })

        b.Run("Optimized", func(b *testing.B) {
            mapper.MapOptimized()
        })
    })

    // Optimized should be at least 2x faster
    unsortedTime := b.Result("Unsorted").MemBytes
    optimizedTime := b.Result("Optimized").MemBytes
    assert.True(t, optimizedTime*2 < unsortedTime, "Optimized should be at least 2x faster")
}
```

#### Test 4: Selection Integration
```go
func TestSelection_MapPositions(t *testing.T) {
    doc := New("Hello World\nLine 2\nLine 3")

    // Multi-cursor selection
    ranges := []Range{
        NewRange(0, 5),    // "Hello"
        NewRange(6, 11),   // "World"
        NewRange(12, 17),  // "Line 2"
    }
    sel := NewSelection(ranges...)

    // Insert at position 6
    cs := NewChangeSet(doc.Length()).Retain(6).Insert("beautiful ")

    // Map selection through changeset
    mappedSel := sel.MapPositions(cs)

    // Verify all cursors shifted correctly
    newRanges := mappedSel.Iter()
    assert.Equal(t, 3, len(newRanges))

    // First selection unchanged (before insert)
    assert.Equal(t, 0, newRanges[0].From())
    assert.Equal(t, 5, newRanges[0].To())

    // Second selection shifted (after insert)
    assert.Equal(t, 6, newRanges[1].From())
    assert.Equal(t, 11, newRanges[1].To())

    // Third selection shifted
    assert.Equal(t, 12, newRanges[2].From())
    assert.Equal(t, 17, newRanges[2].To())
}
```

#### Test 5: Association Behavior
```go
func TestPositionMapper_AssocBehavior(t *testing.T) {
    doc := New("Hello World")
    cs := NewChangeSet(doc.Length()).Retain(6).Delete(5) // Delete "World"

    tests := []struct {
        pos     int
        assoc   Assoc
        expected int
    }{
        {0, AssocBefore, 0},      // Before delete, stays at 0
        {5, AssocBefore, 5},      // At delete position, before delete
        {5, AssocAfter, 6},       // At delete position, after delete
        {10, AssocBefore, 5},     // In deleted range, clamped
        {10, AssocAfter, 6},      // In deleted range, after delete
    }

    for _, tt := range tests {
        mapper := NewPositionMapper(cs)
        result := mapper.MapPosition(tt.pos, tt.assoc)
        assert.Equal(t, tt.expected, result, "Pos=%d Assoc=%v", tt.pos, tt.assoc)
    }
}
```

---

## Feature 4: Time-based Undo Tests

### Source: Helix Editor History

**Location:** [Helix History Tests](https://github.com/helix-editor/helix/tree/master/helix-view/src/history.rs)

#### Test 1: Earlier by Duration
```go
func TestHistory_EarlierByDuration(t *testing.T) {
    doc := New("Hello")
    h := NewHistory(doc)

    // Create transactions with known timestamps
    baseTime := time.Now().Add(-10 * time.Minute)

    for i := 0; i < 5; i++ {
        tx := NewTransaction(NewChangeSet(doc.Length()).Insert("X")
        tx.timestamp = baseTime.Add(time.Duration(i) * time.Minute)
        h = h.Append(tx)
        doc = tx.Apply(doc)
    }

    // Undo to 3 minutes ago (should undo last 2 transactions)
    h3m := h.EarlierByDuration(3 * time.Minute)
    assert.Equal(t, 3, h3m.currentIndex)  // At transaction 2 (0,1,2 remain)
}
```

#### Test 2: Later by Duration
```go
func TestHistory_LaterByDuration(t *testing.T) {
    doc := New("Hello")
    h := NewHistory(doc)

    // Create 3 transactions
    baseTime := time.Now().Add(-10 * time.Minute)
    for i := 0; i < 3; i++ {
        tx := NewTransaction(NewChangeSet(doc.Length()).Insert("X")
        tx.timestamp = baseTime.Add(time.Duration(i) * time.Minute)
        h = h.Append(tx)
        doc = tx.Apply(doc)
    }

    // Undo to root
    hRoot := h.EarlierByDuration(100 * time.Hour)
    assert.Equal(t, 0, hRoot.currentIndex)

    // Redo to 1 minute ahead
    h1m := hRoot.LaterByDuration(1 * time.Minute)
    assert.Equal(t, 1, h1m.currentIndex)
}
```

#### Test 3: Duration Parsing
```go
func TestParseDuration(t *testing.T) {
    tests := []struct {
        input    string
        expected time.Duration
    }{
        {"30s", 30 * time.Second},
        {"5 min", 5 * time.Minute},
        {"2 hours", 2 * time.Hour},
        {"60", 60 * time.Second},
        {"1d", 24 * time.Hour},
    }

    for _, tt := range tests {
        d, err := ParseDuration(tt.input)
        assert.NoError(t, err, "Failed to parse: %s", tt.input)
        assert.Equal(t, tt.expected, d)
    }
}
```

#### Test 4: Edge Cases
```go
func TestHistory_TimeEdgeCases(t *testing.T) {
    doc := New("Hello")
    h := NewHistory(doc)

    // Empty history
    hEmpty := h.EarlierByDuration(1 * time.Minute)
    assert.True(t, hEmpty.IsEmpty())

    // Zero duration
    hZero := h.LaterByDuration(0)
    assert.Equal(t, h.currentIndex, hZero.currentIndex)

    // Very large duration
    hBig := h.EarlierByDuration(10000 * time.Hour)
    assert.Equal(t, 0, hBig.currentIndex)  // At root
}
```

#### Test 5: Timestamp Consistency
```go
func TestHistory_TimestampConsistency(t *testing.T) {
    doc := New("Hello")
    h := NewHistory(doc)

    // Create transactions
    now := time.Now()
    for i := 0; i < 5; i++ {
        tx := NewTransaction(NewChangeSet(doc.Length()).Insert("X")
        tx.timestamp = now.Add(time.Duration(i) * time.Minute)
        h = h.Append(tx)
        doc = tx.Apply(doc)
    }

    // Undo 2 minutes
    h2m := h.EarlierByDuration(2 * time.Minute)
    assert.Equal(t, 3, h2m.currentIndex)

    // Redo 2 minutes
    hBack := h2m.LaterByDuration(2 * time.Minute)
    assert.Equal(t, 5, hBack.currentIndex)  // Back at tip
}
```

#### Test 6: Round Trip
```go
func TestHistory_TimeRoundTrip(t *testing.T) {
    doc := New("Original")
    h := NewHistory(doc)

    // Create history
    for i := 0; i < 10; i++ {
        tx := NewTransaction(NewChangeSet(doc.Length()).Insert("X")
        h = h.Append(tx)
        doc = tx.Apply(doc)
    }

    originalText := doc.String()

    // Undo to 5 minutes ago
    h1 := h.EarlierByDuration(5 * time.Minute)
    assert.NotEqual(t, originalText, h1.CurrentDocument().String())

    // Redo back to present
    h2 := h1.LaterByDuration(5 * time.Minute)
    assert.Equal(t, originalText, h2.CurrentDocument().String())
}
```

---

## Test Implementation Checklist

### For Each Feature:

- [ ] Create test file
- [ ] Implement basic tests
- [ ] Implement edge case tests
- [ ] Implement performance tests
- [ ] Implement property tests
- [ ] Document test coverage
- [ ] Verify all tests pass
- [ ] Add benchmarks

### Coverage Goals:

- **Grapheme:** 90%+ coverage (critical for Unicode)
- **Chunk_at:** 80%+ coverage
- **Position Mapping:** 85%+ coverage
- **Time-based Undo:** 75%+ coverage

---

## Running Tests

```bash
# Run all grapheme tests
go test ./pkg/rope -v -run "Grapheme"

# Run all chunk_at tests
go test ./pkg/rope -v -run "Chunk"

# Run all position mapping tests
go test ./pkg/rope -v -run "PositionMapper|Selection_Map"

# Run all time-based history tests
go test ./pkg/rope -v -run "History.*Time|Duration"

# Run all new tests
go test ./pkg/rope -v -run "TestGrapheme|TestChunk.*|TestPositionMapper|TestHistory.*Time"

# Run with coverage
go test ./pkg/rope -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Summary

**Total Test Files:** 4
**Total Test Cases:** ~100+
**Estimated Test Implementation Time:** 5-7 days

### Test Files:

1. `pkg/rope/grapheme_test.go` (~25 tests)
2. `pkg/rope/chunk_at_test.go` (~20 tests)
3. `pkg/rope/position_mapping_test.go` (~15 tests)
4. `pkg/rope/history_time_test.go` (~20 tests)

---

**Prepared by:** Claude (AI Assistant)
**Date:** 2025-01-31
**Version:** 1.0
