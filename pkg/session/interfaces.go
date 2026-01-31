package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ========== Authentication Interface ==========

// UserInfo represents user information from authentication.
type UserInfo struct {
	UserID       string
	Username     string
	Name         string
	Email        string
	AuthProvider string // e.g., "token", "oauth"
	Permissions  []string
	Metadata     map[string]interface{}
}

// TokenInfo represents information about a token.
type TokenInfo struct {
	Token     string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
	Metadata  map[string]interface{}
}

// Authenticator provides token-based authentication.
// Similar to Jupyter Notebook's token mechanism:
// - Tokens are used as identity credentials
// - Default implementation doesn't validate, only distinguishes sessions
// - Tokens can be stored in cookies or query parameters
type Authenticator interface {
	// Authenticate validates a token and returns user info.
	Authenticate(ctx context.Context, token string) (*UserInfo, error)

	// GenerateToken generates a new token for a user.
	GenerateToken(ctx context.Context, userID string) (string, error)

	// ValidateToken checks if a token is valid and returns user info.
	ValidateToken(ctx context.Context, token string) (bool, *UserInfo)

	// RevokeToken revokes a token.
	RevokeToken(ctx context.Context, token string) error

	// RefreshToken refreshes an existing token.
	RefreshToken(ctx context.Context, token string) (string, error)
}

// TokenAuthenticator provides a default token-based authentication implementation.
// Similar to Jupyter: tokens don't expire by default, used for session identification.
type TokenAuthenticator struct {
	mu     sync.RWMutex
	tokens map[string]*TokenInfo // token -> token info
	users  map[string]*UserInfo   // userID -> user info
}

// NewTokenAuthenticator creates a new token authenticator.
func NewTokenAuthenticator() *TokenAuthenticator {
	return &TokenAuthenticator{
		tokens: make(map[string]*TokenInfo),
		users:  make(map[string]*UserInfo),
	}
}

