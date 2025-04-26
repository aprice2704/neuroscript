// filename: pkg/core/sync_morehelpers.go
package core

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	// Required imports for genai and gitignore
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/google/generative-ai-go/genai"
	gitignore "github.com/sabhiram/go-gitignore"
	"google.golang.org/api/iterator"
)

// initializeSyncState sets up default loggers and the statistics map.
func initializeSyncState(logger interfaces.Logger) (
	stats map[string]interface{},
	incrementStat func(string),
	ilogger interfaces.Logger, // Return modified loggers
) {

	stats = map[string]interface{}{
		"files_scanned": int64(0), "files_ignored": int64(0), "files_filtered": int64(0),
		"files_uploaded": int64(0), "files_updated_api": int64(0), "files_deleted_api": int64(0),
		"files_up_to_date": int64(0), "upload_errors": int64(0), "delete_errors": int64(0),
		"list_api_errors": int64(0), "walk_errors": int64(0), "hash_errors": int64(0),
		"files_processed": int64(0), "files_deleted_locally": int64(0),
	}
	// Use the potentially defaulted logger in the closure
	incrementStat = func(key string) {
		if v, ok := stats[key].(int64); ok {
			stats[key] = v + 1
		} else {
			logger.Error("[CRITICAL SYNC HELPER] Invalid stat key %s", key)
		} // Use logger here
	}
	// Return the (potentially defaulted) loggers along with stats and function
	return stats, incrementStat, logger
}

// listExistingAPIFiles fetches the list of files from the API and returns them as a map.
// Requires access to syncContext definition (from sync_types.go).
func listExistingAPIFiles(sc *syncContext) (map[string]*genai.File, error) {
	// Ensure loggers in context are valid (they should be after initializeSyncState)
	if sc.logger == nil {
		// This indicates a programming error if syncContext wasn't populated correctly
		return nil, errors.New("listExistingAPIFiles called with nil loggers in syncContext")
	}
	if sc.client == nil {
		return nil, errors.New("genai client is nil in listExistingAPIFiles")
	}

	sc.logger.Debug("[API HELPER Sync] Listing current API files...")
	sc.logger.Debug("[API HELPER List] Fetching file list from API...") // Use context's infoLog
	if sc.ctx == nil {
		sc.ctx = context.Background()
	}

	iter := sc.client.ListFiles(sc.ctx)
	results := []*genai.File{}
	fetchErrors := 0
	for {
		file, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			errMsg := fmt.Sprintf("Error fetching file list page: %v", err)
			sc.logger.Error("[API HELPER List] %s", errMsg) // Use context's errorLog
			fetchErrors++
			continue
		}
		results = append(results, file)
	}
	sc.logger.Info("[API HELPER List] Found %d files. Encountered %d errors during fetch.", len(results), fetchErrors) // Use infoLog

	apiFilesMap := make(map[string]*genai.File)
	for _, file := range results {
		if file.DisplayName != "" {
			apiFilesMap[file.DisplayName] = file
			hashPrefix := ""
			if len(file.Sha256Hash) > 0 {
				// Assumes min is accessible globally or via import from helpers.go
				hashPrefix = hex.EncodeToString(file.Sha256Hash)[:min(len(hex.EncodeToString(file.Sha256Hash)), 8)]
			}
			sc.logger.Debug("API HELPER Sync] API File Found: Name=%s, DisplayName=%s, SHA=%s...", file.Name, file.DisplayName, hashPrefix) // Use debugLog
		} else {
			sc.logger.Warn("[WARN API HELPER Sync] API File Found empty DisplayName: %s", file.Name) // Use debugLog
		}
	}
	sc.logger.Info("[API HELPER Sync] Found %d API files.", len(apiFilesMap)) // Use infoLog

	var returnErr error
	if fetchErrors > 0 {
		sc.incrementStat("list_api_errors") // Use incrementStat from context
		returnErr = fmt.Errorf("encountered %d errors fetching file list", fetchErrors)
		sc.logger.Error("[ERROR API HELPER Sync] List API files finished with errors: %v", returnErr) // Use errorLog
	}
	return apiFilesMap, returnErr
}

// initializeGitignore loads the .gitignore file if requested and available.
// Requires access to syncContext definition (from sync_types.go).
func initializeGitignore(sc *syncContext, ignoreGitignore bool) *gitignore.GitIgnore {
	// Ensure loggers in context are valid
	if sc.logger == nil {
		sc.logger.Error("ERROR: initializeGitignore called with nil loggers in syncContext!")
		return nil // Cannot proceed safely
	}
	if ignoreGitignore {
		sc.logger.Debug("[API HELPER Sync] Ignoring .gitignore")
		return nil
	}
	gitignorePath := filepath.Join(sc.absLocalDir, ".gitignore")
	sc.logger.Debug("API HELPER Sync] Loading gitignore: %s", gitignorePath)
	ignorer, gitignoreErr := gitignore.CompileIgnoreFile(gitignorePath)
	if gitignoreErr != nil {
		if os.IsNotExist(gitignoreErr) {
			sc.logger.Debug("[API HELPER Sync] No .gitignore found.")
		} else {
			sc.logger.Error("[WARN API HELPER Sync] Error read gitignore %s: %v", gitignorePath, gitignoreErr)
		}
		return nil
	}
	if ignorer != nil {
		sc.logger.Debug("[API HELPER Sync] Using gitignore rules.")
	}
	return ignorer
}

// Note: Ensure syncContext (from sync_types.go) and min (from helpers.go?) are accessible.
