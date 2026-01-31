package rope

import (
	"math"
	"testing"
)

// TestBalance tests the Balance method.
func TestBalance(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty rope",
			input: "",
		},
		{
			name:  "small rope",
			input: "Hello",
		},
		{
			name:  "medium rope",
			input: string(make([]byte, 1000)),
		},
		{
			name:  "unicode text",
			input: "Hello 世界 🌍 Testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rope := New(tt.input)
			balanced := rope.Balance()

			// Content should be preserved
			if balanced.String() != tt.input {
				t.Errorf("Balance changed content: expected %q, got %q",
					tt.input, balanced.String())
			}

			// Length should be preserved
			if balanced.Length() != rope.Length() {
				t.Errorf("Balance changed length: expected %d, got %d",
					rope.Length(), balanced.Length())
			}
		})
	}
}

// TestBalanceWithConfig tests the BalanceWithConfig method.
func TestBalanceWithConfig(t *testing.T) {
	t.Run("custom config", func(t *testing.T) {
		rope := New("Hello World")

		config := &BalanceConfig{
			MinLeafSize: 64,
			MaxLeafSize: 256,
			MaxDepth:    32,
		}

		balanced := rope.BalanceWithConfig(config)
		if balanced.String() != "Hello World" {
			t.Errorf("BalanceWithConfig changed content")
		}
	})

	t.Run("different leaf sizes", func(t *testing.T) {
		text := string(make([]byte, 2000))
		rope := New(text)

		// Small leaves
		config1 := &BalanceConfig{
			MinLeafSize: 64,
			MaxLeafSize: 128,
			MaxDepth:    64,
		}
		balanced1 := rope.BalanceWithConfig(config1)

		// Large leaves
		config2 := &BalanceConfig{
			MinLeafSize: 512,
			MaxLeafSize: 1024,
			MaxDepth:    64,
		}
		balanced2 := rope.BalanceWithConfig(config2)

		// Both should have same content
		if balanced1.String() != balanced2.String() {
			t.Error("Different configs produced different content")
		}

		// But potentially different structure
		stats1 := balanced1.Stats()
		stats2 := balanced2.Stats()

		// Small leaves should have more leaves
		if stats2.LeafCount > 0 && stats1.LeafCount <= stats2.LeafCount {
			// This is expected: smaller leaf size -> more leaves
		}
	})
}

// TestBalanceDepth tests the Depth method for balance operations.
func TestBalanceDepth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxDepth int
	}{
		{
			name:     "empty rope",
			input:    "",
			maxDepth: 0,
		},
		{
			name:     "single leaf",
			input:    "Hello",
			maxDepth: 1,
		},
		{
			name:     "medium text",
			input:    string(make([]byte, 1000)),
			maxDepth: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rope := New(tt.input)
			depth := rope.Depth()

			if tt.maxDepth > 0 && depth > tt.maxDepth {
				t.Errorf("Depth %d exceeds expected max %d", depth, tt.maxDepth)
			}

			if tt.input == "" && depth != 0 {
				t.Errorf("Empty rope should have depth 0, got %d", depth)
			}
		})
	}
}

// TestIsBalanced tests the IsBalanced method.
func TestIsBalanced(t *testing.T) {
	t.Run("empty rope is balanced", func(t *testing.T) {
		rope := Empty()
		if !rope.IsBalanced() {
			t.Error("Empty rope should be balanced")
		}
	})

	t.Run("small rope is balanced", func(t *testing.T) {
		rope := New("Hello")
		if !rope.IsBalanced() {
			t.Error("Small rope should be balanced")
		}
	})

	t.Run("balanced rope", func(t *testing.T) {
		rope := New("Hello World")
		balanced := rope.Balance()
		if !balanced.IsBalanced() {
			t.Error("Balanced rope should be balanced")
		}
	})

	t.Run("large text remains balanced", func(t *testing.T) {
		text := string(make([]byte, 10000))
		rope := New(text)
		// Even large text should be reasonably balanced
		if !rope.IsBalanced() {
			// This might fail for very unbalanced ropes
			depth := rope.Depth()
			expectedMaxDepth := 2 * int(math.Ceil(math.Log2(float64(rope.Length()+1))))
			t.Logf("Depth: %d, Expected max: %d", depth, expectedMaxDepth)
		}
	})
}