// Authenticate validates a token and returns user info.
func (a *TokenAuthenticator) Authenticate(ctx context.Context, token string) (*UserInfo, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	tokenInfo, ok := a.tokens[token]
	if !ok {
		return nil, ErrInvalidToken
	}

	// Check if expired
	if !tokenInfo.ExpiresAt.IsZero() && time.Now().After(tokenInfo.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	user, ok := a.users[tokenInfo.UserID]
	if !ok {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GenerateToken generates a new token for a user.
// Uses a simple random token generation (similar to Jupyter).
func (a *TokenAuthenticator) GenerateToken(ctx context.Context, userID string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if user exists
	user, ok := a.users[userID]
	if !ok {
		// Create default user
		user = &UserInfo{
			UserID:   userID,
			Metadata: make(map[string]interface{}),
		}
		a.users[userID] = user
	}

	// Generate random token (16 bytes = 32 hex chars)
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(bytes)

	tokenInfo := &TokenInfo{
		Token:     token,
		UserID:    userID,
		CreatedAt: time.Now(),
		// No expiration by default (like Jupyter)
		Metadata:  make(map[string]interface{}),
	}

	a.tokens[token] = tokenInfo

	return token, nil
}

// ValidateToken checks if a token is valid and returns user info.
func (a *TokenAuthenticator) ValidateToken(ctx context.Context, token string) (bool, *UserInfo) {
	user, err := a.Authenticate(ctx, token)
	if err != nil {
		return false, nil
	}
	return true, user
}

// RevokeToken revokes a token.
func (a *TokenAuthenticator) RevokeToken(ctx context.Context, token string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.tokens, token)
	return nil
}

// RefreshToken refreshes an existing token (generates a new token).
func (a *TokenAuthenticator) RefreshToken(ctx context.Context, token string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	tokenInfo, ok := a.tokens[token]
	if !ok {
		return "", ErrInvalidToken
	}

	// Revoke old token
	delete(a.tokens, token)

	// Generate new token for the same user
	return a.GenerateToken(ctx, tokenInfo.UserID)
}

// AddUser adds or updates a user.
func (a *TokenAuthenticator) AddUser(user *UserInfo) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.users[user.UserID] = user
}

// GetUser retrieves a user by ID.
func (a *TokenAuthenticator) GetUser(userID string) (*UserInfo, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	user, ok := a.users[userID]
	if !ok {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// ListUsers returns all users.
func (a *TokenAuthenticator) ListUsers() []*UserInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	users := make([]*UserInfo, 0, len(a.users))
	for _, user := range a.users {
		users = append(users, user)
	}
	return users
}

// ListTokens returns all tokens.
func (a *TokenAuthenticator) ListTokens() []*TokenInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	tokens := make([]*TokenInfo, 0, len(a.tokens))
	for _, info := range a.tokens {
		tokens = append(tokens, info)
	}
	return tokens
}

// ========== Content Storage Interface ==========

// ContentModel represents a document/content.
type ContentModel struct {
	Name        string
	Type        string // "file", "directory", "notebook"
	Content     string
	Format      string // "json", "text", etc.
	MimeType    string
	Size        int64
	Created     string
	Modified    string
	Path        string
	ReadOnly    bool
	Metadata    map[string]interface{}
}

// GetOptions specifies options for getting content.
type GetOptions struct {
	Format      string
	ContentType string
}

// SaveOptions specifies options for saving content.
type SaveOptions struct {
	Overwrite     bool
	CreateParents bool
}

// ContentItem represents an item in a directory listing.
type ContentItem struct {
	Name     string
	Path     string
	Type     string // "file", "directory"
	MimeType string
	Size     int64
	Modified string
}

// ContentStorage provides K/V content storage like Jupyter's contents API.
// Reference: https://jupyter-notebook.readthedocs.io/en/stable/contents.html
type ContentStorage interface {
	// List returns a list of content items at the given path.
	List(ctx context.Context, path string) ([]*ContentItem, error)

	// Get retrieves content at the given path.
	Get(ctx context.Context, contentPath string, options *GetOptions) (*ContentModel, error)

	// Save saves content at the given path.
	Save(ctx context.Context, contentPath string, model *ContentModel, options *SaveOptions) (*ContentModel, error)

	// Delete deletes content at the given path.
	Delete(ctx context.Context, contentPath string) error

	// CheckExists checks if content exists at the given path.
	CheckExists(ctx context.Context, contentPath string) (bool, error)

	// CreateDirectory creates a new directory.
	CreateDirectory(ctx context.Context, dirPath string) error

	// Rename renames/moves content.
	Rename(ctx context.Context, oldPath, newPath string) error

	// GetSize returns the size of content at the given path.
	GetSize(ctx context.Context, contentPath string) (int64, error)
}

// MemoryContentStorage provides an in-memory content storage implementation.
type MemoryContentStorage struct {
	mu       sync.RWMutex
	contents map[string]*ContentModel // path -> content
}

// NewMemoryContentStorage creates a new memory content storage.
func NewMemoryContentStorage() *MemoryContentStorage {
	return &MemoryContentStorage{
		contents: make(map[string]*ContentModel),
	}
}

// List returns a list of content items at the given path.
func (m *MemoryContentStorage) List(ctx context.Context, path string) ([]*ContentItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]*ContentItem, 0)

	// Add trailing slash for proper prefix matching
	prefix := path
	if path != "" && !strings.HasSuffix(path, "/") {
		prefix += "/"
	}

	for p, content := range m.contents {
		// Exact match or prefix match (for subdirectories)
		if p == path || (len(p) > len(path) && strings.HasPrefix(p, prefix)) {
			items = append(items, &ContentItem{
				Name:     content.Name,
				Path:     content.Path,
				Type:     content.Type,
				MimeType: content.MimeType,
				Size:     content.Size,
				Modified: content.Modified,
			})
		}
	}

	return items, nil
}

// Get retrieves content at the given path.
func (m *MemoryContentStorage) Get(ctx context.Context, contentPath string, options *GetOptions) (*ContentModel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	content, ok := m.contents[contentPath]
	if !ok {
		return nil, ErrContentNotFound
	}

	return content, nil
}

// Save saves content at the given path.
func (m *MemoryContentStorage) Save(ctx context.Context, contentPath string, model *ContentModel, options *SaveOptions) (*ContentModel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Set metadata
	model.Path = contentPath
	model.Modified = time.Now().Format(time.RFC3339)

	if model.Created == "" {
		model.Created = model.Modified
	}

	m.contents[contentPath] = model
	return model, nil
}

// Delete deletes content at the given path.
func (m *MemoryContentStorage) Delete(ctx context.Context, contentPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.contents, contentPath)
	return nil
}

// CheckExists checks if content exists at the given path.
func (m *MemoryContentStorage) CheckExists(ctx context.Context, contentPath string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.contents[contentPath]
	return ok, nil
}

