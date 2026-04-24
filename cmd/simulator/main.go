// Command simulator runs the coin-flip economy aggregation pipeline (CLI flags only — no domain rules).
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"coinflip-sim/internal/run"
)

type optionalSeed struct {
	set bool
	val uint64
}

func (o *optionalSeed) String() string {
	if !o.set {
		return ""
	}
	return strconv.FormatUint(o.val, 10)
}

func (o *optionalSeed) Set(s string) error {
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	o.val = u
	o.set = true
	return nil
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	inputPath := fs.String("input", "input_table.csv", "path to players CSV")
	configPath := fs.String("config", "config_table.csv", "path to config key/value CSV")
	outputPath := fs.String("output", "output_results.csv", "path to write aggregated CSV")
	threshold := fs.Float64("threshold", 0, "player-level total points threshold for counting players_above_threshold")
	churnMult := fs.Float64("churn-multiplier", 1.3, "multiplier applied to success probabilities when about_to_churn is true")
	var seed optionalSeed
	fs.Var(&seed, "seed", "RNG seed for deterministic simulation (omit for nondeterministic RNG)")
	workers := fs.Int("workers", 1, "parallel workers for simulation (1 = sequential streaming)")
	analytical := fs.Bool("analytical", false, "analytical (non-simulation) mode using closed-form outcome distributions")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [flags]\n\n", fs.Name())
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}

	opts := run.Options{
		InputPath:       *inputPath,
		ConfigPath:      *configPath,
		OutputPath:      *outputPath,
		Threshold:       *threshold,
		ChurnMultiplier: *churnMult,
		Workers:         *workers,
		Analytical:      *analytical,
	}

	if seed.set {
		opts.Seed = &seed.val
	}

	if err := run.Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "simulator: %v\n", err)
		os.Exit(1)
	}
}
