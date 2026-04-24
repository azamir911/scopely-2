// Package output writes aggregated simulation results to CSV.
package output

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"coinflip-sim/internal/domain"
)

// WriteAggregates writes one header row and one data row.
func WriteAggregates(w io.Writer, maxSuccesses int, a domain.Aggregates) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := make([]string, 0, 4+maxSuccesses+1)
	header = append(header, "total_roll_interactions")
	for k := 0; k <= maxSuccesses; k++ {
		header = append(header, fmt.Sprintf("success_%d_count", k))
	}
	header = append(header, "total_points", "players_above_threshold")

	if err := cw.Write(header); err != nil {
		return err
	}

	row := make([]string, 0, len(header))
	row = append(row, strconv.FormatInt(a.TotalRollInteractions, 10))
	for k := 0; k <= maxSuccesses; k++ {
		var v int64
		if k < len(a.SuccessCounts) {
			v = a.SuccessCounts[k]
		}
		row = append(row, strconv.FormatInt(v, 10))
	}
	row = append(row,
		strconv.FormatFloat(a.TotalPoints, 'f', -1, 64),
		strconv.FormatInt(a.PlayersAboveThreshold, 10),
	)
	return cw.Write(row)
}
