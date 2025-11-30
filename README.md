# PowerGrid

PowerGrid is a small simulation of a regional grid that exposes live telemetry over HTTP and a streaming API consumed by the accompanying React UI.

## Getting started

### Containers

- **Development**: `docker compose up --build` runs the Go API with hot reload via Air on port **8080** and the Vite dev server on **5173**. Set `SIM_TICK_RATE`, `SOLAR_PEAK_MW`, `WIND_BASE_MW`, `GAS_CAPACITY_MW`, `GAS_MIN_MW`, or `GAS_RAMP_MW` to tune the simulation without recompiling. `VITE_API_PROXY` controls which backend the frontend proxies to (defaults to `http://localhost:8080`).
- **Production images**: `Dockerfile.backend` builds a minimal Go image (exposes **8080**) and `web/Dockerfile` builds static assets into an nginx image (exposes **80**). Use the `production` target for multi-stage builds.

Typical CPU load for the backend is well under one vCPU at the default 1s tick rate; decreasing `SIM_TICK_RATE` increases work proportionally. Expect low memory pressure (<100MB) under default settings.

Monitoring/logging: container logs already emit per-tick summaries. Ship them to your log aggregator and set alerts on frequency error (`frequency_hz` drift) or controller saturation (`gas_mw` nearing `capacity_mw`). Basic container health checks can poll `/status` for liveness.

### Backend

The backend is a Go service that advances the simulation in discrete ticks and serves API endpoints under `/status`, `/controller`, and `/stream`.

```bash
make test    # run all Go unit + integration tests
make run     # start the API server on :8080
make lint    # go vet plus frontend lint
make format  # gofmt and Prettier
```

### Frontend

The frontend lives in `web/` and uses Vite + React.

```bash
make frontend-install  # installs npm dependencies
npm --prefix web run start  # launches the dev server
npm --prefix web run build  # builds production assets
npm --prefix web run test   # runs Vitest
```

## Architecture notes

- **Simulation** (`internal/sim`): models solar, wind, gas, and residential demand. Each tick computes net supply/demand and resulting frequency.
- **Control loop** (`internal/control`): PID-based controller that targets grid frequency and can be overridden with manual setpoints. The controller chooses gas plant output which then ramps toward the requested setpoint.
- **API layer** (`internal/api`): exposes JSON endpoints for state and control parameters plus a streaming `/stream` endpoint supporting Server-Sent Events and WebSockets. Clients receive the latest `Snapshot` for each simulation tick.
- **CLI** (`cmd/server`): wires the simulation with the HTTP server and handles graceful shutdown.
- **Web UI** (`web/src`): consumes `/stream` for live charts and `/status`/`/controller` for current readings and settings.
