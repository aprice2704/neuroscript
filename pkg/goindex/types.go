// NeuroScript Go Indexer - Data Structures
// File version: 2.0.2
// Purpose: Defines the Go structs for the code index. Added FQN to MethodDetail.
// filename: pkg/goindex/types.go
package goindex

// ParamDetail holds information about a function/method parameter.
type ParamDetail struct {
	Name string `json:"name"`
	Type string `json:"type"` // Fully qualified type string
}

// FunctionDetail holds information about an exported function.
type FunctionDetail struct {
	Name       string        `json:"name"`       // Fully qualified name (e.g., "github.com/user/repo/pkg.FuncName")
	SourceFile string        `json:"sourceFile"` // Relative to component root
	Parameters []ParamDetail `json:"parameters,omitempty"`
	Returns    []string      `json:"returns,omitempty"` // List of type strings
	IsExported bool          `json:"isExported,omitempty"`
	IsMethod   bool          `json:"isMethod,omitempty"` // True if this FunctionDetail represents a method (used if methods are duplicated here for unified search)
}

// MethodDetail holds information about an exported method.
type MethodDetail struct {
	ReceiverName string        `json:"receiverName,omitempty"` // e.g., "MyType" or "*MyType"
	ReceiverType string        `json:"receiverType,omitempty"` // Type string as it appears in source (e.g., "MyType", "*MyType")
	Name         string        `json:"name"`                   // Simple method name (e.g., "MyMethod")
	FQN          string        `json:"fqn"`                    // Fully qualified name (e.g., "github.com/user/repo/pkg.(MyType).MyMethod")
	SourceFile   string        `json:"sourceFile"`             // Relative to component root
	Parameters   []ParamDetail `json:"parameters,omitempty"`   // Excludes receiver
	Returns      []string      `json:"returns,omitempty"`
	IsExported   bool          `json:"isExported,omitempty"`
}

// FieldDetail holds information about an exported struct field.
type FieldDetail struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // Fully qualified type string
	Tags     string `json:"tags,omitempty"`
	Exported bool   `json:"exported"`
}

// StructDetail holds information about an exported struct.
type StructDetail struct {
	Name       string        `json:"name"` // Simple struct name
	FQN        string        `json:"fqn"`  // Fully qualified struct name "pkg.path.StructName"
	SourceFile string        `json:"sourceFile"`
	Fields     []FieldDetail `json:"fields,omitempty"`
}

// InterfaceDetail holds information about an exported interface.
type InterfaceDetail struct {
	Name       string         `json:"name"` // Simple interface name
	FQN        string         `json:"fqn"`  // Fully qualified interface name "pkg.path.InterfaceName"
	SourceFile string         `json:"sourceFile"`
	Methods    []MethodDetail `json:"methods,omitempty"` // Method set for the interface (MethodDetail.FQN here would be for the interface method itself)
}

// GlobalVarDetail holds information about an exported global variable.
type GlobalVarDetail struct {
	Name       string `json:"name"` // Simple var name
	FQN        string `json:"fqn"`  // Fully qualified var name "pkg.path.VarName"
	Type       string `json:"type"`
	SourceFile string `json:"sourceFile"`
	Value      string `json:"value,omitempty"`
}

// GlobalConstDetail holds information about an exported global constant.
type GlobalConstDetail struct {
	Name       string `json:"name"` // Simple const name
	FQN        string `json:"fqn"`  // Fully qualified const name "pkg.path.ConstName"
	Type       string `json:"type"`
	SourceFile string `json:"sourceFile"`
	Value      string `json:"value"`
}

// TypeAliasDetail holds information about an exported type alias or definition.
type TypeAliasDetail struct {
	Name           string `json:"name"` // Simple type alias name
	FQN            string `json:"fqn"`  // Fully qualified type alias name "pkg.path.TypeName"
	UnderlyingType string `json:"underlyingType"`
	SourceFile     string `json:"sourceFile"`
}

// PackageDetail aggregates all exported entities within a Go package.
type PackageDetail struct {
	PackagePath  string              `json:"packagePath"` // Fully qualified Go package path
	PackageName  string              `json:"packageName"` // Short package name
	Functions    []FunctionDetail    `json:"functions,omitempty"`
	Methods      []MethodDetail      `json:"methods,omitempty"`
	Structs      []StructDetail      `json:"structs,omitempty"`
	Interfaces   []InterfaceDetail   `json:"interfaces,omitempty"`
	GlobalVars   []GlobalVarDetail   `json:"globalVars,omitempty"`
	GlobalConsts []GlobalConstDetail `json:"globalConsts,omitempty"`
	TypeAliases  []TypeAliasDetail   `json:"typeAliases,omitempty"`
}

// NeuroScriptArgDetail (no changes from your v2.0.1)
type NeuroScriptArgDetail struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
}

// NeuroScriptToolDetail (no changes from your v2.0.1)
type NeuroScriptToolDetail struct {
	Name                               string                 `json:"name"`
	Description                        string                 `json:"description"`
	Category                           string                 `json:"category,omitempty"`
	Args                               []NeuroScriptArgDetail `json:"args,omitempty"`
	ReturnType                         string                 `json:"returnType"`
	ReturnHelp                         string                 `json:"returnHelp,omitempty"`
	Variadic                           bool                   `json:"variadic,omitempty"`
	Example                            string                 `json:"example,omitempty"`
	ErrorConditions                    string                 `json:"errorConditions,omitempty"`
	ImplementingGoFunctionFullName     string                 `json:"implementingGoFunctionFullName,omitempty"`
	ImplementingGoFunctionSourceFile   string                 `json:"implementingGoFunctionSourceFile,omitempty"`
	GoImplementingFunctionReturnsError bool                   `json:"goImplementingFunctionReturnsError"`
	ComponentPath                      string                 `json:"componentPath,omitempty"`
}

// ComponentIndex (no changes from your v2.0.1)
type ComponentIndex struct {
	ComponentName      string                    `json:"componentName"`
	ComponentPath      string                    `json:"componentPath"`
	IndexSchemaVersion string                    `json:"indexSchemaVersion"`
	Packages           map[string]*PackageDetail `json:"packages"`
	NeuroScriptTools   []NeuroScriptToolDetail   `json:"neuroscriptTools,omitempty"`
	LastIndexed        string                    `json:"lastIndexed,omitempty"`
	GitBranch          string                    `json:"gitBranch,omitempty"`
	GitCommitHash      string                    `json:"gitCommitHash,omitempty"`
}

// ProjectIndex (no changes from your v2.0.1)
type ProjectIndex struct {
	ProjectRootModulePath string                             `json:"projectRootModulePath"`
	IndexSchemaVersion    string                             `json:"indexSchemaVersion"`
	LastIndexedTimestamp  string                             `json:"lastIndexedTimestamp"`
	Components            map[string]ComponentIndexFileEntry `json:"components"`
	GitBranch             string                             `json:"gitBranch,omitempty"`
	GitCommitHash         string                             `json:"gitCommitHash,omitempty"`
}

// ComponentIndexFileEntry (no changes from your v2.0.1)
type ComponentIndexFileEntry struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	IndexFile   string `json:"indexFile"`
	Description string `json:"description,omitempty"`
}
