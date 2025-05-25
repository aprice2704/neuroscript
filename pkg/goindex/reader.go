// NeuroScript Go Indexer - Index Reader
// File version: 1.0.2 // Fixed unused variables in FindMethod, clarified pkgPath usage.
// Purpose: Provides functions to load and query the Go code index.
// filename: pkg/goindex/reader.go
package goindex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IndexReader provides methods to read and query the project and component indexes.
type IndexReader struct {
	projectRootPath  string
	ProjectIndex     *ProjectIndex
	loadedComponents map[string]*ComponentIndex // Cache for loaded component indexes
	indexDir         string                     // Store the directory where project_index.json was found
}

// NewIndexReader creates a new reader for the code index.
// projectRoot is the absolute path to the root of the indexed project.
// indexDir is the directory where project_index.json and component index files are stored.
func NewIndexReader(projectRoot, indexDir string) (*IndexReader, error) {
	indexPath := filepath.Join(indexDir, "project_index.json")
	projectIndex, err := LoadProjectIndex(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load project index from %s: %w", indexPath, err)
	}

	return &IndexReader{
		projectRootPath:  projectRoot,
		ProjectIndex:     projectIndex,
		loadedComponents: make(map[string]*ComponentIndex),
		indexDir:         indexDir, // Store for loading component indexes
	}, nil
}

// LoadProjectIndex loads the main project index file.
func LoadProjectIndex(filePath string) (*ProjectIndex, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project index file %s: %w", filePath, err)
	}

	var index ProjectIndex
	err = json.Unmarshal(data, &index)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal project index from %s: %w", filePath, err)
	}
	return &index, nil
}

// GetComponentIndex loads a specific component's index.
// It uses the IndexReader's cache to avoid reloading.
func (r *IndexReader) GetComponentIndex(componentName string) (*ComponentIndex, error) {
	if compIndex, ok := r.loadedComponents[componentName]; ok {
		return compIndex, nil
	}

	componentMeta, ok := r.ProjectIndex.Components[componentName]
	if !ok {
		return nil, fmt.Errorf("component '%s' not found in project index", componentName)
	}

	compIndexFilePath := filepath.Join(r.indexDir, componentMeta.IndexFile)

	data, err := os.ReadFile(compIndexFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read component index file %s: %w", compIndexFilePath, err)
	}

	var compIndex ComponentIndex
	err = json.Unmarshal(data, &compIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal component index from %s: %w", compIndexFilePath, err)
	}

	r.loadedComponents[componentName] = &compIndex
	return &compIndex, nil
}

// FindFunction searches for a function by its fully qualified name across all components.
func (r *IndexReader) FindFunction(fullyQualifiedFuncName string) (*FunctionDetail, *ComponentIndex, *PackageDetail, error) {
	pkgPath, funcName := splitFullyQualifiedName(fullyQualifiedFuncName)
	if pkgPath == "" || funcName == "" {
		return nil, nil, nil, fmt.Errorf("invalid fully qualified function name: %s", fullyQualifiedFuncName)
	}

	for compName := range r.ProjectIndex.Components {
		compIndex, err := r.GetComponentIndex(compName)
		if err != nil {
			fmt.Printf("Warning: could not load component %s: %v\n", compName, err)
			continue
		}
		if pkgDetail, ok := compIndex.Packages[pkgPath]; ok {
			for i := range pkgDetail.Functions { // Iterate by index to get a pointer
				f := &pkgDetail.Functions[i]
				// FunctionDetail.Name is already fully qualified
				if f.Name == fullyQualifiedFuncName {
					return f, compIndex, pkgDetail, nil
				}
			}
		}
	}
	return nil, nil, nil, fmt.Errorf("function %s not found", fullyQualifiedFuncName)
}

// FindStruct searches for a struct by its fully qualified name.
// Example: "github.com/aprice2704/neuroscript/pkg/core.MyStruct"
func (r *IndexReader) FindStruct(fullyQualifiedStructName string) (*StructDetail, *ComponentIndex, *PackageDetail, error) {
	pkgPath, structName := splitFullyQualifiedName(fullyQualifiedStructName)
	if pkgPath == "" || structName == "" {
		return nil, nil, nil, fmt.Errorf("invalid fully qualified struct name: %s", fullyQualifiedStructName)
	}

	for compName := range r.ProjectIndex.Components {
		compIndex, err := r.GetComponentIndex(compName)
		if err != nil {
			continue
		}
		if pkgDetail, ok := compIndex.Packages[pkgPath]; ok {
			for i := range pkgDetail.Structs { // Iterate by index to get a pointer
				s := &pkgDetail.Structs[i]
				if s.Name == structName { // StructDetail.Name is the short name
					return s, compIndex, pkgDetail, nil
				}
			}
		}
	}
	return nil, nil, nil, fmt.Errorf("struct %s not found in package %s", structName, pkgPath)
}

