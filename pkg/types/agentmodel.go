// NeuroScript Version: 0.6.0
// File version: 1.0.0
// Purpose: Provides the canonical, shared definition for the AgentModel struct to break import cycles.
// filename: pkg/types/agentmodel.go
// nlines: 25
// risk_rating: LOW

package types

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

	// Fields migrated from agentmodel.Info
	Notes    string
	Disabled bool
}
