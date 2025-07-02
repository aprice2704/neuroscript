// File: pkg/goindex/reader.go
// NeuroScript Go Index Reader
// File version: 0.2.5 (Added GetAllNeuroScriptToolDetails)
// Purpose: Reads and provides access to the Go code index. Includes logic for linking NeuroScript tools.
// filename: pkg/goindex/reader.go

package goindex

import (
	"encoding/json"
	"fmt"
	"log" // Added for logging within the new method
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	// Assuming this is needed for NeuroScriptToolDetail's Args mapping
)

// IndexReader provides thread-safe access to the loaded code indexes.
type IndexReader struct {
	projectRootPath  string
	indexPath        string // Directory where index files (project_index.json, component_*.json) are stored
	projectIndex     *ProjectIndex
	componentIndexes map[string]*ComponentIndex // Key: component name
	mu               sync.RWMutex
}

// FoundCallableItem holds common details from a found FunctionDetail or MethodDetail.
type FoundCallableItem struct {
	FQN        string
	SourceFile string
	Returns    []string
	IsExported bool
	// Add other common fields if necessary, e.g., Parameters
}

// NewIndexReader creates a new IndexReader.
// projectRoot is the absolute path to the root of the Go project being indexed.
// indexPath is the directory where the index JSON files are located.
func NewIndexReader(projectRoot, indexPath string) (*IndexReader, error) {
	if projectRoot == "" || indexPath == "" {
		return nil, fmt.Errorf("projectRoot and indexPath cannot be empty")
	}
	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute project root: %w", err)
	}
	absIndexPath, err := filepath.Abs(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute index path: %w", err)
	}

	return &IndexReader{
		projectRootPath:  absProjectRoot,
		indexPath:        absIndexPath,
		componentIndexes: make(map[string]*ComponentIndex),
	}, nil
}

// LoadProjectIndex loads the main project_index.json file.
func (ir *IndexReader) LoadProjectIndex() error {
	ir.mu.Lock()
	defer ir.mu.Unlock()

	filePath := filepath.Join(ir.indexPath, "project_index.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read project index file %s: %w", filePath, err)
	}

	var pi ProjectIndex
	if err := json.Unmarshal(data, &pi); err != nil {
		return fmt.Errorf("failed to unmarshal project index from %s: %w", filePath, err)
	}
	ir.projectIndex = &pi
	// Clear component indexes as they might be stale if project index changed
	ir.componentIndexes = make(map[string]*ComponentIndex)
	return nil
}

// ProjectIndexLoaded checks if the project index has been loaded.
func (ir *IndexReader) ProjectIndexLoaded() bool {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	return ir.projectIndex != nil
}

// GetProjectIndex returns a copy of the loaded project index.
// Returns nil if not loaded.
func (ir *IndexReader) GetProjectIndex() *ProjectIndex {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	if ir.projectIndex == nil {
		return nil
	}
	piCopy := *ir.projectIndex // Create a shallow copy
	// Deep copy the map if modification by caller is a concern, or document as read-only.
	piCopy.Components = make(map[string]ComponentIndexFileEntry)
	for k, v := range ir.projectIndex.Components {
		piCopy.Components[k] = v
	}
	return &piCopy
}

// GetRepoModulePath returns the repository module path from the project index.
func (ir *IndexReader) GetRepoModulePath() string {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	if ir.projectIndex == nil {
		return ""
	}
	return ir.projectIndex.ProjectRootModulePath
}

// LoadComponentIndex loads a specific component's index file.
func (ir *IndexReader) LoadComponentIndex(componentName string) (*ComponentIndex, error) {
	ir.mu.RLock() // Check cache under RLock first
	if compIdx, exists := ir.componentIndexes[componentName]; exists && compIdx != nil {
		ir.mu.RUnlock()
		return compIdx, nil
	}
	ir.mu.RUnlock()

	ir.mu.Lock() // Acquire full lock to load if not found or stale
	defer ir.mu.Unlock()

	// Re-check cache now that we have the full lock, in case another goroutine loaded it.
	if compIdx, exists := ir.componentIndexes[componentName]; exists && compIdx != nil {
		return compIdx, nil
	}

	if ir.projectIndex == nil {
		return nil, fmt.Errorf("project index not loaded, cannot load component index for %s", componentName)
	}

	compMeta, metaExists := ir.projectIndex.Components[componentName]
	if !metaExists {
		return nil, fmt.Errorf("component '%s' not found in project index", componentName)
	}

	filePath := filepath.Join(ir.indexPath, compMeta.IndexFile)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read component index file %s for component %s: %w", filePath, componentName, err)
	}

	var ci ComponentIndex
	if err := json.Unmarshal(data, &ci); err != nil {
		return nil, fmt.Errorf("failed to unmarshal component index from %s for %s: %w", filePath, componentName, err)
	}
	ir.componentIndexes[componentName] = &ci
	return &ci, nil
}

