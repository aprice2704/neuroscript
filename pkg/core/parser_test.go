package core

import (
	"reflect"
	"testing"
)

// === Test Procedure Header Parsing === (Passed Previously)
func TestParseProcedureHeader(t *testing.T) {
	// ... tests unchanged ...
	tests := []struct {
		name      string
		input     string
		expected  Procedure
		errSubstr string
	}{
		{name: "Simple DEFINE", input: "DEFINE PROCEDURE Add(a, b)", expected: Procedure{Name: "Add", Params: []string{"a", "b"}}},
		{name: "Underscores/nums", input: "DEFINE PROCEDURE P1_test(input_1, output_arg)", expected: Procedure{Name: "P1_test", Params: []string{"input_1", "output_arg"}}},
		{name: "No parameters", input: "DEFINE PROCEDURE NoParams()", expected: Procedure{Name: "NoParams", Params: []string{}}},
		{name: "Spaces around params", input: "DEFINE PROCEDURE Spaces( a , b )", expected: Procedure{Name: "Spaces", Params: []string{"a", "b"}}},
		{name: "Spaces after DEFINE", input: "  DEFINE PROCEDURE   WithSpaces   (  p1 , p2 )  ", expected: Procedure{Name: "WithSpaces", Params: []string{"p1", "p2"}}},
		{name: "Empty name", input: "DEFINE PROCEDURE (a, b)", errSubstr: "procedure name cannot be empty"},
		{name: "Missing name and parens", input: "DEFINE PROCEDURE", errSubstr: "missing '(' for parameter list"},
		{name: "Missing parens", input: "DEFINE PROCEDURE MissingParens", errSubstr: "missing '(' for parameter list"},
		{name: "Missing closing paren", input: "DEFINE PROCEDURE MissingParen(a, b", errSubstr: "missing or misplaced ')' after parameter list"},
		{name: "Missing opening paren", input: "DEFINE PROCEDURE MissingParen a, b)", errSubstr: "missing '(' for parameter list"},
		{name: "Empty param name", input: "DEFINE PROCEDURE BadParams(a, , b)", expected: Procedure{Name: "BadParams", Params: []string{"a", "b"}}}, // Warning printed, empty param ignored by splitParams
		{name: "Single Param", input: "DEFINE PROCEDURE Single(only)", expected: Procedure{Name: "Single", Params: []string{"only"}}},
		{name: "Comment after header", input: "DEFINE PROCEDURE WithComment(x) # This is a comment", expected: Procedure{Name: "WithComment", Params: []string{"x"}}},
		{name: "Invalid char in name", input: "DEFINE PROCEDURE Bad-Name()", errSubstr: "invalid procedure name 'Bad-Name'"},
		{name: "Invalid char in param", input: "DEFINE PROCEDURE BadParam(good, bad-param)", errSubstr: "invalid parameter name 'bad-param'"},
		{name: "Trailing comma in params", input: "DEFINE PROCEDURE TrailingComma(a, )", expected: Procedure{Name: "TrailingComma", Params: []string{"a"}}},
		{name: "Keyword as name", input: "DEFINE PROCEDURE SET()", errSubstr: "invalid procedure name 'SET'"},
		{name: "Keyword as param", input: "DEFINE PROCEDURE UseSet(SET)", errSubstr: "invalid parameter name 'SET'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proc, err := parseProcedureHeader(tt.input)
			if tt.errSubstr != "" {
				checkErrorContains(t, err, tt.errSubstr)
				return
			}
			checkNoError(t, err)
			if proc.Name != tt.expected.Name {
				t.Errorf("Name mismatch:\nExpected: %q\nGot:      %q", tt.expected.Name, proc.Name)
			}
			if !reflect.DeepEqual(proc.Params, tt.expected.Params) {
				t.Errorf("Params mismatch:\nExpected: %v\nGot:      %v", tt.expected.Params, proc.Params)
			}
		})
	}
}

