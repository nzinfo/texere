package main

import (
	"fmt"
	"unicode/utf8"

	"github.com/texere-rope/pkg/rope"
)

func main() {
	s := "Hello Beautiful World"

	fmt.Printf("Original string: %q\n", s)
	fmt.Printf("Length (bytes): %d\n", len(s))
	fmt.Printf("Length (runes): %d\n", utf8.RuneCountInString(s))
	fmt.Printf("\n")

	// Show each character with its position
	fmt.Println("Character positions:")
	runes := []rune(s)
	for i, r := range runes {
		fmt.Printf("  Position %2d: %q (%U)\n", i, r)
	}
	fmt.Printf("\n")

	// Test slicing
	fmt.Println("Slice tests:")
	fmt.Printf("  s[0:5]  = %q (Hello)\n", string(runes[0:5]))
	fmt.Printf("  s[5:6]  = %q (space)\n", string(runes[5:6]))
	fmt.Printf("  s[5:15] = %q ( Beautiful)\n", string(runes[5:15]))
	fmt.Printf("  s[5:16] = %q ( Beautiful )\n", string(runes[5:16]))
	fmt.Printf("  s[6:15] = %q (Beautiful)\n", string(runes[6:15]))
	fmt.Printf("  s[6:16] = %q (Beautiful )\n", string(runes[6:16]))
	fmt.Printf("  s[15:16] = %q (space)\n", string(runes[15:16]))
	fmt.Printf("  s[16:21] = %q (World)\n", string(runes[16:21]))
	fmt.Printf("\n")

	// Test reconstruction
	fmt.Println("Reconstruction tests:")
	result1 := string(runes[0:5]) + string(runes[16:21])
	fmt.Printf("  s[0:5] + s[16:21] = %q\n", result1)

	result2 := string(runes[0:5]) + string(runes[15:21])
	fmt.Printf("  s[0:5] + s[15:21] = %q\n", result2)

	result3 := string(runes[0:5]) + string(runes[6:15]) + string(runes[16:21])
	fmt.Printf("  s[0:5] + s[6:15] + s[16:21] = %q\n", result3)
	fmt.Printf("\n")

	// Test Rope.Delete
	fmt.Println("Rope.Delete tests:")
	r := rope.New(s)
	fmt.Printf("  Original rope: %q\n", r.String())

	r1 := r.Delete(5, 16)
	fmt.Printf("  Delete(5, 16): %q\n", r1.String())
	fmt.Printf("  Expected:      %q\n", "Hello World")
	fmt.Printf("  Match: %v\n", r1.String() == "Hello World")

	r2 := r.Delete(5, 15)
	fmt.Printf("  Delete(5, 15): %q\n", r2.String())

	r3 := r.Delete(6, 16)
	fmt.Printf("  Delete(6, 16): %q\n", r3.String())
	fmt.Printf("  Expected:      %q\n", "Hello World")
	fmt.Printf("  Match: %v\n", r3.String() == "Hello World")
}
