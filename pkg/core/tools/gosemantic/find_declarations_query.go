// NeuroScript Version: 0.3.1
// File version: 0.0.9 // Fix receiver constraint logic in lookupMethodWithReceiver.
// Initial definition for GoGetDeclarationOfSymbol tool.
// filename: pkg/core/tools/gosemantic/find_declarations_query.go

package gosemantic

import (
	"errors"
	"fmt"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"    // Import core
	"github.com/aprice2704/neuroscript/pkg/logging" // Import logging
	// Import packages
)

// +++ Tool Definition: GoGetDeclarationOfSymbol +++
var toolGoGetDeclarationOfSymbolImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name: "GoGetDeclarationOfSymbol",
		Description: "Finds the declaration location of a Go symbol using a semantic query string within an indexed codebase.\n" +
			"The query should be a semicolon-separated string of key:value pairs identifying the symbol.\n" +
			"Required Key: 'package' (e.g., 'package:github.com/example/pkg').\n" +
			"Symbol Keys: 'type', 'interface', 'method', 'function', 'var', 'const'. Use only one.\n" +
			"Context Keys (optional): 'receiver' (for methods, e.g., 'receiver:MyStruct' or 'receiver:*MyStruct'), 'field' (for fields within structs, same as using 'var' within 'type').\n" +
			"Examples:\n" +
			"  'package:github.com/example/pkg; function:ProcessData'\n" +
			"  'package:github.com/example/pkg; type:MyStruct'\n" +
			"  'package:github.com/example/pkg; type:MyStruct; method:DoThing'\n" +
			"  'package:github.com/example/pkg; type:MyStruct; field:counter' (or 'var:counter')\n" +
			"  'package:github.com/example/pkg; var:globalVar'",
		Args: []core.ArgSpec{
			{Name: "index_handle", Type: core.ArgTypeString, Required: true, Description: "Handle returned by GoIndexCode."},
			{Name: "query", Type: core.ArgTypeString, Required: true, Description: "Semantic query string identifying the symbol."},
		},
		ReturnType: core.ArgTypeMap,
	},
	Func: toolGoGetDeclarationOfSymbol,
}

