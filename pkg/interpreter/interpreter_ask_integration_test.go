// NeuroScript Version: 0.7.0
// File version: 28
// Purpose: Corrected the mock provider to return syntactically valid AEIOU envelopes with single, correct sections, fixing the parsing errors.
// filename: pkg/interpreter/interpreter_ask_integration_test.go
// nlines: 180
// risk_rating: MEDIUM

package interpreter_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/provider"
)

const askTestScript = `
func TestBasicSuccess(returns result) means
    ask "default_agent", "What is the capital of BC?" into result
    return result
endfunc

func TestProviderError() means
    ask "default_agent", "This will cause a provider error."
endfunc

func TestWithOptions() means
    ask "default_agent", "A prompt." with {"temperature": 0.85}
endfunc

func TestNonExistentAgent() means
    ask "no_such_agent", "This will fail because the agent is not registered."
endfunc
`

type mockAskProvider struct {
	LastRequest      provider.AIRequest
	ResponseToReturn *provider.AIResponse
	ErrorToReturn    error
}

func (m *mockAskProvider) Chat(ctx context.Context, req provider.AIRequest) (*provider.AIResponse, error) {
	m.LastRequest = req
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	if m.ResponseToReturn == nil {
		// FIX: The default response is now a valid, minimal AEIOU envelope with a single ACTIONS section.
		return &provider.AIResponse{TextContent: strings.Join([]string{
			`<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:START>>>`,
			`<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:ACTIONS>>>`,
			`command`,
			`  emit "default mock response"`,
			`  emit '<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:LOOP:{"control":"done"}>>>'`,
			`endcommand`,
			`<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:END>>>`,
		}, "\n")}, nil
	}
	return m.ResponseToReturn, nil
}

func setupAskTest(t *testing.T) (*interpreter.Interpreter, *mockAskProvider) {
	t.Helper()

	logger := logging.NewTestLogger(t)
	permissivePolicy := &policy.ExecPolicy{
		Context: policy.ContextConfig,
		Allow:   []string{"*"},
		Grants: capability.NewGrantSet(
			[]capability.Capability{
				{Resource: "model", Verbs: []string{"admin", "use"}, Scopes: []string{"*"}},
				{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"*"}},
				{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"*"}},
			},
			capability.Limits{},
		),
	}
	interp := interpreter.NewInterpreter(
		interpreter.WithLogger(logger),
		interpreter.WithExecPolicy(permissivePolicy),
	)

	mockProv := &mockAskProvider{}
	interp.RegisterProvider("mock_provider", mockProv)

	agentConfig := map[string]any{
		"provider": "mock_provider",
		"model":    "mock-model-v1",
		"tools":    map[string]any{"toolLoopPermitted": true},
	}
	if err := interp.AgentModelsAdmin().Register("default_agent", agentConfig); err != nil {
		t.Fatalf("Failed to register default agent model: %v", err)
	}

	parserAPI := parser.NewParserAPI(interp.GetLogger())
	p, pErr := parserAPI.Parse(askTestScript)
	if pErr != nil {
		t.Fatalf("Failed to parse embedded script: %v", pErr)
	}

	program, _, bErr := parser.NewASTBuilder(interp.GetLogger()).Build(p)
	if bErr != nil {
		t.Fatalf("Failed to build AST: %v", bErr)
	}
	if err := interp.Load(program); err != nil {
		t.Fatalf("Failed to load program into interpreter: %v", err)
	}
	return interp, mockProv
}

func TestAskIntegration(t *testing.T) {
	t.Run("Basic ask statement success", func(t *testing.T) {
		interp, mockProv := setupAskTest(t)
		// FIX: Corrected the mock response to be a valid envelope with a single ACTIONS section.
		mockProv.ResponseToReturn = &provider.AIResponse{TextContent: strings.Join([]string{
			`<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:START>>>`,
			`<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:ACTIONS>>>`,
			`command`,
			`  emit "Victoria"`,
			`  emit '<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:LOOP:{"control":"done"}>>>'`,
			`endcommand`,
			`<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:END>>>`,
		}, "\n")}

		finalResult, err := interp.Run("TestBasicSuccess")
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		resultStr, _ := lang.ToString(finalResult)
		if resultStr != "Victoria" {
			t.Errorf("Expected result 'Victoria', got '%s'", resultStr)
		}
	})

	t.Run("Ask statement with provider error", func(t *testing.T) {
		interp, mockProv := setupAskTest(t)
		mockProv.ErrorToReturn = errors.New("provider API key invalid")

		_, err := interp.Run("TestProviderError")
		if err == nil {
			t.Fatal("Script execution was expected to fail, but it succeeded.")
		}

		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || !strings.Contains(rtErr.Error(), "provider API key invalid") {
			t.Errorf("Expected a provider error, but got: %v", err)
		}
	})

	t.Run("Ask statement with 'with' options", func(t *testing.T) {
		interp, _ := setupAskTest(t)
		_, err := interp.Run("TestWithOptions")
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}
	})

	t.Run("Ask with non-existent AgentModel", func(t *testing.T) {
		interp, _ := setupAskTest(t)
		_, err := interp.Run("TestNonExistentAgent")
		if err == nil {
			t.Fatal("Script execution was expected to fail, but it succeeded.")
		}

		var rtErr *lang.RuntimeError
		if !errors.As(err, &rtErr) || rtErr.Code != lang.ErrorCodeKeyNotFound {
			t.Errorf("Expected a KeyNotFound error, but got: %v", err)
		}
	})
}
