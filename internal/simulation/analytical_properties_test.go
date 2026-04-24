package simulation

import (
	"math"
	"testing"
)

func TestAnalyticalDistribution_SumsToOne(t *testing.T) {
	t.Parallel()

	probs := []float64{0.2, 0.35, 0.9}
	d := AnalyticalDistribution(probs)
	var sum float64
	for _, x := range d {
		sum += x
	}
	if math.Abs(sum-1.0) > 1e-12 {
		t.Fatalf("distribution must sum to 1, got %v (parts=%v)", sum, d)
	}
}
