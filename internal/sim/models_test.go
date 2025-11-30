package sim

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"powergrid/internal/control"
)

func TestSolarOutputProfile(t *testing.T) {
	cfg := control.SolarConfig{PeakMW: 80, SunriseHour: 6, SunsetHour: 18}
	solar := NewSolar(cfg)

	morning := time.Date(2024, time.January, 1, 5, 0, 0, 0, time.UTC)
	midday := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)
	evening := time.Date(2024, time.January, 1, 19, 0, 0, 0, time.UTC)

	if got := solar.Output(morning); got != 0 {
		t.Fatalf("expected no production before sunrise, got %.2f", got)
	}

	peak := solar.Output(midday)
	if math.Abs(peak-cfg.PeakMW) > 0.5 {
		t.Fatalf("expected midday output near peak %.2f, got %.2f", cfg.PeakMW, peak)
	}

	if got := solar.Output(evening); got != 0 {
		t.Fatalf("expected no production after sunset, got %.2f", got)
	}
}

func TestWindSmoothing(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	wind := NewWind(control.WindConfig{BaseMW: 40, Variance: 10, Smoothing: 0.25})

	var previous float64
	for i := 0; i < 10; i++ {
		output := wind.Output(rng)
		if output < 0 {
			t.Fatalf("wind output should never be negative, got %.2f", output)
		}
		if i > 0 {
			delta := math.Abs(output - previous)
			if delta > 10 { // smoothed step should be less than a full variance jump
				t.Fatalf("wind smoothing too erratic, delta=%.2f", delta)
			}
		}
		previous = output
	}
}

func TestHouseDemandDistribution(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	demand := NewHouseDemand(control.HouseConfig{
		Count:         1000,
		BaseDemandKW:  3,
		Variance:      0.15,
		PeakHour:      19,
		PeakHourWidth: 3,
	})

	now := time.Date(2024, time.January, 1, 19, 0, 0, 0, time.UTC)
	samples := 1_000
	var total float64
	for i := 0; i < samples; i++ {
		v := demand.Draw(now, rng)
		if v < 0 {
			t.Fatalf("demand should never be negative, got %.4f", v)
		}
		total += v
	}

	mean := total / float64(samples)
	expected := 3.3 // Approximate MW for evening peak with defaults
	if math.Abs(mean-expected) > 0.3 {
		t.Fatalf("unexpected demand mean: got %.3f, want around %.3f", mean, expected)
	}
}
