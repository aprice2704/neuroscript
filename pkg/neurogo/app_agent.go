// filename: pkg/neurogo/app_agent.go
package neurogo

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	// Import interfaces for Logger (and potentially LLMClient if needed directly)
)

// runAgentMode starts the interactive agent loop.
func (app *App) runAgentMode(ctx context.Context) error {
	app.Log.Info("Entering Agent Mode...")

	if app.llmClient == nil {
		return fmt.Errorf("cannot run Agent Mode: LLM client is not initialized")
	}

	// Initialize conversation history
	conversation := core.NewConversation() // Assuming NewConversation exists

	// Agent execution context
	agentCtx := NewAgentContext(app.Log, app.interpreter, app.llmClient, conversation)

	// Register built-in tools for the agent
	app.registerAgentTools(agentCtx)

	// Main interactive loop
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Agent ready. Type your request or 'exit' to quit.")

	for {
		fmt.Print("> ")
		userInput, err := reader.ReadString('\n')
		if err != nil {
			app.Log.Error("Error reading user input", "error", err)
			return fmt.Errorf("failed to read input: %w", err)
		}
		userInput = strings.TrimSpace(userInput)

		if strings.ToLower(userInput) == "exit" {
			fmt.Println("Exiting agent mode.")
			break
		}

		if userInput == "" {
			continue
		}

		// Add user input to conversation
		conversation.AddTurn(core.RoleUser, userInput)

		// Process the turn using the agent context
		// The handleTurn function encapsulates the core agent logic
		err = app.handleTurn(ctx, agentCtx)
		if err != nil {
			app.Log.Error("Error handling turn", "error", err)
			fmt.Println("An error occurred:", err)
			// Decide whether to continue or exit on error
			// For now, let's continue the loop
			// Optionally remove the last user turn if handling failed critically?
		}

		// Display the latest assistant response (or tool results) from the conversation
		lastTurn := conversation.LastTurn()
		if lastTurn != nil {
			// Simple display, TUI mode would format this better
			fmt.Printf("[%s]: %s\n", lastTurn.Role, lastTurn.Content)
			if len(lastTurn.ToolCalls) > 0 {
				fmt.Println("Tool Calls Requested:")
				for _, tc := range lastTurn.ToolCalls {
					fmt.Printf("  - %s(%v)\n", tc.Name, tc.Arguments)
				}
			}
			if len(lastTurn.ToolResults) > 0 {
				fmt.Println("Tool Results:")
				for _, tr := range lastTurn.ToolResults {
					fmt.Printf("  - ID %s: %v\n", tr.ID, tr.Result) // Displaying raw result
				}
			}
		}

		// TODO: Add context management (e.g., pruning conversation history)
	}

	return nil
}

// registerAgentTools registers the tools available to the agent.
func (app *App) registerAgentTools(agentCtx *AgentContext) {
	app.Log.Info("Registering agent tools...")

	// Example: Registering a simple echo tool (implementation would be elsewhere)
	// agentCtx.RegisterTool(core.ToolDefinition{
	// 	 Name:        "echo",
	// 	 Description: "Echoes the input message back.",
	// 	 InputSchema: map[string]any{
	// 		 "type": "object",
	// 		 "properties": map[string]any{
	// 			 "message": map[string]any{"type": "string", "description": "The message to echo."},
	// 		 },
	// 		 "required": []string{"message"},
	// 	 },
	// }, func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// 	 message, ok := args["message"].(string)
	// 	 if !ok {
	// 		 return nil, fmt.Errorf("invalid 'message' argument type: %T", args["message"])
	// 	 }
	// 	 return message, nil
	// })

	// Register tools defined in agent_tools.go
	RegisterCoreAgentTools(agentCtx)

	app.Log.Info("Agent tools registered.")
}

// handleTurn processes a single turn of the conversation.
// This is a placeholder and should delegate to handle_turn.go logic.
func (app *App) handleTurn(ctx context.Context, agentCtx *AgentContext) error {
	app.Log.Debug("Handling agent turn...")
	// This function should contain the logic currently in pkg/neurogo/handle_turn.go
	// Call app.processAgentTurn or similar function defined in handle_turn.go
	return app.processAgentTurn(ctx, agentCtx) // Assuming processAgentTurn exists in handle_turn.go
}

// Helper function to get available tools from the agent context
// This might be better placed within AgentContext itself.
func getAvailableTools(agentCtx *AgentContext) []core.ToolDefinition {
	agentCtx.mu.RLock()
	defer agentCtx.mu.RUnlock()
	tools := make([]core.ToolDefinition, 0, len(agentCtx.tools))
	for _, t := range agentCtx.tools {
		tools = append(tools, t.Definition)
	}
	return tools
}
