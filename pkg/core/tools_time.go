// NeuroScript Version: 0.4.1
// File version: 2
// Purpose: Implements the Go function for the 'Time.Now' tool.
// filename: core/tools_time.go
// nlines: 15
// risk_rating: LOW

package core

import (
	"fmt"
	"time"
)

// toolTimeNow implements the "Time.Now" tool function.
func toolTimeNow(i *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("Time.Now() expects 0 arguments, got %d", len(args))
	}
	return TimedateValue{Value: time.Now()}, nil
}
