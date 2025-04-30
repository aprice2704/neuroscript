// filename: pkg/neurogo/app_interface.go
package neurogo

// import (
// 	 "context" // No longer needed here
//
// 	 "github.com/aprice2704/neuroscript/pkg/core" // No longer needed here
// )

// LLMClient interface definition moved to pkg/interfaces/llm_client.go
// to break import cycles.

// PatchHandler defines the interface for handling patch requests.
type PatchHandler interface {
	HandlePatch(patch string) error
}

// Add other interfaces specific to the neurogo application's internal components here.
// For example, if the TUI needed to interact with the main App logic via an interface:
// type TUIController interface {
//     SubmitInput(input string)
//     GetCurrentState() SomeStateType
// }
