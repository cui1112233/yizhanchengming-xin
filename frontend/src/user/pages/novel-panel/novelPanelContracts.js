function stringValue(...values) {
  for (const value of values) {
    if (typeof value === 'string' && value.trim()) return value.trim();
  }
  return '';
}

export function normalizeNovelProject(value) {
  const project = value?.project && typeof value.project === 'object' ? value.project : value;
  return {
    id: stringValue(project?.id, project?.project_id),
    name: stringValue(project?.name, project?.title),
    data: project?.data && typeof project.data === 'object' ? project.data : {},
    createdAt: stringValue(project?.createdAt, project?.created_at),
    updatedAt: stringValue(project?.updatedAt, project?.updated_at)
  };
}

export function normalizeNovelAsset(value) {
  const asset = value?.asset && typeof value.asset === 'object' ? value.asset : value || {};
  const revision = Number(asset.revision ?? asset.version ?? 0);
  return {
    assetType: stringValue(asset.assetType, asset.asset_type, asset.type),
    assetId: stringValue(asset.assetId, asset.asset_id, asset.id),
    name: stringValue(asset.name, asset.displayName, asset.display_name, asset.title),
    imageUrl: stringValue(asset.imageUrl, asset.image_url, asset.mainUrl, asset.main_url, asset.url),
    revision: Number.isFinite(revision) ? revision : 0
  };
}

export function normalizeNovelPrompt(value) {
  const prompt = value?.prompt && typeof value.prompt === 'object' ? value.prompt : value || {};
  return {
    text: stringValue(prompt.text, prompt.promptText, prompt.prompt_text, prompt.content),
    version: stringValue(prompt.version, prompt.promptVersion, prompt.prompt_version)
  };
}

export function normalizeNovelPanelError(value) {
  const source = value?.error && typeof value.error === 'object' ? value.error : value || {};
  const code = Number(source.code ?? source.status ?? 500);
  return {
    code: Number.isFinite(code) ? code : 500,
    message: stringValue(source.message, source.error, source.detail) || '小说面板请求失败',
    retryable: source.retryable === true || code >= 500
  };
}
