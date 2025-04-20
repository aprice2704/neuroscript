// filename: pkg/core/tool_file_api_sync.go
package core

import (
	"context"
	"encoding/hex" // Keep if walkAndCompareLocalFiles uses it
	"errors"
	"fmt"
	"io/fs" // Keep for walkAndCompareLocalFiles
	"log"   // Keep for walkAndCompareLocalFiles
	"path/filepath"
	"sync" // Keep for waitForWorkersAndProcessResults

	"github.com/google/generative-ai-go/genai"
	// gitignore "github.com/sabhiram/go-gitignore" // Moved to helpers?
)

// --- Main Orchestrator ---

// SyncDirectoryUpHelper synchronizes a local directory up to the File API.
// Returns a map of statistics and an error if critical issues occurred.
func SyncDirectoryUpHelper(
	ctx context.Context,
	absLocalDir string,
	filterPattern string,
	ignoreGitignore bool,
	client *genai.Client,
	infoLog, errorLog, debugLog *log.Logger,
) (map[string]interface{}, error) {

	// Assumes syncContext is defined (likely in helpers file now)
	syncCtx := &syncContext{
		ctx:            ctx,
		absLocalDir:    absLocalDir,
		filterPattern:  filterPattern,
		client:         client,
		infoLog:        infoLog,
		errorLog:       errorLog,
		debugLog:       debugLog,
		apiFilesMap:    make(map[string]*genai.File),
		localFilesSeen: make(map[string]bool),
	}

	var err error

	// 1. Initialize loggers and stats
	// Assumes initializeSyncState is defined in helpers
	syncCtx.stats, syncCtx.incrementStat = initializeSyncState(infoLog, errorLog, debugLog)
	syncCtx.infoLog.Println("[API HELPER Sync] Starting sync 'up' for directory:", syncCtx.absLocalDir)

	// 2. List existing API files
	// Assumes listExistingAPIFiles is defined in helpers
	syncCtx.apiFilesMap, err = listExistingAPIFiles(syncCtx)
	if err != nil {
		return syncCtx.stats, err
	}

	// 3. Initialize Gitignore
	// Assumes initializeGitignore is defined in helpers
	syncCtx.ignorer = initializeGitignore(syncCtx, ignoreGitignore)

	// 4. Start Upload Workers
	// Assumes uploadJob, uploadResult defined (likely in helpers)
	jobsChan := make(chan uploadJob, 100)
	resultsChan := make(chan uploadResult, 100)
	var uploadWg sync.WaitGroup
	// Assumes startUploadWorkers is defined in helpers
	// Pass the receive-only view of jobsChan, send-only view of resultsChan
	startUploadWorkers(syncCtx, &uploadWg, jobsChan, resultsChan) // Pass bidirectional channels

	// 5. Walk Local Directory & Queue Jobs
	// Assumes walkAndCompareLocalFiles is defined below or in helpers
	walkErr := walkAndCompareLocalFiles(syncCtx, jobsChan) // jobsChan needs send-only here

	// --- Synchronization Point 1: After Walk, Before Result Processing ---
	syncCtx.debugLog.Printf("[DEBUG API HELPER Sync] filepath.WalkDir finished. Error: %v", walkErr)
	syncCtx.debugLog.Println("[DEBUG API HELPER Sync] Closing jobsChan.")
	close(jobsChan)
	syncCtx.infoLog.Printf("[API HELPER Sync] Local directory walk finished (WalkErr: %v). Waiting for upload workers...", walkErr)

	// 6. Wait for Workers & Process Results
	// Assumes waitForWorkersAndProcessResults is defined below or in helpers
	// Pass the bidirectional resultsChan so it can be closed inside
	err = waitForWorkersAndProcessResults(syncCtx, &uploadWg, resultsChan)
	if err != nil {
		syncCtx.errorLog.Printf("[ERROR API HELPER Sync] Error during worker wait or result processing: %v", err)
	}

	// 7. Delete Stale API Files
	// Assumes deleteStaleAPIFiles is defined in helpers
	err = deleteStaleAPIFiles(syncCtx)
	if err != nil {
		syncCtx.errorLog.Printf("[ERROR API HELPER Sync] Error during stale file deletion: %v", err)
	}

	// 8. Finalize and Return
	syncCtx.stats["files_processed"] = syncCtx.stats["files_scanned"].(int64) - syncCtx.stats["files_ignored"].(int64) - syncCtx.stats["files_filtered"].(int64)
	syncCtx.infoLog.Printf("[API HELPER Sync] Sync finished. Final Stats: %+v", syncCtx.stats)

	if walkErr != nil {
		syncCtx.errorLog.Printf("[FINAL API HELPER Sync] Sync completed with critical walk error: %v", walkErr)
		return syncCtx.stats, fmt.Errorf("sync failed during directory walk: %w", walkErr)
	}
	if syncCtx.stats["upload_errors"].(int64) > 0 || syncCtx.stats["delete_errors"].(int64) > 0 || syncCtx.stats["list_api_errors"].(int64) > 0 || syncCtx.stats["hash_errors"].(int64) > 0 {
		syncCtx.errorLog.Println("[FINAL API HELPER Sync] Sync completed with non-critical errors reported in stats.")
	} else {
		syncCtx.infoLog.Println("[FINAL API HELPER Sync] Sync completed successfully.")
	}
	return syncCtx.stats, nil
}

