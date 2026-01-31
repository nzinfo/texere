// Code: texere/pkg/rope/naming.go
//
// This file documents the API naming conventions used in the rope package.
// It serves as a reference for understanding the patterns and consistency
// across the API surface.

/*
API NAMING CONVENTIONS
======================

1. POSITION-BASED OPERATIONS
   Pattern: *At(pos int)
   - CharAt(pos) - Get rune at character position
   - ByteAt(pos) - Get byte at byte position
   - LineAt(line) - Get line by line number

2. BYTE-BASED OPERATIONS
   Pattern: *Bytes()
   - LengthBytes() - Get length in bytes
   - Bytes() - Get content as byte slice
   - ByteAt(pos) - Get byte at position

3. CHARACTER/RUNE-BASED OPERATIONS
   Pattern: *Char*() or *Rune()
   - LengthChars() - Get length in characters
   - CharAt(pos) - Get rune at position
   - Runes() - Get all runes
   - InsertChar(pos, rune) - Insert single rune
   - DeleteChar(pos) - Delete single rune
   - ReplaceChar(pos, rune) - Replace single rune

4. GRAPHEME CLUSTER OPERATIONS
   Pattern: *Grapheme*()
   - Graphemes() - Get grapheme iterator
   - GraphemeSlice(start, end) - Slice by grapheme clusters
   - LenGraphemes() - Count grapheme clusters
   - MapGraphemes(fn) - Map over graphemes
   - FilterGraphemes(fn) - Filter graphemes

5. LINE-BASED OPERATIONS
   Pattern: *Line*()
   - Line(lineNum) - Get line content
   - Lines() - Get all lines
   - LineCount() - Count lines
   - InsertLine(lineNum, text) - Insert line
   - DeleteLine(lineNum) - Delete line

6. QUERY OPERATIONS (no mutation)
   Pattern: Is*(), Has*(), Can*()
   - IsBalanced() - Check if balanced
   - IsEmpty() - Check if empty
   - IsLeaf() - Check if node is leaf
   - HasTrailingNewline() - Check for trailing newline
   - CanAppendWithoutRebalancing() - Check if efficient to append

7. METADATA OPERATIONS
   Pattern: Size, Depth, Stats
   - Size() - Get size in bytes (use LengthBytes() instead)
   - Depth() - Get tree depth
   - Stats() - Get detailed tree statistics

8. TRANSFORMATION OPERATIONS (return new Rope)
   - Insert(pos, text) - Insert text (returns *Rope, error)
   - Delete(start, end) - Delete range (returns *Rope, error)
   - Replace(start, end, text) - Replace range (returns *Rope, error)
   - Split(pos) - Split into two (returns *Rope, *Rope, error)
   - Concat(other) - Concatenate (returns *Rope)
   - Clone() - Create copy (returns *Rope)

9. ITERATOR OPERATIONS
   Pattern: *Iterator() or Iter*()
   - NewIterator() - Create new forward rune iterator
   - IterReverse() - Create reverse iterator
   - NewBytesIterator() - Create byte iterator
   - Chunks() - Create chunk iterator
   - Graphemes() - Create grapheme iterator

10. BUILDER OPERATIONS
    Pattern: Method chaining with *Builder
    - NewBuilder() - Create new builder
    - Append(text) - Add to end (returns *Builder)
    - Insert(pos, text) - Insert at position (returns *Builder)
    - Delete(start, end) - Delete range (returns *Builder, error)
    - Build() - Build final Rope (returns *Rope, error)

DEPRECATED NAMES (use alternatives instead)
    - Size() → Use LengthBytes() instead
    - ToRunes() → Use Runes() instead (kept for compatibility)

ERROR HANDLING CONVENTIONS
--------------------------
As of 2026-01, all operations that can fail return (result, error):
- Bounds checking: Insert, Delete, Replace, Split, Slice, CharAt, ByteAt
- Validation: Validate(), Balance(), Build()
- Iterator operations: Next(), Current(), Peek()

Operations that don't fail:
- Length queries: Length(), LengthBytes(), LengthChars()
- Content access: String(), Bytes() (no bounds checking)
- Concatenation: Concat() (always valid)
- Cloning: Clone() (immutable, no copy needed)
- Search: Contains(), Index(), LastIndex() (return -1 if not found)

CONSISTENCY RULES
----------------
1. Always use "start, end" for ranges (end is exclusive)
2. Character positions are 0-indexed
3. Byte positions are 0-indexed
4. Line numbers are 0-indexed
5. All mutation operations return a new Rope (immutable)
6. Methods that can fail MUST return (result, error)
7. Query methods that can't find something return -1 or false, not error
8. Iterator methods follow: Next() bool, Current() T, Reset()
*/
package rope

// This file contains only documentation. No code is needed.
// The conventions are documented here for reference and to ensure
// consistency across future additions to the API.
