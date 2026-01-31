package rope_test

import (
	"fmt"

	"github.com/coreseekdev/texere/pkg/rope"
)

/*
This file contains usage examples for the rope package.
These examples demonstrate common patterns and best practices.
*/

func ExampleRope_basic() {
	// Create a new rope from a string
	r := rope.New("Hello, World!")

	// Query length
	fmt.Printf("Characters: %d\n", r.Length())      // Characters: 13
	fmt.Printf("Bytes: %d\n", r.LengthBytes())     // Bytes: 13
	fmt.Printf("String: %s\n", r.String())         // String: Hello, World!
}

func ExampleRope_insert() {
	r := rope.New("Hello World!")

	// Insert text at position 6
	updated, err := r.Insert(6, "Beautiful ")
	if err != nil {
		panic(err)
	}
	fmt.Println(updated.String()) // Hello Beautiful World!
}

func ExampleRope_delete() {
	r := rope.New("Hello Beautiful World!")

	// Delete from position 6 to 16 (removes "Beautiful ")
	updated, err := r.Delete(6, 16)
	if err != nil {
		panic(err)
	}
	fmt.Println(updated.String()) // Hello World!
}

func ExampleRope_replace() {
	r := rope.New("Hello World!")

	// Replace "World" with "Go"
	updated, err := r.Replace(6, 11, "Go")
	if err != nil {
		panic(err)
	}
	fmt.Println(updated.String()) // Hello Go!
}

func ExampleRope_slice() {
	r := rope.New("Hello, World!")

	// Get substring from position 0 to 5 (character positions)
	s, err := r.Slice(0, 5)
	if err != nil {
		panic(err)
	}
	fmt.Println(s) // Hello
}

func ExampleRope_concat() {
	r1 := rope.New("Hello, ")
	r2 := rope.New("World!")

	// Concatenate two ropes
	r3 := r1.Concat(r2)
	fmt.Println(r3.String()) // Hello, World!
}

func ExampleRope_split() {
	r := rope.New("Hello World")

	// Split at position 6
	left, right, err := r.Split(6)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Left: %s, Right: %s\n", left.String(), right.String())
	// Left: Hello , Right: World
}

func ExampleRope_unicode() {
	// Rope handles Unicode correctly
	r := rope.New("Hello 世界 🌍")

	fmt.Printf("Characters: %d\n", r.Length())      // Characters: 10
	fmt.Printf("Bytes: %d\n", r.LengthBytes())      // Bytes: 18
	fmt.Printf("String: %s\n", r.String())          // String: Hello 世界 🌍
}

func ExampleRope_search() {
	r := rope.New("Hello World")

	// Check if contains substring
	fmt.Println(r.Contains("World"))  // true
	fmt.Println(r.Contains("Worlds")) // false

	// Find position
	fmt.Println(r.Index("World"))    // 6
	fmt.Println(r.Index("Worlds"))   // -1 (not found)

	// Find last occurrence
	fmt.Println(r.LastIndex("o"))    // 7
	fmt.Println(r.LastIndex("xyz"))  // -1 (not found)
}

func ExampleRope_builder() {
	// Build a rope efficiently
	builder := rope.NewBuilder()
	builder.Append("Hello")
	builder.Append(" ")
	builder.Append("World")

	result, err := builder.Build()
	if err != nil {
		panic(err)
	}
	fmt.Println(result.String()) // Hello World
}

func ExampleRope_builder_withInsert() {
	builder := rope.NewBuilder()
	builder.Append("HelloWorld")

	// Insert at position 5 (batched operation)
	builder.Insert(5, " ") // Note: Insert is batched, error checked during Build()

	result, err := builder.Build()
	if err != nil {
		panic(err)
	}
	fmt.Println(result.String()) // HelloWorld
}

func ExampleRope_iterator() {
	r := rope.New("Hello")

	// Iterate over runes
	it := r.NewIterator()
	for it.Next() {
		fmt.Printf("%c\n", it.Current())
	}
	// Output:
	// H
	// e
	// l
	// l
	// o
}

func ExampleRope_iterator_withIndex() {
	r := rope.New("ABC")

	// Iterate with index
	it := r.NewIterator()
	for it.Next() {
		pos := it.Position() - 1 // Position returns next position
		char := it.Current()
		fmt.Printf("Index %d: %c\n", pos, char)
	}
	// Output:
	// Index 0: A
	// Index 1: B
	// Index 2: C
}

