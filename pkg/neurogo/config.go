// filename: pkg/neurogo/config.go
package neurogo

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration defined by command-line flags.
type Config struct {
	ScriptFile          string   // -script: Path to the .ns script file to execute
	SyncDir             string   // -sync-dir: Directory to sync with File API
	SyncFilter          string   // -sync-filter: Glob pattern to filter files during sync
	SyncIgnoreGitignore bool     // -sync-ignore-gitignore: Ignore .gitignore during sync
	SandboxDir          string   // -sandbox: Root directory for agent file operations
	AllowlistFile       string   // -allowlist: Path to the tool allowlist file
	DebugLogFile        string   // -debug-log: Path to the debug log file
	LLMDebugLogFile     string   // -llm-debug-log: Path to the LLM raw communication log file
	InitialAttachments  []string // -attach: List of files to attach initially
	APIKey              string   // API Key (usually from env)
	ModelName           string   // -model: Name of the GenAI model to use
	RunAgentMode        bool     // -agent: Explicitly run in agent mode
	RunSyncMode         bool     // -sync: Explicitly run sync using config dir
	RunTuiMode          bool     // -tui: Explicitly run in TUI mode
	EnableLLM           bool     // -enable-llm: Enable LLM client (default true, mainly affects script mode)
	Insecure            bool     // TODO: What does this do? Is it still needed?

	// Renamed from Nuke
	CleanAPI bool // -clean-api: Delete all files from the File API

	LibPaths  []string // -L: Library paths for script execution
	TargetArg string   // -target: Target argument for the script
	ProcArgs  []string // -arg: Arguments for the script process/procedure

	// Internal fields
	flagSet *flag.FlagSet
}

// NewConfig creates a new Config struct with default values.
func NewConfig() *Config {
	return &Config{
		SyncDir:    ".",
		SandboxDir: ".",
		EnableLLM:  true, // Default LLM client to enabled
	}
}

// StringSliceFlag is a custom flag type for handling multiple occurrences of a flag.
type stringSliceFlag []string

func (i *stringSliceFlag) String() string         { return strings.Join(*i, ", ") }
func (i *stringSliceFlag) Set(value string) error { *i = append(*i, value); return nil }

