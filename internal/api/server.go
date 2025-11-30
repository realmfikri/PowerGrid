package api

import (
	"encoding/json"
	"net/http"
	"time"

	"powergrid/internal/control"
	"powergrid/internal/sim"
)

// Server exposes HTTP endpoints backed by the simulation state.
type Server struct {
	simulation *sim.Simulation
}

// NewServer wires dependencies required by the HTTP layer.
func NewServer(simulation *sim.Simulation) *Server {
	return &Server{simulation: simulation}
}

// Handler returns an http.Handler ready to be used by a net/http server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/controller", s.handleController)
	mux.HandleFunc("/controller/pid", s.handlePIDUpdate)
	mux.HandleFunc("/controller/manual", s.handleManualGas)
	mux.HandleFunc("/stream", s.handleStream)

	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	snapshot := s.simulation.Snapshot()
	writeJSON(w, snapshot)
}

func (s *Server) handleController(w http.ResponseWriter, _ *http.Request) {
	snapshot := s.simulation.Snapshot()

	resp := struct {
		Control    any `json:"control"`
		Gas        any `json:"gas"`
		Controller any `json:"controller"`
	}{
		Control:    s.simulation.ControllerSettings(),
		Gas:        s.simulation.GasSettings(),
		Controller: snapshot.Controller,
	}

	writeJSON(w, resp)
}

func (s *Server) handleManualGas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Enable     *bool    `json:"enable"`
		SetpointMW *float64 `json:"setpoint_mw"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Enable != nil && *req.Enable {
		if req.SetpointMW == nil {
			http.Error(w, "setpoint_mw is required when enabling manual control", http.StatusBadRequest)
			return
		}
		s.simulation.EnableManualGas(*req.SetpointMW)
	} else if req.Enable != nil && !*req.Enable {
		s.simulation.DisableManualGas()
	}

	if req.Enable == nil && req.SetpointMW != nil {
		s.simulation.UpdateManualGasSetpoint(*req.SetpointMW)
	}

	writeJSON(w, s.simulation.Snapshot().Controller)
}

func (s *Server) handlePIDUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	current := s.simulation.ControllerSettings()
	pid := current.PID

	var req struct {
		Kp                *float64 `json:"kp"`
		Ki                *float64 `json:"ki"`
		Kd                *float64 `json:"kd"`
		IntegralMin       *float64 `json:"integral_min"`
		IntegralMax       *float64 `json:"integral_max"`
		TargetFrequencyHz *float64 `json:"target_frequency_hz"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Kp != nil {
		pid.Kp = *req.Kp
	}
	if req.Ki != nil {
		pid.Ki = *req.Ki
	}
	if req.Kd != nil {
		pid.Kd = *req.Kd
	}
	if req.IntegralMin != nil {
		pid.IntegralMin = *req.IntegralMin
	}
	if req.IntegralMax != nil {
		pid.IntegralMax = *req.IntegralMax
	}

	s.simulation.UpdatePIDConfig(pid)

	if req.TargetFrequencyHz != nil {
		s.simulation.UpdateTargetFrequency(*req.TargetFrequencyHz)
	}

	writeJSON(w, struct {
		Control control.ControlConfig `json:"control"`
	}{Control: s.simulation.ControllerSettings()})
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := s.simulation.Subscribe(8)
	ctx := r.Context()

	// Send an initial snapshot for eager clients.
	writeSSE(w, flusher, s.simulation.Snapshot())

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case snap := <-ch:
			writeSSE(w, flusher, snap)
		case <-ticker.C:
			// Periodic flush to keep connection alive for idle periods.
			flusher.Flush()
		}
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte("data: "))
	_, _ = w.Write(data)
	_, _ = w.Write([]byte("\n\n"))
	flusher.Flush()
}
