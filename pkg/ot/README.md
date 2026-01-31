# Concordia - Operational Transformation (OT) Library

Concordia is a Go implementation of Operational Transformation (OT) algorithms for real-time collaborative editing. It's based on the proven ot.js JavaScript library and provides the core functionality needed to build collaborative text editors.

## Overview

Concordia implements the core OT operations including:
- **Insert**: Insert text at a position
- **Delete**: Delete text at a position
- **Retain**: Keep text without modification

Operations can be composed, transformed, and applied to documents, ensuring that all concurrent edits converge to a consistent state.

## Installation

```bash
go get github.com/coreseekdev/texere/pkg/concordia
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/coreseekdev/texere/pkg/concordia"
)

func main() {
    // Create operations using the builder pattern
    op1 := concordia.NewBuilder().
        Retain(5).
        Insert("Hello").
        Build()

    op2 := concordia.NewBuilder().
        Retain(5).
        Insert("World").
        Build()

    // Apply operations
    doc := "_____"
    result1, _ := op1.Apply(doc)
    result2, _ := op2.Apply(doc)

    fmt.Println(result1) // "_Hello"
    fmt.Println(result2) // "_World"

    // Transform concurrent operations
    op1Prime, op2Prime, _ := concordia.Transform(op1, op2)

    // Apply in any order - results converge
    final1, _ := op1Prime.Apply(doc)
    final2, _ := op2Prime.Apply(final1)

    fmt.Println(final2) // Consistent result
}
```

## Core Concepts

### Operations

An Operation is a sequence of ops (retain, insert, delete) that transforms a document from one state to another:

```go
// Create an operation that:
// 1. Skips the first 5 characters
// 2. Inserts "Hello"
// 3. Skips the next 3 characters
// 4. Deletes 2 characters
op := concordia.NewBuilder().
    Retain(5).
    Insert("Hello").
    Retain(3).
    Delete(2).
    Build()
```

### Builder Pattern

The `OperationBuilder` provides a fluent API for constructing operations with automatic optimization:

```go
// Adjacent operations are automatically merged
op := concordia.NewBuilder().
    Retain(5).
    Retain(3).      // Merged with previous retain
    Insert("Hello").
    Insert(" World"). // Merged with previous insert
    Build()

// Result: retain(8).insert("Hello World")
```

### Transformation

Transform is the core OT algorithm. It ensures concurrent operations converge:

```go
// Two users edit at position 0
op1 := concordia.NewBuilder().Insert("Hello").Build()
op2 := concordia.NewBuilder().Insert("Hi").Build()

// Transform against each other
op1Prime, op2Prime, _ := concordia.Transform(op1, op2)

// Can be applied in any order
result1, _ := op1Prime.Apply("")
result2, _ := op2Prime.Apply(result1)
// Both orders produce the same result
```

### Composition

Compose combines consecutive operations:

```go
op1 := concordia.NewBuilder().Insert("Hello ").Build()
op2 := concordia.NewBuilder().Retain(6).Insert("World").Build()

composed, _ := concordia.Compose(op1, op2)

// composed is equivalent to Insert("Hello World")
```

### Document Interface

The Document interface allows OT to work with different document representations:

```go
import "github.com/coreseekdev/texere/pkg/document"

// String document (simple, efficient for small documents)
doc := document.NewStringDocument("Hello World")
result, _ := op.ApplyToDocument(doc)

// Future: Rope document (efficient for large documents)
// doc := document.NewRopeDocument(largeContent)
```

### UndoManager

The UndoManager provides undo/redo functionality compatible with collaborative editing:

```go
um := concordia.NewUndoManager(50) // Keep up to 50 operations

// Apply an operation
op := concordia.NewBuilder().Insert("Hello").Build()
doc, _ := op.Apply(doc)

// Add the inverse to the undo stack
inverse, _ := op.Invert(doc)
um.Add(inverse, true) // compose=true to merge consecutive operations

// Undo
um.PerformUndo(func(op *concordia.Operation) {
    doc, _ = op.Apply(doc)
})

// Redo
um.PerformRedo(func(op *concordia.Operation) {
    doc, _ = op.Apply(doc)
})

// Transform undo/redo stack when receiving remote operations
um.Transform(remoteOp)
```

### Client

The Client manages the state for collaborative editing:

```go
client := concordia.NewClient()

// Apply local operation
op := concordia.NewBuilder().Insert("Hello").Build()
doc, _ = client.ApplyClient(op)

// Send to server
sendOp := client.OutgoingOperation()

// Receive server acknowledgment
client.ServerAck()

// Apply remote operation
remoteOp := // ... from server
doc, _ = client.ApplyServer(revision, remoteOp)
```

## API Reference

### Core Types

- `Operation`: Immutable sequence of OT operations
- `OperationBuilder`: Builder for constructing operations with optimization
- `Op`: Interface for operation types (RetainOp, InsertOp, DeleteOp)
- `Document`: Interface for document representations
- `UndoManager`: Manages undo/redo stacks
- `Client`: Client-side state management

### Key Functions

- `NewBuilder() *OperationBuilder`: Create a new operation builder
- `Transform(op1, op2) (*Operation, *Operation, error)`: Transform two operations
- `Compose(op1, op2) (*Operation, error)`: Compose two operations
- `FromJSON(ops []interface{}) (*Operation, error)`: Deserialize from JSON

## Testing

The library includes comprehensive tests based on the ot.js test suite:

```bash
# Run all tests
go test ./pkg/concordia/...

# Run with coverage
go test -cover ./pkg/concordia/...

# Run benchmarks
go test -bench=. ./pkg/concordia/...
```

## Design Decisions

### Type Safety vs JavaScript

Unlike ot.js which uses:
- Positive integers for retain
- Strings for insert
- Negative integers for delete

Concordia uses explicit types:
- `RetainOp(int)` for retain operations
- `InsertOp(string)` for insert operations
- `DeleteOp(int)` for delete operations (internally negative)

This provides better type safety and performance.

### Immutability

Operations are immutable and safe for concurrent use. The builder pattern constructs optimized operations, then returns an immutable `Operation`.

### Document Abstraction

The Document interface allows different underlying representations (string, rope, piece table) without changing OT algorithms.

## Performance

Concordia is optimized for:
- **Type safety**: Compile-time checks prevent common errors
- **Memory efficiency**: Automatic operation merging reduces overhead
- **Concurrent safety**: Immutable operations are safe for concurrent use
- **Builder optimization**: Adjacent operations are merged during construction

## Comparison with ot.js

| Feature | ot.js | Concordia |
|---------|-------|-----------|
| Type system | Dynamic (JavaScript) | Static (Go) |
| Operation representation | Primitives | Explicit types |
| Immutability | Mutable operations | Immutable operations |
| Builder pattern | Chainable methods | Builder with optimization |
| Document interface | String only | Multiple implementations |
| Concurrency | Single-threaded | Thread-safe |

## Future Work

- [ ] Rope document implementation for large files
- [ ] Performance benchmarks
- [ ] Additional operation types (formatting, attributes)
- [ ] Server implementation
- [ ] WebSocket client/server examples

## License

Part of the Texere project.

## Contributing

Contributions are welcome! Please ensure:
- All tests pass
- New features include tests
- Code follows Go best practices
- Documentation is updated

## Etymology

Concordia is the Roman goddess of harmony and agreement. In the context of OT, it represents the harmony achieved by coordinating concurrent edits from multiple users.
