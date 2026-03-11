import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Bar, BarChart, CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

import { getLiveMetrics } from '../api/client';
import { MetricCard } from '../components/ui/MetricCard';
import { StatePanel } from '../components/ui/StatePanel';
import { formatPercent } from '../utils/format';

const windows = [
  { label: '15m', value: '15m' },
  { label: '30m', value: '30m' },
  { label: '1h', value: '1h' },
];

export function LiveMetricsPage() {
  const [window, setWindow] = useState('15m');
  const metricsQuery = useQuery({
    queryKey: ['live-metrics', window],
    queryFn: () => getLiveMetrics(window),
    refetchInterval: 5000,
  });

  const chartData = useMemo(
    () =>
      (metricsQuery.data?.series ?? []).map((point) => ({
        time: new Date(point.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
        throughput: point.throughput,
        errorRate: Number((point.errorRate * 100).toFixed(2)),
        p95: Number(point.p95LatencyMs.toFixed(0)),
      })),
    [metricsQuery.data?.series],
  );

  if (metricsQuery.isLoading) {
    return <StatePanel title="Loading live metrics" message="Streaming traffic and latency metrics..." variant="loading" />;
  }
  if (metricsQuery.isError || !metricsQuery.data) {
    return <StatePanel title="Metrics unavailable" message="Unable to retrieve live metric snapshots." variant="error" />;
  }

  const metrics = metricsQuery.data;

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h2>Live Metrics</h2>
          <p>Continuous visibility over throughput, latency and error behavior.</p>
        </div>
        <div className="control-row">
          {windows.map((item) => (
            <button
              key={item.value}
              className={item.value === window ? 'chip chip--active' : 'chip'}
              onClick={() => setWindow(item.value)}
              type="button"
            >
              {item.label}
            </button>
          ))}
        </div>
      </header>

      <div className="metrics-grid">
        <MetricCard title="Requests" value={String(metrics.requests)} subtitle={`window ${window}`} />
        <MetricCard title="RPS" value={metrics.rps.toFixed(2)} subtitle="average" />
        <MetricCard title="Error rate" value={formatPercent(metrics.errorRate)} subtitle="5xx responses" />
        <MetricCard title="P95 latency" value={`${metrics.p95LatencyMs.toFixed(0)} ms`} subtitle="tail latency" />
      </div>

      <section className="panel chart-panel">
        <div className="panel__header">
          <h3>Throughput and latency trend</h3>
          <p>Spikes can be correlated with traces and incident records.</p>
        </div>
        {chartData.length === 0 ? (
          <StatePanel title="No traffic in this window" message="Send requests to populate this chart." variant="empty" />
        ) : (
          <ResponsiveContainer width="100%" height={320}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" minTickGap={24} />
              <YAxis yAxisId="left" />
              <YAxis yAxisId="right" orientation="right" />
              <Tooltip />
              <Bar yAxisId="left" dataKey="throughput" fill="#0f766e" radius={[4, 4, 0, 0]} />
              <Line yAxisId="right" type="monotone" dataKey="p95" stroke="#ea580c" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        )}
      </section>

      <section className="panel chart-panel">
        <div className="panel__header">
          <h3>Status code distribution</h3>
          <p>Response quality profile for the selected window.</p>
        </div>
        {metrics.statusCodes.length === 0 ? (
          <StatePanel title="No status code data" message="No requests available for this interval." variant="empty" />
        ) : (
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={metrics.statusCodes}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="code" />
              <YAxis />
              <Tooltip />
              <Bar dataKey="count" fill="#1d4ed8" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        )}
      </section>
    </section>
  );
}
