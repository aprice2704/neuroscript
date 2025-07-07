// filename: pkg/lang/error_gate_test.go
package lang

import (
	"errors"
	"testing"
)

// setupCustomHandler overrides the global critical handler for a single test
// and returns a function to check if it was called. It uses t.Cleanup
// to restore the original state automatically.
func setupCustomHandler(t *testing.T) func() bool {
	t.Helper()

	var handlerCalled bool
	originalHandler := CriticalHandler
	originalCount := CriticalCount.Load()

	RegisterCriticalHandler(func(e *RuntimeError) {
		handlerCalled = true
		// We don't panic in the test handler
	})

	t.Cleanup(func() {
		CriticalHandler = originalHandler
		CriticalCount.Store(originalCount)
	})

	return func() bool {
		return handlerCalled
	}
}

func TestCheck(t *testing.T) {
	t.Run("with nil error", func(t *testing.T) {
		wasCalled := setupCustomHandler(t)
		initialCount := CriticalCount.Load()

		err := Check(nil)

		if err != nil {
			t.Errorf("Check(nil) should return nil, but got: %v", err)
		}
		if wasCalled() {
			t.Error("CriticalHandler should not be called for nil error")
		}
		if CriticalCount.Load() != initialCount {
			t.Error("CriticalCount should not be incremented for nil error")
		}
	})

	t.Run("with non-critical RuntimeError", func(t *testing.T) {
		wasCalled := setupCustomHandler(t)
		initialCount := CriticalCount.Load()
		nonCriticalErr := NewRuntimeError(ErrorCodeRateLimited, "rate limited", nil)

		err := Check(nonCriticalErr)

		if err != nonCriticalErr {
			t.Errorf("Check should pass non-critical errors through, but it was modified")
		}
		if wasCalled() {
			t.Error("CriticalHandler should not be called for non-critical errors")
		}
		if CriticalCount.Load() != initialCount {
			t.Error("CriticalCount should not be incremented for non-critical errors")
		}
	})

	t.Run("with critical RuntimeError", func(t *testing.T) {
		wasCalled := setupCustomHandler(t)
		initialCount := CriticalCount.Load()
		criticalErr := NewRuntimeError(ErrorCodeSecurity, "security breach", nil)

		err := Check(criticalErr)

		if err != criticalErr {
			t.Errorf("Check should return the original critical error, got %v", err)
		}
		if !wasCalled() {
			t.Error("CriticalHandler was not called for a critical error")
		}
		if CriticalCount.Load() != initialCount+1 {
			t.Errorf("CriticalCount should be incremented for critical errors, got %d", CriticalCount.Load())
		}
	})

	t.Run("with plain error", func(t *testing.T) {
		wasCalled := setupCustomHandler(t)
		initialCount := CriticalCount.Load()
		plainErr := errors.New("a plain error")

		err := Check(plainErr)

		re, ok := err.(*RuntimeError)
		if !ok {
			t.Fatalf("Check should have wrapped the plain error in a RuntimeError, but it didn't")
		}
		if re.Code != ErrorCodeInternal {
			t.Errorf("Wrapped error should have code ErrorCodeInternal, but got %v", re.Code)
		}
		if !wasCalled() {
			t.Error("CriticalHandler was not called for a plain error, but it should be (as it's critical)")
		}
		if CriticalCount.Load() != initialCount+1 {
			t.Errorf("CriticalCount should be incremented for plain errors, got %d", CriticalCount.Load())
		}
	})
}

func TestMust(t *testing.T) {
	t.Run("with nil error", func(t *testing.T) {
		val := Must("success", nil)
		if val != "success" {
			t.Errorf("Must should return the value when error is nil, got %s", val)
		}
	})

	t.Run("with non-critical error", func(t *testing.T) {
		// Must() will not panic on a non-critical error
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Must should not panic on non-critical error, but it did with: %v", r)
			}
		}()
		nonCriticalErr := NewRuntimeError(ErrorCodeRateLimited, "slow down", nil)
		val := Must("success", nonCriticalErr)
		if val != "success" {
			t.Errorf("Must should return the value for non-critical errors, got %s", val)
		}
	})

	t.Run("with critical error", func(t *testing.T) {
		// Must() MUST panic on a critical error
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Must did not panic on a critical error")
			}
		}()
		criticalErr := NewRuntimeError(ErrorCodeSecurity, "access denied", nil)
		// This line should panic
		_ = Must("failure", criticalErr)
	})
}

func TestIsCritical(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"critical code", NewRuntimeError(ErrorCodeSecurity, "critical", nil), true},
		{"non-critical code", NewRuntimeError(ErrorCodeRateLimited, "non-critical", nil), false},
		{"plain error (becomes internal)", errors.New("plain"), false}, // IsCritical doesn't wrap, just checks
		{"nil error", nil, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsCritical(tc.err); got != tc.expected {
				t.Errorf("IsCritical(%v) = %v; want %v", tc.err, got, tc.expected)
			}
		})
	}
}
