// NeuroScript Version: 0.7.2
// File version: 2
// Purpose: Corrected test setup to initialize the interpreter with a valid HostContext, fixing a panic.
// filename: pkg/tool/metadata/tools_test.go
// nlines: 94
// risk_rating: LOW
package metadata_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
	toolmeta "github.com/aprice2704/neuroscript/pkg/tool/metadata"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func newMetadataTestInterpreter(t *testing.T) *interpreter.Interpreter {
	t.Helper()

	hostCtx := &interpreter.HostContext{
		Logger: logging.NewTestLogger(t),
		Stdout: &bytes.Buffer{},
		Stdin:  &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}

	testPolicy := policy.NewBuilder(policy.ContextNormal).Allow("tool.metadata.*").Build()
	interp := interpreter.NewInterpreter(
		interpreter.WithHostContext(hostCtx),
		interpreter.WithExecPolicy(testPolicy),
	)

	// for _, toolImpl := range toolmeta.MetadataToolsToRegister {
	// 	if _, err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
	// 		t.Fatalf("Failed to register tool '%s': %v", toolImpl.Spec.Name, err)
	// 	}
	// }
	return interp
}

func TestToolMetadata_Detect(t *testing.T) {
	interp := newMetadataTestInterpreter(t)
	fullname := types.MakeFullName(toolmeta.Group, "Detect")
	toolImpl, _ := interp.ToolRegistry().GetTool(fullname)

	mdContent := "content\n::serialization: md"
	result, err := toolImpl.Func(interp, []interface{}{mdContent})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "md" {
		t.Errorf("Expected result 'md', got %q", result)
	}
}

func TestToolMetadata_NormalizeKey(t *testing.T) {
	interp := newMetadataTestInterpreter(t)
	fullname := types.MakeFullName(toolmeta.Group, "NormalizeKey")
	toolImpl, _ := interp.ToolRegistry().GetTool(fullname)

	result, err := toolImpl.Func(interp, []interface{}{"File-Version_1.0"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "fileversion10" {
		t.Errorf("Expected result 'fileversion10', got %q", result)
	}
}

func TestToolMetadata_Parse(t *testing.T) {
	interp := newMetadataTestInterpreter(t)
	fullname := types.MakeFullName(toolmeta.Group, "Parse")
	toolImpl, _ := interp.ToolRegistry().GetTool(fullname)

	nsContent := "::serialization: ns\n::id: test-1\nbody content"
	result, err := toolImpl.Func(interp, []interface{}{nsContent})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected a map result, got %T", result)
	}

	expected := map[string]interface{}{
		"serialization": "ns",
		"body":          "body content",
		"metadata": map[string]interface{}{
			"serialization": "ns",
			"id":            "test-1",
		},
	}

	if !reflect.DeepEqual(resMap, expected) {
		t.Errorf("Result mismatch.\nGot:    %#v\nWanted: %#v", resMap, expected)
	}
}
