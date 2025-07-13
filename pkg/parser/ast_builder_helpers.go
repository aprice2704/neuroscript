// filename: pkg/parser/ast_builder_helpers.go
// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Added the generic newNode helper to centralize AST node creation.

package parser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// newNode is a generic helper to create and initialize an AST node.
// It sets both the new BaseNode fields and the legacy .Pos field.
func newNode[T ast.Node](node T, token antlr.Token, kind types.Kind) T {
	pos := tokenToPosition(token)

	// Set the new BaseNode fields
	v := reflect.ValueOf(node).Elem()
	baseNodeField := v.FieldByName("BaseNode")
	if baseNodeField.IsValid() && baseNodeField.CanSet() {
		baseNode := baseNodeField.Addr().Interface().(*ast.BaseNode)
		baseNode.StartPos = &pos
		baseNode.NodeKind = kind
		// StopPos would be set here if we had the end token.
	}

	// Set the legacy .Pos field for backward compatibility
	posField := v.FieldByName("Pos")
	if posField.IsValid() && posField.CanSet() {
		posField.Set(reflect.ValueOf(&pos))
	}

	return node
}

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
func ConvertInputSchemaToArgSpec(schema map[string]interface{}) ([]tool.ArgSpec, error) {
	args := []tool.ArgSpec{}
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
		var argType tool.ArgType = tool.ArgTypeAny // Default to Any if unknown
		switch typeStr {
		case "string":
			argType = tool.ArgTypeString
		case "integer":
			argType = tool.ArgTypeInt
		case "number":
			argType = tool.ArgTypeFloat
		case "boolean":
			argType = tool.ArgTypeBool
		case "array":
			// TODO: Could inspect 'items' field for better type, defaults to SliceAny
			argType = tool.ArgTypeSliceAny
		case "object":
			argType = tool.ArgTypeMap
		}

		args = append(args, tool.ArgSpec{
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
func parseNumber(numStr string) (interface{}, error) {
	// Try parsing as int first
	if !strings.Contains(numStr, ".") { // Optimization: Don't try int if decimal present
		if iVal, err := strconv.ParseInt(numStr, 10, 64); err == nil {
			return iVal, nil
		}
	}

	// Try parsing as float
	if fVal, err := strconv.ParseFloat(numStr, 64); err == nil {
		return fVal, nil
	}

	return nil, fmt.Errorf("invalid number literal: %q", numStr)
}

// unescapeString handles standard Go escape sequences within single or double quotes.
func unescapeString(quotedStr string) (string, error) {
	unquoted, err := strconv.Unquote(quotedStr)
	if err != nil {
		return "", fmt.Errorf("invalid string literal %q: %w", quotedStr, err)
	}
	return unquoted, nil
}

// --- types.Position Helper ---

// tokenToPosition converts an ANTLR token to a types.Position.
func tokenToPosition(token antlr.Token) types.Position {
	if token == nil {
		return types.Position{Line: 0, Column: 0, File: "<nil token>"}
	}
	sourceName := "<unknown>"
	if token.GetInputStream() != nil {
		sourceName = token.GetInputStream().GetSourceName()
		if sourceName == "<INVALID>" {
			sourceName = "<input stream>"
		}
	}
	return types.Position{
		Line:   token.GetLine(),
		Column: token.GetColumn() + 1,
		File:   sourceName,
	}
}
