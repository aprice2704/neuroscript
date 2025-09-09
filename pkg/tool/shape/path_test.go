// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides unit tests for the tool.shape.IsValidPath function.
// filename: pkg/tool/shape/path_test.go
// nlines: 35
// risk_rating: LOW

package shape_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestShapeIsValidPath(t *testing.T) {
	testShapeToolHelper(t, "Valid Path", func(t *testing.T, interp tool.Runtime) {
		result, err := runTool(t, interp, "IsValidPath", "a.b[0].c-d_e")
		assertResult(t, result, err, true, nil)
	})

	testShapeToolHelper(t, "Invalid Path (double dot)", func(t *testing.T, interp tool.Runtime) {
		result, err := runTool(t, interp, "IsValidPath", "a..b")
		assertResult(t, result, err, false, nil)
	})

	testShapeToolHelper(t, "Invalid Path (bad index)", func(t *testing.T, interp tool.Runtime) {
		result, err := runTool(t, interp, "IsValidPath", "a[x]")
		assertResult(t, result, err, false, nil)
	})

	testShapeToolHelper(t, "Invalid Argument Type", func(t *testing.T, interp tool.Runtime) {
		_, err := runTool(t, interp, "IsValidPath", 123)
		assertResult(t, nil, err, nil, lang.ErrInvalidArgument)
	})
}
