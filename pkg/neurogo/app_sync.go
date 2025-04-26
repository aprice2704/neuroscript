// filename: pkg/neurogo/app_sync.go
package neurogo

import (
	"context"
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/core" // For core helpers like SecureFilePath and file API helpers
	// Assuming file_api_helpers exist after refactoring tools_file_api.go
)

// runSyncMode handles the file synchronization logic when -sync flag is used.
func (a *App) runSyncMode(ctx context.Context) error {
	a.Logger.Info("Sync Target Directory: %s", a.Config.SyncDir)
	if a.Config.SyncFilter != "" {
		a.Logger.Info("Sync Filter Pattern: %s", a.Config.SyncFilter)
	}
	a.Logger.Info("Sync Ignore Gitignore: %t", a.Config.SyncIgnoreGitignore)

	// 1. Validate Sync Directory
	// Use current working directory as the base for SecureFilePath validation
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	absSyncDir, secErr := core.SecureFilePath(a.Config.SyncDir, cwd) // Validate relative to CWD
	if secErr != nil {
		// SecureFilePath already wraps ErrPathViolation, just return it
		return fmt.Errorf("invalid sync directory path: %w", secErr)
	}
	// Ensure the validated path is actually a directory
	dirInfo, statErr := os.Stat(absSyncDir)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return fmt.Errorf("sync directory does not exist: %s (resolved to %s)", a.Config.SyncDir, absSyncDir)
		}
		return fmt.Errorf("failed to stat sync directory %s: %w", absSyncDir, statErr)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("sync path is not a directory: %s (resolved to %s)", a.Config.SyncDir, absSyncDir)
	}
	a.Logger.Info("Validated absolute sync directory: %s", absSyncDir)

	// 2. Ensure LLM Client is available (checked in Run, but double-check)
	if a.llmClient == nil || a.llmClient.Client() == nil {
		return fmt.Errorf("LLM Client not available for sync operation")
	}

	// 3. Call the refactored sync helper function
	// NOTE: Assumes 'syncDirectoryHelper' exists after refactoring tools_file_api.go
	// It needs context, path, filter, ignore flag, the base *genai.Client*, and loggers.
	stats, syncErr := core.SyncDirectoryUpHelper( // Renamed Helper for clarity
		ctx,
		absSyncDir,
		a.Config.SyncFilter,
		a.Config.SyncIgnoreGitignore,
		a.llmClient.Client(), // Pass the underlying *genai.Client
		a.InfoLog,
		a.ErrorLog,
		a.DebugLog, // Pass DebugLog for potential verbose output in helper
	)

	// 4. Log Summary Stats (Similar to gensync)
	a.Logger.Info("--------------------")
	a.Logger.Info("Sync Summary:")
	logStat := func(key string, value interface{}) { // Simple helper for consistent logging
		a.Logger.Info("  %-25s: %v", key, value)
	}
	logStat("Directory Synced", absSyncDir)
	if stats != nil { // Check if stats map was returned
		logStat("Files Scanned", stats["files_scanned"])
		logStat("Files Ignored", stats["files_ignored"]) // Ensure helper returns this key
		logStat("Files Up-to-date", stats["files_up_to_date"])
		logStat("Files Uploaded", stats["files_uploaded"])
		logStat("Files Updated (API)", stats["files_updated_api"]) // Need consistent keys from helper
		logStat("Files Deleted (API)", stats["files_deleted_api"])
		logStat("Upload Errors", stats["upload_errors"])
		logStat("Delete Errors", stats["delete_errors"])
		logStat("Walk Errors", stats["walk_errors"])
		logStat("Hash Errors", stats["hash_errors"])
		logStat("List API Errors", stats["list_api_errors"])
	} else {
		a.Logger.Info("  (Sync statistics not available)")
	}
	a.Logger.Info("--------------------")

	if syncErr != nil {
		// Log the primary error returned by the sync helper
		a.Logger.Error("Sync operation failed: %v", syncErr)
		// Check stats map for specific error counts if available
		if stats != nil && (stats["upload_errors"].(int64) > 0 || stats["delete_errors"].(int64) > 0 || stats["list_api_errors"].(int64) > 0) {
			a.ErrorLog.Println("Sync completed with errors.")
		}
		return fmt.Errorf("sync operation failed: %w", syncErr) // Return the error
	}

	a.Logger.Info("Sync completed successfully.")
	return nil // Success
}
