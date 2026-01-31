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

	// Slice returns a substring from start to end (character positions relative to this node).
	Slice(start, end int) string

	// IsLeaf reports whether this is a leaf node (contains text).
	IsLeaf() bool
}

// LeafNode stores actual text content.
type LeafNode struct {
	text string
}

// InternalNode is an internal node in the rope tree that maintains balance and caches subtree info.
type InternalNode struct {
	left   RopeNode
	right  RopeNode
	length int // Cached: total characters in left subtree
	size   int // Cached: total bytes in left subtree
}

// ========== RopeNode Implementations ==========

func (n *LeafNode) Length() int {
	return utf8.RuneCountInString(n.text)
}

func (n *LeafNode) Size() int {
	return len(n.text)
}

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

func (n *LeafNode) IsLeaf() bool {
	return true
}

func (n *InternalNode) Length() int {
	return n.length + n.right.Length()
}

func (n *InternalNode) Size() int {
	return n.size + n.right.Size()
}

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

func (n *InternalNode) IsLeaf() bool {
	return false
}

// ========== Rope Constructors ==========

// New creates a Rope from the given string.
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

// Empty returns an empty Rope.
func Empty() *Rope {
	return &Rope{
		root:   &LeafNode{text: ""},
		length: 0,
		size:   0,
	}
}

// ========== Query Operations ==========

// Length returns the number of characters (Unicode code points) in the rope.
func (r *Rope) Length() int {
	if r == nil {
		return 0
	}
	return r.length
}

// Size returns the number of bytes in the rope.
func (r *Rope) Size() int {
	if r == nil {
		return 0
	}
	return r.size
}

// String returns the complete content as a string.
// Uses optimized byte slice building for minimal allocations.
func (r *Rope) String() string {
	if r == nil || r.length == 0 {
		return ""
	}

	// Pre-allocate with exact size and build using byte slice
	// This is faster than strings.Builder for this use case
	result := make([]byte, 0, r.size)

	it := r.Chunks()
	for it.Next() {
		result = append(result, it.Current()...)
	}

	return string(result)
}

// Bytes returns the complete content as a byte slice.
func (r *Rope) Bytes() []byte {
	return []byte(r.String())
}

// Slice returns a substring from start to end (exclusive, in character positions).
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

// CharAt returns the rune at the given character position.
// Panics if position is out of bounds.
func (r *Rope) CharAt(pos int) rune {
	if pos < 0 || pos >= r.length {
		panic("character position out of range")
	}
	// Use optimized iterator instead of []rune conversion
	it := r.IteratorAt(pos)
	it.Next() // Advance to the target position
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

// splitNode splits a node at a character position, returning (left, right).
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

// Insert inserts text at the given character position and returns a new Rope.
// The original Rope is unchanged.
// Panics if position is out of bounds.
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

// Delete removes characters from start to end (exclusive) and returns a new Rope.
// The original Rope is unchanged.
// Panics if range is out of bounds.
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

// Replace replaces characters from start to end (exclusive) with text and returns a new Rope.
// The original Rope is unchanged.
func (r *Rope) Replace(start, end int, text string) *Rope {
	return r.Delete(start, end).Insert(start, text)
}

// Split splits the rope at the given character position.
// Returns (left, right) where left contains [0, pos) and right contains [pos, end).
// Panics if position is out of bounds.
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

// Concat concatenates two ropes and returns a new Rope.
// The original Ropes are unchanged.
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

// Clone returns the rope itself (ropes are immutable, no copy needed).
func (r *Rope) Clone() *Rope {
	return r
}

// Runes returns all runes in the rope as a slice.
func (r *Rope) Runes() []rune {
	if r == nil || r.length == 0 {
		return []rune{}
	}

	it := r.NewIterator()
	runes := make([]rune, 0, r.length)
	for it.Next() {
		runes = append(runes, it.Current())
	}
	return runes
}

// ForEach calls the given function for each rune in the rope.
func (r *Rope) ForEach(f func(rune)) {
	if r == nil || r.length == 0 {
		return
	}

	it := r.NewIterator()
	for it.Next() {
		f(it.Current())
	}
}

// ForEachWithIndex calls the given function for each rune with its index.
func (r *Rope) ForEachWithIndex(f func(int, rune)) {
	if r == nil || r.length == 0 {
		return
	}

	it := r.NewIterator()
	for it.Next() {
		// Position() returns charPos + 1 (next position)
		// We want the current element's index, so subtract 1
		f(it.Position() - 1, it.Current())
	}
}

// Map creates a new rope by applying the given function to each rune.
func (r *Rope) Map(f func(rune) rune) *Rope {
	if r == nil || r.length == 0 {
		return r
	}

	result := make([]rune, 0, r.length)
	it := r.NewIterator()
	for it.Next() {
		result = append(result, f(it.Current()))
	}
	return New(string(result))
}

// Filter creates a new rope containing only runes for which f returns true.
func (r *Rope) Filter(f func(rune) bool) *Rope {
	if r == nil || r.length == 0 {
		return r
	}

	result := make([]rune, 0, r.length)
	it := r.NewIterator()
	for it.Next() {
		ch := it.Current()
		if f(ch) {
			result = append(result, ch)
		}
	}
	return New(string(result))
}

// Count returns the number of runes for which f returns true.
func (r *Rope) Count(f func(rune) bool) int {
	if r == nil || r.length == 0 {
		return 0
	}

	count := 0
	it := r.NewIterator()
	for it.Next() {
		if f(it.Current()) {
			count++
		}
	}
	return count
}

// ========== Utility Functions ==========

// Lines splits the rope into lines, preserving line endings.
func (r *Rope) Lines() []string {
	content := r.String()
	return strings.SplitAfter(content, "\n")
}

// Contains reports whether the rope contains the given substring.
func (r *Rope) Contains(substring string) bool {
	return strings.Contains(r.String(), substring)
}

// Index returns the first character position of substring, or -1 if not found.
func (r *Rope) Index(substring string) int {
	// Convert byte index to character index
	byteIdx := strings.Index(r.String(), substring)
	if byteIdx < 0 {
		return -1
	}
	return utf8.RuneCountInString(r.String()[:byteIdx])
}

// LastIndex returns the last character position of substring, or -1 if not found.
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

// Equals reports whether two ropes have identical content.
func (r *Rope) Equals(other *Rope) bool {
	return r.String() == other.String()
}
