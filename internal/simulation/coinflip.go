package simulation

import (
	"math/rand/v2"

	"coinflip-sim/internal/domain"
)

// InteractionOutcome is the result of one interaction: how many sequential successes before first failure,
// and total points awarded for those successes.
type InteractionOutcome struct {
	Successes int     // exact count in [0, MaxSuccesses]
	Points    float64 // sum of point awards for each successful flip
}

// SimulateOneInteraction runs the sequential flip model using effProbs (already churn-adjusted if needed).
// Stops at the first failed flip; if all flips succeed, Successes == len(effProbs).
// effProbs and cfg.Points must have equal length (= max successes / number of flip stages).
func SimulateOneInteraction(rng *rand.Rand, cfg domain.SimulationConfig, effProbs []float64) InteractionOutcome {
	var successes int
	var pts float64
	n := len(effProbs)
	for i := 0; i < n; i++ {
		if rng.Float64() < effProbs[i] {
			successes++
			pts += cfg.Points[i]
			continue
		}
		break
	}
	return InteractionOutcome{Successes: successes, Points: pts}
}

// SimulateOneInteractionChurn combines probability adjustment and one play. Reuses scratch for eff probs.
func SimulateOneInteractionChurn(rng *rand.Rand, cfg domain.SimulationConfig, scratch []float64, churnMult float64, churn bool) InteractionOutcome {
	eff := EffectiveProbs(scratch, cfg.PSuccess, churnMult, churn)
	return SimulateOneInteraction(rng, cfg, eff)
}
