// NeuroScript Version: 0.3.0
// File version: 0.1.0 // Updated version
// Removed runCleanAPIMode as it's no longer a distinct mode
// filename: pkg/neurogo/app_helpers.go
package neurogo

// Removed unused imports: "bufio", "context", "os", "strings", "sync", "time"
// Removed ApiFileInfo import/placeholder if it was only for runCleanAPIMode

// min function remains if needed elsewhere, otherwise can be removed too.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// runCleanAPIMode function REMOVED
/*
func (a *App) runCleanAPIMode(ctx context.Context) error {
	// ... entire function body removed ...
}
*/
