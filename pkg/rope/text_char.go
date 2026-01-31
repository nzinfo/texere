package rope

// ========== Single Character Operations ==========

// InsertChar inserts a single rune at the specified character position.
// Returns a new Rope, leaving the original unchanged.
//
// This is equivalent to Insert(pos, string(r)), but slightly more efficient.
func (r *Rope) InsertChar(pos int, ch rune) *Rope {
	if r == nil {
		return New(string(ch))
	}
	return r.Insert(pos, string(ch))
}

// InsertCharAt is an alias for InsertChar.
func (r *Rope) InsertCharAt(pos int, ch rune) *Rope {
	return r.InsertChar(pos, ch)
}

// RemoveChar removes a single character at the specified position.
// Returns a new Rope, leaving the original unchanged.
//
// This is equivalent to Delete(pos, pos+1).
func (r *Rope) RemoveChar(pos int) *Rope {
	if r == nil {
		return r
	}
	return r.Delete(pos, pos+1)
}

// DeleteChar is an alias for RemoveChar.
func (r *Rope) DeleteChar(pos int) *Rope {
	return r.RemoveChar(pos)
}

// ========== Character Replacement ==========

// ReplaceChar replaces a single character at the specified position.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) ReplaceChar(pos int, ch rune) *Rope {
	if r == nil {
		return New(string(ch))
	}
	return r.Replace(pos, pos+1, string(ch))
}

// SwapChar swaps two characters at the specified positions.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) SwapChar(pos1, pos2 int) *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	if pos1 < 0 || pos1 >= r.Length() || pos2 < 0 || pos2 >= r.Length() {
		panic("character position out of bounds")
	}

	if pos1 == pos2 {
		return r
	}

	// Get the two characters
	ch1 := r.CharAt(pos1)
	ch2 := r.CharAt(pos2)

	// Replace them
	result := r.ReplaceChar(pos1, ch2)
	result = result.ReplaceChar(pos2, ch1)

	return result
}

// ========== Character Query ==========

// ContainsChar checks if the rope contains the specified character.
func (r *Rope) ContainsChar(ch rune) bool {
	if r == nil {
		return false
	}
	it := r.NewIterator()
	for it.Next() {
		if it.Current() == ch {
			return true
		}
	}
	return false
}

// IndexOfChar returns the first position of the specified character.
// Returns -1 if the character is not found.
func (r *Rope) IndexOfChar(ch rune) int {
	if r == nil {
		return -1
	}
	it := r.NewIterator()
	for it.Next() {
		if it.Current() == ch {
			return it.Position()
		}
	}
	return -1
}

// IndexOfCharFrom returns the first position of the character starting from pos.
// Returns -1 if the character is not found.
func (r *Rope) IndexOfCharFrom(pos int, ch rune) int {
	if r == nil || pos < 0 {
		return -1
	}

	for i := pos; i < r.Length(); i++ {
		if r.CharAt(i) == ch {
			return i
		}
	}
	return -1
}

// LastIndexOfChar returns the last position of the specified character.
// Returns -1 if the character is not found.
func (r *Rope) LastIndexOfChar(ch rune) int {
	if r == nil {
		return -1
	}
	for i := r.Length() - 1; i >= 0; i-- {
		if r.CharAt(i) == ch {
			return i
		}
	}
	return -1
}

// LastIndexOfCharBefore returns the last position of the character before pos.
// Returns -1 if the character is not found.
func (r *Rope) LastIndexOfCharBefore(pos int, ch rune) int {
	if r == nil || pos <= 0 {
		return -1
	}

	if pos > r.Length() {
		pos = r.Length()
	}

	for i := pos - 1; i >= 0; i-- {
		if r.CharAt(i) == ch {
			return i
		}
	}
	return -1
}

// CountChar counts the occurrences of a character in the rope.
func (r *Rope) CountChar(ch rune) int {
	if r == nil {
		return 0
	}

	count := 0
	it := r.NewIterator()
	for it.Next() {
		if it.Current() == ch {
			count++
		}
	}
	return count
}

// ========== Character Collection ==========

// CollectChars collects all characters into a rune slice.
func (r *Rope) CollectChars() []rune {
	if r == nil || r.Length() == 0 {
		return []rune{}
	}

	runes := make([]rune, 0, r.Length())
	it := r.NewIterator()
	for it.Next() {
		runes = append(runes, it.Current())
	}
	return runes
}

// ToRunes is an alias for CollectChars.
func (r *Rope) ToRunes() []rune {
	return r.CollectChars()
}

// UniqueChars returns a slice of unique characters in the rope.
func (r *Rope) UniqueChars() []rune {
	if r == nil || r.Length() == 0 {
		return []rune{}
	}

	seen := make(map[rune]bool)
	var unique []rune

	it := r.NewIterator()
	for it.Next() {
		ch := it.Current()
		if !seen[ch] {
			seen[ch] = true
			unique = append(unique, ch)
		}
	}

	return unique
}

