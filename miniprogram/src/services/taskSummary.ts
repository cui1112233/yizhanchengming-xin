import {sessionRevision} from './sessionBoundary';
import {loadCmTaskFeed,loadNovelTaskFeed,loadCutTaskFeed,loadWaterTaskFeed,limitTaskRequests} from './taskFeeds';
export type TaskSource = "script" | "agent" | "fetch" | "batch" | "shuihuo" | "novel" | "aicut";
export type TaskStatus = "running" | "retry" | "attention" | "completed" | "unknown";
export type TaskFilter = "全部" | "进行中" | "需重试" | "已完成";

export type AggregatedTask = {
  source: TaskSource;
  projectId?: string;
  bookId?:string;
  sessionId?:string;
  timelineId?:string;
  planKind?:string;
  sourceLabel: string;
  id: string | undefined;
  title: string | undefined;
  createdAt: string | undefined;
  rawStatus: string | undefined;
  status: TaskStatus;
  statusDetail: string;
  progress: { completed: number; total: number } | undefined;
  visualHint: string | undefined;
  moduleKey: TaskSource;
  kind: string | undefined;
  stage: string | undefined;
  taskType: string | undefined;
  retryable: boolean;
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
  { source: "agent", sourceLabel: "CM画布", path: "/api/agent/tasks", collection: "tasks" },
  { source: "fetch", sourceLabel: "批量测书", path: "/api/batch-rewrite/tasks", collection: "tasks" },
  { source: "batch", sourceLabel: "批量工厂", path: "/api/batch-factory/v12/batches/summary", collection: "batches" },
  { source: "shuihuo", sourceLabel: "水货生产", path: "/api/shuihuo-production/projects", collection: "projects" },
  {source:"novel",sourceLabel:"小说近期写作",path:"/api/novel-creation/projects",collection:"projects"},
  {source:"aicut",sourceLabel:"AI-CUT导出",path:"/api/ai-cut/projects",collection:"projects"},
];

const COMPLETED_STATUSES = new Set(["done", "completed", "success", "succeeded", "succeeded_with_warnings", "finished", "ai_done", "已完成", "完成", "成功", "已合并"]);
const RETRY_STATUSES = new Set(["failed", "error", "canceled", "cancelled", "stopped", "waiting_retry", "original_failed", "ai_failed", "失败", "异常", "等待重试"]);
const ATTENTION_STATUSES = new Set(["waiting-review", "waiting_review", "pending_confirmation", "accepted_pending", "待确认", "待审核", "awaiting_confirmation", "paused", "save_pending"]);
const RUNNING_STATUSES = new Set(["queued", "pending", "running", "processing", "executing", "waiting", "scheduled_waiting", "ai_processing", "retrying", "进行中", "处理中", "排队中", "等待中", "writing", "revising", "planning", "normalizing", "auditing", "settling", "indexing", "not_started", "analyzing", "preparing", "submitting", "result_ready", "ingesting", "cancel_requested", "canceling"]);

const STATUS_DETAILS: Record<string, string> = {
  scheduled_waiting: "等待定时启动", input_ready: "分类信息已就绪", waiting_retry: "需重试",
  waiting_classification: "待确认分类", waiting_sensitive_ai: "等待敏感词处理",
  waiting_configuration: "待配置", waiting_login: "待登录", pending_submit: "待提交",
  upload_pending: "提交结果待确认", stopped: "已停止", queued: "排队中", pending: "排队中",
  processing: "进行中", running: "进行中", completed: "已完成", succeeded: "已完成",
  failed: "需重试", error: "需重试", awaiting_confirmation:"待确认", planning:"生成创作蓝图", writing:"写作正文", normalizing:"整理正文", auditing:"审核章节", revising:"修订章节", settling:"整理人物与伏笔", indexing:"建立记忆索引", not_started:"等待启动", paused:"已暂停", save_pending:"待恢复保存", cancel_requested:"正在取消",
};

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
  if (RETRY_STATUSES.has(status)) return "retry";
  if (ATTENTION_STATUSES.has(status)) return "attention";
  if (RUNNING_STATUSES.has(status)) return "running";
  return "unknown";
}

export function taskStatusDetail(value: unknown) {
  const status = textValue(value)?.toLowerCase();
  if (!status) return "状态未返回";
  if (STATUS_DETAILS[status]) return STATUS_DETAILS[status];
  const normalized = taskStatusFromServer(status);
  if (normalized === "running") return "进行中";
  if (normalized === "retry") return "需重试";
  if (normalized === "attention") return "待确认";
  if (normalized === "completed") return "已完成";
  return `状态：${status}`;
}

function numberValue(value: unknown) {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim() && Number.isFinite(Number(value))) return Number(value);
  return undefined;
}

