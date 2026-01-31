// Code: texere/pkg/rope/builder_pattern.go
//
// This file documents the builder pattern error handling strategy.
// It explains the mixed API design and provides guidance for future improvements.

/*
BUILDER PATTERN ERROR HANDLING
================================

Current Design (Mixed API):
----------------------------

The RopeBuilder has a MIXED API for error handling:

1. FLUENT METHODS (no error return):
   - Append(text) - Appends to end, returns *Builder
   - AppendLine(text) - Appends line with newline, returns *Builder
   - Insert(pos, text) - Queues insert operation, returns *Builder

   These methods don't return errors because:
   - Append/AppendLine always succeed (no bounds checking needed)
   - Insert is BATCHED - operations are queued and applied during Build()
   - Errors are deferred until Build() is called

2. ERROR-RETURNING METHODS:
   - Delete(start, end) (*Builder, error) - Applied immediately, validates bounds
   - Replace(start, end, text) (*Builder, error) - Applied immediately
   - Build() (*Rope, error) - Flushes pending operations and validates

   These methods return errors because:
   - They perform immediate bounds checking
   - They need to validate the current rope state
   - Errors cannot be deferred

WHY THIS MIXED DESIGN?
-----------------------

Option A: All methods return errors
  ❌ Breaks fluent API (requires: b, _ = b.Append("a"); b, _ = b.Append("b"))
  ❌ Verbose for common success case
  ✅ Explicit error handling

Option B: No methods return errors (current Append/Insert style)
  ❌ Can't report errors to caller
  ❌ Silent failures or panics
  ✅ Clean fluent API

Option C: Store error internally (like bytes.Buffer)
  ✅ Clean fluent API
  ✅ Error accessible via Error() method
  ❌ Errors might be missed if caller forgets to check
  ❌ Unclear which operation failed

Current Design (Option A + B hybrid):
  ✅ Fluent API for common operations (Append, Insert)
  ✅ Explicit errors for operations that can fail (Delete, Replace, Build)
  ✅ Batch operations are efficient (Insert queued)
  ❌ Inconsistent - some methods return errors, some don't

USAGE EXAMPLES:
--------------

// Pattern 1: Append-only (no errors to handle)
builder := NewBuilder()
builder.Append("Hello")
builder.Append(" ")
builder.Append("World")
result, err := builder.Build() // Check error only once at the end
if err != nil {
    return err
}

// Pattern 2: With Insert (no errors to handle)
builder := NewBuilder()
builder.Append("HelloWorld")
builder.Insert(5, " ") // Queued, no error here
result, err := builder.Build() // Error checked here
if err != nil {
    return err
}

// Pattern 3: With Delete (error handling required)
builder := NewBuilder()
builder.Append("Hello Beautiful World")
b, err := builder.Delete(6, 16) // Must handle error here
if err != nil {
    return err
}
result, err := b.Build() // And check Build() too
if err != nil {
    return err
}

FUTURE IMPROVEMENTS:
--------------------

Option 1: Make Insert also return error
  - Pros: Consistent with Delete/Replace
  - Cons: Loses fluent API for inserts

Option 2: Add Error() method (like bytes.Buffer)
  - Pros: Clean fluent API, error accessible
  - Cons: Easy to forget checking error

  Example:
    builder := NewBuilder()
    builder.Append("Hello")
    builder.Delete(100, 200) // Stores error internally
    result, err := builder.Build()
    if err != nil {
        return err
    }
    if builderErr := builder.Error(); builderErr != nil {
        return builderErr
    }

Option 3: Separate validation methods
  - Add Validate() method to check if operations would succeed
  - Keep fluent API, error only on Build()
  - Cons: Requires two passes (validate then build)

Option 4: Keep current design but document clearly
  - Add comprehensive documentation
  - Provide wrapper functions for common patterns
  - Accept the inconsistency as pragmatic choice

RECOMMENDATION:
--------------

For now, keep Option 4 (current design) but:

1. Add clear documentation to each method explaining when errors occur
2. Provide helper functions for common patterns:
   - BuildFromOps(ops []func(*Builder) error) (*Rope, error)
   - SafeBuild(fn func(*Builder)) (*Rope, error)

3. Consider Option 2 (Error() method) for v2.0:
   - Would allow all methods to be fluent
   - Error checking is explicit but optional
   - Pattern works well in Go (see: bytes.Buffer, strings.Builder)

4. Add deprecation warnings if API will change:
   - Document which methods will change in future versions
   - Provide migration guide

PERFORMANCE CONSIDERATIONS:
---------------------------

The current batch design is efficient because:
1. Multiple Insert operations are coalesced
2. Only one tree rebalance needed (during Build/flush)
3. Minimal allocations for bulk operations

Example:
  // WITHOUT batching (hypothetical):
  r = r.Insert(5, "a")  // Rebalance
  r = r.Insert(6, "b")  // Rebalance
  r = r.Insert(7, "c")  // Rebalance
  // 3 rebalances

  // WITH batching (current):
  b := NewBuilder()
  b.Insert(5, "a")  // Just queue
  b.Insert(6, "b")  // Just queue
  b.Insert(7, "c")  // Just queue
  r, _ := b.Build()  // Single flush/rebalance
  // 1 rebalance

THREAD SAFETY:
-------------

The RopeBuilder is NOT thread-safe:
- Do not call methods on the same builder from multiple goroutines
- Each goroutine should have its own builder
- Use mutex if concurrent access is needed

Example:
  // ❌ WRONG: Concurrent access
  var b *RopeBuilder
  go func() { b.Append("a") }()
  go func() { b.Append("b") }()
  // Data race!

  // ✅ CORRECT: Separate builders
  var wg sync.WaitGroup
  results := make(chan *Rope, 2)
  wg.Add(2)
  go func() {
      defer wg.Done()
      b := NewBuilder()
      b.Append("a")
      r, _ := b.Build()
      results <- r
  }()
  go func() {
      defer wg.Done()
      b := NewBuilder()
      b.Append("b")
      r, _ := b.Build()
      results <- r
  }()
  wg.Wait()
  close(results)
  // No data race
*/
package rope

// This file contains only documentation.
// No code is needed.
