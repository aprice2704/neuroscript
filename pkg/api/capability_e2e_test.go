// NeuroScript Version: 0.8.0
// File version: 10
// Purpose: Corrects calls to newTestHostContext to pass the required logger argument.
// filename: pkg/api/capability_e2e_test.go
// nlines: 115
// risk_rating: HIGH

package api_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// secureFileWriteTool is a custom tool for testing that requires 'fs:write' capability.
var secureFileWriteTool = api.ToolImplementation{
	Spec: api.ToolSpec{
		Name:  "writeFile",
		Group: "test",
		Args: []api.ArgSpec{
			{Name: "path", Type: "string", Required: true},
			{Name: "content", Type: "string", Required: true},
		},
		ReturnType: "nil",
	},
	Func: func(rt api.Runtime, args []any) (any, error) {
		path := args[0].(string)
		content := args[1].(string)
		return nil, os.WriteFile(path, []byte(content), 0600)
	},
	RequiredCaps: []capability.Capability{
		{Resource: "fs", Verbs: []string{"write"}},
	},
}

const secureWriteScript = `
func main(needs path, content) means
    call tool.test.writeFile(path, content)
endfunc
`

// TestE2E_CapabilityCheck_Success verifies that a script calling a secure tool
// succeeds when the interpreter is configured with the correct capability grants.
func TestE2E_CapabilityCheck_Success(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "output.txt")
	fileContent := "it worked"

	requiredGrants := []api.Capability{
		{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"*"}},
	}
	allowedTools := []string{"tool.test.writeFile"}

	// FIX: Pass nil for the logger argument.
	hc := newTestHostContext(nil)

	interp := api.NewConfigInterpreter(
		allowedTools,
		requiredGrants,
		api.WithHostContext(hc),
		api.WithTool(secureFileWriteTool),
		api.WithSandboxDir(tempDir),
	)

	tree, err := api.Parse([]byte(secureWriteScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse() failed: %v", err)
	}
	if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("api.LoadFromUnit() failed: %v", err)
	}

	_, err = api.RunProcedure(context.Background(), interp, "main", targetFile, fileContent)
	if err != nil {
		t.Fatalf("api.RunProcedure() failed unexpectedly: %v", err)
	}

	data, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	if string(data) != fileContent {
		t.Errorf("Expected file content '%s', but got '%s'", fileContent, string(data))
	}
}

// TestE2E_CapabilityCheck_Failure verifies that a script calling a secure tool
// fails with a permission error when the interpreter lacks the required grants.
func TestE2E_CapabilityCheck_Failure(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "output.txt")
	allowedTools := []string{"tool.test.writeFile"}

	// FIX: Pass nil for the logger argument.
	hc := newTestHostContext(nil)

	interp := api.NewConfigInterpreter(
		allowedTools,
		[]api.Capability{}, // Empty grant set
		api.WithHostContext(hc),
		api.WithTool(secureFileWriteTool),
		api.WithSandboxDir(tempDir),
	)

	tree, err := api.Parse([]byte(secureWriteScript), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse() failed: %v", err)
	}
	if err := api.LoadFromUnit(interp, &api.LoadedUnit{Tree: tree}); err != nil {
		t.Fatalf("api.LoadFromUnit() failed: %v", err)
	}

	_, err = api.RunProcedure(context.Background(), interp, "main", targetFile, "should not write")

	if err == nil {
		t.Fatal("api.RunProcedure() succeeded but was expected to fail with a policy error.")
	}

	var rtErr *lang.RuntimeError
	if errors.As(err, &rtErr) {
		if rtErr.Code != lang.ErrorCodePolicy {
			t.Errorf("Expected error code %v (ErrorCodePolicy), but got %v", lang.ErrorCodePolicy, rtErr.Code)
		}
	} else {
		t.Errorf("Expected a *lang.RuntimeError, but got a different error type: %v", err)
	}
}