// GetComponentIndex retrieves a loaded component index. Loads it if not already loaded.
func (ir *IndexReader) GetComponentIndex(componentName string) (*ComponentIndex, error) {
	ir.mu.RLock()
	cachedCi, exists := ir.componentIndexes[componentName]
	ir.mu.RUnlock()

	if exists && cachedCi != nil {
		return cachedCi, nil
	}
	return ir.LoadComponentIndex(componentName)
}

// LoadAllComponentIndexes loads all component indexes defined in the project index.
func (ir *IndexReader) LoadAllComponentIndexes() error {
	if !ir.ProjectIndexLoaded() {
		if err := ir.LoadProjectIndex(); err != nil { // Ensure project index is loaded first
			return fmt.Errorf("failed to load project index before loading components: %w", err)
		}
	}

	ir.mu.RLock()
	if ir.projectIndex == nil {
		ir.mu.RUnlock()
		return fmt.Errorf("project index is nil after attempting load, cannot load components")
	}
	// Create a slice of component names to load to avoid holding lock during file I/O
	componentsToLoad := make([]string, 0, len(ir.projectIndex.Components))
	for compName := range ir.projectIndex.Components {
		componentsToLoad = append(componentsToLoad, compName)
	}
	ir.mu.RUnlock()

	for _, compName := range componentsToLoad {
		if _, err := ir.GetComponentIndex(compName); err != nil { // GetComponentIndex handles its own locking
			return fmt.Errorf("failed to load component index for '%s': %w", compName, err)
		}
	}
	return nil
}

// GetAllNeuroScriptToolDetails retrieves all NeuroScriptToolDetail objects
// from all loaded and successfully parsed component indexes.
// It ensures that component indexes are loaded before attempting to retrieve tool details.
func (ir *IndexReader) GetAllNeuroScriptToolDetails() ([]NeuroScriptToolDetail, error) {
	// Ensure project and all component indexes are loaded.
	// LoadAllComponentIndexes internally calls LoadProjectIndex if needed.
	if err := ir.LoadAllComponentIndexes(); err != nil {
		return nil, fmt.Errorf("failed to ensure all component indexes were loaded: %w", err)
	}

	ir.mu.RLock() // Lock for reading componentIndexes map
	defer ir.mu.RUnlock()

	var allToolDetails []NeuroScriptToolDetail
	if len(ir.componentIndexes) == 0 {
		log.Println("[IndexReader] GetAllNeuroScriptToolDetails: No component indexes loaded or available.")
	}

	for componentName, componentIndex := range ir.componentIndexes {
		if componentIndex != nil && len(componentIndex.NeuroScriptTools) > 0 {
			allToolDetails = append(allToolDetails, componentIndex.NeuroScriptTools...)
		} else if componentIndex == nil {
			log.Printf("[IndexReader] GetAllNeuroScriptToolDetails: Component index for '%s' is nil.", componentName)
		}
	}

	if len(allToolDetails) == 0 {
		log.Println("[IndexReader] GetAllNeuroScriptToolDetails: No NeuroScript tool details found across all loaded components. This could be normal if no tools are indexed, or if the indexer did not populate them.")
	}
	return allToolDetails, nil
}

// ... (GenerateEnhancedToolDetails, findCallableByQualifiedNameInternal, FindFunction, FindStruct, FindMethod, GetNeuroScriptTool, etc., remain the same) ...
// ... make sure mapCoreArgSpecsToNeuroScriptArgDetails is also present or moved to types.go and imported ...

// mapCoreArgSpecsToNeuroScriptArgDetails converts  ArgSpec to the index's NeuroScriptArgDetail.
// TODO: This helper is also in cmd/goindexer/main.go. Consider moving to a shared location (e.g., this package).
func mapCoreArgSpecsToNeuroScriptArgDetails(coreArgs []pec) []NeuroScriptArgDetail {
	if coreArgs == nil {
		return nil
	}
	nsArgs := make([]NeuroScriptArgDetail, len(coreArgs))
	for i, ca := range coreArgs {
		nsArgs[i] = NeuroScriptArgDetail{
			Name:         ca.Name,
			Type:         string(ca.Type), // Assumes  ype is string-convertible
			Description:  ca.Description,
			Required:     ca.Required,
			DefaultValue: ca.DefaultValue,
		}
	}
	return nsArgs
}

