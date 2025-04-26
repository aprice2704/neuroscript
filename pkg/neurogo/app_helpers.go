// filename: pkg/neurogo/app_helpers.go
package neurogo

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/generative-ai-go/genai"
)

// runCleanAPIMode handles deleting all files from the File API.
func (a *App) runCleanAPIMode(ctx context.Context) error {
	a.Logger.Info("Initiating Clean API operation.")

	// Use the correct exported interface method GetLLMClient()
	llmClient := a.GetLLMClient()
	if llmClient == nil || llmClient.Client() == nil {
		return errors.New("LLM Client unavailable for clean-api mode")
	}
	client := llmClient.Client() // Get the underlying client

	a.Logger.Info("Listing all files from the API...")
	// Pass the app's DebugLog (via interface method)
	apiFiles, listErr := core.HelperListApiFiles(ctx, client, a.GetLogger())
	if listErr != nil {
		a.Logger.Error("Failed list API files: %v", listErr)
		return fmt.Errorf("failed list API files: %w", listErr)
	}

	fileCount := len(apiFiles)
	if fileCount == 0 {
		a.Logger.Info("No files found in API.")
		fmt.Println("No files found in the API.")
		return nil
	}
	a.Logger.Info("Found %d files in API.", fileCount)
	fmt.Printf("Found %d files in the API.\n", fileCount)

	// Confirmation
	fmt.Printf("\n!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
	fmt.Printf("!! WARNING: This will permanently delete ALL %d files from the API !!\n", fileCount)
	fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n\n")
	fmt.Print("Are you absolutely sure you want to proceed? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	confirmation, err := reader.ReadString('\n')
	if err != nil {
		a.Logger.Error("Failed read confirm: %v", err)
		return fmt.Errorf("failed read confirm: %w", err)
	}
	confirmation = strings.TrimSpace(strings.ToLower(confirmation))
	if confirmation != "yes" {
		a.Logger.Info("Clean API operation cancelled.")
		fmt.Println("Clean API operation cancelled.")
		return nil
	}

	a.Logger.Info("User confirmed. Proceeding with API file deletion...")
	fmt.Println("Proceeding with deletion...")

	// Concurrent Deletion
	const maxConcurrentDeletes = 16
	var deleteWg sync.WaitGroup
	deleteJobsChan := make(chan *genai.File, fileCount)
	errorChan := make(chan error, fileCount) // Channel for collecting errors

	// Use interface method to get logger
	logger := a.GetLogger()

	logger.Debug("Starting %d API delete workers...", maxConcurrentDeletes)
	for i := 0; i < maxConcurrentDeletes; i++ {
		deleteWg.Add(1)
		go func(workerID int) {
			defer deleteWg.Done()
			logger.Debug("API Delete Worker %d: Started.", workerID)
			for fileToDelete := range deleteJobsChan {
				if fileToDelete == nil || fileToDelete.Name == "" {
					logger.Debug("API Delete Worker %d: Received nil/empty file, skipping.", workerID)
					continue
				}
				logger.Debug("API Delete Worker %d: Deleting %s (%s)...", workerID, fileToDelete.Name, fileToDelete.DisplayName)
				delCtx, cancelDel := context.WithTimeout(context.Background(), 30*time.Second)
				deleteErr := client.DeleteFile(delCtx, fileToDelete.Name)
				cancelDel()
				if deleteErr != nil {
					detailedErr := fmt.Errorf("worker %d failed delete %s (%s): %w", workerID, fileToDelete.Name, fileToDelete.DisplayName, deleteErr)
					logger.Error(detailedErr.Error()) // Log specific error
					errorChan <- detailedErr          // Send error to channel
				} else {
					logger.Debug("API Delete Worker %d: Deleted %s (%s)", workerID, fileToDelete.Name, fileToDelete.DisplayName)
				}
			}
			logger.Debug("API Delete Worker %d: Exiting.", workerID)
		}(i)
	}

	logger.Debug("Queueing delete jobs...")
	for _, file := range apiFiles {
		fileCopy := file
		deleteJobsChan <- fileCopy
	}
	close(deleteJobsChan)
	logger.Debug("All delete jobs queued.")

	logger.Debug("Waiting for API delete workers to finish...")
	deleteWg.Wait()
	close(errorChan) // Close error channel only after all workers are done
	logger.Debug("API delete workers finished.")

	// Collect errors
	errorMessages := []string{}
	for err := range errorChan {
		errorMessages = append(errorMessages, err.Error())
	}
	deleteErrors := int64(len(errorMessages))
	deleteSuccess := int64(fileCount) - deleteErrors

	// Report Summary
	summaryTitle := "Clean API Summary"
	a.Logger.Info("--------------------")
	a.Logger.Info(summaryTitle)
	a.Logger.Info("  Files Found:     %d", fileCount)
	a.Logger.Info("  Files Deleted:   %d", deleteSuccess)
	a.Logger.Info("  Deletion Errors: %d", deleteErrors)
	a.Logger.Info("--------------------")
	fmt.Println("--------------------")
	fmt.Println(summaryTitle)
	fmt.Printf("  Files Found:     %d\n", fileCount)
	fmt.Printf("  Files Deleted:   %d\n", deleteSuccess)
	fmt.Printf("  Deletion Errors: %d\n", deleteErrors)
	fmt.Println("--------------------")

	if deleteErrors > 0 {
		a.Logger.Error("Clean API operation completed with %d errors:", deleteErrors)
		maxLoggedErrors := 5
		for i := 0; i < min(len(errorMessages), maxLoggedErrors); i++ {
			a.Logger.Error("  - %s", errorMessages[i])
		}
		if len(errorMessages) > maxLoggedErrors {
			a.Logger.Error("  ... (logged %d of %d errors)", maxLoggedErrors, deleteErrors)
		}
		return fmt.Errorf("clean API operation completed with %d errors", deleteErrors)
	}

	a.Logger.Info("Clean API operation completed successfully.")
	fmt.Println("Clean API operation completed successfully.")
	return nil
}

// Local min function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// initLLMClient initializes the GenAI client if needed for the current mode.
func (a *App) initLLMClient(ctx context.Context) error {

	logger := a.Logger

	// Determine if LLM is needed based on finalized config flags
	// Ensure flags like RunTuiMode, EnableLLM exist on Config struct
	needsLLM := false
	if a.Config != nil {
		needsLLM = a.Config.RunTuiMode || a.Config.RunAgentMode || a.Config.RunSyncMode || a.Config.CleanAPI || (a.Config.ScriptFile != "" && a.Config.EnableLLM)
	}
	strictNeed := false
	if a.Config != nil {
		strictNeed = a.Config.RunTuiMode || a.Config.RunAgentMode || a.Config.RunSyncMode || a.Config.CleanAPI // Modes absolutely requiring LLM client
	}

	apiKeyPresent := false
	enableLLMFlag := true // Default to true if config is nil? Or handle in NewConfig
	if a.Config != nil {
		apiKeyPresent = a.Config.APIKey != ""
		enableLLMFlag = a.Config.EnableLLM // Assumes EnableLLM field exists
	}

	a.Logger.Debug("initLLMClient: NeedsLLM=%v, StrictNeed=%v, APIKeyPresent=%v, EnableLLMFlag=%v",
		needsLLM, strictNeed, apiKeyPresent, enableLLMFlag)

	if !needsLLM {
		logger.Debug("initLLMClient: LLM Client not required for current mode or explicitly disabled.")
		return nil
	}

	// Warn if LLM seems needed but is disabled by flag
	if !enableLLMFlag && (a.Config != nil && (a.Config.RunTuiMode || a.Config.RunAgentMode)) {
		logger.Warn("LLM Client is disabled via flag (-enable-llm=false), but potentially needed for the selected mode (TUI/Agent). Operations requiring LLM may fail.")
	}

	// Check if already initialized
	if a.llmClient != nil && a.llmClient.Client() != nil {
		logger.Debug("initLLMClient: LLM Client already initialized.")
		return nil
	}

	// Check for API Key specifically when needed
	if !apiKeyPresent {
		errMsg := "API key is missing"
		// Fail if strictly needed OR (needed AND enabled)
		if strictNeed || (needsLLM && enableLLMFlag) {
			errMsg = fmt.Sprintf("%s but required for this mode", errMsg)
			errLog := a.GetLogger()
			if errLog != nil {
				logger.Error(errMsg)
			}
			logger.Debug("initLLMClient: Failing - %s", errMsg)
			return errors.New(errMsg) // Return fatal error
		} else {
			logger.Debug("%s; LLM operations will fail if used.", errMsg)
			logger.Debug("initLLMClient: Continuing without LLM client - %s", errMsg)
			return nil // Not a fatal error for this path
		}
	}

	// Proceed with initialization
	logger.Info("Initializing LLM Client...")

	debugLLMEnabled := false
	modelName := ""
	apiKey := ""
	if a.Config != nil {
		debugLLMEnabled = a.Config.LLMDebugLogFile != ""
		modelName = a.Config.ModelName
		apiKey = a.Config.APIKey
	}

	// Ensure API key is definitely not empty before calling core
	if apiKey == "" {
		errMsg := "internal error: API key is empty despite passing initial check"
		errLog := a.GetLogger()
		if errLog != nil {
			logger.Error(errMsg)
		}
		logger.Debug(errMsg)
		return errors.New(errMsg)
	}

	client := core.NewLLMClient(apiKey, modelName, logger, debugLLMEnabled)

	if client == nil || client.Client() == nil {
		errMsg := "LLM client initialization failed (core.NewLLMClient returned nil or client.Client() is nil)"
		// This is generally fatal if we got this far (API key was present)
		errLog := a.GetLogger()
		if errLog != nil {
			logger.Error(errMsg)
		}
		logger.Debug("initLLMClient: Failing - %s", errMsg)
		return errors.New(errMsg) // Return fatal error
	}

	a.llmClient = client // Assign the successfully created client
	logger.Info("LLM Client initialized successfully (Model: %s)", a.GetModelName())
	logger.Info("initLLMClient: Success, a.llmClient is NOT nil: %v", a.llmClient != nil)
	return nil
}
