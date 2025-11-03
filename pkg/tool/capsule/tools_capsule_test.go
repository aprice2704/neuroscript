// NeuroScript Version: 0.8.0
// File version: 16
// Purpose: Fixes test failure by correctly asserting unwrapped slice as []any, not []string.
// filename: pkg/tool/capsule/tools_test.go
// nlines: 242
// risk_rating: MEDIUM
package capsule_test

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
	toolcapsule "github.com/aprice2704/neuroscript/pkg/tool/capsule"
	"github.com/aprice2704/neuroscript/pkg/types"
)

type capsuleTestCase struct {
	name          string
	toolName      types.ToolName
	args          []interface{}
	setupFunc     func(t *testing.T, interp *interpreter.Interpreter) error
	checkFunc     func(t *testing.T, interp tool.Runtime, result interface{}, err error)
	wantResult    interface{}
	wantToolErrIs error
	isPrivileged  bool
	// new field for provider tests
	provider interfaces.CapsuleProvider
}

func newCapsuleTestInterpreter(t *testing.T, isPrivileged bool) *interpreter.Interpreter {
	t.Helper()

	hostCtx := &interpreter.HostContext{
		Logger: logging.NewTestLogger(t),
		Stdout: &bytes.Buffer{},
		Stdin:  &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}

	var testPolicy *policy.ExecPolicy
	var opts []interpreter.InterpreterOption

	if isPrivileged {
		testPolicy = policy.NewBuilder(policy.ContextConfig).
			Allow("tool.capsule.*").
			Grant("capsule:write:*").
			Build()
		adminRegistry := capsule.NewRegistry()
		opts = append(opts, interpreter.WithCapsuleAdminRegistry(adminRegistry))
	} else {
		testPolicy = policy.NewBuilder(policy.ContextNormal).
			Allow(
				"tool.capsule.List",
				"tool.capsule.Read",
				"tool.capsule.GetLatest",
			).Build()
	}
	opts = append(opts, interpreter.WithExecPolicy(testPolicy))
	opts = append(opts, interpreter.WithHostContext(hostCtx))

	interp := interpreter.NewInterpreter(opts...)

	return interp
}

func testCapsuleToolHelper(t *testing.T, tc capsuleTestCase) {
	t.Helper()

	interp := newCapsuleTestInterpreter(t, tc.isPrivileged)

	// Inject the provider if one is specified for this test case
	if tc.provider != nil {
		interp.SetCapsuleProvider(tc.provider)
	}

	if tc.setupFunc != nil {
		if err := tc.setupFunc(t, interp); err != nil {
			t.Fatalf("Setup function failed for test '%s': %v", tc.name, err)
		}
	}

	fullname := types.MakeFullName(toolcapsule.Group, string(tc.toolName))
	toolImpl, found := interp.ToolRegistry().GetTool(fullname)

	if !found {
		t.Fatalf("Tool %q not found in registry", fullname)
	}

	result, err := toolImpl.Func(interp, tc.args)

	if tc.checkFunc != nil {
		tc.checkFunc(t, interp, result, err)
		return
	}

	if tc.wantToolErrIs != nil {
		if !errors.Is(err, tc.wantToolErrIs) {
			t.Errorf("Expected error wrapping [%v], but got: %v", tc.wantToolErrIs, err)
		}
	} else if err != nil {
		t.Fatalf("Unexpected error for test '%s': %v", tc.name, err)
	}

	if err == nil {
		if tc.wantResult != nil && !reflect.DeepEqual(result, tc.wantResult) {
			t.Errorf("Result mismatch.\nGot:    %#v\nWanted: %#v", result, tc.wantResult)
		}
	}
}

// --- Mock Provider ---

type mockCapsuleProvider struct {
	// Control what the mock returns
	listResult any
	getResult  any
	readResult any
	addResult  any
	returnErr  error

	// Spy on what the mock was called with
	calledList      bool
	calledGetLatest bool
	calledRead      bool
	calledAdd       bool

	lastGetName string
	lastReadID  string
	lastAddData string
}

func (m *mockCapsuleProvider) Add(ctx context.Context, capsuleContent string) (any, error) {
	m.calledAdd = true
	m.lastAddData = capsuleContent
	return m.addResult, m.returnErr
}

func (m *mockCapsuleProvider) GetLatest(ctx context.Context, name string) (any, error) {
	m.calledGetLatest = true
	m.lastGetName = name
	return m.getResult, m.returnErr
}

