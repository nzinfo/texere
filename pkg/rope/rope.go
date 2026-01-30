// Package rope implements an efficient Rope data structure for large text editing.
//
// A Rope is a balanced binary tree (B-tree) representation of a string,
// optimized for efficient insertions and deletions in large texts.
//
// Key properties:
// - Immutable: All operations return new Ropes, originals are unchanged
// - Efficient: O(log n) for insert/delete/slice operations
// - Memory: Minimal copying due to tree structure
//
// This implementation is based on:
// - "Ropes: an Alternative to Strings" by Boehm, Atkinson, and Plass (1995)
// - The ropey crate in Rust (used by Helix editor)
package rope

import (
	"strings"
	"unicode/utf8"
)

// Rope represents an immutable string as a balanced tree.
type Rope struct {
	root RopeNode
	// Cached values for O(1) access
	length int // Total characters (Unicode code points)
	size   int // Total bytes
}

// RopeNode is the interface for all rope nodes.
type RopeNode interface {
	// Length returns the number of characters in this subtree.
	Length() int

	// Size returns the number of bytes in this subtree.
	Size() int

	// Slice returns a substring from start to end (character positions).
	// start and end are relative to this node, not the entire rope.
	Slice(start, end int) string

	// IsLeaf returns true if this is a leaf node (contains text).
	IsLeaf() bool
}

// LeafNode stores actual text content.
type LeafNode struct {
	text string
}

// InternalNode represents an internal node in the rope tree.
// It maintains balance and caches subtree information.
type InternalNode struct {
	left   RopeNode
	right  RopeNode
	length int // Cached: total characters in left subtree
	size   int // Cached: total bytes in left subtree
}

// ========== RopeNode Implementations ==========

// Length returns the number of characters in this leaf.
func (n *LeafNode) Length() int {
	return utf8.RuneCountInString(n.text)
}

// Size returns the number of bytes in this leaf.
func (n *LeafNode) Size() int {
	return len(n.text)
}

// Slice returns a substring from this leaf.
func (n *LeafNode) Slice(start, end int) string {
	// Convert character positions to byte positions without []rune conversion
	byteStart := 0
	for i := 0; i < start; i++ {
		_, size := utf8.DecodeRuneInString(n.text[byteStart:])
		byteStart += size
	}

	byteEnd := byteStart
	for i := start; i < end; i++ {
		_, size := utf8.DecodeRuneInString(n.text[byteEnd:])
		byteEnd += size
	}

	return n.text[byteStart:byteEnd]
}

// IsLeaf returns true for leaf nodes.
func (n *LeafNode) IsLeaf() bool {
	return true
}

// Length returns the total characters in this subtree.
func (n *InternalNode) Length() int {
	return n.length + n.right.Length()
}

// Size returns the total bytes in this subtree.
func (n *InternalNode) Size() int {
	return n.size + n.right.Size()
}

// Slice returns a substring from this internal node.
func (n *InternalNode) Slice(start, end int) string {
	leftLen := n.left.Length()

	// Entirely in left subtree
	if end <= leftLen {
		return n.left.Slice(start, end)
	}

	// Entirely in right subtree
	if start >= leftLen {
		return n.right.Slice(start-leftLen, end-leftLen)
	}

	// Spans both subtrees
	leftPart := n.left.Slice(start, leftLen)
	rightPart := n.right.Slice(0, end-leftLen)
	return leftPart + rightPart
}

// IsLeaf returns false for internal nodes.
func (n *InternalNode) IsLeaf() bool {
	return false
}

// ========== Rope Constructors ==========

// New creates a new Rope from the given string.
func New(text string) *Rope {
	if text == "" {
		return Empty()
	}

	return &Rope{
		root:   &LeafNode{text: text},
		length: utf8.RuneCountInString(text),
		size:   len(text),
	}
}

// Empty creates an empty Rope.
func Empty() *Rope {
	return &Rope{
		root:   &LeafNode{text: ""},
		length: 0,
		size:   0,
	}
}

// ========== Basic Query Operations ==========

// Length returns the total number of characters (Unicode code points) in the rope.
func (r *Rope) Length() int {
	if r == nil {
		return 0
	}
	return r.length
}

// Size returns the total number of bytes in the rope.
func (r *Rope) Size() int {
	if r == nil {
		return 0
	}
	return r.size
}

// String returns the complete content of the rope as a string.
// Optimized implementation using strings.Builder for minimal allocations.
func (r *Rope) String() string {
	if r == nil || r.length == 0 {
		return ""
	}

	// Use strings.Builder with pre-allocated capacity
	var b strings.Builder
	b.Grow(r.size)

	// Iterate chunks directly for efficiency
	it := r.Chunks()
	for it.Next() {
		b.WriteString(it.Current())
	}

	return b.String()
}

