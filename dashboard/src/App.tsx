import { Activity } from 'lucide-react'
import { WellLogChart } from './components/WellLogChart'
import { useTelemetryStream } from './hooks/useTelemetryStream'

const WELL_ID = import.meta.env.VITE_WELL_ID ?? 'WELL-DEMO-01'

function App() {
  const { points, latest, status, error, health } = useTelemetryStream(WELL_ID)

  return (
    <div className="min-h-screen bg-[#0b1220] px-6 py-8">
      <header className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div className="flex items-center gap-3">
          <Activity className="h-7 w-7 text-emerald-400" />
          <div>
            <h1 className="text-2xl font-semibold tracking-tight text-slate-100">
              Rig Telemetry Dashboard
            </h1>
            <p className="text-sm text-slate-400">
              Live well-log strip chart — {WELL_ID}
            </p>
          </div>
        </div>
        <div className="text-right text-sm text-slate-400">
          <div>
            Stream:{' '}
            <span className="text-emerald-400 font-medium">{status}</span>
          </div>
          <div>API health: {health}</div>
          {error && <div className="text-rose-400">{error}</div>}
        </div>
      </header>

      <div className="mb-4 flex flex-wrap gap-6 text-sm text-slate-300">
        <Metric label="Bit Depth" value={latest ? `${latest.bit_depth.toFixed(2)} m` : '—'} />
        <Metric label="ROP" value={latest ? `${latest.rop.toFixed(2)}` : '—'} />
        <Metric label="WOB" value={latest ? `${latest.wob.toFixed(2)}` : '—'} />
        <Metric label="Gamma Ray" value={latest ? `${latest.gamma_ray.toFixed(2)} API` : '—'} />
        <Metric label="Points" value={String(points.length)} />
      </div>

      <WellLogChart points={points} />
    </div>
  )
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-xs uppercase tracking-wide text-slate-500">{label}</div>
      <div className="font-medium text-slate-100">{value}</div>
    </div>
  )
}

export default App
