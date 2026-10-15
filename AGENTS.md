# V88 Agent Execution Rules

All agents, chats, Codex sessions, and automation working on this repository must follow these rules before changing or deploying V88.

1. **Git branch `v88` is the canonical maintained source and future total branch.** Do not treat production/ECS files, old V78 branches, historical releases, or hotfix bundles as the source of truth.
2. **Every feature and bug fix must be written into Git first.** Normal flow: feature/fix branch -> tests -> merge into `v88` -> record commit SHA -> deploy/sync production.
3. **Production/ECS is a runtime copy, not a development source.** Do not complete a task by editing only the running server/container. Emergency production edits must be ported back to `v88` immediately; otherwise the task is incomplete.
4. **Before modifying or deploying, verify the current `v88` HEAD.** Never overwrite current V88 with an older `app.js`, old hotfix bundle, old image, or historical branch snapshot.
5. **Daily small changes should use the Go binary release path.** Frontend changes rebuild React and then the embedded Go binary; Go changes rebuild/restart Go; migrations run Goose only when required; Worker changes update/restart Worker only when required.
6. **Do not interpret “no Docker for daily releases” as “delete all existing Docker infrastructure.”** Existing MySQL/Redis/Worker/containerized infrastructure may remain until intentionally migrated. Docker is not the default release mechanism for small application changes.
7. **Every production release must be traceable to a Git SHA** and have a known previous stable SHA/rollback point.
8. **User-visible `121` naming must be `视频管理系统`.** Internal compatibility names may retain `121`.
9. **Novel Fetch confirmed semantics must not regress:** preserve full `originalRaw`; process only up to `maxTxt`; display processed/raw counts such as `4000/27831`; `input_ready` displays `分类信息已就绪`; `running/processing` displays `正在执行中…`; date filters must use real task dates and must not be bypassed with fake dates such as `2099-12-31`.
10. If the user says only **“执行”** or **“发布”**, default to the latest Git `v88`, ensure the change is committed to Git first, and use the currently approved deployment path. Do not silently choose direct production edits or a full Docker rebuild.
10a. **源码优先，禁止编译产物覆盖源码或页面。** Never copy `frontend/dist`, `frontend/public`, server static files, or any other compiled artifact back as the implementation of a page or feature. Maintain the React/source implementation first, run a local build and verification, then publish the newly built result through Go.

## V88 consolidation freeze

The repository is now in the seven-stage consolidation program:

1. inventory and freeze;
2. V88 feature consolidation;
3. Git rule consolidation;
4. multi-chat / GPT collaboration consolidation;
5. release-path consolidation;
6. production-runtime consolidation;
7. delete obsolete branches and workflows.

During stages 1-6, apply these additional rules:

11. **Historical branches are read-only reference sources.** Do not continue maintained development on V78, old feature/fix/integration/ops/design/release/archive/tmp branches. New maintained work must end in `v88`.
12. **Do not delete historical branches or workflows before stage 7.** First prove their functionality and runtime dependencies have been consolidated and replaced.
13. **The approved application release path is the Go embed binary script.** Use `scripts/deploy-go-embedded.ps1`; do not introduce or restore a Node release path.
14. **Historical Docker/Node/GHCR workflows are not release entry points.** They are outside the refactored application and must not be selected for ordinary deployment.
15. **The refactored application runtime is Go-only.** MySQL, Redis, TOS, FFmpeg and external connectors remain infrastructure dependencies, but the application API and embedded frontend are served by the Go binary.
16. **Before starting a V88 task in any chat, read the current consolidation inventory OBJ and `docs/V88_PROJECT_EXECUTION_MEMORY.md`.** Do not rely on stale `master`, historical production snapshots, or prior-chat assumptions.

For the full project memory and deployment rationale, read:

`docs/V88_PROJECT_EXECUTION_MEMORY.md`

For the current consolidation inventory/freeze baseline, read:

`docs/obj/2026-09-08-v88-consolidation-stage1-inventory-freeze.md`

## 重构模式补充规则

本项目当前按全量重构处理，不为旧 Node 发布链路保留兼容入口。`deploy/v88-direct` 已退出维护并删除；部署统一使用 `scripts/deploy-go-embedded.ps1`。历史文档只作为背景资料，不得重新创建已删除的 Node 部署目录或脚本。
