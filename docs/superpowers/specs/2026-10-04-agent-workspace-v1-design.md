# Agent Workspace V1 Design

## Goal

Build a first complete, openable Agent workspace for 一战晟铭 that lets a beginner operate the system by natural language instead of learning every page first. The Agent must be able to answer system questions, navigate the app, create durable tasks, show execution status, and carry media references in chat. Image/video files are never special local Agent files: formal media assets belong in TOS and Agent messages reference `media_asset_id` only.

## Product direction

The experience combines three ideas without copying any one UI:

- Your Dot: persistent responsibilities, in-progress work, waiting/needs-decision/completed states.
- YFAI-style visual creation: image/video results live in the conversation and can become the context for the next request.
- Existing v88: the Agent is a control plane over real product capabilities such as 小说获取、剧本生成、创作漫剧、批量工厂、设置, not a second implementation of those capabilities.

## V1 scope

V1 must be directly openable at `/agent` and provide a complete first working loop:

1. Thread list and new-thread creation.
2. Chat message history persisted in MySQL.
3. Composer with text plus explicit context chips for project/media/resource references.
4. Assistant responses persisted in the same thread.
5. Tool registry with working navigation tools for:
   - 小说获取 `/novel-fetch`
   - 剧本生成 `/script`
   - 水货生产 `/shuihuo-production`
   - 创作漫剧 `/shuihuo-production/creative`
   - 批量工厂 `/batch-factory`
   - Agent `/agent`
   - 设置 `/settings`
   - API 配置 `/api-config`
6. Natural-language navigation requests such as “帮我打开小说获取” are converted into an assistant message plus a `ui.navigate` tool call. The browser executes the navigation; the Agent does not fake completion of backend work.
7. Durable task records with `in_progress / waiting / needs_decision / completed / failed` states and task cards in chat/sidebar.
8. Durable tool-call records so “AI说话”和“AI在执行” are distinct.
9. Media-reference cards in the conversation. V1 can display existing `media_asset_id` references; actual image/video generation and TOS upload adapters are a later tool implementation, but the data contract is already TOS-first.
10. Right-side inspector is contextual and optional, not permanently open. It shows the selected task/tool/media details.
11. The existing Batch Factory page remains available and is not redesigned in this task.

## Non-goals for V1

- Do not recreate old v88 Agent JSON/Canvas storage.
- Do not implement a second novel-fetch/script/video pipeline inside Agent.
- Do not pretend image/video generation works when no provider tool is wired yet.
- Do not store images/videos on the server filesystem as durable business assets.
- Do not redesign other v88 pages in this task.
- GitHub Actions policy is not changed in this task.

## Architecture

```text
React Agent Workspace
        |
        v
Go HTTP API
        |
        v
Agent Service
  |       |       |
  |       |       +--> Tool Registry --> product navigation / future product services
  |       +----------> Task Service
  +------------------> Thread / Message Service
        |
        v
MySQL (durable source of truth)

Future async product tools:
Agent Tool -> existing Go service -> Redis job id -> Go Worker -> provider -> TOS -> media_assets -> Agent result message
```

Redis is transport only. MySQL is authoritative for Agent threads/messages/tasks/tool calls. TOS is authoritative for formal image/video bytes.

## Data model

### `agent_threads`
- `id VARCHAR(64)` primary key
- `owner VARCHAR(191)`
- `title VARCHAR(255)`
- `status VARCHAR(32)` default `active`
- `created_at`, `updated_at`

### `agent_messages`
- `id VARCHAR(64)` primary key
- `thread_id VARCHAR(64)`
- `owner VARCHAR(191)`
- `role VARCHAR(16)` (`user`, `assistant`, `system`)
- `content MEDIUMTEXT`
- `metadata_json JSON NULL`
- `created_at`

### `agent_tasks`
- `id VARCHAR(64)` primary key
- `thread_id VARCHAR(64)`
- `owner VARCHAR(191)`
- `title VARCHAR(255)`
- `status VARCHAR(32)` (`in_progress`, `waiting`, `needs_decision`, `completed`, `failed`)
- `progress_current INT`, `progress_total INT`
- `detail TEXT NULL`
- `created_at`, `updated_at`

