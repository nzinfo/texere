package concordia

import (
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// randomString generates a random string for testing.
// Corresponds to ot.js test/helpers.js: randomString
func randomString(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		if rand.Float64() < 0.15 {
			b.WriteRune('\n')
		} else {
			b.WriteRune('a' + rune(rand.Intn(26)))
		}
	}
	return b.String()
}

// randomOperation generates a random operation for testing.
// Corresponds to ot.js test/helpers.js: randomOperation
//
// This function uses the exact logic from ot.js to ensure test compatibility.
// It tracks operation.baseLength instead of document position to handle
// Insert operations correctly (which don't advance the base length).
func randomOperation(str string) *Operation {
	operation := NewOperation()

	for {
		left := len(str) - operation.BaseLength()
		if left == 0 {
			break
		}

		// Random length between 1 and min(left-1, 20), ensuring we don't consume all remaining chars
		// This allows the loop to continue with more operations
		maxLen := min(left-1, 20)
		if maxLen < 1 {
			maxLen = 1
		}
		l := 1 + rand.Intn(maxLen)

		r := rand.Float64()

		switch {
		case r < 0.2:
			// Insert - doesn't change baseLength, so loop continues
			s := randomString(l)
			operation.Insert(s)
		case r < 0.4:
			// Delete - increases baseLength (consumes characters from base)
			operation.Delete(l)
		default:
			// Retain - increases baseLength
			operation.Retain(l)
		}
	}

	// 30% chance to insert at the end
	if rand.Float64() < 0.3 {
		operation.Insert(randomString(1 + rand.Intn(10)))
	}

	return operation
}

// min returns the minimum of two integers.
