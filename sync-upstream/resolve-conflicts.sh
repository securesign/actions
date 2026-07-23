#!/usr/bin/env bash
set -euo pipefail

SYNC_BRANCH="${SYNC_BRANCH:?SYNC_BRANCH is required}"
TARGET_BRANCH="${TARGET_BRANCH:?TARGET_BRANCH is required}"
RESOLVE_CONFLICTS="${RESOLVE_CONFLICTS:-true}"
GO_CEILING_IMAGE="${GO_CEILING_IMAGE:-registry.redhat.io/ubi9/go-toolset:latest}"
GO_CEILING_OVERRIDE="${GO_CEILING_OVERRIDE:-}"
DOWNSTREAM_ONLY_FILES="${DOWNSTREAM_ONLY_FILES:-}"
TAKE_UPSTREAM_PATTERNS="${TAKE_UPSTREAM_PATTERNS:-}"
TAKE_DOWNSTREAM_PATTERNS="${TAKE_DOWNSTREAM_PATTERNS:-}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

RESOLVE="${SCRIPT_DIR}/resolve-conflicts/resolve-conflicts"

info() { echo "[resolve-conflicts] $*"; }
err()  { echo "::error::$*"; }

fatal() {
    err "$*"
    git merge --abort 2>/dev/null || true
    exit 1
}

build_tool() {
    if [[ ! -x "${RESOLVE}" ]]; then
        echo "::group::Building resolve-conflicts tool"
        (cd "${SCRIPT_DIR}/resolve-conflicts" && go build -o resolve-conflicts .) || fatal "Failed to build resolve-conflicts tool"
        echo "::endgroup::"
    fi
}

merge_target_branch() {
    echo "::group::Merge ${TARGET_BRANCH} into ${SYNC_BRANCH}"
    git config user.name "github-actions[bot]"
    git config user.email "41898282+github-actions[bot]@users.noreply.github.com"
    git checkout "${SYNC_BRANCH}"

    local merge_exit=0
    git merge "origin/${TARGET_BRANCH}" --no-edit || merge_exit=$?

    if [[ "${merge_exit}" -eq 0 ]]; then
        info "Merge completed without conflicts"
        echo "::endgroup::"
        return 1
    fi

    # Verify merge is actually in progress (not a different failure like missing identity)
    if [[ ! -f .git/MERGE_HEAD ]]; then
        echo "::endgroup::"
        fatal "git merge failed (exit ${merge_exit}) but no merge in progress — check git config and permissions"
    fi

    info "Merge conflicts detected"
    echo "::endgroup::"
    return 0
}

get_conflicting_files() {
    git diff --name-only --diff-filter=U
}

get_go_ceiling() {
    if [[ -n "${GO_CEILING_OVERRIDE}" ]]; then
        echo "${GO_CEILING_OVERRIDE}"
        return
    fi
    "${RESOLVE}" go-ceiling --image "${GO_CEILING_IMAGE}" || fatal "Failed to detect Go version ceiling from ${GO_CEILING_IMAGE}"
}

restore_downstream_only_files() {
    echo "::group::Restoring downstream-only files"

    local files
    files=$(git diff --name-only --diff-filter=A "${MERGE_BASE}" "origin/${TARGET_BRANCH}" 2>/dev/null || true)
    if [[ -n "${DOWNSTREAM_ONLY_FILES}" ]]; then
        files="${files}"$'\n'"${DOWNSTREAM_ONLY_FILES}"
    fi

    local count=0
    while IFS= read -r file; do
        [[ -z "${file}" ]] && continue
        if git show "origin/${TARGET_BRANCH}:${file}" &>/dev/null; then
            git checkout "origin/${TARGET_BRANCH}" -- "${file}" && {
                git add "${file}"
                info "Restored: ${file}"
                count=$((count + 1))
            }
        fi
    done <<< "${files}"
    info "Restored ${count} downstream-only files"
    echo "::endgroup::"
}

resolve_fips() {
    echo "::group::Resolving FIPS-related conflicts"
    "${RESOLVE}" fips || true
    echo "::endgroup::"
}

resolve_image_versions() {
    echo "::group::Resolving image version conflicts"
    local conflicts
    conflicts=$(get_conflicting_files | grep -iE '(^|/)(Dockerfile|docker-compose)' || true)
    if [[ -z "${conflicts}" ]]; then
        info "No image version conflicts"
        echo "::endgroup::"
        return
    fi

    while IFS= read -r f; do
        [[ -z "${f}" ]] && continue
        if ! "${RESOLVE}" dockerfile --file "${f}"; then
            info "Cannot auto-resolve ${f} — will remain as conflict"
            continue
        fi
        git add "${f}"
        info "Resolved: ${f}"
    done <<< "${conflicts}"
    echo "::endgroup::"
}

