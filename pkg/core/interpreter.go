// filename: pkg/core/interpreter.go
package core

import (
	"errors"
	"fmt"
	"io"
	"log"

	// "os" // No longer needed for API key here

	"github.com/aprice2704/neuroscript/pkg/core/prompts"
	"github.com/google/generative-ai-go/genai"
	// "google.golang.org/api/option" // No longer needed here
)

// --- Interpreter ---
type Interpreter struct {
	variables       map[string]interface{}
	knownProcedures map[string]Procedure
	lastCallResult  interface{}
	vectorIndex     map[string][]float32
	embeddingDim    int
	currentProcName string
	sandboxDir      string
	toolRegistry    *ToolRegistry
	logger          *log.Logger
	objectCache     map[string]interface{}
	handleTypes     map[string]string
	// --- REMOVED: genaiClient ---
	// genaiClient     *genai.Client
	// +++ ADDED: LLMClient field +++
	llmClient *LLMClient
	modelName string // Keep modelName for non-LLMClient use cases? Maybe remove? Let's keep for now.
}

// --- MODIFIED: Setter for Model Name might be less relevant if LLMClient manages it ---
// Consider if this should configure the LLMClient's model instead.
func (i *Interpreter) SetModelName(name string) error {
	if name == "" {
		return errors.New("model name cannot be empty")
	}
	i.modelName = name
	i.logger.Printf("[INFO INTERP] Interpreter model name set to: %s (Note: LLMClient might use its own)", name)
	// TODO: Decide if this should also attempt to update i.llmClient's model if i.llmClient is not nil.
	return nil
}

// ToolRegistry getter (unchanged)
func (i *Interpreter) ToolRegistry() *ToolRegistry {
	if i.toolRegistry == nil {
		if i.logger != nil {
			i.logger.Println("[WARN] ToolRegistry accessed before initialization, creating new one.")
		}
		i.toolRegistry = NewToolRegistry()
	}
	return i.toolRegistry
}

// --- MODIFIED: GenAIClient getter uses LLMClient ---
// GenAIClient returns the underlying *genai.Client from the LLMClient instance.
// Useful for tools that interact directly with the base client (e.g., File API).
func (i *Interpreter) GenAIClient() *genai.Client {
	if i.llmClient == nil {
		// Log this potential issue
		if i.logger != nil {
			i.logger.Println("[WARN INTERP] GenAIClient() called but internal LLMClient is nil.")
		}
		return nil
	}
	// Use the getter method we added to LLMClient
	return i.llmClient.Client()
}

// --- END MODIFIED Getter ---

// AddProcedure (unchanged)
func (i *Interpreter) AddProcedure(proc Procedure) error {
	if i.knownProcedures == nil {
		i.knownProcedures = make(map[string]Procedure)
	}
	if _, exists := i.knownProcedures[proc.Name]; exists {
		return fmt.Errorf("procedure '%s' already defined", proc.Name)
	}
	i.knownProcedures[proc.Name] = proc
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP] Added procedure '%s' to known procedures.", proc.Name)
	}
	return nil
}

// KnownProcedures getter (unchanged)
func (i *Interpreter) KnownProcedures() map[string]Procedure {
	if i.knownProcedures == nil {
		return make(map[string]Procedure) // Return empty map if not initialized
	}
	return i.knownProcedures
}

// Logger getter (unchanged)
func (i *Interpreter) Logger() *log.Logger {
	if i.logger == nil {
		// Ensure a non-nil logger is always returned
		return log.New(io.Discard, "", 0)
	}
	return i.logger
}

// Vector Index methods (unchanged)
func (i *Interpreter) GetVectorIndex() map[string][]float32   { /*...*/ return i.vectorIndex }
func (i *Interpreter) SetVectorIndex(vi map[string][]float32) { /*...*/ i.vectorIndex = vi }

// Handle Cache methods (unchanged)
func (i *Interpreter) storeObjectInCache(obj interface{}, typeTag string) string { /*...*/ return "" }
func (i *Interpreter) retrieveObjectFromCache(handleID string, expectedTypeTag string) (interface{}, error) { /*...*/
	return nil, nil
}
func (i *Interpreter) getCachedObjectAndType(handleID string) (object interface{}, typeTag string, found bool) { /*...*/
	return nil, "", false
}

// --- MODIFIED: NewInterpreter accepts LLMClient ---
func NewInterpreter(logger *log.Logger, llmClient *LLMClient) *Interpreter {
	effectiveLogger := logger
	if effectiveLogger == nil {
		effectiveLogger = log.New(io.Discard, "", 0)
	}

	// --- REMOVED: Direct GenAI Client Init ---

	interp := &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16, // Default or configure?
		toolRegistry:    NewToolRegistry(),
		logger:          effectiveLogger,
		objectCache:     make(map[string]interface{}),
		handleTypes:     make(map[string]string),
		// --- MODIFIED: Store passed LLMClient ---
		llmClient: llmClient,
		// Initialize modelName from client if possible, otherwise default
		modelName: "gemini-1.5-pro-latest", // Default, may be overridden
	}
	// Optionally sync modelName if client is valid
	if llmClient != nil && llmClient.modelName != "" {
		interp.modelName = llmClient.modelName
	}

	interp.variables["NEUROSCRIPT_DEVELOP_PROMPT"] = prompts.PromptDevelop
	interp.variables["NEUROSCRIPT_EXECUTE_PROMPT"] = prompts.PromptExecute

	return interp
}

// --- END MODIFIED NewInterpreter ---

// RunProcedure (unchanged)
func (i *Interpreter) RunProcedure(procName string, args ...string) (result interface{}, err error) {
	// ... (implementation unchanged) ...
	return result, err
}

// executeSteps (unchanged)
func (i *Interpreter) executeSteps(steps []Step) (result interface{}, wasReturn bool, err error) {
	// ... (implementation unchanged) ...
	return result, wasReturn, err
}

// executeBlock (unchanged)
func (i *Interpreter) executeBlock(blockValue interface{}, parentStepNum int, blockType string) (result interface{}, wasReturn bool, err error) {
	// ... (implementation unchanged) ...
	return result, wasReturn, err
}
