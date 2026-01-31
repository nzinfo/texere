package rope

import (
	"strings"
	"testing"
)

// TestNewBuilder tests creating a new builder.
func TestNewBuilder(t *testing.T) {
	t.Run("empty builder", func(t *testing.T) {
		builder := NewBuilder()
		if builder == nil {
			t.Fatal("NewBuilder() returned nil")
		}
		if builder.Length() != 0 {
			t.Errorf("Expected length 0, got %d", builder.Length())
		}
		if builder.Size() != 0 {
			t.Errorf("Expected size 0, got %d", builder.Size())
		}
	})
}

// TestBuilderAppend tests the Append method.
func TestBuilderAppend(t *testing.T) {
	tests := []struct {
		name     string
		appends  []string
		expected string
	}{
		{
			name:     "single append",
			appends:  []string{"Hello"},
			expected: "Hello",
		},
		{
			name:     "multiple appends",
			appends:  []string{"Hello", " ", "World"},
			expected: "Hello World",
		},
		{
			name:     "append empty string",
			appends:  []string{"Hello", "", "World"},
			expected: "HelloWorld",
		},
		{
			name:     "append unicode",
			appends:  []string{"Hello", " 世界", " 🌍"},
			expected: "Hello 世界 🌍",
		},
		{
			name:     "append with newlines",
			appends:  []string{"Line1\n", "Line2\n", "Line3"},
			expected: "Line1\nLine2\nLine3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder()
			for _, s := range tt.appends {
				builder.Append(s)
			}
			rope := builder.Build()
			if rope.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, rope.String())
			}
		})
	}
}

// TestBuilderAppendBytes tests the AppendBytes method.
func TestBuilderAppendBytes(t *testing.T) {
	t.Run("append bytes", func(t *testing.T) {
		builder := NewBuilder()
		data := []byte("Hello World")
		builder.AppendBytes(data)
		rope := builder.Build()
		if rope.String() != "Hello World" {
			t.Errorf("Expected 'Hello World', got %q", rope.String())
		}
	})

	t.Run("append empty bytes", func(t *testing.T) {
		builder := NewBuilder()
		builder.AppendBytes([]byte{})
		rope := builder.Build()
		if rope.Length() != 0 {
			t.Errorf("Expected length 0, got %d", rope.Length())
		}
	})

	t.Run("append multiple byte slices", func(t *testing.T) {
		builder := NewBuilder()
		builder.AppendBytes([]byte("Hello"))
		builder.AppendBytes([]byte(" "))
		builder.AppendBytes([]byte("World"))
		rope := builder.Build()
		if rope.String() != "Hello World" {
			t.Errorf("Expected 'Hello World', got %q", rope.String())
		}
	})
}

// TestBuilderInsert tests the Insert method.
func TestBuilderInsert(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		inserts  []struct {
			pos int
			text string
		}
		expected string
	}{
		{
			name:    "insert at beginning",
			initial: "World",
			inserts: []struct {
				pos int
				text string
			}{{0, "Hello "}},
			expected: "Hello World",
		},
		{
			name:    "insert at end",
			initial: "Hello",
			inserts: []struct {
				pos int
				text string
			}{{5, " World"}},
			expected: "Hello World",
		},
		{
			name:    "insert in middle",
			initial: "Hi World",
			inserts: []struct {
				pos int
				text string
			}{{3, "Hello "}},
			expected: "Hi Hello World",
		},
		{
			name:    "multiple inserts",
			initial: "AC",
			inserts: []struct {
				pos int
				text string
			}{{1, "B"}, {3, "D"}},
			expected: "ABCD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilderFromRope(New(tt.initial))
			for _, ins := range tt.inserts {
				builder.Insert(ins.pos, ins.text)
			}
			rope := builder.Build()
			if rope.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, rope.String())
			}
		})
	}
}

