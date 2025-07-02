// filename: pkg/neurogo/app_sync.go
package neurogo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"	// Import filepath

	"github.com/aprice2704/neuroscript/pkg/core"	// For core helpers like SecureFilePath and file API helpers
	// Assuming file_api_helpers exist after refactoring tools_file_api.go
)

// runSyncMode handles the file synchronization logic when -sync flag is used.
func (a *App) runSyncMode(ctx context.Context) error {
	// Use a.Log consistently
	a.Log.Info("--- Running in Sync Mode ---")
	a.Log.Info("Sync Target Directory:", "path", a.Config.SyncDir)
	if a.Config.SyncFilter != "" {
		a.Log.Info("Sync Filter Pattern:", "filter", a.Config.SyncFilter)
	}
	a.Log.Info("Sync Ignore Gitignore:", "ignore", a.Config.SyncIgnoreGitignore)

	// 1. Validate Sync Directory
	// Use current working directory as the base for SecureFilePath validation
	_, err := os.Getwd()
	if err != nil {
		a.Log.Error("Failed to get current working directory", "error", err)
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	// SecureFilePath requires the allowed directory first, then the path to check relative to it.
	// To validate SyncDir relative to CWD, SyncDir IS the path to check. The allowedDir is CWD.
	// Let's assume SyncDir is relative to CWD or absolute. We need its absolute path first.
	absSyncDir, err := filepath.Abs(a.Config.SyncDir)
	if err != nil {
		a.Log.Error("Failed to get absolute path for sync directory", "path", a.Config.SyncDir, "error", err)
		return fmt.Errorf("failed to resolve absolute path for sync dir '%s': %w", a.Config.SyncDir, err)
	}
	absSyncDir = filepath.Clean(absSyncDir)	// Clean the absolute path

	// TODO: Security Check - Ensure absSyncDir is within an allowed base path if necessary.
	// The original SecureFilePath call here might have been intended differently,
	// possibly ensuring SyncDir was *within* SandboxDir? Re-evaluate security needs.
	// For now, we just check existence and if it's a directory.

	a.Log.Debug("Resolved absolute sync directory", "path", absSyncDir)

	// Ensure the path exists and is a directory
	dirInfo, statErr := os.Stat(absSyncDir)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			a.Log.Error("Sync directory does not exist.", "path", absSyncDir)
			return fmt.Errorf("sync directory does not exist: %s (resolved to %s)", a.Config.SyncDir, absSyncDir)
		}
		a.Log.Error("Failed to stat sync directory", "path", absSyncDir, "error", statErr)
		return fmt.Errorf("failed to stat sync directory %s: %w", absSyncDir, statErr)
	}
	if !dirInfo.IsDir() {
		a.Log.Error("Sync path is not a directory.", "path", absSyncDir)
		return fmt.Errorf("sync path is not a directory: %s (resolved to %s)", a.Config.SyncDir, absSyncDir)
	}
	a.Log.Info("Validated sync directory exists and is a directory.", "path", absSyncDir)

	// 2. Ensure Interpreter is available (needed by SyncDirectoryUpHelper)
	if a.interpreter == nil {
		a.Log.Error("Interpreter not available for sync operation.")
		return fmt.Errorf("interpreter not available for sync operation")
	}
	// LLM Client check (needed by interpreter which is needed by SyncDirectoryUpHelper)
	if a.llmClient == nil || a.llmClient.Client() == nil {
		a.Log.Error("LLM Client not available for sync operation (required by interpreter).")
		return fmt.Errorf("LLM Client not available for sync operation")
	}

	// 3. Call the sync helper function with correct arguments
	a.Log.Info("Starting directory sync operation...")
	// Signature: (ctx context.Context, absLocalDir string, filterPattern string, ignoreGitignore bool, interp *Interpreter) (map[string]interface{}, error)
	stats, syncErr := core.SyncDirectoryUpHelper(
		ctx,
		absSyncDir,			// Correctly passed abs path
		a.Config.SyncFilter,		// Pass filter pattern
		a.Config.SyncIgnoreGitignore,	// Pass ignore flag
		a.interpreter,			// Pass the Interpreter
	)

	// 4. Log Summary Stats
	a.Log.Info("--------------------")
	a.Log.Info("Sync Summary:")
	logStat := func(key string, value interface{}) {
		// Log only if value is non-zero (or non-nil/non-empty for non-numeric)
		logVal := false
		switch v := value.(type) {
		case int:
			if v != 0 {
				logVal = true
			}
		case int64:
			if v != 0 {
				logVal = true
			}
		case string:
			if v != "" {
				logVal = true
			}
		case bool:
			if v {
				logVal = true
			}	// Log if true
		case nil:
			// don't log nil
		default:
			logVal = true	// Log unknown types
		}
		if logVal {
			a.Log.Info(fmt.Sprintf("  %-25s: %v", key, value))	// Use fmt.Sprintf with logger
		}
	}

	logStat("Directory Synced", absSyncDir)
	if stats != nil {	// Check if stats map was returned
		// Access stats using map keys, as per user's code and function signature
		logStat("Files Scanned", stats["files_scanned"])
		logStat("Files Ignored", stats["files_ignored"])
		logStat("Files Up-to-date", stats["files_up_to_date"])
		logStat("Files To Upload", stats["files_to_upload"])		// Key from sync_logic.go
		logStat("Files To Update (API)", stats["files_to_update"])	// Key from sync_logic.go
		logStat("Files To Delete (API)", stats["files_to_delete"])	// Key from sync_logic.go
		logStat("Uploads Attempted", stats["uploads_attempted"])	// Key from sync_workers.go via context
		logStat("Uploads Succeeded", stats["uploads_succeeded"])	// Key from sync_workers.go via context
		logStat("Upload Errors", stats["upload_errors"])		// Key from sync_workers.go via context
		logStat("Deletes Attempted", stats["deletes_attempted"])	// Key from sync_workers.go via context
		logStat("Deletes Succeeded", stats["deletes_succeeded"])	// Key from sync_workers.go via context
		logStat("Delete Errors", stats["delete_errors"])		// Key from sync_workers.go via context
		logStat("Walk Errors", stats["walk_errors"])			// Key from sync_logic.go
		logStat("Hash Errors", stats["hash_errors"])			// Key from sync_logic.go
		logStat("List API Errors", stats["list_api_errors"])		// Key from sync_morehelpers.go via context
		logStat("Workers Used", stats["workers_used"])			// Added in previous helper version

	} else {
		a.Log.Info("  (Sync statistics map not available)")
	}
	a.Log.Info("--------------------")

	if syncErr != nil {
		a.Log.Error("Sync operation failed overall.", "error", syncErr)

		// Add robust check for specific error counts from stats map
		var uploadErrors, deleteErrors, listErrors int64
		var ok bool
		if stats != nil {
			if ue, exists := stats["upload_errors"]; exists {
				uploadErrors, ok = ue.(int64)
				if !ok {
					a.Log.Warn("Stats key 'upload_errors' has unexpected type.", "type", fmt.Sprintf("%T", ue))
				}
			}
			if de, exists := stats["delete_errors"]; exists {
				deleteErrors, ok = de.(int64)
				if !ok {
					a.Log.Warn("Stats key 'delete_errors' has unexpected type.", "type", fmt.Sprintf("%T", de))
				}
			}
			if le, exists := stats["list_api_errors"]; exists {
				listErrors, ok = le.(int64)
				if !ok {
					a.Log.Warn("Stats key 'list_api_errors' has unexpected type.", "type", fmt.Sprintf("%T", le))
				}
			}
		}

		if uploadErrors > 0 || deleteErrors > 0 || listErrors > 0 {
			a.Log.Warn("Sync completed with specific errors.", "upload_errors", uploadErrors, "delete_errors", deleteErrors, "list_api_errors", listErrors)
			// The primary syncErr might already reflect these, but logging helps.
		}
		// Return the primary error from the helper function
		return fmt.Errorf("sync operation failed: %w", syncErr)
	}

	a.Log.Info("Sync completed successfully.")
	return nil	// Success
}