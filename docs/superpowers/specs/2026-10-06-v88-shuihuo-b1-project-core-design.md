# V88 Shuihuo B1 Project Core Design

## Scope

B1 restores only the ordinary Shuihuo Project Core needed by the old v88 product. It does not implement the global shell, Batch Factory V11, shared assets/media infrastructure, model/provider ownership, or a second runtime.

## Architectural boundaries

- `shuihuo_projects` and `shuihuo_segments` are Shuihuo-specific business entities and persist in MySQL through Goose migrations.
- Assets, asset images, media, storage, model catalog, provider config, credentials, Redis queues, schedulers, workers and VIDEO runtime are explicitly out of B1.
- Authentication remains the existing Go Session/Auth stack. No token/localStorage auth is introduced.
- Browser localStorage is not a project/source/segment/task/media/provider fact source.
- Shuihuo smart segmentation consumes an injected text-segmentation capability. B1 does not own a model catalog, provider registry or credential store.
- B1 mutations use the existing same-origin/CSRF semantics and existing capability middleware.

## Domain model

### Project

A Shuihuo project stores:

- ID
- owner user ID
- optional team ID
- name
- production mode
- original source text
- segmentation status (`draft` or `confirmed`)
- created/updated timestamps

Ownership permits the owner, a member of the same non-zero team, or an explicit administrative bypass used only after the existing Auth layer has authorized the request.

### Segment

A segment belongs to one Shuihuo project and stores:

- ID
- project ID
- 1-based contiguous position
- source text
- subtitle text
- speaker
- created/updated timestamps

Deleting/reordering segments keeps positions contiguous.

## Required behaviors

A normal Shuihuo project must support:

1. create/list/read/delete;
2. replace source text, which resets segmentation to draft and clears stale confirmed segments;
3. paragraph segmentation;
4. fixed-line segmentation;
5. imported segmentation, including v88 exported `index<TAB>text` rows;
6. smart segmentation through an injected shared text capability;
7. confirm segmentation, replacing persisted segments transactionally and marking the project confirmed;
8. segment create/update/delete/reorder;
9. ownership isolation across users/teams;
10. reload persistence from MySQL.

## HTTP surface

B1 owns the Shuihuo-specific API surface below `/api/v1/shuihuo-production`:

- `GET /projects`
- `POST /projects`
- `GET /projects/{id}`
- `DELETE /projects/{id}`
- `PUT /projects/{id}/source`
- `POST /projects/{id}/segmentation/paragraphs`
- `POST /projects/{id}/segmentation/fixed`
- `POST /projects/{id}/segmentation/import`
- `POST /projects/{id}/segmentation/smart`
- `POST /projects/{id}/segmentation/confirm`
- `POST /projects/{id}/segments`
- `PUT /segments/{id}`
- `DELETE /segments/{id}`
- `PUT /projects/{id}/segments/order`

Reads use the existing authenticated/capability boundary. Mutations additionally use the existing same-origin boundary.

## Smart segmentation boundary

B1 defines a narrow `SmartSegmenter` consumer interface. The production adapter may reuse the already-wired shared text generation provider, but B1 must not create any `shuihuo_model_catalog`, provider configuration, credential table, or provider registry. Task F remains the owner of system model/provider configuration.

## Explicit non-goals

- shared assets domain;
- asset images/version/primary image/project binding/segment binding;
- shared storage/TOS implementation;
- Shuihuo media production;
- taskruntime relation and media task status;
- health/readiness aggregation;
- Batch Factory production workbench;
- global Router/UserLayout/AuthBoundary changes;
- frontend product UI.

Those belong to B2/B3/B4 or tasks A/C/F as already assigned.
