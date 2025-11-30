package control

import (
	"math"
	"sync"
	"time"
)

// Controller manages dispatchable generation to keep grid frequency near a target.
type Controller struct {
	mu sync.Mutex

	targetHz float64
	pid      PIDConfig
	gas      GasConfig

	integral     float64
	prevError    float64
	lastTime     time.Time
	lastSetpoint float64

	manualEnabled  bool
	manualSetpoint float64
}

// ControllerSnapshot captures a consistent view of controller state for logging or monitoring.
type ControllerSnapshot struct {
	ManualEnabled bool
	SetpointMW    float64
	Integral      float64
	Error         float64
}

// NewController builds a controller using control loop tuning and generator limits.
func NewController(ctrlCfg ControlConfig, gasCfg GasConfig) *Controller {
	return &Controller{
		targetHz: ctrlCfg.TargetFrequencyHz,
		pid:      ctrlCfg.PID,
		gas:      gasCfg,
	}
}

// NextSetpoint determines the desired gas output in MW for the upcoming tick.
func (c *Controller) NextSetpoint(measuredHz float64, deficitMW float64, tickDuration time.Duration) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	if measuredHz == 0 {
		measuredHz = c.targetHz
	}

	now := time.Now()
	dt := tickDuration.Seconds()
	if !c.lastTime.IsZero() {
		dt = now.Sub(c.lastTime).Seconds()
	}
	c.lastTime = now

	var target float64
	if c.manualEnabled {
		target = c.manualSetpoint
	} else {
		errorValue := c.targetHz - measuredHz
		c.integral += errorValue * dt
		c.integral = clamp(c.integral, c.pid.IntegralMin, c.pid.IntegralMax)
		derivative := 0.0
		if dt > 0 {
			derivative = (errorValue - c.prevError) / dt
		}
		c.prevError = errorValue

		correction := c.pid.Kp*errorValue + c.pid.Ki*c.integral + c.pid.Kd*derivative
		target = deficitMW + correction
	}

	c.lastSetpoint = c.clampTarget(target)
	return c.lastSetpoint
}

// ManualEnable switches the controller into manual override mode and sets the target output.
func (c *Controller) ManualEnable(setpointMW float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.manualEnabled = true
	c.manualSetpoint = setpointMW
}

// ManualDisable returns the controller to automatic PID mode while preserving the last setpoint.
func (c *Controller) ManualDisable() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.manualEnabled = false
}

// ManualSetpoint adjusts the manual output target without toggling the mode.
func (c *Controller) ManualSetpoint(setpointMW float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.manualSetpoint = setpointMW
}

// Snapshot returns a point-in-time view of the controller internals.
func (c *Controller) Snapshot() ControllerSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	return ControllerSnapshot{
		ManualEnabled: c.manualEnabled,
		SetpointMW:    c.lastSetpoint,
		Integral:      c.integral,
		Error:         c.prevError,
	}
}

func (c *Controller) clampTarget(target float64) float64 {
	target = clamp(target, c.gas.MinMW, c.gas.CapacityMW)
	return target
}

func clamp(value, min, max float64) float64 {
	return math.Max(min, math.Min(max, value))
}
