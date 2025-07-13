// NeuroScript Version: 0.5.2
// File version: 6
// Purpose: Provides the final execution entrypoint, correcting the import path for the secret package.
// filename: pkg/api/exec.go
// nlines: 41
// risk_rating: HIGH

package api

import (
	"context"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interp"

	"github.com/aprice2704/neuroscript/pkg/api/secret"
)

// ExecConfig holds the configuration for an execution run.
type ExecConfig struct {
	Cache         Cache
	SecretPrivKey []byte
	MaxGas        uint64
}

// Exec executes a pre-loaded and vetted script.
func Exec(ctx context.Context, lu *LoadedUnit, cfg ExecConfig) (*interfaces.ExecResult, error) {
	if lu == nil || lu.Tree == nil {
		return nil, fmt.Errorf("cannot execute a nil LoadedUnit or unit with a nil tree")
	}

	// The internal interpreter expects a resolver with the signature `func(*ast.SecretRef) (string, error)`.
	resolver := func(ref *ast.SecretRef) (string, error) {
		// Convert the internal *ast.SecretRef to the public secret.Ref type for decoding.
		secretRef := secret.Ref{
			Path: ref.Path,
			Enc:  ref.Enc,
			Raw:  ref.Raw,
		}
		return secret.Decode(secretRef, cfg.SecretPrivKey)
	}

	// Create the configuration for the internal interpreter.
	interpCfg := interp.Config{
		ResolveSecret: resolver,
	}

	// Call the internal execution function with the verified AST and config.
	return interp.ExecCommand(ctx, lu.Tree, interpCfg)
}
