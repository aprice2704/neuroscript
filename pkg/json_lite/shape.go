// NeuroScript Version: 0.5.2
// File version: 17
// Purpose: Implements shape validation with options for case-insensitivity and float-to-int coercion.
// filename: pkg/json-lite/shape.go
// nlines: 302
// risk_rating: MEDIUM

package json_lite

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

const (
	maxShapeDepth = 64
)

var (
	emailRegex       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	urlRegex         = regexp.MustCompile(`^https?:\/\/(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{2,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&//=]*)?$`)
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

// ValidateOptions provides options for the Validate method.
type ValidateOptions struct {
	AllowExtra      bool
	CaseInsensitive bool
}

func ParseShape(rawShape map[string]any) (*Shape, error) {
	return parseShapeRecursive(rawShape, 0)
}

func parseShapeRecursive(rawShape map[string]any, depth int) (*Shape, error) {
	if depth > maxShapeDepth {
		return nil, fmt.Errorf("%w: shape definition exceeds maximum nesting depth of %d", ErrNestingDepthExceeded, maxShapeDepth)
	}
	if rawShape == nil {
		return nil, fmt.Errorf("%w: raw shape definition cannot be nil", ErrInvalidArgument)
	}
	s := &Shape{Fields: make(map[string]*FieldSpec, len(rawShape))}
	for key, typeDef := range rawShape {
		keyName := key
		var isOptional, isList bool
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
			return nil, fmt.Errorf("%w: shape key '%s' is invalid because it has no name part", ErrInvalidArgument, key)
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
			return nil, fmt.Errorf("%w: invalid type definition for key '%s'", ErrValidationTypeMismatch, key)
		}
		s.Fields[keyName] = spec
	}
	return s, nil
}

func (s *Shape) Validate(value any, options *ValidateOptions) error {
	valMap, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("%w: expected a map, but got %T", ErrValidationTypeMismatch, value)
	}
	if options == nil {
		options = &ValidateOptions{} // Use defaults if nil
	}
	return s.validateMap(valMap, options, "", 0)
}

func (s *Shape) validateMap(valMap map[string]any, options *ValidateOptions, currentPath string, depth int) error {
	if depth > maxShapeDepth {
		return fmt.Errorf("%w: data structure exceeds maximum nesting depth of %d at path '%s'", ErrNestingDepthExceeded, maxShapeDepth, currentPath)
	}

	validatedKeys := make(map[string]bool)

	for specKeyName, spec := range s.Fields {
		path := buildPath(currentPath, specKeyName, false)
		var actualValue any
		var exists bool
		var originalKey string

		if options.CaseInsensitive {
			lowerSpecKeyName := strings.ToLower(specKeyName)
			for k, v := range valMap {
				if strings.ToLower(k) == lowerSpecKeyName {
					actualValue = v
					exists = true
					originalKey = k
					break
				}
			}
		} else {
			actualValue, exists = valMap[specKeyName]
			originalKey = specKeyName
		}

		if exists {
			validatedKeys[originalKey] = true
		} else {
			if !spec.IsOptional {
				return fmt.Errorf("%w: missing required key '%s' at path '%s'", ErrValidationRequiredArgMissing, specKeyName, currentPath)
			}
			continue
		}

		if spec.IsList {
			if err := spec.validateList(actualValue, options, path, depth); err != nil {
				return err
			}
		} else {
			if err := spec.validateSingle(actualValue, options, path, depth); err != nil {
				return err
			}
		}
	}

	if !options.AllowExtra {
		for key := range valMap {
			if !validatedKeys[key] {
				return fmt.Errorf("%w: unexpected key '%s' at path '%s'", ErrInvalidArgument, key, buildPath(currentPath, key, false))
			}
		}
	}
	return nil
}

func (fs *FieldSpec) validateSingle(value any, options *ValidateOptions, path string, depth int) error {
	if value == nil {
		if fs.PrimitiveType == "any" {
			return nil
		}
		return fmt.Errorf("%w: at path '%s', got nil value for required type '%s'", ErrValidationTypeMismatch, path, fs.PrimitiveType)
	}

	if fs.NestedShape != nil {
		valMap, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%w: at path '%s', expected a map but got %T", ErrValidationTypeMismatch, path, value)
		}
		return fs.NestedShape.validateMap(valMap, options, path, depth+1)
	}

	if err := validatePrimitive(value, fs.PrimitiveType, path); err != nil {
		return err
	}
	return nil
}

func (fs *FieldSpec) validateList(value any, options *ValidateOptions, path string, depth int) error {
	valList, ok := value.([]any)
	if !ok {
		return fmt.Errorf("%w: for key '%s', expected a list but got %T", ErrValidationTypeMismatch, path, value)
	}
	for i, item := range valList {
		itemPath := buildPath(path, fmt.Sprintf("%d", i), true)
		if err := fs.validateSingle(item, options, itemPath, depth); err != nil {
			return err
		}
	}
	return nil
}

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

	// Handle special string types (email, url, etc.) first.
	isSpecial, err := isSpecialType(value, shapeType, path)
	if err != nil {
		return err // This was a format validation failure (e.g., bad email)
	}
	if isSpecial {
		return nil // This was a successful special type validation
	}

	typeName, _ := getTypeName(value)

	// Special case: If the shape expects an "int", also allow a "float"
	// value IF AND ONLY IF it represents a whole number. This handles
	// type mismatches from systems like NeuroScript that unwrap all
	// numbers as float64.
	if shapeType == "int" && typeName == "float" {
		var f64 float64
		var ok bool
		if f32, isF32 := value.(float32); isF32 {
			f64 = float64(f32)
			ok = true
		} else if f64, ok = value.(float64); ok {
			// f64 is already assigned
		}

		if ok {
			// Use math.Modf to check for a fractional part
			_, frac := math.Modf(f64)
			if frac == 0.0 {
				return nil // It's a whole number, so it's a valid "int"
			}
			// If frac != 0.0, it's a real float, fall through to the error
		}
	}

	if typeName != shapeType {
		return fmt.Errorf("%w: at path '%s', expected type '%s' but got '%s'", ErrValidationTypeMismatch, path, shapeType, typeName)
	}
	return nil
}

func isSpecialType(value any, shapeType string, path string) (bool, error) {
	str, isString := value.(string)
	if !isString {
		if shapeType == "email" || shapeType == "url" || shapeType == "isoDatetime" {
			return false, fmt.Errorf("%w: at path '%s', expected a string for special type '%s' but got %T", ErrValidationTypeMismatch, path, shapeType, value)
		}
		return false, nil
	}

	switch shapeType {
	case "email":
		if !emailRegex.MatchString(str) {
			return true, fmt.Errorf("%w: at path '%s', value '%s' is not a valid email format", ErrValidationFailed, path, str)
		}
		return true, nil
	case "url":
		if !urlRegex.MatchString(str) {
			return true, fmt.Errorf("%w: at path '%s', value '%s' is not a valid URL format", ErrValidationFailed, path, str)
		}
		return true, nil
	case "isoDatetime":
		if !isoDatetimeRegex.MatchString(str) {
			return true, fmt.Errorf("%w: at path '%s', value '%s' is not a valid ISO 8061 datetime format", ErrValidationFailed, path, str)
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
