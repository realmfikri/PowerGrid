import './App.css'
import type { ControllerSettingsResponse, SimulationSnapshot } from './types/simulation'

const placeholderSnapshot: SimulationSnapshot = {
  tick_count: 482,
  timestamp: new Date().toISOString(),
  solar_mw: 72.4,
  wind_mw: 41.6,
  gas_mw: 55.2,
  supply_mw: 169.2,
  demand_mw: 164.1,
  net_balance_mw: 5.1,
  frequency_hz: 60.03,
  controller: {
    manual_enabled: false,
    manual_setpoint_mw: 58,
    setpoint_mw: 55.2,
    integral: 3.4,
    error: -0.02,
  },
}

const controllerSettings: ControllerSettingsResponse = {
  control: {
    target_frequency_hz: 60,
    pid: {
      kp: 0.8,
      ki: 0.3,
      kd: 0.05,
      integral_min: -80,
      integral_max: 80,
    },
  },
  gas: {
    capacity_mw: 140,
    min_mw: 20,
    ramp_mw: 12,
  },
  controller: placeholderSnapshot.controller,
}

const navLinks = ['Overview', 'Streams', 'Controls', 'Settings']

const highlights = [
  {
    label: 'Grid frequency',
    value: `${placeholderSnapshot.frequency_hz.toFixed(2)} Hz`,
    detail: 'Target 60.00 Hz',
  },
  {
    label: 'Supply vs demand',
    value: `${placeholderSnapshot.supply_mw.toFixed(1)} / ${placeholderSnapshot.demand_mw.toFixed(1)} MW`,
    detail: `${placeholderSnapshot.net_balance_mw.toFixed(1)} MW reserve`,
  },
  {
    label: 'Dispatchable output',
    value: `${placeholderSnapshot.gas_mw.toFixed(1)} MW`,
    detail: controllerSettings.controller.manual_enabled ? 'Manual override' : 'PID managed',
  },
]

