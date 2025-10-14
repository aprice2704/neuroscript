// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Corrected grant satisfaction logic to be case-insensitive by reusing the canonical implementation in matcher.go.
// filename: pkg/capability/check.go
// nlines: 15
// risk_rating: HIGH

package capability

// Check determines if the GrantSet satisfies a required capability by calling
// the canonical, case-insensitive matching logic. This ensures all capability
// checks in the system behave consistently.
func (gs *GrantSet) Check(required Capability) bool {
	// Re-use the correct, case-insensitive implementation from matcher.go.
	// This consolidates the logic and fixes the case-sensitivity bug system-wide.
	return CapsSatisfied([]Capability{required}, gs.Grants)
}
