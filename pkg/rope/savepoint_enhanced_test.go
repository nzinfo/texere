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

// ============================================================================
// QueryPreallocated Tests
// ============================================================================

func TestEnhancedSavePointManager_QueryPreallocated(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow)

	now := time.Now()

	// Create multiple savepoints
	r1 := New("Content1")
	sm.Create(r1, 1, SavePointMetadata{
		UserID: "user1",
		Tags:   []string{"tag1", "important"},
	})

	time.Sleep(5 * time.Millisecond)

	r2 := New("Content2")
	sm.Create(r2, 2, SavePointMetadata{
		UserID: "user2",
		Tags:   []string{"tag2"},
	})

	time.Sleep(5 * time.Millisecond)

	r3 := New("Content3")
	sm.Create(r3, 3, SavePointMetadata{
		UserID: "user1",
		Tags:   []string{"tag1"},
	})

	t.Run("query all with preallocated slice", func(t *testing.T) {
		results := make([]SavePointResult, 0, 16)
		results = sm.QueryPreallocated(SavePointQuery{}, results)

		assert.Equal(t, 3, len(results))
	})

	t.Run("query with user filter", func(t *testing.T) {
		userID := "user1"
		results := make([]SavePointResult, 0, 16)
		results = sm.QueryPreallocated(SavePointQuery{
			UserID: &userID,
		}, results)

		assert.Equal(t, 2, len(results))
		for _, r := range results {
			assert.Equal(t, "user1", r.Metadata.UserID)
		}
	})

	t.Run("query with tag filter", func(t *testing.T) {
		tag := "tag1"
		results := make([]SavePointResult, 0, 16)
		results = sm.QueryPreallocated(SavePointQuery{
			Tag: &tag,
		}, results)

		assert.Equal(t, 2, len(results))
	})

	t.Run("query with time range", func(t *testing.T) {
		start := now
		end := now.Add(7 * time.Millisecond) // Between first and second (5ms and 10ms)
		results := make([]SavePointResult, 0, 16)
		results = sm.QueryPreallocated(SavePointQuery{
			StartTime: &start,
			EndTime:   &end,
		}, results)

		// Should get 1 or 2 results depending on timing
		count := len(results)
		assert.True(t, count >= 1 && count <= 2, "Expected 1-2 results, got %d", count)
	})

	t.Run("query with limit", func(t *testing.T) {
		results := make([]SavePointResult, 0, 16)
		results = sm.QueryPreallocated(SavePointQuery{
			Limit: 2,
		}, results)

		assert.Equal(t, 2, len(results))
	})

	t.Run("reuse slice multiple times", func(t *testing.T) {
		results := make([]SavePointResult, 0, 16)

		// First query
		results = sm.QueryPreallocated(SavePointQuery{
			UserID: func() *string { s := "user1"; return &s }(),
		}, results)
		count1 := len(results)
		assert.Equal(t, 2, count1)

		// Reuse slice for second query
		results = sm.QueryPreallocated(SavePointQuery{
			UserID: func() *string { s := "user2"; return &s }(),
		}, results)
		count2 := len(results)
		assert.Equal(t, 1, count2)
	})

	t.Run("nil slice creates new one", func(t *testing.T) {
		var results []SavePointResult
		results = sm.QueryPreallocated(SavePointQuery{}, results)

		assert.Equal(t, 3, len(results))
	})
}

// ============================================================================
// QueryByTime Tests
// ============================================================================

func TestEnhancedSavePointManager_QueryByTime(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow)

	now := time.Now()

	// Create savepoints at different times
	sm.Create(New("V1"), 1, SavePointMetadata{UserID: "user1"})
	time.Sleep(10 * time.Millisecond)

	sm.Create(New("V2"), 2, SavePointMetadata{UserID: "user1"})
	time.Sleep(10 * time.Millisecond)

	sm.Create(New("V3"), 3, SavePointMetadata{UserID: "user1"})

	t.Run("query with time range", func(t *testing.T) {
		start := now
		end := now.Add(15 * time.Millisecond)

		results := sm.ByTime(start, end, 0)

		assert.Equal(t, 2, len(results))
	})

	t.Run("query with limit", func(t *testing.T) {
		start := now
		end := time.Now().Add(100 * time.Millisecond)

		results := sm.ByTime(start, end, 2)

		assert.Equal(t, 2, len(results))
	})

	t.Run("query returns newest first", func(t *testing.T) {
		start := now
		end := time.Now().Add(100 * time.Millisecond)

		results := sm.ByTime(start, end, 0)

		// Results should be sorted by timestamp (newest first)
		if len(results) >= 2 {
			assert.True(t, results[0].Timestamp.After(results[1].Timestamp))
		}
	})

	t.Run("empty time range", func(t *testing.T) {
		future := time.Now().Add(1 * time.Hour)
		start := future
		end := future.Add(1 * time.Hour)

		results := sm.ByTime(start, end, 0)

		assert.Equal(t, 0, len(results))
	})
}

// ============================================================================
// QueryByTag Tests
// ============================================================================

