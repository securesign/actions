# RHTAS Consistency Report — Team-Owned Repositories

Generated: 2026-07-23

## Repositories

| Repository | Branch |
|------------|--------|
| actions | main |
| artifact-signer-ansible | main |
| fbc | main |
| model-validation-operator | main |
| pipelines | main |
| policy-controller-operator | main |
| releases | main |
| renovate-config | main |
| rekor-search-ui | main |
| rhtas-benchmark | main |
| rhtas-console | main |
| rhtas-console-ui | main |
| secure-sign-operator | main |
| segment-backup-job | main |
| sigstore-e2e | main |
| structural-tests | main |
| tufcli | main |

These repos are fully under the team's control. Standards can be enforced across the board.

---

## 1. Go Versions

Target is **1.26.3**. Five different versions are in use across team-owned repos.

| Version | Repos |
|---------|-------|
| 1.26.3 | secure-sign-operator, tufcli |
| 1.26.0 | policy-controller-operator, rhtas-console |
| 1.25.7 | model-validation-operator (main), secure-sign-operator/pkg/analyzer |
| 1.25.0 | rhtas-benchmark |
| 1.24 | sigstore-e2e ⚠️ |
| 1.23.0 | structural-tests ⚠️ |
| 1.22.6 | pipelines/pipeline-tests ⚠️ |

sigstore-e2e, structural-tests, and pipelines/pipeline-tests are 2–4 minor versions behind and should be updated immediately.

---

## 2. GitHub Actions Versions

### Pinning strategy

**Use SHA pins with a version comment everywhere:**

```yaml
uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v7.0.0
```

Version tags (`v4`, `v6`) are mutable — an attacker who compromises an action repo can silently redirect them to malicious code. SHA pins are immutable. For a supply chain security product, SHA pins are the only correct choice.

Prerequisite: **fix Renovate first** so pins are kept current automatically. Without that, SHA pins go stale.

### actions/checkout

| Version | Repos |
|---------|-------|
| v2 ⚠️ | actions |
| v3 | rekor-search-ui, releases |
| v4 | artifact-signer-ansible, fbc, rekor-search-ui, releases, rhtas-console, secure-sign-operator, sigstore-e2e, structural-tests |
| v6 | policy-controller-operator, rhtas-console |
| v6.0.2 | model-validation-operator |
| v7 | pipelines, rhtas-console-ui |
| SHA pin | model-validation-operator, rhtas-console-ui |

**Target: SHA pin for v7**

### actions/setup-go

| Version | Repos |
|---------|-------|
| v5 | releases, rhtas-console, secure-sign-operator, sigstore-e2e, structural-tests |
| v5.0.2 | releases |
| v6 | policy-controller-operator, rhtas-console |
| v6.3.0 | model-validation-operator |
| SHA pin | model-validation-operator |

**Target: SHA pin for v6**

### actions/upload-artifact

| Version | Repos |
|---------|-------|
| v4 | artifact-signer-ansible, fbc, rekor-search-ui, secure-sign-operator, sigstore-e2e |
| v7 | pipelines, rhtas-console-ui |
| SHA pin | model-validation-operator |

**Target: SHA pin for v7**

### actions/download-artifact

| Version | Repos |
|---------|-------|
| v4 | fbc, secure-sign-operator |
| v8 | rhtas-console-ui |

**Target: SHA pin for v8**

### actions/cache

| Version | Repos |
|---------|-------|
| v3 | artifact-signer-ansible |
| v4 | rekor-search-ui, secure-sign-operator |
| SHA pin | rekor-monitor |

**Target: SHA pin for v4**

### actions/setup-node

| Version | Repos |
|---------|-------|
| v3 | rekor-search-ui |
| v4 | rekor-search-ui |
| v6 | rhtas-console-ui |
| SHA pin | rhtas-console-ui |

**Target: SHA pin for v6**

### actions/create-github-app-token

| Version | Repos |
|---------|-------|
| v1 ⚠️ | fbc, releases |
| v3 | model-validation-operator |

**Target: SHA pin for v3** — fbc and releases are two versions behind.

### actions/github-script

| Version | Repos |
|---------|-------|
| v7 | releases |
| v9 | pipelines |

**Target: SHA pin for v9**

### actions/attest-build-provenance

rhtas-console and rhtas-console-ui use different SHA pins — must be consolidated to the same one.

### actions/configure-pages / deploy-pages / upload-pages-artifact

Used in rekor-search-ui and rhtas-console-ui with different versions. Align to the same SHA pin.

### golangci/golangci-lint-action

