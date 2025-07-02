// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 01:05:00 AM PDT
// filename: pkg/neurodata/checklist/checklist_status.go

package checklist

import (
	"fmt"
	// "github.com/aprice2704/neuroscript/pkg/logging" // Logger might not be needed here if errors are returned
)

// calculateAutomaticStatus determines the parent status based purely on child statuses and symbols.
// It encapsulates Rules 1, 2, 3, 4 from the checklist specification. Rule 5 (no children)
// is handled by the caller (`updateAutomaticNodeStatus`).
// Returns the calculated status string and the special symbol string (if applicable).
func calculateAutomaticStatus(childStatuses []string, childSymbols map[int]string) (string, string, error) {
	calculatedStatus := "open" // Default status (Rule 4 / fallback)
	calculatedSymbol := ""

	if len(childStatuses) == 0 {
		// Should be handled by caller (Rule 5), but return open for safety.
		return "open", "", nil
	}

	// --- Check priorities (Rule 1) ---
	hasBlocked := false
	hasQuestion := false
	hasInProgress := false
	hasSpecial := false
	var firstSpecialSymbol string = "?" // Default symbol if lookup fails

	// Determine highest priority status present among children
	for idx, status := range childStatuses {
		switch status {
		case "blocked":
			hasBlocked = true
		case "question":
			hasQuestion = true
		case "inprogress":
			hasInProgress = true
		case "special":
			if !hasSpecial { // Only capture the *first* special status encountered
				hasSpecial = true
				sym, ok := childSymbols[idx]
				if ok && sym != "" {
					firstSpecialSymbol = sym
				} else {
					// This indicates an issue upstream where a special child status was determined
					// but no symbol was provided in the map. Return an error.
					return "", "", fmt.Errorf("internal inconsistency: child %d has status 'special' but symbol is missing or empty in map %v", idx, childSymbols)
				}
			}
		}
		// Optimization: Break early if highest priority is found
		if hasBlocked {
			break
		}
	}

	// Apply Rule 1 based on highest priority found
	if hasBlocked {
		calculatedStatus = "blocked"
	} else if hasQuestion {
		calculatedStatus = "question"
	} else if hasInProgress {
		calculatedStatus = "inprogress"
	} else if hasSpecial {
		calculatedStatus = "special"
		calculatedSymbol = firstSpecialSymbol
	} else {
		// --- No priority statuses found, apply Rules 2, 3, 4 using counts ---
		doneCount := 0
		skippedCount := 0
		partialCount := 0
		openCount := 0
		otherCount := 0
		totalChildren := len(childStatuses)

		for _, status := range childStatuses {
			switch status {
			case "open":
				openCount++
			case "done":
				doneCount++
			case "skipped":
				skippedCount++
			case "partial":
				partialCount++
			default:
				otherCount++ // Includes potential invalid/unknown statuses from children
			}
		}

		// Validate counts add up (sanity check)
		if openCount+doneCount+skippedCount+partialCount+otherCount != totalChildren {
			return "", "", fmt.Errorf("internal inconsistency: child status counts (%d) do not sum to total children (%d)", openCount+doneCount+skippedCount+partialCount+otherCount, totalChildren)
		}

		// Apply rules based on counts, respecting precedence: Rule 3 -> Rule 2 -> Rule 4
		if doneCount == totalChildren && totalChildren > 0 { // Rule 3: All children are 'done' (and there's at least one child)
			calculatedStatus = "done"
		} else if doneCount > 0 || skippedCount > 0 || partialCount > 0 { // Rule 2: Any done/skipped/partial trigger (and not all were 'done')
			calculatedStatus = "partial"
			// } else if openCount == totalChildren && totalChildren > 0 { // Rule 4: All children are 'open' (explicit check might be redundant due to default)
			// 	calculatedStatus = "open"
		} else if otherCount > 0 {
			// If there are unknown statuses and no other rules apply, what should happen?
			// Let's default to 'open' but maybe log this condition in the calling function.
			calculatedStatus = "open" // Fallback, potentially log warning in caller
		}
		// If none of the above, the default "open" remains (covers Rule 4 implicitly and mixed 'open'/'other')
	}

	return calculatedStatus, calculatedSymbol, nil
}
