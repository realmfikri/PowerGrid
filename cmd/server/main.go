package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"powergrid/internal/api"
	"powergrid/internal/control"
	"powergrid/internal/sim"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := control.Defaults()

	simulation := sim.NewSimulation(cfg)
	go func() {
		if err := simulation.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("simulation stopped: %v", err)
		}
	}()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: api.NewServer(simulation).Handler(),
	}

	go func() {
		log.Printf("HTTP server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