func toolGoGetDeclarationOfSymbol(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	logger := interpreter.Logger()

	if len(args) != 2 {
		return nil, fmt.Errorf("%w: GoGetDeclarationOfSymbol requires 2 arguments (index_handle, query)", core.ErrInvalidArgument)
	}
	handle, okH := args[0].(string)
	query, okQ := args[1].(string)
	if !okH || !okQ {
		return nil, fmt.Errorf("%w: GoGetDeclarationOfSymbol invalid argument types (expected string, string)", core.ErrInvalidArgument)
	}
	if query == "" {
		return nil, fmt.Errorf("%w: GoGetDeclarationOfSymbol query cannot be empty", core.ErrInvalidArgument)
	}
	logger.Debug("[TOOL-GOGETDECLSYM] Request", "handle", handle, "query", query)
	indexValue, err := interpreter.GetHandleValue(handle, semanticIndexTypeTag)
	if err != nil {
		logger.Error("[TOOL-GOGETDECLSYM] Failed get handle", "handle", handle, "error", err)
		return nil, err
	}
	index, ok := indexValue.(*SemanticIndex)
	if !ok {
		logger.Error("[TOOL-GOGETDECLSYM] Handle not *SemanticIndex", "handle", handle, "type", fmt.Sprintf("%T", indexValue))
		return nil, fmt.Errorf("%w: handle '%s' is not a SemanticIndex", core.ErrHandleWrongType, handle)
	}
	if index.Fset == nil || index.Packages == nil {
		logger.Error("[TOOL-GOGETDECLSYM] Index FileSet or Packages nil", "handle", handle)
		return nil, fmt.Errorf("%w: index '%s' is incomplete", ErrIndexNotReady, handle)
	}

	// Parse Query
	parsedQuery, parseErr := parseSemanticQuery(query)
	if parseErr != nil {
		logger.Error("[TOOL-GOGETDECLSYM] Failed to parse query", "query", query, "error", parseErr)
		return nil, fmt.Errorf("%w: %w", core.ErrInvalidArgument, parseErr)
	}
	logger.Debug("[TOOL-GOGETDECLSYM] Parsed query", "parsed", parsedQuery)

	// Find Symbol Object - passing logger down
	foundObj, findErr := findObjectInIndex(index, parsedQuery, logger)
	if findErr != nil {
		if errors.Is(findErr, ErrSymbolNotFound) || errors.Is(findErr, ErrPackageNotFound) || errors.Is(findErr, ErrWrongKind) {
			logger.Warn("[TOOL-GOGETDECLSYM] Symbol lookup failed or not found", "query", query, "error", findErr)
			return nil, nil // Return nil for "not found" or "wrong kind" errors as per tool spec convention
		}
		logger.Error("[TOOL-GOGETDECLSYM] Error during symbol lookup", "query", query, "error", findErr)
		return nil, findErr // Return other internal errors
	}
	if foundObj == nil {
		// This case should ideally not happen if findErr is nil, but guard anyway
		logger.Error("[TOOL-GOGETDECLSYM] Internal inconsistency: findObjectInIndex returned nil object without error", "query", query)
		return nil, fmt.Errorf("%w: findObjectInIndex returned nil object unexpectedly for query '%s'", core.ErrInternal, query)
	}
	logger.Debug("[TOOL-GOGETDECLSYM] Found object", "query", query, "objName", foundObj.Name(), "objType", fmt.Sprintf("%T", foundObj), "objPos", index.Fset.Position(foundObj.Pos()))

	// Check Object Type & Filter
	if _, isPkgName := foundObj.(*types.PkgName); isPkgName {
		logger.Info("[TOOL-GOGETDECLSYM] Query resolved to PkgName, ignoring.", "query", query, "pkgPath", foundObj.(*types.PkgName).Imported().Path())
		return nil, nil
	}
	declPos := foundObj.Pos()
	kind := getObjectKind(foundObj)
	logger.Debug("[TOOL-GOGETDECLSYM] Using obj.Pos() for declaration", "kind", kind, "pos", declPos, "objReported", index.Fset.Position(foundObj.Pos()))
	if !declPos.IsValid() {
		logger.Warn("[TOOL-GOGETDECLSYM] Declaration position invalid", "object", foundObj.Name(), "query", query)
		return nil, nil
	}
	declPosition := index.Fset.Position(declPos)
	declFilenameAbs := declPosition.Filename
	if declFilenameAbs == "" {
		logger.Warn("[TOOL-GOGETDECLSYM] Declaration has no filename", "object", foundObj.Name(), "query", query)
		return nil, nil
	}

	// Path filtering logic (relative to index.LoadDir)
	logger.Debug("[TOOL-GOGETDECLSYM] Path Filter Check", "declPathAbs", declFilenameAbs, "indexLoadDir", index.LoadDir)
	relDeclPathCheck, errCheck := filepath.Rel(index.LoadDir, declFilenameAbs)
	cleanLoadDir := filepath.Clean(index.LoadDir)
	cleanDeclFilenameAbs := filepath.Clean(declFilenameAbs)
	isOutside := errCheck != nil || (!strings.HasPrefix(cleanDeclFilenameAbs, cleanLoadDir+string(filepath.Separator)) && cleanDeclFilenameAbs != cleanLoadDir)
	logger.Debug("[TOOL-GOGETDECLSYM] Path Filter Result", "isOutside", isOutside, "relErr", errCheck, "cleanDeclPath", cleanDeclFilenameAbs, "cleanLoadDir", cleanLoadDir)
	if isOutside {
		logger.Info("[TOOL-GOGETDECLSYM] Declaration outside indexed dir, filtering.", "object", foundObj.Name(), "query", query, "decl_path", declFilenameAbs, "load_dir", cleanLoadDir)
		return nil, nil // Outside indexed scope
	}

	// Return Result
	relDeclPath := filepath.ToSlash(relDeclPathCheck)
	name := foundObj.Name()
	logger.Debug("[TOOL-GOGETDECLSYM] Found declaration", "query", query, "name", name, "kind", kind, "path", relDeclPath, "L", declPosition.Line, "C", declPosition.Column)
	return map[string]interface{}{
		"path":   relDeclPath,
		"line":   int64(declPosition.Line),
		"column": int64(declPosition.Column),
		"name":   name,
		"kind":   kind,
	}, nil
}

