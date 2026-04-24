package run

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestRun_Simulation_ZeroFloorInteractions_NoPointsNoThresholdCount ensures rolls_sink < avg_multiplier
// yields floor 0 interactions: no rolls, no points, and no player above any non-negative threshold.
func TestRun_Simulation_ZeroFloorInteractions_NoPointsNoThresholdCount(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input_table.csv")
	configPath := filepath.Join(dir, "config_table.csv")
	outputPath := filepath.Join(dir, "output_results.csv")

	inputContent := "user_id,rolls_sink,avg_multiplier,about_to_churn\n" +
		"lonely,3,10,false\n"
	if err := os.WriteFile(inputPath, []byte(inputContent), 0o600); err != nil {
		t.Fatal(err)
	}

	configContent := "Input,Value\n" +
		"max_successes,1\n" +
		"p_success_1,50%\n" +
		"points_success_1,100\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatal(err)
	}

	seed := uint64(1)
	opts := Options{
		InputPath:       inputPath,
		ConfigPath:      configPath,
		OutputPath:      outputPath,
		Threshold:       0,
		ChurnMultiplier: 1.3,
		Seed:            &seed,
		Workers:         1,
		Analytical:      false,
	}
	if err := Run(opts); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil || len(rows) < 2 {
		t.Fatalf("parse output: rows=%v err=%v", rows, err)
	}
	row := rows[1]

	totalInteractions, _ := strconv.ParseInt(row[0], 10, 64)
	totalPoints, _ := strconv.ParseFloat(row[len(row)-2], 64)
	above, _ := strconv.ParseInt(row[len(row)-1], 10, 64)

	if totalInteractions != 0 {
		t.Fatalf("total_roll_interactions want 0 got %d", totalInteractions)
	}
	if totalPoints != 0 {
		t.Fatalf("total_points want 0 got %v", totalPoints)
	}
	if above != 0 {
		t.Fatalf("players_above_threshold want 0 (strict > and zero points) got %d", above)
	}
}
