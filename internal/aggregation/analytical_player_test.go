package aggregation

import (
	"testing"

	"coinflip-sim/internal/domain"
)

// Locks analytical mode to the same interaction count rule as simulation (floor, not fractional n).
func TestAnalyticalPlayer_UsesFloorInteractions(t *testing.T) {
	t.Parallel()

	var a domain.Aggregates
	a.SuccessCounts = NewBuckets(1)
	p := domain.Player{RollsSink: 109, AvgMultiplier: 10, UserID: "x"} // 10.9 -> floor 10

	dist := []float64{1.0, 0.0} // always 0 successes in one interaction
	pointsPer := 5.0

	AnalyticalPlayer(&a, p, dist, pointsPer, 0)

	if a.TotalRollInteractions != 10 {
		t.Fatalf("TotalRollInteractions=%d want 10 (floor of 10.9)", a.TotalRollInteractions)
	}
	if a.TotalPoints != 50 {
		t.Fatalf("TotalPoints=%v want 50 (10 * 5)", a.TotalPoints)
	}
}
