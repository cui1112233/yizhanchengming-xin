const PROJECT_TYPES = new Set(['video', 'comic']);

export function normalizeAgentProjectType(value) {
  const normalized = String(value || '').trim().toLowerCase();
  return PROJECT_TYPES.has(normalized) ? normalized : 'video';
}

export function agentProjectTypeLabel(value) {
  return normalizeAgentProjectType(value) === 'comic' ? '漫画作品' : '视频作品';
}

export function agentProjectCapabilities(value) {
  if (normalizeAgentProjectType(value) === 'comic') {
    return {
      nodeTypes: ['script', 'text', 'image'],
      modelKinds: ['text', 'image'],
      allowVideoGeneration: false,
      allowAudio: false
    };
  }
  return {
    nodeTypes: ['script', 'text', 'image', 'video', 'audio', 'director'],
    modelKinds: ['text', 'image', 'video', 'audio'],
    allowVideoGeneration: true,
    allowAudio: true
  };
}

export function agentProjectTypeOptions() {
  return [
    { value: 'video', label: '视频作品', description: '剧本、图片、视频、音频完整制作链路' },
    { value: 'comic', label: '漫画作品', description: '剧本、文本、图片和漫画提示词生成链路' }
  ];
}