function recordProgress(record: Record<string, unknown>) {
  const completed = numberValue(record.completed ?? record.progressCurrent ?? record.ready);
  const total = numberValue(record.total ?? record.progressTotal);
  return completed !== undefined && total !== undefined && total > 0 ? { completed, total } : undefined;
}

function historyVideoState(entry: Record<string, unknown>) {
  const videoTasks = entry.videoTasks;
  if (!isRecord(videoTasks)) return { status: "unknown" as TaskStatus, progress: undefined };

  const tasks = Object.values(videoTasks).filter(isRecord);
  if (!tasks.length) return { status: "unknown" as TaskStatus, progress: undefined };

  const statuses = tasks.map((task) => taskStatusFromServer(firstText(task, ["status"])));
  const completed = statuses.filter((status) => status === "completed").length;
  const progress = { completed, total: tasks.length };

  if (statuses.includes("retry")) return { status: "retry" as TaskStatus, progress };
  if (statuses.includes("running")) return { status: "running" as TaskStatus, progress };
  if (statuses.every((status) => status === "completed")) return { status: "completed" as TaskStatus, progress };
  return { status: "unknown" as TaskStatus, progress };
}

function entryFromRecord(definition: SourceDefinition, record: Record<string, unknown>, sourceOrder: number, entryOrder: number): InternalTask {
  const rawStatus = firstText(record, ["status", "state", "effectiveStatus", "effective_status", "aiStatus", "ai_status", "originalStatus", "original_status"]);
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
    ...(firstText(record,['projectId','batchId'])?{projectId:firstText(record,['projectId','batchId'])}:{}),
    ...(textValue(record.bookId)?{bookId:textValue(record.bookId)}:{}),
    ...(textValue(record.sessionId)?{sessionId:textValue(record.sessionId)}:{}),
    ...(textValue(record.timelineId)?{timelineId:textValue(record.timelineId)}:{}),
    ...(textValue(record.planKind)?{planKind:textValue(record.planKind)}:{}),
    id: firstText(record, ["id", "bookId", "book_id"]),
    title: firstText(record, titleKeys) || (definition.source === "agent" ? firstText(request, ["instruction"]) : undefined),
    createdAt: firstText(record, ["updatedAt", "updated_at", "createdAt", "created_at"]),
    rawStatus,
    status,
    statusDetail: definition.source==='batch'&&record.kind==='automation'&&taskStatusFromServer(rawStatus)==='completed'?'流程已结束':taskStatusDetail(rawStatus),
    progress: video.progress || recordProgress(record),
    visualHint: firstText(record, ["kind", "type", "stage", "taskType", "task_type"]),
    moduleKey: definition.source,
    kind: firstText(record, ["kind", "type"]),
    stage: firstText(record, ["stage", "progressPhase"]),
    taskType: firstText(record, ["taskType", "task_type"]),
    retryable: record.retryable === true || record.canRetry === true || record.can_retry === true || (definition.source === "fetch" && status === "retry"),
    sourceOrder,
    entryOrder,
  };
}

