// NeuroScript Version: 0.6.0
// File version: 10
// Purpose: Corrects a faulty type assertion in the test setup (*api.Program -> *ast.Program), which was causing the return value to be lost.
// filename: pkg/api/e2e_test.go
// nlines: 105
// risk_rating: MEDIUM

package api_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api" // Correctly import the ast package
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// TestEndToEndGoldenPath provides a full integration test of the public API,
// following the golden path outlined in the integration guide.
func TestEndToEndGoldenPath(t *testing.T) {
	// 1. Define Source and Keys
	src := `
func only_command_blocks_run_automatically(returns msg) means
  set msg = "hello world"
  emit msg
  return msg
endfunc
`
	srcBytes := []byte(src)

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ed25519 key pair: %v", err)
	}

	// 2. Parse the source code into an AST.
	tree, err := api.Parse(srcBytes, api.ParseSkipComments)
	if err != nil {
		t.Fatalf("Step 2: api.Parse failed: %v", err)
	}
	if tree == nil || tree.Root == nil {
		t.Fatal("Step 2: api.Parse returned a nil tree.")
	}
	t.Log("Step 2: Parse successful.")

	// 3. Canonicalise the AST.
	blob, sum, err := api.Canonicalise(tree)
	if err != nil {
		t.Fatalf("Step 3: api.Canonicalise failed: %v", err)
	}
	if len(blob) == 0 {
		t.Fatal("Step 3: api.Canonicalise returned an empty blob.")
	}
	t.Log("Step 3: Canonicalise successful.")

	// 4. Sign the hash of the canonical blob.
	sig := ed25519.Sign(privKey, sum[:])
	signedAST := &api.SignedAST{Blob: blob, Sum: sum, Sig: sig}
	t.Log("Step 4: Sign successful.")

	// 5. Load the signed AST, which verifies the signature.
	loadedUnit, err := api.Load(context.Background(), signedAST, api.LoaderConfig{}, pubKey)
	if err != nil {
		t.Fatalf("Step 5: api.Load failed: %v", err)
	}
	if loadedUnit == nil {
		t.Fatal("Step 5: api.Load returned a nil unit.")
	}
	t.Log("Step 5: Load successful.")

	// 6. Execute using the stateful model.
	var stdout bytes.Buffer
	interp := api.New(api.WithStdout(&stdout))

	// 7. Load the program by executing the loaded unit.
	_, execErr := api.ExecWithInterpreter(context.Background(), interp, loadedUnit.Tree)
	if execErr != nil {
		t.Fatalf("Step 7: api.ExecWithInterpreter failed during load: %v", execErr)
	}

	// 8. Verify that no execution happened automatically.
	if stdout.Len() > 0 {
		t.Fatalf("Step 8: Expected no output after loading, but got %q", stdout.String())
	}
	t.Log("Step 8: Confirmed loading was a passive operation.")

	// 9. Now, explicitly run the function by name on the same interpreter.
	result, runErr := interp.Run("only_command_blocks_run_automatically")
	if runErr != nil {
		t.Fatalf("Step 9: interp.Run failed: %v", runErr)
	}
	t.Log("Step 9: Explicit Run call successful.")

	// 10. Verify the final execution results.
	if got := stdout.String(); got != "hello world\n" {
		t.Errorf("Step 10: Expected output 'hello world\\n', but got %q", got)
	}

	expectedResult := lang.StringValue{Value: "hello world"}
	if result != expectedResult {
		t.Errorf("Step 10: Expected result value %#v, but got %#v", expectedResult, result)
	}
}
