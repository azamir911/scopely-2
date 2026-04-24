// Package input streams player rows from CSV without loading the full file into memory.
package input

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"coinflip-sim/internal/domain"
)

// Column indices resolved from header.
type headerMap struct {
	userID        int
	rollsSink     int
	avgMultiplier int
	churn         int // -1 if absent
}

// ParseHeader maps required columns; about_to_churn is optional.
func ParseHeader(header []string) (headerMap, error) {
	if len(header) == 0 {
		return headerMap{}, fmt.Errorf("empty header")
	}
	var hm headerMap
	hm.userID = -1
	hm.rollsSink = -1
	hm.avgMultiplier = -1
	hm.churn = -1
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "user_id":
			hm.userID = i
		case "rolls_sink":
			hm.rollsSink = i
		case "avg_multiplier":
			hm.avgMultiplier = i
		case "about_to_churn":
			hm.churn = i
		}
	}
	if hm.userID < 0 || hm.rollsSink < 0 || hm.avgMultiplier < 0 {
		return headerMap{}, fmt.Errorf("header must include user_id, rolls_sink, avg_multiplier")
	}
	if err := hm.validate(len(header)); err != nil {
		return headerMap{}, err
	}
	return hm, nil
}

func (hm headerMap) validate(nCol int) error {
	max := hm.userID
	if hm.rollsSink > max {
		max = hm.rollsSink
	}
	if hm.avgMultiplier > max {
		max = hm.avgMultiplier
	}
	if hm.churn >= 0 && hm.churn > max {
		max = hm.churn
	}
	if max >= nCol {
		return fmt.Errorf("header references missing columns")
	}
	return nil
}

// ReadPlayers streaming: invokes cb for each parsed player. cb returns error to abort.
func ReadPlayers(r io.Reader, cb func(domain.Player) error) error {
	cr := csv.NewReader(r)
	cr.ReuseRecord = true

	header, err := cr.Read()
	if err != nil {
		return err
	}
	hm, err := ParseHeader(header)
	if err != nil {
		return err
	}

	for {
		rec, err := cr.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		p, err := parseRow(hm, rec)
		if err != nil {
			return err
		}
		if err := cb(p); err != nil {
			return err
		}
	}
}

func parseRow(hm headerMap, rec []string) (domain.Player, error) {
	get := func(i int) string {
		if i < len(rec) {
			return strings.TrimSpace(rec[i])
		}
		return ""
	}

	rs, err := strconv.ParseFloat(get(hm.rollsSink), 64)
	if err != nil || rs < 0 {
		return domain.Player{}, fmt.Errorf("rolls_sink: %w", domain.ErrNegativeRollsSink)
	}
	am, err := strconv.ParseFloat(get(hm.avgMultiplier), 64)
	if err != nil || am <= 0 {
		return domain.Player{}, fmt.Errorf("avg_multiplier: %w", domain.ErrNonPositiveAvgMult)
	}

	var churn bool
	hasChurn := hm.churn >= 0 && hm.churn < len(rec)
	if hasChurn {
		v := strings.ToLower(get(hm.churn))
		churn = v == "true" || v == "1" || v == "yes"
	}

	return domain.Player{
		UserID:         get(hm.userID),
		RollsSink:      rs,
		AvgMultiplier:  am,
		AboutToChurn:   churn,
		HasChurnColumn: hasChurn,
	}, nil
}
