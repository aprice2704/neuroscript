// filename: pkg/core/tools_file_api.go
// UPDATED: Add HelperUploadStringAndPollFile, toolUpsertAs, and register UpsertAs
package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"mime" // Using standard library mime package
	"os"
	"path/filepath"
	"strings"
	"time"

	// Remove gabriel-vasile/mimetype if only using standard library now
	// "github.com/gabriel-vasile/mimetype"
	"github.com/google/generative-ai-go/genai"
)

// --- Constants, init, and Hash Helper --- (Unchanged)
const (
	emptyFileContentForHash = " "
)

var (
	emptyFileHash string
	// ErrSkippedBinaryFile defined in errors.go
)

func init() {
	hasher := sha256.New()
	hasher.Write([]byte(emptyFileContentForHash))
	emptyFileHash = hex.EncodeToString(hasher.Sum(nil))
}

// --- Helper: Upload File and Poll --- (Unchanged from fetch)
func HelperUploadAndPollFile(ctx context.Context, absLocalPath string, displayName string, client *genai.Client, logger *log.Logger) (*genai.File, error) {
	if client == nil {
		return nil, errors.New("genai client is nil")
	}
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	debugLog := logger
	infoLog := logger
	errorLog := logger // Simple assignment for this helper
	debugLog.Printf("[API HELPER Upload] Processing: %s (Display: %s)", absLocalPath, displayName)
	fileInfo, err := os.Stat(absLocalPath)
	if err != nil {
		return nil, fmt.Errorf("stat file %s: %w", absLocalPath, err)
	}
	isZeroByte := fileInfo.Size() == 0
	fileExt := filepath.Ext(absLocalPath)
	uploadMimeType := "application/octet-stream" // Default to generic byte stream

	if isZeroByte {
		uploadMimeType = "text/plain"
		debugLog.Printf("[API HELPER Upload] Handling zero-byte file: %s as text/plain", absLocalPath)
	} else {
		// Use standard library mime detection based on extension first
		stdMime := mime.TypeByExtension(fileExt)
		if stdMime != "" {
			uploadMimeType = stdMime
			debugLog.Printf("[API HELPER Upload] Detected MIME by extension: %s for %s", uploadMimeType, absLocalPath)
		} else {
			debugLog.Printf("[API HELPER Upload] MIME not detected by extension for %s, keeping default %s", absLocalPath, uploadMimeType)
			// Could add sniffing here if needed, but keeping simple for now
		}

		// Force text/plain for known text types or Go files for compatibility
		// This logic might need refinement based on File API actual compatibility
		if fileExt == ".go" || strings.HasPrefix(uploadMimeType, "text/") {
			uploadMimeType = "text/plain"
			debugLog.Printf("[API HELPER Upload] Forcing upload MIME type to text/plain for compatibility: %s", absLocalPath)
		} else {
			// If not obviously text, warn and consider skipping (or allow generic upload)
			warnMsg := fmt.Sprintf("File type for %s is %s (not text/* or .go). Uploading as %s, but processing might fail.", absLocalPath, stdMime, uploadMimeType)
			infoLog.Printf("[WARN API HELPER Upload] %s", warnMsg)
			// Decide whether to return an error or proceed with generic upload
			// Proceeding for now, but this could be where ErrSkippedBinaryFile is used
		}
	}

	options := &genai.UploadFileOptions{MIMEType: uploadMimeType, DisplayName: displayName}
	var reader io.Reader
	if isZeroByte {
		reader = strings.NewReader("")
		debugLog.Printf("[API HELPER Upload] Using empty string reader for zero-byte file: %s", absLocalPath)
	} else {
		fileHandle, err := os.Open(absLocalPath)
		if err != nil {
			return nil, fmt.Errorf("open file %s for upload: %w", absLocalPath, err)
		}
		defer fileHandle.Close()
		reader = fileHandle
	}

	if ctx == nil {
		ctx = context.Background()
	}
	infoLog.Printf("[API HELPER Upload] Starting upload for %s (Display: %s, MIME: %s)...", absLocalPath, displayName, uploadMimeType)
	apiFile, err := client.UploadFile(ctx, "", reader, options)
	if err != nil {
		return nil, fmt.Errorf("api upload failed for %q (Display: %s): %w", absLocalPath, displayName, err)
	}
	debugLog.Printf("[API HELPER Upload] Upload initiated -> API Name: %s (URI: %s)", apiFile.Name, apiFile.URI)

	// Polling Logic (Unchanged from fetch)
	startTime := time.Now()
	pollInterval := 1 * time.Second
	const maxPollInterval = 10 * time.Second
	const timeout = 2 * time.Minute
	for apiFile.State == genai.FileStateProcessing {
		if time.Since(startTime) > timeout {
			errMsg := fmt.Sprintf("polling timeout for file %s (API Name: %s, Display: %s)", absLocalPath, apiFile.Name, displayName)
			errorLog.Printf("[ERROR API HELPER Upload] %s. Attempting to delete orphaned API file.", errMsg)
			_ = client.DeleteFile(context.Background(), apiFile.Name) // Best effort delete
			return nil, fmt.Errorf(errMsg)
		}
		time.Sleep(pollInterval)
		debugLog.Printf("[DEBUG API HELPER Upload] Polling status for API file %s...", apiFile.Name)
		getCtx, cancelGet := context.WithTimeout(context.Background(), 30*time.Second)
		updatedFile, getErr := client.GetFile(getCtx, apiFile.Name)
		cancelGet()
		if getErr != nil {
			errorLog.Printf("[WARN API HELPER Upload] Error getting status for %s (will retry): %v", apiFile.Name, getErr)
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
		_ = client.DeleteFile(context.Background(), apiFile.Name) // Best effort delete
		return nil, fmt.Errorf(errMsg)
	}
	infoLog.Printf("[API HELPER Upload] Upload successful and ACTIVE: %s -> %s", displayName, apiFile.Name)
	return apiFile, nil
}

// +++ NEW HELPER: Upload String Content and Poll +++
// HelperUploadStringAndPollFile handles uploading string content and waiting for it to be ACTIVE.
func HelperUploadStringAndPollFile(ctx context.Context, content string, displayName string, client *genai.Client, logger *log.Logger) (*genai.File, error) {
	if client == nil {
		return nil, errors.New("genai client is nil")
	}
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	debugLog := logger
	infoLog := logger
	errorLog := logger

	// Assume string content should always be uploaded as text/plain
	uploadMimeType := "text/plain"
	debugLog.Printf("[API HELPER UploadString] Processing content (Display: %s, Length: %d) as %s", displayName, len(content), uploadMimeType)

	options := &genai.UploadFileOptions{MIMEType: uploadMimeType, DisplayName: displayName}
	reader := strings.NewReader(content) // Use strings.NewReader for the content

	if ctx == nil {
		ctx = context.Background()
	}

	infoLog.Printf("[API HELPER UploadString] Starting upload for content (Display: %s, MIME: %s)...", displayName, uploadMimeType)
	apiFile, err := client.UploadFile(ctx, "", reader, options)
	if err != nil {
		return nil, fmt.Errorf("api upload failed for content (Display: %s): %w", displayName, err)
	}
	debugLog.Printf("[API HELPER UploadString] Upload initiated -> API Name: %s (URI: %s)", apiFile.Name, apiFile.URI)

	// Polling Logic (Identical to HelperUploadAndPollFile)
	startTime := time.Now()
	pollInterval := 1 * time.Second
	const maxPollInterval = 10 * time.Second
	const timeout = 2 * time.Minute
	for apiFile.State == genai.FileStateProcessing {
		if time.Since(startTime) > timeout {
			errMsg := fmt.Sprintf("polling timeout for content (API Name: %s, Display: %s)", apiFile.Name, displayName)
			errorLog.Printf("[ERROR API HELPER UploadString] %s. Attempting to delete orphaned API file.", errMsg)
			_ = client.DeleteFile(context.Background(), apiFile.Name) // Best effort delete
			return nil, fmt.Errorf(errMsg)
		}
		time.Sleep(pollInterval)
		debugLog.Printf("[DEBUG API HELPER UploadString] Polling status for API file %s...", apiFile.Name)
		getCtx, cancelGet := context.WithTimeout(context.Background(), 30*time.Second)
		updatedFile, getErr := client.GetFile(getCtx, apiFile.Name)
		cancelGet()
		if getErr != nil {
			errorLog.Printf("[WARN API HELPER UploadString] Error getting status for %s (will retry): %v", apiFile.Name, getErr)
			continue
		}
		apiFile = updatedFile
		pollInterval *= 2
		if pollInterval > maxPollInterval {
			pollInterval = maxPollInterval
		}
		debugLog.Printf("[DEBUG API HELPER UploadString] Poll %s successful, state: %s (Next poll in %v)", apiFile.Name, apiFile.State, pollInterval)
	}
	if apiFile.State != genai.FileStateActive {
		errMsg := fmt.Sprintf("file processing failed for content (API Name: %s, Display: %s). Final State: %s", apiFile.Name, displayName, apiFile.State)
		errorLog.Printf("[ERROR API HELPER UploadString] %s. Attempting to delete failed API file.", errMsg)
		_ = client.DeleteFile(context.Background(), apiFile.Name) // Best effort delete
		return nil, fmt.Errorf(errMsg)
	}
	infoLog.Printf("[API HELPER UploadString] Upload successful and ACTIVE: %s -> %s", displayName, apiFile.Name)
	return apiFile, nil
}

// --- Tool: ListAPIFiles --- (Unchanged from fetch)
func toolListAPIFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.ListAPIFiles: %w", clientErr)
	}
	apiFiles, err := HelperListApiFiles(context.Background(), client, interpreter.logger)
	if err != nil {
		interpreter.logger.Printf("[TOOL ListAPIFiles] Warning: Error from helper (returning partial list if any): %v", err)
	}
	results := []map[string]interface{}{}
	for _, file := range apiFiles {
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
		fileInfo := map[string]interface{}{"name": file.Name, "displayName": file.DisplayName, "mimeType": file.MIMEType, "sizeBytes": file.SizeBytes, "createTime": createTimeStr, "updateTime": updateTimeStr, "state": string(file.State), "uri": file.URI, "sha256Hash": ""}
		if len(file.Sha256Hash) > 0 {
			fileInfo["sha256Hash"] = hex.EncodeToString(file.Sha256Hash)
		}
		results = append(results, fileInfo)
	}
	return map[string]interface{}{"files": results}, err
}

