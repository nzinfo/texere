package session

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages multiple editing sessions.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]Session
	auth     Authenticator
	storage  ContentStorage
}

// NewManager creates a new session manager.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]Session),
		auth:     NewTokenAuthenticator(),
		storage:   NewMemoryContentStorage(),
	}
}

// SetAuth sets the authenticator.
func (m *Manager) SetAuth(auth Authenticator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.auth = auth
}

// SetStorage sets the content storage.
func (m *Manager) SetStorage(storage ContentStorage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.storage = storage
}

// CreateSession creates a new session.
func (m *Manager) CreateSession(ctx context.Context, config SessionConfig) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[config.DocID]; exists {
		return nil, fmt.Errorf("session already exists: %s", config.DocID)
	}

	var session Session
	var err error

	switch config.DocID {
	default:
		// Create SimpleSession by default
		session, err = NewSimpleSession(ctx, config)
		if err != nil {
			return nil, err
		}
	}

	m.sessions[config.DocID] = session
	return session, nil
}

// GetSession retrieves a session by ID.
func (m *Manager) GetSession(docID string) (Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[docID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

// DeleteSession deletes a session.
func (m *Manager) DeleteSession(docID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[docID]
	if !ok {
		return ErrSessionNotFound
	}

	// Close the session
	if err := session.Close(); err != nil {
		return err
	}

	delete(m.sessions, docID)
	return nil
}

// ListSessions returns all session IDs.
func (m *Manager) ListSessions() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	return ids
}

// Authenticate authenticates a token and returns user info.
func (m *Manager) Authenticate(ctx context.Context, token string) (*UserInfo, error) {
	return m.auth.Authenticate(ctx, token)
}

// GenerateToken generates a new token for a user.
func (m *Manager) GenerateToken(ctx context.Context, userID string) (string, error) {
	return m.auth.GenerateToken(ctx, userID)
}

// ValidateToken validates a token.
func (m *Manager) ValidateToken(ctx context.Context, token string) (bool, *UserInfo) {
	return m.auth.ValidateToken(ctx, token)
}

// GetContent retrieves content from storage.
func (m *Manager) GetContent(ctx context.Context, path string, options *GetOptions) (*ContentModel, error) {
	return m.storage.Get(ctx, path, options)
}

// SaveContent saves content to storage.
func (m *Manager) SaveContent(ctx context.Context, path string, model *ContentModel, options *SaveOptions) (*ContentModel, error) {
	return m.storage.Save(ctx, path, model, options)
}

// DeleteContent deletes content from storage.
func (m *Manager) DeleteContent(ctx context.Context, path string) error {
	return m.storage.Delete(ctx, path)
}

// ListContent lists content at a path.
func (m *Manager) ListContent(ctx context.Context, path string) ([]*ContentItem, error) {
	return m.storage.List(ctx, path)
}
