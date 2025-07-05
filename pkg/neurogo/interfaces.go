// NeuroScript Version: 0.3.0
// File version: 0.0.6
// Corrected package for AIWorkerManager and updated TUIController method signature.
// filename: pkg/neurogo/interfaces.go
package neurogo

import (
	"io/fs"
	"os"
	// Corrected import for AIWorkerManager
)

// WMStatusViewDataProvider defines the methods required by the WMStatusScreen
// (or its Formatter) to fetch the data it needs to display.
// The main App struct will implement this interface.
// type WMStatusViewDataProvider interface {
// 	GetLogger() interfaces.Logger
// 	GetAIWorkerManager() *wm.AIWorkerManager // Corrected package
// }

// TUIController defines the methods the App can use to interact with the TUI.
// This helps decouple the App logic from the concrete TUI implementation.
type TUIController interface {
	// CreateAndShowNewChatScreen instructs the TUI to create a new visual representation
	// for a chat session, make it active, and handle focusing relevant UI elements.
	// This method MUST be callable via app.QueueUpdateDraw by the App to ensure
	// it runs on the TUI's main goroutine.
	// MODIFIED: Now accepts a *ChatSession object directly.
	CreateAndShowNewChatScreen(session *ChatSession) // Corrected signature

	// Add other methods here if App needs to command TUI for other specific UI actions.
}

// FileSystemOperations defines an interface for basic file system interactions.
// This allows for easier testing by mocking the file system.
type FileSystemOperations interface {
	Stat(name string) (fs.FileInfo, error)
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	Remove(name string) error
	UserHomeDir() (string, error)
	Getenv(key string) string
}

// StandardFileSystem implements FileSystemOperations using the os package.
type StandardFileSystem struct{}

func (sfs *StandardFileSystem) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }
func (sfs *StandardFileSystem) ReadFile(name string) ([]byte, error)  { return os.ReadFile(name) }
func (sfs *StandardFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (sfs *StandardFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}
func (sfs *StandardFileSystem) Remove(name string) error     { return os.Remove(name) }
func (sfs *StandardFileSystem) UserHomeDir() (string, error) { return os.UserHomeDir() }
func (sfs *StandardFileSystem) Getenv(key string) string     { return os.Getenv(key) }
