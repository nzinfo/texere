package rope

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Enhanced SavePoint Tests
// ============================================================================

func TestEnhancedSavePoint_Basic(t *testing.T) {
	r := New("Hello World")

	metadata := SavePointMetadata{
		UserID:      "user1",
		ViewID:      "view1",
		Tags:        []string{"important", "checkpoint"},
		Description: "Initial savepoint",
	}

	sp := NewEnhancedSavePoint(r, 1, metadata)

	assert.NotNil(t, sp)
	assert.Equal(t, "user1", sp.Metadata().UserID)
	assert.Equal(t, "view1", sp.Metadata().ViewID)
	assert.True(t, sp.HasTag("important"))
	assert.True(t, sp.HasTag("checkpoint"))
	assert.False(t, sp.HasTag("other"))
	assert.NotEmpty(t, sp.Hash())
}

func TestEnhancedSavePoint_Metadata(t *testing.T) {
	r := New("Hello World")
	sp := NewEnhancedSavePoint(r, 1, SavePointMetadata{})

	// Set metadata
	newMetadata := SavePointMetadata{
		UserID:      "user2",
		ViewID:      "view2",
		Tags:        []string{"tag1", "tag2"},
		Description: "Updated metadata",
	}
	sp.SetMetadata(newMetadata)

	metadata := sp.Metadata()
	assert.Equal(t, "user2", metadata.UserID)
	assert.Equal(t, "view2", metadata.ViewID)
	assert.Equal(t, "tag1", metadata.Tags[0])
	assert.Equal(t, "tag2", metadata.Tags[1])
}

func TestEnhancedSavePoint_Tags(t *testing.T) {
	r := New("Hello World")
	metadata := SavePointMetadata{
		Tags: []string{"tag1"},
	}
	sp := NewEnhancedSavePoint(r, 1, metadata)

	// Add tags
	sp.AddTags("tag2", "tag3")
	assert.True(t, sp.HasTag("tag1"))
	assert.True(t, sp.HasTag("tag2"))
	assert.True(t, sp.HasTag("tag3"))

	// Try to add duplicate
	sp.AddTags("tag1")
	assert.True(t, sp.HasTag("tag1"))

	// Remove tag
	sp.RemoveTag("tag2")
	assert.False(t, sp.HasTag("tag2"))
	assert.True(t, sp.HasTag("tag1"))
	assert.True(t, sp.HasTag("tag3"))
}

// ============================================================================
// Enhanced SavePointManager Tests
// ============================================================================

func TestEnhancedSavePointManager_Basic(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	r := New("Hello World")

	metadata := SavePointMetadata{
		UserID:      "user1",
		Tags:        []string{"checkpoint"},
		Description: "Test savepoint",
	}

	id, isDup := sm.Create(r, 1, metadata)

	assert.False(t, isDup)
	assert.Equal(t, 0, id)
	assert.Equal(t, 1, sm.Count())

	// Retrieve savepoint
	sp := sm.Get(id)
	assert.NotNil(t, sp)
	assert.Equal(t, "user1", sp.Metadata().UserID)
}

func TestEnhancedSavePointManager_DuplicateDetection_Skip(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeSkip)

	r1 := New("Hello World")
	r2 := New("Hello World") // Same content

	metadata := SavePointMetadata{
		UserID: "user1",
		Tags:   []string{"checkpoint"},
	}

	// First savepoint
	id1, isDup1 := sm.Create(r1, 1, metadata)
	assert.False(t, isDup1)
	assert.Equal(t, 0, id1)
	assert.Equal(t, 1, sm.Count())

	// Second savepoint with same content should be skipped
	id2, isDup2 := sm.Create(r2, 2, metadata)
	assert.True(t, isDup2)
	assert.Equal(t, 0, id2) // Returns first ID
	assert.Equal(t, 1, sm.Count()) // Still only 1 savepoint
}

func TestEnhancedSavePointManager_DuplicateDetection_Allow(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow)

	r1 := New("Hello World")
	r2 := New("Hello World")

	metadata := SavePointMetadata{
		UserID: "user1",
		Tags:   []string{"checkpoint"},
	}

	// Both savepoints should be created
	id1, isDup1 := sm.Create(r1, 1, metadata)
	assert.False(t, isDup1)
	assert.Equal(t, 0, id1)

	id2, isDup2 := sm.Create(r2, 2, metadata)
	assert.False(t, isDup2)
	assert.Equal(t, 1, id2)
	assert.Equal(t, 2, sm.Count())
}

