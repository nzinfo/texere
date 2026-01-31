package rope

import (
	"unicode/utf8"
)

// ========== Byte Position Cache ==========

// BytePosCache caches character-to-byte position mappings.
// This avoids repeated UTF-8 traversal for the same string.
type BytePosCache struct {
	text       string
	charToByte []int // Sparse cache of character -> byte positions
}

// NewBytePosCache creates a new cache for the given text.
func NewBytePosCache(text string) *BytePosCache {
	return &BytePosCache{
		text:       text,
		charToByte: make([]int, 0, 32), // Start with small capacity
	}
}

// GetBytePos returns the byte position for the given character position.
// Uses cache with exponential backoff for efficiency.
func (c *BytePosCache) GetBytePos(charPos int) int {
	// Fast path: if we have this exact position cached (and cache is long enough)
	if charPos >= 0 && charPos < len(c.charToByte) && c.charToByte[charPos] != 0 {
		return c.charToByte[charPos]
	}

	// Find nearest cached position before charPos
	cachePos := c.findNearestCachePos(charPos)

	// Start from cached position and iterate forward
	startChar := 0
	startByte := 0
	if cachePos >= 0 && cachePos < len(c.charToByte) {
		startChar = cachePos
		startByte = c.charToByte[cachePos]
	}

	// Iterate from start position
	bytePos := startByte
	for i := startChar; i < charPos; i++ {
		_, size := utf8.DecodeRuneInString(c.text[bytePos:])
		bytePos += size

		// Cache every 8th position (exponential backoff)
		if (i+1)%8 == 0 && i+2 <= cap(c.charToByte) {
			c.ensureCapacity(i + 2) // Need length i+2 to access index i+1
			c.charToByte[i+1] = bytePos
		}
	}

	return bytePos
}

// findNearestCachePos finds the nearest cached position <= charPos.
func (c *BytePosCache) findNearestCachePos(charPos int) int {
	// Binary search for nearest cached position
	left, right := 0, len(c.charToByte)-1
	result := -1

	for left <= right {
		mid := (left + right) / 2
		if c.charToByte[mid] != 0 && mid <= charPos {
			result = mid
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return result
}

// ensureCapacity ensures the cache has at least minCapacity.
func (c *BytePosCache) ensureCapacity(minCapacity int) {
	if minCapacity > cap(c.charToByte) {
		newCache := make([]int, minCapacity, minCapacity*2)
		copy(newCache, c.charToByte)
		c.charToByte = newCache
	}
	if minCapacity > len(c.charToByte) {
		// Extend with zeros
		c.charToByte = append(c.charToByte, make([]int, minCapacity-len(c.charToByte))...)
	}
}

// Reset clears the cache.
func (c *BytePosCache) Reset() {
	c.charToByte = c.charToByte[:0]
}

// ========== Cached Leaf Node ==========

// CachedLeaf is a leaf node with built-in byte position cache.
type CachedLeaf struct {
	text  string
	cache *BytePosCache
}

// NewCachedLeaf creates a new cached leaf node.
func NewCachedLeaf(text string) *CachedLeaf {
	return &CachedLeaf{
		text:  text,
		cache: NewBytePosCache(text),
	}
}

// Length returns the number of characters.
func (n *CachedLeaf) Length() int {
	return utf8.RuneCountInString(n.text)
}

// Size returns the number of bytes.
func (n *CachedLeaf) Size() int {
	return len(n.text)
}

// Slice returns a substring.
func (n *CachedLeaf) Slice(start, end int) string {
	startByte := n.cache.GetBytePos(start)
	endByte := n.cache.GetBytePos(end)
	return n.text[startByte:endByte]
}

// IsLeaf returns true.
func (n *CachedLeaf) IsLeaf() bool {
	return true
}

// SplitAt splits the leaf at the given character position.
func (n *CachedLeaf) SplitAt(pos int) (*CachedLeaf, *CachedLeaf) {
	if pos <= 0 {
		return nil, n
	}
	if pos >= n.Length() {
		return n, nil
	}

	splitByte := n.cache.GetBytePos(pos)
	left := NewCachedLeaf(n.text[:splitByte])
	right := NewCachedLeaf(n.text[splitByte:])
	return left, right
}

// ========== Fast Byte Position Utilities ==========

// FindBytePositions finds all byte positions for the given character positions.
// Uses cache for efficiency when multiple positions are needed.
func FindBytePositions(text string, charPositions []int) []int {
	cache := NewBytePosCache(text)
	result := make([]int, len(charPositions))

	for i, pos := range charPositions {
		result[i] = cache.GetBytePos(pos)
	}

	return result
}

// FindBytePosFast finds byte position for a single character position.
// Optimized for small texts without cache overhead.
func FindBytePosFast(text string, charPos int) int {
	bytePos := 0
	for i := 0; i < charPos; i++ {
		_, size := utf8.DecodeRuneInString(text[bytePos:])
		bytePos += size
	}
	return bytePos
}

// CountRunesInBytes counts runes in a byte range efficiently.
func CountRunesInBytes(bytes []byte, start, end int) int {
	count := 0
	pos := start
	for pos < end {
		_, size := utf8.DecodeRune(bytes[pos:])
		pos += size
		count++
	}
	return count
}

// CountRunesInString counts runes in a string byte range.
func CountRunesInString(text string, startByte, endByte int) int {
	count := 0
	pos := startByte
	for pos < endByte {
		_, size := utf8.DecodeRuneInString(text[pos:])
		pos += size
		count++
	}
	return count
}
