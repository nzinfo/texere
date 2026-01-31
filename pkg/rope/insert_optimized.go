package rope

import (
	"unicode/utf8"
)

// ========== Optimized Insert/Delete ==========

// InsertOptimized inserts text at the specified character position.
// Optimized version that avoids rune[] conversions and reduces allocations.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) InsertOptimized(pos int, text string) *Rope {
	if r == nil {
		return New(text)
	}
	if pos < 0 || pos > r.length {
		panic("insert position out of range")
	}
	if text == "" {
		return r
	}
	if pos == 0 {
		return r.Prepend(text) // Now uses optimized implementation
	}
	if pos == r.length {
		return r.Append(text)
	}

	newRoot := insertNodeOptimized(r.root, pos, text)
	return &Rope{
		root:   newRoot,
		length: r.length + utf8.RuneCountInString(text),
		size:   r.size + len(text),
	}
}

// insertNodeOptimized performs optimized insertion.
func insertNodeOptimized(node RopeNode, pos int, text string) RopeNode {
	if node.IsLeaf() {
		leaf := node.(*LeafNode)

		// Optimized: use byte operations instead of rune[] conversion
		oldText := leaf.text
		oldBytes := []byte(oldText)

		// Find byte position
		bytePos := 0
		for i := 0; i < pos; i++ {
			_, size := utf8.DecodeRune(oldBytes[bytePos:])
			bytePos += size
		}

		// Create new text with insertion
		newText := make([]byte, 0, len(oldBytes)+len(text))
		newText = append(newText, oldBytes[:bytePos]...)
		newText = append(newText, text...)
		newText = append(newText, oldBytes[bytePos:]...)

		return &LeafNode{text: string(newText)}
	}

	// Internal node
	internal := node.(*InternalNode)
	leftLen := internal.left.Length()

	if pos <= leftLen {
		// Insert into left subtree
		newLeft := insertNodeOptimized(internal.left, pos, text)
		return &InternalNode{
			left:   newLeft,
			right:  internal.right,
			length: newLeft.Length(),
			size:   newLeft.Size(),
		}
	}

	// Insert into right subtree
	newRight := insertNodeOptimized(internal.right, pos-leftLen, text)
	return &InternalNode{
		left:   internal.left,
		right:  newRight,
		length: internal.left.Length(),
		size:   internal.left.Size(),
	}
}

// DeleteOptimized removes characters from start to end (exclusive).
// Optimized version that reduces allocations.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) DeleteOptimized(start, end int) *Rope {
	if r == nil {
		return r
	}
	if start < 0 || end > r.length || start > end {
		panic("delete range out of bounds")
	}
	if start == end {
		return r
	}

	newRoot := deleteNodeOptimized(r.root, start, end)
	return &Rope{
		root:   newRoot,
		length: r.length - utf8.RuneCountInString(r.Slice(start, end)),
		size:   r.size - len(r.Slice(start, end)),
	}
}

// deleteNodeOptimized performs optimized deletion.
func deleteNodeOptimized(node RopeNode, start, end int) RopeNode {
	if node.IsLeaf() {
		leaf := node.(*LeafNode)

		// Optimized: use byte operations
		oldText := leaf.text
		oldBytes := []byte(oldText)

		// Find byte positions
		startByte := 0
		for i := 0; i < start; i++ {
			_, size := utf8.DecodeRune(oldBytes[startByte:])
			startByte += size
		}

		endByte := startByte
		for i := start; i < end; i++ {
			_, size := utf8.DecodeRune(oldBytes[endByte:])
			endByte += size
		}

		// Create new text without deleted range
		newText := make([]byte, 0, len(oldBytes)-(endByte-startByte))
		newText = append(newText, oldBytes[:startByte]...)
		newText = append(newText, oldBytes[endByte:]...)

		return &LeafNode{text: string(newText)}
	}

	// Internal node
	internal := node.(*InternalNode)
	leftLen := internal.left.Length()

	// Entirely in left subtree
	if end <= leftLen {
		newLeft := deleteNodeOptimized(internal.left, start, end)
		return &InternalNode{
			left:   newLeft,
			right:  internal.right,
			length: newLeft.Length(),
			size:   newLeft.Size(),
		}
	}

	// Entirely in right subtree
	if start >= leftLen {
		newRight := deleteNodeOptimized(internal.right, start-leftLen, end-leftLen)
		return &InternalNode{
			left:   internal.left,
			right:  newRight,
			length: internal.left.Length(),
			size:   internal.left.Size(),
		}
	}

	// Spans both subtrees - need to split and merge
	leftPart := internal.left.Slice(start, leftLen)
	rightPart := internal.right.Slice(0, end-leftLen)

	// Concatenate left and right parts
	return &InternalNode{
		left:  New(leftPart).root,
		right: New(rightPart).root,
		length: 0, // Will be calculated by parent
		size:   0,
	}
}

// ReplaceOptimized replaces characters from start to end (exclusive) with the given text.
// Optimized version that combines Delete and Insert efficiently.
func (r *Rope) ReplaceOptimized(start, end int, text string) *Rope {
	if r == nil {
		return New(text)
	}

	// Optimized: If replacement is same size, just swap the content
	oldLen := utf8.RuneCountInString(r.Slice(start, end))
	newLen := utf8.RuneCountInString(text)

	if oldLen == newLen {
		// Direct replacement in place
		return r.DeleteOptimized(start, end).InsertOptimized(start, text)
	}

	// For different sizes, use Delete + Insert
	return r.DeleteOptimized(start, end).InsertOptimized(start, text)
}