| Action version | golangci-lint `version:` | Repos |
|----------------|--------------------------|-------|
| v6 (SHA pin) ⚠️ | v1.61 | structural-tests |
| v9 (SHA pin) | v2.4.0 | model-validation-operator |
| v9 (SHA pin) | v2.8.0 | sigstore-e2e |
| v9 (SHA pin) | v2.12.2 | secure-sign-operator |
| v7 tag ⚠️ | latest | rhtas-console |
| v9 tag | — | sigstore-e2e |

**Target: SHA pin for v9, golangci-lint `version: v2.12.x`**

### codecov/codecov-action

| Version | Repos |
|---------|-------|
| v4.0.1 ⚠️ | rekor-search-ui |
| v6 | policy-controller-operator, rhtas-console, secure-sign-operator |
| v7 | rhtas-console-ui |
| SHA pin (older) | model-validation-operator |

**Target: SHA pin for v7** — rekor-search-ui is three major versions behind.

### redhat-actions/podman-login

| Version | Repos |
|---------|-------|
| v1 ⚠️ | secure-sign-operator, sigstore-e2e |
| SHA pin | rhtas-console, rhtas-console-ui, secure-sign-operator |

secure-sign-operator mixes v1 tag and SHA pin across workflows. **Target: one consistent SHA pin.**

### docker/login-action

| Version | Repos |
|---------|-------|
| v3 | fbc |
| v4 | rhtas-console-ui |

**Target: SHA pin for v4**

### github/codeql-action/*

Only used in rhtas-console-ui, currently with a v4 tag. **Target: latest SHA pin.**

### chainguard-dev/actions/*

rekor-search-ui references these with `@main` — no pin at all. **Target: latest SHA pin.**

### Summary

All targets are SHA pins with a `# vX.Y.Z` comment.

