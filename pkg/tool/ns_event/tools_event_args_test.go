// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Contains tests for ns_event tools with a focus on argument validation.
// filename: pkg/tool/ns_event/tools_event_args_test.go
// nlines: 42
// risk_rating: LOW
package ns_event_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestToolEvent_Compose_OptionalArgFailures(t *testing.T) {
	tests := []eventTestCase{
		{
			name:          "Fail: id is not a string",
			toolName:      "Compose",
			args:          []interface{}{"kind.string", map[string]interface{}{}, 12345},
			wantToolErrIs: lang.ErrInvalidArgument,
		},
		{
			name:          "Fail: agent_id is not a string",
			toolName:      "Compose",
			args:          []interface{}{"kind.string", map[string]interface{}{}, "valid-id", 999},
			wantToolErrIs: lang.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testEventToolHelper(t, tt)
		})
	}
}
