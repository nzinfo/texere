package rope

// ========== String Utilities ==========

// CommonPrefix returns the length of the common prefix between two ropes.
func (r *Rope) CommonPrefix(other *Rope) int {
	if r == nil || other == nil {
		return 0
	}

	it1 := r.NewIterator()
	it2 := other.NewIterator()
	count := 0

	for it1.Next() && it2.Next() {
		if it1.Current() != it2.Current() {
			break
		}
		count++
	}

	return count
}

// CommonPrefixString returns the common prefix string of two ropes.
func (r *Rope) CommonPrefixString(other *Rope) string {
	length := r.CommonPrefix(other)
	if length == 0 {
		return ""
	}
	return r.Slice(0, length)
}

// CommonSuffix returns the length of the common suffix between two ropes.
func (r *Rope) CommonSuffix(other *Rope) int {
	if r == nil || other == nil {
		return 0
	}

	len1 := r.Length()
	len2 := other.Length()
	count := 0

	for count < len1 && count < len2 {
		if r.CharAt(len1-1-count) != other.CharAt(len2-1-count) {
			break
		}
		count++
	}

	return count
}

// CommonSuffixString returns the common suffix string of two ropes.
func (r *Rope) CommonSuffixString(other *Rope) string {
	length := r.CommonSuffix(other)
	if length == 0 {
		return ""
	}
	return r.Slice(r.Length()-length, r.Length())
}

// ========== String Comparison ==========

// StartsWith checks if the rope starts with the given prefix.
func (r *Rope) StartsWith(prefix string) bool {
	if r == nil {
		return prefix == ""
	}
	if len(prefix) == 0 {
		return true
	}
	if r.Length() < len(prefix) {
		return false
	}

	it := r.NewIterator()
	for _, ch := range prefix {
		if !it.Next() || it.Current() != ch {
			return false
		}
	}
	return true
}

// StartsWithRope checks if the rope starts with another rope.
func (r *Rope) StartsWithRope(prefix *Rope) bool {
	if r == nil {
		return prefix == nil || prefix.Length() == 0
	}
	if prefix == nil || prefix.Length() == 0 {
		return true
	}
	if r.Length() < prefix.Length() {
		return false
	}

	it1 := r.NewIterator()
	it2 := prefix.NewIterator()

	for it1.Next() && it2.Next() {
		if it1.Current() != it2.Current() {
			return false
		}
	}

	return !it2.Next() // Prefix should be exhausted
}

// EndsWith checks if the rope ends with the given suffix.
func (r *Rope) EndsWith(suffix string) bool {
	if r == nil {
		return suffix == ""
	}
	if len(suffix) == 0 {
		return true
	}
	if r.Length() < len(suffix) {
		return false
	}

	start := r.Length() - len(suffix)
	for i, ch := range suffix {
		if r.CharAt(start+i) != ch {
			return false
		}
	}
	return true
}

// EndsWithRope checks if the rope ends with another rope.
func (r *Rope) EndsWithRope(suffix *Rope) bool {
	if r == nil {
		return suffix == nil || suffix.Length() == 0
	}
	if suffix == nil || suffix.Length() == 0 {
		return true
	}
	if r.Length() < suffix.Length() {
		return false
	}

	start := r.Length() - suffix.Length()
	it1 := r.NewIterator()
	it1.Seek(start)
	it2 := suffix.NewIterator()

	for it1.Next() && it2.Next() {
		if it1.Current() != it2.Current() {
			return false
		}
	}

	return !it2.Next() // Suffix should be exhausted
}

// ========== String Transformations ==========

// ToUpper converts all characters to uppercase.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) ToUpper() *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	return r.MapChars(func(ch rune) rune {
		// Simple uppercase conversion for ASCII
		// For full Unicode support, use unicode.ToUpper
		if ch >= 'a' && ch <= 'z' {
			return ch - ('a' - 'A')
		}
		return ch
	})
}

// ToLower converts all characters to lowercase.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) ToLower() *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	return r.MapChars(func(ch rune) rune {
		// Simple lowercase conversion for ASCII
		// For full Unicode support, use unicode.ToLower
		if ch >= 'A' && ch <= 'Z' {
			return ch + ('a' - 'A')
		}
		return ch
	})
}

