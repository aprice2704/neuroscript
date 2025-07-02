// NeuroScript Version: 0.3.0
// File version: 0.1.0
// Defines constants used by the neurogo application package.
// filename: pkg/neurogo/constants.go
package neurogo

// DefaultModelName specifies the default LLM model to be used by the application
// if not overridden by specific configurations (e.g., AIWorkerDefinition).
const DefaultModelName = "gemini-1.5-flash-latest"	// Or your preferred default

// DefaultDefinitionsFilename is the standard filename for storing AIWorkerDefinition JSON data
// within the sandbox directory.
const DefaultDefinitionsFilename = "ai_worker_definitions.json"

// DefaultPerformanceDataFilename is the standard filename for storing AIWorker performance records
// as JSON data within the sandbox directory.
const DefaultPerformanceDataFilename = "ai_worker_performance_data.json"