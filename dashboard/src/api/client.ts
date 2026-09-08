import type { HealthResponse, TelemetryPoint, TelemetryRangeResponse } from './types'

const API_BASE = import.meta.env.VITE_API_BASE ?? 'http://localhost:8080'

export async function fetchHealth(): Promise<HealthResponse> {
  const res = await fetch(`${API_BASE}/api/v1/health`)
  if (!res.ok) throw new Error(`health ${res.status}`)
  return res.json()
}

export async function fetchTelemetryRange(
  wellID: string,
  from: Date,
  to: Date,
  limit = 5000,
): Promise<TelemetryPoint[]> {
  const params = new URLSearchParams({
    from: from.toISOString(),
    to: to.toISOString(),
    limit: String(limit),
  })
  const res = await fetch(
    `${API_BASE}/api/v1/wells/${encodeURIComponent(wellID)}/telemetry?${params}`,
  )
  if (!res.ok) throw new Error(`telemetry ${res.status}`)
  const body = (await res.json()) as TelemetryRangeResponse
  return body.points ?? []
}

export function websocketURL(wellID: string): string {
  const base = API_BASE.replace(/^http/, 'ws')
  return `${base}/api/v1/ws/telemetry?well_id=${encodeURIComponent(wellID)}`
}
