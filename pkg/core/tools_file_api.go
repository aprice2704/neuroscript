// filename: pkg/core/tools_file_api.go
package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log" // Keep standard mime for extension fallback if needed
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype" // Added import for the library
	"github.com/google/generative-ai-go/genai"
)

// --- Constants, init, and Hash Helper ---
const (
	emptyFileContentForHash = " "
)

var (
	emptyFileHash string
	// ErrSkippedBinaryFile should be defined in errors.go
	// Example: var ErrSkippedBinaryFile = errors.New("skipped potentially binary file")
)

func init() {
	hasher := sha256.New()
	hasher.Write([]byte(emptyFileContentForHash))
	emptyFileHash = hex.EncodeToString(hasher.Sum(nil))
	// Optional: Configure mimetype library if needed (e.g., SetLimit)
	// mimetype.SetLimit(0) // Example: remove read limit
}

// --- Helper: Upload File and Poll ---
// HelperUploadAndPollFile handles the core logic of uploading a single file and waiting for it to be ACTIVE.
// MODIFIED: Use mimetype library, force text/plain for non-Go text files, skip binary
// MODIFIED: Use strings.NewReader("") for zero-byte files
// MODIFIED: Upload .go files as text/plain
func HelperUploadAndPollFile(ctx context.Context, absLocalPath string, displayName string, client *genai.Client, logger *log.Logger) (*genai.File, error) {
	if client == nil {
		return nil, errors.New("genai client is nil")
	}
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	debugLog := logger
	infoLog := logger
	errorLog := logger

	debugLog.Printf("[API HELPER Upload] Processing: %s (Display: %s)", absLocalPath, displayName)

	fileInfo, err := os.Stat(absLocalPath)
	if err != nil {
		return nil, fmt.Errorf("stat file %s: %w", absLocalPath, err)
	}

	isZeroByte := fileInfo.Size() == 0
	fileExt := filepath.Ext(absLocalPath)
	uploadMimeType := "" // The MIME type we will actually use for the upload

	// --- MIME Type Detection and Handling ---
	if isZeroByte {
		uploadMimeType = "text/plain"
		debugLog.Printf("[API HELPER Upload] Handling zero-byte file: %s as text/plain", absLocalPath)
		// *** CHANGE GO FILE HANDLING: Upload as text/plain ***
	} else if fileExt == ".go" {
		// Treat Go files as plain text for GenerateContent compatibility
		uploadMimeType = "text/plain"
		debugLog.Printf("[API HELPER Upload] Detected Go file, using MIME type: text/plain for upload compatibility for: %s", absLocalPath)
		// *** END CHANGE ***
	} else {
		// Use mimetype library for non-Go, non-empty files
		debugLog.Printf("[API HELPER Upload] Detecting MIME type for non-Go file: %s", absLocalPath)
		detectedMime, detectErr := mimetype.DetectFile(absLocalPath)
		if detectErr != nil {
			errorLog.Printf("[ERROR API HELPER Upload] MIME detection failed for %s: %v", absLocalPath, detectErr)
			return nil, fmt.Errorf("mime detection failed for %s: %w", absLocalPath, detectErr)
		}

		detectedMimeStr := detectedMime.String()
		debugLog.Printf("[API HELPER Upload] Detected MIME by library: %s for %s", detectedMimeStr, absLocalPath)

		// Check if the detected type is text-based by checking parent hierarchy
		isText := false
		for m := detectedMime; m != nil; m = m.Parent() {
			if m.Is("text/plain") {
				isText = true
				break
			}
		}

		if !isText {
			// Skip potentially binary files
			warnMsg := fmt.Sprintf("Skipping potentially binary file (detected: %s): %s", detectedMimeStr, absLocalPath)
			infoLog.Printf("[WARN API HELPER Upload] %s", warnMsg)
			// NOTE: Using temporary error string until ErrSkippedBinaryFile is properly defined/used
			return nil, fmt.Errorf("skipped potentially binary file (detected: %s): %s", detectedMimeStr, absLocalPath)
		} else {
			// Force text/plain for all other non-Go, text-based files
			uploadMimeType = "text/plain"
			debugLog.Printf("[API HELPER Upload] Non-Go text file detected (Library: %s). Forcing upload MIME type to text/plain for: %s", detectedMimeStr, absLocalPath)
		}
	}
	// --- End MIME Type Logic ---

	if uploadMimeType == "" {
		// Should not happen if logic above is correct, but as a safeguard
		errorLog.Printf("[ERROR API HELPER Upload] Failed to determine upload MIME type for %s", absLocalPath)
		return nil, fmt.Errorf("failed to determine upload MIME type for %s", absLocalPath)
	}

	debugLog.Printf("[API HELPER Upload] Final determined upload MIME type: %s for %s", uploadMimeType, absLocalPath)

	options := &genai.UploadFileOptions{MIMEType: uploadMimeType, DisplayName: displayName}

	// Handle reader for zero-byte files explicitly
	var reader io.Reader
	if isZeroByte {
		reader = strings.NewReader("") // Use an empty string reader for zero-byte files
		debugLog.Printf("[API HELPER Upload] Using empty string reader for zero-byte file: %s", absLocalPath)
	} else {
		// Use actual file content for non-zero-byte files.
		fileHandle, err := os.Open(absLocalPath)
		if err != nil {
			return nil, fmt.Errorf("open file %s for upload: %w", absLocalPath, err)
		}
		defer fileHandle.Close() // Close only if opened
		reader = fileHandle      // Use the file handle as the reader
	}

	if ctx == nil {
		ctx = context.Background()
	}

	infoLog.Printf("[API HELPER Upload] Starting upload for %s (Display: %s, MIME: %s)...", absLocalPath, displayName, uploadMimeType)
	// The 'reader' variable now correctly holds either the fileHandle or the empty string reader
	apiFile, err := client.UploadFile(ctx, "", reader, options)
	if err != nil {
		if strings.Contains(err.Error(), "Unsupported MIME type") {
			errorLog.Printf("[ERROR API HELPER Upload] API rejected upload MIME type %q for file %q (Display: %s)", uploadMimeType, absLocalPath, displayName)
		}
		return nil, fmt.Errorf("api upload failed for %q (Display: %s): %w", absLocalPath, displayName, err)
	}

	debugLog.Printf("[API HELPER Upload] Upload initiated -> API Name: %s (URI: %s)", apiFile.Name, apiFile.URI)

	// --- Polling Logic (Unchanged) ---
	startTime := time.Now()
	pollInterval := 1 * time.Second
	const maxPollInterval = 10 * time.Second
	const timeout = 2 * time.Minute

	for apiFile.State == genai.FileStateProcessing {
		if time.Since(startTime) > timeout {
			errMsg := fmt.Sprintf("polling timeout for file %s (API Name: %s, Display: %s)", absLocalPath, apiFile.Name, displayName)
			errorLog.Printf("[ERROR API HELPER Upload] %s. Attempting to delete orphaned API file.", errMsg)
			_ = client.DeleteFile(context.Background(), apiFile.Name)
			return nil, fmt.Errorf(errMsg)
		}
		time.Sleep(pollInterval)
		debugLog.Printf("[DEBUG API HELPER Upload] Polling status for API file %s...", apiFile.Name)
		getCtx, cancelGet := context.WithTimeout(context.Background(), 30*time.Second)
		updatedFile, getErr := client.GetFile(getCtx, apiFile.Name)
		cancelGet()
		if getErr != nil {
			errorLog.Printf("[WARN API HELPER Upload] Error getting status for %s (will retry): %v", apiFile.Name, getErr)
			// Consider adding backoff or error threshold here
			time.Sleep(pollInterval) // Basic retry delay
			continue
		}
		apiFile = updatedFile
		pollInterval *= 2
		if pollInterval > maxPollInterval {
			pollInterval = maxPollInterval
		}
		debugLog.Printf("[DEBUG API HELPER Upload] Poll %s successful, state: %s (Next poll in %v)", apiFile.Name, apiFile.State, pollInterval)
	}

	if apiFile.State != genai.FileStateActive {
		errMsg := fmt.Sprintf("file processing failed for %s (API Name: %s, Display: %s). Final State: %s", absLocalPath, apiFile.Name, displayName, apiFile.State)
		errorLog.Printf("[ERROR API HELPER Upload] %s. Attempting to delete failed API file.", errMsg)
		_ = client.DeleteFile(context.Background(), apiFile.Name)
		return nil, fmt.Errorf(errMsg)
	}

	infoLog.Printf("[API HELPER Upload] Upload successful and ACTIVE: %s -> %s", displayName, apiFile.Name)
	return apiFile, nil
}

