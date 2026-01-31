package rope

// EditOperation represents a single edit operation with (from, to, replacement).
type EditOperation struct {
	From int
	To   int
	Text string // Empty string for deletion
}

// Deletion represents a deletion range (from, to).
type Deletion struct {
	From int
	To   int
}
