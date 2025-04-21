// filename: pkg/core/tools_file_api.go
package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
)

// --- Constants, init, and Hash Helper ---
const emptyFileContentForHash = " "

var emptyFileHash string

func init() {
	hasher := sha256.New()
	hasher.Write([]byte(emptyFileContentForHash))
	emptyFileHash = hex.EncodeToString(hasher.Sum(nil))
}

// calculateFileHash calculates the SHA256 hash of a file content or returns a default hash for empty files.
func calculateFileHash(filePath string) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		// Handle file not existing specifically? For hashing, maybe return error.
		return "", fmt.Errorf("stat file %s: %w", filePath, err)
	}
	// Treat empty file as having a specific default hash
	if fileInfo.Size() == 0 {
		return emptyFileHash, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file %s for hashing: %w", filePath, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("read file %s for hashing: %w", filePath, err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// --- Helper: Check/Get GenAI Client ---
// checkGenAIClient retrieves the underlying genai.Client from the interpreter's LLMClient.
func checkGenAIClient(interpreter *Interpreter) (*genai.Client, error) {
	if interpreter == nil || interpreter.llmClient == nil || interpreter.llmClient.Client() == nil {
		// Return a more specific error indicating the client isn't ready/configured
		return nil, errors.New("genai client is not initialized (API key potentially missing or invalid)")
	}
	return interpreter.llmClient.Client(), nil
}

// --- Helper: Upload File and Poll ---
// HelperUploadAndPollFile handles the core logic of uploading a single file and waiting for it to be ACTIVE.
func HelperUploadAndPollFile(ctx context.Context, absLocalPath string, displayName string, client *genai.Client, logger *log.Logger) (*genai.File, error) {
	if client == nil {
		return nil, errors.New("genai client is nil")
	}
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	// Use specific logger levels if available, otherwise default logger
	debugLog := logger // Assume debug for now if only one logger passed
	infoLog := logger
	errorLog := logger

	debugLog.Printf("[API HELPER Upload] Processing: %s (Display: %s)", absLocalPath, displayName)

	fileInfo, err := os.Stat(absLocalPath)
	if err != nil {
		return nil, fmt.Errorf("stat file %s: %w", absLocalPath, err)
	}

	isZeroByte := fileInfo.Size() == 0
	mimeType := ""
	if isZeroByte {
		mimeType = "text/plain" // Treat zero-byte as plain text for upload
	} else {
		mimeType = mime.TypeByExtension(filepath.Ext(absLocalPath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		} // Default MIME type
	}
	debugLog.Printf("[API HELPER Upload] Determined MIME type: %s", mimeType)

	options := &genai.UploadFileOptions{MIMEType: mimeType, DisplayName: displayName}

	var reader io.Reader
	var fileHandle *os.File // Keep track to close it later if opened

	if isZeroByte {
		// Use predefined content for zero-byte files to avoid API issues
		reader = strings.NewReader(emptyFileContentForHash)
		debugLog.Printf("[API HELPER Upload] Handling zero-byte file: %s", absLocalPath)
	} else {
		fileHandle, err = os.Open(absLocalPath)
		if err != nil {
			return nil, fmt.Errorf("open file %s for upload: %w", absLocalPath, err)
		}
		defer fileHandle.Close() // Ensure file handle is closed
		reader = fileHandle
	}

	if ctx == nil {
		ctx = context.Background()
	} // Ensure context is not nil

	infoLog.Printf("[API HELPER Upload] Starting upload for %s (Display: %s)...", absLocalPath, displayName)
	apiFile, err := client.UploadFile(ctx, "", reader, options) // Use empty string for name, API generates it
	if err != nil {
		return nil, fmt.Errorf("api upload failed for %q (Display: %s): %w", absLocalPath, displayName, err)
	}

	debugLog.Printf("[API HELPER Upload] Upload initiated -> API Name: %s (URI: %s)", apiFile.Name, apiFile.URI)

	// --- Polling Logic ---
	startTime := time.Now()
	pollInterval := 1 * time.Second
	const maxPollInterval = 10 * time.Second
	const timeout = 2 * time.Minute // Maybe make configurable?

	for apiFile.State == genai.FileStateProcessing {
		if time.Since(startTime) > timeout {
			errMsg := fmt.Sprintf("polling timeout for file %s (API Name: %s, Display: %s)", absLocalPath, apiFile.Name, displayName)
			errorLog.Printf("[ERROR API HELPER Upload] %s. Attempting to delete orphaned API file.", errMsg)
			// Attempt cleanup, ignore error as we're already in an error state
			_ = client.DeleteFile(context.Background(), apiFile.Name)
			return nil, fmt.Errorf(errMsg) // Return timeout error
		}

		time.Sleep(pollInterval)
		debugLog.Printf("[DEBUG API HELPER Upload] Polling status for API file %s...", apiFile.Name)
		getCtx, cancelGet := context.WithTimeout(context.Background(), 30*time.Second) // Timeout for GetFile call itself
		updatedFile, getErr := client.GetFile(getCtx, apiFile.Name)
		cancelGet()

		if getErr != nil {
			// Don't return immediately on transient Get error, maybe log and retry?
			// But if it persists, it could indicate a problem.
			errorLog.Printf("[WARN API HELPER Upload] Error getting status for %s (will retry): %v", apiFile.Name, getErr)
			// Continue loop, maybe slow down polling slightly?
			time.Sleep(pollInterval) // Add extra delay on error
			continue
		}
		apiFile = updatedFile // Update status

		// Exponential backoff for polling interval
		pollInterval *= 2
		if pollInterval > maxPollInterval {
			pollInterval = maxPollInterval
		}
		debugLog.Printf("[DEBUG API HELPER Upload] Poll %s successful, state: %s (Next poll in %v)", apiFile.Name, apiFile.State, pollInterval)
	}

	if apiFile.State != genai.FileStateActive {
		errMsg := fmt.Sprintf("file processing failed for %s (API Name: %s, Display: %s). Final State: %s", absLocalPath, apiFile.Name, displayName, apiFile.State)
		errorLog.Printf("[ERROR API HELPER Upload] %s. Attempting to delete failed API file.", errMsg)
		// Attempt cleanup
		_ = client.DeleteFile(context.Background(), apiFile.Name)
		return nil, fmt.Errorf(errMsg)
	}

	infoLog.Printf("[API HELPER Upload] Upload successful and ACTIVE: %s -> %s", displayName, apiFile.Name)
	return apiFile, nil
}

// --- Helper: List API Files Helper ---
// HelperListApiFiles fetches all files from the API.
func HelperListApiFiles(ctx context.Context, client *genai.Client, logger *log.Logger) ([]*genai.File, error) {
	if client == nil {
		return nil, errors.New("genai client is nil")
	}
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	logger.Println("[API HELPER List] Fetching file list from API...")
	if ctx == nil {
		ctx = context.Background()
	}

	iter := client.ListFiles(ctx)
	results := []*genai.File{}
	fetchErrors := 0
	fileCount := 0

	for {
		file, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			errMsg := fmt.Sprintf("Error fetching file list page: %v", err)
			logger.Printf("[API HELPER List] %s", errMsg)
			fetchErrors++
			// Optionally break or implement retry logic here? For now, continue.
			continue
		}
		results = append(results, file)
		fileCount++
	}

	logger.Printf("[API HELPER List] Found %d files. Encountered %d errors during fetch.", fileCount, fetchErrors)
	if fetchErrors > 0 {
		// Return partial list and an error
		return results, fmt.Errorf("encountered %d errors fetching file list", fetchErrors)
	}
	return results, nil
}

// --- Tool: ListAPIFiles (Wrapper) ---
// toolListAPIFiles provides the ListAPIFiles tool implementation.
func toolListAPIFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.ListAPIFiles: %w", clientErr)
	}

	// Call the helper
	apiFiles, err := HelperListApiFiles(context.Background(), client, interpreter.logger)
	if err != nil {
		interpreter.logger.Printf("[TOOL ListAPIFiles] Warning: Error from helper (returning partial list if any): %v", err)
		// Continue processing any files that *were* retrieved
	}

	// Convert []*genai.File to []map[string]interface{} for NeuroScript
	results := []map[string]interface{}{}
	for _, file := range apiFiles {
		// Ensure required fields are present, handle potential nil times
		createTimeStr := ""
		if !file.CreateTime.IsZero() {
			createTimeStr = file.CreateTime.Format(time.RFC3339)
		}
		updateTimeStr := ""
		if !file.UpdateTime.IsZero() {
			updateTimeStr = file.UpdateTime.Format(time.RFC3339)
		}

		fileInfo := map[string]interface{}{
			"name":        file.Name,
			"displayName": file.DisplayName,
			"mimeType":    file.MIMEType,
			"sizeBytes":   file.SizeBytes,
			"createTime":  createTimeStr,
			"updateTime":  updateTimeStr,
			"state":       string(file.State),
			"uri":         file.URI,
			"sha256Hash":  hex.EncodeToString(file.Sha256Hash),
		}
		results = append(results, fileInfo)
	}

	// Return results map, pass through any error from the helper
	return map[string]interface{}{"files": results}, err
}

