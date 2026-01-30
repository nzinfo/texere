package rope

// SplitOff splits the rope at the given character position, returning
// a new rope containing the text after the split point, and a new rope
// containing the text before the split point.
//
// This is the inverse operation of Append(). The original rope is unchanged.
//
// Example:
//   r := rope.New("Hello World")
//   left, right := r.SplitOff(5)
//   fmt.Println(left.String())   // Output: "Hello"
//   fmt.Println(right.String())  // Output: " World"
func (r *Rope) SplitOff(pos int) (*Rope, *Rope) {
	if pos <= 0 {
		// Split at beginning: return empty rope for left, full rope for right
		return Empty(), r.Clone()
	}
	if pos >= r.Length() {
		// Split at end: return full rope for left, empty rope for right
		return r.Clone(), Empty()
	}

	// Split the rope using the existing Split method
	left, right := r.Split(pos)

	return left, right
}