// TestBuilderDelete tests the Delete method.
func TestBuilderDelete(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		start    int
		end      int
		expected string
	}{
		{
			name:     "delete from beginning",
			initial:  "Hello World",
			start:    0,
			end:      6,
			expected: "World",
		},
		{
			name:     "delete from end",
			initial:  "Hello World",
			start:    5,
			end:      11,
			expected: "Hello",
		},
		{
			name:     "delete from middle",
			initial:  "Hello World",
			start:    5,
			end:      6,
			expected: "HelloWorld",
		},
		{
			name:     "delete all",
			initial:  "Hello",
			start:    0,
			end:      5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilderFromRope(New(tt.initial))
			builder.Delete(tt.start, tt.end)
			rope := builder.Build()
			if rope.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, rope.String())
			}
		})
	}
}

// TestBuilderReplace tests the Replace method.
func TestBuilderReplace(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		start    int
		end      int
		text     string
		expected string
	}{
		{
			name:     "replace at beginning",
			initial:  "Hello World",
			start:    0,
			end:      5,
			text:     "Goodbye",
			expected: "Goodbye World",
		},
		{
			name:     "replace at end",
			initial:  "Hello World",
			start:    6,
			end:      11,
			text:     "Universe",
			expected: "Hello Universe",
		},
		{
			name:     "replace in middle",
			initial:  "Hello World",
			start:    5,
			end:      6,
			text:     " Beautiful ",
			expected: "Hello Beautiful World",
		},
		{
			name:     "replace with empty string (delete)",
			initial:  "Hello World",
			start:    5,
			end:      11,
			text:     "",
			expected: "Hello",
		},
		{
			name:     "replace all",
			initial:  "ABC",
			start:    0,
			end:      3,
			text:     "XYZ",
			expected: "XYZ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilderFromRope(New(tt.initial))
			builder.Replace(tt.start, tt.end, tt.text)
			rope := builder.Build()
			if rope.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, rope.String())
			}
		})
	}
}

// TestBuilderReset tests the Reset method.
func TestBuilderReset(t *testing.T) {
	t.Run("reset after operations", func(t *testing.T) {
		builder := NewBuilder()
		builder.Append("Hello")
		builder.Append(" World")
		builder.Reset()

		if builder.Length() != 0 {
			t.Errorf("Expected length 0 after reset, got %d", builder.Length())
		}
		if builder.Size() != 0 {
			t.Errorf("Expected size 0 after reset, got %d", builder.Size())
		}

		builder.Append("New")
		rope := builder.Build()
		if rope.String() != "New" {
			t.Errorf("Expected 'New', got %q", rope.String())
		}
	})
}

// TestBuilderResetFromRope tests the ResetFromRope method.
func TestBuilderResetFromRope(t *testing.T) {
	t.Run("reset from existing rope", func(t *testing.T) {
		builder := NewBuilder()
		builder.Append("Hello")

		newRope := New("World")
		builder.ResetFromRope(newRope)

		rope := builder.Build()
		if rope.String() != "World" {
			t.Errorf("Expected 'World', got %q", rope.String())
		}
	})

	t.Run("reset and continue building", func(t *testing.T) {
		builder := NewBuilder()
		builder.Append("Hello")

		newRope := New("World")
		builder.ResetFromRope(newRope)
		builder.Append("!")

		rope := builder.Build()
		if rope.String() != "World!" {
			t.Errorf("Expected 'World!', got %q", rope.String())
		}
	})
}

// TestBuilderInsertString tests the InsertString method.
func TestBuilderInsertString(t *testing.T) {
	t.Run("insert string at position", func(t *testing.T) {
		builder := NewBuilderFromRope(New("AC"))
		builder.InsertString(1, "B")
		rope := builder.Build()
		if rope.String() != "ABC" {
			t.Errorf("Expected 'ABC', got %q", rope.String())
		}
	})

	t.Run("method chaining", func(t *testing.T) {
		builder := NewBuilderFromRope(New("ACE"))
		builder.InsertString(1, "B").InsertString(3, "D")
		rope := builder.Build()
		if rope.String() != "ABCDE" {
			t.Errorf("Expected 'ABCDE', got %q", rope.String())
		}
	})
}

