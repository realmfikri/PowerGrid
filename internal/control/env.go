package control

import (
	"log"
	"os"
	"strconv"
	"time"
)

// FromEnv returns a Config hydrated with defaults and optional overrides
// provided via environment variables. Invalid values are ignored with a log
// message so the service can continue using defaults.
func FromEnv() Config {
	cfg := Defaults()

	if v := os.Getenv("SIM_TICK_RATE"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.TickRate = d
		} else {
			log.Printf("invalid SIM_TICK_RATE %q: %v", v, err)
		}
	}

	if v := os.Getenv("SOLAR_PEAK_MW"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Generator.Solar.PeakMW = f
		} else {
			log.Printf("invalid SOLAR_PEAK_MW %q: %v", v, err)
		}
	}

	if v := os.Getenv("WIND_BASE_MW"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Generator.Wind.BaseMW = f
		} else {
			log.Printf("invalid WIND_BASE_MW %q: %v", v, err)
		}
	}

	if v := os.Getenv("GAS_CAPACITY_MW"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Generator.Gas.CapacityMW = f
		} else {
			log.Printf("invalid GAS_CAPACITY_MW %q: %v", v, err)
		}
	}

	if v := os.Getenv("GAS_MIN_MW"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Generator.Gas.MinMW = f
		} else {
			log.Printf("invalid GAS_MIN_MW %q: %v", v, err)
		}
	}

	if v := os.Getenv("GAS_RAMP_MW"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Generator.Gas.RampMW = f
		} else {
			log.Printf("invalid GAS_RAMP_MW %q: %v", v, err)
		}
	}

	return cfg
}
