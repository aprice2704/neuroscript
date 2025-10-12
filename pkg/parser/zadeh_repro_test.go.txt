// filename: pkg/parser/zadeh_repro_test.go
// NeuroScript Version: 0.6.2
// File version: 1
// Purpose: Faithfully reproduces the zadeh e2e test failure within the parser package to isolate the AST construction bug for for-each loops.
// nlines: 70
// risk_rating: HIGH

package parser

import (
	"strings"
	"testing"
)

// TestZadehE2EScriptRepro takes the exact script from the failing e2e test
// and runs it through the parser and AST builder. It then inspects every
// 'for' loop in the resulting AST.
//
// A failure here will prove that the AST builder is creating a malformed 'for'
// step node with a nil 'Collection' field under the specific conditions of
// this script, which is the root cause of the runtime error in the zadeh tool.
func TestZadehE2EScriptRepro(t *testing.T) {
	ingestionScript := `
:: target: main

func main() means
    emit "--- Starting Ingestion with Filters ---"
    set plan_id = tool.FDM.Plan.Create(".")

    set glob_patterns = ["**/node_modules", "**/*.jar"]
    set plan_id = tool.FDM.Plan.AddIgnorePatterns(plan_id, glob_patterns)

    set mime_types = ["image/png"]
    set plan_id = tool.FDM.Plan.AddExcludeMIMETypes(plan_id, mime_types)

    set plan_id = tool.FDM.Plan.AddOverlay(plan_id, leaf.OverlayFSO)

    emit "--- Executing Ingest Plan ---"
    set success = tool.FDM.Repo.Ingest(plan_id)
    if not success
        fail "Ingestion failed unexpectedly."
    endif

    emit "--- Verifying Ingestion Results ---"
    set dir_entries = tool.FDM.FS.ListDir(".", true)

    set ingested_paths = []
    for each entry in dir_entries
        if entry["fs:kind"] == "fs_file"
            set ingested_paths = tool.List.Append(ingested_paths, entry["fs:path"])
        endif
    endfor

    set allowed_paths = ["src/main.go", "assets/empty.txt", "assets/config.txt"]
    set denied_paths = ["node_modules/some-lib/index.js", "libs/my-library.jar", "assets/logo.png"]

    for each path in allowed_paths
        if not tool.List.Contains(ingested_paths, path)
            fail "Verification FAILED: Allowed path was not found: " + path
        endif
    endfor

    for each path in denied_paths
        if tool.List.Contains(ingested_paths, path)
            fail "Verification FAILED: Denied path was found: " + path
        endif
    endfor

    emit "--- Verification Successful ---"
endfunc
`
	prog := testParseAndBuild(t, ingestionScript)
	if prog == nil {
		t.Fatal("testParseAndBuild returned a nil program")
	}

	proc, ok := prog.Procedures["main"]
	if !ok {
		t.Fatal("Procedure 'main' not found in AST")
	}

	forLoopCount := 0
	for i, step := range proc.Steps {
		// We only care about 'for' loop steps.
		if step.Type != "for" {
			continue
		}
		forLoopCount++

		// THIS IS THE CRITICAL ASSERTION
		if step.Collection == nil {
			var collectionName string
			// Heuristic to find the likely name of the collection for the error message
			if len(step.Body) > 0 && len(step.Body[0].Comments) > 0 {
				collectionName = strings.Fields(step.Body[0].Comments[0].Text)[3]
			}
			t.Fatalf("FAILURE: The 'Collection' field for the 'for' loop at step %d (looping over '%s') is nil.", i, collectionName)
		}
	}

	if forLoopCount != 3 {
		t.Errorf("Expected to find 3 'for' loops in the script, but found %d", forLoopCount)
	}
}
