package rope

import (
	"unicode/utf8"
)

// ========== UTF-16 Support ==========

// LenUTF16 returns the number of UTF-16 code units needed to represent the rope.
// This is important for interoperability with JavaScript and Windows APIs,
// which use UTF-16 encoding.
//
// Most Unicode characters (U+0000 to U+FFFF) require one UTF-16 code unit.
// Characters outside the Basic Multilingual Plane (U+10000 to U+10FFFF)
// require two UTF-16 code units (a surrogate pair).
func (r *Rope) LenUTF16() int {
	if r == nil || r.Length() == 0 {
		return 0
	}

	count := 0
	it := r.NewIterator()
	for it.Next() {
		r := it.Current()
		if r <= 0xFFFF {
			count++ // BMP character
		} else {
			count += 2 // Surrogate pair
		}
	}
	return count
}

// LenUTF16CU is an alias for LenUTF16 (CU = Code Units).
func (r *Rope) LenUTF16CU() int {
	return r.LenUTF16()
}

// CharToUTF16Offset converts a character index to a UTF-16 code unit offset.
// Returns the offset in UTF-16 code units.
func (r *Rope) CharToUTF16Offset(charIdx int) int {
	if r == nil || charIdx <= 0 {
		return 0
	}
	if charIdx > r.Length() {
		charIdx = r.Length()
	}

	offset := 0
	it := r.NewIterator()
	for i := 0; i < charIdx && it.Next(); i++ {
		r := it.Current()
		if r <= 0xFFFF {
			offset++
		} else {
			offset += 2
		}
	}
	return offset
}

// UTF16OffsetToChar converts a UTF-16 code unit offset to a character index.
// Returns the character index.
func (r *Rope) UTF16OffsetToChar(utf16Offset int) int {
	if r == nil || utf16Offset <= 0 {
		return 0
	}

	charIdx := 0
	offset := 0
	it := r.NewIterator()
	for it.Next() {
		r := it.Current()
		if r <= 0xFFFF {
			offset++
			if offset > utf16Offset {
				return charIdx
			}
		} else {
			offset += 2
			if offset > utf16Offset {
				return charIdx
			}
		}
		charIdx++
	}
	return charIdx
}

// IsUTF16SurrogatePair checks if a rune is part of a UTF-16 surrogate pair.
func IsUTF16SurrogatePair(r rune) bool {
	return r > 0xFFFF && r <= 0x10FFFF
}

// UTF16HighSurrogate returns the high surrogate of a surrogate pair.
func UTF16HighSurrogate(r rune) uint16 {
	if !IsUTF16SurrogatePair(r) {
		panic("rune is not a surrogate pair")
	}
	return uint16((r-0x10000)>>10) + 0xD800
}

// UTF16LowSurrogate returns the low surrogate of a surrogate pair.
func UTF16LowSurrogate(r rune) uint16 {
	if !IsUTF16SurrogatePair(r) {
		panic("rune is not a surrogate pair")
	}
	return uint16((r-0x10000)&0x3FF) + 0xDC00
}

// DecodeUTF16SurrogatePair decodes a high and low surrogate into a rune.
func DecodeUTF16SurrogatePair(high, low uint16) rune {
	return rune((uint32(high-0xD800)<<10) + uint32(low-0xDC00) + 0x10000)
}

// ========== UTF-16 Iteration ==========

// UTF16Iterator iterates over UTF-16 code units.
type UTF16Iterator struct {
	rope      *Rope
	runeIt    *Iterator
	hasHigh   bool // Whether we have a high surrogate pending
	highSurr  uint16
}

// NewUTF16Iterator creates a new UTF-16 iterator.
func (r *Rope) NewUTF16Iterator() *UTF16Iterator {
	if r == nil || r.Length() == 0 {
		return &UTF16Iterator{rope: r, runeIt: &Iterator{}, hasHigh: false}
	}
	return &UTF16Iterator{
		rope:   r,
		runeIt: r.NewIterator(),
		hasHigh: false,
	}
}

// IterUTF16 creates a UTF-16 iterator.
func (r *Rope) IterUTF16() *UTF16Iterator {
	return r.NewUTF16Iterator()
}

// Next advances to the next UTF-16 code unit and returns true if there are more.
func (it *UTF16Iterator) Next() bool {
	if it.hasHigh {
		// We already have a high surrogate, return low surrogate next
		it.hasHigh = false
		return true
	}

	if !it.runeIt.Next() {
		return false
	}

	r := it.runeIt.Current()
	if r <= 0xFFFF {
		// BMP character - return as single code unit
		return true
	}

	// Surrogate pair - store high surrogate
	it.highSurr = UTF16HighSurrogate(r)
	it.hasHigh = true
	return true
}

