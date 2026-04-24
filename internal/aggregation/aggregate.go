// Package aggregation merges simulation outcomes into global counters (pure logic).
package aggregation

import (
	"math"

	"coinflip-sim/internal/domain"
)

// NewBuckets allocates success count buckets for outcomes 0..maxSuccesses inclusive.
func NewBuckets(maxSuccesses int) []int64 {
	return make([]int64, maxSuccesses+1)
}

// Merge adds partial aggregates into dst.
func Merge(dst *domain.Aggregates, src domain.Aggregates) {
	dst.TotalRollInteractions += src.TotalRollInteractions
	dst.TotalPoints += src.TotalPoints
	dst.PlayersAboveThreshold += src.PlayersAboveThreshold
	if len(src.SuccessCounts) != len(dst.SuccessCounts) {
		panic("aggregation.Merge: bucket length mismatch")
	}
	for i := range dst.SuccessCounts {
		dst.SuccessCounts[i] += src.SuccessCounts[i]
	}
}

// ApplyInteraction adds one interaction outcome to aggregates (single-threaded helper).
func ApplyInteraction(a *domain.Aggregates, successes int, points float64) {
	if successes < 0 || successes >= len(a.SuccessCounts) {
		panic("ApplyInteraction: successes out of range")
	}
	a.TotalRollInteractions++
	a.SuccessCounts[successes]++
	a.TotalPoints += points
}

// ApplyPlayerThreshold increments PlayersAboveThreshold if condition holds.
func ApplyPlayerThreshold(a *domain.Aggregates, playerTotalPoints float64, threshold float64) {
	if playerTotalPoints > threshold {
		a.PlayersAboveThreshold++
	}
}

// FloorInteractions is the non-simulated interaction count: floor(rolls_sink/avg_multiplier) with
// domain-specific handling of non-finite and non-positive values (treated as zero interactions).
// This must match between simulation and analytical modes so both answer the same “how many plays” question.
func FloorInteractions(v float64) int64 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 {
		return 0
	}
	return int64(math.Floor(v))
}

// AnalyticalPlayer merges one player's analytical contribution without per-interaction RNG.
// pointsPerInteraction must already include the player's avg_multiplier (see internal/run orchestration).
func AnalyticalPlayer(a *domain.Aggregates, p domain.Player, dist []float64, pointsPerInteraction float64, threshold float64) {
	n := FloorInteractions(p.Interactions())
	if n <= 0 {
		return
	}

	a.TotalRollInteractions += n
	// Expected depth points per interaction × multiplier × integer interaction count (aligned with simulation).
	totalPlayerPts := float64(n) * pointsPerInteraction
	a.TotalPoints += totalPlayerPts

	maxK := len(a.SuccessCounts) - 1 // outcome buckets 0..max_successes inclusive
	for k := 0; k <= maxK && k < len(dist); k++ {
		a.SuccessCounts[k] += int64(math.Round(float64(n) * dist[k]))
	}

	if totalPlayerPts > threshold {
		a.PlayersAboveThreshold++
	}
}
