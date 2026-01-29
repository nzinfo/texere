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
func randomOperation(str string) *Operation {
	builder := NewBuilder()

	// Track position in the document as we apply the operation
	// This is a simplified version - ot.js uses operation.baseLength
	docPos := 0
	originalLen := len(str)

	for docPos < originalLen {
		left := originalLen - docPos
		if left <= 0 {
			break
		}

		// Random length between 1 and min(left, 20)
		maxLen := min(left, 20)
		l := 1 + rand.Intn(maxLen)

		r := rand.Float64()

		switch {
		case r < 0.2:
			// Insert
			s := randomString(l)
			builder.Insert(s)
			// docPos doesn't change for insert (we insert at current position)
		case r < 0.4:
			// Delete
			builder.Delete(l)
			docPos += l
		default:
			// Retain
			builder.Retain(l)
			docPos += l
		}
	}

	// 30% chance to insert at the end
	if rand.Float64() < 0.3 {
		builder.Insert(randomString(1 + rand.Intn(10)))
	}

	return builder.Build()
}

// min returns the minimum of two integers.