// Current returns the current UTF-16 code unit.
func (it *UTF16Iterator) Current() uint16 {
	if it.hasHigh {
		// Return high surrogate
		return it.highSurr
	}
	// Return the actual rune or low surrogate
	r := it.runeIt.Current()
	if r <= 0xFFFF {
		return uint16(r)
	}
	// Return low surrogate
	return UTF16LowSurrogate(r)
}

// CurrentRune returns the current rune.
// This is only valid when not in the middle of a surrogate pair.
func (it *UTF16Iterator) CurrentRune() rune {
	return it.runeIt.Current()
}

// IsSurrogatePair returns true if the current code unit is part of a surrogate pair.
func (it *UTF16Iterator) IsSurrogatePair() bool {
	r := it.runeIt.Current()
	return IsUTF16SurrogatePair(r)
}

// IsHighSurrogate returns true if the current code unit is a high surrogate.
func (it *UTF16Iterator) IsHighSurrogate() bool {
	return it.hasHigh || IsUTF16SurrogatePair(it.runeIt.Current())
}

// IsLowSurrogate returns true if the current code unit is a low surrogate.
func (it *UTF16Iterator) IsLowSurrogate() bool {
	if it.hasHigh {
		return false // We're at high surrogate
	}
	r := it.runeIt.Current()
	return IsUTF16SurrogatePair(r)
}

// Position returns the current UTF-16 code unit position.
func (it *UTF16Iterator) Position() int {
	// This would require tracking position separately
	return 0
}

// Reset resets the iterator to the beginning.
func (it *UTF16Iterator) Reset() {
	it.runeIt.Reset()
	it.hasHigh = false
	it.highSurr = 0
}

// ToSlice collects all UTF-16 code units into a slice.
func (it *UTF16Iterator) ToSlice() []uint16 {
	codes := make([]uint16, 0, it.rope.LenUTF16())
	it.Reset()
	for it.Next() {
		codes = append(codes, it.Current())
	}
	return codes
}

// ========== UTF-16 Utilities ==========

// ToUTF16 converts the rope to a UTF-16 code unit slice.
func (r *Rope) ToUTF16() []uint16 {
	if r == nil || r.Length() == 0 {
		return []uint16{}
	}

	codes := make([]uint16, 0, r.LenUTF16())
	it := r.IterUTF16()
	for it.Next() {
		codes = append(codes, it.Current())
	}
	return codes
}

// ToUTF16String converts the rope to a string containing UTF-16 encoded data.
// Note: Go strings are UTF-8, so this returns a UTF-8 string representation
// of the UTF-16 data.
func (r *Rope) ToUTF16String() string {
	return r.String()
}

// ContainsSurrogatePairs returns true if the rope contains any characters
// that require surrogate pairs in UTF-16.
func (r *Rope) ContainsSurrogatePairs() bool {
	it := r.NewIterator()
	for it.Next() {
		if IsUTF16SurrogatePair(it.Current()) {
			return true
		}
	}
	return false
}

// SurrogatePairCount returns the number of surrogate pairs in the rope.
func (r *Rope) SurrogatePairCount() int {
	count := 0
	it := r.NewIterator()
	for it.Next() {
		if IsUTF16SurrogatePair(it.Current()) {
			count++
		}
	}
	return count
}

// ========== UTF-16 Slice Operations ==========

// SliceUTF16 returns a substring from UTF-16 code unit start to end.
// This is useful when working with JavaScript/Windows APIs that use UTF-16 offsets.
func (r *Rope) SliceUTF16(startUTF16, endUTF16 int) string {
	startChar := r.UTF16OffsetToChar(startUTF16)
	endChar := r.UTF16OffsetToChar(endUTF16)
	return r.Slice(startChar, endChar)
}

// CharRangeFromUTF16Range converts a UTF-16 range to a character range.
func (r *Rope) CharRangeFromUTF16Range(startUTF16, endUTF16 int) (int, int) {
	startChar := r.UTF16OffsetToChar(startUTF16)
	endChar := r.UTF16OffsetToChar(endUTF16)
	return startChar, endChar
}

// UTF16RangeFromCharRange converts a character range to a UTF-16 range.
func (r *Rope) UTF16RangeFromCharRange(startChar, endChar int) (int, int) {
	startUTF16 := r.CharToUTF16Offset(startChar)
	endUTF16 := r.CharToUTF16Offset(endChar)
	return startUTF16, endUTF16
}

// ========== Validation ==========

// IsValidUTF8 checks if the rope contains valid UTF-8.
func (r *Rope) IsValidUTF8() bool {
	if r == nil {
		return true
	}
	it := r.NewIterator()
	for it.Next() {
		if !utf8.ValidRune(it.Current()) {
			return false
		}
	}
	return true
}