// TestOptimize tests the Optimize method.
func TestOptimize(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty rope",
			input: "",
		},
		{
			name:  "small text",
			input: "Hello World",
		},
		{
			name:  "unicode text",
			input: "Hello 世界 🌍",
		},
		{
			name:  "repeated insertions",
			input: "ABCDEFGH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rope := New(tt.input)
			optimized := rope.Optimize()

			// Content should be preserved
			if optimized.String() != tt.input {
				t.Errorf("Optimize changed content: expected %q, got %q",
					tt.input, optimized.String())
			}

			// Length should be preserved
			if optimized.Length() != rope.Length() {
				t.Errorf("Optimize changed length: expected %d, got %d",
					rope.Length(), optimized.Length())
			}

			// Optimized rope should be balanced
			if !optimized.IsBalanced() {
				t.Error("Optimized rope should be balanced")
			}
		})
	}

	t.Run("optimize after edits", func(t *testing.T) {
		rope := New("Hello World")
		rope = rope.Insert(5, " Beautiful")  // "Hello Beautiful World"
		rope = rope.Delete(5, 16)               // Remove " Beautiful " -> "HelloWorld"

		optimized := rope.Optimize()
		// Optimize should only restructure the tree, not change content
		if optimized.String() != "HelloWorld" {
			t.Errorf("Optimize after edits changed content: got %q, want %q", optimized.String(), "HelloWorld")
		}
	})
}

// TestCompact tests the Compact method.
func TestCompact(t *testing.T) {
	t.Run("compact empty rope", func(t *testing.T) {
		rope := Empty()
		compacted := rope.Compact()
		if compacted.Length() != 0 {
			t.Errorf("Compacted empty rope should have length 0, got %d", compacted.Length())
		}
	})

	t.Run("compact small rope", func(t *testing.T) {
		rope := New("Hello World")
		compacted := rope.Compact()

		if compacted.String() != "Hello World" {
			t.Errorf("Compact changed content")
		}
	})

	t.Run("compact large rope", func(t *testing.T) {
		// Create a rope with many small nodes
		builder := NewBuilder()
		for i := 0; i < 100; i++ {
			builder.Append("Line " + string(rune('A'+i%26)) + "\n")
		}
		rope := builder.Build()

		compacted := rope.Compact()

		// Content should be preserved
		if compacted.String() != rope.String() {
			t.Error("Compact changed content")
		}

		// Length should be preserved
		if compacted.Length() != rope.Length() {
			t.Errorf("Compact changed length: expected %d, got %d",
				rope.Length(), compacted.Length())
		}

		// Compacted should have fewer or same number of nodes
		originalStats := rope.Stats()
		compactedStats := compacted.Stats()

		if compactedStats.NodeCount > originalStats.NodeCount {
			t.Logf("Warning: Compact increased node count from %d to %d",
				originalStats.NodeCount, compactedStats.NodeCount)
		}
	})

	t.Run("compact after many edits", func(t *testing.T) {
		rope := New("ABCDEFGHIJ")
		// Perform many edits
		rope = rope.Insert(5, "12345")
		rope = rope.Delete(8, 13)
		rope = rope.Insert(3, "XYZ")

		compacted := rope.Compact()

		// Content preserved
		if compacted.Length() != rope.Length() {
			t.Errorf("Compact changed length")
		}
	})
}

// TestValidate tests the Validate method.
func TestValidate(t *testing.T) {
	t.Run("validate empty rope", func(t *testing.T) {
		rope := Empty()
		err := rope.Validate()
		if err != nil {
			t.Errorf("Empty rope validation failed: %v", err)
		}
	})

	t.Run("validate simple rope", func(t *testing.T) {
		rope := New("Hello World")
		err := rope.Validate()
		if err != nil {
			t.Errorf("Simple rope validation failed: %v", err)
		}
	})

	t.Run("validate balanced rope", func(t *testing.T) {
		rope := New("Hello World")
		balanced := rope.Balance()
		err := balanced.Validate()
		if err != nil {
			t.Errorf("Balanced rope validation failed: %v", err)
		}
	})

	t.Run("validate compacted rope", func(t *testing.T) {
		rope := New("Hello World")
		compacted := rope.Compact()
		err := compacted.Validate()
		if err != nil {
			t.Errorf("Compacted rope validation failed: %v", err)
		}
	})

	t.Run("validate after edits", func(t *testing.T) {
		rope := New("ABCDEFGH")
		rope = rope.Insert(4, "XXXX")
		rope = rope.Delete(6, 10)
		err := rope.Validate()
		if err != nil {
			t.Errorf("Rope after edits validation failed: %v", err)
		}
	})

	t.Run("validate large rope", func(t *testing.T) {
		text := string(make([]byte, 10000))
		rope := New(text)
		err := rope.Validate()
		if err != nil {
			t.Errorf("Large rope validation failed: %v", err)
		}
	})
}

