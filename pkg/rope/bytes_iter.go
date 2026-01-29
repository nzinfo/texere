package rope

// ========== Bytes Iterator ==========

// BytesIterator iterates over the bytes of a rope.
type BytesIterator struct {
	rope       *Rope
	position   int // Current byte position
	totalBytes int
	exhausted  bool
}

// NewBytesIterator creates a new bytes iterator.
func (r *Rope) NewBytesIterator() *BytesIterator {
	if r == nil || r.Length() == 0 {
		return &BytesIterator{
			rope:       r,
			position:   0,
			totalBytes: 0,
			exhausted:  true,
		}
	}
	return &BytesIterator{
		rope:       r,
		position:   -1, // Start before first byte
		totalBytes: r.Size(),
		exhausted:  false,
	}
}

// IterBytes creates an iterator over the rope's bytes.
func (r *Rope) IterBytes() *BytesIterator {
	return r.NewBytesIterator()
}

// Next advances to the next byte and returns true if there are more bytes.
func (it *BytesIterator) Next() bool {
	if it.exhausted {
		return false
	}

	it.position++
	if it.position >= it.totalBytes {
		it.exhausted = true
		return false
	}

	return true
}

// Current returns the current byte.
func (it *BytesIterator) Current() byte {
	if it.position < 0 || it.position >= it.totalBytes {
		panic("iterator out of bounds")
	}
	return it.rope.ByteAt(it.position)
}

// Position returns the current byte position.
func (it *BytesIterator) Position() int {
	return it.position
}

// BytePosition returns the current position (alias for Position).
func (it *BytesIterator) BytePosition() int {
	return it.position
}

// HasNext returns true if there are more bytes to iterate.
func (it *BytesIterator) HasNext() bool {
	return !it.exhausted && (it.position+1) < it.totalBytes
}

// Reset resets the iterator to the beginning.
func (it *BytesIterator) Reset() {
	it.position = -1
	it.exhausted = (it.rope == nil || it.rope.Size() == 0)
}

// IsExhausted returns true if the iterator has been exhausted.
func (it *BytesIterator) IsExhausted() bool {
	return it.exhausted
}

// Collect collects all bytes into a slice.
func (it *BytesIterator) Collect() []byte {
	bytes := make([]byte, 0, it.totalBytes)
	it.Reset()
	for it.Next() {
		bytes = append(bytes, it.Current())
	}
	return bytes
}

// ToBytes is an alias for Collect.
func (it *BytesIterator) ToBytes() []byte {
	return it.Collect()
}

// Skip skips n bytes.
func (it *BytesIterator) Skip(n int) bool {
	if n < 0 {
		return false
	}
	for i := 0; i < n && it.Next(); i++ {
	}
	return it.HasNext() || it.position < it.totalBytes-1
}

// Peek returns the next byte without advancing the iterator.
func (it *BytesIterator) Peek() byte {
	if it.position+1 >= it.totalBytes {
		panic("no next byte")
	}
	return it.rope.ByteAt(it.position + 1)
}

// HasPeek returns true if there is a next byte to peek.
func (it *BytesIterator) HasPeek() bool {
	return it.position+1 < it.totalBytes
}

// ========== Bytes Iterator At Position ==========

// BytesIteratorAt creates a bytes iterator starting at a specific byte position.
func (r *Rope) BytesIteratorAt(byteIdx int) *BytesIterator {
	if r == nil || r.Size() == 0 {
		return &BytesIterator{rope: r, exhausted: true}
	}

	if byteIdx < 0 || byteIdx > r.Size() {
		panic("byte index out of bounds")
	}

	if byteIdx == r.Size() {
		return &BytesIterator{
			rope:       r,
			position:   byteIdx - 1,
			totalBytes: r.Size(),
			exhausted:  true,
		}
	}

	return &BytesIterator{
		rope:       r,
		position:   byteIdx - 1, // Next() will move to byteIdx
		totalBytes: r.Size(),
		exhausted:  false,
	}
}

// IterBytesAt creates an iterator starting at a specific byte position.
func (r *Rope) IterBytesAt(byteIdx int) *BytesIterator {
	return r.BytesIteratorAt(byteIdx)
}

// Seek seeks to a specific byte position.
// Returns true if the position is valid.
func (it *BytesIterator) Seek(byteIdx int) bool {
	if byteIdx < 0 || byteIdx >= it.totalBytes {
		it.exhausted = true
		return false
	}

	it.position = byteIdx - 1 // Next() will move to byteIdx
	it.exhausted = false
	return true
}

// ========== Advanced Bytes Operations ==========

// ForEachByte applies a function to each byte in the rope.
func (r *Rope) ForEachByte(fn func(byte) bool) bool {
	it := r.IterBytes()
	for it.Next() {
		if !fn(it.Current()) {
			return false
		}
	}
	return true
}

// ForEachByteWithIndex applies a function to each byte with its index.
func (r *Rope) ForEachByteWithIndex(fn func(int, byte) bool) bool {
	it := r.IterBytes()
	for it.Next() {
		if !fn(it.Position(), it.Current()) {
			return false
		}
	}
	return true
}

// MapBytes maps each byte through a function and returns a new byte slice.
func (r *Rope) MapBytes(fn func(byte) byte) []byte {
	it := r.IterBytes()
	result := make([]byte, 0, r.Size())
	for it.Next() {
		result = append(result, fn(it.Current()))
	}
	return result
}

// FilterBytes filters bytes by a predicate function.
func (r *Rope) FilterBytes(fn func(byte) bool) []byte {
	it := r.IterBytes()
	result := make([]byte, 0, r.Size())
	for it.Next() {
		b := it.Current()
		if fn(b) {
			result = append(result, b)
		}
	}
	return result
}

// FindByte finds the first byte that satisfies the predicate.
// Returns the byte position and true if found, -1 and false otherwise.
func (r *Rope) FindByte(fn func(byte) bool) (int, bool) {
	it := r.IterBytes()
	for it.Next() {
		if fn(it.Current()) {
			return it.Position(), true
		}
	}
	return -1, false
}

// FindByteFrom finds the first byte starting from a given position.
func (r *Rope) FindByteFrom(startByte int, fn func(byte) bool) (int, bool) {
	it := r.IterBytesAt(startByte)
	for it.Next() {
		if fn(it.Current()) {
			return it.Position(), true
		}
	}
	return -1, false
}

// AllBytes checks if all bytes satisfy the predicate.
func (r *Rope) AllBytes(fn func(byte) bool) bool {
	it := r.IterBytes()
	for it.Next() {
		if !fn(it.Current()) {
			return false
		}
	}
	return true
}

// AnyByte checks if any byte satisfies the predicate.
func (r *Rope) AnyByte(fn func(byte) bool) bool {
	it := r.IterBytes()
	for it.Next() {
		if fn(it.Current()) {
			return true
		}
	}
	return false
}

// CountBytes counts bytes that satisfy the predicate.
func (r *Rope) CountBytes(fn func(byte) bool) int {
	count := 0
	it := r.IterBytes()
	for it.Next() {
		if fn(it.Current()) {
			count++
		}
	}
	return count
}

// BytesEquals checks if the rope's bytes equal the given byte slice.
func (r *Rope) BytesEquals(bytes []byte) bool {
	if r.Size() != len(bytes) {
		return false
	}

	it := r.IterBytes()
	for _, b := range bytes {
		if !it.Next() || it.Current() != b {
			return false
		}
	}
	return !it.Next()
}

// ToBytes converts the rope to a byte slice.
func (r *Rope) ToBytes() []byte {
	return []byte(r.String())
}
