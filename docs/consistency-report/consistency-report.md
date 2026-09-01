# RHTAS Multi-Repo Consistency Report

Generated: 2026-07-23

29 repositories analyzed across two categories: 17 team-owned and 12 upstream forks. Because forks diverge on every upstream rebase, the findings and recommendations are split into two separate reports.

## Reports

- [consistency-report-owned.md](consistency-report-owned.md) — Team-owned repositories. Full standardization is achievable. Covers Go versions, GitHub Actions targets, Renovate config issues, and go-toolset base images.

- [consistency-report-forks.md](consistency-report-forks.md) — Upstream fork repositories. Scope is limited to `.rh` Dockerfiles and RH-added workflow jobs. Covers the lint failure root cause (Dockerfile Go version extraction), incomplete SHA pin update sweeps, and go-toolset inconsistencies.

## Key findings at a glance

| Finding | Owned repos | Fork repos |
|---------|------------|------------|
| Go versions in use | 5 different (1.22–1.26.3) | 3 different (1.25–1.26.3) |
| Actions checkout versions | v2–v7 | SHA pins at 6 different commits |
| Renovate config | Broken — Go capped at <1.24, 2 images missing from ignoreDeps | Inherits from owned config |
| go-toolset SHA | 3 different tags + 1 `:latest` | 5 different tags + 2 `:latest` |
| Lint failure cause | — | Dockerfile Go version extraction in cosign, fulcio, rekor, timestamp-authority |
