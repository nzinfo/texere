package rope

import (
	"fmt"
)

// ========== Chunk Operations ==========

// ChunkInfo contains information about a chunk.
type ChunkInfo struct {
	ByteIdx   int // Byte index of the chunk's start
	CharIdx   int // Character index of the chunk's start
	LineIdx   int // Line index of the chunk's start
	ByteLen   int // Length in bytes
	CharLen   int // Length in characters
	Text      string // The chunk text
	IsEmpty   bool // Whether the chunk is empty
}

// ChunkAtChar returns the chunk containing the given character position.
// Returns the chunk info and the character index where the chunk starts.
func (r *Rope) ChunkAtChar(charIdx int) (ChunkInfo, int) {
	if r == nil || r.root == nil {
		return ChunkInfo{}, 0
	}

	if charIdx < 0 || charIdx > r.Length() {
		panic(fmt.Sprintf("character index %d out of bounds (len: %d)", charIdx, r.Length()))
	}

	// Find the chunk containing charIdx
	byteIdx := r.charToByte(charIdx)
	chunkStart, chunkText := findChunkAtByte(r.root, byteIdx)

	// Calculate indices
	startCharIdx := r.byteToChar(chunkStart)
	startLineIdx := r.LineAtChar(startCharIdx)

	return ChunkInfo{
		ByteIdx:   chunkStart,
		CharIdx:   startCharIdx,
		LineIdx:   startLineIdx,
		ByteLen:   len(chunkText),
		CharLen:   runeCount(chunkText),
		Text:      chunkText,
		IsEmpty:   len(chunkText) == 0,
	}, startCharIdx
}

// ChunkAtByte returns the chunk containing the given byte position.
// Returns the chunk info and the byte index where the chunk starts.
func (r *Rope) ChunkAtByte(byteIdx int) (ChunkInfo, int) {
	if r == nil || r.root == nil {
		return ChunkInfo{}, 0
	}

	if byteIdx < 0 || byteIdx > r.Size() {
		panic(fmt.Sprintf("byte index %d out of bounds (size: %d)", byteIdx, r.Size()))
	}

	chunkStart, chunkText := findChunkAtByte(r.root, byteIdx)

	// Calculate indices
	startCharIdx := r.byteToChar(chunkStart)
	startLineIdx := 0
	if startCharIdx > 0 {
		startLineIdx = r.LineAtChar(startCharIdx)
	}

	return ChunkInfo{
		ByteIdx:   chunkStart,
		CharIdx:   startCharIdx,
		LineIdx:   startLineIdx,
		ByteLen:   len(chunkText),
		CharLen:   runeCount(chunkText),
		Text:      chunkText,
		IsEmpty:   len(chunkText) == 0,
	}, chunkStart
}

// findChunkAtByte finds the leaf chunk containing the given byte index.
// Returns the byte index where the chunk starts and the chunk text.
func findChunkAtByte(n RopeNode, byteIdx int) (int, string) {
	switch node := n.(type) {
	case *LeafNode:
		return 0, node.text
	case *InternalNode:
		leftSize := node.left.Size()
		if byteIdx < leftSize {
			return findChunkAtByte(node.left, byteIdx)
		}
		offset, text := findChunkAtByte(node.right, byteIdx-leftSize)
		return leftSize + offset, text
	}
	return 0, ""
}

// Chunks creates an iterator over the rope's chunks.
func (r *Rope) Chunks() *ChunksIterator {
	if r == nil || r.root == nil {
		return &ChunksIterator{rope: r}
	}
	return &ChunksIterator{
		rope:       r,
		chunkInfos: r.collectChunkInfos(),
		index:      -1,
	}
}

// collectChunkInfos recursively collects all chunk information.
func (r *Rope) collectChunkInfos() []ChunkInfo {
	if r == nil || r.root == nil {
		return []ChunkInfo{}
	}

	var infos []ChunkInfo
	byteIdx := 0
	charIdx := 0
	lineIdx := 0

	collectChunks(r.root, &infos, &byteIdx, &charIdx, &lineIdx)
	return infos
}

