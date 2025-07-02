package parser

import (
	"reflect"
	"testing"
)

func TestConvertInputSchemaToArgSpec_SuccessScenarios(t *testing.T) {
	testCases := []struct {
		name           string
		schema         map[string]interface{}
		expectedArgs   []ArgSpec
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:         "schema with no properties",
			schema:       map[string]interface{}{"type": "object"},
			expectedArgs: []ArgSpec{},
			expectError:  false,
		},
		{
			name: "schema with properties but no required field",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"param1": map[string]interface{}{"type": "string"},
				},
			},
			expectedArgs: []ArgSpec{
				{Name: "param1", Type: ArgTypeString, Required: false},
			},
			expectError: false,
		},
		{
			name: "schema with empty required field",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"param1": map[string]interface{}{"type": "integer"},
				},
				"required": []string{},
			},
			expectedArgs: []ArgSpec{
				{Name: "param1", Type: ArgTypeInt, Required: false},
			},
			expectError: false,
		},
		{
			name: "invalid required array element type",
			schema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []interface{}{123},
			},
			expectError:    true,
			expectedErrMsg: "invalid schema: 'required' array element 0 is not a string (int)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args, err := ConvertInputSchemaToArgSpec(tc.schema)

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected an error but got nil")
				}
				if err.Error() != tc.expectedErrMsg {
					t.Errorf("Expected error message '%s', got '%s'", tc.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Did not expect an error, but got: %v", err)
				}
				if !reflect.DeepEqual(args, tc.expectedArgs) {
					t.Errorf("Expected args %+v, but got %+v", tc.expectedArgs, args)
				}
			}
		})
	}
}

func TestParseMetadataLine(t *testing.T) {
	testCases := []struct {
		name        string
		line        string
		expectedKey string
		expectedVal string
		expectedOk  bool
	}{
		{"valid full line", ":: key: value", "key", "value", true},
		{"valid with extra space", "  ::  key  :  value  ", "key", "value", true},
		{"valid key only", ":: key_only", "key_only", "", true},
		{"valid with no space after colon", ":: key:value", "key", "value", true},
		{"invalid no key", ":: : value", "", "", false},
		{"invalid not a metadata line", "key: value", "", "", false},
		{"invalid empty line", "::", "", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key, val, ok := ParseMetadataLine(tc.line)
			if ok != tc.expectedOk {
				t.Errorf("Expected ok to be %v, but got %v", tc.expectedOk, ok)
			}
			if key != tc.expectedKey {
				t.Errorf("Expected key '%s', got '%s'", tc.expectedKey, key)
			}
			if val != tc.expectedVal {
				t.Errorf("Expected value '%s', got '%s'", tc.expectedVal, val)
			}
		})
	}
}