func TestEnhancedSavePointManager_DuplicateDetection_Replace(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeReplace)

	r1 := New("Hello World")
	r2 := New("Hello World")

	metadata := SavePointMetadata{
		UserID: "user1",
		Tags:   []string{"checkpoint"},
	}

	// First savepoint
	id1, _ := sm.Create(r1, 1, metadata)
	assert.Equal(t, 0, id1)
	assert.Equal(t, 1, sm.Count())

	// Second savepoint should replace the first
	id2, _ := sm.Create(r2, 2, metadata)
	assert.Equal(t, 1, id2) // New ID
	assert.Equal(t, 1, sm.Count()) // Still only 1 savepoint

	// First savepoint should no longer exist
	assert.Nil(t, sm.Get(id1))
	assert.NotNil(t, sm.Get(id2))
}

func TestEnhancedSavePointManager_Query(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow) // Allow duplicates for testing

	// Create savepoints at different times
	now := time.Now()

	r1 := New("Hello")
	sm.Create(r1, 1, SavePointMetadata{
		UserID: "user1",
		Tags:   []string{"tag1", "important"},
	})

	time.Sleep(10 * time.Millisecond)

	r2 := New("World")
	sm.Create(r2, 2, SavePointMetadata{
		UserID: "user2",
		Tags:   []string{"tag2", "important"},
	})

	time.Sleep(10 * time.Millisecond)

	r3 := New("Test")
	sm.Create(r3, 3, SavePointMetadata{
		UserID: "user1",
		Tags:   []string{"tag1"},
	})

	// Query all
	all := sm.Query(SavePointQuery{})
	assert.Equal(t, 3, len(all))

	// Query by user
	user1Results := sm.ByUser("user1", 0)
	assert.Equal(t, 2, len(user1Results))

	user2Results := sm.ByUser("user2", 0)
	assert.Equal(t, 1, len(user2Results))

	// Query by tag
	tag1Results := sm.ByTag("tag1", 0)
	assert.Equal(t, 2, len(tag1Results))

	importantResults := sm.ByTag("important", 0)
	assert.Equal(t, 2, len(importantResults))

	// Query with limit
	limited := sm.Recent(2)
	assert.Equal(t, 2, len(limited))

	// Query by time range
	start := now
	end := now.Add(5 * time.Millisecond)
	inRange := sm.ByTime(start, end, 0)
	assert.Equal(t, 1, len(inRange)) // Only the first one
}

func TestEnhancedSavePointManager_HasDuplicate(t *testing.T) {
	sm := NewEnhancedSavePointManager()

	r1 := New("Hello World")
	r2 := New("Hello World")
	r3 := New("Different")

	metadata := SavePointMetadata{UserID: "user1"}

	sm.Create(r1, 1, metadata)

	assert.True(t, sm.HasDuplicate(r2))
	assert.False(t, sm.HasDuplicate(r3))

	// Get duplicates
	dups := sm.GetDuplicates(r2)
	assert.Equal(t, 1, len(dups))
}

func TestEnhancedSavePointManager_Restore(t *testing.T) {
	sm := NewEnhancedSavePointManager()

	r := New("Hello World")
	metadata := SavePointMetadata{UserID: "user1"}

	id, _ := sm.Create(r, 1, metadata)

	// Modify original
	r = r.Append(" Modified")

	// Restore should give original content
	restored := sm.Restore(id)
	assert.NotNil(t, restored)
	assert.Equal(t, "Hello World", restored.String())
	assert.NotEqual(t, r.String(), restored.String())
}

func TestEnhancedSavePointManager_Release(t *testing.T) {
	sm := NewEnhancedSavePointManager()

	r := New("Hello World")
	metadata := SavePointMetadata{UserID: "user1"}

	id, _ := sm.Create(r, 1, metadata)
	assert.Equal(t, 1, sm.Count())

	// Get increments ref count
	sp := sm.Get(id)
	assert.NotNil(t, sp)
	assert.Equal(t, 2, sp.RefCount()) // Initial 1 + Get

	// Release decrements
	sm.Release(id)
	assert.Equal(t, 1, sp.RefCount()) // Back to initial

	// Release again should remove (ref count goes to 0)
	sm.Release(id)
	assert.Equal(t, 0, sm.Count())
	assert.Nil(t, sm.Get(id))
}

