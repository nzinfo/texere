# Refactoring TODO

This file tracks deferred refactoring tasks and documents completed work.

## Completed Refactoring (2026-01)

### 1. Length Methods Unification ✅
**Status:** COMPLETED

Added explicit `LengthBytes()` and `LengthChars()` methods across all packages:
- `pkg/rope/rope.go`: Added `LengthBytes()` and `LengthChars()` methods
- `pkg/ot/document.go`: Updated interface to include length methods
- `pkg/ot/string_document.go`: Implemented length methods using `unicode/utf8`
- `pkg/concordia/document.go`: Added length methods delegating to rope

**Decision:** Kept `Length()` returning characters for backward compatibility (Plan B).

### 2. Error Handling - Replaced Panics ✅
**Status:** COMPLETED

Created `pkg/rope/errors.go` with structured error types:
- `ErrOutOfBounds`: For index/position errors
- `ErrInvalidRange`: For invalid range errors
- `ErrIteratorState`: For iterator state errors
- `ErrInvalidInput`: For invalid input parameters

Updated all core operations to return errors instead of panicking:
- `Slice(start, end int) (string, error)`
- `CharAt(pos int) (rune, error)`
- `ByteAt(pos int) (byte, error)`
- `Insert(pos int, text string) (*Rope, error)`
- `Delete(start, end int) (*Rope, error)`
- `Replace(start, end int, text string) (*Rope, error)`
- `Split(pos int) (*Rope, *Rope, error)`

Updated 15+ files in rope package to propagate errors:
- `builder.go`, `reverse_iter.go`, `line_ops.go`, `text_char.go`, `text_graphemes.go`
- `chunk_ops.go`, `changeset.go`, `composition.go`, `hash.go`
- `insert_optimized.go`, `cow_optimization.go`, `text_word_boundary.go`
- `text_utf16.go`, `rope_split.go`, `profiling.go`, `micro_optimizations.go`
- `text_crlf.go`, `balance.go`, `rope_concat.go`, `rope_io.go`

Updated dependent packages:
- `pkg/concordia/document.go`: Updated to handle errors from rope operations
- `pkg/ot/`: Already compatible (no changes needed)

### 3. Interface Segregation (ISP) ✅
**Status:** COMPLETED

Created `pkg/rope/interfaces.go` with focused interfaces:
- `ReadOnlyDocument`: Read-only content access
- `CharAtAccessor`: Character-by-character access
- `ByteAtAccessor`: Byte-by-byte access
- `MutableDocument`: Document modification operations
- `SplittableDocument`: Split operations
- `Concatenable`: Concatenation operations
- `Cloneable`: Cloning operations
- `Searchable`: Search operations
- `Validatable`: Validation operations
- `Balanceable`: Balance operations
- `DocumentMetrics`: Document structure metrics

Composite interfaces:
- `FullDocument`: All capabilities combined
- `ReadOnly`: Read capabilities including search
- `ReadWrite`: Read and write capabilities
- `Editable`: Mutation and splitting capabilities

## Recently Completed (2026-01)

### 5. Documentation and Examples ✅
**Status:** COMPLETED

Added comprehensive documentation:
- `pkg/rope/naming.go` - API naming conventions reference
- `pkg/rope/builder_pattern.go` - Builder pattern error handling strategy
- `pkg/rope/examples_test.go` - Comprehensive usage examples (30+ examples)

### 6. Test Suite Updates ✅
**Status:** COMPLETED

Updated all test files to handle error returns:
- Fixed 20 test files to use new API with error returns
- All tests compile and pass successfully
- Performance benchmarks verified working

### 7. Performance Baseline Established ✅
**Status:** COMPLETED

Ran benchmarks to establish baseline performance:
- SplitOff_Small: 267 ns/op, 96 B/op, 4 allocs/op
- SplitOff_Medium: 1701 ns/op, 96 B/op, 4 allocs/op
- SplitOff_Large: 16875 ns/op, 96 B/op, 4 allocs/op

## Recently Completed (2026-01) - Continued

### 8. Documentation Improvements ✅
**Status:** COMPLETED

Added comprehensive godoc documentation:
- Enhanced package-level documentation with usage guidelines
- Documented when to use Rope vs String (10KB threshold)
- Added performance characteristics table with O notation
- Documented thread-safety guarantees (immutable, concurrent reads safe)
- Added usage examples throughout

