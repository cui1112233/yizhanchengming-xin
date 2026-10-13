# Public Nginx connection reuse Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the tracked V88 public Nginx configuration match the live HTTPS entry and reuse repeat HTTPS connections.

**Architecture:** The existing public Compose stack mounts one Nginx server configuration file. It will retain Node-only proxying and add the active HTTPS entry with bounded TLS session reuse.

**Tech Stack:** Nginx 1.27, Docker Compose, Node built-in test runner.

## Global Constraints

- Maintain V88 as the source of truth; do not patch production only.
- Do not commit certificates, keys, credentials, sessions, or user data.
- Preserve Node-only public ingress and dynamic API behavior.

---

### Task 1: Lock the public ingress contract

**Files:**
- Modify: `tests/v88-public-unified-topology.test.js`

- [ ] Add failing assertions for the read-only `/etc/letsencrypt` mount,
  `yzcm.site` HTTPS server, and `ssl_session_cache shared:SSL:10m;`.
- [ ] Run `node --test tests/v88-public-unified-topology.test.js` and confirm
  it fails before the configuration change.

### Task 2: Restore HTTPS ingress and connection reuse

**Files:**
- Modify: `deploy/v88-public/nginx.conf`
- Modify: `deploy/v88-public/docker-compose.yml`

- [ ] Add `./certs:/etc/letsencrypt:ro` to Nginx volumes.
- [ ] Restore the active HTTP redirect and HTTP/2 HTTPS `yzcm.site` servers;
  proxy application traffic only to `v88-node:3000`.
- [ ] Add `ssl_session_cache shared:SSL:10m;`, `ssl_session_timeout 1d;`,
  `ssl_session_tickets on;`, and `keepalive_requests 1000;`.
- [ ] Run Compose rendering and the public-topology test.
- [ ] Commit only the Nginx configuration, its test, and these task documents.