func (m *mockCapsuleProvider) List(ctx context.Context) (any, error) {
	m.calledList = true
	return m.listResult, m.returnErr
}

func (m *mockCapsuleProvider) Read(ctx context.Context, id string) (any, error) {
	m.calledRead = true
	m.lastReadID = id
	return m.readResult, m.returnErr
}

// --- New Tests for Injected Provider ---

func TestToolCapsule_WithProvider(t *testing.T) {
	t.Run("GetLatest uses provider", func(t *testing.T) {
		mock := &mockCapsuleProvider{
			getResult: map[string]any{"id": "from-mock-provider", "version": "99"},
		}
		testCase := capsuleTestCase{
			name:     "GetLatest uses provider",
			toolName: "GetLatest",
			args:     []interface{}{"capsule/test"},
			provider: mock,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if !mock.calledGetLatest {
					t.Error("Provider.GetLatest was not called")
				}
				if mock.lastGetName != "capsule/test" {
					t.Errorf("Provider.GetLatest called with wrong name: got %q, want %q", mock.lastGetName, "capsule/test")
				}
				unwrapped := lang.Unwrap(result.(lang.Value))
				resMap, _ := unwrapped.(map[string]any)
				if resMap["id"] != "from-mock-provider" {
					t.Errorf("Unexpected result: got %v, want %v", resMap["id"], "from-mock-provider")
				}
			},
		}
		testCapsuleToolHelper(t, testCase)
	})

	t.Run("List uses provider", func(t *testing.T) {
		mock := &mockCapsuleProvider{
			listResult: []string{"capsule/from-mock@1"},
		}
		testCase := capsuleTestCase{
			name:     "List uses provider",
			toolName: "List",
			args:     []interface{}{},
			provider: mock,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if !mock.calledList {
					t.Error("Provider.List was not called")
				}
				unwrapped := lang.Unwrap(result.(lang.Value))

				// FIX: lang.Unwrap(SliceValue) returns []any, not []string
				resList, ok := unwrapped.([]any)
				if !ok {
					t.Fatalf("Expected unwrapped result to be []any, but got %T", unwrapped)
				}
				if len(resList) != 1 {
					t.Fatalf("Expected slice of length 1, got %d", len(resList))
				}
				elem, ok := resList[0].(string)
				if !ok {
					t.Fatalf("Expected slice element to be string, got %T", resList[0])
				}
				if elem != "capsule/from-mock@1" {
					t.Errorf("Unexpected result: got %v", resList)
				}
			},
		}
		testCapsuleToolHelper(t, testCase)
	})

	t.Run("Read uses provider", func(t *testing.T) {
		mock := &mockCapsuleProvider{
			readResult: map[string]any{"id": "from-mock-read", "content": "mocked"},
		}
		testCase := capsuleTestCase{
			name:     "Read uses provider",
			toolName: "Read",
			args:     []interface{}{"capsule/test@1"},
			provider: mock,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if !mock.calledRead {
					t.Error("Provider.Read was not called")
				}
				if mock.lastReadID != "capsule/test@1" {
					t.Errorf("Provider.Read called with wrong id: got %q, want %q", mock.lastReadID, "capsule/test@1")
				}
				unwrapped := lang.Unwrap(result.(lang.Value))
				resMap, _ := unwrapped.(map[string]any)
				if resMap["id"] != "from-mock-read" {
					t.Errorf("Unexpected result: got %v", resMap["id"])
				}
			},
		}
		testCapsuleToolHelper(t, testCase)
	})

	t.Run("Add uses provider", func(t *testing.T) {
		mock := &mockCapsuleProvider{
			addResult: map[string]any{"id": "capsule/added-by-mock", "version": "1"},
		}
		testCase := capsuleTestCase{
			name:         "Add uses provider",
			toolName:     "Add",
			args:         []interface{}{"content-for-mock"},
			provider:     mock,
			isPrivileged: true, // Still need privilege to even access the tool
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if !mock.calledAdd {
					t.Error("Provider.Add was not called")
				}
				if mock.lastAddData != "content-for-mock" {
					t.Errorf("Provider.Add called with wrong content: got %q", mock.lastAddData)
				}
				unwrapped := lang.Unwrap(result.(lang.Value))
				resMap, _ := unwrapped.(map[string]any)
				if resMap["id"] != "capsule/added-by-mock" {
					t.Errorf("Unexpected result: got %v", resMap["id"])
				}
			},
		}
		testCapsuleToolHelper(t, testCase)
	})
}
