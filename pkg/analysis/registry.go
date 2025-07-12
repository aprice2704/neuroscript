// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Implements the registry and Vet function for running static analysis passes.
// filename: pkg/analysis/registry.go
// nlines: 40
// risk_rating: MEDIUM

package analysis

import (
	"sync"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

var (
	passesMu sync.RWMutex
	passes   = make(map[string]interfaces.Pass)
)

// RegisterPass adds a new analysis pass to the global registry.
// It will panic if a pass with the same name is registered twice.
func RegisterPass(p interfaces.Pass) {
	passesMu.Lock()
	defer passesMu.Unlock()
	if p == nil {
		panic("cannot register a nil analysis pass")
	}
	name := p.Name()
	if _, dup := passes[name]; dup {
		panic("analysis pass registered twice: " + name)
	}
	passes[name] = p
}

// Vet runs all registered analysis passes on the given AST and returns a
// consolidated list of diagnostics.
func Vet(tree *interfaces.Tree) []interfaces.Diag {
	passesMu.RLock()
	// Create a snapshot of the passes to run so we don't hold the lock
	// during the analysis itself.
	passesToRun := make([]interfaces.Pass, 0, len(passes))
	for _, p := range passes {
		passesToRun = append(passesToRun, p)
	}
	passesMu.RUnlock()

	var allDiags []interfaces.Diag
	for _, p := range passesToRun {
		diags := p.Analyse(tree)
		if len(diags) > 0 {
			allDiags = append(allDiags, diags...)
		}
	}
	return allDiags
}
