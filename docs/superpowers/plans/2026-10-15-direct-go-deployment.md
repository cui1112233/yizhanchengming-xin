# Direct Go Binary Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the active application deployment path with direct execution of a Go binary, while keeping MySQL, Redis, TOS, Nginx, and rollback data safe.

**Architecture:** Build React locally, copy the resulting assets into the Go embed tree, cross-compile `qiantie` for Linux amd64, and transfer a versioned release archive to the server. A systemd service will run the selected binary directly; activation will atomically switch `current`, restart the service, verify `/health` and `/api/runtime-build-info`, and restore the previous release on failure. Historical Docker/GitHub workflow files remain archived/read-only until the repository consolidation gate permits physical deletion.

**Tech Stack:** PowerShell, Bash, Go, React/Vite, systemd, Nginx, MySQL, Redis, TOS.

**Spec:** `docs/superpowers/specs/2026-10-14-go-runtime-consolidation-design.md`

## Global Constraints

- Git `v88` remains the only maintained source line.
- The application runtime is a directly launched Go binary; Node is build-time only.
- Docker is not used by the active application deployment path.
- MySQL, Redis, TOS, FFmpeg, and Nginx are not deleted or re-created by an application release.
- Every release carries an exact Git SHA and a rollback pointer.
- No deployment script may print credentials or copy secret files into Git.

## Review Focus

- Dirty worktrees must be rejected before packaging so a release is traceable.
- A missing or non-executable Go binary must fail before changing `current`.
- A failed health check must restore the previous binary and service state.
- Nginx must proxy to the host Go listener rather than a container DNS name.
- Historical Docker/Action files must not remain an active release entry point.

---

### Task 1: Direct release package and service contract

**Files:**
- Modify: `deploy/v88-direct/build-direct-release.ps1`
- Create: `deploy/v88-direct/qiantie.service`
- Create: `deploy/v88-direct/direct-release-contract.test.js`

**Interfaces:**
- Produces `release/v88-direct-<40-char-sha>.tar.gz` containing `go/qiantie`, `runtime-release.json`, and the service template.
- The service reads `/etc/qiantie/v88.env` and serves on `:3000`.

- [ ] Write a failing contract test that rejects Docker/GHCR commands and requires the service template and exact-SHA archive layout.
- [ ] Run the contract test and verify it fails against the current compose-based activation flow.
- [ ] Implement the package/service changes.
- [ ] Run the contract test and verify it passes.
- [ ] Commit the focused package change.

### Task 2: Direct activation, rollback, and Nginx target

**Files:**
- Modify: `deploy/v88-direct/send-direct-release.ps1`
- Modify: `deploy/v88-direct/activate-direct-release.sh`
- Modify: `deploy/v88-public/nginx.conf`
- Create: `deploy/v88-direct/direct-activation.test.js`

**Interfaces:**
- `activate-direct-release.sh <exact-40-char-sha>` switches `/opt/qiantie/v88/direct/current`, restarts `qiantie.service`, checks health/build identity, and rolls back on failure.
- Nginx proxies to `127.0.0.1:3000`.

- [ ] Write a failing test covering direct service activation, rollback commands, and the host Nginx upstream.
- [ ] Run it and verify the current scripts fail because they call Docker Compose.
- [ ] Replace the activation path with systemd and host-process checks.
- [ ] Run the focused test and shell syntax checks.
- [ ] Commit the direct activation change.

### Task 3: Disable automated Docker/Action release entry points

**Files:**
- Modify: `deploy/v88-public/README.md`
- Modify: `deploy/v88-public/runtime-topology.manifest.json`
- Create: `deploy/v88-direct/README.md`
- Archive: `.github/workflows/*` as read-only historical references outside the active workflow directory when the consolidation gate allows it.

**Interfaces:**
- Documentation and checks identify the direct binary scripts as the only active release path.
- No GitHub Action compiles or publishes the application.

- [ ] Add a failing static test that detects an active Docker/GHCR release entry point.
- [ ] Run it and verify the current repository fails.
- [ ] Update the active deployment documentation and guard.
- [ ] Run the complete relevant test suite.
- [ ] Commit the release-path consolidation.

### Task 4: End-to-end static verification

- [ ] Run `go test ./...` from `backend`.
- [ ] Run all direct deployment contract tests.
- [ ] Run PowerShell parser checks for modified `.ps1` files and Bash syntax checks for modified `.sh` files.
- [ ] Record the resulting commit SHA and explicitly report any historical files preserved by the consolidation gate.
