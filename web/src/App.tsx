import { useEffect, useMemo, useState } from 'react'
import {
  Area,
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
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

const placeholderController: ControllerSettingsResponse = {
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

const HISTORY_LENGTH = 180

function formatTickLabel(timestamp: string) {
  const date = new Date(timestamp)
  return `${date.getMinutes().toString().padStart(2, '0')}:${date
    .getSeconds()
    .toString()
    .padStart(2, '0')}`
}

function frequencyColor(frequency: number, target: number) {
  if (frequency < 59) return '#ff7b7b'
  if (frequency < target) return '#f2c84b'
  if (frequency > target + 0.35) return '#5bd8ff'
  return '#7cf0a3'
}

function gaugeRotation(frequency: number, target: number) {
  const min = target - 1.2
  const max = target + 1.2
  const clamped = Math.min(Math.max(frequency, min), max)
  const ratio = (clamped - min) / (max - min)
  return -110 + ratio * 220
}

function App() {
  const [snapshot, setSnapshot] = useState<SimulationSnapshot>(placeholderSnapshot)
  const [controllerSettings, setControllerSettings] = useState<ControllerSettingsResponse>(placeholderController)
  const [history, setHistory] = useState<SimulationSnapshot[]>([placeholderSnapshot])
  const [flash, setFlash] = useState(false)

  useEffect(() => {
    fetch('/api/controller')
      .then((res) => res.json())
      .then((data: ControllerSettingsResponse) => {
        setControllerSettings(data)
      })
      .catch(() => {
        // keep placeholder settings
      })
  }, [])

  useEffect(() => {
    const eventSource = new EventSource('/stream')

    eventSource.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data) as SimulationSnapshot
        setSnapshot(payload)
        setHistory((prev) => {
          const next = [...prev, payload]
          if (next.length > HISTORY_LENGTH) {
            next.shift()
          }
          return next
        })
      } catch (err) {
        console.error('failed to parse stream event', err)
      }
    }

    eventSource.onerror = () => {
      eventSource.close()
    }

    return () => eventSource.close()
  }, [])

  useEffect(() => {
    if (snapshot.frequency_hz < 59) {
      setFlash(true)
      const timeout = setTimeout(() => setFlash(false), 300)
      return () => clearTimeout(timeout)
    }
    setFlash(false)
    return undefined
  }, [snapshot.frequency_hz])

  const highlights = useMemo(
    () => [
      {
        label: 'Grid frequency',
        value: `${snapshot.frequency_hz.toFixed(2)} Hz`,
        detail: `Target ${controllerSettings.control.target_frequency_hz.toFixed(2)} Hz`,
      },
      {
        label: 'Supply vs demand',
        value: `${snapshot.supply_mw.toFixed(1)} / ${snapshot.demand_mw.toFixed(1)} MW`,
        detail: `${snapshot.net_balance_mw.toFixed(1)} MW reserve`,
      },
      {
        label: 'Dispatchable output',
        value: `${snapshot.gas_mw.toFixed(1)} MW`,
        detail: snapshot.controller.manual_enabled ? 'Manual override' : 'PID managed',
      },
    ],
    [snapshot, controllerSettings.control.target_frequency_hz],
  )

  const chartData = useMemo(
    () =>
      history.map((point) => ({
        ...point,
        label: formatTickLabel(point.timestamp),
      })),
    [history],
  )

  const recentEvents = useMemo(() => history.slice(-6).reverse(), [history])

  const freqColor = frequencyColor(snapshot.frequency_hz, controllerSettings.control.target_frequency_hz)
  const rotation = gaugeRotation(snapshot.frequency_hz, controllerSettings.control.target_frequency_hz)

  return (
    <div className={`app-shell ${flash ? 'flash' : ''}`}>
      <header className="top-bar">
        <div className="brand">
          <div className="brand-mark">⚡</div>
          <div>
            <div className="brand-name">PowerGrid</div>
            <div className="brand-subtitle">Simulation control panel</div>
          </div>
        </div>
        <nav className="main-nav">
          {['Overview', 'Streams', 'Controls', 'Settings'].map((link) => (
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
              Live status snapshots, controller state, and dispatch controls ready for visualization. Charts and gauges now
              stream live from the simulation backend.
            </p>
            <div className="hero-actions">
              <button type="button" className="primary">
                View status stream
              </button>
              <button type="button" className="secondary">
                Open control sandbox
              </button>
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
                <p className="panel-subtitle">Live overlay of supply stack versus consumer demand from the stream.</p>
              </div>
              <span className="live-chip">Live</span>
            </div>
            <div className="chart-area">
              <ResponsiveContainer width="100%" height={320}>
                <LineChart data={chartData} margin={{ left: 8, right: 8, top: 8, bottom: 8 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.08)" />
                  <XAxis dataKey="label" stroke="#9fb4d3" tick={{ fontSize: 12 }} minTickGap={24} />
                  <YAxis stroke="#9fb4d3" tick={{ fontSize: 12 }} domain={[0, 'dataMax + 40']} />
                  <Tooltip
                    contentStyle={{ background: '#0f1623', border: '1px solid rgba(255,255,255,0.1)' }}
                    labelStyle={{ color: '#cbd5e5' }}
                  />
                  <Legend />
                  <Area type="monotone" dataKey="solar_mw" stackId="supply" stroke="#f2c84b" fill="#f2c84b33" name="Solar" />
                  <Area type="monotone" dataKey="wind_mw" stackId="supply" stroke="#7cf0a3" fill="#7cf0a333" name="Wind" />
                  <Area type="monotone" dataKey="gas_mw" stackId="supply" stroke="#5bd8ff" fill="#5bd8ff33" name="Gas" />
                  <Line type="monotone" dataKey="demand_mw" stroke="#ff9f7b" strokeWidth={2.5} dot={false} name="Demand" />
                  <Line type="monotone" dataKey="supply_mw" stroke="#2f7cf3" strokeWidth={2} dot={false} name="Supply" />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </article>

          <article className="panel">
            <div className="panel-heading">
              <div>
                <p className="eyebrow">Grid stability</p>
                <h2>Frequency gauge</h2>
                <p className="panel-subtitle">Dynamic color and alerting when frequency drifts low.</p>
              </div>
              <span className={`mode-pill ${snapshot.controller.manual_enabled ? 'manual' : 'pid'}`}>
                {snapshot.controller.manual_enabled ? 'Manual mode' : 'PID mode'}
              </span>
            </div>
            <div className="gauge">
              <div className="gauge-ring" style={{ borderColor: `${freqColor}55`, boxShadow: `0 0 30px ${freqColor}33` }}>
                <div className="gauge-needle" style={{ transform: `rotate(${rotation}deg)`, background: freqColor }} />
                <div className="gauge-center" style={{ color: freqColor }}>
                  {snapshot.frequency_hz.toFixed(2)} Hz
                </div>
                <div className="gauge-band" />
              </div>
              <p className="gauge-footnote">Target {controllerSettings.control.target_frequency_hz.toFixed(2)} Hz</p>
            </div>
          </article>

          <article className="panel">
            <div className="panel-heading">
              <div>
                <p className="eyebrow">Generation</p>
                <h2>Supply stack</h2>
                <p className="panel-subtitle">Live generator outputs and demand summary.</p>
              </div>
              <span className="pill subtle">Net {snapshot.net_balance_mw.toFixed(1)} MW</span>
            </div>
            <div className="status-grid">
              {[
                { label: 'Solar', value: snapshot.solar_mw, accent: '#f2c84b', detail: 'Irradiance-driven PV' },
                { label: 'Wind', value: snapshot.wind_mw, accent: '#7cf0a3', detail: 'Turbines online' },
                {
                  label: 'Gas',
                  value: snapshot.gas_mw,
                  accent: '#5bd8ff',
                  detail: snapshot.controller.manual_enabled ? 'Manual dispatch' : 'PID controlled',
                },
              ].map((gen) => (
                <div key={gen.label} className="status-card">
                  <div className="status-header">
                    <span className="status-dot" style={{ background: gen.accent }} />
                    <p className="status-label">{gen.label}</p>
                  </div>
                  <p className="status-value">{gen.value.toFixed(1)} MW</p>
                  <p className="status-detail">{gen.detail}</p>
                </div>
              ))}
              <div className="status-card demand">
                <div className="status-header">
                  <span className="status-dot demand" />
                  <p className="status-label">Demand</p>
                </div>
                <p className="status-value">{snapshot.demand_mw.toFixed(1)} MW</p>
                <p className="status-detail">Consumers + industrial load</p>
              </div>
            </div>
          </article>

          <article className="panel">
            <div className="panel-heading">
              <div>
                <p className="eyebrow">Controller</p>
                <h2>PID + manual dispatch</h2>
                <p className="panel-subtitle">Live PID output, manual override flag, and gas capacity.</p>
              </div>
              <span className="pill">Gas {controllerSettings.gas.min_mw}–{controllerSettings.gas.capacity_mw} MW</span>
            </div>
            <div className="controls-grid">
              <div className="control-tile">
                <p className="control-label">Mode</p>
                <p className="control-value">{snapshot.controller.manual_enabled ? 'Manual' : 'PID'}</p>
                <p className="control-helper">
                  {snapshot.controller.manual_enabled
                    ? `Setpoint ${snapshot.controller.manual_setpoint_mw.toFixed(1)} MW`
                    : `PID output ${snapshot.controller.setpoint_mw.toFixed(1)} MW`}
                </p>
              </div>
              <div className="control-tile">
                <p className="control-label">PID tuning</p>
                <p className="control-value">
                  kp {controllerSettings.control.pid.kp}, ki {controllerSettings.control.pid.ki}, kd{' '}
                  {controllerSettings.control.pid.kd}
                </p>
                <p className="control-helper">Integral {snapshot.controller.integral.toFixed(2)}</p>
              </div>
              <div className="control-tile">
                <p className="control-label">Frequency error</p>
                <p className="control-value">{snapshot.controller.error.toFixed(3)}</p>
                <p className="control-helper">Target {controllerSettings.control.target_frequency_hz.toFixed(2)} Hz</p>
              </div>
            </div>
          </article>

          <article className="panel">
            <div className="panel-heading">
              <div>
                <p className="eyebrow">Streaming</p>
                <h2>Status events</h2>
                <p className="panel-subtitle">Latest ticks from /stream rendered as they arrive.</p>
              </div>
              <span className="pill subtle">{history.length} pts buffered</span>
            </div>
            <div className="stream-list">
              {recentEvents.map((item) => (
                <div key={item.tick_count} className="stream-item">
                  <span className="bullet" />
                  <div>
                    <p className="stream-title">Tick #{item.tick_count}</p>
                    <p className="stream-meta">{new Date(item.timestamp).toLocaleTimeString()}</p>
                  </div>
                  <div className="stream-values">
                    {item.net_balance_mw.toFixed(1)} MW net · {item.frequency_hz.toFixed(2)} Hz
                  </div>
                </div>
              ))}
            </div>
          </article>
        </section>
      </main>
    </div>
  )
}

export default App
