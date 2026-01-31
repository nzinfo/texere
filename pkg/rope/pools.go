package rope

import (
	"sync"
)

// ========== Node Pools for Memory Reuse ==========

// NodePool manages a pool of reusable nodes.
type NodePool struct {
	leafPool     sync.Pool
	internalPool sync.Pool
}

// globalNodePool is the global node pool instance.
var globalNodePool = &NodePool{
	leafPool: sync.Pool{
		New: func() interface{} {
			return &LeafNode{
				text: "",
			}
		},
	},
	internalPool: sync.Pool{
		New: func() interface{} {
			return &InternalNode{
				left:  nil,
				right: nil,
			}
		},
	},
}

// AcquireLeaf acquires a leaf node from the pool.
func AcquireLeaf() *LeafNode {
	node := globalNodePool.leafPool.Get().(*LeafNode)
	// Reset text to empty
	node.text = ""
	return node
}

// ReleaseLeaf releases a leaf node back to the pool.
func ReleaseLeaf(node *LeafNode) {
	if node != nil {
		globalNodePool.leafPool.Put(node)
	}
}

// AcquireInternal acquires an internal node from the pool.
func AcquireInternal() *InternalNode {
	node := globalNodePool.internalPool.Get().(*InternalNode)
	// Reset fields
	node.left = nil
	node.right = nil
	node.length = 0
	node.size = 0
	return node
}

// ReleaseInternal releases an internal node back to the pool.
func ReleaseInternal(node *InternalNode) {
	if node != nil {
		globalNodePool.internalPool.Put(node)
	}
}

// ========== Buffer Pools ==========

// BufferPool manages reusable byte buffers.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024) // 1KB initial buffer
	},
}

// AcquireBuffer acquires a buffer from the pool.
func AcquireBuffer() []byte {
	return bufferPool.Get().([]byte)[:0]
}

// ReleaseBuffer releases a buffer back to the pool.
func ReleaseBuffer(buf []byte) {
	if cap(buf) <= 64*1024 { // Only pool buffers <= 64KB
		bufferPool.Put(buf[:0])
	}
}

// AcquireBufferSize acquires a buffer with minimum size.
func AcquireBufferSize(minSize int) []byte {
	if minSize <= 1024 {
		return AcquireBuffer()
	}

	// For larger buffers, just allocate (don't pool them)
	return make([]byte, 0, minSize)
}

// ========== Builder Pool Integration ==========

// AcquireBuilder acquires a builder from the pool.
func AcquireBuilder() *RopeBuilder {
	builder := &RopeBuilder{
		rope:    Empty(),
		pending: make([]pendingInsert, 0, 16),
	}
	return builder
}

// ReleaseBuilder releases a builder back to the pool.
func ReleaseBuilder(builder *RopeBuilder) {
	if builder != nil {
		builder.Reset()
	}
}

// ========== Pool Statistics ==========

// PoolStats contains statistics about pool usage.
type PoolStats struct {
	LeafAllocations     uint64
	InternalAllocations uint64
	BufferAllocations   uint64
}

// GetPoolStats returns current pool statistics.
// Note: This requires instrumenting the pools, which is not done by default.
func GetPoolStats() PoolStats {
	return PoolStats{}
}
