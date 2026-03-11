import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { createError, getErrors } from '../api/client';
import { StatePanel } from '../components/ui/StatePanel';
import { StatusBadge } from '../components/ui/StatusBadge';
import { formatDateTime } from '../utils/format';

const severityOptions = ['all', 'critical', 'high', 'medium', 'low'] as const;
const statusOptions = ['all', 'open', 'investigating', 'resolved'] as const;

export function ErrorExplorerPage() {
  const [severity, setSeverity] = useState<(typeof severityOptions)[number]>('all');
  const [status, setStatus] = useState<(typeof statusOptions)[number]>('all');
  const [query, setQuery] = useState('');
  const [page, setPage] = useState(1);
  const [newMessage, setNewMessage] = useState('');
  const [newSeverity, setNewSeverity] = useState<'critical' | 'high' | 'medium' | 'low'>('high');

  const queryClient = useQueryClient();
  const errorQuery = useQuery({
    queryKey: ['errors', severity, status, query, page],
    queryFn: () =>
      getErrors({
        severity: severity === 'all' ? '' : severity,
        status: status === 'all' ? '' : status,
        q: query,
        page,
        pageSize: 12,
      }),
    refetchInterval: 8000,
  });

  const createMutation = useMutation({
    mutationFn: () =>
      createError({
        severity: newSeverity,
        status: 'open',
        message: newMessage,
      }),
    onSuccess: async () => {
      setNewMessage('');
      await queryClient.invalidateQueries({ queryKey: ['errors'] });
      await queryClient.invalidateQueries({ queryKey: ['overview'] });
    },
  });

  const tableRows = useMemo(() => errorQuery.data?.rows ?? [], [errorQuery.data?.rows]);

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h2>Error Explorer</h2>
          <p>Filterable incident list to support post-mortem clarity and stakeholder updates.</p>
        </div>
      </header>

      <section className="panel filters-panel">
        <div className="filters-row">
          <label>
            Severity
            <select value={severity} onChange={(event) => { setSeverity(event.target.value as typeof severity); setPage(1); }}>
              {severityOptions.map((option) => (
                <option key={option} value={option}>
                  {option}
                </option>
              ))}
            </select>
          </label>
          <label>
            Status
            <select value={status} onChange={(event) => { setStatus(event.target.value as typeof status); setPage(1); }}>
              {statusOptions.map((option) => (
                <option key={option} value={option}>
                  {option}
                </option>
              ))}
            </select>
          </label>
          <label className="filters-row__search">
            Search
            <input
              placeholder="service, message..."
              value={query}
              onChange={(event) => {
                setQuery(event.target.value);
                setPage(1);
              }}
            />
          </label>
        </div>
      </section>

      <section className="panel">
        <div className="panel__header">
          <h3>Incident feed</h3>
          <p>{errorQuery.data?.pagination.total ?? 0} matching records</p>
        </div>

        {errorQuery.isLoading ? <StatePanel title="Loading incidents" message="Retrieving incident rows..." variant="loading" /> : null}
        {errorQuery.isError ? <StatePanel title="Incident query failed" message="Could not fetch incident data." variant="error" /> : null}

        {!errorQuery.isLoading && !errorQuery.isError && tableRows.length === 0 ? (
          <StatePanel title="No incidents found" message="Try widening filters or query text." variant="empty" />
        ) : null}

        {tableRows.length > 0 ? (
          <div className="table-wrapper">
            <table>
              <thead>
                <tr>
                  <th>Created</th>
                  <th>Severity</th>
                  <th>Status</th>
                  <th>Service</th>
                  <th>Message</th>
                  <th>Trace</th>
                </tr>
              </thead>
              <tbody>
                {tableRows.map((row) => (
                  <tr key={row.id}>
                    <td>{formatDateTime(row.createdAt)}</td>
                    <td>
                      <StatusBadge
                        tone={row.severity === 'critical' ? 'error' : row.severity === 'high' ? 'warning' : 'neutral'}
                        label={row.severity}
                      />
                    </td>
                    <td>{row.status}</td>
                    <td>{row.service}</td>
                    <td>{row.message}</td>
                    <td>{row.traceId ? row.traceId.slice(0, 16) : 'n/a'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}

        {errorQuery.data?.pagination.totalPages && errorQuery.data.pagination.totalPages > 1 ? (
          <div className="pagination">
            <button type="button" disabled={page <= 1} onClick={() => setPage((value) => value - 1)}>
              Previous
            </button>
            <span>
              Page {errorQuery.data.pagination.page} / {errorQuery.data.pagination.totalPages}
            </span>
            <button
              type="button"
              disabled={page >= errorQuery.data.pagination.totalPages}
              onClick={() => setPage((value) => value + 1)}
            >
              Next
            </button>
          </div>
        ) : null}
      </section>

      <section className="panel">
        <div className="panel__header">
          <h3>Create demo incident</h3>
          <p>Inject realistic errors for stakeholder demos.</p>
        </div>
        <form
          className="inline-form"
          onSubmit={(event) => {
            event.preventDefault();
            if (newMessage.trim().length < 8) return;
            createMutation.mutate();
          }}
        >
          <label>
            Severity
            <select value={newSeverity} onChange={(event) => setNewSeverity(event.target.value as typeof newSeverity)}>
              <option value="critical">critical</option>
              <option value="high">high</option>
              <option value="medium">medium</option>
              <option value="low">low</option>
            </select>
          </label>
          <label className="inline-form__grow">
            Message
            <input
              value={newMessage}
              onChange={(event) => setNewMessage(event.target.value)}
              placeholder="API timeout for enterprise tenant"
            />
          </label>
          <button type="submit" className="btn-primary" disabled={createMutation.isPending}>
            {createMutation.isPending ? 'Creating...' : 'Add incident'}
          </button>
        </form>
        {createMutation.isError ? <p className="hint">Incident creation failed. Check API/auth settings.</p> : null}
      </section>
    </section>
  );
}
