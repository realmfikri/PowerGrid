package sim

import (
	"context"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"powergrid/internal/control"
)

// Simulation drives the state of the power grid over discrete ticks.
type Simulation struct {
	cfg control.Config

	mu          sync.RWMutex
	tick        Snapshot
	rng         *rand.Rand
	subscribers []chan Snapshot

	controller  *control.Controller
	solar       Solar
	wind        Wind
	gas         Gas
	houseDemand HouseDemand
}

// Snapshot exposes read-only state for external consumers such as the API.
type Snapshot struct {
	TickCount    int64                      `json:"tick_count"`
	Timestamp    time.Time                  `json:"timestamp"`
	SolarMW      float64                    `json:"solar_mw"`
	WindMW       float64                    `json:"wind_mw"`
	GasMW        float64                    `json:"gas_mw"`
	SupplyMW     float64                    `json:"supply_mw"`
	DemandMW     float64                    `json:"demand_mw"`
	NetBalanceMW float64                    `json:"net_balance_mw"`
	FrequencyHz  float64                    `json:"frequency_hz"`
	Controller   control.ControllerSnapshot `json:"controller"`
}

// NewSimulation constructs a simulation seeded with configuration and defaults.
func NewSimulation(cfg control.Config) *Simulation {
	sim := &Simulation{
		cfg:         cfg,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
		controller:  control.NewController(cfg.Control, cfg.Generator.Gas),
		solar:       NewSolar(cfg.Generator.Solar),
		wind:        NewWind(cfg.Generator.Wind),
		gas:         NewGas(cfg.Generator.Gas),
		houseDemand: NewHouseDemand(cfg.Generator.Houses),
	}

	return sim
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

	return s.tick
}

// ControllerSettings returns static configuration for the control loop.
func (s *Simulation) ControllerSettings() control.ControlConfig {
	return s.cfg.Control
}

// GasSettings exposes the generator configuration for the dispatchable plant.
func (s *Simulation) GasSettings() control.GasConfig {
	return s.cfg.Generator.Gas
}

// Subscribe returns a channel that receives snapshots on each tick.
func (s *Simulation) Subscribe(buffer int) <-chan Snapshot {
	ch := make(chan Snapshot, buffer)

	s.mu.Lock()
	s.subscribers = append(s.subscribers, ch)
	if s.tick.TickCount > 0 {
		ch <- s.tick
	}
	s.mu.Unlock()

	return ch
}

// EnableManualGas switches control of the gas plant to manual mode.
func (s *Simulation) EnableManualGas(setpointMW float64) {
	s.controller.ManualEnable(setpointMW)
}

// DisableManualGas returns control to PID automation.
func (s *Simulation) DisableManualGas() {
	s.controller.ManualDisable()
}

// UpdateManualGasSetpoint adjusts the manual gas output target.
func (s *Simulation) UpdateManualGasSetpoint(setpointMW float64) {
	s.controller.ManualSetpoint(setpointMW)
}

// UpdatePIDConfig replaces the controller PID tuning values and persists them to the config.
func (s *Simulation) UpdatePIDConfig(cfg control.PIDConfig) {
	s.controller.UpdatePID(cfg)
	s.cfg.Control.PID = cfg
}

// UpdateTargetFrequency adjusts the desired grid frequency for the controller loop.
func (s *Simulation) UpdateTargetFrequency(target float64) {
	s.controller.UpdateTarget(target)
	s.cfg.Control.TargetFrequencyHz = target
}

func (s *Simulation) step() {
	s.mu.Lock()
	now := time.Now()
	solar := s.solar.Output(now)
	wind := s.wind.Output(s.rng)
	demand := s.houseDemand.Draw(now, s.rng)
	renewable := solar + wind

	measuredHz := s.tick.FrequencyHz
	if measuredHz == 0 {
		measuredHz = s.cfg.Control.TargetFrequencyHz
	}

	deficit := demand - renewable
	gasSetpoint := s.controller.NextSetpoint(measuredHz, deficit, s.cfg.TickRate)
	gas := s.gas.Setpoint(gasSetpoint)
	supply := renewable + gas
	net := supply - demand
	freq := s.cfg.Grid.BaseFrequencyHz + net*s.cfg.Grid.SensitivityHz

	s.tick.TickCount++
	s.tick.Timestamp = now
	s.tick.SolarMW = solar
	s.tick.WindMW = wind
	s.tick.GasMW = gas
	s.tick.SupplyMW = supply
	s.tick.DemandMW = demand
	s.tick.NetBalanceMW = net
	s.tick.FrequencyHz = freq
	controllerSnapshot := s.controller.Snapshot()
	s.tick.Controller = controllerSnapshot

	subscribers := append([]chan Snapshot(nil), s.subscribers...)
	s.mu.Unlock()

	log.Printf(
		"tick=%d freq=%.2fHz gas=%.2fMW manual=%t error=%.3f net=%.2fMW deficit=%.2fMW setpoint=%.2fMW",
		s.tick.TickCount,
		freq,
		gas,
		controllerSnapshot.ManualEnabled,
		controllerSnapshot.Error,
		net,
		deficit,
		controllerSnapshot.SetpointMW,
	)

	for _, ch := range subscribers {
		select {
		case ch <- s.tick:
		default:
		}
	}
}