// --- Tool: UploadFile --- (Unchanged from fetch)
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
	securePath, secErr := ResolveAndSecurePath(localPath, interpreter.sandboxDir)
	if secErr != nil {
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
		if strings.HasPrefix(uploadErr.Error(), "skipped potentially binary file") {
			reason := uploadErr.Error()
			return map[string]interface{}{"status": "skipped", "reason": reason, "path": localPath}, nil
		}
		return nil, fmt.Errorf("TOOL.UploadFile: %w", uploadErr)
	}
	if apiFile == nil {
		return nil, errors.New("TOOL.UploadFile: helper returned nil file without error")
	}
	createTimeStr := ""
	if !apiFile.CreateTime.IsZero() {
		createTimeStr = apiFile.CreateTime.Format(time.RFC3339)
	}
	updateTimeStr := ""
	if !apiFile.UpdateTime.IsZero() {
		updateTimeStr = apiFile.UpdateTime.Format(time.RFC3339)
	}
	resultMap := map[string]interface{}{"name": apiFile.Name, "displayName": apiFile.DisplayName, "mimeType": apiFile.MIMEType, "sizeBytes": apiFile.SizeBytes, "createTime": createTimeStr, "updateTime": updateTimeStr, "state": string(apiFile.State), "uri": apiFile.URI, "sha256Hash": ""}
	if len(apiFile.Sha256Hash) > 0 {
		resultMap["sha256Hash"] = hex.EncodeToString(apiFile.Sha256Hash)
	}
	return resultMap, nil
}

