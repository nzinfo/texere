package rope

import (
	"sync/atomic"
	"unicode/utf8"
)

// ========== Copy-on-Write Optimization ==========

// CowNode is a wrapper for copy-on-write support.
type CowNode struct {
	node       RopeNode
	refCount   int32 // Reference count for sharing
	isShared   bool  // Whether this node is shared
}

// NewCowNode creates a new COW node.
func NewCowNode(node RopeNode) *CowNode {
	return &CowNode{
		node:     node,
		refCount: 1,
		isShared: false,
	}
}

// Retain increases reference count.
func (n *CowNode) Retain() {
	if n != nil {
		atomic.AddInt32(&n.refCount, 1)
		n.isShared = n.refCount > 1
	}
}

// Release decreases reference count.
func (n *CowNode) Release() {
	if n != nil {
		atomic.AddInt32(&n.refCount, -1)
	}
}

// IsShared returns true if this node is shared.
func (n *CowNode) IsShared() bool {
	return n != nil && n.isShared
}

// CloneIfNeeded returns a cloned node if shared, otherwise returns the node itself.
func (n *CowNode) CloneIfNeeded() RopeNode {
	if n == nil {
		return nil
	}
	if n.IsShared() {
		return cloneNode(n.node)
	}
	return n.node
}

// cloneNode deep clones a node for copy-on-write.
func cloneNode(node RopeNode) RopeNode {
	if node == nil {
		return nil
	}

	if node.IsLeaf() {
		leaf := node.(*LeafNode)
		// Leaf nodes are immutable, so we can share them
		return leaf
	}

	internal := node.(*InternalNode)
	// Recursively clone children
	return &InternalNode{
		left:   cloneNode(internal.left),
		right:  cloneNode(internal.right),
		length: internal.length,
		size:   internal.size,
	}
}

// ========== Optimized Rope with COW ==========

// CowRope is a rope with copy-on-write optimization.
type CowRope struct {
	root   *CowNode
	length int
	size   int
}

// NewCowRope creates a new COW rope.
func NewCowRope(text string) *CowRope {
	node := &LeafNode{text: text}
	return &CowRope{
		root:   NewCowNode(node),
		length: utf8.RuneCountInString(text),
		size:   len(text),
	}
}

// Length returns the total characters.
func (r *CowRope) Length() int {
	if r == nil {
		return 0
	}
	return r.length
}

// Size returns the total bytes.
func (r *CowRope) Size() int {
	if r == nil {
		return 0
	}
	return r.size
}

// String returns the content as string.
func (r *CowRope) String() string {
	if r == nil || r.length == 0 {
		return ""
	}
	return nodeToString(r.root.node)
}

// nodeToString converts node to string.
func nodeToString(node RopeNode) string {
	if node == nil {
		return ""
	}

	if node.IsLeaf() {
		return node.(*LeafNode).text
	}

	internal := node.(*InternalNode)
	return nodeToString(internal.left) + nodeToString(internal.right)
}

// Insert inserts text at position with COW optimization.
func (r *CowRope) Insert(pos int, text string) *CowRope {
	if r == nil {
		return NewCowRope(text)
	}
	if pos < 0 || pos > r.length {
		panic("insert position out of range")
	}
	if text == "" {
		return r
	}

	newRoot := cowInsert(r.root.CloneIfNeeded(), pos, text)
	result := &CowRope{
		root:   NewCowNode(newRoot),
		length: r.length + utf8.RuneCountInString(text),
		size:   r.size + len(text),
	}
	r.root.Retain() // Retain old root for sharing
	return result
}

// cowInsert performs COW insertion.
func cowInsert(node RopeNode, pos int, text string) RopeNode {
	if node.IsLeaf() {
		leaf := node.(*LeafNode)
		if pos == 0 {
			return concatNodes(&LeafNode{text: text}, leaf)
		}
		if pos == leaf.Length() {
			return concatNodes(leaf, &LeafNode{text: text})
		}
		return splitAndInsert(leaf, pos, text)
	}

	internal := node.(*InternalNode)
	leftLen := internal.length

	if pos <= leftLen {
		newLeft := cowInsert(internal.left, pos, text)
		return &InternalNode{
			left:   newLeft,
			right:  internal.right, // Share right subtree
			length: newLeft.Length(),
			size:   newLeft.Size(),
		}
	}

	newRight := cowInsert(internal.right, pos-leftLen, text)
	return &InternalNode{
		left:   internal.left, // Share left subtree
		right:  newRight,
		length: internal.left.Length(),
		size:   internal.left.Size(),
	}
}

