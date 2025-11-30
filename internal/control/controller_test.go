package control

import (
	"testing"
	"time"
)

func TestControllerPIDResponse(t *testing.T) {
	ctrl := NewController(ControlConfig{
		TargetFrequencyHz: 60,
		PID:               PIDConfig{Kp: 1.0, Ki: 0.5, Kd: 0.1, IntegralMin: -1, IntegralMax: 1},
	}, GasConfig{CapacityMW: 100, MinMW: 10})

	// Prime the controller with a 1 Hz deficit over a 1 second interval.
	ctrl.lastTime = time.Now().Add(-time.Second)
	first := ctrl.NextSetpoint(59, 20, time.Second)
	snap := ctrl.Snapshot()
	if snap.Integral != 1 {
		t.Fatalf("expected integral to accumulate to 1, got %.2f", snap.Integral)
	}
	expectedFirst := 20 + (1.0*1 + 0.5*1 + 0.1*1)
	if diff := expectedFirst - first; diff > 0.01 || diff < -0.01 {
		t.Fatalf("unexpected first setpoint: got %.2f want %.2f", first, expectedFirst)
	}

	// A larger error should clamp the integral while increasing the setpoint.
	ctrl.lastTime = time.Now().Add(-time.Second)
	second := ctrl.NextSetpoint(50, 20, time.Second)
	snap = ctrl.Snapshot()
	if snap.Integral != 1 { // integral should be clamped at max
		t.Fatalf("expected integral clamp to 1, got %.2f", snap.Integral)
	}
	if second <= first {
		t.Fatalf("expected second setpoint %.2f to exceed first %.2f", second, first)
	}
	if second > 100 {
		t.Fatalf("setpoint should respect gas capacity limits, got %.2f", second)
	}
}

func TestControllerManualMode(t *testing.T) {
	ctrl := NewController(ControlConfig{TargetFrequencyHz: 60, PID: PIDConfig{}}, GasConfig{CapacityMW: 80, MinMW: 20})

	ctrl.ManualEnable(50)
	auto := ctrl.NextSetpoint(59, 10, time.Second)
	if auto != 50 {
		t.Fatalf("manual setpoint should override calculations, got %.2f", auto)
	}

	ctrl.ManualSetpoint(30)
	ctrl.lastTime = time.Now().Add(-time.Second)
	updated := ctrl.NextSetpoint(50, 10, time.Second)
	if updated != 30 {
		t.Fatalf("manual adjustments should persist, got %.2f", updated)
	}

	ctrl.ManualDisable()
	ctrl.lastTime = time.Now().Add(-time.Second)
	resumed := ctrl.NextSetpoint(59, 10, time.Second)
	if resumed < ctrl.gas.MinMW || resumed > ctrl.gas.CapacityMW {
		t.Fatalf("automatic mode should respect bounds, got %.2f", resumed)
	}
}
