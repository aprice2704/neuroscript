package main

import (
	"context"
	"flag"
	"fmt"
	"io" // Required for io.Discard with logger
	"log"
	"os"
	"strings" // Required for confirmation prompt

	"github.com/aprice2704/neuroscript/pkg/nssync"
)

// stringSliceFlag is a custom flag type for repeatable string flags
type stringSliceFlag []string

func (i *stringSliceFlag) String() string {
	// Check for nil to prevent panic during flag processing if flag is not set
	if i == nil {
		return "[]"
	}
	return fmt.Sprintf("%v", *i)
}

func (i *stringSliceFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	// --- Flags ---
	dir := flag.String("dir", "", "Local directory to sync (required, must be subdirectory of CWD)")
	var includePatterns stringSliceFlag
	flag.Var(&includePatterns, "include", "Glob pattern for files to include (can be repeated, e.g., '**/*.go'). Defaults to all files ('**/*').")
	var excludePatterns stringSliceFlag
	flag.Var(&excludePatterns, "exclude", "Glob pattern for files or directories to exclude (can be repeated, e.g., '.git/**')")
	ignoreGitignore := flag.Bool("ignore-gitignore", false, "Ignore .gitignore file(s) in the sync directory")
	dryRun := flag.Bool("dry-run", false, "Show what would be done without making changes")
	clearRemote := flag.Bool("clear-remote", false, "Delete ALL files from the remote API storage (requires confirmation)")
	apiKey := flag.String("api-key", "", "Gemini API Key (defaults to GOOGLE_API_KEY env var)")
	verbose := flag.Bool("verbose", false, "Enable detailed logging output to stdout")
	// Add a flag to skip confirmation for clear-remote, useful for scripts
	forceClear := flag.Bool("force-clear", false, "Skip confirmation prompt when using -clear-remote (use with extreme caution!)")

	flag.Parse()

	// --- Validation ---
	if *dir == "" && !*clearRemote {
		// Dir is not required if only clearing remote, but is required for sync
		fmt.Fprintln(os.Stderr, "Error: -dir flag is required for sync operations.")
		flag.Usage()
		os.Exit(1)
	}
	// If clearing, dir is still needed for the path safety check within NewSyncer, unless we relax that constraint for clear-remote specifically.
	// Let's keep it simple: require -dir for now, even for clear-remote, to ensure NewSyncer can run its checks.
	if *dir == "" && *clearRemote {
		fmt.Fprintln(os.Stderr, "Error: -dir flag is currently required even for -clear-remote to establish context.")
		flag.Usage()
		os.Exit(1)
	}

	if *clearRemote && (len(includePatterns) > 0 || len(excludePatterns) > 0 || *ignoreGitignore) {
		// Allow -dir with clear-remote for safety check, but not filtering flags
		fmt.Fprintln(os.Stderr, "Error: Filtering flags (-include, -exclude, -ignore-gitignore) cannot be used with -clear-remote")
		flag.Usage()
		os.Exit(1)
	}
	if *forceClear && !*clearRemote {
		fmt.Fprintln(os.Stderr, "Warning: -force-clear flag has no effect without -clear-remote")
	}

	// --- Setup Logger ---
	logger := log.New(io.Discard, "[nssync-cli] ", log.LstdFlags) // Default discard
	if *verbose {
		// Set output to Stdout for verbose mode to distinguish from errors on Stderr
		logger = log.New(os.Stdout, "[nssync-cli] ", log.LstdFlags)
		logger.Println("Verbose logging enabled.")
	}

	// --- Get API Key ---
	key := *apiKey
	if key == "" {
		key = os.Getenv("GOOGLE_API_KEY")
		if key != "" && *verbose {
			logger.Println("Using API Key from GOOGLE_API_KEY environment variable.")
		}
	} else {
		if *verbose {
			logger.Println("Using API Key from -api-key flag.")
		}
	}

	if key == "" {
		fmt.Fprintln(os.Stderr, "Error: API Key not provided via -api-key flag or GOOGLE_API_KEY environment variable")
		os.Exit(1)
	}

	// --- Initialize Syncer ---
	// Note: NewSyncer performs the crucial path safety check on *dir
	config := nssync.SyncerConfig{
		LocalDir:        *dir,
		APIKey:          key,
		DryRun:          *dryRun,
		IgnoreGitignore: *ignoreGitignore,
	}

	// Pass the logger instance to the syncer library
	syncer, err := nssync.NewSyncer(config, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing syncer: %v\n", err)
		os.Exit(1)
	}
	logger.Println("Syncer initialized successfully.")

	ctx := context.Background() // Create a background context

	// --- Execute Action ---
	if *clearRemote {
		logger.Println("Executing Clear Remote operation.")
		fmt.Fprintln(os.Stdout, "WARNING: You requested to clear all remote files associated with the API key.") // Use Stdout for user messages

		// Confirmation prompt unless -force-clear is used
		if !*dryRun && !*forceClear {
			fmt.Print("Are you absolutely sure you want to delete all remote files? (yes/no): ")
			var confirmation string
			_, err := fmt.Scanln(&confirmation)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nError reading confirmation: %v\nAborted.\n", err)
				os.Exit(1)
			}
			if strings.ToLower(strings.TrimSpace(confirmation)) != "yes" {
				fmt.Fprintln(os.Stdout, "Aborted by user.")
				os.Exit(0)
			}
			fmt.Fprintln(os.Stdout, "Confirmation received.") // Provide feedback
		} else if *dryRun {
			fmt.Fprintln(os.Stdout, "Dry run mode enabled. No files will actually be deleted.")
		} else if *forceClear {
			fmt.Fprintln(os.Stdout, "Skipping confirmation prompt due to -force-clear flag.")
		}

		err = syncer.ClearRemote(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing remote files: %v\n", err)
			os.Exit(1) // Exit with error status
		}

		if *dryRun {
			fmt.Fprintln(os.Stdout, "Dry run: Clear Remote operation simulated successfully.")
		} else {
			fmt.Fprintln(os.Stdout, "Clear Remote operation completed successfully.")
		}

	} else {
		// --- Sync Operation ---
		logger.Println("Executing Sync operation.")

		// Find local files matching the criteria using the library function
		logger.Printf("Scanning directory '%s' for files to sync...", *dir)
		// Pass include/exclude patterns directly from flags
		localFiles, err := nssync.FindFiles(*dir, includePatterns, excludePatterns, logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding local files: %v\n", err)
			os.Exit(1) // Exit with error status
		}

		if len(localFiles) == 0 {
			// Use Stdout for user-facing warnings/info
			fmt.Fprintf(os.Stdout, "Warning: No local files found matching the specified criteria in '%s'.\n", *dir)
			// Decide if we should still proceed to potentially delete remote files.
			// The Sync function handles deleting remotes not present in the (empty) list.
			fmt.Fprintln(os.Stdout, "Proceeding with sync. This may delete remote files if they existed previously under this configuration.")
		} else {
			fmt.Fprintf(os.Stdout, "Found %d local file(s) matching criteria for sync.\n", len(localFiles))
			if *verbose {
				// Optionally log the files found in verbose mode
				logger.Println("Matching local files:")
				for _, f := range localFiles {
					logger.Printf("- %s\n", f)
				}
			}
		}

		// Call the Sync method with the discovered list of relative file paths
		err = syncer.Sync(ctx, localFiles)
		if err != nil {
			// Report sync errors to Stderr
			fmt.Fprintf(os.Stderr, "Error during sync operation: %v\n", err)
			os.Exit(1) // Exit with error status
		}

		// Report success to Stdout
		if *dryRun {
			fmt.Fprintln(os.Stdout, "Dry run: Sync operation simulated successfully.")
		} else {
			fmt.Fprintln(os.Stdout, "Sync operation completed successfully.")
		}
	}

	// Explicitly exit with success status
	os.Exit(0)
}
