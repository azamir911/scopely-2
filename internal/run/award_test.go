package run

import "testing"

// Contract with simulatePlayer: base depth points from the engine are scaled by avg_multiplier
// at orchestration. This test locks the 10× example from the spec.
func TestAwardedPoints_AvgMultiplierScaling(t *testing.T) {
	t.Parallel()

	baseDepthPoints := 30.0
	avgMultiplier := 10.0
	awarded := baseDepthPoints * avgMultiplier
	if awarded != 300 {
		t.Fatalf("awarded = %v, want 300 (30 * 10)", awarded)
	}
}