// +++ NEW TOOL: UpsertAs +++
// toolUpsertAs takes content string and display name, uploads it, returns map.
func toolUpsertAs(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter) // Check client first
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.UpsertAs: %w", clientErr)
	}
	if len(args) != 2 {
		return nil, fmt.Errorf("TOOL.UpsertAs: expected 2 arguments (contents, display_name), got %d", len(args))
	}
	contents, okC := args[0].(string)
	if !okC {
		return nil, fmt.Errorf("TOOL.UpsertAs: contents argument must be a string, got %T", args[0])
	}
	displayName, okN := args[1].(string)
	if !okN {
		return nil, fmt.Errorf("TOOL.UpsertAs: display_name argument must be a string, got %T", args[1])
	}
	if displayName == "" {
		return nil, errors.New("TOOL.UpsertAs: display_name cannot be empty")
	}
	// Content can be empty

	// Use the new helper function for uploading string content
	apiFile, uploadErr := HelperUploadStringAndPollFile(context.Background(), contents, displayName, client, interpreter.logger)
	if uploadErr != nil {
		// Don't need to check for binary skip here as we force text/plain
		return nil, fmt.Errorf("TOOL.UpsertAs: failed to upload content for display name '%s': %w", displayName, uploadErr)
	}

	// Check for nil API file just in case helper returns nil without error
	if apiFile == nil || apiFile.Name == "" { // apiFile.Name contains the URI like "files/..."
		return nil, fmt.Errorf("TOOL.UpsertAs: upload helper returned nil or empty File/URI for display name '%s'", displayName)
	}

	// Return map required by AgentPin: {"displayName": string, "uri": string}
	// Using displayName passed in, as relative path isn't applicable here.
	resultMap := map[string]interface{}{
		"displayName": displayName,
		"uri":         apiFile.Name, // Use Name which contains the URI
	}
	interpreter.logger.Printf("[TOOL UpsertAs] Successfully uploaded content '%s' -> URI: %s", displayName, apiFile.Name)
	return resultMap, nil
}