### 9. Iterator Pooling ✅
**Status:** COMPLETED

Enhanced iterator pooling implementation:
- Added pooled variants for all iterator types
- NewIteratorPooled(), IterReversePooled(), NewBytesIteratorPooled()
- Added corresponding release functions
- Added benchmarks comparing pooled vs non-pooled performance

### 10. Builder Pattern Refactoring ✅
**Status:** COMPLETED

Implemented Option B for builder pattern (error accumulation):
- Added `err` field to `RopeBuilder` and `Error()` method
- All builder methods now maintain fluent API (return `*RopeBuilder`)
- Errors are stored internally and accessible via `Error()`
- Updated `DocumentBuilder` in concordia to match pattern
- Added bounds checking to `InsertFast()` and `DeleteFast()`

### 11. Iterator Unification ✅
**Status:** COMPLETED

Created unified iterator interfaces in `pkg/rope/iterator_interfaces.go`:
- Core generic interfaces: `Seq[T]`, `PositionalSeq[T]`, `FullSeq[T]`, etc.
- Type-specific behavior interfaces for each iterator type
- Go 1.23+ adapter functions: `IterRunes()`, `IterBytes()`, `IterGraphemes()`, etc.
- Named interfaces "Seq" to avoid conflict with existing `Iterator` type

## Deferred Tasks
**Reason:** Defer for later as it requires significant design work

### File Reorganization
**Reason:** Requires separate project, too many files to reorganize

The `pkg/rope` directory has 50+ files. Some files could be reorganized:
- Test files that were already merged/renamed
- `*_test.go` files should each correspond to a source file
- Consider grouping related functionality:
  - `rope_*.go` → `core/`, `ops/`, `iter/`, `utils/`

**Future Work:**
1. Audit all 50+ files in `pkg/rope`
2. Group into logical subdirectories:
   ```
   pkg/rope/
   ├── core.go          # Main Rope type and core operations
   ├── node.go          # RopeNode, LeafNode, InternalNode
   ├── builder.go       # RopeBuilder
   ├── ops/             # Operations (insert, delete, replace, split)
   ├── iter/            # Iterators (forward, reverse, bytes, chunks)
   ├── search/          # Search operations
   ├── text/            # Text operations (chars, graphemes, words, lines)
   ├── utils/           # Utilities (validation, metrics, balancing)
   └── errors.go        # Error types
   ```
3. Update all imports across the codebase
4. Run full test suite to ensure nothing broke

### 13. Documentation Improvements ✅
**Status:** COMPLETED

Enhanced documentation for Rope public APIs with examples and clarifications:

**Functional Methods Enhanced (rope.go):**
- `ForEach` - Added description and usage example
- `ForEachWithIndex` - Clarified 0-based index
- `Map` - Added example showing uppercase conversion
- `Filter` - Added example showing vowel filtering
- `Count` - Added example showing digit counting

**Utility Methods Enhanced (rope.go):**
- `Lines` - Clarified line ending preservation
- `Contains` - Added note about grapheme search alternatives
- `Index`/`LastIndex` - Clarified character vs byte position
- `Compare`/`Equals` - Enhanced with examples

**Word Boundary Detection Enhanced (text_word_boundary.go):**
- `WordBoundary` type - Added comprehensive description
- `IsWordChar` - Clarified \w pattern matching
- `IsWhitespace` - Documented Unicode support
- `PrevWordStart/NextWordStart/PrevWordEnd/NextWordEnd` - Added examples

**Documentation Standards Applied:**
- Clear descriptions of method behavior
- Parameter descriptions where applicable
- Return value descriptions
- Usage examples for common use cases
- Notes on edge case behavior
- Thread-safety and immutability notes (already documented in package godoc)

### Iterator Unification ✅
**Status:** COMPLETED

Created unified iterator interfaces in `pkg/rope/iterator_interfaces.go`:

**Core Interfaces:**
- `Seq[T any]` - Minimal iterator interface (Next + Current)
- `PositionalSeq[T]` - Adds position tracking
- `ResettableSeq[T]` - Adds reset capability
- `StatefulSeq[T]` - Adds state query (HasNext, IsExhausted)
- `SeekableSeq[T]` - Adds random access (Seek)
- `PeekableSeq[T]` - Adds lookahead (Peek)
- `CollectingSeq[T]` - Adds bulk collection (Collect)
- `FullSeq[T]` - Combines all capabilities

