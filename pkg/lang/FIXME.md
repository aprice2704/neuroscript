# NS -- things that need to be fixed

1. some odd syntax does not trigger parse error
a. e.g. let instead of set

b. must err == nil is redundant

2. need to make a lib for handles and use it everywhere

this script:
// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Registers provider accounts from environment variables.
// filename: config/accounts.nsc
// risk_rating: HIGH (contains credential logic)

print("--- Registering Provider Accounts ---")

// It's best practice to fetch secrets from the environment
// rather than hardcoding them. We assume a tool like `env.Get` exists.
let openai_api_key = env.Get("OPENAI_API_KEY")
let google_api_key = env.Get("GOOGLE_API_KEY")

if openai_api_key == nil || openai_api_key == "" {
    print("WARN: OPENAI_API_KEY environment variable not set. Skipping.")
} else {
    let success = account.Register("openai-main", {
        "kind": "llm",
        "provider": "openai",
        "apiKey": openai_api_key,
        "notes": "Primary account for general purpose tasks."
    })

    if success {
        print("✅ Registered account: openai-main")
    }
}


if google_api_key == nil || google_api_key == "" {
    print("WARN: GOOGLE_API_KEY environment variable not set. Skipping.")
} else {
    let success = account.Register("google-gemini", {
        "kind": "llm",
        "provider": "google",
        "apiKey": google_api_key,
        "notes": "Primary account for Gemini models."
    })

    if success {
        print("✅ Registered account: google-gemini")
    }
}

print("\n--- Verification: All Registered Accounts ---")
let current_accounts = account.List()
print(current_accounts)
print("------------------------------------------")

doesn't throw errors on the prints, even though they should be emit


4. should allow whitespace after \ to be line continuations (or render unneccessary)

5. Need to rationalize types:

Okay, let's sort this out.

You're right, the type names used in the tool specs (`slice_string`, `list`, `any`, etc.) seem inconsistent and are causing these coercion headaches. Rationalizing them is a good idea.

First, here's the corrected code block for `SetRespectGitignore` in `tooldefs_fdm.go`. I've removed the erroneous line:

```go
		{
			Spec: api.ToolSpec{
				Name:        "SetRespectGitignore",
				Group:       "FDM.Plan",
				Description: "Sets whether the ingestion plan should respect .gitignore files.",
				Args: []api.ArgSpec{
					{Name: "planID", Type: "string", Required: true, Description: "The NodeID of the ingest_plan to modify."},
					// Corrected line below:
					{Name: "respect", Type: "bool", Required: true, Description: "Set to true to respect .gitignore, false to ignore it."},
				},
				ReturnType: "string", // Returns new NodeID
				ReturnHelp: "Returns the NodeID of the new, updated ingest_plan node.",
			},
			Func: mediator.ToolPlanSetRespectGitignore,
		},
```

-----

## Tool Argument Type Options

Based on the `lang` package files and the `tooldefs_fdm.go` file, here are the type names I can identify being used or implied for `api.ArgSpec.Type`:

**Directly from `lang/type_names.go`:**

  * `string`
  * `number` (maps to Go `float64`)
  * `boolean` (seems `bool` is used in practice)
  * `bytes`
  * `list`
  * `map`
  * `nil` (unlikely to be used as an *expected* argument type)
  * `function` (unlikely for tool args)
  * `tool` (unlikely for tool args)
  * `error` (unlikely for tool args)
  * `event` (unlikely for tool args)
  * `timedate`
  * `fuzzy`

**Observed in `tooldefs_fdm.go`:**

  * `string`
  * `map`
  * `int` (This implies the coercion layer *might* handle Go `int`/`int64` separately from `number`/`float64`)
  * `bool` (Used instead of `boolean`)
  * `slice_string` (The source of our current problem, likely intended to map to Go `[]string`)
  * `any` (The most flexible type, likely mapping to Go `any` or `interface{}`)

**Implicit/Potential:**

  * `float` / `float64`: Possibly distinct from `number` or `int`.
  * `slice_map` / `list_map`: Used as a *return* type, could potentially be an input type.

**Summary of Confusion:**

The core issue seems to be the inconsistent mapping between:

1.  NeuroScript's internal types (`lang.Value` implementations).
2.  The string names used in `api.ArgSpec.Type`.
3.  The expected Go types in the tool's implementation function signature.
4.  How the `lang.Unwrap` function converts NS types to Go types.
5.  How the tool-calling machinery (coercion layer) uses the `Type` string to validate the unwrapped Go type against the tool function's signature.

The types `slice_string`, `int`, and `bool` (vs `boolean`) strongly suggest the `api.ArgSpec.Type` string is intended to map more closely to Go types than to the core NeuroScript types. `list` and `map` seem to map to the NS/Go concepts, while `any` is the catch-all.

We definitely need to clarify and document the canonical set of allowed `Type` strings and exactly which Go types they correspond to during coercion. Using `"any"` in the spec for the list arguments seems like the most robust path forward for now, given the implementation already handles `[]any`.