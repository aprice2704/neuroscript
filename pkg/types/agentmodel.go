// NeuroScript Version: 0.7.0
// File version: 7
// Purpose: Expanded AgentModel to include comprehensive generation, tool, and safety controls.
// filename: pkg/types/agentmodel.go
// nlines: 110
// risk_rating: MEDIUM

package types

import "encoding/json"

// ConstitutionLevel indicates how a model's constitution is applied.
type ConstitutionLevel uint8

const (
	ConstitutionNone       ConstitutionLevel = iota // No constitution applied
	ConstitutionSysPrompt                           // Constitution applied via system prompt
	ConstitutionInTraining                          // Constitution baked in during model training
)

// ResponseFormat specifies the desired output format from the model.
type ResponseFormat string

const (
	ResponseFormatDefault ResponseFormat = ""
	ResponseFormatText    ResponseFormat = "text"
	ResponseFormatJSON    ResponseFormat = "json_object"
)

// ToolChoice sets the mode for tool execution.
type ToolChoice string

const (
	ToolChoiceDefault ToolChoice = ""
	ToolChoiceAuto    ToolChoice = "auto"
	ToolChoiceAny     ToolChoice = "any"
	ToolChoiceNone    ToolChoice = "none"
)

// GenerationConfig holds parameters that control the model's creative output.
type GenerationConfig struct {
	Temperature       float64
	TopP              float64
	TopK              int
	MaxOutputTokens   int
	StopSequences     []string
	PresencePenalty   float64
	FrequencyPenalty  float64
	RepetitionPenalty float64 // For Meta Llama
	Seed              *int64  // Use a pointer to distinguish between 0 and not set
	LogitBias         map[string]int
	LogProbs          bool
	ResponseFormat    ResponseFormat
}

// ToolConfig defines how the model should interact with tools.
type ToolConfig struct {
	ToolLoopPermitted bool
	AutoLoopEnabled   bool // If true, a single 'ask' will automatically start the multi-turn loop.
	ToolChoice        ToolChoice
}

// SafetyConfig defines safety-related parameters.
type SafetyConfig struct {
	// For providers like Mistral, this can be used to enable a "safe mode".
	SafePrompt bool
	// For providers like Google, this allows for fine-grained control over safety categories.
	// The key is the category (e.g., "HARM_CATEGORY_HARASSMENT") and the value is the
	// threshold (e.g., "BLOCK_MEDIUM_AND_ABOVE").
	Settings map[string]string
}

// AgentModel holds the validated and parsed configuration for a specific AI model endpoint.
// This is the canonical definition used throughout the system.
type AgentModel struct {
	Name           AgentModelName
	Provider       string
	Model          string
	SecretRef      string
	BaseURL        string
	BudgetCurrency string
	PriceTable     map[string]float64

	// Core Generation Controls
	Generation GenerationConfig

	// Tool & Loop Controls
	Tools ToolConfig

	// Safety Controls
	Safety SafetyConfig

	// Context and Constitution
	ContextKTok        int // Maximum advertised context in thousands of tokens (e.g., 1000 for 1M tokens)
	IdealContextKTok   int // The recommended context size to use in kTok.
	ConstitutionLevel  ConstitutionLevel
	ConstitutionSource string // Source identifier for the constitution (e.g., a file path or URL)

	// Rate & Token Limits
	MaxReqPerSec    int
	MaxReqPerMinute int
	MaxReqPerHour   int
	MaxReqPerDay    int
	MaxTokPerSec    int
	MaxTokPerMinute int

	// Ask Loop Control (Legacy fields, kept for backward compatibility during transition)
	// These will be migrated to ToolConfig and GenerationConfig over time.
	// New implementations should prefer the structured configs.
	ToolLoopPermitted bool    `json:"-"` // Deprecated: use Tools.ToolLoopPermitted
	AutoLoopEnabled   bool    `json:"-"` // Deprecated: use Tools.AutoLoopEnabled
	Temperature       float64 `json:"-"` // Deprecated: use Generation.Temperature
	MaxTurns          int
	MaxRetries        int

	// Metadata
	Notes    string
	Disabled bool
}

// String provides a JSON representation of the AgentModel for logging/debugging.
func (am AgentModel) String() string {
	b, err := json.MarshalIndent(am, "", "  ")
	if err != nil {
		return "failed to marshal AgentModel"
	}
	return string(b)
}
