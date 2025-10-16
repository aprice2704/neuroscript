// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Defines the core Actor and ActorProvider interfaces for identity-aware execution.
// filename: pkg/interfaces/actor.go
// nlines: 17
// risk_rating: LOW

package interfaces

// Actor represents any entity that can perform an action in the system.
// This interface is defined here to be accessible by the core interpreter
// and tools without creating import cycles.
type Actor interface {
	// DID returns the canonical Decentralized Identifier for the actor.
	DID() string
}

// ActorProvider is an interface for any object that can provide an Actor.
// This is the contract that allows tools to securely access the identity
// bound to an interpreter's runtime context.
type ActorProvider interface {
	Actor() (Actor, bool)
}
