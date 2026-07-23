# RHTAS Consistency Report — Upstream Fork Repositories

Generated: 2026-07-23

## Repositories

| Repository | Branch | Upstream |
|------------|--------|----------|
| certificate-transparency-go | main | google/certificate-transparency-go |
| cosign | main | sigstore/cosign |
| fulcio | main | sigstore/fulcio |
| gitsign | main | sigstore/gitsign |
| model-transparency | main | sigstore/model-transparency |
| model-transparency-go | main | sigstore/model-transparency |
| policy-controller | main | sigstore/policy-controller |
| rekor | main | sigstore/rekor |
| rekor-monitor | main | sigstore/rekor-monitor |
| timestamp-authority | main | sigstore/timestamp-authority |
| tough | develop ⚠️ | awslabs/tough |
| trillian | main | google/trillian |

## Standardization scope

These repos track upstream open-source projects and carry Red Hat-specific patches. Standardizing upstream workflow files creates merge conflicts on every rebase — upstream will not adopt the team's conventions. The practical scope for changes is limited to:

- **`.rh` Dockerfiles** — fully owned, standardize freely.
- **RH-added workflow jobs** — any CI job added by the team (not from upstream), apply standards there.
- **Upstream workflow files** — accept divergence; only fix issues that directly cause RH build failures.

---

## 1. Go Versions

Most forks are on 1.26.x, which is expected as they largely track upstream. The Go version in these repos is owned by upstream and will update on their schedule.

| Version | Repos |
|---------|-------|
| 1.26.3 | cosign, fulcio, gitsign, model-transparency-go, rekor |
| 1.26.0 | certificate-transparency-go, policy-controller, rekor-monitor, trillian |
| 1.26 *(no patch)* | timestamp-authority |
| 1.25.0 | `hack/tools` in fulcio, rekor, timestamp-authority |
| N/A (Rust) | tough |
| N/A (Python) | model-transparency |

`hack/tools` sub-modules in fulcio, rekor, and timestamp-authority are one minor version behind the main module — these are tooling-only and lower priority but worth aligning.

---

## 2. Go Version Source in Workflows (lint failure root cause)

Three approaches are used to specify the Go version for CI. One is the direct cause of intermittent lint failures and exists only in upstream forks.

### Approach A — Parsed from `Dockerfile` (fragile) ⚠️

Used in **~25 workflow files** across cosign, fulcio, rekor, timestamp-authority.

```yaml
- name: Extract version of Go to use
  run: echo "GOVERSION=$(awk -F'[:@]' '/FROM golang/{print $2; exit}' Dockerfile)" >> $GITHUB_ENV

- uses: actions/setup-go@...
  with:
    go-version: ${{ env.GOVERSION }}
```

This fails silently in two scenarios:

