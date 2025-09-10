// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Implements tools for OS process and time functions. Corrected policy access in Sleep to use new Policy() getter.
// filename: pkg/tool/os/tools_os_proc.go
// nlines: 61
// risk_rating: HIGH

package os

import (
	"fmt"
	"os"
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func toolSleep(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Sleep: expected 1 argument (duration_seconds)", lang.ErrArgumentMismatch)
	}
	duration, ok := args[0].(float64)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Sleep: duration_seconds must be a number, got %T", args[0]), lang.ErrArgumentMismatch)
	}
	if duration < 0 {
		duration = 0
	}

	// The tool.Runtime interface doesn't expose the policy directly. We perform a
	// type assertion to access the concrete interpreter's Policy() method.
	interpImpl, ok := interpreter.(interface{ Policy() *policy.ExecPolicy })
	if !ok {
		// This should not happen in the standard interpreter.
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "Sleep: could not access policy from runtime", lang.ErrInternal)
	}

	// Enforce policy limit before sleeping
	if err := interpImpl.Policy().Grants.CheckSleep(duration); err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodePolicy, err.Error(), err)
	}

	interpreter.GetLogger().Debug("Tool: Sleep", "duration_seconds", duration)
	time.Sleep(time.Duration(duration * float64(time.Second)))
	return nil, nil
}

func toolNow(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Now: expected 0 arguments", lang.ErrArgumentMismatch)
	}
	now := float64(time.Now().Unix())
	interpreter.GetLogger().Debug("Tool: Now", "timestamp", now)
	return now, nil
}

func toolHostname(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Hostname: expected 0 arguments", lang.ErrArgumentMismatch)
	}
	hostname, err := os.Hostname()
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("Hostname: failed to get hostname: %v", err), lang.ErrInternal)
	}
	interpreter.GetLogger().Debug("Tool: Hostname", "hostname", hostname)
	return hostname, nil
}