// ParseFlags parses command-line arguments into the Config struct.
func (c *Config) ParseFlags(args []string) error {
	fs := flag.NewFlagSet("neurogo", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	// --- Define Execution Mode Flags ---
	fs.StringVar(&c.ScriptFile, "script", "", "Path to the .ns script file to execute.")
	fs.BoolVar(&c.RunSyncMode, "sync", false, "Run in sync mode using -sync-dir (or default './').")
	fs.BoolVar(&c.RunAgentMode, "agent", false, "Run in interactive agent mode.")
	fs.BoolVar(&c.RunTuiMode, "tui", false, "Run in interactive TUI mode.") // Added -tui flag
	// Renamed from -nuke
	fs.BoolVar(&c.CleanAPI, "clean-api", false, "Delete ALL files from the File API (use with caution!). Must be used alone.")

	// --- Define Configuration Flags ---
	fs.StringVar(&c.SyncDir, "sync-dir", c.SyncDir, "Directory to sync with File API (used by /sync cmd, ignored if -sync flag is set).")
	fs.StringVar(&c.SyncFilter, "sync-filter", "", "Glob pattern (filename only) to filter files during sync.")
	fs.BoolVar(&c.SyncIgnoreGitignore, "sync-ignore-gitignore", false, "Ignore .gitignore file during sync.")
	fs.StringVar(&c.SandboxDir, "sandbox", c.SandboxDir, "Root directory for safe agent file operations.")
	fs.StringVar(&c.AllowlistFile, "allowlist", "", "Path to the tool allowlist file.")
	fs.StringVar(&c.DebugLogFile, "debug-log", "", "Path to write detailed debug logs.")
	fs.StringVar(&c.LLMDebugLogFile, "llm-debug-log", "", "Path to write raw LLM request/response logs.")
	fs.StringVar(&c.ModelName, "model", "", "Optional: GenAI model name (e.g., gemini-1.5-flash-latest).")
	fs.BoolVar(&c.EnableLLM, "enable-llm", c.EnableLLM, "Enable LLM client (default true, use -enable-llm=false to disable).") // Added -enable-llm flag

	// --- Flags for Agent/Script Context ---
	var attachments stringSliceFlag
	fs.Var(&attachments, "attach", "File path to attach to the agent session initially (can be used multiple times).")
	var libPaths stringSliceFlag
	fs.Var(&libPaths, "L", "Library path for NeuroScript execution (can be used multiple times).")
	fs.StringVar(&c.TargetArg, "target", "", "Target argument passed to the main script procedure.")
	var procArgs stringSliceFlag
	fs.Var(&procArgs, "arg", "Argument passed to the main script procedure (can be used multiple times).")

	// Configure Usage message
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage of neurogo:\n")
		fmt.Fprintf(fs.Output(), "  neurogo [flags]\n\n")
		// Updated precedence and flag name
		fmt.Fprintf(fs.Output(), "Modes (mutually exclusive, precedence: -clean-api > -sync > -script > -tui > -agent (default)):\n")
		fmt.Fprintf(fs.Output(), "  -clean-api           : Delete all files from API (requires confirmation).\n")
		fmt.Fprintf(fs.Output(), "  -sync                : Run file synchronization and exit.\n")
		fmt.Fprintf(fs.Output(), "  -script <file.ns>    : Execute the specified NeuroScript file.\n")
		fmt.Fprintf(fs.Output(), "  -tui                 : Run in interactive Text User Interface (TUI) mode.\n") // Added TUI
		fmt.Fprintf(fs.Output(), "  -agent               : Run in interactive command-line agent mode (default).\n")
		fmt.Fprintf(fs.Output(), "\nCommon Flags:\n")
		fmt.Fprintf(fs.Output(), "  -sandbox <dir>       : Root directory for agent file operations (default: %q)\n", c.SandboxDir)
		fmt.Fprintf(fs.Output(), "  -allowlist <file>    : Path to the tool allowlist file.\n")
		fmt.Fprintf(fs.Output(), "  -attach <file>       : File to attach initially (repeatable).\n")
		fmt.Fprintf(fs.Output(), "  -model <name>        : GenAI model name (optional).\n")
		fmt.Fprintf(fs.Output(), "  -enable-llm=<bool>   : Enable LLM client (default: %t).\n", c.EnableLLM) // Added EnableLLM
		fmt.Fprintf(fs.Output(), "\nSync Flags (used with -sync or /sync command):\n")
		fmt.Fprintf(fs.Output(), "  -sync-dir <dir>      : Directory to sync (default: %q)\n", c.SyncDir)
		fmt.Fprintf(fs.Output(), "  -sync-filter <pat>   : Glob pattern for filenames.\n")
		fmt.Fprintf(fs.Output(), "  -sync-ignore-gitignore: Ignore .gitignore file.\n")
		fmt.Fprintf(fs.Output(), "\nScript Flags (used with -script):\n")
		fmt.Fprintf(fs.Output(), "  -L <path>            : Library path for NeuroScript (repeatable).\n")
		fmt.Fprintf(fs.Output(), "  -target <arg>        : Target argument for the script.\n")
		fmt.Fprintf(fs.Output(), "  -arg <arg>           : Argument for the script (repeatable).\n")
		fmt.Fprintf(fs.Output(), "\nLogging Flags:\n")
		fmt.Fprintf(fs.Output(), "  -debug-log <file>    : Path for detailed debug logs.\n")
		fmt.Fprintf(fs.Output(), "  -llm-debug-log <file>: Path for raw LLM request/response logs.\n")
		fmt.Fprintf(fs.Output(), "\nOther:\n")
		fmt.Fprintf(fs.Output(), "  -h, -help            : Show this help message.\n")
	}

	// Parse the flags
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	c.InitialAttachments = attachments
	c.LibPaths = libPaths
	c.ProcArgs = procArgs
	c.flagSet = fs

	// --- Validate Flag Combinations ---
	if c.CleanAPI {
		cleanApiOnly := true
		for _, arg := range args {
			isCleanApiFlag := arg == "-clean-api"
			isLogFlag := arg == "-debug-log" || arg == "-llm-debug-log" || strings.HasPrefix(arg, "-debug-log=") || strings.HasPrefix(arg, "-llm-debug-log=")
			isModelFlag := arg == "-model" || strings.HasPrefix(arg, "-model=")
			// Allow other global flags here?

			if !isCleanApiFlag && !isLogFlag && !isModelFlag && strings.HasPrefix(arg, "-") {
				cleanApiOnly = false
				break
			}
			if !strings.HasPrefix(arg, "-") {
				cleanApiOnly = false
				break
			}
		}
		if !cleanApiOnly {
			fs.Usage()
			return errors.New("the -clean-api flag must be used alone (potentially with logging or model flags)")
		}
	}

	// Check other mode combinations
	otherModeCount := 0
	if c.RunSyncMode {
		otherModeCount++
	}
	if c.ScriptFile != "" {
		otherModeCount++
	}
	if c.RunAgentMode {
		otherModeCount++
	}
	if c.RunTuiMode { // Added TUI check
		otherModeCount++
	}

	if otherModeCount > 1 {
		fs.Usage()
		return errors.New("flags -sync, -script, -tui, and -agent are mutually exclusive")
	}

	// Default to agent mode if no other primary mode flag is set (check CleanAPI)
	if !c.CleanAPI && !c.RunSyncMode && c.ScriptFile == "" && !c.RunTuiMode && !c.RunAgentMode {
		c.RunAgentMode = true
		fmt.Fprintln(os.Stderr, "Defaulting to interactive agent mode.")
	}

	// API Key Check
	c.APIKey = os.Getenv("GEMINI_API_KEY")
	if c.APIKey == "" {
		helpRequested := false
		for _, arg := range args {
			if arg == "-h" || arg == "-help" {
				helpRequested = true
				break
			}
		}
		// Modes that *might* need an API key (LLM can be disabled for script)
		// TUI and Agent modes inherently need the LLM.
		// Sync and CleanAPI need it for file API access.
		needsKey := c.RunAgentMode || c.RunTuiMode || c.RunSyncMode || c.CleanAPI || (c.ScriptFile != "" && c.EnableLLM)
		if !helpRequested && needsKey {
			return errors.New("required environment variable for API Key (e.g., GEMINI_API_KEY) is not set")
		}
	}

	return nil // Success
}
