// filename: pkg/neurodata/models/models.go
package models

// Schema represents a basic data structure definition.
// This serves as a root type for different kinds of schemas you might define.
// Replace or extend this with the actual structures needed by your application.
type Schema struct {
	Name        string           `json:"name"`
	Version     string           `json:"version"`
	Description string           `json:"description,omitempty"`
	Fields      map[string]Field `json:"fields"`
	// Consider adding a 'Kind' field if you have multiple types of schemas (e.g., "data", "tool_input", "config")
	Kind string `json:"kind,omitempty"`
}

// Field represents a single field definition within a schema.
type Field struct {
	Type        string      `json:"type"` // e.g., "string", "number", "boolean", "list", "map", or a custom type name
	Required    bool        `json:"required,omitempty"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"` // Default value if applicable
	// Add constraints, nested schemas, enum values etc. if needed
	// Example: Constraints map[string]interface{} `json:"constraints,omitempty"`
}

// You can add more specific schema types here if needed, potentially embedding the base Schema
// type ToolInputSchema struct {
//    Schema
//    // Tool-specific fields
// }

// --- Loading/Parsing Logic (Optional) ---
// You might add functions here to load/parse schema definitions from files (e.g., JSON, YAML)
// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// )
//
// func LoadSchemaFromFile(filePath string) (*Schema, error) {
// 	data, err := os.ReadFile(filePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read schema file %s: %w", filePath, err)
// 	}
//
// 	var schema Schema
// 	err = json.Unmarshal(data, &schema)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal schema file %s: %w", filePath, err)
// 	}
//
// 	// TODO: Add validation logic for the loaded schema if needed
//
// 	return &schema, nil
// }
