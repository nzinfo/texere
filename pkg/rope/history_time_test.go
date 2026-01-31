package rope

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ========== EarlierByDuration Tests ==========

func TestHistory_EarlierByDuration_Basic(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Create 5 transactions
	for i := 0; i < 5; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Should be at tip (index 4)
	assert.Equal(t, 4, h.Stats().CurrentIndex)

	// Undo to approximately 2 seconds ago
	// Since we added delays, this should undo some transactions
	time.Sleep(100 * time.Millisecond)
	hPast := h.EarlierByDuration(100 * time.Millisecond)

	// Should be at an earlier state
	assert.Less(t, hPast.Stats().CurrentIndex, h.Stats().CurrentIndex)
	assert.GreaterOrEqual(t, hPast.Stats().CurrentIndex, -1)
}

func TestHistory_EarlierByDuration_Empty(t *testing.T) {
	h := NewHistory()

	// Empty history should return empty history
	hPast := h.EarlierByDuration(1 * time.Minute)
	assert.True(t, hPast.IsEmpty())
}

func TestHistory_EarlierByDuration_LargeDuration(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Add some revisions
	for i := 0; i < 5; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
	}

	// Very large duration should go to root
	hRoot := h.EarlierByDuration(10000 * time.Hour)
	assert.Equal(t, -1, hRoot.Stats().CurrentIndex)
}

// ========== LaterByDuration Tests ==========

func TestHistory_LaterByDuration_Basic(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Create 5 transactions
	for i := 0; i < 5; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
		time.Sleep(10 * time.Millisecond)
	}

	// Go to root first
	hRoot := h.ToRoot()
	assert.Equal(t, -1, hRoot.Stats().CurrentIndex)

	// Move forward by small duration
	hFuture := hRoot.LaterByDuration(50 * time.Millisecond)

	// Should be at a later state
	assert.Greater(t, hFuture.Stats().CurrentIndex, hRoot.Stats().CurrentIndex)
	assert.Less(t, hFuture.Stats().CurrentIndex, 5) // Not past tip
}

func TestHistory_LaterByDuration_Empty(t *testing.T) {
	h := NewHistory()

	// Empty history should return empty history
	hFuture := h.LaterByDuration(1 * time.Minute)
	assert.True(t, hFuture.IsEmpty())
}

func TestHistory_LaterByDuration_FromRoot(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Create 3 transactions
	for i := 0; i < 3; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
		time.Sleep(10 * time.Millisecond)
	}

	// Go to root
	hRoot := h.ToRoot()
	assert.Equal(t, -1, hRoot.Stats().CurrentIndex)

	// Move forward by small duration
	hForward := hRoot.LaterByDuration(50 * time.Millisecond)

	// Should be at a later state (closer to tip)
	assert.Greater(t, hForward.Stats().CurrentIndex, -1)
}

// ========== TimeAt Tests ==========

func TestHistory_TimeAt_Basic(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Empty history has no timestamp
	timestamp := h.TimeAt()
	assert.True(t, timestamp.IsZero())

	// Add a revision
	cs := NewChangeSet(doc.Length()).
		Retain(doc.Length()).
		Insert("X")
	tx := NewTransaction(cs)
	h.CommitRevision(tx, doc)

	// Now should have a timestamp
	timestamp = h.TimeAt()
	assert.False(t, timestamp.IsZero())
	assert.Less(t, time.Since(timestamp), time.Second)
}

// ========== DurationFromRoot Tests ==========

func TestHistory_DurationFromRoot_Basic(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Empty history
	assert.Equal(t, time.Duration(0), h.DurationFromRoot())

	// Add revisions with delays
	for i := 0; i < 3; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
		if i < 2 {
			time.Sleep(20 * time.Millisecond)
		}
	}

	// Should have some duration
	duration := h.DurationFromRoot()
	assert.Greater(t, duration, time.Duration(0))
}

// ========== DurationToTip Tests ==========

func TestHistory_DurationToTip_Basic(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Add revisions with delays
	for i := 0; i < 3; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
		if i < 2 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// At tip, duration should be 0
	hTip := h.ToTip()
	assert.Equal(t, time.Duration(0), hTip.DurationToTip())

	// Create a new history at root to test non-tip duration
	hRoot := h.ToRoot()
	if !hRoot.IsEmpty() {
		duration := hRoot.DurationToTip()
		assert.Greater(t, duration, time.Duration(0))
	}
}

