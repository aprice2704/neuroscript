// filename: pkg/neurogo/app_helpers.go
package neurogo

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	// Keep for FileState potentially
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// runCleanAPIMode deletes all files from the configured File API service.
func (a *App) runCleanAPIMode(ctx context.Context) error {
	a.Log.Info("--- Running in Clean API Mode ---")
	if !a.Config.CleanAPI {
		return fmt.Errorf("internal error: runCleanAPIMode called but CleanAPI flag is not set")
	}

	// User Confirmation
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("!! WARNING: This will permanently delete ALL files from  !!")
	fmt.Printf("!! the Google AI File API associated with this API key. !!\n")
	fmt.Println("!! There is NO UNDO.                                     !!")
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Print("Type 'yes' to confirm deletion: ")
	reader := bufio.NewReader(os.Stdin)
	confirmation, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(confirmation)) != "yes" {
		a.Log.Info("Clean API operation cancelled by user.")
		fmt.Println("Operation cancelled.")
		return nil
	}
	a.Log.Warn("Proceeding with Clean API operation.")

	// Get Clients
	llmClient := a.GetLLMClient()
	if llmClient == nil {
		return fmt.Errorf("cannot clean API: LLM client is nil")
	}
	genaiClient := llmClient.Client()
	if genaiClient == nil {
		return fmt.Errorf("cannot clean API: underlying genai client is nil")
	}
	logger := a.GetLogger()
	if logger == nil {
		// Fallback if logger somehow nil despite earlier checks
		fmt.Println("Error: Logger not available for Clean API.")
		return fmt.Errorf("cannot clean API: logger is nil")
	}

	// List Files
	logger.Info("Listing all files from File API for deletion...")
	// TODO: Find or implement HelperListApiFiles in pkg/core or use genaiClient directly
	// apiFiles, listErr := core.HelperListApiFiles(ctx, genaiClient, logger)
	apiFiles := []*ApiFileInfo{} // Use placeholder type defined in app_interface.go
	listErr := fmt.Errorf("core.HelperListApiFiles is undefined - Clean API needs implementation")

	if listErr != nil {
		logger.Error("Failed to list API files for deletion.", "error", listErr)
		fmt.Println("Error listing files from API. Aborting.")
		return fmt.Errorf("failed to list API files: %w", listErr)
	}

	if len(apiFiles) == 0 {
		logger.Info("No files found in File API to delete.")
		fmt.Println("No files found to delete.")
		return nil
	}

	logger.Info("Found files to delete.", "count", len(apiFiles))
	fmt.Printf("Found %d files. Starting deletion...\n", len(apiFiles))

	// Delete Files Concurrently
	var deleteWg sync.WaitGroup
	errorChan := make(chan error, len(apiFiles))
	// Use placeholder type ApiFileInfo
	deleteJobsChan := make(chan *ApiFileInfo, len(apiFiles))

	numWorkers := 10 // Adjust as needed
	for i := 0; i < numWorkers; i++ {
		deleteWg.Add(1)
		go func(workerID int) {
			defer deleteWg.Done()
			logger.Debug("API Delete Worker started.", "worker_id", workerID)
			for fileToDelete := range deleteJobsChan {
				if fileToDelete == nil || fileToDelete.Name == "" {
					logger.Debug("API Delete Worker received nil/empty file, skipping.", "worker_id", workerID)
					continue
				}
				logger.Debug("API Delete Worker deleting file.", "worker_id", workerID, "file_name", fileToDelete.Name, "display_name", fileToDelete.DisplayName)
				delCtx, cancelDel := context.WithTimeout(ctx, 30*time.Second)
				// Use genaiClient directly for deletion
				deleteErr := genaiClient.DeleteFile(delCtx, fileToDelete.Name)
				cancelDel()
				if deleteErr != nil {
					detailedErr := fmt.Errorf("worker %d failed delete %s (%s): %w", workerID, fileToDelete.Name, fileToDelete.DisplayName, deleteErr)
					logger.Error("API file deletion failed.", "worker_id", workerID, "file_name", fileToDelete.Name, "error", detailedErr)
					errorChan <- detailedErr
				} else {
					logger.Debug("API Delete Worker deleted file.", "worker_id", workerID, "file_name", fileToDelete.Name)
				}
			}
			logger.Debug("API Delete Worker exiting.", "worker_id", workerID)
		}(i)
	}

	logger.Debug("Sending delete jobs to workers.")
	for _, file := range apiFiles {
		if file != nil {
			deleteJobsChan <- file
		}
	}
	close(deleteJobsChan)
	logger.Debug("All delete jobs sent.")

	logger.Debug("Waiting for delete workers to complete...")
	deleteWg.Wait()
	close(errorChan)
	logger.Debug("Delete workers finished.")

	// Collect and Report Errors
	var deleteErrors []error
	for err := range errorChan {
		deleteErrors = append(deleteErrors, err)
	}

	if len(deleteErrors) > 0 {
		logger.Error("Encountered errors during File API cleanup.", "error_count", len(deleteErrors))
		fmt.Printf("Finished with %d errors:\n", len(deleteErrors))
		for i, err := range deleteErrors {
			fmt.Printf("  %d: %v\n", i+1, err)
			if i < 5 {
				logger.Error("Cleanup error detail", "index", i+1, "error", err)
			}
		}
		return fmt.Errorf("encountered %d errors during API file cleanup", len(deleteErrors))
	}

	logger.Info("Successfully deleted all files from File API.", "count", len(apiFiles))
	fmt.Printf("Successfully deleted %d files.\n", len(apiFiles))
	logger.Info("--- Clean API Mode Finished ---")
	return nil
}
