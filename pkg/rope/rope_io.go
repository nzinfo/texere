package rope

import (
	"bufio"
	"io"
)

// FromReader reads content from an io.Reader and creates a new Rope.
//
// This is useful for efficiently loading large files without loading
// everything into memory at once.
//
// Example:
//   file, _ := os.Open("large_file.txt")
//   defer file.Close()
//   rope, err := rope.FromReader(file)
//
//   // Or with buffered reading for better performance:
//   file, _ := os.Open("large_file.txt")
//   defer file.Close()
//   rope, err := rope.FromReader(bufio.NewReader(file))
func FromReader(reader io.Reader) (*Rope, error) {
	b := NewBuilder()
	bufReader := bufio.NewReader(reader)
	buf := make([]byte, 4096)

	for {
		n, err := bufReader.Read(buf)
		if n > 0 {
			b.Append(string(buf[:n]))
		}
		if err != nil {
			if err == io.EOF {
				return b.Build(), nil
			}
			// Clean up on error
			return nil, err
		}
	}
}

// WriteTo writes the rope's content to an io.Writer.
//
// Returns the number of bytes written and any error encountered.
//
// Example:
//   r := rope.New("Hello World")
//   var buf bytes.Buffer
//   n, err := r.WriteTo(&buf)
func (r *Rope) WriteTo(writer io.Writer) (int, error) {
	// Convert to string and write
	// This is efficient for most use cases
	str := r.String()
	return writer.Write([]byte(str))
}

// WriteToBuffer writes the rope's content to a bytes.Buffer.
//
// This is a convenience method for writing to a buffer.
//
// Example:
//   r := rope.New("Hello World")
//   var buf bytes.Buffer
//   r.WriteToBuffer(&buf)
func (r *Rope) WriteToBuffer(buf interface{ Write([]byte) (int, error) }) (int, error) {
	return r.WriteTo(buf)
}

// Reader returns a new io.Reader that reads from the rope.
//
// This allows using a Rope anywhere an io.Reader is expected.
//
// Example:
//   r := rope.New("Hello World")
//   reader := r.Reader()
//   data, _ := io.ReadAll(reader)
func (r *Rope) Reader() io.Reader {
	return &ropeReader{rope: r, pos: 0}
}

// ropeReader implements io.Reader for Rope
type ropeReader struct {
	rope *Rope
	pos  int
}

func (rr *ropeReader) Read(p []byte) (int, error) {
	if rr.pos >= rr.rope.Size() {
		return 0, io.EOF
	}

	// Read available bytes up to len(p)
	available := rr.rope.Size() - rr.pos
	toRead := len(p)
	if toRead > available {
		toRead = available
	}

	// Get bytes from rope
	bytes := rr.rope.IterBytes()
	bytes.Seek(rr.pos)

	count := 0
	for count < toRead && bytes.Next() {
		b := bytes.Current()
		p[count] = b
		count++
	}

	rr.pos += count
	if count < toRead {
		return count, io.EOF
	}
	return count, nil
}
