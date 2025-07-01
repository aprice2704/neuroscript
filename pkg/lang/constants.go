// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines shared string constants for map keys and other literals used across the interpreter.
// filename: pkg/core/constants.go
// nlines: 20
// risk_rating: LOW

package lang

// Standardized keys for map-like Value types.
const (
	// EventKeyName is the key for an event's name in an EventValue map.
	EventKeyName	= "name"
	// EventKeySource is the key for an event's source in an EventValue map.
	EventKeySource	= "source"
	// EventKeyPayload is the key for an event's payload in an EventValue map.
	EventKeyPayload	= "payload"

	// ErrorKeyMessage is the key for an error's message in an ErrorValue map.
	ErrorKeyMessage	= "message"
	// ErrorKeyCode is the key for an error's code in an ErrorValue map.
	ErrorKeyCode	= "code"
	// ErrorKeyDetails is the key for an error's details in an ErrorValue map.
	ErrorKeyDetails	= "details"
)
