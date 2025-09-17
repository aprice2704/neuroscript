// NeuroScript Version: 0.7.2
// File version: 3
// Purpose: Corrects the test to set the sandbox directory explicitly after interpreter creation to ensure correct initialization order.
// filename: pkg/api/interpreter_options_test.go
// nlines: 72
// risk_rating: MEDIUM

package api_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
)

// TestInterpreter_WithSandboxDir provides an end-to-end test to ensure that the
// WithSandboxDir option correctly configures the interpreter's security sandbox
// for filesystem operations.
func TestInterpreter_WithSandboxDir(t *testing.T) {
	// 1. Setup: Create a temporary directory to act as the sandbox.
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "output.txt")
	fileContent := "sandbox test successful"

	// 2. Define a script that writes to a file.
	script := `
func main(needs path, content) means
    must tool.fs.write(path, content)
endfunc
`
	// 3. Create a trusted configuration interpreter.
	allowedTools := []string{"tool.fs.write"}
	requiredGrants := []api.Capability{
		api.NewCapability(api.ResFS, api.VerbWrite, "*"),
	}
	// FIX: Create the interpreter first, then explicitly set the sandbox directory.
	// This avoids potential initialization order issues with options.
	interp := api.NewConfigInterpreter(
		allowedTools,
		requiredGrants,
	)
	interp.SetSandboxDir(tempDir)

	// 4. Parse and load the script into the interpreter.
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse() failed: %v", err)
	}
	if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("api.LoadFromUnit() failed: %v", err)
	}

	// 5. Run the procedure.
	_, err = api.RunProcedure(context.Background(), interp, "main", "output.txt", fileContent)
	if err != nil {
		t.Fatalf("api.RunProcedure() failed unexpectedly: %v", err)
	}

	// 6. Verify that the file was created in the correct location with the correct content.
	data, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file from sandbox: %v", err)
	}
	if string(data) != fileContent {
		t.Errorf("Expected file content '%s', but got '%s'", fileContent, string(data))
	}
}