// Bytes returns the complete content of the rope as a byte slice.
func (r *Rope) Bytes() []byte {
	return []byte(r.String())
}

// Slice returns a substring from start to end (exclusive).
// The indices are character positions (not byte positions).
// Panics if indices are out of bounds.
func (r *Rope) Slice(start, end int) string {
	if r == nil {
		return ""
	}
	if start < 0 || end > r.length || start > end {
		panic("slice bounds out of range")
	}
	if start == end {
		return ""
	}
	return r.root.Slice(start, end)
}

// CharAt returns the character (rune) at the given character position.
// Panics if position is out of bounds.
func (r *Rope) CharAt(pos int) rune {
	if pos < 0 || pos >= r.length {
		panic("character position out of range")
	}
	// Use optimized iterator instead of []rune conversion
	it := r.IteratorAt(pos)
	return it.Current()
}

// ByteAt returns the byte at the given byte position.
// Panics if position is out of bounds.
func (r *Rope) ByteAt(pos int) byte {
	if pos < 0 || pos >= r.size {
		panic("byte position out of range")
	}
	// Use optimized bytes iterator instead of Bytes()
	it := r.NewBytesIterator()
	it.Seek(pos)
	it.Next() // Move to the target position
	return it.Current()
}

// ========== Helper Functions ==========

// concatNodes concatenates two nodes and returns a new node.
// This is the low-level operation used by Concat.
func concatNodes(left, right RopeNode) RopeNode {
	// If one side is empty, return the other
	if left.Length() == 0 {
		return right
	}
	if right.Length() == 0 {
		return left
	}

	return &InternalNode{
		left:   left,
		right:  right,
		length: left.Length(),
		size:   left.Size(),
	}
}

// splitNode splits a node at a character position.
// Returns (leftNode, rightNode).
func splitNode(node RopeNode, pos int) (RopeNode, RopeNode) {
	if node.IsLeaf() {
		leaf := node.(*LeafNode)
		// Find byte position without []rune conversion
		splitByte := 0
		for i := 0; i < pos; i++ {
			_, size := utf8.DecodeRuneInString(leaf.text[splitByte:])
			splitByte += size
		}

		leftText := leaf.text[:splitByte]
		rightText := leaf.text[splitByte:]

		var left, right RopeNode
		if leftText != "" {
			left = &LeafNode{text: leftText}
		}
		if rightText != "" {
			right = &LeafNode{text: rightText}
		}

		return left, right
	}

	internal := node.(*InternalNode)
	leftLen := internal.left.Length()

	if pos <= leftLen {
		leftLeft, leftRight := splitNode(internal.left, pos)
		return leftLeft, concatNodes(leftRight, internal.right)
	}

	rightLeft, rightRight := splitNode(internal.right, pos-leftLen)
	return concatNodes(internal.left, rightLeft), rightRight
}

// insertNode inserts text at a character position in a node.
func insertNode(node RopeNode, pos int, text string) RopeNode {
	if node.Length() == 0 {
		return &LeafNode{text: text}
	}

	if node.IsLeaf() {
		leaf := node.(*LeafNode)
		// Find byte position without []rune conversion
		insertByte := 0
		for i := 0; i < pos; i++ {
			_, size := utf8.DecodeRuneInString(leaf.text[insertByte:])
			insertByte += size
		}

		leftPart := leaf.text[:insertByte]
		rightPart := leaf.text[insertByte:]

		return concatNodes(
			&LeafNode{text: leftPart + text},
			&LeafNode{text: rightPart},
		)
	}

	internal := node.(*InternalNode)
	leftLen := internal.left.Length()

	if pos <= leftLen {
		newLeft := insertNode(internal.left, pos, text)
		return &InternalNode{
			left:   newLeft,
			right:  internal.right,
			length: newLeft.Length(),
			size:   newLeft.Size(),
		}
	}

	newRight := insertNode(internal.right, pos-leftLen, text)
	return &InternalNode{
		left:   internal.left,
		right:  newRight,
		length: internal.left.Length(),
		size:   internal.left.Size(),
	}
}

