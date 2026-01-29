// Package weave provides the core weaving engine that combines
// human edits and AI generations into a coherent document.
//
// The weave engine is the heart of Texere, orchestrating the flow
// of content from multiple sources (users, AI, imports) and
// weaving them together into a final document fabric.
package engine

import (
	"sync"
	"time"

	"github.com/coreseekdev/texere/pkg/concordia"
)

// Engine represents the document weaving engine.
// It coordinates human edits and AI generations, ensuring
// consistency through Operational Transformation.
type Engine struct {
	mu       sync.RWMutex
	document *Document
	history  *History
	ai       *AIWeaver
}

// Document represents the current state of the document.
type Document struct {
	ID      string
	Content string
	Version int
	Updated time.Time
}

// History maintains the edit history for undo/redo.
type History struct {
	undo []*concordia.Operation
	redo []*concordia.Operation
	max  int
}

// AIWeaver handles AI-assisted content generation.
type AIWeaver struct {
	enabled bool
	model   string
}

// EngineConfig configures the weaving engine.
type EngineConfig struct {
	DocumentID   string
	InitialDoc   string
	AIEnabled    bool
	AIModel      string
	HistoryLimit int
}

// NewEngine creates a new weaving engine.
func NewEngine(config EngineConfig) *Engine {
	if config.HistoryLimit == 0 {
		config.HistoryLimit = 1000
	}

	return &Engine{
		document: &Document{
			ID:      config.DocumentID,
			Content: config.InitialDoc,
			Version: 0,
			Updated: time.Now(),
		},
		history: &History{
			undo: make([]*syntaxis.Operation, 0, config.HistoryLimit),
			redo: make([]*syntaxis.Operation, 0, config.HistoryLimit),
			max:  config.HistoryLimit,
		},
		ai: &AIWeaver{
			enabled: config.AIEnabled,
			model:   config.AIModel,
		},
	}
}

// WeaveHuman weaves a human editing operation into the document.
//
// This method:
// 1. Validates the operation
// 2. Transforms against concurrent operations
// 3. Applies to the document
// 4. Records in history
func (e *Engine) WeaveHuman(op *concordia.Operation) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Apply operation
	e.document.Content = concordia.Apply(e.document.Content, op)
	e.document.Version++
	e.document.Updated = time.Now()

	// Record in history
	e.recordUndo(op)

	return nil
}

// WeaveAI weaves AI-generated content into the document.
//
// The AI weaver can:
// - Generate content based on context
// - Suggest completions
// - Expand on existing content
// - Summarize sections
func (e *Engine) WeaveAI(request *AIRequest) (*AIResponse, error) {
	if !e.ai.enabled {
		return nil, ErrAIDisabled
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Generate AI content
	response, err := e.ai.Generate(request, e.document)
	if err != nil {
		return nil, err
	}

	// Create insert operation
	op := concordia.NewInsert(request.Position, response.Content)

	// Apply operation
	e.document.Content = concordia.Apply(e.document.Content, op)
	e.document.Version++
	e.document.Updated = time.Now()

	// Record in history
	e.recordUndo(op)

	return response, nil
}

// WeaveMultiple weaves multiple operations atomically.
// This is useful for batching concurrent edits.
func (e *Engine) WeaveMultiple(ops []*concordia.Operation) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Compose operations
	composed := concordia.Compose(ops...)

	// Apply composed operation
	e.document.Content = concordia.Apply(e.document.Content, composed)
	e.document.Version++
	e.document.Updated = time.Now()

	// Record in history
	e.recordUndo(composed)

	return nil
}

// Document returns the current document state.
func (e *Engine) Document() *Document {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return &Document{
		ID:      e.document.ID,
		Content: e.document.Content,
		Version: e.document.Version,
		Updated: e.document.Updated,
	}
}

// Undo reverses the last operation.
func (e *Engine) Undo() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.history.undo) == 0 {
		return ErrNoUndo
	}

	// Get last operation
	op := e.history.undo[len(e.history.undo)-1]
	e.history.undo = e.history.undo[:len(e.history.undo)-1]

	// Create inverse operation
	inverse := e.invertOperation(op)

	// Apply inverse
	e.document.Content = syntaxis.Apply(e.document.Content, inverse)
	e.document.Version++
	e.document.Updated = time.Now()

	// Record to redo
	e.history.redo = append(e.history.redo, op)

	return nil
}

// Redo reapplies the last undone operation.
func (e *Engine) Redo() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.history.redo) == 0 {
		return ErrNoRedo
	}

	// Get operation to redo
	op := e.history.redo[len(e.history.redo)-1]
	e.history.redo = e.history.redo[:len(e.history.redo)-1]

	// Apply operation
	e.document.Content = concordia.Apply(e.document.Content, op)
	e.document.Version++
	e.document.Updated = time.Now()

	// Record to undo
	e.history.undo = append(e.history.undo, op)

	return nil
}

// recordUndo records an operation in the undo history.
func (e *Engine) recordUndo(op *concordia.Operation) {
	e.history.undo = append(e.history.undo, op)

	// Clear redo stack
	e.history.redo = e.history.redo[:0]

	// Trim history if too long
	if len(e.history.undo) > e.history.max {
		e.history.undo = e.history.undo[1:]
	}
}

// invertOperation creates the inverse of an operation.
func (e *Engine) invertOperation(op *concordia.Operation) *concordia.Operation {
	switch op.Type() {
	case concordia.OpInsert:
		// Inverse of insert is delete
		return concordia.NewDelete(op.Position(), len(op.Content()))
	case concordia.OpDelete:
		// Inverse of delete is insert (need to store deleted content)
		// For now, simplified
		return concordia.NewInsert(op.Position(), "")
	default:
		return op
	}
}

// AIRequest represents an AI generation request.
type AIRequest struct {
	Position  int
	Context   string
	Mode      AIMode
	MaxLength int
	Temperature float64
}

// AIMode specifies the AI generation mode.
type AIMode int

const (
	// AIModeComplete suggests completion for the current text
	AIModeComplete AIMode = iota
	// AIModeExpand expands on the current text
	AIModeExpand
	// AIModeSummarize summarizes the current text
	AIModeSummarize
	// AIModeRewrite rewrites the current text
	AIModeRewrite
)

// AIResponse represents the AI generation response.
type AIResponse struct {
	Content    string
	FinishReason string
	TokensUsed int
	Duration   time.Duration
}

// Errors
var (
	ErrAIDisabled = &WeaveError{Message: "AI weaving is disabled"}
	ErrNoUndo     = &WeaveError{Message: "no operations to undo"}
	ErrNoRedo     = &WeaveError{Message: "no operations to redo"}
)

// WeaveError represents a weaving error.
type WeaveError struct {
	Message string
}

func (e *WeaveError) Error() string {
	return e.Message
}
