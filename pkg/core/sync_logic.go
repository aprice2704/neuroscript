// filename: pkg/core/sync_logic.go
package core

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log" // Added for nil logger check fallback
	"os"
	"path/filepath"
	"strings"

	// Added for string comparison below
	"github.com/google/generative-ai-go/genai"
	// Assumes sync_types.go, and other necessary imports like gitignore exist
)

// gatherLocalFiles walks the local directory and collects info about files that pass filters.
// *** MODIFIED: Calls calculateFileHash with interpreter from syncContext ***
func gatherLocalFiles(sc *syncContext) (map[string]LocalFileInfo, error) {
	// Ensure loggers and interp are valid
	if sc.logger == nil || sc.interp == nil { // Check interp too
		log.Println("ERROR: gatherLocalFiles called with nil loggers or interpreter in syncContext!")
		return nil, errors.New("internal error: loggers or interpreter not initialized in sync context")
	}

	sc.logger.Info("[API HELPER Sync] Scanning local directory: %s", sc.absLocalDir)
	localFiles := make(map[string]LocalFileInfo)
	var firstWalkError error

	walkErr := filepath.WalkDir(sc.absLocalDir, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			sc.logger.Error("[ERROR API HELPER Sync Walk] Initial error for %q: %v", currentPath, err)
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
		sc.logger.Debug("API HELPER Sync Walk] Visiting: %q (IsDir: %t)", currentPath, d.IsDir())
		if currentPath == sc.absLocalDir {
			return nil
		} // Skip root

		relPath, relErr := filepath.Rel(sc.absLocalDir, currentPath)
		if relErr != nil {
			sc.logger.Error("[ERROR API HELPER Sync Walk] Failed RelPath %q: %v", currentPath, relErr)
			sc.incrementStat("walk_errors")
			return nil // Skip entry
		}
		relPath = filepath.ToSlash(relPath)

		// Gitignore Check
		if sc.ignorer != nil && sc.ignorer.MatchesPath(relPath) {
			sc.incrementStat("files_ignored")
			if d.IsDir() {
				sc.logger.Debug("API HELPER Sync Walk] Gitignored dir: %s", relPath)
				return filepath.SkipDir
			}
			sc.logger.Debug("API HELPER Sync Walk] Gitignored file: %s", relPath)
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
				sc.logger.Error("[ERROR API HELPER Sync Walk] Invalid filter %q: %v", sc.filterPattern, matchErr)
				sc.incrementStat("walk_errors")
				return fmt.Errorf("invalid filter pattern: %w", matchErr) // Stop on bad filter
			}
			if !match {
				sc.incrementStat("files_filtered")
				sc.logger.Debug("API HELPER Sync Walk] Filtered out file: %s", relPath)
				return nil
			}
		}

		// Passed checks, process the file
		sc.incrementStat("files_scanned")
		sc.logger.Debug("API HELPER Sync Walk] Processing file: %s", relPath)

		// Calculate Hash
		// *** MODIFIED CALL: Pass sc.interp ***
		localHash, hashErr := calculateFileHash(sc.interp, relPath) // Pass interpreter and relative path
		if hashErr != nil {
			sc.logger.Error("[ERROR API HELPER Sync Walk] Hash failed %s: %v", relPath, hashErr)
			sc.incrementStat("hash_errors")
			// Decide if hash error is fatal or just skips file
			if errors.Is(hashErr, ErrFileNotFound) {
				sc.logger.Warn("API HELPER Sync Walk] Skipping file %s as it was not found during hashing (possibly deleted during walk)", relPath)
				return nil // Skip file if not found during hash
			}
			// For other hash errors, maybe stop the walk? For now, skip file.
			return nil // Skip file on other hash errors
		}

		// Store file info
		localFiles[relPath] = LocalFileInfo{
			RelPath: relPath,
			AbsPath: currentPath, // Store absolute path for upload worker
			Hash:    localHash,
		}
		hashPrefix := localHash
		if len(hashPrefix) > 8 {
			hashPrefix = hashPrefix[:8]
		} // Safe slice
		sc.logger.Debug("API HELPER Sync Walk] Stored local info for: %s (Hash: %s...)", relPath, hashPrefix)
		return nil
	}) // End WalkDir Callback

	sc.logger.Info("[API HELPER Sync] Local scan completed. Found %d candidate files. WalkErr: %v", len(localFiles), walkErr)
	// Return combined error
	if walkErr != nil {
		return localFiles, walkErr
	}
	return localFiles, firstWalkError
}

