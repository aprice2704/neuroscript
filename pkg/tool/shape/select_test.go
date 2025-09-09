// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Provides unit tests for the tool.shape.Select tool.
// filename: pkg/tool/shape/select_test.go
// nlines: 80
// risk_rating: LOW

package shape_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestShapeSelect(t *testing.T) {
	data := map[string]interface{}{
		"user": map[string]interface{}{"NAME": "Ada"},
		"items": []interface{}{
			map[string]interface{}{"id": float64(100)}, // Corrected: Use float64 to simulate JSON unmarshaling.
		},
	}

	testShapeToolHelper(t, "Simple Path", func(t *testing.T, interp tool.Runtime) {
		result, err := runTool(t, interp, "Select", data, "items[0].id")
		assertResult(t, result, err, float64(100), nil)
	})

	testShapeToolHelper(t, "Array Path", func(t *testing.T, interp tool.Runtime) {
		result, err := runTool(t, interp, "Select", data, []interface{}{"items", 0, "id"})
		assertResult(t, result, err, float64(100), nil)
	})

	testShapeToolHelper(t, "Missing Key Not OK", func(t *testing.T, interp tool.Runtime) {
		_, err := runTool(t, interp, "Select", data, "user.email")
		assertResult(t, nil, err, nil, json_lite.ErrMapKeyNotFound)
	})

	testShapeToolHelper(t, "Missing Key OK", func(t *testing.T, interp tool.Runtime) {
		options := map[string]interface{}{"missing_ok": true}
		result, err := runTool(t, interp, "Select", data, "user.email", options)
		assertResult(t, result, err, nil, nil)
	})

	testShapeToolHelper(t, "Case Sensitive Fail", func(t *testing.T, interp tool.Runtime) {
		_, err := runTool(t, interp, "Select", data, "user.name")
		assertResult(t, nil, err, nil, json_lite.ErrMapKeyNotFound)
	})

	testShapeToolHelper(t, "Case Insensitive Pass", func(t *testing.T, interp tool.Runtime) {
		options := map[string]interface{}{"case_insensitive": true}
		result, err := runTool(t, interp, "Select", data, "user.name", options)
		assertResult(t, result, err, "Ada", nil)
	})

	testShapeToolHelper(t, "Invalid Path", func(t *testing.T, interp tool.Runtime) {
		_, err := runTool(t, interp, "Select", data, "user..name")
		assertResult(t, nil, err, nil, json_lite.ErrInvalidPath)
	})
}
