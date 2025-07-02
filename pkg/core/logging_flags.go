// filename: pkg/core/logging_flags.go
// pkg/core/logging_flags.go
// file version: 1
package core

import "flag"

// TestVerbose enables noisy test logging when set before NewTestLogger is called.
var TestVerbose = flag.Bool("nslog", false, "enable verbose NeuroScript test logs")