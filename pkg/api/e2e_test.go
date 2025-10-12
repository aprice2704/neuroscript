// NeuroScript Version: 0.8.0
// File version: 15
// Purpose: Updates interpreter creation to use the new HostContextBuilder, resolving a compile error.
// filename: pkg/api/e2e_test.go
// nlines: 132
// risk_rating: HIGH

package api_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"io"
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// TestCanonicalise_Determinism is a critical test to ensure that the byte output
// of the canonicalizer is stable across multiple runs on the same AST.
// Non-determinism is a primary cause of signature verification failures.
func TestCanonicalise_Determinism(t *testing.T) {
	// FIX: Corrected the map literal syntax by adding enclosing curly braces.
	src := `
func main(returns data) means
	set data = {\
		"c": 3,\
		"a": 1,\
		"b": 2\
	}
	return data
endfunc
`
	// 1. Parse the source code into a single, reusable AST.
	tree, err := api.Parse([]byte(src), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse failed: %v", err)
	}

	// 2. Canonicalize the AST multiple times.
	const numRuns = 5
	results := make([][]byte, numRuns)
	for i := 0; i < numRuns; i++ {
		blob, _, err := api.Canonicalise(tree)
		if err != nil {
			t.Fatalf("Run %d: api.Canonicalise failed: %v", i+1, err)
		}
		results[i] = blob
	}

	// 3. Compare all results to the first result. They must be identical.
	for i := 1; i < numRuns; i++ {
		if !bytes.Equal(results[0], results[i]) {
			t.Errorf("Canonicalization is not deterministic!")
			t.Logf("Run 1 Output: %x", results[0])
			t.Logf("Run %d Output: %x", i+1, results[i])
			t.Fatal("Byte blobs do not match, which will cause signature validation to fail.")
		}
	}
}

// TestEndToEnd_GoldenPath_SignatureVerification provides a full integration test of the
// public API's signing and loading workflow, which would have caught the previous
// signature verification bug.
func TestEndToEnd_GoldenPath_SignatureVerification(t *testing.T) {
	// 1. Define Source and Keys
	src := `func main() means
		emit "hello"
	endfunc`
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

	// 3. Canonicalise the AST to get the byte blob and its hash.
	blob, sum, err := api.Canonicalise(tree)
	if err != nil {
		t.Fatalf("Step 3: api.Canonicalise failed: %v", err)
	}
	if len(blob) == 0 {
		t.Fatal("Step 3: api.Canonicalise returned an empty blob.")
	}

	// 4. Sign the hash of the canonical blob.
	sig := ed25519.Sign(privKey, sum[:])
	signedAST := &api.SignedAST{Blob: blob, Sum: sum, Sig: sig}

	// 5. Load the signed AST. This is the critical step where verification happens.
	// This call would have failed before the bug was fixed.
	_, err = api.Load(context.Background(), signedAST, api.LoaderConfig{}, pubKey)
	if err != nil {
		t.Fatalf("Step 5: api.Load failed with a signature verification error: %v", err)
	}
}

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

	// 3. Canonicalise the AST.
	blob, sum, err := api.Canonicalise(tree)
	if err != nil {
		t.Fatalf("Step 3: api.Canonicalise failed: %v", err)
	}

	// 4. Sign the hash of the canonical blob.
	sig := ed25519.Sign(privKey, sum[:])
	signedAST := &api.SignedAST{Blob: blob, Sum: sum, Sig: sig}

	// 5. Load the signed AST, which verifies the signature.
	loadedUnit, err := api.Load(context.Background(), signedAST, api.LoaderConfig{}, pubKey)
	if err != nil {
		t.Fatalf("Step 5: api.Load failed: %v", err)
	}

	// 6. Execute using the stateful model.
	var stdout bytes.Buffer
	hc, err := api.NewHostContextBuilder().
		WithLogger(logging.NewNoOpLogger()).
		WithStdout(&stdout).
		WithStdin(os.Stdin).
		WithStderr(io.Discard).
		Build()
	if err != nil {
		t.Fatalf("Failed to build host context: %v", err)
	}
	interp := api.New(api.WithHostContext(hc))

	// 7. Load the program by executing the loaded unit.
	_, execErr := api.ExecWithInterpreter(context.Background(), interp, loadedUnit.Tree)
	if execErr != nil {
		t.Fatalf("Step 7: api.ExecWithInterpreter failed during load: %v", execErr)
	}

	// 8. Explicitly run the function by name on the same interpreter.
	result, runErr := interp.Run("only_command_blocks_run_automatically")
	if runErr != nil {
		t.Fatalf("Step 9: interp.Run failed: %v", runErr)
	}

	// 9. Verify the final execution results.
	expectedResult := lang.StringValue{Value: "hello world"}
	if result != expectedResult {
		t.Errorf("Step 10: Expected result value %#v, but got %#v", expectedResult, result)
	}
}
