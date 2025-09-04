// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: Aligned struct with the snake_case JSON key standard.
// filename: pkg/account/types.go
// nlines: 18
// risk_rating: LOW

package account

// Account holds the credentials and metadata for a specific provider account.
// This decouples authentication details from model or service configurations.
type Account struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`       // The type of account, e.g., "llm", "database", "rag"
	Provider  string `json:"provider"`   // Provider name, e.g., "google", "openai", "postgres"
	APIKey    string `json:"api_key"`    // The actual API key or password, loaded securely
	OrgID     string `json:"org_id"`     // Optional: For providers like OpenAI
	ProjectID string `json:"project_id"` // Optional: For providers like Google
	Notes     string `json:"notes"`      // Optional: For operator convenience
}
