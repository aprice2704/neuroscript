// NeuroScript Version: 0.5.2
// File version: 15
// Purpose: Reverted to stricter, more robust regex for email and URL validation.
// filename: pkg/json-lite/shape.go
// nlines: 242
// risk_rating: LOW

package json_lite

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

const (
	maxShapeDepth = 64
)

var (
	// A robust, commonly used regex for email validation that correctly enforces TLDs.
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	// A flexible and robust regex for http/https URLs that requires a valid TLD.
	urlRegex = regexp.MustCompile(`^https?:\/\/(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{2,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&//=]*)?$`)
	// Regex for ISO 8601 datetime format.
	isoDatetimeRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$`)
)

type Shape struct {
	Fields map[string]*FieldSpec
}

type FieldSpec struct {
	Name          string
	IsOptional    bool
	IsList        bool
	PrimitiveType string
	NestedShape   *Shape
}

func ParseShape(rawShape map[string]any) (*Shape, error) {
	return parseShapeRecursive(rawShape, 0)
}

func parseShapeRecursive(rawShape map[string]any, depth int) (*Shape, error) {
	if depth > maxShapeDepth {
		return nil, fmt.Errorf("%w: shape definition exceeds maximum nesting depth of %d", lang.ErrNestingDepthExceeded, maxShapeDepth)
	}
	if rawShape == nil {
		return nil, fmt.Errorf("%w: raw shape definition cannot be nil", lang.ErrInvalidArgument)
	}
	s := &Shape{Fields: make(map[string]*FieldSpec, len(rawShape))}
	for key, typeDef := range rawShape {
		// Corrected, robust suffix parsing
		keyName := key
		var isOptional, isList bool
		// Iteratively strip suffixes to handle any order, e.g. '[]?' or '?[]'
		for hasSuffix := true; hasSuffix; {
			if strings.HasSuffix(keyName, "?") {
				keyName = keyName[:len(keyName)-1]
				isOptional = true
			} else if strings.HasSuffix(keyName, "[]") {
				keyName = keyName[:len(keyName)-2]
				isList = true
			} else {
				hasSuffix = false
			}
		}

		if keyName == "" {
			return nil, fmt.Errorf("%w: shape key '%s' is invalid because it has no name part", lang.ErrInvalidArgument, key)
		}

		spec := &FieldSpec{Name: keyName, IsOptional: isOptional, IsList: isList}

		switch td := typeDef.(type) {
		case string:
			spec.PrimitiveType = td
		case map[string]any:
			nestedShape, err := parseShapeRecursive(td, depth+1)
			if err != nil {
				return nil, fmt.Errorf("failed to parse nested shape for key '%s': %w", key, err)
			}
			spec.NestedShape = nestedShape
		default:
			return nil, fmt.Errorf("%w: invalid type definition for key '%s'", lang.ErrValidationTypeMismatch, key)
		}
		s.Fields[keyName] = spec // Use the normalized key name for storage
	}
	return s, nil
}

func (s *Shape) Validate(value any, allowExtra bool) error {
	valMap, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("%w: expected a map, but got %T", lang.ErrValidationTypeMismatch, value)
	}
	return s.validateMap(valMap, allowExtra, "", 0)
}

func (s *Shape) validateMap(valMap map[string]any, allowExtra bool, currentPath string, depth int) error {
	if depth > maxShapeDepth {
		return fmt.Errorf("%w: data structure exceeds maximum nesting depth of %d at path '%s'", lang.ErrNestingDepthExceeded, maxShapeDepth, currentPath)
	}

	validatedKeys := make(map[string]bool)

	for keyName, spec := range s.Fields {
		path := buildPath(currentPath, keyName, false)
		actualValue, exists := valMap[keyName]
		validatedKeys[keyName] = true

		if !exists {
			if !spec.IsOptional {
				return fmt.Errorf("%w: missing required key '%s' at path '%s'", lang.ErrValidationRequiredArgMissing, keyName, currentPath)
			}
			continue
		}

		// Depth increases when we descend into a nested structure
		if spec.IsList {
			if err := spec.validateList(actualValue, allowExtra, path, depth); err != nil {
				return err
			}
		} else {
			if err := spec.validateSingle(actualValue, allowExtra, path, depth); err != nil {
				return err
			}
		}
	}

	if !allowExtra {
		for key := range valMap {
			if !validatedKeys[key] {
				return fmt.Errorf("%w: unexpected key '%s' at path '%s'", lang.ErrInvalidArgument, key, buildPath(currentPath, key, false))
			}
		}
	}
	return nil
}

func (fs *FieldSpec) validateSingle(value any, allowExtra bool, path string, depth int) error {
	// Corrected: Handle nil values as a type mismatch unless the type is 'any'
	if value == nil {
		if fs.PrimitiveType == "any" {
			return nil
		}
		// A nil value for a present key is a type mismatch.
		return fmt.Errorf("%w: at path '%s', got nil value for required type '%s'", lang.ErrValidationTypeMismatch, path, fs.PrimitiveType)
	}

	if fs.NestedShape != nil {
		valMap, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%w: at path '%s', expected a map but got %T", lang.ErrValidationTypeMismatch, path, value)
		}
		return fs.NestedShape.validateMap(valMap, allowExtra, path, depth+1)
	}

	// Validate against the primitive type
	if err := validatePrimitive(value, fs.PrimitiveType, path); err != nil {
		return err
	}
	return nil
}

func (fs *FieldSpec) validateList(value any, allowExtra bool, path string, depth int) error {
	valList, ok := value.([]any)
	if !ok {
		return fmt.Errorf("%w: for key '%s', expected a list but got %T", lang.ErrValidationTypeMismatch, path, value)
	}
	for i, item := range valList {
		itemPath := buildPath(path, fmt.Sprintf("%d", i), true)
		// Pass the same depth for items in a list, but validateSingle will increment it if an item is a map
		if err := fs.validateSingle(item, allowExtra, itemPath, depth); err != nil {
			return err
		}
	}
	return nil
}

// --- Helpers ---
func getTypeName(value any) (string, bool) {
	switch value.(type) {
	case string:
		return "string", true
	case int, int8, int16, int32, int64:
		return "int", true
	case float32, float64:
		return "float", true
	case bool:
		return "bool", true
	default:
		return fmt.Sprintf("%T", value), false
	}
}

func validatePrimitive(value any, shapeType string, path string) error {
	if shapeType == "any" {
		return nil
	}

	typeName, _ := getTypeName(value)
	isSpecial, err := isSpecialType(value, shapeType, path)
	if err != nil {
		return err
	}
	if isSpecial {
		return nil
	}

	// Fallback to basic type check if not a special type
	if typeName != shapeType {
		return fmt.Errorf("%w: at path '%s', expected type '%s' but got '%s'", lang.ErrValidationTypeMismatch, path, shapeType, typeName)
	}
	return nil
}

func isSpecialType(value any, shapeType string, path string) (bool, error) {
	str, isString := value.(string)
	if !isString {
		// Special types must have an underlying string type
		if shapeType == "email" || shapeType == "url" || shapeType == "isoDatetime" {
			return false, fmt.Errorf("%w: at path '%s', expected a string for special type '%s' but got %T", lang.ErrValidationTypeMismatch, path, shapeType, value)
		}
		return false, nil
	}

	switch shapeType {
	case "email":
		if !emailRegex.MatchString(str) {
			return true, fmt.Errorf("%w: at path '%s', value '%s' is not a valid email format", lang.ErrValidationFailed, path, str)
		}
		return true, nil
	case "url":
		if !urlRegex.MatchString(str) {
			return true, fmt.Errorf("%w: at path '%s', value '%s' is not a valid URL format", lang.ErrValidationFailed, path, str)
		}
		return true, nil
	case "isoDatetime":
		if !isoDatetimeRegex.MatchString(str) {
			return true, fmt.Errorf("%w: at path '%s', value '%s' is not a valid ISO 8601 datetime format", lang.ErrValidationFailed, path, str)
		}
		return true, nil
	}
	return false, nil
}

func buildPath(base, addition string, isIndex bool) string {
	if base == "" {
		return addition
	}
	if isIndex {
		return fmt.Sprintf("%s[%s]", base, addition)
	}
	return base + "." + addition
}
