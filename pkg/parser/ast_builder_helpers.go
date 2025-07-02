// filename: pkg/parser/ast_builder_helpers.go

package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/lang"
	// Import other necessary core types if needed by helpers
)

// ArgType defines the type of an argument.
type ArgType int

const (
	ArgTypeAny ArgType = iota
	ArgTypeString
	ArgTypeInt
	ArgTypeFloat
	ArgTypeBool
	ArgTypeSliceAny
	ArgTypeMap
)

// ArgSpec defines the specification for a tool or function argument.
type ArgSpec struct {
	Name        string
	Type        ArgType
	Description string
	Required    bool
}

// ParseMetadataLine attempts to parse a line potentially containing metadata (e.g., ":: key: value").
// It returns the extracted key, value, and a boolean indicating if the line was a valid metadata line.
// Key and value are trimmed of whitespace.
func ParseMetadataLine(line string) (key string, value string, ok bool) {
	trimmedLine := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmedLine, "::") {
		return "", "", false // Not a metadata line
	}

	// Remove "::" prefix and trim surrounding space
	content := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "::"))

	// Find the first colon
	colonIndex := strings.Index(content, ":")
	if colonIndex == -1 {
		// Treat as a key-only metadata line (value is empty)
		key = strings.TrimSpace(content)
		value = ""
		// Basic validation: key cannot be empty
		if key == "" {
			return "", "", false
		}
		return key, value, true
		// Alternatively, consider this invalid: return "", "", false
	}

	// Extract key and value based on the first colon
	key = strings.TrimSpace(content[:colonIndex])
	value = strings.TrimSpace(content[colonIndex+1:])

	// Basic validation: key cannot be empty
	if key == "" {
		return "", "", false
	}

	return key, value, true
}

// --- Added Helper Function for Schema Conversion ---

// ConvertInputSchemaToArgSpec converts a JSON Schema-like map (from old ToolDefinition)
// into the []ArgSpec required by ToolSpec.
// Moved here from interpreter_steps_ask.go/llm_tools.go corrections.
func ConvertInputSchemaToArgSpec(schema map[string]interface{}) ([]ArgSpec, error) {
	args := []ArgSpec{}
	propsVal, okProps := schema["properties"]
	if !okProps {
		// If no properties, return empty args (valid for tools with no args)
		return args, nil
	}
	props, okPropsMap := propsVal.(map[string]interface{})
	if !okPropsMap {
		return nil, fmt.Errorf("invalid schema: 'properties' field is not a map[string]interface{}")
	}

	// Handle 'required' field - it might be missing or nil
	required := []string{}
	reqVal, okReq := schema["required"]
	if okReq {
		reqSlice, okReqSlice := reqVal.([]string)
		if !okReqSlice {
			// Check if it's []interface{} and try converting
			reqIntSlice, okReqIntSlice := reqVal.([]interface{})
			if okReqIntSlice {
				required = make([]string, 0, len(reqIntSlice))
				for i, item := range reqIntSlice {
					if strItem, okStr := item.(string); okStr {
						required = append(required, strItem)
					} else {
						return nil, fmt.Errorf("invalid schema: 'required' array element %d is not a string (%T)", i, item)
					}
				}
			} else {
				return nil, fmt.Errorf("invalid schema: 'required' field is not []string or []interface{} of strings")
			}
		} else {
			required = reqSlice
		}
	}
	// Build a map for quick lookup of required args
	reqMap := make(map[string]bool)
	for _, r := range required {
		reqMap[r] = true
	}

	for name, propSchemaIntf := range props {
		propSchema, okSchema := propSchemaIntf.(map[string]interface{})
		if !okSchema {
			return nil, fmt.Errorf("invalid schema: property '%s' is not a map[string]interface{}", name)
		}

		typeStrVal, _ := propSchema["type"]
		typeStr, _ := typeStrVal.(string) // JSON schema type
		descStrVal, _ := propSchema["description"]
		descStr, _ := descStrVal.(string) // Description

		// Convert JSON schema type to internal ArgType
		var argType ArgType = ArgTypeAny // Default to Any if unknown
		switch typeStr {
		case "string":
			argType = ArgTypeString
		case "integer":
			argType = ArgTypeInt
		case "number":
			argType = ArgTypeFloat
		case "boolean":
			argType = ArgTypeBool
		case "array":
			// TODO: Could inspect 'items' field for better type, defaults to SliceAny
			argType = ArgTypeSliceAny
		case "object":
			argType = ArgTypeMap
			// default: // Keep default as ArgTypeAny
		}

		args = append(args, ArgSpec{
			Name:        name,
			Type:        argType,
			Description: descStr,
			Required:    reqMap[name], // Check if name was in the required list
		})
	}
	return args, nil
}

// --- Added Helper Functions for Literal Parsing ---

// parseNumber attempts to parse a string as int64 or float64.
// Moved here from ast_builder_terminators.go correction.
func parseNumber(numStr string) (interface{}, error) {
	// Try parsing as int first
	if !strings.Contains(numStr, ".") { // Optimization: Don't try int if decimal present
		if iVal, err := strconv.ParseInt(numStr, 10, 64); err == nil {
			return iVal, nil
		}
		// Int parsing failed, fall through to try float
	}

	// Try parsing as float
	if fVal, err := strconv.ParseFloat(numStr, 64); err == nil {
		return fVal, nil
	}

	// Both failed
	return nil, fmt.Errorf("invalid number literal: %q", numStr)
}

// unescapeString handles standard Go escape sequences within single or double quotes.
// Moved here from ast_builder_terminators.go correction.
func unescapeString(quotedStr string) (string, error) {
	// strconv.Unquote handles both ' and " delimited strings and standard escapes
	unquoted, err := strconv.Unquote(quotedStr)
	if err != nil {
		return "", fmt.Errorf("invalid string literal %q: %w", quotedStr, err)
	}
	return unquoted, nil
}

// --- lang.Position Helper ---

// tokenToPosition converts an ANTLR token to a lang.Position.
// It sets the exported fields Line, Column, and File.
func tokenToPosition(token antlr.Token) lang.Position {
	if token == nil {
		return lang.Position{Line: 0, Column: 0, File: "<nil token>"} // Return a default invalid lang.Position
	}
	// Handle potential nil InputStream or SourceName gracefully
	sourceName := "<unknown>"
	if token.GetInputStream() != nil {
		sourceName = token.GetInputStream().GetSourceName()
		if sourceName == "<INVALID>" { // Use a more descriptive name if ANTLR provides one
			sourceName = "<input stream>"
		}
	}
	return lang.Position{
		Line:   token.GetLine(),
		Column: token.GetColumn() + 1, // ANTLR columns are 0-based, prefer 1-based
		File:   sourceName,
		// Length: len(token.GetText()), // Add if needed by lang.Position struct consumers
	}
}
