// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Removed the generic 'Config' map, as the registration process will now populate the explicit fields directly.
// filename: pkg/types/agentmodel.go
// nlines: 50
// risk_rating: LOW

package types

// ConstitutionLevel indicates how a model's constitution is applied.
type ConstitutionLevel uint8

const (
	ConstitutionNone       ConstitutionLevel = iota // No constitution applied
	ConstitutionSysPrompt                           // Constitution applied via system prompt
	ConstitutionInTraining                          // Constitution baked in during model training
)

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
	Temperature    float64

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

	// Ask Loop Control
	ToolLoopPermitted bool
	MaxTurns          int
	MaxRetries        int

	// Fields migrated from agentmodel.Info
	Notes    string
	Disabled bool
}