// Capitalize capitalizes the first character of the rope.
func (r *Rope) Capitalize() *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	b := NewBuilder()
	it := r.NewIterator()
	first := true

	for it.Next() {
		ch := it.Current()
		if first {
			if ch >= 'a' && ch <= 'z' {
				ch = ch - ('a' - 'A')
			}
			first = false
		}
		b.AppendRune(ch)
	}

	return b.Build()
}

// Title capitalizes the first character of each word.
func (r *Rope) Title() *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	b := NewBuilder()
	it := r.NewIterator()
	newWord := true

	for it.Next() {
		ch := it.Current()
		if ch == ' ' || ch == '\t' || ch == '\n' {
			newWord = true
		} else if newWord {
			if ch >= 'a' && ch <= 'z' {
				ch = ch - ('a' - 'A')
			}
			newWord = false
		}
		b.AppendRune(ch)
	}

	return b.Build()
}

// SwapCase swaps uppercase to lowercase and vice versa.
func (r *Rope) SwapCase() *Rope {
	if r == nil || r.Length() == 0 {
		return r
	}

	return r.MapChars(func(ch rune) rune {
		if ch >= 'a' && ch <= 'z' {
			return ch - ('a' - 'A')
		}
		if ch >= 'A' && ch <= 'Z' {
			return ch + ('a' - 'A')
		}
		return ch
	})
}

// ========== String Padding ==========

// PadLeft pads the rope on the left with the given character to reach the target length.
func (r *Rope) PadLeft(targetLen int, padChar rune) *Rope {
	if r == nil {
		r = Empty()
	}

	currentLen := r.Length()
	if currentLen >= targetLen {
		return r
	}

	padding := targetLen - currentLen
	b := NewBuilder()

	// Add padding
	for i := 0; i < padding; i++ {
		b.AppendRune(padChar)
	}

	// Add original content
	b.Append(r.String())

	return b.Build()
}

// PadRight pads the rope on the right with the given character.
func (r *Rope) PadRight(targetLen int, padChar rune) *Rope {
	if r == nil {
		r = Empty()
	}

	currentLen := r.Length()
	if currentLen >= targetLen {
		return r
	}

	padding := targetLen - currentLen
	result := r.Clone()

	// Add padding
	for i := 0; i < padding; i++ {
		result = result.Append(string(padChar))
	}

	return result
}

// PadCenter centers the rope with padding on both sides.
func (r *Rope) PadCenter(targetLen int, padChar rune) *Rope {
	if r == nil {
		r = Empty()
	}

	currentLen := r.Length()
	if currentLen >= targetLen {
		return r
	}

	padding := targetLen - currentLen
	leftPadding := padding / 2
	rightPadding := padding - leftPadding

	b := NewBuilder()

	// Left padding
	for i := 0; i < leftPadding; i++ {
		b.AppendRune(padChar)
	}

	// Original content
	b.Append(r.String())

	// Right padding
	for i := 0; i < rightPadding; i++ {
		b.AppendRune(padChar)
	}

	return b.Build()
}

// ========== String Truncation ==========

// Truncate truncates the rope to the specified length.
// If ellipsis is true and truncation occurs, adds "..." at the end.
func (r *Rope) Truncate(maxLen int, ellipsis bool) *Rope {
	if r == nil || r.Length() <= maxLen {
		return r
	}

	if !ellipsis {
		return New(r.Slice(0, maxLen))
	}

	ellipsisLen := 3
	if maxLen <= ellipsisLen {
		return New("...")
	}

	return New(r.Slice(0, maxLen-ellipsisLen)).Append("...")
}

// TruncateLeft truncates from the left side.
func (r *Rope) TruncateLeft(maxLen int, ellipsis bool) *Rope {
	if r == nil || r.Length() <= maxLen {
		return r
	}

	if !ellipsis {
		return New(r.Slice(r.Length()-maxLen, r.Length()))
	}

	ellipsisLen := 3
	if maxLen <= ellipsisLen {
		return New("...")
	}

	return New("...").Append(r.Slice(r.Length()-maxLen+ellipsisLen, r.Length()))
}

