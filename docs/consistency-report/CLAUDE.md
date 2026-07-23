# RHTAS Multi-Repo Consistency Analysis

This directory contains all RHTAS-related repositories cloned locally for cross-repo analysis. The goal is to identify and resolve consistency problems that make the product harder to maintain: Go version drift, fragmented GitHub Actions versions, broken Renovate automation, and inconsistent base images.

## Cloning the repositories

Create an empty directory and run the script below from inside it. It clones all 29 repositories from the `securesign` GitHub organization. SSH access to the org is required.

```bash
#!/usr/bin/env bash
set -euo pipefail

BASE=git@github.com:securesign

repos=(
  actions
  artifact-signer-ansible
  certificate-transparency-go
  cosign
  fbc
  fulcio
  gitsign
  model-transparency
  model-transparency-go
  model-validation-operator
  pipelines
  policy-controller
  policy-controller-operator
  rekor
  rekor-monitor
  rekor-search-ui
  releases
  renovate-config
  rhtas-benchmark
  rhtas-console
  rhtas-console-ui
  secure-sign-operator
  segment-backup-job
  sigstore-e2e
  structural-tests
  timestamp-authority
  trillian
  tufcli
)

for repo in "${repos[@]}"; do
  echo "Cloning $repo..."
  git clone "$BASE/$repo"
done

# tough uses a non-default branch
echo "Cloning tough (develop branch)..."
git clone "$BASE/tough" --branch develop
```

Save the script as `clone-all.sh`, make it executable (`chmod +x clone-all.sh`), and run it from the directory where you want all repos to live. Once complete, open that directory in Claude Code.

## Repository categories

The 29 repositories split into two categories with different maintenance implications — described in detail in the reports.

**Team-owned (17):** actions, artifact-signer-ansible, fbc, model-validation-operator, pipelines, policy-controller-operator, releases, renovate-config, rekor-search-ui, rhtas-benchmark, rhtas-console, rhtas-console-ui, secure-sign-operator, segment-backup-job, sigstore-e2e, structural-tests, tufcli

**Upstream forks (12):** certificate-transparency-go, cosign, fulcio, gitsign, model-transparency, model-transparency-go, policy-controller, rekor, rekor-monitor, timestamp-authority, tough, trillian

For forks, only RH-specific files (`.rh` Dockerfiles, RH-added workflow jobs) should be standardized. Upstream-controlled files will diverge on every rebase.

## Analysis reports

| File | Contents |
|------|----------|
| `consistency-report.md` | Index and summary of key findings |
| `consistency-report-owned.md` | Full analysis for team-owned repos — Go versions, GitHub Actions targets, Renovate config, recommendations with code examples |
| `consistency-report-forks.md` | Full analysis for upstream forks — lint failure root cause, go-toolset inconsistency, Renovate scoping, recommendations with code examples |

## What has been analyzed

- **Go versions** across all `go.mod` files (including sub-modules like `hack/tools`)
- **GitHub Actions versions** — every `uses:` across all `.github/workflows/` files; version per repo, SHA pins vs tags, recommended targets
- **Renovate configuration** — `centralRenovate.json` (in `actions` repo) and `org-inherited-config.json` (in `renovate-config` repo); which repos inherit which config; bugs found (Go/Python version caps, missing `ignoreDeps` entries, dead `github-actions` manager)
- **Go version source in workflows** — three patterns found: Dockerfile extraction (fragile, causes lint failures), `go-version-file: go.mod` (correct), hardcoded version (stale)
- **RH go-toolset base image versions** — across all `.rh` Dockerfiles
- **`ubi9/ubi-minimal` SHA pins** — consistency across `.rh` Dockerfiles
- **`ignoreDeps` completeness** — which product images are missing and need to be added

## Key findings summary

| Problem | Where | Priority |
|---------|-------|----------|
| Go version cap `<1.24.0` blocks Renovate updates | `actions/centralRenovate.json` | High |
| `cosign-rhel9`, `updatetree-rhel9` missing from `ignoreDeps` | `actions/centralRenovate.json` | High |
| Dockerfile Go version extraction causes intermittent lint failures | cosign, fulcio, rekor, timestamp-authority workflows | High |
| `actions/checkout@v2` still in use | `actions` repo | High |
| Go versions 1.22–1.24 still in use | sigstore-e2e, structural-tests, pipelines | High |
| `github-actions` manager missing from `enabledManagers` | `renovate-config/org-inherited-config.json` | Medium |
| 12 owned repos have no Renovate config at all | — | Medium |
| GitHub Actions use mix of version tags and SHA pins | all repos | Medium |
| golangci-lint spans v1.61–v2.12.2 | all Go repos | Medium |
| 6 different RH go-toolset build tags in `.rh` Dockerfiles | all forks | Medium |
| trillian fork inherits owned-repo Renovate config (wrong scope) | `trillian/renovate.json` | High |

## Continuing the analysis

Useful questions to ask Claude when continuing:

- "Are there any other configuration files that should be consistent across repos?" (e.g., `.golangci.yml` rules, `Makefile` targets, `CODEOWNERS`)
- "Which repos are missing a particular file?" (e.g., `SECURITY.md`, `.golangci.yml`)
- "Show me all the places where X version is referenced across repos"
- "Compare the `.golangci.yml` configs across repos — are the linter rules consistent?"
- "Which repos don't have CodeQL scanning set up?"
- "Are there any GitHub Actions used only in one repo that could be shared?"

To update the reports after making changes to repositories, ask Claude to re-run the relevant analysis and update the appropriate report file.
