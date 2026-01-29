package engine

import (
	"context"
	"fmt"
	"time"
)

// Generate generates AI content based on the request.
func (aw *AIWeaver) Generate(req *AIRequest, doc *Document) (*AIResponse, error) {
	// This is a placeholder implementation
	// Real implementation would call an LLM API (OpenAI, Claude, etc.)

	startTime := time.Now()

	// Extract context around the position
	context := aw.extractContext(doc.Content, req.Position)

	// Generate content based on mode
	content, err := aw.generateContent(req.Mode, context, req.MaxLength)
	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)

	return &AIResponse{
		Content:      content,
		FinishReason: "stop",
		TokensUsed:   len(content) / 4, // Rough estimate
		Duration:     duration,
	}, nil
}

// extractContext extracts text context around a position.
func (aw *AIWeaver) extractContext(content string, position int) string {
	// Extract 500 characters before and after the position
	start := position - 500
	if start < 0 {
		start = 0
	}

	end := position + 500
	if end > len(content) {
		end = len(content)
	}

	return content[start:end]
}

// generateContent generates content based on the mode.
func (aw *AIWeaver) generateContent(mode AIMode, context string, maxLength int) (string, error) {
	// Placeholder implementation
	// Real implementation would use an LLM API

	switch mode {
	case AIModeComplete:
		return aw.generateCompletion(context, maxLength)
	case AIModeExpand:
		return aw.generateExpansion(context, maxLength)
	case AIModeSummarize:
		return aw.generateSummary(context)
	case AIModeRewrite:
		return aw.generateRewrite(context)
	default:
		return "", fmt.Errorf("unknown AI mode: %d", mode)
	}
}

// generateCompletion suggests completion for the current text.
func (aw *AIWeaver) generateCompletion(context string, maxLength int) (string, error) {
	// Placeholder: return a simple completion
	return "completion", nil
}

// generateExpansion expands on the current text.
func (aw *AIWeaver) generateExpansion(context string, maxLength int) (string, error) {
	// Placeholder: return an expansion
	return "expanded content", nil
}

// generateSummary summarizes the current text.
func (aw *AIWeaver) generateSummary(context string) (string, error) {
	// Placeholder: return a summary
	return "summary", nil
}

// generateRewrite rewrites the current text.
func (aw *AIWeaver) generateRewrite(context string) (string, error) {
	// Placeholder: return a rewritten version
	return "rewritten content", nil
}

// StreamGenerate streams AI content generation.
func (aw *AIWeaver) StreamGenerate(
	ctx context.Context,
	req *AIRequest,
	doc *Document,
) (<-chan string, <-chan error) {
	contentCh := make(chan string)
	errCh := make(chan error, 1)

	go func() {
		defer close(contentCh)
		defer close(errCh)

		response, err := aw.Generate(req, doc)
		if err != nil {
			errCh <- err
			return
		}

		// Stream the content word by word
		words := splitWords(response.Content)
		for _, word := range words {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case contentCh <- word:
			}
		}
	}()

	return contentCh, errCh
}

// splitWords splits text into words for streaming.
func splitWords(text string) []string {
	// Simple word splitting
	// Real implementation would be more sophisticated
	words := make([]string, 0)
	current := ""

	for _, r := range text {
		if r == ' ' || r == '\n' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
			words = append(words, string(r))
		} else {
			current += string(r)
		}
	}

	if current != "" {
		words = append(words, current)
	}

	return words
}