function batchRuntimeRecords(batch: Record<string, unknown>, runtime: unknown) {
  if (!isRecord(runtime)) throw new Error("批量工厂运行状态响应不完整");
  const batchID = firstText(batch, ["id"]);
  const batchTitle = firstText(batch, ["title", "name"]) || "批量任务";
  const rows: Record<string, unknown>[] = [];
  const push = (record: Record<string, unknown>, fallback: Record<string, unknown>) => rows.push({ ...fallback, ...record, moduleKey: "batch", batchId: batchID, projectId:batchID });

  const automation = isRecord(runtime.automation) ? runtime.automation : undefined;
  const automationState = automation && firstText(automation, ["state", "status"]);
  const bookRows=Array.isArray(automation?.books)?automation!.books.filter(isRecord):isRecord(automation?.books)?Object.values(automation!.books).filter(isRecord):[];
  const bookTitle=(bookId:unknown)=>{const book=bookRows.find(b=>firstText(b,['bookId','id'])===textValue(bookId));return book?firstText(book,['title','name','bookTitle'])||batchTitle:batchTitle;};
  const stageNames:Record<string,string>={assets:'人物场景',director:'剧本分镜',opening:'开篇',visual:'视觉风格',comic:'漫画分镜',image:'漫画图片',video:'视频生成',merge:'成片合成'};

  if (automation && automationState && automationState !== "idle") {
    const counts = isRecord(automation.counts) ? automation.counts : {};
    push(automation, {
      id: firstText(automation, ["id"]) || `${batchID || "batch"}-automation`, title: `${batchTitle} · 自动生产`,
      kind: "automation", stage: "automation", status: automationState,
      completed: numberValue(counts.ready), total: numberValue(counts.total), updatedAt: firstText(batch, ["updatedAt", "updated_at"]),
    });
  }

  const production = isRecord(runtime.production) ? runtime.production : {};
  const productionJobs = Array.isArray(production.jobs) ? production.jobs.filter(isRecord) : [];
  productionJobs.forEach((job) => {
    const tasks = Array.isArray(job.tasks) ? job.tasks.filter(isRecord) : [];
    tasks.forEach((task) => push(task, {
      title: `${bookTitle(job.bookId)} · 视频生成`, kind: "video", stage: "video", bookId:textValue(job.bookId),
      status: firstText(task, ["status"]) || firstText(job, ["status"]),
      createdAt: firstText(task, ["updatedAt", "createdAt"]) || firstText(job, ["updatedAt", "createdAt"]),
    }));
  });

  const merge = isRecord(runtime.merge) ? runtime.merge : {};
  const mergeJobs = Array.isArray(merge.jobs) ? merge.jobs.filter(isRecord) : [];
  mergeJobs.forEach((job) => push(job, { title: `${bookTitle(job.bookId)} · 成片合成`, kind: "export", stage: "merge" }));

  const stageSummaries = isRecord(runtime.stageSummaries) ? runtime.stageSummaries : {};
  Object.values(stageSummaries).filter(isRecord).forEach((summary) => {
    const runs = Array.isArray(summary.runs) ? summary.runs.filter(isRecord) : [];
    runs.forEach((run) => {
      const stage = firstText(run, ["stage"]) || "stage";
      push(run, { title: `${bookTitle(run.bookId||summary.bookId)} · ${stageNames[stage]||stage}`, bookId:run.bookId||summary.bookId,kind: stage, stage });
    });
  });
  return rows;
}