// --- Tool: DeleteAPIFile ---
// toolDeleteAPIFile provides the DeleteAPIFile tool implementation.
func toolDeleteAPIFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.DeleteAPIFile: %w", clientErr)
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.DeleteAPIFile: expected 1 arg (api_file_name), got %d", len(args))
	}
	apiFileName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.DeleteAPIFile: arg must be string, got %T", args[0])
	}
	if apiFileName == "" {
		return nil, errors.New("TOOL.DeleteAPIFile: API file name cannot be empty")
	}

	interpreter.logger.Printf("[TOOL DeleteAPIFile] Attempting delete: %s", apiFileName)
	err := client.DeleteFile(context.Background(), apiFileName)
	if err != nil {
		errMsg := fmt.Sprintf("Failed delete %s: %v", apiFileName, err)
		interpreter.logger.Printf("[TOOL DeleteAPIFile] Error: %s", errMsg)
		// Return error message *and* Go error for interpreter
		return map[string]interface{}{"error": errMsg}, fmt.Errorf("TOOL.DeleteAPIFile: %w", err)
	}
	successMsg := fmt.Sprintf("Successfully deleted: %s", apiFileName)
	interpreter.logger.Printf("[TOOL DeleteAPIFile] %s", successMsg)
	return map[string]interface{}{"status": "success", "message": successMsg}, nil
}

