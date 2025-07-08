// filename: pkg/lang/error_gate.go
// NeuroScript Version: 0.5.2
// File version: 0.1.0
// Purpose: Central error gate – spot critical errors and invoke a
//          pluggable handler (panic by default).
// nlines: 120
// risk_rating: LOW

package lang

import (
	"errors"
	"fmt"
	"sync/atomic"
)

// ---------------------------------------------------------------------------
// 0.  Critical-code registry
// ---------------------------------------------------------------------------

// criticalCodes lists every ErrorCode that should trigger the global handler.
// Extend this slice when you add new severity levels.
var criticalCodes = [...]ErrorCode{
	ErrorCodeSecurity,
	ErrorCodeAttackProbable,
	ErrorCodeAttackCertain,
	ErrorCodeSubsystemCompromised,
	ErrorCodeSubsystemQuarantined,
	ErrorCodeInternal,
}

func isCritical(code ErrorCode) bool {
	for _, c := range criticalCodes {
		if c == code {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// 1.  Handler plumbing
// ---------------------------------------------------------------------------

// CriticalHandler is invoked exactly once per *critical* RuntimeError.
// Swap it in neurogo / main() to integrate with slog, Sentry, etc.
var CriticalHandler = func(e *RuntimeError) {
	// default behaviour: crash fast so ops notice
	panic(fmt.Sprintf("[FDM-CRITICAL] %s", e.Error()))
}

// Statistic for dashboards / tests
var CriticalCount atomic.Uint64

// RegisterCriticalHandler lets the host override the default panic.
func RegisterCriticalHandler(h func(*RuntimeError)) {
	if h != nil {
		CriticalHandler = h
	}
}

// ---------------------------------------------------------------------------
// 2.  One helper the whole code-base should call
// ---------------------------------------------------------------------------

// Check is a thin wrapper to inspect and forward errors.
//
// Call it at EVERY return-point where an error bubbles up:
//
//     if err := lang.Check(err); err != nil { return nil, err }

func Check(err error) error {
	if err == nil {
		return nil
	}

	// -- ADD THIS BLOCK --
	// Specifically prevent ErrToolNotAllowed from triggering a critical panic.
	if errors.Is(err, ErrToolNotFound) {
		return err
	}
	// -- END BLOCK --

	var re *RuntimeError
	if errors.As(err, &re) {
		if isCritical(re.Code) {
			CriticalCount.Add(1)
			CriticalHandler(re)
		}
		return re
	}

	// Not a RuntimeError – wrap it so callers see a consistent type.
	wrapped := NewRuntimeError(ErrorCodeInternal, "wrapped non-runtime error", err)
	CriticalCount.Add(1)
	CriticalHandler(wrapped)
	return wrapped
}

// ---------------------------------------------------------------------------
// 3.  Example replacement handler (drop in neurogo/engine init)
// ---------------------------------------------------------------------------

// func init() {
//     lang.RegisterCriticalHandler(func(e *lang.RuntimeError) {
//         log.Printf("[CRIT] %v", e)
//         // clean shutdown, alerting, etc.
//         os.Exit(1)
//     })
// }

// ---------------------------------------------------------------------------
// 4.  Convenience helpers for callers
// ---------------------------------------------------------------------------

// IsCritical returns true if err (or any wrapped error) carries
// a critical ErrorCode.
func IsCritical(err error) bool {
	var re *RuntimeError
	if errors.As(err, &re) {
		return isCritical(re.Code)
	}
	return false
}

// Must is syntactic sugar: panic on critical, else return the value.
func Must[T any](v T, err error) T {
	if err != nil {
		Check(err) // will panic if critical
	}
	return v
}
