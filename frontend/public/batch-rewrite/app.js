document.documentElement.classList.add("qiantie-novel-fetch-hydrating");

const state = {
  config: null,
  tasks: [],
  allTasks: [],
  selectedId: "",
  selectedIds: new Set(),
  sensitiveProcessingIds: new Set(),
  aiProcessingIds: new Set(),
  pendingRuleSuggestions: null,
  pendingOpeningItem: null,
  activeProcessJobId: "",
  taskDate: "",
  currentBatchIds: new Set(),
  currentBatchDate: "",
  textModels: [],
  viewMode: "current",
  modalBusy: new Set(),
  detailRefreshTimer: null,
  detailViewKind: "",
};

const DEFAULT_COLUMN_ORDER = "书籍ID,书名,推荐理由,男女频,标签,评级";
const WORK_FORM_STORAGE_KEY = "batchRewrite.workForm.v1";
const LEGACY_WORK_INPUT_SAMPLE_PREFIX = "7674515088685943832\t";
const API_ROOT = "/api/batch-rewrite";
const PLATFORM_API_TIMEOUT_MS = 15000;
const PER_BOOK_MATERIAL_LIMIT = 8;
const PROCESS_JOB_TIMEOUT_MS = 10 * 60 * 1000;
const REWRITE_METHOD_OPTIONS = [
  { id: "", name: "自动轮换" },
  { id: "high_imitation", name: "高仿文章库" },
  { id: "opening_instruction", name: "爆款开头词库" },
  { id: "instruction", name: "批量改文指令库" },
];
let workFormSaveTimer = null;
let textModelPersistPromise = null;

const $ = (id) => document.getElementById(id);

document.documentElement.dataset.theme = new URLSearchParams(window.location.search).get('theme') === 'light' ? 'light' : 'dark';
window.addEventListener('message', event => {
  if (event.origin !== window.location.origin || event.data?.type !== 'qiantie-theme-sync') return;
  document.documentElement.dataset.theme = event.data.theme === 'light' ? 'light' : 'dark';
});

function pretty(value) {
  return JSON.stringify(value, null, 2);
}

function clone(value) {
  return JSON.parse(JSON.stringify(value || {}));
}

function mergeConfigResponse(nextConfig) {
  const knowledgeLoaded = state.config?.knowledge_loaded === true || nextConfig?.knowledge_loaded === true;
  state.config = { ...(nextConfig || {}), knowledge_loaded: knowledgeLoaded };
  if (knowledgeLoaded) renderKnowledgeSummary(state.config.knowledge_summary || {});
  return state.config;
}

function numberValue(id, fallback) {
  const value = Number($(id).value);
  return Number.isFinite(value) ? value : fallback;
}

function materialJieyaValue(value) {
  const count = Math.round(Number(value));
  return Number.isFinite(count) ? Math.max(0, Math.min(PER_BOOK_MATERIAL_LIMIT, count)) : 4;
}

function parseJsonInput(id, fallback) {
  const text = String($(id)?.value || '').trim();
  if (!text) return fallback;
  try { return JSON.parse(text); } catch (_) { throw new Error(`${id === 'webProfilesJson' ? '上传配置档' : '版本绑定'} 不是有效 JSON`); }
}

function syncGunpingMaterialCount() {
  const jieya = materialJieyaValue($("webJieyaNum")?.value);
  if ($("webJieyaNum")) $("webJieyaNum").value = jieya;
  if ($("webGunpingNum")) $("webGunpingNum").value = PER_BOOK_MATERIAL_LIMIT - jieya;
}

async function api(path, options = {}) {
  if (window.QiantieNovelFetchRuntime && !options.bypassRuntime) {
    return window.QiantieNovelFetchRuntime.api(path, options);
  }
  const legacyPath = String(path || "").replace(/^\/api/, "");
  const headers = { "Content-Type": "application/json", ...(options.headers || {}) };
  const token = localStorage.getItem("auth_token") || "";
  if (token) headers.Authorization = `Bearer ${token}`;
  let response;
  try {
    const requestUrl = /^\/api\/models(?:\?|$)/.test(String(path || ""))
      ? String(path)
      : `${API_ROOT}${legacyPath}`;
    response = await fetch(requestUrl, { headers, ...options });
  } catch (error) {
    reportBatchIssue("network", path, error.message || "网络请求失败");
    throw error;
  }
  const text = await response.text();
  let data = {};
  try {
    data = text ? JSON.parse(text) : {};
  } catch {
    data = { raw: text };
  }
  if (!response.ok) {
    // 兼容本工作台接口使用的 { error } 结构，避免把可操作的原因吞成“HTTP 400”。
    const error = new Error(normalizeBatchErrorMessage(data.error || data.detail || data.message || data.raw, response.status));
    error.status = response.status;
    reportBatchIssue("response", path, error.message, response.status, options.method || "GET");
    throw error;
  }
  return data;
}

function normalizeBatchErrorMessage(value, status) {
  const source = String(value || "");
  const raw = source.replace(/<!--[\s\S]*?-->/g, " ").replace(/<[^>]*>/g, " ").replace(/&lt;|&gt;|&amp;/g, " ").replace(/\s+/g, " ").trim();
  const match = source.match(/\b(502|503|504)\b/);
  const code = Number(status) || (match ? Number(match[1]) : 0);
  if (code === 502) return "网关暂时不可用（502），任务已保留，请稍后重试";
  if (code === 503) return "服务暂时不可用（503），任务已保留，请稍后重试";
  if (code === 504) return "服务响应超时（504），任务已保留，请稍后重试";
  if (!raw || /^(bad gateway|service unavailable|gateway timeout)$/i.test(raw)) return `请求失败${code ? `（${code}）` : ""}，任务已保留，请稍后重试`;
  return raw.slice(0, 4000);
}

// External 121 checks must always settle so the UI cannot remain in a loading state.
async function novelFetchPlatformApi(path, options = {}) {
  const headers = { "Content-Type": "application/json", ...(options.headers || {}) };
  const token = localStorage.getItem("auth_token") || "";
  if (token) headers.Authorization = `Bearer ${token}`;
  const controller = new AbortController();
  const timeoutId = window.setTimeout(() => controller.abort(), PLATFORM_API_TIMEOUT_MS);
  try {
    let response;
    try {
      response = await fetch(path, { ...options, headers, signal: controller.signal });
    } catch (error) {
      const message = error?.name === "AbortError" ? "验证请求超时，请检查网络后重试" : (error?.message || "网络请求失败");
      reportBatchIssue("network", path, message, undefined, options.method || "GET");
      throw new Error(message);
    }
    const text = await response.text();
    let data = {};
    try { data = text ? JSON.parse(text) : {}; } catch (_) { data = {}; }
    if (!response.ok) {
      const message = data.error || `HTTP ${response.status}`;
      reportBatchIssue("response", path, message, response.status, options.method || "GET");
      throw new Error(message);
    }
    return data;
  } finally {
    window.clearTimeout(timeoutId);
  }
}

function reportBatchIssue(kind, path, message, status, method) {
  const section = String(path || "").includes("web-submit") ? "submit"
    : String(path || "").includes("rules") ? "rules"
      : String(path || "").includes("process") ? "processing" : "api";
  const token = localStorage.getItem("auth_token") || "";
  const payload = JSON.stringify({
    kind: `batch-rewrite.${section}.${kind}`,
    message: String(message || "小说获取操作失败").slice(0, 4000),
    source: "小说获取",
    path: `${API_ROOT}${String(path || "")}`,
    method: method || "GET",
    status: Number.isInteger(status) ? status : undefined
  });
  void fetch("/api/client-errors", { method: "POST", headers: { "Content-Type": "application/json", ...(token ? { Authorization: `Bearer ${token}` } : {}) }, body: payload, keepalive: true }).catch(() => {});
}

function activateTab(name) {
  document.querySelectorAll(".tab").forEach((button) => {
    button.classList.toggle("active", button.dataset.tab === name);
  });
  document.querySelectorAll(".panel").forEach((panel) => {
    panel.classList.toggle("active", panel.id === name);
  });
}

function visiblePlatforms() {
  const items = state.config?.platforms || [];
  return items.filter((item) => item.visible !== false);
}

function renderOptions(select, items, valueKey = "id", labelKey = "name") {
  select.innerHTML = "";
  for (const item of items) {
    const option = document.createElement("option");
    option.value = item[valueKey];
    option.textContent = item[labelKey];
    select.appendChild(option);
  }
}

function collectWorkFormState() {
  return {
    platform_id: $("platformSelect")?.value || "",
    input_text: $("inputText")?.value || "",
    selected_versions: selectedProcessVersions(),
    ai_slot_methods: processAiMethods(),
    sensitive_ai_enabled: sensitiveAiProcessEnabled(),
    updated_at: new Date().toISOString(),
  };
}

function saveWorkFormState() {
  const formState = collectWorkFormState();
  try {
    localStorage.setItem(WORK_FORM_STORAGE_KEY, JSON.stringify(formState));
  } catch {
    // localStorage may be unavailable in some locked-down browsers.
  }
  clearTimeout(workFormSaveTimer);
  workFormSaveTimer = setTimeout(() => {
    api("/api/work-form", {
      method: "POST",
      body: JSON.stringify({ state: formState }),
    }).catch(() => {});
  }, 800);
}

async function saveWorkFormStateNow() {
  clearTimeout(workFormSaveTimer);
  workFormSaveTimer = null;
  const formState = collectWorkFormState();
  try {
    localStorage.setItem(WORK_FORM_STORAGE_KEY, JSON.stringify(formState));
  } catch {
    // ignore local save failure
  }
  await api("/api/work-form", {
    method: "POST",
    body: JSON.stringify({ state: formState }),
  }).catch(() => {});
}

function persistWorkFormChoiceNow() {
  void saveWorkFormStateNow();
}

function newerWorkFormState(localState, serverState) {
  if (!localState) return serverState || null;
  if (!serverState) return localState;
  const localTime = Date.parse(localState.updated_at || "") || 0;
  const serverTime = Date.parse(serverState.updated_at || "") || 0;
  return localTime >= serverTime ? localState : serverState;
}

function removeLegacyWorkInputSample(formState) {
  if (!formState || typeof formState !== "object") return { formState, removed: false };
  if (typeof formState.input_text !== "string" || !formState.input_text.startsWith(LEGACY_WORK_INPUT_SAMPLE_PREFIX)) {
    return { formState, removed: false };
  }
  return { formState: { ...formState, input_text: "" }, removed: true };
}

function applyWorkFormStateToVersionControls(saved) {
  if (!saved || typeof saved !== "object") return false;
  const selected = new Set(asArray(saved.selected_versions).map(version => String(version).toLowerCase()));
  const outerControls = [
    ["original", $("v78TargetOriginal")],
    ...Array.from({ length: 5 }, (_, index) => [`ai${index + 1}`, $(`v78TargetAi${index + 1}`)]),
  ];
  const hasOuterControls = outerControls.some(([, input]) => input);
  if (Array.isArray(saved.selected_versions)) {
    for (const [version, input] of outerControls) if (input) input.checked = selected.has(version);
    document.querySelectorAll('.process-version').forEach(input => { input.checked = selected.has(String(input.value || '').toLowerCase()); });
  }
  for (let index = 1; index <= 5; index += 1) {
    const method = saved.ai_slot_methods?.[`ai${index}`] || "";
    const outer = $(`v78RunAiMethod${index}`);
    const legacy = $(`processAiMethod${index}`);
    if (outer) outer.value = method;
    if (legacy) legacy.value = method;
  }
  if (state.config?.app_config) {
    state.config.app_config.rewrite = state.config.app_config.rewrite || {};
    state.config.app_config.rewrite.ai_slot_methods = {
      ...(state.config.app_config.rewrite.ai_slot_methods || {}),
      ...(saved.ai_slot_methods || {}),
    };
  }
  return hasOuterControls;
}

window.batchRewriteApplyWorkFormState = applyWorkFormStateToVersionControls;

function restoreWorkFormState() {
  let localState = null;
  try {
    localState = JSON.parse(localStorage.getItem(WORK_FORM_STORAGE_KEY) || "null");
  } catch {
    localState = null;
  }
  const local = removeLegacyWorkInputSample(localState);
  const server = removeLegacyWorkInputSample(state.config?.work_form || null);
  const saved = newerWorkFormState(local.formState, server.formState);
  if (!saved || typeof saved !== "object") return;
  const platformSelect = $("platformSelect");
  if (saved.platform_id && [...platformSelect.options].some((option) => option.value === String(saved.platform_id))) {
    platformSelect.value = String(saved.platform_id);
  }
  if (typeof saved.input_text === "string") $("inputText").value = saved.input_text;
  if (!applyWorkFormStateToVersionControls(saved)) window.__batchRewritePendingWorkFormState = saved;
  if (typeof saved.sensitive_ai_enabled === "boolean" && $("sensitiveAiProcessEnabled")) {
    $("sensitiveAiProcessEnabled").checked = saved.sensitive_ai_enabled;
  }
  if (local.removed || server.removed) window.setTimeout(() => { saveWorkFormStateNow(); }, 0);
  updateVersionConfigSummary();
}

function bindWorkFormPersistence() {
  for (const id of ["platformSelect"]) {
    const element = $(id);
    if (element) element.addEventListener("change", persistWorkFormChoiceNow);
  }
  for (const id of ["inputText"]) {
    const element = $(id);
    if (element) element.addEventListener("input", saveWorkFormState);
  }
  const sensitiveAiToggle = $("sensitiveAiProcessEnabled");
  if (sensitiveAiToggle) sensitiveAiToggle.addEventListener("change", persistWorkFormChoiceNow);
  const persistVersionSelection = event => {
    if (!event.target.matches('.process-version, [id^="processAiMethod"], #v78TargetVersions input, [id^="v78RunAiMethod"]')) return;
    persistWorkFormChoiceNow();
    updateVersionConfigSummary();
  };
  document.addEventListener('change', persistVersionSelection);
}

function presetOptions() {
  const presets = state.config?.app_config?.ai_presets || [];
  return [
    { id: "__current__", name: "当前AI配置" },
    ...presets.map((item) => ({ id: item.id, name: `${item.name || item.id} (${item.model || "未填模型"})` })),
  ];
}

function processAiMethods() {
  const result = {};
  for (let index = 1; index <= 5; index += 1) {
    const element = $(`v78RunAiMethod${index}`) || $(`processAiMethod${index}`);
    const value = element?.value || "";
    if (value) result[`ai${index}`] = value;
  }
  return result;
}

function selectedProcessVersions() {
  const outerControls = [
    ["original", $("v78TargetOriginal")],
    ...Array.from({ length: 5 }, (_, index) => [`ai${index + 1}`, $(`v78TargetAi${index + 1}`)]),
  ];
  if (outerControls.some(([, input]) => input)) return outerControls.filter(([, input]) => input?.checked).map(([version]) => version);
  return [...document.querySelectorAll('.process-version:checked')].map(input => String(input.value || '').toLowerCase());
}

function updateVersionConfigSummary() {
  const summary = $("versionConfigSummary");
  if (!summary) return;
  const versions = selectedProcessVersions();
  summary.textContent = versions.length
    ? versions.map(version => version === "original" ? "原文" : version.toUpperCase()).join("、")
    : "未选择版本";
}

function setModalBusy(id, busy) {
  const key = String(id || "").trim();
  if (!key) return;
  if (busy) state.modalBusy.add(key);
  else state.modalBusy.delete(key);
  const dialog = $(key);
  if (dialog) {
    dialog.dataset.modalBusy = busy ? "true" : "false";
    dialog.setAttribute("aria-busy", busy ? "true" : "false");
  }
}

function isModalBusy(id) {
  return state.modalBusy.has(String(id || "").trim());
}

function openVersionConfigCard() {
  const dialog = $("versionConfigCard");
  if (!dialog) return;
  renderVersionPromptConfig();
  updateVersionConfigSummary();
  if (typeof dialog.showModal === "function") dialog.showModal();
  else dialog.setAttribute("open", "");
}

function renderVersionPromptConfig() {
  const rewrite = state.config?.app_config?.rewrite || {};
  const fields = {
    versionConfigRewritePrompt: rewrite.prompt || "",
    versionConfigProcessingRulePrompt: rewrite.processing_rule_prompt || "",
    versionConfigKnowledgeUsagePrompt: state.config?.knowledge?.usage_prompt || "",
  };
  for (const [id, value] of Object.entries(fields)) {
    const element = $(id);
    if (element) element.value = value;
  }
}

function syncVersionPromptConfigToForm() {
  const pairs = [
    ["versionConfigRewritePrompt", "rewritePrompt"],
    ["versionConfigProcessingRulePrompt", "processingRulePrompt"],
    ["versionConfigKnowledgeUsagePrompt", "knowledgeUsagePrompt"],
  ];
  for (const [sourceId, targetId] of pairs) {
    const source = $(sourceId);
    const target = $(targetId);
    if (source && target) target.value = source.value;
  }
}

function closeVersionConfigCard() {
  const dialog = $("versionConfigCard");
  if (!dialog) return;
  if (isModalBusy("versionConfigCard")) return;
  if (typeof dialog.close === "function") dialog.close();
  else dialog.removeAttribute("open");
}

function renderConfig() {
  const appCfg = state.config.app_config || {};
  ensureKnowledgeConfig();
  ensureSensitiveConfig();
  ensureWebSubmitConfig();

  const platformSelect = $("platformSelect");
  platformSelect.innerHTML = "";
  for (const item of visiblePlatforms()) {
    const option = document.createElement("option");
    option.value = item.id;
    option.textContent = `${item.name} (${item.id})`;
    platformSelect.appendChild(option);
  }

  renderWorkflowConfig(appCfg);
  renderVersionPromptConfig();
  renderAiConfig(appCfg);
  renderPresetControls(appCfg);
  renderKnowledgeSummary(state.config.knowledge_summary || {});
  renderLibraryManager();
  renderRuleEditor();
  renderWebSubmitConfig(state.config.web_submit || {});
  if ($("knowledgeUsagePrompt")) $("knowledgeUsagePrompt").value = state.config.knowledge?.usage_prompt || "";

  $("platformsText").value = pretty(state.config.platforms || []);
  $("stylesText").value = pretty(state.config.styles || []);
  restoreWorkFormState();
  updatePlatformHint();
}

function renderWorkflowConfig(appCfg) {
  const workflow = appCfg.workflow || {};
  const fetch = appCfg.fetch || {};
  const rewrite = appCfg.rewrite || {};
  const storage = appCfg.storage || {};
  $("workflowAutoClassify").checked = workflow.auto_classify_missing !== false;
  $("workflowAutoFetch").checked = workflow.auto_fetch_original !== false;
  $("workflowAutoRewrite").checked = workflow.auto_rewrite_after_fetch !== false;
  if ($("sensitiveAiProcessEnabled")) {
    $("sensitiveAiProcessEnabled").checked = appCfg.sensitive_ai?.enabled === true;
  }
  $("fetchEndpoint").value = fetch.endpoint || "https://txt.121w.com/api.php";
  $("fetchMaxTxt").value = fetch.default_max_txt || 4000;
  $("fetchConcurrency").value = fetch.concurrency || 4;
  $("fetchRetries").value = fetch.retries ?? 1;
  $("fetchTimeout").value = fetch.timeout_seconds || 30;
  $("fetchAutoDetectPlatform").checked = fetch.auto_detect_platform !== false;
  $("processLineCount").value = rewrite.process_line_count || 5;
  $("anchorLineCount").value = rewrite.anchor_line_count || 5;
  $("rewriteTemp").value = rewrite.temperature ?? 0.45;
  $("methodSequenceInput").value = Array.isArray(rewrite.method_sequence)
    ? rewrite.method_sequence.join(",")
    : (rewrite.method_sequence || "high_imitation,opening_instruction,instruction");
  $("openingPhraseMode").value = rewrite.opening_phrase_mode || "auto";
  $("highImitationMode").value = rewrite.high_imitation_mode || "auto";
  renderRewriteTemplateOptions(rewrite.default_template_id || "");
  $("rewritePrompt").value = rewrite.prompt || "";
  $("processingRulePrompt").value = rewrite.processing_rule_prompt || "";
}

function asArray(value) {
  if (Array.isArray(value)) return value;
  if (value === undefined || value === null || value === "") return [];
  return [value];
}

