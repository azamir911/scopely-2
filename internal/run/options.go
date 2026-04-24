package run

// Options configures a simulation or analytical run (populated by CLI flags).
type Options struct {
	InputPath       string
	ConfigPath      string
	OutputPath      string
	Threshold       float64
	ChurnMultiplier float64
	Seed            *uint64 // nil = non-deterministic (crypto-random PCG seeds)
	Workers         int     // 0 or 1 = sequential; >1 parallelize per-player
	Analytical      bool    // skip RNG; use closed-form expectations
}
