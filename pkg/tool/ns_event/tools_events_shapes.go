// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides canonical shape definitions for ns_event structures.
// filename: pkg/tool/ns_event/tools_event_shapes.go
// nlines: 40
// risk_rating: LOW

package ns_event

import "github.com/aprice2704/neuroscript/pkg/tool"

// eventEnvelopeShape defines the structure of a single event envelope.
var eventEnvelopeShape = map[string]interface{}{
	"ID":       "string",
	"Kind":     "string",
	"AgentID?": "string", // AgentID is optional
	"TS":       "int",    // Timestamp
	"Payload":  "any",    // Payload can be any structure, so we use 'any'
}

// fdmEventShape defines the top-level structure containing event envelopes.
var fdmEventShape = map[string]interface{}{
	// The top-level object has a 'payload' key which is a list of envelopes.
	"payload[]": eventEnvelopeShape,
}

// toolGetEventShape implements the tool.ns_event.GetEventShape function.
func toolGetEventShape(rt tool.Runtime, args []interface{}) (interface{}, error) {
	// Return a copy to prevent modification of the canonical shape definition.
	newMap := make(map[string]interface{})
	for k, v := range fdmEventShape {
		newMap[k] = v
	}
	return newMap, nil
}
