export type TelemetryPoint = {
  time: string
  well_id: string
  bit_depth: number
  rop: number
  wob: number
  gamma_ray: number
}

export type TelemetryRangeResponse = {
  well_id: string
  from: string
  to: string
  points: TelemetryPoint[]
}

export type HealthResponse = {
  status: string
  db: string
  mqtt: string
}

export type WsEnvelope = {
  type: string
  well_id?: string
  data?: TelemetryPoint
  message?: string
}