1. **No `golang:` base image** — `.rh` Dockerfiles use `registry.redhat.io/ubi9/go-toolset`. When a workflow runs in that context the `awk` pattern never matches, `GOVERSION` is empty, and `setup-go` errors or picks an unpredictable version.
2. **Version skew** — the Go version in the Dockerfile and `go.mod` are already out of sync (e.g., rekor's `Dockerfile` pins `golang:1.26.4` while `go.mod` declares `go 1.26.3`).

**Fix:** replace this block with `go-version-file: 'go.mod'` in all affected workflow files in cosign, fulcio, rekor, and timestamp-authority. This is an RH-side fix and will not conflict with upstream since it improves reliability without changing behaviour.

### Approach B — Read from `go.mod` (reliable, canonical)

Used in: certificate-transparency-go, gitsign, model-transparency-go, policy-controller, rekor-monitor, trillian.

```yaml
- uses: actions/setup-go@...
  with:
    go-version-file: 'go.mod'
    check-latest: true
```

Single source of truth. No skew possible.

### Approach C — Hardcoded version (stale)

One file: `rekor-monitor/.github/workflows/test-rekor-integration.yml`

```yaml
go-version: 1.25
```

Already one minor version behind the repo's own `go.mod` (`go 1.26.0`). Should be replaced with `go-version-file: 'go.mod'`.

### Summary

| Approach | Repos | Risk |
|----------|-------|------|
| Dockerfile extraction ⚠️ | cosign (~11 wf), fulcio (~5 wf), rekor (~5 wf), timestamp-authority (~4 wf) | High — silently fails with `.rh` Dockerfiles |
| `go-version-file: go.mod` | certificate-transparency-go, gitsign, model-transparency-go, policy-controller, rekor-monitor, trillian | None — recommended |
| Hardcoded ⚠️ | rekor-monitor/test-rekor-integration.yml | Medium — goes stale silently |

---

## 3. GitHub Actions Versions

Upstream forks own their workflow files. Version divergence across forks is expected and will continue as upstream projects update independently. The focus here is on identifying incomplete update sweeps and unpinned references that pose security risk.

### SHA pinning strategy

All forks already use SHA pinning in upstream-owned workflows — this is good practice inherited from upstream. The concern is that the SHA pins across forks are at many different versions because Renovate is not running effectively (broken central config). Once Renovate is fixed, pins will converge automatically.

### Incomplete update sweeps

Several actions show two distinct SHA batches, indicating a manual update was done but not completed consistently across all forks:

**chainguard-dev/actions/*** — all six sub-actions in two batches:

| SHA | Repos |
|-----|-------|
| `3e8a2a22` (older) | cosign, fulcio, policy-controller |
| `d67380d0` (newer) | policy-controller |

policy-controller has both batches in different workflow files.

**github/codeql-action/*** — six distinct SHA batches:

| SHA batch | Repos |
|-----------|-------|
| `28deaeda` (oldest) | certificate-transparency-go |
| `38697555` | policy-controller, trillian |
| `5d4e8d1a` | rekor-monitor |
| `65c74964` | cosign, fulcio |
| `89a39a4e` | model-transparency |
| `8aad20d1` (newest) | model-transparency-go, rekor, timestamp-authority |

**ko-build/setup-ko** — rekor uses two different SHA pins in different workflow files.

**mikefarah/yq** — policy-controller uses two different SHA pins in different workflow files.

### Unpinned `@main` references ⚠️

Using `@main` provides no immutability guarantee. These should be pinned to a SHA:

| Action | Repos |
|--------|-------|
| `sigstore/community` reusable dependency review | cosign, model-transparency-go |
| `sigstore-conformance/extremely-dangerous-public-oidc-beacon` | gitsign |
| `sigstore/sigstore-conformance` | cosign |

### Actions used only in forks (divergence accepted)

These actions are not used in team-owned repos. Version divergence across forks is expected and low priority — Renovate will align them once running correctly.

| Action | Versions across forks |
|--------|----------------------|
| sigstore/cosign-installer | 4 distinct SHA pins |
| ossf/scorecard-action | 2 SHA pins (certificate-transparency-go one behind the rest) |
| goreleaser/goreleaser-action | 3 distinct SHA pins |
| anchore/sbom-action/download-syft | 3 distinct SHA pins |
| ko-build/setup-ko | 2 SHA pins |

---

## 4. RH go-toolset Base Image Inconsistency

`.rh` Dockerfiles across forks use six different build numbers for `registry.redhat.io/ubi9/go-toolset:9.8-XXXXX`. This is the most actionable item for forks since the `.rh` Dockerfiles are fully owned by the team.

| Build tag | Repos |
|-----------|-------|
| `9.8-1784190466` | rekor, model-transparency-go, timestamp-authority, tufcli (Dockerfile.rh) |
| `9.8-1784090680` | trillian |
| `9.8-1783931515` | certificate-transparency-go, gitsign |
| `9.8-1783679445` | cosign, fulcio, rekor-monitor, tufcli (Dockerfile.tufcli-init) |
| `9.8-1782980183` | (model-validation-operator — owned) |
| `:latest` ⚠️ | policy-controller, tough |

policy-controller and tough use `:latest` with no SHA pin — non-reproducible builds.

`registry.access.redhat.com/ubi9/ubi-minimal` SHA digests also vary: `062c52ff`, `463cae32`, `8201445b`, and `932acc91` are all in use. All `.rh` Dockerfiles should be pinned to the same SHA.

---

## 5. Renovate Config Approach for Forks

The Renovate approach for upstream forks is fundamentally different from team-owned repos and must be treated separately.

### Why the approach differs

For **team-owned repos**, Renovate covers everything: Go toolchain, Docker base images, GitHub Actions, npm, etc.

For **upstream forks**, this approach would be destructive:
- Upstream projects maintain their own dependency update process (Renovate, Dependabot, or manual). If RH Renovate also opens PRs touching upstream-controlled files (`go.mod`, upstream workflow files), these changes create merge conflicts on every rebase from upstream and pollute the git history with churn.
- The team does not control the Go version, upstream GitHub Actions versions, or upstream Go module dependencies in these repos. Those update on upstream's schedule.

The correct scope for Renovate in a fork is: **RH-specific files only**.

### Current state

Only one fork (`trillian`) has a Renovate config, and it extends `centralRenovate.json` — the same config designed for owned repos. This means Renovate will try to update upstream-controlled files in trillian on the same basis as owned repos, creating potential rebase conflicts.

All other forks have no Renovate config at all, so their `.rh` Dockerfiles are never updated automatically.

### Recommended approach for forks

Each fork should have a scoped `renovate.json` that limits Renovate strictly to RH-specific files:

```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "enabledManagers": ["dockerfile"],
  "ignorePaths": ["Dockerfile", "Containerfile"],
  "includePaths": [
    "Dockerfile*.rh",
    "Containerfile*.rh",
    ".tekton/**"
  ],
  "packageRules": [
    {
      "description": "Group all RH Docker image updates together",
      "matchManagers": ["dockerfile"],
      "groupName": "RH Base Images",
      "groupSlug": "rh-base-images"
    }
  ],
  "ignoreDeps": [
    "registry.redhat.io/rhtas/rhtas-rhel9-operator",
    "registry.redhat.io/rhtas/trillian-logsigner-rhel9"
    // ... full ignoreDeps list from centralRenovate.json
  ]
}
```

Key properties of this config:
- `enabledManagers: ["dockerfile"]` — only scans Dockerfiles, never `go.mod` or workflow files.
- `includePaths` — limits scanning to `.rh` files and `.tekton/` (RH-added Tekton configs). Upstream `Dockerfile` and `Containerfile` are explicitly excluded.
- `ignoreDeps` — inherited from the central list so product images are never bumped.
- No `github-actions` manager — upstream workflow file Actions versions are owned by upstream.

### Current Renovate config status per fork

| Repository | Renovate config | Problem |
|------------|----------------|---------|
| trillian | `centralRenovate.json` | Wrong scope — covers upstream-controlled files |
| certificate-transparency-go | None | `.rh` Dockerfiles never auto-updated |
| cosign | None | `.rh` Dockerfiles never auto-updated |
| fulcio | None | `.rh` Dockerfiles never auto-updated |
| gitsign | None | `.rh` Dockerfiles never auto-updated |
| model-transparency | None | `.rh` Dockerfiles never auto-updated |
| model-transparency-go | None | `.rh` Dockerfiles never auto-updated |
| policy-controller | None | `.rh` Dockerfiles never auto-updated |
| rekor | None | `.rh` Dockerfiles never auto-updated |
| rekor-monitor | None | `.rh` Dockerfiles never auto-updated |
| timestamp-authority | None | `.rh` Dockerfiles never auto-updated |
| tough | None | `.rh` Dockerfiles never auto-updated |

---

## 6. Recommended Actions

Changes here are limited to what is safe to make in a fork without causing rebase conflicts.

---

### HIGH — Replace Dockerfile Go version extraction with `go-version-file: 'go.mod'`

Affects: cosign (~11 workflow files), fulcio (~5), rekor (~5), timestamp-authority (~4).

**Problematic** (`rekor/.github/workflows/verify.yml` and others):
```yaml
- name: Extract version of Go to use
  run: echo "GOVERSION=$(awk -F'[:@]' '/FROM golang/{print $2; exit}' Dockerfile)" >> $GITHUB_ENV

- uses: actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 # v6.5.0
  with:
    go-version: ${{ env.GOVERSION }}
```
Silently produces an empty `GOVERSION` when the Dockerfile has no `golang:` base image (e.g., `.rh` Dockerfiles).

**Better** (`gitsign/.github/workflows/verify.yml`):
```yaml
- uses: actions/setup-go@7a3fe6cf4cb3a834922a1244abfce67bcef6a0c5 # v6.2.0
  with:
    go-version-file: 'go.mod'
    check-latest: true
```

---

### HIGH — Fix hardcoded Go version in rekor-monitor

**Problematic** (`rekor-monitor/.github/workflows/test-rekor-integration.yml`):
```yaml
- uses: actions/setup-go@v4
  with:
    go-version: 1.25
```
One minor version behind the repo's own `go.mod` (`go 1.26.0`) and will silently stay stale.

**Better** — same fix as above:
```yaml
- uses: actions/setup-go@... # latest SHA pin
  with:
    go-version-file: 'go.mod'
```

---

### HIGH — Add scoped `renovate.json` to all forks

No fork (except trillian with the wrong config) has Renovate managing its `.rh` Dockerfiles. See Section 5 for the recommended config template.

**Problematic** — no `renovate.json` in cosign, fulcio, rekor, gitsign, etc. → `.rh` base images never auto-updated.

**Better** — scoped config limiting Renovate to `.rh` files only (full template in Section 5).

---

### HIGH — Fix trillian's Renovate config

**Problematic** (`trillian/renovate.json`):
```json
{ "extends": ["github>securesign/actions:centralRenovate.json"] }
```
Inherits the owned-repo config — Renovate will try to update upstream-controlled `go.mod` and workflow files in trillian, creating rebase conflicts.

**Better** — replace with the fork-scoped config from Section 5, limiting scope to `Dockerfile*.rh` files only.

---

### MEDIUM — Standardize RH go-toolset to a single build tag across all `.rh` Dockerfiles

**Problematic** — two different approaches in active `.rh` Dockerfiles:
```dockerfile
# policy-controller/Dockerfile.rh — mutable tag
FROM registry.redhat.io/ubi9/go-toolset:latest@sha256:ad9d042375cef55890db3378ced9d0cebd74656bc8dc4c3b0cdbea31b85ce459

# rekor/Dockerfile.rekor-server.rh — specific build tag
FROM registry.redhat.io/ubi9/go-toolset:9.8-1784190466@sha256:f99dd81b20e5971ef9f63a51ac27cf0aa591ff9921d021490548b67fd9b17144
```

**Better** — use the specific build tag everywhere:
```dockerfile
FROM registry.redhat.io/ubi9/go-toolset:9.8-1784190466@sha256:f99dd81b20e5971ef9f63a51ac27cf0aa591ff9921d021490548b67fd9b17144
```

---

### MEDIUM — Pin `ubi9/ubi-minimal` to a consistent SHA across all `.rh` Dockerfiles

**Problematic** — different SHAs across forks:
```dockerfile
# tufcli/Dockerfile.rh
FROM registry.access.redhat.com/ubi9/ubi-minimal@sha256:6c79f4fb38a20d496c859025d57e4074835e849d5d14819c4e021ad78446bce8
# rekor/Dockerfile.rekor-cli.rh
FROM registry.access.redhat.com/ubi9/ubi-minimal@sha256:062c52ff973065752b0965787649db2bcf551a6c727a00e95a3eb42cebadbdab
```

**Better** — pick one SHA and use it in all `.rh` Dockerfiles across all forks.

---

### MEDIUM — Pin `@main` refs to SHA in upstream workflow files

**Problematic** (`cosign/.github/workflows/depsreview.yml`):
```yaml
uses: sigstore/community/.github/workflows/reusable-dependency-review.yml@main
```

**Problematic** (`cosign/.github/workflows/conformance-nightly.yml`):
```yaml
- uses: sigstore/sigstore-conformance@main
```
`@main` is mutable — the referenced workflow can change at any time without a PR in this repo.

**Better** — pin to a specific commit SHA:
```yaml
uses: sigstore/community/.github/workflows/reusable-dependency-review.yml@9b1b5aca605f92ec5b1bf3681b1e61b3dbc420cc
```

---

### LOW — Align `hack/tools` Go version to match the main module

**Problematic** (`fulcio/hack/tools/go.mod`, `rekor/hack/tools/go.mod`, `timestamp-authority/hack/tools/go.mod`):
```
go 1.25.0
```

**Better** — match the main module version:
```
go 1.26.3
```
