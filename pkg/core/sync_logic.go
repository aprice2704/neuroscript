// filename: pkg/core/sync_logic.go
package core

import (
	// Still needed for logging the local hash prefix if desired
	"errors"
	"fmt"
	"io/fs"
	"log" // Added for nil logger check fallback
	"path/filepath"

	// Added for string comparison below
	"github.com/google/generative-ai-go/genai"
	// Assumes sync_types.go, and other necessary imports like gitignore exist
)

// gatherLocalFiles walks the local directory and collects info about files that pass filters.
// Requires access to syncContext definition (from sync_types.go) and helpers.
func gatherLocalFiles(sc *syncContext) (map[string]LocalFileInfo, error) {
	// Ensure loggers are valid
	if sc.infoLog == nil || sc.errorLog == nil || sc.debugLog == nil {
		log.Println("ERROR: gatherLocalFiles called with nil loggers in syncContext!")
		// Return error as we cannot safely proceed without logging/stats
		return nil, errors.New("internal error: loggers not initialized in sync context")
	}

	sc.infoLog.Printf("[API HELPER Sync] Scanning local directory: %s", sc.absLocalDir)
	localFiles := make(map[string]LocalFileInfo)
	var firstWalkError error

	walkErr := filepath.WalkDir(sc.absLocalDir, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			sc.errorLog.Printf("[ERROR API HELPER Sync Walk] Initial error for %q: %v", currentPath, err)
			sc.incrementStat("walk_errors")
			if firstWalkError == nil {
				firstWalkError = err
			}
			if errors.Is(err, fs.ErrPermission) {
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil // Skip file
			}
			return err // Stop walk on other errors
		}
		sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Visiting: %q (IsDir: %t)", currentPath, d.IsDir())
		if currentPath == sc.absLocalDir {
			return nil
		} // Skip root

		relPath, relErr := filepath.Rel(sc.absLocalDir, currentPath)
		if relErr != nil {
			sc.errorLog.Printf("[ERROR API HELPER Sync Walk] Failed RelPath %q: %v", currentPath, relErr)
			sc.incrementStat("walk_errors")
			return nil // Skip entry
		}
		relPath = filepath.ToSlash(relPath)

		// Gitignore Check
		// Assumes sc.ignorer was initialized correctly earlier
		if sc.ignorer != nil && sc.ignorer.MatchesPath(relPath) {
			sc.incrementStat("files_ignored")
			if d.IsDir() {
				sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Gitignored dir: %s", relPath)
				return filepath.SkipDir
			}
			sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Gitignored file: %s", relPath)
			return nil
		}
		if d.IsDir() {
			return nil
		} // Skip other directories

		// Filter Check
		if sc.filterPattern != "" {
			baseName := filepath.Base(currentPath)
			match, matchErr := filepath.Match(sc.filterPattern, baseName)
			if matchErr != nil {
				sc.errorLog.Printf("[ERROR API HELPER Sync Walk] Invalid filter %q: %v", sc.filterPattern, matchErr)
				sc.incrementStat("walk_errors")
				return fmt.Errorf("invalid filter pattern: %w", matchErr) // Stop on bad filter
			}
			if !match {
				sc.incrementStat("files_filtered")
				sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Filtered out file: %s", relPath)
				return nil
			}
		}

		// Passed checks, process the file
		sc.incrementStat("files_scanned")
		sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Processing file: %s", relPath)

		// Calculate Hash
		// Assumes calculateFileHash is accessible (e.g., from tools_file_api.go)
		localHash, hashErr := calculateFileHash(currentPath)
		if hashErr != nil {
			sc.errorLog.Printf("[ERROR API HELPER Sync Walk] Hash failed %s: %v", relPath, hashErr)
			sc.incrementStat("hash_errors")
			return nil // Skip file on hash error
		}

		// Store file info
		// Assumes LocalFileInfo struct is defined (e.g., in sync_types.go)
		localFiles[relPath] = LocalFileInfo{
			RelPath: relPath,
			AbsPath: currentPath,
			Hash:    localHash,
		}
		// Assumes min is accessible (e.g. from helpers.go)
		sc.debugLog.Printf("[DEBUG API HELPER Sync Walk] Stored local info for: %s (Hash: %s...)", relPath, localHash[:min(len(localHash), 8)])
		return nil
	}) // End WalkDir Callback

	sc.infoLog.Printf("[API HELPER Sync] Local scan completed. Found %d candidate files. WalkErr: %v", len(localFiles), walkErr)
	// Return combined error
	if walkErr != nil {
		return localFiles, walkErr
	}
	return localFiles, firstWalkError
}