// collectChunks recursively collects chunks.
func collectChunks(n RopeNode, infos *[]ChunkInfo, byteIdx, charIdx, lineIdx *int) {
	switch node := n.(type) {
	case *LeafNode:
		textLen := len(node.text)
		charCount := runeCount(node.text)

		// Skip empty leaf nodes to match ropey behavior
		// Empty rope should have 0 chunks, not 1 empty chunk
		if textLen > 0 {
			*infos = append(*infos, ChunkInfo{
				ByteIdx:   *byteIdx,
				CharIdx:   *charIdx,
				LineIdx:   *lineIdx,
				ByteLen:   textLen,
				CharLen:   charCount,
				Text:      node.text,
				IsEmpty:   textLen == 0,
			})
		}

		*byteIdx += textLen
		*charIdx += charCount
		// Update line count
		for _, ch := range node.text {
			if ch == '\n' {
				*lineIdx++
			}
		}

	case *InternalNode:
		collectChunks(node.left, infos, byteIdx, charIdx, lineIdx)
		collectChunks(node.right, infos, byteIdx, charIdx, lineIdx)
	}
}

// ========== Chunks Iterator ==========

// ChunksIterator iterates over the chunks of a rope.
type ChunksIterator struct {
	rope       *Rope
	chunkInfos []ChunkInfo
	index      int
}

// NewChunksIterator creates a new chunks iterator.
func (r *Rope) NewChunksIterator() *ChunksIterator {
	return r.Chunks()
}

// Next advances to the next chunk and returns true if there are more chunks.
func (it *ChunksIterator) Next() bool {
	if it.chunkInfos == nil {
		it.chunkInfos = it.rope.collectChunkInfos()
	}
	it.index++
	return it.index < len(it.chunkInfos)
}

// Current returns the current chunk text.
func (it *ChunksIterator) Current() string {
	if it.index < 0 || it.index >= len(it.chunkInfos) {
		panic("iterator out of bounds")
	}
	return it.chunkInfos[it.index].Text
}

// CurrentInfo returns the current chunk info.
func (it *ChunksIterator) CurrentInfo() ChunkInfo {
	if it.index < 0 || it.index >= len(it.chunkInfos) {
		panic("iterator out of bounds")
	}
	return it.chunkInfos[it.index]
}

// Info returns the current chunk's information.
func (it *ChunksIterator) Info() ChunkInfo {
	return it.CurrentInfo()
}

// Position returns the current iterator position.
func (it *ChunksIterator) Position() int {
	return it.index
}

// Count returns the total number of chunks.
func (it *ChunksIterator) Count() int {
	return len(it.chunkInfos)
}

// Reset resets the iterator to the beginning.
func (it *ChunksIterator) Reset() {
	it.index = -1
}

// ToSlice collects all chunks into a slice.
func (it *ChunksIterator) ToSlice() []string {
	it.Reset()
	chunks := make([]string, 0, it.Count())
	for it.Next() {
		chunks = append(chunks, it.Current())
	}
	return chunks
}

// ToInfoSlice collects all chunk infos into a slice.
func (it *ChunksIterator) ToInfoSlice() []ChunkInfo {
	it.Reset()
	infos := make([]ChunkInfo, 0, it.Count())
	for it.Next() {
		infos = append(infos, it.CurrentInfo())
	}
	return infos
}

// ========== Chunk Utilities ==========

// ChunkCount returns the total number of chunks in the rope.
func (r *Rope) ChunkCount() int {
	return r.LeafCount()
}

// AverageChunkSize returns the average chunk size in bytes.
func (r *Rope) AverageChunkSize() float64 {
	if r == nil {
		return 0
	}
	count := r.ChunkCount()
	if count == 0 {
		return 0
	}
	return float64(r.Size()) / float64(count)
}

