package rope

import "testing"

func TestBytesIteratorAt_Zero(t *testing.T) {
	r := New("Hello")

	// This should not panic
	it := r.IterBytesAt(0)

	if it == nil {
		t.Fatal("Iterator should not be nil")
	}

	// First Next() should move to byte 0
	if !it.Next() {
		t.Fatal("Next() should return true")
	}

	if it.Position() != 0 {
		t.Errorf("Position should be 0, got %d", it.Position())
	}

	if it.Current() != 'H' {
		t.Errorf("Current should be 'H', got %c", it.Current())
	}
}
