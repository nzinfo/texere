package rope

// ========== Byte/Char Conversion Utilities ==========

// CharToByte converts a character index to a byte index.
func (r *Rope) CharToByte(charIdx int) int {
	if r == nil || charIdx <= 0 {
		return 0
	}
	if charIdx >= r.Length() {
		return r.Size()
	}

	// Count bytes up to character index
	byteIdx := 0
	it := r.NewIterator()
	for i := 0; i < charIdx && it.Next(); i++ {
		r := it.Current()
		byteIdx += lenRune(r)
	}
	return byteIdx
}

// ByteToChar converts a byte index to a character index.
func (r *Rope) ByteToChar(byteIdx int) int {
	if r == nil || byteIdx <= 0 {
		return 0
	}
	if byteIdx >= r.Size() {
		return r.Length()
	}

	// Count characters up to byte index
	charIdx := 0
	byteCount := 0
	it := r.NewIterator()
	for it.Next() {
		r := it.Current()
		runeLen := lenRune(r)
		if byteCount+runeLen > byteIdx {
			break
		}
		byteCount += runeLen
		charIdx++
	}
	return charIdx
}

// charToByte is an alias for CharToByte (for backward compatibility).
func (r *Rope) charToByte(charIdx int) int {
	return r.CharToByte(charIdx)
}

// byteToChar is an alias for ByteToChar (for backward compatibility).
func (r *Rope) byteToChar(byteIdx int) int {
	return r.ByteToChar(byteIdx)
}

// lenRune returns the byte length of a rune.
func lenRune(r rune) int {
	if r <= 0x7F {
		return 1
	}
	if r <= 0x7FF {
		return 2
	}
	if r <= 0xFFFF {
		return 3
	}
	return 4
}
