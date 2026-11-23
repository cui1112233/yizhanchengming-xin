export type TaskSource = "script" | "agent" | "fetch" | "batch" | "shuihuo";
export type TaskStatus = "running" | "attention" | "failed" | "completed" | "unknown";
export type TaskFilter = "全部" | "进行中" | "待确认" | "已完成";

export type AggregatedTask = {
  source: TaskSource;
  projectId?: string;
  sourceLabel: string;
  id: string | undefined;
  title: string | undefined;
  createdAt: string | undefined;
  rawStatus: string | undefined;
  status: TaskStatus;
  progress: { completed: number; total: number } | undefined;
};

export type TaskSourceResult = {
  source: TaskSource;
  sourceLabel: string;
  state: "ready" | "error";
  error: string | undefined;
};

export type TaskSummary = {
  entries: AggregatedTask[];
  sources: TaskSourceResult[];
};

export type TaskRequest = (path: string) => Promise<unknown>;

type SourceDefinition = {
  source: TaskSource;
  sourceLabel: string;
  path: string;
  collection: "entries" | "tasks" | "batches" | "projects" | "runs";
};

type InternalTask = AggregatedTask & {
  sourceOrder: number;
  entryOrder: number;
};

const TASK_SOURCE_DEFINITIONS: SourceDefinition[] = [
  { source: "script", sourceLabel: "剧本", path: "/api/history", collection: "entries" },
  { source: "agent", sourceLabel: "Agent", path: "/api/studio/runs", collection: "runs" },
  { source: "fetch", sourceLabel: "小说获取", path: "/api/batch-rewrite/tasks", collection: "tasks" },
  { source: "batch", sourceLabel: "批量工厂", path: "/api/batch-factory/v12/batches/summary", collection: "batches" },
  { source: "shuihuo", sourceLabel: "水货生产", path: "/api/shuihuo-production/projects", collection: "projects" },
];

const COMPLETED_STATUSES = new Set(["done", "completed", "success", "succeeded", "finished", "已完成", "完成", "成功", "已合并"]);
const FAILED_STATUSES = new Set(["failed", "error", "canceled", "cancelled", "original_failed", "ai_failed", "失败", "异常"]);
const ATTENTION_STATUSES = new Set(["waiting-review", "waiting_review", "pending_confirmation", "accepted_pending", "待确认", "待审核"]);
const RUNNING_STATUSES = new Set(["queued", "pending", "running", "processing", "executing", "waiting", "ai_processing", "进行中", "处理中", "排队中", "等待中"]);

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}

function textValue(value: unknown) {
  if (typeof value === "string") {
    const normalized = value.trim();
    return normalized || undefined;
  }
  if (typeof value === "number") return String(value);
  return undefined;
}

function firstText(record: Record<string, unknown>, keys: string[]) {
  for (const key of keys) {
    const value = textValue(record[key]);
    if (value) return value;
  }
  return undefined;
}

function collectionFrom(payload: unknown, key: SourceDefinition["collection"]) {
  const value = isRecord(payload) ? payload[key] : undefined;
  return Array.isArray(value) ? value.filter(isRecord) : [];
}

export function taskStatusFromServer(value: unknown): TaskStatus {
  const status = textValue(value)?.toLowerCase();
  if (!status) return "unknown";
  if (COMPLETED_STATUSES.has(status)) return "completed";
  if (FAILED_STATUSES.has(status)) return "failed";
  if (ATTENTION_STATUSES.has(status)) return "attention";
  if (RUNNING_STATUSES.has(status)) return "running";
  return "unknown";
}

function historyVideoState(entry: Record<string, unknown>) {
  const videoTasks = entry.videoTasks;
  if (!isRecord(videoTasks)) return { status: "unknown" as TaskStatus, progress: undefined };

  const tasks = Object.values(videoTasks).filter(isRecord);
  if (!tasks.length) return { status: "unknown" as TaskStatus, progress: undefined };

  const statuses = tasks.map((task) => taskStatusFromServer(firstText(task, ["status"])));
  const completed = statuses.filter((status) => status === "completed").length;
  const progress = { completed, total: tasks.length };

  if (statuses.includes("failed")) return { status: "failed" as TaskStatus, progress };
  if (statuses.includes("running")) return { status: "running" as TaskStatus, progress };
  if (statuses.every((status) => status === "completed")) return { status: "completed" as TaskStatus, progress };
  return { status: "unknown" as TaskStatus, progress };
}

function entryFromRecord(definition: SourceDefinition, record: Record<string, unknown>, sourceOrder: number, entryOrder: number): InternalTask {
  const rawStatus = firstText(record, ["status", "aiStatus", "ai_status", "originalStatus", "original_status"]);
  const directStatus = taskStatusFromServer(rawStatus);
  const video = definition.source === "script" ? historyVideoState(record) : { status: "unknown" as TaskStatus, progress: undefined };
  const status = definition.source === "script" && video.status !== "unknown" ? video.status : directStatus;
  const titleKeys = definition.source === "script"
    ? ["preview", "title"]
    : definition.source === "agent"
      ? ["title", "name", "instruction"]
      : ["title", "name", "bookName", "book_name"];
  const request = isRecord(record.request) ? record.request : {};

  return {
    source: definition.source,
    sourceLabel: definition.sourceLabel,
    ...(definition.source === 'shuihuo' ? { projectId: textValue(record.projectId) } : {}),
    id: firstText(record, ["id", "bookId", "book_id"]),
    title: firstText(record, titleKeys) || (definition.source === "agent" ? firstText(request, ["instruction"]) : undefined),
    createdAt: firstText(record, ["updatedAt", "updated_at", "createdAt", "created_at"]),
    rawStatus,
    status,
    progress: video.progress,
    sourceOrder,
    entryOrder,
  };
}

