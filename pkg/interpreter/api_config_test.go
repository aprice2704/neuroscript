// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: Added test to verify the creation of a default execution policy.
// filename: pkg/interpreter/api_config_test.go
// nlines: 182
// risk_rating: LOW

package interpreter_test

import (
	"bytes"
	"sync"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestInterpreter_ConfigurationOptions(t *testing.T) {

	t.Run("Optional HostContext callbacks are correctly wired", func(t *testing.T) {
		t.Logf("[DEBUG] Starting test: Optional HostContext callbacks are correctly wired")
		var (
			emitCalled    bool
			whisperCalled bool
			mu            sync.Mutex
		)

		harness := NewTestHarness(t)
		hc, err := interpreter.NewHostContextBuilder().
			WithLogger(logging.NewTestLogger(t)).
			WithStdout(&bytes.Buffer{}).
			WithStdin(&bytes.Buffer{}).
			WithStderr(&bytes.Buffer{}).
			WithEmitFunc(func(v lang.Value) {
				mu.Lock()
				defer mu.Unlock()
				emitCalled = true
			}).
			WithWhisperFunc(func(h, d lang.Value) {
				mu.Lock()
				defer mu.Unlock()
				whisperCalled = true
			}).
			Build()
		if err != nil {
			t.Fatalf("HostContextBuilder failed: %v", err)
		}

		interp := interpreter.NewInterpreter(interpreter.WithHostContext(hc))
		script := `
			command
				emit "hello"
				whisper "self", "data"
			endcommand
		`
		tree, _ := harness.Parser.Parse(script)
		program, _, _ := harness.ASTBuilder.Build(tree)

		_, execErr := interp.Execute(program)
		if execErr != nil {
			t.Fatalf("Script execution failed: %v", execErr)
		}

		mu.Lock()
		defer mu.Unlock()
		if !emitCalled {
			t.Error("EmitFunc set via builder was not called.")
		}
		if !whisperCalled {
			t.Error("WhisperFunc set via builder was not called.")
		}
		t.Logf("[DEBUG] Test passed.")
	})

	t.Run("Defaults are used when HostContext callbacks are nil", func(t *testing.T) {
		t.Logf("[DEBUG] Starting test: Defaults are used when HostContext callbacks are nil")

		harness := NewTestHarness(t)
		var stdoutBuffer bytes.Buffer

		hc, err := interpreter.NewHostContextBuilder().
			WithLogger(logging.NewTestLogger(t)).
			WithStdout(&stdoutBuffer).
			WithStdin(&bytes.Buffer{}).
			WithStderr(&bytes.Buffer{}).
			Build()
		if err != nil {
			t.Fatalf("HostContextBuilder failed: %v", err)
		}

		interp := interpreter.NewInterpreter(interpreter.WithHostContext(hc))

		script := `
			command
				emit "hello default"
				whisper self, "whisper default"
			endcommand
		`
		tree, _ := harness.Parser.Parse(script)
		program, _, _ := harness.ASTBuilder.Build(tree)

		_, execErr := interp.Execute(program)
		if execErr != nil {
			t.Fatalf("Script execution failed: %v", execErr)
		}

		expectedEmit := "hello default\n"
		if got := stdoutBuffer.String(); got != expectedEmit {
			t.Errorf("Default emit behavior failed. Expected stdout to be '%s', got '%s'", expectedEmit, got)
		}

		expectedWhisper := "whisper default\n"
		if got := interp.GetAndClearWhisperBuffer(); got != expectedWhisper {
			t.Errorf("Default whisper behavior failed. Expected buffer to contain '%s', got '%s'", expectedWhisper, got)
		}

		t.Logf("[DEBUG] Test passed.")
	})

	t.Run("Default ExecPolicy is created if not provided", func(t *testing.T) {
		t.Logf("[DEBUG] Starting test: Default ExecPolicy is created if not provided")

		// Don't use the harness as it provides a policy. Build from scratch.
		hc, err := interpreter.NewHostContextBuilder().
			WithLogger(logging.NewTestLogger(t)).
			WithStdout(&bytes.Buffer{}).
			WithStdin(&bytes.Buffer{}).
			WithStderr(&bytes.Buffer{}).
			Build()
		if err != nil {
			t.Fatalf("HostContextBuilder failed: %v", err)
		}

		// Create an interpreter without the WithExecPolicy option.
		interp := interpreter.NewInterpreter(interpreter.WithHostContext(hc))

		if interp.GetExecPolicy() == nil {
			t.Fatal("Interpreter's ExecPolicy is nil, but a default was expected.")
		}

		if interp.GetExecPolicy().Context != policy.ContextNormal {
			t.Errorf("Expected default policy context to be ContextNormal, but got %v", interp.GetExecPolicy().Context)
		}
		t.Logf("[DEBUG] Test passed.")
	})

	t.Run("WithAccountStore and WithAgentModelStore options work", func(t *testing.T) {
		t.Logf("[DEBUG] Starting test: WithAccountStore and WithAgentModelStore options work")
		h := NewTestHarness(t)

		accountStore := account.NewStore()
		modelStore := agentmodel.NewAgentModelStore()

		privilegedPolicy := policy.NewBuilder(policy.ContextConfig).
			Allow("*").
			Grant("model:admin:*").
			Grant("account:admin:*").
			Build()

		interp := interpreter.NewInterpreter(
			interpreter.WithHostContext(h.HostContext),
			interpreter.WithExecPolicy(privilegedPolicy),
			interpreter.WithAccountStore(accountStore),
			interpreter.WithAgentModelStore(modelStore),
		)

		if err := interp.AgentModelsAdmin().Register("test_agent", map[string]any{"provider": "p", "model": "m"}); err != nil {
			t.Fatalf("Failed to register agent model: %v", err)
		}
		accountConfig := map[string]any{
			"kind":     "test",
			"provider": "test_provider",
			"api_key":  "12345",
		}
		if err := interp.AccountsAdmin().Register("test_account", accountConfig); err != nil {
			t.Fatalf("Failed to register account: %v", err)
		}

		_, accExists := interp.Accounts().Get("test_account")
		if !accExists {
			t.Error("WithAccountStore failed: account registered via interpreter not found.")
		}

		_, agentExists := interp.AgentModels().Get(types.AgentModelName("test_agent"))
		if !agentExists {
			t.Error("WithAgentModelStore failed: agent model registered via interpreter not found.")
		}
		t.Logf("[DEBUG] Test passed.")
	})
}
