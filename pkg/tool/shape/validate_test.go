// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides unit tests for the tool.shape.Validate tool.
// filename: pkg/tool/shape/validate_test.go
// nlines: 95
// risk_rating: LOW

package shape_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/json_lite"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestShapeValidate(t *testing.T) {
	shape := map[string]interface{}{
		"name":  "string",
		"email": "email",
	}
	validData := map[string]interface{}{
		"name":  "Ada",
		"email": "ada@example.com",
	}

	testShapeToolHelper(t, "Valid Data", func(t *testing.T, interp tool.Runtime) {
		result, err := runTool(t, interp, "Validate", validData, shape)
		assertResult(t, result, err, true, nil)
	})

	testShapeToolHelper(t, "Missing Required Field", func(t *testing.T, interp tool.Runtime) {
		invalidData := map[string]interface{}{"name": "Ada"}
		_, err := runTool(t, interp, "Validate", invalidData, shape)
		assertResult(t, nil, err, nil, json_lite.ErrValidationRequiredArgMissing)
	})

	testShapeToolHelper(t, "Type Mismatch", func(t *testing.T, interp tool.Runtime) {
		invalidData := map[string]interface{}{"name": 123, "email": "ada@example.com"}
		_, err := runTool(t, interp, "Validate", invalidData, shape)
		assertResult(t, nil, err, nil, json_lite.ErrValidationTypeMismatch)
	})

	testShapeToolHelper(t, "Extra Field Not Allowed", func(t *testing.T, interp tool.Runtime) {
		extraData := map[string]interface{}{"name": "Ada", "email": "a@b.com", "extra": true}
		_, err := runTool(t, interp, "Validate", extraData, shape)
		assertResult(t, nil, err, nil, json_lite.ErrInvalidArgument)
	})

	testShapeToolHelper(t, "Extra Field Allowed", func(t *testing.T, interp tool.Runtime) {
		extraData := map[string]interface{}{"name": "Ada", "email": "a@b.com", "extra": true}
		options := map[string]interface{}{"allow_extra": true}
		result, err := runTool(t, interp, "Validate", extraData, shape, options)
		assertResult(t, result, err, true, nil)
	})

	testShapeToolHelper(t, "Case Insensitive Pass", func(t *testing.T, interp tool.Runtime) {
		caseData := map[string]interface{}{"NAME": "Ada", "EMAIL": "a@b.com"}
		options := map[string]interface{}{"case_insensitive": true}
		result, err := runTool(t, interp, "Validate", caseData, shape, options)
		assertResult(t, result, err, true, nil)
	})

	testShapeToolHelper(t, "Invalid Shape", func(t *testing.T, interp tool.Runtime) {
		invalidShape := map[string]interface{}{"name": 123}
		_, err := runTool(t, interp, "Validate", validData, invalidShape)
		assertResult(t, nil, err, nil, json_lite.ErrValidationTypeMismatch)
	})

	testShapeToolHelper(t, "Invalid Arguments", func(t *testing.T, interp tool.Runtime) {
		_, err := runTool(t, interp, "Validate", "not a map", shape)
		assertResult(t, nil, err, nil, lang.ErrInvalidArgument)

		_, err = runTool(t, interp, "Validate", validData, "not a map")
		assertResult(t, nil, err, nil, lang.ErrInvalidArgument)

		_, err = runTool(t, interp, "Validate", validData, shape, "not a map")
		assertResult(t, nil, err, nil, lang.ErrInvalidArgument)
	})
}
