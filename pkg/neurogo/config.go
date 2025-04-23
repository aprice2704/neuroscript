// filename: pkg/neurogo/config.go
package neurogo

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration.
// Flag parsing is primarily handled in main.go, but this struct holds the values.
type Config struct {
	// --- Execution Mode Flags ---
	// These are set by main.go based on which flag was provided or default logic.
	RunAgentMode    bool // -agent
	RunScriptMode   bool // -script (Implied by ScriptFile != "")
	RunSyncMode     bool // -sync
	RunTuiMode      bool // -tui
	RunCleanAPIMode bool // -clean-api

	// --- Script/Agent Execution ---
	ScriptFile    string   // -script: Path to the .ns script file to execute
	StartupScript string   // -startup-script: Path to agent initialization script (NEW)
	LibPaths      []string // -L: Library paths for script execution
	TargetArg     string   // -target: Target argument for the script
	ProcArgs      []string // -arg: Arguments for the script process/procedure

	// --- Sync Operation ---
	SyncDir             string // -sync-dir: Directory to sync with File API
	SyncFilter          string // -sync-filter: Glob pattern to filter files during sync
	SyncIgnoreGitignore bool   // -sync-ignore-gitignore: Ignore .gitignore during sync

	// --- Agent/General Configuration (May be overridden by startup script) ---
	SandboxDir         string   // -sandbox: Root directory for agent file operations (set by main, used by app?)
	AllowlistFile      string   // -allowlist: Path to the tool allowlist file (set by main, used by app?)
	InitialAttachments []string // -attach: List of files to attach initially (DEPRECATED? Handle via startup?)
	APIKey             string   // API Key (usually from env)
	ModelName          string   // -model: Name of the GenAI model to use (set by main, potentially overridden by startup)
	EnableLLM          bool     // -enable-llm: Enable LLM client (default true, mainly affects script mode) (set by main)
	Insecure           bool     // -insecure: Disable security checks (Use with extreme caution!) (set by main)
	CleanAPI           bool     // -clean-api: Delete all files from the File API (set by main)

	// --- Logging ---
	DebugLogFile    string // -debug-log: Path to the debug log file
	LLMDebugLogFile string // -llm-debug-log: Path to the LLM raw communication log file

	// Internal fields (Consider removing flagSet if parsing is fully in main.go)
	// flagSet *flag.FlagSet
}

// NewConfig creates a new Config struct with default values.
// Defaults for flags removed from here are handled in main.go's flag definitions.
func NewConfig() *Config {
	return &Config{
		// Defaults for SyncDir/SandboxDir now set in main.go's flag definitions
		EnableLLM: true, // Default LLM client to enabled
	}
}

// --- NOTE: The ParseFlags method below is likely now obsolete ---
// Flag parsing logic has been moved to main.go.
// Keeping the method signature here might be useful if some validation logic
// specific to Config fields needs to live here, but the flag definitions and parsing
// should primarily occur in main.go.
// Consider removing or refactoring this method based on final design.

// StringSliceFlag is a custom flag type for handling multiple occurrences of a flag.
type stringSliceFlag []string

func (i *stringSliceFlag) String() string         { return strings.Join(*i, ", ") }
func (i *stringSliceFlag) Set(value string) error { *i = append(*i, value); return nil }

// ParseFlags parses command-line arguments into the Config struct.
// OBSOLETE? Flag parsing is now primarily in main.go.
// This method might need removal or significant refactoring.
func (c *Config) ParseFlags(args []string) error {
	// If this method is kept, it should likely only perform validation
	// on the fields already populated by main.go, rather than defining/parsing flags itself.
	fmt.Fprintln(os.Stderr, "Warning: Config.ParseFlags called, but flag parsing should occur in main.go.")

	// Example validation (can be expanded):
	if c.RunCleanAPIMode {
		modeCount := 0
		if c.RunAgentMode {
			modeCount++
		}
		if c.RunScriptMode {
			modeCount++
		}
		if c.RunSyncMode {
			modeCount++
		}
		if c.RunTuiMode {
			modeCount++
		}
		if modeCount > 0 {
			return errors.New("the -clean-api mode cannot be combined with -agent, -script, -sync, or -tui")
		}
	} else {
		modeCount := 0
		if c.RunAgentMode {
			modeCount++
		}
		if c.RunScriptMode {
			modeCount++
		}
		if c.RunSyncMode {
			modeCount++
		}
		if c.RunTuiMode {
			modeCount++
		}
		if modeCount > 1 {
			return errors.New("modes -agent, -script, -sync, and -tui are mutually exclusive")
		}
		// Defaulting logic now happens in main.go
	}

	// API Key Check (still relevant if done here)
	// This check might be better placed in App.Run or where the client is initialized.
	if c.APIKey == "" {
		needsKey := c.RunAgentMode || c.RunTuiMode || c.RunSyncMode || c.RunCleanAPIMode || (c.RunScriptMode && c.EnableLLM)
		if needsKey {
			// Check if help was requested? This info isn't easily available here anymore.
			// Assume key is needed if mode requires it.
			return errors.New("required environment variable for API Key (e.g., GEMINI_API_KEY) is not set")
		}
	}

	return nil // Success (or return validation errors)
}
