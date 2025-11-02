:: version: 1.0.0
:: component: Interpreter Core
:: author: AJP
:: purpose: To track the refactoring of the ProviderRegistry and implementation of the generic http.Provider.

- | | **Phase 1: Fix ProviderRegistry Integration** #(proj-reg)
  - | | Modify `pkg/interpreter/interpreter.go` #(p1-interp)
    - [ ] Add `providerRegistry *provider.Registry` field to the `Interpreter` struct. #(c1k4)
    - [ ] Add `SetProviderRegistry(registry *provider.Registry)` method (sets `i.rootInterpreter().providerRegistry`). #(s9p7)
    - [ ] Remove the old `RegisterProvider` method (the one modifying `i.state.providers`). #(r2d6)
  - [ ] Modify `pkg/interpreter/options.go` to correctly call `i.SetProviderRegistry()` in `WithProviderRegistry`. #(p1-opts)
  - [ ] Modify `pkg/interpreter/clone.go` to ensure `clone.providerRegistry = i.providerRegistry` in the `fork()` method. #(p1-clone)
  - | | Modify `pkg/interpreter/api.go` #(p1-api)
    - [ ] Update `GetProvider(name string)` to use `i.rootInterpreter().providerRegistry.Get(name)`. #(g8p3)
    - [ ] Remove the obsolete `RegisterProvider` method. #(r4p1)
  - [ ] Modify `pkg/interpreter/state_2.go` to remove `providers` map and `providersMu` from `interpreterState`. #(p1-state2)

- | | **Phase 2: Implement Generic `http.Provider`** #(proj-http)
  - [ ] Create new directory `pkg/provider/httpprovider/`. #(p2-dir)
  - | | Implement `pkg/provider/httpprovider/httpprovider.go` #(p2-impl)
    - [ ] Define a `Provider` struct that implements `provider.AIProvider`. #(p2-struct)
    - [ ] Implement the `Chat()` method. #(p2-chat)
    - [ ] Add logic to `Chat()` to read `api_url`, `api_headers`, `api_body_template`, `api_response_path` from `req.Model`. #(p2-config)
    - [ ] Add logic to replace `{API_KEY}`, `{MODEL}`, `{PROMPT}` tokens in headers and body. #(p2-token)
    - [ ] Add `http.Post` call to execute the request. #(p2-post)
    - [ ] Add JSONPath/JMESPath logic to extract response text using `api_response_path`. #(p2-extract)

- | | **Phase 3: Update Tests** #(proj-test)
  - | | Modify `pkg/interpreter/testing_helpers_test.go` (`NewTestHarness`) #(p3-harness)
    - [ ] Create a `*provider.Registry` in the harness setup. #(p3-reg)
    - [ ] Register all common mock providers (like `test.New()`) into this registry. #(p3-reg-mocks)
    - [ ] Pass the registry to `interpreter.NewInterpreter` using `WithProviderRegistry()`. #(p3-pass)
  - | | Refactor all `_test.go` files #(p3-refactor)
    - [ ] Search all `interpreter/*_test.go` files. #(p3-search)
    - [ ] Remove all calls to `interp.RegisterProvider()`. #(p3-remove)
    - [ ] Ensure tests correctly get mock providers (e.g., "mock_ask_provider") from the harness's registry. #(p3-verify)

- | | **Phase 4: Validation** #(proj-validate)
  - [ ] Confirm all tests pass, especially `ask_e2e_test.go` and `context_propagation_test.go`. #(v1-tests)
  - [ ] Manually verify that no changes were made to `steps_ask_hostloop.go` or `steps_ask_aeiou.go`. #(v2-askloop)