// TestBuilderInsertRune tests the InsertRune method.
func TestBuilderInsertRune(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		position int
		rune     rune
		expected string
	}{
		{
			name:     "insert ASCII rune",
			initial:  "AC",
			position: 1,
			rune:     'B',
			expected: "ABC",
		},
		{
			name:     "insert unicode rune",
			initial:  "Hi",
			position: 2,
			rune:     '🌍',
			expected: "Hi🌍",
		},
		{
			name:     "insert at beginning",
			initial:  "BC",
			position: 0,
			rune:     'A',
			expected: "ABC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilderFromRope(New(tt.initial))
			builder.InsertRune(tt.position, tt.rune)
			rope := builder.Build()
			if rope.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, rope.String())
			}
		})
	}
}

// TestBuilderInsertByte tests the InsertByte method.
func TestBuilderInsertByte(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		position int
		byteVal  byte
		expected string
	}{
		{
			name:     "insert ASCII byte",
			initial:  "AC",
			position: 1,
			byteVal:  'B',
			expected: "ABC",
		},
		{
			name:     "insert space",
			initial:  "HelloWorld",
			position: 5,
			byteVal:  ' ',
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilderFromRope(New(tt.initial))
			builder.InsertByte(tt.position, tt.byteVal)
			rope := builder.Build()
			if rope.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, rope.String())
			}
		})
	}
}

// TestBuilderAppendRune tests the AppendRune method.
func TestBuilderAppendRune(t *testing.T) {
	t.Run("append single rune", func(t *testing.T) {
		builder := NewBuilder()
		builder.AppendRune('H')
		builder.AppendRune('i')
		rope := builder.Build()
		if rope.String() != "Hi" {
			t.Errorf("Expected 'Hi', got %q", rope.String())
		}
	})

	t.Run("append unicode rune", func(t *testing.T) {
		builder := NewBuilder()
		builder.Append("Hello ")
		builder.AppendRune('🌍')
		rope := builder.Build()
		if rope.String() != "Hello 🌍" {
			t.Errorf("Expected 'Hello 🌍', got %q", rope.String())
		}
	})

	t.Run("append multiple runes", func(t *testing.T) {
		builder := NewBuilder()
		for _, r := range "ABC" {
			builder.AppendRune(r)
		}
		rope := builder.Build()
		if rope.String() != "ABC" {
			t.Errorf("Expected 'ABC', got %q", rope.String())
		}
	})
}

// TestBuilderAppendByte tests the AppendByte method.
func TestBuilderAppendByte(t *testing.T) {
	t.Run("append single byte", func(t *testing.T) {
		builder := NewBuilder()
		builder.AppendByte('H')
		builder.AppendByte('i')
		rope := builder.Build()
		if rope.String() != "Hi" {
			t.Errorf("Expected 'Hi', got %q", rope.String())
		}
	})

	t.Run("append bytes", func(t *testing.T) {
		builder := NewBuilder()
		builder.Append("Hello")
		builder.AppendByte(' ')
		builder.AppendByte('W')
		rope := builder.Build()
		if rope.String() != "Hello W" {
			t.Errorf("Expected 'Hello W', got %q", rope.String())
		}
	})
}

// TestBuilderAppendLine tests the AppendLine method.
func TestBuilderAppendLine(t *testing.T) {
	t.Run("append lines", func(t *testing.T) {
		builder := NewBuilder()
		builder.AppendLine("Line1")
		builder.AppendLine("Line2")
		builder.Append("Line3")
		rope := builder.Build()
		if rope.String() != "Line1\nLine2\nLine3" {
			t.Errorf("Expected 'Line1\\nLine2\\nLine3', got %q", rope.String())
		}
	})
}

