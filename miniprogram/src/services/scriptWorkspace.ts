export type ScriptEntity = { id?: string; name?: string; description?: string; data?: unknown; imageUrls?: string[]; mainImageUrl?: string };
export type ScriptVideoTask = { taskId?: string; status?: string; videoUrl?: string; error?: string; [key: string]: unknown };
export type ScriptConstraintLayer = { enabled?: boolean; source?: string; presetId?: string; personalPromptId?: string; body?: string; smartUnifiedStyle?: string };
export type ScriptConstraints = { enabled?: boolean; baseSetup?: ScriptConstraintLayer; prefix?: ScriptConstraintLayer; quality?: ScriptConstraintLayer; restriction?: ScriptConstraintLayer; negative?: ScriptConstraintLayer; entityReferences?: unknown[] };

type LayerKey = "baseSetup" | "prefix" | "quality" | "restriction" | "negative";
export function scriptConstraintChoices<T extends { constraintCategory?: string }>(catalog: T[], category: LayerKey) {
  return catalog.filter((item) => item.constraintCategory === category);
}
export function updateScriptConstraint(current: ScriptConstraints, key: LayerKey, patch: ScriptConstraintLayer): ScriptConstraints {
  const next = { ...current, [key]: { ...current[key], ...patch } };
  next.enabled = (["baseSetup", "prefix", "quality", "restriction", "negative"] as const).some((layer) => next[layer]?.enabled);
  return next;
}

export function scriptEntityKey(entity: ScriptEntity, index: number) {
  const data = entity.data as Record<string, unknown> | undefined;
  return String(entity.id || data?.name || data?.角色名称 || data?.名称 || entity.name || index);
}

function focusedCharacters(state: ScriptWorkspaceState) {
  return state.characters.filter((entity, index) => state.protagonists?.includes(scriptEntityKey(entity, index)));
}

export type ScriptWorkspaceState = {
  novelText: string;
  characters: ScriptEntity[];
  scenes: ScriptEntity[];
  opening: string;
  format: string;
  duration: string;
  matchAudio: boolean;
  audioTotalSeconds?: number | null;
  textModelId?: string;
  imageModelId?: string;
  extractionPreset?: string;
  visualStyle?: string;
  protagonists?: string[];
  constraints?: ScriptConstraints;
  videoTasks?: Record<string, ScriptVideoTask>;
};

const openingValues: Record<string, string> = { 连续开头: "continuous", 爆款开头: "hook", 分段开头: "segmented" };
const formatValues: Record<string, string> = { 画布: "storyboard", 剧本: "shortdrama", 剧情: "screenplay", 分镜: "shotlist", Q版: "q版" };
const durationValues: Record<string, string> = { "10 秒": "10s", "15 秒": "15s" };

function normalizedOptions(state: ScriptWorkspaceState) {
  return {
    mode: openingValues[state.opening] || state.opening,
    format: formatValues[state.format] || state.format,
    duration: durationValues[state.duration] || state.duration,
  };
}

export function buildExtractRequest(state: ScriptWorkspaceState) {
  return {
    promptType: "extract",
    novelText: state.novelText.trim(),
    extractionPreset: state.extractionPreset || undefined,
    ...(state.textModelId ? { textModelId: state.textModelId } : {}),
    max_tokens: 4096,
    temperature: 0.3,
    stream: false,
  };
}

export function buildGenerateRequest(state: ScriptWorkspaceState) {
  const options = normalizedOptions(state);
  return {
    promptType: "script",
    ...options,
    novelText: state.novelText.trim(),
    characters: state.characters,
    scenes: state.scenes,
    visualStyle: state.visualStyle || "",
    protagonists: focusedCharacters(state),
    constraints: state.constraints || {},
    matchAudio: state.matchAudio,
    audioTotalSeconds: state.matchAudio ? state.audioTotalSeconds ?? null : null,
    ...(state.textModelId ? { textModelId: state.textModelId } : {}),
    ...(state.imageModelId ? { imageModelId: state.imageModelId } : {}),
    max_tokens: options.format === "shotlist" ? 16000 : 8192,
    temperature: 0.7,
    stream: false,
  };
}

