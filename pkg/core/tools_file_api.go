// filename: pkg/core/tools_file_api.go
package core

import (
	"context"
	"crypto/sha256" // Added for hashing
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs" // Added for WalkDir
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
)

// --- Constants and Vars for Hashing (adapted from gensync) ---
const emptyFileContentForHash = " " // Use a single space for empty file representation
var emptyFileHash string            // Will be calculated in init()

func init() {
	hasher := sha256.New()
	hasher.Write([]byte(emptyFileContentForHash))
	emptyFileHash = hex.EncodeToString(hasher.Sum(nil))
}

// --- Helper: Calculate Local File Hash (adapted from gensync) ---
// For zero-byte files, returns the pre-calculated hash of emptyFileContentForHash
func calculateFileHash(filePath string) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("stat failed for hashing %s: %w", filePath, err)
	}

	if fileInfo.Size() == 0 {
		return emptyFileHash, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file for hashing %s: %w", filePath, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("copying file data for hashing %s: %w", filePath, err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// --- Helper: Check GenAI Client ---
// Returns a standard error if the client is nil.
func checkGenAIClient(interpreter *Interpreter) (*genai.Client, error) {
	client := interpreter.GenAIClient()
	if client == nil {
		return nil, errors.New("GenAI client is not initialized (missing API key?)")
	}
	return client, nil
}

// --- Tool: ListAPIFiles ---

func toolListAPIFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.ListAPIFiles: %w", clientErr)
	}

	interpreter.logger.Println("[TOOL ListAPIFiles] Fetching file list from API...")
	ctx := context.Background() // Or use a context from interpreter if available later
	iter := client.ListFiles(ctx)
	results := []map[string]interface{}{}
	fetchErrors := 0

	for {
		file, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			errMsg := fmt.Sprintf("Error fetching file list page: %v", err)
			interpreter.logger.Printf("[TOOL ListAPIFiles] %s", errMsg)
			fetchErrors++
			// Continue trying to fetch remaining pages if possible
			// Optionally add an error marker to results? For now, just log and count.
			continue // Skip adding this potential partial error result
		}

		fileInfo := map[string]interface{}{
			"name":        file.Name,
			"displayName": file.DisplayName,
			"mimeType":    file.MIMEType,
			"sizeBytes":   file.SizeBytes,
			"createTime":  file.CreateTime.Format(time.RFC3339),
			"updateTime":  file.UpdateTime.Format(time.RFC3339),
			"state":       string(file.State),
			"uri":         file.URI,
			"sha256Hash":  hex.EncodeToString(file.Sha256Hash),
		}
		results = append(results, fileInfo)
	}

	interpreter.logger.Printf("[TOOL ListAPIFiles] Found %d files. Encountered %d errors during fetch.", len(results), fetchErrors)
	if fetchErrors > 0 {
		// Return partial list but also a Go error to signal issues
		return results, fmt.Errorf("TOOL.ListAPIFiles: encountered %d errors fetching file list", fetchErrors)
	}
	return results, nil
}

// --- Tool: DeleteAPIFile ---

func toolDeleteAPIFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.DeleteAPIFile: %w", clientErr)
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.DeleteAPIFile: expected 1 argument (apiFileName), got %d", len(args))
	}
	apiFileName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.DeleteAPIFile: argument must be a string (API file name), got %T", args[0])
	}
	if apiFileName == "" {
		return nil, fmt.Errorf("TOOL.DeleteAPIFile: API file name cannot be empty")
	}
	if !strings.HasPrefix(apiFileName, "files/") {
		interpreter.logger.Printf("[WARN TOOL.DeleteAPIFile] API file name '%s' does not start with 'files/'. Proceeding anyway.", apiFileName)
	}

	interpreter.logger.Printf("[TOOL DeleteAPIFile] Attempting to delete API file: %s", apiFileName)
	ctx := context.Background()
	err := client.DeleteFile(ctx, apiFileName)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to delete API file '%s': %v", apiFileName, err)
		interpreter.logger.Printf("[TOOL DeleteAPIFile] %s", errMsg)
		// Return error message string AND Go error
		return errMsg, fmt.Errorf("TOOL.DeleteAPIFile: %w", err)
	}

	successMsg := fmt.Sprintf("Successfully deleted API file: %s", apiFileName)
	interpreter.logger.Printf("[TOOL DeleteAPIFile] %s", successMsg)
	return successMsg, nil
}

