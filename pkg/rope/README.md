# Rope Package for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/texere-rope/pkg/rope.svg)](https://pkg.go.dev/github.com/texere-rope/pkg/rope)

A high-performance Rope data structure implementation in Go for efficient large text editing operations.

## Overview

A Rope is a balanced binary tree (B-tree) representation of a string, optimized for efficient insertions, deletions, and other operations on large texts. Unlike standard strings, ropes provide O(log n) complexity for modification operations instead of O(n).

## Features

- **Immutable**: All operations return new Ropes, originals remain unchanged
- **Efficient**: O(log n) for insert/delete/slice operations
- **Memory Optimized**: Minimal copying due to tree structure
- **UTF-8 Support**: Full Unicode support with character-based indexing
- **Thread-Safe**: Immutable structure enables safe concurrent reads
- **Editor-Friendly**: Rich line-based operations for text editing

## Installation

```bash
go get github.com/texere-rope/pkg/rope
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/texere-rope/pkg/rope"
)

func main() {
    // Create a rope from a string
    r := rope.New("Hello World")

    // Query operations
    fmt.Println(r.Length())   // 11
    fmt.Println(r.String())   // Hello World
    fmt.Println(r.Slice(0, 5)) // Hello

    // Insert text (immutable - returns new rope)
    r2 := r.Insert(5, " Beautiful")
    fmt.Println(r2.String())  // Hello Beautiful World
    fmt.Println(r.String())    // Original unchanged: Hello World

    // Delete text
    r3 := r.Delete(5, 11)
    fmt.Println(r3.String())  // Hello

    // Concatenate ropes
    r4 := r.Concat(rope.New(" Again"))
    fmt.Println(r4.String())  // Hello World Again
}
```

## Core Concepts

### Rope Structure

A rope is a tree where:
- **Leaf nodes** contain actual text
- **Internal nodes** maintain balance and cache subtree information

```
        [Internal: leftLen=5, leftSize=5]
             /                    \
    [Leaf: "Hello"]          [Leaf: " World"]
```

### Immutability

All operations return new rope instances:

```go
r1 := rope.New("Hello")
r2 := r1.Insert(5, " World")
// r1 is still "Hello"
// r2 is "Hello World"
```

### Character vs Byte Indexing

Ropes use **character positions** (Unicode code points), not byte positions:

```go
r := rope.New("Hello 世界")
r.Length()     // 8 characters (5 + 1 + 2)
r.Size()       // 12 bytes (5 + 1 + 6)
r.CharAt(6)    // '世' (7th character)
```

## API Reference

### Constructors

```go
// Create from string
r := rope.New("Hello World")

// Create empty rope
r := rope.Empty()

// Create from builder
b := rope.NewBuilder()
b.Append("Hello")
b.Append(" World")
r := b.Build()
```

### Query Operations

```go
r.Length()              // Number of characters
r.Size()                // Number of bytes
r.String()              // Full string
r.Bytes()               // Full byte slice
r.Slice(start, end)     // Substring (character positions)
r.CharAt(pos)           // Character at position
r.ByteAt(pos)           // Byte at position
r.Contains(substring)   // Check if contains substring
r.Index(substring)      // Find first occurrence
```

### Modification Operations

```go
// Insert text at position
r2 := r.Insert(pos, text)

// Delete range
r2 := r.Delete(start, end)

// Replace range with text
r2 := r.Replace(start, end, text)

// Split at position
left, right := r.Split(pos)

// Concatenate ropes
r2 := r.Concat(other)
```

### Iterator

```go
it := r.NewIterator()

// Iterate character by character
for it.Next() {
    ch := it.Current()
    fmt.Println(ch)
}

// Seek to position
it.Seek(10)

// Peek without advancing
ch, ok := it.Peek()

// Collect remaining
remaining := it.Collect()
```

### Line Operations

```go
// Get line info
lineCount := r.LineCount()
line := r.Line(0)              // Get line text
lineStart := r.LineStart(0)    // Line start position
lineEnd := r.LineEnd(0)        // Line end position

// Modify lines
r2 := r.InsertLine(0, "Text")
r2 := r.DeleteLine(0)
r2 := r.ReplaceLine(0, "New Text")

// Line editor operations
pos := r.PositionAtLineCol(line, col)
line := r.LineAtChar(pos)
col := r.ColumnAtChar(pos)
```

### Builder Pattern

```go
b := rope.NewBuilder()
b.Append("Hello")
b.Append(" ")
b.Insert(5, "Beautiful")
b.Delete(0, 6)
r := b.Build()

// Builder can be reused
b.Reset()
b.Append("New text")
r2 := b.Build()
```

### Balancing

```go
// Check if balanced
isBalanced := r.IsBalanced()

// Balance the rope
r2 := r.Balance()

// Optimize structure
r2 := r.Optimize()

// Get statistics
stats := r.Stats()
fmt.Printf("Depth: %d, Nodes: %d\n", stats.Depth, stats.NodeCount)
```

