# PowerGrid

PowerGrid is a small simulation of a regional grid that exposes live telemetry over HTTP and a streaming API consumed by the accompanying React UI.

## Getting started

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
