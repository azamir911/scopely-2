package run

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestRun_Simulation_SmallEndToEndDeterministic exercises the full pipeline (config + input + run + output)
// with temp files, fixed seed, and workers=1 so totals are stable for submission review.
func TestRun_Simulation_SmallEndToEndDeterministic(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input_table.csv")
	configPath := filepath.Join(dir, "config_table.csv")
	outputPath := filepath.Join(dir, "output_results.csv")

	inputContent := "user_id,rolls_sink,avg_multiplier,about_to_churn\n" +
		"p1,100,10,false\n" +
		"p2,35,10,false\n"
	// floor(100/10)=10, floor(35/10)=3 -> 13 interactions total
	if err := os.WriteFile(inputPath, []byte(inputContent), 0o600); err != nil {
		t.Fatal(err)
	}

	configContent := "Input,Value\n" +
		"max_successes,2\n" +
		"p_success_1,50%\n" +
		"p_success_2,50%\n" +
		"points_success_1,1\n" +
		"points_success_2,2\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatal(err)
	}

	seed := uint64(42_4242)
	opts := Options{
		InputPath:       inputPath,
		ConfigPath:      configPath,
		OutputPath:      outputPath,
		Threshold:       1e12,
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
		t.Fatalf("read output: %v", err)
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse output CSV: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want header + 1 data row, got %d rows", len(rows))
	}

	wantHeader := []string{
		"total_roll_interactions",
		"success_0_count", "success_1_count", "success_2_count",
		"total_points", "players_above_threshold",
	}
	if len(rows[0]) != len(wantHeader) {
		t.Fatalf("header len got %d want %d: %v", len(rows[0]), len(wantHeader), rows[0])
	}
	for i, w := range wantHeader {
		if rows[0][i] != w {
			t.Fatalf("header col %d: want %q got %q", i, w, rows[0][i])
		}
	}

	totalInteractions, err := strconv.ParseInt(rows[1][0], 10, 64)
	if err != nil {
		t.Fatalf("total_roll_interactions: %v", err)
	}
	if totalInteractions != 13 {
		t.Fatalf("total_roll_interactions want 13 got %d", totalInteractions)
	}

	totalPoints, err := strconv.ParseFloat(rows[1][len(rows[1])-2], 64)
	if err != nil {
		t.Fatalf("total_points: %v", err)
	}
	if totalPoints < 0 {
		t.Fatalf("total_points must be non-negative, got %v", totalPoints)
	}

	// success bucket columns exist and parse as integers
	for i := 1; i <= 3; i++ {
		if _, err := strconv.ParseInt(rows[1][i], 10, 64); err != nil {
			t.Fatalf("success bucket col %d: %v", i, err)
		}
	}
}
