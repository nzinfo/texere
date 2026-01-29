package rope

import (
	"bytes"
)

// ========== CRLF-Aware Operations ==========

// IsCRLF checks if the byte sequence at the given position is a CRLF (\r\n).
func IsCRLF(data []byte, pos int) bool {
	if pos < 0 || pos+1 >= len(data) {
		return false
	}
	return data[pos] == '\r' && data[pos+1] == '\n'
}

// FindGoodSplit finds a good position to split text without breaking CRLF pairs.
// Returns a split position that is safe (doesn't split CRLF).
//
// Parameters:
//   - preferredPos: The preferred position to split at
//   - text: The text to split
//   - minSplit: If true, will return the closest position before preferredPos;
//               if false, will return the closest position after preferredPos
//
// Returns a safe split position.
func FindGoodSplit(preferredPos int, text []byte, minSplit bool) int {
	if len(text) == 0 {
		return 0
	}

	// Clamp preferredPos to valid range
	if preferredPos < 0 {
		preferredPos = 0
	}
	if preferredPos > len(text) {
		preferredPos = len(text)
	}

	// Check if we're at a boundary (start or end)
	if preferredPos == 0 || preferredPos == len(text) {
		return preferredPos
	}

	// Check if preferred position is good
	if IsGoodSplit(preferredPos, text) {
		return preferredPos
	}

	// Position is not good, find nearest good position
	if minSplit {
		// Find closest position before
		for i := preferredPos - 1; i > 0; i-- {
			if IsGoodSplit(i, text) {
				return i
			}
		}
		// No good position found, return 0
		return 0
	} else {
		// Find closest position after
		for i := preferredPos + 1; i < len(text); i++ {
			if IsGoodSplit(i, text) {
				return i
			}
		}
		// No good position found, return end
		return len(text)
	}
}

// IsGoodSplit checks if splitting at the given position would break a CRLF pair.
// Returns true if it's safe to split at this position.
func IsGoodSplit(pos int, text []byte) bool {
	if len(text) == 0 {
		return true
	}
	if pos <= 0 || pos >= len(text) {
		return true
	}

	// Check if position is between \r and \n
	if text[pos-1] == '\r' && pos < len(text) && text[pos] == '\n' {
		return false // Would split CRLF
	}

	return true // Safe to split
}

// FindBadSplit finds a position that would split a CRLF pair.
// Returns -1 if no such position exists.
func FindBadSplit(text []byte) int {
	for i := 1; i < len(text); i++ {
		if !IsGoodSplit(i, text) {
			return i
		}
	}
	return -1
}

// HasBadSplits checks if the text has any bad split points.
func HasBadSplits(text []byte) bool {
	return FindBadSplit(text) >= 0
}

// CountCRLF counts the number of CRLF sequences in the text.
func CountCRLF(text []byte) int {
	count := 0
	for i := 0; i < len(text)-1; i++ {
		if text[i] == '\r' && text[i+1] == '\n' {
			count++
		}
	}
	return count
}

// CountCRLFInRope counts CRLF sequences in a rope.
func (r *Rope) CountCRLF() int {
	if r == nil || r.Length() == 0 {
		return 0
	}

	data := r.String()
	return CountCRLF([]byte(data))
}

// ========== CRLF-Aware Splitting ==========

// SplitCRLFSafe splits text at preferredPos without breaking CRLF pairs.
// Returns (left, right) parts.
func SplitCRLFSafe(text []byte, preferredPos int) ([]byte, []byte) {
	safePos := FindGoodSplit(preferredPos, text, false)
	return text[:safePos], text[safePos:]
}

// SplitCRLFSafeString splits a string without breaking CRLF pairs.
func SplitCRLFSafeString(text string, preferredPos int) (string, string) {
	safePos := FindGoodSplit(preferredPos, []byte(text), false)
	return text[:safePos], text[safePos:]
}

// ========== CRLF Detection ==========

// HasCRLF checks if text contains any CRLF sequences.
func HasCRLF(text []byte) bool {
	return bytes.Count(text, []byte{'\r', '\n'}) > 0
}

// HasCRLFString checks if a string contains CRLF.
func HasCRLFString(text string) bool {
	return HasCRLF([]byte(text))
}

// DetectLineEnding detects the dominant line ending style in text.
// Returns "CRLF" for \r\n, "LF" for \n, "CR" for \r, or "NONE" if no line endings.
func DetectLineEnding(text []byte) string {
	crlfCount := bytes.Count(text, []byte{'\r', '\n'})
	lfCount := bytes.Count(text, []byte{'\n'}) - crlfCount
	crCount := bytes.Count(text, []byte{'\r'}) - crlfCount

	if crlfCount > lfCount && crlfCount > crCount {
		return "CRLF"
	}
	if lfCount >= crlfCount && lfCount > crCount {
		return "LF"
	}
	if crCount > crlfCount && crCount > lfCount {
		return "CR"
	}
	return "NONE"
}

// DetectLineEndingInRope detects the line ending style in a rope.
func (r *Rope) DetectLineEnding() string {
	if r == nil || r.Length() == 0 {
		return "NONE"
	}
	return DetectLineEnding([]byte(r.String()))
}

// ========== CRLF Conversion ==========

