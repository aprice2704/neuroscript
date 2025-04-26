// filename: pkg/core/tool_file_api_sync.go
package core

import (
	"context"
	"errors"
	"fmt"
	"os" // Added for direct output in progress printer
	"sync"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/google/generative-ai-go/genai"
	// Assumes sync_types.go, sync_morehelpers.go, sync_logic.go, sync_workers.go
	// and their necessary imports (like path/filepath, io, etc.) exist
	// within this package ('core').
)

// SyncDirectoryUpHelper orchestrates the directory synchronization process.
// It now gathers local/remote state first, computes actions, then executes.
func SyncDirectoryUpHelper(
	ctx context.Context,
	absLocalDir string,
	filterPattern string,
	ignoreGitignore bool,
	client *genai.Client,
	logger interfaces.Logger,
) (map[string]interface{}, error) {

	// 1. Initialize Context, Loggers, Stats
	// Create syncCtx first, loggers will be assigned next.
	// Ensure syncContext is defined (e.g., in sync_types.go)
	syncCtx := &syncContext{
		ctx:           ctx,
		absLocalDir:   absLocalDir,
		filterPattern: filterPattern,
		client:        client,
		// Loggers will be set below
	}

	// *** Use returned loggers from initializeSyncState ***
	var stats map[string]interface{}
	var incrementStat func(string)
	// Call initializeSyncState (assumed in sync_morehelpers.go) and get back the potentially defaulted loggers
	// Ensure initializeSyncState signature returns the loggers.
	stats, incrementStat, syncCtx.logger = initializeSyncState(logger)
	syncCtx.stats = stats
	syncCtx.incrementStat = incrementStat
	// *** End Fix ***

	// Now syncCtx loggers are guaranteed to be non-nil
	syncCtx.logger.Debug("[API HELPER Sync] Starting sync 'up' for directory:", syncCtx.absLocalDir)

	// --- Phase 1: Gather State ---
	// Ensure listExistingAPIFiles is defined (e.g., in sync_morehelpers.go)
	remoteFilesMap, listErr := listExistingAPIFiles(syncCtx)
	if listErr != nil {
		syncCtx.logger.Error("[ERROR API HELPER Sync] Failed to list initial API files: %v", listErr)
		// Return stats map even on error, as some might have been initialized
		return syncCtx.stats, listErr // Return critical list error
	}
	// Ensure initializeGitignore is defined (e.g., in sync_morehelpers.go)
	syncCtx.ignorer = initializeGitignore(syncCtx, ignoreGitignore) // Store ignorer in context
	// Ensure gatherLocalFiles is defined (e.g., in sync_logic.go)
	localFilesMap, walkErr := gatherLocalFiles(syncCtx)
	// Check if walkErr indicates a critical failure vs. just skipped files
	// For now, treat any error returned by gatherLocalFiles as critical for stopping sync
	if walkErr != nil {
		syncCtx.logger.Error("[ERROR API HELPER Sync] Critical error during local file scan: %v", walkErr)
		return syncCtx.stats, fmt.Errorf("local file scan failed: %w", walkErr)
	}
	syncCtx.logger.Info("[API HELPER Sync] Local scan complete, found %d files passing filters.", len(localFilesMap))

	// --- Phase 2: Compare and Plan ---
	// Ensure computeSyncActions and SyncActions are defined (e.g., in sync_logic.go and sync_types.go)
	actions := computeSyncActions(syncCtx, localFilesMap, remoteFilesMap)

	// --- Phase 3: Execute Actions ---
	totalPlannedUploadsUpdates := len(actions.FilesToUpload) + len(actions.FilesToUpdate)
	totalPlannedDeletes := len(actions.FilesToDelete)
	totalOps := totalPlannedUploadsUpdates + totalPlannedDeletes

	// Print Plan Summary
	if totalOps > 0 {
		// Print summary directly to stdout for user visibility
		fmt.Printf("Syncing: Uploads=%d Updates=%d Deletes=%d Total=%d\n",
			len(actions.FilesToUpload), len(actions.FilesToUpdate), totalPlannedDeletes, totalOps)
		// Start progress line only if uploads/updates exist
		if totalPlannedUploadsUpdates > 0 {
			fmt.Printf("Progress [Upd/Add]: ") // Indicate Upload/Update phase
		}
	} else {
		// Handle case where no operations are needed
		syncCtx.logger.Debug("[API HELPER Sync] No sync operations required.")
		syncCtx.stats["files_processed"] = syncCtx.stats["files_scanned"].(int64) // All scanned files processed (by doing nothing)
		syncCtx.logger.Info("[API HELPER Sync] Sync finished. Final Stats: %+v", syncCtx.stats)
		syncCtx.logger.Debug("[FINAL API HELPER Sync] Sync completed successfully (No operations needed).")
		return syncCtx.stats, nil // Success, nothing to do
	}

	// Execute Uploads/Updates
	var uploadWg sync.WaitGroup
	uploadErr := errors.New("no upload/update operations performed") // Default info if none scheduled
	if totalPlannedUploadsUpdates > 0 {
		// Ensure uploadResult is defined (e.g., in sync_types.go)
		resultsChan := make(chan uploadResult, totalPlannedUploadsUpdates)
		// Ensure startUploadWorkers is defined (e.g., in sync_workers.go)
		startUploadWorkers(syncCtx, &uploadWg, actions, resultsChan)

		// Wait for Uploads/Updates and Process Results (Prints progress chars)
		uploadErr = waitForUploadResultsAndPrintProgress(syncCtx, &uploadWg, resultsChan, totalPlannedUploadsUpdates)
		if uploadErr != nil {
			syncCtx.logger.Error("[ERROR API HELPER Sync] Error during upload/update phase: %v", uploadErr)
		}
		fmt.Println(" Done.") // Finish progress line for uploads/updates
	} else {
		syncCtx.logger.Debug("[API HELPER Sync] Skipping upload/update phase (0 operations).")
		uploadErr = nil // No error occurred if skipped
	}

	// Execute Deletions
	var deleteWg sync.WaitGroup
	// Ensure startDeleteWorkers is defined (e.g., in sync_workers.go)
	startDeleteWorkers(syncCtx, &deleteWg, actions.FilesToDelete)
	if totalPlannedDeletes > 0 {
		syncCtx.logger.Debug("[DEBUG API HELPER Sync] Waiting for deleteWg.Wait()...")
		deleteWg.Wait()
		syncCtx.logger.Debug("[DEBUG API HELPER Sync] deleteWg.Wait() finished.")
		syncCtx.logger.Debug("[API HELPER Sync] Deletion phase complete.")
	} else {
		syncCtx.logger.Debug("[API HELPER Sync] Skipping delete phase (0 operations).")
	}

	// --- Finalize ---
	// Update final stats based on plan and errors encountered
	syncCtx.stats["files_processed"] = syncCtx.stats["files_scanned"].(int64)
	// Update stats based on *planned* actions. Actual success is inferred from error counts/return values.
	syncCtx.stats["files_uploaded"] = int64(len(actions.FilesToUpload))
	syncCtx.stats["files_updated_api"] = int64(len(actions.FilesToUpdate))
	// Note: delete_errors and files_deleted_api counts are incremented within startDeleteWorkers/workers

	syncCtx.logger.Info("[API HELPER Sync] Sync finished. Final Stats: %+v", syncCtx.stats)

	// Determine overall success/failure
	finalError := walkErr // Start with potential critical walk error
	if finalError == nil {
		finalError = uploadErr // Prioritize upload errors if no walk error
	}
	if finalError == nil && syncCtx.stats["delete_errors"].(int64) > 0 {
		finalError = fmt.Errorf("sync completed with %d delete errors", syncCtx.stats["delete_errors"].(int64))
	}

	if finalError != nil {
		syncCtx.logger.Error("[FINAL API HELPER Sync] Sync completed with errors: %v", finalError)
	} else {
		syncCtx.logger.Debug("[FINAL API HELPER Sync] Sync completed successfully.")
	}

	// Return stats and the first critical error encountered (or nil if successful)
	return syncCtx.stats, finalError
}

