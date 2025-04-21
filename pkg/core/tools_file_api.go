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
	"mime" // Keep standard mime for extension fallback if needed
	"os"
	"path/filepath"
	"strings"
	"time"

	// Removed unicode imports as binary check is replaced

	"github.com/gabriel-vasile/mimetype" // Added import for the library
	"github.com/google/generative-ai-go/genai"
)

// --- Constants, init, and Hash Helper ---
const (
	emptyFileContentForHash = " "
	// binaryCheckBufferSize removed, library handles reading internally
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

// isPotentiallyBinary function removed, replaced by mimetype library logic

// --- Helper: Upload File and Poll ---
// HelperUploadAndPollFile handles the core logic of uploading a single file and waiting for it to be ACTIVE.
// *** MODIFIED: Use mimetype library, force text/plain for non-Go text files, skip binary ***
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
	} else if fileExt == ".go" {
		// Prioritize specific type for Go files
		uploadMimeType = mime.TypeByExtension(fileExt)
		if uploadMimeType == "" {
			uploadMimeType = "text/x-go" // Fallback
		}
		debugLog.Printf("[API HELPER Upload] Detected Go file, using specific MIME type: %s", uploadMimeType)
	} else {
		// Use mimetype library for non-Go, non-empty files
		debugLog.Printf("[API HELPER Upload] Detecting MIME type for non-Go file: %s", absLocalPath)
		detectedMime, detectErr := mimetype.DetectFile(absLocalPath)
		if detectErr != nil {
			errorLog.Printf("[ERROR API HELPER Upload] MIME detection failed for %s: %v", absLocalPath, detectErr)
			// Fail upload if detection fails? Or default to octet-stream/skip? Let's fail.
			return nil, fmt.Errorf("mime detection failed for %s: %w", absLocalPath, detectErr)
		}

		detectedMimeStr := detectedMime.String()
		debugLog.Printf("[API HELPER Upload] Detected MIME by library: %s for %s", detectedMimeStr, absLocalPath)

		// Check if the detected type is text-based by checking parent hierarchy
		isText := false
		for m := detectedMime; m != nil; m = m.Parent() {
			// Use Is("text/plain") as the indicator for text-based as per library examples
			if m.Is("text/plain") {
				isText = true
				break
			}
		}

		if !isText {
			// Skip potentially binary files (those not having text/plain in hierarchy)
			warnMsg := fmt.Sprintf("Skipping potentially binary file (detected: %s): %s", detectedMimeStr, absLocalPath)
			infoLog.Printf("[WARN API HELPER Upload] %s", warnMsg)
			// NOTE: Ensure ErrSkippedBinaryFile is defined in pkg/core/errors.go and imported or defined locally
			// return nil, ErrSkippedBinaryFile // Use this once defined and imported
			return nil, fmt.Errorf("skipped potentially binary file (detected: %s): %s", detectedMimeStr, absLocalPath) // Temporary fallback
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

	// Use actual file content for upload.
	fileHandle, err := os.Open(absLocalPath)
	if err != nil {
		return nil, fmt.Errorf("open file %s for upload: %w", absLocalPath, err)
	}
	defer fileHandle.Close()
	reader := fileHandle

	if ctx == nil {
		ctx = context.Background()
	}

	infoLog.Printf("[API HELPER Upload] Starting upload for %s (Display: %s, MIME: %s)...", absLocalPath, displayName, uploadMimeType)
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
			time.Sleep(pollInterval)
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
// (Function largely unchanged, fixed potential nil pointer dereference for time fields)
func toolListAPIFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.ListAPIFiles: %w", clientErr)
	}
	apiFiles, err := HelperListApiFiles(context.Background(), client, interpreter.logger) // Assumes HelperListApiFiles is defined elsewhere now
	if err != nil {
		interpreter.logger.Printf("[TOOL ListAPIFiles] Warning: Error from helper (returning partial list if any): %v", err)
	}
	results := []map[string]interface{}{}
	for _, file := range apiFiles {
		createTimeStr := ""
		// *** CORRECTED CHECK: Use IsZero() only ***
		if !file.CreateTime.IsZero() { // Check if time is not the zero value
			createTimeStr = file.CreateTime.Format(time.RFC3339)
		}
		updateTimeStr := ""
		// *** CORRECTED CHECK: Use IsZero() only ***
		if !file.UpdateTime.IsZero() { // Check if time is not the zero value
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
	return map[string]interface{}{"files": results}, err
}

// --- Tool: UploadFile (Wrapper) ---
// *** MODIFIED: Check for specific binary skip error from helper ***
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
		if !ok && args[1] != nil {
			return nil, fmt.Errorf("TOOL.UploadFile: display_name must be string or null, got %T", args[1])
		}
	}
	// Assumes ErrValidationArgValue exists in scope or is defined elsewhere
	// Assumes ResolveAndSecurePath is defined elsewhere
	securePath, secErr := ResolveAndSecurePath(localPath, interpreter.sandboxDir)
	if secErr != nil {
		// Ensure errors.Join is appropriate here, might need specific error handling
		return nil, fmt.Errorf("TOOL.UploadFile: invalid path %q: %w", localPath, errors.Join(secErr))
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
		// *** Check for the specific binary skip error string (temporary) ***
		// Replace with errors.Is(uploadErr, ErrSkippedBinaryFile) once defined and used in helper
		if strings.HasPrefix(uploadErr.Error(), "skipped potentially binary file") {
			// Return a map indicating skip, and nil Go error (tool didn't fail, it skipped as designed)
			reason := uploadErr.Error() // Use the full error message as reason
			return map[string]interface{}{"status": "skipped", "reason": reason, "path": localPath}, nil
		}
		// Wrap other errors
		return nil, fmt.Errorf("TOOL.UploadFile: %w", uploadErr)
	}
	// Return file info map on success (fixed nil time checks)
	createTimeStr := ""
	// *** CORRECTED CHECK: Use IsZero() only ***
	if !apiFile.CreateTime.IsZero() { // Check if time is not the zero value
		createTimeStr = apiFile.CreateTime.Format(time.RFC3339)
	}
	updateTimeStr := ""
	// *** CORRECTED CHECK: Use IsZero() only ***
	if !apiFile.UpdateTime.IsZero() { // Check if time is not the zero value
		updateTimeStr = apiFile.UpdateTime.Format(time.RFC3339)
	}
	resultMap := map[string]interface{}{
		"name": apiFile.Name, "displayName": apiFile.DisplayName, "mimeType": apiFile.MIMEType,
		"sizeBytes": apiFile.SizeBytes, "createTime": createTimeStr, "updateTime": updateTimeStr,
		"state": string(apiFile.State), "uri": apiFile.URI, "sha256Hash": "",
	}
	if len(apiFile.Sha256Hash) > 0 {
		resultMap["sha256Hash"] = hex.EncodeToString(apiFile.Sha256Hash)
	}
	return resultMap, nil
}

// --- Registration ---
// Assumes ToolRegistry, ToolImplementation, ArgSpec, ArgTypeAny etc. are defined elsewhere
// Assumes toolDeleteAPIFile, toolSyncFiles are defined elsewhere
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
			return fmt.Errorf("failed register tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil
}
