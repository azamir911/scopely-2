package simulation

import (
	"math"
	"testing"
)

func TestEffectiveProbs_ChurnBoostAndClamp(t *testing.T) {
	t.Parallel()

	src := []float64{0.5, 0.8, 0.99}
	dst := make([]float64, len(src))

	t.Run("no churn copies baseline", func(t *testing.T) {
		got := EffectiveProbs(dst, src, 1.3, false)
		for i := range src {
			if got[i] != src[i] {
				t.Fatalf("idx %d: want %v got %v", i, src[i], got[i])
			}
		}
	})

	t.Run("churn multiplies and clamps to 1", func(t *testing.T) {
		got := EffectiveProbs(dst, src, 1.3, true)
		want := []float64{0.65, 1.0, 1.0}
		for i := range want {
			if math.Abs(got[i]-want[i]) > 1e-9 {
				t.Fatalf("idx %d: want %v got %v", i, want[i], got[i])
			}
		}
	})
}