// --- Helper Functions ---

// parseSemanticQuery parses the query string into a map.
func parseSemanticQuery(query string) (map[string]string, error) {
	parsed := make(map[string]string)
	parts := strings.Split(query, ";")
	if len(parts) == 0 || (len(parts) == 1 && strings.TrimSpace(parts[0]) == "") {
		return nil, fmt.Errorf("%w: query string is empty or contains no parts", ErrInvalidQueryFormat)
	}
	knownKeys := map[string]bool{"package": true, "type": true, "interface": true, "method": true, "function": true, "var": true, "const": true, "receiver": true, "field": true}
	symbolKeyCount := 0
	symbolKeys := []string{"type", "interface", "method", "function", "var", "const", "field"}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("%w: invalid part format (missing ':'): '%s'", ErrInvalidQueryFormat, part)
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if key == "" || value == "" {
			return nil, fmt.Errorf("%w: invalid part format (empty key or value): '%s'", ErrInvalidQueryFormat, part)
		}
		if !knownKeys[key] {
			return nil, fmt.Errorf("%w: unknown key '%s' in query part '%s'", ErrInvalidQueryFormat, key, part)
		}
		if _, exists := parsed[key]; exists {
			return nil, fmt.Errorf("%w: duplicate key '%s' found in query", ErrInvalidQueryFormat, key)
		}
		isSymbolKey := false
		for _, sk := range symbolKeys {
			if key == sk {
				isSymbolKey = true
				break
			}
		}
		if isSymbolKey {
			symbolKeyCount++
		}
		// Allow 'field' as an alias for 'var' when 'type' is present
		if key == "field" {
			key = "var" // Treat field as var internally for lookup
			if _, varExists := parsed["var"]; varExists {
				return nil, fmt.Errorf("%w: duplicate key 'var' found (via 'field' alias) in query", ErrInvalidQueryFormat)
			}
		}
		parsed[key] = value
	}
	if len(parsed) == 0 {
		return nil, fmt.Errorf("%w: query string resulted in no valid key-value pairs", ErrInvalidQueryFormat)
	}
	if _, ok := parsed["package"]; !ok {
		return nil, fmt.Errorf("%w: query must contain a 'package:' key", ErrInvalidQueryFormat)
	}
	if symbolKeyCount == 0 {
		return nil, fmt.Errorf("%w: query must contain exactly one symbol key (e.g., 'type', 'function', 'var', etc.)", ErrInvalidQueryFormat)
	}
	// Allow 'var'/'field' or 'method' together with 'type'/'interface'
	if symbolKeyCount > 1 &&
		!((symbolKeyCount == 2) &&
			((parsed["type"] != "" || parsed["interface"] != "") && (parsed["method"] != "" || parsed["var"] != ""))) {
		return nil, fmt.Errorf("%w: query contains conflicting symbol keys (%d found, check valid combinations like type+method or type+var)", ErrInvalidQueryFormat, symbolKeyCount)
	}
	if parsed["method"] != "" && parsed["type"] == "" && parsed["interface"] == "" {
		return nil, fmt.Errorf("%w: 'method:' key requires a 'type:' or 'interface:' key", ErrInvalidQueryFormat)
	}
	// Allow 'var' (field alias) only with 'type' or 'interface'
	if parsed["var"] != "" && parsed["type"] == "" && parsed["interface"] == "" && symbolKeyCount > 1 {
		return nil, fmt.Errorf("%w: 'var:' (or 'field:') key used as a qualifier requires a 'type:' or 'interface:' key", ErrInvalidQueryFormat)
	}
	if parsed["receiver"] != "" && parsed["method"] == "" {
		return nil, fmt.Errorf("%w: 'receiver:' key is only valid with 'method:' key", ErrInvalidQueryFormat)
	}
	return parsed, nil
}