// TestBuilderLength tests the Length method.
func TestBuilderLength(t *testing.T) {
	t.Run("length with pending operations", func(t *testing.T) {
		builder := NewBuilder()
		if builder.Length() != 0 {
			t.Errorf("Expected initial length 0, got %d", builder.Length())
		}

		builder.Append("Hello")
		if builder.Length() != 5 {
			t.Errorf("Expected length 5, got %d", builder.Length())
		}

		builder.Append(" World")
		if builder.Length() != 11 {
			t.Errorf("Expected length 11, got %d", builder.Length())
		}
	})

	t.Run("length with unicode", func(t *testing.T) {
		builder := NewBuilder()
		builder.Append("Hi")
		builder.AppendRune('🌍')
		if builder.Length() != 3 {
			t.Errorf("Expected length 3, got %d", builder.Length())
		}
	})
}

// TestBuilderSize tests the Size method.
func TestBuilderSize(t *testing.T) {
	t.Run("size with pending operations", func(t *testing.T) {
		builder := NewBuilder()
		if builder.Size() != 0 {
			t.Errorf("Expected initial size 0, got %d", builder.Size())
		}

		builder.Append("Hello")
		if builder.Size() != 5 {
			t.Errorf("Expected size 5, got %d", builder.Size())
		}

		builder.Append(" World")
		if builder.Size() != 11 {
			t.Errorf("Expected size 11, got %d", builder.Size())
		}
	})

	t.Run("size with unicode", func(t *testing.T) {
		builder := NewBuilder()
		builder.Append("Hi")
		builder.AppendRune('🌍')
		// '🌍' is 4 bytes in UTF-8
		if builder.Size() != 6 {
			t.Errorf("Expected size 6, got %d", builder.Size())
		}
	})
}

