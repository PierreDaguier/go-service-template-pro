import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';

import { getTraces } from '../api/client';
import { StatePanel } from '../components/ui/StatePanel';
import { StatusBadge } from '../components/ui/StatusBadge';
import { formatDateTime, formatLatency } from '../utils/format';

export function TraceExplorerPage() {
  const [query, setQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'success' | 'error'>('all');

  const tracesQuery = useQuery({
    queryKey: ['traces'],
    queryFn: () => getTraces(150),
    refetchInterval: 5000,
  });

  const filtered = useMemo(() => {
    const rows = tracesQuery.data ?? [];
    return rows.filter((row) => {
      const matchesQuery = query.trim() === '' || row.route.toLowerCase().includes(query.toLowerCase()) || row.traceId.includes(query);
      const matchesStatus =
        statusFilter === 'all' || (statusFilter === 'error' ? row.statusCode >= 500 : row.statusCode < 500);
      return matchesQuery && matchesStatus;
    });
  }, [tracesQuery.data, query, statusFilter]);

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h2>Trace Explorer</h2>
          <p>Simplified trace stream for quick isolation of latency and failure paths.</p>
        </div>
      </header>

      <section className="panel filters-panel">
        <div className="filters-row">
          <label className="filters-row__search">
            Route or trace id
            <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="/api/v1/errors or trace id" />
          </label>
          <label>
            Status
            <select value={statusFilter} onChange={(event) => setStatusFilter(event.target.value as typeof statusFilter)}>
              <option value="all">all</option>
              <option value="success">success</option>
              <option value="error">error</option>
            </select>
          </label>
        </div>
      </section>

      <section className="panel">
        {tracesQuery.isLoading ? <StatePanel title="Loading traces" message="Collecting recent spans..." variant="loading" /> : null}
        {tracesQuery.isError ? <StatePanel title="Trace stream unavailable" message="Unable to retrieve trace records." variant="error" /> : null}

        {!tracesQuery.isLoading && !tracesQuery.isError && filtered.length === 0 ? (
          <StatePanel title="No traces match current filters" message="Adjust filters or generate API traffic." variant="empty" />
        ) : null}

        {filtered.length > 0 ? (
          <div className="table-wrapper">
            <table>
              <thead>
                <tr>
                  <th>Timestamp</th>
                  <th>Status</th>
                  <th>Route</th>
                  <th>Duration</th>
                  <th>Trace ID</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((trace) => (
                  <tr key={`${trace.traceId}-${trace.timestamp}`}>
                    <td>{formatDateTime(trace.timestamp)}</td>
                    <td>
                      <StatusBadge tone={trace.statusCode >= 500 ? 'error' : 'healthy'} label={String(trace.statusCode)} />
                    </td>
                    <td>{trace.method} {trace.route}</td>
                    <td>{formatLatency(trace.durationMs)}</td>
                    <td className="mono">{trace.traceId || 'n/a'}</td>
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
