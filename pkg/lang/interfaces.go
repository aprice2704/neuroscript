// filename: pkg/lang/interfaces.go
package lang

// Callable represents anything that can be called like a function.
type Callable interface {
	IsCallable()
	Name() string // Getter for the function's name
}
