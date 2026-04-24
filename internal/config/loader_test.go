package config

import (
	"strings"
	"testing"
)

func TestLoad_AssignmentCanonicalHeader_PercentProbabilities(t *testing.T) {
	raw := `Input,Value
max_successes,2
p_success_1,60%
p_success_2,50%
points_success_1,1
points_success_2,2
`
	cfg, err := Load(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxSuccesses != 2 {
		t.Fatalf("MaxSuccesses=%d", cfg.MaxSuccesses)
	}
	if cfg.PSuccess[0] != 0.6 || cfg.PSuccess[1] != 0.5 {
		t.Fatalf("PSuccess=%v", cfg.PSuccess)
	}
	if cfg.Points[0] != 1 || cfg.Points[1] != 2 {
		t.Fatalf("Points=%v", cfg.Points)
	}
}

func TestLoad_HeaderCaseInsensitiveAndSpaces(t *testing.T) {
	raw := ` Input , Value 
max_successes,1
p_success_1,100%
points_success_1,3
`
	cfg, err := Load(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxSuccesses != 1 || cfg.PSuccess[0] != 1.0 || cfg.Points[0] != 3 {
		t.Fatalf("%+v err=%v", cfg, err)
	}
}

func TestLoad_InputValueLowercase(t *testing.T) {
	raw := `input,value
max_successes,1
p_success_1,0.75
points_success_1,10
`
	cfg, err := Load(strings.NewReader(raw))
	if err != nil || cfg.PSuccess[0] != 0.75 {
		t.Fatalf("%+v %v", cfg, err)
	}
}

func TestLoad_DecimalProbabilities(t *testing.T) {
	raw := `Input,Value
max_successes,2
p_success_1,0.6
p_success_2,0.25
points_success_1,1
points_success_2,2
`
	cfg, err := Load(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PSuccess[0] != 0.6 || cfg.PSuccess[1] != 0.25 {
		t.Fatal(cfg.PSuccess)
	}
}

func TestLoad_TrimmedPercentValues(t *testing.T) {
	raw := `Input,Value
max_successes,1
p_success_1, 60% 
points_success_1, 5 
`
	cfg, err := Load(strings.NewReader(raw))
	if err != nil || cfg.PSuccess[0] != 0.6 || cfg.Points[0] != 5 {
		t.Fatalf("%+v %v", cfg, err)
	}
}

func TestLoad_LegacyKeyValueHeaders(t *testing.T) {
	raw := `key,value
max_successes,2
p_success_0,0.5
p_success_1,0.25
points_0,10
points_1,20
`
	cfg, err := Load(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxSuccesses != 2 || cfg.PSuccess[0] != 0.5 {
		t.Fatal(cfg)
	}
}

func TestLoad_InvalidHeader(t *testing.T) {
	raw := `K1,K2
max_successes,1
`
	_, err := Load(strings.NewReader(raw))
	if err == nil || !strings.Contains(err.Error(), "header") {
		t.Fatalf("want header error, got %v", err)
	}
}

func TestLoad_MissingMaxSuccesses(t *testing.T) {
	raw := `Input,Value
p_success_1,50%
points_success_1,1
`
	_, err := Load(strings.NewReader(raw))
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "max_successes") {
		t.Fatalf("want missing max_successes, got %v", err)
	}
}

func TestLoad_MissingProbabilityStep(t *testing.T) {
	raw := `Input,Value
max_successes,3
p_success_1,50%
p_success_3,50%
points_success_1,1
points_success_2,2
points_success_3,3
`
	_, err := Load(strings.NewReader(raw))
	if err == nil || !strings.Contains(err.Error(), "p_success_2") {
		t.Fatalf("want missing step error, got %v", err)
	}
}

func TestLoad_MissingPointsStep(t *testing.T) {
	raw := `Input,Value
max_successes,2
p_success_1,50%
p_success_2,50%
points_success_1,1
`
	_, err := Load(strings.NewReader(raw))
	if err == nil || !strings.Contains(err.Error(), "points_success_2") {
		t.Fatalf("want missing points step, got %v", err)
	}
}

func TestLoad_NegativePercentProbability(t *testing.T) {
	raw := `Input,Value
max_successes,1
p_success_1,-5%
points_success_1,1
`
	_, err := Load(strings.NewReader(raw))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_InvalidPercentProbability(t *testing.T) {
	raw := `Input,Value
max_successes,1
p_success_1,120%
points_success_1,1
`
	_, err := Load(strings.NewReader(raw))
	if err == nil || !strings.Contains(err.Error(), "120%") && !strings.Contains(err.Error(), "percent") {
		t.Fatalf("want invalid percent error, got %v", err)
	}
}

func TestLoad_InvalidDecimalProbability(t *testing.T) {
	raw := `Input,Value
max_successes,1
p_success_1,1.7
points_success_1,1
`
	_, err := Load(strings.NewReader(raw))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_InvalidProbabilityGarbage(t *testing.T) {
	raw := `Input,Value
max_successes,1
p_success_1,abc
points_success_1,1
`
	_, err := Load(strings.NewReader(raw))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_DuplicateKey(t *testing.T) {
	raw := `Input,Value
max_successes,2
max_successes,3
p_success_1,50%
p_success_2,50%
points_success_1,1
points_success_2,2
`
	_, err := Load(strings.NewReader(raw))
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		t.Fatalf("want duplicate key error, got %v", err)
	}
}

func TestLoad_MaxSuccessesZero(t *testing.T) {
	raw := `Input,Value
max_successes,0
`
	_, err := Load(strings.NewReader(raw))
	if err == nil {
		t.Fatal("expected error")
	}
}