// walkAndCompareLocalFiles walks the local directory, compares files with the API map, and sends jobs to workers.
// Kept here as it's closely tied to the main sync logic flow initiated by SyncDirectoryUpHelper.
// Takes jobsChan as send-only (chan<-)
func walkAndCompareLocalFiles(sc *syncContext, jobsChan chan<- uploadJob) error {
	sc.infoLog.Printf("[API HELPER Sync] Walking local directory: %s", sc.absLocalDir)
	sc.debugLog.Printf("[DEBUG API HELPER Sync] About to call filepath.WalkDir on: %s", sc.absLocalDir)

	walkErr := filepath.WalkDir(sc.absLocalDir, func(currentPath string, d fs.DirEntry, err error) error {
		// +++ DEBUG: Entry point (First line in callback) +++
		sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Visiting: %q (IsDir: %t, Err: %v)", currentPath, d.IsDir(), err)

		if err != nil {
			sc.incrementStat("walk_errors")
			sc.errorLog.Printf("[ERROR API HELPER Sync Walk] Access error visiting %q: %v", currentPath, err)
			if errors.Is(err, fs.ErrPermission) {
				if d.IsDir() {
					sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Skipping permission-denied directory: %q", currentPath)
					return filepath.SkipDir
				}
				sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Skipping permission-denied file: %q", currentPath)
				return nil
			}
			sc.errorLog.Printf("[ERROR API HELPER Sync Walk] Stopping walk due to non-permission error: %v", err)
			return err
		}
		if currentPath == sc.absLocalDir {
			sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Skipping root directory entry: %q", currentPath)
			return nil
		}

		relPath, relErr := filepath.Rel(sc.absLocalDir, currentPath)
		if relErr != nil {
			sc.incrementStat("walk_errors")
			sc.errorLog.Printf("[ERROR API HELPER Sync Walk] Failed to get relative path for %q (base %q): %v", currentPath, sc.absLocalDir, relErr)
			return nil
		}
		relPath = filepath.ToSlash(relPath)
		sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Relative path: %q", relPath)

		// Gitignore Check
		if sc.ignorer != nil && sc.ignorer.MatchesPath(relPath) {
			sc.incrementStat("files_ignored")
			if d.IsDir() {
				sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Gitignored dir, skipping subtree: %s", relPath)
				return filepath.SkipDir
			}
			sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Gitignored file: %s", relPath)
			return nil
		}
		if d.IsDir() {
			sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Skipping directory entry: %s", relPath)
			return nil
		} // Skip other dirs

		// Process File
		sc.stats["files_scanned"] = sc.stats["files_scanned"].(int64) + 1
		sc.localFilesSeen[relPath] = true // Mark file as seen locally

		// Filter Check
		if sc.filterPattern != "" {
			baseName := filepath.Base(currentPath)
			match, matchErr := filepath.Match(sc.filterPattern, baseName)
			if matchErr != nil {
				sc.incrementStat("walk_errors")
				sc.errorLog.Printf("[ERROR API HELPER Sync Walk] Invalid filter pattern %q: %v", sc.filterPattern, matchErr)
				return fmt.Errorf("invalid sync filter pattern: %w", matchErr)
			}
			if !match {
				sc.incrementStat("files_filtered")
				sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Filtered out by pattern %q: %s", sc.filterPattern, relPath)
				return nil
			}
			sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Passed filter check: %s", relPath)
		}

		// Calculate Hash
		sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Calculating SHA256 hash for: %s", relPath)
		// Assumes calculateFileHash is available in package core
		localHash, hashErr := calculateFileHash(currentPath)
		if hashErr != nil {
			sc.incrementStat("hash_errors")
			sc.errorLog.Printf("[ERROR API HELPER Sync Walk] Failed to hash file %s: %v", relPath, hashErr)
			return nil
		}
		// Assumes min is available in package core
		sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Hash OK for: %s (Hash: %s...)", relPath, localHash[:min(len(localHash), 8)])

		// Compare with API map & Queue Job if needed
		apiFileInfo, existsInAPI := sc.apiFilesMap[relPath]
		apiHashHex := ""
		if existsInAPI && apiFileInfo != nil {
			apiHashHex = hex.EncodeToString(apiFileInfo.Sha256Hash)
		}

		if !existsInAPI || apiFileInfo == nil || apiHashHex != localHash {
			jobType := "NEW"
			if existsInAPI && apiFileInfo != nil {
				jobType = "UPDATE"
			}
			sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Queuing %s upload job for: %s", jobType, relPath)
			// Assumes uploadJob is defined (likely in helpers)
			jobsChan <- uploadJob{ // Sending to send-only channel is allowed
				localAbsPath:    currentPath,
				relPath:         relPath,
				localHash:       localHash,
				existingApiFile: apiFileInfo,
			}
		} else {
			sc.incrementStat("files_up_to_date")
			sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] File is up-to-date: %s", relPath)
		}
		return nil
	}) // End WalkDir Callback
	return walkErr // Return the error from WalkDir itself
}

