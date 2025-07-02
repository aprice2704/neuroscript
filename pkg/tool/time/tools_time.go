// NeuroScript Version: 0.4.1
// File version: 2
// Purpose: Implements the Go function for the 'Time.Now' tool.
// filename: pkg/tool/time/tools_time.go
// nlines: 15
// risk_rating: LOW

package time

import (
	"fmt"
	"time"

	"github.com/aprice2704/neuroscript/pkg/tool"
)

// ================================================================================
//
//	Adapter/Bridge Layer (Matches the ToolFunc signature)
//
// ================================================================================

func adaptToolTimeNow(interp tool.RunTime, args []interface{}) (interface{}, error) {
	if err := validateTimeNow(args); err != nil {
		return nil, err
	}
	return implTimeNow()
}

func adaptToolTimeSleep(interp tool.RunTime, args []interface{}) (interface{}, error) {
	if err := validateTimeSleep(args); err != nil {
		return nil, err
	}
	// We know from validation that args[0] is a float64.
	durationSeconds := args[0].(float64)
	return implTimeSleep(durationSeconds)
}

// ================================================================================
//
//	Validation Layer
//
// ================================================================================

func validateTimeNow(args []interface{}) error {
	if len(args) != 0 {
		return fmt.Errorf("validation error for Time.Now: expected 0 arguments, but got %d", len(args))
	}
	return nil
}

func validateTimeSleep(args []interface{}) error {
	if len(args) != 1 {
		return fmt.Errorf("validation error for Time.Sleep: expected 1 argument, but got %d", len(args))
	}
	if _, ok := args[0].(float64); !ok {
		return fmt.Errorf("validation error for Time.Sleep: argument must be a number, but got %T", args[0])
	}
	return nil
}

// ================================================================================
//
//	Raw Tool Implementation Layer
//
// ================================================================================

func implTimeNow() (time.Time, error) {
	return time.Now(), nil
}

func implTimeSleep(durationSeconds float64) (bool, error) {
	if durationSeconds < 0 {
		return false, fmt.Errorf("sleep duration cannot be negative, got %f", durationSeconds)
	}
	sleepDuration := time.Duration(durationSeconds * float64(time.Second))
	time.Sleep(sleepDuration)
	return true, nil
}