// --- Tool: UploadFile ---

const emptyFileContent = " " // Use a single space for empty file representation

func toolUploadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.UploadFile: %w", clientErr)
	}

	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("TOOL.UploadFile: expected 1 or 2 arguments (localPath, [displayName]), got %d", len(args))
	}
	localPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.UploadFile: localPath argument must be a string, got %T", args[0])
	}
	if localPath == "" {
		return nil, fmt.Errorf("TOOL.UploadFile: localPath cannot be empty")
	}

	var displayName string
	if len(args) == 2 {
		displayName, ok = args[1].(string)
		if !ok {
			if args[1] != nil {
				return nil, fmt.Errorf("TOOL.UploadFile: displayName argument must be a string or null, got %T", args[1])
			}
		}
	}

	securePath, secErr := SecureFilePath(localPath, interpreter.sandboxDir)
	if secErr != nil {
		return nil, fmt.Errorf("TOOL.UploadFile: invalid localPath '%s': %w", localPath, errors.Join(ErrValidationArgValue, secErr))
	}
	interpreter.logger.Printf("[TOOL UploadFile] Validated local path: %s -> %s", localPath, securePath)

	uploadDisplayName := displayName
	if uploadDisplayName == "" {
		if interpreter.sandboxDir != "" {
			relPath, err := filepath.Rel(interpreter.sandboxDir, securePath)
			if err == nil {
				uploadDisplayName = filepath.ToSlash(relPath)
			} else {
				interpreter.logger.Printf("[WARN TOOL.UploadFile] Failed to get relative path for %s from sandbox %s: %v. Using basename.", securePath, interpreter.sandboxDir, err)
				uploadDisplayName = filepath.Base(securePath)
			}
		} else {
			uploadDisplayName = filepath.Base(securePath)
		}
		interpreter.logger.Printf("[TOOL UploadFile] Using default display name: %s", uploadDisplayName)
	} else {
		interpreter.logger.Printf("[TOOL UploadFile] Using provided display name: %s", uploadDisplayName)
	}

	fileInfo, err := os.Stat(securePath)
	if err != nil {
		return nil, fmt.Errorf("TOOL.UploadFile: stat failed for %s: %w", securePath, err)
	}
	isZeroByte := fileInfo.Size() == 0

	var mimeType string
	if isZeroByte {
		mimeType = "text/plain"
	} else {
		mimeType = mime.TypeByExtension(filepath.Ext(securePath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}

	options := &genai.UploadFileOptions{
		MIMEType:    mimeType,
		DisplayName: uploadDisplayName,
	}

	var reader io.Reader
	var fileReader *os.File
	if isZeroByte {
		interpreter.logger.Printf("[TOOL UploadFile] Handling zero-byte file %q by uploading minimal content.", localPath)
		reader = strings.NewReader(emptyFileContent)
	} else {
		fileReader, err = os.Open(securePath)
		if err != nil {
			return nil, fmt.Errorf("TOOL.UploadFile: opening local file %s: %w", securePath, err)
		}
		defer fileReader.Close()
		reader = fileReader
	}

	interpreter.logger.Printf("[TOOL UploadFile] Initiating API upload for %s (DisplayName: %s, Mime: %s)", localPath, uploadDisplayName, mimeType)
	ctx := context.Background()
	apiFile, err := client.UploadFile(ctx, "", reader, options)

	if err != nil {
		return nil, fmt.Errorf("TOOL.UploadFile: API upload call failed for %q: %w", localPath, err)
	}
	interpreter.logger.Printf("[TOOL UploadFile] Upload initiated -> API Name: %s", apiFile.Name)

	startTime := time.Now()
	pollInterval := 1 * time.Second
	const maxPollInterval = 10 * time.Second
	const timeout = 2 * time.Minute

	for apiFile.State == genai.FileStateProcessing {
		if time.Since(startTime) > timeout {
			errMsg := fmt.Sprintf("file %s (API: %s) timed out in processing state after %v", localPath, apiFile.Name, timeout)
			interpreter.logger.Printf("[ERROR TOOL.UploadFile] %s. Attempting delete.", errMsg)
			_ = client.DeleteFile(context.Background(), apiFile.Name)
			return nil, fmt.Errorf("TOOL.UploadFile: %s", errMsg)
		}
		time.Sleep(pollInterval)
		updatedFile, err := client.GetFile(ctx, apiFile.Name)
		if err != nil {
			errMsg := fmt.Errorf("checking processing status failed for %s: %w", apiFile.Name, err)
			interpreter.logger.Printf("[ERROR TOOL.UploadFile] %s. Attempting delete.", errMsg)
			_ = client.DeleteFile(context.Background(), apiFile.Name)
			return nil, fmt.Errorf("TOOL.UploadFile: %s", errMsg)
		}
		apiFile = updatedFile
		pollInterval += 500 * time.Millisecond
		if pollInterval > maxPollInterval {
			pollInterval = maxPollInterval
		}
		interpreter.logger.Printf("[DEBUG TOOL.UploadFile] Polled file %s, state: %s", apiFile.Name, apiFile.State)
	}

	if apiFile.State != genai.FileStateActive {
		errMsg := fmt.Sprintf("file %s finished processing but is not ACTIVE (State: %s, URI: %s)", localPath, apiFile.State, apiFile.URI)
		interpreter.logger.Printf("[ERROR TOOL.UploadFile] %s (API: %s)", errMsg, apiFile.Name)
		_ = client.DeleteFile(context.Background(), apiFile.Name)
		return nil, fmt.Errorf("TOOL.UploadFile: %s", errMsg)
	}

	interpreter.logger.Printf("[TOOL UploadFile] Upload successful and file ACTIVE: %s -> %s", localPath, apiFile.Name)

	resultMap := map[string]interface{}{
		"name":        apiFile.Name,
		"displayName": apiFile.DisplayName,
		"mimeType":    apiFile.MIMEType,
		"sizeBytes":   apiFile.SizeBytes,
		"createTime":  apiFile.CreateTime.Format(time.RFC3339),
		"updateTime":  apiFile.UpdateTime.Format(time.RFC3339),
		"state":       string(apiFile.State),
		"uri":         apiFile.URI,
		"sha256Hash":  hex.EncodeToString(apiFile.Sha256Hash),
	}
	return resultMap, nil
}

// --- Tool: SyncFiles (NEW) ---

func toolSyncFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// --- Argument Parsing & Validation ---
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("TOOL.SyncFiles: expected 2 or 3 arguments (direction, localDir, [filterPattern]), got %d", len(args))
	}
	direction, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.SyncFiles: direction argument must be a string, got %T", args[0])
	}
	localDir, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.SyncFiles: localDir argument must be a string, got %T", args[1])
	}
	if localDir == "" {
		return nil, fmt.Errorf("TOOL.SyncFiles: localDir cannot be empty")
	}

	var filterPattern string
	if len(args) == 3 {
		filterPattern, ok = args[2].(string)
		if !ok {
			if args[2] != nil {
				return nil, fmt.Errorf("TOOL.SyncFiles: filterPattern argument must be a string or null, got %T", args[2])
			}
			// nil is fine, filterPattern remains ""
		}
	}

	// Validate direction (only "up" supported initially)
	direction = strings.ToLower(direction)
	if direction != "up" {
		return nil, fmt.Errorf("TOOL.SyncFiles: direction '%s' not supported. Only 'up' is currently implemented", direction)
	}

	// --- Path Security Validation for localDir ---
	absLocalDir, secErr := SecureFilePath(localDir, interpreter.sandboxDir)
	if secErr != nil {
		return nil, fmt.Errorf("TOOL.SyncFiles: invalid localDir '%s': %w", localDir, errors.Join(ErrValidationArgValue, secErr))
	}
	// Check if it's actually a directory
	dirInfo, statErr := os.Stat(absLocalDir)
	if statErr != nil {
		return nil, fmt.Errorf("TOOL.SyncFiles: cannot access localDir '%s': %w", localDir, statErr)
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("TOOL.SyncFiles: localDir '%s' is not a directory", localDir)
	}
	interpreter.logger.Printf("[TOOL SyncFiles] Validated local directory: %s (Direction: %s, Filter: '%s')", absLocalDir, direction, filterPattern)

	// --- Stats Counters ---
	stats := map[string]int64{
		"files_scanned":    0,
		"files_filtered":   0, // Count files skipped due to filter
		"files_uploaded":   0,
		"files_deleted":    0,
		"files_up_to_date": 0,
		"upload_errors":    0,
		"delete_errors":    0,
		"list_errors":      0, // Errors listing API files
		"walk_errors":      0, // Errors during local walk
		"hash_errors":      0, // Errors calculating local hash
	}

	// --- Core Logic (Direction="up") ---

	// 1. List API Files
	interpreter.logger.Println("[TOOL SyncFiles] Listing current API files...")
	apiFilesResult, listErr := toolListAPIFiles(interpreter, []interface{}{})
	if listErr != nil {
		stats["list_errors"]++
		interpreter.logger.Printf("[ERROR TOOL SyncFiles] Failed to list API files: %v. Aborting sync.", listErr)
		// Return stats map with error? Or just the error? Let's return the error directly.
		return nil, fmt.Errorf("TOOL.SyncFiles: failed to list API files: %w", listErr)
	}

	// Create map for efficient lookup: displayName -> apiFileInfo map
	apiFilesMap := make(map[string]map[string]interface{})
	if apiFilesList, ok := apiFilesResult.([]map[string]interface{}); ok {
		for _, fileInfo := range apiFilesList {
			// Assuming displayName uniquely identifies the file for sync purposes
			if dispName, ok := fileInfo["displayName"].(string); ok && dispName != "" {
				// Filter out files that may have failed previously or are not ACTIVE?
				// Let's sync based on display name regardless of current state for now.
				apiFilesMap[dispName] = fileInfo
			} else {
				// Log files without display names? They can't be synced reliably by this method.
				if name, ok := fileInfo["name"].(string); ok {
					interpreter.logger.Printf("[WARN TOOL SyncFiles] API file %s has no display name, cannot sync.", name)
				}
			}
		}
	} else {
		// This shouldn't happen if toolListAPIFiles worked correctly
		return nil, fmt.Errorf("TOOL.SyncFiles: internal error - toolListAPIFiles returned unexpected type %T", apiFilesResult)
	}
	interpreter.logger.Printf("[TOOL SyncFiles] Found %d API files with display names.", len(apiFilesMap))

	// 2. Walk Local Directory
	interpreter.logger.Printf("[TOOL SyncFiles] Walking local directory: %s", absLocalDir)
	localFilesSeen := make(map[string]bool) // track relative paths seen locally

	walkErr := filepath.WalkDir(absLocalDir, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			stats["walk_errors"]++
			interpreter.logger.Printf("[ERROR TOOL SyncFiles] Error accessing %q during walk: %v", currentPath, err)
			// Decide whether to skip file or abort walk? Let's try skipping.
			return nil // Continue walk if possible
		}

		// Skip root directory itself
		if currentPath == absLocalDir {
			return nil
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		stats["files_scanned"]++

		// Calculate relative path (use this as the key/display name)
		relPath, relErr := filepath.Rel(absLocalDir, currentPath)
		if relErr != nil {
			stats["walk_errors"]++
			interpreter.logger.Printf("[ERROR TOOL SyncFiles] Cannot get relative path for %s: %v", currentPath, relErr)
			return nil // Skip this file
		}
		relPath = filepath.ToSlash(relPath)
		localFilesSeen[relPath] = true

		// Apply filter pattern
		if filterPattern != "" {
			match, matchErr := filepath.Match(filterPattern, filepath.Base(currentPath)) // Match base name
			if matchErr != nil {
				stats["walk_errors"]++
				interpreter.logger.Printf("[ERROR TOOL SyncFiles] Invalid filter pattern '%s': %v", filterPattern, matchErr)
				return fmt.Errorf("invalid filter pattern: %w", matchErr) // Abort walk on bad pattern
			}
			if !match {
				stats["files_filtered"]++
				interpreter.logger.Printf("[DEBUG TOOL SyncFiles] Filtered out: %s", relPath)
				return nil // Skip file
			}
		}

		// Calculate local hash
		localHash, hashErr := calculateFileHash(currentPath)
		if hashErr != nil {
			stats["hash_errors"]++
			interpreter.logger.Printf("[ERROR TOOL SyncFiles] Cannot calculate hash for %s: %v", relPath, hashErr)
			return nil // Skip this file
		}

		// Compare with API file map
		apiFileInfo, existsInAPI := apiFilesMap[relPath]
		needsUpload := false
		if !existsInAPI {
			needsUpload = true
			interpreter.logger.Printf("[TOOL SyncFiles] File needs upload (new): %s", relPath)
		} else {
			// Compare hash if API file info has it
			apiHash, hasApiHash := apiFileInfo["sha256Hash"].(string)
			if !hasApiHash || apiHash == "" {
				interpreter.logger.Printf("[WARN TOOL SyncFiles] API file %s (%s) has no hash, assuming change.", relPath, apiFileInfo["name"])
				needsUpload = true // Re-upload if API hash missing
			} else if apiHash != localHash {
				needsUpload = true
				interpreter.logger.Printf("[TOOL SyncFiles] File needs upload (hash mismatch): %s (Local: %s.., API: %s..)", relPath, localHash[:min(len(localHash), 8)], apiHash[:min(len(apiHash), 8)])
			} else {
				// Hashes match, file is up-to-date
				stats["files_up_to_date"]++
				interpreter.logger.Printf("[DEBUG TOOL SyncFiles] File up-to-date: %s", relPath)
			}
		}

		// Perform Upload if needed
		if needsUpload {
			// Call toolUploadFile internally, passing the *absolute* local path
			// and the *relative* path as the desired display name.
			uploadArgs := []interface{}{currentPath, relPath}
			_, uploadErr := toolUploadFile(interpreter, uploadArgs)
			if uploadErr != nil {
				stats["upload_errors"]++
				interpreter.logger.Printf("[ERROR TOOL SyncFiles] Upload failed for %s: %v", relPath, uploadErr)
				// Continue sync if one file fails
			} else {
				stats["files_uploaded"]++
				interpreter.logger.Printf("[TOOL SyncFiles] Upload successful for: %s", relPath)
			}
		}

		return nil // Continue walk
	}) // End WalkDir

	if walkErr != nil {
		// This error would be from a critical issue like invalid pattern
		interpreter.logger.Printf("[ERROR TOOL SyncFiles] Walk completed with critical error: %v", walkErr)
		// Return partial stats? Or just the error?
		return nil, fmt.Errorf("TOOL.SyncFiles: critical error during directory walk: %w", walkErr)
	}

	// 3. Delete API files not found locally
	interpreter.logger.Println("[TOOL SyncFiles] Checking for remotely deleted files...")
	for displayName, apiFileInfo := range apiFilesMap {
		if !localFilesSeen[displayName] {
			// Check if this file should be skipped by the filter pattern
			if filterPattern != "" {
				match, _ := filepath.Match(filterPattern, filepath.Base(displayName))
				// Ignore match errors here, just skip if doesn't match
				if !match {
					interpreter.logger.Printf("[DEBUG TOOL SyncFiles] Skipping remote delete for filtered file: %s", displayName)
					continue // Skip deletion if it would have been filtered locally
				}
			}

			apiFileName, nameOk := apiFileInfo["name"].(string)
			if !nameOk || apiFileName == "" {
				interpreter.logger.Printf("[WARN TOOL SyncFiles] Cannot delete remote file %s, missing API name.", displayName)
				continue
			}
			interpreter.logger.Printf("[TOOL SyncFiles] File needs deletion (removed locally): %s (API Name: %s)", displayName, apiFileName)
			// Call toolDeleteAPIFile internally
			_, deleteErr := toolDeleteAPIFile(interpreter, []interface{}{apiFileName})
			if deleteErr != nil {
				stats["delete_errors"]++
				interpreter.logger.Printf("[ERROR TOOL SyncFiles] Delete failed for %s: %v", apiFileName, deleteErr)
				// Continue sync
			} else {
				stats["files_deleted"]++
				interpreter.logger.Printf("[TOOL SyncFiles] Delete successful for: %s", apiFileName)
			}
		}
	}

	interpreter.logger.Printf("[TOOL SyncFiles] Sync ('up' direction) finished. Stats: %+v", stats)

	// Return the stats map
	return stats, nil
}