func TestEnhancedSavePointManager_QueryByTag(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow)

	// Create savepoints with different tags
	sm.Create(New("C1"), 1, SavePointMetadata{
		Tags: []string{"important", "checkpoint"},
	})
	sm.Create(New("C2"), 2, SavePointMetadata{
		Tags: []string{"checkpoint"},
	})
	sm.Create(New("C3"), 3, SavePointMetadata{
		Tags: []string{"important"},
	})
	sm.Create(New("C4"), 4, SavePointMetadata{
		Tags: []string{"other"},
	})

	t.Run("query by single tag", func(t *testing.T) {
		results := sm.ByTag("important", 0)

		assert.Equal(t, 2, len(results))
		for _, r := range results {
			assert.True(t, r.SavePoint.HasTag("important"))
		}
	})

	t.Run("query with limit", func(t *testing.T) {
		results := sm.ByTag("checkpoint", 1)

		assert.Equal(t, 1, len(results))
	})

	t.Run("query non-existent tag", func(t *testing.T) {
		results := sm.ByTag("nonexistent", 0)

		assert.Equal(t, 0, len(results))
	})

	t.Run("multiple savepoints same tag", func(t *testing.T) {
		sm.Create(New("C5"), 5, SavePointMetadata{
			Tags: []string{"important", "new"},
		})

		results := sm.ByTag("important", 0)

		assert.Equal(t, 3, len(results))
	})
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

func TestEnhancedSavePointManager_ConcurrentCreate(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow)

	done := make(chan bool)
	numGoroutines := 10
	savesPerGoroutine := 10

	// Concurrent creates
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < savesPerGoroutine; j++ {
				r := New("Content" + string(rune('A'+id%26)))
				sm.Create(r, id*savesPerGoroutine+j, SavePointMetadata{
					UserID: "user" + string(rune('0'+id%10)),
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify count
	expectedCount := numGoroutines * savesPerGoroutine
	assert.Equal(t, expectedCount, sm.Count())
}

func TestEnhancedSavePointManager_ConcurrentQuery(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow)

	// Create some savepoints first
	for i := 0; i < 20; i++ {
		r := New("Content" + string(rune('A'+i%26)))
		sm.Create(r, i, SavePointMetadata{
			UserID: "user" + string(rune('0'+i%5)),
			Tags:   []string{"tag" + string(rune('0'+i%3))},
		})
	}

	done := make(chan bool)

	// Concurrent queries
	for i := 0; i < 10; i++ {
		go func(id int) {
			// Different query types
			switch id % 4 {
			case 0:
				sm.Query(SavePointQuery{})
			case 1:
				userID := "user0"
				sm.ByUser(userID, 10)
			case 2:
				sm.ByTag("tag0", 10)
			case 3:
				sm.Recent(5)
			}
			done <- true
		}(i)
	}

	// Wait for all queries
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no data loss
	assert.Equal(t, 20, sm.Count())
}

func TestEnhancedSavePointManager_ConcurrentCreateQuery(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow)

	done := make(chan bool)
	numOps := 50

	// Mix of creates and queries
	for i := 0; i < numOps; i++ {
		go func(id int) {
			if id%2 == 0 {
				// Create
				r := New("Content" + string(rune('A'+id%26)))
				sm.Create(r, id, SavePointMetadata{
					UserID: "user" + string(rune('0'+id%3)),
					Tags:   []string{"tag" + string(rune('0'+id%2))},
				})
			} else {
				// Query
				sm.Query(SavePointQuery{})
				sm.Recent(5)
			}
			done <- true
		}(i)
	}

	// Wait for all operations
	for i := 0; i < numOps; i++ {
		<-done
	}

	// Verify no deadlocks or panics
	assert.Greater(t, sm.Count(), 0)
}

func TestEnhancedSavePointManager_ConcurrentGetRelease(t *testing.T) {
	sm := NewEnhancedSavePointManager()

	// Create initial savepoints
	for i := 0; i < 10; i++ {
		r := New("Content" + string(rune('A'+i%26)))
		sm.Create(r, i, SavePointMetadata{UserID: "user1"})
	}

	done := make(chan bool)

	// Concurrent gets and releases
	for i := 0; i < 20; i++ {
		go func(id int) {
			// Get and release random savepoint
			spID := id % 10
			sp := sm.Get(spID)
			if sp != nil {
				sm.Release(spID)
			}
			done <- true
		}(i)
	}

	// Wait for all operations
	for i := 0; i < 20; i++ {
		<-done
	}

	// Some savepoints may still exist
	// The important thing is no deadlock or panic
}

