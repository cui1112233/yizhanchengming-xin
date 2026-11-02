const supportedNodeTypes = new Set(['text', 'script', 'image', 'video', 'audio', 'director']);
const mediaNodeTypes = new Set(['image', 'video', 'audio']);
const nodeTypeLabels = { text: '文本', script: '剧本', image: '图片', video: '视频', audio: '音频', director: '导演台' };
const nodeStatusLabels = {
  draft: '草稿', pending: '等待中', queued: '排队中', running: '生成中', processing: '处理中',
  awaiting_confirmation: '等待选择', confirmed: '已确认', completed: '已完成', done: '已完成',
  failed: '需处理', canceled: '已取消', cancelled: '已取消'
};

export function normalizeCanvasNodeType(value) {
  const type = String(value || '').trim().toLowerCase();
  return supportedNodeTypes.has(type) ? type : 'text';
}

export function isMediaCanvasNodeType(value) {
  return mediaNodeTypes.has(normalizeCanvasNodeType(value));
}

export function canvasNodeTypeLabel(value) {
  return nodeTypeLabels[normalizeCanvasNodeType(value)];
}

export function canvasNodeStatusLabel(value) {
  const status = String(value || '').trim().toLowerCase();
  return nodeStatusLabels[status] || '待处理';
}

export function usesPinnedBatchFactoryComposer(node) {
  const config = node?.config && typeof node.config === 'object' ? node.config : {};
  return normalizeCanvasNodeType(node?.type) === 'script'
    && String(config.source || '').trim() === 'batch_factory'
    && String(config.resourceType || '').trim() === 'book_script';
}

export function canvasEditorFocusOptions(zoom = 1) {
  const currentZoom = Number.isFinite(Number(zoom)) ? Number(zoom) : 1;
  if (currentZoom >= 0.85) return null;
  return { padding: 0.16, duration: 260, minZoom: 0.75, maxZoom: 1 };
}
