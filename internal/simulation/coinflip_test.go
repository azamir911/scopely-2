package simulation

import (
	"math/rand/v2"
	"testing"

	"coinflip-sim/internal/domain"
)

func TestSimulateOneInteraction_StopsOnFirstFailure(t *testing.T) {
	t.Parallel()

	// Deterministic RNG: fail the first flip always.
	cfg := domain.SimulationConfig{
		MaxSuccesses: 3,
		PSuccess:     []float64{0, 1, 1},
		Points:       []float64{10, 20, 30},
	}
	rng := rand.New(rand.NewPCG(1, 2))
	eff := []float64{0, 1, 1}

	out := SimulateOneInteraction(rng, cfg, eff)
	if out.Successes != 0 || out.Points != 0 {
		t.Fatalf("expected 0 successes when p0=0, got %+v", out)
	}

	// Succeed depth 0, fail depth 1.
	cfg2 := domain.SimulationConfig{
		MaxSuccesses: 3,
		PSuccess:     []float64{1, 0, 1},
		Points:       []float64{5, 9, 7},
	}
	rng2 := rand.New(rand.NewPCG(3, 4))
	eff2 := []float64{1, 0, 1}
	out2 := SimulateOneInteraction(rng2, cfg2, eff2)
	if out2.Successes != 1 || out2.Points != 5 {
		t.Fatalf("expected exactly one success then stop, got %+v", out2)
	}

	// All successes.
	cfg3 := domain.SimulationConfig{
		MaxSuccesses: 2,
		PSuccess:     []float64{1, 1},
		Points:       []float64{2, 3},
	}
	rng3 := rand.New(rand.NewPCG(5, 6))
	eff3 := []float64{1, 1}
	out3 := SimulateOneInteraction(rng3, cfg3, eff3)
	if out3.Successes != 2 || out3.Points != 5 {
		t.Fatalf("expected full streak, got %+v", out3)
	}
}
