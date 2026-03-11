import {
  mockConfig,
  mockIncidents,
  mockIncidentsPagination,
  mockLogs,
  mockMetrics,
  mockOverview,
  mockTraces,
} from './mock';
import type {
  ConfigStatus,
  CreateIncidentPayload,
  Incident,
  IncidentPagination,
  LogEntry,
  MetricsSnapshot,
  Overview,
  TraceRecord,
} from '../types/ops';

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? '';
const API_KEY = import.meta.env.VITE_API_KEY ?? 'client-demo-key-2026';
const USE_MOCKS = import.meta.env.VITE_USE_MOCKS === 'true';

interface Envelope<T> {
  data: T;
}

interface PagedEnvelope<T> {
  data: T;
  pagination: IncidentPagination;
}

export class ApiError extends Error {
  status: number;
  code?: string;

  constructor(message: string, status: number, code?: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': API_KEY,
      ...(init?.headers ?? {}),
    },
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => ({ error: { message: 'Unknown API error' } }));
    throw new ApiError(payload.error?.message ?? 'API request failed', response.status, payload.error?.code);
  }

  return response.json() as Promise<T>;
}

async function withFallback<T>(promise: Promise<T>, fallback: () => T | Promise<T>): Promise<T> {
  if (!USE_MOCKS) return promise;
  try {
    return await promise;
  } catch {
    return fallback();
  }
}

export async function getOverview(): Promise<Overview> {
  return withFallback(
    request<Envelope<Overview>>('/api/v1/overview').then((payload) => payload.data),
    () => mockOverview,
  );
}

export async function getLiveMetrics(window = '15m'): Promise<MetricsSnapshot> {
  return withFallback(
    request<Envelope<MetricsSnapshot>>(`/api/v1/metrics/live?window=${window}`).then((payload) => payload.data),
    () => mockMetrics,
  );
}

export async function getErrors(params: {
  severity?: string;
  status?: string;
  q?: string;
  page?: number;
  pageSize?: number;
}): Promise<{ rows: Incident[]; pagination: IncidentPagination }> {
  const query = new URLSearchParams();
  if (params.severity) query.set('severity', params.severity);
  if (params.status) query.set('status', params.status);
  if (params.q) query.set('q', params.q);
  query.set('page', String(params.page ?? 1));
  query.set('page_size', String(params.pageSize ?? 20));

  return withFallback(
    request<PagedEnvelope<Incident[]>>(`/api/v1/errors?${query.toString()}`).then((payload) => ({
      rows: payload.data,
      pagination: payload.pagination,
    })),
    () => {
      const severity = params.severity?.toLowerCase() ?? '';
      const status = params.status?.toLowerCase() ?? '';
      const term = params.q?.toLowerCase() ?? '';
      const rows = mockIncidents.filter((incident) => {
        const matchSeverity = !severity || incident.severity === severity;
        const matchStatus = !status || incident.status === status;
        const matchTerm =
          term === '' ||
          incident.message.toLowerCase().includes(term) ||
          incident.service.toLowerCase().includes(term);
        return matchSeverity && matchStatus && matchTerm;
      });
      return {
        rows,
        pagination: { ...mockIncidentsPagination, total: rows.length, totalPages: Math.max(1, Math.ceil(rows.length / 12)) },
      };
    },
  );
}

export async function createError(payload: CreateIncidentPayload): Promise<Incident> {
  return withFallback(
    request<Envelope<Incident>>('/api/v1/errors', {
      method: 'POST',
      body: JSON.stringify(payload),
    }).then((result) => result.data),
    () => ({
      id: crypto.randomUUID(),
      service: payload.service ?? 'go-service-template-pro',
      environment: payload.environment ?? 'staging',
      severity: payload.severity,
      status: payload.status,
      message: payload.message,
      traceId: payload.traceId,
      metadata: payload.metadata,
      createdAt: new Date().toISOString(),
    }),
  );
}

export async function getTraces(limit = 120): Promise<TraceRecord[]> {
  return withFallback(
    request<Envelope<TraceRecord[]>>(`/api/v1/traces?limit=${limit}`).then((payload) => payload.data),
    () => mockTraces.slice(0, limit),
  );
}

export async function getConfigStatus(): Promise<ConfigStatus> {
  return withFallback(
    request<Envelope<ConfigStatus>>('/api/v1/config').then((payload) => payload.data),
    () => mockConfig,
  );
}

export async function getLogs(level = '', limit = 80): Promise<LogEntry[]> {
  const query = new URLSearchParams();
  if (level) query.set('level', level);
  query.set('limit', String(limit));

  return withFallback(
    request<Envelope<LogEntry[]>>(`/api/v1/logs?${query.toString()}`).then((payload) => payload.data),
    () => mockLogs.slice(0, limit),
  );
}