// --- Registration ---
// UPDATED: Add UpsertAs registration
func registerFileAPITools(registry *ToolRegistry) error {
	var err error
	tools := []ToolImplementation{
		{Spec: ToolSpec{Name: "ListAPIFiles", Description: "Lists files previously uploaded to the API.", Args: []ArgSpec{}, ReturnType: ArgTypeAny}, Func: toolListAPIFiles},
		{Spec: ToolSpec{Name: "DeleteAPIFile", Description: "Deletes a file from the API by its name (e.g., 'files/abc123xyz').", Args: []ArgSpec{{Name: "api_file_name", Type: ArgTypeString, Required: true, Description: "The full API name of the file (e.g., files/xyz)."}}, ReturnType: ArgTypeAny}, Func: toolDeleteAPIFile},
		{Spec: ToolSpec{Name: "UploadFile", Description: "Uploads a local file (relative to sandbox) to the API.", Args: []ArgSpec{{Name: "local_path", Type: ArgTypeString, Required: true, Description: "Relative path to the local file."}, {Name: "display_name", Type: ArgTypeString, Required: false, Description: "Optional display name (defaults to relative path)."}}, ReturnType: ArgTypeAny}, Func: toolUploadFile},
		{Spec: ToolSpec{Name: "SyncFiles", Description: "Syncs local directory (relative to sandbox) to API ('up' only).", Args: []ArgSpec{{Name: "direction", Type: ArgTypeString, Required: true, Description: "Sync direction ('up')."}, {Name: "local_dir", Type: ArgTypeString, Required: true, Description: "Relative path to local directory."}, {Name: "filter_pattern", Type: ArgTypeString, Required: false, Description: "Optional filename glob pattern."}, {Name: "ignore_gitignore", Type: ArgTypeBool, Required: false, Description: "Ignore .gitignore files if true (default: false)."}}, ReturnType: ArgTypeAny}, Func: toolSyncFiles},
		// +++ NEW: UpsertAs Registration +++
		{
			Spec: ToolSpec{
				Name:        "UpsertAs",
				Description: "Uploads string content to the File API with a specified display name.",
				Args: []ArgSpec{
					{Name: "contents", Type: ArgTypeString, Required: true, Description: "The string content to upload."},
					{Name: "display_name", Type: ArgTypeString, Required: true, Description: "The desired display name for the file in the API."},
				},
				// Returns map: {"displayName": string, "uri": string}
				ReturnType: ArgTypeMap, // Return type is Map
			},
			Func: toolUpsertAs,
		},
		// +++ END NEW +++
	}
	for _, tool := range tools {
		if err = registry.RegisterTool(tool); err != nil {
			log.Printf("Error registering tool %s: %v", tool.Spec.Name, err)
			return fmt.Errorf("failed register tool %s: %w", tool.Spec.Name, err) // Return on first error
		}
	}
	return nil
}

// Helper function checkGenAIClient (Assume defined elsewhere or add stub)
// func checkGenAIClient(interpreter *Interpreter) (*genai.Client, error) {
// 	if interpreter == nil || interpreter.llmClient == nil || interpreter.llmClient.Client() == nil {
// 		return nil, errors.New("GenAI client not initialized")
// 	}
// 	return interpreter.llmClient.Client(), nil
// }

// Assumed functions (ensure defined elsewhere):
// - func HelperListApiFiles(ctx context.Context, client *genai.Client, logger *log.Logger) ([]*genai.File, error) // Likely in sync_helpers.go or sync_morehelpers.go
// - func toolDeleteAPIFile(interpreter *Interpreter, args []interface{}) (interface{}, error) // Likely here or another tools_file_api_*.go
// - func toolSyncFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) // Likely here or another tools_file_api_*.go
// - func ResolveAndSecurePath(localPath string, sandboxDir string) (string, error) // Likely in security_helpers.go or tools_helpers.go
