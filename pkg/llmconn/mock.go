// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Updated scenario helpers to generate V3-compliant ACTION blocks using 'tool.aeiou.magic' instead of the obsolete V2 'aeiou.Wrap' function. Adds a UserData section to conform to V3 spec.
// filename: pkg/llmconn/mock.go
// nlines: 106
// risk_rating: LOW

package llmconn

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

// ScenarioTurn defines a single response from the mock LLM, which can be either a success or an error.
type ScenarioTurn struct {
	Response *provider.AIResponse
	Err      error
}

// MockConn is a mock implementation of askloop.Connector for testing.
type MockConn struct {
	t        *testing.T
	mu       sync.Mutex
	scenario []ScenarioTurn
	turn     int
}

// NewMock creates a new MockConn with a defined scenario.
func NewMock(t *testing.T, scenario ...ScenarioTurn) *MockConn {
	return &MockConn{
		t:        t,
		scenario: scenario,
	}
}

// Converse pops the next response from the scenario queue.
func (m *MockConn) Converse(ctx context.Context, input *aeiou.Envelope) (*provider.AIResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.turn >= len(m.scenario) {
		m.t.Logf("MockConn: Converse called more times than there are turns in the scenario (called %d times for a %d-turn scenario)", m.turn+1, len(m.scenario))
		return nil, fmt.Errorf("mockconn: scenario ended")
	}

	next := m.scenario[m.turn]
	m.turn++

	if next.Err != nil {
		return nil, next.Err
	}
	return next.Response, nil
}

// --- Scenario Helper Functions ---

// composeAction creates a valid AEIOU envelope with a given action string.
func composeAction(action string) (string, error) {
	env := &aeiou.Envelope{
		UserData: "{}", // V3 requires a UserData section.
		Actions:  action,
	}
	return env.Compose()
}

// Continue creates a scenario turn that emits a message and signals the loop to continue.
func Continue(message string) ScenarioTurn {
	sanitized := strings.ReplaceAll(message, "\"", "\\\"")
	action := fmt.Sprintf(`
	command
		emit "%s"
		set p = {"action": "continue", "notes": "Continuing mock scenario."}
		emit tool.aeiou.magic("LOOP", p)
	endcommand`, sanitized)

	envelope, err := composeAction(action)
	if err != nil {
		panic(fmt.Sprintf("failed to build 'Continue' scenario turn: %v", err))
	}
	return ScenarioTurn{Response: &provider.AIResponse{TextContent: envelope}}
}

// Done creates a final scenario turn that emits a message and signals the loop is done.
func Done(message string) ScenarioTurn {
	sanitized := strings.ReplaceAll(message, "\"", "\\\"")
	action := fmt.Sprintf(`
	command
		emit "%s"
		set p = {"action": "done", "notes": "Completing mock scenario."}
		emit tool.aeiou.magic("LOOP", p)
	endcommand`, sanitized)

	envelope, err := composeAction(action)
	if err != nil {
		panic(fmt.Sprintf("failed to build 'Done' scenario turn: %v", err))
	}
	return ScenarioTurn{Response: &provider.AIResponse{TextContent: envelope}}
}

// Error creates a scenario turn that returns an error, simulating a provider failure.
func Error(err error) ScenarioTurn {
	return ScenarioTurn{Err: err}
}