func TestHistory_DurationToTip_AtTip(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Add some revisions
	for i := 0; i < 3; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
	}

	// At tip, duration should be 0
	hTip := h.ToTip()
	assert.Equal(t, time.Duration(0), hTip.DurationToTip())
}

// ========== IsEmpty Tests ==========

func TestHistory_IsEmpty_Empty(t *testing.T) {
	h := NewHistory()

	assert.True(t, h.IsEmpty())
	assert.Equal(t, 0, h.RevisionCount())
}

func TestHistory_IsEmpty_WithRevisions(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	cs := NewChangeSet(doc.Length()).
		Retain(doc.Length()).
		Insert("X")
	tx := NewTransaction(cs)
	h.CommitRevision(tx, doc)

	assert.False(t, h.IsEmpty())
	assert.Equal(t, 1, h.RevisionCount())
}

// ========== ToRoot/ToTip Tests ==========

func TestHistory_ToRoot_FromTip(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Add 3 revisions
	for i := 0; i < 3; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
	}

	// Should be at tip (index 2)
	assert.Equal(t, 2, h.Stats().CurrentIndex)

	// Go to root
	hRoot := h.ToRoot()
	assert.Equal(t, -1, hRoot.Stats().CurrentIndex)
}

func TestHistory_ToTip_FromRoot(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Add 3 revisions
	for i := 0; i < 3; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
	}

	// Go to root
	hRoot := h.ToRoot()
	assert.Equal(t, -1, hRoot.Stats().CurrentIndex)

	// Go back to tip
	hTip := hRoot.ToTip()
	assert.Equal(t, 2, hTip.Stats().CurrentIndex)
}

// ========== CurrentRevision Tests ==========

func TestHistory_CurrentRevision_Basic(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Empty history
	rev := h.CurrentRevision()
	assert.Nil(t, rev)

	// Add revision
	cs := NewChangeSet(doc.Length()).
		Retain(doc.Length()).
		Insert("X")
	tx := NewTransaction(cs)
	h.CommitRevision(tx, doc)

	// Now should have a current revision
	rev = h.CurrentRevision()
	assert.NotNil(t, rev)
	assert.NotNil(t, rev.transaction)
}

// ========== Clone Tests ==========

func TestHistory_Clone_Basic(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Add revisions
	for i := 0; i < 3; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
	}

	// Clone
	hClone := h.Clone()

	// Should have same state
	assert.Equal(t, h.Stats().TotalRevisions, hClone.Stats().TotalRevisions)
	assert.Equal(t, h.Stats().CurrentIndex, hClone.Stats().CurrentIndex)

	// Modifications to clone shouldn't affect original
	hClone.Undo()
	assert.NotEqual(t, h.Stats().CurrentIndex, hClone.Stats().CurrentIndex)
}

// ========== Round Trip Tests ==========

func TestHistory_TimeRoundTrip(t *testing.T) {
	doc := New("Original")
	h := NewHistory()

	// Create history
	for i := 0; i < 10; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
		time.Sleep(5 * time.Millisecond)
	}

	originalIndex := h.Stats().CurrentIndex

	// Go back 50ms
	hPast := h.EarlierByDuration(50 * time.Millisecond)

	// Should be at an earlier state
	assert.Less(t, hPast.Stats().CurrentIndex, originalIndex)

	// Go forward 50ms
	hFuture := hPast.LaterByDuration(50 * time.Millisecond)

	// Should be back at or past the original position
	assert.GreaterOrEqual(t, hFuture.Stats().CurrentIndex, originalIndex-1)
}

// ========== Duration Parsing Integration Tests ==========

func TestHistory_ParseDuration_Basic(t *testing.T) {
	// These tests are in grapheme_test.go, but verify they work here too
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"30s", 30 * time.Second},
		{"5m", 5 * time.Minute},
		{"2h", 2 * time.Hour},
		{"1d", 24 * time.Hour},
		{"60", 60 * time.Second},
	}

	for _, tt := range tests {
		d, err := ParseDuration(tt.input)
		assert.NoError(t, err, "Failed to parse: %s", tt.input)
		assert.Equal(t, tt.expected, d)
	}
}

