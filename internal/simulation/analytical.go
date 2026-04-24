package simulation

// AnalyticalDistribution returns P(exactly k successes in one interaction) for k = 0..max,
// using the sequential stop-on-failure model with per-depth probabilities probs.
func AnalyticalDistribution(probs []float64) []float64 {
	n := len(probs)
	out := make([]float64, n+1)
	if n == 0 {
		out[0] = 1
		return out
	}

	prefix := 1.0
	for k := 0; k < n; k++ {
		// Exactly k successes: succeed 0..k-1, fail at k.
		fail := 1 - probs[k]
		if fail < 0 {
			fail = 0
		}
		if fail > 1 {
			fail = 1
		}
		out[k] = prefix * fail
		prefix *= probs[k]
	}
	out[n] = prefix
	return out
}

// ExpectedPointsOneInteraction computes E[points] for one draw using distribution and cumulative point weights.
func ExpectedPointsOneInteraction(dist []float64, points []float64) float64 {
	var exp float64
	for k := 0; k <= len(points); k++ {
		pk := 0.0
		if k < len(dist) {
			pk = dist[k]
		}
		exp += pk * PointsForSuccesses(k, points)
	}
	return exp
}

// PointsForSuccesses returns sum(points[0:k]) for k successes (k may be 0..len(points)).
func PointsForSuccesses(successes int, points []float64) float64 {
	if successes <= 0 {
		return 0
	}
	up := successes
	if up > len(points) {
		up = len(points)
	}
	var s float64
	for i := 0; i < up; i++ {
		s += points[i]
	}
	return s
}
