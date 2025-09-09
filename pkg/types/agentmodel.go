// NeuroScript Version: 0.7.0
// File version: 6
// Purpose: Adds mapstructure tags to the AgentModel struct for automated config parsing.
// filename: pkg/types/agentmodel.go
// nlines: 81
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
	Name           AgentModelName `json:"name" mapstructure:"name"`
	Provider       string         `json:"provider" mapstructure:"provider"`
	Model          string         `json:"model" mapstructure:"model"`
	AccountName    string         `json:"account_name,omitempty" mapstructure:"account_name"`
	BaseURL        string         `json:"base_url,omitempty" mapstructure:"base_url"`
	BudgetCurrency string         `json:"budget_currency,omitempty" mapstructure:"budget_currency"`
	Notes          string         `json:"notes,omitempty" mapstructure:"notes"`
	Disabled       bool           `json:"disabled,omitempty" mapstructure:"disabled"`
	ContextKTok    int            `json:"context_ktok,omitempty" mapstructure:"context_ktok"`
	MaxTurns       int            `json:"max_turns,omitempty" mapstructure:"max_turns"`
	MaxRetries     int            `json:"max_retries,omitempty" mapstructure:"max_retries"`

	// APIKey is resolved at runtime from the AccountName. It is not persisted
	// or parsed from configuration files.
	APIKey string `json:"-" mapstructure:"-"`

	PriceTable PriceTable       `json:"price_table,omitempty" mapstructure:",squash"`
	Generation GenerationConfig `json:"generation,omitempty" mapstructure:",squash"`
	Tools      ToolConfig       `json:"tools,omitempty" mapstructure:",squash"`
	Safety     SafetyConfig     `json:"safety,omitempty" mapstructure:",squash"`

	// --- Deprecated Fields (for backward compatibility) ---
	Temperature       float64 `json:"temperature,omitempty" mapstructure:"temperature"`
	ToolLoopPermitted bool    `json:"tool_loop_permitted,omitempty" mapstructure:"tool_loop_permitted"`
	AutoLoopEnabled   bool    `json:"auto_loop_enabled,omitempty" mapstructure:"auto_loop_enabled"`
}

// PriceTable defines the cost of using a model.
type PriceTable struct {
	InputPerMTok  float64 `json:"input_per_mtok,omitempty" mapstructure:"input_per_mtok"`
	OutputPerMTok float64 `json:"output_per_mtok,omitempty" mapstructure:"output_per_mtok"`
}

// GenerationConfig holds parameters that control the model's output generation.
type GenerationConfig struct {
	Temperature       float64        `json:"temperature,omitempty" mapstructure:"temperature"`
	TopP              float64        `json:"top_p,omitempty" mapstructure:"top_p"`
	TopK              int            `json:"top_k,omitempty" mapstructure:"top_k"`
	MaxOutputTokens   int            `json:"max_output_tokens,omitempty" mapstructure:"max_output_tokens"`
	StopSequences     []string       `json:"stop_sequences,omitempty" mapstructure:"stop_sequences"`
	PresencePenalty   float64        `json:"presence_penalty,omitempty" mapstructure:"presence_penalty"`
	FrequencyPenalty  float64        `json:"frequency_penalty,omitempty" mapstructure:"frequency_penalty"`
	RepetitionPenalty float64        `json:"repetition_penalty,omitempty" mapstructure:"repetition_penalty"`
	Seed              *int64         `json:"seed,omitempty" mapstructure:"seed"`
	LogProbs          bool           `json:"log_probs,omitempty" mapstructure:"log_probs"`
	ResponseFormat    ResponseFormat `json:"response_format,omitempty" mapstructure:"response_format"`
}

// ToolConfig holds parameters related to tool use.
type ToolConfig struct {
	ToolLoopPermitted bool       `json:"tool_loop_permitted,omitempty" mapstructure:"tool_loop_permitted"`
	AutoLoopEnabled   bool       `json:"auto_loop_enabled,omitempty" mapstructure:"auto_loop_enabled"`
	ToolChoice        ToolChoice `json:"tool_choice,omitempty" mapstructure:"tool_choice"`
}

// SafetyConfig holds parameters for content safety.
type SafetyConfig struct {
	SafePrompt bool              `json:"safe_prompt,omitempty" mapstructure:"safe_prompt"`
	Settings   map[string]string `json:"settings,omitempty" mapstructure:"safety_settings"`
}
