// NeuroScript Version: 0.3.0
// File version: 0.1.4
// Correct ToolRegistry type in getAvailableTools
// filename: pkg/neurogo/app_agent.go
// nlines: 285
// risk_rating: MEDIUM
package neurogo

import (
	"bufio"
	"context" // Keep context for other parts of the file
	"fmt"
	"os"
	"strings"

	// "time" // No longer needed directly in this file

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/generative-ai-go/genai"
)

// runAgentMode starts the interactive agent mode.
func (app *App) runAgentMode(ctx context.Context) error {
	// ... (setup unchanged) ...
	app.Log.Info("--- Running in Agent Mode ---")
	if app.llmClient == nil {
		return fmt.Errorf("cannot run agent mode: LLM client is nil")
	}
	if app.interpreter == nil {
		return fmt.Errorf("cannot run agent mode: Interpreter is nil")
	}

	convoManager := core.NewConversationManager(app.Log)
	if convoManager == nil {
		return fmt.Errorf("failed to create conversation manager")
	}

	agentCtx := NewAgentContext(app.Log)
	if agentCtx == nil {
		return fmt.Errorf("failed to create agent context")
	}
	if app.interpreter != nil {
		agentCtx.SetSandboxDir(app.Config.SandboxDir)
	} else {
		return fmt.Errorf("cannot set agent sandbox: interpreter is nil")
	}

	if app.interpreter != nil {
		err := app.registerAgentTools(agentCtx)
		if err != nil {
			app.Log.Error("Failed to register agent tools", "error", err)
		}
	} else {
		return fmt.Errorf("cannot register agent tools: interpreter is nil")
	}

	if app.Config.StartupScript != "" {
		app.Log.Info("Executing startup script.", "path", app.Config.StartupScript)
		if app.interpreter != nil {
			// Pass ctx here for potential future use, even if RunProcedure doesn't take it now
			err := app.executeStartupScript(ctx, app.Config.StartupScript, agentCtx)
			if err != nil {
				app.Log.Error("Failed to execute startup script.", "path", app.Config.StartupScript, "error", err)
				fmt.Printf("[AGENT] Warning: Startup script '%s' failed: %v\n", app.Config.StartupScript, err)
			} else {
				app.Log.Info("Startup script executed successfully.", "path", app.Config.StartupScript)
			}
		} else {
			app.Log.Error("Cannot execute startup script: interpreter is nil.")
		}
	} else {
		app.Log.Info("No startup script specified.")
	}

	fmt.Println("Entering interactive agent mode. Type 'exit' or 'quit' to end.")
	reader := bufio.NewReader(os.Stdin)

	for {
		// ... (input loop unchanged) ...
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			app.Log.Error("Error reading user input", "error", err)
			fmt.Println("Error reading input:", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if strings.ToLower(input) == "quit" || strings.ToLower(input) == "exit" {
			app.Log.Info("Exiting agent mode.")
			break
		}

		convoManager.AddUserMessage(input)

		err = app.handleTurn(ctx, convoManager, agentCtx)
		if err != nil {
			app.Log.Error("Error handling agent turn", "error", err)
			fmt.Println("Error processing turn:", err)
		}

		// Display the latest model response
		history := convoManager.GetHistory()
		if len(history) > 0 {
			lastContent := history[len(history)-1]
			contentRole := lastContent.Role
			if contentRole == "model" {
				var modelTextResponse strings.Builder
				if lastContent.Parts != nil {
					for _, part := range lastContent.Parts {
						if textPart, ok := part.(genai.Text); ok {
							modelTextResponse.WriteString(string(textPart))
						}
					}
				}
				responseText := modelTextResponse.String()
				if responseText != "" {
					fmt.Println("<", responseText)
				} else if err == nil {
					fmt.Println("< (Turn processed, check logs for details or tool output)")
				}
			} else if err == nil {
				fmt.Println("< (Waiting for model response...)")
			}
		} else if err == nil {
			fmt.Println("< (No response and no history?)")
		}
		fmt.Println()

	} // End input loop

	fmt.Println("Agent mode finished.")
	return nil
}

// registerAgentTools registers tools specifically needed for the agent mode.
func (app *App) registerAgentTools(agentCtx *AgentContext) error {
	// ... (unchanged) ...
	app.Log.Info("Registering agent tools...")
	if app.interpreter == nil {
		return fmt.Errorf("cannot register agent tools: interpreter is nil")
	}
	// app.interpreter.ToolRegistry() returns core.ToolRegistry (interface type)
	registry := app.interpreter.ToolRegistry()
	if registry == nil {
		return fmt.Errorf("interpreter tool registry is nil")
	}

	// This call is now correct as RegisterAgentTools expects core.ToolRegistry
	err := RegisterAgentTools(registry)
	if err != nil {
		return fmt.Errorf("failed during agent tool registration: %w", err)
	}

	app.Log.Info("Agent tools registered.")
	return nil
}

// handleTurn processes a single turn of the conversation.
func (app *App) handleTurn(ctx context.Context, convoManager *core.ConversationManager, agentCtx *AgentContext) error {
	// ... (setup unchanged) ...
	app.Log.Debug("Handling agent turn...")

	llmClient := app.GetLLMClient()
	if app.interpreter == nil {
		app.Log.Error("Interpreter is nil within handleTurn.")
		return fmt.Errorf("cannot handle turn: interpreter is nil")
	}
	if llmClient == nil {
		return fmt.Errorf("cannot handle turn: LLM client is nil")
	}

	// app.interpreter.ToolRegistry() returns core.ToolRegistry (interface type)
	registry := app.interpreter.ToolRegistry()
	if registry == nil {
		app.Log.Error("Interpreter's ToolRegistry is nil.")
		return fmt.Errorf("cannot handle turn: tool registry is nil")
	}

	var allowedTools []string = nil
	var allowedPaths map[string]bool = nil
	sandboxDir := app.interpreter.SandboxDir()
	securityLayer := core.NewSecurityLayer(
		allowedTools,
		allowedPaths,
		sandboxDir,
		registry, // registry is core.ToolRegistry, which is expected by NewSecurityLayer if it takes the interface
		app.Log,
	)
	if securityLayer == nil {
		return fmt.Errorf("failed to create security layer")
	}

	// This call is now correct as getAvailableTools expects core.ToolRegistry
	availableTools := getAvailableTools(agentCtx, registry)

	fileInfoList := agentCtx.GetURIsForNextContext()
	stringURIs := make([]string, 0, len(fileInfoList))
	for _, fileInfo := range fileInfoList {
		if fileInfo != nil && fileInfo.URI != "" {
			stringURIs = append(stringURIs, fileInfo.URI)
		}
	}
	accumulatedContextURIsForCall := stringURIs

	// Call handleAgentTurn, passing app.interpreter
	// Note: handleAgentTurn might need context passed if *it* calls context-aware funcs
	err := app.handleAgentTurn(
		ctx, // Pass context along
		llmClient,
		convoManager,
		app.interpreter,
		securityLayer,
		availableTools,
		accumulatedContextURIsForCall,
	)
	if err != nil {
		app.Log.Error("handleAgentTurn implementation failed", "error", err)
		return err
	}

	app.Log.Debug("Agent turn processing complete in handleTurn.")
	return nil
}

// executeStartupScript handles running the initial agent configuration script.
func (app *App) executeStartupScript(ctx context.Context, scriptPath string, agentCtx *AgentContext) error {
	app.Log.Info("Executing startup script.", "path", scriptPath)

	if app.interpreter == nil {
		app.Log.Error("Interpreter is nil before executing startup script.")
		return fmt.Errorf("cannot execute startup script: interpreter is nil")
	}

	procDefs, fileMeta, err := app.processNeuroScriptFile(scriptPath, app.interpreter)
	if err != nil {
		return fmt.Errorf("failed to process startup script %s: %w", scriptPath, err)
	}
	app.Log.Debug("Startup script processed.", "path", scriptPath, "procedures_found", len(procDefs), "metadata", fileMeta)

	startupProcName := "main"
	found := false
	for _, proc := range procDefs {
		if proc != nil && proc.Name == startupProcName {
			found = true
			break
		}
	}

	if !found {
		app.Log.Warn("No 'main' procedure found in startup script, nothing to execute.", "path", scriptPath)
		return nil
	}

	app.Log.Info("Running startup procedure.", "name", startupProcName, "script", scriptPath)

	var results interface{}
	procName := startupProcName
	// arguments map is nil, so we pass no variadic args

	// <<< FIX: Call RunProcedure matching gopls signature (procName string, args ...interface{}) >>>
	results, runErr := app.interpreter.RunProcedure(procName) // Pass only procName

	if runErr != nil {
		// Don't need the specific context error check anymore
		app.Log.Error("Error running startup procedure.", "proc", startupProcName, "error", runErr)
		return fmt.Errorf("error running startup procedure '%s' from %s: %w", startupProcName, scriptPath, runErr)
	}

	app.Log.Info("Startup procedure finished.", "name", startupProcName, "result_type", fmt.Sprintf("%T", results))
	return nil
}

// getAvailableTools prepares the list of genai.Tools for the LLM call.
// CORRECTED: Changed registry type from *core.ToolRegistry to core.ToolRegistry
func getAvailableTools(agentCtx *AgentContext, registry core.ToolRegistry) []*genai.Tool {
	// ... (unchanged) ...
	if registry == nil {
		fmt.Println("[AGENT] Warning: Tool registry is nil in getAvailableTools.")
		return []*genai.Tool{}
	}

	// This call is now correct because 'registry' is core.ToolRegistry (interface)
	// and 'ListTools' is a method on that interface.
	allTools := registry.ListTools()
	genaiTools := make([]*genai.Tool, 0, len(allTools))
	for _, toolImpl := range allTools {
		// Use qualified name for declaration if tools are registered/called that way
		qualifiedName := "TOOL." + toolImpl.Name // Assuming tools need TOOL. prefix for LLM
		genaiFunc := &genai.FunctionDeclaration{
			Name:        qualifiedName,
			Description: toolImpl.Description,
		}
		genaiTools = append(genaiTools, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{genaiFunc},
		})
	}
	return genaiTools
}

// snippet returns the first n characters of a string.
func snippet(s string, n int) string {
	// ... (unchanged) ...
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
