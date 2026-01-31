package transport

import (
	"context"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
)

// TCPTransport implements Transport over TCP.
type TCPTransport struct {
	*BaseTransport
	conn   net.Conn
	enc    *gob.Encoder
	dec    *gob.Decoder
	mu     sync.Mutex
	closed bool
}

// NewTCPTransport creates a new TCP transport.
func NewTCPTransport(id, clientID, docID string) *TCPTransport {
	base := NewBaseTransport(id, clientID, docID)
	return &TCPTransport{
		BaseTransport: base,
	}
}

// Connect establishes a TCP connection.
func (t *TCPTransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	// For client mode, connect to a server
	// For server mode, this would be called with an accepted connection
	t.connected = true

	// Start receive loop
	go t.receiveLoop()

	return nil
}

// SetConnection sets the underlying TCP connection (for server mode).
func (t *TCPTransport) SetConnection(conn net.Conn) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.conn = conn
	t.enc = gob.NewEncoder(conn)
	t.dec = gob.NewDecoder(conn)
	t.connected = true

	// Start receive loop
	go t.receiveLoop()
}

// Send sends a message over TCP.
func (t *TCPTransport) Send(ctx context.Context, msg *Message) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed || t.conn == nil {
		return ErrTransportClosed
	}

	err := t.enc.Encode(msg)
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	return nil
}

// Close closes the TCP connection.
func (t *TCPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	t.connected = false

	if t.conn != nil {
		t.conn.Close()
	}

	return t.BaseTransport.Close()
}

// receiveLoop receives messages from TCP connection.
func (t *TCPTransport) receiveLoop() {
	for {
		t.mu.Lock()
		if t.closed || t.conn == nil {
			t.mu.Unlock()
			return
		}
		t.mu.Unlock()

		var msg Message
		err := t.dec.Decode(&msg)
		if err != nil {
			t.Close()
			return
		}

		select {
		case t.recvCh <- &msg:
		case <-t.closeCh:
			return
		}
	}
}

// ========== TCP Server ==========

// TCPServer handles incoming TCP connections.
type TCPServer struct {
	addr     string
	mu       sync.RWMutex
	conns    map[string]net.Conn
	acceptCh chan net.Conn
	closeCh  chan struct{}
}

// NewTCPServer creates a new TCP server.
func NewTCPServer(addr string) *TCPServer {
	return &TCPServer{
		addr:     addr,
		conns:    make(map[string]net.Conn),
		acceptCh: make(chan net.Conn, 10),
		closeCh:  make(chan struct{}),
	}
}

// Start starts accepting connections.
func (s *TCPServer) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-s.closeCh:
					return
				default:
					continue
				}
			}

			select {
			case s.acceptCh <- conn:
			case <-s.closeCh:
				conn.Close()
				return
			}
		}
	}()

	return nil
}

// Accept accepts the next connection.
func (s *TCPServer) Accept() (net.Conn, error) {
	select {
	case conn := <-s.acceptCh:
		s.mu.Lock()
		s.conns[conn.RemoteAddr().String()] = conn
		s.mu.Unlock()
		return conn, nil
	case <-s.closeCh:
		return nil, fmt.Errorf("server closed")
	}
}

// Close closes the server.
func (s *TCPServer) Close() error {
	select {
	case <-s.closeCh:
		return nil
	default:
		close(s.closeCh)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, conn := range s.conns {
		conn.Close()
	}

	s.conns = make(map[string]net.Conn)
	return nil
}

// ========== TCP Client ==========

// DialTCP creates a TCP transport by connecting to an address.
func DialTCP(addr string) (*TCPTransport, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	transport := NewTCPTransport(
		conn.RemoteAddr().String(),
		"client",
		"",
	)
	transport.SetConnection(conn)

	return transport, nil
}
