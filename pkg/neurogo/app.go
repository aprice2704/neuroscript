// filename: pkg/neurogo/app.go
package neurogo

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/generative-ai-go/genai"
)

// App encapsulates the application's state and configuration.
type App struct {
	Config   *Config // Holds configuration like Insecure, APIKey, ModelName etc.
	InfoLog  *log.Logger
	WarnLog  *log.Logger // Keep this field
	ErrorLog *log.Logger
	DebugLog *log.Logger
	LLMLog   *log.Logger
	Insecure *bool

	// Runtime state / clients
	llmClient *core.LLMClient // Unexported, managed internally
}

// NewApp creates a new App instance with default loggers.
func NewApp() *App { /* ... unchanged ... */
	return &App{
		Config:   NewConfig(),
		InfoLog:  log.New(os.Stdout, "INFO: ", log.LstdFlags),
		ErrorLog: log.New(os.Stderr, "ERROR: ", log.LstdFlags|log.Lshortfile),
		DebugLog: log.New(io.Discard, "DEBUG: ", log.LstdFlags|log.Lshortfile),
		LLMLog:   log.New(io.Discard, "DEBUG-LLM: ", log.LstdFlags|log.Lshortfile),
	}
}

// initLogging sets up logging based on Config.
func (a *App) initLogging() error { /* ... unchanged ... */
	// Debug Log
	if a.Config.DebugLogFile != "" {
		f, err := os.OpenFile(a.Config.DebugLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed open debug log %s: %w", a.Config.DebugLogFile, err)
		}
		a.DebugLog.SetOutput(f)
		a.DebugLog.Printf("--- Debug Logging Enabled to %s ---", a.Config.DebugLogFile)
	} else {
		a.DebugLog.SetOutput(io.Discard)
	}
	// LLM Debug Log
	if a.Config.LLMDebugLogFile != "" {
		f, err := os.OpenFile(a.Config.LLMDebugLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed open LLM debug log %s: %w", a.Config.LLMDebugLogFile, err)
		}
		a.LLMLog.SetOutput(f)
		a.LLMLog.Printf("--- LLM Debug Logging Enabled to %s ---", a.Config.LLMDebugLogFile)
		fmt.Println("--- LLM Debug Logging Enabled ---")
	} else {
		a.LLMLog.SetOutput(io.Discard)
	}
	return nil
}

// initLLMClient initializes the GenAI client if needed for the current mode.
func (a *App) initLLMClient(ctx context.Context) error {
	// Updated check to include CleanAPI
	needsLLM := a.Config.RunAgentMode || a.Config.ScriptFile != "" || a.Config.RunSyncMode || a.Config.CleanAPI
	if !needsLLM {
		a.DebugLog.Println("LLM Client not required.")
		return nil
	}
	if a.llmClient != nil && a.llmClient.Client() != nil {
		a.DebugLog.Println("LLM Client already initialized.")
		return nil
	}
	if a.Config.APIKey == "" {
		return errors.New("API key is missing")
	}

	a.InfoLog.Println("Initializing LLM Client...") // Simplified log
	debugLLMEnabled := a.Config.LLMDebugLogFile != ""
	client := core.NewLLMClient(a.Config.APIKey, a.Config.ModelName, a.LLMLog, debugLLMEnabled)
	if client == nil || client.Client() == nil {
		return errors.New("LLM client initialization failed")
	}

	a.llmClient = client
	a.InfoLog.Println("LLM Client initialized successfully.")
	return nil
}

// ParseFlags wraps the config's ParseFlags method.
func (a *App) ParseFlags(args []string) error {
	return a.Config.ParseFlags(args)
}

