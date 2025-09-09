// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Provides pre-compiled shapes and convenience functions for validating and composing common data structures.
// filename: pkg/api/shape/predefined.go
// nlines: 77
// risk_rating: LOW

package shape

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	// NSEvent is a pre-compiled shape that can be used to validate a canonical
	// NeuroScript event object.
	NSEvent *Shape
)

// nsEventShapeMap is the raw Shape-Lite definition for a standard event.
var nsEventShapeMap = map[string]interface{}{
	"payload[]": map[string]interface{}{
		"ID":       "string",
		"Kind":     "string",
		"AgentID?": "string",
		"TS":       "int",
		"Payload":  "any",
	},
}

// NSEventComposeOptions provides optional parameters for the ComposeNSEvent function.
type NSEventComposeOptions struct {
	ID      string
	AgentID string
}

func init() {
	var err error
	// Parse the canonical shape definition at package initialization time.
	NSEvent, err = ParseShape(nsEventShapeMap)
	if err != nil {
		panic(fmt.Sprintf("failed to parse predefined NSEvent shape: %v", err))
	}
}

// ValidateNSEvent is a convenience function that validates a map against the
// canonical NeuroScript event shape.
func ValidateNSEvent(value map[string]interface{}, options *ValidateOptions) error {
	return NSEvent.Validate(value, options)
}

// ComposeNSEvent creates a valid NeuroScript event object. It uses the pre-compiled
// NSEvent shape to validate its own output, guaranteeing a correct structure.
func ComposeNSEvent(kind string, payload map[string]interface{}, options *NSEventComposeOptions) (map[string]interface{}, error) {
	id := ""
	agentID := ""
	if options != nil {
		id = options.ID
		agentID = options.AgentID
	}
	if id == "" {
		id = uuid.New().String()
	}

	eventEnvelope := map[string]interface{}{
		"ID":      id,
		"Kind":    kind,
		"AgentID": agentID,
		"TS":      time.Now().UnixNano(),
		"Payload": payload,
	}

	eventObject := map[string]interface{}{
		"payload": []interface{}{eventEnvelope},
	}

	// Self-validation to ensure correctness. This should never fail unless the
	// hardcoded shape and this function's logic are out of sync.
	if err := NSEvent.Validate(eventObject, nil); err != nil {
		return nil, fmt.Errorf("internal consistency error: composed event failed validation: %w", err)
	}

	return eventObject, nil
}