// computeSyncActions compares local and remote file lists and determines necessary actions.
// FIXED: Hash comparison treats API bytes as ASCII hex representation.
func computeSyncActions(sc *syncContext, localFiles map[string]LocalFileInfo, remoteFiles map[string]*genai.File) SyncActions {
	// Ensure loggers are valid
	if sc.logger == nil {
		panic("ERROR: computeSyncActions called with nil loggers in syncContext!")
	}
	sc.logger.Debug("[API HELPER Sync] Comparing local and remote file states...")
	actions := SyncActions{}
	localProcessed := make(map[string]bool)
	upToDateCount := int64(0) // Local counter for logging

	for relPath, localInfo := range localFiles { // Check local files
		localProcessed[relPath] = true
		apiFileInfo, existsInAPI := remoteFiles[relPath]

		if !existsInAPI || apiFileInfo == nil { // Upload case: Exists locally, not remotely
			sc.logger.Debug("Compare] Action=Upload for %s (Not found in API map)", relPath)
			actions.FilesToUpload = append(actions.FilesToUpload, localInfo)
		} else { // Exists in both, compare hashes

			// *** FIX: Treat API hash bytes as ASCII hex string ***
			apiHashStr := hex.EncodeToString(apiFileInfo.Sha256Hash) // Convert API hash bytes to hex string

			if apiHashStr != localInfo.Hash { // Update case: Hashes differ
				sc.logger.Debug(
					"[DEBUG Compare] Action=Update for %s --- HASH MISMATCH --- Local: [%s] != Remote Hex: [%s]",
					relPath, localInfo.Hash, apiHashStr,
				)
				// sc.logger.Debug("Compare] API File Details: Name=%s, Raw Hash Bytes: %x", apiFileInfo.Name, apiFileInfo.Sha256Hash) // Log raw bytes for inspection

				actions.FilesToUpdate = append(actions.FilesToUpdate, uploadJob{
					localAbsPath:    localInfo.AbsPath,
					relPath:         localInfo.RelPath,
					localHash:       localInfo.Hash,
					existingApiFile: apiFileInfo,
				})
			} else { // Up-to-date case: Hashes match
				sc.incrementStat("files_up_to_date") // Increment global stat counter
				upToDateCount++                      // Increment local counter for summary log
				hashPrefix := localInfo.Hash
				if len(hashPrefix) > 8 {
					hashPrefix = hashPrefix[:8]
				} // Safe slice
				sc.logger.Debug("Compare] Action=None (Up-to-date) for %s (Hashes Match: %s...)", relPath, hashPrefix)
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
					sc.logger.Debug("Compare] Skipping delete for remote %s (doesn't match filter %q)", dispName, sc.filterPattern)
					continue
				}
			}
			// Passed filter or no filter exists
			sc.logger.Debug("Compare] Action=Delete for remote %s (API Name: %s)", dispName, apiFileInfo.Name)
			sc.incrementStat("files_deleted_locally") // Stat counts files *identified* for deletion
			actions.FilesToDelete = append(actions.FilesToDelete, apiFileInfo)
		}
	}

	// Use the local counter for the summary message
	sc.logger.Info("[API HELPER Sync] Comparison complete. Plan: %d uploads, %d updates (%d up-to-date), %d deletes.",
		len(actions.FilesToUpload), len(actions.FilesToUpdate), upToDateCount, len(actions.FilesToDelete))
	return actions
}

// --- Tool: SyncFiles (Wrapper) ---
// *** MODIFIED: Calls checkGenAIClient with interpreter ***
func toolSyncFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// *** MODIFIED: Call checkGenAIClient helper ***
	_, clientErr := checkGenAIClient(interpreter) // Check and get client
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.SyncFiles: %w", clientErr)
	}
	// *** END MODIFIED ***

	// Argument parsing (unchanged)
	if len(args) < 2 || len(args) > 4 {
		return nil, fmt.Errorf("TOOL.SyncFiles: expected 2-4 arguments (direction, local_dir, [filter_pattern], [ignore_gitignore]), got %d", len(args))
	}
	direction, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.SyncFiles: direction must be string")
	}
	localDir, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.SyncFiles: local_dir must be string")
	}
	if localDir == "" {
		return nil, errors.New("TOOL.SyncFiles: local_dir empty")
	} // Use errors.New

	var filterPattern string
	if len(args) >= 3 {
		if args[2] != nil { // Check if optional arg was provided and not nil
			filterPattern, ok = args[2].(string)
			if !ok {
				return nil, fmt.Errorf("TOOL.SyncFiles: filter_pattern must be string or null")
			}
		}
	}

	var ignoreGitignore bool = false // Default value
	if len(args) == 4 {
		if args[3] != nil { // Check if optional arg was provided and not nil
			ignoreGitignore, ok = args[3].(bool)
			if !ok {
				return nil, fmt.Errorf("TOOL.SyncFiles: ignore_gitignore must be boolean or null")
			}
		}
	}

	direction = strings.ToLower(direction)
	if direction != "up" {
		return nil, fmt.Errorf("TOOL.SyncFiles: direction '%s' not supported", direction)
	}

	// Path validation (unchanged)
	absLocalDir, secErr := ResolveAndSecurePath(localDir, interpreter.sandboxDir)
	if secErr != nil {
		return nil, fmt.Errorf("TOOL.SyncFiles: invalid local_dir '%s': %w", localDir, secErr)
	} // Wrap directly

	dirInfo, statErr := os.Stat(absLocalDir)
	if statErr != nil {
		return nil, fmt.Errorf("TOOL.SyncFiles: cannot access local_dir '%s': %w", localDir, statErr)
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("TOOL.SyncFiles: local_dir '%s' is not a directory", localDir)
	}

	interpreter.logger.Info("Tool: SyncFiles] Validated dir: %s (Ignore .gitignore: %t)", absLocalDir, ignoreGitignore)

	// Call the main helper (passing interpreter)
	// *** MODIFIED: Pass interpreter instead of client/logger ***
	statsMap, syncErr := SyncDirectoryUpHelper(context.Background(), absLocalDir, filterPattern, ignoreGitignore, interpreter)

	return statsMap, syncErr
}
