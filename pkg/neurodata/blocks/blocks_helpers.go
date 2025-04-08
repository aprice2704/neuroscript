package blocks

import (
	"reflect"
	"testing"
)

var fixtureDir string = "test_fixtures"

// --- ADDED HELPER FUNCTIONS ---

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func compareBlockSlices(t *testing.T, got, want []FencedBlock, sourceName string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Mismatch in extracted blocks from %s", sourceName)
		maxLen := minInt(len(got), len(want)) // Use minInt helper
		for i := 0; i < maxLen; i++ {
			if !reflect.DeepEqual(got[i], want[i]) {
				t.Errorf("--- Block %d Mismatch ---", i)
				t.Errorf("  Got : LangID=%q, Start=%d, End=%d, Content=\n---\n%s\n---", got[i].LanguageID, got[i].StartLine, got[i].EndLine, got[i].RawContent)
				t.Errorf("  Want: LangID=%q, Start=%d, End=%d, Content=\n---\n%s\n---", want[i].LanguageID, want[i].StartLine, want[i].EndLine, want[i].RawContent)
			}
		}
		if len(got) > maxLen {
			t.Errorf("--- Extra blocks extracted: ---")
			for i := maxLen; i < len(got); i++ {
				t.Errorf("  Index %d: LangID=%q, Start=%d, End=%d, Content=\n---\n%s\n---", i, got[i].LanguageID, got[i].StartLine, got[i].EndLine, got[i].RawContent)
			}
		}
		if len(want) > maxLen {
			t.Errorf("--- Expected blocks missing: ---")
			for i := maxLen; i < len(want); i++ {
				t.Errorf("  Index %d: LangID=%q, Start=%d, End=%d, Content=\n---\n%s\n---", i, want[i].LanguageID, want[i].StartLine, want[i].EndLine, want[i].RawContent)
			}
		}
		// Use Logf instead of Errorf for the full dump
		t.Logf("\nFull Got Blocks:\n%#v\n", got)
		t.Logf("\nFull Want Blocks:\n%#v\n", want)
		t.Errorf("Block comparison failed (details above).") // Ensure test fails
	} else {
		t.Logf("Blocks extracted from %s match expected blocks.", sourceName)
	}
}

// --- END ADDED HELPER FUNCTIONS ---
