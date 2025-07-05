// NeuroScript Version: 0.3.0
// File version: 0.1.1
// Corrected package for SyncDirectoryUpHelper from core to tool.
// filename: pkg/neurogo/app_sync.go
package neurogo

import (
	"context"
	// Corrected import from core
)

// runSyncMode handles the file synchronization logic when -sync flag is used.
func (a *App) runSyncMode(ctx context.Context) error {

	a.Log.Info("--- SYNC DISABLED in app_sync.go Running in Sync Mode ---")
	// a.Log.Info("Sync Target Directory:", "path", a.Config.SyncDir)
	// if a.Config.SyncFilter != "" {
	// 	a.Log.Info("Sync Filter Pattern:", "filter", a.Config.SyncFilter)
	// }
	// a.Log.Info("Sync Ignore Gitignore:", "ignore", a.Config.SyncIgnoreGitignore)

	// // 1. Validate Sync Directory
	// _, err := os.Getwd()
	// if err != nil {
	// 	a.Log.Error("Failed to get current working directory", "error", err)
	// 	return fmt.Errorf("failed to get current working directory: %w", err)
	// }

	// absSyncDir, err := filepath.Abs(a.Config.SyncDir)
	// if err != nil {
	// 	a.Log.Error("Failed to get absolute path for sync directory", "path", a.Config.SyncDir, "error", err)
	// 	return fmt.Errorf("failed to resolve absolute path for sync dir '%s': %w", a.Config.SyncDir, err)
	// }
	// absSyncDir = filepath.Clean(absSyncDir)

	// a.Log.Debug("Resolved absolute sync directory", "path", absSyncDir)

	// dirInfo, statErr := os.Stat(absSyncDir)
	// if statErr != nil {
	// 	if os.IsNotExist(statErr) {
	// 		a.Log.Error("Sync directory does not exist.", "path", absSyncDir)
	// 		return fmt.Errorf("sync directory does not exist: %s (resolved to %s)", a.Config.SyncDir, absSyncDir)
	// 	}
	// 	a.Log.Error("Failed to stat sync directory", "path", absSyncDir, "error", statErr)
	// 	return fmt.Errorf("failed to stat sync directory %s: %w", absSyncDir, statErr)
	// }
	// if !dirInfo.IsDir() {
	// 	a.Log.Error("Sync path is not a directory.", "path", absSyncDir)
	// 	return fmt.Errorf("sync path is not a directory: %s (resolved to %s)", a.Config.SyncDir, absSyncDir)
	// }
	// a.Log.Info("Validated sync directory exists and is a directory.", "path", absSyncDir)

	// // 2. Ensure Interpreter is available (needed by SyncDirectoryUpHelper)
	// if a.interpreter == nil {
	// 	a.Log.Error("Interpreter not available for sync operation.")
	// 	return fmt.Errorf("interpreter not available for sync operation")
	// }
	// if a.llmClient == nil || a.llmClient.Client() == nil {
	// 	a.Log.Error("LLM Client not available for sync operation (required by interpreter).")
	// 	return fmt.Errorf("LLM Client not available for sync operation")
	// }

	// // 3. Call the sync helper function with correct arguments
	// a.Log.Info("FIXME Starting directory sync operation...")
	// // Corrected to use tool package
	// stats, syncErr := tool.SyncDirectoryUpHelper(
	// 	ctx,
	// 	absSyncDir,
	// 	a.Config.SyncFilter,
	// 	a.Config.SyncIgnoreGitignore,
	// 	a.interpreter,
	// )

	// // 4. Log Summary Stats
	// a.Log.Info("--------------------")
	// a.Log.Info("Sync Summary:")
	// logStat := func(key string, value interface{}) {
	// 	logVal := false
	// 	switch v := value.(type) {
	// 	case int:
	// 		if v != 0 {
	// 			logVal = true
	// 		}
	// 	case int64:
	// 		if v != 0 {
	// 			logVal = true
	// 		}
	// 	case string:
	// 		if v != "" {
	// 			logVal = true
	// 		}
	// 	case bool:
	// 		if v {
	// 			logVal = true
	// 		}
	// 	case nil:
	// 	default:
	// 		logVal = true
	// 	}
	// 	if logVal {
	// 		a.Log.Info(fmt.Sprintf("  %-25s: %v", key, value))
	// 	}
	// }

	// logStat("Directory Synced", absSyncDir)
	// if stats != nil {
	// 	logStat("Files Scanned", stats["files_scanned"])
	// 	logStat("Files Ignored", stats["files_ignored"])
	// 	logStat("Files Up-to-date", stats["files_up_to_date"])
	// 	logStat("Files To Upload", stats["files_to_upload"])
	// 	logStat("Files To Update (API)", stats["files_to_update"])
	// 	logStat("Files To Delete (API)", stats["files_to_delete"])
	// 	logStat("Uploads Attempted", stats["uploads_attempted"])
	// 	logStat("Uploads Succeeded", stats["uploads_succeeded"])
	// 	logStat("Upload Errors", stats["upload_errors"])
	// 	logStat("Deletes Attempted", stats["deletes_attempted"])
	// 	logStat("Deletes Succeeded", stats["deletes_succeeded"])
	// 	logStat("Delete Errors", stats["delete_errors"])
	// 	logStat("Walk Errors", stats["walk_errors"])
	// 	logStat("Hash Errors", stats["hash_errors"])
	// 	logStat("List API Errors", stats["list_api_errors"])
	// 	logStat("Workers Used", stats["workers_used"])

	// } else {
	// 	a.Log.Info("  (Sync statistics map not available)")
	// }
	// a.Log.Info("--------------------")

	// if syncErr != nil {
	// 	a.Log.Error("Sync operation failed overall.", "error", syncErr)

	// 	var uploadErrors, deleteErrors, listErrors int64
	// 	var ok bool
	// 	if stats != nil {
	// 		if ue, exists := stats["upload_errors"]; exists {
	// 			uploadErrors, ok = ue.(int64)
	// 			if !ok {
	// 				a.Log.Warn("Stats key 'upload_errors' has unexpected type.", "type", fmt.Sprintf("%T", ue))
	// 			}
	// 		}
	// 		if de, exists := stats["delete_errors"]; exists {
	// 			deleteErrors, ok = de.(int64)
	// 			if !ok {
	// 				a.Log.Warn("Stats key 'delete_errors' has unexpected type.", "type", fmt.Sprintf("%T", de))
	// 			}
	// 		}
	// 		if le, exists := stats["list_api_errors"]; exists {
	// 			listErrors, ok = le.(int64)
	// 			if !ok {
	// 				a.Log.Warn("Stats key 'list_api_errors' has unexpected type.", "type", fmt.Sprintf("%T", le))
	// 			}
	// 		}
	// 	}

	// 	if uploadErrors > 0 || deleteErrors > 0 || listErrors > 0 {
	// 		a.Log.Warn("Sync completed with specific errors.", "upload_errors", uploadErrors, "delete_errors", deleteErrors, "list_api_errors", listErrors)
	// 	}
	// 	return fmt.Errorf("sync operation failed: %w", syncErr)
	// }

	// a.Log.Info("Sync completed successfully.")
	return nil
}