// FindMethod searches for a specific method on a receiver type.
// receiverTypeFQN: Fully qualified name of the receiver type (e.g., "github.com/aprice2704/neuroscript/pkg/core.MyType" or "*github.com/aprice2704/neuroscript/pkg/core.MyType")
// methodName: The short name of the method.
func (r *IndexReader) FindMethod(receiverTypeFQN string, methodName string) (*MethodDetail, *ComponentIndex, *PackageDetail, error) {
	derivedPkgPath, _ := splitFullyQualifiedName(strings.TrimPrefix(receiverTypeFQN, "*"))

	if derivedPkgPath == "" { // Could not derive package path, search all packages (less efficient)
		for compName := range r.ProjectIndex.Components {
			compIndex, err := r.GetComponentIndex(compName)
			if err != nil {
				continue
			}
			for _, pkgDetail := range compIndex.Packages {
				for i := range pkgDetail.Methods { // Iterate by index to get a pointer
					method := &pkgDetail.Methods[i]
					// Normalize both index receiver type and query receiver type for pointer comparison
					normalizedIndexReceiverType := strings.TrimPrefix(method.ReceiverType, "*")
					normalizedQueryReceiverType := strings.TrimPrefix(receiverTypeFQN, "*")

					// Check if the base types match and if the pointer characteristic matches (or if one is pointer and other is not, for flexibility if needed)
					// A more precise match would ensure pointer stars align or one implies the other based on context.
					// For now, matching base type and method name.
					// Also check direct equality for cases where FQN matches exactly including *.
					if (normalizedIndexReceiverType == normalizedQueryReceiverType || method.ReceiverType == receiverTypeFQN) && method.Name == methodName {
						return method, compIndex, pkgDetail, nil
					}
				}
			}
		}
	} else { // Package path was derived, search only that package
		for compName := range r.ProjectIndex.Components {
			compIndex, err := r.GetComponentIndex(compName)
			if err != nil {
				continue
			}
			if pkgDetail, ok := compIndex.Packages[derivedPkgPath]; ok {
				for i := range pkgDetail.Methods { // Iterate by index to get a pointer
					method := &pkgDetail.Methods[i]
					normalizedIndexReceiverType := strings.TrimPrefix(method.ReceiverType, "*")
					normalizedQueryReceiverType := strings.TrimPrefix(receiverTypeFQN, "*")
					if (normalizedIndexReceiverType == normalizedQueryReceiverType || method.ReceiverType == receiverTypeFQN) && method.Name == methodName {
						return method, compIndex, pkgDetail, nil
					}
				}
			}
		}
	}
	return nil, nil, nil, fmt.Errorf("method %s.%s not found", receiverTypeFQN, methodName)
}

// splitFullyQualifiedName splits a FQN like "github.com/aprice2704/neuroscript/pkg/core.MyFunction"
// into package path ("github.com/aprice2704/neuroscript/pkg/core") and entity name ("MyFunction").
func splitFullyQualifiedName(fqn string) (pkgPath string, entityName string) {
	lastDot := strings.LastIndex(fqn, ".")
	if lastDot == -1 {
		// Could be a package-level entity in a root package (e.g. package main, func main)
		// or an unqualified name. For unqualified, assume no package path.
		return "", fqn
	}
	pkgPath = fqn[:lastDot]
	entityName = fqn[lastDot+1:]
	return
}

// GetNeuroScriptTool retrieves the specification for a named NeuroScript tool.
func (r *IndexReader) GetNeuroScriptTool(toolName string) (*NeuroScriptToolDetail, *ComponentIndex, error) {
	for compName := range r.ProjectIndex.Components {
		compIndex, err := r.GetComponentIndex(compName)
		if err != nil {
			fmt.Printf("Warning: could not load component %s for tool search: %v\n", compName, err)
			continue
		}
		if compIndex.NeuroScriptTools != nil {
			for i := range compIndex.NeuroScriptTools { // Iterate by index to get a pointer
				tool := &compIndex.NeuroScriptTools[i]
				if tool.Name == toolName {
					return tool, compIndex, nil
				}
			}
		}
	}
	return nil, nil, fmt.Errorf("NeuroScript tool '%s' not found", toolName)
}
