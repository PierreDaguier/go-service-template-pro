export type Severity = 'critical' | 'high' | 'medium' | 'low';
export type IncidentStatus = 'open' | 'investigating' | 'resolved';

export interface Overview {
  serviceName: string;
  environment: string;
  version: string;
  uptimeSeconds: number;
  openIncidents: number;
  criticalIncidents: number;
  systemStatus: 'healthy' | 'degraded';
  traffic: MetricsSnapshot;
}

export interface MetricsSnapshot {
  windowSeconds: number;
  requests: number;
  rps: number;
  errorRate: number;
  p95LatencyMs: number;
  series: TimePoint[];
  statusCodes: StatusCount[];
  topRoutes: RouteCount[];
}

export interface TimePoint {
  timestamp: string;
  throughput: number;
  errorRate: number;
  p95LatencyMs: number;
}

export interface StatusCount {
  code: number;
  count: number;
}

export interface RouteCount {
  route: string;
  count: number;
}

export interface Incident {
  id: string;
  service: string;
  environment: string;
  severity: Severity;
  status: IncidentStatus;
  message: string;
  traceId?: string;
  metadata?: Record<string, unknown>;
  createdAt: string;
  resolvedAt?: string;
}

export interface IncidentPagination {
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

export interface TraceRecord {
  timestamp: string;
  traceId: string;
  method: string;
  route: string;
  statusCode: number;
  durationMs: number;
  source: string;
}

export interface LogEntry {
  timestamp: string;
  level: string;
  message: string;
  fields: Record<string, unknown>;
}

export interface ConfigStatus {
  service: {
    name: string;
    environment: string;
    version: string;
  };
  auth: {
    apiKeyEnabled: boolean;
    jwtEnabled: boolean;
  };
  telemetry: {
    otlpEndpoint: string;
  };
  database: {
    configured: boolean;
  };
  rateLimit: {
    requestsPerSecond: number;
    burst: number;
  };
  allowedOrigins: string[];
}

export interface CreateIncidentPayload {
  service?: string;
  environment?: string;
  severity: Severity;
  status: IncidentStatus;
  message: string;
  traceId?: string;
  metadata?: Record<string, unknown>;
}

export interface ApiErrorPayload {
  code: string;
  message: string;
  requestId?: string;
  details?: unknown;
}
