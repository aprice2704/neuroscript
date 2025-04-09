// filename: pkg/neurogo/config.go
package neurogo

import (
	"flag"
	"fmt"
	"os" // Import os for Stderr
	"path/filepath"
	"strings"
)

// stringSlice helps handle repeatable flags.
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(value string) error {
	*s = append(*s, filepath.Clean(value)) // Clean paths immediately
	return nil
}

// Config holds all application configuration, typically populated from flags.
type Config struct {
	// General
	LibPaths         []string
	DebugAST         bool
	DebugInterpreter bool

	// Agent Mode
	AgentMode     bool
	AllowlistFile string
	DenylistFiles []string // Paths to additional denylist files (from -denylist flag)
	SandboxDir    string
	APIKey        string // Can be empty, will fallback to ENV

	// Script Mode Specific (Positional Args)
	TargetArg string
	ProcArgs  []string
}

// ParseFlags parses command-line arguments into the App's Config.
func (a *App) ParseFlags(args []string) error {
	// Use os.Stderr for usage output
	fs := flag.NewFlagSet("neurogo", flag.ContinueOnError)
	fs.SetOutput(os.Stderr) // Print usage/errors to stderr

	// Define flags using pointers to Config fields
	var libPathsFlag stringSlice
	var denyPathsFlag stringSlice // Temporary holder for repeatable denylist flag

	fs.Var(&libPathsFlag, "lib", "Specify a library path (file or directory) to load procedures from (repeatable)")
	fs.BoolVar(&a.Config.DebugAST, "debug-ast", false, "Enable verbose AST node logging")
	fs.BoolVar(&a.Config.DebugInterpreter, "debug-interpreter", false, "Enable verbose interpreter execution logging")

	fs.BoolVar(&a.Config.AgentMode, "agent", false, "Run neurogo in LLM agent mode (experimental)")
	fs.StringVar(&a.Config.AllowlistFile, "allowlist", "agent_allowlist.txt", "Path to agent tool allowlist file")
	fs.Var(&denyPathsFlag, "denylist", "Path to an additional agent tool denylist file (repeatable)") // New Denylist Flag
	fs.StringVar(&a.Config.SandboxDir, "sandbox", "./agent_sandbox", "Root directory for agent filesystem sandboxing")
	fs.StringVar(&a.Config.APIKey, "apikey", "", "Override LLM API key (defaults to GEMINI_API_KEY env var)")

	if err := fs.Parse(args); err != nil {
		// Usage is already printed by the flag set on error
		return fmt.Errorf("error parsing flags: %w", err)
	}

	a.Config.LibPaths = libPathsFlag
	a.Config.DenylistFiles = denyPathsFlag // Store optional denylist paths

	// Handle positional arguments
	positionalArgs := fs.Args()
	if !a.Config.AgentMode {
		if len(positionalArgs) == 0 {
			// Explicitly print usage if no target/agent flag provided
			fmt.Fprintf(os.Stderr, "Usage: neurogo [flags] <ProcedureToRun | FileToRun.ns.txt> [proc_args...]\n")
			fmt.Fprintf(os.Stderr, "       neurogo -agent [agent_flags...]\n")
			fmt.Fprintf(os.Stderr, "\nError: Missing procedure name or filename for script mode, and -agent flag not specified.\n")
			fmt.Fprintf(os.Stderr, "\nFlags:\n")
			fs.PrintDefaults() // Print flags again for clarity
			return fmt.Errorf("operation mode unclear: missing script target and -agent flag")
		}
		a.Config.TargetArg = positionalArgs[0]
		a.Config.ProcArgs = positionalArgs[1:]
	} else {
		if len(positionalArgs) > 0 {
			fmt.Fprintf(os.Stderr, "Usage: neurogo [flags] <ProcedureToRun | FileToRun.ns.txt> [proc_args...]\n")
			fmt.Fprintf(os.Stderr, "       neurogo -agent [agent_flags...]\n")
			fmt.Fprintf(os.Stderr, "\nError: Unexpected positional arguments provided in agent mode: %v\n", positionalArgs)
			fmt.Fprintf(os.Stderr, "\nFlags:\n")
			fs.PrintDefaults()
			return fmt.Errorf("unexpected positional arguments in agent mode")
		}
	}

	return nil
}
