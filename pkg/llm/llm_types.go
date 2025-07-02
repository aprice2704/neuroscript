// filename: pkg/core/llm_types.go
package core

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	// TokenUsageMetrics is defined in ai_worker_types.go, ensure it's accessible
	// If not directly, this implies ai_worker_types.go is in the same package or imported.
	// Assuming it's accessible as `TokenUsageMetrics` directly or as `core.TokenUsageMetrics`
	// For this file, direct accessibility implies it's in the same package 'core',
	// which aligns with `ai_worker_types.go` also being in `package core`.
)

// String returns a string representation of the interfaces.ConversationTurn.
func String(t *interfaces.ConversationTurn) string {
	base := fmt.Sprintf("[%s]: %s", t.Role, t.Content)
	if len(t.ToolCalls) > 0 {
		calls := ""
		for _, tc := range t.ToolCalls {
			calls += fmt.Sprintf("\n  interfaces.ToolCall(ID: %s, Name: %s, Args: %v)", tc.ID, tc.Name, tc.Arguments)
		}
		base += calls
	}
	if len(t.ToolResults) > 0 {
		results := ""
		for _, tr := range t.ToolResults {
			resStr := fmt.Sprintf("%v", tr.Result)
			if tr.Error != "" {
				resStr = fmt.Sprintf("Error: %s", tr.Error)
			}
			results += fmt.Sprintf("\n  interfaces.ToolResult(ID: %s, Result: %s)", tr.ID, resStr)
		}
		base += results
	}
	//	if t.TokenUsage.TotalTokens > 0 || t.TokenUsage.InputTokens > 0 || t.TokenUsage.OutputTokens > 0 { // Only show if non-zero
	//		base += fmt.Sprintf("\n  Tokens(In: %d, Out: %d, Total: %d)", t.TokenUsage.InputTokens, t.TokenUsage.OutputTokens, t.TokenUsage.TotalTokens)
	//	}
	return base
}
