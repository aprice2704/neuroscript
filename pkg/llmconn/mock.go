// NeuroScript Version: 0.7.0
// File version: 6
// Purpose: Implemented the provider.AIProvider interface on MockConn to allow its use in interpreter tests.
// filename: pkg/llmconn/mock.go
// nlines: 125
// risk_rating: MEDIUM

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
// It also implements provider.AIProvider to be used directly in interpreter tests.
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

// Chat implements the provider.AIProvider interface, allowing the MockConn to be
// used where a provider is expected in tests. It parses the raw prompt string
// into an envelope and then delegates to the Converse method.
func (m *MockConn) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	// The mock provider needs to simulate what a real provider does: parse the envelope.
	// We find the last envelope to correctly skip over bootstrap text.
	promptToParse := req.Prompt
	if markerPos := strings.LastIndex(req.Prompt, aeiou.Wrap(aeiou.SectionStart)); markerPos != -1 {
		promptToParse = req.Prompt[markerPos:]
	}

	env, _, err := aeiou.Parse(strings.NewReader(promptToParse))
	if err != nil {
		// This simulates a provider failing because the prompt isn't a valid envelope.
		return nil, fmt.Errorf("mock provider could not parse AEIOU envelope from prompt: %w", err)
	}

	// Delegate to the existing scenario-based logic.
	return m.Converse(ctx, env)
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