// ========== Character Transformations ==========

// MapChars maps each character through a function.
// Returns a new Rope with the transformed characters.
func (r *Rope) MapChars(fn func(rune) rune) *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	b := NewBuilder()
	it := r.NewIterator()
	for it.Next() {
		b.AppendRune(fn(it.Current()))
	}
	return b.Build()
}

// FilterChars filters characters by a predicate function.
// Returns a new Rope with only the characters that satisfy the predicate.
func (r *Rope) FilterChars(fn func(rune) bool) *Rope {
	if r == nil || r.Length() == 0 {
		return Empty()
	}

	b := NewBuilder()
	it := r.NewIterator()
	for it.Next() {
		ch := it.Current()
		if fn(ch) {
			b.AppendRune(ch)
		}
	}
	return b.Build()
}

// RemoveChars removes all occurrences of the specified characters.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) RemoveChars(charsToRemove ...rune) *Rope {
	if r == nil || len(charsToRemove) == 0 {
		return r
	}

	removeSet := make(map[rune]bool)
	for _, ch := range charsToRemove {
		removeSet[ch] = true
	}

	b := NewBuilder()
	it := r.NewIterator()
	for it.Next() {
		ch := it.Current()
		if !removeSet[ch] {
			b.AppendRune(ch)
		}
	}
	return b.Build()
}

// ReplaceChar replaces all occurrences of oldChar with newChar.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) ReplaceAllChar(oldChar, newChar rune) *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	return r.MapChars(func(ch rune) rune {
		if ch == oldChar {
			return newChar
		}
		return ch
	})
}

// ReverseChars reverses all characters in the rope.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) ReverseChars() *Rope {
	if r == nil || r.Length() <= 1 {
		return r
	}

	runes := r.ToRunes()
	// Reverse in place
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	b := NewBuilder()
	for _, r := range runes {
		b.AppendRune(r)
	}
	return b.Build()
}

// ========== Character Categories ==========

// IsWhitespace checks if a rune is whitespace.
func IsWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// IsDigit checks if a rune is a decimal digit.
func IsDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// IsLetter checks if a rune is a letter.
func IsLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// IsLower checks if a rune is lowercase.
func IsLower(ch rune) bool {
	return ch >= 'a' && ch <= 'z'
}

// IsUpper checks if a rune is uppercase.
func IsUpper(ch rune) bool {
	return ch >= 'A' && ch <= 'Z'
}

// CountWhitespace counts whitespace characters in the rope.
func (r *Rope) CountWhitespace() int {
	if r == nil {
		return 0
	}

	count := 0
	it := r.NewIterator()
	for it.Next() {
		if IsWhitespace(it.Current()) {
			count++
		}
	}
	return count
}

// CountDigits counts digit characters in the rope.
func (r *Rope) CountDigits() int {
	if r == nil {
		return 0
	}

	count := 0
	it := r.NewIterator()
	for it.Next() {
		if IsDigit(it.Current()) {
			count++
		}
	}
	return count
}

// CountLetters counts letter characters in the rope.
func (r *Rope) CountLetters() int {
	if r == nil {
		return 0
	}

	count := 0
	it := r.NewIterator()
	for it.Next() {
		if IsLetter(it.Current()) {
			count++
		}
	}
	return count
}

// TrimLeftChar removes leading characters that satisfy the predicate.
func (r *Rope) TrimLeftChar(fn func(rune) bool) *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	it := r.NewIterator()
	start := 0
	for it.Next() {
		if !fn(it.Current()) {
			break
		}
		start++
	}

	if start == 0 {
		return r
	}
	return New(r.Slice(start, r.Length()))
}

// TrimRightChar removes trailing characters that satisfy the predicate.
func (r *Rope) TrimRightChar(fn func(rune) bool) *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	end := r.Length()
	for end > 0 {
		if !fn(r.CharAt(end - 1)) {
			break
		}
		end--
	}

	if end == r.Length() {
		return r
	}
	return New(r.Slice(0, end))
}

// TrimChar removes leading and trailing characters that satisfy the predicate.
func (r *Rope) TrimChar(fn func(rune) bool) *Rope {
	return r.TrimLeftChar(fn).TrimRightChar(fn)
}

// TrimLeftWhitespace removes leading whitespace.
func (r *Rope) TrimLeftWhitespace() *Rope {
	return r.TrimLeftChar(IsWhitespace)
}

// TrimRightWhitespace removes trailing whitespace.
func (r *Rope) TrimRightWhitespace() *Rope {
	return r.TrimRightChar(IsWhitespace)
}

// TrimWhitespace removes leading and trailing whitespace.
func (r *Rope) TrimWhitespace() *Rope {
	return r.TrimChar(IsWhitespace)
}