func TestEnhancedSavePointManager_ConcurrentRestore(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow)

	// Create savepoints
	for i := 0; i < 10; i++ {
		r := New("Content" + string(rune('A'+i%26)))
		sm.Create(r, i, SavePointMetadata{UserID: "user1"})
	}

	done := make(chan bool)

	// Concurrent restores
	for i := 0; i < 20; i++ {
		go func(id int) {
			spID := id % 10
			restored := sm.Restore(spID)
			assert.NotNil(t, restored)
			assert.NotEmpty(t, restored.String())
			done <- true
		}(i)
	}

	// Wait for all operations
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestEnhancedSavePointManager_ConcurrentStats(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow)

	done := make(chan bool)

	// Mix of creates, queries, and stats
	for i := 0; i < 30; i++ {
		go func(id int) {
			switch id % 3 {
			case 0:
				r := New("Content" + string(rune('A'+id%26)))
				sm.Create(r, id, SavePointMetadata{
					UserID: "user" + string(rune('0'+id%3)),
				})
			case 1:
				sm.Query(SavePointQuery{})
			case 2:
				sm.Stats()
			}
			done <- true
		}(i)
	}

	// Wait for all operations
	for i := 0; i < 30; i++ {
		<-done
	}

	// Verify stats are consistent
	stats := sm.Stats()
	assert.Greater(t, stats.TotalSavepoints, 0)
	assert.Greater(t, stats.TotalUsers, 0)
}

// ============================================================================
// QueryOptimized Tests
// ============================================================================

func TestEnhancedSavePointManager_QueryOptimized(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow)

	// Create savepoints
	for i := 0; i < 10; i++ {
		r := New("Content" + string(rune('A'+i%26)))
		sm.Create(r, i, SavePointMetadata{
			UserID: "user" + string(rune('0'+i%3)),
			Tags:   []string{"tag" + string(rune('0'+i%2))},
		})
	}

	t.Run("query optimized returns results", func(t *testing.T) {
		results := sm.QueryOptimized(SavePointQuery{})

		assert.Equal(t, 10, len(results))
	})

	t.Run("query optimized with filter", func(t *testing.T) {
		userID := "user0"
		results := sm.QueryOptimized(SavePointQuery{
			UserID: &userID,
		})

		assert.Greater(t, len(results), 0)
		for _, r := range results {
			assert.Equal(t, "user0", r.Metadata.UserID)
		}
	})

	t.Run("query optimized with limit", func(t *testing.T) {
		results := sm.QueryOptimized(SavePointQuery{
			Limit: 5,
		})

		assert.Equal(t, 5, len(results))
	})

	t.Run("repeated queries use pool", func(t *testing.T) {
		// Run multiple queries to test pool usage
		for i := 0; i < 10; i++ {
			results := sm.QueryOptimized(SavePointQuery{
				Limit: 3,
			})
			assert.Equal(t, 3, len(results))
		}
	})
}

// ============================================================================
// Edge Cases and Error Handling
// ============================================================================

func TestEnhancedSavePointManager_EmptyQueries(t *testing.T) {
	sm := NewEnhancedSavePointManager()

	t.Run("query empty manager", func(t *testing.T) {
		results := sm.Query(SavePointQuery{})
		assert.Equal(t, 0, len(results))
	})

	t.Run("query preallocated empty manager", func(t *testing.T) {
		results := make([]SavePointResult, 0, 16)
		results = sm.QueryPreallocated(SavePointQuery{}, results)
		assert.Equal(t, 0, len(results))
	})

	t.Run("query optimized empty manager", func(t *testing.T) {
		results := sm.QueryOptimized(SavePointQuery{})
		assert.Equal(t, 0, len(results))
	})

	t.Run("by time empty manager", func(t *testing.T) {
		now := time.Now()
		results := sm.ByTime(now, now.Add(1*time.Hour), 0)
		assert.Equal(t, 0, len(results))
	})

	t.Run("by tag empty manager", func(t *testing.T) {
		results := sm.ByTag("any", 0)
		assert.Equal(t, 0, len(results))
	})

	t.Run("by user empty manager", func(t *testing.T) {
		results := sm.ByUser("any", 0)
		assert.Equal(t, 0, len(results))
	})
}

func TestEnhancedSavePointManager_QueryFilters(t *testing.T) {
	sm := NewEnhancedSavePointManager()
	sm.SetDuplicateMode(DuplicateModeAllow)

	now := time.Now()

	r1 := New("Hello")
	sm.Create(r1, 1, SavePointMetadata{
		UserID:      "user1",
		Tags:        []string{"tag1", "important"},
		Description: "Test1",
	})

	time.Sleep(5 * time.Millisecond)

	r2 := New("World")
	sm.Create(r2, 2, SavePointMetadata{
		UserID:      "user2",
		Tags:        []string{"tag2"},
		Description: "Test2",
	})

	t.Run("combined filters", func(t *testing.T) {
		userID := "user1"
		tag := "tag1"
		start := now
		end := time.Now().Add(100 * time.Millisecond)

		results := sm.Query(SavePointQuery{
			UserID:    &userID,
			Tag:       &tag,
			StartTime: &start,
			EndTime:   &end,
			Limit:     10,
		})

		assert.Equal(t, 1, len(results))
		assert.Equal(t, "user1", results[0].Metadata.UserID)
		assert.True(t, results[0].SavePoint.HasTag("tag1"))
	})

	t.Run("hash filter", func(t *testing.T) {
		hash := r1.HashCode()
		hashStr := HashToString(hash)

		results := sm.Query(SavePointQuery{
			Hash: &hashStr,
		})

		assert.Greater(t, len(results), 0)
		for _, r := range results {
			assert.Equal(t, hashStr, r.SavePoint.Hash())
		}
	})
}
