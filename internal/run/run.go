// Package run orchestrates config load, CSV input, simulation or analytical engine, and output.
// Business rules and RNG live in domain/simulation/aggregation; this package only wires I/O and control flow.
package run

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	mrand "math/rand/v2"

	"coinflip-sim/internal/aggregation"
	"coinflip-sim/internal/config"
	"coinflip-sim/internal/domain"
	"coinflip-sim/internal/input"
	"coinflip-sim/internal/output"
	"coinflip-sim/internal/simulation"
)

// Run executes the pipeline for the given options.
func Run(opts Options) error {
	cfgFile, err := os.Open(opts.ConfigPath)
	if err != nil {
		return err
	}
	defer cfgFile.Close()

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	inFile, err := os.Open(opts.InputPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(opts.OutputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	var ag domain.Aggregates
	ag.SuccessCounts = aggregation.NewBuckets(cfg.MaxSuccesses)

	if opts.Analytical {
		err = analyticalPass(inFile, cfg, opts, &ag)
	} else {
		err = simulatePass(inFile, cfg, opts, &ag)
	}
	if err != nil {
		return err
	}

	return output.WriteAggregates(outFile, cfg.MaxSuccesses, ag)
}

func simulatePass(in *os.File, cfg domain.SimulationConfig, opts Options, ag *domain.Aggregates) error {
	if opts.Workers <= 1 {
		rng := newRNG(opts.Seed)
		scratch := make([]float64, cfg.MaxSuccesses)
		return input.ReadPlayers(in, func(p domain.Player) error {
			simulatePlayer(rng, cfg, scratch, opts.ChurnMultiplier, opts.Threshold, ag, p)
			return nil
		})
	}
	return simulateParallel(in, cfg, opts, ag)
}

func simulatePlayer(
	rng *mrand.Rand,
	cfg domain.SimulationConfig,
	scratch []float64,
	churnMult float64,
	threshold float64,
	ag *domain.Aggregates,
	p domain.Player,
) {
	n := aggregation.FloorInteractions(p.Interactions())
	var playerPts float64
	for i := int64(0); i < n; i++ {
		out := simulation.SimulateOneInteractionChurn(rng, cfg, scratch, churnMult, p.AboutToChurn && p.HasChurnColumn)
		// Simulation returns depth-only points; economy rules apply avg_multiplier at orchestration so the engine stays pure.
		awarded := out.Points * p.AvgMultiplier
		aggregation.ApplyInteraction(ag, out.Successes, awarded)
		playerPts += awarded
	}
	aggregation.ApplyPlayerThreshold(ag, playerPts, threshold)
}

func newRNG(seed *uint64) *mrand.Rand {
	if seed != nil {
		s := *seed
		return mrand.New(mrand.NewPCG(s, s^0x9e3779b97f4a7c15))
	}
	var b [16]byte
	if _, err := crand.Read(b[:]); err != nil {
		panic(err)
	}
	s1 := binary.LittleEndian.Uint64(b[:8])
	s2 := binary.LittleEndian.Uint64(b[8:])
	return mrand.New(mrand.NewPCG(s1, s2))
}

type playerJob struct {
	seq    uint64 // stable order from input CSV (1-based)
	player domain.Player
}

func simulateParallel(in *os.File, cfg domain.SimulationConfig, opts Options, ag *domain.Aggregates) error {
	w := opts.Workers
	if w < 2 {
		w = 2
	}

	jobs := make(chan playerJob, w*4)
	var readErr atomic.Pointer[error]

	go func() {
		defer close(jobs)
		var seq uint64
		err := input.ReadPlayers(in, func(p domain.Player) error {
			seq++
			jobs <- playerJob{seq: seq, player: p}
			return nil
		})
		if err != nil {
			readErr.Store(&err)
		}
	}()

	baseSeed := uint64(0x123456789abcdef0)
	if opts.Seed != nil {
		baseSeed = *opts.Seed
	} else {
		var b [8]byte
		if _, err := crand.Read(b[:]); err != nil {
			return err
		}
		baseSeed = binary.LittleEndian.Uint64(b[:])
	}

	var wg sync.WaitGroup
	partials := make(chan domain.Aggregates, w)
	wg.Add(w)
	for wid := 0; wid < w; wid++ {
		go func() {
			defer wg.Done()
			scratch := make([]float64, cfg.MaxSuccesses)
			var local domain.Aggregates
			local.SuccessCounts = aggregation.NewBuckets(cfg.MaxSuccesses)
			for j := range jobs {
				// Deterministic per-player stream: reproducible regardless of worker scheduling order.
				s := baseSeed ^ (j.seq * 0x9e3779b97f4a7c15)
				rng := mrand.New(mrand.NewPCG(s, s^0xdeadbeefbeefdead))
				simulatePlayer(rng, cfg, scratch, opts.ChurnMultiplier, opts.Threshold, &local, j.player)
			}
			partials <- local
		}()
	}

	go func() {
		wg.Wait()
		close(partials)
	}()

	for partial := range partials {
		aggregation.Merge(ag, partial)
	}

	if errp := readErr.Load(); errp != nil {
		return *errp
	}
	return nil
}

func analyticalPass(in *os.File, cfg domain.SimulationConfig, opts Options, ag *domain.Aggregates) error {
	cm := opts.ChurnMultiplier
	return input.ReadPlayers(in, func(p domain.Player) error {
		churn := p.AboutToChurn && p.HasChurnColumn
		probs := simulation.EffectiveProbabilitiesSlice(cfg.PSuccess, cm, churn)
		dist := simulation.AnalyticalDistribution(probs)
		// Expected depth points for one interaction, scaled by multiplier — mirrors simulatePlayer's awarded := out.Points * p.AvgMultiplier.
		expPer := simulation.ExpectedPointsOneInteraction(dist, cfg.Points) * p.AvgMultiplier
		aggregation.AnalyticalPlayer(ag, p, dist, expPer, opts.Threshold)
		return nil
	})
}
