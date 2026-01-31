package rope

import (
	"hash/fnv"
)

// ========== Hash Support ==========

// HashCode returns a 32-bit hash code of the rope's content.
// Uses FNV-1a hashing algorithm.
func (r *Rope) HashCode() uint32 {
	if r == nil || r.Length() == 0 {
		return 0
	}

	h := fnv.New32a()
	it := r.NewIterator()
	for it.Next() {
		r := it.Current()
		h.Write([]byte(string(r)))
	}
	return h.Sum32()
}

// HashCode32 returns a 32-bit hash code (alias for HashCode).
func (r *Rope) HashCode32() uint32 {
	return r.HashCode()
}

// HashCode64 returns a 64-bit hash code of the rope's content.
func (r *Rope) HashCode64() uint64 {
	if r == nil || r.Length() == 0 {
		return 0
	}

	h := fnv.New64a()
	it := r.NewIterator()
	for it.Next() {
		r := it.Current()
		h.Write([]byte(string(r)))
	}
	return h.Sum64()
}

// HashString returns a string representation of the hash code.
// Useful for debugging or display purposes.
func (r *Rope) HashString() string {
	if r == nil || r.Length() == 0 {
		return "0"
	}
	h := r.HashCode()
	return uint32ToString(h)
}

// ========== Hash Equality ==========

// HashEquals checks if two ropes have the same hash code.
// Note: This is not a guarantee of content equality (hash collisions possible),
// but can be used as a fast pre-check before full comparison.
func (r *Rope) HashEquals(other *Rope) bool {
	if r == nil && other == nil {
		return true
	}
	if r == nil || other == nil {
		return false
	}
	return r.HashCode() == other.HashCode()
}

// ========== Hash Combining ==========

// CombineHash combines multiple hash codes into one.
func CombineHash(codes ...uint32) uint32 {
	if len(codes) == 0 {
		return 0
	}

	result := codes[0]
	for i := 1; i < len(codes); i++ {
		result = result*31 + codes[i]
	}
	return result
}

// CombineHash64 combines multiple 64-bit hash codes into one.
func CombineHash64(codes ...uint64) uint64 {
	if len(codes) == 0 {
		return 0
	}

	result := codes[0]
	for i := 1; i < len(codes); i++ {
		result = result*31 + codes[i]
	}
	return result
}

// ========== Chunk-based Hashing ==========

// ChunkHashes returns hash codes for each chunk in the rope.
// This can be useful for incremental hashing or diffing.
func (r *Rope) ChunkHashes() []uint32 {
	if r == nil || r.Length() == 0 {
		return []uint32{}
	}

	it := r.Chunks()
	hashes := make([]uint32, 0, it.Count())

	for it.Next() {
		chunk := it.Current()
		h := fnv.New32a()
		h.Write([]byte(chunk))
		hashes = append(hashes, h.Sum32())
	}

	return hashes
}

// CombinedChunkHash combines all chunk hashes into a single hash.
func (r *Rope) CombinedChunkHash() uint32 {
	hashes := r.ChunkHashes()
	return CombineHash(hashes...)
}

// ========== Rolling Hash ==========

// RollingHasher supports incremental rolling hash computation.
type RollingHasher struct {
	rope   *Rope
	window int
	hash   uint32
}

// NewRollingHasher creates a new rolling hasher for the rope.
func (r *Rope) NewRollingHasher(windowSize int) *RollingHasher {
	if windowSize <= 0 || windowSize > r.Length() {
		windowSize = r.Length()
	}

	hasher := &RollingHasher{
		rope:   r,
		window: windowSize,
		hash:   0,
	}

	hasher.initialize()
	return hasher
}

// initialize computes the initial hash.
func (rh *RollingHasher) initialize() {
	if rh.rope == nil || rh.window == 0 {
		return
	}

	h := fnv.New32a()
	it := rh.rope.NewIterator()
	for i := 0; i < rh.window && it.Next(); i++ {
		h.Write([]byte(string(it.Current())))
	}
	rh.hash = h.Sum32()
}

// Current returns the current hash value.
func (rh *RollingHasher) Current() uint32 {
	return rh.hash
}

// Roll advances the window by one character.
func (rh *RollingHasher) Roll() bool {
	if rh.rope == nil || rh.window >= rh.rope.Length() {
		return false
	}

	// This is a simplified rolling hash implementation
	// For a true rolling hash, you'd need to remove the outgoing
	// character and add the incoming character
	h := fnv.New32a()
	start := rh.rope.CharToByte(rh.window)
	end := rh.rope.CharToByte(rh.window + 1)
	if end > start {
		slice := rh.rope.Slice(start, end)
		h.Write([]byte(slice))
		rh.hash = rh.hash ^ h.Sum32()
	}

	return true
}

// ========== Hash-based Comparison ==========

// LikelyEquals checks if two ropes are likely equal by comparing hash codes first.
// Returns false immediately if hashes differ, otherwise does full comparison.
func (r *Rope) LikelyEquals(other *Rope) bool {
	// Fast path: check hashes first
	if r.HashCode() != other.HashCode() {
		return false
	}

	// Hashes match, do full comparison
	return r.Equals(other)
}

// ========== Hash Map Utilities ==========

// HashKey returns a value suitable for use as a map key.
// This is the HashCode() but with a more semantic name for map usage.
func (r *Rope) HashKey() uint32 {
	return r.HashCode()
}

// ========== Hash Set Utilities ==========

// HashSlice returns a slice of hash codes for a slice of ropes.
func HashSlice(ropes []*Rope) []uint32 {
	hashes := make([]uint32, len(ropes))
	for i, r := range ropes {
		if r != nil {
			hashes[i] = r.HashCode()
		}
	}
	return hashes
}

// ========== Incremental Hashing ==========

// IncrementalHasher supports incremental hashing of rope modifications.
type IncrementalHasher struct {
	baseHash uint32
}

// NewIncrementalHasher creates an incremental hasher starting from a base hash.
func NewIncrementalHasher(baseHash uint32) *IncrementalHasher {
	return &IncrementalHasher{
		baseHash: baseHash,
	}
}

// AddString adds a string to the hash.
func (ih *IncrementalHasher) AddString(s string) {
	h := fnv.New32a()
	h.Write([]byte(s))
	ih.baseHash ^= h.Sum32()
}

// AddRope adds a rope to the hash.
func (ih *IncrementalHasher) AddRope(r *Rope) {
	if r != nil {
		ih.baseHash ^= r.HashCode()
	}
}

// Current returns the current hash value.
func (ih *IncrementalHasher) Current() uint32 {
	return ih.baseHash
}

// Reset resets to zero.
func (ih *IncrementalHasher) Reset() {
	ih.baseHash = 0
}

// ========== Helper Functions ==========

// uint32ToString converts a uint32 to a hex string.
func uint32ToString(h uint32) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		result[i] = hexChars[h&0xf]
		h >>= 4
	}
	return string(result)
}

// HashBytes returns a hash of a byte slice.
func HashBytes(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}

// HashString returns a hash of a string.
func HashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// HashRunes returns a hash of a rune slice.
func HashRunes(runes []rune) uint32 {
	h := fnv.New32a()
	for _, r := range runes {
		h.Write([]byte(string(r)))
	}
	return h.Sum32()
}
