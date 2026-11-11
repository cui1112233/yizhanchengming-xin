const adjustableQuality = { resolution: '1080p', fps: 30, locked: false };

export const aiSegmentModeContract = Object.freeze({
  subtitle: Object.freeze({
    label: '按字幕切割',
    hidden: true,
    sections: Object.freeze(['video', 'srt', 'output', 'duration', 'quality', 'progress', 'log']),
    quality: Object.freeze({ ...adjustableQuality }),
    primaryAction: '分析切点',
    exportAction: '开始分段'
  }),
  'main-track': Object.freeze({
    label: '按剪映主轨道',
    sections: Object.freeze(['video', 'draft', 'output', 'duration', 'quality', 'progress', 'log']),
    quality: Object.freeze({ ...adjustableQuality }),
    primaryAction: '分析切点',
    exportAction: '开始分段'
  }),
  'text-track': Object.freeze({
    label: '按剪映文本轨道',
    sections: Object.freeze(['video', 'draft', 'text-track', 'output', 'quality', 'progress', 'log']),
    quality: Object.freeze({ ...adjustableQuality }),
    primaryAction: '分析切点',
    exportAction: '开始分段'
  }),
  ai: Object.freeze({
    label: 'AI分段',
    sections: Object.freeze(['model', 'prompt', 'video', 'draft', 'script', 'analysis', 'output', 'quality', 'progress', 'log']),
    quality: Object.freeze({ resolution: '1080p', fps: 30, locked: true }),
    primaryAction: 'AI 分析切点',
    exportAction: '开始按AI切点分段'
  })
});

export const defaultAISegmentMode = 'main-track';

export const visibleAISegmentModes = Object.freeze(
  Object.entries(aiSegmentModeContract)
    .filter(([, item]) => !item.hidden)
    .map(([value, item]) => Object.freeze({ value, label: item.label }))
);

export function aiSegmentModeHas(mode, section) {
  return aiSegmentModeContract[mode]?.sections.includes(section) === true;
}