### `agent_tool_calls`
- `id VARCHAR(64)` primary key
- `thread_id VARCHAR(64)`
- `message_id VARCHAR(64) NULL`
- `owner VARCHAR(191)`
- `tool_name VARCHAR(96)`
- `status VARCHAR(32)` (`proposed`, `completed`, `failed`)
- `arguments_json JSON`
- `result_json JSON NULL`
- `created_at`, `updated_at`

### `agent_message_media`
- `message_id VARCHAR(64)`
- `media_asset_id VARCHAR(64)`
- `ordinal INT`
- no raw file path, provider URL, or local artifact path as source of truth.

A future global `media_assets` table owns TOS `bucket/key/content_type/size/checksum` and signing metadata. V1 deliberately stores only media references so it cannot create a conflicting storage model.

## Agent response strategy

V1 uses a deterministic command interpreter first so the page is useful even before a general LLM configuration is added. It recognizes system-navigation intents and common help questions. Examples:

- “打开小说获取” -> `ui.navigate { path: "/novel-fetch" }`
- “批量工厂在哪里” -> assistant explanation + `ui.navigate { path: "/batch-factory" }`
- “图片模型在哪里设置” -> explanation + `ui.navigate { path: "/api-config", section: "image-models" }`

Unrecognized requests receive a transparent response explaining that the current Agent V1 can operate navigation/system help and that creative execution tools are being added to the same registry. It must not claim a generation/job happened when it did not.

A future `AgentResponder` can use the unified text-model configuration, but must produce structured tool calls that use the same registry.

## UI structure

Desktop:

```text
+----------------+------------------------------------------+--------------------+
| Threads        | Agent Chat                               | Inspector (optional)|
|                |                                          |                    |
| New chat       | header: Agent / status                   | selected task/tool |
| Today          |                                          | or media reference |
| thread...      | user message                             |                    |
|                | assistant message                        |                    |
| Tasks          | tool status card                         |                    |
| in progress 3  | task card                                |                    |
| needs decide 1 | media cards                              |                    |
| completed      |                                          |                    |
|                |------------------------------------------|                    |
|                | context chips + composer + send          |                    |
+----------------+------------------------------------------+--------------------+
```

The inspector collapses by default. On narrower screens the left sidebar becomes a drawer and the inspector becomes a modal/drawer.

Visual language: dark workbench, restrained borders, large readable chat column, soft blue/violet Agent accent, media/result cards with high contrast. Tool-call cards use product language (“正在打开小说获取”) rather than developer function syntax.

## API contract

- `GET /api/agent/threads`
- `POST /api/agent/threads` `{title?}`
- `GET /api/agent/threads/{threadID}` -> thread + messages + tasks + tool_calls
- `POST /api/agent/threads/{threadID}/messages` `{content, media_asset_ids?}` -> persists user message, interprets request, persists assistant message and any tool call, returns updated records.
- `GET /api/agent/tasks?thread_id=...`

Owner comes from the server-side owner resolver, never a client-supplied owner field.

## Security and correctness

- Thread/task/message reads and writes always filter by owner.
- Unknown thread IDs return 404, not empty cross-user data.
- JSON bodies are bounded and strict enough to reject malformed payloads.
- Navigation tool only accepts paths from a server-side allow-list; arbitrary URL redirects are not allowed.
- Tool-call arguments are JSON and the frontend executes only known `ui.*` tools.
- Media IDs are opaque references; V1 never accepts local filesystem paths as media references.

## Success criteria

V1 is complete when:

- `/agent` opens in the embedded React app.
- A user can create a thread, send messages, refresh, and see the same history from MySQL.
- “帮我打开小说获取” returns a visible tool card and the UI can navigate to `/novel-fetch`.
- Common system-location questions return actionable answers.
- The sidebar displays durable task counts/statuses.
- Tool and task cards are distinct from normal assistant text.
- Existing `/batch-factory` still opens.
- No Agent durable-media field introduces local file paths or provider URLs as formal assets.