function batchListRecord(batch: Record<string, unknown>) {
  return {
    id: firstText(batch, ["id"]),
    projectId:firstText(batch,["id"]),
    title: firstText(batch, ["title", "name"]),
    createdAt: firstText(batch, ["updatedAt", "updated_at", "createdAt", "created_at"]),
    status: firstText(batch, ["status", "state"]),
    kind: firstText(batch, ["productionMode", "production_mode"]) || "batch",
    stage: "batch",
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

export async function loadTaskSummary(request: TaskRequest, options: { projectsOnly?: boolean; sources?: TaskSource[]; shouldContinue?: () => boolean; onProgress?: (summary:TaskSummary)=>void } = {}): Promise<TaskSummary> {
  const accountRevision=sessionRevision();
  const assertCurrent=()=>{if(accountRevision!==sessionRevision()||options.shouldContinue?.()===false)throw Error('页面已离开或账号已切换，已停止后续同步');};
  const originalRequest=request;
  // Check inside the limiter: queued reads must not start after the user leaves.
  request=limitTaskRequests(async path=>{assertCurrent();const result=await originalRequest(path);assertCurrent();return result;});
  const definitions=TASK_SOURCE_DEFINITIONS.filter(definition=>!options.sources||options.sources.includes(definition.source));
  const partial:Array<PromiseSettledResult<unknown>|undefined>=definitions.map(()=>undefined);
  const publish=(index:number,result:PromiseSettledResult<unknown>)=>{partial[index]=result;if(accountRevision===sessionRevision()&&options.shouldContinue?.()!==false)options.onProgress?.(summarize(partial));};
  const settled = await Promise.allSettled(definitions.map(async (definition,index) => {
    try {const result=await readSource(definition);publish(index,{status:'fulfilled',value:result});return result;}
    catch(reason){publish(index,{status:'rejected',reason});throw reason;}
  }));
  assertCurrent();
  return summarize(settled);

  async function readSource(definition:SourceDefinition){
    const payload = await request(definition.path);
    if (!isRecord(payload) || !Array.isArray(payload[definition.collection])) throw new Error("服务端任务列表响应不完整");

    if(['agent','novel','aicut'].includes(definition.source)){
      const projects=collectionFrom(payload,definition.collection);
      if(options.projectsOnly)return{[definition.collection]:projects.map(p=>({...p,projectId:textValue(p.id),taskType:'project'}))};
      const feed=definition.source==='agent'?await loadCmTaskFeed(request,projects):definition.source==='novel'?await loadNovelTaskFeed(request,projects):await loadCutTaskFeed(request,projects);
      return{[definition.collection]:feed.rows,warning:feed.warning};
    }
    if (definition.source === "batch") {
      const batches = collectionFrom(payload, "batches");
      const runtimeRows = await Promise.all(batches.map(async (batch) => {
        const id = firstText(batch, ["id"]); if (!id) throw new Error("批量工厂批次缺少 ID");
        try {
          const runtime = await request(`/api/batch-factory/v12/batches/${encodeURIComponent(id)}/runtime-summary`);
          const rows = batchRuntimeRecords(batch, runtime);
          return rows.length ? rows : [batchListRecord(batch)];
        } catch {
          // The summary list is already durable server truth. Runtime summary is
          // a heavier optional enrichment and one slow batch must not erase all
          // of the user's existing task records.
          return [batchListRecord(batch)];
        }
      }));
      return { batches: runtimeRows.flat() };
    }
    if (definition.source !== "shuihuo") return payload;
    if (options.projectsOnly) return { projects: collectionFrom(payload, "projects").map(project => ({ ...project, projectId: textValue(project.id) })) };
    const feed=await loadWaterTaskFeed(request,collectionFrom(payload,'projects'));
    return{projects:feed.rows,warning:feed.warning};
  }
  function summarize(results:Array<PromiseSettledResult<unknown>|undefined>):TaskSummary{
  const entries: InternalTask[] = [];
  const sources = results.flatMap((result, sourceOrder): TaskSourceResult[] => {
    if(!result)return [];
    const definition = definitions[sourceOrder];
    if (result.status === "rejected") {
      return [{
        source: definition.source,
        sourceLabel: definition.sourceLabel,
        state: "error",
        error: errorMessage(result.reason),
      }];
    }

    collectionFrom(result.value, definition.collection).forEach((record, entryOrder) => {
      entries.push(entryFromRecord(definition, record, sourceOrder, entryOrder));
    });

    const warning=(result.value as Record<string,unknown>).warning;
    return [{
      source: definition.source,
      sourceLabel: definition.sourceLabel,
      state: warning ? "error" : "ready",
      error: typeof warning==='string'&&warning?`部分任务未同步：${warning}`:undefined,
    }];
  });

  return {
    entries: entries.sort(compareEntries).map(({ sourceOrder: _sourceOrder, entryOrder: _entryOrder, ...entry }) => entry),
    sources,
  };
  }
}

export function filterTaskEntries(entries: AggregatedTask[], filter: TaskFilter) {
  if (filter === "全部") return entries;
  const expected: Record<Exclude<TaskFilter, "全部">, TaskStatus> = {
    "进行中": "running",
    "需重试": "retry",
    "已完成": "completed",
  };
  return entries.filter((entry) => entry.status === expected[filter]);
}

export function taskCounts(entries: AggregatedTask[]) {
  return entries.reduce((counts, entry) => {
    if (entry.status === "running") counts.running += 1;
    if (entry.status === "retry") counts.retry += 1;
    if (entry.status === "completed") counts.completed += 1;
    return counts;
  }, { running: 0, retry: 0, completed: 0 });
}

export function taskDestination(entry: AggregatedTask) {
  const id = encodeURIComponent(entry.id || '');
  if (entry.source === 'batch') return `/pages/batch/index?batchId=${encodeURIComponent(entry.projectId||entry.id||'')}${entry.bookId?`&bookId=${encodeURIComponent(entry.bookId)}`:''}`;
  if (entry.source === 'shuihuo') return `/pages/module-detail/index?key=shuihuo&projectId=${encodeURIComponent(entry.projectId || '')}`;
  if (entry.source === 'agent') return `/pages/module-detail/index?key=agent&taskId=${encodeURIComponent(entry.projectId||entry.id||'')}&tab=tasks`;
  if(entry.source==='novel'&&entry.taskType==='project')return `/comic/pages/novel/index?projectId=${encodeURIComponent(entry.projectId||entry.id||'')}`;
  if(entry.source==='aicut'&&entry.taskType==='project')return `/comic/pages/aicut/index?projectId=${encodeURIComponent(entry.projectId||entry.id||'')}`;
  if(entry.source==='novel'&&entry.taskType==='writing-plan')return `/comic/pages/novel/index?projectId=${encodeURIComponent(entry.projectId||'')}&sessionId=${encodeURIComponent(entry.sessionId||'')}&tab=automation&planId=${id}&planKind=${entry.planKind==='batches'?'batches':'schedules'}`;
  if(entry.source==='novel')return `/comic/pages/novel/index?projectId=${encodeURIComponent(entry.projectId||'')}&sessionId=${encodeURIComponent(entry.sessionId||'')}&runId=${id}`;
  if(entry.source==='aicut')return `/comic/pages/aicut/index?projectId=${encodeURIComponent(entry.projectId||'')}&timelineId=${encodeURIComponent(entry.timelineId||'default')}&tab=exports`;
  if (entry.source === 'fetch') return `/pages/module-detail/index?key=booktest&taskId=${id}`;
  return `/pages/module-detail/index?key=script&historyId=${id}`;
}
