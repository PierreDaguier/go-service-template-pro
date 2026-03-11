# GitHub Governance

This document defines the governance baseline for `go-service-template-pro`.
The goal is to reduce delivery risk by enforcing predictable review, CI, and release discipline.

## Branch Policy

- `main` is protected.
- Direct pushes to `main` are blocked.
- All changes go through pull requests.
- Linear history is enforced on `main`.
- Admins are also subject to branch protection.

## Required Checks on `main`

The following checks must pass before merge:

- `backend-lint-test-build`
- `frontend-lint-build`
- `docker-build`

## Pull Request Rules

- At least `1` approving review is required.
- Stale approvals are dismissed when new commits are pushed.
- Conversation resolution is required before merge.

## Label Taxonomy

Use labels to improve triage and reporting consistency.

### Type

- `type/feature`
- `type/bug`
- `type/docs`
- `type/ci`
- `type/refactor`

### Priority

- `priority/p0`
- `priority/p1`
- `priority/p2`

### Area

- `area/api`
- `area/ops-panel`
- `area/telemetry`
- `area/infra`

## Milestones

Roadmap tracking milestones:

- `MVP`
- `Reliability`
- `UX Polish`
- `v1.0.0`

## Release Governance

- Release notes are drafted automatically with Release Drafter.
- Configuration lives in `.github/release.yml`.
- Workflow lives in `.github/workflows/release-drafter.yml`.

## Expected Team Workflow

1. Create an issue and apply `type/*`, `priority/*`, and `area/*` labels.
2. Assign a milestone (`MVP`, `Reliability`, `UX Polish`, or `v1.0.0`).
3. Implement on a topic branch and open a PR.
4. Ensure required checks pass, collect approval, resolve conversations, then merge.
