// filename: pkg/core/tool_file_api_sync.go
package core

import (
	"context"
	"errors"
	"fmt"
	"os" // Added for direct output in progress printer
	"sync"
	// Assumes sync_types.go, sync_morehelpers.go, sync_logic.go, sync_workers.go
	// and their necessary imports (like path/filepath, io, etc.) exist
	// within this package ('core').
)

// SyncDirectoryUpHelper orchestrates the directory synchronization process.
// *** MODIFIED: Removed check involving 'direction' ***
func SyncDirectoryUpHelper(
	ctx context.Context,
	absLocalDir string,
	filterPattern string,
	ignoreGitignore bool,
	interp *Interpreter, // Pass Interpreter
) (map[string]interface{}, error) {

	// --- Get Logger and Client from Interpreter ---
	logger := interp.Logger()
	client := interp.GenAIClient()
	// *** REMOVED check involving 'direction' ***
	if client == nil {
		// Sync 'up' requires a client.
		logger.Error("[API HELPER Sync] Sync 'up' requires a valid GenAI Client, but it's nil.")
		return nil, errors.New("sync 'up' requires a configured GenAI client")
	}
	// --- End Get Logger/Client ---

	// 1. Initialize Context, Stats, and Loggers
	stats, incrementStat, effectiveLogger := initializeSyncState(logger)

	syncCtx := &syncContext{
		ctx:           ctx,
		absLocalDir:   absLocalDir,
		filterPattern: filterPattern,
		client:        client,
		logger:        effectiveLogger,
		stats:         stats,
		incrementStat: incrementStat,
		interp:        interp,
	}

	syncCtx.logger.Debug("[API HELPER Sync] Starting sync 'up' for directory:", syncCtx.absLocalDir)

	// --- Phase 1: Gather State ---
	remoteFilesMap, listErr := listExistingAPIFiles(syncCtx)
	if listErr != nil {
		syncCtx.logger.Error("[ERROR API HELPER Sync] Failed to list initial API files: %v", listErr)
		return syncCtx.stats, listErr
	}
	syncCtx.ignorer = initializeGitignore(syncCtx, ignoreGitignore)
	localFilesMap, walkErr := gatherLocalFiles(syncCtx)
	if walkErr != nil {
		syncCtx.logger.Error("[ERROR API HELPER Sync] Critical error during local file scan: %v", walkErr)
		return syncCtx.stats, fmt.Errorf("local file scan failed: %w", walkErr)
	}
	syncCtx.logger.Debug("[API HELPER Sync] Local scan complete, found %d files passing filters.", len(localFilesMap))

	// --- Phase 2: Compare and Plan ---
	actions := computeSyncActions(syncCtx, localFilesMap, remoteFilesMap)

	// --- Phase 3: Execute Actions ---
	totalPlannedUploadsUpdates := len(actions.FilesToUpload) + len(actions.FilesToUpdate)
	totalPlannedDeletes := len(actions.FilesToDelete)
	totalOps := totalPlannedUploadsUpdates + totalPlannedDeletes

	// Print Plan Summary
	if totalOps > 0 {
		fmt.Printf("Syncing: Uploads=%d Updates=%d Deletes=%d Total=%d\n",
			len(actions.FilesToUpload), len(actions.FilesToUpdate), totalPlannedDeletes, totalOps)
		if totalPlannedUploadsUpdates > 0 {
			fmt.Printf("Progress [Upd/Add]: ")
		}
	} else {
		syncCtx.logger.Debug("[API HELPER Sync] No sync operations required.")
		scannedCount, ok := syncCtx.stats["files_scanned"].(int64)
		if !ok {
			scannedCount = 0
		}
		syncCtx.stats["files_processed"] = scannedCount
		syncCtx.logger.Debug("[API HELPER Sync] Sync finished. Final Stats: %+v", syncCtx.stats)
		syncCtx.logger.Debug("[FINAL API HELPER Sync] Sync completed successfully (No operations needed).")
		return syncCtx.stats, nil
	}

	// Execute Uploads/Updates
	var uploadWg sync.WaitGroup
	uploadErr := errors.New("no upload/update operations performed")
	if totalPlannedUploadsUpdates > 0 {
		resultsChan := make(chan uploadResult, totalPlannedUploadsUpdates)
		startUploadWorkers(syncCtx, &uploadWg, actions, resultsChan)
		uploadErr = waitForUploadResultsAndPrintProgress(syncCtx, &uploadWg, resultsChan, totalPlannedUploadsUpdates)
		if uploadErr != nil {
			syncCtx.logger.Error("[ERROR API HELPER Sync] Error during upload/update phase: %v", uploadErr)
		}
		fmt.Println(" Done.")
	} else {
		syncCtx.logger.Debug("[API HELPER Sync] Skipping upload/update phase (0 operations).")
		uploadErr = nil
	}

	// Execute Deletions
	var deleteWg sync.WaitGroup
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
	scannedCountFinal, ok := syncCtx.stats["files_scanned"].(int64)
	if !ok {
		scannedCountFinal = 0
	}
	syncCtx.stats["files_processed"] = scannedCountFinal
	syncCtx.stats["files_uploaded"] = int64(len(actions.FilesToUpload))
	syncCtx.stats["files_updated_api"] = int64(len(actions.FilesToUpdate))

	syncCtx.logger.Debug("[API HELPER Sync] Sync finished. Final Stats: %+v", syncCtx.stats)

	// Determine overall success/failure
	finalError := walkErr
	if finalError == nil {
		finalError = uploadErr
	}
	deleteErrorsCount, ok := syncCtx.stats["delete_errors"].(int64)
	if !ok {
		deleteErrorsCount = 0
	}

	if finalError == nil && deleteErrorsCount > 0 {
		finalError = fmt.Errorf("sync completed with %d delete errors", deleteErrorsCount)
	}

	if finalError != nil {
		syncCtx.logger.Error("[FINAL API HELPER Sync] Sync completed with errors: %v", finalError)
	} else {
		syncCtx.logger.Debug("[FINAL API HELPER Sync] Sync completed successfully.")
	}

	return syncCtx.stats, finalError
}

