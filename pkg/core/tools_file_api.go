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
	"strings" // Added for WaitGroup in sync helper
	"time"

	"github.com/google/generative-ai-go/genai" // Added for gitignore handling in sync helper
	"google.golang.org/api/iterator"
)

// --- Constants, init, and Hash Helper (Unchanged) ---
const emptyFileContentForHash = " "

var emptyFileHash string

func init() {
	hasher := sha256.New()
	hasher.Write([]byte(emptyFileContentForHash))
	emptyFileHash = hex.EncodeToString(hasher.Sum(nil))
}
func calculateFileHash(filePath string) (string, error) { /* ... implementation unchanged ... */
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("stat: %w", err)
	}
	if fileInfo.Size() == 0 {
		return emptyFileHash, nil
	}
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open: %w", err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("copy: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// --- Helper: Check/Get GenAI Client (Unchanged) ---
func checkGenAIClient(interpreter *Interpreter) (*genai.Client, error) { /* ... implementation unchanged ... */
	if interpreter == nil || interpreter.llmClient == nil || interpreter.llmClient.Client() == nil {
		return nil, errors.New("GenAI client is not initialized")
	}
	return interpreter.llmClient.Client(), nil
}

// --- Helper: Upload File and Poll (Unchanged) ---
// HelperUploadAndPollFile handles the core logic of uploading a single file and waiting for it to be ACTIVE.
func HelperUploadAndPollFile(ctx context.Context, absLocalPath string, displayName string, client *genai.Client, logger *log.Logger) (*genai.File, error) { /* ... implementation unchanged from previous step ... */
	if client == nil {
		return nil, errors.New("genai client is nil")
	}
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	logger.Printf("[API HELPER Upload] Processing: %s (Display: %s)", absLocalPath, displayName)
	fileInfo, err := os.Stat(absLocalPath)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", absLocalPath, err)
	}
	isZeroByte := fileInfo.Size() == 0
	mimeType := ""
	if isZeroByte {
		mimeType = "text/plain"
	} else {
		mimeType = mime.TypeByExtension(filepath.Ext(absLocalPath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}
	options := &genai.UploadFileOptions{MIMEType: mimeType, DisplayName: displayName}
	var reader io.Reader
	var fileHandle *os.File
	if isZeroByte {
		reader = strings.NewReader(emptyFileContentForHash)
	} else {
		fileHandle, err = os.Open(absLocalPath)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", absLocalPath, err)
		}
		defer fileHandle.Close()
		reader = fileHandle
	}
	if ctx == nil {
		ctx = context.Background()
	} // Default context
	apiFile, err := client.UploadFile(ctx, "", reader, options)
	if err != nil {
		return nil, fmt.Errorf("api upload %q: %w", absLocalPath, err)
	}
	logger.Printf("[API HELPER Upload] Initiated -> API Name: %s", apiFile.Name)
	startTime := time.Now()
	pollInterval := 1 * time.Second
	const maxPollInterval = 10 * time.Second
	const timeout = 2 * time.Minute
	for apiFile.State == genai.FileStateProcessing {
		if time.Since(startTime) > timeout {
			errMsg := fmt.Sprintf("timeout %s (API: %s)", absLocalPath, apiFile.Name)
			logger.Printf("[ERROR API HELPER Upload] %s. Deleting.", errMsg)
			_ = client.DeleteFile(context.Background(), apiFile.Name)
			return nil, fmt.Errorf(errMsg)
		}
		time.Sleep(pollInterval)
		updatedFile, err := client.GetFile(context.Background(), apiFile.Name)
		if err != nil {
			errMsg := fmt.Errorf("get status %s: %w", apiFile.Name, err)
			logger.Printf("[ERROR API HELPER Upload] %s.", errMsg)
			return nil, errMsg
		} // Don't delete on transient Get error
		apiFile = updatedFile
		pollInterval += 500 * time.Millisecond
		if pollInterval > maxPollInterval {
			pollInterval = maxPollInterval
		}
		logger.Printf("[DEBUG API HELPER Upload] Poll %s, state: %s", apiFile.Name, apiFile.State)
	}
	if apiFile.State != genai.FileStateActive {
		errMsg := fmt.Sprintf("not ACTIVE %s (State: %s)", absLocalPath, apiFile.State)
		logger.Printf("[ERROR API HELPER Upload] %s (API: %s). Deleting.", errMsg, apiFile.Name)
		_ = client.DeleteFile(context.Background(), apiFile.Name)
		return nil, fmt.Errorf(errMsg)
	}
	logger.Printf("[API HELPER Upload] Success ACTIVE: %s -> %s", absLocalPath, apiFile.Name)
	return apiFile, nil
}

