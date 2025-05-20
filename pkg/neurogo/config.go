// NeuroScript Version: 0.3.0
// File version: 0.1.1
// Added TuiMode field to Config struct.
// filename: pkg/neurogo/config.go
// nlines: 55 // Approximate
// risk_rating: LOW
package neurogo

import (
	"strings"
)

// Config holds the application configuration.
type Config struct {
	// --- Execution Flags (Informational/Control, not strict modes) ---
	TuiMode bool // -tui: Explicitly run in TUI mode. If false and other modes (like -script) are not set, TUI may still be default.

	// --- Script/Agent Execution ---
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
	Insecure      bool   // -insecure: Disable security checks

	// --- Logging ---
	// Log related flags are typically handled in main.go during logger setup.
	// LogFile string
	// LogLevel string

	// --- Schema ---
	SchemaPath string // Path to the schema definition file (Placeholder for future use)
}

// NewConfig creates a new Config struct with default values.
func NewConfig() *Config {
	return &Config{
		// Set any necessary defaults here
		// TuiMode could default to false, and main.go can set it to true if -tui is present
		// or if no other exclusive mode flags are set.
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
func NewStringSliceFlag() *stringSliceFlag {
	return &stringSliceFlag{}
}
