import { useEffect, useRef, useState } from 'react'
import { fetchHealth, fetchTelemetryRange, websocketURL } from '../api/client'
import type { TelemetryPoint, WsEnvelope } from '../api/types'

export type StreamStatus = 'connecting' | 'live' | 'reconnecting' | 'error'

const MAX_POINTS = 2000

function mergePoint(points: TelemetryPoint[], next: TelemetryPoint): TelemetryPoint[] {
  const last = points[points.length - 1]
  if (last && last.time === next.time && last.bit_depth === next.bit_depth) {
    return points
  }
  const merged = [...points, next]
  if (merged.length > MAX_POINTS) {
    return merged.slice(merged.length - MAX_POINTS)
  }
  return merged
}

export function useTelemetryStream(wellID: string) {
  const [points, setPoints] = useState<TelemetryPoint[]>([])
  const [status, setStatus] = useState<StreamStatus>('connecting')
  const [error, setError] = useState<string | null>(null)
  const [health, setHealth] = useState<string>('unknown')
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    let cancelled = false
    let retryTimer: number | undefined
    let pingTimer: number | undefined

    async function bootstrap() {
      setStatus('connecting')
      setError(null)
      try {
        const h = await fetchHealth()
        if (!cancelled) setHealth(`${h.status} (db=${h.db}, mqtt=${h.mqtt})`)
        const to = new Date()
        const from = new Date(to.getTime() - 60 * 60 * 1000)
        const history = await fetchTelemetryRange(wellID, from, to)
        if (!cancelled) setPoints(history)
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'bootstrap failed')
          setStatus('error')
        }
      }
      if (!cancelled) connectWS()
    }

    function connectWS() {
      if (cancelled) return
      const ws = new WebSocket(websocketURL(wellID))
      wsRef.current = ws

      ws.onopen = () => {
        if (cancelled) return
        setStatus('live')
        pingTimer = window.setInterval(() => {
          if (ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'ping' }))
          }
        }, 25000)
      }

      ws.onmessage = (ev) => {
        try {
          const msg = JSON.parse(ev.data as string) as WsEnvelope
          if (msg.type === 'point' && msg.data) {
            setPoints((prev) => mergePoint(prev, msg.data!))
          }
        } catch {
          // ignore malformed
        }
      }

      ws.onerror = () => {
        if (!cancelled) setStatus('error')
      }

      ws.onclose = () => {
        if (pingTimer) window.clearInterval(pingTimer)
        if (cancelled) return
        setStatus('reconnecting')
        retryTimer = window.setTimeout(connectWS, 2000)
      }
    }

    void bootstrap()

    return () => {
      cancelled = true
      if (retryTimer) window.clearTimeout(retryTimer)
      if (pingTimer) window.clearInterval(pingTimer)
      wsRef.current?.close()
    }
  }, [wellID])

  const latest = points.length > 0 ? points[points.length - 1] : null
  return { points, latest, status, error, health }
}
