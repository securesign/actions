#!/usr/bin/env bash
# Create a GitHub-signed commit through the REST API (Git Data API).
#
# GitHub cryptographically signs commits that bots / GitHub Apps author via
# the API (they show "Verified" and satisfy require-signed-commits rulesets),
# unlike plain `git commit` + `git push`. This script therefore sends NO
# author/committer, so the commit is attributed to the token's identity and
# is signed by GitHub.
#
# Environment:
#   GH_TOKEN        required   token (GitHub App installation token recommended)
#   BRANCH          required   branch to update, or to create when CREATE_BRANCH=true
#   MESSAGE         required   commit message
#   BASE_REF        optional   diff base. Default: origin/<BRANCH>, or
#                              origin/<default branch> when CREATE_BRANCH=true
#   SOURCE_REV      optional   commit-ish to read file contents from.
#                              Default: the git index (stage changes with git add)
#   CREATE_BRANCH   optional   "true": create <BRANCH> at the new commit
#   EXTRA_PARENT    optional   additional parent rev (creates a merge commit;
#                              recorded even when the diff is empty)
#   PARENT_SHA      optional   override the first parent SHA. The new commit is
#                              stacked on the parent's tree, so parent changes
#                              are never reverted (use when the parent is an API
#                              commit that does not exist locally)
#
# Outputs ($GITHUB_OUTPUT): commit-sha, committed
#
# Content is uploaded byte-exact (returned blob SHAs are verified against the
# local ones). File names must not contain tabs, newlines, quotes, or
# backslashes.
set -euo pipefail

: "${GH_TOKEN:?GH_TOKEN is required}"
: "${BRANCH:?BRANCH is required}"
: "${MESSAGE:?MESSAGE is required}"

# gh api --input sends no Content-Type header (only -f/-F bodies get one
# automatically); GitHub's API requires application/json for JSON bodies.
JSON_HEADER="Content-Type: application/json"

REPO_DATA=$(gh repo view --json nameWithOwner,defaultBranchRef)
REPO=$(jq -r .nameWithOwner <<<"$REPO_DATA")
DEFAULT_BRANCH=$(jq -r .defaultBranchRef.name <<<"$REPO_DATA")

if [ -z "${BASE_REF:-}" ]; then
  if [ "${CREATE_BRANCH:-false}" = "true" ]; then
    BASE_REF="origin/${DEFAULT_BRANCH}"
  else
    BASE_REF="origin/${BRANCH}"
  fi
fi

if ! BASE_SHA=$(git rev-parse --verify --quiet "${BASE_REF}^{commit}"); then
  echo "::error::Base ref '${BASE_REF}' not found locally. Fetch it first (actions/checkout with fetch-depth: 0, or git fetch)."
  exit 1
fi
BASE_TREE=$(git rev-parse "${BASE_REF}^{tree}")

# First parent: explicit override (API commit not present locally), else the
# base commit. When overriding, build on the parent's tree so the new commit
# stacks on it instead of reverting parent changes missing from the index.
PARENT1="${PARENT_SHA:-$BASE_SHA}"
if [ -n "${PARENT_SHA:-}" ] && [ "$PARENT_SHA" != "$BASE_SHA" ]; then
  BASE_TREE=$(gh api "repos/${REPO}/git/commits/${PARENT_SHA}" --jq .tree.sha)
fi
if [ -n "${EXTRA_PARENT:-}" ]; then
  PARENT2=$(git rev-parse "${EXTRA_PARENT}^{commit}")
  PARENTS=$(jq -n --arg p1 "$PARENT1" --arg p2 "$PARENT2" '[$p1, $p2]')
else
  PARENTS=$(jq -n --arg p1 "$PARENT1" '[$p1]')
fi

# Changed files: SOURCE_REV vs base, or the index (staged changes) vs base.
# core.quotePath=false keeps non-ASCII paths unquoted and parseable.
if [ -n "${SOURCE_REV:-}" ]; then
  DIFF=$(git -c core.quotePath=false diff --name-status --no-renames "$BASE_SHA" "$SOURCE_REV")
else
  DIFF=$(git -c core.quotePath=false diff --name-status --no-renames --cached "$BASE_SHA")
fi

# An empty diff still records a merge commit when EXTRA_PARENT is set
if [ -z "$DIFF" ] && [ -z "${EXTRA_PARENT:-}" ]; then
  echo "No changes against ${BASE_REF} and no merge parent. Nothing to commit."
  if [ -n "${GITHUB_OUTPUT:-}" ]; then
    echo "commit-sha=" >> "$GITHUB_OUTPUT"
    echo "committed=false" >> "$GITHUB_OUTPUT"
  fi
  exit 0
fi

