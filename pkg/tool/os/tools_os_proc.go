// NeuroScript Version: 0.5.2
// File version: 5
// Purpose: Implements OS tools. Corrected Sleep to use the GetGrantSet() interface method.
// filename: pkg/tool/os/tools_os_proc.go
// nlines: 53
// risk_rating: HIGH

package os

import (
	"fmt"
	"os"
	"time"

	"github.com/aprice2704/neuroscript/pkg/lang"
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

	// Enforce policy limit before sleeping by using the GetGrantSet interface method.
	grantSet := interpreter.GetGrantSet()
	if grantSet == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "Sleep: could not access grant set from runtime", lang.ErrInternal)
	}
	if err := grantSet.CheckSleep(duration); err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodePolicy, err.Error(), err)
	}

	// interpreter.GetLogger().Debug("Tool: Sleep", "duration_seconds", duration)
	time.Sleep(time.Duration(duration * float64(time.Second)))
	return nil, nil
}

func toolNow(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Now: expected 0 arguments", lang.ErrArgumentMismatch)
	}
	now := float64(time.Now().Unix())
	// interpreter.GetLogger().Debug("Tool: Now", "timestamp", now)
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
	// interpreter.GetLogger().Debug("Tool: Hostname", "hostname", hostname)
	return hostname, nil
}
