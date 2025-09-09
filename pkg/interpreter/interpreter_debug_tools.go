// NeuroScript Version: 0.7.1
// File version: 1
// Purpose: Implements a debug tool to dump the state of all registered interpreter clones.
// filename: pkg/interpreter/interpreter_debug_tools.go
// nlines: 45
// risk_rating: LOW

package interpreter

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/tool"
)

func registerDebugTools(r tool.ToolRegistry) error {
	impl := tool.ToolImplementation{
		Spec: tool.ToolSpec{
			Name:        "dumpClones",
			Group:       "debug",
			Description: "Logs the state of all registered interpreter clones to the host's stdout.",
			ReturnType:  tool.ArgTypeString,
		},
		Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) {
			interp, ok := rt.(*Interpreter)
			if !ok {
				return nil, fmt.Errorf("internal error: runtime is not an *Interpreter")
			}

			root := interp.rootInterpreter()
			root.cloneRegistryMu.Lock()
			defer root.cloneRegistryMu.Unlock()

			var report strings.Builder
			report.WriteString("\n\n--- Interpreter Clone Dump ---\n")
			report.WriteString(fmt.Sprintf("Root: %s (customEmitFunc: %t)\n", root.id, root.customEmitFunc != nil))
			report.WriteString(fmt.Sprintf("Total Clones Registered: %d\n", len(root.cloneRegistry)))

			for i, clone := range root.cloneRegistry {
				report.WriteString(fmt.Sprintf(
					"  [%d] ID: %s | Root: %s | customEmitFunc: %t\n",
					i,
					clone.id,
					clone.root.id,
					clone.customEmitFunc != nil,
				))
			}
			report.WriteString("----------------------------\n\n")

			// Use Println to ensure it goes to the interpreter's configured stdout,
			// which is captured by the test harness.
			interp.Println(report.String())

			return report.String(), nil
		},
	}
	_, err := r.RegisterTool(impl)
	return err
}
