package transport

import (
	"context"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// TestPatchManager_ComputePatch tests basic patch computation.
func TestPatchManager_ComputePatch(t *testing.T) {
	pm := NewPatchManager()

	oldText := "Hello World"
	newText := "Hello Beautiful World"

	result := pm.ComputePatch(oldText, newText)

	if result.Patch == "" {
		t.Error("Expected patch to be generated")
	}

	if result.OldSize != len(oldText) {
		t.Errorf("Expected OldSize %d, got %d", len(oldText), result.OldSize)
	}

	if result.NewSize != len(newText) {
		t.Errorf("Expected NewSize %d, got %d", len(newText), result.NewSize)
	}

	// Patch should be smaller than the full text
	if result.PatchSize >= result.NewSize {
		t.Logf("Warning: Patch size %d >= new text size %d (no compression achieved)", result.PatchSize, result.NewSize)
	}

	t.Logf("Patch: %s", result.Patch)
	t.Logf("Old size: %d, New size: %d, Patch size: %d, Saved: %d bytes",
		result.OldSize, result.NewSize, result.PatchSize, result.SavedBytes)
}

// TestPatchManager_ApplyPatch tests patch application.
func TestPatchManager_ApplyPatch(t *testing.T) {
	pm := NewPatchManager()

	oldText := "The quick brown fox"
	newText := "The quick brown fox jumps over the lazy dog"

	// Compute patch
	patchResult := pm.ComputePatch(oldText, newText)

	// Apply patch
	applyResult := pm.ApplyPatch(oldText, patchResult.Patch)

	if !applyResult.Success {
		t.Error("Expected patch application to succeed")
	}

	if applyResult.Content != newText {
		t.Errorf("Expected reconstructed text '%s', got '%s'", newText, applyResult.Content)
	}

	t.Logf("Successfully applied patch: %d patches applied", applyResult.PatchesApplied)
}

// TestPatchManager_RoundTrip tests patch computation and application round-trip.
func TestPatchManager_RoundTrip(t *testing.T) {
	pm := NewPatchManager()

	testCases := []struct {
		name     string
		oldText  string
		newText  string
	}{
		{
			name:    "Simple insertion",
			oldText: "Hello World",
			newText: "Hello Beautiful World",
		},
		{
			name:    "Simple deletion",
			oldText: "Hello Beautiful World",
			newText: "Hello World",
		},
		{
			name:    "Replacement",
			oldText: "The quick brown fox",
			newText: "The slow red fox",
		},
		{
			name:    "Multiple changes",
			oldText: "ABC",
			newText: "A B C",
		},
		{
			name:    "Large text",
			oldText: "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
			newText: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Compute patch
			patchResult := pm.ComputePatch(tc.oldText, tc.newText)

			// Apply patch
			applyResult := pm.ApplyPatch(tc.oldText, patchResult.Patch)

			if !applyResult.Success {
				t.Errorf("Patch application failed for '%s'", tc.name)
				return
			}

			if applyResult.Content != tc.newText {
				t.Errorf("Round-trip failed for '%s': expected '%s', got '%s'",
					tc.name, tc.newText, applyResult.Content)
			}

			t.Logf("'%s': Old=%d, New=%d, Patch=%d, Saved=%d bytes",
				tc.name, patchResult.OldSize, patchResult.NewSize, patchResult.PatchSize, patchResult.SavedBytes)
		})
	}
}

// TestPatchManager_EmptyPatch tests handling of identical texts.
func TestPatchManager_EmptyPatch(t *testing.T) {
	pm := NewPatchManager()

	text := "Hello World"
	result := pm.ComputePatch(text, text)

	// Patch should be very small or empty for identical texts
	t.Logf("Empty patch result: PatchSize=%d, Patch='%s'", result.PatchSize, result.Patch)
}