func TestEnhancedSavePointManager_CleanOlderThan(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow) // Allow duplicates for testing

	r := New("Hello")
	metadata := SavePointMetadata{UserID: "user1"}

	sm.Create(r, 1, metadata)

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	sm.Create(r, 2, metadata)

	// Clean older than 5ms should remove first one
	removed := sm.CleanOlderThan(5 * time.Millisecond)
	assert.Equal(t, 1, removed)
	assert.Equal(t, 1, sm.Count())
}

func TestEnhancedSavePointManager_CleanByTag(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow) // Allow duplicates for testing

	r := New("Hello")

	// Create savepoints with different tags
	sm.Create(r, 1, SavePointMetadata{
		Tags: []string{"tag1", "important"},
	})

	sm.Create(r, 2, SavePointMetadata{
		Tags: []string{"tag2", "important"},
	})

	sm.Create(r, 3, SavePointMetadata{
		Tags: []string{"tag3"},
	})

	// Clean by tag1
	removed := sm.CleanByTag("tag1")
	assert.Equal(t, 1, removed)
	assert.Equal(t, 2, sm.Count())

	// Remaining should be tag2 and tag3
	assert.Nil(t, sm.Get(0))
	assert.NotNil(t, sm.Get(1))
	assert.NotNil(t, sm.Get(2))
}

func TestEnhancedSavePointManager_Stats(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow) // Allow duplicates for testing

	r := New("Hello")

	// Create savepoints
	sm.Create(r, 1, SavePointMetadata{
		UserID: "user1",
		Tags:   []string{"tag1", "tag2"},
	})

	sm.Create(r, 2, SavePointMetadata{
		UserID: "user2",
		Tags:   []string{"tag1"},
	})

	stats := sm.Stats()
	assert.Equal(t, 2, stats.TotalSavepoints)
	assert.Equal(t, 2, stats.TotalUsers) // user1, user2
	assert.Equal(t, 2, stats.TotalTags)  // tag1, tag2
	assert.Equal(t, 1, stats.UniqueHashes) // Same content
	assert.Greater(t, stats.AvgRefCount, 0.0)
}

func TestEnhancedSavePointManager_Clear(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow) // Allow duplicates for testing

	r := New("Hello")
	metadata := SavePointMetadata{UserID: "user1"}

	sm.Create(r, 1, metadata)
	sm.Create(r, 2, metadata)

	assert.Equal(t, 2, sm.Count())

	sm.Clear()

	assert.Equal(t, 0, sm.Count())
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestEnhancedSavePointManager_MultiUser(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow) // Allow duplicates for testing

	r := New("Hello World")

	// User 1 creates savepoint
	id1, _ := sm.Create(r, 1, SavePointMetadata{
		UserID:      "user1",
		Description: "User 1 checkpoint",
	})

	// User 2 creates savepoint
	id2, _ := sm.Create(r, 2, SavePointMetadata{
		UserID:      "user2",
		Description: "User 2 checkpoint",
	})

	// Each user should see their own savepoints
	user1SPs := sm.ByUser("user1", 0)
	assert.Equal(t, 1, len(user1SPs))
	assert.Equal(t, id1, user1SPs[0].ID)

	user2SPs := sm.ByUser("user2", 0)
	assert.Equal(t, 1, len(user2SPs))
	assert.Equal(t, id2, user2SPs[0].ID)
}

func TestEnhancedSavePointManager_ComplexWorkflow(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow) // Allow for testing

	// Simulate a document editing workflow
	r := New("Hello")

	// Initial savepoint
	id1, _ := sm.Create(r, 1, SavePointMetadata{
		UserID:      "user1",
		Tags:        []string{"initial", "auto"},
		Description: "Initial state",
	})

	// Make edits and create savepoints
	r = r.Append(" World")
	id2, _ := sm.Create(r, 2, SavePointMetadata{
		UserID:      "user1",
		Tags:        []string{"checkpoint"},
		Description: "After adding World",
	})

	r = r.Insert(5, " Beautiful")
	id3, _ := sm.Create(r, 3, SavePointMetadata{
		UserID:      "user1",
		Tags:        []string{"checkpoint"},
		Description: "After adding Beautiful",
	})

	// Query checkpoints
	checkpoints := sm.ByTag("checkpoint", 0)
	assert.Equal(t, 2, len(checkpoints))

	// Restore to earlier state
	restored := sm.Restore(id2)
	assert.Equal(t, "Hello World", restored.String())

	// Verify all savepoints exist
	assert.NotNil(t, sm.Get(id1))
	assert.NotNil(t, sm.Get(id2))
	assert.NotNil(t, sm.Get(id3))
}
