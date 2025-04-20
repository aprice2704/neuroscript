// filename: pkg/neurogo/config.go
package neurogo

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// stringSlice helper (unchanged)
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(value string) error {
	*s = append(*s, filepath.Clean(value))
	return nil
}

const defaultSandboxSubdir = "agent_sandbox"

// Config holds all application configuration, typically populated from flags.
type Config struct {
	// General
	LibPaths         []string // -lib
	DebugAST         bool     // -debug-ast
	DebugInterpreter bool     // -debug-interpreter
	// +++ ADDED: DebugLLM field +++
	DebugLLM  bool   // -debug-llm
	ModelName string // -model

	// Agent Mode
	AgentMode     bool     // -agent
	AllowlistFile string   // -allowlist
	DenylistFiles []string // -denylist
	SandboxDir    string   // -sandbox
	APIKey        string   // -apikey

	// Script Mode Specific
	TargetArg string   // Positional arg 1
	ProcArgs  []string // Positional args 2+
}

// ParseFlags parses command-line arguments into the App's Config.
func (a *App) ParseFlags(args []string) error {
	if a == nil {
		return fmt.Errorf("internal error: ParseFlags called on nil App")
	}

	fs := flag.NewFlagSet("neurogo", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var libPathsFlag stringSlice
	var denyPathsFlag stringSlice

	// General Flags
	fs.Var(&libPathsFlag, "lib", "Specify a library path (file or directory) to load procedures from (repeatable)")
	fs.BoolVar(&a.Config.DebugAST, "debug-ast", false, "Enable verbose AST node logging")
	fs.BoolVar(&a.Config.DebugInterpreter, "debug-interpreter", false, "Enable verbose interpreter execution logging")
	// +++ ADDED: Define -debug-llm flag +++
	fs.BoolVar(&a.Config.DebugLLM, "debug-llm", false, "Enable verbose LLM API interaction logging")
	fs.StringVar(&a.Config.ModelName, "model", "gemini-1.5-pro-latest", "Name of the generative model to use.")

	// Agent Flags
	fs.BoolVar(&a.Config.AgentMode, "agent", false, "Run neurogo in LLM agent mode (experimental)")
	fs.StringVar(&a.Config.AllowlistFile, "allowlist", "agent_allowlist.txt", "Path to agent tool allowlist file")
	fs.Var(&denyPathsFlag, "denylist", "Path to an additional agent tool denylist file (repeatable)")
	exePath, errExe := os.Executable()
	defaultSandbox := defaultSandboxSubdir
	if errExe == nil {
		defaultSandbox = filepath.Join(filepath.Dir(exePath), defaultSandboxSubdir)
	} else {
		fmt.Fprintf(os.Stderr, "[Warning] Could not determine executable path, using default sandbox subdir '%s': %v\n", defaultSandboxSubdir, errExe)
	}
	fs.StringVar(&a.Config.SandboxDir, "sandbox", defaultSandbox, "Root directory for agent filesystem sandboxing")
	fs.StringVar(&a.Config.APIKey, "apikey", "", "Override LLM API key (defaults to GEMINI_API_KEY env var)")

	// Parse
	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("error parsing flags: %w", err)
	}

	// Assign repeatable flags
	a.Config.LibPaths = libPathsFlag
	a.Config.DenylistFiles = denyPathsFlag

	// Handle positional arguments (unchanged)
	positionalArgs := fs.Args()
	if !a.Config.AgentMode { // Script Mode
		if len(positionalArgs) == 0 {
			fmt.Fprintf(os.Stderr, "Usage: neurogo [flags] <ProcedureToRun | FileToRun.ns.txt> [proc_args...]\n")
			fmt.Fprintf(os.Stderr, "       neurogo -agent [agent_flags...]\n")
			fmt.Fprintf(os.Stderr, "\nError: Missing procedure name or filename for script mode, and -agent flag not specified.\n")
			fmt.Fprintf(os.Stderr, "\nFlags:\n")
			fs.PrintDefaults()
			return fmt.Errorf("operation mode unclear: missing script target and -agent flag not specified")
		}
		a.Config.TargetArg = positionalArgs[0]
		a.Config.ProcArgs = positionalArgs[1:]
	} else { // Agent Mode
		if len(positionalArgs) > 0 {
			fmt.Fprintf(os.Stderr, "Usage: neurogo [flags] <ProcedureToRun | FileToRun.ns.txt> [proc_args...]\n")
			fmt.Fprintf(os.Stderr, "       neurogo -agent [agent_flags...]\n")
			fmt.Fprintf(os.Stderr, "\nError: Unexpected positional arguments provided in agent mode: %v\n", positionalArgs)
			fmt.Fprintf(os.Stderr, "\nFlags:\n")
			fs.PrintDefaults()
			return fmt.Errorf("unexpected positional arguments in agent mode")
		}
		a.Config.TargetArg = ""
		a.Config.ProcArgs = nil
	}

	// Final Validation Checks (unchanged)
	if a.Config.AgentMode && a.Config.APIKey == "" && os.Getenv("GEMINI_API_KEY") == "" {
		return fmt.Errorf("must provide -apikey or set GEMINI_API_KEY environment variable for -agent mode")
	}

	return nil
}