// GenerateEnhancedToolDetails dynamically links ToolSpecs from an interpreter
// with their Go implementation details using a loaded code index.
func (ir *IndexReader) GenerateEnhancedToolDetails(
	interpreter interpreter.Interpreter,
) ([]NeuroScriptToolDetail, error) {
	if interpreter == nil {
		return nil, fmt.Errorf("interpreter cannot be nil for generating enhanced tool details")
	}
	// Ensure indexes are loaded
	if err := ir.LoadAllComponentIndexes(); err != nil {
		return nil, fmt.Errorf("IndexReader: failed to load necessary indexes for tool detail generation: %w", err)
	}

	var allEnhancedDetails []NeuroScriptToolDetail
	toolSpecs := interpreter.ListTools()

	for _, spec := range toolSpecs {
		toolImpl, exists := interpreter.GetTool(spec.Name)
		if !exists || toolImpl.Func == nil {
			fmt.Printf("goindex.GenerateEnhancedToolDetails: Warning - Tool %s found in spec list but not in GetTool or has nil Func\n", spec.Name)
			// Create a partial detail based only on the spec
			allEnhancedDetails = append(allEnhancedDetails, NeuroScriptToolDetail{
				Name:            spec.Name,
				Description:     spec.Description,
				Category:        spec.Category,
				Args:            mapCoreArgSpecsToNeuroScriptArgDetails(spec.Args),
				ReturnType:      string(spec.ReturnType),
				ReturnHelp:      spec.ReturnHelp,
				Variadic:        spec.Variadic,
				Example:         spec.Example,
				ErrorConditions: spec.ErrorConditions,
			})
			continue
		}

		goFuncName := runtime.FuncForPC(reflect.ValueOf(toolImpl.Func).Pointer()).Name()
		foundItem, componentNameFound, err := ir.findCallableByQualifiedNameInternal(goFuncName)

		detail := NeuroScriptToolDetail{
			Name:                           spec.Name,
			Description:                    spec.Description,
			Category:                       spec.Category,
			Args:                           mapCoreArgSpecsToNeuroScriptArgDetails(spec.Args),
			ReturnType:                     string(spec.ReturnType),
			ReturnHelp:                     spec.ReturnHelp,
			Variadic:                       spec.Variadic,
			Example:                        spec.Example,
			ErrorConditions:                spec.ErrorConditions,
			ImplementingGoFunctionFullName: goFuncName, // Store what runtime reported
		}

		if err != nil {
			fmt.Printf("goindex.GenerateEnhancedToolDetails: Warning - Could not find callable detail for Go entity %s (tool %s): %v\n", goFuncName, spec.Name, err)
			// ImplementingGoFunctionSourceFile will be empty, ComponentPath empty, GoImplementingFunctionReturnsError default false
		} else if foundItem != nil {
			detail.ImplementingGoFunctionFullName = foundItem.FQN // Prefer FQN from index
			detail.ImplementingGoFunctionSourceFile = foundItem.SourceFile

			compMeta, metaExists := ir.getComponentMeta(componentNameFound)
			if !metaExists {
				fmt.Printf("goindex.GenerateEnhancedToolDetails: Warning - Component metadata for component name '%s' not found (tool %s, func %s)\n", componentNameFound, spec.Name, goFuncName)
				// ComponentPath will be empty
			} else {
				detail.ComponentPath = compMeta.Path
			}

			for _, retType := range foundItem.Returns {
				if retType == "error" {
					detail.GoImplementingFunctionReturnsError = true
					break
				}
			}
		}
		allEnhancedDetails = append(allEnhancedDetails, detail)
	}
	return allEnhancedDetails, nil
}

// findCallableByQualifiedNameInternal searches for a function or method by its fully qualified name
// across all loaded component indexes.
func (ir *IndexReader) findCallableByQualifiedNameInternal(qualifiedName string) (*FoundCallableItem, string, error) {
	ir.mu.RLock() // Assumes components are already loaded by a public method like LoadAllComponentIndexes if needed.
	defer ir.mu.RUnlock()

	searchNameForMethod := qualifiedName
	if strings.HasSuffix(qualifiedName, "-fm") {
		searchNameForMethod = strings.TrimSuffix(qualifiedName, "-fm")
	}

	var componentsChecked []string
	for componentName, componentIndex := range ir.componentIndexes {
		componentsChecked = append(componentsChecked, componentName) // For error reporting
		if componentIndex == nil {
			continue
		}
		for _, pkgDetail := range componentIndex.Packages {
			if pkgDetail == nil {
				continue
			}
			// Search Functions
			for i := range pkgDetail.Functions {
				if pkgDetail.Functions[i].Name == qualifiedName {
					f := pkgDetail.Functions[i]
					return &FoundCallableItem{
						FQN: f.Name, SourceFile: f.SourceFile, Returns: f.Returns, IsExported: f.IsExported,
					}, componentName, nil
				}
			}
			// Search Methods
			for i := range pkgDetail.Methods {
				if pkgDetail.Methods[i].FQN == searchNameForMethod {
					m := pkgDetail.Methods[i]
					return &FoundCallableItem{
						FQN: m.FQN, SourceFile: m.SourceFile, Returns: m.Returns, IsExported: m.IsExported,
					}, componentName, nil
				}
			}
		}
	}
	return nil, "", fmt.Errorf("callable %q (method search '%s') not found. Searched components: %s", qualifiedName, searchNameForMethod, strings.Join(componentsChecked, ", "))
}

