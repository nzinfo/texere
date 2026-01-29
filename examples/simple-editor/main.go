// Package main provides a simple example of using Texere
// for collaborative editing with AI assistance.
package main

import (
	"fmt"
	"log"

	"github.com/coreseekdev/texere/pkg/concordia"
	"github.com/coreseekdev/texere/pkg/weave/engine"
)

func main() {
	fmt.Println("🧵 Texere - Document Weaving Engine")
	fmt.Println("===================================\n")

	// Example 1: Basic OT Operations
	fmt.Println("Example 1: Basic OT Operations")
	basicOTExample()

	fmt.Println("\n---\n")

	// Example 2: Weaving Human Edits
	fmt.Println("Example 2: Weaving Human Edits")
	humanWeaveExample()

	fmt.Println("\n---\n")

	// Example 3: AI-Assisted Editing
	fmt.Println("Example 3: AI-Assisted Editing")
	aiWeaveExample()
}

// basicOTExample demonstrates basic OT operations.
func basicOTExample() {
	// Create initial document
	doc := "Hello"

	fmt.Printf("Initial document: \"%s\"\n", doc)

	// User A inserts at position 5
	op1 := concordia.NewInsert(5, " World")
	doc = concordia.Apply(doc, op1)
	fmt.Printf("After insert at 5: \"%s\"\n", doc)

	// User B inserts at position 0 (concurrent)
	op2 := concordia.NewInsert(0, "Hi ")
	doc = concordia.Apply(doc, op2)
	fmt.Printf("After insert at 0: \"%s\"\n", doc)

	// Delete operation
	op3 := concordia.NewDelete(0, 3)
	doc = concordia.Apply(doc, op3)
	fmt.Printf("After delete at 0: \"%s\"\n", doc)
}

// humanWeaveExample demonstrates weaving human edits.
func humanWeaveExample() {
	// Create weaving engine
	config := engine.EngineConfig{
		DocumentID:   "doc-001",
		InitialDoc:   "Hello",
		AIEnabled:    false,
		HistoryLimit: 100,
	}
	e := engine.NewEngine(config)

	fmt.Printf("Initial document: \"%s\"\n", e.Document().Content)

	// User inserts text
	op1 := concordia.NewInsert(5, " World")
	if err := e.WeaveHuman(op1); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("After user edit: \"%s\"\n", e.Document().Content)

	// Another user edits concurrently
	op2 := concordia.NewInsert(0, "Greetings: ")
	if err := e.WeaveHuman(op2); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("After concurrent edit: \"%s\"\n", e.Document().Content)

	// Undo
	if err := e.Undo(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("After undo: \"%s\"\n", e.Document().Content)

	// Redo
	if err := e.Redo(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("After redo: \"%s\"\n", e.Document().Content)
}

// aiWeaveExample demonstrates AI-assisted editing.
func aiWeaveExample() {
	// Create weaving engine with AI enabled
	config := engine.EngineConfig{
		DocumentID:   "doc-002",
		InitialDoc:   "The quick brown fox",
		AIEnabled:    true,
		AIModel:      "gpt-4",
		HistoryLimit: 100,
	}
	e := engine.NewEngine(config)

	fmt.Printf("Initial document: \"%s\"\n", e.Document().Content)

	// User starts typing
	op1 := concordia.NewInsert(19, " jumps over")
	if err := e.WeaveHuman(op1); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("After user edit: \"%s\"\n", e.Document().Content)

	// AI suggests completion
	request := &engine.AIRequest{
		Position:  28,
		Context:   e.Document().Content,
		Mode:      engine.AIModeComplete,
		MaxLength: 100,
	}

	response, err := e.WeaveAI(request)
	if err != nil {
		fmt.Printf("AI generation error: %v\n", err)
		// This is expected since we don't have a real AI backend
		return
	}

	fmt.Printf("AI generated: \"%s\"\n", response.Content)
	fmt.Printf("Final document: \"%s\"\n", e.Document().Content)
	fmt.Printf("Tokens used: %d, Duration: %v\n", response.TokensUsed, response.Duration)
}