// --- Tool: UploadFile (Wrapper) ---
// toolUploadFile provides the UploadFile tool implementation.
func toolUploadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.UploadFile: %w", clientErr)
	}
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("TOOL.UploadFile: expected 1-2 args (local_path, [display_name]), got %d", len(args))
	}

	localPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.UploadFile: local_path must be string, got %T", args[0])
	}
	if localPath == "" {
		return nil, errors.New("TOOL.UploadFile: local_path cannot be empty")
	}

	var displayName string
	if len(args) == 2 {
		displayName, ok = args[1].(string)
		if !ok && args[1] != nil { // Allow null/nil to signify default
			return nil, fmt.Errorf("TOOL.UploadFile: display_name must be string or null, got %T", args[1])
		}
	}

	// Use ResolveAndSecurePath for validation, as path comes from script relative to CWD
	securePath, secErr := ResolveAndSecurePath(localPath, interpreter.sandboxDir)
	if secErr != nil {
		return nil, fmt.Errorf("TOOL.UploadFile: invalid path %q: %w", localPath, errors.Join(ErrValidationArgValue, secErr))
	}
	interpreter.logger.Printf("[TOOL UploadFile] Validated path: %s -> %s", localPath, securePath)

	// Default display name logic (if not provided or empty)
	if displayName == "" {
		// Calculate path relative to sandbox root for default display name
		relPath, err := filepath.Rel(interpreter.sandboxDir, securePath) // Use absolute sandbox path
		if err == nil {
			displayName = filepath.ToSlash(relPath) // Use clean relative path
		} else {
			// Fallback to basename if Rel fails (shouldn't happen if validation passed)
			displayName = filepath.Base(securePath)
			interpreter.logger.Printf("[WARN TOOL UploadFile] Could not get relative path for default display name, using basename: %v", err)
		}
		interpreter.logger.Printf("[TOOL UploadFile] Using default display name: %s", displayName)
	}

	// Call the helper
	apiFile, uploadErr := HelperUploadAndPollFile(context.Background(), securePath, displayName, client, interpreter.logger)
	if uploadErr != nil {
		return nil, fmt.Errorf("TOOL.UploadFile: %w", uploadErr)
	} // Wrap error

	// Return file info map on success
	createTimeStr := ""
	if !apiFile.CreateTime.IsZero() {
		createTimeStr = apiFile.CreateTime.Format(time.RFC3339)
	}
	updateTimeStr := ""
	if !apiFile.UpdateTime.IsZero() {
		updateTimeStr = apiFile.UpdateTime.Format(time.RFC3339)
	}
	resultMap := map[string]interface{}{
		"name": apiFile.Name, "displayName": apiFile.DisplayName, "mimeType": apiFile.MIMEType,
		"sizeBytes": apiFile.SizeBytes, "createTime": createTimeStr, "updateTime": updateTimeStr,
		"state": string(apiFile.State), "uri": apiFile.URI, "sha256Hash": hex.EncodeToString(apiFile.Sha256Hash),
	}
	return resultMap, nil
}

