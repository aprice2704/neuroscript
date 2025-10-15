// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Corrected test setup to initialize the interpreter with a valid HostContext, fixing a panic.
// filename: pkg/tool/account/tools_account_test.go
// nlines: 139
// risk_rating: MEDIUM
package account_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
	toolaccount "github.com/aprice2704/neuroscript/pkg/tool/account"
	"github.com/aprice2704/neuroscript/pkg/types"
)

type accountTestCase struct {
	name          string
	toolName      types.ToolName
	args          []interface{}
	setupFunc     func(t *testing.T, interp *interpreter.Interpreter) error
	checkFunc     func(t *testing.T, interp tool.Runtime, result interface{}, err error)
	wantResult    interface{}
	wantToolErrIs error
}

func newAccountTestInterpreter(t *testing.T) *interpreter.Interpreter {
	t.Helper()

	hostCtx := &interpreter.HostContext{
		Logger: logging.NewTestLogger(t),
		Stdout: &bytes.Buffer{},
		Stdin:  &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}

	testPolicy := policy.AllowAll()
	testPolicy.Context = policy.ContextConfig
	testPolicy.Grants = capability.NewGrantSet(
		[]capability.Capability{
			{Resource: "account", Verbs: []string{"admin", "read"}, Scopes: []string{"*"}},
		},
		capability.Limits{},
	)

	interp := interpreter.NewInterpreter(
		interpreter.WithHostContext(hostCtx),
		interpreter.WithExecPolicy(testPolicy),
	)

	for _, toolImpl := range toolaccount.AccountToolsToRegister {
		if _, err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
			t.Fatalf("Failed to register tool '%s': %v", toolImpl.Spec.Name, err)
		}
	}
	return interp
}

func testAccountToolHelper(t *testing.T, tc accountTestCase) {
	t.Helper()

	interp := newAccountTestInterpreter(t)

	if tc.setupFunc != nil {
		if err := tc.setupFunc(t, interp); err != nil {
			t.Fatalf("Setup function failed for test '%s': %v", tc.name, err)
		}
	}

	fullname := types.MakeFullName(toolaccount.Group, string(tc.toolName))
	toolImpl, found := interp.ToolRegistry().GetTool(fullname)
	if !found {
		t.Fatalf("Tool %q not found in registry", tc.toolName)
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
		t.Fatalf("Unexpected error: %v", err)
	}

	if err == nil {
		if !reflect.DeepEqual(result, tc.wantResult) {
			t.Errorf("Result mismatch.\nGot:    %#v\nWanted: %#v", result, tc.wantResult)
		}
	}
}

func newValidAccountConfig() map[string]interface{} {
	return map[string]interface{}{
		"kind":     "llm",
		"provider": "test-provider",
		"api_key":  "test-key",
	}
}

func TestToolAccount_Register(t *testing.T) {
	tests := []accountTestCase{
		{
			name:       "Success: Register a new account",
			toolName:   "Register",
			args:       []interface{}{"test_account_1", newValidAccountConfig()},
			wantResult: true,
		},
		{
			name:          "Fail: Missing required kind field",
			toolName:      "Register",
			args:          []interface{}{"test_account_2", map[string]interface{}{"provider": "p", "api_key": "k"}},
			wantToolErrIs: account.ErrInvalidConfiguration,
		},
		{
			name:     "Fail: Duplicate registration",
			toolName: "Register",
			args:     []interface{}{"test_account_1", newValidAccountConfig()},
			setupFunc: func(t *testing.T, interp *interpreter.Interpreter) error {
				return interp.AccountsAdmin().Register("test_account_1", newValidAccountConfig())
			},
			wantToolErrIs: lang.ErrDuplicateKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAccountToolHelper(t, tt)
		})
	}
}

func TestToolAccount_Delete(t *testing.T) {
	setup := func(t *testing.T, interp *interpreter.Interpreter) error {
		return interp.AccountsAdmin().Register("acct_to_delete", newValidAccountConfig())
	}

	tests := []accountTestCase{
		{
			name:       "Success: Delete existing account",
			toolName:   "Delete",
			args:       []interface{}{"acct_to_delete"},
			setupFunc:  setup,
			wantResult: true,
		},
		{
			name:       "Success: Delete non-existent account",
			toolName:   "Delete",
			args:       []interface{}{"nonexistent_acct"},
			setupFunc:  setup,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAccountToolHelper(t, tt)
		})
	}
}

func TestToolAccount_List(t *testing.T) {
	setup := func(t *testing.T, interp *interpreter.Interpreter) error {
		if err := interp.AccountsAdmin().Register("b-acct", newValidAccountConfig()); err != nil {
			return err
		}
		if err := interp.AccountsAdmin().Register("a-acct", newValidAccountConfig()); err != nil {
			return err
		}
		return nil
	}

	tests := []accountTestCase{
		{
			name:      "Success: List accounts",
			toolName:  "List",
			args:      []interface{}{},
			setupFunc: setup,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				names, ok := result.([]string)
				if !ok {
					t.Fatalf("Expected []string, got %T", result)
				}
				// The order is not guaranteed, so check for presence.
				if len(names) != 2 || !((names[0] == "a-acct" && names[1] == "b-acct") || (names[0] == "b-acct" && names[1] == "a-acct")) {
					t.Errorf("Expected [a-acct, b-acct] (in any order), got %v", names)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAccountToolHelper(t, tt)
		})
	}
}
