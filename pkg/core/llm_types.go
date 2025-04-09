// filename: pkg/core/llm_types.go
package core

// --- Gemini API Request/Response Structures (including Function Calling) ---
// Based on documentation: https://ai.google.dev/api/rest/v1beta/models/generateContent

// GeminiContent represents a single piece of content in the conversation history.
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"` // "user" or "model" or "function"
}

// GeminiPart represents a unit within content (text or function call/response).
type GeminiPart struct {
	Text             string                  `json:"text,omitempty"`
	FunctionCall     *GeminiFunctionCall     `json:"functionCall,omitempty"`     // LLM -> Agent request
	FunctionResponse *GeminiFunctionResponse `json:"functionResponse,omitempty"` // Agent -> LLM result
	// Add other part types like inlineData (FileData) if needed later
	// FileData         *GeminiFileData         `json:"fileData,omitempty"`
}

// GeminiFileData represents file data sent to the API (if needed).
// type GeminiFileData struct {
//  MimeType string `json:"mimeType"`
//  FileURI  string `json:"fileUri"`
// }

// GeminiFunctionCall represents the LLM's request to call a tool.
type GeminiFunctionCall struct {
	Name string                 `json:"name"` // Tool name (e.g., "TOOL.ReadFile")
	Args map[string]interface{} `json:"args"` // Arguments as a map
}

// GeminiFunctionResponse represents the result of a tool execution sent back to the LLM.
type GeminiFunctionResponse struct {
	Name     string                 `json:"name"`     // Tool name (must match the call)
	Response map[string]interface{} `json:"response"` // Result map (e.g., {"content": "...", "success": true})
}

// GeminiRequest is the top-level structure for sending requests to the Gemini API.
type GeminiRequest struct {
	Contents         []GeminiContent         `json:"contents"`                   // Conversation history
	Tools            []GeminiTool            `json:"tools,omitempty"`            // Tool declarations
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"` // Optional generation settings
	SafetySettings   []GeminiSafetySetting   `json:"safetySettings,omitempty"`   // Optional safety settings
	// SystemInstruction *GeminiContent `json:"systemInstruction,omitempty"` // If needed later
	// tool_config, etc. if needed
}

// GeminiTool wraps the function declarations.
type GeminiTool struct {
	FunctionDeclarations []GeminiFunctionDeclaration `json:"functionDeclarations"`
	// Add CodeExecution or other tool types if needed later
}

// GeminiFunctionDeclaration describes a tool (function) to the LLM.
type GeminiFunctionDeclaration struct {
	Name        string                 `json:"name"`                 // Tool name (e.g., "TOOL.ReadFile")
	Description string                 `json:"description"`          // Description for the LLM
	Parameters  *GeminiParameterSchema `json:"parameters,omitempty"` // OpenAPI schema for parameters
}

// GeminiParameterSchema defines the structure (parameters) of a tool.
// Based on OpenAPI Schema Object: https://swagger.io/specification/v3/#schema-object
type GeminiParameterSchema struct {
	Type       string                            `json:"type"` // Usually "object" for functions
	Properties map[string]GeminiParameterDetails `json:"properties"`
	Required   []string                          `json:"required,omitempty"`
	// Add other schema properties if needed: description, nullable, etc.
}

// GeminiParameterDetails describes a single tool parameter.
type GeminiParameterDetails struct {
	Type        string   `json:"type"` // e.g., "string", "integer", "number", "boolean", "array", "object"
	Description string   `json:"description,omitempty"`
	Format      string   `json:"format,omitempty"` // e.g., "int64", "float", "double" (for type: "integer" or "number")
	Nullable    bool     `json:"nullable,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	// For type: "array"
	Items *GeminiParameterDetails `json:"items,omitempty"` // Schema for array items
	// Add other schema properties if needed
}

// GeminiGenerationConfig (Optional): Controls generation parameters.
type GeminiGenerationConfig struct {
	Temperature      float32  `json:"temperature,omitempty"`
	TopP             float32  `json:"topP,omitempty"`
	TopK             int      `json:"topK,omitempty"`
	CandidateCount   int      `json:"candidateCount,omitempty"` // Usually 1
	MaxOutputTokens  int      `json:"maxOutputTokens,omitempty"`
	StopSequences    []string `json:"stopSequences,omitempty"`
	ResponseMimeType string   `json:"responseMimeType,omitempty"` // e.g., "text/plain", "application/json"
	// Add responseSchema if using JSON mode
}

// GeminiSafetySetting (Optional): Configures safety filters.
type GeminiSafetySetting struct {
	Category  string `json:"category"`  // e.g., "HARM_CATEGORY_DANGEROUS_CONTENT"
	Threshold string `json:"threshold"` // e.g., "BLOCK_MEDIUM_AND_ABOVE"
}

// --- Gemini API Response Structures ---

// GeminiResponse is the top-level structure received from the Gemini API.
type GeminiResponse struct {
	Candidates     []GeminiCandidate     `json:"candidates"`
	PromptFeedback *GeminiPromptFeedback `json:"promptFeedback,omitempty"`
	UsageMetadata  *GeminiUsageMetadata  `json:"usageMetadata,omitempty"`
}

// GeminiCandidate represents one possible response from the model.
type GeminiCandidate struct {
	Content          GeminiContent           `json:"content"`                // The model's response content (can include text or functionCall)
	FinishReason     string                  `json:"finishReason,omitempty"` // e.g., "STOP", "MAX_TOKENS", "SAFETY", "RECITATION", "TOOL_CODE", "FUNCTION_CALL"
	Index            int                     `json:"index"`
	SafetyRatings    []GeminiSafetyRating    `json:"safetyRatings,omitempty"`
	CitationMetadata *GeminiCitationMetadata `json:"citationMetadata,omitempty"`
	TokenCount       int                     `json:"tokenCount,omitempty"` // Note: This might be within usageMetadata instead
}

// GeminiSafetyRating for response content.
type GeminiSafetyRating struct {
	Category         string  `json:"category"`
	Probability      string  `json:"probability"` // e.g., "NEGLIGIBLE", "LOW", "MEDIUM", "HIGH"
	ProbabilityScore float32 `json:"probabilityScore,omitempty"`
	Severity         string  `json:"severity,omitempty"`
	SeverityScore    float32 `json:"severityScore,omitempty"`
}

// GeminiCitationMetadata contains citation information if finishReason is "RECITATION".
type GeminiCitationMetadata struct {
	CitationSources []GeminiCitationSource `json:"citationSources,omitempty"`
}

// GeminiCitationSource provides details about a single citation.
type GeminiCitationSource struct {
	StartIndex int    `json:"startIndex,omitempty"`
	EndIndex   int    `json:"endIndex,omitempty"`
	URI        string `json:"uri,omitempty"`
	License    string `json:"license,omitempty"`
}

// GeminiPromptFeedback provides feedback on the input prompt.
type GeminiPromptFeedback struct {
	BlockReason        string               `json:"blockReason,omitempty"`
	BlockReasonMessage string               `json:"blockReasonMessage,omitempty"`
	SafetyRatings      []GeminiSafetyRating `json:"safetyRatings,omitempty"`
}

// GeminiUsageMetadata provides token counts.
type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"` // Sum of tokens across all candidates
	TotalTokenCount      int `json:"totalTokenCount"`
}
