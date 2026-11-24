const labels = {
  empty: '暂无文字页', pending: '待生成', rendering: '生成中', ready: '已就绪',
  failed: '生成失败', locked: '上传成品页', stale: '待重新生成',
};

// listTextPages returns page rows and artifact history separately. Only the active ID
// selects an image; later candidates and previously rendered versions are not current.
export function textPagePresentation(page, artifacts) {
  const id = page?.activeArtifactId;
  const artifact = artifacts?.find(item => item.id === id && item.pageId === page?.id && item.kind === 'page') ?? null;
  const hasImage = Boolean(id && (artifacts == null || artifact));
  let state = 'pending';
  if (!page) state = 'empty';
  else if (page.locked && hasImage) state = 'locked';
  else if (page.status === 'ready' && hasImage) state = 'ready';
  else if (page.status === 'failed' || page.status === 'rendering') state = page.status;
  else if (page.status === 'stale' || page.renderVersion > 0 || id || page.locked) state = 'stale';
  const ready = state === 'locked' || state === 'ready';
  return {
    state, label: labels[state], ready,
    artifactId: ready ? id : '', artifact: ready ? artifact : null,
    error: state === 'failed' ? page.lastError || labels.failed : '',
  };
}

export function textBookCompletion(pages = []) {
  const total = pages.length;
  const ready = pages.reduce((sum,page) => sum + Number(textPagePresentation(page).ready), 0);
  return { ready, total, percent: total ? Math.round(ready * 100 / total) : 0 };
}

// The work cover is edited and displayed separately. A collection is only the
// ordered set of ready text-page images, so cover readiness cannot block it.
export function canCreateTextCollection(pages = [], ids, document, artifacts) {
  if (!document?.activeVersionId || !Array.isArray(artifacts)) return false;
  const byArtifactId = new Map(artifacts.map(artifact => [artifact.id, artifact]));
  const byId = new Map(pages.map(page => [page.id, page]));
  const selected = ids ?? [...byId.keys()];
  if (!selected.length || new Set(selected).size !== selected.length) return false;
  return selected.every(id => {
    const page = byId.get(id);
    if (page?.status !== 'ready' || page.versionId !== document.activeVersionId
      || page.paginationVersion !== document.paginationVersion) return false;
    const image = byArtifactId.get(page.activeArtifactId);
    return image?.kind === 'page' && image.pageId === id && image.renderVersion === page.renderVersion
      && image.sourceKind === (page.locked ? 'user_upload' : 'system');
  });
}

// Compare editable properties, ignoring server-generated revisions and IDs.
// Accept the server page layout or the equivalent flattened settings form.
const geometryKeys = {
  paginationMode:'textPaginationMode', aspectRatio: 'imageAspectRatio', targetCharacters: 'textPageTarget', fontId: 'textFontId',
  fontSize: 'textFontSize', fontWeight: 'textFontWeight', lineHeight: 'textLineHeight',
  paragraphSpacing: 'textParagraphSpacing', padding: 'textPadding', maxHeightRatio: 'textMaxHeightRatio',
};
const paintKeys = { align: 'textAlign', textColor: 'textColor', highlightColor: 'textHighlightColor' };
function layoutChanged(before, after, keys) {
  return Object.entries(keys).some(([key, formKey]) => {
    const value = page => page.layout?.[key] ?? page[key] ?? page[formKey];
    return value(before) !== value(after);
  });
}

export function textPageDirtyImpact(before = {}, after = {}) {
  if (before.text !== after.text || before.contentRange !== after.contentRange || layoutChanged(before, after, geometryKeys))
    return 'reflow';
  const left = before.background ?? before.textBackground;
  const right = after.background ?? after.textBackground;
  if (layoutChanged(before, after, paintKeys) || before.highlightReuseEnabled !== after.highlightReuseEnabled
    || ['kind','color','id'].some(key => left?.[key] !== right?.[key])) return 'render';
  return 'none';
}