// MaxChunkSize returns the size of the largest chunk in bytes.
func (r *Rope) MaxChunkSize() int {
	it := r.Chunks()
	maxSize := 0
	for it.Next() {
		size := len(it.Current())
		if size > maxSize {
			maxSize = size
		}
	}
	return maxSize
}

// MinChunkSize returns the size of the smallest chunk in bytes.
func (r *Rope) MinChunkSize() int {
	it := r.Chunks()
	if it.Count() == 0 {
		return 0
	}

	it.Next()
	minSize := len(it.Current())
	for it.Next() {
		size := len(it.Current())
		if size < minSize {
			minSize = size
		}
	}
	return minSize
}

// ========== Advanced Chunk Operations ==========

// ChunksAtChar creates an iterator starting at the chunk containing charIdx.
// Also returns the starting byte and character indices of that chunk.
func (r *Rope) ChunksAtChar(charIdx int) (*ChunksIterator, int, int, int) {
	if charIdx < 0 || charIdx > r.Length() {
		panic(fmt.Sprintf("character index %d out of bounds (len: %d)", charIdx, r.Length()))
	}

	if charIdx == r.Length() {
		// Return iterator at end
		it := r.Chunks()
		it.index = it.Count() // Move to end
		return it, r.Size(), r.Length(), r.LineCount()
	}

	info, startChar := r.ChunkAtChar(charIdx)
	it := r.Chunks()

	// Find the chunk index
	for i := 0; i < len(it.chunkInfos); i++ {
		if it.chunkInfos[i].CharIdx == startChar {
			it.index = i - 1 // Next() will advance to i
			break
		}
	}

	return it, info.ByteIdx, info.CharIdx, info.LineIdx
}

// ChunksAtByte creates an iterator starting at the chunk containing byteIdx.
// Also returns the starting byte and character indices of that chunk.
func (r *Rope) ChunksAtByte(byteIdx int) (*ChunksIterator, int, int, int) {
	if byteIdx < 0 || byteIdx > r.Size() {
		panic(fmt.Sprintf("byte index %d out of bounds (size: %d)", byteIdx, r.Size()))
	}

	if byteIdx == r.Size() {
		// Return iterator at end
		it := r.Chunks()
		it.index = it.Count() // Move to end
		return it, r.Size(), r.Length(), r.LineCount()
	}

	info, startByte := r.ChunkAtByte(byteIdx)
	it := r.Chunks()

	// Find the chunk index
	for i := 0; i < len(it.chunkInfos); i++ {
		if it.chunkInfos[i].ByteIdx == startByte {
			it.index = i - 1 // Next() will advance to i
			break
		}
	}

	return it, info.ByteIdx, info.CharIdx, info.LineIdx
}

// ChunksAtLine creates an iterator starting at the chunk containing lineIdx.
func (r *Rope) ChunksAtLine(lineIdx int) (*ChunksIterator, int, int, int) {
	if lineIdx < 0 || lineIdx > r.LineCount() {
		panic(fmt.Sprintf("line index %d out of bounds (lines: %d)", lineIdx, r.LineCount()))
	}

	if lineIdx == r.LineCount() {
		it := r.Chunks()
		it.index = it.Count()
		return it, r.Size(), r.Length(), r.LineCount()
	}

	// Find the chunk containing the start of this line
	it := r.Chunks()
	for i := 0; i < len(it.chunkInfos); i++ {
		if it.chunkInfos[i].LineIdx == lineIdx {
			it.index = i - 1
			info := it.chunkInfos[i]
			return it, info.ByteIdx, info.CharIdx, info.LineIdx
		}
	}

	// Should not reach here
	return it, 0, 0, 0
}

// ========== Helper Functions ==========

// runeCount returns the number of runes in a string.
func runeCount(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}