resolve_gomod() {
    local go_ceiling="$1"
    echo "::group::Resolving go.mod conflicts"
    local conflicts
    conflicts=$(get_conflicting_files | grep '/\?go\.mod$' || true)
    if [[ -z "${conflicts}" ]]; then
        info "No go.mod conflicts"
        echo "::endgroup::"
        return
    fi

    local tmpdir
    tmpdir=$(mktemp -d)

    # Collect directories containing resolved go.mod files for later tidy
    local gomod_dirs=()

    while IFS= read -r f; do
        [[ -z "${f}" ]] && continue
        git show "${MERGE_BASE}:${f}" > "${tmpdir}/base.mod" 2>/dev/null || echo "module unknown" > "${tmpdir}/base.mod"
        if ! git show "MERGE_HEAD:${f}" > "${tmpdir}/ours.mod" 2>/dev/null; then
            info "Cannot auto-resolve ${f}: modify/delete conflict — will remain as conflict"
            continue
        fi
        if ! git show "HEAD:${f}" > "${tmpdir}/theirs.mod" 2>/dev/null; then
            info "Cannot auto-resolve ${f}: modify/delete conflict — will remain as conflict"
            continue
        fi
        if ! "${RESOLVE}" gomod --base "${tmpdir}/base.mod" --ours "${tmpdir}/ours.mod" --theirs "${tmpdir}/theirs.mod" \
            --go-ceiling "${go_ceiling}" --output "${f}"; then
            info "Cannot auto-resolve ${f} — will remain as conflict"
            continue
        fi
        git add "${f}"
        gomod_dirs+=("$(dirname "${f}")")
        info "Resolved: ${f} (ceiling: ${go_ceiling})"
    done <<< "${conflicts}"
    rm -rf "${tmpdir}"
    echo "::endgroup::"

    echo "::group::Running go mod tidy"
    for dir in "${gomod_dirs[@]}"; do
        local sum="${dir}/go.sum"
        [[ "${dir}" == "." ]] && sum="go.sum"
        if [[ -f "${sum}" ]] && grep -q '^<<<<<<< ' "${sum}" 2>/dev/null; then
            rm "${sum}"
            info "Removed conflicted ${sum} (will be regenerated)"
        fi

        info "Running go mod tidy in ${dir}"
        if ! (cd "${dir}" && go mod tidy); then
            err "go mod tidy failed in ${dir} — go.sum may need manual fix"
            git checkout -- "${dir}/go.mod" "${dir}/go.sum" 2>/dev/null || true
            echo "${dir}" >> /tmp/go-mod-tidy-failed.txt
            continue
        fi
        git add "${dir}/go.mod" "${dir}/go.sum"
    done
    echo "::endgroup::"
}

resolve_workflows() {
    echo "::group::Resolving workflow conflicts"
    local conflicts
    conflicts=$(get_conflicting_files | grep -E '^\.github/workflows/.*\.ya?ml$' || true)
    if [[ -z "${conflicts}" ]]; then
        info "No workflow conflicts"
        echo "::endgroup::"
        return
    fi

    while IFS= read -r f; do
        [[ -z "${f}" ]] && continue
        if ! "${RESOLVE}" workflow --file "${f}"; then
            info "Cannot auto-resolve ${f} — will remain as conflict"
            continue
        fi
        git add "${f}"
        info "Resolved: ${f}"
    done <<< "${conflicts}"
    echo "::endgroup::"
}

resolve_take_upstream() {
    [[ -z "${TAKE_UPSTREAM_PATTERNS}" ]] && return
    echo "::group::Resolving conflicts (take upstream)"
    local remaining
    remaining=$(get_conflicting_files)
    [[ -z "${remaining}" ]] && { echo "::endgroup::"; return; }

    while IFS= read -r pattern; do
        [[ -z "${pattern}" ]] && continue
        local matches
        matches=$(echo "${remaining}" | grep -E "${pattern}" || true)
        [[ -z "${matches}" ]] && continue
        while IFS= read -r f; do
            [[ -z "${f}" ]] && continue
            # Handle modify/delete conflicts: upstream may have deleted the file
            if git ls-tree HEAD -- "${f}" &>/dev/null; then
                git checkout --ours "${f}" || fatal "Failed to resolve ${f}"
            else
                git rm "${f}" || fatal "Failed to resolve ${f}"
            fi
            git add "${f}"
            echo "${f}" >> /tmp/resolved-take-upstream.txt
            info "Resolved (upstream pattern '${pattern}'): ${f}"
        done <<< "${matches}"
        remaining=$(get_conflicting_files)
    done <<< "${TAKE_UPSTREAM_PATTERNS}"
    echo "::endgroup::"
}

