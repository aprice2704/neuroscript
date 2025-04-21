// filename: pkg/core/sync_helpers.go
package core

import (
	"context" // Added context import
	"encoding/hex"
	"errors"
	"fmt"

	// Required imports for genai and gitignore
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator" // Added for listExistingAPIFiles
)

// listExistingAPIFiles fetches the list of files from the API and returns them as a map.
// Requires access to syncContext definition (from sync_types.go).
// VERSION 2: Uses specific loggers from syncContext
func listExistingAPIFiles_v2(sc *syncContext) (map[string]*genai.File, error) {
	sc.infoLog.Println("[API HELPER Sync] Listing current API files...")

	if sc.client == nil {
		return nil, errors.New("genai client is nil in listExistingAPIFiles")
	}

	sc.infoLog.Println("[API HELPER List] Fetching file list from API...") // Use infoLog
	if sc.ctx == nil {
		sc.ctx = context.Background()
	}
	iter := sc.client.ListFiles(sc.ctx)
	results := []*genai.File{}
	fetchErrors := 0

	for {
		file, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			errMsg := fmt.Sprintf("Error fetching file list page: %v", err)
			sc.errorLog.Printf("[API HELPER List] %s", errMsg) // Use errorLog
			fetchErrors++
			continue
		}
		results = append(results, file)
	}
	sc.infoLog.Printf("[API HELPER List] Found %d files. Encountered %d errors during fetch.", len(results), fetchErrors) // Use infoLog

	apiFilesMap := make(map[string]*genai.File)
	for _, file := range results {
		if file.DisplayName != "" {
			apiFilesMap[file.DisplayName] = file
			hashPrefix := ""
			if len(file.Sha256Hash) > 0 {
				hashPrefix = hex.EncodeToString(file.Sha256Hash)[:min(len(hex.EncodeToString(file.Sha256Hash)), 8)]
			}
			sc.debugLog.Printf("[DEBUG API HELPER Sync] API File Found: Name=%s, DisplayName=%s, SHA=%s...", file.Name, file.DisplayName, hashPrefix) // Use debugLog
		} else {
			sc.debugLog.Printf("[WARN API HELPER Sync] API File Found empty DisplayName: %s", file.Name) // Use debugLog
		}
	}
	sc.infoLog.Printf("[API HELPER Sync] Found %d API files.", len(apiFilesMap)) // Use infoLog

	var returnErr error
	if fetchErrors > 0 {
		sc.incrementStat("list_api_errors")
		returnErr = fmt.Errorf("encountered %d errors fetching file list", fetchErrors)
		sc.errorLog.Printf("[ERROR API HELPER Sync] List API files finished with errors: %v", returnErr) // Use errorLog
	}

	return apiFilesMap, returnErr
}