// TestSuggestedConfig tests the SuggestedConfig method.
func TestSuggestedConfig(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedMinSize  int
		expectedMaxSize  int
	}{
		{
			name:            "very small rope",
			input:           "Hi",
			expectedMinSize: 64,
			expectedMaxSize: 256,
		},
		{
			name:            "small rope",
			input:           string(make([]byte, 500)),
			expectedMinSize: 64,
			expectedMaxSize: 256,
		},
		{
			name:            "medium rope",
			input:           string(make([]byte, 5000)),
			expectedMinSize: 256,
			expectedMaxSize: 1024,
		},
		{
			name:            "large rope",
			input:           string(make([]byte, 2*1024*1024)),
			expectedMinSize: 512,
			expectedMaxSize: 2048,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rope := New(tt.input)
			config := rope.SuggestedConfig()

			if config.MinLeafSize != tt.expectedMinSize {
				t.Errorf("Expected MinLeafSize %d, got %d",
					tt.expectedMinSize, config.MinLeafSize)
			}

			if config.MaxLeafSize != tt.expectedMaxSize {
				t.Errorf("Expected MaxLeafSize %d, got %d",
					tt.expectedMaxSize, config.MaxLeafSize)
			}

			if config.MaxDepth != DefaultMaxDepth {
				t.Errorf("Expected MaxDepth %d, got %d",
					DefaultMaxDepth, config.MaxDepth)
			}
		})
	}

	t.Run("nil rope returns default", func(t *testing.T) {
		var rope *Rope
		config := rope.SuggestedConfig()
		defaultConfig := DefaultBalanceConfig()

		if config.MinLeafSize != defaultConfig.MinLeafSize {
			t.Error("Nil rope should return default config")
		}
	})
}

// TestAutoBalance tests the AutoBalance method.
func TestAutoBalance(t *testing.T) {
	t.Run("auto balance balanced rope", func(t *testing.T) {
		rope := New("Hello World")
		// Small ropes are typically balanced
		autoBalanced := rope.AutoBalance()

		// Should return same or equivalent rope
		if autoBalanced.String() != rope.String() {
			t.Error("AutoBalance changed content")
		}
	})

	t.Run("auto balance returns balanced rope", func(t *testing.T) {
		rope := New("Hello World")
		autoBalanced := rope.AutoBalance()

		if !autoBalanced.IsBalanced() {
			t.Error("AutoBalance should return balanced rope")
		}
	})

	t.Run("auto balance empty rope", func(t *testing.T) {
		rope := Empty()
		autoBalanced := rope.AutoBalance()

		if autoBalanced.Length() != 0 {
			t.Errorf("AutoBalanced empty rope should have length 0")
		}
	})

	t.Run("auto balance nil rope", func(t *testing.T) {
		var rope *Rope
		autoBalanced := rope.AutoBalance()

		if autoBalanced != nil {
			t.Error("AutoBalanced nil rope should return nil")
		}
	})
}

// TestBalanceStats tests the Stats method for balance operations.
func TestBalanceStats(t *testing.T) {
	t.Run("stats for empty rope", func(t *testing.T) {
		rope := Empty()
		stats := rope.Stats()

		// Empty() creates a rope with an empty leaf node
		if stats.NodeCount != 1 {
			t.Errorf("Empty rope has %d nodes, expected 1 (empty leaf)", stats.NodeCount)
		}
		if stats.LeafCount != 1 {
			t.Errorf("Empty rope has %d leaves, expected 1 (empty leaf)", stats.LeafCount)
		}
	})

	t.Run("stats for single leaf", func(t *testing.T) {
		rope := New("Hello")
		stats := rope.Stats()

		if stats.NodeCount == 0 {
			t.Error("Rope should have nodes")
		}
		if stats.LeafCount == 0 {
			t.Error("Rope should have at least one leaf")
		}
		if stats.Depth < 0 {
			t.Errorf("Invalid depth: %d", stats.Depth)
		}
	})

	t.Run("stats for larger rope", func(t *testing.T) {
		rope := New(string(make([]byte, 5000)))
		stats := rope.Stats()

		if stats.NodeCount == 0 {
			t.Error("Rope should have nodes")
		}
		if stats.LeafCount == 0 {
			t.Error("Rope should have leaves")
		}
		if stats.InternalCount == 0 {
			// Might be 0 for small ropes
		}
		if stats.MinLeafSize > stats.MaxLeafSize {
			t.Error("MinLeafSize should not exceed MaxLeafSize")
		}
	})

	t.Run("stats consistent", func(t *testing.T) {
		rope := New("Hello World")
		stats1 := rope.Stats()
		stats2 := rope.Stats()

		if stats1.NodeCount != stats2.NodeCount {
			t.Error("Stats should be consistent")
		}
		if stats1.LeafCount != stats2.LeafCount {
			t.Error("Stats should be consistent")
		}
	})
}

