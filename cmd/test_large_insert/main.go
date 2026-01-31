package main

import (
	"fmt"
	"strings"

	"github.com/texere-rope/pkg/rope"
)

func main() {
	// Create a large text (1MB)
	large := strings.Repeat("a", 1024*1024)
	r := rope.New(large)

	fmt.Printf("Original rope length: %d (expected: %d)\n", r.Length(), 1024*1024)
	fmt.Printf("Original rope size: %d bytes\n", r.Size())

	// Insert in the middle
	insertPos := 512 * 1024
	textToInsert := "INSERTED"
	r2 := r.Insert(insertPos, textToInsert)

	fmt.Printf("\nAfter Insert(%d, %q):\n", insertPos, textToInsert)
	fmt.Printf("New rope length: %d (expected: %d)\n", r2.Length(), 1024*1024+7)
	fmt.Printf("New rope size: %d bytes (expected: %d)\n", r2.Size(), 1024*1024+len(textToInsert))
	fmt.Printf("Difference: %d\n", r2.Length()-1024*1024)

	// Check if inserted text is present
	result := r2.String()
	if strings.Contains(result, textToInsert) {
		fmt.Printf("✓ Inserted text found\n")

		// Find position
		pos := strings.Index(result, textToInsert)
		fmt.Printf("  Position in string: %d\n", pos)
	} else {
		fmt.Printf("✗ Inserted text NOT found\n")
	}

	// Verify the rope has the correct content
	expectedLength := 1024*1024 + len(textToInsert)
	actualLength := r2.Length()
	if actualLength != expectedLength {
		fmt.Printf("\n✗ LENGTH MISMATCH!\n")
		fmt.Printf("  Expected: %d\n", expectedLength)
		fmt.Printf("  Actual:   %d\n", actualLength)
		fmt.Printf("  Diff:     %d\n", actualLength-expectedLength)
	} else {
		fmt.Printf("\n✓ Length is correct\n")
	}
}