// splitAndInsert splits leaf and inserts text.
func splitAndInsert(leaf *LeafNode, pos int, text string) RopeNode {
	runes := []rune(leaf.text)
	leftText := string(runes[:pos])
	rightText := string(runes[pos:])

	return concatNodes(
		&LeafNode{text: leftText + text},
		&LeafNode{text: rightText},
	)
}

// Delete removes characters with COW optimization.
func (r *CowRope) Delete(start, end int) *CowRope {
	if r == nil {
		return r
	}
	if start < 0 || end > r.length || start > end {
		panic("delete range out of bounds")
	}
	if start == end {
		return r
	}

	deletedLen := utf8.RuneCountInString(nodeSlice(r.root.node, start, end))
	deletedSize := len(nodeSlice(r.root.node, start, end))

	newRoot := cowDelete(r.root.CloneIfNeeded(), start, end)
	result := &CowRope{
		root:   NewCowNode(newRoot),
		length: r.length - deletedLen,
		size:   r.size - deletedSize,
	}
	r.root.Retain() // Retain old root for sharing
	return result
}

// cowDelete performs COW deletion.
func cowDelete(node RopeNode, start, end int) RopeNode {
	if node.IsLeaf() {
		leaf := node.(*LeafNode)
		runes := []rune(leaf.text)
		newText := string(runes[:start]) + string(runes[end:])
		return &LeafNode{text: newText}
	}

	internal := node.(*InternalNode)
	leftLen := internal.length

	if end <= leftLen {
		newLeft := cowDelete(internal.left, start, end)
		if newLeft.Length() == 0 {
			return internal.right
		}
		return &InternalNode{
			left:   newLeft,
			right:  internal.right,
			length: newLeft.Length(),
			size:   newLeft.Size(),
		}
	}

	if start >= leftLen {
		newRight := cowDelete(internal.right, start-leftLen, end-leftLen)
		if newRight.Length() == 0 {
			return internal.left
		}
		return &InternalNode{
			left:   internal.left,
			right:  newRight,
			length: internal.left.Length(),
			size:   internal.left.Size(),
		}
	}

	// Spans both subtrees
	newLeft := cowDelete(internal.left, start, leftLen)
	newRight := cowDelete(internal.right, 0, end-leftLen)
	return concatNodes(newLeft, newRight)
}

// nodeSlice extracts substring from node.
func nodeSlice(node RopeNode, start, end int) string {
	if node.IsLeaf() {
		leaf := node.(*LeafNode)
		runes := []rune(leaf.text)
		return string(runes[start:end])
	}

	internal := node.(*InternalNode)
	leftLen := internal.length

	if end <= leftLen {
		return nodeSlice(internal.left, start, end)
	}
	if start >= leftLen {
		return nodeSlice(internal.right, start-leftLen, end-leftLen)
	}

	leftPart := nodeSlice(internal.left, start, leftLen)
	rightPart := nodeSlice(internal.right, 0, end-leftLen)
	return leftPart + rightPart
}

// ========== Tree Depth for CowRope ==========

// Depth returns the depth of the rope tree.
func (r *CowRope) Depth() int {
	return cowNodeDepth(r.root.node)
}

// cowNodeDepth calculates the depth of a node for CowRope.
func cowNodeDepth(node RopeNode) int {
	if node == nil || node.IsLeaf() {
		return 0
	}

	internal := node.(*InternalNode)
	leftDepth := cowNodeDepth(internal.left)
	rightDepth := cowNodeDepth(internal.right)
	maxDepth := leftDepth
	if rightDepth > maxDepth {
		maxDepth = rightDepth
	}
	return maxDepth + 1
}

// ShouldRebalance returns true if the rope should be rebalanced.
func (r *CowRope) ShouldRebalance() bool {
	if r == nil || r.root == nil {
		return false
	}
	return r.Depth() > DefaultMaxDepth
}
