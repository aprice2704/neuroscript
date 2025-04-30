// filename: pkg/neurogo/config.go
package neurogo

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration.
type Config struct {
	// --- Execution Mode Flags ---
	RunAgentMode    bool // -agent
	RunScriptMode   bool // -script (Implied by ScriptFile != "")
	RunSyncMode     bool // -sync
	RunTuiMode      bool // -tui (Corrected case)
	RunCleanAPIMode bool // -clean-api

	// --- Script/Agent Execution ---
	ScriptFile    string   // -script: Path to the .ns script file to execute
	StartupScript string   // -startup-script: Path to agent initialization script
	LibPaths      []string // -L: Library paths for script execution
	TargetArg     string   // -target: Target argument for the script
	ProcArgs      []string // -arg: Arguments for the script process/procedure

	// --- Sync Operation ---
	SyncDir             string // -sync-dir: Directory to sync with File API
	SyncFilter          string // -sync-filter: Glob pattern to filter files during sync
	SyncIgnoreGitignore bool   // -sync-ignore-gitignore: Ignore .gitignore during sync

	// --- Agent/General Configuration ---
	SandboxDir    string // -sandbox: Root directory for agent file operations
	AllowlistFile string // -allowlist: Path to the tool allowlist file
	APIKey        string // API Key (usually from env)
	APIHost       string // API Host / Endpoint (e.g., for custom LLM providers) <<< ADDED
	ModelID       string // -model / ModelName field: Specific model identifier <<< ADDED (using ModelID for clarity)
	EnableLLM     bool   // -enable-llm: Enable LLM client
	Insecure      bool   // -insecure: Disable security checks
	CleanAPI      bool   // -clean-api: Delete all files from the File API

	// --- Logging ---
	DebugLogFile    string // -debug-log: Path to the debug log file
	LLMDebugLogFile string // -llm-debug-log: Path to the LLM raw communication log file

	// --- Schema ---
	SchemaPath string // Path to the schema definition file <<< ADDED (Placeholder for previous discussion)

	// Note: Removed InitialAttachments as it seems deprecated/handled by startup script
}

// NewConfig creates a new Config struct with default values.
func NewConfig() *Config {
	return &Config{
		EnableLLM: true, // Default LLM client to enabled
		// Default other fields like APIHost if applicable
		// APIHost: "default-llm-host.com",
	}
}

// StringSliceFlag is a custom flag type for handling multiple occurrences of a flag.
type stringSliceFlag []string

func (i *stringSliceFlag) String() string         { return strings.Join(*i, ", ") }
func (i *stringSliceFlag) Set(value string) error { *i = append(*i, value); return nil }

// ParseFlags performs validation after flags have been parsed (likely in main.go).
func (c *Config) ParseFlags(args []string) error {
	// This function assumes flags have already been parsed into 'c' by main.go
	fmt.Fprintln(os.Stderr, "Warning: Config.ParseFlags called, performing post-parse validation.")

	// Mode Exclusivity Check
	modeCount := 0
	modes := []bool{c.RunAgentMode, c.RunScriptMode, c.RunSyncMode, c.RunTuiMode, c.RunCleanAPIMode}
	for _, mode := range modes {
		if mode {
			modeCount++
		}
	}

	if c.RunCleanAPIMode && modeCount > 1 {
		return errors.New("the -clean-api mode cannot be combined with -agent, -script, -sync, or -tui")
	}
	if !c.RunCleanAPIMode && modeCount > 1 {
		// Identify which modes are conflicting for a better error message?
		return errors.New("modes -agent, -script, -sync, and -tui are mutually exclusive")
	}
	if modeCount == 0 {
		// If no mode is explicitly set, default logic should apply (handled in App.Run now)
		// return errors.New("no execution mode specified (e.g., -agent, -script, -sync, -tui, -clean-api)")
		fmt.Fprintln(os.Stderr, "Info: No specific mode flag set, will default based on LLM enablement.")
	}

	// API Key Check (Crucial if LLM is needed)
	needsKey := c.EnableLLM && (c.RunAgentMode || c.RunTuiMode || c.RunScriptMode || c.RunSyncMode || c.RunCleanAPIMode) // Refined condition
	if needsKey && c.APIKey == "" {
		// Removed environment check here, assumes key is populated by main.go if found
		return errors.New("API Key is required for the selected mode with LLM enabled, but is missing")
	}

	// Validate other required fields based on modes if necessary

	return nil // Success
}
