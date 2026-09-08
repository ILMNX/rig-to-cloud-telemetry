import ReactECharts from 'echarts-for-react'
import type { TelemetryPoint } from '../api/types'

type Props = {
  points: TelemetryPoint[]
}

export function WellLogChart({ points }: Props) {
  const gammaSeries = points.map((p) => [p.gamma_ray, p.bit_depth])

  const option = {
    backgroundColor: 'transparent',
    animation: false,
    grid: { left: 64, right: 28, top: 28, bottom: 48 },
    tooltip: {
      trigger: 'axis',
      formatter: (params: Array<{ data: [number, number] }>) => {
        const d = params?.[0]?.data
        if (!d) return ''
        return `Gamma: ${d[0].toFixed(2)} API<br/>Depth: ${d[1].toFixed(2)} m`
      },
    },
    xAxis: {
      type: 'value',
      name: 'Gamma Ray (API)',
      nameLocation: 'middle',
      nameGap: 30,
      axisLabel: { color: '#9fb0c3' },
      splitLine: { lineStyle: { color: '#1e2a3a' } },
    },
    yAxis: {
      type: 'value',
      name: 'Bit Depth (m)',
      inverse: true,
      scale: true,
      axisLabel: { color: '#9fb0c3' },
      splitLine: { lineStyle: { color: '#1e2a3a' } },
    },
    series: [
      {
        name: 'Gamma Ray',
        type: 'line',
        showSymbol: false,
        data: gammaSeries,
        lineStyle: { width: 2, color: '#3dd68c' },
        areaStyle: { color: 'rgba(61, 214, 140, 0.12)' },
      },
    ],
  }

  return (
    <div className="h-[70vh] w-full max-w-3xl rounded-lg border border-slate-800 bg-slate-950/60 p-2">
      <ReactECharts option={option} style={{ height: '100%', width: '100%' }} notMerge />
    </div>
  )
}