// waitForUploadResultsAndPrintProgress (Implementation unchanged)
func waitForUploadResultsAndPrintProgress(sc *syncContext, wg *sync.WaitGroup, resultsChan chan uploadResult, totalPlannedOps int) error {
	waitDoneChan := make(chan struct{})
	go func() {
		sc.logger.Debug("[DEBUG WaitGroup] Starting wg.Wait()...")
		wg.Wait()
		sc.logger.Debug("[DEBUG WaitGroup] wg.Wait() finished.")
		sc.logger.Debug("[DEBUG WaitGroup] Closing resultsChan.")
		close(resultsChan)
		sc.logger.Debug("[DEBUG WaitGroup] Closed resultsChan.")
		close(waitDoneChan)
	}()

	sc.logger.Debug("[DEBUG Progress] Starting results processing loop...")
	processedCount := 0
	charsOnLine := 0
	const maxCharsPerLine = 80
	var firstError error

	for result := range resultsChan {
		processedCount++
		sc.logger.Debug("Progress] Result %d: %s (Err: %v)", processedCount, result.job.relPath, result.err)
		if result.err != nil {
			sc.incrementStat("upload_errors")
			sc.logger.Error("[ERROR API HELPER Sync] Fail: %s: %v", result.job.relPath, result.err)
			if firstError == nil {
				firstError = result.err
			}
			fmt.Print("E")
		} else {
			if result.apiFile != nil {
				sc.logger.Debug("API HELPER Sync] Upload successful: %s -> %s", result.job.relPath, result.apiFile.Name)
			}
			fmt.Print(".")
		}
		os.Stdout.Sync()
		charsOnLine++
		if charsOnLine >= maxCharsPerLine {
			fmt.Println()
			fmt.Printf("Progress [Upd/Add]: ")
			charsOnLine = 0
		}
	}

	sc.logger.Debug("Progress] Finished results loop (%d results).", processedCount)
	sc.logger.Debug("[API HELPER Sync] Finished processing %d upload/update results.", processedCount)
	sc.logger.Debug("[DEBUG Progress] Waiting for waitDoneChan...")
	<-waitDoneChan
	sc.logger.Debug("[DEBUG Progress] Received waitDoneChan.")
	return firstError
}
