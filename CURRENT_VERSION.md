# V88 Current Version

Updated: 2026-09-08

## Canonical source

- Maintained branch: `v88`.
- `v88` is the only maintained application source and the future total branch.
- GitHub default branch is still `master`; changing the default branch belongs to consolidation stage 3 and has not been performed yet.
- Historical V78 / feature / fix / integration / ops / release / archive / tmp branches are reference-only during the consolidation freeze.
- Repository inventory currently contains 147 branches. Do not delete them before consolidation stage 7.

## Current public application lineage

The public Node application is now deployed from the `v88` Git lineage.

Last fully verified public Node release from the current Direct Stage -> Cutover path:

`7390234d52031e63b1f6d169b86f90f510e3c9e2`

That release passed:

- exact-SHA Stage deployment;
- parallel host Node health verification;
- public route preservation before cutover;
- Cutover;
- external `/api/build-info` exact-SHA verification;
- rollback protection.

The Git `v88` HEAD may be newer than the public runtime SHA because documentation, rules, tests, or future un-released work can be committed after a production release. Do not assume `HEAD == production` without checking `/api/build-info` or the release record.

## Current formal Node release path

Normal Node/public releases use:

`Git v88 -> V88 Direct Deploy Node Stage -> exact staged SHA -> CUTOVER-REQUEST -> V88 Direct Deploy Node Cutover -> external exact-SHA verification`

The old automatic Docker/GHCR public image release is retired as the normal application release path.

`v88-node-emergency-redeploy.yml`, `*-once.yml`, registry diagnostics, and other historical recovery workflows are not normal release entry points during consolidation.

## Current production runtime mode

Production is currently **hybrid**, not fully migrated away from Docker.

The Git-direct host Node stage uses:

- pinned Node runtime `24.19.0`;
- release directories under `/opt/qiantie/releases/v88-stage/`;
- systemd service `qiantie-v88-node-stage.service`;
- stage port `18081`;
- current persistent data/output mounts.

It still depends on the existing production Docker network and services, including:

- Go API;
- Browser Worker;
- Nginx;
- the current Docker network used to resolve those upstreams;
- existing production environment/mount information inherited from the running Node container during staging.

Therefore “daily Node releases are Git-direct” must not be interpreted as “Docker can be deleted now”. Production-runtime consolidation is stage 6.

## Deploy configuration warning

Git `v88/deploy/v88-public/` currently contains the Browser Worker Compose overlay, while existing production workflows also reference an ECS-side base file at:

`/opt/qiantie/v88/deploy/v88-public/docker-compose.yml`

Until stage 6 makes all required production runtime configuration formally traceable from Git, do not remove or overwrite that ECS runtime configuration.

## Maintained application scope already present in V88

The V88 consolidation line includes the actively maintained website/application work, including:

- Batch Factory V11 frontend and Go runtime integration;
- Novel Fetch / 视频管理系统 integration and Browser Worker contract;
- Script / storyboard / constraint / prompt-pipeline work;
- Node-to-Go bridge runtime;
- local executor server integration;
- maintained desktop local executor source under `executors/local-executor/`;
- Windows local executor build/update pipeline;
- account/member/team and other website functionality inherited and consolidated into the V88 source line;
- current regression and deployment contracts required to keep these functions stable.

Feature completeness is not assumed merely because a file exists in V88. Consolidation stage 2 must explicitly verify that every production-used capability is present and authoritative in `v88` before old branches are removed.

## Technology conventions

- API/backend target: Go binary.
- Admin: React + Ant Design.
- User frontend: React + Ant Design.
- Database: MySQL + Goose migrations.
- Queue / locks / non-relational runtime: Redis.
- Object storage: TOS.
- Prompt templates: backend-managed and replaced with real prompts at runtime.
- Go embed remains the preferred final packaging direction where applicable, while the current production stack is still in a mixed Node/Go migration state.

## Consolidation program

1. Inventory and freeze — baseline established.
2. V88 feature consolidation — next active stage.
3. Git rule consolidation.
4. GPT multi-chat collaboration consolidation.
5. Release-path consolidation.
6. Production-runtime consolidation.
7. Delete obsolete branches / workflows.

Read before any new V88 work:

- `AGENTS.md`
- `docs/V88_PROJECT_EXECUTION_MEMORY.md`
- `docs/obj/2026-09-08-v88-consolidation-stage1-inventory-freeze.md`

## Historical production snapshot

`PRODUCTION_SNAPSHOT.md` is a historical V78.3.0.3 extraction record. It is retained for provenance only and must not be treated as the current production source of truth.
