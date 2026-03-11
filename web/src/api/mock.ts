import type {
  ConfigStatus,
  Incident,
  IncidentPagination,
  LogEntry,
  MetricsSnapshot,
  Overview,
  TraceRecord,
} from '../types/ops';

const now = Date.now();

export const mockMetrics: MetricsSnapshot = {
  windowSeconds: 900,
  requests: 482,
  rps: 0.54,
  errorRate: 0.024,
  p95LatencyMs: 146,
  series: Array.from({ length: 15 }).map((_, index) => {
    const timestamp = new Date(now - (15 - index) * 60_000).toISOString();
    return {
      timestamp,
      throughput: 18 + (index % 5) * 2,
      errorRate: index % 6 === 0 ? 0.08 : 0.01,
      p95LatencyMs: 95 + (index % 4) * 22,
    };
  }),
  statusCodes: [
    { code: 200, count: 443 },
    { code: 401, count: 17 },
    { code: 429, count: 10 },
    { code: 500, count: 12 },
  ],
  topRoutes: [
    { route: 'GET /api/v1/overview', count: 130 },
    { route: 'GET /api/v1/metrics/live', count: 122 },
    { route: 'GET /api/v1/errors', count: 91 },
    { route: 'GET /api/v1/traces', count: 72 },
  ],
};

export const mockOverview: Overview = {
  serviceName: 'go-service-template-pro',
  environment: 'staging',
  version: '1.0.0',
  uptimeSeconds: 158_400,
  openIncidents: 4,
  criticalIncidents: 1,
  systemStatus: 'degraded',
  traffic: mockMetrics,
};

export const mockIncidents: Incident[] = [
  {
    id: 'e8f799af-9409-4c5e-a47e-8d37be4a10f1',
    service: 'payments-gateway',
    environment: 'staging',
    severity: 'critical',
    status: 'open',
    message: 'Gateway timeout spike above SLO for enterprise tenants',
    traceId: '2f6f7ca62f9a42218ed6a481ee56bbf8',
    createdAt: new Date(now - 11 * 60_000).toISOString(),
  },
  {
    id: '2809b863-eef0-4f44-80b8-42efeb7f75eb',
    service: 'auth-service',
    environment: 'staging',
    severity: 'high',
    status: 'investigating',
    message: 'JWT signing key rotation created transient token validation failures',
    traceId: '6d9ad2f3ccce4141ab706a4a9fd4d337',
    createdAt: new Date(now - 44 * 60_000).toISOString(),
  },
  {
    id: 'af7f06ee-4d8e-406a-8d8f-832dd2f74f14',
    service: 'metrics-exporter',
    environment: 'staging',
    severity: 'medium',
    status: 'resolved',
    message: 'Prometheus scrape errors recovered after endpoint timeout adjustment',
    createdAt: new Date(now - 110 * 60_000).toISOString(),
    resolvedAt: new Date(now - 94 * 60_000).toISOString(),
  },
];

export const mockIncidentsPagination: IncidentPagination = {
  page: 1,
  pageSize: 12,
  total: mockIncidents.length,
  totalPages: 1,
};

export const mockTraces: TraceRecord[] = Array.from({ length: 24 }).map((_, idx) => ({
  timestamp: new Date(now - idx * 45_000).toISOString(),
  traceId: `trace-${idx.toString().padStart(4, '0')}-7f2a`,
  method: idx % 5 === 0 ? 'POST' : 'GET',
  route: idx % 5 === 0 ? '/api/v1/errors' : '/api/v1/overview',
  statusCode: idx % 6 === 0 ? 502 : 200,
  durationMs: idx % 6 === 0 ? 320 + idx * 3 : 56 + idx * 2,
  source: 'mock',
}));

export const mockLogs: LogEntry[] = [
  {
    timestamp: new Date(now - 23_000).toISOString(),
    level: 'error',
    message: 'request completed',
    fields: { route: '/api/v1/errors', status: 502, duration_ms: 347 },
  },
  {
    timestamp: new Date(now - 52_000).toISOString(),
    level: 'warn',
    message: 'request completed',
    fields: { route: '/api/v1/overview', status: 429, duration_ms: 12 },
  },
  {
    timestamp: new Date(now - 84_000).toISOString(),
    level: 'info',
    message: 'service started',
    fields: { environment: 'staging', version: '1.0.0' },
  },
];

export const mockConfig: ConfigStatus = {
  service: {
    name: 'go-service-template-pro',
    environment: 'staging',
    version: '1.0.0',
  },
  auth: {
    apiKeyEnabled: true,
    jwtEnabled: true,
  },
  telemetry: {
    otlpEndpoint: 'otel-collector:4317',
  },
  database: {
    configured: true,
  },
  rateLimit: {
    requestsPerSecond: 15,
    burst: 30,
  },
  allowedOrigins: ['https://ops.acme-client.io', 'https://staging.acme-client.io'],
};
