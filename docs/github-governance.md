# GitHub Governance Protocol

- Protect `main`: PR mandatory + required checks.
- Create `develop` branch.
- Templates: PR, issue feature/bug, release template.
- Labels:
  - `type:feature`, `type:bug`, `type:docs`, `type:ci`, `type:refactor`
  - `area:api`, `area:ops-panel`, `area:telemetry`, `area:infra`
  - `risk:low`, `risk:medium`, `risk:high`
- Milestones: MVP, Reliability, UX polish.
- Project board: Backlog / In progress / Review / Done.

Use `scripts/github-bootstrap.sh` to apply most of this automatically.
