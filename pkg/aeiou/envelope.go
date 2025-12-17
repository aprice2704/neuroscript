// :: product: FDM/NS
// :: majorVersion: 0
// :: fileVersion: 6
// :: description: Updated comments to reflect AEIOU V4.
// :: latestChange: Updated comments from V3 to V4.
// :: filename: pkg/aeiou/envelope.go
// :: serialization: go
package aeiou

import (
	"errors"
	"strings"
)

// --- AEIOU Envelope Sentinel Errors ---
var (
	// ErrMarkerInvalid indicates a malformed or unrecognizable envelope marker.
	ErrMarkerInvalid = errors.New("invalid envelope marker")
	// ErrSectionOrder indicates sections are not in the canonical order.
	ErrSectionOrder = errors.New("envelope sections out of order")
	// ErrSectionMissing indicates a required section is missing.
	ErrSectionMissing = errors.New("envelope missing required section")
	// ErrDuplicateSection indicates a section appears more than once.
	ErrDuplicateSection = errors.New("envelope duplicate section")
	// ErrPayloadTooLarge indicates the envelope or a section exceeds size limits.
	ErrPayloadTooLarge = errors.New("envelope or section exceeds size limits")
	// ErrMaxRecursionDepth indicates JSON nesting exceeds the maximum recursion depth.
	ErrMaxRecursionDepth = errors.New("json nesting exceeds max recursion depth")
	// ErrUserDataSchema indicates the USERDATA payload does not conform to the required schema.
	ErrUserDataSchema = errors.New("userdata does not conform to schema")
)

// --- V3 Token Sentinel Errors (Legacy support) ---
var (
	// ErrTokenInvalid indicates a malformed or unparseable token string.
	ErrTokenInvalid = errors.New("token is malformed")
	// ErrTokenSignature indicates the token's signature is invalid.
	ErrTokenSignature = errors.New("token signature verification failed")
	// ErrTokenScope indicates the token is not valid for the current context (SID, turn, nonce).
	ErrTokenScope = errors.New("token scope mismatch")
	// ErrTokenExpired indicates the token's TTL has passed.
	ErrTokenExpired = errors.New("token has expired")
	// ErrTokenReplay indicates the token's JTI has already been seen.
	ErrTokenReplay = errors.New("token replay detected")
	// ErrTokenUnknownKID indicates the key ID in the token is not recognized.
	ErrTokenUnknownKID = errors.New("token has unknown key id")
	// ErrMintingFailed indicates both primary and fallback minters failed.
	ErrMintingFailed = errors.New("token minting failed for both primary and fallback signers")
)

// SectionType defines the valid sections in an AEIOU envelope.
type SectionType string

// AEIOU Section Types
const (
	SectionStart      SectionType = "START"
	SectionEnd        SectionType = "END"
	SectionUserData   SectionType = "USERDATA"
	SectionScratchpad SectionType = "SCRATCHPAD"
	SectionOutput     SectionType = "OUTPUT"
	SectionActions    SectionType = "ACTIONS"
)

// Envelope holds the parsed content of the AEIOU sections.
type Envelope struct {
	// UserData is the host-provided JSON object for the program.
	UserData string
	// Scratchpad contains private notes from the prior turn's 'whisper' emissions.
	Scratchpad string
	// Output contains the public log from the prior turn's 'emit' emissions.
	Output string
	// Actions contains the NeuroScript code to be executed this turn.
	Actions string
}

// Compose constructs the full envelope string from an Envelope struct,
// ensuring the canonical section order.
func (e *Envelope) Compose() (string, error) {
	if e.UserData == "" || e.Actions == "" {
		return "", ErrSectionMissing
	}

	var sb strings.Builder

	sb.WriteString(Wrap(SectionStart))
	sb.WriteString("\n")

	// USERDATA (required)
	sb.WriteString(Wrap(SectionUserData))
	sb.WriteString("\n")
	sb.WriteString(e.UserData)
	sb.WriteString("\n")

	// SCRATCHPAD (optional)
	if e.Scratchpad != "" {
		sb.WriteString(Wrap(SectionScratchpad))
		sb.WriteString("\n")
		sb.WriteString(e.Scratchpad)
		sb.WriteString("\n")
	}

	// OUTPUT (optional)
	if e.Output != "" {
		sb.WriteString(Wrap(SectionOutput))
		sb.WriteString("\n")
		sb.WriteString(e.Output)
		sb.WriteString("\n")
	}

	// ACTIONS (required)
	sb.WriteString(Wrap(SectionActions))
	sb.WriteString("\n")
	sb.WriteString(e.Actions)
	sb.WriteString("\n")

	sb.WriteString(Wrap(SectionEnd))
	sb.WriteString("\n")

	return sb.String(), nil
}
