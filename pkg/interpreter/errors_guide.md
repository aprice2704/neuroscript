# NeuroScript / FDM — Error-Handling Cookbook

This guide shows **when** and **how** to create, propagate, and classify
errors across the entire code-base.  It is aligned with **the actual Go
code you have today**:

* `pkg/lang/errors.go`          – type `RuntimeError`, enum `ErrorCode`
* `pkg/lang/error_gate.go`      – central “critical” filter (`lang.Check`)

---

## TL;DR

Quick Guide: Adding `lang.Check` to Existing Files
To integrate a file with the central error gate, you only need to modify the standard Go error check. The goal is to ensure every error that gets returned "bubbles up" through 

```golang
lang.Check.
``` 

The pattern is a simple, one-line change.

### Before
Your original code will look something like this:

Go
```golang
value, err := someFunction()
if err != nil {
    return nil, err
}
```

### After

Wrap the err variable with lang.Check inside the if statement. The := shadows the original err variable, which is standard Go practice.

Go
```golang
value, err := someFunction()
if err := lang.Check(err); err != nil {
    return nil, err
}
```

This simple change ensures that every potential error is inspected by the central handler, which will automatically panic on critical errors or pass non-critical ones through. 

## 1  Canonical error type & helpers

```go
// errors.go  (already exists)
type ErrorCode int

const (
    ErrorCodeGeneric ErrorCode = iota
    // ... many other codes
    ErrorCodeSecurity
    ErrorCodeAttackProbable
    ErrorCodeAttackCertain
    ErrorCodeSubsystemCompromised
    ErrorCodeInternal
)

type RuntimeError struct {
    Code     ErrorCode
    Message  string // updated from Msg
    Wrapped  error  // updated from Err
    Position *Position
}

func NewRuntimeError(code ErrorCode, message string, wrapped error) *RuntimeError // updated signature
```

---

## 2  Critical-vs-non-critical mapping

| `ErrorCode`                         | Critical? | Typical source                |
|------------------------------------|-----------|--------------------------------|
| `ErrorCodeSyntax`                  | no        | Parser / AST builder           |
| `ErrorCodeEvaluation`              | no        | Interpreter step               |
| `ErrorCodeResourceExhaustion`      | no* | Quota exhausted → Script stop  |
| `ErrorCodeSecurity`                | **YES** | Path escape, ACL break         |
| `ErrorCodeAttackProbable` / `Certain` | **YES**| Heuristic agent      |
| `ErrorCodeSubsystemCompromised`    | **YES** | Integrity checker              |
| `ErrorCodeSubsystemQuarantined`    | **YES** | Security agent                 |
| `ErrorCodeInternal`                | **YES** | Logic bug, invariant break     |

`ResourceExhaustion` becomes critical only if the system fails to throttle; normally
scripts just receive an error and stop.

**Critical codes** are listed in `error_gate.go::criticalCodes` and trigger the global handler.

---

## 3  The *one* helper every layer calls

```go
err := doSomething()
if err := lang.Check(err); err != nil {
    return nil, err   // bubble up
}
```

Behaviour of `lang.Check` (_error_gate.go_):

1. **nil** → returns nil. 
2. **RuntimeError**    * critical   → increments metric `lang.CriticalCount` and invokes
     `lang.CriticalHandler` (default is `panic`).
   * non-critical → passes straight through.
3. **plain error** (not `RuntimeError`)  
   → automatically wrapped with `ErrorCodeInternal`, counted as critical.

Keep code‐paths clean: *do not* log or panic locally; rely on `Check`.

---

## 4  Overriding the critical handler (engine init)

```go
func init() {
    lang.RegisterCriticalHandler(func(e *lang.RuntimeError) {
        slog.Error("critical", "code", e.Code, "msg", e.Message)
        metrics.Inc("critical_total")
        os.Exit(1)   // or supervisor restart
    })
}
```

---

## 5  Convenience helpers

* **`lang.IsCritical(err)`** – true if the wrapped code is critical.
* **`lang.Must(v, err)`** – panic via `Check` if critical, else return v.

```go
fd := lang.Must(os.Open(path))
```

---

## 6  Checklist for contributors

* [ ] Construct **only** `*lang.RuntimeError` in production paths.
* [ ] Call `lang.Check(err)` exactly once per return edge.
* [ ] Use the correct `ErrorCode…` – pick **Security** for sandbox or ACL
      violations.
* [ ] Let `CriticalHandler` own logging / shutdown; avoid duplicate logs.

Do this and every critical failure bubbles to the central gate, ensuring
consistent metrics, clean shutdown, and predictable behaviour across FDM,
NeuroScript, the Gateway, and all tool-sets.