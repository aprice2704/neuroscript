// filename: pkg/core/llm_tools.go
package core

// Keep fmt
// Does not need other llm specific imports, only core types

// --- Function Declaration Generation ---

// GenerateToolDeclarations creates the GeminiTool list (containing FunctionDeclarations)
// for tools that are present in the allowlist.
// RENAMED: Exported by capitalizing the first letter.
func GenerateToolDeclarations(registry *ToolRegistry, allowlist []string) []GeminiTool {
	if registry == nil || len(allowlist) == 0 {
		return nil
	}

	allowlistMap := make(map[string]bool)
	for _, toolName := range allowlist {
		allowlistMap[toolName] = true
	}

	declarations := make([]GeminiFunctionDeclaration, 0)

	registeredTools := registry.tools
	if registeredTools == nil {
		// This path might need logging if the registry can realistically be nil here
		// logger.Printf("[WARN LLM Tools] Tool registry map is nil during declaration generation.")
		return nil
	}

	for toolName, impl := range registeredTools {
		if !allowlistMap[toolName] {
			continue
		}

		properties := make(map[string]GeminiParameterDetails)
		required := make([]string, 0)
		for _, argSpec := range impl.Spec.Args {
			paramType := "string"
			paramFormat := ""
			var itemsSchema *GeminiParameterDetails // For arrays

			switch argSpec.Type {
			case ArgTypeInt:
				paramType = "integer"
				paramFormat = "int64"
			case ArgTypeFloat:
				paramType = "number"
				paramFormat = "double"
			case ArgTypeBool:
				paramType = "boolean"
			case ArgTypeSliceString:
				paramType = "array"
				itemsSchema = &GeminiParameterDetails{Type: "string"}
			case ArgTypeSliceAny:
				paramType = "array"
				// Default items to string, but ideally, this could be more specific
				// if the 'any' slice typically holds a known type.
				itemsSchema = &GeminiParameterDetails{Type: "string"}
			case ArgTypeString:
				paramType = "string"
			case ArgTypeAny:
				// Represent 'any' as string for simplicity. Could also omit type or use a generic object type.
				paramType = "string"
			default:
				// Log unknown types? For now, default to string.
				paramType = "string"
			}

			propDetail := GeminiParameterDetails{
				Type:        paramType,
				Description: argSpec.Description,
			}
			if paramFormat != "" {
				propDetail.Format = paramFormat
			}
			if itemsSchema != nil {
				propDetail.Items = itemsSchema // Set Items for array type
			}
			properties[argSpec.Name] = propDetail // Add detail to properties map

			if argSpec.Required {
				required = append(required, argSpec.Name)
			}
		}

		var schema *GeminiParameterSchema
		if len(properties) > 0 {
			schema = &GeminiParameterSchema{
				Type:       "object",
				Properties: properties,
			}
			if len(required) > 0 {
				schema.Required = required
			}
		}

		declaration := GeminiFunctionDeclaration{
			Name:        impl.Spec.Name,
			Description: impl.Spec.Description,
			Parameters:  schema,
		}
		declarations = append(declarations, declaration)
	}

	if len(declarations) == 0 {
		return nil
	}

	return []GeminiTool{{FunctionDeclarations: declarations}}
}
