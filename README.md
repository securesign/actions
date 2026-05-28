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
Automates syncing downstream forks with upstream repository releases. Detects the latest upstream release tag (or accepts an explicit ref), pushes it as a branch, and creates a PR for review. Merge conflicts are visible in the GitHub PR UI and must be resolved before merging.

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
  generate-token:
    runs-on: ubuntu-latest
    outputs:
      token: ${{ steps.app-token.outputs.token }}
    steps:
      - uses: actions/create-github-app-token@v1
        id: app-token
        with:
          app-id: ${{ vars.SYNC_APP_ID }}
          private-key: ${{ secrets.SYNC_APP_PRIVATE_KEY }}

  sync:
    needs: generate-token
    uses: securesign/actions/.github/workflows/sync-upstream.yaml@main
    with:
      upstream_repo: https://github.com/sigstore/rekor
      upstream_ref: ${{ inputs.upstream_ref || '' }}
      target_branch: main
    secrets:
      token: ${{ needs.generate-token.outputs.token }}
```

#### Inputs
| Input | Type | Required | Description |
|-------|------|----------|-------------|
| `upstream_repo` | string | yes | Upstream repository URL (e.g., `https://github.com/sigstore/rekor`) |
| `upstream_ref` | string | no | Specific upstream ref (tag/branch) to merge. Auto-detects latest release if empty. |
| `target_branch` | string | yes | Downstream branch to merge into (e.g., `main`) |

#### Secrets
| Secret | Required | Description |
|--------|----------|-------------|
| `token` | yes | GitHub token with contents write + pull-requests write. Use a GitHub App token (via `actions/create-github-app-token`) so PRs automatically trigger CI workflows. |

#### GitHub App Setup
1. Create a GitHub App in the securesign org with permissions: **Contents: Read & Write**, **Pull Requests: Read & Write**
2. Install the app on the repositories that need upstream sync
3. Store the App ID as an org-level variable: `SYNC_APP_ID`
4. Store the private key as an org-level secret: `SYNC_APP_PRIVATE_KEY`

#### Features
- **Auto-detection**: Queries the upstream repo's latest GitHub Release when no `upstream_ref` is provided
- **Idempotent**: Updates existing sync PRs instead of creating duplicates
- **Conflict-safe**: Merge conflicts are shown in the GitHub PR UI for manual resolution
- **Skip logic**: Skips if the target branch already contains the upstream ref

#### Repository Settings
1. Actions need to be able to create pull requests (Settings -> Actions -> General -> Workflow permissions)

# Contributing
If you want to add a reusable GitHub action, please refer to the documentation [here](https://docs.github.com/en/actions/using-workflows/reusing-workflows).