// findObjectInIndex performs the actual lookup in the SemanticIndex.
func findObjectInIndex(index *SemanticIndex, pq map[string]string, logger logging.Logger) (types.Object, error) {
	logger.Debug("[FINDOBJ] Starting lookup", "query", pq)
	pkgPath := pq["package"]
	var targetPkgInfo *PackageInfo
	for _, pkgInfo := range index.Packages {
		if pkgInfo == nil || pkgInfo.Types == nil {
			continue
		}
		logger.Debug("[FINDOBJ] Checking package", "queryPath", pkgPath, "pkgInfo.PkgPath", pkgInfo.PkgPath, "pkgInfo.ID", pkgInfo.ID, "pkgInfo.Types.Path", pkgInfo.Types.Path())
		// Match primarily on ID (module path) or PkgPath (import path)
		if pkgInfo.ID == pkgPath || pkgInfo.PkgPath == pkgPath {
			logger.Debug("[FINDOBJ] Matched package", "matchedOn", "ID or PkgPath", "pkgID", pkgInfo.ID)
			targetPkgInfo = pkgInfo
			break
		}
		// Fallback check on types.Package path (can sometimes differ)
		if pkgInfo.Types.Path() == pkgPath {
			logger.Debug("[FINDOBJ] Matched package on Types.Path()", "pkgID", pkgInfo.ID)
			targetPkgInfo = pkgInfo
			break
		}
	}

	if targetPkgInfo == nil {
		logger.Warn("[FINDOBJ] Package not found in index", "queryPath", pkgPath)
		return nil, fmt.Errorf("%w: package '%s'", ErrPackageNotFound, pkgPath)
	}
	if targetPkgInfo.Types == nil || targetPkgInfo.Types.Scope() == nil {
		logger.Warn("[FINDOBJ] Package found but scope is nil", "queryPath", pkgPath, "pkgID", targetPkgInfo.ID)
		// This often indicates the package had load errors (e.g., compilation errors)
		return nil, fmt.Errorf("%w: package '%s' scope is nil (check index load errors)", ErrPackageNotFound, pkgPath)
	}
	logger.Debug("[FINDOBJ] Found target package", "pkgID", targetPkgInfo.ID, "pkgPath", targetPkgInfo.PkgPath)

	scope := targetPkgInfo.Types.Scope()
	targetTypesPkg := targetPkgInfo.Types
	var obj types.Object
	var err error

	// --- Determine Search Target ---
	if name, ok := pq["function"]; ok {
		logger.Debug("[FINDOBJ] Looking up function in package scope", "name", name)
		obj = scope.Lookup(name)
		if obj == nil {
			err = fmt.Errorf("%w: function '%s'", ErrSymbolNotFound, name)
		} else {
			logger.Debug("[FINDOBJ] Found obj via scope.Lookup", "name", name, "type", fmt.Sprintf("%T", obj))
			if !isKind(obj, types.Func{}) {
				err = fmt.Errorf("%w: '%s' is %T, not a function", ErrWrongKind, name, obj)
			}
		}
	} else if name, ok := pq["var"]; ok {
		// Handle field lookup (var + type) vs global var lookup
		if typeName, typeOk := pq["type"]; typeOk { // Could also be interface{} but less common for fields
			logger.Debug("[FINDOBJ] Looking up field in type", "typeName", typeName, "fieldName", name)
			typeObj := scope.Lookup(typeName)
			if typeObj == nil {
				err = fmt.Errorf("%w: type '%s' for field lookup", ErrSymbolNotFound, typeName)
			} else if !isKind(typeObj, types.TypeName{}) {
				err = fmt.Errorf("%w: '%s' is %T, not a type name", ErrWrongKind, typeName, typeObj)
			} else {
				obj, err = findFieldInType(typeObj, name, logger) // Pass logger
			}
		} else if ifaceName, ifaceOk := pq["interface"]; ifaceOk {
			// Technically interfaces don't have fields, but maybe query intended type?
			// Let's treat this as an error for now unless a use case emerges.
			err = fmt.Errorf("%w: interfaces ('%s') do not have fields ('%s')", ErrInvalidQueryFormat, ifaceName, name)
		} else {
			// Assume global variable lookup
			logger.Debug("[FINDOBJ] Looking up top-level var in package scope", "name", name)
			obj = scope.Lookup(name)
			if obj == nil {
				err = fmt.Errorf("%w: variable '%s'", ErrSymbolNotFound, name)
			} else {
				logger.Debug("[FINDOBJ] Found obj via scope.Lookup", "name", name, "type", fmt.Sprintf("%T", obj))
				if !isKind(obj, types.Var{}) {
					err = fmt.Errorf("%w: '%s' is %T, not a variable", ErrWrongKind, name, obj)
				} else if obj.(*types.Var).IsField() {
					// Found a field when looking for a global var
					err = fmt.Errorf("%w: '%s' is a field, not a global variable", ErrWrongKind, name)
				}
			}
		}
	} else if name, ok := pq["const"]; ok {
		logger.Debug("[FINDOBJ] Looking up const in package scope", "name", name)
		obj = scope.Lookup(name)
		if obj == nil {
			err = fmt.Errorf("%w: constant '%s'", ErrSymbolNotFound, name)
		} else {
			logger.Debug("[FINDOBJ] Found obj via scope.Lookup", "name", name, "type", fmt.Sprintf("%T", obj))
			if !isKind(obj, types.Const{}) {
				err = fmt.Errorf("%w: '%s' is %T, not a constant", ErrWrongKind, name, obj)
			}
		}
	} else if name, ok := pq["type"]; ok { // Includes named types like structs, interfaces, aliases
		logger.Debug("[FINDOBJ] Looking up type/interface in package scope", "name", name)
		obj = scope.Lookup(name)
		if obj == nil {
			err = fmt.Errorf("%w: type/interface '%s'", ErrSymbolNotFound, name)
		} else {
			logger.Debug("[FINDOBJ] Found obj via scope.Lookup", "name", name, "type", fmt.Sprintf("%T", obj))
			if !isKind(obj, types.TypeName{}) {
				err = fmt.Errorf("%w: '%s' is %T, not a type name", ErrWrongKind, name, obj)
			}
		}
		if err == nil {
			if methodName, methodOk := pq["method"]; methodOk {
				receiverName := pq["receiver"] // Optional constraint
				logger.Debug("[FINDOBJ] Looking up method on type/interface", "type", name, "method", methodName, "receiverConstraint", receiverName)
				obj, err = findMethodOnType(targetTypesPkg, obj, methodName, receiverName, logger) // Pass logger
			}
			// Field lookup is handled under the 'var' key check above
		}
	} else if name, ok := pq["interface"]; ok { // Specific key for interfaces, acts like 'type'
		logger.Debug("[FINDOBJ] Looking up interface specifically", "name", name)
		obj = scope.Lookup(name)
		if obj == nil {
			err = fmt.Errorf("%w: interface '%s'", ErrSymbolNotFound, name)
		} else {
			logger.Debug("[FINDOBJ] Found obj via scope.Lookup", "name", name, "type", fmt.Sprintf("%T", obj))
			if tn, ok := obj.(*types.TypeName); !ok || !types.IsInterface(tn.Type()) {
				err = fmt.Errorf("%w: '%s' is %T, not an interface type name", ErrWrongKind, name, obj)
			}
		}
		if err == nil {
			if methodName, methodOk := pq["method"]; methodOk {
				receiverName := pq["receiver"] // Optional constraint
				logger.Debug("[FINDOBJ] Looking up method on interface", "interface", name, "method", methodName, "receiverConstraint", receiverName)
				obj, err = findMethodOnType(targetTypesPkg, obj, methodName, receiverName, logger) // Pass logger
			}
			// Field lookup is handled under the 'var' key check above
		}
	} else {
		// This case should be caught by parseSemanticQuery, but acts as a fallback
		err = fmt.Errorf("%w: no valid symbol key found in query", ErrInvalidQueryFormat)
	}

	// --- Error Handling & Return ---
	if err != nil {
		logger.Warn("[FINDOBJ] Lookup failed", "query", pq, "error", err)
		// Wrap error with package context for clarity
		return nil, fmt.Errorf("in package '%s': %w", pkgPath, err)
	}
	if obj == nil {
		// This should not happen if err is nil, indicates internal logic error
		logger.Error("[FINDOBJ] Internal error: object nil without error", "query", pq)
		return nil, fmt.Errorf("%w: internal error: object nil without error for query '%v'", core.ErrInternal, pq)
	}

	logger.Debug("[FINDOBJ] Lookup successful", "query", pq, "foundObjName", obj.Name(), "foundObjType", fmt.Sprintf("%T", obj))
	return obj, nil
}