**Type-Specific Interfaces:**
- `RuneIteratorBehavior` - Character iteration
- `ReverseIteratorBehavior` - Reverse character iteration
- `BytesIteratorBehavior` - Byte iteration
- `LinesIteratorBehavior` - Line iteration
- `GraphemeIteratorBehavior` - Grapheme cluster iteration

**Go 1.23+ Compatibility:**
Added adapter functions for use with for-range loops:
- `IterRunes(r)` - Iterate over runes
- `IterBytes(r)` - Iterate over bytes
- `IterGraphemes(r)` - Iterate over grapheme clusters
- `IterLines(r)` - Iterate over lines
- `IterReverse(r)` - Iterate in reverse

**Note:** Named interfaces "Seq" instead of "Iterator" to avoid conflict with existing `Iterator` type.

### 12. API Naming Consistency ✅
**Status:** COMPLETED

Audited all method names for consistency and added deprecation notices:

**Deprecated Methods:**
- `Size()` → Use `LengthBytes()` instead
- `ToRunes()` → Use `Runes()` instead
- `InsertCharAt()` → Use `InsertChar()` instead
- `RemoveChar()` → Use `DeleteChar()` instead (for consistency with Delete operations)
- `GraphemeIterator.ToSlice()` → Use `GraphemeIterator.Collect()` instead

**Naming Improvements:**
- Inverted `DeleteChar`/`RemoveChar` relationship:
  - `DeleteChar` is now the primary implementation (directly calls `Delete`)
  - `RemoveChar` is now the deprecated alias (calls `DeleteChar`)
  - This follows the naming convention: deletion operations use `Delete*()`

**Documentation Updates:**
- Updated `pkg/rope/naming.go` with complete list of deprecated methods
- Added deprecation notices to all alias methods with clear guidance on alternatives
- All deprecated methods maintain backward compatibility

### Performance Optimization Opportunities
**Reason:** Code already has optimizations, these are future enhancements

Potential optimizations identified:
1. **Iterator Pooling**: Reuse iterator objects instead of allocating
2. **Lazy Evaluation**: Defer string conversion until absolutely needed
3. **Caching**: Add more caching for frequently accessed data
4. **Memory Pooling**: Reuse node objects in operations

**Future Work:**
1. Profile with `pprof` to identify actual bottlenecks
2. Benchmark with realistic workloads
3. Optimize only hot paths identified by profiling
4. Add benchmark tests for performance regressions

## Build Status

All packages build successfully:
```bash
go build ./pkg/rope/...   ✅
go build ./pkg/ot/...      ✅
go build ./pkg/concordia/... ✅
go build ./...              ✅
```

## Testing Notes

- Many test files exist but may need updates for new error returns
- Test files previously reorganized (orphaned tests merged)
- `byte_char_conv_test.go` has comprehensive UTF-8 conversion tests
- Consider adding integration tests for error handling paths

## Migration Guide for Consumers

If you're using this package, here's how to update your code:

### Before (panics on errors):
```go
r := rope.New("Hello World")
r = r.Insert(5, " Beautiful")  // Would panic if out of bounds
```

### After (handle errors):
```go
r := rope.New("Hello World")
r, err := r.Insert(5, " Beautiful")
if err != nil {
    // Handle error: out of bounds, etc.
}
```

### Document API Changes:

**Before:**
```go
doc := concordia.NewRopeDocument("Hello")
doc = doc.Insert(5, " World")  // No error handling
```

**After:**
```go
doc := concordia.NewRopeDocument("Hello")
doc, err := doc.Insert(5, " World")
if err != nil {
    // Handle error
}
```

### Slice Behavior (Document interface compatibility):

The `ot.Document` interface's `Slice()` method still returns `string` (not error) for compatibility.
Errors are handled internally by returning empty string:

```go
// This is safe - returns "" on error
s := doc.Slice(0, 5)

// Direct rope Slice returns error
s, err := doc.Rope().Slice(0, 5)
```

## Next Steps

1. **Immediate:** Update any consumers of rope/ot/concordia to handle errors
2. **Short-term:** Add comprehensive error handling tests
3. **Medium-term:** Implement deferred tasks above
4. **Long-term:** Performance profiling and optimization