// TestPatchManager_ComputeDiff tests diff computation.
func TestPatchManager_ComputeDiff(t *testing.T) {
	pm := NewPatchManager()

	oldText := "Hello World"
	newText := "Hello Beautiful World"

	diffs := pm.ComputeDiff(oldText, newText)

	if len(diffs) == 0 {
		t.Error("Expected diffs to be computed")
	}

	t.Logf("Diffs: %d operations", len(diffs))
	for i, diff := range diffs {
		opType := ""
		switch diff.Type {
		case diffmatchpatch.DiffEqual:
			opType = "EQUAL"
		case diffmatchpatch.DiffInsert:
			opType = "INSERT"
		case diffmatchpatch.DiffDelete:
			opType = "DELETE"
		}
		t.Logf("  Diff %d: %s '%s'", i, opType, diff.Text)
	}
}

// TestPatchManager_PrettyPrintDiff tests pretty printing of diffs.
func TestPatchManager_PrettyPrintDiff(t *testing.T) {
	pm := NewPatchManager()

	oldText := "The quick brown fox"
	newText := "The quick red fox"

	diffs := pm.ComputeDiff(oldText, newText)
	pretty := pm.PrettyPrintDiff(diffs)

	if pretty == "" {
		t.Error("Expected pretty print to generate output")
	}

	t.Logf("Pretty printed diff:\n%s", pretty)
}

// TestPatchManager_CreateRollbackPatch tests creating rollback patches.
func TestPatchManager_CreateRollbackPatch(t *testing.T) {
	pm := NewPatchManager()

	original := "Hello World"
	modified := "Hello Beautiful World"

	// Create forward patch
	forwardPatch := pm.ComputePatch(original, modified)

	// Apply forward patch
	applied := pm.ApplyPatch(original, forwardPatch.Patch)
	if !applied.Success {
		t.Fatal("Failed to apply forward patch")
	}

	// Create rollback patch
	rollbackPatch := pm.CreateRollbackPatch(original, forwardPatch.Patch)
	if rollbackPatch == "" {
		t.Fatal("Failed to create rollback patch")
	}

	// Apply rollback patch to get back to original
	rolledBack := pm.ApplyPatch(applied.Content, rollbackPatch)
	if !rolledBack.Success {
		t.Error("Failed to apply rollback patch")
	}

	if rolledBack.Content != original {
		t.Errorf("Rollback failed: expected '%s', got '%s'", original, rolledBack.Content)
	}

	t.Logf("Successfully rolled back: '%s' -> '%s' -> '%s'",
		original, modified, rolledBack.Content)
}

// TestPatchManager_GetPatchStats tests getting patch statistics.
func TestPatchManager_GetPatchStats(t *testing.T) {
	pm := NewPatchManager()

	oldText := "Hello World"
	newText := "Hello Beautiful World"

	patchResult := pm.ComputePatch(oldText, newText)
	stats := pm.GetPatchStats(patchResult.Patch)

	if stats.TotalDiffs == 0 {
		t.Error("Expected at least one patch operation")
	}

	t.Logf("Patch stats: TotalDiffs=%d", stats.TotalDiffs)
}

// TestRedisHistoryService_ReconstructSnapshot tests reconstructing snapshots from patches.
func TestRedisHistoryService_ReconstructSnapshot(t *testing.T) {
	miniRedis := NewMiniRedis()
	historyService := NewRedisHistoryServiceWithOpts(miniRedis, true) // Use patch mode
	defer historyService.Close()

	sessionID := "test-session-reconstruct"

	// Create multiple versions with patches
	versions := []struct {
		version  int64
		content  string
	}{
		{0, "Hello"},           // First snapshot (full content)
		{1, "Hello World"},     // Patch 1
		{2, "Hello Beautiful World"}, // Patch 2
		{3, "Hello Beautiful World!"}, // Patch 3
	}

	// Store all versions
	for _, v := range versions {
		event := &HistoryEvent{
			SessionID:  sessionID,
			EventType:  "snapshot",
			VersionID:  v.version,
			Content:    v.content,
			Operations: []interface{}{},
			CreatedAt:  1234567890 + v.version,
			CreatedBy:  "test-client",
		}

		// Directly store snapshot (bypass async event processing for testing)
		historyService.storeSnapshotWithPatch(event)
	}

	// Test reconstruction for each version
	for _, targetVersion := range versions {
		reconstructed, err := historyService.ReconstructSnapshot(context.Background(), sessionID, targetVersion.version)
		if err != nil {
			t.Errorf("Failed to reconstruct version %d: %v", targetVersion.version, err)
			continue
		}

		if reconstructed != targetVersion.content {
			t.Errorf("Version %d: expected '%s', got '%s'",
				targetVersion.version, targetVersion.content, reconstructed)
		} else {
			t.Logf("Successfully reconstructed version %d: '%s'", targetVersion.version, reconstructed)
		}
	}
}

