// Package domain holds core types for the coin-flip economy simulation.
// It has no I/O or framework dependencies.
package domain

// Player is one row from the input CSV.
type Player struct {
	UserID         string
	RollsSink      float64
	AvgMultiplier  float64
	AboutToChurn   bool
	HasChurnColumn bool // false if column was omitted (treat as not churning)
}

// Interactions returns rolls_sink / avg_multiplier per product rules.
// Caller must ensure AvgMultiplier > 0 (validated at input boundary).
// The count of runs per player is floor(Interactions())—see aggregation.FloorInteractions.
func (p Player) Interactions() float64 {
	return p.RollsSink / p.AvgMultiplier
}

// SimulationConfig drives flip probabilities, depth cap, and point awards.
type SimulationConfig struct {
	MaxSuccesses int       // maximum sequential flips attempted (depth cap)
	PSuccess     []float64 // p_success_n for n = 0 .. MaxSuccesses-1
	Points       []float64 // per-depth base awards; final points per interaction also scale by the player’s avg_multiplier in run/aggregation
}

// Validate checks structural consistency. Probabilities should already be in [0,1].
func (c SimulationConfig) Validate() error {
	if c.MaxSuccesses < 0 {
		return ErrInvalidMaxSuccesses
	}
	if len(c.PSuccess) != c.MaxSuccesses || len(c.Points) != c.MaxSuccesses {
		return ErrConfigLengthMismatch
	}
	return nil
}

// Aggregates are global counters after processing all players (or one analytical pass).
type Aggregates struct {
	TotalRollInteractions int64
	SuccessCounts         []int64 // index i = interactions ending with exactly i successes (i from 0 to MaxSuccesses)
	TotalPoints           float64
	PlayersAboveThreshold int64
}