resolve_take_downstream() {
    [[ -z "${TAKE_DOWNSTREAM_PATTERNS}" ]] && return
    echo "::group::Resolving conflicts (take downstream)"
    local remaining
    remaining=$(get_conflicting_files)
    [[ -z "${remaining}" ]] && { echo "::endgroup::"; return; }

    while IFS= read -r pattern; do
        [[ -z "${pattern}" ]] && continue
        local matches
        matches=$(echo "${remaining}" | grep -E "${pattern}" || true)
        [[ -z "${matches}" ]] && continue
        while IFS= read -r f; do
            [[ -z "${f}" ]] && continue
            # Handle modify/delete conflicts: downstream may have deleted the file
            if git ls-tree MERGE_HEAD -- "${f}" &>/dev/null; then
                git checkout --theirs "${f}" || fatal "Failed to resolve ${f}"
            else
                git rm "${f}" || fatal "Failed to resolve ${f}"
            fi
            git add "${f}"
            echo "${f}" >> /tmp/resolved-take-downstream.txt
            info "Resolved (downstream pattern '${pattern}'): ${f}"
        done <<< "${matches}"
        remaining=$(get_conflicting_files)
    done <<< "${TAKE_DOWNSTREAM_PATTERNS}"
    echo "::endgroup::"
}

main() {
    info "Starting: ${SYNC_BRANCH} <- ${TARGET_BRANCH}"
    : > /tmp/resolved-take-upstream.txt
    : > /tmp/resolved-take-downstream.txt
    rm -f /tmp/go-mod-tidy-failed.txt

    if [[ "${RESOLVE_CONFLICTS}" == "true" ]]; then
        build_tool
    fi

    if ! merge_target_branch; then
        info "No conflicts to resolve"
    elif [[ "${RESOLVE_CONFLICTS}" != "true" ]]; then
        info "Conflict resolution disabled, pushing unresolved branch"
    else
        MERGE_BASE=$(git merge-base HEAD "origin/${TARGET_BRANCH}") || fatal "Failed to compute merge-base"
        info "Merge base: ${MERGE_BASE}"

        restore_downstream_only_files

        local go_ceiling
        go_ceiling=$(get_go_ceiling)
        info "Go version ceiling: ${go_ceiling}"

        resolve_fips
        resolve_image_versions
        resolve_gomod "${go_ceiling}"
        resolve_workflows
        resolve_take_upstream
        resolve_take_downstream

        echo "::group::Checking for remaining conflicts"
        local remaining
        remaining=$(get_conflicting_files)
        if [[ -n "${remaining}" ]]; then
            local remaining_count
            remaining_count=$(echo "${remaining}" | grep -c .)
            err "Unresolved conflicts remain (${remaining_count} files):"
            echo "${remaining}" | while IFS= read -r f; do err "  - ${f}"; done
            echo "::endgroup::"

            echo "${remaining}" > /tmp/unresolved-conflicts.txt

            info "Pushing sync branch (upstream commit only) for local continuation"
            git merge --abort 2>/dev/null || true
            git push -f origin "${SYNC_BRANCH}" 2>/dev/null || true

            err "Cannot proceed with unresolved conflicts — see job summary for resolution guide"
            exit 1
        fi
        info "All conflicts resolved"
        echo "::endgroup::"

        if [[ ! -f .git/MERGE_HEAD ]]; then
            fatal "Merge state lost during resolution — .git/MERGE_HEAD missing"
        fi

        echo "::group::Committing"
        git add -A
        git commit --no-edit || fatal "git commit failed"
        echo "::endgroup::"
    fi

    echo "::group::Pushing"
    git push -f origin "${SYNC_BRANCH}" || fatal "git push failed"
    echo "::endgroup::"

    if [[ -s /tmp/go-mod-tidy-failed.txt ]]; then
        err "go mod tidy failed for some modules — see PR description for details"
        exit 1
    fi

    info "Done"
}

main "$@"
