# V88 public runtime topology

`runtime-topology.manifest.json` is the non-secret, source-owned topology contract. Git `v88` is the maintained source authority; ECS is a runtime copy. This document does not authorize a release, cleanup or service retirement. Existing approved release procedures and exact-SHA rollback requirements still apply.

## Evidence and ownership

The manifest records the locally inspected `origin/v88` SHA as a source-reference snapshot, not as the current production release. Its service declarations come from the tracked `docker-compose.yml`, `nginx.conf`, build files and the dated cleanup inventory. No ECS inspection was performed for this task. Concrete deploy-time image references and component release SHAs must be verified against normalized `docker compose config --format json` and `docker inspect` evidence before making runtime changes. Record only non-secret image identities, service dependencies, symbolic storage purposes and listener ownership; never commit raw command output containing environment values or mount locations.

All Compose inputs must exist in Git except secret overlays. This includes base Compose, non-secret overrides, network/service definitions and Nginx routing configuration. Runtime changes to those inputs must be ported back into `v88`. Source-owned image selectors declare which deployment input owns an image; they are not proof that the running image was built from that source. Concrete image digests and exact release identities belong in a verified release record.

The only permitted secret material is deploy-time environment values and certificate/session storage outside Git. Secret overlays are supplied at deployment and kept outside Git. Passwords, tokens, private keys, populated environment files, host-local mount paths, certificates, login/session content, database contents and business media must not be copied into the manifest or tracked deployment files. Persistent mounts are represented by purpose only; backup and restore procedures must retain their real storage separately.

## Service roles

| Manifest ID | Runtime role | Source-owned configuration | Dependencies |
| --- | --- | --- | --- |
| `node` | Legacy opt-in compatibility entry only; never part of the default public path | `docker-compose.yml`, root `Dockerfile`, `server.js` | Go API, Browser Worker, Shuihuo compatibility |
| `go-api` | Default frontend, authenticated application/API entry, batch/business API and persistent local executor artifacts | `docker-compose.yml`, `backend/Dockerfile` | MySQL |
| `browser-worker` | 视频管理系统 browser login, session reuse and upload operations | `docker-compose.yml`, `services/121-browser-worker/Dockerfile` | None declared internally |
| `shuihuo-compat` | Retained compatibility API with isolated database and object storage | `docker-compose.yml`; retained image provenance remains a gap | MySQL |
| `mysql` | Retained account/business database and isolated compatibility database | Public startup definition remains a Git ownership gap | None declared internally |
| `nginx` | Public ingress forwarding application traffic to Go | `docker-compose.yml`, `nginx.conf` | Go API |

The manifest includes runtime dependencies even when Compose does not declare startup ordering: both Go and Shuihuo compatibility use the internal MySQL alias. The current base Compose declares five services; MySQL remains external. Its exact image, startup inputs, network ownership and live listener bindings require sanitized evidence before adding a source-owned definition. Do not infer a public MySQL image from a local preview configuration.

Nginx owns the declared public port mappings: 80 to 80 and 3000 to 80. Application services expose only internal ports. HTTPS is intentionally not published: a tracked TLS listener, certificate renewal contract and external verification are required before adding port 443.

The dated inventory also mentions a separate legacy Novel Fetch worker outside the six-service target contract. Its callers and replacement relationship still require verification before any retirement. This manifest does not prove that all runtime containers have already been consolidated.

## Validation and retirement

Run `node --test tests/v88-runtime-topology.test.js` to check the Go-owned public contract, retained legacy service retirement gates, dependency references, existing repository-relative configuration paths, public listener ownership and exclusion of host paths/populated environment material.

Each service carries its own `retirementGate`. Passing this contract test validates the manifest only. Retirement additionally requires source-owned replacement configuration, authenticated business-route evidence, durable data/media readback where relevant, restorable storage and a known rollback point. Do not remove containers, networks, volumes, legacy implementations or release paths based solely on this declaration.