| Action | Current versions | Target |
|--------|-----------------|--------|
| actions/checkout | v2, v3, v4, v6, v6.0.2, v7 | **v7** |
| actions/setup-go | v5, v5.0.2, v6, v6.3.0 | **v6** |
| actions/upload-artifact | v4, v7 | **v7** |
| actions/download-artifact | v4, v8 | **v8** |
| actions/cache | v3, v4 | **v4** |
| actions/setup-node | v3, v4, v6 | **v6** |
| actions/create-github-app-token | v1, v3 | **v3** |
| actions/github-script | v7, v9 | **v9** |
| golangci/golangci-lint-action | v6, v7, v9 + mixed lint versions | **v9, lint v2.12.x** |
| codecov/codecov-action | v4.0.1, v6, v7 | **v7** |
| redhat-actions/podman-login | v1, mixed SHA pins | **latest** |
| docker/login-action | v3, v4 | **v4** |
| chainguard-dev/actions/* | `@main` unpinned ⚠️ | **latest** |
| github/codeql-action/* | v4 tag | **latest** |

---

## 3. Central Renovate Config

**File:** `actions/centralRenovate.json`

### 3.1 Broken version caps

Two `allowedVersions` rules block automated updates beyond already-outdated versions:

```json
{
  "matchDatasources": ["golang-version"],
  "allowedVersions": "<1.24.0"
},
{
  "matchPackageNames": ["golang"],
  "allowedVersions": "<1.24.0"
},
{
  "matchPackageNames": ["python"],
  "allowedVersions": "<3.12.0"
}
```

- **Go is capped at <1.24.0** — yet most repos already run 1.26.x. Renovate will not open PRs to update the Go toolchain.
- **Python is capped at <3.12.0** — yet `model-transparency` already uses `python:3.13-slim`.

These caps must be removed or raised.

### 3.2 Fragmented config inheritance

There are two Renovate base configs maintained by the team — `centralRenovate.json` (in the `actions` repo) and `org-inherited-config.json` (in the `renovate-config` repo). They have different capabilities and different bugs. Neither is used consistently.

#### Two base configs compared

| Property | `centralRenovate.json` | `org-inherited-config.json` |
|----------|----------------------|---------------------------|
| Location | `actions` repo | `renovate-config` repo |
| Extends | `konflux-ci/mintmaker` | — |
| Go version cap | `<1.24.0` ⚠️ | None — no cap |
| Python version cap | `<3.12.0` ⚠️ | `<3.12.0` ⚠️ |
| GitHub Actions manager | Not configured | In `packageRules` but **missing from `enabledManagers`** — dead rule, Actions never updated ⚠️ |
| Dockerfile manager | Yes (grouped) | Yes (grouped) |
| Go modules manager | Yes (grouped) | Yes (grouped) with `gomodTidy` |
| Tekton manager | No | Yes |
| RPM lockfile manager | No | Yes |
| Ansible tools grouping | No | Yes |
| `ignoreDeps` (product images) | 15 images (incomplete) | 15 images (incomplete — same list) |

Notable issues in `org-inherited-config.json`:
- **GitHub Actions group rule is dead**: `github-actions` appears in `packageRules` with a group name, but `github-actions` is not listed in `enabledManagers`. Renovate will never scan workflow files, so the grouping rule does nothing.
- **Python cap is the same bug** as `centralRenovate.json`.
- **`ignoreDeps` is identical and incomplete** — missing `cosign-rhel9` and `updatetree-rhel9` (see Section 3.3).

#### Which repos use which config

Of 17 team-owned repos, only 5 have any Renovate config, and none inherit from `centralRenovate.json`:

| Repository | Inherits from | Notes |
|------------|---------------|-------|
| artifact-signer-ansible | `org-inherited-config.json` | Actions never updated (dead rule) |
| secure-sign-operator | `org-inherited-config.json` | Actions never updated (dead rule) |
| pipelines | `konflux-ci/mintmaker//config/renovate/renovate.json` | Tekton manager only; no Go or Docker coverage |
| policy-controller-operator | None — fully standalone | Has its own Go cap at `<=1.24.0` (same bug, independently maintained) |
| rhtas-console-ui | None — fully standalone | npm/node constraints only; no Dockerfile or Go management |
| All other 12 owned repos | No Renovate config | Renovate does not run for these repos at all |

Key problems:

- **12 owned repos have no Renovate config** — dependency updates are entirely manual.
- **No owned repo inherits `centralRenovate.json`** — fixes to it have zero effect on owned repos (only the fork `trillian` uses it; see the forks report).
- **`org-inherited-config.json` never updates GitHub Actions** — the manager is missing from `enabledManagers` despite the grouping rule being defined.
- **policy-controller-operator** independently replicates the Go version cap bug. Must be fixed separately.
- **pipelines** uses the Konflux mintmaker config — appropriate for Tekton but does not cover Go or Docker.
- **rhtas-console-ui** only manages npm.

The two base configs should be consolidated into one. All team-owned repos should then extend it, with repo-specific rules layered on top.

### 3.3 ignoreDeps — purpose and gaps

`ignoreDeps` lists the RHTAS product's own output images so Renovate never tries to open a PR bumping them. These images are managed by the release/promotion process, not by dependency PRs.

The current list of 15 images is incomplete. Two additional images have pinned version tags in active configs that Renovate will pick up:

| Image | Referenced in | Version seen |
|-------|--------------|--------------|
| `registry.redhat.io/rhtas/cosign-rhel9` | `pipelines/pipelines/integration-test/*.yaml` (4 files) | `:1.3.1` |
| `registry.redhat.io/rhtas/updatetree-rhel9` | `artifact-signer-ansible/molecule/key_rotation/` | `:1.1.0` |

Both must be added to `ignoreDeps`.

Note: `client-server-rhel9` maps to `cosign/Dockerfile.clients.rh` (the combined CLI stack: cosign + gitsign + rekor-cli). The name is not obvious.

---

## 4. RH go-toolset Base Images

Team-owned repos with `.rh` Dockerfiles use different go-toolset build tags:

| Build tag | Repos |
|-----------|-------|
| `9.8-1784190466` | secure-sign-operator |
| `9.8-1784076237` | rhtas-console |
| `9.8-1782980183` | model-validation-operator |
| `:latest` ⚠️ | policy-controller-operator |

policy-controller-operator uses `:latest` with no SHA pin — non-reproducible builds.

`registry.access.redhat.com/ubi9/ubi-minimal` SHA digests also vary across repos. All owned repos should use the same SHA.

---

## 5. Recommended Actions

---

### HIGH — Fix `centralRenovate.json`: remove the Go and Python version caps

**Problematic** (`actions/centralRenovate.json`):
```json
{ "matchDatasources": ["golang-version"], "allowedVersions": "<1.24.0" },
{ "matchPackageNames": ["golang"],        "allowedVersions": "<1.24.0" },
{ "matchPackageNames": ["python"],        "allowedVersions": "<3.12.0" }
```

**Better** — remove all three rules entirely. Let Renovate propose the latest stable version; pin specific versions in `allowedVersions` only if there is a documented reason to hold back.

---

### HIGH — Add missing images to `ignoreDeps` in `centralRenovate.json`

`cosign-rhel9` and `updatetree-rhel9` are referenced with pinned versions in active configs but absent from `ignoreDeps`. Renovate will open unwanted PRs bumping them.

**Problematic** — `pipelines/pipelines/integration-test/rhtas-operator-e2e.yaml`:
```yaml
image: registry.redhat.io/rhtas/cosign-rhel9:1.3.1
```
No entry in `ignoreDeps` → Renovate will try to update `:1.3.1`.

**Better** — add to `actions/centralRenovate.json`:
```json
"ignoreDeps": [
  ...
  "registry.redhat.io/rhtas/cosign-rhel9",
  "registry.redhat.io/rhtas/updatetree-rhel9"
]
```

---

### HIGH — Update `actions/checkout` in the `actions` repo from v2 to current SHA pin

**Problematic** (`actions/sync-upstream/resolve-conflicts/action.yml`):
```yaml
- uses: actions/checkout@v2
```

**Better** (as used in rekor and tufcli):
```yaml
- uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7.0.0
```

---

### HIGH — Lift Go version in lagging owned repos

**Problematic** (`sigstore-e2e/go.mod`):
```
go 1.24
```

**Better** (target, matching cosign/secure-sign-operator):
```
go 1.26.3
```

Same fix needed for `structural-tests` (1.23.0) and `pipelines/pipeline-tests` (1.22.6).

---

### MEDIUM — Fix dead `github-actions` rule in `org-inherited-config.json`

The grouping rule exists but the manager is never enabled, so GitHub Actions in `artifact-signer-ansible` and `secure-sign-operator` are never updated by Renovate.

**Problematic** (`renovate-config/org-inherited-config.json`):
```json
"enabledManagers": ["dockerfile", "npm", "pip_requirements", "gomod", "cargo", "tekton", "custom.regex", "rpm-lockfile"],
"packageRules": [
  { "matchManagers": ["github-actions"], "groupName": "GitHub Actions" }
]
```
`github-actions` is in `packageRules` but absent from `enabledManagers` — the rule is never triggered.

**Better**:
```json
"enabledManagers": ["dockerfile", "npm", "pip_requirements", "gomod", "cargo", "tekton", "custom.regex", "rpm-lockfile", "github-actions"],
```

---

### MEDIUM — Consolidate the two base configs into one

`centralRenovate.json` and `org-inherited-config.json` are maintained separately and have diverged. All owned repos should extend a single config. The consolidated config should take the best of both: no Go cap (from `org-inherited-config.json`), `gomodTidy` post-update options, all managers including `github-actions`.

---

### MEDIUM — Standardize GitHub Actions to SHA pins with version comments

**Problematic** (multiple owned repos):
```yaml
uses: actions/checkout@v4
uses: golangci/golangci-lint-action@v7
  with:
    version: latest
```

**Better** (`rekor`, `tufcli`, `secure-sign-operator`):
```yaml
uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7.0.0
uses: golangci/golangci-lint-action@1e7e51e771db61008b38414a730f564565cf7c20 # v9.2.0
  with:
    version: v2.12.2
```

---

### MEDIUM — Standardize golangci-lint version across owned repos

**Problematic** (`rhtas-console/.github/workflows/linter.yml`):
```yaml
uses: golangci/golangci-lint-action@v7
with:
  version: latest
```

**Better** (`secure-sign-operator/.github/workflows/linter.yml`):
```yaml
uses: golangci/golangci-lint-action@v9
with:
  version: v2.12.2
```

---

### LOW — Standardize RH go-toolset to a single build tag across owned `.rh` Dockerfiles

**Problematic** (`policy-controller-operator/Dockerfile`):
```dockerfile
FROM registry.redhat.io/ubi9/go-toolset:latest@sha256:f99dd81b20e5971ef9f63a51ac27cf0aa591ff9921d021490548b67fd9b17144
```
Using `:latest` makes the tag mutable — a future image push changes what this resolves to even though the SHA is pinned today.

**Better** (`rekor/Dockerfile.rekor-server.rh`):
```dockerfile
FROM registry.redhat.io/ubi9/go-toolset:9.8-1784190466@sha256:f99dd81b20e5971ef9f63a51ac27cf0aa591ff9921d021490548b67fd9b17144
```
Specific build tag makes the version explicit and auditable.

---

### LOW — Pin `ubi9/ubi-minimal` to a consistent SHA across owned `.rh` Dockerfiles

**Problematic** — two different SHAs in use across owned repos:
```dockerfile
# tufcli/Dockerfile.rh
FROM registry.access.redhat.com/ubi9/ubi-minimal@sha256:6c79f4fb38a20d496c859025d57e4074835e849d5d14819c4e021ad78446bce8
# rekor/Dockerfile.rekor-cli.rh
FROM registry.access.redhat.com/ubi9/ubi-minimal@sha256:062c52ff973065752b0965787649db2bcf551a6c727a00e95a3eb42cebadbdab
```

**Better** — pick one SHA and use it everywhere. Renovate will keep it current once configured.