// --- Tool: ListAPIFiles (Wrapper) ---
func toolListAPIFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.ListAPIFiles: %w", clientErr)
	}
	// Assumes HelperListApiFiles exists (potentially in sync_helpers.go or sync_morehelpers.go)
	apiFiles, err := HelperListApiFiles(context.Background(), client, interpreter.logger)
	if err != nil {
		interpreter.logger.Printf("[TOOL ListAPIFiles] Warning: Error from helper (returning partial list if any): %v", err)
		// Continue processing any files that were returned before the error
	}

	results := []map[string]interface{}{}
	for _, file := range apiFiles {
		// Check for nil file pointer, although ListFiles usually doesn't return nils
		if file == nil {
			continue
		}
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
			"sha256Hash":  "", // Default to empty
		}
		if len(file.Sha256Hash) > 0 {
			fileInfo["sha256Hash"] = hex.EncodeToString(file.Sha256Hash)
		}
		results = append(results, fileInfo)
	}
	// Return the original listing error if one occurred
	return map[string]interface{}{"files": results}, err
}

// --- Tool: UploadFile (Wrapper) ---
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
		if !ok && args[1] != nil { // Allow nil for display name
			return nil, fmt.Errorf("TOOL.UploadFile: display_name must be string or null, got %T", args[1])
		}
	}

	// Assumes ResolveAndSecurePath is defined elsewhere (likely security_helpers.go or tools_helpers.go)
	securePath, secErr := ResolveAndSecurePath(localPath, interpreter.sandboxDir)
	if secErr != nil {
		// Use errors.Is for specific checks if ResolveAndSecurePath returns wrapped errors
		return nil, fmt.Errorf("TOOL.UploadFile: invalid path %q: %w", localPath, secErr)
	}
	interpreter.logger.Printf("[TOOL UploadFile] Validated path: %s -> %s", localPath, securePath)

	if displayName == "" {
		relPath, err := filepath.Rel(interpreter.sandboxDir, securePath)
		if err == nil {
			displayName = filepath.ToSlash(relPath)
		} else {
			displayName = filepath.Base(securePath)
			interpreter.logger.Printf("[WARN TOOL UploadFile] Could not get relative path for default display name, using basename: %v", err)
		}
		interpreter.logger.Printf("[TOOL UploadFile] Using default display name: %s", displayName)
	}

	apiFile, uploadErr := HelperUploadAndPollFile(context.Background(), securePath, displayName, client, interpreter.logger)
	if uploadErr != nil {
		// Check for the specific binary skip error string (temporary)
		// Replace with errors.Is(uploadErr, ErrSkippedBinaryFile) once defined and used in helper
		if strings.HasPrefix(uploadErr.Error(), "skipped potentially binary file") {
			reason := uploadErr.Error()
			return map[string]interface{}{"status": "skipped", "reason": reason, "path": localPath}, nil
		}
		// Wrap other errors
		return nil, fmt.Errorf("TOOL.UploadFile: %w", uploadErr)
	}

	// Check for nil API file just in case helper returns nil without error
	if apiFile == nil {
		return nil, errors.New("TOOL.UploadFile: helper returned nil file without error")
	}

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
		"name":        apiFile.Name,
		"displayName": apiFile.DisplayName,
		"mimeType":    apiFile.MIMEType,
		"sizeBytes":   apiFile.SizeBytes,
		"createTime":  createTimeStr,
		"updateTime":  updateTimeStr,
		"state":       string(apiFile.State),
		"uri":         apiFile.URI,
		"sha256Hash":  "",
	}
	if len(apiFile.Sha256Hash) > 0 {
		resultMap["sha256Hash"] = hex.EncodeToString(apiFile.Sha256Hash)
	}
	return resultMap, nil
}