function App() {
  return (
    <div className="app-shell">
      <header className="top-bar">
        <div className="brand">
          <div className="brand-mark">⚡</div>
          <div>
            <div className="brand-name">PowerGrid</div>
            <div className="brand-subtitle">Simulation control panel</div>
          </div>
        </div>
        <nav className="main-nav">
          {navLinks.map((link) => (
            <a key={link} href="#" className="nav-link">
              {link}
            </a>
          ))}
        </nav>
        <div className="env-pill">Dev proxy → http://localhost:8080</div>
      </header>

      <main className="content">
        <section className="hero-panel">
          <div className="hero-copy">
            <p className="eyebrow">Realtime ops</p>
            <h1>Monitor, tune, and steer the grid</h1>
            <p className="lede">
              Live status snapshots, controller state, and dispatch controls ready for visualization. Charts and
              gauges will stream data from the backend once wired to the API.
            </p>
            <div className="hero-actions">
              <button type="button" className="primary">View status stream</button>
              <button type="button" className="secondary">Open control sandbox</button>
            </div>
          </div>
          <div className="summary">
            {highlights.map((item) => (
              <div key={item.label} className="summary-card">
                <p className="summary-label">{item.label}</p>
                <p className="summary-value">{item.value}</p>
                <p className="summary-detail">{item.detail}</p>
              </div>
            ))}
          </div>
        </section>

        <section className="panel-grid">
          <article className="panel wide">
            <div className="panel-heading">
              <div>
                <p className="eyebrow">Supply & demand</p>
                <h2>Energy balance timeline</h2>
                <p className="panel-subtitle">Placeholder for a chart overlaying solar, wind, gas, and demand curves.</p>
              </div>
              <span className="placeholder-chip">Chart placeholder</span>
            </div>
            <div className="chart-placeholder">
              <div className="bar" style={{ width: '60%' }} />
              <div className="bar" style={{ width: '35%' }} />
              <div className="bar" style={{ width: '80%' }} />
              <div className="bar" style={{ width: '50%' }} />
            </div>
          </article>

          <article className="panel">
            <div className="panel-heading">
              <div>
                <p className="eyebrow">Grid stability</p>
                <h2>Frequency gauge</h2>
                <p className="panel-subtitle">Add a dial or gauge to highlight frequency drift.</p>
              </div>
              <span className="placeholder-chip">Gauge placeholder</span>
            </div>
            <div className="gauge">
              <div className="gauge-ring">
                <div className="gauge-needle" style={{ transform: 'rotate(-4deg)' }} />
                <div className="gauge-center">{placeholderSnapshot.frequency_hz.toFixed(2)} Hz</div>
              </div>
              <p className="gauge-footnote">Target {controllerSettings.control.target_frequency_hz.toFixed(2)} Hz</p>
            </div>
          </article>

          <article className="panel">
            <div className="panel-heading">
              <div>
                <p className="eyebrow">Controller</p>
                <h2>PID + manual dispatch</h2>
                <p className="panel-subtitle">Space for controls to toggle manual gas and tune PID gains.</p>
              </div>
              <span className="placeholder-chip">Control placeholders</span>
            </div>
            <div className="controls-grid">
              <div className="control-tile">
                <p className="control-label">Manual override</p>
                <p className="control-value">
                  {controllerSettings.controller.manual_enabled ? 'Enabled' : 'Off'}
                </p>
                <p className="control-helper">
                  Setpoint {controllerSettings.controller.manual_setpoint_mw.toFixed(1)} MW ready to dispatch.
                </p>
              </div>
              <div className="control-tile">
                <p className="control-label">PID tuning</p>
                <p className="control-value">
                  kp {controllerSettings.control.pid.kp}, ki {controllerSettings.control.pid.ki}, kd{' '}
                  {controllerSettings.control.pid.kd}
                </p>
                <p className="control-helper">Inputs for sliders/inputs to adjust gains and integral bounds.</p>
              </div>
              <div className="control-tile">
                <p className="control-label">Gas plant limits</p>
                <p className="control-value">
                  {controllerSettings.gas.min_mw}–{controllerSettings.gas.capacity_mw} MW
                </p>
                <p className="control-helper">Include ramp-rate guardrails and manual setpoint input.</p>
              </div>
            </div>
          </article>

          <article className="panel">
            <div className="panel-heading">
              <div>
                <p className="eyebrow">Streaming</p>
                <h2>Status events</h2>
                <p className="panel-subtitle">SSE or WebSocket events from /stream will render here.</p>
              </div>
              <span className="placeholder-chip">Live feed</span>
            </div>
            <div className="stream-list">
              <div className="stream-item">
                <span className="bullet" />
                <div>
                  <p className="stream-title">Tick #{placeholderSnapshot.tick_count}</p>
                  <p className="stream-meta">{placeholderSnapshot.timestamp}</p>
                </div>
                <div className="stream-values">{placeholderSnapshot.net_balance_mw.toFixed(1)} MW net</div>
              </div>
              <div className="stream-item">
                <span className="bullet" />
                <div>
                  <p className="stream-title">Controller error</p>
                  <p className="stream-meta">Integral {placeholderSnapshot.controller.integral.toFixed(2)}</p>
                </div>
                <div className="stream-values">{placeholderSnapshot.controller.error.toFixed(2)}</div>
              </div>
              <div className="stream-item">
                <span className="bullet" />
                <div>
                  <p className="stream-title">Dispatch setpoint</p>
                  <p className="stream-meta">PID output</p>
                </div>
                <div className="stream-values">{placeholderSnapshot.controller.setpoint_mw.toFixed(1)} MW</div>
              </div>
            </div>
          </article>

          <article className="panel">
            <div className="panel-heading">
              <div>
                <p className="eyebrow">Operations</p>
                <h2>Operator controls</h2>
                <p className="panel-subtitle">Reserved space for start/stop actions and scenario toggles.</p>
              </div>
              <span className="placeholder-chip">Control surface</span>
            </div>
            <div className="controls-grid">
              <div className="control-tile">
                <p className="control-label">Scenario selector</p>
                <p className="control-helper">Dropdown or segmented control placeholder.</p>
              </div>
              <div className="control-tile">
                <p className="control-label">Safety checks</p>
                <p className="control-helper">Checklist or status indicators placeholder.</p>
              </div>
              <div className="control-tile">
                <p className="control-label">Export</p>
                <p className="control-helper">Button to download snapshots or controller configs.</p>
              </div>
            </div>
          </article>
        </section>
      </main>
    </div>
  )
}

export default App
