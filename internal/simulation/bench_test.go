package simulation

import (
	"math/rand/v2"
	"testing"

	"coinflip-sim/internal/domain"
)

func BenchmarkSimulateOneInteraction(b *testing.B) {
	cfg := domain.SimulationConfig{
		MaxSuccesses: 5,
		PSuccess:     []float64{0.5, 0.45, 0.4, 0.35, 0.3},
		Points:       []float64{1, 2, 3, 4, 5},
	}
	scratch := make([]float64, len(cfg.PSuccess))
	rng := rand.New(rand.NewPCG(42, 43))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SimulateOneInteractionChurn(rng, cfg, scratch, 1.3, i%2 == 0)
	}
}