// CreateDirectory creates a new directory.
func (m *MemoryContentStorage) CreateDirectory(ctx context.Context, dirPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.contents[dirPath] = &ContentModel{
		Name:     dirPath[strings.LastIndex(dirPath, "/")+1:],
		Path:     dirPath,
		Type:     "directory",
		Modified: time.Now().Format(time.RFC3339),
		Created:  time.Now().Format(time.RFC3339),
		Metadata: make(map[string]interface{}),
	}

	return nil
}

// Rename renames/moves content.
func (m *MemoryContentStorage) Rename(ctx context.Context, oldPath, newPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	content, ok := m.contents[oldPath]
	if !ok {
		return ErrContentNotFound
	}

	delete(m.contents, oldPath)
	content.Path = newPath
	m.contents[newPath] = content

	return nil
}

// GetSize returns the size of content at the given path.
func (m *MemoryContentStorage) GetSize(ctx context.Context, contentPath string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	content, ok := m.contents[contentPath]
	if !ok {
		return 0, ErrContentNotFound
	}

	return content.Size, nil
}

// ========== Error Definitions ==========

var (
	// ErrInvalidToken is returned when a token is invalid.
	ErrInvalidToken = &SessionError{Code: "invalid_token", Message: "invalid token"}

	// ErrTokenExpired is returned when a token has expired.
	ErrTokenExpired = &SessionError{Code: "token_expired", Message: "token expired"}

	// ErrContentNotFound is returned when content is not found.
	ErrContentNotFound = &SessionError{Code: "not_found", Message: "content not found"}

	// ErrSessionNotFound is returned when a session is not found.
	ErrSessionNotFound = &SessionError{Code: "session_not_found", Message: "session not found"}

	// ErrUserNotFound is returned when a user is not found.
	ErrUserNotFound = &SessionError{Code: "user_not_found", Message: "user not found"}

	// ErrAlreadyExists is returned when content already exists.
	ErrAlreadyExists = &SessionError{Code: "already_exists", Message: "already exists"}

	// ErrPermissionDenied is returned when user lacks permission.
	ErrPermissionDenied = &SessionError{Code: "permission_denied", Message: "permission denied"}

	// ErrUnauthorized is returned when authentication fails.
	ErrUnauthorized = &SessionError{Code: "unauthorized", Message: "unauthorized"}

	// ErrInvalidRequest is returned when the request is invalid.
	ErrInvalidRequest = &SessionError{Code: "bad_request", Message: "invalid request"}

	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = &SessionError{Code: "not_found", Message: "not found"}

	// ErrConflict is returned when there's a conflict.
	ErrConflict = &SessionError{Code: "conflict", Message: "conflict"}

	// ErrServerError is returned for internal server errors.
	ErrServerError = &SessionError{Code: "internal_error", Message: "internal server error"}

	// ErrServiceUnavailable is returned when the service is unavailable.
	ErrServiceUnavailable = &SessionError{Code: "service_unavailable", Message: "service unavailable"}

	// ErrForbidden is returned when access is forbidden.
	ErrForbidden = &SessionError{Code: "forbidden", Message: "forbidden"}

	// ErrTooManyRequests is returned when rate limit is exceeded.
	ErrTooManyRequests = &SessionError{Code: "too_many_requests", Message: "too many requests"}

	// ErrContentTooLarge is returned when content is too large.
	ErrContentTooLarge = &SessionError{Code: "content_too_large", Message: "content too large"}

	// ErrNo such file or directory
	ErrNoSuchFileOrDirectory = &SessionError{Code: "no such file or directory", Message: "no such file or directory"}
)

// SessionError represents a session-related error.
type SessionError struct {
	Code    string
	Message string
}

func (e *SessionError) Error() string {
	return e.Message
}

// IsRetryable returns true if the error is retryable.
func (e *SessionError) IsRetryable() bool {
	switch e.Code {
	case "service_unavailable", "too_many_requests":
		return true
	default:
		return false
	}
}

// StatusCode returns the HTTP status code for this error.
func (e *SessionError) StatusCode() int {
	switch e.Code {
	case "invalid_token", "unauthorized":
		return 401
	case "forbidden", "permission_denied":
		return 403
	case "not_found":
		return 404
	case "conflict":
		return 409
	case "already_exists":
		return 409
	case "content_too_large":
		return 413
	case "too_many_requests":
		return 429
	case "internal_error":
		return 500
	case "not_implemented":
		return 501
	case "service_unavailable":
		return 503
	default:
		return 500
	}
}