// isKind checks if an object is of a certain underlying kind.
func isKind(obj types.Object, kind interface{}) bool {
	if obj == nil {
		return false
	}
	switch kind.(type) {
	case types.Func:
		_, ok := obj.(*types.Func)
		return ok
	case types.Var:
		_, ok := obj.(*types.Var)
		return ok
	case types.Const:
		_, ok := obj.(*types.Const)
		return ok
	case types.TypeName:
		_, ok := obj.(*types.TypeName)
		return ok
	default:
		return false
	}
}

// findFieldInType looks for a field within a given type object.
func findFieldInType(typeObj types.Object, fieldName string, logger logging.Logger) (types.Object, error) {
	tn, ok := typeObj.(*types.TypeName)
	if !ok {
		return nil, fmt.Errorf("%w: expected type object for field lookup, got %T", ErrWrongKind, typeObj)
	}
	logger.Debug("[FINDFIELD] Looking for field", "type", tn.Name(), "field", fieldName)

	// We need the underlying type, handling pointers automatically
	currentType := tn.Type()
	if ptr, ok := currentType.(*types.Pointer); ok {
		logger.Debug("[FINDFIELD] Type is pointer, getting element", "type", tn.Name())
		currentType = ptr.Elem()
	}

	// Check if the underlying type is a struct
	if structType, ok := currentType.Underlying().(*types.Struct); ok {
		logger.Debug("[FINDFIELD] Type is struct, searching fields", "type", tn.Name(), "numFields", structType.NumFields())
		for i := 0; i < structType.NumFields(); i++ {
			field := structType.Field(i)
			logger.Debug("[FINDFIELD] Checking field", "index", i, "fieldName", field.Name())
			if field.Name() == fieldName {
				logger.Debug("[FINDFIELD] Found field", "type", tn.Name(), "field", fieldName)
				return field, nil // Found the field
			}
		}
		// Field not found in the struct
		logger.Warn("[FINDFIELD] Field not found in struct", "type", tn.Name(), "field", fieldName)
		return nil, fmt.Errorf("%w: field '%s' not found in struct '%s'", ErrSymbolNotFound, fieldName, tn.Name())
	}

	// The type is not a struct or a pointer to a struct
	logger.Warn("[FINDFIELD] Type is not a struct", "type", tn.Name(), "underlying", fmt.Sprintf("%T", currentType.Underlying()))
	return nil, fmt.Errorf("%w: type '%s' (%T) is not a struct or pointer-to-struct, cannot lookup field '%s'", ErrWrongKind, tn.Name(), tn.Type(), fieldName)
}

