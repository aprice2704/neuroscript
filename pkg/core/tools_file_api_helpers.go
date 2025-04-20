// filename: pkg/core/tools_file_api_helpers.go
package core

import (
	"context"
	"encoding/hex" // Added for processUploadJob error return example
	"fmt"
	"io"  // Added for initializeSyncState
	"log" // Added for initializeSyncState
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/generative-ai-go/genai"
	gitignore "github.com/sabhiram/go-gitignore"
)

// --- Helper Structs (Moved here for clarity if not defined elsewhere) ---

// syncContext holds shared state and configuration for the sync operation.
type syncContext struct {
	ctx            context.Context
	absLocalDir    string
	filterPattern  string
	client         *genai.Client
	infoLog        *log.Logger
	errorLog       *log.Logger
	debugLog       *log.Logger
	stats          map[string]interface{}
	incrementStat  func(string)
	ignorer        *gitignore.GitIgnore
	apiFilesMap    map[string]*genai.File // Map DisplayName -> API File info
	localFilesSeen map[string]bool        // Set of relative paths seen locally
}

// uploadJob defines the data needed for an upload worker.
type uploadJob struct {
	localAbsPath    string
	relPath         string
	localHash       string
	existingApiFile *genai.File // nil if it's a new upload
}

// uploadResult defines the result of an upload worker's job.
type uploadResult struct {
	job     uploadJob
	apiFile *genai.File // nil on error
	err     error
}

// --- Helper Functions ---

// initializeSyncState sets up default loggers and the statistics map.
func initializeSyncState(infoLog, errorLog, debugLog *log.Logger) (map[string]interface{}, func(string)) {
	if infoLog == nil {
		infoLog = log.New(io.Discard, "INFO: ", log.LstdFlags)
	}
	if errorLog == nil {
		errorLog = log.New(io.Discard, "ERROR: ", log.LstdFlags)
	}
	if debugLog == nil {
		debugLog = log.New(io.Discard, "DEBUG: ", log.LstdFlags|log.Lshortfile)
	}

	stats := map[string]interface{}{
		"files_scanned": int64(0), "files_ignored": int64(0), "files_filtered": int64(0),
		"files_uploaded": int64(0), "files_updated_api": int64(0), "files_deleted_api": int64(0),
		"files_up_to_date": int64(0), "upload_errors": int64(0), "delete_errors": int64(0),
		"list_api_errors": int64(0), "walk_errors": int64(0), "hash_errors": int64(0),
		"files_processed": int64(0), "files_deleted_locally": int64(0),
	}
	incrementStat := func(key string) {
		if v, ok := stats[key].(int64); ok {
			stats[key] = v + 1
		} else {
			errorLog.Printf("[CRITICAL API HELPER Sync] Invalid stat key type for %s", key)
		}
	}
	return stats, incrementStat
}

// listExistingAPIFiles fetches the list of files from the API and returns them as a map.
func listExistingAPIFiles(sc *syncContext) (map[string]*genai.File, error) {
	sc.infoLog.Println("[API HELPER Sync] Listing current API files...")
	apiFiles, listErr := HelperListApiFiles(sc.ctx, sc.client, sc.debugLog)
	if listErr != nil {
		sc.incrementStat("list_api_errors")
		sc.errorLog.Printf("[ERROR API HELPER Sync] Failed to list API files: %v", listErr)
		return nil, fmt.Errorf("failed to list API files: %w", listErr) // Return critical error
	}

	apiFilesMap := make(map[string]*genai.File)
	for _, file := range apiFiles {
		if file.DisplayName != "" {
			apiFilesMap[file.DisplayName] = file
			hashPrefix := ""
			if len(file.Sha256Hash) > 0 {
				// Ensure calculateFileHash and min are accessible if used here
				hashPrefix = hex.EncodeToString(file.Sha256Hash)[:min(len(hex.EncodeToString(file.Sha256Hash)), 8)]
			}
			sc.debugLog.Printf("[DEBUG API HELPER Sync] API File Found: Name=%s, DisplayName=%s, SHA=%s...", file.Name, file.DisplayName, hashPrefix)
		} else {
			sc.debugLog.Printf("[WARN API HELPER Sync] API File Found with empty DisplayName: Name=%s", file.Name)
		}
	}
	sc.infoLog.Printf("[API HELPER Sync] Found %d API files.", len(apiFilesMap))
	return apiFilesMap, nil
}