// +++ ADDED: Reusable List API Files Helper +++
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

	for {
		file, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			errMsg := fmt.Sprintf("Error fetching file list page: %v", err)
			logger.Printf("[API HELPER List] %s", errMsg)
			fetchErrors++
			continue // Skip this page/error
		}
		results = append(results, file)
	}

	logger.Printf("[API HELPER List] Found %d files. Encountered %d errors during fetch.", len(results), fetchErrors)
	if fetchErrors > 0 {
		// Return partial list and an error
		return results, fmt.Errorf("encountered %d errors fetching file list", fetchErrors)
	}
	return results, nil
}

// --- Tool: ListAPIFiles (Wrapper) ---
func toolListAPIFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.ListAPIFiles: %w", clientErr)
	}

	// Call the helper
	apiFiles, err := HelperListApiFiles(context.Background(), client, interpreter.logger)
	if err != nil {
		// Helper logs details, tool should return error if helper failed significantly
		interpreter.logger.Printf("[TOOL ListAPIFiles] Error from helper: %v", err)
		// Format results even if there was a partial error
	}

	// Convert []*genai.File to []map[string]interface{} for NeuroScript
	results := []map[string]interface{}{}
	for _, file := range apiFiles {
		fileInfo := map[string]interface{}{
			"name": file.Name, "displayName": file.DisplayName, "mimeType": file.MIMEType,
			"sizeBytes": file.SizeBytes, "createTime": file.CreateTime.Format(time.RFC3339),
			"updateTime": file.UpdateTime.Format(time.RFC3339), "state": string(file.State),
			"uri": file.URI, "sha256Hash": hex.EncodeToString(file.Sha256Hash),
		}
		results = append(results, fileInfo)
	}

	// Return results, include Go error if helper returned one
	return results, err
}

// --- Tool: DeleteAPIFile (Unchanged - uses client directly) ---
func toolDeleteAPIFile(interpreter *Interpreter, args []interface{}) (interface{}, error) { /* ... implementation unchanged ... */
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.DeleteAPIFile: %w", clientErr)
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("TOOL.DeleteAPIFile: expected 1 arg, got %d", len(args))
	}
	apiFileName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.DeleteAPIFile: arg must be string, got %T", args[0])
	}
	if apiFileName == "" {
		return nil, fmt.Errorf("TOOL.DeleteAPIFile: API name empty")
	}
	interpreter.logger.Printf("[TOOL DeleteAPIFile] Attempting delete: %s", apiFileName)
	err := client.DeleteFile(context.Background(), apiFileName)
	if err != nil {
		errMsg := fmt.Sprintf("Failed delete %s: %v", apiFileName, err)
		interpreter.logger.Printf("[TOOL DeleteAPIFile] %s", errMsg)
		return errMsg, fmt.Errorf("TOOL.DeleteAPIFile: %w", err)
	}
	successMsg := fmt.Sprintf("Successfully deleted: %s", apiFileName)
	interpreter.logger.Printf("[TOOL DeleteAPIFile] %s", successMsg)
	return successMsg, nil
}

// --- Tool: UploadFile (Wrapper) ---
func toolUploadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.UploadFile: %w", clientErr)
	}
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("TOOL.UploadFile: expected 1-2 args, got %d", len(args))
	}
	localPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.UploadFile: localPath must be string, got %T", args[0])
	}
	if localPath == "" {
		return nil, fmt.Errorf("TOOL.UploadFile: localPath empty")
	}
	var displayName string
	if len(args) == 2 {
		displayName, ok = args[1].(string)
		if !ok && args[1] != nil {
			return nil, fmt.Errorf("TOOL.UploadFile: displayName must be string or null, got %T", args[1])
		}
	}

	securePath, secErr := SecureFilePath(localPath, interpreter.sandboxDir)
	if secErr != nil {
		return nil, fmt.Errorf("TOOL.UploadFile: invalid path %q: %w", localPath, errors.Join(ErrValidationArgValue, secErr))
	}
	interpreter.logger.Printf("[TOOL UploadFile] Validated path: %s -> %s", localPath, securePath)

	if displayName == "" { /* Default display name logic */
		if interpreter.sandboxDir != "" {
			relPath, err := filepath.Rel(interpreter.sandboxDir, securePath)
			if err == nil {
				displayName = filepath.ToSlash(relPath)
			} else {
				displayName = filepath.Base(securePath)
			}
		} else {
			displayName = filepath.Base(securePath)
		}
		interpreter.logger.Printf("[TOOL UploadFile] Using default display name: %s", displayName)
	}

	apiFile, uploadErr := HelperUploadAndPollFile(context.Background(), securePath, displayName, client, interpreter.logger)
	if uploadErr != nil {
		return nil, fmt.Errorf("TOOL.UploadFile: %w", uploadErr)
	}

	resultMap := map[string]interface{}{
		"name": apiFile.Name, "displayName": apiFile.DisplayName, "mimeType": apiFile.MIMEType, "sizeBytes": apiFile.SizeBytes,
		"createTime": apiFile.CreateTime.Format(time.RFC3339), "updateTime": apiFile.UpdateTime.Format(time.RFC3339),
		"state": string(apiFile.State), "uri": apiFile.URI, "sha256Hash": hex.EncodeToString(apiFile.Sha256Hash),
	}
	return resultMap, nil
}

