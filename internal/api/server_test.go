package api

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"

	"powergrid/internal/control"
	"powergrid/internal/sim"
)

func TestStreamWebsocket(t *testing.T) {
	cfg := control.Defaults()
	cfg.TickRate = 50 * time.Millisecond
	simulation := sim.NewSimulation(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		_ = simulation.Run(ctx)
		close(done)
	}()

	ts := httptest.NewServer(NewServer(simulation).Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/stream"
	conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket stream: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test complete")

	var snapshots []sim.Snapshot
	seenTicks := make(map[int64]bool)
	readCtx, cancelRead := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelRead()

	for len(seenTicks) < 2 || len(snapshots) < 3 {
		_, data, err := conn.Read(readCtx)
		if err != nil {
			t.Fatalf("failed to read from stream: %v", err)
		}
		var snap sim.Snapshot
		if err := json.Unmarshal(data, &snap); err != nil {
			t.Fatalf("failed to decode snapshot: %v", err)
		}
		snapshots = append(snapshots, snap)
		seenTicks[snap.TickCount] = true
	}

	cancel()
	<-done

	if len(seenTicks) < 2 {
		t.Fatalf("expected advancing ticks in stream messages: %+v", snapshots)
	}
	if snapshots[len(snapshots)-1].FrequencyHz == 0 {
		t.Fatalf("expected frequency to be computed in stream snapshots")
	}
}