func (ir *IndexReader) getComponentMeta(componentName string) (*ComponentIndexFileEntry, bool) {
	// Assumes ir.projectIndex is loaded and ir.mu (RLock) is held by caller if necessary
	if ir.projectIndex == nil {
		return nil, false
	}
	compMeta, exists := ir.projectIndex.Components[componentName]
	return &compMeta, exists // Return copy or pointer based on usage; here pointer is fine as it's read.
}

// FindFunction, FindStruct, FindMethod, GetNeuroScriptTool (for direct access from component index)
// These methods remain useful for targeted lookups if you already know the component and package.

func (ir *IndexReader) FindFunction(componentName, packagePath, funcName string) (*FunctionDetail, error) {
	compIdx, err := ir.GetComponentIndex(componentName)
	if err != nil {
		return nil, fmt.Errorf("FindFunction: failed to get component index for %s: %w", componentName, err)
	}
	if compIdx == nil {
		return nil, fmt.Errorf("FindFunction: component index for %s is nil", componentName)
	}
	// RLock for reading the specific component index's packages
	ir.mu.RLock() // This lock might be too broad or needs careful thought if GetComponentIndex already loaded.
	// However, accessing compIdx.Packages requires a lock.
	defer ir.mu.RUnlock()

	pkgDetail, pkgExists := compIdx.Packages[packagePath]
	if !pkgExists || pkgDetail == nil {
		return nil, fmt.Errorf("package '%s' not found in component '%s'", packagePath, componentName)
	}
	var expectedFQN string
	if strings.Contains(funcName, ".") && strings.HasPrefix(funcName, packagePath) {
		expectedFQN = funcName
	} else {
		expectedFQN = packagePath + "." + funcName
	}
	for i := range pkgDetail.Functions {
		if pkgDetail.Functions[i].Name == expectedFQN {
			return &pkgDetail.Functions[i], nil
		}
	}
	return nil, fmt.Errorf("function with FQN '%s' not found in package '%s' of component '%s'", expectedFQN, packagePath, componentName)
}

func (ir *IndexReader) FindStruct(componentName, packagePath, structName string) (*StructDetail, error) {
	compIdx, err := ir.GetComponentIndex(componentName)
	if err != nil {
		return nil, fmt.Errorf("FindStruct: failed to get component index for %s: %w", componentName, err)
	}
	if compIdx == nil {
		return nil, fmt.Errorf("FindStruct: component index for %s is nil", componentName)
	}
	ir.mu.RLock()
	defer ir.mu.RUnlock()

	pkgDetail, pkgExists := compIdx.Packages[packagePath]
	if !pkgExists || pkgDetail == nil {
		return nil, fmt.Errorf("package '%s' not found in component '%s'", packagePath, componentName)
	}
	expectedStructFQN := packagePath + "." + structName
	for i := range pkgDetail.Structs {
		if pkgDetail.Structs[i].Name == structName && pkgDetail.Structs[i].FQN == expectedStructFQN {
			return &pkgDetail.Structs[i], nil
		}
	}
	return nil, fmt.Errorf("struct '%s' (expected FQN: %s) not found in package '%s' of component '%s'", structName, expectedStructFQN, packagePath, componentName)
}

func (ir *IndexReader) FindMethod(componentName, packagePath, methodFQN string) (*MethodDetail, error) {
	compIdx, err := ir.GetComponentIndex(componentName)
	if err != nil {
		return nil, fmt.Errorf("FindMethod: failed to get component index for %s: %w", componentName, err)
	}
	if compIdx == nil {
		return nil, fmt.Errorf("FindMethod: component index for %s is nil", componentName)
	}
	ir.mu.RLock()
	defer ir.mu.RUnlock()

	pkgDetail, pkgExists := compIdx.Packages[packagePath]
	if !pkgExists || pkgDetail == nil {
		return nil, fmt.Errorf("package '%s' not found in component '%s' for method lookup", packagePath, componentName)
	}
	for i := range pkgDetail.Methods {
		if pkgDetail.Methods[i].FQN == methodFQN {
			return &pkgDetail.Methods[i], nil
		}
	}
	return nil, fmt.Errorf("method with FQN '%s' not found in package '%s' of component '%s'", methodFQN, packagePath, componentName)
}

