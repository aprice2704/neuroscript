// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 19
// :: description: Tests for the capsule toolset, including registry operations and the new metadata Parse tool.
// :: latestChange: Fixed Parse tests by removing leading newlines in content and using ErrInvalidCapsuleData.
// :: filename: pkg/tool/capsule/tools_test.go
// :: serialization: go

package capsule_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
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
		writableStore := capsule.NewStore(capsule.NewRegistry(), capsule.BuiltInRegistry())
		opts = append(opts, interpreter.WithCapsuleStore(writableStore))
	} else {
		testPolicy = policy.NewBuilder(policy.ContextNormal).
			Allow(
				"tool.capsule.List",
				"tool.capsule.Read",
				"tool.capsule.GetLatest",
				"tool.capsule.Parse",
			).Build()
		opts = append(opts, interpreter.WithCapsuleStore(capsule.DefaultStore()))
	}
	opts = append(opts, interpreter.WithExecPolicy(testPolicy))
	opts = append(opts, interpreter.WithHostContext(hostCtx))

	interp := interpreter.NewInterpreter(opts...)

	return interp
}

func testCapsuleToolHelper(t *testing.T, tc capsuleTestCase) {
	t.Helper()

	interp := newCapsuleTestInterpreter(t, tc.isPrivileged)

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

func TestCapsuleTools_Parse(t *testing.T) {
	t.Run("Parse Markdown Content", func(t *testing.T) {
		testCapsuleToolHelper(t, capsuleTestCase{
			name:     "Valid MD with metadata at end",
			toolName: "Parse",
			args: []interface{}{`# MD Test
Body content.

---
::id: capsule/test-md
::version: 1
::description: Test MD
::serialization: md`},
			wantResult: map[string]interface{}{
				"handle":      "capsule/test-md",
				"version":     "1",
				"description": "Test MD",
				"mime":        "text/markdown",
			},
		})
	})

	t.Run("Parse NeuroScript Content", func(t *testing.T) {
		testCapsuleToolHelper(t, capsuleTestCase{
			name:     "Valid NS with metadata at start",
			toolName: "Parse",
			args: []interface{}{`::id: capsule/test-ns
::version: 5
::description: Test NS
::serialization: ns

emit "test"
`},
			wantResult: map[string]interface{}{
				"handle":      "capsule/test-ns",
				"version":     "5",
				"description": "Test NS",
				"mime":        "application/x-neuroscript",
			},
		})
	})

	t.Run("Fails on Missing Required Fields", func(t *testing.T) {
		testCapsuleToolHelper(t, capsuleTestCase{
			name:     "Missing version",
			toolName: "Parse",
			args: []interface{}{`---
::id: capsule/fail
::description: Missing version
::serialization: md`},
			wantToolErrIs: toolcapsule.ErrInvalidCapsuleData,
		})
	})
}
