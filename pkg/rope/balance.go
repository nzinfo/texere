package rope

import (
	"math"
)

// Balance operations maintain the B-tree properties of the rope.
// A balanced rope ensures O(log n) operations and minimal memory overhead.

const (
	// DefaultMinLeafSize is the minimum size of a leaf node in characters.
	DefaultMinLeafSize = 256

	// DefaultMaxLeafSize is the maximum size of a leaf node in characters.
	DefaultMaxLeafSize = 1024

	// DefaultMaxDepth is the maximum allowed depth of the rope tree.
	DefaultMaxDepth = 64
)

// BalanceConfig contains configuration for balancing operations.
type BalanceConfig struct {
	MinLeafSize int // Minimum leaf size in characters
	MaxLeafSize int // Maximum leaf size in characters
	MaxDepth    int // Maximum tree depth
}

// DefaultBalanceConfig returns the default balancing configuration.
func DefaultBalanceConfig() *BalanceConfig {
	return &BalanceConfig{
		MinLeafSize: DefaultMinLeafSize,
		MaxLeafSize: DefaultMaxLeafSize,
		MaxDepth:    DefaultMaxDepth,
	}
}

// Balance rebalances the rope to optimize performance.
// This operation creates a new rope with balanced tree structure.
func (r *Rope) Balance() *Rope {
	return r.BalanceWithConfig(DefaultBalanceConfig())
}

// BalanceWithConfig rebalances the rope with the given configuration.
func (r *Rope) BalanceWithConfig(config *BalanceConfig) *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	builder := NewBuilder()
	rebalanceNode(r.root, builder, config)
	return builder.Build()
}

// rebalanceNode recursively rebalances a node.
func rebalanceNode(node RopeNode, builder *RopeBuilder, config *BalanceConfig) {
	if node == nil {
		return
	}

	if node.IsLeaf() {
		leaf := node.(*LeafNode)
		text := leaf.text

		// Split large leaves into smaller chunks
		for len(text) > 0 {
			chunkSize := len(text)
			if chunkSize > config.MaxLeafSize {
				// Split at a character boundary
				chunkSize = config.MaxLeafSize
				// Ensure we don't split in the middle of a multi-byte UTF-8 sequence
				for chunkSize > 0 && (text[chunkSize]&0xC0) == 0x80 {
					chunkSize--
				}
			}

			builder.Append(text[:chunkSize])
			text = text[chunkSize:]
		}

		return
	}

	internal := node.(*InternalNode)
	rebalanceNode(internal.left, builder, config)
	rebalanceNode(internal.right, builder, config)
}

// Depth returns the maximum depth of the rope tree.
func (r *Rope) Depth() int {
	if r == nil || r.root == nil {
		return 0
	}
	return nodeDepth(r.root)
}

// nodeDepth returns the depth of a node.
func nodeDepth(node RopeNode) int {
	if node == nil || node.IsLeaf() {
		return 0
	}

	internal := node.(*InternalNode)
	leftDepth := nodeDepth(internal.left)
	rightDepth := nodeDepth(internal.right)

	return 1 + max(leftDepth, rightDepth)
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// IsBalanced checks if the rope is reasonably balanced.
// A rope is balanced if its depth is O(log n).
func (r *Rope) IsBalanced() bool {
	if r == nil || r.Length() == 0 {
		return true
	}

	depth := r.Depth()
	// A balanced tree should have depth <= 2 * log2(n)
	maxDepth := 2 * int(math.Ceil(math.Log2(float64(r.Length()+1))))
	return depth <= maxDepth
}

// Optimize optimizes the rope structure for common operations.
// This includes balancing and merging small leaves.
func (r *Rope) Optimize() *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	config := DefaultBalanceConfig()
	config.MinLeafSize = DefaultMinLeafSize / 2 // Allow smaller leaves for better granularity

	return r.BalanceWithConfig(config)
}

// ========== Node Balancing Helpers ==========

// shouldMerge returns true if two small nodes should be merged.
func shouldMerge(left, right RopeNode, config *BalanceConfig) bool {
	if left == nil || right == nil {
		return false
	}

	if !left.IsLeaf() || !right.IsLeaf() {
		return false
	}

	leftLeaf := left.(*LeafNode)
	rightLeaf := right.(*LeafNode)

	totalSize := leftLeaf.Size() + rightLeaf.Size()
	return totalSize <= config.MaxLeafSize
}

// mergeLeaves merges two leaf nodes into one.
func mergeLeaves(left, right *LeafNode) *LeafNode {
	return &LeafNode{text: left.text + right.text}
}

