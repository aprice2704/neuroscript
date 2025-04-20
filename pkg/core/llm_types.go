// filename: pkg/core/llm_types.go
package core

import (
	"github.com/google/generative-ai-go/genai" // Ensure genai is imported if not already
)

// Existing types (if any) remain here...

// LLMRequestContext encapsulates the necessary context for an LLM call,
// including conversation history and any specific file references.
type LLMRequestContext struct {
	History  []*genai.Content // Existing conversation history
	FileURIs []string         // List of File API URIs to include in this specific call
	// Add other context fields here if needed in the future
}

// LLMResponse encapsulates the response from the LLM, including content and potential errors.
// Assuming a structure like this exists or is needed for CallLLMAgent return.
// If CallLLMAgent returns *genai.GenerateContentResponse directly, this might not be needed.
type LLMResponse struct {
	Response *genai.GenerateContentResponse
	Error    error
}
