// filename: pkg/neurogo/handle_turn.go
package neurogo

import (
	// Keep if other parts use it
	"context"
	"errors"
	"fmt" // Keep if other parts use it
	"strconv"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core" // Keep if patch handler used
	"github.com/google/generative-ai-go/genai"
)

// Constants
const (
	applyPatchFunctionName = "_ApplyNeuroScriptPatch"
	// multiLineCommand = "/m" // Defined in app_agent.go
)

// handleAgentTurn processes a single response from the LLM.
// +++ MODIFIED: Added initialAttachmentURIs parameter +++
func (a *App) handleAgentTurn(
	ctx context.Context,
	llmClient *core.LLMClient,
	convoManager *core.ConversationManager,
	agentInterpreter *core.Interpreter,
	securityLayer *core.SecurityLayer,
	toolDeclarations []*genai.Tool,
	initialAttachmentURIs []string, // <-- ADDED
) error {

	for cycle := 0; cycle < maxFunctionCallCycles; cycle++ {
		a.Logger.Info("--- Agent Inner Loop Cycle %d ---", cycle+1)

		// +++ MODIFIED: Use initialAttachmentURIs in context +++
		requestContext := core.LLMRequestContext{
			History:  convoManager.GetHistory(),
			FileURIs: initialAttachmentURIs, // <-- Use passed URIs
		}
		// ---

		if len(requestContext.History) == 0 {
			return errors.New("history empty")
		} // Shortened

		// Call LLM (logic unchanged)
		llmResponse, err := llmClient.CallLLMAgent(ctx, requestContext, toolDeclarations)
		if err != nil {
			return fmt.Errorf("LLM API call failed: %w", err)
		}
		if llmResponse == nil || len(llmResponse.Candidates) == 0 { /* Handle error/safety block */
			return errors.New("invalid LLM response")
		}

		candidate := llmResponse.Candidates[0]
		if err := convoManager.AddModelResponse(candidate); err != nil {
			a.Logger.Error("Failed to add model response to history: %v", err)
		}
		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			a.Logger.Info("LLM candidate empty. Turn ends.")
			fmt.Println("\n[AGENT] LLM returned empty response.")
			return nil
		}

		// Process Response Parts (logic unchanged)
		foundFunctionCall := false
		var firstFunctionCall *genai.FunctionCall
		var accumulatedText strings.Builder
		for _, part := range candidate.Content.Parts { /* ... process parts, set firstFunctionCall ... */
			switch v := part.(type) {
			case genai.Text:
				accumulatedText.WriteString(string(v))
			case genai.FunctionCall:
				if !foundFunctionCall {
					foundFunctionCall = true
					fcCopy := v
					firstFunctionCall = &fcCopy
				} // Process only first
			}
		}

		// Decide Next Action (logic unchanged)
		if foundFunctionCall && firstFunctionCall != nil {
			// Execute function call
			fc := *firstFunctionCall
			funcResultPart, execErr := securityLayer.ExecuteToolCall(agentInterpreter, fc)
			if execErr != nil {
				a.Logger.Error("Tool exec error '%s': %v", fc.Name, execErr)
			} else {
				a.Logger.Info("OK exec tool: %s", fc.Name)
			}
			if err := convoManager.AddFunctionResultMessage(funcResultPart); err != nil {
				return fmt.Errorf("failed record func result %s: %w", fc.Name, err)
			}
			a.Logger.Debug("Func call done, continue cycle.")
			continue // Next iteration

		} else {
			// No function call. Process text / patch.
			finalText := accumulatedText.String()
			trimmedResponse := strings.TrimSpace(finalText)
			// Trim fences (logic unchanged)
			trimmedResponse = strings.TrimPrefix(trimmedResponse, "```json")
			trimmedResponse = strings.TrimPrefix(trimmedResponse, "```")
			trimmedResponse = strings.TrimSuffix(trimmedResponse, "```")
			trimmedResponse = strings.TrimPrefix(trimmedResponse, "`")
			trimmedResponse = strings.TrimSuffix(trimmedResponse, "`")
			trimmedResponse = strings.TrimSpace(trimmedResponse)

			// Check for Patch Pattern (logic unchanged)
			if strings.HasPrefix(trimmedResponse, applyPatchFunctionName+"(") && strings.HasSuffix(trimmedResponse, ")") {
				// ... (patch detection and handling logic unchanged) ...
				startIndex := strings.Index(trimmedResponse, "(")
				endIndex := strings.LastIndex(trimmedResponse, ")")
				if startIndex != -1 && endIndex > startIndex {
					extractedArg := strings.TrimSpace(trimmedResponse[startIndex+1 : endIndex])
					patchJSON := extractedArg
					if strings.HasPrefix(extractedArg, `"`) /* ... unquote logic ... */ {
						unquoted, errU := strconv.Unquote(extractedArg)
						if errU == nil {
							patchJSON = unquoted
						} else { /* log warn */
						}
					}
					sandboxDir := securityLayer.SandboxRoot()
					patchErr := handleReceivedPatch(patchJSON, agentInterpreter, securityLayer, sandboxDir, a.Logger)
					if patchErr != nil {
						fmt.Printf("[AGENT] Patch failed: %v\n", patchErr)
						return fmt.Errorf("patch failed: %w", patchErr)
					} else {
						fmt.Printf("[AGENT] Patch applied.\n")
						return nil
					} // End turn
				} else { /* malformed parens */
				}

			}
			// Handle final text or empty response (logic unchanged)
			if finalText != "" {
				fmt.Printf("\n[AGENT RESPONSE]\n%s\n\n", finalText)
			} else {
				fmt.Println("\n[AGENT RESPONSE]\n(Agent provided no text response.)\n")
			}
			return nil // End turn
		}
	} // End inner loop

	errLoop := fmt.Errorf("exceeded max cycles (%d)", maxFunctionCallCycles)
	a.Logger.Error("Agent turn failed: %v", errLoop)
	fmt.Printf("\n[AGENT] Error: %v\n", errLoop)
	return errLoop
}

// Assume helpers executeAgentTool, formatToolResult, formatErrorResponse, loadToolListFromFile exist