// findMethodOnType looks for a method on a given type object (struct or interface).
func findMethodOnType(pkg *types.Package, typeObj types.Object, methodName string, receiverConstraint string, logger logging.Logger) (types.Object, error) {
	tn, ok := typeObj.(*types.TypeName)
	if !ok {
		return nil, fmt.Errorf("%w: expected type object for method lookup, got %T", ErrWrongKind, typeObj)
	}

	baseType := tn.Type()
	logger.Debug("[FINDMETHOD] Looking for method", "typeName", tn.Name(), "methodName", methodName, "receiverConstraint", receiverConstraint, "isInterface", types.IsInterface(baseType))

	// Handle interfaces separately as MethodSet works directly
	if types.IsInterface(baseType) {
		mset := types.NewMethodSet(baseType)
		sel := mset.Lookup(pkg, methodName)
		if sel != nil && sel.Kind() == types.MethodVal {
			// Interface methods don't have concrete receiver types in the same way,
			// so the receiverConstraint might be less meaningful or require different interpretation.
			// For now, ignore receiverConstraint for interface methods.
			logger.Debug("[FINDMETHOD] Found method on interface", "ifaceName", tn.Name(), "methodName", methodName)
			return sel.Obj(), nil
		}
		logger.Warn("[FINDMETHOD] Method not found on interface", "ifaceName", tn.Name(), "methodName", methodName)
		return nil, fmt.Errorf("%w: method '%s' not found in interface '%s'", ErrSymbolNotFound, methodName, tn.Name())
	}

	// Handle concrete types (structs, named basic types)
	// We need to check both value and pointer receivers if no constraint is given,
	// or the specific receiver type if a constraint IS given.

	var foundMethodObj types.Object

	if receiverConstraint == "" {
		// No constraint: Check value receiver first, then pointer receiver
		logger.Debug("[FINDMETHOD] No constraint, checking value receiver first")
		foundMethodObj = lookupMethodWithReceiver(pkg, baseType, methodName, "", logger) // Empty constraint means don't check
		if foundMethodObj == nil {
			logger.Debug("[FINDMETHOD] Not found on value receiver, checking pointer receiver")
			foundMethodObj = lookupMethodWithReceiver(pkg, types.NewPointer(baseType), methodName, "", logger)
		}
	} else {
		// Constraint provided: Check only the specified receiver type
		var targetReceiverType types.Type
		if strings.HasPrefix(receiverConstraint, "*") {
			targetReceiverType = types.NewPointer(baseType)
		} else {
			targetReceiverType = baseType
		}
		logger.Debug("[FINDMETHOD] Constraint provided, checking specific receiver", "targetType", types.TypeString(targetReceiverType, nil))
		foundMethodObj = lookupMethodWithReceiver(pkg, targetReceiverType, methodName, receiverConstraint, logger)
	}

	// Check result
	if foundMethodObj != nil {
		logger.Debug("[FINDMETHOD] Found method successfully", "typeName", tn.Name(), "methodName", methodName)
		return foundMethodObj, nil
	}

	// Not found
	if receiverConstraint != "" {
		logger.Warn("[FINDMETHOD] Method with specific receiver constraint not found", "typeName", tn.Name(), "methodName", methodName, "constraint", receiverConstraint)
		return nil, fmt.Errorf("%w: method '%s' with receiver '%s' not found on type '%s'", ErrSymbolNotFound, methodName, receiverConstraint, tn.Name())
	} else {
		logger.Warn("[FINDMETHOD] Method not found on type (checked value and pointer receivers)", "typeName", tn.Name(), "methodName", methodName)
		return nil, fmt.Errorf("%w: method '%s' not found on type '%s' (checked value and pointer receivers)", ErrSymbolNotFound, methodName, tn.Name())
	}
}

