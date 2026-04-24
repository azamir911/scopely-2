package aggregation

import (
	"math"
	"testing"
)

func TestFloorInteractions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		v    float64
		want int64
	}{
		{10.9, 10},
		{10.0, 10},
		{0.9, 0},
		{0, 0},
		{-1, 0},
		{math.NaN(), 0},
		{math.Inf(1), 0},
		{math.Inf(-1), 0},
	}
	for _, tc := range cases {
		got := FloorInteractions(tc.v)
		if got != tc.want {
			t.Fatalf("FloorInteractions(%v) = %d, want %d", tc.v, got, tc.want)
		}
	}
}
