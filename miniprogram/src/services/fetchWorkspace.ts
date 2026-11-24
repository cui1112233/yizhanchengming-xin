export type FetchPlatform = { id: string; name: string };
export type FetchTextModel = { id: string; name: string };

export type FetchTask = {
  bookId: string;
  bookName: string;
  platformId: string;
  platformName: string;
  status: string;
  originalStatus: string;
  aiStatus: string;
  aiGeneratedCount: number;
  selectedVersions: string[];
  gender: string;
  style: string;
  siteSubmitStatus: string;
  classifyStatus: string;
  createdAt: string;
  updatedAt: string;
  originalError: string;
  aiError: string;
  siteSubmitDoneVersions: string[];
  siteSubmitFailedVersions: string[];
};

export type FetchTaskStage = {
  key: "classify" | "original" | "ai" | "submit";
  label: "分类" | "原文" | "AI" | "提交";
  state: "pending" | "active" | "done" | "failed" | "skipped";
};

export type FetchWorkspaceState = {
  platforms: FetchPlatform[];
  platformId: string;
  models: FetchTextModel[];
  textModelId: string;
  maxTxt: number;
  selectedVersions: string[];
  aiSlotMethods: Record<string, string>;
  sensitiveAiEnabled: boolean;
};

export type FetchProcessInput = {
  input: string;
  platformId: string;
  maxTxt: number;
  selectedVersions: string[];
  aiSlotMethods: Record<string, string>;
  sensitiveAiEnabled: boolean;
};

type RecordValue = Record<string, unknown>;

function record(value: unknown): RecordValue {
  return value && typeof value === "object" && !Array.isArray(value) ? value as RecordValue : {};
}

function text(value: unknown) {
  return typeof value === "string" ? value.trim() : value === undefined || value === null ? "" : String(value);
}

function platformLabel(value:unknown){
  const name=text(value);
  return /^(?:\\?u[0-9a-f]{4}){2,}$/i.test(name)?name.replace(/\\?u([0-9a-f]{4})/gi,(_,code)=>String.fromCharCode(parseInt(code,16))):name;
}

