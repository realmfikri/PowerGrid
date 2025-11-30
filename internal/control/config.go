package control

import "time"

// Config defines runtime options for the power grid simulation and API server.
type Config struct {
	// TickRate controls how frequently the simulation advances.
	TickRate time.Duration

	// Generator holds parameters for power generation behaviour.
	Generator GeneratorConfig

	// HouseCount determines how many consumers participate in the simulation.
	HouseCount int
}

// GeneratorConfig captures tunable parameters for the generator model.
type GeneratorConfig struct {
	// BaseOutput represents the nominal megawatt output per tick.
	BaseOutput float64
	// Variance represents allowable fluctuation from the base output.
	Variance float64
}

// Defaults provides sensible starting values for local development.
func Defaults() Config {
	return Config{
		TickRate: time.Second,
		Generator: GeneratorConfig{
			BaseOutput: 500.0,
			Variance:   25.0,
		},
		HouseCount: 12,
	}
}