// --- Tool: SyncFiles (Wrapper) ---
func toolSyncFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	client, clientErr := checkGenAIClient(interpreter)
	if clientErr != nil {
		return nil, fmt.Errorf("TOOL.SyncFiles: %w", clientErr)
	}

	// --- Argument Parsing & Validation ---
	if len(args) < 2 || len(args) > 4 { // Updated count for ignore-gitignore
		return nil, fmt.Errorf("TOOL.SyncFiles: expected 2-4 arguments (direction, localDir, [filterPattern], [ignoreGitignore]), got %d", len(args))
	}
	direction, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.SyncFiles: direction must be string, got %T", args[0])
	}
	localDir, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("TOOL.SyncFiles: localDir must be string, got %T", args[1])
	}
	if localDir == "" {
		return nil, fmt.Errorf("TOOL.SyncFiles: localDir empty")
	}
	var filterPattern string
	if len(args) >= 3 {
		filterPattern, ok = args[2].(string)
		if !ok && args[2] != nil {
			return nil, fmt.Errorf("TOOL.SyncFiles: filterPattern must be string or null, got %T", args[2])
		}
	}
	var ignoreGitignore bool = false
	if len(args) == 4 {
		ignoreGitignore, ok = args[3].(bool)
		if !ok {
			return nil, fmt.Errorf("TOOL.SyncFiles: ignoreGitignore must be boolean, got %T", args[3])
		}
	}

	direction = strings.ToLower(direction)
	if direction != "up" {
		return nil, fmt.Errorf("TOOL.SyncFiles: direction '%s' not supported. Only 'up' implemented", direction)
	}

	absLocalDir, secErr := SecureFilePath(localDir, interpreter.sandboxDir)
	if secErr != nil {
		return nil, fmt.Errorf("TOOL.SyncFiles: invalid localDir '%s': %w", localDir, errors.Join(ErrValidationArgValue, secErr))
	}
	dirInfo, statErr := os.Stat(absLocalDir)
	if statErr != nil {
		return nil, fmt.Errorf("TOOL.SyncFiles: cannot access localDir '%s': %w", localDir, statErr)
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("TOOL.SyncFiles: localDir '%s' is not a directory", localDir)
	}
	interpreter.logger.Printf("[TOOL SyncFiles] Validated dir: %s (Ignore .gitignore: %t)", absLocalDir, ignoreGitignore)

	// Call the helper function
	statsMap, syncErr := SyncDirectoryUpHelper(
		context.Background(),
		absLocalDir,
		filterPattern,
		ignoreGitignore, // Pass flag
		client,
		interpreter.logger, // Use main logger as info logger
		interpreter.logger, // Use main logger as error logger
		interpreter.logger, // Use main logger as debug logger (adjust if separate log levels needed)
	)

	// HelperSyncDirectoryUp returns error for critical issues, stats map otherwise
	// The tool function should return the map AND the error if one occurred.
	return statsMap, syncErr
}

// --- Registration (Unchanged) ---
func registerFileAPITools(registry *ToolRegistry) error {
	// ... uses previous registration code ...
	var err error
	tools := []ToolImplementation{{Spec: ToolSpec{Name: "ListAPIFiles" /*...*/}, Func: toolListAPIFiles}, {Spec: ToolSpec{Name: "DeleteAPIFile" /*...*/}, Func: toolDeleteAPIFile}, {Spec: ToolSpec{Name: "UploadFile" /*...*/}, Func: toolUploadFile}, {Spec: ToolSpec{Name: "SyncFiles", Description: "Syncs local dir to API ('up' only). Returns stats map.", Args: []ArgSpec{{Name: "direction", Type: ArgTypeString, Required: true}, {Name: "local_dir", Type: ArgTypeString, Required: true}, {Name: "filter_pattern", Type: ArgTypeString, Required: false}, {Name: "ignore_gitignore", Type: ArgTypeBool, Required: false, Description: "Ignore .gitignore files if true."}}, ReturnType: ArgTypeAny}, Func: toolSyncFiles}}
	for _, tool := range tools {
		if err = registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed register %s: %w", tool.Spec.Name, err)
		}
	}
	return nil
}