// computeSyncActions compares local and remote file lists and determines necessary actions.
// Assumes syncContext, LocalFileInfo, SyncActions, uploadJob defined (sync_types.go).
// FIXED: Hash comparison treats API bytes as ASCII hex representation.
func computeSyncActions(sc *syncContext, localFiles map[string]LocalFileInfo, remoteFiles map[string]*genai.File) SyncActions {
	// Ensure loggers are valid
	if sc.infoLog == nil || sc.errorLog == nil || sc.debugLog == nil {
		log.Println("ERROR: computeSyncActions called with nil loggers in syncContext!")
		// Return empty actions, although this indicates a programming error elsewhere.
		return SyncActions{}
	}
	sc.infoLog.Println("[API HELPER Sync] Comparing local and remote file states...")
	actions := SyncActions{}
	localProcessed := make(map[string]bool)
	upToDateCount := int64(0) // Local counter for logging

	for relPath, localInfo := range localFiles { // Check local files
		localProcessed[relPath] = true
		apiFileInfo, existsInAPI := remoteFiles[relPath]

		if !existsInAPI || apiFileInfo == nil { // Upload case: Exists locally, not remotely
			sc.debugLog.Printf("[DEBUG Compare] Action=Upload for %s (Not found in API map)", relPath)
			actions.FilesToUpload = append(actions.FilesToUpload, localInfo)
		} else { // Exists in both, compare hashes

			// *** FIX: Treat API hash bytes as ASCII hex string ***
			apiHashStr := string(apiFileInfo.Sha256Hash)

			if apiHashStr != localInfo.Hash { // Update case: Hashes differ
				sc.debugLog.Printf(
					"[DEBUG Compare] Action=Update for %s --- HASH MISMATCH --- Local: [%s] != Remote Str: [%s]",
					relPath, localInfo.Hash, apiHashStr,
				)
				sc.debugLog.Printf("[DEBUG Compare] API File Details: Name=%s, Raw Hash Bytes: %x",
					apiFileInfo.Name, apiFileInfo.Sha256Hash) // Log raw bytes for inspection

				actions.FilesToUpdate = append(actions.FilesToUpdate, uploadJob{
					localAbsPath:    localInfo.AbsPath,
					relPath:         localInfo.RelPath,
					localHash:       localInfo.Hash,
					existingApiFile: apiFileInfo,
				})
			} else { // Up-to-date case: Hashes match
				sc.incrementStat("files_up_to_date") // Increment global stat counter
				upToDateCount++                      // Increment local counter for summary log
				// Reduce noise: only log up-to-date in debug
				sc.debugLog.Printf("[DEBUG Compare] Action=None (Up-to-date) for %s (Hashes Match: %s)", relPath, localInfo.Hash[:min(len(localInfo.Hash), 8)])
			}
		}
	}

	// Check remote files for deletions
	for dispName, apiFileInfo := range remoteFiles {
		if _, existsLocally := localProcessed[dispName]; !existsLocally {
			// Only delete if it would have matched the filter pattern
			if sc.filterPattern != "" {
				baseName := filepath.Base(dispName)
				match, _ := filepath.Match(sc.filterPattern, baseName)
				if !match {
					sc.debugLog.Printf("[DEBUG Compare] Skipping delete for remote %s (doesn't match filter %q)", dispName, sc.filterPattern)
					continue
				}
			}
			// Passed filter or no filter exists
			sc.debugLog.Printf("[DEBUG Compare] Action=Delete for remote %s (API Name: %s)", dispName, apiFileInfo.Name)
			sc.incrementStat("files_deleted_locally") // Stat counts files *identified* for deletion
			actions.FilesToDelete = append(actions.FilesToDelete, apiFileInfo)
		}
	}

	// Use the local counter for the summary message
	sc.infoLog.Printf("[API HELPER Sync] Comparison complete. Plan: %d uploads, %d updates (%d up-to-date), %d deletes.",
		len(actions.FilesToUpload), len(actions.FilesToUpdate), upToDateCount, len(actions.FilesToDelete))
	return actions
}

// Ensure calculateFileHash and min are accessible (defined in core package, e.g., tools_file_api.go and helpers.go)