// --- Registration ---
// Assumes ToolRegistry, ToolImplementation, ArgSpec, ArgTypeAny etc. are defined elsewhere
// Assumes toolDeleteAPIFile, toolSyncFiles, checkGenAIClient, HelperListApiFiles are defined elsewhere
func registerFileAPITools(registry *ToolRegistry) error {
	var err error
	tools := []ToolImplementation{
		{Spec: ToolSpec{Name: "ListAPIFiles", Description: "Lists files previously uploaded to the API.", Args: []ArgSpec{}, ReturnType: ArgTypeAny}, Func: toolListAPIFiles},
		{Spec: ToolSpec{Name: "DeleteAPIFile", Description: "Deletes a file from the API by its name (e.g., 'files/abc123xyz').", Args: []ArgSpec{{Name: "api_file_name", Type: ArgTypeString, Required: true, Description: "The full API name of the file (e.g., files/xyz)."}}, ReturnType: ArgTypeAny}, Func: toolDeleteAPIFile},
		{Spec: ToolSpec{Name: "UploadFile", Description: "Uploads a local file (relative to sandbox) to the API.", Args: []ArgSpec{{Name: "local_path", Type: ArgTypeString, Required: true, Description: "Relative path to the local file."}, {Name: "display_name", Type: ArgTypeString, Required: false, Description: "Optional display name (defaults to relative path)."}}, ReturnType: ArgTypeAny}, Func: toolUploadFile},
		{Spec: ToolSpec{Name: "SyncFiles", Description: "Syncs local directory (relative to sandbox) to API ('up' only).", Args: []ArgSpec{{Name: "direction", Type: ArgTypeString, Required: true, Description: "Sync direction ('up')."}, {Name: "local_dir", Type: ArgTypeString, Required: true, Description: "Relative path to local directory."}, {Name: "filter_pattern", Type: ArgTypeString, Required: false, Description: "Optional filename glob pattern."}, {Name: "ignore_gitignore", Type: ArgTypeBool, Required: false, Description: "Ignore .gitignore files if true (default: false)."}}, ReturnType: ArgTypeAny}, Func: toolSyncFiles},
	}
	for _, tool := range tools {
		if err = registry.RegisterTool(tool); err != nil {
			log.Printf("Error registering tool %s: %v", tool.Spec.Name, err)
			// Consider returning the error immediately or collecting all errors
			return fmt.Errorf("failed register tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil
}

// Helper function (ensure defined elsewhere, e.g., tools_helpers.go or security_helpers.go)
// func checkGenAIClient(interpreter *Interpreter) (*genai.Client, error) {
// 	if interpreter == nil || interpreter.llmClient == nil || interpreter.llmClient.Client() == nil {
// 		return nil, errors.New("GenAI client not initialized")
// 	}
// 	return interpreter.llmClient.Client(), nil
// }

// Assumed functions (ensure defined elsewhere):
// - func HelperListApiFiles(ctx context.Context, client *genai.Client, logger *log.Logger) ([]*genai.File, error) // Likely in sync_helpers.go or sync_morehelpers.go
// - func toolDeleteAPIFile(interpreter *Interpreter, args []interface{}) (interface{}, error) // Likely in this file or another tools_file_api_*.go
// - func toolSyncFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) // Likely in this file or another tools_file_api_*.go
// - func ResolveAndSecurePath(localPath string, sandboxDir string) (string, error) // Likely in security_helpers.go or tools_helpers.go
