package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/texere-rope/pkg/rope"
)

func main() {
	// Create a large text (1MB)
	large := strings.Repeat("a", 1024*1024)
	r := rope.New(large)

	// Insert in the middle
	insertPos := 512 * 1024
	textToInsert := "INSERTED"
	r2 := r.Insert(insertPos, textToInsert)

	fmt.Printf("Text to insert: %q\n", textToInsert)
	fmt.Printf("Length of text to insert: %d bytes\n", len(textToInsert))
	fmt.Printf("Rune count of text to insert: %d\n", utf8.RuneCountInString(textToInsert))
	fmt.Printf("\n")

	fmt.Printf("Original rope length: %d\n", r.Length())
	fmt.Printf("After insert rope length: %d\n", r2.Length())
	fmt.Printf("Difference: %d\n", r2.Length()-r.Length())
	fmt.Printf("\n")

	// Check the content around the insertion point
	result := r2.String()

	// Check 10 characters before and after insertion point
	start := insertPos - 10
	if start < 0 {
		start = 0
	}
	end := insertPos + len(textToInsert) + 10
	if end > len(result) {
		end = len(result)
	}

	fmt.Printf("Context around insertion (bytes %d-%d):\n", start, end)
	fmt.Printf("%q\n", result[start:end])

	// Count characters in the context
	contextRunes := []rune(result[start:end])
	fmt.Printf("Context as runes (%d chars):\n", len(contextRunes))
	for i, r := range contextRunes {
		fmt.Printf("  [%d] %q (U+%04X)\n", i, r, r)
	}
	fmt.Printf("\n")

	// Check if there's any unexpected character
	// Look specifically at position 524288 (insertion point)
	fmt.Printf("Character at insertion point (%d):\n", insertPos)
	resultRunes := []rune(result)
	if insertPos < len(resultRunes) {
		fmt.Printf("  %q (U+%04X)\n", resultRunes[insertPos], resultRunes[insertPos])
	}
	if insertPos-1 >= 0 && insertPos-1 < len(resultRunes) {
		fmt.Printf("Character before: %q (U+%04X)\n", resultRunes[insertPos-1], resultRunes[insertPos-1])
	}
	fmt.Printf("\n")

	// Find the inserted text
	pos := strings.Index(result, textToInsert)
	fmt.Printf("Found inserted text at byte position: %d\n", pos)

	// Check character position
	charPos := 0
	for i := range result {
		if i == pos {
			fmt.Printf("Found inserted text at character position: %d\n", charPos)
			break
		}
		charPos++
	}
}
