package sim

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"powergrid/internal/control"
)

// Simulation drives the state of the power grid over discrete ticks.
type Simulation struct {
	cfg control.Config

	mu          sync.RWMutex
	tickCount   int64
	lastOutput  float64
	houseDemand float64

	rng *rand.Rand
}

// Snapshot exposes read-only state for external consumers such as the API.
type Snapshot struct {
	TickCount   int64   `json:"tick_count"`
	Generation  float64 `json:"generation"`
	HouseDemand float64 `json:"house_demand"`
}

// NewSimulation constructs a simulation seeded with configuration and defaults.
func NewSimulation(cfg control.Config) *Simulation {
	return &Simulation{
		cfg: cfg,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Run advances the simulation until the provided context is cancelled.
func (s *Simulation) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.cfg.TickRate)
	defer ticker.Stop()

	// Generate an initial state before the first tick for eager consumers.
	s.step()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.step()
		}
	}
}

// Snapshot returns a copy of the current state.
func (s *Simulation) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Snapshot{
		TickCount:   s.tickCount,
		Generation:  s.lastOutput,
		HouseDemand: s.houseDemand,
	}
}

func (s *Simulation) step() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Simulate generator output with simple bounded noise.
	variation := (s.rng.Float64()*2 - 1) * s.cfg.Generator.Variance
	s.lastOutput = s.cfg.Generator.BaseOutput + variation

	// Model demand as a proportion of the generator output scaled by house count.
	householdLoad := 2.5 // MW per house baseline
	s.houseDemand = float64(s.cfg.HouseCount) * householdLoad

	s.tickCount++
}