// lookupMethodWithReceiver is a helper for findMethodOnType.
// It looks for a method with a specific name on a specific receiver type (value or pointer).
// If receiverConstraint is non-empty, it additionally validates the receiver matches.
func lookupMethodWithReceiver(pkg *types.Package, receiverType types.Type, methodName string, receiverConstraint string, logger logging.Logger) types.Object {
	mset := types.NewMethodSet(receiverType)
	sel := mset.Lookup(pkg, methodName)

	if sel == nil || sel.Kind() != types.MethodVal {
		logger.Debug("[LOOKUPMETHOD] Method not found in method set", "receiverType", types.TypeString(receiverType, nil), "methodName", methodName)
		return nil // Method not in the set for this receiver type
	}

	methodObj := sel.Obj() // The *types.Func object for the method

	// If no constraint, we found it.
	if receiverConstraint == "" {
		logger.Debug("[LOOKUPMETHOD] Method found, no constraint check needed", "receiverType", types.TypeString(receiverType, nil), "methodName", methodName)
		return methodObj
	}

	// Constraint provided, need to validate the receiver type string
	sig, ok := methodObj.Type().(*types.Signature)
	if !ok {
		logger.Error("[LOOKUPMETHOD] Internal error: Method object type is not *types.Signature", "methodName", methodName, "objType", fmt.Sprintf("%T", methodObj.Type()))
		return nil // Should not happen
	}
	recvVar := sig.Recv()
	if recvVar == nil {
		logger.Error("[LOOKUPMETHOD] Internal error: Method signature has nil receiver", "methodName", methodName)
		return nil // Should not happen for methods
	}

	actualReceiverTypeString := types.TypeString(recvVar.Type(), nil)   // e.g., "*mytestmodule/pkga.MyStruct" or "mytestmodule/pkga.MyStruct"
	constraintIsPointer := strings.HasPrefix(receiverConstraint, "*")   // e.g., true if "*MyStruct"
	actualIsPointer := strings.HasPrefix(actualReceiverTypeString, "*") // e.g., true if "*mytestmodule/pkga.MyStruct"

	// 1. Check pointer/non-pointer match
	if constraintIsPointer != actualIsPointer {
		logger.Debug("[LOOKUPMETHOD] Pointer mismatch", "constraint", receiverConstraint, "actual", actualReceiverTypeString)
		return nil // Pointer doesn't match constraint
	}

	// 2. Compare base type names (stripping pointer prefix first)
	constraintBaseName := strings.TrimPrefix(receiverConstraint, "*")          // e.g., "MyStruct"
	actualTypeNameWithPkg := strings.TrimPrefix(actualReceiverTypeString, "*") // e.g., "mytestmodule/pkga.MyStruct"

	// Extract the simple type name from the potentially qualified actual name
	actualBaseName := actualTypeNameWithPkg
	// Handle package path separator '.' vs potentially just type name if defined locally
	// Use LastIndex to get the part after the last '.', covering simple cases and pkg.Type
	if idx := strings.LastIndex(actualTypeNameWithPkg, "."); idx != -1 {
		actualBaseName = actualTypeNameWithPkg[idx+1:]
	}
	// If no '.', actualBaseName remains actualTypeNameWithPkg (type defined in same package, no qualifier needed)

	if constraintBaseName == actualBaseName {
		logger.Debug("[LOOKUPMETHOD] Receiver matched constraint", "constraint", receiverConstraint, "actual", actualReceiverTypeString)
		return methodObj // Match!
	} else {
		logger.Debug("[LOOKUPMETHOD] Base name mismatch", "constraintBase", constraintBaseName, "actualBase", actualBaseName, "actualFull", actualReceiverTypeString)
		return nil // Base names don't match
	}
}