// TruncateCenter truncates from the middle, replacing with ellipsis.
func (r *Rope) TruncateCenter(maxLen int, ellipsisStr string) *Rope {
	if r == nil || r.Length() <= maxLen {
		return r
	}

	if ellipsisStr == "" {
		ellipsisStr = "..."
	}

	ellipsisLen := len([]rune(ellipsisStr))
	if maxLen <= ellipsisLen {
		return New(ellipsisStr)
	}

	leftLen := (maxLen - ellipsisLen) / 2
	rightLen := maxLen - ellipsisLen - leftLen

	return New(r.Slice(0, leftLen)).Append(ellipsisStr).Append(
		r.Slice(r.Length()-rightLen, r.Length()),
	)
}

// ========== String Splitting ==========

// SplitBySep splits the rope by a separator string (different from Split(pos)).
func (r *Rope) SplitBySep(separator string) []*Rope {
	if r == nil {
		return []*Rope{}
	}

	if separator == "" {
		// Split by each character
		result := make([]*Rope, r.Length())
		it := r.NewIterator()
		i := 0
		for it.Next() {
			result[i] = New(string(it.Current()))
			i++
		}
		return result
	}

	// Find all occurrences
	var parts []*Rope
	start := 0
	sepRunes := []rune(separator)
	sepLen := len(sepRunes)

	for {
		idx := r.IndexOfCharFrom(start, sepRunes[0])
		if idx == -1 {
			// No more separators
			parts = append(parts, New(r.Slice(start, r.Length())))
			break
		}

		// Check if full separator matches
		match := true
		for i := 0; i < sepLen; i++ {
			if idx+i >= r.Length() || r.CharAt(idx+i) != sepRunes[i] {
				match = false
				break
			}
		}

		if match {
			// Found separator
			parts = append(parts, New(r.Slice(start, idx)))
			start = idx + sepLen
		} else {
			start = idx + 1
		}

		if start >= r.Length() {
			break
		}
	}

	if len(parts) == 0 {
		return []*Rope{r.Clone()}
	}

	return parts
}

// SplitBySepN splits the rope into at most N parts.
func (r *Rope) SplitBySepN(separator string, n int) []*Rope {
	if n <= 0 {
		return []*Rope{r.Clone()}
	}

	parts := r.SplitBySep(separator)
	if len(parts) <= n {
		return parts
	}

	// Merge the remaining parts
	result := parts[:n]
	merged := Concat(parts[n:]...)
	result[n-1] = merged

	return result
}

// ========== String Joining ==========

// JoinStrings joins strings with the rope as separator.
func (r *Rope) JoinStrings(parts []string) *Rope {
	if len(parts) == 0 {
		return Empty()
	}
	if len(parts) == 1 {
		return New(parts[0])
	}

	b := NewBuilder()
	b.Append(parts[0])

	for i := 1; i < len(parts); i++ {
		b.Append(r.String())
		b.Append(parts[i])
	}

	return b.Build()
}

// JoinRopesWith joins ropes with the rope as separator.
func (r *Rope) JoinRopesWith(parts []*Rope) *Rope {
	if len(parts) == 0 {
		return Empty()
	}
	if len(parts) == 1 {
		return parts[0].Clone()
	}

	result := parts[0].Clone()
	for i := 1; i < len(parts); i++ {
		result = result.AppendRope(r)
		result = result.AppendRope(parts[i])
	}

	return result
}

// ========== Repeat ==========

// Repeat repeats the rope n times.
func (r *Rope) Repeat(n int) *Rope {
	if r == nil || r.Length() == 0 || n <= 0 {
		return Empty()
	}
	if n == 1 {
		return r.Clone()
	}

	// Build balanced tree for efficiency
	ropes := make([]*Rope, n)
	for i := 0; i < n; i++ {
		ropes[i] = r
	}

	return Concat(ropes...)
}

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
		ru := it.Current()
		byteIdx += lenRune(ru)
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
		ru := it.Current()
		runeLen := lenRune(ru)
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
