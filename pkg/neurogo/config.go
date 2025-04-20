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

func (s *stringSlice) String() string         { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(value string) error { *s = append(*s, filepath.Clean(value)); return nil }

const defaultSandboxSubdir = "agent_sandbox"

// Config holds all application configuration.
type Config struct {
	// General
	LibPaths         []string // -lib
	DebugAST         bool     // -debug-ast
	DebugInterpreter bool     // -debug-interpreter
	DebugLLM         bool     // -debug-llm
	ModelName        string   // -model
	APIKey           string   // -apikey // Moved to General as multiple modes use it

	// Agent Mode
	AgentMode     bool     // -agent
	AllowlistFile string   // -allowlist
	DenylistFiles []string // -denylist
	SandboxDir    string   // -sandbox
	// +++ ADDED: Repeatable attach flag for agent +++
	InitialAttachments stringSlice // -attach

	// Script Mode Specific
	TargetArg string   // Positional arg 1
	ProcArgs  []string // Positional args 2+

	// Sync Mode Flags
	SyncMode            bool   // -sync
	SyncDir             string // -sync-dir
	SyncFilter          string // -sync-filter
	SyncIgnoreGitignore bool   // -sync-ignore-gitignore
}

// ParseFlags parses command-line arguments.
func (a *App) ParseFlags(args []string) error {
	if a == nil {
		return fmt.Errorf("internal error: ParseFlags called on nil App")
	}

	fs := flag.NewFlagSet("neurogo", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var libPathsFlag stringSlice
	var denyPathsFlag stringSlice
	// +++ ADDED: Variable for attach flag +++
	var attachPathsFlag stringSlice

	// General Flags
	fs.Var(&libPathsFlag, "lib", "Specify a library path (file or directory) (repeatable)")
	fs.BoolVar(&a.Config.DebugAST, "debug-ast", false, "Enable verbose AST logging")
	fs.BoolVar(&a.Config.DebugInterpreter, "debug-interpreter", false, "Enable verbose interpreter execution logging")
	fs.BoolVar(&a.Config.DebugLLM, "debug-llm", false, "Enable verbose LLM API interaction logging")
	fs.StringVar(&a.Config.ModelName, "model", "gemini-1.5-pro-latest", "Name of the generative model to use")
	fs.StringVar(&a.Config.APIKey, "apikey", "", "LLM API key (defaults to GEMINI_API_KEY env var)")

	// Agent Flags
	fs.BoolVar(&a.Config.AgentMode, "agent", false, "Run neurogo in LLM agent mode")
	fs.StringVar(&a.Config.AllowlistFile, "allowlist", "agent_allowlist.txt", "Path to agent tool allowlist file")
	fs.Var(&denyPathsFlag, "denylist", "Path to an additional agent tool denylist file (repeatable)")
	// Default sandbox logic
	exePath, errExe := os.Executable()
	defaultSandbox := defaultSandboxSubdir
	if errExe == nil {
		defaultSandbox = filepath.Join(filepath.Dir(exePath), defaultSandboxSubdir)
	} else {
		fmt.Fprintf(os.Stderr, "[Warning] Could not determine executable path for sandbox default: %v\n", errExe)
	}
	fs.StringVar(&a.Config.SandboxDir, "sandbox", defaultSandbox, "Root directory for agent filesystem sandboxing")
	// +++ ADDED: Define -attach flag for agent mode +++
	fs.Var(&attachPathsFlag, "attach", "Path to a local file (in sandbox) to upload and attach to the agent session context (repeatable)")

	// Sync Mode Flags
	fs.BoolVar(&a.Config.SyncMode, "sync", false, "Run in file sync mode (upload local to API)")
	fs.StringVar(&a.Config.SyncDir, "sync-dir", ".", "Local directory to synchronize with File API")
	fs.StringVar(&a.Config.SyncFilter, "sync-filter", "", "Optional glob pattern to filter files within sync-dir")
	fs.BoolVar(&a.Config.SyncIgnoreGitignore, "sync-ignore-gitignore", false, "Ignore .gitignore files within sync-dir")

	// Custom Usage function (consider updating to show -attach)
	fs.Usage = func() { /* ... existing usage ... */
		// Add -attach under Agent Mode Flags
		fmt.Fprintf(os.Stderr, "  -%s\t%s\n", "attach", "Path to file to upload for agent context (repeatable)")
	}

	// Parse
	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("error parsing flags: %w", err)
	}

	// Assign repeatable flags
	a.Config.LibPaths = libPathsFlag
	a.Config.DenylistFiles = denyPathsFlag
	// +++ ADDED: Assign -attach flag value +++
	a.Config.InitialAttachments = attachPathsFlag

	// Determine operation mode and handle positional args (logic unchanged)
	// ...
	modes := 0
	if a.Config.AgentMode {
		modes++
	}
	if a.Config.SyncMode {
		modes++
	}
	isScriptMode := !a.Config.AgentMode && !a.Config.SyncMode
	positionalArgs := fs.Args()
	if modes > 1 { /* error */
		return fmt.Errorf("cannot use -agent and -sync flags simultaneously")
	}
	if (a.Config.AgentMode || a.Config.SyncMode) && len(positionalArgs) > 0 { /* error */
		return fmt.Errorf("unexpected positional args in -agent or -sync mode: %v", positionalArgs)
	}
	if isScriptMode {
		if len(positionalArgs) == 0 {
			fs.Usage()
			return fmt.Errorf("missing <Procedure | ScriptFile.ns.txt> argument")
		}
		a.Config.TargetArg = positionalArgs[0]
		a.Config.ProcArgs = positionalArgs[1:]
	} else {
		a.Config.TargetArg = ""
		a.Config.ProcArgs = nil
	}
	// ---

	// Final Validation Checks
	if (a.Config.AgentMode || a.Config.SyncMode) && a.Config.APIKey == "" && os.Getenv("GEMINI_API_KEY") == "" {
		return fmt.Errorf("must provide -apikey or set GEMINI_API_KEY environment variable for -agent or -sync mode")
	}
	if len(a.Config.InitialAttachments) > 0 && !a.Config.AgentMode {
		return fmt.Errorf("-attach flag can only be used with -agent mode")
	}

	return nil
}