function timestamp(entry: AggregatedTask) {
  if (!entry.createdAt) return undefined;
  const value = new Date(entry.createdAt).getTime();
  return Number.isNaN(value) ? undefined : value;
}

function compareEntries(left: InternalTask, right: InternalTask) {
  const leftTime = timestamp(left);
  const rightTime = timestamp(right);
  if (leftTime !== undefined && rightTime !== undefined && leftTime !== rightTime) return rightTime - leftTime;
  if (leftTime !== undefined && rightTime === undefined) return -1;
  if (leftTime === undefined && rightTime !== undefined) return 1;
  if (left.sourceOrder !== right.sourceOrder) return left.sourceOrder - right.sourceOrder;
  return left.entryOrder - right.entryOrder;
}

function errorMessage(cause: unknown) {
  if (cause && typeof cause === "object" && "message" in cause) {
    const message = textValue((cause as { message?: unknown }).message);
    if (message) return message;
  }
  return "服务端请求失败";
}

export async function loadTaskSummary(request: TaskRequest, options: { projectsOnly?: boolean } = {}): Promise<TaskSummary> {
  const settled = await Promise.allSettled(TASK_SOURCE_DEFINITIONS.map(async (definition) => {
    const payload = await request(definition.path);
    if (!isRecord(payload) || !Array.isArray(payload[definition.collection])) throw new Error("服务端任务列表响应不完整");
    if (definition.source !== "shuihuo") return payload;
    if (options.projectsOnly) return { projects: collectionFrom(payload, "projects").map(project => ({ ...project, projectId: textValue(project.id) })) };
    const projects = collectionFrom(payload, 'projects');
    const results = await Promise.all(projects.map(async project => {
      const id = textValue(project.id); if (!id) throw new Error('水货项目缺少 ID');
      const tasks = await request(`/api/shuihuo-production/projects/${encodeURIComponent(id)}/tasks`);
      if (!isRecord(tasks) || !Array.isArray(tasks.tasks)) throw new Error('水货任务响应不完整');
      return collectionFrom(tasks, 'tasks').map(task => ({ ...task, projectId: id, title: firstText(task, ['title', 'name']) || `${textValue(project.name) || id} · ${textValue(task.kind) || '任务'}` }));
    }));
    return { projects: results.flat() };
  }));
  const entries: InternalTask[] = [];
  const sources = settled.map((result, sourceOrder): TaskSourceResult => {
    const definition = TASK_SOURCE_DEFINITIONS[sourceOrder];
    if (result.status === "rejected") {
      return {
        source: definition.source,
        sourceLabel: definition.sourceLabel,
        state: "error",
        error: errorMessage(result.reason),
      };
    }

    collectionFrom(result.value, definition.collection).forEach((record, entryOrder) => {
      entries.push(entryFromRecord(definition, record, sourceOrder, entryOrder));
    });

    return {
      source: definition.source,
      sourceLabel: definition.sourceLabel,
      state: "ready",
      error: undefined,
    };
  });

  return {
    entries: entries.sort(compareEntries).map(({ sourceOrder: _sourceOrder, entryOrder: _entryOrder, ...entry }) => entry),
    sources,
  };
}

export function filterTaskEntries(entries: AggregatedTask[], filter: TaskFilter) {
  if (filter === "全部") return entries;
  const expected: Record<Exclude<TaskFilter, "全部">, TaskStatus> = {
    "进行中": "running",
    "待确认": "attention",
    "已完成": "completed",
  };
  return entries.filter((entry) => entry.status === expected[filter]);
}

export function taskCounts(entries: AggregatedTask[]) {
  return entries.reduce((counts, entry) => {
    if (entry.status === "running") counts.running += 1;
    if (entry.status === "attention") counts.attention += 1;
    if (entry.status === "completed") counts.completed += 1;
    return counts;
  }, { running: 0, attention: 0, completed: 0 });
}

export function taskDestination(entry: AggregatedTask) {
  const id = encodeURIComponent(entry.id || '');
  if (entry.source === 'batch') return `/pages/batch/index?batchId=${id}`;
  if (entry.source === 'shuihuo') return `/pages/module-detail/index?key=shuihuo&projectId=${encodeURIComponent(entry.projectId || '')}`;
  if (entry.source === 'agent') return `/pages/module-detail/index?key=agent&taskId=${id}`;
  if (entry.source === 'fetch') return `/pages/module-detail/index?key=fetch&taskId=${id}`;
  return `/pages/module-detail/index?key=script&entryId=${id}`;
}