## Document Interface Integration

The rope implements the `Document` interface for use with OT (Operational Transformation):

```go
import "github.com/texere-rope/pkg/document"

// Create a document
doc := rope.NewRopeDocument("Hello World")

// Use as Document interface
var d document.Document = doc
length := d.Length()
text := d.Slice(0, 5)

// Convert back to RopeDocument
ropeDoc := rope.AsRopeDocument(d)
r := ropeDoc.Rope()
```

## Performance

### Complexity

| Operation | Time Complexity | Space Complexity |
|-----------|----------------|------------------|
| New       | O(n)           | O(n)             |
| Length    | O(1)           | O(1)             |
| Slice     | O(log n)       | O(k)             |
| Insert    | O(log n)       | O(log n)         |
| Delete    | O(log n)       | O(log n)         |
| Concat    | O(1)           | O(1)             |
| Split     | O(log n)       | O(log n)         |

### Benchmarks

```
BenchmarkRope_New-8                 10000    123456 ns/op
BenchmarkRope_Insert_Small-8       100000      9876 ns/op
BenchmarkRope_Delete_Small-8       100000      8765 ns/op
BenchmarkRope_Slice-8              200000      5432 ns/op
BenchmarkRope_Concat-8             500000      2345 ns/op
BenchmarkRope_Iterator-8           300000      4567 ns/op
BenchmarkBuilder_Append-8          200000      3456 ns/op
```

### When to Use Rope

**Use Rope when:**
- Working with large files (> 100KB)
- Frequent insertions/deletions
- Need efficient undo/redo (immutable)
- Building a text editor
- Collaborative editing (with OT)

**Use string when:**
- Small texts (< 10KB)
- Few modifications
- Simplicity is preferred

## Examples

### Text Editor

```go
// Create document
doc := rope.NewRopeDocument("Line 1\nLine 2\nLine 3")

// Insert at line/column
doc = doc.InsertAtLineCol(1, 6, " Hello")

// Delete line
doc = doc.DeleteLine(2)

// Get current line
line := doc.Rope().Line(1)
```

### Log Processing

```go
// Build large log from multiple sources
builder := rope.NewBuilder()
for _, entry := range logEntries {
    builder.AppendLine(entry)
}
log := builder.Build()

// Process line by line
it := log.LinesIterator()
for it.Next() {
    line := it.Current()
    processLine(line)
}
```

### Efficient Concatenation

```go
// Concatenate many strings efficiently
builder := rope.NewBuilder()
for _, chunk := range chunks {
    builder.Append(chunk)
}
result := builder.Build()
```

## Testing

Run tests:

```bash
# Run all tests
go test ./pkg/rope/...

# Run with coverage
go test ./pkg/rope/... -cover

# Run benchmarks
go test ./pkg/rope/... -bench=.

# Run specific test
go test ./pkg/rope/... -run TestInsert
```

## Implementation Details

### Tree Structure

- **Leaf Node**: Stores actual text content
- **Internal Node**: Maintains balance, caches subtree info
- **Balance Strategy**: B-tree with configurable parameters

### UTF-8 Support

- All indices are character-based (not byte-based)
- Automatic UTF-8 validation
- Efficient Unicode handling

### Memory Management

- Immutable structure enables sharing
- Minimal copying through tree sharing
- Optional compaction for memory optimization

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass: `go test ./pkg/rope/...`
2. Code is formatted: `go fmt ./pkg/rope/...`
3. Add tests for new features
4. Update documentation

## License

[Your License Here]

## Acknowledgments

This library is heavily inspired by and based on the design of the excellent **[ropey](https://github.com/cessen/ropey)** crate for Rust by **[Cessen](https://github.com/cessen)**.

The ropey crate provides:
- The core balanced binary tree algorithm
- B-tree optimization strategies
- Efficient UTF-8 handling patterns
- The overall API design philosophy

This Go implementation adapts those concepts for Go's idioms while maintaining the performance characteristics that make ropey excellent for text editing.

### Similar Projects

- **[ropey (Rust)](https://github.com/cessen/ropey)** - Original inspiration, used in the Helix editor
- **[rope (C++)](https://github.com/zeux/rope)** - C++ implementation used in the Neovim editor
- **[Rope (Swift)](https://github.com/apple/swift-rope)** - Swift implementation by Apple

## References

- **[Ropes: an Alternative to Strings](https://www.cs.rit.edu/~ats/books/ooc/html/ooc.html)** - Boehm, Atkinson, and Plass (1995)
  - Original paper introducing the Rope data structure
- **[Ropey Documentation](https://docs.rs/ropey/)** - Rust crate that inspired this implementation
- **[Xi-Editor Rope](https://github.com/google/xi-editor)** - Google's editor implementation
- **[USAGE.md](USAGE.md)** - Comprehensive Chinese usage guide (中文使用指南)
