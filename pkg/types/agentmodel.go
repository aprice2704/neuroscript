// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Adds a runtime-only APIKey field to the AgentModel struct to plumb credentials from the account store to the LLM connector.
// filename: pkg/types/agentmodel.go
// nlines: 78
// risk_rating: MEDIUM

package types

// AgentModelName is a typed string for agent model identifiers.
type AgentModelName string

// ResponseFormat specifies the format for the model's output (e.g., json_object).
type ResponseFormat string

const (
	ResponseFormatText ResponseFormat = "text"
	ResponseFormatJSON ResponseFormat = "json_object"
)

// ToolChoice controls how the model uses tools.
type ToolChoice string

const (
	ToolChoiceAuto ToolChoice = "auto"
	ToolChoiceAny  ToolChoice = "any"
	ToolChoiceNone ToolChoice = "none"
)

// AgentModel is the central, unified configuration for an AI agent. It defines
// the model to use, its parameters, and associated settings.
type AgentModel struct {
	Name           AgentModelName `json:"name"`
	Provider       string         `json:"provider"`
	Model          string         `json:"model"`
	AccountName    string         `json:"account_name,omitempty"`
	BaseURL        string         `json:"base_url,omitempty"`
	BudgetCurrency string         `json:"budget_currency,omitempty"`
	Notes          string         `json:"notes,omitempty"`
	Disabled       bool           `json:"disabled,omitempty"`
	ContextKTok    int            `json:"context_ktok,omitempty"`
	MaxTurns       int            `json:"max_turns,omitempty"`
	MaxRetries     int            `json:"max_retries,omitempty"`

	// APIKey is resolved at runtime from the AccountName. It is not persisted
	// or parsed from configuration files.
	APIKey string `json:"-"`

	PriceTable PriceTable       `json:"price_table,omitempty"`
	Generation GenerationConfig `json:"generation,omitempty"`
	Tools      ToolConfig       `json:"tools,omitempty"`
	Safety     SafetyConfig     `json:"safety,omitempty"`

	// --- Deprecated Fields (for backward compatibility) ---
	Temperature       float64 `json:"temperature,omitempty"`
	ToolLoopPermitted bool    `json:"tool_loop_permitted,omitempty"`
	AutoLoopEnabled   bool    `json:"auto_loop_enabled,omitempty"`
}

// PriceTable defines the cost of using a model.
type PriceTable struct {
	InputPerMTok  float64 `json:"input_per_mtok,omitempty"`
	OutputPerMTok float64 `json:"output_per_mtok,omitempty"`
}

// GenerationConfig holds parameters that control the model's output generation.
type GenerationConfig struct {
	Temperature       float64        `json:"temperature,omitempty"`
	TopP              float64        `json:"top_p,omitempty"`
	TopK              int            `json:"top_k,omitempty"`
	MaxOutputTokens   int            `json:"max_output_tokens,omitempty"`
	StopSequences     []string       `json:"stop_sequences,omitempty"`
	PresencePenalty   float64        `json:"presence_penalty,omitempty"`
	FrequencyPenalty  float64        `json:"frequency_penalty,omitempty"`
	RepetitionPenalty float64        `json:"repetition_penalty,omitempty"`
	Seed              *int64         `json:"seed,omitempty"`
	LogProbs          bool           `json:"log_probs,omitempty"`
	ResponseFormat    ResponseFormat `json:"response_format,omitempty"`
}

// ToolConfig holds parameters related to tool use.
type ToolConfig struct {
	ToolLoopPermitted bool       `json:"tool_loop_permitted,omitempty"`
	AutoLoopEnabled   bool       `json:"auto_loop_enabled,omitempty"`
	ToolChoice        ToolChoice `json:"tool_choice,omitempty"`
}

// SafetyConfig holds parameters for content safety.
type SafetyConfig struct {
	SafePrompt bool              `json:"safe_prompt,omitempty"`
	Settings   map[string]string `json:"settings,omitempty"`
}