// TestRedisHistoryService_PatchModeCompression tests compression ratio in patch mode.
func TestRedisHistoryService_PatchModeCompression(t *testing.T) {
	miniRedis := NewMiniRedis()
	historyService := NewRedisHistoryServiceWithOpts(miniRedis, true) // Use patch mode
	defer historyService.Close()

	sessionID := "test-compression"

	// Create a large document and make small incremental changes
	baseContent := ""
	for i := 0; i < 100; i++ {
		baseContent += "This is line " + string(rune('0'+i%10)) + ". "
	}

	// Store initial version
	event0 := &HistoryEvent{
		SessionID:  sessionID,
		EventType:  "snapshot",
		VersionID:  0,
		Content:    baseContent,
		Operations: []interface{}{},
		CreatedAt:  1234567890,
		CreatedBy:  "test-client",
	}
	historyService.storeSnapshotWithPatch(event0)

	// Make small incremental changes
	totalContentSize := len(baseContent)
	totalPatchSize := 0

	for i := 1; i <= 10; i++ {
		// Small change: replace one character
		modifiedContent := baseContent[:100] + "X" + baseContent[101:]

		event := &HistoryEvent{
			SessionID:  sessionID,
			EventType:  "snapshot",
			VersionID:  int64(i),
			Content:    modifiedContent,
			Operations: []interface{}{},
			CreatedAt:  1234567890 + int64(i),
			CreatedBy:  "test-client",
		}
		historyService.storeSnapshotWithPatch(event)

		// Get patch size
		snapshotKey := "snapshot:" + sessionID + ":" + string(rune('0'+i))
		data, _ := miniRedis.Get(snapshotKey)

		totalContentSize += len(modifiedContent)
		totalPatchSize += len(data)
	}

	compressionRatio := float64(totalPatchSize) / float64(totalContentSize) * 100

	t.Logf("Content size: %d bytes", totalContentSize)
	t.Logf("Patch size: %d bytes", totalPatchSize)
	t.Logf("Compression ratio: %.2f%% (lower is better)", compressionRatio)

	if compressionRatio > 50 {
		t.Logf("Warning: Compression ratio is %.2f%%, expected < 50%%", compressionRatio)
	}
}

// BenchmarkPatchManager_ComputePatch benchmarks patch computation.
func BenchmarkPatchManager_ComputePatch(b *testing.B) {
	pm := NewPatchManager()

	oldText := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
		"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	newText := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
		"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
		"Ut enim ad minim veniam, quis nostrud exercitation ullamco."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.ComputePatch(oldText, newText)
	}
}

// BenchmarkPatchManager_ApplyPatch benchmarks patch application.
func BenchmarkPatchManager_ApplyPatch(b *testing.B) {
	pm := NewPatchManager()

	oldText := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
		"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	newText := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
		"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
		"Ut enim ad minim veniam, quis nostrud exercitation ullamco."

	patchResult := pm.ComputePatch(oldText, newText)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.ApplyPatch(oldText, patchResult.Patch)
	}
}