// shouldSplit returns true if a leaf should be split.
func shouldSplit(leaf *LeafNode, config *BalanceConfig) bool {
	return leaf.Size() > config.MaxLeafSize
}

// splitLeaf splits a leaf node at a character position.
func splitLeaf(leaf *LeafNode, pos int) (*LeafNode, *LeafNode) {
	runes := []rune(leaf.text)
	left := &LeafNode{text: string(runes[:pos])}
	right := &LeafNode{text: string(runes[pos:])}
	return left, right
}

// rebalanceTree rebalances an internal node and its children.
func rebalanceTree(node RopeNode, config *BalanceConfig) RopeNode {
	if node == nil || node.IsLeaf() {
		return node
	}

	internal := node.(*InternalNode)

	// Recursively balance children
	internal.left = rebalanceTree(internal.left, config)
	internal.right = rebalanceTree(internal.right, config)

	// Merge small adjacent leaves
	if shouldMerge(internal.left, internal.right, config) {
		leftLeaf := internal.left.(*LeafNode)
		rightLeaf := internal.right.(*LeafNode)
		return mergeLeaves(leftLeaf, rightLeaf)
	}

	// Update cached values
	internal.length = internal.left.Length()
	internal.size = internal.left.Size()

	return internal
}

// ========== Tree Health Metrics ==========

// TreeStats contains statistics about a rope's tree structure.
type TreeStats struct {
	NodeCount     int // Total number of nodes
	LeafCount     int // Number of leaf nodes
	InternalCount int // Number of internal nodes
	Depth         int // Maximum depth
	AvgDepth      float64 // Average depth of leaves
	MinLeafSize   int // Smallest leaf size
	MaxLeafSize   int // Largest leaf size
	AvgLeafSize   float64 // Average leaf size
}

// Stats returns statistics about the rope's tree structure.
func (r *Rope) Stats() *TreeStats {
	if r == nil || r.root == nil {
		return &TreeStats{}
	}

	stats := &TreeStats{}
	collectStats(r.root, 0, stats)

	if stats.LeafCount > 0 {
		stats.AvgDepth = float64(stats.Depth) / float64(stats.LeafCount)
		stats.AvgLeafSize = float64(stats.MaxLeafSize) / float64(stats.LeafCount)
	}

	return stats
}

// collectStats collects statistics from a node.
func collectStats(node RopeNode, depth int, stats *TreeStats) {
	if node == nil {
		return
	}

	stats.NodeCount++

	if node.IsLeaf() {
		leaf := node.(*LeafNode)
		stats.LeafCount++
		stats.Depth = max(stats.Depth, depth)

		size := leaf.Size()
		stats.MaxLeafSize += size

		if stats.MinLeafSize == 0 || size < stats.MinLeafSize {
			stats.MinLeafSize = size
		}
		if size > stats.MaxLeafSize {
			stats.MaxLeafSize = size
		}
	} else {
		internal := node.(*InternalNode)
		stats.InternalCount++

		collectStats(internal.left, depth+1, stats)
		collectStats(internal.right, depth+1, stats)
	}
}

// ========== Memory Optimization ==========

// Compact reduces memory usage by merging small nodes and eliminating redundancy.
func (r *Rope) Compact() *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	newRoot := rebuildOptimal(r.root, DefaultMinLeafSize, DefaultMaxLeafSize)
	return &Rope{
		root:   newRoot,
		length: r.length,
		size:   r.size,
	}
}

// rebuildOptimal rebuilds a subtree with optimal node sizes.
func rebuildOptimal(node RopeNode, minSize, maxSize int) RopeNode {
	if node == nil {
		return nil
	}

	if node.IsLeaf() {
		return node
	}

	// Collect all leaves in order
	leaves := collectLeaves(node)

	// Merge small leaves
	leaves = mergeLeavesOptimal(leaves, minSize, maxSize)

	// Build balanced tree
	return buildBalancedTree(leaves, 0, len(leaves))
}

// collectLeaves collects all leaves from a subtree in order.
func collectLeaves(node RopeNode) []*LeafNode {
	if node == nil {
		return nil
	}

	if node.IsLeaf() {
		leaf := node.(*LeafNode)
		return []*LeafNode{leaf}
	}

	internal := node.(*InternalNode)
	leftLeaves := collectLeaves(internal.left)
	rightLeaves := collectLeaves(internal.right)

	return append(leftLeaves, rightLeaves...)
}

