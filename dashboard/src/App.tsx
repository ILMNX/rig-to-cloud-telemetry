import { Activity } from 'lucide-react'
import ReactECharts from 'echarts-for-react'

const demoDepths = Array.from({ length: 40 }, (_, i) => 1000 + i * 5)
const demoGamma = demoDepths.map((_, i) => 60 + 30 * Math.sin(i / 4) + (i % 7))

function App() {
  const option = {
    backgroundColor: 'transparent',
    grid: { left: 56, right: 24, top: 24, bottom: 40 },
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'value',
      name: 'Gamma Ray (API)',
      nameLocation: 'middle',
      nameGap: 28,
      axisLabel: { color: '#9fb0c3' },
      splitLine: { lineStyle: { color: '#1e2a3a' } },
    },
    yAxis: {
      type: 'value',
      name: 'Bit Depth (m)',
      inverse: true,
      axisLabel: { color: '#9fb0c3' },
      splitLine: { lineStyle: { color: '#1e2a3a' } },
    },
    series: [
      {
        name: 'Gamma Ray',
        type: 'line',
        showSymbol: false,
        data: demoDepths.map((d, i) => [demoGamma[i], d]),
        lineStyle: { width: 2, color: '#3dd68c' },
        areaStyle: { color: 'rgba(61, 214, 140, 0.12)' },
      },
    ],
  }

  return (
    <div className="min-h-screen bg-[#0b1220] px-6 py-8">
      <header className="mb-6 flex items-center gap-3">
        <Activity className="h-7 w-7 text-emerald-400" />
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-slate-100">
            Rig Telemetry Dashboard
          </h1>
          <p className="text-sm text-slate-400">
            Vertical well-log strip chart (demo data)
          </p>
        </div>
      </header>
      <div className="h-[70vh] w-full max-w-3xl rounded-lg border border-slate-800 bg-slate-950/60 p-2">
        <ReactECharts option={option} style={{ height: '100%', width: '100%' }} />
      </div>
    </div>
  )
}

export default App
