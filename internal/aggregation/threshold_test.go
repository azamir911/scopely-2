package aggregation

import (
	"testing"

	"coinflip-sim/internal/domain"
)

// Verifies product policy: only strict '>' counts (not >=), so exact threshold is excluded.
func TestApplyPlayerThreshold_StrictComparison(t *testing.T) {
	t.Parallel()

	var a domain.Aggregates
	a.SuccessCounts = NewBuckets(0)

	ApplyPlayerThreshold(&a, 100, 100)
	if a.PlayersAboveThreshold != 0 {
		t.Fatalf("at threshold: should not count, got %d", a.PlayersAboveThreshold)
	}

	ApplyPlayerThreshold(&a, 100.0001, 100)
	if a.PlayersAboveThreshold != 1 {
		t.Fatalf("just above threshold: want 1, got %d", a.PlayersAboveThreshold)
	}
}
