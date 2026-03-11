import { useQuery } from '@tanstack/react-query';

import { getLogs, getOverview } from '../api/client';
import { MetricCard } from '../components/ui/MetricCard';
import { StatePanel } from '../components/ui/StatePanel';
import { StatusBadge } from '../components/ui/StatusBadge';
import { formatDateTime, formatPercent, formatUptime } from '../utils/format';

export function OverviewPage() {
  const overviewQuery = useQuery({
    queryKey: ['overview'],
    queryFn: getOverview,
    refetchInterval: 7000,
  });

  const logsQuery = useQuery({
    queryKey: ['logs-overview'],
    queryFn: () => getLogs('', 8),
    refetchInterval: 7000,
  });

  if (overviewQuery.isLoading) {
    return <StatePanel title="Loading overview" message="Collecting service heartbeat and KPI data..." variant="loading" />;
  }

  if (overviewQuery.isError || !overviewQuery.data) {
    return <StatePanel title="Overview unavailable" message="The API did not return service overview data." variant="error" />;
  }

  const overview = overviewQuery.data;

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h2>Service Overview</h2>
          <p>High-level reliability posture with context understandable by technical and business stakeholders.</p>
        </div>
        <StatusBadge tone={overview.systemStatus === 'healthy' ? 'healthy' : 'degraded'} label={overview.systemStatus} />
      </header>

      <div className="metrics-grid">
        <MetricCard title="Service" value={overview.serviceName} subtitle={`${overview.environment} · v${overview.version}`} />
        <MetricCard title="Uptime" value={formatUptime(overview.uptimeSeconds)} subtitle="Since last restart" />
        <MetricCard title="Open incidents" value={String(overview.openIncidents)} subtitle="Current unresolved issues" />
        <MetricCard title="Critical incidents" value={String(overview.criticalIncidents)} subtitle="Escalation required" />
        <MetricCard title="Error rate" value={formatPercent(overview.traffic.errorRate)} subtitle="Last 15 minutes" />
        <MetricCard title="P95 latency" value={`${overview.traffic.p95LatencyMs.toFixed(0)} ms`} subtitle="Last 15 minutes" />
      </div>

      <section className="panel">
        <div className="panel__header">
          <h3>Recent operational logs</h3>
          <p>Fast signal for support conversations with clients.</p>
        </div>

        {logsQuery.isLoading ? <p className="hint">Loading logs...</p> : null}
        {logsQuery.isError ? <p className="hint">Unable to fetch logs.</p> : null}

        {logsQuery.data && logsQuery.data.length > 0 ? (
          <div className="table-wrapper">
            <table>
              <thead>
                <tr>
                  <th>Timestamp</th>
                  <th>Level</th>
                  <th>Message</th>
                </tr>
              </thead>
              <tbody>
                {logsQuery.data.map((entry) => (
                  <tr key={`${entry.timestamp}-${entry.message}`}>
                    <td>{formatDateTime(entry.timestamp)}</td>
                    <td>
                      <StatusBadge
                        tone={entry.level === 'error' ? 'error' : entry.level === 'warn' ? 'warning' : 'neutral'}
                        label={entry.level}
                      />
                    </td>
                    <td>{entry.message}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </section>
    </section>
  );
}