// --- Registration ---

func registerFileAPITools(registry *ToolRegistry) error {
	var err error

	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "ListAPIFiles",
			Description: "Lists files previously uploaded to the Gemini File API associated with the current API key.",
			Args:        []ArgSpec{},
			ReturnType:  ArgTypeAny, // Returns a list of maps -> Any
		},
		Func: toolListAPIFiles,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool ListAPIFiles: %w", err)
	}

	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "DeleteAPIFile",
			Description: "Deletes a file from the Gemini File API using its specific API name (e.g., 'files/abc123xyz').",
			Args: []ArgSpec{
				{Name: "api_file_name", Type: ArgTypeString, Required: true, Description: "The unique API name of the file to delete (e.g., 'files/abc123xyz')."},
			},
			ReturnType: ArgTypeString, // Returns success message or error details
		},
		Func: toolDeleteAPIFile,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool DeleteAPIFile: %w", err)
	}

	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name: "UploadFile",
			Description: "Uploads a local file to the Gemini File API. Waits for the file to become ACTIVE. " +
				"Handles zero-byte files correctly. Returns a map containing the API file details upon success.",
			Args: []ArgSpec{
				{Name: "local_path", Type: ArgTypeString, Required: true, Description: "Path to the local file within the sandbox."},
				{Name: "display_name", Type: ArgTypeString, Required: false, Description: "Optional display name for the file in the API. Defaults to relative path from sandbox root or basename if sandbox not set."},
			},
			ReturnType: ArgTypeAny, // Returns map -> Any
		},
		Func: toolUploadFile,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool UploadFile: %w", err)
	}

	// Register SyncFiles (NEW)
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name: "SyncFiles",
			Description: "Synchronizes files between a local directory and the Gemini File API. " +
				"Currently only supports 'up' direction (local -> API). Returns a map summarizing the operation.",
			Args: []ArgSpec{
				{Name: "direction", Type: ArgTypeString, Required: true, Description: "Sync direction. Currently only 'up' is supported."},
				{Name: "local_dir", Type: ArgTypeString, Required: true, Description: "Path to the local directory within the sandbox to sync."},
				{Name: "filter_pattern", Type: ArgTypeString, Required: false, Description: "Optional glob pattern (e.g., '*.go', 'docs/*') to filter files by basename during sync. If omitted, all files are synced."},
			},
			ReturnType: ArgTypeAny, // Returns map[string]int64 -> Any
		},
		Func: toolSyncFiles,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool SyncFiles: %w", err)
	}

	return nil
}
