# Trusted Artifact Signer Github actions  
This repository hosts all reusable GitHub actions utilized by the 'securesign' organization on GitHub.

## Current list of actions
The current actions included in this repository include:

### Check image version
This GitHub Action utilizes skopeo to verify that the images specified in the inputs, are always using the most up-to-date SHA. If they aren't, the action will create a PR to update them in any of the branches specified in the matrix.

#### Usage

```    
check-image-version:
  uses: securesign/actions/.github/workflows/check-image-version.yaml@main
  strategy:
    matrix:
      branch: [main, midstream-vx-y-z, ....]
    with:
      branch: ${{ matrix.branch }}
      images: '["img1","img2","img3","img4"]'
    secrets:
      token: ${{ secrets.GITHUB_TOKEN }}
      registry_redhat_io_username: ${{ secrets.REGISTRY_REDHAT_IO_USERNAME }}
      registry_redhat_io_password: ${{ secrets.REGISTRY_REDHAT_IO_PASSWORD }}
```

In order for the action to work correctly there are two settings that need to be changed for the repo.

1. Actions need to be able to create pull requests (settings -> Actions -> General -> Workflow permissions)
2. Actions need read and write permissions (settings -> Actions -> General -> Workflow permissions)

### Trigger konflux build
This Github action opens a pr with a timestamp file, with the aim of easily triggering a build on Konflux. (May need to add the file to cel expressions in the tekton pipelines)

### Usage
```
name: Trigger Konflux build
on:
  workflow_dispatch:

jobs:
  trigger-konflux-build:
    uses: securesign/actions/.github/workflows/trigger-konflux-build.yaml@main
    with:
      branch: main
    secrets:
      token: ${{ secrets.GITHUB_TOKEN }}
```

In order for the action to work correctly there are two settings that need to be changed for the repo.

1. Actions need to be able to create pull requests (settings -> Actions -> General -> Workflow permissions)
2. Actions need read and write permissions (settings -> Actions -> General -> Workflow permissions)

### Sync Upstream
Automates syncing downstream forks with upstream repository releases. Detects the latest upstream release tag (or accepts an explicit ref), merges the downstream branch into it, auto-resolves known conflict patterns, and creates a PR for review.

#### Auto-resolved conflict patterns
| Pattern | Resolution |
|---------|-----------|
| Dockerfile `FROM` version conflicts | Picks the side with the newer image version |
| Dockerfile digest-only conflicts | Takes upstream when images differ only in `@sha256:` digest |
| `go.mod` dependency version conflicts | Three-way merge: picks newest semver per dependency, clamps `go` directive to ubi9/go-toolset ceiling |
| GitHub Actions version bumps (`uses: action@SHA # vX.Y.Z`) | Picks the side with the newer version |
| Downstream-only files (`.tekton/`, custom workflows, etc.) | Auto-detected and restored from the target branch |
| `go.sum` conflicts | Removed and regenerated via `go mod tidy` |
| User-configured upstream patterns | Takes upstream content for files matching `take_upstream_patterns` |
| User-configured downstream patterns | Takes downstream content for files matching `take_downstream_patterns` |

Unknown conflict types (e.g., source code changes) will cause the action to fail without pushing, requiring manual resolution.

#### Usage

```yaml
name: Sync Upstream
on:
  schedule:
    - cron: '17 8 * * 1'  # Weekly Monday 8:17 UTC
  workflow_dispatch:
    inputs:
      upstream_ref:
        description: 'Upstream ref to merge (leave empty for latest release)'
        required: false
        default: ''

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/create-github-app-token@v3
        id: app-token
        with:
          client-id: ${{ vars.SYNC_APP_CLIENT_ID }}
          private-key: ${{ secrets.SYNC_APP_PRIVATE_KEY }}

      - uses: securesign/actions/sync-upstream@main
        with:
          token: ${{ steps.app-token.outputs.token }}
          upstream_repo: https://github.com/sigstore/rekor
          upstream_ref: ${{ inputs.upstream_ref || '' }}
          target_branch: main
```

For repos with generated protobuf files or downstream-specific test fixtures, use `take_upstream_patterns` and `take_downstream_patterns` to auto-resolve additional conflict types:

```yaml
      - uses: securesign/actions/sync-upstream@main
        with:
          token: ${{ steps.app-token.outputs.token }}
          upstream_repo: https://github.com/sigstore/fulcio
          target_branch: main
          take_upstream_patterns: |
            \.pb\.go$
            _grpc\.pb\.go$
          take_downstream_patterns: |
            testdata/
```

#### Inputs
| Input | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `token` | string | yes | | GitHub token with contents write + pull-requests write. Use a GitHub App token so PRs trigger CI. |
| `upstream_repo` | string | yes | | Upstream repository URL (e.g., `https://github.com/sigstore/rekor`) |
| `upstream_ref` | string | no | (latest release) | Specific upstream ref (tag/branch) to merge. Auto-detects latest release if empty. |
| `target_branch` | string | yes | | Downstream branch to merge into (e.g., `main`) |
| `resolve_conflicts` | string | no | `true` | Auto-resolve known merge conflict patterns. Set to `false` to push the unresolved branch. |
| `go_version_ceiling_image` | string | no | `registry.redhat.io/ubi9/go-toolset:latest` | Container image used to determine the max allowed Go version via the Red Hat Pyxis API. |
| `go_version_ceiling` | string | no | | Explicit Go version ceiling (e.g., `1.26.2`). Skips Pyxis API lookup when set. |
| `downstream_only_files` | string | no | | Newline-separated list of additional files to always restore from the target branch. |
| `take_upstream_patterns` | string | no | `(^|/)CHANGELOG` | Newline-separated grep -E patterns for files to auto-resolve with upstream content. Default includes CHANGELOG files. |
| `take_downstream_patterns` | string | no | | Newline-separated grep -E patterns for files to auto-resolve with downstream content (e.g., `testdata/`). |

#### GitHub App Setup
1. Create a GitHub App in the securesign org with permissions: **Contents: Read & Write**, **Pull Requests: Read & Write**
2. Install the app on the repositories that need upstream sync
3. Store the App Client ID as an org-level variable: `SYNC_APP_CLIENT_ID`
4. Store the private key as an org-level secret: `SYNC_APP_PRIVATE_KEY`

#### Security
The caller workflow generates a short-lived GitHub App installation token (1hr expiry) and passes only this token to the composite action. The private key never leaves the caller workflow.

#### Features
- **Auto-detection**: Queries the upstream repo's latest GitHub Release when no `upstream_ref` is provided
- **Auto-conflict-resolution**: Resolves Dockerfile, go.mod, workflow, and downstream-only file conflicts automatically
- **Safe push**: Only pushes after all conflicts are resolved and committed. Failures push the sync branch for local continuation.
- **Idempotent**: Updates existing sync PRs instead of creating duplicates
- **Skip logic**: Skips if the target branch already contains the upstream ref
- **Resolution guide**: When conflicts remain, the job summary includes categorized files and step-by-step local resolution instructions

#### Manual Resolution

When the action cannot auto-resolve all conflicts, it pushes the sync branch (with the upstream commit) and writes a resolution guide to the GitHub Actions job summary. To continue locally:

```bash
# 1. Check out the sync branch and start the merge
git fetch origin
git checkout sync-upstream/main/<tag>
git merge origin/main

# 2. Auto-resolve known patterns (Dockerfiles, go.mod, workflow version bumps)
go install github.com/securesign/actions/sync-upstream/resolve-conflicts@main
resolve-conflicts all

# 3. Resolve remaining conflicts manually, then commit and push
git diff --name-only --diff-filter=U   # list what's left
# ... fix remaining files ...
git add -A && git commit
git push origin sync-upstream/main/<tag>
```

The `resolve-conflicts all` command auto-detects all conflicting files, resolves Dockerfiles, go.mod (three-way merge), and workflow version bumps, then reports what's left for manual attention.

#### Repository Settings
1. Actions need to be able to create pull requests (Settings -> Actions -> General -> Workflow permissions)

# Contributing
If you want to add a reusable GitHub action, please refer to the documentation [here](https://docs.github.com/en/actions/using-workflows/reusing-workflows).
