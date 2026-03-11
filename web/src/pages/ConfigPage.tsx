import { useQuery } from '@tanstack/react-query';

import { getConfigStatus } from '../api/client';
import { StatePanel } from '../components/ui/StatePanel';
import { StatusBadge } from '../components/ui/StatusBadge';

const USE_MOCKS = import.meta.env.VITE_USE_MOCKS === 'true';

async function getHealthReady(): Promise<{ status: string; reason?: string }> {
  if (USE_MOCKS) {
    return { status: 'ready' };
  }
  const response = await fetch('/health/ready');
  return response.json();
}

export function ConfigPage() {
  const configQuery = useQuery({
    queryKey: ['config'],
    queryFn: getConfigStatus,
    refetchInterval: 15000,
  });

  const healthQuery = useQuery({
    queryKey: ['health-ready'],
    queryFn: getHealthReady,
    refetchInterval: 10000,
  });

  if (configQuery.isLoading) {
    return <StatePanel title="Loading configuration" message="Gathering runtime status..." variant="loading" />;
  }

  if (configQuery.isError || !configQuery.data) {
    return <StatePanel title="Config unavailable" message="Could not read runtime configuration status." variant="error" />;
  }

  const config = configQuery.data;

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h2>Config & Environment Status</h2>
          <p>Sanitized runtime configuration to validate deployment integrity.</p>
        </div>
        {healthQuery.data ? (
          <StatusBadge tone={healthQuery.data.status === 'ready' ? 'healthy' : 'degraded'} label={healthQuery.data.status} />
        ) : null}
      </header>

      <section className="panel">
        <div className="panel__header">
          <h3>Service profile</h3>
        </div>
        <div className="kv-grid">
          <div><span>Service</span><strong>{config.service.name}</strong></div>
          <div><span>Environment</span><strong>{config.service.environment}</strong></div>
          <div><span>Version</span><strong>{config.service.version}</strong></div>
          <div><span>Database configured</span><strong>{String(config.database.configured)}</strong></div>
          <div><span>API key auth</span><strong>{String(config.auth.apiKeyEnabled)}</strong></div>
          <div><span>JWT auth</span><strong>{String(config.auth.jwtEnabled)}</strong></div>
          <div><span>Rate limit</span><strong>{config.rateLimit.requestsPerSecond}/s burst {config.rateLimit.burst}</strong></div>
          <div><span>OTLP endpoint</span><strong>{config.telemetry.otlpEndpoint}</strong></div>
        </div>
      </section>

      <section className="panel">
        <div className="panel__header">
          <h3>Allowed origins</h3>
        </div>
        {config.allowedOrigins.length === 0 ? (
          <StatePanel title="No origins configured" message="CORS is currently unrestricted or disabled." variant="warning" />
        ) : (
          <ul className="pill-list">
            {config.allowedOrigins.map((origin) => (
              <li key={origin} className="pill">{origin}</li>
            ))}
          </ul>
        )}
      </section>

      {healthQuery.data?.reason ? <p className="hint">Readiness reason: {healthQuery.data.reason}</p> : null}
    </section>
  );
}
