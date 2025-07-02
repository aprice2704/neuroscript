// NeuroScript Version: 0.4.2
// File version: 1
// Purpose: Contains getters and setters for various internal state properties of the interpreter.
// filename: pkg/runtime/interpreter_state.go
// nlines: 77
// risk_rating: LOW
package runtime

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/google/generative-ai-go/genai"
)

func (i *Interpreter) SetAIWorkerManager(manager *AIWorkerManager) {
	i.aiWorkerManager = manager
}

func (i *Interpreter) AIWorkerManager() *AIWorkerManager {
	return i.aiWorkerManager
}

func (i *Interpreter) SandboxDir() string { return i.sandboxDir }

func (i *Interpreter) Logger() interfaces.Logger {
	if i.logger == nil {
		panic("FATAL: Interpreter logger is nil")
	}
	return i.logger
}

func (i *Interpreter) FileAPI() *FileAPI {
	if i.fileAPI == nil {
		panic("FATAL: Interpreter fileAPI not initialized")
	}
	return i.fileAPI
}

func (i *Interpreter) SetSandboxDir(newSandboxDir string) error {
	var cleanNewSandboxDir string
	if newSandboxDir == "" {
		cleanNewSandboxDir = "."
	} else {
		absPath, err := filepath.Abs(newSandboxDir)
		if err != nil {
			return fmt.Errorf("invalid sandbox directory '%s': %w", newSandboxDir, err)
		}
		cleanNewSandboxDir = filepath.Clean(absPath)
	}
	if i.sandboxDir != cleanNewSandboxDir {
		i.sandboxDir = cleanNewSandboxDir
		i.fileAPI = NewFileAPI(i.sandboxDir, i.logger)
	}
	return nil
}

func (i *Interpreter) SetVariable(name string, value Value) error {
	i.variablesMu.Lock()
	defer i.variablesMu.Unlock()
	if i.variables == nil {
		i.variables = make(map[string]Value)
	}
	if name == "" {
		return errors.New("variable name cannot be empty")
	}
	i.variables[name] = value
	return nil
}

func (i *Interpreter) GetVariable(name string) (Value, bool) {
	i.variablesMu.RLock()
	defer i.variablesMu.RUnlock()
	if i.variables == nil {
		return nil, false
	}
	val, exists := i.variables[name]
	return val, exists
}

func (i *Interpreter) GenAIClient() *genai.Client {
	if i.llmClient == nil {
		return nil
	}
	return i.llmClient.Client()
}

func (i *Interpreter) GetVectorIndex() map[string][]float32 {
	if i.vectorIndex == nil {
		i.vectorIndex = make(map[string][]float32)
	}
	return i.vectorIndex
}

func (i *Interpreter) SetVectorIndex(vi map[string][]float32) { i.vectorIndex = vi }