// initializeGitignore loads the .gitignore file if requested and available.
func initializeGitignore(sc *syncContext, ignoreGitignore bool) *gitignore.GitIgnore {
	if ignoreGitignore {
		sc.infoLog.Println("[API HELPER Sync] Ignoring .gitignore file as requested.")
		return nil
	}

	gitignorePath := filepath.Join(sc.absLocalDir, ".gitignore")
	sc.debugLog.Printf("[DEBUG API HELPER Sync] Attempting to load gitignore from: %s", gitignorePath)
	ignorer, gitignoreErr := gitignore.CompileIgnoreFile(gitignorePath)
	if gitignoreErr != nil {
		if os.IsNotExist(gitignoreErr) {
			sc.infoLog.Println("[API HELPER Sync] No .gitignore file found, proceeding without gitignore rules.")
		} else {
			sc.errorLog.Printf("[WARN API HELPER Sync] Error reading .gitignore file at %s: %v", gitignorePath, gitignoreErr)
		}
		return nil // Proceed without ignorer on error or not found
	}

	if ignorer != nil {
		sc.infoLog.Println("[API HELPER Sync] Using gitignore rules.")
	}
	return ignorer
}

// startUploadWorkers initializes and starts the pool of goroutines that handle file uploads/updates.
// FIXED: jobsChan is now receive-only (<-chan) as workers only receive from it.
func startUploadWorkers(sc *syncContext, wg *sync.WaitGroup, jobsChan <-chan uploadJob, resultsChan chan<- uploadResult) {
	const maxConcurrentUploads = 8 // Can be made configurable later
	sc.debugLog.Printf("[DEBUG API HELPER Sync] Starting %d upload workers...", maxConcurrentUploads)

	for i := 0; i < maxConcurrentUploads; i++ {
		wg.Add(1) // Increment counter for each worker goroutine start
		go func(workerID int) {
			defer wg.Done() // Decrement counter when goroutine exits
			sc.debugLog.Printf("[API HELPER Sync Worker %d] Started.", workerID)
			// FIXED: Ranging over jobsChan (receive-only) is now allowed.
			for job := range jobsChan {
				sc.debugLog.Printf("[API HELPER Sync Worker %d] Processing job for: %s", workerID, job.relPath)
				apiFile, uploadErr := processUploadJob(sc, job)                         // Extracted job processing logic
				resultsChan <- uploadResult{job: job, apiFile: apiFile, err: uploadErr} // resultsChan is send-only view here (correct)
				sc.debugLog.Printf("[API HELPER Sync Worker %d] Finished job for: %s (API File: %v, Err: %v)", workerID, job.relPath, apiFile != nil, uploadErr != nil)
			}
			sc.debugLog.Printf("[API HELPER Sync Worker %d] Exiting (jobsChan closed).", workerID)
		}(i)
	}
	sc.debugLog.Printf("[DEBUG API HELPER Sync] %d upload workers started.", maxConcurrentUploads)
}

// processUploadJob handles the logic for a single upload/update job within a worker goroutine.
func processUploadJob(sc *syncContext, job uploadJob) (*genai.File, error) {
	var uploadErr error
	var apiFile *genai.File

	// If updating, delete the old file first
	if job.existingApiFile != nil {
		sc.debugLog.Printf("[API HELPER Sync Worker] Deleting existing API file %s for update: %s", job.existingApiFile.Name, job.relPath)
		delCtx, cancelDel := context.WithTimeout(context.Background(), 30*time.Second) // Use background context for API calls within worker
		deleteErr := sc.client.DeleteFile(delCtx, job.existingApiFile.Name)
		cancelDel()
		if deleteErr != nil {
			sc.errorLog.Printf("[ERROR API HELPER Sync Worker] Pre-delete failed for API file %s (rel: %s): %v", job.existingApiFile.Name, job.relPath, deleteErr)
			// Log error but proceed with upload attempt
		} else {
			sc.debugLog.Printf("[API HELPER Sync Worker] Pre-delete OK for API file %s (rel: %s)", job.existingApiFile.Name, job.relPath)
		}
		time.Sleep(100 * time.Millisecond) // Brief pause after delete
	}

	// Perform the upload using the dedicated helper
	uploadCtx, cancelUpload := context.WithTimeout(context.Background(), 5*time.Minute) // 5 min timeout for upload+poll
	// Ensure HelperUploadAndPollFile is accessible
	apiFile, uploadErr = HelperUploadAndPollFile(uploadCtx, job.localAbsPath, job.relPath, sc.client, sc.debugLog)
	cancelUpload()

	return apiFile, uploadErr
}

