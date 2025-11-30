package control

import "time"

// Config defines runtime options for the power grid simulation and API server.
type Config struct {
	// TickRate controls how frequently the simulation advances.
	TickRate time.Duration

	// Generator holds parameters for power generation behaviour.
	Generator GeneratorConfig

	// Grid determines how the simulated frequency responds to imbalances.
	Grid GridConfig
}

// GeneratorConfig captures tunable parameters for the generator model.
type GeneratorConfig struct {
	// Solar drives deterministic production based on time of day.
	Solar SolarConfig
	// Wind captures smooth, noisy output similar to Perlin noise.
	Wind WindConfig
	// Gas provides dispatchable capacity to balance the grid.
	Gas GasConfig
	// Houses defines residential demand assumptions.
	Houses HouseConfig
}

// SolarConfig defines sunrise/sunset profile for solar output.
type SolarConfig struct {
	PeakMW      float64
	SunriseHour float64
	SunsetHour  float64
}

// WindConfig sets expectations for wind-based generation.
type WindConfig struct {
	BaseMW    float64
	Variance  float64
	Smoothing float64
}

// GasConfig sets bounds for dispatchable generation.
type GasConfig struct {
	CapacityMW float64
	MinMW      float64
	RampMW     float64
}

// HouseConfig captures assumptions for residential load.
type HouseConfig struct {
	Count         int
	BaseDemandKW  float64
	Variance      float64
	PeakHour      float64
	PeakHourWidth float64
}

// GridConfig tunes how grid frequency responds to power imbalance.
type GridConfig struct {
	BaseFrequencyHz float64
	SensitivityHz   float64
}

// Defaults provides sensible starting values for local development.
func Defaults() Config {
	return Config{
		TickRate: time.Second,
		Generator: GeneratorConfig{
			Solar: SolarConfig{
				PeakMW:      80,
				SunriseHour: 6.0,
				SunsetHour:  20.0,
			},
			Wind: WindConfig{
				BaseMW:    40,
				Variance:  15,
				Smoothing: 0.2,
			},
			Gas: GasConfig{
				CapacityMW: 140,
				MinMW:      20,
				RampMW:     12,
			},
			Houses: HouseConfig{
				Count:         1000,
				BaseDemandKW:  3.0,
				Variance:      0.15,
				PeakHour:      19.0,
				PeakHourWidth: 3.0,
			},
		},
		Grid: GridConfig{
			BaseFrequencyHz: 60.0,
			SensitivityHz:   0.01,
		},
	}
}