// deleteNode deletes characters from start to end (exclusive) from a node.
func deleteNode(node RopeNode, start, end int) RopeNode {
	if node.Length() == 0 || start >= end {
		return node
	}

	if node.IsLeaf() {
		leaf := node.(*LeafNode)
		// Find byte positions without []rune conversion
		startByte := 0
		for i := 0; i < start; i++ {
			_, size := utf8.DecodeRuneInString(leaf.text[startByte:])
			startByte += size
		}

		endByte := startByte
		for i := start; i < end; i++ {
			_, size := utf8.DecodeRuneInString(leaf.text[endByte:])
			endByte += size
		}

		newText := leaf.text[:startByte] + leaf.text[endByte:]
		return &LeafNode{text: newText}
	}

	internal := node.(*InternalNode)
	leftLen := internal.left.Length()

	// Entirely in left subtree
	if end <= leftLen {
		newLeft := deleteNode(internal.left, start, end)
		return concatNodes(newLeft, internal.right)
	}

	// Entirely in right subtree
	if start >= leftLen {
		newRight := deleteNode(internal.right, start-leftLen, end-leftLen)
		return concatNodes(internal.left, newRight)
	}

	// Spans both subtrees
	newLeft := deleteNode(internal.left, start, leftLen)
	newRight := deleteNode(internal.right, 0, end-leftLen)
	return concatNodes(newLeft, newRight)
}

// ========== Modification Operations ==========

// Insert inserts text at the given character position.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) Insert(pos int, text string) *Rope {
	if pos < 0 || pos > r.length {
		panic("insert position out of range")
	}
	if text == "" {
		return r
	}

	newRoot := insertNode(r.root, pos, text)
	return &Rope{
		root:   newRoot,
		length: r.length + utf8.RuneCountInString(text),
		size:   r.size + len(text),
	}
}

// Delete removes characters from start to end (exclusive).
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) Delete(start, end int) *Rope {
	if start < 0 || end > r.length || start > end {
		panic("delete range out of bounds")
	}
	if start == end {
		return r
	}

	deletedLength := utf8.RuneCountInString(r.Slice(start, end))
	deletedSize := len(r.Slice(start, end))

	newRoot := deleteNode(r.root, start, end)
	return &Rope{
		root:   newRoot,
		length: r.length - deletedLength,
		size:   r.size - deletedSize,
	}
}

// Replace replaces characters from start to end (exclusive) with the given text.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) Replace(start, end int, text string) *Rope {
	return r.Delete(start, end).Insert(start, text)
}

// Split splits the rope at the given character position.
// Returns (left, right) where left contains [0, pos) and right contains [pos, end).
func (r *Rope) Split(pos int) (*Rope, *Rope) {
	if pos < 0 || pos > r.length {
		panic("split position out of range")
	}
	if pos == 0 {
		return Empty(), r
	}
	if pos == r.length {
		return r, Empty()
	}

	leftRoot, rightRoot := splitNode(r.root, pos)

	left := &Rope{
		root:   leftRoot,
		length: pos,
		size:   0, // Will be recalculated
	}
	left.size = left.root.Size()

	right := &Rope{
		root:   rightRoot,
		length: r.length - pos,
		size:   0, // Will be recalculated
	}
	right.size = right.root.Size()

	return left, right
}

// Concat concatenates two ropes.
// Returns a new Rope, leaving both originals unchanged.
func (r *Rope) Concat(other *Rope) *Rope {
	if r == nil || r.length == 0 {
		return other
	}
	if other == nil || other.length == 0 {
		return r
	}

	newRoot := concatNodes(r.root, other.root)
	return &Rope{
		root:   newRoot,
		length: r.length + other.length,
		size:   r.size + other.size,
	}
}

// Clone creates a shallow copy of the rope.
// Since ropes are immutable, this returns the same instance.
func (r *Rope) Clone() *Rope {
	return r
}

// ========== Utility Functions ==========

// Lines splits the rope into lines, preserving line endings.
func (r *Rope) Lines() []string {
	content := r.String()
	return strings.SplitAfter(content, "\n")
}

// Contains returns true if the rope contains the given substring.
func (r *Rope) Contains(substring string) bool {
	return strings.Contains(r.String(), substring)
}

// Index returns the first character position of the given substring,
// or -1 if not found.
func (r *Rope) Index(substring string) int {
	// Convert byte index to character index
	byteIdx := strings.Index(r.String(), substring)
	if byteIdx < 0 {
		return -1
	}
	return utf8.RuneCountInString(r.String()[:byteIdx])
}

// LastIndex returns the last character position of the given substring,
// or -1 if not found.
func (r *Rope) LastIndex(substring string) int {
	// Convert byte index to character index
	byteIdx := strings.LastIndex(r.String(), substring)
	if byteIdx < 0 {
		return -1
	}
	return utf8.RuneCountInString(r.String()[:byteIdx])
}

// Compare compares two ropes lexicographically.
// Returns -1 if r < other, 0 if r == other, 1 if r > other.
func (r *Rope) Compare(other *Rope) int {
	return strings.Compare(r.String(), other.String())
}

// Equals returns true if two ropes have identical content.
func (r *Rope) Equals(other *Rope) bool {
	return r.String() == other.String()
}