// TestBuilderPool tests the BuilderPool.
func TestBuilderPool(t *testing.T) {
	t.Run("create pool", func(t *testing.T) {
		pool := NewBuilderPool(5)
		if pool == nil {
			t.Fatal("NewBuilderPool() returned nil")
		}
	})

	t.Run("get and put builder", func(t *testing.T) {
		pool := NewBuilderPool(2)

		// Get a builder
		builder1 := pool.Get()
		if builder1 == nil {
			t.Fatal("pool.Get() returned nil")
		}

		// Use it
		builder1.Append("Hello")
		rope1 := builder1.Build()
		if rope1.String() != "Hello" {
			t.Errorf("Expected 'Hello', got %q", rope1.String())
		}

		// Return to pool
		pool.Put(builder1)

		// Get again (should get the same reset builder)
		builder2 := pool.Get()
		if builder2.Length() != 0 {
			t.Errorf("Expected reset builder with length 0, got %d", builder2.Length())
		}

		builder2.Append("World")
		rope2 := builder2.Build()
		if rope2.String() != "World" {
			t.Errorf("Expected 'World', got %q", rope2.String())
		}
	})

	t.Run("pool overflow", func(t *testing.T) {
		pool := NewBuilderPool(1)

		builder1 := pool.Get()
		builder2 := pool.Get()
		builder3 := pool.Get()

		// Put all back
		pool.Put(builder1)
		pool.Put(builder2)
		pool.Put(builder3)

		// Pool should only hold 1
		if builder1.Length() != 0 {
			t.Error("Builder was not reset when put in pool")
		}
	})

	t.Run("concurrent usage", func(t *testing.T) {
		pool := NewBuilderPool(10)
		done := make(chan bool)

		// Concurrent gets and puts
		for i := 0; i < 10; i++ {
			go func() {
				builder := pool.Get()
				builder.Append("test")
				builder.Build()
				pool.Put(builder)
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// TestBuilderWrite tests the Write method (io.Writer interface).
func TestBuilderWrite(t *testing.T) {
	t.Run("write bytes", func(t *testing.T) {
		builder := NewBuilder()
		data := []byte("Hello World")

		n, err := builder.Write(data)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
		}

		rope := builder.Build()
		if rope.String() != "Hello World" {
			t.Errorf("Expected 'Hello World', got %q", rope.String())
		}
	})
}

// TestBuilderWriteString tests the WriteString method (io.StringWriter interface).
func TestBuilderWriteString(t *testing.T) {
	t.Run("write string", func(t *testing.T) {
		builder := NewBuilder()
		s := "Hello World"

		n, err := builder.WriteString(s)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if n != len(s) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(s), n)
		}

		rope := builder.Build()
		if rope.String() != "Hello World" {
			t.Errorf("Expected 'Hello World', got %q", rope.String())
		}
	})
}

// TestBuilderBuildReuse tests that Build allows reuse of the builder.
func TestBuilderBuildReuse(t *testing.T) {
	t.Run("build multiple times", func(t *testing.T) {
		builder := NewBuilder()

		// First build
		builder.Append("Hello")
		rope1 := builder.Build()
		if rope1.String() != "Hello" {
			t.Errorf("Expected 'Hello', got %q", rope1.String())
		}

		// Second build (should add to existing rope)
		builder.Append(" World")
		rope2 := builder.Build()
		if rope2.String() != "Hello World" {
			t.Errorf("Expected 'Hello World', got %q", rope2.String())
		}
	})
}

// TestBuilderComplexOperations tests complex combinations of operations.
func TestBuilderComplexOperations(t *testing.T) {
	t.Run("mixed operations", func(t *testing.T) {
		builder := NewBuilder()
		builder.Append("Hello")
		builder.Append(" Beautiful")
		builder.Append(" World")
		rope := builder.Build()
		// After appends: "Hello Beautiful World"
		expected := "Hello Beautiful World"
		if rope.String() != expected {
			t.Errorf("Expected %q, got %q", expected, rope.String())
		}
	})

	t.Run("large text building", func(t *testing.T) {
		builder := NewBuilder()
		var expected strings.Builder

		// Build a large text
		for i := 0; i < 1000; i++ {
			line := strings.Repeat("Line ", i%10)
			builder.AppendLine(line)
			expected.WriteString(line)
			expected.WriteString("\n")
		}

		rope := builder.Build()
		if rope.String() != expected.String() {
			t.Errorf("Large text mismatch")
		}
	})
}

// TestBuilderEdgeCases tests edge cases and boundary conditions.
func TestBuilderEdgeCases(t *testing.T) {
	t.Run("empty operations", func(t *testing.T) {
		builder := NewBuilder()
		builder.Append("")
		builder.Insert(0, "")
		builder.Delete(0, 0)
		rope := builder.Build()
		if rope.Length() != 0 {
			t.Errorf("Expected empty rope, got length %d", rope.Length())
		}
	})

	t.Run("insert at bounds", func(t *testing.T) {
		builder := NewBuilderFromRope(New("AB"))
		builder.Insert(0, "X")  // At beginning
		builder.Insert(2, "Y")  // In middle
		builder.Insert(4, "Z")  // At end
		rope := builder.Build()
		if rope.String() != "XAYBZ" {
			t.Errorf("Expected 'XAYBZ', got %q", rope.String())
		}
	})

	t.Run("replace entire content", func(t *testing.T) {
		builder := NewBuilderFromRope(New("Old"))
		builder.Replace(0, 3, "New")
		rope := builder.Build()
		if rope.String() != "New" {
			t.Errorf("Expected 'New', got %q", rope.String())
		}
	})

	t.Run("append after build", func(t *testing.T) {
		builder := NewBuilder()
		builder.Append("Hello")
		_ = builder.Build()
		builder.Append(" World")
		rope := builder.Build()
		if rope.String() != "Hello World" {
			t.Errorf("Expected 'Hello World', got %q", rope.String())
		}
	})
}