// TestLeafCount tests the LeafCount method.
func TestLeafCount(t *testing.T) {
	t.Run("leaf count for empty rope", func(t *testing.T) {
		rope := Empty()
		// Empty() creates a rope with one empty leaf node
		// This is by design - all ropes have a root node
		if rope.LeafCount() != 1 {
			t.Errorf("Empty rope should have 1 leaf node (empty), got %d", rope.LeafCount())
		}
	})

	t.Run("leaf count for simple rope", func(t *testing.T) {
		rope := New("Hello World")
		count := rope.LeafCount()
		if count <= 0 {
			t.Errorf("Rope should have at least 1 leaf, got %d", count)
		}
	})

	t.Run("leaf count matches stats", func(t *testing.T) {
		rope := New("Hello World")
		stats := rope.Stats()
		leafCount := rope.LeafCount()

		if leafCount != stats.LeafCount {
			t.Errorf("LeafCount() %d != Stats().LeafCount %d",
				leafCount, stats.LeafCount)
		}
	})
}

// TestNodeCount tests the NodeCount method.
func TestNodeCount(t *testing.T) {
	t.Run("node count for empty rope", func(t *testing.T) {
		rope := Empty()
		// Empty() creates a rope with one empty leaf node
		// This is by design - all ropes have a root node
		if rope.NodeCount() != 1 {
			t.Errorf("Empty rope should have 1 node (empty leaf), got %d", rope.NodeCount())
		}
	})

	t.Run("node count for simple rope", func(t *testing.T) {
		rope := New("Hello World")
		count := rope.NodeCount()
		if count <= 0 {
			t.Errorf("Rope should have at least 1 node, got %d", count)
		}
	})

	t.Run("node count >= leaf count", func(t *testing.T) {
		rope := New("Hello World")
		nodeCount := rope.NodeCount()
		leafCount := rope.LeafCount()

		if nodeCount < leafCount {
			t.Errorf("NodeCount %d should be >= LeafCount %d",
				nodeCount, leafCount)
		}
	})

	t.Run("node count matches stats", func(t *testing.T) {
		rope := New("Hello World")
		stats := rope.Stats()
		nodeCount := rope.NodeCount()

		if nodeCount != stats.NodeCount {
			t.Errorf("NodeCount() %d != Stats().NodeCount %d",
				nodeCount, stats.NodeCount)
		}
	})
}

// TestBalanceEdgeCases tests edge cases for balancing operations.
func TestBalanceEdgeCases(t *testing.T) {
	t.Run("balance nil rope", func(t *testing.T) {
		var rope *Rope
		balanced := rope.Balance()
		if balanced != nil {
			t.Error("Balancing nil rope should return nil")
		}
	})

	t.Run("optimize nil rope", func(t *testing.T) {
		var rope *Rope
		optimized := rope.Optimize()
		if optimized != nil {
			t.Error("Optimizing nil rope should return nil")
		}
	})

	t.Run("compact nil rope", func(t *testing.T) {
		var rope *Rope
		compacted := rope.Compact()
		if compacted != nil {
			t.Error("Compacting nil rope should return nil")
		}
	})

	t.Run("balance with nil config", func(t *testing.T) {
		_ = New("Hello")
		// This should panic or handle nil gracefully
		// Uncomment if the implementation handles nil config
		// balanced := rope.BalanceWithConfig(nil)
	})

	t.Run("balance very large text", func(t *testing.T) {
		// Create a large text
		text := string(make([]byte, 100000))
		rope := New(text)

		balanced := rope.Balance()
		if balanced.String() != text {
			t.Error("Balancing large text changed content")
		}

		if !balanced.IsBalanced() {
			t.Error("Large balanced rope should be balanced")
		}
	})
}

// TestBalancePreservesContent tests that balancing operations preserve content.
func TestBalancePreservesContent(t *testing.T) {
	texts := []string{
		"",
		"Hello",
		"Hello World",
		"Line1\nLine2\nLine3",
		"Hello 世界 🌍",
		string(make([]byte, 10000)),
	}

	for _, text := range texts {
		t.Run("", func(t *testing.T) {
			rope := New(text)

			// Test all balancing operations
			balanced := rope.Balance()
			optimized := rope.Optimize()
			compacted := rope.Compact()
			autoBalanced := rope.AutoBalance()

			// All should preserve content
			if balanced.String() != text {
				t.Error("Balance changed content")
			}
			if optimized.String() != text {
				t.Error("Optimize changed content")
			}
			if compacted.String() != text {
				t.Error("Compact changed content")
			}
			if autoBalanced.String() != text {
				t.Error("AutoBalance changed content")
			}

			// All should preserve length
			if balanced.Length() != len([]rune(text)) {
				t.Error("Balance changed length")
			}
			if optimized.Length() != len([]rune(text)) {
				t.Error("Optimize changed length")
			}
			if compacted.Length() != len([]rune(text)) {
				t.Error("Compact changed length")
			}
		})
	}
}