// mergeLeavesOptimal merges leaves to optimal sizes.
func mergeLeavesOptimal(leaves []*LeafNode, minSize, maxSize int) []*LeafNode {
	if len(leaves) == 0 {
		return leaves
	}

	merged := make([]*LeafNode, 0)
	current := leaves[0]

	for i := 1; i < len(leaves); i++ {
		combinedSize := current.Size() + leaves[i].Size()

		if combinedSize <= maxSize {
			// Merge
			current = &LeafNode{text: current.text + leaves[i].text}
		} else {
			// Don't merge
			merged = append(merged, current)
			current = leaves[i]
		}
	}

	merged = append(merged, current)
	return merged
}

// buildBalancedTree builds a balanced tree from a slice of leaves.
func buildBalancedTree(leaves []*LeafNode, start, end int) RopeNode {
	if start >= end {
		return nil
	}

	if start == end-1 {
		return leaves[start]
	}

	mid := (start + end) / 2
	left := buildBalancedTree(leaves, start, mid)
	right := buildBalancedTree(leaves, mid, end)

	if left == nil {
		return right
	}
	if right == nil {
		return left
	}

	return &InternalNode{
		left:   left,
		right:  right,
		length: left.Length(),
		size:   left.Size(),
	}
}

// ========== Validation ==========

// Validate checks the integrity of the rope structure.
// Returns nil if the rope is valid, or an error describing the problem.
func (r *Rope) Validate() error {
	if r == nil || r.root == nil {
		return nil
	}

	return validateNode(r.root, r.length, r.size)
}

// validateNode validates a node and its children.
func validateNode(node RopeNode, expectedLength, expectedSize int) error {
	if node == nil {
		return nil
	}

	// Check length
	if node.Length() != expectedLength {
		return &RopeError{
			Type:    "LengthMismatch",
			Message: "node length mismatch",
		}
	}

	// Check size
	if node.Size() != expectedSize {
		return &RopeError{
			Type:    "SizeMismatch",
			Message: "node size mismatch",
		}
	}

	if node.IsLeaf() {
		return nil
	}

	internal := node.(*InternalNode)

	// Validate left subtree
	if err := validateNode(internal.left, internal.length, internal.size); err != nil {
		return err
	}

	// Validate right subtree
	rightLength := expectedLength - internal.length
	rightSize := expectedSize - internal.size
	if err := validateNode(internal.right, rightLength, rightSize); err != nil {
		return err
	}

	return nil
}

// RopeError represents an error in the rope structure.
type RopeError struct {
	Type    string
	Message string
}

func (e *RopeError) Error() string {
	return e.Type + ": " + e.Message
}

// ========== Performance Tuning ==========

// SuggestedConfig returns a balance configuration based on rope size.
func (r *Rope) SuggestedConfig() *BalanceConfig {
	if r == nil {
		return DefaultBalanceConfig()
	}

	size := r.Length()

	config := DefaultBalanceConfig()

	// Adjust for very small ropes
	if size < 1024 {
		config.MinLeafSize = 64
		config.MaxLeafSize = 256
		return config
	}

	// Adjust for very large ropes
	if size > 1024*1024 {
		config.MinLeafSize = 512
		config.MaxLeafSize = 2048
		return config
	}

	// Default for medium ropes
	return config
}

// AutoBalance automatically balances the rope if needed.
// Returns the balanced rope, or the original if no balancing was needed.
func (r *Rope) AutoBalance() *Rope {
	if r == nil || r.IsBalanced() {
		return r
	}

	return r.BalanceWithConfig(r.SuggestedConfig())
}

// LeafCount returns the number of leaf nodes in the rope tree.
func (r *Rope) LeafCount() int {
	if r == nil || r.root == nil {
		return 0
	}
	return nodeLeafCount(r.root)
}

// nodeLeafCount recursively counts leaf nodes.
func nodeLeafCount(node RopeNode) int {
	if node == nil {
		return 0
	}
	if node.IsLeaf() {
		return 1
	}

	internal := node.(*InternalNode)
	return nodeLeafCount(internal.left) + nodeLeafCount(internal.right)
}

// NodeCount returns the total number of nodes in the rope tree.
func (r *Rope) NodeCount() int {
	if r == nil || r.root == nil {
		return 0
	}
	return nodeCountTotal(r.root)
}

// nodeCountTotal recursively counts all nodes.
func nodeCountTotal(node RopeNode) int {
	if node == nil {
		return 0
	}
	if node.IsLeaf() {
		return 1
	}

	internal := node.(*InternalNode)
	return 1 + nodeCountTotal(internal.left) + nodeCountTotal(internal.right)
}
