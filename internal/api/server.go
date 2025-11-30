package api

import (
	"encoding/json"
	"net/http"

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
	mux.HandleFunc("/state", s.handleState)

	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleState(w http.ResponseWriter, _ *http.Request) {
	snapshot := s.simulation.Snapshot()
	writeJSON(w, snapshot)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
