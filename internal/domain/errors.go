package domain

import "errors"

var (
	ErrInvalidMaxSuccesses  = errors.New("max_successes must be non-negative")
	ErrConfigLengthMismatch = errors.New("p_success and points lengths must equal max_successes")
	ErrNonPositiveAvgMult   = errors.New("avg_multiplier must be positive")
	ErrInvalidProbability   = errors.New("probability must be in [0,1]")
	ErrMissingConfig        = errors.New("required configuration is missing")
	ErrNegativeRollsSink    = errors.New("rolls_sink must be non-negative")
	ErrInvalidInteractions  = errors.New("computed interactions must be finite and non-negative")
)
