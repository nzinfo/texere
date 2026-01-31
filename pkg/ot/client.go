package ot

// ClientState represents the state of an OT client.
type ClientState int

const (
	// StateSynchronized means the client is in sync with the server.
	StateSynchronized ClientState = iota
	// StateAwaitingConfirm means the client has sent an operation and is waiting for acknowledgment.
	StateAwaitingConfirm
	// StateAwaitingWithBuffer means the client has sent an operation and has buffered another.
	StateAwaitingWithBuffer
)

// Client represents an OT client for collaborative editing.
//
// The client manages the local state, handles operation transformation,
// and ensures consistency with the server. This is based on the ot.js
// Client implementation.
//
// Example usage:
//
//	client := NewClient()
//
//	// Apply a local operation
//	op := NewBuilder().Insert("Hello").Build()
//	client.ApplyClient(op)
//
//	// Send to server
//	sendOp := client.OutgoingOperation()
type Client struct {
	state     ClientState
	revision  int
	document  string // Current document state
	clientOp  *Operation
	bufferOp  *Operation
	serverOps []*Operation // Server operations since last sync
}

// NewClient creates a new OT client.
//
// Returns:
//   - a new Client in Synchronized state
func NewClient() *Client {
	return &Client{
		state:     StateSynchronized,
		revision:  0,
		document:  "",
		clientOp:  nil,
		bufferOp:  nil,
		serverOps: make([]*Operation, 0),
	}
}

// State returns the current client state.
func (c *Client) State() ClientState {
	return c.state
}

// Revision returns the current revision number.
func (c *Client) Revision() int {
	return c.revision
}

// Document returns the current document state.
func (c *Client) Document() string {
	return c.document
}

// ApplyClient applies a client-side operation.
//
// This transforms the operation against any buffered operations and
// updates the client state accordingly.
//
// Parameters:
//   - op: the operation to apply
//
// Returns:
//   - the new document state
//   - an error if the operation cannot be applied
func (c *Client) ApplyClient(op *Operation) (string, error) {
	newDoc, err := op.Apply(c.document)
	if err != nil {
		return "", err
	}

	switch c.state {
	case StateSynchronized:
		c.state = StateAwaitingConfirm
		c.clientOp = op
	case StateAwaitingConfirm:
		c.state = StateAwaitingWithBuffer
		c.bufferOp = op
	case StateAwaitingWithBuffer:
		// Compose with buffer
		composed, err := Compose(c.bufferOp, op)
		if err != nil {
			return "", err
		}
		c.bufferOp = composed
	}

	c.document = newDoc
	return c.document, nil
}

// ApplyServer applies a server-side operation.
//
// This transforms the server operation against any pending client
// operations and applies it to the concordia.
//
// Parameters:
//   - revision: the revision number of the server operation
//   - op: the operation to apply
//
// Returns:
//   - the new document state
//   - an error if the operation cannot be applied
func (c *Client) ApplyServer(revision int, op *Operation) (string, error) {
	// Validate revision
	if revision != c.revision {
		return "", ErrInvalidBaseLength
	}

	var transformedOp *Operation
	var err error

	switch c.state {
	case StateSynchronized:
		// Just apply the operation
		transformedOp = op
	case StateAwaitingConfirm:
		// Transform against client operation
		c.clientOp, transformedOp, err = Transform(c.clientOp, op)
		if err != nil {
			return "", err
		}
	case StateAwaitingWithBuffer:
		// Transform against both client and buffer
		c.clientOp, transformedOp, err = Transform(c.clientOp, op)
		if err != nil {
			return "", err
		}
		c.bufferOp, _, err = Transform(c.bufferOp, op)
		if err != nil {
			return "", err
		}
	}

	// Apply the transformed operation
	newDoc, err := transformedOp.Apply(c.document)
	if err != nil {
		return "", err
	}

	c.document = newDoc
	c.revision++
	return c.document, nil
}

// ServerAck handles a server acknowledgment.
//
// This is called when the server acknowledges receipt of a client operation.
//
// Returns:
//   - an error if the client state is invalid
func (c *Client) ServerAck() error {
	if c.state != StateAwaitingConfirm && c.state != StateAwaitingWithBuffer {
		return ErrInvalidBaseLength
	}

	c.revision++

	switch c.state {
	case StateAwaitingConfirm:
		c.state = StateSynchronized
		c.clientOp = nil
	case StateAwaitingWithBuffer:
		c.state = StateAwaitingConfirm
		c.clientOp = c.bufferOp
		c.bufferOp = nil
	}

	return nil
}

// OutgoingOperation returns the operation to send to the server.
//
// Returns:
//   - the operation to send, or nil if there's no pending operation
func (c *Client) OutgoingOperation() *Operation {
	switch c.state {
	case StateAwaitingConfirm:
		return c.clientOp
	case StateAwaitingWithBuffer:
		return c.clientOp
	default:
		return nil
	}
}
