package rope

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByteCacheBug(t *testing.T) {
	text := strings.Repeat("Hello, World! ", 100)
	cache := NewBytePosCache(text)
	pos := 50

	// This should not panic
	result := cache.GetBytePos(pos)
	t.Logf("Position 50 -> byte %d", result)

	assert.True(t, result >= 0)
	assert.True(t, result < len(text))
}
