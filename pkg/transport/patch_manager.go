package transport

import (
	"github.com/sergi/go-diff/diffmatchpatch"
)

// PatchManager handles diff-match-patch operations for version history.
// Uses Google's diff-match-patch algorithm for efficient text patching.
type PatchManager struct {
	dmp *diffmatchpatch.DiffMatchPatch
}

// NewPatchManager creates a new patch manager.
func NewPatchManager() *PatchManager {
	return &PatchManager{
		dmp: diffmatchpatch.New(),
	}
}

// PatchResult represents the result of computing a patch between two texts.
type PatchResult struct {
	Patch      string // Patch in text format (compact)
	PatchSize  int    // Size of patch in bytes
	OldSize    int    // Size of old text in bytes
	NewSize    int    // Size of new text in bytes
	SavedBytes int    // Bytes saved by using patch instead of full content
}

// ApplyPatchResult represents the result of applying a patch.
type ApplyPatchResult struct {
	Content     string // Reconstructed content
	Success     bool   // Whether patch application succeeded
	PatchesApplied int  // Number of patches applied
}

// ComputePatch computes a patch from oldText to newText.
// Returns a PatchResult containing the patch in compact text format.
func (pm *PatchManager) ComputePatch(oldText, newText string) *PatchResult {
	// 1. Compute diffs between the texts
	// The third parameter (timeout) is set to 0 for no timeout
	diffs := pm.dmp.DiffMain(oldText, newText, false)

	// 2. Create patch from diffs
	patch := pm.dmp.PatchMake(oldText, diffs)

	// 3. Convert to text format for compact storage
	patchText := pm.dmp.PatchToText(patch)

	oldSize := len(oldText)
	newSize := len(newText)
	patchSize := len(patchText)

	return &PatchResult{
		Patch:      patchText,
		PatchSize:  patchSize,
		OldSize:    oldSize,
		NewSize:    newSize,
		SavedBytes: newSize - patchSize,
	}
}

// ApplyPatch applies a patch to oldText to reconstruct newText.
// Returns ApplyPatchResult with the reconstructed content.
func (pm *PatchManager) ApplyPatch(oldText, patchText string) *ApplyPatchResult {
	// 1. Parse patch from text format
	patches, _ := pm.dmp.PatchFromText(patchText)

	// 2. Apply patches to old text
	newText, applied := pm.dmp.PatchApply(patches, oldText)

	// Count successfully applied patches
	appliedCount := 0
	for _, success := range applied {
		if success {
			appliedCount++
		}
	}

	// Check if all patches were applied successfully
	allSuccess := appliedCount == len(applied)

	return &ApplyPatchResult{
		Content:        newText,
		Success:        allSuccess,
		PatchesApplied: appliedCount,
	}
}

// ComputeDiff computes differences between two texts (for debugging/analysis).
// Returns a list of Diff objects: Equal, Insert, Delete.
func (pm *PatchManager) ComputeDiff(oldText, newText string) []diffmatchpatch.Diff {
	return pm.dmp.DiffMain(oldText, newText, false)
}

// ComputeDiffCleanup computes diffs with cleanup for more readable output.
// The cleanup parameter merges nearby diffs for more compact representation.
func (pm *PatchManager) ComputeDiffCleanup(oldText, newText string, cleanup bool) []diffmatchpatch.Diff {
	diffs := pm.dmp.DiffMain(oldText, newText, cleanup)
	if cleanup {
		// DiffCleanupMerge merges adjacent diffs of the same type
		diffs = pm.dmp.DiffCleanupMerge(diffs)
		// DiffCleanupSemanticLossless makes diffs more human-readable
		diffs = pm.dmp.DiffCleanupSemanticLossless(diffs)
	}
	return diffs
}

// PrettyPrintDiff converts diffs to a human-readable string.
func (pm *PatchManager) PrettyPrintDiff(diffs []diffmatchpatch.Diff) string {
	return pm.dmp.DiffPrettyText(diffs)
}

// CreateRollbackPatch creates a patch that can rollback a change.
// Given original text and the applied patch, creates a reverse patch.
func (pm *PatchManager) CreateRollbackPatch(originalText, appliedPatchText string) string {
	// Apply the patch to get the modified text
	result := pm.ApplyPatch(originalText, appliedPatchText)
	if !result.Success {
		return ""
	}

	// Create a reverse patch from modified back to original
	reversePatch := pm.ComputePatch(result.Content, originalText)
	return reversePatch.Patch
}

// PatchStats returns statistics about a patch.
type PatchStats struct {
	TotalDiffs int // Total number of patch operations
}

// GetPatchStats analyzes a patch and returns statistics.
// Note: This computes stats from the patch text length as a simple metric.
// For detailed diff statistics, use ComputeDiff directly.
func (pm *PatchManager) GetPatchStats(patchText string) *PatchStats {
	patches, _ := pm.dmp.PatchFromText(patchText)

	stats := &PatchStats{
		TotalDiffs: len(patches),
	}

	// Count patches (each patch represents a change region)
	// For detailed statistics, apply the patch and analyze diffs
	return stats
}