function positiveNumber(value: unknown, fallback: number) {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

function nonNegativeNumber(value: unknown) {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : 0;
}

function stringRecord(value: unknown): Record<string, string> {
  return Object.fromEntries(Object.entries(record(value)).flatMap(([key, item]) => {
    const normalized = text(item);
    return normalized ? [[key, normalized]] : [];
  }));
}

function stringList(value: unknown) {
  return Array.isArray(value) ? value.map(text).filter(Boolean) : [];
}

export function normalizeFetchTask(value: unknown): FetchTask {
  const task = record(value);
  return {
    bookId: text(task.bookId || task.book_id || task.id),
    bookName: text(task.bookName || task.book_name),
    platformId: text(task.platformId || task.platform_id),
    platformName: text(task.platformName || task.platform_name),
    status: text(task.status),
    originalStatus: text(task.originalStatus || task.original_status),
    aiStatus: text(task.aiStatus || task.ai_status),
    aiGeneratedCount: nonNegativeNumber(task.aiGeneratedCount ?? task.ai_generated_count ?? (Array.isArray(task.ai_files) ? task.ai_files.length : 0)),
    selectedVersions: stringList(task.selectedVersions || task.selected_versions || task.target_versions),
    gender: text(task.gender),
    style: text(task.style),
    siteSubmitStatus: text(task.siteSubmitStatus || task.site_submit_status),
    classifyStatus: text(task.classifyStatus || task.classify_status),
    createdAt: text(task.createdAt || task.created_at || task.created),
    updatedAt: text(task.updatedAt || task.updated_at || task.updated),
    originalError: text(task.originalError || task.original_error || task.originalErrorMessage || task.error),
    aiError: text(task.aiError || task.ai_error),
    siteSubmitDoneVersions: stringList(task.siteSubmitDoneVersions || task.site_submit_done_versions),
    siteSubmitFailedVersions: stringList(task.siteSubmitFailedVersions || task.site_submit_failed_versions),
  };
}

export function normalizeFetchTasks(value: unknown) {
  return (Array.isArray(value) ? value : []).map(normalizeFetchTask).filter((task) => Boolean(task.bookId));
}

export function taskDateKey(task: Pick<FetchTask, "createdAt" | "updatedAt">) {
  const date = new Date(task.createdAt || task.updatedAt);
  if (!Number.isFinite(date.getTime())) return "";
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return local.toISOString().slice(0, 10);
}

function stageState(value: string): FetchTaskStage["state"] {
  const status = value.toLowerCase();
  if (/(failed|error|失败)/.test(status)) return "failed";
  if (/(done|success|submitted|complete|completed|classified)/.test(status)) return "done";
  if (/(running|processing|pending|accepted|queue|处理)/.test(status)) return "active";
  return "pending";
}

export function fetchTaskStages(task: FetchTask): FetchTaskStage[] {
  const classify = stageState(task.classifyStatus) !== "pending"
    ? stageState(task.classifyStatus)
    : task.gender || task.style ? "done" : "pending";
  const original = stageState(task.originalStatus);
  const expectedAi = task.selectedVersions.filter((version) => /^ai[1-5]$/i.test(version)).length;
  let ai = stageState(task.aiStatus);
  if (ai === "pending" && expectedAi > 0 && task.aiGeneratedCount >= expectedAi) ai = "done";
  else if (ai === "pending" && task.aiGeneratedCount > 0) ai = "active";
  else if (expectedAi === 0) ai = "skipped";

  let submit = stageState(task.siteSubmitStatus);
  if (submit === "pending" && task.siteSubmitDoneVersions.length) submit = "done";
  if (task.siteSubmitFailedVersions.length) submit = "failed";

  return [
    { key: "classify", label: "分类", state: classify },
    { key: "original", label: "原文", state: original },
    { key: "ai", label: "AI", state: ai },
    { key: "submit", label: "提交", state: submit },
  ];
}

export function fetchTaskOverallStatus(task: FetchTask) {
  if (['running','processing'].includes(task.status.toLowerCase())) return '正在执行中…';
  if (task.status === 'input_ready') return '分类信息已就绪';
  const stages = fetchTaskStages(task);
  if (stages.some((stage) => stage.state === "failed")) return "失败";
  if (stages.find((stage) => stage.key === "submit")?.state === "done") return "已提交";
  if (stages.find((stage) => stage.key === "ai")?.state === "done") return "AI完成";
  if (stages.find((stage) => stage.key === "original")?.state === "done") return "原文完成";
  if (stages.some((stage) => stage.state === "active")) return "处理中";
  return task.status || "等待中";
}

export function fetchWorkspaceState(configPayload: unknown, modelPayload: unknown): FetchWorkspaceState {
  const config = record(configPayload);
  const appConfig = record(config.app_config);
  const fetchConfig = record(appConfig.fetch);
  const workForm = record(config.work_form || appConfig.work_form);
  const platforms = (Array.isArray(config.platforms) ? config.platforms : []).flatMap((item) => {
    const platform = record(item);
    if (platform.visible === false) return [];
    const id = text(platform.id);
    return id ? [{ id, name: platformLabel(platform.name) || id }] : [];
  });
  const models = (Array.isArray(modelPayload) ? modelPayload : []).flatMap((item) => {
    const model = record(item);
    if (model.enabled === false) return [];
    const id = text(model.id || model.modelId);
    return id ? [{ id, name: text(model.displayName || model.name || model.modelId || model.id) || id }] : [];
  });
  const configuredModelId = text(appConfig.text_model_id || appConfig.textModelId);
  const selectedVersions = Array.isArray(workForm.selected_versions) && workForm.selected_versions.length
    ? workForm.selected_versions.map(text).filter(Boolean)
    : ["original", "ai1"];

  return {
    platforms,
    platformId: platforms[0]?.id || "",
    models,
    textModelId: models.some((model) => model.id === configuredModelId) ? configuredModelId : models[0]?.id || "",
    maxTxt: positiveNumber(fetchConfig.default_max_txt, 4000),
    selectedVersions,
    aiSlotMethods: { ...stringRecord(record(appConfig.rewrite).ai_slot_methods), ...stringRecord(workForm.ai_slot_methods) },
    sensitiveAiEnabled: (workForm.sensitive_ai_enabled ?? record(appConfig.sensitive_ai).enabled) === true,
  };
}

export function buildFetchProcessPayload(input: FetchProcessInput) {
  if (!input.platformId.trim()) throw new Error("请先选择服务端提供的书城平台");
  return {
    input_text: input.input.trim(),
    platform_id: input.platformId,
    max_txt: positiveNumber(input.maxTxt, 4000),
    selected_versions: input.selectedVersions.length ? input.selectedVersions : ["original", "ai1"],
    ai_slot_methods: input.aiSlotMethods,
    sensitive_ai_enabled: input.sensitiveAiEnabled,
  };
}