// Solar models deterministic day-night production using a sinusoidal profile.
type Solar struct {
	peakMW      float64
	sunriseHour float64
	sunsetHour  float64
}

// NewSolar constructs a Solar generator from configuration.
func NewSolar(cfg control.SolarConfig) Solar {
	return Solar{
		peakMW:      cfg.PeakMW,
		sunriseHour: cfg.SunriseHour,
		sunsetHour:  cfg.SunsetHour,
	}
}

// Output returns the solar output for the provided moment in MW.
func (s Solar) Output(now time.Time) float64 {
	hour := float64(now.Hour()) + float64(now.Minute())/60 + float64(now.Second())/3600
	if hour <= s.sunriseHour || hour >= s.sunsetHour || s.peakMW <= 0 {
		return 0
	}
	dayFraction := (hour - s.sunriseHour) / (s.sunsetHour - s.sunriseHour)
	angle := math.Pi * dayFraction
	return math.Max(0, s.peakMW*math.Sin(angle))
}

// Wind produces smooth pseudo-random output akin to Perlin noise.
type Wind struct {
	baseMW    float64
	variance  float64
	smoothing float64
	current   float64
}

// NewWind creates a wind model using configured parameters.
func NewWind(cfg control.WindConfig) Wind {
	smoothing := cfg.Smoothing
	if smoothing <= 0 || smoothing > 1 {
		smoothing = 0.15
	}
	return Wind{
		baseMW:    cfg.BaseMW,
		variance:  cfg.Variance,
		smoothing: smoothing,
	}
}

// Output advances the internal noise state and returns current generation in MW.
func (w *Wind) Output(rng *rand.Rand) float64 {
	if w.current == 0 {
		w.current = w.baseMW
	}
	target := w.baseMW + rng.NormFloat64()*w.variance
	// Smooth towards the new random target to mimic Perlin-like continuity.
	w.current = w.current*(1-w.smoothing) + target*w.smoothing
	if w.current < 0 {
		w.current = 0
	}
	return w.current
}

// Gas represents dispatchable generation used to balance demand.
type Gas struct {
	capacityMW float64
	minMW      float64
	rampMW     float64
	current    float64
}

// NewGas constructs a gas generator model from config.
func NewGas(cfg control.GasConfig) Gas {
	return Gas{
		capacityMW: cfg.CapacityMW,
		minMW:      cfg.MinMW,
		rampMW:     cfg.RampMW,
	}
}

// Setpoint moves the gas plant towards a target output while respecting limits.
func (g *Gas) Setpoint(target float64) float64 {
	if target < g.minMW {
		target = g.minMW
	}
	if target > g.capacityMW {
		target = g.capacityMW
	}
	if g.rampMW <= 0 {
		g.current = target
		return g.current
	}
	// Move towards target with ramp constraint.
	delta := target - g.current
	if delta > g.rampMW {
		delta = g.rampMW
	} else if delta < -g.rampMW {
		delta = -g.rampMW
	}
	g.current += delta
	return g.current
}

// Dispatch attempts to cover the provided deficit (MW) while respecting ramp limits.
func (g *Gas) Dispatch(deficit float64) float64 {
	return g.Setpoint(deficit)
}

// HouseDemand models aggregated stochastic residential demand.
type HouseDemand struct {
	homes         int
	baseDemandKW  float64
	variance      float64
	peakHour      float64
	peakHourWidth float64
}

// NewHouseDemand creates a demand model for residential consumers.
func NewHouseDemand(cfg control.HouseConfig) HouseDemand {
	return HouseDemand{
		homes:         cfg.Count,
		baseDemandKW:  cfg.BaseDemandKW,
		variance:      cfg.Variance,
		peakHour:      cfg.PeakHour,
		peakHourWidth: cfg.PeakHourWidth,
	}
}

// Draw samples a demand figure (in MW) for the provided time.
func (h HouseDemand) Draw(now time.Time, rng *rand.Rand) float64 {
	hour := float64(now.Hour()) + float64(now.Minute())/60 + float64(now.Second())/3600
	morningBump := math.Exp(-math.Pow((hour-7)/2, 2))
	eveningPeak := math.Exp(-math.Pow((hour-h.peakHour)/h.peakHourWidth, 2))
	cycle := 0.6 + 0.25*morningBump + 0.5*eveningPeak
	meanKW := h.baseDemandKW * cycle * float64(h.homes)
	stddevKW := h.variance * meanKW
	demandKW := meanKW + rng.NormFloat64()*stddevKW
	if demandKW < 0 {
		demandKW = 0
	}
	return demandKW / 1000 // convert kW to MW
}