// === Test Docstring Parsing (Using parseDocstringBlock) === (Passed Previously)
func TestParseDocstringBlock(t *testing.T) {
	// ... tests unchanged ...
	tests := []struct {
		name        string
		inputLines  []string
		expectedDoc Docstring
		errSubstr   string
	}{
		{
			name: "Valid multi-line",
			inputLines: []string{
				"    PURPOSE: Test procedure", // Content starts on header line
				"      with multiple lines.",
				"    INPUTS:", // No content on header line
				"      - x (int): Input value",
				"            spread over lines.",    // Indented continuation
				"      - y: Another input, simple.", // Single line input
				"",                                  // Blank line between inputs
				"      - z (string): Third input.",
				"    OUTPUT: Result description",
				"          also multi-line.",
				"    ALGORITHM:",
				"      1. Step one indented.",
				"         2. Step two.",
				"    CAVEATS: None really.",
				"    EXAMPLES:",
				"      Example usage 1",
				"      Example usage 2",
			},
			expectedDoc: Docstring{
				Purpose:   "Test procedure\n      with multiple lines.",
				Inputs:    map[string]string{"x": "Input value\n            spread over lines.", "y": "Another input, simple.", "z": "Third input."},
				Output:    "Result description\n          also multi-line.",
				Algorithm: "1. Step one indented.\n         2. Step two.",
				Caveats:   "None really.",
				Examples:  "Example usage 1\n      Example usage 2",
			},
		},
		{name: "Sections out of order", inputLines: []string{"ALGORITHM: Steps.", "PURPOSE: Purpose.", "OUTPUT: Result.", "INPUTS: - a: A"}, expectedDoc: Docstring{Purpose: "Purpose.", Inputs: map[string]string{"a": "A"}, Output: "Result.", Algorithm: "Steps."}},
		{name: "Missing optional sections", inputLines: []string{"PURPOSE: P", "INPUTS: - i:I", "OUTPUT: O", "ALGORITHM: A"}, expectedDoc: Docstring{Purpose: "P", Inputs: map[string]string{"i": "I"}, Output: "O", Algorithm: "A"}},
		{name: "Malformed inputs - bad line", inputLines: []string{"PURPOSE: T", "INPUTS:", " Bad line.", " - x: Ok", "OUTPUT: O", "ALGORITHM: A"}, errSubstr: "malformed INPUTS content"},
		{name: "Content before first section", inputLines: []string{"Stray text.", "PURPOSE: P"}, errSubstr: "content before first section header"},
		{name: "Inputs None", inputLines: []string{"PURPOSE: P", "INPUTS: None", "OUTPUT: O", "ALGORITHM: A"}, expectedDoc: Docstring{Purpose: "P", Inputs: map[string]string{}, Output: "O", Algorithm: "A"}},
		{name: "Inputs none case insensitive", inputLines: []string{"PURPOSE: P", "INPUTS: none", "OUTPUT: O", "ALGORITHM: A"}, expectedDoc: Docstring{Purpose: "P", Inputs: map[string]string{}, Output: "O", Algorithm: "A"}},
		{name: "Inputs empty", inputLines: []string{"PURPOSE: P", "INPUTS:", "OUTPUT: O", "ALGORITHM: A"}, expectedDoc: Docstring{Purpose: "P", Inputs: map[string]string{}, Output: "O", Algorithm: "A"}},
		{name: "Duplicate section", inputLines: []string{"PURPOSE: P1", "PURPOSE: P2"}, errSubstr: "duplicate PURPOSE section"},
		{name: "Duplicate input param", inputLines: []string{"PURPOSE: P", "INPUTS:", "- x: Desc 1", "- x: Desc 2", "OUTPUT: O", "ALGORITHM: A"}, errSubstr: "duplicate input parameter name 'x'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parseDocstringBlock(tt.inputLines)
			if tt.errSubstr != "" {
				checkErrorContains(t, err, tt.errSubstr)
				return
			}
			checkNoError(t, err)
			if doc.Purpose != tt.expectedDoc.Purpose {
				t.Errorf("Purpose:\nExpected: %q\nGot:      %q", tt.expectedDoc.Purpose, doc.Purpose)
			}
			if !reflect.DeepEqual(doc.Inputs, tt.expectedDoc.Inputs) {
				t.Errorf("Inputs:\nExpected: %v\nGot:      %v", tt.expectedDoc.Inputs, doc.Inputs)
			}
			if doc.Output != tt.expectedDoc.Output {
				t.Errorf("Output:\nExpected: %q\nGot:      %q", tt.expectedDoc.Output, doc.Output)
			}
			if doc.Algorithm != tt.expectedDoc.Algorithm {
				t.Errorf("Algorithm:\nExpected: %q\nGot:      %q", tt.expectedDoc.Algorithm, doc.Algorithm)
			}
			if doc.Caveats != tt.expectedDoc.Caveats {
				t.Errorf("Caveats:\nExpected: %q\nGot:      %q", tt.expectedDoc.Caveats, doc.Caveats)
			}
			if doc.Examples != tt.expectedDoc.Examples {
				t.Errorf("Examples:\nExpected: %q\nGot:      %q", tt.expectedDoc.Examples, doc.Examples)
			}
		})
	}
}
