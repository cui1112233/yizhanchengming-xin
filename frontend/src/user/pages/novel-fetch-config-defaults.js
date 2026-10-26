function defaultEnabled(value) {
  return typeof value === 'boolean' ? value : true;
}

export const NOVEL_FETCH_PROFILE_VERSIONS = ['original', 'ai1', 'ai2', 'ai3', 'ai4', 'ai5'];

export function withNovelFetchVersionProfiles(bindings = {}) {
  const source = bindings && typeof bindings === 'object' ? bindings : {};
  return Object.fromEntries(NOVEL_FETCH_PROFILE_VERSIONS.map(version => [version, String(source[version] || '')]));
}

export function withNovelFetchConfigDefaults(config = {}) {
  const workflow = { ...(config.workflow || {}) };
  const fetch = { ...(config.fetch || {}) };
  const sensitiveAI = { ...(config.sensitive_ai || {}) };
  const webSubmit = { ...(config.web_submit || {}) };

  workflow.auto_classify_missing = defaultEnabled(workflow.auto_classify_missing);
  workflow.auto_fetch_original = defaultEnabled(workflow.auto_fetch_original);
  workflow.auto_rewrite_after_fetch = defaultEnabled(workflow.auto_rewrite_after_fetch);
  fetch.auto_detect_platform = defaultEnabled(fetch.auto_detect_platform);
  sensitiveAI.enabled = defaultEnabled(sensitiveAI.enabled);
  sensitiveAI.group_hits = typeof sensitiveAI.group_hits === 'boolean' ? sensitiveAI.group_hits : true;
  sensitiveAI.concurrency = Number.isFinite(Number(sensitiveAI.concurrency)) ? Number(sensitiveAI.concurrency) : 4;
  sensitiveAI.max_hits_per_task = Number.isFinite(Number(sensitiveAI.max_hits_per_task)) ? Number(sensitiveAI.max_hits_per_task) : 80;
  sensitiveAI.context_chars = Number.isFinite(Number(sensitiveAI.context_chars)) ? Number(sensitiveAI.context_chars) : 12;
  sensitiveAI.retries = Number.isFinite(Number(sensitiveAI.retries)) ? Number(sensitiveAI.retries) : 1;
  sensitiveAI.temperature = Number.isFinite(Number(sensitiveAI.temperature)) ? Number(sensitiveAI.temperature) : 0.2;
  webSubmit.version_profiles = withNovelFetchVersionProfiles(webSubmit.version_profiles || webSubmit.profile_bindings);
  webSubmit.profile_bindings = { ...webSubmit.version_profiles };

  return { ...config, workflow, fetch, sensitive_ai: sensitiveAI, web_submit: webSubmit };
}
