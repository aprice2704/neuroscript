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