// Run executes the appropriate application mode based on parsed flags.
func (a *App) Run(ctx context.Context) error {
	if err := a.initLogging(); err != nil {
		log.Printf("ERROR: Logging init failed: %v\n", err)
		return err
	}
	if err := a.initLLMClient(ctx); err != nil {
		a.ErrorLog.Printf("LLM Client init failed: %v", err)
		// Updated check
		needsLLM := a.Config.RunAgentMode || a.Config.ScriptFile != "" || a.Config.RunSyncMode || a.Config.CleanAPI
		if needsLLM {
			return err
		} // Return error only if LLM was needed
		a.InfoLog.Println("Proceeding without LLM client.")
	}

	// Select Mode based on Config flags (updated check for CleanAPI)
	if a.Config.CleanAPI {
		a.InfoLog.Println("--- Running in Clean API Mode ---") // Updated log
		if a.llmClient == nil || a.llmClient.Client() == nil {
			err := errors.New("LLM Client required for clean-api mode")
			a.ErrorLog.Println(err.Error())
			return err
		}
		return a.runCleanAPIMode(ctx) // Call renamed function
	} else if a.Config.RunSyncMode { /* ... unchanged sync ... */
		a.InfoLog.Println("--- Running in Sync Mode ---")
		if a.llmClient == nil || a.llmClient.Client() == nil {
			err := errors.New("LLM Client required for sync mode")
			a.ErrorLog.Println(err.Error())
			return err
		}
		return a.runSyncMode(ctx)
	} else if a.Config.ScriptFile != "" { /* ... unchanged script ... */
		a.InfoLog.Println("--- Running in Script Mode ---")
		if a.llmClient == nil {
			err := errors.New("LLM Client wrapper required for script mode")
			a.ErrorLog.Println(err.Error())
			return err
		}
		return a.runScriptMode(ctx)
	} else if a.Config.RunAgentMode { /* ... unchanged agent ... */
		a.InfoLog.Println("--- Running in Agent Mode ---")
		if a.llmClient == nil || a.llmClient.Client() == nil {
			err := errors.New("LLM Client required for agent mode")
			a.ErrorLog.Println(err.Error())
			return err
		}
		return a.runAgentMode(ctx)
	} else {
		a.ErrorLog.Println("Error: No execution mode selected.")
		return errors.New("no execution mode specified")
	}
}

// --- Mode Execution Functions ---
// Stubs for modes assumed defined elsewhere
// func (a *App) runScriptMode(ctx context.Context) error { return errors.New("not implemented") }
// func (a *App) runAgentMode(ctx context.Context) error { return errors.New("not implemented") }
// func (a *App) runSyncMode(ctx context.Context) error { return errors.New("not implemented") }

