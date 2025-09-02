// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Defines the core Account struct, now with a 'Kind' field to support different account types.
// filename: pkg/account/types.go
// nlines: 16
// risk_rating: LOW

package account

// Account holds the credentials and metadata for a specific provider account.
// This decouples authentication details from model or service configurations.
type Account struct {
	Name      string // Logical name, e.g., "prod-openai"
	Kind      string // The type of account, e.g., "llm", "database", "rag"
	Provider  string // Provider name, e.g., "google", "openai", "postgres"
	APIKey    string // The actual API key or password, loaded securely
	OrgID     string // Optional: For providers like OpenAI
	ProjectID string // Optional: For providers like Google
	Notes     string // Optional: For operator convenience
}