// --- Tool: SyncFiles (Wrapper) ---
// toolSyncFiles provides the SyncFiles tool implementation.
// FIXED: Ensure call is to SyncDirectoryUpHelper and uses ResolveAndSecurePath.
func toolSyncFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.SyncFiles: %w", clientErr)
	}

	// --- Argument Parsing & Validation ---
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
	}
	var filterPattern string
	if len(args) >= 3 {
		filterPattern, ok = args[2].(string)
		if !ok && args[2] != nil {
			return nil, fmt.Errorf("TOOL.SyncFiles: filter_pattern must be string or null")
		}
	}
	var ignoreGitignore bool = false
	if len(args) == 4 {
		ignoreGitignore, ok = args[3].(bool)
		if !ok {
			return nil, fmt.Errorf("TOOL.SyncFiles: ignore_gitignore must be boolean")
		}
	}
	direction = strings.ToLower(direction)
	if direction != "up" {
		return nil, fmt.Errorf("TOOL.SyncFiles: direction '%s' not supported", direction)
	}

	// Use ResolveAndSecurePath for validation, as localDir comes from script (CWD-relative)
	absLocalDir, secErr := ResolveAndSecurePath(localDir, interpreter.sandboxDir)
	if secErr != nil {
		return nil, fmt.Errorf("TOOL.SyncFiles: invalid local_dir '%s': %w", localDir, errors.Join(ErrValidationArgValue, secErr))
	}
	dirInfo, statErr := os.Stat(absLocalDir)
	if statErr != nil {
		return nil, fmt.Errorf("TOOL.SyncFiles: cannot access local_dir '%s': %w", localDir, statErr)
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("TOOL.SyncFiles: local_dir '%s' is not a directory", localDir)
	}
	interpreter.logger.Printf("[TOOL SyncFiles] Validated dir: %s (Ignore .gitignore: %t)", absLocalDir, ignoreGitignore)
	// --- End Argument Parsing ---

	// *** FIXED: Call the correct helper function SyncDirectoryUpHelper ***
	// Ensure SyncDirectoryUpHelper is accessible (defined in tool_file_api_sync.go)
	statsMap, syncErr := SyncDirectoryUpHelper(
		context.Background(), // Use background context for tool's sync operation
		absLocalDir,
		filterPattern,
		ignoreGitignore,
		client,
		interpreter.logger, // Pass interpreter's logger for all levels
		interpreter.logger,
		interpreter.logger,
	)

	// Return stats map and any error from the sync operation
	// The NeuroScript side will handle interpreting the map/error.
	return statsMap, syncErr
}

// --- Registration ---
// registerFileAPITools registers all file API related tools.
func registerFileAPITools(registry *ToolRegistry) error {
	var err error
	// Ensure specifications match the implementation arguments
	tools := []ToolImplementation{
		{Spec: ToolSpec{
			Name: "ListAPIFiles", Description: "Lists files previously uploaded to the API.", Args: []ArgSpec{}, ReturnType: ArgTypeAny, // Returns map {"files": [...]}
		}, Func: toolListAPIFiles},
		{Spec: ToolSpec{
			Name: "DeleteAPIFile", Description: "Deletes a file from the API by its name (e.g., 'files/abc123xyz').", Args: []ArgSpec{{Name: "api_file_name", Type: ArgTypeString, Required: true, Description: "The full API name of the file (e.g., files/xyz)."}}, ReturnType: ArgTypeAny, // Returns map {"status": "success", "message": ...} or {"error": ...}
		}, Func: toolDeleteAPIFile},
		{Spec: ToolSpec{
			Name: "UploadFile", Description: "Uploads a local file (relative to sandbox) to the API.", Args: []ArgSpec{{Name: "local_path", Type: ArgTypeString, Required: true, Description: "Relative path to the local file."}, {Name: "display_name", Type: ArgTypeString, Required: false, Description: "Optional display name (defaults to relative path)."}}, ReturnType: ArgTypeAny, // Returns map of file info
		}, Func: toolUploadFile},
		{Spec: ToolSpec{
			Name: "SyncFiles", Description: "Syncs local directory (relative to sandbox) to API ('up' only).", Args: []ArgSpec{{Name: "direction", Type: ArgTypeString, Required: true, Description: "Sync direction ('up')."}, {Name: "local_dir", Type: ArgTypeString, Required: true, Description: "Relative path to local directory."}, {Name: "filter_pattern", Type: ArgTypeString, Required: false, Description: "Optional filename glob pattern."}, {Name: "ignore_gitignore", Type: ArgTypeBool, Required: false, Description: "Ignore .gitignore files if true (default: false)."}}, ReturnType: ArgTypeAny, // Returns map of sync statistics
		}, Func: toolSyncFiles},
	}
	for _, tool := range tools {
		if err = registry.RegisterTool(tool); err != nil {
			// Log the specific tool that failed registration
			log.Printf("Error registering tool %s: %v", tool.Spec.Name, err)
			return fmt.Errorf("failed register tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil
}
