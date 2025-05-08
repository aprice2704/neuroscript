// NeuroScript Version: 0.3.0
// File version: 0.1.0 // Updated version
// Refactored Config for AI Worker Manager integration
// filename: pkg/neurogo/config.go
package neurogo

import (
	// "errors" // No longer needed here
	// "fmt" // No longer needed here
	// "os" // No longer needed here
	"strings"
)

// Config holds the application configuration.
type Config struct {
	// --- Execution Flags (Informational/Control, not strict modes) ---
	// RunTuiMode      bool // -tui flag still used in main.go to launch TUI
	// RunCleanAPIMode bool // Functionality likely moved to a tool

	// --- Script/Agent Execution ---
	// ScriptFile is now handled by StartupScript
	StartupScript string   // -script: Path to the .ns script file to execute on startup
	LibPaths      []string // -L: Library paths for script execution
	TargetArg     string   // -target: Target argument for the script (if StartupScript is run)
	ProcArgs      []string // -arg: Arguments for the script process/procedure (if StartupScript is run)

	// --- Sync Operation (Values potentially used by tools) ---
	SyncDir             string // -sync-dir: Directory to sync with File API
	SyncFilter          string // -sync-filter: Glob pattern to filter files during sync
	SyncIgnoreGitignore bool   // -sync-ignore-gitignore: Ignore .gitignore during sync

	// --- General Configuration ---
	SandboxDir    string // -sandbox: Root directory for file operations & ai_wm persistence
	AllowlistFile string // -allowlist: Path to the tool allowlist file (might be superseded by AI WM definitions)
	APIKey        string // API Key (usually from env)
	APIHost       string // API Host / Endpoint (e.g., for custom LLM providers)
	ModelName     string // -model: Default/Fallback model identifier
	// EnableLLM flag removed, LLM client creation is now standard
	Insecure bool // -insecure: Disable security checks

	// --- Logging ---
	// DebugLogFile    string // Handled directly in main.go logger setup
	// LLMDebugLogFile string // Handled directly in main.go logger setup

	// --- Schema ---
	SchemaPath string // Path to the schema definition file (Placeholder for future use)
}

// NewConfig creates a new Config struct with default values.
func NewConfig() *Config {
	return &Config{
		// Set any necessary defaults here
	}
}

// StringSliceFlag remains the same.
type stringSliceFlag struct {
	Value []string
}

func (f *stringSliceFlag) String() string {
	return strings.Join(f.Value, ", ")
}

func (f *stringSliceFlag) Set(value string) error {
	f.Value = append(f.Value, value)
	return nil
}

// NewStringSliceFlag creates a new empty StringSliceFlag.
// Required because flag.Var needs a pointer to a non-nil value.
func NewStringSliceFlag() *stringSliceFlag {
	return &stringSliceFlag{}
}

// ParseFlags is removed as validation logic is simplified and moved mostly to main.go
// func (c *Config) ParseFlags(args []string) error { ... }