func ExampleRope_reverseIterator() {
	r := rope.New("Hello")

	// Iterate in reverse
	it := r.IterReverse()
	for it.Next() {
		fmt.Printf("%c\n", it.Current())
	}
	// Output:
	// o
	// l
	// l
	// e
	// H
}

func ExampleRope_errorHandling() {
	r := rope.New("Hello")

	// This will return an error (position out of bounds)
	_, err := r.Insert(100, "!")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		// Error: Insert: position 100 out of bounds (valid range: [0, 6])
	}

	// Safe insertion
	updated, err := r.Insert(5, " World!")
	if err != nil {
		panic(err)
	}
	fmt.Println(updated.String()) // Hello World!
}

func ExampleRope_validation() {
	r := rope.New("Hello")

	// Validate rope structure (check tree integrity)
	if err := r.Validate(); err != nil {
		fmt.Printf("Invalid rope: %v\n", err)
	} else {
		fmt.Println("Rope is valid")
	}
}

func ExampleRope_balancing() {
	// Create an unbalanced rope through many insertions
	r := rope.New("")
	for i := 0; i < 1000; i++ {
		var err error
		r, err = r.Insert(r.Length(), "a")
		if err != nil {
			panic(err)
		}
	}

	// Check if balanced
	fmt.Printf("Is balanced: %v\n", r.IsBalanced()) // May be false
	fmt.Printf("Depth: %d\n", r.Depth())           // May be deep

	// Balance the rope
	balanced := r.Balance()
	fmt.Printf("Is balanced after: %v\n", balanced.IsBalanced()) // true
	fmt.Printf("Depth after: %d\n", balanced.Depth())             // Lower
}

func ExampleRope_bytesIteration() {
	r := rope.New("Hello")

	// Iterate over bytes
	it := r.NewBytesIterator()
	for it.Next() {
		fmt.Printf("Byte %d: %c\n", it.Position()-1, it.Current())
	}
	// Output:
	// Byte 0: H
	// Byte 1: e
	// Byte 2: l
	// Byte 3: l
	// Byte 4: o
}

func ExampleRope_charAt() {
	r := rope.New("Hello")

	// Get character at position
	ch, err := r.CharAt(1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Character at 1: %c\n", ch) // Character at 1: e
}

func ExampleRope_immutability() {
	r := rope.New("Hello")

	// Insert returns a NEW rope, original is unchanged
	updated, err := r.Insert(5, " World!")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Original: %s\n", r.String())         // Original: Hello
	fmt.Printf("Updated: %s\n", updated.String())     // Updated: Hello World!
}

func ExampleRope_clone() {
	r := rope.New("Hello")

	// Clone creates a new rope (but since ropes are immutable,
	// this is very cheap - no actual copying)
	cloned := r.Clone()

	// Both point to the same underlying data
	fmt.Println(r.String() == cloned.String()) // true
}

func ExampleRope_concatMultiple() {
	// Efficiently concatenate multiple ropes
	parts := []*rope.Rope{
		rope.New("Hello"),
		rope.New(" "),
		rope.New("World"),
		rope.New("!"),
	}

	// Build using builder for efficiency
	builder := rope.NewBuilder()
	for _, part := range parts {
		builder.Append(part.String())
	}
	result, err := builder.Build()
	if err != nil {
		panic(err)
	}

	fmt.Println(result.String()) // Hello World!
}

func ExampleRope_map() {
	r := rope.New("hello")

	// Transform each character
	upper, err := r.Map(func(ch rune) rune {
		// Convert to uppercase
		if ch >= 'a' && ch <= 'z' {
			return ch - ('a' - 'A')
		}
		return ch
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(upper.String()) // HELLO
}

func ExampleRope_filter() {
	r := rope.New("Hello123World!")

	// Keep only alphabetic characters
	filtered, err := r.Filter(func(ch rune) bool {
		return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(filtered.String()) // HelloWorld
}

func ExampleRope_count() {
	r := rope.New("Hello World")

	// Count characters matching predicate
	count := r.Count(func(ch rune) bool {
		return ch == 'l' || ch == 'o'
	})

	fmt.Printf("Count of 'l' and 'o': %d\n", count) // Count of 'l' and 'o': 3
}

func ExampleRope_lines() {
	r := rope.New("Line 1\nLine 2\nLine 3")

	// Split into lines
	lines := r.Lines()
	for i, line := range lines {
		fmt.Printf("Line %d: %s", i, line)
	}
	// Output:
	// Line 0: Line 1
	// Line 1: Line 2
	// Line 2: Line 3
}
