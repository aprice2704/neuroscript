package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
)

// --- Tool: DeleteAPIFile ---
// (Function unchanged)
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
	interpreter.logger.Info("Tool: DeleteAPIFile] Attempting delete: %s", apiFileName)
	err := client.DeleteFile(context.Background(), apiFileName)
	if err != nil {
		errMsg := fmt.Sprintf("Failed delete %s: %v", apiFileName, err)
		interpreter.logger.Info("Tool: DeleteAPIFile] Error: %s", errMsg)
		return map[string]interface{}{"error": errMsg}, fmt.Errorf("TOOL.DeleteAPIFile: %w", err)
	}
	successMsg := fmt.Sprintf("Successfully deleted: %s", apiFileName)
	interpreter.logger.Info("Tool: DeleteAPIFile] %s", successMsg)
	return map[string]interface{}{"status": "success", "message": successMsg}, nil
}

// --- Helper: List API Files Helper ---
// (Function unchanged)
func HelperListApiFiles(ctx context.Context, client *genai.Client, logger interfaces.Logger) ([]*genai.File, error) {
	if client == nil {
		return nil, errors.New("genai client is nil")
	}
	if logger == nil {
		panic("List API files needs a valid logger")
	}
	logger.Debug("[API HELPER List] Fetching file list from API...")
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
			logger.Debug("[API HELPER List] %s", errMsg)
			fetchErrors++
			continue
		}
		results = append(results, file)
		fileCount++
	}
	logger.Debug("[API HELPER List] Found %d files. Encountered %d errors during fetch.", fileCount, fetchErrors)
	if fetchErrors > 0 {
		return results, fmt.Errorf("encountered %d errors fetching file list", fetchErrors)
	}
	return results, nil
}

// --- Helper: Check/Get GenAI Client ---
// (Function unchanged)
func checkGenAIClient(interpreter *Interpreter) (*genai.Client, error) {
	if interpreter == nil || interpreter.llmClient == nil || interpreter.llmClient.Client() == nil {
		return nil, errors.New("genai client is not initialized (API key potentially missing or invalid)")
	}
	return interpreter.llmClient.Client(), nil
}

// calculateFileHash calculates the SHA256 hash of a file content or returns a default hash for empty files.
// (Function unchanged)
func calculateFileHash(filePath string) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("stat file %s: %w", filePath, err)
	}
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
