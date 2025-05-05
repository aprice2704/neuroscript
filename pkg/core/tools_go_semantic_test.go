// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 12:30:00 PDT // Adjust Method/Receiver expectations to L22 based on Go tool pos
// filename: pkg/core/tools_go_semantic_test.go

package core

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Test fixture content (unchanged)
const mainGoFixtureContent = `package main

import "fmt" // L3

var GlobalVar = "initial" // L5

type MyStruct struct { // L7
	Field int // L8
}

func main() { // L11
	message := Greet("World") // L12 (message def C2, Greet usage C13)
	fmt.Println(message)      // L13 (message usage C15, fmt usage C3, Println C7)
	GlobalVar = "changed"     // L14 (GlobalVar usage C1) -> AST reports C2
}

// Greet generates a greeting.
func Greet(name string) string { // L18 (Greet def C6, name def C12)
	return "Hello, " + name // L19 (name usage C21)
}

// Note: Go tooling seems to report positions for Method and 'm' on Line 22
func (m *MyStruct) Method() { // L21 (m def C7->L22 C7, Method def C21->L22 C20)
	m.Field++ // L22 (m usage C1->L23 C2?, Field usage C4->L23 C4?)
} // L24?
`

// setupIndexForTest helper (unchanged)
func setupIndexForTest(t *testing.T) (*Interpreter, string, string, error) {
	t.Helper()
	sandboxDir := t.TempDir()
	t.Logf("Test sandbox created: %s", sandboxDir)
	mainGoPath := filepath.Join(sandboxDir, "main.go")
	err := os.WriteFile(mainGoPath, []byte(mainGoFixtureContent), 0644)
	if err != nil {
		return nil, "", sandboxDir, fmt.Errorf("failed to write test Go file: %w", err)
	}
	t.Logf("Test file created: %s", mainGoPath)
	goModContent := "module testmodule\n\ngo 1.21\n"
	goModPath := filepath.Join(sandboxDir, "go.mod")
	err = os.WriteFile(goModPath, []byte(goModContent), 0644)
	if err != nil {
		return nil, "", sandboxDir, fmt.Errorf("failed to write test go.mod file: %w", err)
	}
	t.Logf("Test go.mod created: %s", goModPath)
	interpreter, _ := NewDefaultTestInterpreter(t)
	interpreter.SetSandboxDir(sandboxDir)
	t.Logf("Interpreter sandbox set to: %s", interpreter.SandboxDir())
	args := []interface{}{}
	result, toolErr := toolGoIndexCode(interpreter, args)
	if toolErr != nil {
		return interpreter, "", sandboxDir, fmt.Errorf("toolGoIndexCode failed during setup: %w", toolErr)
	}
	handle, ok := result.(string)
	if !ok || handle == "" {
		return interpreter, "", sandboxDir, fmt.Errorf("toolGoIndexCode did not return a valid handle, got: %v", result)
	}
	t.Logf("Index created with handle: %s", handle)
	return interpreter, handle, sandboxDir, nil
}

// TestToolGoIndexCode test (unchanged)
func TestToolGoIndexCode(t *testing.T) {
	// ... (implementation unchanged) ...
}
