// Package config loads and validates simulation configuration from CSV (Input/Value or key/value rows).
package config

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"coinflip-sim/internal/domain"
)

// Load reads two-column CSV. Canonical assignment format uses headers Input and Value (case-insensitive,
// whitespace-tolerant). Legacy headers key,value are still accepted.
//
// Assignment keys are 1-based in the file (p_success_1 … p_success_N, points_success_1 … points_success_N).
// They are mapped to the internal 0-based slices PSuccess[i] / Points[i] for engine depth i (flip index i).
func Load(r io.Reader) (domain.SimulationConfig, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true

	header, err := cr.Read()
	if err != nil {
		return domain.SimulationConfig{}, fmt.Errorf("config CSV: header: %w", err)
	}
	keyIdx, valIdx, err := findColumns(header)
	if err != nil {
		return domain.SimulationConfig{}, err
	}

	kv := make(map[string]string)
	firstRow := make(map[string]int)

	rowNum := 2 // 1-based line number after header (for reviewer-facing errors)
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return domain.SimulationConfig{}, fmt.Errorf("config CSV row %d: %w", rowNum, err)
		}
		if len(rec) <= keyIdx || len(rec) <= valIdx {
			rowNum++
			continue
		}
		rawKey := strings.TrimSpace(rec[keyIdx])
		if rawKey == "" {
			rowNum++
			continue
		}
		key := strings.ToLower(rawKey)
		val := strings.TrimSpace(rec[valIdx])
		if prev, exists := kv[key]; exists {
			return domain.SimulationConfig{}, fmt.Errorf(
				"config CSV row %d: duplicate key %q (previous value %q at row %d)",
				rowNum, rawKey, prev, firstRow[key])
		}
		kv[key] = val
		firstRow[key] = rowNum
		rowNum++
	}

	maxStr, ok := kv["max_successes"]
	if !ok || strings.TrimSpace(maxStr) == "" {
		return domain.SimulationConfig{}, fmt.Errorf("config CSV: missing required key max_successes")
	}
	maxN, err := strconv.Atoi(strings.TrimSpace(maxStr))
	if err != nil || maxN <= 0 {
		return domain.SimulationConfig{}, fmt.Errorf("max_successes must be a positive integer (got %q)", maxStr)
	}

	assignment := isAssignmentStyle(kv)
	var ps, pt []float64
	if assignment {
		ps = make([]float64, maxN)
		pt = make([]float64, maxN)
		for j := 1; j <= maxN; j++ {
			pKey := fmt.Sprintf("p_success_%d", j)
			ptKey := fmt.Sprintf("points_success_%d", j)
			pStr, okp := kv[pKey]
			ptStr, okpt := kv[ptKey]
			if !okp {
				return domain.SimulationConfig{}, fmt.Errorf("config CSV: missing required key %s (need p_success_1..p_success_%d for assignment format)", pKey, maxN)
			}
			if !okpt {
				return domain.SimulationConfig{}, fmt.Errorf("config CSV: missing required key %s (need points_success_1..points_success_%d for assignment format)", ptKey, maxN)
			}
			p, err := parseProbability(pStr, pKey)
			if err != nil {
				return domain.SimulationConfig{}, err
			}
			ptv, err := parsePointsValue(ptStr, ptKey)
			if err != nil {
				return domain.SimulationConfig{}, err
			}
			ps[j-1] = p
			pt[j-1] = ptv
		}
	} else {
		ps = make([]float64, maxN)
		pt = make([]float64, maxN)
		for i := 0; i < maxN; i++ {
			pKey := fmt.Sprintf("p_success_%d", i)
			ptKey := fmt.Sprintf("points_%d", i)
			pStr, okp := kv[pKey]
			ptStr, okpt := kv[ptKey]
			if !okp {
				return domain.SimulationConfig{}, fmt.Errorf("config CSV: missing required key %s (legacy format needs p_success_0..p_success_%d)", pKey, maxN-1)
			}
			if !okpt {
				return domain.SimulationConfig{}, fmt.Errorf("config CSV: missing required key %s (legacy format needs points_0..points_%d)", ptKey, maxN-1)
			}
			p, err := parseProbability(pStr, pKey)
			if err != nil {
				return domain.SimulationConfig{}, err
			}
			ptv, err := parsePointsValue(ptStr, ptKey)
			if err != nil {
				return domain.SimulationConfig{}, err
			}
			ps[i] = p
			pt[i] = ptv
		}
	}

	cfg := domain.SimulationConfig{MaxSuccesses: maxN, PSuccess: ps, Points: pt}
	if err := cfg.Validate(); err != nil {
		return domain.SimulationConfig{}, err
	}
	return cfg, nil
}

// isAssignmentStyle prefers the assignment format when points_success_* is used; otherwise legacy if
// p_success_0 or points_0 appears; else default to assignment (1-based keys without legacy anchors).
func isAssignmentStyle(kv map[string]string) bool {
	if _, ok := kv["points_success_1"]; ok {
		return true
	}
	if _, ok := kv["p_success_0"]; ok {
		return false
	}
	if _, ok := kv["points_0"]; ok {
		return false
	}
	return true
}

func findColumns(header []string) (keyIdx, valIdx int, err error) {
	keyIdx, valIdx = -1, -1
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "input", "key":
			keyIdx = i
		case "value":
			valIdx = i
		}
	}
	if keyIdx < 0 || valIdx < 0 {
		return -1, -1, fmt.Errorf("config CSV header must include input/key and value columns (got %v)", header)
	}
	return keyIdx, valIdx, nil
}

func parseProbability(s, key string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("config key %q: empty probability", key)
	}
	if strings.HasSuffix(s, "%") {
		numStr := strings.TrimSpace(strings.TrimSuffix(s, "%"))
		x, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, fmt.Errorf("config key %q: invalid percent %q: %w", key, s, err)
		}
		if x < 0 || x > 100 {
			return 0, fmt.Errorf("config key %q: percent must be between 0%% and 100%% (got %q)", key, s)
		}
		return x / 100.0, nil
	}
	x, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("config key %q: invalid probability %q: %w", key, s, err)
	}
	if x < 0 || x > 1 {
		return 0, fmt.Errorf("config key %q: decimal probability must be between 0 and 1 (got %v)", key, x)
	}
	return x, nil
}

func parsePointsValue(s, key string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("config key %q: empty points value", key)
	}
	x, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("config key %q: invalid number %q: %w", key, s, err)
	}
	if x < 0 {
		return 0, fmt.Errorf("config key %q: points must be non-negative (got %v)", key, x)
	}
	return x, nil
}
