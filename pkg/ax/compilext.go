package ax

// Registry is the front door to add builtins/types without init-magic.
type Registry interface {
	RegisterBuiltin(name string, fn any) error
	RegisterType(name string, factory any) error
	// Add cautiously: RegisterOp, RegisterMacro, ...
}

// Extension is what external packages implement to add features.
type Extension interface {
	Name() string
	Register(Registry) error
}
