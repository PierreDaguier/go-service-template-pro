# Phase 0 - Official Stable Version Validation (March 11, 2026)

This project is intentionally positioned as an **industrial demonstration template**. Version choices are based on official stable channels at the date above.

## Verified versions

| Domain | Version selected | Official source |
|---|---:|---|
| Go runtime | **1.26.1** | https://go.dev/dl/ |
| Node.js runtime (LTS) | **24.14.0 LTS** | https://nodejs.org/en/about/previous-releases |
| React | **19.2** | https://react.dev/versions |
| Vite | **7.3.1** | https://vite.dev/releases |
| OpenTelemetry Collector | **0.147.0** | https://github.com/open-telemetry/opentelemetry-collector/releases |
| OpenTelemetry Go SDK | **1.42.0** | https://github.com/open-telemetry/opentelemetry-go/releases |
| Prometheus | **3.10.0** | https://prometheus.io/download/ |
| Grafana | **12.4.1** | https://github.com/grafana/grafana/releases |
| Loki | **3.6.7** | https://github.com/grafana/loki/releases |
| Tempo | **2.10.1** | https://github.com/grafana/tempo/releases |
| PostgreSQL | **18.3** | https://www.postgresql.org/ |

## Stack decision (2026 rationale)

- **Go 1.26 + clean architecture**: high-performance API core, maintainable structure for freelance handover and scaling.
- **React 19 + Vite 7 + TypeScript**: fast UX iteration and strong typed UI reliability for operations dashboards.
- **PostgreSQL 18**: trusted relational persistence with mature operational tooling.
- **OpenTelemetry + Prometheus + Grafana + Tempo + Loki**: vendor-neutral observability baseline expected in modern production-grade delivery.