// deleteStaleAPIFiles identifies and deletes files existing in the API but not locally.
func deleteStaleAPIFiles(sc *syncContext) error {
	sc.debugLog.Println("[DEBUG API HELPER Sync] Entering deletion phase.")
	sc.infoLog.Println("[API HELPER Sync] Checking for remote files to delete...")

	var deleteWg sync.WaitGroup
	const maxConcurrentDeletes = 4
	deleteJobsChan := make(chan *genai.File, len(sc.apiFilesMap)) // Buffer size can be adjusted

	// Start delete workers
	sc.debugLog.Printf("[DEBUG API HELPER Sync] Starting %d delete workers...", maxConcurrentDeletes)
	for i := 0; i < maxConcurrentDeletes; i++ {
		deleteWg.Add(1)
		go func(workerID int) {
			defer deleteWg.Done()
			sc.debugLog.Printf("[API HELPER Sync Delete Worker %d] Started.", workerID)
			for fileToDelete := range deleteJobsChan { // Ranging is fine here
				if fileToDelete == nil || fileToDelete.Name == "" {
					continue
				}
				displayName := fileToDelete.DisplayName
				if displayName == "" {
					displayName = fileToDelete.Name
				}

				sc.debugLog.Printf("[API HELPER Sync Delete Worker %d] Deleting API File: Name=%s, DisplayName=%s", workerID, fileToDelete.Name, displayName)
				delCtx, cancelDel := context.WithTimeout(context.Background(), 30*time.Second)
				// Ensure sc.client is accessible
				deleteErr := sc.client.DeleteFile(delCtx, fileToDelete.Name)
				cancelDel()

				if deleteErr != nil {
					sc.incrementStat("delete_errors")
					sc.errorLog.Printf("[ERROR API HELPER Sync Delete Worker %d] Failed to delete API file %s (%s): %v", workerID, fileToDelete.Name, displayName, deleteErr)
				} else {
					sc.incrementStat("files_deleted_api")
					sc.infoLog.Printf("[API HELPER Sync Delete Worker %d] Deleted API file %s (%s)", workerID, fileToDelete.Name, displayName)
				}
				time.Sleep(50 * time.Millisecond) // Small delay between deletes
			}
			sc.debugLog.Printf("[API HELPER Sync Delete Worker %d] Exiting (deleteJobsChan closed).", workerID)
		}(i)
	}

	// Identify and queue files for deletion
	filesToDeleteCount := 0
	for displayName, apiFileInfo := range sc.apiFilesMap {
		if _, foundLocally := sc.localFilesSeen[displayName]; !foundLocally {
			// Apply filter check before deleting
			if sc.filterPattern != "" {
				baseName := filepath.Base(displayName)
				// Ensure filepath.Match is accessible
				match, _ := filepath.Match(sc.filterPattern, baseName)
				if !match {
					sc.debugLog.Printf("[DEBUG API HELPER Sync] Skipping deletion of remote file %s (doesn't match filter %q)", displayName, sc.filterPattern)
					continue
				}
			}
			if apiFileInfo == nil || apiFileInfo.Name == "" {
				sc.errorLog.Printf("[WARN API HELPER Sync] Skipping deletion for %s due to nil/empty API info.", displayName)
				continue
			}
			filesToDeleteCount++
			sc.incrementStat("files_deleted_locally")
			sc.debugLog.Printf("[DEBUG API HELPER Sync] Queuing remote file for deletion: %s (API Name: %s)", displayName, apiFileInfo.Name)
			deleteJobsChan <- apiFileInfo // Sending is fine here
		}
	}
	close(deleteJobsChan) // Signal delete workers no more jobs
	sc.infoLog.Printf("[API HELPER Sync] Queued %d remote files for deletion. Waiting for delete workers...", filesToDeleteCount)

	// Wait for deletions to complete
	sc.debugLog.Println("[DEBUG API HELPER Sync] Waiting for deleteWg.Wait()...")
	deleteWg.Wait()
	sc.debugLog.Println("[DEBUG API HELPER Sync] deleteWg.Wait() finished.")
	sc.infoLog.Println("[API HELPER Sync] Delete workers finished.")
	sc.debugLog.Println("[DEBUG API HELPER Sync] Exited deletion phase.")
	return nil // Indicate success or non-critical errors logged
}

// Ensure utility functions like min, calculateFileHash, HelperListApiFiles, HelperUploadAndPollFile
// are accessible within this package (either defined here or in another file in pkg/core)