// ConvertCRLFToLF converts all CRLF (\r\n) to LF (\n).
func ConvertCRLFToLF(text []byte) []byte {
	return bytes.ReplaceAll(text, []byte{'\r', '\n'}, []byte{'\n'})
}

// ConvertLFToCRLF converts all LF (\n) to CRLF (\r\n).
func ConvertLFToCRLF(text []byte) []byte {
	// First convert any existing CRLF to LF to avoid double conversion
	temp := bytes.ReplaceAll(text, []byte{'\r', '\n'}, []byte{'\n'})
	return bytes.ReplaceAll(temp, []byte{'\n'}, []byte{'\r', '\n'})
}

// ConvertToCRLF converts any line ending style to CRLF.
func ConvertToCRLF(text []byte) []byte {
	// First normalize all line endings to LF
	temp := NormalizeLineEndingsToLF(text)
	// Then convert all LF to CRLF
	return ConvertLFToCRLF(temp)
}

// NormalizeLineEndingsToLF normalizes all line endings to LF.
func NormalizeLineEndingsToLF(text []byte) []byte {
	// Replace CRLF with LF
	result := bytes.ReplaceAll(text, []byte{'\r', '\n'}, []byte{'\n'})
	// Replace any remaining CR with LF
	result = bytes.ReplaceAll(result, []byte{'\r'}, []byte{'\n'})
	return result
}

// ========== CRLF-Aware Rope Operations ==========

// SplitCRLFSafe splits a rope at a position without breaking CRLF pairs.
func (r *Rope) SplitCRLFSafe(pos int) (*Rope, *Rope) {
	if r == nil {
		return Empty(), Empty()
	}

	if pos <= 0 {
		return Empty(), r.Clone()
	}
	if pos >= r.Length() {
		return r.Clone(), Empty()
	}

	// Convert position to bytes
	bytePos := r.charToByte(pos)

	// Check if this would split a CRLF
	str := r.String()
	if bytePos > 0 && bytePos < len(str) &&
		str[bytePos-1] == '\r' && str[bytePos] == '\n' {
		// Would split CRLF, adjust position
		bytePos--
		pos = r.byteToChar(bytePos)
	}

	return r.Split(pos)
}

// JoinWithCRLF joins multiple ropes with CRLF as separator.
func JoinWithCRLF(ropes []*Rope) *Rope {
	if len(ropes) == 0 {
		return Empty()
	}
	if len(ropes) == 1 {
		return ropes[0].Clone()
	}

	crlf := New("\r\n")
	result := ropes[0].Clone()

	for i := 1; i < len(ropes); i++ {
		result = result.AppendRope(crlf)
		result = result.AppendRope(ropes[i])
	}

	return result
}

// ========== CRLF Validation ==========

// ValidateCRLFPairs checks that all CRLF pairs are intact.
// Returns true if no CRLF pairs are broken.
func (r *Rope) ValidateCRLFPairs() bool {
	if r == nil || r.Length() == 0 {
		return true
	}

	str := r.String()
	for i := 0; i < len(str)-1; i++ {
		if str[i] == '\r' && str[i+1] != '\n' {
			// Found \r not followed by \n
			// This might be valid (Mac Classic style), but check if it was intended as CRLF
			// For strict CRLF validation, return false
			return false
		}
	}

	return true
}

// ========== Line Ending Statistics ==========

// LineEndingStats contains statistics about line endings in text.
type LineEndingStats struct {
	CRLF int // \r\n count
	LF   int // \n count (excluding CRLF)
	CR   int // \r count (excluding CRLF)
	Total int // Total line count
}

// LineEndingStats returns statistics about line endings in the rope.
func (r *Rope) LineEndingStats() LineEndingStats {
	if r == nil || r.Length() == 0 {
		return LineEndingStats{}
	}

	str := r.String()
	stats := LineEndingStats{}

	for i := 0; i < len(str); i++ {
		if i < len(str)-1 && str[i] == '\r' && str[i+1] == '\n' {
			stats.CRLF++
			stats.Total++
			i++ // Skip \n
		} else if str[i] == '\n' {
			stats.LF++
			stats.Total++
		} else if str[i] == '\r' {
			stats.CR++
			stats.Total++
		}
	}

	// Handle last line if no trailing newline
	if len(str) > 0 && str[len(str)-1] != '\n' && str[len(str)-1] != '\r' {
		stats.Total++
	}

	return stats
}

// ========== CRLF Utilities ==========

// EnsureTrailingCRLF ensures the rope ends with CRLF.
// If it doesn't, adds CRLF to the end.
func (r *Rope) EnsureTrailingCRLF() *Rope {
	if r == nil {
		return New("\r\n")
	}

	str := r.String()
	if len(str) >= 2 && str[len(str)-2:] == "\r\n" {
		return r
	}

	return r.Append("\r\n")
}

// StripTrailingCRLF removes trailing CRLF (or LF/CR) from the rope.
func (r *Rope) StripTrailingCRLF() *Rope {
	if r == nil {
		return r
	}

	str := r.String()
	end := len(str)

	// Check for CRLF
	if end >= 2 && str[end-2:] == "\r\n" {
		return New(str[:end-2])
	}

	// Check for LF or CR
	if end >= 1 && (str[end-1] == '\n' || str[end-1] == '\r') {
		return New(str[:end-1])
	}

	return r
}
