package transport

import (
	"context"
)

// ========== History Service Interface ==========

// SnapshotInfo contains information about a snapshot.
type SnapshotInfo struct {
	SnapshotVersion          int64 `json:"snapshot_version"`
	LastSnapshotTime         int64 `json:"last_snapshot_time"`
	RecentChangeCount        int   `json:"recent_change_count"`
	MaxChangesBeforeSnapshot int   `json:"max_changes_before_snapshot"`
	MaxSnapshotInterval      int64 `json:"max_snapshot_interval"` // seconds
	TimeUntilSnapshot        int64 `json:"time_until_snapshot"`   // seconds
}

// HistoryService provides version history storage and retrieval.
// Similar to Jupyter's checkpoint mechanism and HedgeDoc's revision history.
//
// Implementations can use different storage backends:
// - Redis: For distributed systems
// - Memory: For testing and single-instance deployments
// - Database: For persistent storage (PostgreSQL, MySQL, etc.)
//
// The interface supports two storage modes:
// 1. Full content: Stores complete content for each version (simple, more storage)
// 2. Patch mode: Stores only diffs using diff-match-patch (efficient, less storage)
type HistoryService interface {
	// OnSnapshot handles snapshot events from edit sessions.
	// Called when a new snapshot is created (either by operation count or timeout).
	OnSnapshot(event *HistoryEvent) error

	// OnOperation handles operation events from edit sessions.
	// Called for each OT operation applied to a document.
	OnOperation(event *HistoryEvent) error

	// GetSnapshot retrieves a specific snapshot from storage.
	// Returns the HistoryEvent with content and metadata for the given version.
	GetSnapshot(ctx context.Context, sessionID string, versionID int64) (*HistoryEvent, error)

	// GetSessionHistory retrieves history for a session from storage.
	// Returns up to `limit` most recent events (snapshots and operations).
	GetSessionHistory(ctx context.Context, sessionID string, limit int64) ([]*HistoryEvent, error)

	// ReconstructSnapshot reconstructs the content of a specific version.
	// Necessary when using patch mode, where only the first snapshot has full content.
	// Starting from version 0, applies all patches up to the target version.
	ReconstructSnapshot(ctx context.Context, sessionID string, targetVersionID int64) (string, error)

	// ListSnapshots lists all snapshots for a session.
	// Returns metadata for each snapshot (version, time, creator).
	ListSnapshots(ctx context.Context, sessionID string) ([]*SnapshotInfo, error)

	// Close closes the history service and releases resources.
	Close() error
}

// HistoryOptions specifies options for creating a history service.
type HistoryOptions struct {
	// UsePatchMode enables diff-match-patch for efficient storage.
	// When true, stores only patches between versions (like HedgeDoc).
	// When false, stores full content for each version (simpler, more storage).
	UsePatchMode bool

	// MaxChangesBeforeSnapshot triggers snapshot creation after N operations.
	// Default: 200 operations.
	MaxChangesBeforeSnapshot int

	// MaxSnapshotInterval triggers snapshot creation after N seconds of inactivity.
	// Default: 300 seconds (5 minutes).
	MaxSnapshotInterval int64

	// StorageBackend specifies which storage backend to use.
	// Options: "redis", "memory", "database"
	StorageBackend string

	// RedisAddr specifies Redis server address (if using Redis backend).
	RedisAddr string

	// RedisPassword specifies Redis password (if required).
	RedisPassword string

	// RedisDB specifies Redis database number.
	RedisDB int
}

// NewHistoryService creates a new history service based on options.
// Factory function that returns the appropriate implementation.
func NewHistoryService(opts *HistoryOptions) HistoryService {
	if opts == nil {
		opts = &HistoryOptions{}
	}

	// Set defaults
	if opts.MaxChangesBeforeSnapshot == 0 {
		opts.MaxChangesBeforeSnapshot = 200
	}
	if opts.MaxSnapshotInterval == 0 {
		opts.MaxSnapshotInterval = 300
	}

	// Create appropriate implementation based on backend
	switch opts.StorageBackend {
	case "redis", "":
		// Redis backend (default)
		var redisClient RedisClient
		if opts.RedisAddr != "" {
			// TODO: Create real Redis client
			// For now, fall back to MiniRedis
			redisClient = NewMiniRedis()
		} else {
			redisClient = NewMiniRedis()
		}
		return NewRedisHistoryService(redisClient)

	case "memory":
		// In-memory implementation
		return NewMemoryHistoryService(opts.UsePatchMode)

	default:
		// Default to MiniRedis
		return NewRedisHistoryService(NewMiniRedis())
	}
}
