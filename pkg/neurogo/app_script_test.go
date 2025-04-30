// filename: pkg/neurogo/app_script_test.go
package neurogo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/llm/noop" // Use NoOpLLMClient
)

// Helper to create a temporary script file
func createTempScript(t *testing.T, dir string, filename string, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp script file '%s': %v", filePath, err)
	}
	return filePath
}

// Test for executing a script with multi-return functions
func TestApp_RunScriptMode_MultiReturn(t *testing.T) {
	// Setup: Create temp directory and script file
	tempDir := t.TempDir()
	scriptContent := `
:: title: Multi-Return Test Script
:: version: 1.0

func multiReturnFunc() returns val1, val2, val3 means
    emit "Inside multiReturnFunc"
    set val1 = 100
    set val2 = "Success"
    set val3 = true
    return val1, val2, val3
endfunc

func main() means
    emit "Calling multiReturnFunc..."
    set resultList = multiReturnFunc()
    emit "Call complete. Result list:", resultList
    must resultList[0] == 100
    must resultList[1] == "Success"
    must resultList[2] == true
    emit "Checks passed."
endfunc
`
	scriptPath := createTempScript(t, tempDir, "multi_return.ns.txt", scriptContent)

	// Configure the App
	// Using NoOpLogger and NoOpLLMClient for testing script execution logic
	logger := &interfaces.NoOpLogger{}
	llmClient := noop.NewNoOpLLMClient(logger) // Use the NoOp LLM Client
	cfg := Config{
		Mode:       "script", // Set mode to script
		ScriptFile: scriptPath,
		TargetArg:  "main", // Target the main procedure
		// LibPaths and ProcArgs are empty for this test
	}

	// Create and run the App
	app := NewApp(cfg, logger, llmClient) // Pass LLM client to NewApp
	err := app.Run(context.Background())

	// Assert: Check for errors during execution
	if err != nil {
		t.Errorf("Expected successful execution of multi-return script, but got error: %v", err)
	}
}

// Add more tests for app_script.go functionality as needed