// Renamed function runCleanAPIMode (was runNukeMode)
// runCleanAPIMode handles deleting all files from the File API.
func (a *App) runCleanAPIMode(ctx context.Context) error {
	// Updated log messages
	a.InfoLog.Println("Initiating Clean API operation.")

	client := a.llmClient.Client() // Assumed checked in Run

	a.InfoLog.Println("Listing all files from the API...")
	apiFiles, listErr := core.HelperListApiFiles(ctx, client, a.DebugLog)
	if listErr != nil {
		a.ErrorLog.Printf("Failed list API files: %v", listErr)
		return fmt.Errorf("failed list API files: %w", listErr)
	}

	fileCount := len(apiFiles)
	if fileCount == 0 {
		a.InfoLog.Println("No files found in API.")
		fmt.Println("No files found in the API.")
		return nil
	}
	a.InfoLog.Printf("Found %d files in API.", fileCount)
	fmt.Printf("Found %d files in the API.\n", fileCount)

	// Confirmation (Updated wording)
	fmt.Printf("\n!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
	fmt.Printf("!! WARNING: This will permanently delete ALL %d files from the API !!\n", fileCount)
	fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n\n")
	fmt.Print("Are you absolutely sure you want to proceed? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	confirmation, err := reader.ReadString('\n')
	if err != nil {
		a.ErrorLog.Printf("Failed read confirm: %v", err)
		return fmt.Errorf("failed read confirm: %w", err)
	}
	confirmation = strings.TrimSpace(strings.ToLower(confirmation))
	if confirmation != "yes" {
		a.InfoLog.Println("Clean API operation cancelled.")
		fmt.Println("Clean API operation cancelled.")
		return nil
	} // Updated message

	// Updated log message
	a.InfoLog.Println("User confirmed. Proceeding with API file deletion...")
	fmt.Println("Proceeding with deletion...")

	// Concurrent Deletion
	const maxConcurrentDeletes = 16 // Keep increased worker pool
	var deleteWg sync.WaitGroup
	deleteJobsChan := make(chan *genai.File, fileCount)
	deleteErrors := int64(0)
	errorChan := make(chan error, fileCount)

	// Renamed worker log prefix
	a.DebugLog.Printf("Starting %d API delete workers...", maxConcurrentDeletes)
	for i := 0; i < maxConcurrentDeletes; i++ {
		deleteWg.Add(1)
		go func(workerID int) {
			defer deleteWg.Done()
			// Updated worker log prefix
			a.DebugLog.Printf("API Delete Worker %d: Started.", workerID)
			for fileToDelete := range deleteJobsChan {
				if fileToDelete == nil || fileToDelete.Name == "" {
					continue
				}
				// Updated worker log prefix
				a.DebugLog.Printf("API Delete Worker %d: Deleting %s (%s)...", workerID, fileToDelete.Name, fileToDelete.DisplayName)
				delCtx, cancelDel := context.WithTimeout(context.Background(), 30*time.Second)
				deleteErr := client.DeleteFile(delCtx, fileToDelete.Name)
				cancelDel()
				if deleteErr != nil {
					// Updated worker log prefix
					a.ErrorLog.Printf("API Delete Worker %d: FAILED Delete %s (%s): %v", workerID, fileToDelete.Name, fileToDelete.DisplayName, deleteErr)
					errorChan <- fmt.Errorf("failed delete %s: %w", fileToDelete.Name, deleteErr)
				} else {
					// Updated worker log prefix
					a.DebugLog.Printf("API Delete Worker %d: Deleted %s (%s)", workerID, fileToDelete.Name, fileToDelete.DisplayName)
				}
				time.Sleep(50 * time.Millisecond)
			}
			// Updated worker log prefix
			a.DebugLog.Printf("API Delete Worker %d: Exiting.", workerID)
		}(i)
	}

	a.DebugLog.Println("Queueing delete jobs...")
	for _, file := range apiFiles {
		fileCopy := file
		deleteJobsChan <- fileCopy
	}
	close(deleteJobsChan)
	a.DebugLog.Println("All delete jobs queued.")
	// Updated log prefix
	a.DebugLog.Println("Waiting for API delete workers...")
	deleteWg.Wait()
	close(errorChan)
	a.DebugLog.Println("API delete workers finished.")

	errorMessages := []string{}
	for err := range errorChan {
		deleteErrors++
		errorMessages = append(errorMessages, err.Error())
	}
	deleteSuccess := int64(fileCount) - deleteErrors

	// Report Summary (Updated title)
	summaryTitle := "Clean API Summary"
	a.InfoLog.Println("--------------------")
	a.InfoLog.Println(summaryTitle)
	a.InfoLog.Printf("  Files Found:     %d", fileCount)
	a.InfoLog.Printf("  Files Deleted:   %d", deleteSuccess)
	a.InfoLog.Printf("  Deletion Errors: %d", deleteErrors)
	a.InfoLog.Println("--------------------")
	fmt.Println("--------------------")
	fmt.Println(summaryTitle)
	fmt.Printf("  Files Found:     %d\n", fileCount)
	fmt.Printf("  Files Deleted:   %d\n", deleteSuccess)
	fmt.Printf("  Deletion Errors: %d\n", deleteErrors)
	fmt.Println("--------------------")

	if deleteErrors > 0 {
		// Updated error message
		a.ErrorLog.Printf("Clean API operation completed with %d errors:", deleteErrors)
		for i := 0; i < min(len(errorMessages), 5); i++ {
			a.ErrorLog.Printf("  - %s", errorMessages[i])
		}
		return fmt.Errorf("clean API operation completed with %d errors", deleteErrors)
	}
	// Updated success message
	a.InfoLog.Println("Clean API operation completed successfully.")
	fmt.Println("Clean API operation completed successfully.")
	return nil
}

// Add min if needed
// func min(a, b int) int { if a < b { return a }; return b }
// Add the InitLoggingAndLLMClient method to the App struct in app.go if it doesn't exist
// Example (needs to be added in pkg/neurogo/app.go):

func (a *App) InitLoggingAndLLMClient(ctx context.Context) error {
	if err := a.initLogging(); err != nil {
		// Use standard log here as app loggers might not be fully set
		log.Printf("ERROR: Logging init failed: %v\\n", err)
		return err
	}
	// Always attempt LLM init here, let initLLMClient decide if key exists
	if err := a.initLLMClient(ctx); err != nil {
		// Log but don't necessarily return error unless LLM is mandatory for TUI
		a.ErrorLog.Printf("LLM Client init failed: %v", err)
		// return err // Uncomment if LLM must succeed for TUI mode
	}
	return nil
}