// waitForWorkersAndProcessResults waits for upload workers to finish and processes their results.
// It updates the apiFilesMap and stats based on the results.
// FIXED: resultsChan is now bidirectional (chan) so it can be closed.
func waitForWorkersAndProcessResults(sc *syncContext, wg *sync.WaitGroup, resultsChan chan uploadResult) error {

	waitDoneChan := make(chan struct{})
	go func() {
		sc.debugLog.Println("[DEBUG API HELPER Sync] Goroutine starting wg.Wait()...")
		wg.Wait()
		sc.debugLog.Println("[DEBUG API HELPER Sync] wg.Wait() finished.")
		sc.debugLog.Println("[DEBUG API HELPER Sync] Closing resultsChan.")
		// FIXED: Closing bidirectional channel is allowed.
		close(resultsChan)
		sc.debugLog.Println("[DEBUG API HELPER Sync] Goroutine finished closing resultsChan.")
		close(waitDoneChan) // Signal completion
	}()

	sc.debugLog.Println("[DEBUG API HELPER Sync] Starting loop to process results from resultsChan...")
	sc.infoLog.Println("[API HELPER Sync] Processing upload results...")
	processedCount := 0
	// FIXED: Ranging over resultsChan (bidirectional) is fine.
	for result := range resultsChan {
		processedCount++
		sc.debugLog.Printf("[DEBUG API HELPER Sync] Received result %d for: %s (Error: %v)", processedCount, result.job.relPath, result.err)
		if result.err != nil {
			sc.incrementStat("upload_errors")
			sc.errorLog.Printf("[ERROR API HELPER Sync] Upload/Update failed for %s: %v", result.job.relPath, result.err)
			delete(sc.apiFilesMap, result.job.relPath) // Remove from map on error
		} else if result.apiFile != nil {
			if result.job.existingApiFile == nil {
				sc.incrementStat("files_uploaded")
			} else {
				sc.incrementStat("files_updated_api")
			}
			sc.infoLog.Printf("[API HELPER Sync] Upload successful (%s): %s -> API Name: %s", result.job.relPath, result.job.relPath, result.apiFile.Name) // Simplified log
			sc.apiFilesMap[result.job.relPath] = result.apiFile                                                                                            // Update map with new/updated file info
		} else {
			sc.errorLog.Printf("[CRITICAL API HELPER Sync] Upload result for %s had nil error but also nil apiFile", result.job.relPath)
			sc.incrementStat("upload_errors")
			delete(sc.apiFilesMap, result.job.relPath)
		}
	}
	sc.debugLog.Printf("[DEBUG API HELPER Sync] Finished processing results loop (%d results).", processedCount)
	sc.infoLog.Printf("[API HELPER Sync] Finished processing %d upload results.", processedCount)

	// Wait for the goroutine above to finish closing the channel etc.
	sc.debugLog.Println("[DEBUG API HELPER Sync] Waiting for waitDoneChan signal...")
	<-waitDoneChan
	sc.debugLog.Println("[DEBUG API HELPER Sync] Received waitDoneChan signal.")
	return nil // Indicate success (or non-critical errors logged)
}

// Note: Ensure definitions for calculateFileHash, min, HelperListApiFiles, HelperUploadAndPollFile
// and the structs (syncContext, uploadJob, uploadResult) are accessible. If helpers were moved,
// ensure necessary structs/functions are either also moved or remain accessible in the original file.
