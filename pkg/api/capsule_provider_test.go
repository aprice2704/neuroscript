// NeuroScript Version: 0.8.0
// File version: 6
// Purpose: Fixes test failure by asserting the expected 'ErrorCodeType' error occurs.
// Latest change: Commented out broken test for the deleted CapsuleProvider mechanism.
// filename: pkg/api/capsule_provider_test.go
// nlines: 97
// risk_rating: LOW

package api_test

// "github.com/aprice2704/neuroscript/pkg/logging" // No longer needed

/*
// mockAPIProvider implements the api.CapsuleProvider interface for testing.
type mockAPIProvider struct {
	calledGetLatest bool
	lastGetName     string
	getResult       any
	returnErr       error
}

func (m *mockAPIProvider) Add(ctx context.Context, capsuleContent string) (any, error) {
	return nil, m.returnErr
}
func (m *mockAPIProvider) GetLatest(ctx context.Context, name string) (any, error) {
	m.calledGetLatest = true
	m.lastGetName = name
	return m.getResult, m.returnErr
}
func (m *mockAPIProvider) List(ctx context.Context) (any, error) {
	return nil, m.returnErr
}
func (m *mockAPIProvider) Read(ctx context.Context, id string) (any, error) {
	return nil, m.returnErr
}
*/

// NOTE: newTestHostContext(t) is defined in another test file in this package (harness_test.go or capsule_admin_test.go)
// We are removing the duplicate definition from this file.

/*
func TestAPI_WithCapsuleProvider(t *testing.T) {
	// 1. Define the mock provider and what it should return
	mock := &mockAPIProvider{
		getResult: map[string]any{
			"id":      "capsule/from-provider@1",
			"name":    "capsule/from-provider",
			"version": "1",
			"content": "This content came from the mock provider",
		},
	}

	// 2. Define the script that will call the tool
	script := `
func main(returns string) means
    set c = tool.capsule.GetLatest("capsule/from-provider")
    return c["content"]
endfunc
`

	// 3. Create a standard interpreter, injecting the mock provider
	policy := api.NewPolicyBuilder(api.ContextNormal).
		Allow("tool.capsule.getlatest").
		Build()

	interp := api.New(
		api.WithHostContext(newTestHostContext(nil)),
		api.WithCapsuleStore(mock), // <-- THIS IS THE COMPILE ERROR
		api.WithExecPolicy(policy),
	)

	// 4. Parse and load the script
	tree, err := api.Parse([]byte(script), api.ParseSkipComments)
	if err != nil {
		t.Fatalf("api.Parse() failed: %v", err)
	}
	if err := interp.Load(tree); err != nil {
		t.Fatalf("interp.Load() failed: %v", err)
	}

	// 5. Run the 'main' procedure
	result, err := api.RunProcedure(context.Background(), interp, "main")
	if err != nil {
		t.Fatalf("api.RunProcedure() failed: %v", err)
	}

	// 6. Verify the mock was called correctly
	if !mock.calledGetLatest {
		t.Error("The mock CapsuleProvider.GetLatest method was not called")
	}
	if mock.lastGetName != "capsule/from-provider" {
		t.Errorf("Mock provider was called with wrong name: got %q, want %q",
			mock.lastGetName, "capsule/from-provider")
	}

	// 7. Verify the script got the mock's data
	unwrapped, _ := api.Unwrap(result)
	content, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected a string result, but got %T", unwrapped)
	}

	expectedContent := "This content came from the mock provider"
	if !strings.Contains(content, expectedContent) {
		t.Errorf("Read incorrect capsule content.\n  Expected to contain: %q\n  Got: %q",
			expectedContent, content)
	}

	// 8. Verify fallback is NOT used
	// Run again with a name the provider won't find (it will return nil)
	// This proves the tool doesn't fall back to the internal registry.
	mock.getResult = lang.NilValue{} // Return nil
	_, err = api.RunProcedure(context.Background(), interp, "main")

	// FIX: We *expect* an error here because the script tries c["content"] on nil
	if err == nil {
		t.Fatal("api.RunProcedure() for nil check succeeded, but was expected to fail with a runtime error")
	}

	var rtErr *lang.RuntimeError
	if errors.As(err, &rtErr) {
		// FIX: Use the correct error code from pkg/lang/errors.go
		// Error 7 is ErrorCodeType, which is returned for "invalid operation" / "cannot access elements on type nil"
		if rtErr.Code != lang.ErrorCodeType {
			t.Errorf("Expected ErrorCodeType (%d), but got code %d: %v",
				lang.ErrorCodeType, rtErr.Code, err)
		} else {
			// This is the desired outcome
			t.Logf("Correctly received expected error: %v", err)
		}
	} else {
		t.Errorf("Expected a *lang.RuntimeError, but got %T: %v", err, err)
	}
}
*/
