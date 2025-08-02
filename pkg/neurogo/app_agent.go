// NeuroScript Version: 0.3.1
// File version: 0.2.4
// Corrected all remaining compiler errors from typos and refactoring.
// filename: pkg/neurogo/app_agent.go
package neurogo

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/llm"
	"github.com/aprice2704/neuroscript/pkg/security"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/google/generative-ai-go/genai"
)

// runAgentMode starts the interactive agent mode.
func (app *App) runAgentMode(ctx context.Context) error {
	app.Log.Info("--- Running in Agent Mode ---")
	if app.llmClient == nil {
		return fmt.Errorf("cannot run agent mode: LLM client is nil")
	}
	if app.interpreter == nil {
		return fmt.Errorf("cannot run agent mode: Interpreter is nil")
	}

	convoManager := llm.NewConversationManager(app.Log)
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
	}

	fmt.Println("Agent mode finished.")
	return nil
}

// registerAgentTools registers tools specifically needed for the agent mode.
func (app *App) registerAgentTools(agentCtx *AgentContext) error {
	app.Log.Info("Registering agent tools...")
	if app.interpreter == nil {
		return fmt.Errorf("cannot register agent tools: interpreter is nil")
	}
	registry := app.interpreter.ToolRegistry()
	if registry == nil {
		return fmt.Errorf("interpreter tool registry is nil")
	}

	err := RegisterAgentTools(registry)
	if err != nil {
		return fmt.Errorf("failed during agent tool registration: %w", err)
	}

	app.Log.Info("Agent tools registered.")
	return nil
}

// handleTurn processes a single turn of the conversation.
func (app *App) handleTurn(ctx context.Context, convoManager *llm.ConversationManager, agentCtx *AgentContext) error {
	app.Log.Debug("Handling agent turn...")

	llmClient := app.GetLLMClient()
	if app.interpreter == nil {
		app.Log.Error("Interpreter is nil within handleTurn.")
		return fmt.Errorf("cannot handle turn: interpreter is nil")
	}
	if llmClient == nil {
		return fmt.Errorf("cannot handle turn: LLM client is nil")
	}

	registry := app.interpreter.ToolRegistry()
	if registry == nil {
		app.Log.Error("Interpreter's ToolRegistry is nil.")
		return fmt.Errorf("cannot handle turn: tool registry is nil")
	}

	var allowedTools security.ADlist = nil
	var deniedTools security.ADlist = nil
	//var allowedPaths map[string]bool = nil
	sandboxDir := app.interpreter.SandboxDir()
	securityLayer := security.NewSecurityLayer(
		allowedTools,
		deniedTools,
		sandboxDir,
		registry,
		app.Log,
	)
	if securityLayer == nil {
		return fmt.Errorf("failed to create security layer")
	}

	availableTools := getAvailableTools(agentCtx, registry)

	fileInfoList := agentCtx.GetURIsForNextContext()
	stringURIs := make([]string, 0, len(fileInfoList))
	for _, fileInfo := range fileInfoList {
		if fileInfo != nil && fileInfo.URI != "" {
			stringURIs = append(stringURIs, fileInfo.URI)
		}
	}
	accumulatedContextURIsForCall := stringURIs

	err := app.handleAgentTurn(
		ctx,
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
		return fmt.Errorf("cannot execute startup script: interpreter is nil")
	}

	filepathArg, err := lang.Wrap(scriptPath)
	if err != nil {
		return fmt.Errorf("internal error wrapping script path '%s': %w", scriptPath, err)
	}
	toolArgs := map[string]lang.Value{"filepath": filepathArg}
	// CORRECTED: Use the canonical tool name 'fs.read'
	contentValue, err := app.interpreter.ExecuteTool("fs.read", toolArgs)
	if err != nil {
		return fmt.Errorf("failed to read startup script file '%s': %w", scriptPath, err)
	}
	scriptContent, ok := lang.Unwrap(contentValue).(string)
	if !ok {
		return fmt.Errorf("internal error: 'fs.read' did not return a string for '%s'", scriptPath)
	}

	if _, err := app.LoadScriptString(ctx, scriptContent); err != nil {
		return fmt.Errorf("failed to load startup script %s: %w", scriptPath, err)
	}
	app.Log.Debug("Startup script processed and loaded.", "path", scriptPath)

	startupProcName := "main"
	app.Log.Info("Running startup procedure.", "name", startupProcName, "script", scriptPath)
	_, runErr := app.RunProcedure(ctx, startupProcName, nil)

	if runErr != nil {
		var rErr *lang.RuntimeError
		if errors.As(runErr, &rErr) && rErr.Code == lang.ErrorCodeProcNotFound {
			app.Log.Warn("No 'main' procedure found in startup script, nothing to execute.", "path", scriptPath)
			return nil
		}
		app.Log.Error("Error running startup procedure.", "proc", startupProcName, "error", runErr)
		return fmt.Errorf("error running startup procedure '%s' from %s: %w", startupProcName, scriptPath, runErr)
	}

	app.Log.Info("Startup procedure finished successfully.", "name", startupProcName)
	return nil
}

// getAvailableTools prepares the list of genai.Tools for the LLM call.
func getAvailableTools(agentCtx *AgentContext, registry tool.ToolRegistry) []*genai.Tool {
	if registry == nil {
		fmt.Println("[AGENT] Warning: Tool registry is nil in getAvailableTools.")
		return []*genai.Tool{}
	}
	allTools := registry.ListTools()
	genaiTools := make([]*genai.Tool, 0, len(allTools))
	for _, toolImpl := range allTools {
		genaiFunc := &genai.FunctionDeclaration{
			Name:        string(toolImpl.FullName),
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
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