payload=$(mktemp)
entries=$(mktemp)

while IFS=$'\t' read -r status path; do
  [ -z "$path" ] && continue
  case "$status" in
    D)
      # Mode of the deleted file comes from the base tree
      mode=$(git -c core.quotePath=false ls-tree "$BASE_SHA" -- ":(literal)$path" | awk '{print $1}')
      entry=$(jq -n --arg p "$path" --arg m "$mode" '{path: $p, mode: $m, type: "blob", sha: null}')
      ;;
    A | M)
      # ":(literal)" disables pathspec globbing: a file named "[.]env" must not
      # resolve to ".env"'s blob
      if [ -n "${SOURCE_REV:-}" ]; then
        meta=$(git -c core.quotePath=false ls-tree "$SOURCE_REV" -- ":(literal)$path")  # <mode> <type> <sha>\t<path>
        mode=$(echo "$meta" | awk '{print $1}')
        blob=$(echo "$meta" | awk '{print $3}')
      else
        meta=$(git -c core.quotePath=false ls-files -s -- ":(literal)$path")            # <mode> <sha> <stage>\t<path>
        mode=$(echo "$meta" | awk '{print $1}')
        blob=$(echo "$meta" | awk '{print $2}')
      fi
      # Upload the blob. Payload goes through a file + gh api --input: content
      # can exceed the shell per-argument limit (catalog files reach ~1 MB).
      # base64 runs outside jq: jq's raw text input replaces non-UTF-8 bytes.
      git cat-file blob "$blob" | base64 | tr -d '\n' | jq -Rs '{content: ., encoding: "base64"}' > "$payload"
      blob_sha=$(gh api --method POST "repos/${REPO}/git/blobs" -H "$JSON_HEADER" --input "$payload" --jq .sha)
      # Blobs are content-addressed: a mismatch means the upload altered bytes
      if [ "$blob_sha" != "$blob" ]; then
        echo "::error::Blob upload for '${path}' altered content (expected ${blob}, got ${blob_sha})."
        exit 1
      fi
      entry=$(jq -n --arg p "$path" --arg m "$mode" --arg s "$blob_sha" '{path: $p, mode: $m, type: "blob", sha: $s}')
      ;;
    *)
      echo "::error::Unsupported diff status '${status}' for '${path}'"
      exit 1
      ;;
  esac
  printf '%s\n' "$entry" >> "$entries"
done <<< "$DIFF"

if [ -s "$entries" ]; then
  jq -s --arg bt "$BASE_TREE" '{base_tree: $bt, tree: .}' "$entries" > "$payload"
  tree_sha=$(gh api --method POST "repos/${REPO}/git/trees" -H "$JSON_HEADER" --input "$payload" --jq .sha)
else
  # Content-free merge: the base tree is already the intended result
  tree_sha=$BASE_TREE
fi

# No author/committer on purpose: the commit is attributed to the token's
# identity and signed by GitHub (required for the Verified badge).
jq -n --arg m "$MESSAGE" --arg t "$tree_sha" --argjson p "$PARENTS" \
  '{message: $m, tree: $t, parents: $p}' > "$payload"
commit=$(gh api --method POST "repos/${REPO}/git/commits" -H "$JSON_HEADER" --input "$payload")
new_sha=$(jq -r .sha <<<"$commit")

if [ "$(jq -r .verification.verified <<<"$commit")" != "true" ]; then
  echo "::error::Commit ${new_sha} was created but is NOT verified. GitHub signs API commits only when no author/committer is set and the token belongs to a bot/App."
  exit 1
fi

if [ "${CREATE_BRANCH:-false}" = "true" ]; then
  if ! gh api --method POST "repos/${REPO}/git/refs" -f "ref=refs/heads/${BRANCH}" -f "sha=${new_sha}" > /dev/null; then
    echo "::error::Failed to create branch '${BRANCH}'. If it already exists, drop create-branch (the branch is updated instead) or delete it first."
    exit 1
  fi
else
  # Percent-encode the whole branch name for the ref path (slashes, '#', ...)
  branch_enc=$(jq -Rr @uri <<<"$BRANCH")
  if ! gh api --method PATCH "repos/${REPO}/git/refs/heads/${branch_enc}" \
      -f "sha=${new_sha}" > /dev/null; then
    echo "::error::Failed to update '${BRANCH}'. If this was a non-fast-forward rejection, the branch moved since checkout — re-run the workflow with a fresh checkout."
    exit 1
  fi
fi

echo "Created verified commit ${new_sha} on ${BRANCH} (${REPO})"
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  echo "commit-sha=${new_sha}" >> "$GITHUB_OUTPUT"
  echo "committed=true" >> "$GITHUB_OUTPUT"
fi
