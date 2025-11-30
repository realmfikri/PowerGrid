export interface PIDConfig {
  kp: number
  ki: number
  kd: number
  integral_min: number
  integral_max: number
}

export interface ControlConfig {
  target_frequency_hz: number
  pid: PIDConfig
}

export interface GasConfig {
  capacity_mw: number
  min_mw: number
  ramp_mw: number
}

export interface SolarConfig {
  peak_mw: number
  sunrise_hour: number
  sunset_hour: number
}

export interface WindConfig {
  base_mw: number
  variance: number
  smoothing: number
}

export interface HouseConfig {
  count: number
  base_demand_kw: number
  variance: number
  peak_hour: number
  peak_hour_width: number
}

export interface GridConfig {
  base_frequency_hz: number
  sensitivity_hz: number
}

export interface GeneratorConfig {
  solar: SolarConfig
  wind: WindConfig
  gas: GasConfig
  houses: HouseConfig
}

export interface ControllerSnapshot {
  manual_enabled: boolean
  manual_setpoint_mw: number
  setpoint_mw: number
  integral: number
  error: number
}

export interface SimulationSnapshot {
  tick_count: number
  timestamp: string
  solar_mw: number
  wind_mw: number
  gas_mw: number
  supply_mw: number
  demand_mw: number
  net_balance_mw: number
  frequency_hz: number
  controller: ControllerSnapshot
}

export interface ControllerSettingsResponse {
  control: ControlConfig
  gas: GasConfig
  controller: ControllerSnapshot
}

export interface ManualGasRequest {
  enable?: boolean
  setpoint_mw?: number
}
