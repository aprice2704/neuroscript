// NeuroScript Version: 0.3.0
// File version: 0.0.4
// Defines the comprehensive AppAccess interface required by the TUI package.
// Added missing methods used by TUI components.
// filename: pkg/neurogo/interfaces.go
// nlines: 38 // Approximate
// risk_rating: HIGH
package neurogo

import (
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// WMStatusViewDataProvider defines the methods required by the WMStatusScreen
// (or its Formatter) to fetch the data it needs to display.
// The main App struct will implement this interface.
type WMStatusViewDataProvider interface {
	GetLogger() logging.Logger
	GetAIWorkerManager() *core.AIWorkerManager
	// GetModel() *model // If WMStatusScreen needs to update m.lastDisplayedWMDefinitions directly
	// For now, FormatWMStatusView returns the list, and WMStatusScreen.Update passes it to app.model
}