// GetNeuroScriptTool retrieves a specific NeuroScriptToolDetail directly from a component's index.
// This assumes the NeuroScriptTools array is populated in the component index JSON.
func (ir *IndexReader) GetNeuroScriptTool(componentName, toolName string) (*NeuroScriptToolDetail, error) {
	compIdx, err := ir.GetComponentIndex(componentName) // Ensures component is loaded
	if err != nil {
		return nil, fmt.Errorf("GetNeuroScriptTool: failed to get component index for %s: %w", componentName, err)
	}
	if compIdx == nil { // Should be caught by err above, but good practice
		return nil, fmt.Errorf("GetNeuroScriptTool: component index for %s is nil after attempting load", componentName)
	}

	ir.mu.RLock() // For reading compIdx.NeuroScriptTools
	defer ir.mu.RUnlock()

	if compIdx.NeuroScriptTools == nil {
		// This is a valid state if the component has no tools or they weren't indexed.
		return nil, fmt.Errorf("NeuroScript tool list is nil/empty for component '%s'", componentName)
	}

	for i := range compIdx.NeuroScriptTools {
		if compIdx.NeuroScriptTools[i].Name == toolName {
			return &compIdx.NeuroScriptTools[i], nil
		}
	}
	return nil, fmt.Errorf("NeuroScript tool '%s' not found in component '%s'", toolName, componentName)
}

// SaveComponentIndex and SaveAllLoadedComponentIndexes are not strictly reader functions but utilities.
// If they are only used by the indexer, they could live in cmd/goindexer/main.go or a utils package there.
// For now, keeping them here if IndexReader might be used for updates in some contexts.
func (ir *IndexReader) SaveComponentIndex(componentName string) error {
	ir.mu.RLock()
	compIdx, exists := ir.componentIndexes[componentName]
	if !exists || compIdx == nil {
		ir.mu.RUnlock()
		return fmt.Errorf("component index for '%s' not loaded, cannot save", componentName)
	}
	if ir.projectIndex == nil { // Should be loaded if compIdx is present
		ir.mu.RUnlock()
		return fmt.Errorf("project index not loaded, cannot determine save path for component '%s'", componentName)
	}
	compMeta, metaExists := ir.projectIndex.Components[componentName]
	if !metaExists {
		ir.mu.RUnlock()
		return fmt.Errorf("metadata for component '%s' not found in project index", componentName)
	}
	ir.mu.RUnlock() // Release read lock before potential write operations

	// To update LastIndexed, we need a write lock on the specific componentIndex's data,
	// or make a copy and marshal that. For simplicity, let's assume modifying in place is okay
	// IF ONLY ONE GOROUTINE SAVES AT A TIME or if saves are rare.
	// A safer approach is to lock compIdx specifically if it had its own mutex,
	// or pass a copy to MarshalIndent if LastIndexed needs frequent updates.

	// For this context, let's assume it's acceptable to update LastIndexed here
	// under the general IndexReader lock (if it were a full ir.mu.Lock() for save).
	// However, if we only held RLock, we should not modify compIdx.LastIndexed.
	// Let's assume for now that saving does not modify the in-memory compIdx.LastIndexed here.
	// The caller should update it if needed before calling save or this func should take a copy.

	data, err := json.MarshalIndent(compIdx, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal component index for %s: %w", componentName, err)
	}

	filePath := filepath.Join(ir.indexPath, compMeta.IndexFile)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for component index %s: %w", filePath, err)
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write component index file %s: %w", filePath, err)
	}
	return nil
}

func (ir *IndexReader) SaveAllLoadedComponentIndexes() error {
	ir.mu.RLock()
	componentNames := make([]string, 0, len(ir.componentIndexes))
	for name, comp := range ir.componentIndexes {
		if comp != nil { // Ensure we only try to save components that are actually loaded
			componentNames = append(componentNames, name)
		}
	}
	ir.mu.RUnlock()

	for _, name := range componentNames {
		if err := ir.SaveComponentIndex(name); err != nil {
			return fmt.Errorf("failed to save component index for %s: %w", name, err)
		}
	}
	return nil
}
