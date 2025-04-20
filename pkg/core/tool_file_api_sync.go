package core

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/generative-ai-go/genai"
	gitignore "github.com/sabhiram/go-gitignore"
)

// +++ ADDED: Reusable Sync Directory Helper +++
// HelperSyncDirectoryUp synchronizes a local directory up to the File API.
// Returns a map of statistics and an error if critical issues occurred.
func SyncDirectoryUpHelper( // Renamed Helper for clarity
	ctx context.Context,
	absLocalDir string, // Assumes already validated, absolute path
	filterPattern string,
	ignoreGitignore bool,
	client *genai.Client,
	infoLog, errorLog, debugLog *log.Logger, // Use multiple loggers
) (map[string]interface{}, error) { // Return map[string]interface{} to match tool's expected output type easily

	if client == nil {
		return nil, errors.New("genai client is nil")
	}
	// Ensure loggers are non-nil
	if infoLog == nil {
		infoLog = log.New(io.Discard, "", 0)
	}
	if errorLog == nil {
		errorLog = log.New(io.Discard, "", 0)
	}
	if debugLog == nil {
		debugLog = log.New(io.Discard, "", 0)
	}

	infoLog.Println("[API HELPER Sync] Starting sync 'up' for directory:", absLocalDir)

	// Stats Counters
	stats := map[string]interface{}{ // Use interface{} for values to match tool return easily
		"files_scanned": int64(0), "files_ignored": int64(0), "files_uploaded": int64(0),
		"files_updated_api": int64(0), "files_deleted_api": int64(0), "files_up_to_date": int64(0),
		"upload_errors": int64(0), "delete_errors": int64(0), "list_api_errors": int64(0),
		"walk_errors": int64(0), "hash_errors": int64(0),
		// Added keys matching gensync for potential consistency
		"files_processed":       int64(0), // Scanned minus ignored dirs/files
		"files_deleted_locally": int64(0), // Count of files detected as deleted locally
	}
	incrementStat := func(key string) {
		if v, ok := stats[key].(int64); ok {
			stats[key] = v + 1
		}
	}

	// 1. List API Files
	infoLog.Println("[API HELPER Sync] Listing current API files...")
	apiFiles, listErr := HelperListApiFiles(ctx, client, debugLog) // Use DebugLog for listing details
	if listErr != nil {
		incrementStat("list_api_errors")
		errorLog.Printf("[ERROR API HELPER Sync] Failed to list API files: %v. Aborting sync.", listErr)
		return stats, fmt.Errorf("failed to list API files: %w", listErr)
	}
	// Create map: displayName -> *genai.File
	apiFilesMap := make(map[string]*genai.File)
	for _, file := range apiFiles {
		if file.DisplayName != "" {
			// Handle potential duplicates? Last one wins for now.
			if _, exists := apiFilesMap[file.DisplayName]; exists {
				debugLog.Printf("[DEBUG API HELPER Sync] Duplicate display name found in API list: %s (API Name: %s overwriting previous)", file.DisplayName, file.Name)
			}
			apiFilesMap[file.DisplayName] = file
		} else {
			infoLog.Printf("[WARN API HELPER Sync] API file %s has no display name, cannot sync.", file.Name)
		}
	}
	infoLog.Printf("[API HELPER Sync] Found %d API files with display names.", len(apiFilesMap))

	// 2. Initialize Gitignore
	var ignorer *gitignore.GitIgnore
	if !ignoreGitignore {
		gitignorePath := filepath.Join(absLocalDir, ".gitignore")
		ignorer, err := gitignore.CompileIgnoreFile(gitignorePath)
		if err != nil && !os.IsNotExist(err) {
			errorLog.Printf("[WARN API HELPER Sync] Could not compile .gitignore file %s: %v", gitignorePath, err)
		} else if ignorer != nil {
			infoLog.Println("[API HELPER Sync] Initialized gitignore rules.")
		}
	} else {
		infoLog.Println("[API HELPER Sync] Ignoring .gitignore file.")
	}

	// 3. Walk Local Directory & Compare
	infoLog.Printf("[API HELPER Sync] Walking local directory: %s", absLocalDir)
	localFilesSeen := make(map[string]bool) // track relative paths seen locally

	// --- Concurrency Setup ---
	const maxConcurrentUploads = 8 // Make configurable later?
	type uploadJob struct {
		localAbsPath, relPath, localHash string
		existingApiFile                  *genai.File
	}
	type uploadResult struct {
		job     uploadJob
		apiFile *genai.File
		err     error
	}
	jobsChan := make(chan uploadJob, 100) // Buffered channel
	resultsChan := make(chan uploadResult, 100)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < maxConcurrentUploads; i++ {
		go func(workerID int) {
			debugLog.Printf("[API HELPER Sync Worker %d] Started.", workerID)
			for job := range jobsChan {
				debugLog.Printf("[API HELPER Sync Worker %d] Processing job for: %s", workerID, job.relPath)
				var uploadErr error
				var apiFile *genai.File
				// Delete existing file first if needed (match gensync behavior)
				if job.existingApiFile != nil {
					debugLog.Printf("[API HELPER Sync Worker %d] Deleting existing API file %s for update of %s", workerID, job.existingApiFile.Name, job.relPath)
					delErr := client.DeleteFile(ctx, job.existingApiFile.Name)
					if delErr != nil {
						// Log but proceed with upload; API might clean up later
						errorLog.Printf("[ERROR API HELPER Sync Worker %d] Failed pre-update delete for API file %s: %v", workerID, job.existingApiFile.Name, delErr)
						// We don't have a specific stat for pre-delete errors, maybe add one? Count as upload_error for now?
					} else {
						debugLog.Printf("[API HELPER Sync Worker %d] Pre-update delete successful for API file %s", workerID, job.existingApiFile.Name)
					}
					// Small delay after delete?
					time.Sleep(50 * time.Millisecond)
				}
				// Call upload helper
				apiFile, uploadErr = HelperUploadAndPollFile(ctx, job.localAbsPath, job.relPath, client, debugLog) // Use debugLog for upload details
				resultsChan <- uploadResult{job: job, apiFile: apiFile, err: uploadErr}
				wg.Done()
				debugLog.Printf("[API HELPER Sync Worker %d] Finished job for: %s (Error: %v)", workerID, job.relPath, uploadErr)
			}
			debugLog.Printf("[API HELPER Sync Worker %d] Exiting.", workerID)
		}(i)
	}
	// --- End Concurrency Setup ---

	walkErr := filepath.WalkDir(absLocalDir, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			incrementStat("walk_errors")
			errorLog.Printf("[ERROR API HELPER Sync] Access error %q: %v", currentPath, err)
			return nil // Skip path on access error
		}
		if currentPath == absLocalDir {
			return nil
		} // Skip root

		relPath, relErr := filepath.Rel(absLocalDir, currentPath)
		if relErr != nil {
			incrementStat("walk_errors")
			errorLog.Printf("[ERROR API HELPER Sync] Cannot get relative path for %s: %v", currentPath, relErr)
			return nil
		}
		relPath = filepath.ToSlash(relPath) // Use consistent slashes

		// Check gitignore first
		if ignorer != nil && ignorer.MatchesPath(relPath) {
			if d.IsDir() {
				incrementStat("files_ignored")
				debugLog.Printf("[DEBUG API HELPER Sync] Ignoring directory: %s", relPath)
				return filepath.SkipDir
			}
			incrementStat("files_ignored")
			debugLog.Printf("[DEBUG API HELPER Sync] Ignoring file: %s", relPath)
			return nil
		}

		// Skip directories after gitignore check
		if d.IsDir() {
			return nil
		}

		// Should be a file now
		stats["files_processed"] = stats["files_processed"].(int64) + 1
		localFilesSeen[relPath] = true // Mark as seen locally

		// Apply filter pattern by basename
		if filterPattern != "" {
			match, matchErr := filepath.Match(filterPattern, filepath.Base(currentPath))
			if matchErr != nil {
				incrementStat("walk_errors")
				errorLog.Printf("[ERROR API HELPER Sync] Invalid pattern '%s': %v", filterPattern, matchErr)
				return fmt.Errorf("invalid filter: %w", matchErr)
			} // Abort on bad pattern
			if !match {
				incrementStat("files_filtered")
				debugLog.Printf("[DEBUG API HELPER Sync] Filtered out by pattern: %s", relPath)
				return nil
			}
		}

		// Calculate hash
		localHash, hashErr := calculateFileHash(currentPath)
		if hashErr != nil {
			incrementStat("hash_errors")
			errorLog.Printf("[ERROR API HELPER Sync] Hash failed %s: %v", relPath, hashErr)
			return nil
		} // Skip file on hash error

		// Compare with API map
		apiFileInfo, existsInAPI := apiFilesMap[relPath]
		if !existsInAPI || hex.EncodeToString(apiFileInfo.Sha256Hash) != localHash {
			// Needs upload/update - Create job
			debugLog.Printf("[API HELPER Sync] Queuing job for: %s (Exists: %t, Hash Match: %t)", relPath, existsInAPI, existsInAPI && hex.EncodeToString(apiFileInfo.Sha256Hash) == localHash)
			wg.Add(1)                                                                                                              // Increment WaitGroup counter before sending job
			jobsChan <- uploadJob{localAbsPath: currentPath, relPath: relPath, localHash: localHash, existingApiFile: apiFileInfo} // Pass existing info if found
		} else {
			incrementStat("files_up_to_date") // Hashes match
			debugLog.Printf("[DEBUG API HELPER Sync] File up-to-date: %s", relPath)
		}
		return nil
	}) // End WalkDir

	// Close jobs channel once walk is done
	close(jobsChan)
	infoLog.Printf("[API HELPER Sync] Local directory walk finished. Waiting for uploads/updates...")

	// Wait for workers in separate goroutine, then close results
	go func() {
		wg.Wait()
		close(resultsChan)
		debugLog.Println("[API HELPER Sync] All upload workers finished.")
	}()

	// Process results as they come in
	for result := range resultsChan {
		if result.err != nil {
			incrementStat("upload_errors")
			errorLog.Printf("[ERROR API HELPER Sync] Upload/Update failed for %s: %v", result.job.relPath, result.err)
		} else {
			if result.job.existingApiFile == nil {
				incrementStat("files_uploaded") // Count as new upload
				infoLog.Printf("[API HELPER Sync] Upload successful (New): %s -> %s", result.job.relPath, result.apiFile.Name)
			} else {
				incrementStat("files_updated_api") // Count as update
				infoLog.Printf("[API HELPER Sync] Upload successful (Update): %s -> %s", result.job.relPath, result.apiFile.Name)
			}
		}
	}
	infoLog.Println("[API HELPER Sync] Finished processing upload/update results.")

	// 4. Delete API files not found locally
	infoLog.Println("[API HELPER Sync] Checking for remotely deleted files...")
	deleteCount := 0
	for displayName, apiFileInfo := range apiFilesMap {
		if !localFilesSeen[displayName] {
			// Check filter pattern again before deleting
			if filterPattern != "" {
				match, _ := filepath.Match(filterPattern, filepath.Base(displayName))
				if !match {
					debugLog.Printf("[DEBUG API HELPER Sync] Skipping remote delete for filtered file: %s", displayName)
					continue
				}
			}

			if apiFileInfo.Name == "" {
				infoLog.Printf("[WARN API HELPER Sync] Cannot delete remote file %s, missing API name.", displayName)
				continue
			}

			incrementStat("files_deleted_locally") // Mark intent
			deleteCount++
			infoLog.Printf("[API HELPER Sync] Deleting API file %s (for local: %s)", apiFileInfo.Name, displayName)
			// Use background context for delete?
			deleteErr := client.DeleteFile(context.Background(), apiFileInfo.Name)
			if deleteErr != nil {
				incrementStat("delete_errors")
				errorLog.Printf("[ERROR API HELPER Sync] Delete failed for %s: %v", apiFileInfo.Name, deleteErr)
			} else {
				incrementStat("files_deleted_api") // Count actual success
				debugLog.Printf("[DEBUG API HELPER Sync] Delete successful for: %s", apiFileInfo.Name)
			}
			// Optional small delay between deletes
			time.Sleep(50 * time.Millisecond)
		}
	}
	if deleteCount > 0 {
		infoLog.Printf("[API HELPER Sync] Finished processing %d potential remote deletions.", deleteCount)
	} else {
		infoLog.Println("[API HELPER Sync] No remotely deleted files found matching criteria.")
	}

	// Add total files processed (scanned - ignored) to stats if useful
	stats["files_processed"] = stats["files_scanned"].(int64) - stats["files_ignored"].(int64)

	// Return final stats and the first critical error encountered during walk
	infoLog.Printf("[API HELPER Sync] Sync finished. Stats: %+v", stats)
	return stats, walkErr // Return walkErr if it was critical (e.g., bad pattern)
}