export function buildHistoryEntry(state: ScriptWorkspaceState, output: string) {
  return {
    id: `mini-${Date.now()}-${Math.random().toString(36).slice(2,9)}`,
    ...normalizedOptions(state),
    output: output.trim(),
    novelText: state.novelText.trim(),
    extractInfo: { characters: state.characters, scenes: state.scenes },
    matchAudio: state.matchAudio,
    audioTotalSeconds: state.matchAudio ? state.audioTotalSeconds ?? null : null,
    videoTasks: {},
    textModelId: state.textModelId || "",
    imageModelId: state.imageModelId || "",
    extractionPreset: state.extractionPreset || "",
    visualStyle: state.visualStyle || "",
    protagonists: focusedCharacters(state),
    constraints: state.constraints || {},
  };
}

export function getScriptStepState(state: Pick<ScriptWorkspaceState, "novelText" | "characters" | "scenes">) {
  return {
    source: state.novelText.trim() ? "done" : "active",
    entities: state.characters.length || state.scenes.length ? "done" : state.novelText.trim() ? "active" : "idle",
    result: "idle",
  } as const;
}

export function normalizeScriptEntities(payload: any) {
  let source = payload?.extractInfo || payload?.data || payload || {};
  if (![source.characters, source.people, source.scenes, source.人物设定, source.人物, source.角色设定, source.场景设定, source.场景].some(Array.isArray)) {
    const content = parseScriptOutput(payload).replace(/^\s*```(?:json)?\s*/i, "").replace(/\s*```\s*$/, "").trim();
    const start = content.indexOf("{");
    const end = content.lastIndexOf("}");
    if (start < 0 || end < start) throw new Error("人物场景提取未返回有效 JSON");
    try { source = JSON.parse(content.slice(start, end + 1)); }
    catch { throw new Error("人物场景提取格式无效，请检查提取模板"); }
  }
  const characters = [source.characters, source.people, source.人物设定, source.人物, source.角色设定].find(Array.isArray) || [];
  const scenes = [source.scenes, source.场景设定, source.场景, source.scene_options].find(Array.isArray) || [];
  return { characters, scenes };
}

export function parseScriptOutput(payload: any): string {
  const source = payload?.data || payload?.result || payload?.response || payload;
  const value = typeof source === "string" ? source : source?.choices?.[0]?.message?.content ?? source?.message?.content ?? source?.output_text ?? source?.output ?? source?.text ?? source?.content;
  if (typeof value !== "string" || !value.trim()) throw new Error("模型未返回可用的剧本内容");
  return value.trim();
}

export function restoreScriptHistory(payload: any): ScriptWorkspaceState & { output: string } {
  const entry = payload?.entry || payload;
  if (!entry || typeof entry !== "object" || typeof entry.novelText !== "string") throw new Error("服务端历史未返回可恢复的原文");
  const extraction = normalizeScriptEntities(entry.extractInfo || { characters: [], scenes: [] });
  const label = (values: Record<string, string>, value: any, fallback: string) => Object.keys(values).find((key) => values[key] === value) || value || fallback;
  return {
    novelText: entry.novelText, ...extraction,
    opening: label(openingValues, entry.mode, "分段开头"),
    format: label(formatValues, entry.format, "分镜"), duration: label(durationValues, entry.duration, "15 秒"),
    matchAudio: entry.matchAudio === true, audioTotalSeconds: entry.audioTotalSeconds ?? null,
    textModelId: entry.textModelId || "", imageModelId: entry.imageModelId || "", extractionPreset: entry.extractionPreset || "",
    visualStyle: entry.visualStyle || "", protagonists: Array.isArray(entry.protagonists) ? entry.protagonists.map((item: string | ScriptEntity) => typeof item === "string" ? item : scriptEntityKey(item, extraction.characters.findIndex((entity: ScriptEntity) => JSON.stringify(entity) === JSON.stringify(item)))) : [],
    constraints: entry.constraints || {}, output: typeof entry.output === "string" ? entry.output : "",
    videoTasks: entry.videoTasks && typeof entry.videoTasks === "object" && !Array.isArray(entry.videoTasks) ? entry.videoTasks : {},
  };
}
