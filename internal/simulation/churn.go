package simulation

// EffectiveProbs applies the churn multiplier to baseline success probabilities when AboutToChurn is true.
// Each probability is multiplied by churnMult then clamped to [0, 1].
// dst must have length >= len(src); src is not modified.
func EffectiveProbs(dst, src []float64, churnMult float64, churn bool) []float64 {
	if !churn || churnMult <= 0 {
		n := copy(dst, src)
		return dst[:n]
	}
	out := dst[:len(src)]
	for i := range src {
		v := src[i] * churnMult
		if v > 1 {
			v = 1
		}
		if v < 0 {
			v = 0
		}
		out[i] = v
	}
	return out
}

// Clamp01 clamps x to [0,1]. Exported for tests that validate churn behavior explicitly.
func Clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// MulClamp multiplies p by m and clamps to [0,1].
func MulClamp(p, m float64) float64 {
	return Clamp01(p * m)
}

// EffectiveProbabilitiesSlice returns a new slice (allocate) — use EffectiveProbs for zero-alloc paths.
func EffectiveProbabilitiesSlice(src []float64, churnMult float64, churn bool) []float64 {
	if !churn {
		out := make([]float64, len(src))
		copy(out, src)
		return out
	}
	out := make([]float64, len(src))
	for i := range src {
		out[i] = MulClamp(src[i], churnMult)
	}
	return out
}
