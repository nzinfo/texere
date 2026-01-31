# Rope Project Quick Start Guide

## Overview

This guide will help you get started with the Rope data structure implementation in under 5 minutes.

## Installation

The Rope package is located at `S:/workspace/texere-rope/pkg/rope`.

```bash
cd S:/workspace/texere-rope
go mod tidy
```

## Your First Rope

### Basic Usage

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
    fmt.Println(r.String())        // "Hello World"
    fmt.Println(r.Length())        // 11
    fmt.Println(r.Slice(0, 5))     // "Hello"

    // Insert (returns new rope, original unchanged)
    r2 := r.Insert(5, " Beautiful")
    fmt.Println(r2.String())       // "Hello Beautiful World"
    fmt.Println(r.String())        // Still "Hello World"

    // Delete
    r3 := r.Delete(5, 16)
    fmt.Println(r3.String())       // "Hello World"
}
```

### Builder Pattern

For building ropes incrementally:

```go
b := rope.NewBuilder()
b.Append("Hello")
b.Append(" ")
b.Append("World")
r := b.Build()
fmt.Println(r.String())  // "Hello World"
```

### Line Operations

Perfect for text editors:

```go
r := rope.New("Line 1\nLine 2\nLine 3")

// Query lines
fmt.Println(r.LineCount())       // 3
fmt.Println(r.Line(0))           // "Line 1"
fmt.Println(r.Line(1))           // "Line 2"

// Modify lines
r2 := r.InsertLine(1, "Inserted")
fmt.Println(r2.Line(1))          // "Inserted"

r3 := r.DeleteLine(2)
fmt.Println(r3.LineCount())      // 2
```

### Iterator

Efficient character-by-character traversal:

```go
r := rope.New("Hello")
it := r.NewIterator()

for it.Next() {
    ch := it.Current()
    fmt.Printf("%c\n", ch)
}

// Or use functional API
r.ForEach(func(ch rune) {
    fmt.Printf("%c\n", ch)
})
```

## Running Tests

```bash
# Run all rope tests
go test ./pkg/rope/... -v

# Run with coverage
go test ./pkg/rope/... -cover

# Run benchmarks
go test ./pkg/rope/... -bench=.

# Run specific test
go test ./pkg/rope/... -run TestInsert
```

## Running the Demo

```bash
cd S:/workspace/texere-rope
go run examples/rope-demo/main.go
```

The demo showcases:
1. Basic operations
2. Immutability
3. UTF-8 support
4. Large text operations
5. Line operations
6. Builder pattern
7. Iterator
8. Document interface
9. Performance comparison

## Document Interface

For integration with OT (Operational Transformation):

```go
import "github.com/texere-rope/pkg/rope"

// Create document
doc := rope.NewRopeDocument("Hello World")

// Use as Document interface
length := doc.Length()
text := doc.Slice(0, 5)

// Modify
doc2 := doc.Insert(5, " Beautiful")

// Convert back to Rope
r := doc.Rope()
```

## Common Patterns

### Concatenating Many Strings

```go
// Efficient for many small strings
builder := rope.NewBuilder()
for _, chunk := range chunks {
    builder.Append(chunk)
}
result := builder.Build()
```

### Large Text Processing

```go
// Process line by line
r := rope.New(largeText)
it := r.LinesIterator()
for it.Next() {
    line := it.Current()
    processLine(line)
}
```

### Text Transformations

```go
r := rope.New("hello world")

// Uppercase
r2 := r.Map(func(ch rune) rune {
    if ch >= 'a' && ch <= 'z' {
        return ch - 32
    }
    return ch
})
// r2 is "HELLO WORLD"

// Filter
r3 := r.Filter(func(ch rune) bool {
    return ch != ' '
})
// r3 is "helloworld"
```

## When to Use Rope

### Use Rope When:
- Working with large files (> 100KB)
- Frequent insertions/deletions
- Building a text editor
- Need efficient undo/redo (immutable)
- Collaborative editing (with OT)

### Use string When:
- Small texts (< 10KB)
- Few modifications
- Simplicity is preferred

## Performance Tips

1. **Use Builder for batch operations**
   ```go
   builder := rope.NewBuilder()
   // Multiple appends
   r := builder.Build()
   ```

2. **Balance periodically for large texts**
   ```go
   if !r.IsBalanced() {
       r = r.Balance()
   }
   ```

3. **Use iterator for sequential access**
   ```go
   it := r.NewIterator()
   for it.Next() {
       // Process
   }
   ```

## API Quick Reference

### Constructors
- `rope.New(text)` - Create from string
- `rope.Empty()` - Create empty rope
- `rope.NewBuilder()` - Create builder

### Query
- `Length()` - Character count
- `Size()` - Byte count
- `String()` - Full string
- `Slice(start, end)` - Substring
- `CharAt(pos)` - Character at position

### Modify
- `Insert(pos, text)` - Insert text
- `Delete(start, end)` - Delete range
- `Replace(start, end, text)` - Replace range
- `Split(pos)` - Split into two
- `Concat(other)` - Concatenate ropes

### Lines
- `LineCount()` - Number of lines
- `Line(n)` - Get line text
- `LineStart(n)` - Line start position
- `LineEnd(n)` - Line end position

### Iterator
- `NewIterator()` - Create iterator
- `Next()` - Advance to next
- `Current()` - Current character
- `Seek(pos)` - Jump to position

## Further Reading

- [Complete API Documentation](pkg/rope/README.md)
- [Initialization Report](ROPE_INITIALIZATION_REPORT.md)
- [Implementation Plan](S:/src.editor/ROPE_IMPLEMENTATION_PLAN.md)

## Getting Help

1. Check the [README](pkg/rope/README.md) for detailed documentation
2. Run the demo: `go run examples/rope-demo/main.go`
3. Look at test cases in `pkg/rope/rope_test.go`
4. Review examples in `examples/rope-demo/main.go`

## Next Steps

1. ✅ Run the tests to verify everything works
2. ✅ Run the demo to see Rope in action
3. ✅ Read the API documentation
4. ✅ Try the examples in your own code
5. ✅ Integrate with your OT layer using Document interface

Happy coding!
