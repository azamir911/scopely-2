package simulation

import "testing"

func TestPointsForSuccesses_CumulativeDepth(t *testing.T) {
	t.Parallel()

	points := []float64{10, 20, 30, 5}
	cases := []struct {
		k    int
		want float64
	}{
		{0, 0},
		{1, 10},
		{2, 30},
		{3, 60},
		{4, 65},
		{99, 65},
	}
	for _, tc := range cases {
		if got := PointsForSuccesses(tc.k, points); got != tc.want {
			t.Fatalf("successes=%d: want %v got %v", tc.k, tc.want, got)
		}
	}
}

func TestExpectedPointsOneInteraction_MatchesManual(t *testing.T) {
	t.Parallel()

	// Two stages: P(0)=0.5 fail first, P(1)=0.5 fail second after first success.
	probs := []float64{0.5, 0.5}
	dist := AnalyticalDistribution(probs)
	pointSeries := []float64{3.0, 4.0}

	exp := ExpectedPointsOneInteraction(dist, pointSeries)
	// P(0)=0.5*0 + P(1)=0.25*3 + P(2)=0.25*7 = 0 + 0.75 + 1.75 = 2.5
	want := 0.5*0 + 0.25*3.0 + 0.25*(3.0+4.0)
	if exp != want {
		t.Fatalf("want %v got %v", want, exp)
	}
}