function splitLooseList(value) {
  return String(value || "")
    .split(/[\n,，、|/]+/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function joinLooseList(value) {
  return asArray(value).map((item) => String(item || "").trim()).filter(Boolean).join("，");
}

function ensureKnowledgeConfig() {
  state.config.knowledge = state.config.knowledge || {};
  state.config.knowledge.high_imitation = state.config.knowledge.high_imitation || { prompts: [], references: [] };
  const high = state.config.knowledge.high_imitation;
  high.prompts = asArray(high.prompts);
  high.references = asArray(high.references);
  if (!high.references.length && Array.isArray(high.items) && high.items.length) {
    high.references = high.items.map((item, index) => {
      const source = item && typeof item === "object" ? item : {};
      return {
        ...source,
        id: source.id || `high_${String(index + 1).padStart(3, "0")}`,
        name: source.name || source.title || "",
        reference_text: source.reference_text || source.content || "",
        enabled: source.enabled !== false,
      };
    });
  }
  state.config.knowledge.opening_phrases = state.config.knowledge.opening_phrases || { items: [] };
  state.config.knowledge.opening_phrases.items = asArray(state.config.knowledge.opening_phrases.items);
  state.config.knowledge.opening_analyzer = state.config.knowledge.opening_analyzer || {};
  state.config.knowledge.rewrite_templates = state.config.knowledge.rewrite_templates || { profiles: [], temporary_instructions: [] };
  state.config.knowledge.rewrite_templates.profiles = asArray(state.config.knowledge.rewrite_templates.profiles);
  state.config.knowledge.layout_rules = state.config.knowledge.layout_rules || {};
  state.config.knowledge.chapter_rules = state.config.knowledge.chapter_rules || { chapter_exact_rules: [], chapter_inline_rules: [] };
  state.config.knowledge.symbol_rules = state.config.knowledge.symbol_rules || { symbol_rules: [], pair_fill_rules: [] };
}

function ensureSensitiveConfig() {
  const current = state.config.sensitive;
  if (Array.isArray(current)) {
    state.config.sensitive = {
      groups: [
        {
          name: "默认敏感词组",
          enabled: true,
          apply_to_original: true,
          apply_to_ai: true,
          rules: current,
        },
      ],
    };
  }
  if (!state.config.sensitive || typeof state.config.sensitive !== "object") {
    state.config.sensitive = { groups: [] };
  }
  state.config.sensitive.groups = asArray(state.config.sensitive.groups);
  if (!state.config.sensitive.groups.length) {
    state.config.sensitive.groups.push({
      name: "默认敏感词组",
      enabled: true,
      apply_to_original: true,
      apply_to_ai: true,
      rules: [],
    });
  }
  for (const group of state.config.sensitive.groups) {
    group.rules = asArray(group.rules);
  }
}

function ensureSensitiveAiConfig() {
  const appCfg = state.config.app_config = state.config.app_config || {};
  const current = appCfg.sensitive_ai || {};
  appCfg.sensitive_ai = {
    enabled: Boolean(current.enabled),
    context_chars: current.context_chars ?? 12,
    max_hits_per_task: current.max_hits_per_task ?? 80,
    concurrency: current.concurrency ?? 4,
    retries: current.retries ?? 1,
    temperature: current.temperature ?? 0.2,
    prompt: current.prompt || [
      "你是内容合规改写助手。请只改写下面命中敏感词的小段内容。",
      "要求：保留原剧情意思、人物关系和情绪；不增加新剧情；去掉违规、擦边、色情、低俗表达。",
      "只返回改写后的小段，不要解释。",
      "",
      "任务ID：{book_id}",
      "命中词：{keyword}",
      "",
      "原片段：",
      "{snippet}",
    ].join("\n"),
  };
  return appCfg.sensitive_ai;
}

function sensitiveAiProcessEnabled() {
  const toggle = $("sensitiveAiProcessEnabled");
  if (toggle) return toggle.checked === true;
  return state.config?.app_config?.sensitive_ai?.enabled === true;
}

function ensureWebSubmitConfig() {
  state.config = state.config || {};
  const current = state.config.web_submit || {};
  const jieyaNum = materialJieyaValue(current.advanced?.jieyaNum ?? 4);
  state.config.web_submit = {
    // 保留旧字段供服务端和历史配置兼容，但网络提交已由任务列表直接触发。
    enabled: true,
    username: current.username || "",
    password: "",
    password_masked: Boolean(current.password_masked),
    skip_submitted: current.skip_submitted !== false,
    min_text_chars: Math.max(0, Number(current.min_text_chars) || 0),
    retry_times: Math.max(0, Number(current.retry_times) || 1),
    upload_profiles: asArray(current.upload_profiles),
    profile_bindings: current.profile_bindings && typeof current.profile_bindings === 'object' ? current.profile_bindings : {},
    organization: String(current.organization || '1'),
    organization_catalog: asArray(current.organization_catalog),
    advanced: {
      tl5: Number(current.advanced?.tl5) === 1 ? 1 : 0,
      jieyaNum,
      jieyaAiHead: Number(current.advanced?.jieyaAiHead ?? 0),
      jieyaSpeed: Number(current.advanced?.jieyaSpeed ?? 1.7),
      jieyaPitch: Number(current.advanced?.jieyaPitch ?? 0),
      gunpingNum: PER_BOOK_MATERIAL_LIMIT - jieyaNum,
      gunpingSpeed: Number(current.advanced?.gunpingSpeed ?? 1),
      ziti: Number(current.advanced?.ziti ?? 1),
      zitidx: Number(current.advanced?.zitidx ?? 62),
      biaohong: String(current.advanced?.biaohong || ''),
      keywords: String(current.advanced?.keywords || '')
    }
  };
  return state.config.web_submit;
}

function getHighPrompt() {
  ensureKnowledgeConfig();
  const prompts = state.config.knowledge.high_imitation.prompts;
  if (!prompts.length) {
    prompts.push({
      id: "mimic_prompt_default",
      name: "高仿专用指令",
      content: "请学习参考文案的开口方式、节奏和冲突表达，再根据原文进行高仿改写。不能照搬参考剧情，不能新增重大设定。",
      enabled: true,
    });
  }
  return prompts[0];
}

function setHighPromptContent(content) {
  const prompt = getHighPrompt();
  prompt.content = content;
  prompt.prompt = content;
  prompt.system_prompt = content;
}

function libraryDefinition(type = "") {
  ensureKnowledgeConfig();
  const defs = {
    high_imitation: {
      title: "高仿文章库",
      getItems: () => {
        const high = state.config.knowledge.high_imitation;
        const references = asArray(high.references);
        return references.length ? references : asArray(high.items);
      },
      setItems: (items) => {
        state.config.knowledge.high_imitation.references = items;
      },
      create: () => ({
        id: `mimic_${Date.now()}`,
        name: "新的高仿参考",
        gender: "",
        style: "",
        tags: [],
        reference_text: "",
        analysis: "",
        enabled: true,
      }),
      label: (item, index) => `${item.name || item.id || `高仿参考${index + 1}`} / ${item.style || "未分类"} / ${item.gender || "未分频"}`,
      renderForm: (item) => {
        const prompt = getHighPrompt();
        return `
          <div class="notice">高仿文章库是单独一套改文方法，只给“高仿文章”方案调用。</div>
          <label>
            高仿专用指令
            <textarea id="libHighPrompt" class="small-area" spellcheck="false">${escapeHtml(prompt.content || prompt.prompt || prompt.system_prompt || "")}</textarea>
          </label>
          <div class="control-grid two">
            ${inputField("libId", "条目ID", item.id)}
            ${inputField("libName", "名称", item.name)}
            ${inputField("libStyle", "风格类型", item.style)}
            ${inputField("libGender", "男女频", item.gender)}
            ${inputField("libTags", "标签", joinLooseList(item.tags))}
            ${checkboxField("libEnabled", "启用", item.enabled !== false)}
          </div>
          <label>
            高仿参考正文
            <textarea id="libReferenceText" class="large-area" spellcheck="false">${escapeHtml(item.reference_text || "")}</textarea>
          </label>
          <label>
            拆解备注
            <textarea id="libAnalysis" class="small-area" spellcheck="false">${escapeHtml(item.analysis || "")}</textarea>
          </label>
        `;
      },
      readForm: (oldItem) => {
        setHighPromptContent($("libHighPrompt").value);
        return {
          ...oldItem,
          id: $("libId").value.trim() || oldItem.id || `mimic_${Date.now()}`,
          name: $("libName").value.trim(),
          style: $("libStyle").value.trim(),
          gender: $("libGender").value.trim(),
          tags: splitLooseList($("libTags").value),
          reference_text: $("libReferenceText").value,
          analysis: $("libAnalysis").value,
          enabled: $("libEnabled").checked,
        };
      },
    },
    opening_phrases: {
      title: "爆款开头词库",
      getItems: () => state.config.knowledge.opening_phrases.items,
      setItems: (items) => {
        state.config.knowledge.opening_phrases.items = items;
      },
      create: () => ({
        id: `opening_${Date.now()}`,
        name: "新的开头词",
        style: "",
        content: "",
        analysis: "",
        slot_hint: "",
        tags: [],
        enabled: true,
        linked_template_ids: [],
      }),
      label: (item, index) => `${item.name || item.id || `开头词${index + 1}`} / ${item.style || item.category || "自动风格"}`,
      renderForm: (item) => `
        <div class="notice">爆款开头词库会配合“批量改文指令库”使用，不单独替代整套改文指令。</div>
        <div class="control-grid two">
          ${inputField("libId", "条目ID", item.id)}
          ${inputField("libName", "名称", item.name)}
          ${openingStyleSelect("libStyle", "适用风格类型", item.style || item.category || "")}
          ${inputField("libTags", "标签", joinLooseList(item.tags))}
          ${inputField("libLinkedTemplates", "绑定指令模板ID", joinLooseList(item.linked_template_ids))}
          ${checkboxField("libEnabled", "启用", item.enabled !== false)}
        </div>
        <label>
          开头词骨架
          <textarea id="libContent" class="large-area" spellcheck="false">${escapeHtml(item.content || "")}</textarea>
        </label>
        <label>
          拆解说明
          <textarea id="libAnalysis" class="small-area" spellcheck="false">${escapeHtml(item.analysis || "")}</textarea>
        </label>
        <label>
          槽位 / 适配提示
          <textarea id="libSlotHint" class="small-area" spellcheck="false">${escapeHtml(item.slot_hint || "")}</textarea>
        </label>
      `,
      readForm: (oldItem) => ({
        id: $("libId").value.trim() || oldItem.id || `opening_${Date.now()}`,
        name: $("libName").value.trim(),
        style: $("libStyle").value.trim(),
        tags: splitLooseList($("libTags").value),
        linked_template_ids: splitLooseList($("libLinkedTemplates").value),
        content: $("libContent").value,
        analysis: $("libAnalysis").value,
        slot_hint: $("libSlotHint").value,
        enabled: $("libEnabled").checked,
      }),
    },
    rewrite_templates: {
      title: "批量改文指令库",
      getItems: () => state.config.knowledge.rewrite_templates.profiles,
      setItems: (items) => {
        state.config.knowledge.rewrite_templates.profiles = items;
      },
      create: () => ({
        id: `rewrite_${Date.now()}`,
        name: "新的改文指令",
        system_prompt: "",
        user_prompt_template: "",
        required_variables: ["content_id", "book_name", "style", "gender", "full_text", "target_lines_text", "anchor_lines_text", "line_count", "extra_instruction"],
        response_mode: "changed_lines_json",
        enabled: true,
      }),
      label: (item, index) => `${item.name || item.id || `改文指令${index + 1}`} / ${item.response_mode || "普通输出"}`,
      renderForm: (item) => `
        <div class="notice">这里就是 AI指令模板与提示词逻辑。用户提示词模板会直接参与 AI 改文。</div>
        <div class="control-grid two">
          ${inputField("libId", "模板ID", item.id)}
          ${inputField("libName", "模板名称", item.name)}
          <label>
            返回方式
            <select id="libResponseMode">
              <option value="changed_lines_json" ${item.response_mode === "changed_lines_json" ? "selected" : ""}>只返回改写块JSON</option>
              <option value="plain" ${item.response_mode === "plain" ? "selected" : ""}>直接返回正文</option>
            </select>
          </label>
          ${checkboxField("libEnabled", "启用", item.enabled !== false)}
        </div>
        <label>
          系统提示词
          <textarea id="libSystemPrompt" class="large-area" spellcheck="false">${escapeHtml(item.system_prompt || "")}</textarea>
        </label>
        <label>
          用户提示词模板
          <textarea id="libUserPromptTemplate" class="large-area" spellcheck="false">${escapeHtml(item.user_prompt_template || "")}</textarea>
        </label>
        <label>
          可用变量
          <input id="libRequiredVariables" value="${escapeHtml(joinLooseList(item.required_variables))}" />
        </label>
      `,
      readForm: (oldItem) => ({
        ...oldItem,
        id: $("libId").value.trim() || oldItem.id || `rewrite_${Date.now()}`,
        name: $("libName").value.trim(),
        system_prompt: $("libSystemPrompt").value,
        user_prompt_template: $("libUserPromptTemplate").value,
        required_variables: splitLooseList($("libRequiredVariables").value),
        response_mode: $("libResponseMode").value,
        enabled: $("libEnabled").checked,
      }),
    },
  };
  return defs[type] || defs.high_imitation;
}

function inputField(id, label, value = "") {
  return `<label>${escapeHtml(label)}<input id="${escapeHtml(id)}" value="${escapeHtml(value || "")}" /></label>`;
}

function checkboxField(id, label, checked = false) {
  return `<label class="inline-check"><input id="${escapeHtml(id)}" type="checkbox" ${checked ? "checked" : ""} /> ${escapeHtml(label)}</label>`;
}

function textAreaField(id, label, value = "", className = "small-area") {
  return `<label>${escapeHtml(label)}<textarea id="${escapeHtml(id)}" class="${escapeHtml(className)}" spellcheck="false">${escapeHtml(value || "")}</textarea></label>`;
}

function openingStyleSelect(id, label, selected = "") {
  const options = asArray(state.config?.styles).map((style) => {
    const value = String(style || "").trim();
    return `<option value="${escapeHtml(value)}" ${value === selected ? "selected" : ""}>${escapeHtml(value)}</option>`;
  }).join("");
  return `
    <label>
      ${escapeHtml(label)}
      <select id="${escapeHtml(id)}">
        <option value="">自动选择</option>
        ${options}
      </select>
    </label>
  `;
}

function renderLibraryManager(selectedIndex = null) {
  const select = $("libraryItemSelect");
  const form = $("libraryForm");
  if (!select || !form) return;
  const type = $("libraryTypeSelect")?.value || "high_imitation";
  renderOpeningAnalyzerPanel(type);
  const def = libraryDefinition(type);
  const items = def.getItems();
  const keyword = ($("librarySearchInput")?.value || "").trim().toLowerCase();
  const filtered = items
    .map((item, index) => ({ item, index }))
    .filter(({ item }) => !keyword || JSON.stringify(item).toLowerCase().includes(keyword));
  select.innerHTML = "";
  for (const { item, index } of filtered) {
    const option = document.createElement("option");
    option.value = String(index);
    option.textContent = def.label(item, index);
    select.appendChild(option);
  }
  if (!filtered.length) {
    select.innerHTML = `<option value="">暂无条目</option>`;
    form.innerHTML = `<div class="detail-empty">没有找到条目，可以新增一个。</div>`;
    return;
  }
  const wanted = selectedIndex === null ? Number(select.value || filtered[0].index) : selectedIndex;
  const match = filtered.find((entry) => entry.index === wanted) || filtered[0];
  select.value = String(match.index);
  form.innerHTML = def.renderForm(items[match.index], match.index);
}

function renderOpeningAnalyzerPanel(type = "") {
  const panel = $("openingAnalyzerSection");
  if (!panel) return;
  panel.classList.toggle("hidden", type !== "opening_phrases");
  if (type !== "opening_phrases") return;
  const select = $("openingAnalyzeStyle");
  if (!select) return;
  const currentValue = select.value || "";
  select.innerHTML = `<option value="">AI自动选择适用风格</option>`;
  for (const style of asArray(state.config?.styles)) {
    const value = String(style || "").trim();
    if (!value) continue;
    const option = document.createElement("option");
    option.value = value;
    option.textContent = value;
    select.appendChild(option);
  }
  select.value = currentValue;
}

function captureCurrentLibraryForm() {
  const select = $("libraryItemSelect");
  const form = $("libraryForm");
  if (!select || !form || !form.querySelector("input, textarea, select")) return false;
  const type = $("libraryTypeSelect")?.value || "high_imitation";
  const def = libraryDefinition(type);
  const items = [...def.getItems()];
  const index = Number(select.value);
  if (!items.length || !Number.isFinite(index) || !items[index]) return false;
  items[index] = def.readForm(items[index]);
  def.setItems(items);
  return true;
}

function addLibraryItem() {
  captureCurrentLibraryForm();
  const type = $("libraryTypeSelect").value;
  const def = libraryDefinition(type);
  const items = [...def.getItems(), def.create()];
  def.setItems(items);
  renderLibraryManager(items.length - 1);
  $("knowledgeStatus").textContent = "已新增，记得保存知识库配置";
}

function duplicateLibraryItem() {
  captureCurrentLibraryForm();
  const type = $("libraryTypeSelect").value;
  const def = libraryDefinition(type);
  const items = [...def.getItems()];
  const index = Number($("libraryItemSelect").value);
  if (!items[index]) return;
  const next = clone(items[index]);
  next.id = `${next.id || "copy"}_${Date.now()}`;
  next.name = `${next.name || "复制条目"} 副本`;
  items.splice(index + 1, 0, next);
  def.setItems(items);
  renderLibraryManager(index + 1);
  $("knowledgeStatus").textContent = "已复制，记得保存知识库配置";
}

function saveLibraryItem(silent = false) {
  if (!captureCurrentLibraryForm()) return;
  renderLibraryManager(Number($("libraryItemSelect").value || 0));
  renderRewriteTemplateOptions(state.config?.app_config?.rewrite?.default_template_id || "");
  if (!silent) $("knowledgeStatus").textContent = "条目已保存到页面，记得保存知识库配置";
}

function deleteLibraryItem() {
  const type = $("libraryTypeSelect").value;
  const def = libraryDefinition(type);
  const items = [...def.getItems()];
  const index = Number($("libraryItemSelect").value);
  if (!items.length || !Number.isFinite(index) || !items[index]) return;
  if (!confirm("确认删除当前知识库条目？")) return;
  items.splice(index, 1);
  def.setItems(items);
  renderLibraryManager(Math.max(0, index - 1));
  $("knowledgeStatus").textContent = "条目已删除，记得保存知识库配置";
}

async function optimizeLibraryItem() {
  if (!captureCurrentLibraryForm()) return;
  const type = $("libraryTypeSelect").value;
  const def = libraryDefinition(type);
  const items = [...def.getItems()];
  const index = Number($("libraryItemSelect").value);
  if (!items[index]) return;
  $("knowledgeStatus").textContent = "AI优化中...";
  try {
    const result = await api("/api/knowledge/optimize", {
      method: "POST",
      body: JSON.stringify({
        library_type: type,
        item: items[index],
        goal: $("promptImproveGoal").value,
      }),
    });
    items[index] = { ...items[index], ...(result.item || {}) };
    def.setItems(items);
    renderLibraryManager(index);
    $("knowledgeStatus").textContent = "AI建议已填入，确认后保存配置";
  } catch (error) {
    $("knowledgeStatus").textContent = error.message;
  }
}

function ruleList(type) {
  ensureKnowledgeConfig();
  ensureSensitiveConfig();
  if (type === "chapter_exact_rules") return state.config.knowledge.chapter_rules.chapter_exact_rules = asArray(state.config.knowledge.chapter_rules.chapter_exact_rules);
  if (type === "chapter_inline_rules") return state.config.knowledge.chapter_rules.chapter_inline_rules = asArray(state.config.knowledge.chapter_rules.chapter_inline_rules);
  if (type === "symbol_rules") return state.config.knowledge.symbol_rules.symbol_rules = asArray(state.config.knowledge.symbol_rules.symbol_rules);
  if (type === "pair_fill_rules") return state.config.knowledge.symbol_rules.pair_fill_rules = asArray(state.config.knowledge.symbol_rules.pair_fill_rules);
  return [];
}

function renderRuleSelect(list, selectedIndex = 0) {
  const options = list.map((item, index) => `<option value="${index}" ${index === selectedIndex ? "selected" : ""}>${escapeHtml(item.find || item.name || `规则${index + 1}`)}${item.enabled === false ? "（停用）" : ""}</option>`).join("");
  return `<select id="ruleItemSelect" class="list-select" size="10">${options || `<option value="">暂无规则</option>`}</select>`;
}

function ruleHelp(type) {
  return {
    sensitive: "敏感词规则统一管理命中词：开启AI时原文命中小句会交给AI修复；关闭AI时才按替换内容处理。",
    layout_rules: "排版规则会在抓取后和AI返回后自动执行。",
    chapter_exact_rules: "整行命中章节规则时删除整行。",
    chapter_inline_rules: "行内命中章节规则时按设置清洗。",
    symbol_rules: "字位符号会按位置把句子切开或清理。",
    pair_fill_rules: "成对补符号会自动补齐左右符号。",
  }[type] || "添加规则后保存配置";
}

function ruleAiMeta(type) {
  return {
    sensitive: {
      title: "敏感词AI助手",
      sample: "粘贴一段原文或AI文案，AI会提取需要命中的敏感词。",
      goal: "例如：提取违规称呼、敏感动作、平台敏感词；开启AI处理时只需要命中词。",
    },
    layout_rules: {
      title: "批量排版AI助手",
      sample: "粘贴你想对标的排版样本文案，或填入当前处理效果。",
      goal: "例如：建议空行、缩进、是否删除空行、是否清理首尾空格。换行只交给字位符号规则。",
    },
    chapter_exact_rules: {
      title: "章节整行清洗AI助手",
      sample: "粘贴带章节标题的文本，AI会生成整行删除规则。",
      goal: "例如：识别“第1章”“第一章 标题”“章节序号”这类整行清洗规则。",
    },
    chapter_inline_rules: {
      title: "章节行内清洗AI助手",
      sample: "粘贴章节序号和正文在同一行的文本。",
      goal: "例如：识别行内开头的章节序号并生成替换为空的规则。",
    },
    symbol_rules: {
      title: "字位符号AI助手",
      sample: "粘贴符号位置异常、需要按指定位置切句的文本。",
      goal: "例如：分析哪些符号要切分、命中第几位、是否允许相邻命中。",
    },
    pair_fill_rules: {
      title: "成对补符号AI助手",
      sample: "粘贴缺引号、缺括号、符号不成对的文本。",
      goal: "例如：生成左符号和右符号补齐规则。",
    },
  }[type] || {};
}

function renderRuleAiAssistant(type) {
  const meta = ruleAiMeta(type);
  if ($("ruleAiTitle")) $("ruleAiTitle").textContent = meta.title || "规则AI助手";
  if ($("ruleAiSample")) $("ruleAiSample").placeholder = meta.sample || "粘贴样本文案";
  if ($("ruleAiGoal")) $("ruleAiGoal").placeholder = meta.goal || "填写AI处理要求";
  if ($("ruleAiStatus")) $("ruleAiStatus").textContent = "";
  if ($("ruleAiResult")) $("ruleAiResult").textContent = "AI生成后会显示待加入规则。";
}

function renderRuleEditor(selectedIndex = 0) {
  const box = $("ruleManager");
  if (!box) return;
  ensureKnowledgeConfig();
  ensureSensitiveConfig();
  const type = $("ruleTypeSelect")?.value || "sensitive";
  $("ruleHelpText").value = ruleHelp(type);
  state.pendingRuleSuggestions = null;
  renderRuleAiAssistant(type);
  if (type === "sensitive") {
    renderSensitiveRuleEditor();
    return;
  }
  if (type === "layout_rules") {
    renderLayoutRuleEditor();
    return;
  }
  if (type === "chapter_exact_rules" || type === "chapter_inline_rules") {
    renderChapterRuleEditor(type, selectedIndex);
    return;
  }
  renderSymbolRuleEditor(type, selectedIndex);
}

function renderSensitiveRuleEditor() {
  const groups = state.config.sensitive.groups;
  const groupIndex = Math.max(0, Math.min(Number($("sensitiveGroupSelect")?.value || 0), groups.length - 1));
  const group = groups[groupIndex];
  const rules = asArray(group.rules);
  const ruleIndex = Math.max(0, Math.min(Number($("sensitiveRuleSelect")?.value || 0), Math.max(rules.length - 1, 0)));
  const rule = rules[ruleIndex] || { find: "", replace: "", enabled: true };
  const sensitiveAi = ensureSensitiveAiConfig();
  $("ruleManager").innerHTML = `
    <div class="notice">
      敏感词 AI 模式由“处理”页面的开关控制。开启时先调用 AI 修复，AI 失败会自动回退为“查找内容 → 替换为”；关闭时始终直接替换。这里仅维护词组、替换内容和 AI 参数。
    </div>
    <div class="control-grid two">
      ${inputField("sensitiveAiConcurrency", "AI修复并发数", sensitiveAi.concurrency)}
      ${inputField("sensitiveAiMaxHits", "每任务最大命中数", sensitiveAi.max_hits_per_task)}
      ${inputField("sensitiveAiRetries", "AI修复重试次数", sensitiveAi.retries)}
      ${inputField("sensitiveAiContext", "命中句前后扩展字数", sensitiveAi.context_chars)}
      ${inputField("sensitiveAiTemperature", "AI修复温度", sensitiveAi.temperature)}
    </div>
    <label>
      敏感词AI修复提示词
      <textarea id="sensitiveAiPrompt" class="small-area" spellcheck="false">${escapeHtml(sensitiveAi.prompt || "")}</textarea>
    </label>
    <div class="control-grid two">
      <label>
        敏感词组
        <select id="sensitiveGroupSelect">${groups.map((item, index) => `<option value="${index}" ${index === groupIndex ? "selected" : ""}>${escapeHtml(item.name || `词组${index + 1}`)}</option>`).join("")}</select>
      </label>
      ${inputField("sensitiveGroupName", "词组名称", group.name)}
      ${checkboxField("sensitiveGroupEnabled", "启用词组", group.enabled !== false)}
      ${checkboxField("sensitiveApplyOriginal", "处理原文", group.apply_to_original !== false)}
      ${checkboxField("sensitiveApplyAi", "处理AI文案", group.apply_to_ai !== false)}
    </div>
    <div class="actions">
      <button data-rule-action="add-sensitive-group">新增词组</button>
      <button data-rule-action="delete-sensitive-group" class="danger">删除词组</button>
    </div>
    <div class="manager-grid">
      <div>
        <label>规则列表</label>
        <select id="sensitiveRuleSelect" class="list-select" size="10">${rules.map((item, index) => `<option value="${index}" ${index === ruleIndex ? "selected" : ""}>${escapeHtml(item.find || `敏感词${index + 1}`)}${item.enabled === false ? "（停用）" : ""}</option>`).join("") || `<option value="">暂无规则</option>`}</select>
      </div>
      <div class="form-stack">
        ${inputField("sensitiveFind", "查找内容 / 敏感词", rule.find)}
        ${inputField("sensitiveReplace", "替换为（AI失败或关闭时使用）", rule.replace)}
        ${checkboxField("sensitiveRuleEnabled", "启用规则", rule.enabled !== false)}
        <div class="actions">
          <button data-rule-action="add-sensitive-rule">新增敏感词</button>
          <button data-rule-action="delete-sensitive-rule" class="danger">删除敏感词</button>
        </div>
      </div>
    </div>
    <label>
      批量添加敏感词
      <textarea id="sensitiveBulkText" class="large-area" spellcheck="false" placeholder="一行一个：敏感词 或 敏感词=>替换词。AI 失败或关闭时使用替换词。"></textarea>
    </label>
    <div class="actions">
      <button data-rule-action="bulk-sensitive-rule">批量加入当前词组</button>
    </div>
  `;
}

function renderLayoutRuleEditor() {
  const cfg = state.config.app_config = state.config.app_config || {};
  const layoutCfg = cfg.layout = cfg.layout || {};
  const rules = state.config.knowledge.layout_rules = state.config.knowledge.layout_rules || {};
  $("ruleManager").innerHTML = `
    <div class="notice">
      批量排版只负责空行、缩进、敏感词、章节清洗和符号补齐。正文换行只由“字位符号规则”处理。
    </div>
    <div class="control-grid two">
      ${checkboxField("layoutApplyOriginal", "自动处理原文", layoutCfg.apply_to_original !== false)}
      ${checkboxField("layoutApplyAi", "自动处理AI文案", layoutCfg.apply_to_ai !== false)}
      ${checkboxField("layoutApplySensitive", "启用敏感词", layoutCfg.apply_sensitive !== false)}
      ${checkboxField("layoutApplyChapter", "启用章节清洗", layoutCfg.apply_chapter_cleanup !== false)}
      ${checkboxField("layoutApplySymbol", "启用字位符号", layoutCfg.apply_symbol_rules !== false)}
      ${checkboxField("layoutApplyPair", "启用成对补符号", layoutCfg.apply_pair_fill !== false)}
      ${checkboxField("layoutDropEmpty", "删除空行", layoutCfg.drop_empty_lines !== false)}
      ${checkboxField("layoutTrimLines", "清理行首尾空格", layoutCfg.trim_lines !== false)}
      ${inputField("layoutIndentSpaces", "段首空格数", rules.indent_spaces ?? 0)}
      ${inputField("layoutBlankLines", "段落间空行数", rules.blank_lines_before ?? 0)}
      ${checkboxField("layoutSensitiveReplace", "排版时替换敏感词", rules.apply_sensitive_replace !== false)}
    </div>
  `;
}

function renderChapterRuleEditor(type, selectedIndex = 0) {
  const chapter = state.config.knowledge.chapter_rules;
  const list = ruleList(type);
  const index = Math.max(0, Math.min(selectedIndex, Math.max(list.length - 1, 0)));
  const rule = list[index] || { find: "", replace: "", enabled: true };
  $("ruleManager").innerHTML = `
    <div class="control-grid two">
      ${checkboxField("chapterEnabled", "启用章节清洗", chapter.chapter_cleanup_enabled !== false)}
      ${checkboxField("chapterRemoveStandalone", "删除独立章节行", chapter.chapter_remove_standalone !== false)}
      ${checkboxField("chapterRemoveInline", "启用行内章节规则", Boolean(chapter.chapter_remove_inline))}
      ${inputField("chapterMaxN", "最大章节序号", chapter.chapter_max_n ?? 200)}
    </div>
    <div class="manager-grid">
      <div>
        <label>章节规则列表</label>
        ${renderRuleSelect(list, index)}
      </div>
      <div class="form-stack">
        ${textAreaField("ruleFind", "命中规则", rule.find)}
        ${textAreaField("ruleReplace", "替换为", rule.replace)}
        ${checkboxField("ruleEnabled", "启用规则", rule.enabled !== false)}
        <div class="actions">
          <button data-rule-action="add-generic-rule">新增规则</button>
          <button data-rule-action="delete-generic-rule" class="danger">删除规则</button>
        </div>
      </div>
    </div>
  `;
}

function renderSymbolRuleEditor(type, selectedIndex = 0) {
  const symbol = state.config.knowledge.symbol_rules;
  const list = ruleList(type);
  const index = Math.max(0, Math.min(selectedIndex, Math.max(list.length - 1, 0)));
  const rule = list[index] || { find: "", replace: "", enabled: true };
  const title = type === "pair_fill_rules" ? "成对补符号规则" : "字位符号规则";
  $("ruleManager").innerHTML = `
    ${type === "symbol_rules" ? `
      <div class="control-grid two">
        ${checkboxField("symbolTrimEnabled", "启用字位符号切分", Boolean(symbol.enable_symbol_trim))}
        ${checkboxField("symbolAllowAdjacent", "允许相邻位置命中", Boolean(symbol.symbol_trim_allow_adjacent))}
        ${inputField("symbolTrimPositions", "命中位置", symbol.symbol_trim_positions)}
      </div>
    ` : ""}
    <div class="manager-grid">
      <div>
        <label>${title}列表</label>
        ${renderRuleSelect(list, index)}
      </div>
      <div class="form-stack">
        ${textAreaField("ruleFind", type === "pair_fill_rules" ? "左符号" : "符号", rule.find)}
        ${textAreaField("ruleReplace", type === "pair_fill_rules" ? "右符号" : "替换为", rule.replace)}
        ${checkboxField("ruleEnabled", "启用规则", rule.enabled !== false)}
        <div class="actions">
          <button data-rule-action="add-generic-rule">新增规则</button>
          <button data-rule-action="delete-generic-rule" class="danger">删除规则</button>
        </div>
      </div>
    </div>
  `;
}

function captureCurrentRuleForm() {
  const type = $("ruleTypeSelect")?.value || "sensitive";
  ensureKnowledgeConfig();
  ensureSensitiveConfig();
  if (type === "sensitive") {
    const sensitiveAi = ensureSensitiveAiConfig();
    sensitiveAi.concurrency = numberValue("sensitiveAiConcurrency", 4);
    sensitiveAi.max_hits_per_task = numberValue("sensitiveAiMaxHits", 80);
    sensitiveAi.retries = numberValue("sensitiveAiRetries", 1);
    sensitiveAi.context_chars = numberValue("sensitiveAiContext", 12);
    sensitiveAi.temperature = numberValue("sensitiveAiTemperature", 0.2);
    sensitiveAi.prompt = $("sensitiveAiPrompt").value;
    const groupIndex = Number($("sensitiveGroupSelect")?.value || 0);
    const group = state.config.sensitive.groups[groupIndex];
    if (!group) return false;
    group.name = $("sensitiveGroupName").value.trim() || group.name || `词组${groupIndex + 1}`;
    group.enabled = $("sensitiveGroupEnabled").checked;
    group.apply_to_original = $("sensitiveApplyOriginal").checked;
    group.apply_to_ai = $("sensitiveApplyAi").checked;
    const ruleIndex = Number($("sensitiveRuleSelect")?.value || 0);
    if (group.rules[ruleIndex]) {
      group.rules[ruleIndex] = {
        ...group.rules[ruleIndex],
        find: $("sensitiveFind").value,
        replace: $("sensitiveReplace").value,
        enabled: $("sensitiveRuleEnabled").checked,
      };
    }
    return true;
  }
  if (type === "layout_rules") {
    const cfg = state.config.app_config = state.config.app_config || {};
    const layoutCfg = cfg.layout = cfg.layout || {};
    layoutCfg.apply_to_original = $("layoutApplyOriginal").checked;
    layoutCfg.apply_to_ai = $("layoutApplyAi").checked;
    layoutCfg.apply_sensitive = $("layoutApplySensitive").checked;
    layoutCfg.apply_chapter_cleanup = $("layoutApplyChapter").checked;
    layoutCfg.apply_symbol_rules = $("layoutApplySymbol").checked;
    layoutCfg.apply_pair_fill = $("layoutApplyPair").checked;
    layoutCfg.drop_empty_lines = $("layoutDropEmpty").checked;
    layoutCfg.trim_lines = $("layoutTrimLines").checked;
    delete layoutCfg.line_break_mode;
    const rules = state.config.knowledge.layout_rules;
    rules.indent_spaces = Number($("layoutIndentSpaces").value || 0);
    rules.blank_lines_before = Number($("layoutBlankLines").value || 0);
    rules.apply_sensitive_replace = $("layoutSensitiveReplace").checked;
    delete rules.line_break_mode;
    delete rules.line_len_min;
    delete rules.line_len_max;
    delete rules.target_line_len;
    delete rules.enable_len_wrap;
    delete rules.delimiters;
    return true;
  }
  if (type === "chapter_exact_rules" || type === "chapter_inline_rules") {
    const chapter = state.config.knowledge.chapter_rules;
    chapter.chapter_cleanup_enabled = $("chapterEnabled").checked;
    chapter.chapter_remove_standalone = $("chapterRemoveStandalone").checked;
    chapter.chapter_remove_inline = $("chapterRemoveInline").checked;
    chapter.chapter_max_n = Number($("chapterMaxN").value || 200);
  }
  if (type === "symbol_rules") {
    const symbol = state.config.knowledge.symbol_rules;
    symbol.enable_symbol_trim = $("symbolTrimEnabled").checked;
    symbol.symbol_trim_allow_adjacent = $("symbolAllowAdjacent").checked;
    symbol.symbol_trim_positions = $("symbolTrimPositions").value;
  }
  const list = ruleList(type);
  const index = Number($("ruleItemSelect")?.value || 0);
  if (list[index]) {
    list[index] = {
      ...list[index],
      find: $("ruleFind").value,
      replace: $("ruleReplace").value,
      enabled: $("ruleEnabled").checked,
    };
  }
  return true;
}

function saveRuleEditor(silent = false) {
  if (!captureCurrentRuleForm()) return;
  if (!silent) $("ruleEditorStatus").textContent = "规则已保存到页面，记得保存处理规则配置";
}

function handleRuleAction(action) {
  captureCurrentRuleForm();
  const type = $("ruleTypeSelect").value;
  if (action === "add-sensitive-group") {
    state.config.sensitive.groups.push({ name: "新词组", enabled: true, apply_to_original: true, apply_to_ai: true, rules: [] });
    renderSensitiveRuleEditor();
    return;
  }
  if (action === "delete-sensitive-group") {
    if (state.config.sensitive.groups.length <= 1) return;
    state.config.sensitive.groups.splice(Number($("sensitiveGroupSelect").value || 0), 1);
    renderSensitiveRuleEditor();
    return;
  }
  if (action === "add-sensitive-rule") {
    const group = state.config.sensitive.groups[Number($("sensitiveGroupSelect").value || 0)];
    group.rules.push({ find: "", replace: "", enabled: true });
    renderSensitiveRuleEditor();
    return;
  }
  if (action === "delete-sensitive-rule") {
    const group = state.config.sensitive.groups[Number($("sensitiveGroupSelect").value || 0)];
    group.rules.splice(Number($("sensitiveRuleSelect").value || 0), 1);
    renderSensitiveRuleEditor();
    return;
  }
  if (action === "bulk-sensitive-rule") {
    const group = state.config.sensitive.groups[Number($("sensitiveGroupSelect").value || 0)];
    const lines = $("sensitiveBulkText").value.split(/\n+/).map((line) => line.trim()).filter(Boolean);
    for (const line of lines) {
      const parts = line.split(/=>|=|→/);
      group.rules.push({ find: (parts[0] || "").trim(), replace: (parts[1] || "").trim(), enabled: true });
    }
    renderSensitiveRuleEditor();
    return;
  }
  if (action === "add-generic-rule") {
    const list = ruleList(type);
    list.push({ find: "", replace: "", enabled: true });
    renderRuleEditor(list.length - 1);
    return;
  }
  if (action === "delete-generic-rule") {
    const list = ruleList(type);
    list.splice(Number($("ruleItemSelect")?.value || 0), 1);
    renderRuleEditor(0);
  }
}

function currentRuleConfigForAi(type) {
  ensureKnowledgeConfig();
  ensureSensitiveConfig();
  if (type === "sensitive") {
    const group = state.config.sensitive.groups[Number($("sensitiveGroupSelect")?.value || 0)] || {};
    return { group_name: group.name || "", ai_fix: ensureSensitiveAiConfig(), rules: group.rules || [] };
  }
  if (type === "layout_rules") {
    return {
      layout: state.config.app_config?.layout || {},
      layout_rules: state.config.knowledge.layout_rules || {},
    };
  }
  if (type === "chapter_exact_rules" || type === "chapter_inline_rules") {
    return state.config.knowledge.chapter_rules || {};
  }
  return state.config.knowledge.symbol_rules || {};
}

async function suggestCurrentRuleWithAi() {
  const type = $("ruleTypeSelect").value || "sensitive";
  saveRuleEditor(true);
  $("ruleAiStatus").textContent = "AI生成中...";
  $("ruleAiResult").textContent = "正在分析样本文案...";
  try {
    const result = await api("/api/rules/ai-suggest", {
      method: "POST",
      body: JSON.stringify({
        rule_type: type,
        sample_text: $("ruleAiSample").value,
        goal: $("ruleAiGoal").value,
        current_config: currentRuleConfigForAi(type),
      }),
    });
    state.pendingRuleSuggestions = result.suggestions || null;
    renderRuleSuggestionResult();
    $("ruleAiStatus").textContent = "已生成，确认后加入规则";
  } catch (error) {
    state.pendingRuleSuggestions = null;
    $("ruleAiStatus").textContent = error.message;
    $("ruleAiResult").textContent = error.message;
  }
}

function renderRuleSuggestionResult() {
  const box = $("ruleAiResult");
  const suggestions = state.pendingRuleSuggestions;
  if (!box) return;
  if (!suggestions) {
    box.textContent = "AI生成后会显示待加入规则。";
    return;
  }
  const rows = [];
  for (const key of ["rules", "chapter_exact_rules", "chapter_inline_rules", "symbol_rules", "pair_fill_rules"]) {
    for (const item of asArray(suggestions[key])) {
      if (!item || typeof item !== "object") continue;
      rows.push(`<div class="suggestion-item"><b>${escapeHtml(key)}</b><span>${escapeHtml(item.find || "")}${item.replace ? ` → ${escapeHtml(item.replace)}` : ""}</span></div>`);
    }
  }
  if (suggestions.layout_rules || suggestions.layout) {
    rows.push(`<div class="suggestion-item"><b>排版参数</b><span>${escapeHtml(Object.keys(suggestions.layout_rules || {}).concat(Object.keys(suggestions.layout || {})).join("，") || "无")}</span></div>`);
  }
  box.innerHTML = rows.join("") || `<div class="suggestion-item"><b>没有生成可加入规则</b><span>可以调整样本文案或AI处理要求后重试。</span></div>`;
}

function appendRulesUnique(target, incoming) {
  const list = asArray(target);
  const seen = new Set(list.map((item) => `${item?.find || ""}=>${item?.replace || ""}`));
  for (const item of asArray(incoming)) {
    if (!item || typeof item !== "object" || !String(item.find || "").trim()) continue;
    const next = { find: String(item.find || "").trim(), replace: String(item.replace || ""), enabled: item.enabled !== false };
    const key = `${next.find}=>${next.replace}`;
    if (seen.has(key)) continue;
    list.push(next);
    seen.add(key);
  }
  return list;
}

function applyRuleSuggestions() {
  const suggestions = state.pendingRuleSuggestions;
  if (!suggestions) {
    $("ruleAiStatus").textContent = "还没有AI生成的规则";
    return;
  }
  const type = suggestions.rule_type || $("ruleTypeSelect").value || "sensitive";
  ensureKnowledgeConfig();
  ensureSensitiveConfig();
  if (type === "sensitive") {
    const group = state.config.sensitive.groups[Number($("sensitiveGroupSelect")?.value || 0)] || state.config.sensitive.groups[0];
    group.rules = appendRulesUnique(group.rules, suggestions.rules);
    $("ruleAiStatus").textContent = `已加入 ${asArray(suggestions.rules).length} 条敏感词规则`;
    renderSensitiveRuleEditor();
  } else if (type === "layout_rules") {
    state.config.knowledge.layout_rules = {
      ...(state.config.knowledge.layout_rules || {}),
      ...(suggestions.layout_rules || {}),
    };
    state.config.app_config = state.config.app_config || {};
    state.config.app_config.layout = {
      ...(state.config.app_config.layout || {}),
      ...(suggestions.layout || {}),
    };
    $("ruleAiStatus").textContent = "已应用排版参数";
    renderLayoutRuleEditor();
  } else if (type === "chapter_exact_rules" || type === "chapter_inline_rules") {
    const chapter = state.config.knowledge.chapter_rules;
    chapter.chapter_exact_rules = appendRulesUnique(chapter.chapter_exact_rules, suggestions.chapter_exact_rules);
    chapter.chapter_inline_rules = appendRulesUnique(chapter.chapter_inline_rules, suggestions.chapter_inline_rules);
    if (suggestions.settings && typeof suggestions.settings === "object") {
      Object.assign(chapter, suggestions.settings);
    }
    renderRuleEditor(0);
    $("ruleAiStatus").textContent = "已加入章节清洗规则";
  } else if (type === "symbol_rules") {
    const symbol = state.config.knowledge.symbol_rules;
    symbol.symbol_rules = appendRulesUnique(symbol.symbol_rules, suggestions.symbol_rules);
    if (suggestions.settings && typeof suggestions.settings === "object") {
      Object.assign(symbol, suggestions.settings);
    }
    renderRuleEditor(0);
    $("ruleAiStatus").textContent = "已加入字位符号规则";
  } else if (type === "pair_fill_rules") {
    const symbol = state.config.knowledge.symbol_rules;
    symbol.pair_fill_rules = appendRulesUnique(symbol.pair_fill_rules, suggestions.pair_fill_rules);
    renderRuleEditor(0);
    $("ruleAiStatus").textContent = "已加入成对补符号规则";
  }
  state.pendingRuleSuggestions = null;
  if ($("ruleAiResult")) $("ruleAiResult").textContent = "已加入对应规则，记得保存处理规则配置。";
}

function renderRewriteTemplateOptions(selectedId = "") {
  const select = $("rewriteTemplateSelect");
  const profiles = state.config?.knowledge?.rewrite_templates?.profiles || [];
  const options = profiles
    .filter((item) => item && item.enabled !== false)
    .map((item) => ({ id: item.id || "", name: item.name || item.id || "" }));
  select.innerHTML = `<option value="">自动选择</option>`;
  for (const item of options) {
    const option = document.createElement("option");
    option.value = item.id || "";
    option.textContent = item.name || item.id || "";
    select.appendChild(option);
  }
  select.value = selectedId || "";
}

function renderKnowledgeSummary(summary) {
  const box = $("knowledgeSummary");
  if (!box) return;
  const items = [
    ["开头词", summary.opening_phrases || 0],
    ["适用风格", summary.opening_styles || summary.opening_categories || 0],
    ["改文模板", summary.rewrite_profiles || 0],
    ["临时指令", summary.temporary_instructions || 0],
    ["高仿提示词", summary.high_imitation_prompts || 0],
    ["高仿参考", summary.high_imitation_references || 0],
    ["敏感词", summary.sensitive_rules || 0],
    ["字位符号", summary.symbol_rules || 0],
    ["成对补符号", summary.pair_fill_rules || 0],
    ["章节整行", summary.chapter_exact_rules || 0],
    ["章节开头", summary.chapter_inline_rules || 0],
  ];
  box.innerHTML = items.map(([label, value]) => `
    <div class="knowledge-pill">
      <b>${escapeHtml(label)}</b>
      <span>${escapeHtml(value)}</span>
    </div>
  `).join("");
}

function renderAiConfig(appCfg) {
  applyAiSettingsToForm(appCfg.ai || {});
  const selectedId = appCfg.text_model_id || appCfg.textModelId || "";
  const select = $("textModelSelect");
  if (select) {
    select.onchange = () => {
      const modelId = String(select.value || "").trim();
      void persistSelectedTextModel(modelId).catch(error => {
        if ($("textModelStatus")) $("textModelStatus").textContent = `保存文本模型失败：${error.message || "请稍后重试"}`;
      });
    };
  }
  void loadTextModels(selectedId);
}

function renderTextModelReadOnly(modelId) {
  const output = $("textModelReadOnly");
  if (!output) return;
  const id = String(modelId || "").trim();
  const model = (state.textModels || []).find(item => String(item.id || item.modelId || "") === id);
  output.textContent = model?.displayName || model?.name || model?.modelId || model?.id || id || "暂未选择";
}

async function persistSelectedTextModel(modelId) {
  const id = String(modelId || "").trim();
  if (!id || !state.config?.app_config) return null;
  if (state.config.app_config.text_model_id === id && state.__persistedTextModelId === id) return textModelPersistPromise || null;
  state.config.app_config.text_model_id = id;
  state.__persistedTextModelId = id;
  if (textModelPersistPromise) await textModelPersistPromise;
  textModelPersistPromise = api("/api/config", {
    method: "POST",
    body: JSON.stringify({ app_config: { text_model_id: id } }),
  }).then(result => {
    if (result?.config) mergeConfigResponse({ ...state.config, ...result.config });
    renderTextModelReadOnly(id);
    return result;
  }).catch(error => {
    if (state.config?.app_config?.text_model_id === id) state.config.app_config.text_model_id = "";
    state.__persistedTextModelId = "";
    throw error;
  }).finally(() => {
    textModelPersistPromise = null;
  });
  return textModelPersistPromise;
}

async function loadTextModels(selectedId = "") {
  const select = $("textModelSelect");
  const status = $("textModelStatus");
  if (!select) return;
  try {
    const result = await api("/api/models?kind=text");
    const models = Array.isArray(result.models) ? result.models : [];
    state.textModels = models;
    select.innerHTML = "";
    if (!models.length) {
      select.innerHTML = '<option value="">暂无可用文本模型，请先在 API 配置中启用</option>';
      select.value = "";
      if (status) status.textContent = "暂无已启用的文本模型";
      renderTextModelReadOnly("");
      return;
    }
    for (const model of models) {
      const option = document.createElement("option");
      option.value = model.id || model.modelId || "";
      option.textContent = model.displayName || model.name || model.modelId || model.id;
      select.appendChild(option);
    }
    const value = models.some(model => String(model.id || model.modelId) === String(selectedId))
      ? String(selectedId)
      : String(models[0].id || models[0].modelId || "");
    select.value = value;
    renderTextModelReadOnly(value);
    if (value && String(selectedId || "") !== value) {
      void persistSelectedTextModel(value).catch(error => {
        if (status) status.textContent = `保存文本模型失败：${error.message || "请稍后重试"}`;
      });
    }
    if (status) status.textContent = "已连接导航栏 API 配置，执行时使用所选模型";
  } catch (error) {
    select.innerHTML = `<option value="">读取文本模型失败</option>`;
    if (status) status.textContent = `读取模型失败：${error.message || "请稍后重试"}`;
    renderTextModelReadOnly("");
  }
}

function renderPresetControls(appCfg) {
  const presets = appCfg.ai_presets || [];
  const presetSelect = $("presetSelect");
  presetSelect.innerHTML = "";
  if (!presets.length) {
    presetSelect.innerHTML = `<option value="">暂无预设</option>`;
  } else {
    for (const preset of presets) {
      const option = document.createElement("option");
      option.value = preset.id;
      option.textContent = `${preset.name || preset.id} (${preset.model || "未填模型"})`;
      presetSelect.appendChild(option);
    }
  }

  const assignmentOptions = presetOptions();
  renderOptions($("classifierSelect"), assignmentOptions);
  renderOptions($("rewriteSelect"), assignmentOptions);
  if ($("sensitiveFixSelect")) renderOptions($("sensitiveFixSelect"), assignmentOptions);
  const assignments = appCfg.ai_assignments || {};
  $("classifierSelect").value = assignments.classifier || "__current__";
  $("rewriteSelect").value = assignments.rewrite || "__current__";
  if ($("sensitiveFixSelect")) $("sensitiveFixSelect").value = assignments.sensitive_fix || "__current__";
}

function applyAiSettingsToForm(settings) {
  $("aiTimeout").value = settings.timeout_seconds || 180;
  $("aiConcurrency").value = settings.max_concurrency || 6;
  $("aiRetries").value = settings.retry_times ?? 2;
  $("aiStream").checked = Boolean(settings.stream);
  $("aiJsonMode").checked = Boolean(settings.json_mode);
  $("aiMaxTokens").value = settings.max_tokens || 1200;
  $("aiTemperature").value = settings.temperature ?? 0.45;
  $("aiTopP").value = settings.top_p ?? 0.9;
  $("aiPresence").value = settings.presence_penalty ?? 0;
  $("aiFrequency").value = settings.frequency_penalty ?? 0;
  $("aiEnableThinking").checked = Boolean(settings.enable_thinking);
  $("aiDisableThinking").checked = Boolean(settings.disable_thinking);
  $("aiExtraJson").value = settings.extra_body_json || "";
}

function readAiSettingsFromForm() {
  return {
    timeout_seconds: numberValue("aiTimeout", 180),
    max_concurrency: numberValue("aiConcurrency", 6),
    retry_times: numberValue("aiRetries", 2),
    stream: $("aiStream").checked,
    json_mode: $("aiJsonMode").checked,
    max_tokens: numberValue("aiMaxTokens", 1200),
    temperature: numberValue("aiTemperature", 0.45),
    top_p: numberValue("aiTopP", 0.9),
    presence_penalty: numberValue("aiPresence", 0),
    frequency_penalty: numberValue("aiFrequency", 0),
    enable_thinking: $("aiEnableThinking").checked,
    disable_thinking: $("aiDisableThinking").checked,
    extra_body_json: $("aiExtraJson").value.trim(),
  };
}

// 当前填写的 AI 接口不是一次性表单值：保存时同步为一个可见的“当前预设”，
// 并让三个工作用途优先使用它。手动创建的其他预设不受影响。
function syncCurrentAiPreset(cfg) {
  const settings = cfg.ai || {};
  if (!settings.base_url || !settings.api_key || !settings.model) return "";
  const id = "preset_current_auto";
  const preset = {
    ...settings,
    id,
    name: `当前预设（自动同步）· ${settings.model}`,
    auto_managed: true,
  };
  const presets = Array.isArray(cfg.ai_presets) ? [...cfg.ai_presets] : [];
  const index = presets.findIndex((item) => item?.id === id);
  if (index >= 0) presets[index] = preset;
  else presets.unshift(preset);
  cfg.ai_presets = presets;
  cfg.ai_assignments = {
    ...(cfg.ai_assignments || {}),
    classifier: id,
    rewrite: id,
    sensitive_fix: id,
  };
  return id;
}

function updatePlatformHint() {
  const id = $("platformSelect").value;
  const found = (state.config?.platforms || []).find((item) => String(item.id) === String(id));
  $("platformHint").textContent = found ? `平台ID：${found.id}` : "";
}

function renderWebSubmitConfig(settings = {}) {
  if (!$("webAllowResubmit")) return;
  const cfg = ensureWebSubmitConfig();
  Object.assign(cfg, settings || {});
  $("webAllowResubmit").checked = cfg.skip_submitted === false;
  $("webMinTextChars").value = cfg.min_text_chars ?? 0;
  $("webRetryTimes").value = cfg.retry_times ?? 1;
  $("webProfilesJson").value = JSON.stringify(cfg.upload_profiles || [], null, 2);
  $("webProfileBindingsJson").value = JSON.stringify(cfg.profile_bindings || {}, null, 2);
  renderWebDefaultProfileOptions(cfg.upload_profiles || [], cfg.selected_profile || "");
  renderWebVersionProfileBindings(cfg.upload_profiles || [], cfg.profile_bindings || {});
  renderWebOrganizationOptions(cfg.organization_catalog || [], cfg.organization || "1");

  updateResubmitHint();
  renderWebSubmitMode();
  renderWebLoginStatus(cfg);
}

function renderWebOrganizationOptions(catalog, selected) {
  const select = $("webOrganization");
  if (!select) return;
  const items = asArray(catalog);
  const options = items.length ? items.map(item => {
    const id = String(item?.id || "").trim();
    const name = String(item?.name || id).trim();
    return id ? `<option value="${escapeHtml(id)}" ${id === String(selected || "") ? "selected" : ""}>${escapeHtml(name)}（ID ${escapeHtml(id)}）</option>` : "";
  }).join("") : `<option value="${escapeHtml(String(selected || "1"))}">默认组织（ID ${escapeHtml(String(selected || "1"))}）</option>`;
  select.innerHTML = options;
  select.value = String(selected || "1");
}

function renderWebLoginStatus(settings = {}) {
  const box = $("webLoginStatus");
  if (!box) return;
  const username = String(settings.username || "").trim();
  const stateClass = state.webLoginSession === false ? "error" : username && state.webLoginSession ? "ok" : "not-logged-in";
  box.className = `web-login-status ${stateClass}`;
  const label = username && state.webLoginSession ? username : (state.webLoginSession === false ? (username || "登录异常") : "未登录");
  box.innerHTML = `<span class="web-login-dot" aria-hidden="true"></span><span>${escapeHtml(label)}</span>`;
}

async function openWebLoginDialog() {
  let dialog = $("webLoginDialog");
  if (!dialog) {
    dialog = document.createElement("dialog");
    dialog.id = "webLoginDialog";
    dialog.innerHTML = `<form method="dialog" class="login-dialog-form"><h3>登录批量后台</h3><p class="form-hint">登录后处理页会显示账号和状态点，提交任务时自动复用此会话。</p><label>账号<input id="webLoginUsername" autocomplete="username"></label><label>密码<input id="webLoginPassword" type="password" autocomplete="current-password"></label><div class="actions"><button value="cancel">取消</button><button id="webLoginSubmit" value="default" class="primary">保存并验证</button></div><span id="webLoginResult"></span></form>`;
    document.body.appendChild(dialog);
    dialog.addEventListener("pointerdown", event => {
      if (event.target === dialog && !isModalBusy("webLoginDialog")) dialog.close();
    });
    dialog.addEventListener("click", event => {
      if (event.target === dialog && !isModalBusy("webLoginDialog")) dialog.close();
    });
    dialog.addEventListener("cancel", event => {
      if (isModalBusy("webLoginDialog")) event.preventDefault();
    });
    dialog.addEventListener("submit", async (event) => {
      if (event.submitter?.id !== "webLoginSubmit") return;
      event.preventDefault();
      const result = $("webLoginResult"); result.textContent = "验证中...";
      const loginButtons = dialog.querySelectorAll("button");
      loginButtons.forEach(button => { button.disabled = true; });
      setModalBusy("webLoginDialog", true);
      try {
        const username = $("webLoginUsername").value.trim();
        const password = $("webLoginPassword").value;
        const operation = await api("/api/web-submit/operations", {
          method: "POST",
          body: JSON.stringify({ kind: "login", username, password }),
        });
        const completed = await waitWebSubmitOperation(operation.id);
        if (completed.status === "failed") throw new Error(completed.error || "登录验证失败");
        const settings = completed.result?.settings || {};
        state.webLoginSession = true;
        state.config.web_submit = { ...(state.config.web_submit || {}), ...settings, username, password_masked: true };
        renderWebLoginStatus(state.config.web_submit);
        result.textContent = "登录批量后台成功";
        if (state.webLoginSession) setTimeout(() => dialog.close(), 500);
      } catch (error) { state.webLoginSession = false; renderWebLoginStatus(state.config.web_submit || {}); result.textContent = error.message; }
      finally {
        const passwordInput = $("webLoginPassword");
        if (passwordInput) passwordInput.value = "";
        loginButtons.forEach(button => { button.disabled = false; });
        setModalBusy("webLoginDialog", false);
      }
    });
  }
  $("webLoginUsername").value = state.config?.web_submit?.username || "";
  $("webLoginPassword").value = "";
  dialog.showModal();
}

function webSubmitModeFromForm() {
  return "version";
}

function renderWebSubmitMode() {
  $("webVersionConfigSection")?.classList.remove("hidden");
}

function renderWebDefaultProfileOptions(profiles, selectedProfile) {
  const select = $("webDefaultProfile");
  if (!select) return;
  const items = asArray(profiles);
  const saved = String(selectedProfile || "");
  const selected = items.find((profile) => String(profile?.id || "") === saved || String(profile?.name || "") === saved)?.id || "";
  const options = ["<option value=\"\">按书籍自动匹配</option>"];
  for (const profile of items) {
    const id = String(profile?.id || "").trim();
    const name = String(profile?.name || id).trim();
    if (!id) continue;
    options.push(`<option value="${escapeHtml(id)}" ${id === selected ? "selected" : ""}>${escapeHtml(name)}</option>`);
  }
  select.innerHTML = options.join("");
}

function profileIdFromIdentity(profiles, identity) {
  const value = String(identity || "");
  return asArray(profiles).find((profile) => String(profile?.id || "") === value || String(profile?.name || "") === value)?.id || "";
}

function renderWebVersionProfileBindings(profiles, bindings) {
  const fields = { original: "webProfileBindingOriginal", ai1: "webProfileBindingAi1", ai2: "webProfileBindingAi2", ai3: "webProfileBindingAi3", ai4: "webProfileBindingAi4", ai5: "webProfileBindingAi5" };
  const items = asArray(profiles);
  for (const [version, id] of Object.entries(fields)) {
    const select = $(id);
    if (!select) continue;
    const selected = profileIdFromIdentity(items, bindings?.[version]);
    const options = ["<option value=\"\">跟随默认配置档</option>"];
    for (const profile of items) {
      const profileId = String(profile?.id || "").trim();
      const name = String(profile?.name || profileId).trim();
      if (!profileId) continue;
      options.push(`<option value="${escapeHtml(profileId)}" ${profileId === selected ? "selected" : ""}>${escapeHtml(name)}</option>`);
    }
    select.innerHTML = options.join("");
  }
}

function webProfileBindingsFromForm() {
  return {
    original: $("webProfileBindingOriginal")?.value || "",
    ai1: $("webProfileBindingAi1")?.value || "",
    ai2: $("webProfileBindingAi2")?.value || "",
    ai3: $("webProfileBindingAi3")?.value || "",
    ai4: $("webProfileBindingAi4")?.value || "",
    ai5: $("webProfileBindingAi5")?.value || ""
  };
}

function updateResubmitHint() {
  const allow = $("webAllowResubmit")?.checked === true;
  const hint = $("webResubmitHint");
  if (hint) hint.textContent = allow
    ? "已开启：已成功版本会再次上传。"
    : "默认保护：已成功版本会跳过。";
}

function syncFormToWebSubmitConfig() {
  const cfg = clone(ensureWebSubmitConfig());
  // 兼容旧配置字段；是否执行由任务列表的“提交网络”按钮决定。
  cfg.enabled = true;
  // 账号只在登录弹窗中维护，提交设置保存时保留当前凭据。
  if ($("webUsername")) cfg.username = $("webUsername").value.trim();
  if ($("webPassword")) cfg.password = $("webPassword").value.trim();
  cfg.skip_submitted = !$("webAllowResubmit").checked;
  cfg.submit_mode = "version";
  cfg.min_text_chars = numberValue("webMinTextChars", 0);
  cfg.retry_times = numberValue("webRetryTimes", 1);
  cfg.upload_profiles = parseJsonInput("webProfilesJson", []);
  cfg.profile_bindings = webProfileBindingsFromForm();
  $("webProfileBindingsJson").value = JSON.stringify(cfg.profile_bindings);
  cfg.selected_profile = $("webDefaultProfile").value;
  cfg.organization = $("webOrganization")?.value || cfg.organization || "1";
  cfg.submit_versions = selectedProcessVersions();
  return cfg;
}

function setSiteSubmitStatus(text) {
  if ($("siteSubmitStatus")) $("siteSubmitStatus").textContent = text || "";
}

function setVersionConfigStatus(text) {
  if ($("webSubmitSelectionStatus")) $("webSubmitSelectionStatus").textContent = text || "";
}

async function saveWebSubmitConfig(silent = false) {
  if (!silent) setSiteSubmitStatus("网站提交配置保存中...");
  try {
    const result = await api("/api/web-submit/config", {
      method: "POST",
      body: JSON.stringify({ settings: syncFormToWebSubmitConfig() }),
    });
    state.config.web_submit = result.settings || state.config.web_submit;
    renderWebSubmitConfig(state.config.web_submit);
    if (result.tasks) renderTasks(result.tasks || []);
    if (!silent) setSiteSubmitStatus("网站提交配置已保存");
    return result;
  } catch (error) {
    if (!silent) setSiteSubmitStatus(error.message);
    if (silent) throw error;
    return null;
  }
}

async function saveVersionConfigAuthority() {
  clearTimeout(workFormSaveTimer);
  workFormSaveTimer = null;
  syncVersionPromptConfigToForm();
  const formState = collectWorkFormState();
  const appConfig = syncFormToAppConfig();
  appConfig.rewrite = {
    ...(appConfig.rewrite || {}),
    ai_slot_methods: { ...(appConfig.rewrite?.ai_slot_methods || {}), ...processAiMethods() },
  };
  appConfig.work_form = formState;
  const webSubmit = syncFormToWebSubmitConfig();
  delete webSubmit.password;
  appConfig.web_submit = webSubmit;
  const payload = {
    app_config: appConfig,
    platforms: JSON.parse($("platformsText").value),
    styles: JSON.parse($("stylesText").value),
    sensitive: state.config.sensitive || { groups: [] },
    ...(state.config.knowledge_loaded ? { knowledge: state.config.knowledge || {} } : {}),
  };
  const data = await api("/api/config", { method: "POST", body: JSON.stringify(payload) });
  if (!data?.config) throw new Error("版本配置保存失败：服务器没有返回保存结果");
  mergeConfigResponse(data.config);
  try { localStorage.setItem(WORK_FORM_STORAGE_KEY, JSON.stringify(formState)); } catch (_) {}
  renderConfig();
  return data.config;
}

async function confirmWebSubmitSelection() {
  const status = $("webSubmitSelectionStatus");
  const versions = selectedProcessVersions();
  if (!versions.length) {
    if (status) status.textContent = "请至少选择一个文案版本";
    return;
  }
  if (status) status.textContent = "正在保存版本配置...";
  setModalBusy("versionConfigCard", true);
  try {
    await saveVersionConfigAuthority();
    if (status) status.textContent = `已保存：${versions.map(version => version === "original" ? "原文" : version.toUpperCase()).join("、")}；任务会直接使用这些版本。`;
    updateVersionConfigSummary();
    setModalBusy("versionConfigCard", false);
    closeVersionConfigCard();
  } catch (error) {
    if (status) status.textContent = error.message;
  } finally {
    setModalBusy("versionConfigCard", false);
  }
}

function renderWebSubmitOperation(operation) {
  const text = asArray(operation?.steps).map(step => step.message).filter(Boolean).join("\n");
  if (text) {
    setSiteSubmitStatus(text);
    setVersionConfigStatus(text);
    if ($("webLoginResult")) $("webLoginResult").textContent = text;
  }
}

async function waitWebSubmitOperation(id) {
  const deadline = Date.now() + (5 * 60 * 1000);
  for (;;) {
    const operation = await api(`/api/web-submit/operations/${encodeURIComponent(id)}`);
    renderWebSubmitOperation(operation);
    if (operation.status === "done" || operation.status === "failed") return operation;
    if (Date.now() >= deadline) throw new Error("121 后台操作等待超时，请稍后刷新查看状态");
    await new Promise(resolve => setTimeout(resolve, 500));
  }
}

async function syncWebSubmit(kind) {
  const label = kind === "styles" ? "批量风格类型" : kind === "organizations" ? "组织归属" : "批量后台配置";
  setSiteSubmitStatus(`正在同步${label}...`);
  setVersionConfigStatus(`正在同步${label}...`);
  setModalBusy("versionConfigCard", true);
  try {
    await saveWebSubmitConfig(true);
    const operation = await api("/api/web-submit/operations", {
      method: "POST",
      body: JSON.stringify({ kind }),
    });
    const completed = await waitWebSubmitOperation(operation.id);
    if (completed.status === "failed") throw new Error(completed.error || `${label}同步失败`);
    const result = completed.result || {};
    state.config.web_submit = result.settings || state.config.web_submit;
    if (result.styles) state.config.styles = result.styles;
    renderWebSubmitConfig(state.config.web_submit);
    const styleText = result.style_sync
      ? `，AI风格 ${result.style_sync.new_count || 0} 个，新增 ${result.style_sync.added?.length || 0}，删除 ${result.style_sync.removed?.length || 0}`
      : "";
    const configText = kind === "configs" ? `：${result.groups?.length || 0} 个配置档已更新` : "";
    const message = `已同步${label}${configText}${styleText}`;
    setSiteSubmitStatus(message);
    setVersionConfigStatus(message);
  } catch (error) {
    setSiteSubmitStatus(error.message);
    setVersionConfigStatus(error.message);
  } finally {
    setModalBusy("versionConfigCard", false);
  }
}

async function checkWebEnvironment() {
  setSiteSubmitStatus("正在检查内置运行环境...");
  try {
    await saveWebSubmitConfig(true);
    const result = await api("/api/web-submit/environment");
    const rows = asArray(result.checks).map((item) => `
      <div class="site-skipped-row">
        <code>${escapeHtml(item.ok ? "正常" : "异常")}</code>
        <span>${escapeHtml(item.name || "")}</span>
        <span>${escapeHtml(item.detail || "")}</span>
      </div>
    `).join("");
    if ($("siteSubmitGroups")) {
      $("siteSubmitGroups").innerHTML = `
        <div class="site-group-card">
          <div class="site-group-head"><b>环境自检</b><span class="${result.ok ? "status-ok" : "status-error"}">${result.ok ? "通过" : "存在问题"}</span></div>
          <div class="site-skipped-list">${rows}</div>
        </div>
      `;
    }
    setSiteSubmitStatus(result.ok ? "环境自检通过" : "环境自检发现问题");
  } catch (error) {
    setSiteSubmitStatus(error.message);
  }
}

async function testVisibleWebFlow() {
  const button = $("testVisibleWebBtn");
  const originalLabel = button.textContent;
  button.disabled = true;
  button.textContent = "验证中…";
  setSiteSubmitStatus("正在验证 121 登录会话...");
  try {
    await saveWebSubmitConfig(true);
    const result = await novelFetchPlatformApi("/api/batch-rewrite/web-submit/test-visible", {
      method: "POST",
      body: JSON.stringify(webSubmitRequestPayload(state.selectedIds.size ? "selected" : "all", true)),
    });
    const checks = result.result?.checks || {};
    const checkRows = Object.keys(checks).map((key) => `
      <div class="site-skipped-row">
        <code>${escapeHtml(key)}</code>
        <span>${escapeHtml(String(checks[key]))}</span>
      </div>
    `).join("");
    if ($("siteSubmitGroups")) {
      $("siteSubmitGroups").innerHTML = `
        <div class="site-group-card">
          <div class="site-group-head"><b>121 登录会话验证</b><span class="${result.ok ? "status-ok" : "status-error"}">${result.ok ? "完成" : "失败"}</span></div>
          <div class="site-skipped-list">${checkRows || "暂无检查项"}</div>
          <pre class="site-output">${escapeHtml(asArray(result.output).join("\n"))}</pre>
        </div>
      `;
    }
    setSiteSubmitStatus(result.ok ? "121 登录会话有效；此检查不会上传文件或确认生成任务" : "121 登录会话验证失败");
  } catch (error) {
    setSiteSubmitStatus(error.message);
    if ($("siteSubmitGroups")) {
      $("siteSubmitGroups").innerHTML = `
        <div class="site-group-card">
          <div class="site-group-head"><b>121 登录会话验证</b><span class="status-error">未通过</span></div>
          <div class="site-skipped-list">
            <div class="site-skipped-row"><code>需要处理</code><span>${escapeHtml(error.message || "验证失败")}</span></div>
          </div>
        </div>
      `;
    }
  } finally {
    button.disabled = false;
    button.textContent = originalLabel;
  }
}

function clearWebPreview() {
  if ($("siteSubmitGroups")) $("siteSubmitGroups").innerHTML = "";
  setSiteSubmitStatus("预览已隐藏，不影响任务和提交记录");
}

function renderWebSubmitHistory(records = []) {
  const box = $("siteSubmitHistory");
  if (!box) return;
  const rows = asArray(records).map((item) => {
    const allocation = item.material_allocation || {};
    const allocationText = allocation.jieyaNum === undefined ? "" : `解压 ${allocation.jieyaNum} / 滚屏 ${allocation.gunpingNum}`;
    return `<div class="site-history-row">
      <code title="${escapeHtml(item.book_id || "")}">${escapeHtml(item.book_id || "")}</code>
      <span class="site-history-name" title="${escapeHtml(item.book_name || "")}">${escapeHtml(item.book_name || "-")}</span>
      <span>${escapeHtml(item.version || "-")}</span>
      <span class="${statusClass(item.status)}">${escapeHtml(item.status || "-")}</span>
      <span>${escapeHtml(allocationText || "-")}</span>
      <time title="${escapeHtml(item.time || "")}">${escapeHtml(item.time || "-")}</time>
      <span class="site-history-error" title="${escapeHtml(item.result || item.error || "")}">${escapeHtml(item.result || item.error || "-")}</span>
    </div>`;
  }).join("");
  box.innerHTML = rows ? `<div class="site-history-table">
    <div class="site-history-row site-history-head"><span>书籍 ID</span><span>书名</span><span>版本</span><span>状态</span><span>素材</span><span>提交时间</span><span>结果 / 错误</span></div>
    ${rows}
  </div>` : `<div class="log-row">暂无实际网站提交记录。</div>`;
}

async function loadWebSubmitHistory() {
  try {
    const data = await api("/api/web-submit/history");
    renderWebSubmitHistory(data.records || []);
  } catch (error) {
    const box = $("siteSubmitHistory");
    if (box) box.innerHTML = `<div class="log-row">${escapeHtml(error.message || "提交历史读取失败")}</div>`;
  }
}

function webSubmitRequestPayload(mode, force = false, explicitIds = null) {
  const ids = mode === "selected" ? (Array.isArray(explicitIds) ? explicitIds.map(id => String(id || "").trim()).filter(Boolean) : selectedTaskIds()) : [];
  return {
    mode,
    ids,
    force,
    grouped: true,
  };
}

function showWebSubmitSelectionRequired() {
  if (!$('siteSubmitGroups')) return;
  $("siteSubmitGroups").innerHTML = `
    <div class="site-group-card">
      <div class="site-group-head"><b>还没有选中任务</b><span class="status-warn">需要先选择</span></div>
      <div class="site-skipped-list">
        <div class="site-skipped-row"><code>下一步</code><span>切换到“任务列表”，勾选每条任务最左侧的复选框；再回到这里预览或提交选中任务。</span></div>
      </div>
    </div>
  `;
}

async function previewWebSubmit(mode) {
  if (mode === "selected" && !selectedTaskIds().length) {
    setSiteSubmitStatus("先选择任务");
    showWebSubmitSelectionRequired();
    return;
  }
  setSiteSubmitStatus("正在预览网站提交分组...");
  try {
    await saveWebSubmitConfig(true);
    const result = await api("/api/web-submit/preview", {
      method: "POST",
      body: JSON.stringify(webSubmitRequestPayload(mode, false)),
    });
    renderWebSubmitGroups(result);
    setSiteSubmitStatus(`预览完成：${result.groups?.length || 0} 个提交组，跳过 ${result.skipped?.length || 0} 条`);
  } catch (error) {
    setSiteSubmitStatus(error.message);
  }
}

async function submitWebSubmit(mode, explicitIds = null) {
  const selectedIdsForSubmit = mode === "selected" && Array.isArray(explicitIds) ? explicitIds.map(id => String(id || "").trim()).filter(Boolean) : selectedTaskIds();
  if (mode === "selected" && !selectedIdsForSubmit.length) {
    setSiteSubmitStatus("先选择任务");
    setBatchStatus("先选择任务");
    showWebSubmitSelectionRequired();
    return;
  }
  const label = mode === "all" ? "全部任务" : (mode === "failed" ? "提交失败任务" : `${selectedIdsForSubmit.length} 个选中任务`);
  if (!confirm(`确认提交${label}到网站？系统会按平台、男女频、风格类型自动分组上传。`)) return;
  setBatchStatus(`排队中：${label}`);
  setSiteSubmitStatus(`正在提交${label}...`);
  let polling = false;
  try {
    await saveWebSubmitConfig(true);
    setBatchStatus(`上传中：${label}`);
    const submitPromise = api("/api/web-submit/submit", {
      method: "POST",
      body: JSON.stringify(webSubmitRequestPayload(mode, mode === "failed", selectedIdsForSubmit)),
    });
    polling = true;
    const poll = (async () => {
      while (polling) {
        await sleep(700);
        if (!polling) break;
        try { await loadTasks(); } catch (_) {}
      }
    })();
    const result = await submitPromise;
    polling = false;
    await poll;
    renderWebSubmitGroups(result);
    renderTasks(result.tasks || state.tasks);
    await loadWebSubmitHistory();
    if (Number(result.failed_groups) > 0) reportBatchIssue("result", "/api/web-submit/submit", `网站提交失败 ${result.failed_groups} 组`);
    const summary = `提交完成：已确认 ${result.success_groups || 0} 组，121 待确认 ${result.accepted_groups || 0} 组，失败 ${result.failed_groups || 0} 组`;
    setSiteSubmitStatus(summary);
    setBatchStatus(summary);
  } catch (error) {
    polling = false;
    setSiteSubmitStatus(error.message);
    setBatchStatus(`提交失败：${error.message}`);
    await loadTasks().catch(() => {});
  }
}

function siteSubmitText(task) {
  const queued = asArray(task.site_submit_queued_versions);
  const uploading = asArray(task.site_submit_uploading_versions);
  const done = asArray(task.site_submit_done_versions);
  const accepted = asArray(task.site_submit_accepted_versions);
  const failed = asArray(task.site_submit_failed_versions);
  // 同一版本曾失败但后来已成功提交时，成功结果才是当前状态；
  // 失败记录仍保留在“记录”和“问题日志”中供追溯。
  const parts = [];
  if (done.length) parts.push(`已提交：${done.map(displayVersionLabel).join(",")}`);
  if (uploading.length) parts.push(`上传中：${uploading.map(displayVersionLabel).join(",")}`);
  if (queued.length) parts.push(`排队中：${queued.map(displayVersionLabel).join(",")}`);
  if (accepted.length) parts.push(`待确认：${accepted.map(displayVersionLabel).join(",")}`);
  if (failed.length) parts.push(`失败：${failed.map(displayVersionLabel).join(",")}`);
  if (parts.length) return parts.join("；");
  if (task.site_submit_status) return task.site_submit_status;
  return "未提交";
}

function displayVersionLabel(version) {
  const value = String(version || "").trim().toLowerCase();
  if (value === "original") return "原文";
  if (/^ai[1-5]$/.test(value)) return value.toUpperCase();
  return String(version || "").trim();
}

function groupCardHtml(group) {
  const summary = group.summary || {};
  const advanced = group.advanced || {};
  const items = asArray(group.items);
  const status = group.status || "";
  const itemRows = items.slice(0, 24).map((item) => {
    const itemAdvanced = item.advanced || advanced;
    return `
    <div class="site-file-row">
      <code>${escapeHtml(item.id || "")}</code>
      <span>${escapeHtml(item.version || group.version || "")}</span>
      <span>${escapeHtml(`解压 ${itemAdvanced.jieyaNum ?? 0} / 滚屏 ${itemAdvanced.gunpingNum ?? 0}`)}</span>
      <span>${escapeHtml(item.size ? `${item.size}字节` : "")}</span>
    </div>
  `;
  }).join("");
  return `
    <div class="site-group-card">
      <div class="site-group-head">
        <b>${escapeHtml(group.group_id || "提交组")}</b>
        <span class="${statusClass(status)}">${escapeHtml(status || "待提交")}</span>
      </div>
      <div class="meta-grid compact">
        ${metaItem("平台", `${summary.platform_name || ""} ${summary.platform_id || ""}`)}
        ${metaItem("男女频", `${summary.gender || ""} ${summary.gender_value || ""}`)}
        ${metaItem("风格", `${summary.style || ""} ${summary.style_value || ""}`)}
        ${metaItem("版本", group.version || summary.version || "")}
        ${metaItem("配置档", summary.profile_name || "默认上传参数")}
        ${metaItem("任务数", String(items.length || group.count || 0))}
        ${metaItem("时长", Number(advanced.tl5) === 1 ? "限制" : "不限制")}
        ${metaItem("每本书解压总量", `${advanced.jieyaNum ?? 4} 个 / ${advanced.jieyaSpeed ?? 1.7}x`)}
        ${metaItem("每本书滚屏总量", `${advanced.gunpingNum ?? 4} 个 / ${advanced.gunpingSpeed ?? 1}x`)}
        ${metaItem("错误", group.error || "")}
      </div>
      <div class="site-file-list">${itemRows || "暂无文件"}</div>
      ${group.helper_output ? `<pre class="site-output">${escapeHtml(asArray(group.helper_output).join("\n"))}</pre>` : ""}
    </div>
  `;
}

function renderWebSubmitGroups(data = {}) {
  if (!$("siteSubmitGroups")) return;
  const groups = asArray(data.groups || data.results);
  const skipped = asArray(data.skipped);
  const skippedRows = skipped.slice(0, 80).map((item) => `
    <div class="site-skipped-row">
      <code>${escapeHtml(item.id || "")}</code>
      <span>${escapeHtml(item.version || "")}</span>
      <span>${escapeHtml(item.status || "")}</span>
      <span>${escapeHtml(item.error || "")}</span>
    </div>
  `).join("");
  $("siteSubmitGroups").innerHTML = `
    ${groups.map(groupCardHtml).join("") || `<div class="detail-empty small">暂无可提交分组。请先选择任务、确认AI文案文件和网站风格映射。</div>`}
    ${skippedRows ? `
      <div class="site-group-card">
        <div class="site-group-head"><b>跳过/待配置</b><span>${escapeHtml(String(skipped.length))} 条</span></div>
        <div class="site-skipped-list">${skippedRows}</div>
      </div>
    ` : ""}
  `;
}

function statusClass(value) {
  const text = String(value || "");
  if (text.includes("failed") || text.includes("失败")) return "status-error";
  if (text.includes("waiting") || text.includes("等待") || text.includes("pending") || text.includes("待确认") || text.includes("排队中") || text.includes("上传中") || text.includes("accepted") || text.includes("partial") || text.includes("submitting") || text.includes("提交中") || text.includes("generating") || text.includes("dry_run")) return "status-warn";
  if (text.includes("done") || text.includes("完成") || text.includes("classified") || text.includes("submitted") || text.includes("已提交")) return "status-ok";
  return "";
}

function taskIsProblem(task) {
  const text = [
    task.status,
    task.original_status,
    task.ai_status,
    task.classify_status,
    task.error,
  ].filter(Boolean).join(" ").toLowerCase();
  return text.includes("failed") || text.includes("失败") || text.includes("waiting") || text.includes("等待") || text.includes("partial");
}

function updateSelectedCount() {
  const count = state.selectedIds.size;
  $("selectedCount").textContent = `已选 ${count} 个`;
  const allBox = $("selectAllTasks");
  if (allBox) {
    allBox.checked = Boolean(state.tasks.length) && state.tasks.every((task) => state.selectedIds.has(String(task.id || "")));
    allBox.indeterminate = count > 0 && !allBox.checked;
  }
}

function originalStatusText(task) {
  if (task.original_status === "done") {
    const raw = Number(task.original_raw_chars ?? task.originalRawChars ?? 0);
    const maxTxt = Number(task.max_txt ?? task.maxTxt ?? 0);
    const processed = raw > 0 && maxTxt > 0 ? Math.min(maxTxt, raw) : Number(task.original_chars ?? task.originalChars ?? 0);
    const text = raw > 0 ? `${processed}/${raw}` : `${processed}字`;
    return task.original_refresh_error ? `${text}（刷新失败，已保留）` : text;
  }
  if (task.original_status === "process_failed") {
    return `已抓取${task.original_raw_chars ? ` ${task.original_raw_chars}字` : ""} / 规则失败`;
  }
  if (task.original_raw_status === "done" && String(task.original_status || "").includes("failed")) {
    return `已抓取${task.original_raw_chars ? ` ${task.original_raw_chars}字` : ""} / 规则失败`;
  }
  if (task.original_raw_status === "done") {
    return `已抓取${task.original_raw_chars ? ` ${task.original_raw_chars}字` : ""} / 待处理`;
  }
  return taskStatusText(task.original_status, "未抓取");
}

function taskStatusText(value, fallback = "") {
  const text = String(value || "").trim();
  const labels = {
    created: "待处理", queued: "排队中", running: "正在执行中…", processing: "正在执行中…", input_ready: "分类信息已就绪", classified: "已完成判断", classifying: "AI判断中…", classify_failed: "判断失败",
    original: "原文", fetching: "正在抓取", fetched: "已抓取", done: "已完成", original_processing: "原文处理中", original_done: "原文已就绪", original_failed: "原文抓取失败",
    ai_processing: "AI文案处理中",
    generating: "正在生成AI文案…", generated: "已生成", ai_done: "AI文案已生成", ai_partial: "AI文案部分完成", ai_failed: "AI生成失败", process_failed: "处理失败",
    waiting_ai_config: "等待 AI 配置", waiting_original: "等待原文", waiting_classifier_config: "等待分类模型配置",
    uploading: "上传中", submitted: "已提交", accepted_pending: "已接收，待确认", failed: "失败", waiting_config: "等待配置", skipped: "已跳过", interrupted: "已中断"
  };
  return labels[text] || text || fallback;
}

function classifyStatusDisplay(task = {}) {
  const label = taskStatusText(task.classify_status, task.classifier_model ? "已完成判断" : "待判断");
  const reason = String(task.classify_error || task.classifyError || "").trim();
  if (!reason) return escapeHtml(label);
  return `<span class="task-status-main">${escapeHtml(label)}</span><small class="task-status-reason">分类错误：${escapeHtml(reason)}</small>`;
}

function classifyDetailText(meta = {}) {
  const label = taskStatusText(meta.classify_status, "待判断");
  const reason = String(meta.classify_error || meta.classifyError || "").trim();
  return reason ? `${label}（分类错误：${reason}）` : label;
}

function aiStatusDisplay(task = {}, aiText = '') {
  const reason = String(task.ai_error || task.aiError || '').trim();
  if (task.ai_status === 'generating' || task.ai_status === 'ai_processing') return `<span class="task-spinner" aria-hidden="true"></span>${escapeHtml(aiText)}`;
  if (!reason) return escapeHtml(aiText);
  return `<span class="task-status-main">${escapeHtml(aiText)}</span><small class="task-status-reason">AI错误：${escapeHtml(reason)}</small>`;
}

function aiCopyStatusText(task = {}, selectedAi = [], generatedAi = []) {
  const status = String(task.ai_status || task.aiStatus || '').trim().toLowerCase();
  const current = String(task.ai_current_version || task.aiCurrentVersion || '').trim().toUpperCase();
  const selectedCount = selectedAi.length;
  const generatedCount = generatedAi.filter(version => !selectedAi.length || selectedAi.includes(version)).length;
  const count = selectedCount ? `${generatedCount}/${selectedCount}` : (generatedCount ? `${generatedCount}/${generatedCount}` : '');
  const reason = String(task.ai_error || task.aiError || '').trim();
  const reasonText = reason ? normalizeBatchErrorMessage(reason, task.ai_upstream_code || task.aiUpstreamCode) : '';
  if (status === 'generating' || status === 'ai_processing') return `生成中${current ? `（${current}）` : ''}`;
  if (status === 'waiting_original') return '等待原文';
  if (status === 'waiting_ai_config') return '等待 AI 配置';
  if (status === 'failed' || status === 'ai_failed') return `失败${current ? `（${current}）` : ''}${reasonText ? `：${reasonText.slice(0, 120)}` : ''}`;
  if (status === 'partial' || status === 'ai_partial') return `部分完成${count ? `（${count}）` : ''}${reasonText ? `：${reasonText.slice(0, 120)}` : ''}`;
  if (status === 'done' || status === 'ai_done') return `已生成${count ? `（${count}）` : ''}`;
  if (generatedCount) return `已生成${count}`;
  if (selectedCount) return `待生成（0/${selectedCount}）`;
  return '未生成';
}

function taskOverallStatusText(task = {}) {
  const aiStatus = String(task.ai_status || task.aiStatus || '').trim().toLowerCase();
  if (aiStatus === 'failed' || aiStatus === 'ai_failed') return 'AI文案失败';
  if (aiStatus === 'partial' || aiStatus === 'ai_partial') return 'AI文案部分完成';
  if (aiStatus === 'done' || aiStatus === 'ai_done') return 'AI文案已完成';
  if (aiStatus === 'waiting_original') return '等待原文';
  if (aiStatus === 'waiting_ai_config') return '等待 AI 配置';
  return taskStatusText(task.status, '待处理');
}

function taskDateKey(task) {
  const value = task.batch_created_at || task.batchCreatedAt || task.created_at || task.createdAt || task.updated_at || task.updatedAt || "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

function dateKeyFromValue(value) {
  const date = new Date(value || "");
  if (Number.isNaN(date.getTime())) return "";
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

function todayDateKey() {
  const date = new Date();
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

function shiftTaskDateKey(value, offset) {
  const match = String(value || "").match(/^(\d{4})-(\d{2})-(\d{2})$/);
  if (!match) return todayDateKey();
  const date = new Date(Date.UTC(Number(match[1]), Number(match[2]) - 1, Number(match[3])));
  date.setUTCDate(date.getUTCDate() + Number(offset || 0));
  return date.toISOString().slice(0, 10);
}

function setCurrentBatchFromResult(result = {}, timestamp = "") {
  const ids = Array.isArray(result.current_batch_ids)
    ? result.current_batch_ids.map(id => String(id || "")).filter(Boolean)
    : [];
  const batchDate = dateKeyFromValue(timestamp) || todayDateKey();
  state.currentBatchDate = batchDate === todayDateKey() ? batchDate : todayDateKey();
  state.currentBatchIds = batchDate === todayDateKey() ? new Set(ids) : new Set();
  state.viewMode = "current";
  state.taskDate = state.currentBatchDate;
}

function currentTaskList(tasks) {
  const list = Array.isArray(tasks) ? tasks : [];
  const selectedDate = state.taskDate || todayDateKey();
  if (state.viewMode === "current" && state.currentBatchIds.size) {
    return list.filter(task => state.currentBatchIds.has(String(task.id || task.book_id || "")));
  }
  return list.filter(task => taskDateKey(task) === selectedDate);
}

function taskIdentity(task) {
  return String(task?.id || task?.book_id || task?.bookId || '').trim();
}

function mergeTasksKeepingIds(incomingTasks, preserveIds = []) {
  const incoming = Array.isArray(incomingTasks) ? incomingTasks : [];
  const wanted = new Set((Array.isArray(preserveIds) ? preserveIds : [preserveIds]).map(id => String(id || '').trim()).filter(Boolean));
  if (!wanted.size || !Array.isArray(state.allTasks) || !state.allTasks.length) return incoming;
  const incomingIds = new Set(incoming.map(taskIdentity));
  const missing = state.allTasks.filter(task => {
    const id = taskIdentity(task);
    return wanted.has(id) && id && !incomingIds.has(id);
  });
  const merged = incoming.slice();
  for (const task of missing) {
    const originalIndex = state.allTasks.indexOf(task);
    merged.splice(Math.min(Math.max(0, originalIndex), merged.length), 0, task);
  }
  return merged;
}

function renderTasks(tasks, options = {}) {
  const selectedDate = state.taskDate || todayDateKey();
  state.taskDate = selectedDate;
  if ($("taskDateFilter")) $("taskDateFilter").value = selectedDate;
  const list = Array.isArray(tasks) ? tasks : [];
  state.tasks = options.preserveVisible ? list : currentTaskList(list);
  const visibleIds = new Set(state.tasks.map((task) => String(task.id || "")));
  state.selectedIds = new Set([...state.selectedIds].filter((id) => visibleIds.has(id)));
  const body = $("tasksBody");
  body.innerHTML = "";
  if (!state.tasks.length) {
    body.innerHTML = `<tr><td colspan="12">${selectedDate === todayDateKey() ? "今日暂无任务" : `${selectedDate} 暂无任务`}</td></tr>`;
    updateSelectedCount();
    return;
  }
  for (const task of state.tasks) {
    const id = String(task.id || "");
    const sensitiveRunning = state.sensitiveProcessingIds.has(id);
    const aiBusy = state.aiProcessingIds.has(id);
    const sensitiveHits = Number(task.sensitive_hit_count || 0);
    const sensitiveFixed = Number(task.sensitive_fixed_count || 0);
    const sensitiveFailed = Number(task.sensitive_failed_count || 0);
    const sensitiveLabel = sensitiveRunning
      ? `<span class="task-spinner" aria-hidden="true"></span>敏感词执行中…`
      : sensitiveHits > 0
        ? `敏感日志 ${sensitiveHits}/${sensitiveFixed}${sensitiveFailed ? `/${sensitiveFailed}` : ""}`
        : "无敏感词命中";
    const tr = document.createElement("tr");
    const originalText = originalStatusText(task);
    const selectedAi = asArray(task.selected_versions).filter(version => /^ai[1-5]$/.test(version));
    const generatedAi = asArray(task.ai_files);
    const aiTask = aiBusy ? { ...task, ai_status: 'generating', ai_current_version: task.ai_current_version || selectedAi[0] || 'ai1' } : task;
    const aiText = aiCopyStatusText(aiTask, selectedAi, generatedAi);
    const siteText = siteSubmitText(task);
    tr.innerHTML = `
      <td><input class="task-check" type="checkbox" data-id="${escapeHtml(id)}" ${state.selectedIds.has(id) ? "checked" : ""} /></td>
      <td class="id-cell"><button class="task-id-link" data-action="detail" data-id="${escapeHtml(id)}" title="查看任务详情">${escapeHtml(id)}</button></td>
      <td>${escapeHtml(task.book_name || "")}</td>
      <td>${escapeHtml(task.platform_name || "")}</td>
      <td>${escapeHtml(task.style || "")}</td>
      <td>${escapeHtml(task.gender || "")}</td>
      <td class="${statusClass(task.classify_status)}" title="${escapeHtml(task.classify_error || task.classifyError || "")}">${classifyStatusDisplay(task)}</td>
      <td class="${statusClass(task.original_status)}">${escapeHtml(originalText)}${task.original_error ? `<div class="task-error-detail" title="${escapeHtml(task.original_error)}">${escapeHtml(task.original_error)}</div>` : ""}</td>
      <td class="${statusClass(aiTask.ai_status)}" title="${escapeHtml(task.ai_error || task.aiError || '')}">${aiStatusDisplay(aiTask, aiText)}</td>
      <td class="${statusClass(siteText)}">${escapeHtml(siteText)}</td>
      <td class="${statusClass(taskOverallStatusText(task))}">${escapeHtml(taskOverallStatusText(task))}</td>
      <td class="task-actions">
        <button data-action="detail" data-id="${escapeHtml(id)}">查看</button>
        <button data-action="fetch" data-id="${escapeHtml(id)}">抓原文</button>
        <button data-action="ai" data-id="${escapeHtml(id)}" ${aiBusy ? "disabled" : ""} aria-busy="${aiBusy ? "true" : "false"}">${aiBusy ? "生成中…" : "生成AI文案"}</button>
        <button data-action="sensitive" data-id="${escapeHtml(id)}" ${sensitiveRunning ? "disabled" : ""}>${sensitiveLabel}</button>
        <button data-action="siteLog" data-id="${escapeHtml(id)}">提交日志</button>
      </td>
    `;
    body.appendChild(tr);
  }
  updateSelectedCount();
}

function renderDetail(data) {
  const rawMeta = data.meta || {};
  const meta = {
    ...rawMeta,
    book_id: rawMeta.book_id || rawMeta.bookId || rawMeta.id || "",
    book_name: rawMeta.book_name || rawMeta.bookName || "",
    platform_name: rawMeta.platform_name || rawMeta.platformName || "",
    platform_id: rawMeta.platform_id || rawMeta.platformId || "",
    original_error: rawMeta.original_error || rawMeta.originalErrorMessage || rawMeta.error || "",
    original_error_code: rawMeta.original_error_code || rawMeta.originalErrorCode || "",
    original_upstream_code: rawMeta.original_upstream_code ?? rawMeta.originalUpstreamCode ?? "",
    classify_error: rawMeta.classify_error || rawMeta.classifyError || "",
    ai_status: rawMeta.ai_status || rawMeta.aiStatus || "",
    ai_error: rawMeta.ai_error || rawMeta.aiError || "",
    ai_current_version: rawMeta.ai_current_version || rawMeta.aiCurrentVersion || "",
    ai_generated_count: rawMeta.ai_generated_count ?? rawMeta.aiGeneratedCount ?? 0,
    ai_last_attempt_error: rawMeta.ai_last_attempt_error || rawMeta.aiLastAttemptError || "",
    site_submit_error: rawMeta.site_submit_error || rawMeta.siteSubmitError || "",
    created_at: rawMeta.created_at || rawMeta.createdAt || "",
    updated_at: rawMeta.updated_at || rawMeta.updatedAt || "",
    error: rawMeta.error || "",
  };
  const detailId = meta.book_id || meta.id || "";
  const detailAiBusy = state.aiProcessingIds.has(String(detailId)) || meta.ai_status === "generating" || meta.ai_status === "ai_processing";
  $("detailTitle").textContent = detailId || "详情";
  const knowledgeHistory = meta.rewrite_knowledge_history || [];
  const sensitiveFixed = data.sensitive_fixed || {};
  const sensitiveItems = asArray(sensitiveFixed.items).filter((item) => item && item.status === "done");
  const sensitiveRows = sensitiveItems.map((item) => `
    <div class="meta-item">
      <b>${escapeHtml(`敏感词${item.hit_index || ""}`)}</b>
      <span>${escapeHtml(item.keyword || "")}</span>
    </div>
  `).join("");
  const aiBlocks = (data.ai_texts || []).map((item) => `
    <div class="text-block">
      <h3>${escapeHtml(item.name)}</h3>
      <pre>${escapeHtml(item.text || "")}</pre>
    </div>
  `).join("");
  const knowledge = meta.rewrite_knowledge || {};
  const knowledgeRows = knowledgeHistory.map((item) => `
    <div class="meta-item">
      <b>${escapeHtml(`AI${item.ai_index || ""}`)}</b>
      <span>${escapeHtml(item.strategy_name || item.strategy || "")}${item.opening_phrase_name ? ` / ${escapeHtml(item.opening_phrase_name)}` : ""}${item.high_imitation_reference_name ? ` / ${escapeHtml(item.high_imitation_reference_name)}` : ""}</span>
    </div>
  `).join("");
  $("detail").className = "detail-box";
  $("detail").innerHTML = `
    <div class="meta-grid">
      ${metaItem("ID", meta.book_id || meta.id)}
      ${metaItem("书名", meta.book_name)}
      ${metaItem("平台", `${meta.platform_name || ""} ${meta.platform_id || ""}`)}
      ${metaItem("风格", meta.style)}
      ${metaItem("男女频", meta.gender)}
      ${metaItem("AI判断", classifyDetailText(meta))}
      ${metaItem("分类模型（历史分类）", meta.classifier_model)}
      ${metaItem("状态", taskStatusText(meta.status))}
      ${metaItem("原文失败原因", meta.original_error || "")}
      ${metaItem("原文错误码", meta.original_error_code || "")}
      ${metaItem("上游错误码", meta.original_upstream_code ?? "")}
      ${metaItem("原文刷新失败原因", meta.original_refresh_error || "")}
      ${metaItem("原文刷新错误码", meta.original_refresh_error_code || "")}
      ${metaItem("原文刷新时间", meta.original_refresh_at || "")}
      ${metaItem("原文字数", meta.original_chars || 0)}
      ${metaItem("敏感词处理", `${meta.sensitive_mode || ""} ${meta.sensitive_status || ""}`)}
      ${metaItem("敏感词模型（本次执行）", meta.sensitive_model)}
      ${metaItem("命中/修复", `${meta.sensitive_hit_count || 0} / ${meta.sensitive_fixed_count || 0}`)}
      ${metaItem("AI状态", taskStatusText(meta.ai_status))}
      ${metaItem("AI失败原因", meta.ai_error || "")}
      ${metaItem("AI当前版本", meta.ai_current_version || "")}
      ${metaItem("AI已生成数量", meta.ai_generated_count || 0)}
      ${metaItem("AI尝试次数", meta.ai_last_attempt_count || 0)}
      ${metaItem("AI最后执行时间", meta.ai_last_attempt_at || "")}
      ${metaItem("AI最近失败", meta.ai_last_attempt_error || "无")}
      ${metaItem("分类失败原因", meta.classify_error || "")}
      ${metaItem("网站提交失败原因", meta.site_submit_error || "")}
      ${metaItem("任务错误", meta.error || "")}
      ${metaItem("创建时间", meta.created_at || "")}
      ${metaItem("最后更新时间", meta.updated_at || "")}
      ${metaItem("改文模型", meta.rewrite_model)}
      ${metaItem("改文方案", knowledge.strategy_name || knowledge.strategy || "")}
      ${metaItem("改文模板", knowledge.rewrite_template_name || "")}
      ${metaItem("开头词", knowledge.opening_phrase_name || "")}
      ${metaItem("高仿参考", knowledge.high_imitation_reference_name || "")}
    </div>
    ${knowledgeRows ? `<div class="meta-grid">${knowledgeRows}</div>` : ""}
    ${sensitiveRows ? `<div class="meta-grid">${sensitiveRows}</div>` : ""}
    <div class="actions">
      <button id="detailPrevBtn" ${adjacentTaskId(-1) ? "" : "disabled"}>上一条</button>
      <button id="detailNextBtn" ${adjacentTaskId(1) ? "" : "disabled"}>下一条</button>
      <button id="detailCloseBtn" type="button">关闭</button>
      <button id="detailRefreshBtn" type="button">刷新详情</button>
      <button id="detailFetchBtn">重新抓原文</button>
      <button id="detailAiBtn" ${detailAiBusy ? "disabled" : ""} aria-busy="${detailAiBusy ? "true" : "false"}">${detailAiBusy ? "生成中…" : "生成AI文案"}</button>
      <button id="detailTraceBtn">规则追踪</button>
      <button id="detailSensitiveBtn">重跑敏感词（当前文案）</button>
      ${data.has_original_raw ? `<button id="detailRestoreBtn">从备份恢复原文</button>` : ""}
    </div>
    <div class="text-block">
      <h3>处理后原文</h3>
      <pre>${escapeHtml(data.original || "")}</pre>
    </div>
    ${aiBlocks}
    <div class="text-block">
      <h3>最近处理日志</h3>
      <div class="logs-list">${asArray(data.logs).slice(-20).reverse().map(item => {
        const event = item?.event || "事件";
        const time = item?.time || "";
        const payload = item?.data && typeof item.data === "object" ? item.data : {};
        const reason = payload.error || payload.message || payload.reason || payload.status || "";
        return `<div class="log-row"><b>${escapeHtml(time)}</b> ${escapeHtml(event)}${reason ? `：${escapeHtml(reason)}` : ""}</div>`;
      }).join("") || `<div class="log-row">暂无处理日志</div>`}</div>
    </div>
  `;
  bindDetailCloseControl();
  $("detailPrevBtn").onclick = () => { const id = adjacentTaskId(-1); if (id) showTask(id); };
  $("detailNextBtn").onclick = () => { const id = adjacentTaskId(1); if (id) showTask(id); };
  $("detailFetchBtn").onclick = () => refetchTask(meta.book_id || meta.id);
  $("detailRefreshBtn").onclick = () => { void refreshTaskDetail(meta.book_id || meta.id); };
  $("detailAiBtn").onclick = () => generateAi(meta.book_id || meta.id);
  $("detailTraceBtn").onclick = () => showRulesTrace(meta.book_id || meta.id);
  $("detailSensitiveBtn").onclick = () => reprocessSensitive([meta.book_id || meta.id], false);
  const restoreBtn = $("detailRestoreBtn");
  if (restoreBtn) restoreBtn.onclick = () => restoreOriginal(meta.book_id || meta.id);
}

function renderSensitiveLog(data) {
  const meta = data.meta || {};
  const hitsData = data.sensitive_hits || {};
  const fixedData = data.sensitive_fixed || {};
  const hits = asArray(hitsData.hits);
  const fixedItems = asArray(fixedData.items);
  const fixedByIndex = new Map(fixedItems.map((item) => [String(item.hit_index || ""), item]));
  const rows = hits.map((hit) => {
    const fixed = fixedByIndex.get(String(hit.hit_index || "")) || {};
    const status = fixed.status || fixedData.status || hitsData.status || "";
    return `
      <div class="sensitive-log-card">
        <div class="sensitive-log-head">
          <b>#${escapeHtml(hit.hit_index || "")} ${escapeHtml(hit.keyword || "")}</b>
          <span class="${statusClass(status)}">${escapeHtml(status || "命中")}</span>
        </div>
        <div class="meta-grid compact">
          ${metaItem("命中位置", `${hit.start ?? ""} - ${hit.end ?? ""}`)}
          ${metaItem("处理模型", fixed.model || "")}
          ${metaItem("尝试次数", fixed.attempt || "")}
          ${metaItem("错误", fixed.error || "")}
        </div>
        <div class="compare-grid">
          <div class="text-block">
            <h3>命中片段</h3>
            <pre>${escapeHtml(hit.snippet || fixed.before || "")}</pre>
          </div>
          <div class="text-block">
            <h3>AI替换结果</h3>
            <pre>${escapeHtml(fixed.after || "暂无AI替换内容")}</pre>
          </div>
        </div>
      </div>
    `;
  }).join("");
  const eventRows = asArray(data.logs).map((item) => `
    <div class="log-row">
      <code>${escapeHtml(item.time || "")}</code>
      ${escapeHtml(item.event || item.raw || "")}
      ${item.data ? `<pre>${escapeHtml(JSON.stringify(item.data, null, 2))}</pre>` : ""}
    </div>
  `).join("");
  $("detailTitle").textContent = `${meta.book_id || meta.id || "任务"} 敏感日志`;
  $("detail").className = "detail-box";
  $("detail").innerHTML = `
    <div class="meta-grid">
      ${metaItem("ID", meta.book_id || meta.id)}
      ${metaItem("书名", meta.book_name || "")}
      ${metaItem("处理模式", hitsData.mode || fixedData.mode || meta.sensitive_mode || "")}
      ${metaItem("处理状态", hitsData.status || fixedData.status || meta.sensitive_status || "")}
      ${metaItem("命中数量", hitsData.hit_count ?? meta.sensitive_hit_count ?? hits.length)}
      ${metaItem("AI修复数量", fixedData.fixed_count ?? meta.sensitive_fixed_count ?? fixedItems.length)}
      ${metaItem("命中更新时间", hitsData.updated_at || "")}
      ${metaItem("替换更新时间", fixedData.updated_at || "")}
    </div>
    ${rows || `<div class="detail-empty small">暂无敏感词命中记录。抓取原文或规则处理后，如果命中敏感词，会在这里显示命中片段和 AI 替换结果。</div>`}
    <div class="text-block">
      <h3>相关处理日志</h3>
      <div class="logs-list">${eventRows || `<div class="log-row">暂无相关日志</div>`}</div>
    </div>
    <div class="actions">${detailCloseControl()}</div>
  `;
  bindDetailCloseControl();
}

function renderRulesTrace(data) {
  const meta = data.meta || {};
  const stages = asArray(data.stages);
  const rows = stages.map((stage, index) => {
    const samples = asArray(stage.samples).map((sample) => `
      <div class="log-row">
        <b>第 ${escapeHtml(sample.line || "")} 行</b>
        <div class="compare-grid">
          <div class="text-block">
            <h3>处理前</h3>
            <pre>${escapeHtml(sample.before || "")}</pre>
          </div>
          <div class="text-block">
            <h3>处理后</h3>
            <pre>${escapeHtml(sample.after || "")}</pre>
          </div>
        </div>
      </div>
    `).join("");
    return `
      <div class="text-block">
        <h3>${index + 1}. ${escapeHtml(stage.title || stage.name || "")}</h3>
        <div class="meta-grid compact">
          ${metaItem("字数", stage.chars ?? 0)}
          ${metaItem("行数", stage.lines ?? 0)}
          ${metaItem("是否改动", stage.changed ? "有改动" : "无改动")}
        </div>
        ${samples ? `<div>${samples}</div>` : `<div class="detail-empty small">这一步没有产生可见改动。</div>`}
        <pre>${escapeHtml(stage.preview || "")}</pre>
      </div>
    `;
  }).join("");
  $("detailTitle").textContent = `${meta.book_id || meta.id || "任务"} 规则追踪`;
  $("detail").className = "detail-box";
  $("detail").innerHTML = `
    <div class="meta-grid">
      ${metaItem("ID", meta.book_id || meta.id)}
      ${metaItem("书名", meta.book_name || "")}
      ${metaItem("追踪来源", data.source || "")}
      ${metaItem("来源字数", data.source_chars || 0)}
      ${metaItem("当前原文字数", data.saved_original_chars || 0)}
      ${metaItem("说明", "只追踪系统规则，不触发AI改文")}
    </div>
    ${rows || `<div class="detail-empty small">暂无规则追踪内容。</div>`}
  `;
}

function renderSiteSubmitLog(data) {
  const meta = data.meta || {};
  const result = data.result || {};
  const versions = result.versions || {};
  const savedVersions = Object.keys(versions).length ? versions : asArray(data.logs).reduce((result, entry) => {
    const version = entry?.version || entry?.data?.version;
    if (version) result[version] = { ...(entry?.data || entry), status: entry?.status || entry?.data?.status || "" };
    return result;
  }, {});
  const versionRows = Object.keys(savedVersions).map((key) => {
    const item = savedVersions[key] || {};
    const traceRows = asArray(item.execution_trace).map((entry) => `
      <div class="site-skipped-row submit-trace-row">
        <code>${escapeHtml(entry.step || "执行步骤")}</code>
        <span class="${statusClass(entry.status)}">${escapeHtml(entry.status || "")}</span>
        <span>${escapeHtml(entry.detail || "")}</span>
        <time>${escapeHtml(entry.time || "")}</time>
      </div>
    `).join("");
    const receipt = item.remote_receipt || {};
    const remote = receipt.remote_record || {};
    return `
      <div class="sensitive-log-card">
        <div class="sensitive-log-head">
          <b>${escapeHtml(key)}</b>
          <span class="${statusClass(item.status)}">${escapeHtml(item.status || "")}</span>
        </div>
        <div class="meta-grid compact">
          ${metaItem("提交组", item.group_id || "")}
          ${metaItem("文件", item.file || "")}
          ${metaItem("更新时间", item.updated_at || item.time || "")}
          ${metaItem("素材分配", item.material_allocation ? `解压 ${item.material_allocation.jieyaNum ?? 0} / 滚屏 ${item.material_allocation.gunpingNum ?? 0}` : "")}
          ${metaItem("错误", item.error || "")}
          ${metaItem("121记录", remote.found ? `已找到 #${remote.remote_id || ""}` : remote.detail || "未核验")}
        </div>
        ${traceRows ? `<div class="text-block submit-trace"><h3>执行链路</h3><div class="site-skipped-list">${traceRows}</div></div>` : ""}
        ${item.remote_receipt ? `<details class="submit-receipt"><summary>查看 121 原始回执</summary><pre class="site-output">${escapeHtml(JSON.stringify(item.remote_receipt, null, 2))}</pre></details>` : ""}
        ${item.profile ? `<pre class="site-output">${escapeHtml(JSON.stringify(item.profile, null, 2))}</pre>` : ""}
      </div>
    `;
  }).join("");
  const eventRows = asArray(data.logs).map((item) => `
    <div class="log-row">
      <code>${escapeHtml(item.time || "")}</code>
      ${escapeHtml(item.event || item.raw || "")}
      ${item.data ? `<pre>${escapeHtml(JSON.stringify(item.data, null, 2))}</pre>` : ""}
    </div>
  `).join("");
  $("detailTitle").textContent = `${meta.book_id || meta.id || "任务"} 网站提交日志`;
  $("detail").className = "detail-box";
  $("detail").innerHTML = `
    <div class="meta-grid">
      ${metaItem("ID", meta.book_id || meta.id)}
      ${metaItem("书名", meta.book_name || "")}
      ${metaItem("平台", `${meta.platform_name || ""} ${meta.platform_id || ""}`)}
      ${metaItem("风格", meta.style || "")}
      ${metaItem("男女频", meta.gender || "")}
      ${metaItem("总状态", result.status || "未提交")}
      ${metaItem("最后提交组", result.last_group_id || "")}
      ${metaItem("错误", result.error || "")}
    </div>
    ${versionRows || `<div class="detail-empty small">暂无网站提交记录。提交后会显示每个版本的状态和对应上传组。</div>`}
    <div class="text-block">
      <h3>上传过程日志</h3>
      <div class="logs-list">${eventRows || `<div class="log-row">暂无相关日志</div>`}</div>
    </div>
    <div class="actions">${detailCloseControl()}</div>
  `;
  bindDetailCloseControl();
}

function metaItem(label, value) {
  return `<div class="meta-item"><b>${escapeHtml(label)}</b><span>${escapeHtml(String(value || ""))}</span></div>`;
}

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

async function loadConfig() {
  state.config = await api("/api/bootstrap");
  state.config.knowledge_loaded = state.config.knowledge_loaded === true;
  renderConfig();
}

async function ensureKnowledgeLoaded() {
  if (state.config?.knowledge_loaded) return;
  const status = $("knowledgeStatus");
  if (status) status.textContent = "正在读取知识库...";
  state.config = { ...state.config, ...await api("/api/config"), knowledge_loaded: true };
  renderKnowledgeSummary(state.config.knowledge_summary || {});
  if (status) status.textContent = "";
}

async function loadTasks(options = {}) {
  const taskPath = state.viewMode === "date" && state.taskDate
    ? `/api/tasks?${new URLSearchParams({ date: state.taskDate }).toString()}`
    : "/api/tasks";
  const data = await api(taskPath);
  const incomingTasks = Array.isArray(data.tasks) ? data.tasks : [];
  if (options.preserveOnEmpty && incomingTasks.length === 0 && Array.isArray(state.allTasks) && state.allTasks.length > 0) {
    const fallbackTasks = Array.isArray(state.tasks) && state.tasks.length ? state.tasks : state.allTasks;
    renderTasks(fallbackTasks, { preserveVisible: true });
    return data;
  }
  state.allTasks = mergeTasksKeepingIds(incomingTasks, options.preserveIds || []);
  renderTasks(state.allTasks);
  $("summaryText").textContent = state.viewMode === "current"
    ? `${state.taskDate} 当前批次显示 ${state.tasks.length} 个任务；历史任务请切换查看日期`
    : `${state.taskDate} 显示 ${state.tasks.length} 个任务；历史任务可切换日期查看`;
}

async function refreshTasksAndSubmitHistory() {
  setBatchStatus("正在刷新任务和网站提交记录...");
  try {
    await Promise.all([loadTasks(), loadWebSubmitHistory()]);
    setBatchStatus("任务和网站提交记录已更新");
  } catch (error) {
    setBatchStatus(`刷新失败：${error.message || "请稍后重试"}`);
  }
}

async function loadLogs() {
  const data = await api("/api/logs");
  const list = $("logsList");
  list.innerHTML = "";
  for (const item of (data.logs || [])) {
    const div = document.createElement("div");
    div.className = "log-row";
    div.innerHTML = `<code>${escapeHtml(item.book_id || "")}</code> ${escapeHtml(item.time || "")} ${escapeHtml(item.event || "")}<br>${escapeHtml(JSON.stringify(item.data || item.raw || {}, null, 0))}`;
    list.appendChild(div);
  }
  if (!list.innerHTML) list.textContent = "暂无日志";
}

async function loadRecords() {
  await Promise.all([loadWebSubmitHistory(), loadLogs()]);
}

function renderProcessResult(result) {
  const lines = [
    "处理完成",
    `读取行数：${result.parsed || 0}`,
    `有效任务：${result.unique_tasks || 0}`,
    `重复ID：${result.duplicate_count || 0}`,
    `空ID行：${result.empty_id_count || 0}`,
    `raw原文抓到：${result.raw_fetched || 0}`,
    `原文处理成功：${result.fetched || 0}`,
    `系统规则/敏感词处理失败：${result.process_failed || 0}`,
    `接口抓取失败：${result.fetch_failed || 0}`,
    `AI文案：${result.generated_ai_files || 0}`,
    `AI文案生成失败：${result.rewrite_failed || 0}`,
  ];
  if (result.style_sync?.style_sync) {
    lines.push(`网站风格同步：新增 ${result.style_sync.style_sync.added?.length || 0}，删除 ${result.style_sync.style_sync.removed?.length || 0}`);
  } else if (result.style_sync?.error) {
    lines.push(`网站风格同步失败：${result.style_sync.error}`);
  }
  if (result.site_submit?.requested !== undefined) {
    lines.push(`网站提交：提交组 ${result.site_submit.group_count || 0}，成功 ${result.site_submit.success_groups || 0}，失败 ${result.site_submit.failed_groups || 0}`);
  } else if (result.site_submit?.error) {
    lines.push(`网站提交失败：${result.site_submit.error}`);
  }
  if (result.classify_errors?.length) {
    lines.push(`分类提示：${result.classify_errors.join("；")}`);
  }
  const steps = Array.isArray(result.steps) ? result.steps : [];
  if (steps.length) {
    lines.push("", "执行明细：");
    steps.forEach((item, index) => {
      const label = {
        done: "完成",
        warning: "注意",
        failed: "失败",
        skipped: "跳过",
      }[item.status] || item.status || "状态";
      lines.push(`${index + 1}. ${label}：${item.message || ""}`);
    });
  }
  lines.push("", "任务列表已刷新，可切换到“任务”查看每个ID。");
  return lines.filter((line) => line !== "").join("\n");
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function renderProcessJob(job) {
  if (job.result) return renderProcessResult(job.result);
  const lines = [
    `处理中：${job.id || ""}`,
    `状态：${job.status || "running"}`,
  ];
  const steps = Array.isArray(job.steps) ? job.steps : [];
  if (steps.length) {
    lines.push("", "实时进度：");
    steps.forEach((item, index) => {
      const label = {
        done: "完成",
        running: "执行中",
        warning: "注意",
        failed: "失败",
        skipped: "跳过",
      }[item.status] || item.status || "状态";
      lines.push(`${index + 1}. ${label}：${item.message || ""}`);
    });
  } else {
    lines.push("", "正在准备执行...");
  }
  if (job.error) lines.push("", `错误：${job.error}`);
  return lines.join("\n");
}

async function restoreLatestProcessJob() {
  const box = $("processResult");
  if (!box) return;
  try {
    const data = await api("/api/process/jobs/latest");
    const job = data.latest || {};
    if (!job.id) {
      box.textContent = "暂无正在处理的任务。";
      return;
    }
    if (job.status === "done" && Array.isArray(job.result?.current_batch_ids)) {
      setCurrentBatchFromResult(job.result, job.completed_at || job.updated_at || job.started_at);
      renderTasks(state.allTasks);
      $("summaryText").textContent = `${state.taskDate} 当前批次显示 ${state.tasks.length} 个任务；历史任务请切换查看日期`;
    } else if (job.status !== "running") {
      setCurrentBatchFromResult({}, new Date().toISOString());
      renderTasks(state.allTasks);
      $("summaryText").textContent = `${state.taskDate} 当前批次暂无任务；历史任务请切换查看日期`;
    }
    // A completed job is represented by the current-batch view above. Do not
    // replay its old summary into the live status area on every page entry.
    if (job.status !== "running") {
      box.textContent = "暂无正在处理的任务。";
      return;
    }
    state.activeProcessJobId = job.id;
    box.textContent = renderProcessJob(job);
    if (job.status === "running") {
      const result = await pollProcessJob(job.id);
      if (state.activeProcessJobId === job.id) {
        box.textContent = renderProcessResult(result);
        if (result.tasks) renderTasks(result.tasks || []);
      }
    }
  } catch (error) {
    if (!box.textContent.trim()) box.textContent = "最近处理记录读取失败：" + error.message;
  }
}

async function pollProcessJob(jobId) {
  let latest = null;
  const startedAt = Date.now();
  for (;;) {
    latest = await api(`/api/process/jobs/${encodeURIComponent(jobId)}`);
    $("processResult").textContent = renderProcessJob(latest);
    if (latest.status === "done") return latest.result || {};
    if (latest.status === "failed") throw new Error(latest.error || "处理失败");
    if (Date.now() - startedAt >= PROCESS_JOB_TIMEOUT_MS) {
      throw new Error("处理任务超过 10 分钟仍未完成，请到“任务”页刷新并查看状态");
    }
    await sleep(1000);
  }
}

async function processInput() {
  const button = $("processBtn");
  const versions = selectedProcessVersions();
  if (!versions.length) {
    $("processResult").textContent = "请至少选择一个文案版本；请打开版本对应配置档勾选版本。";
    openVersionConfigCard();
    return;
  }
  button.disabled = true;
  try {
    setCurrentBatchFromResult({}, new Date().toISOString());
    const textModelId = String($("textModelSelect")?.value || "").trim();
    if (!textModelId) throw new Error("请先在小说获取页面选择文本模型");
    await persistSelectedTextModel(textModelId);
    await saveWorkFormStateNow();
    await saveWebSubmitConfig(true);
    $("processResult").textContent = [
      "处理中...",
      "1. 正在读取版本配置",
      "2. 正在补齐缺失的风格类型/男女频",
      "3. 正在按所选平台抓取原文",
      "4. 正在生成所选文案并执行处理规则",
    ].join("\n");
    const payload = {
      platform_id: $("platformSelect").value,
      input_text: $("inputText").value,
      selected_versions: versions,
      ai_slot_methods: processAiMethods(),
      profile_bindings: webProfileBindingsFromForm(),
      sensitive_ai_enabled: sensitiveAiProcessEnabled(),
      text_model_id: textModelId,
    };
    const job = await api("/api/process/start", {
      method: "POST",
      body: JSON.stringify(payload),
    });
    state.activeProcessJobId = job.id;
    $("processResult").textContent = renderProcessJob(job);
    const result = await pollProcessJob(job.id);
    $("processResult").textContent = renderProcessResult(result);
    setCurrentBatchFromResult(result, new Date().toISOString());
    state.allTasks = result.tasks || state.allTasks;
    renderTasks(state.allTasks);
    $("summaryText").textContent = `${state.taskDate} 当前批次显示 ${state.tasks.length} 个任务；历史任务请切换查看日期`;
  } catch (error) {
    $("processResult").textContent = error.message;
  } finally {
    button.disabled = false;
  }
}

function stopTaskDetailRefresh() {
  if (state.detailRefreshTimer) window.clearInterval(state.detailRefreshTimer);
  state.detailRefreshTimer = null;
}

async function refreshTaskDetail(id) {
  if (!id || state.detailViewKind !== "task" || String(state.selectedId) !== String(id)) return;
  try {
    const data = await api(`/api/tasks/${id}`);
    if (state.detailViewKind === "task" && String(state.selectedId) === String(id)) renderDetail(data);
  } catch (error) {
    setBatchStatus(`详情刷新失败：${error.message || "请稍后重试"}`);
  }
}

function startTaskDetailRefresh(id) {
  stopTaskDetailRefresh();
  state.detailViewKind = "task";
  state.detailRefreshTimer = window.setInterval(() => { void refreshTaskDetail(id); }, 2500);
}

async function showTask(id) {
  stopTaskDetailRefresh();
  state.selectedId = id;
  const data = await api(`/api/tasks/${id}`);
  state.detailViewKind = "task";
  renderDetail(data);
  activateTab("tasks");
  document.body.classList.add("detail-modal-open");
  startTaskDetailRefresh(id);
  focusTaskDetail();
}

function closeTaskDetail() { stopTaskDetailRefresh(); state.detailViewKind = ""; document.body.classList.remove("detail-modal-open"); }
function detailCloseControl() { return '<button id="detailCloseBtn" type="button">关闭</button>'; }
function bindDetailCloseControl() {
  $("detailCloseBtn")?.addEventListener("click", event => { event.preventDefault(); event.stopPropagation(); closeTaskDetail(); });
}

function adjacentTaskId(direction) {
  const index = state.tasks.findIndex((task) => String(task.id || "") === String(state.selectedId || ""));
  const next = index + direction;
  return next >= 0 && next < state.tasks.length ? String(state.tasks[next].id || "") : "";
}

function focusTaskDetail() {
  requestAnimationFrame(() => $("detail")?.scrollIntoView({ behavior: "smooth", block: "start" }));
}

async function showSensitiveLog(id) {
  stopTaskDetailRefresh();
  state.detailViewKind = "sensitive";
  state.selectedId = id;
  const data = await api(`/api/tasks/${id}/sensitive-log`);
  renderSensitiveLog(data);
  activateTab("tasks");
  document.body.classList.add("detail-modal-open");
  focusTaskDetail();
}

async function showRulesTrace(id) {
  stopTaskDetailRefresh();
  state.detailViewKind = "rules";
  state.selectedId = id;
  const data = await api(`/api/tasks/${id}/rules-trace`);
  renderRulesTrace(data);
  activateTab("tasks");
  document.body.classList.add("detail-modal-open");
  focusTaskDetail();
}

async function showSiteSubmitLog(id) {
  stopTaskDetailRefresh();
  state.detailViewKind = "site-submit";
  state.selectedId = id;
  const data = await api(`/api/tasks/${id}/site-submit-log`);
  renderSiteSubmitLog(data);
  activateTab("tasks");
  document.body.classList.add("detail-modal-open");
  focusTaskDetail();
}

async function refetchTask(id) {
  if (!id) return;
  const result = await api(`/api/tasks/${id}/fetch`, {
    method: "POST",
    body: JSON.stringify({ sensitive_ai_enabled: sensitiveAiProcessEnabled() }),
  });
  await loadTasks();
  await showTask(id);
  if (result?.status === "refresh_failed") setBatchStatus(result.message || "本次刷新原文失败，已保留原文");
}

async function restoreOriginal(id) {
  if (!id) return;
  if (!confirm("确认从备份恢复原文？恢复时会自动重新套用系统处理规则。")) return;
  await api(`/api/tasks/${id}/restore-original`, {
    method: "POST",
    body: JSON.stringify({ sensitive_ai_enabled: sensitiveAiProcessEnabled() }),
  });
  await loadTasks();
  await showTask(id);
}

async function generateAi(id) {
  if (!id || state.aiProcessingIds.has(id)) return;
  const task = state.tasks.find(item => String(item.id || item.book_id || "") === String(id));
  if (!task) {
    setBatchStatus("未找到要生成AI文案的任务，请刷新任务列表");
    return;
  }
  const label = task.book_name || task.book_id || id;
  let tasksLoaded = false;
  state.aiProcessingIds.add(id);
  renderTasks(state.allTasks);
  setBatchStatus(`正在生成AI文案：${label}`);
  try {
    const selectedVersions = asArray(task?.selected_versions).length ? task.selected_versions : selectedProcessVersions();
    const response = await api(`/api/tasks/${id}/generate-ai`, {
      method: "POST",
      body: JSON.stringify({ selected_versions: selectedVersions, ai_slot_methods: task?.ai_slot_methods || processAiMethods(), sensitive_ai_enabled: sensitiveAiProcessEnabled() }),
    });
    if (response?.status && response.status !== "done") {
      throw new Error(response.error || response.result?.error || `AI文案生成${response.status}`);
    }
    const verification = await api(`/api/tasks/${encodeURIComponent(id)}`);
    const verifiedMeta = verification?.meta || {};
    const requestedAi = selectedVersions.filter(version => /^ai[1-5]$/.test(String(version || "").toLowerCase())).map(version => String(version).toLowerCase());
    const generatedAi = asArray(verifiedMeta.ai_generated_versions || verifiedMeta.ai_files).map(version => String(version || "").toLowerCase());
    const verificationError = verifiedMeta.ai_error || verifiedMeta.aiError || "";
    if (verificationError) throw new Error(verificationError);
    if (requestedAi.length && !requestedAi.every(version => generatedAi.includes(version))) {
      throw new Error("AI文案生成未完成，请查看任务详情和失败原因");
    }
    await loadTasks({ preserveOnEmpty: true, preserveIds: [id] });
    tasksLoaded = true;
    await refreshTaskDetail(id);
    setBatchStatus(`AI文案生成已完成：${label}`);
  } catch (error) {
    setBatchStatus(`生成AI文案失败：${normalizeBatchErrorMessage(error?.message, error?.status)}`);
  } finally {
    if (!tasksLoaded) {
      try { await loadTasks({ preserveOnEmpty: true, preserveIds: [id] }); } catch (_) { /* keep the original generation error visible */ }
    }
    await refreshTaskDetail(id);
    state.aiProcessingIds.delete(id);
    renderTasks(state.allTasks);
  }
}

function selectedTaskIds() {
  return [...state.selectedIds];
}

// The V2 parsed-book picker is rendered by a separate script. Keep its
// selection authoritative for the legacy task actions that it exposes.
window.addEventListener("qiantie-v78-task-selection", event => {
  const ids = Array.isArray(event?.detail?.ids) ? event.detail.ids.map(id => String(id || "").trim()).filter(Boolean) : [];
  state.selectedIds = new Set(ids);
  updateSelectedCount();
});

function setBatchStatus(text) {
  $("batchStatus").textContent = text || "";
}

async function platformApi(path, options = {}) {
  const token = localStorage.getItem("auth_token") || "";
  const response = await fetch(path, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options.headers || {}),
    },
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.error || `请求失败（${response.status}）`);
  return data;
}

async function transferSelectedToBatchFactory() {
  const ids = selectedTaskIds();
  if (!ids.length) {
    setBatchStatus("先选择要转入批量工厂的任务");
    return;
  }
  const button = $("transferBatchFactoryBtn");
  button.disabled = true;
  setBatchStatus(`正在检查 ${ids.length} 个任务的原文…`);
  try {
    const details = await Promise.all(ids.map(id => api(`/api/tasks/${encodeURIComponent(id)}`)));
    const items = details.flatMap((detail, index) => {
      const meta = detail?.meta || {};
      const sourceTaskId = String(meta.book_id || ids[index]).trim();
      const bookId = String(meta.book_id || ids[index]).trim();
      if (!sourceTaskId || !bookId) return [];
      const baseTitle = String(meta.book_name || bookId).trim();
      const versions = [{ version: "original", text: detail?.original }]
        .concat(Array.isArray(detail?.ai_texts) ? detail.ai_texts : []);
      const seenVersions = new Set();
      return versions.flatMap(item => {
        const version = String(item?.version || item?.name || "").trim().toLowerCase();
        const sourceText = String(item?.text || "").trim();
        if (!version || !sourceText || seenVersions.has(version)) return [];
        seenVersions.add(version);
        return [{
          sourceTaskId,
          bookId,
          title: version === "original" ? baseTitle : `${version.toUpperCase()}·${baseTitle}`,
          platform: String(meta.platform_name || "").trim(),
          sourceText,
          txtText: sourceText,
          sourceMetadata: { ...meta, sourceContentVersion: version },
        }];
      });
    });
    const transferredTaskCount = details.filter(detail => {
      const hasOriginal = Boolean(String(detail?.original || "").trim());
      const hasAiVersion = Array.isArray(detail?.ai_texts) && detail.ai_texts.some(item => String(item?.text || "").trim());
      return hasOriginal || hasAiVersion;
    }).length;
    const skipped = ids.length - transferredTaskCount;
    if (!items.length) throw new Error("选中的任务都没有可用原文，请先完成原文获取");
    const result = await platformApi("/api/batch-factory/v12/intakes/novel-fetch", {
      method: "POST",
      body: JSON.stringify({
        books: items.map(item => ({
          id: item.bookId,
          bookId: item.bookId,
          sourceTaskId: item.sourceTaskId,
          title: item.title,
          platform: item.platform,
          sourceText: item.sourceText,
          txtText: item.txtText,
          txtFileName: `${item.bookId}.txt`,
          sourceMetadata: item.sourceMetadata,
        })),
        metadata: {
          name: `小说获取转入 ${items.length} 本`,
          source: "novel-fetch",
          transferredAt: new Date().toISOString(),
        },
      }),
    });
    const intakeId = String(result?.intake?.id || "").trim();
    const redirectTo = intakeId ? `/shuihuo-production?intake=${encodeURIComponent(intakeId)}` : "";
    if (!redirectTo) throw new Error("小说获取交接创建成功但未返回交接编号");
    setBatchStatus(`已转入 ${items.length} 本${skipped ? `, 跳过 ${skipped} 本未完成任务` : ""}`);
    if (window.parent !== window) {
      window.parent.postMessage({ type: "qiantie:batch-factory-intake", intakeId, redirectTo }, window.location.origin);
    } else {
      window.location.assign(redirectTo);
    }
  } catch (error) {
    setBatchStatus(error.message || "转入批量工厂失败");
  } finally {
    button.disabled = false;
  }
}

let batchToastTimer = null;
function showBatchToast(text, type = "success") {
  if (window.parent !== window) {
    window.parent.postMessage({
      type: "qiantie:task-notification",
      source: "小说获取",
      status: type === "warning" || type === "error" ? "error" : "success",
      title: type === "warning" || type === "error" ? "小说获取任务需要处理" : "小说获取任务已完成",
      detail: String(text || "").replace(/\s+/g, " ").trim().slice(0, 280),
      page: "小说获取",
      pagePath: "/novel-fetch",
    }, "*");
  }
  let toast = $("batchToast");
  if (!toast) {
    toast = document.createElement("div");
    toast.id = "batchToast";
    toast.setAttribute("role", "status");
    toast.setAttribute("aria-live", "polite");
    document.body.appendChild(toast);
  }
  toast.className = `batch-toast ${type}`;
  toast.textContent = text || "";
  requestAnimationFrame(() => toast.classList.add("visible"));
  clearTimeout(batchToastTimer);
  batchToastTimer = setTimeout(() => toast.classList.remove("visible"), 4200);
}

async function batchDelete(mode) {
  const ids = mode === "selected" ? selectedTaskIds() : [];
  if (mode === "selected" && !ids.length) {
    setBatchStatus("先选择任务");
    return;
  }
  const label = mode === "all" ? "全部任务" : (mode === "failed" ? "失败任务" : `${ids.length} 个选中任务`);
  if (!confirm(`确认删除${label}？会删除对应 ID 的本地内容文件。`)) return;
  setBatchStatus("删除中...");
  try {
    const result = await api("/api/tasks/batch-delete", {
      method: "POST",
      body: JSON.stringify({ mode, ids }),
    });
    state.selectedIds.clear();
    if (state.selectedId && !(result.tasks || []).some((task) => String(task.id) === String(state.selectedId))) {
      state.selectedId = "";
      $("detailTitle").textContent = "未选择";
      $("detail").className = "detail-empty";
      $("detail").textContent = "在“任务”页点击书籍 ID 或“查看”，这里会展示该书的原文、AI 文案、处理规则和提交记录。";
    }
    renderTasks(result.tasks || []);
    setBatchStatus(`已删除 ${result.deleted || 0} 个，失败 ${result.failed || 0} 个`);
  } catch (error) {
    setBatchStatus(error.message);
  }
}

function failedTaskIds(tasks) {
  const ids = new Set();
  for (const task of (Array.isArray(tasks) ? tasks : [])) {
    const status = [
      task.status, task.classify_status, task.classify_error,
      task.original_status, task.original_error, task.original_error_code,
      task.sensitive_status, task.ai_status, task.ai_error,
      task.site_submit_status, task.site_submit_error
    ].filter(Boolean).join(' ');
    const failedSubmit = asArray(task.site_submit_failed_versions).length > 0;
    const sensitiveFailed = Number(task.sensitive_failed_count || 0) > 0;
    if (failedSubmit || sensitiveFailed || /(failed|error|timeout|interrupted|incomplete|失败|错误|超时|中断|未完成|121异常|partial|部分)/i.test(status)) {
      const id = String(task.id || task.book_id || '').trim();
      if (id) ids.add(id);
    }
  }
  return [...ids];
}

async function batchRetry(mode) {
  const ids = mode === "selected" ? selectedTaskIds() : failedTaskIds(state.tasks);
  if (!ids.length) {
    setBatchStatus(mode === "selected" ? "先选择任务" : "当前列表没有失败任务");
    return;
  }
  const label = mode === "failed" ? "失败任务" : `${ids.length} 个选中任务`;
  setBatchStatus(`重试${label}中...`);
  try {
    const result = await api("/api/tasks/batch-retry", {
      method: "POST",
      // “重试失败”固定使用当前任务表可见的失败 ID，保持表格从上到下的执行顺序，
      // 不把日期筛选外的历史失败任务混入本次队列。
      body: JSON.stringify({ mode: "selected", ids, sensitive_ai_enabled: sensitiveAiProcessEnabled() }),
    });
    renderTasks(result.tasks || []);
    const stageLabels = { classify: "AI 判断", original: "原文抓取", sensitive: "敏感词处理", rewrite: "AI 文案生成", submit: "网站提交" };
    const stageCounts = result.retry_stage_counts || {};
    const stageSummary = Object.entries(stageCounts)
      .map(([stage, count]) => `${stageLabels[stage] || stage} ${count} 个`)
      .join("，");
    const deduplicated = Number(result.deduplicated) || 0;
    const throughput = result.throughput || {};
    const active = throughput.active || {};
    const limits = throughput.limits || {};
    const capacity = Object.keys(limits).length
      ? `；资源池 抓取 ${active.fetch || 0}/${limits.fetch || 0}，改文 ${active.rewrite || 0}/${limits.rewrite || 0}，提交 ${active.submit || 0}/${limits.submit || 0}${throughput.paused ? "（负载保护中）" : ""}`
      : "";
    const duplicateText = deduplicated ? `；已在队列 ${deduplicated} 个` : "";
    setBatchStatus((stageSummary
      ? `正在按失败步骤重试：${stageSummary}；已加入队列 ${result.retried || 0} 个`
      : `已重试 ${result.retried || 0} 个，失败 ${result.failed || 0} 个`) + duplicateText + capacity);
    if (Number(result.retried) > 0) {
      window.dispatchEvent(new CustomEvent("qiantie-novel-fetch-retry-queued", {
        detail: { ids: ids.slice(), retried: Number(result.retried) || 0 }
      }));
    }
    return result;
  } catch (error) {
    setBatchStatus(error.message);
    return null;
  }
}

async function applyRules(mode) {
  const ids = mode === "selected" ? selectedTaskIds() : [];
  if (mode === "selected" && !ids.length) {
    setBatchStatus("先选择任务");
    return;
  }
  const label = mode === "all" ? "全部任务" : `${ids.length} 个选中任务`;
  setBatchStatus(`规则处理${label}中...`);
  try {
    const result = await api("/api/tasks/apply-rules", {
      method: "POST",
      body: JSON.stringify({
        mode,
        ids,
        scope: $("ruleApplyScope").value || "both",
        sensitive_ai_enabled: sensitiveAiProcessEnabled(),
      }),
    });
    renderTasks(result.tasks || []);
    setBatchStatus(`规则处理完成 ${result.applied || 0} 个，失败 ${result.failed || 0} 个`);
  } catch (error) {
    setBatchStatus(error.message);
  }
}

async function reprocessSensitive(ids, restoreFromBackup) {
  const selected = Array.isArray(ids) ? ids.filter(Boolean) : [];
  if (!selected.length) {
    setBatchStatus("先选择任务");
    return;
  }
  if (restoreFromBackup && !confirm("将从原文备份恢复后，按当前敏感词规则重新处理。当前处理后原文会被覆盖，是否继续？")) return;
  const label = restoreFromBackup ? "从备份重跑敏感词" : "重跑敏感词";
  let processed = 0;
  let restored = 0;
  let failed = 0;
  let latestTasks = state.tasks;
  let firstError = "";
  for (let index = 0; index < selected.length; index += 1) {
    const id = selected[index];
    state.sensitiveProcessingIds.add(id);
    setBatchStatus(`${label}中：第 ${index + 1}/${selected.length} 条`);
    renderTasks(latestTasks);
    try {
      const result = await api("/api/tasks/reprocess-sensitive", {
        method: "POST",
        body: JSON.stringify({
          ids: [id],
          restore_from_backup: restoreFromBackup,
          sensitive_ai_enabled: sensitiveAiProcessEnabled(),
        }),
      });
      processed += Number(result.processed || 0);
      restored += Number(result.restored || 0);
      failed += Number(result.failed || 0);
      latestTasks = result.tasks || latestTasks;
    } catch (error) {
      failed += 1;
      firstError ||= error.message || "请求失败";
    } finally {
      state.sensitiveProcessingIds.delete(id);
      renderTasks(latestTasks);
    }
  }
  const summary = `${label}完成：已处理 ${processed} 个${restoreFromBackup ? `，已恢复 ${restored} 个` : ""}，失败 ${failed} 个`;
  const detail = firstError ? `${summary}：${firstError}` : summary;
  setBatchStatus(detail);
  const processResult = $("processResult");
  if (processResult) processResult.textContent = detail;
  if (state.selectedId && selected.includes(String(state.selectedId))) await showTask(state.selectedId);
}

// The V2 login hotfix is loaded after this legacy client. Both layers need the
// same environment result, but must never probe the Browser Worker twice in
// parallel during a page switch (the second result used to overwrite a good
// session with a transient timeout).
window.qiantieEnsureWebLoginEnvironment = function qiantieEnsureWebLoginEnvironment() {
  if (window.__qiantieWebLoginEnvironmentPromise) return window.__qiantieWebLoginEnvironmentPromise;
  window.__qiantieWebLoginEnvironmentPromise = api("/api/web-submit/environment")
    .then(environment => {
      state.webLoginSession = environment?.ok === true;
      renderWebLoginStatus(state.config?.web_submit || {});
      return environment;
    })
    .catch(error => {
      state.webLoginSession = false;
      renderWebLoginStatus(state.config?.web_submit || {});
      throw error;
    });
  return window.__qiantieWebLoginEnvironmentPromise;
};

async function updateSelectedAiCount() {
  const ids = selectedTaskIds();
  const count = Number($("batchAiCount")?.value || 1);
  if (!ids.length) return setBatchStatus("先选择任务");
  if (!Number.isInteger(count) || count < 1 || count > 20) return setBatchStatus("AI数量需为 1 到 20");
  setBatchStatus("正在调整 AI 数量...");
  try {
    const result = await api("/api/tasks/batch-ai-count", { method: "POST", body: JSON.stringify({ ids, ai_count: count }) });
    setBatchStatus(`已调整 ${result.updated || 0} 个任务的 AI 数量为 ${count}`);
    await refreshTasksAndSubmitHistory();
  } catch (error) { setBatchStatus(error.message); }
}

async function startSensitiveProcessing() {
  const ids = state.tasks.map(task => String(task.id || '')).filter(Boolean);
  if (!ids.length) {
    setBatchStatus('当前没有可处理任务');
    return;
  }
  await reprocessSensitive(ids, false);
  await loadTasks();
}

function openWebSubmitFromTasks() {
  void submitWebSubmit("selected");
}

// V2 action bars use this explicit bridge instead of dispatching a synthetic click
// against the legacy button, which may be replaced while switching task panels.
window.qiantieSubmitSelectedTasks = ids => void submitWebSubmit("selected", ids);

function selectAllVisibleTasks() {
  for (const task of state.tasks) {
    if (task.id) state.selectedIds.add(String(task.id));
  }
  renderTasks(state.tasks);
}

function clearSelectedTasks() {
  state.selectedIds.clear();
  renderTasks(state.tasks);
}

function syncFormToAppConfig() {
  const cfg = clone(state.config.app_config || {});
  cfg.workflow = {
    ...(cfg.workflow || {}),
    auto_classify_missing: $("workflowAutoClassify").checked,
    auto_fetch_original: $("workflowAutoFetch").checked,
    auto_rewrite_after_fetch: $("workflowAutoRewrite").checked,
    auto_submit_after_rewrite: false,
    auto_submit_confirmed: false,
  };
  cfg.sensitive_ai = {
    ...(cfg.sensitive_ai || {}),
    enabled: sensitiveAiProcessEnabled(),
  };
  cfg.fetch = {
    ...(cfg.fetch || {}),
    endpoint: $("fetchEndpoint").value.trim() || "https://txt.121w.com/api.php",
    default_max_txt: numberValue("fetchMaxTxt", 4000),
    concurrency: numberValue("fetchConcurrency", 4),
    retries: numberValue("fetchRetries", 1),
    timeout_seconds: numberValue("fetchTimeout", 30),
    auto_detect_platform: $("fetchAutoDetectPlatform").checked,
  };
  cfg.rewrite = {
    ...(cfg.rewrite || {}),
    process_line_count: numberValue("processLineCount", 5),
    anchor_line_count: numberValue("anchorLineCount", 5),
    temperature: numberValue("rewriteTemp", 0.45),
    method_sequence: $("methodSequenceInput").value.split(/[\s,，、|\/]+/).map((item) => item.trim()).filter(Boolean),
    default_template_id: $("rewriteTemplateSelect").value || "",
    opening_phrase_mode: $("openingPhraseMode").value || "auto",
    high_imitation_mode: $("highImitationMode").value || "auto",
    prompt: $("rewritePrompt").value,
    processing_rule_prompt: $("processingRulePrompt").value,
  };
  ensureKnowledgeConfig();
  state.config.knowledge.usage_prompt = $("knowledgeUsagePrompt")?.value || "";
  cfg.ai = { ...(cfg.ai || {}), ...readAiSettingsFromForm() };
  cfg.text_model_id = $("textModelSelect")?.value || cfg.text_model_id || cfg.textModelId || "";
  cfg.ai_assignments = {
    classifier: $("classifierSelect").value || "__current__",
    rewrite: $("rewriteSelect").value || "__current__",
    sensitive_fix: $("sensitiveFixSelect")?.value || "__current__",
  };
  cfg.ai_presets = cfg.ai_presets || [];
  syncCurrentAiPreset(cfg);
  return cfg;
}

async function saveConfig(throwOnError = false) {
  $("configStatus").textContent = "保存中...";
  try {
    saveLibraryItem(true);
    saveRuleEditor(true);
    const appConfig = syncFormToAppConfig();
    const activePresetId = appConfig.ai_assignments?.rewrite === "preset_current_auto" ? "preset_current_auto" : "";
    const payload = {
      app_config: appConfig,
      platforms: JSON.parse($("platformsText").value),
      styles: JSON.parse($("stylesText").value),
      sensitive: state.config.sensitive || { groups: [] },
      ...(state.config.knowledge_loaded ? { knowledge: state.config.knowledge || {} } : {}),
    };
    const data = await api("/api/config", {
      method: "POST",
      body: JSON.stringify(payload),
    });
    mergeConfigResponse(data.config);
    renderConfig();
    if (activePresetId && $("presetSelect")) {
      $("presetSelect").value = activePresetId;
      $("presetName").value = state.config.app_config?.ai_presets?.find((item) => item?.id === activePresetId)?.name || "当前预设（自动同步）";
    }
    $("configStatus").textContent = activePresetId ? "已保存，并已更新当前预设" : "已保存";
    const knowledgeSave = $("knowledgeSaveStatus");
    if (knowledgeSave) knowledgeSave.textContent = "已保存";
    const rulesSave = $("rulesSaveStatus");
    if (rulesSave) rulesSave.textContent = "已保存";
    return data;
  } catch (error) {
    $("configStatus").textContent = error.message;
    if (throwOnError) throw error;
  }
}

async function saveCurrentPreset() {
  const cfg = syncFormToAppConfig();
  const presets = cfg.ai_presets || [];
  const selectedId = $("presetSelect").value;
  const name = $("presetName").value.trim() || $("textModelSelect")?.selectedOptions?.[0]?.textContent?.trim() || `预设${presets.length + 1}`;
  let id = selectedId && selectedId !== "" ? selectedId : `preset_${Date.now()}`;
  if (!presets.some((item) => item.id === id)) {
    id = `preset_${Date.now()}`;
  }
  const nextPreset = { ...readAiSettingsFromForm(), id, name };
  const index = presets.findIndex((item) => item.id === id);
  if (index >= 0) {
    presets[index] = nextPreset;
  } else {
    presets.push(nextPreset);
  }
  state.config.app_config = { ...cfg, ai_presets: presets };
  renderPresetControls(state.config.app_config);
  $("presetSelect").value = id;
  await saveConfig();
  $("configStatus").textContent = "预设已保存";
}

function applySelectedPreset() {
  const id = $("presetSelect").value;
  const preset = (state.config?.app_config?.ai_presets || []).find((item) => item.id === id);
  if (!preset) return;
  applyAiSettingsToForm(preset);
  $("presetName").value = preset.name || "";
  $("configStatus").textContent = "已应用预设，记得保存配置";
}

async function deleteSelectedPreset() {
  const id = $("presetSelect").value;
  if (!id) return;
  const cfg = syncFormToAppConfig();
  cfg.ai_presets = (cfg.ai_presets || []).filter((item) => item.id !== id);
  if (cfg.ai_assignments?.classifier === id) cfg.ai_assignments.classifier = "__current__";
  if (cfg.ai_assignments?.rewrite === id) cfg.ai_assignments.rewrite = "__current__";
  if (cfg.ai_assignments?.sensitive_fix === id) cfg.ai_assignments.sensitive_fix = "__current__";
  state.config.app_config = cfg;
  renderPresetControls(cfg);
  await saveConfig();
  $("configStatus").textContent = "预设已删除并保存";
}

async function testAi() {
  $("aiTestStatus").textContent = "测试中...";
  try {
    const result = await api("/api/ai/test", {
      method: "POST",
      body: JSON.stringify({
        purpose: "rewrite",
      }),
    });
    $("aiTestStatus").textContent = `成功：${result.content || "ok"}（已使用所选文本模型）`;
  } catch (error) {
    $("aiTestStatus").textContent = error.message;
  }
}

async function refreshKnowledge() {
  $("knowledgeStatus").textContent = "刷新中...";
  try {
    const result = await api("/api/knowledge");
    state.config.knowledge_summary = result.summary || {};
    renderKnowledgeSummary(state.config.knowledge_summary);
    renderRewriteTemplateOptions(state.config?.app_config?.rewrite?.default_template_id || "");
    $("knowledgeStatus").textContent = "已刷新";
  } catch (error) {
    $("knowledgeStatus").textContent = error.message;
  }
}

async function analyzeOpeningPhrase() {
  const button = $("openingAnalyzeBtn");
  const status = $("openingAnalyzeStatus");
  const resultBox = $("openingAnalyzeResult");
  button.disabled = true;
  status.textContent = "拆解中...";
  resultBox.value = "";
  state.pendingOpeningItem = null;
  try {
    const result = await api("/api/opening/analyze", {
      method: "POST",
      body: JSON.stringify({
        source_text: $("openingSourceText").value,
        style: $("openingAnalyzeStyle").value,
        extra_requirement: $("openingAnalyzeExtra").value,
      }),
    });
    state.pendingOpeningItem = result.item || null;
    resultBox.value = pretty(result.item || {});
    status.textContent = `已拆解${result.model ? `，模型：${result.model}` : ""}`;
  } catch (error) {
    status.textContent = error.message;
  } finally {
    button.disabled = false;
  }
}

async function saveAnalyzedOpeningPhrase() {
  const status = $("openingAnalyzeStatus");
  let item = state.pendingOpeningItem;
  if (!item) {
    try {
      item = JSON.parse($("openingAnalyzeResult").value || "{}");
    } catch {
      status.textContent = "拆解结果不是有效格式";
      return;
    }
  }
  if (!item || !item.content) {
    status.textContent = "没有可入库的开头词内容";
    return;
  }
  status.textContent = "入库中...";
  try {
    const result = await api("/api/opening/save", {
      method: "POST",
      body: JSON.stringify({ item }),
    });
    state.config = result.config || state.config;
    state.pendingOpeningItem = null;
    $("libraryTypeSelect").value = "opening_phrases";
    renderConfig();
    status.textContent = "已加入开头词库";
  } catch (error) {
    status.textContent = error.message;
  }
}

async function normalizeOpeningLibrary() {
  const status = $("knowledgeStatus");
  status.textContent = "矫正中...";
  try {
    const result = await api("/api/opening/normalize", { method: "POST" });
    state.config = result.config || state.config;
    $("libraryTypeSelect").value = "opening_phrases";
    renderConfig();
    status.textContent = `已矫正开头词库：${result.summary?.opening_phrases || 0} 条`;
  } catch (error) {
    status.textContent = error.message;
  }
}

function clearOpeningAnalyzer() {
  $("openingSourceText").value = "";
  $("openingAnalyzeExtra").value = "";
  $("openingAnalyzeResult").value = "";
  $("openingAnalyzeStatus").textContent = "";
  state.pendingOpeningItem = null;
}

async function previewRuleResult() {
  saveRuleEditor(true);
  const input = $("rulePreviewInput").value;
  $("rulePreviewOutput").value = "处理中...";
  try {
    await saveConfig();
    const result = await api("/api/rules/preview", {
      method: "POST",
      body: JSON.stringify({
        text: input,
        scope: $("ruleTypeSelect").value === "sensitive" ? "original" : "ai",
      }),
    });
    $("rulePreviewOutput").value = result.processed || "";
  } catch (error) {
    $("rulePreviewOutput").value = error.message;
  }
}

document.addEventListener("click", async (event) => {
  const loginStatus = event.target.closest("#webLoginStatus");
  if (loginStatus) {
    await openWebLoginDialog();
    return;
  }
  const button = event.target.closest("button");
  if (!button) return;
  if (button.dataset.ruleAction) {
    handleRuleAction(button.dataset.ruleAction);
    return;
  }
  if (button.classList.contains("tab")) {
    activateTab(button.dataset.tab);
    if (button.dataset.tab === "tasks") await refreshTasksAndSubmitHistory();
    if (button.dataset.tab === "knowledge") {
      try {
        await ensureKnowledgeLoaded();
        renderLibraryManager(Number($("libraryItemSelect")?.value || 0));
      } catch (error) {
        const status = $("knowledgeStatus");
        if (status) status.textContent = error.message || "知识库读取失败";
      }
    }
    if (button.dataset.tab === "rules") renderRuleEditor();
    if (button.dataset.tab === "logs") await loadRecords();
    return;
  }
  const action = button.dataset.action;
  const id = button.dataset.id;
  if (action === "detail") await showTask(id);
  if (action === "fetch") await refetchTask(id);
  if (action === "ai") await generateAi(id);
  if (action === "sensitive") await showSensitiveLog(id);
  if (action === "siteLog") await showSiteSubmitLog(id);
});

document.addEventListener("change", (event) => {
  const target = event.target;
  if (!target) return;
  if (target.id === "libraryItemSelect") renderLibraryManager(Number(target.value || 0));
  if (target.id === "sensitiveGroupSelect" || target.id === "sensitiveRuleSelect") renderSensitiveRuleEditor();
  if (target.id === "ruleItemSelect") renderRuleEditor(Number(target.value || 0));
  if (target.id === "taskDateFilter") {
    state.viewMode = "date";
    state.taskDate = target.value || todayDateKey();
    void loadTasks();
  }
});

document.addEventListener("pointerdown", event => {
  if (!document.body.classList.contains("detail-modal-open")) return;
  if (event.target.closest("#detail")) return;
  closeTaskDetail();
}, true);

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") closeTaskDetail();
  const loginStatus = event.target.closest?.("#webLoginStatus");
  if (loginStatus && (event.key === "Enter" || event.key === " ")) {
    event.preventDefault();
    void openWebLoginDialog();
  }
});

document.addEventListener("change", (event) => {
  const input = event.target.closest("input");
  if (!input) return;
  if (input.id === "selectAllTasks") {
    if (input.checked) {
      selectAllVisibleTasks();
    } else {
      clearSelectedTasks();
    }
  }
  if (input.classList.contains("task-check")) {
    const id = String(input.dataset.id || "");
    if (!id) return;
    if (input.checked) {
      state.selectedIds.add(id);
    } else {
      state.selectedIds.delete(id);
    }
    updateSelectedCount();
  }
});

window.addEventListener("DOMContentLoaded", async () => {
  function bindOptionalControl(id, property, handler) {
    const control = $(id);
    if (control) control[property] = handler;
  }
  function listenOptionalControl(id, event, handler) {
    $(id)?.addEventListener(event, handler);
  }
  const sitePanel = $("siteSubmit");
  const mount = $("webSubmitMount");
  if (sitePanel && mount) {
    sitePanel.classList.remove("panel", "hidden");
    sitePanel.removeAttribute("id");
    const details = document.createElement("details");
    details.className = "submit-settings-details";
    details.open = false;
    details.innerHTML = "<summary>提交文案与提交方式</summary>";
    while (sitePanel.firstChild) details.appendChild(sitePanel.firstChild);
    mount.appendChild(details);
    sitePanel.remove();
  }
  bindWorkFormPersistence();
  bindOptionalControl("versionConfigBtn", "onclick", openVersionConfigCard);
  bindOptionalControl("versionConfigCloseBtn", "onclick", closeVersionConfigCard);
  bindOptionalControl("versionConfigCloseBtnBottom", "onclick", closeVersionConfigCard);
  listenOptionalControl("versionConfigCard", "cancel", event => {
    event.preventDefault();
    if (!isModalBusy("versionConfigCard")) closeVersionConfigCard();
  });
  listenOptionalControl("versionConfigCard", "pointerdown", event => {
    if (event.target === $("versionConfigCard") && !isModalBusy("versionConfigCard")) closeVersionConfigCard();
  });
  listenOptionalControl("versionConfigCard", "click", event => {
    if (event.target === $("versionConfigCard") && !isModalBusy("versionConfigCard")) closeVersionConfigCard();
  });
  bindOptionalControl("processBtn", "onclick", processInput);
  bindOptionalControl("refreshBtn", "onclick", refreshTasksAndSubmitHistory);
  bindOptionalControl("taskRefreshBtn", "onclick", refreshTasksAndSubmitHistory);
  bindOptionalControl("taskTodayBtn", "onclick", async () => { state.viewMode = "date"; state.taskDate = todayDateKey(); await refreshTasksAndSubmitHistory(); });
  bindOptionalControl("taskPrevDayBtn", "onclick", async () => { state.viewMode = "date"; state.taskDate = shiftTaskDateKey(state.taskDate || todayDateKey(), -1); await refreshTasksAndSubmitHistory(); });
  bindOptionalControl("taskNextDayBtn", "onclick", async () => { state.viewMode = "date"; state.taskDate = shiftTaskDateKey(state.taskDate || todayDateKey(), 1); await refreshTasksAndSubmitHistory(); });
  bindOptionalControl("taskDefaultViewBtn", "onclick", async () => { state.viewMode = "current"; state.taskDate = state.currentBatchDate || todayDateKey(); await refreshTasksAndSubmitHistory(); });
  bindOptionalControl("taskToggleBtn", "onclick", () => { const details = $("taskListDetails"); details.open = !details.open; $("taskToggleBtn").textContent = details.open ? "收起任务" : "展开任务"; });
  bindOptionalControl("selectAllBtn", "onclick", selectAllVisibleTasks);
  bindOptionalControl("clearSelectedBtn", "onclick", clearSelectedTasks);
  bindOptionalControl("retrySelectedBtn", "onclick", () => batchRetry("selected"));
  bindOptionalControl("retryFailedBtn", "onclick", () => batchRetry("failed"));
  bindOptionalControl("transferBatchFactoryBtn", "onclick", transferSelectedToBatchFactory);
  bindOptionalControl("applyRulesSelectedBtn", "onclick", () => applyRules("selected"));
  bindOptionalControl("applyRulesAllBtn", "onclick", () => applyRules("all"));
  bindOptionalControl("openWebSubmitBtn", "onclick", openWebSubmitFromTasks);
  bindOptionalControl("submitTaskAllBtn", "onclick", () => void submitWebSubmit("all"));
  bindOptionalControl("updateAiCountSelectedBtn", "onclick", updateSelectedAiCount);
  bindOptionalControl("deleteSelectedBtn", "onclick", () => batchDelete("selected"));
  bindOptionalControl("deleteFailedBtn", "onclick", () => batchDelete("failed"));
  bindOptionalControl("deleteAllBtn", "onclick", () => batchDelete("all"));
  bindOptionalControl("logsRefreshBtn", "onclick", loadRecords);
  bindOptionalControl("saveConfigBtn", "onclick", saveConfig);
  bindOptionalControl("saveKnowledgeConfigBtn", "onclick", saveConfig);
  bindOptionalControl("saveRulesConfigBtn", "onclick", saveConfig);
  bindOptionalControl("saveWorkflowBtn", "onclick", saveConfig);
  bindOptionalControl("saveAiConfigBtn", "onclick", saveConfig);
  bindOptionalControl("savePresetBtn", "onclick", saveCurrentPreset);
  bindOptionalControl("applyPresetBtn", "onclick", applySelectedPreset);
  bindOptionalControl("deletePresetBtn", "onclick", deleteSelectedPreset);
  bindOptionalControl("testAiBtn", "onclick", testAi);
  bindOptionalControl("libraryTypeSelect", "onchange", () => renderLibraryManager(0));
  bindOptionalControl("librarySearchInput", "oninput", () => renderLibraryManager(Number($("libraryItemSelect").value || 0)));
  bindOptionalControl("addLibraryItemBtn", "onclick", addLibraryItem);
  bindOptionalControl("duplicateLibraryItemBtn", "onclick", duplicateLibraryItem);
  bindOptionalControl("saveLibraryItemBtn", "onclick", () => saveLibraryItem(false));
  bindOptionalControl("deleteLibraryItemBtn", "onclick", deleteLibraryItem);
  bindOptionalControl("optimizeLibraryItemBtn", "onclick", optimizeLibraryItem);
  bindOptionalControl("openingAnalyzeBtn", "onclick", analyzeOpeningPhrase);
  bindOptionalControl("openingSaveAnalyzedBtn", "onclick", saveAnalyzedOpeningPhrase);
  bindOptionalControl("openingClearAnalyzeBtn", "onclick", clearOpeningAnalyzer);
  bindOptionalControl("normalizeOpeningBtn", "onclick", normalizeOpeningLibrary);
  bindOptionalControl("ruleTypeSelect", "onchange", () => renderRuleEditor(0));
  bindOptionalControl("saveRuleEditorBtn", "onclick", () => saveRuleEditor(false));
  bindOptionalControl("previewRuleBtn", "onclick", previewRuleResult);
  bindOptionalControl("ruleAiSuggestBtn", "onclick", suggestCurrentRuleWithAi);
  bindOptionalControl("ruleAiApplyBtn", "onclick", applyRuleSuggestions);
  bindOptionalControl("syncWebProfilesBtn", "onclick", () => syncWebSubmit("configs"));
  bindOptionalControl("syncWebStylesBtn", "onclick", () => syncWebSubmit("styles"));
  bindOptionalControl("syncWebOrganizationsBtn", "onclick", () => syncWebSubmit("organizations"));
  bindOptionalControl("confirmWebSubmitSelectionBtn", "onclick", confirmWebSubmitSelection);
  bindOptionalControl("webAllowResubmit", "onchange", updateResubmitHint);
  bindOptionalControl("platformSelect", "onchange", updatePlatformHint);
  // 121 环境验证可能触发 Browser Worker，最慢时会等待超时。它只影响
  // 提交能力，不应阻塞小说获取工作台首次进入；配置和任务先完成后立即展示。
  void window.qiantieEnsureWebLoginEnvironment().catch(error => {
    setBatchStatus(error?.message || "121 登录状态检查失败");
  });
  try {
    await Promise.allSettled([loadConfig(), loadTasks()]);
  } finally {
    document.documentElement.classList.remove("qiantie-novel-fetch-hydrating");
    window.parent?.postMessage({ type: "qiantie:novel-fetch-ready" }, window.location.origin);
  }
  await restoreLatestProcessJob();
});