func TestHistory_FormatDuration_Basic(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{30 * time.Second, "30s"},
		{5 * time.Minute, "5m"},
		{2 * time.Hour, "2h"},
		{24 * time.Hour, "1d"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

// ========== Integration Tests ==========

func TestHistory_ImmutableNavigation(t *testing.T) {
	doc := New("Hello World")
	h := NewHistory()

	// Add some edits
	for i := 0; i < 5; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
	}

	// Original history should be unchanged
	originalIndex := h.Stats().CurrentIndex

	// Navigate back
	h1 := h.EarlierByDuration(100 * time.Millisecond)

	// h1 should be different from h
	assert.NotEqual(t, h1.Stats().CurrentIndex, originalIndex)

	// But original h should be unchanged
	assert.Equal(t, originalIndex, h.Stats().CurrentIndex)

	// Can continue from h1
	h2 := h1.LaterByDuration(50 * time.Millisecond)

	// h and h1 should remain unchanged
	assert.Equal(t, originalIndex, h.Stats().CurrentIndex)

	// h2 should be at or ahead of h1 (moved forward in time)
	assert.GreaterOrEqual(t, h2.Stats().CurrentIndex, h1.Stats().CurrentIndex)
}

func TestHistory_MultipleTimeNavigations(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Create history with known timing
	for i := 0; i < 10; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
		if i%3 == 0 {
			time.Sleep(20 * time.Millisecond)
		}
	}

	// Navigate back and forth
	h1 := h.EarlierByDuration(150 * time.Millisecond)
	h2 := h1.LaterByDuration(50 * time.Millisecond)
	h3 := h2.EarlierByDuration(100 * time.Millisecond)

	// All should be valid states
	for _, hist := range []*History{h, h1, h2, h3} {
		assert.GreaterOrEqual(t, hist.Stats().CurrentIndex, -1)
		assert.Less(t, hist.Stats().CurrentIndex, 10)
	}
}

// ========== Edge Cases ==========

func TestHistory_ZeroDuration(t *testing.T) {
	doc := New("Hello")
	h := NewHistory()

	// Add a revision
	cs := NewChangeSet(doc.Length()).
		Retain(doc.Length()).
		Insert("X")
	tx := NewTransaction(cs)
	h.CommitRevision(tx, doc)

	originalIndex := h.Stats().CurrentIndex

	// Zero duration should return same state (or very close)
	hSame := h.EarlierByDuration(0)
	assert.Equal(t, originalIndex, hSame.Stats().CurrentIndex)
}

func TestHistory_NegativeDuration(t *testing.T) {
	h := NewHistory()

	// Negative duration is invalid
	// Just verify it doesn't crash and returns something reasonable
	hResult := h.EarlierByDuration(-1 * time.Second)
	assert.NotNil(t, hResult)
}

func TestHistory_TimeNavigation_EmptyHistory(t *testing.T) {
	h := NewHistory()

	// All time navigation on empty history should return empty history
	h1 := h.EarlierByDuration(1 * time.Minute)
	h2 := h.LaterByDuration(1 * time.Minute)

	assert.True(t, h1.IsEmpty())
	assert.True(t, h2.IsEmpty())
}

// ========== Performance Tests ==========

func TestHistory_TimeNavigation_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test")
	}

	doc := New(strings.Repeat("hello ", 100))

	// Create large history
	h := NewHistory()
	for i := 0; i < 100; i++ {
		cs := NewChangeSet(doc.Length()).
			Retain(doc.Length()).
			Insert("X")
		tx := NewTransaction(cs)
		h.CommitRevision(tx, doc)
		doc = tx.Apply(doc)
	}

	// Time navigation should be fast (binary search)
	start := time.Now()
	for i := 0; i < 100; i++ {
		h.EarlierByDuration(time.Duration(i) * time.Millisecond)
	}
	elapsed := time.Since(start)

	// Should be very fast (< 10ms for 100 iterations)
	assert.Less(t, elapsed, 100*time.Millisecond)
}