// waitForUploadResultsAndPrintProgress waits for workers and prints progress.
// Prints '.' for success, 'E' for error to stdout. Manages line wrapping.
// Returns the first error encountered during upload/update processing.
func waitForUploadResultsAndPrintProgress(sc *syncContext, wg *sync.WaitGroup, resultsChan chan uploadResult, totalPlannedOps int) error {
	waitDoneChan := make(chan struct{})
	go func() {
		sc.logger.Debug("[DEBUG WaitGroup] Starting wg.Wait()...")
		wg.Wait()
		sc.logger.Debug("[DEBUG WaitGroup] wg.Wait() finished.")
		sc.logger.Debug("[DEBUG WaitGroup] Closing resultsChan.")
		close(resultsChan) // Close after all workers are done
		sc.logger.Debug("[DEBUG WaitGroup] Closed resultsChan.")
		close(waitDoneChan)
	}()

	sc.logger.Debug("[DEBUG Progress] Starting results processing loop...")
	processedCount := 0
	charsOnLine := 0
	const maxCharsPerLine = 80 // Characters per line for progress bar
	var firstError error

	for result := range resultsChan { // Process results and print progress chars
		processedCount++
		sc.logger.Debug("Progress] Result %d: %s (Err: %v)", processedCount, result.job.relPath, result.err)
		if result.err != nil {
			sc.incrementStat("upload_errors")
			// Log error details via ErrorLog
			sc.logger.Error("[ERROR API HELPER Sync] Fail: %s: %v", result.job.relPath, result.err)
			if firstError == nil {
				firstError = result.err // Store first error
			}
			fmt.Print("E") // Print 'E' for error to stdout progress
		} else {
			// Log detailed success only to debug log
			if result.apiFile != nil {
				sc.logger.Debug("API HELPER Sync] Upload successful: %s -> %s", result.job.relPath, result.apiFile.Name)
			}
			fmt.Print(".") // Print '.' for success to stdout progress
		}
		// Flush stdout buffer to ensure progress character is immediately visible
		os.Stdout.Sync()

		charsOnLine++
		if charsOnLine >= maxCharsPerLine {
			fmt.Println()                      // Wrap line
			fmt.Printf("Progress [Upd/Add]: ") // Start new progress line prefix
			charsOnLine = 0
		}
	} // End results processing loop

	sc.logger.Debug("Progress] Finished results loop (%d results).", processedCount)
	sc.logger.Info("[API HELPER Sync] Finished processing %d upload/update results.", processedCount) // Summary log
	sc.logger.Debug("[DEBUG Progress] Waiting for waitDoneChan...")
	<-waitDoneChan
	sc.logger.Debug("[DEBUG Progress] Received waitDoneChan.")
	return firstError // Return the first upload/update error encountered
}

// Note: This file assumes the following are defined in other files within the 'core' package:
// - Structs: syncContext, uploadResult, LocalFileInfo, SyncActions (likely in sync_types.go)
// - Helper funcs: initializeSyncState, listExistingAPIFiles, initializeGitignore (likely in sync_morehelpers.go)
// - Helper funcs: gatherLocalFiles, computeSyncActions (likely in sync_logic.go)
// - Helper funcs: startUploadWorkers, startDeleteWorkers (likely in sync_workers.go)
