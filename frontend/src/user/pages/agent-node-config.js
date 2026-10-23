const IMAGE_DEFAULTS = Object.freeze({ ratio: '16:9', quality: 'standard', resolution: '2K', count: 1 });
const VIDEO_DEFAULTS = Object.freeze({ ratio: '16:9', resolution: '720P', durationSeconds: 5, count: 1, generateAudio: false });
const TEXT_DEFAULTS = Object.freeze({ operation: 'generate', length: 'standard', creativity: 'balanced' });
const AUDIO_DEFAULTS = Object.freeze({ operation: 'speech', speed: 1, pitch: 0, emotion: 'neutral', format: 'mp3' });

const DEFAULTS = { image: IMAGE_DEFAULTS, video: VIDEO_DEFAULTS, text: TEXT_DEFAULTS, audio: AUDIO_DEFAULTS };
const PARAMETER_KEYS = {
  image: Object.keys(IMAGE_DEFAULTS),
  video: Object.keys(VIDEO_DEFAULTS),
  text: Object.keys(TEXT_DEFAULTS),
  audio: [...Object.keys(AUDIO_DEFAULTS), 'startMs', 'endMs', 'segmentSeconds', 'ranges']
};

const LABELS = {
  quality: { low: '快速画质', standard: '标准画质', high: '高清画质' },
  operation: { generate: '生成', rewrite: '改写', continue: '续写', speech: '语音', music: '音乐', effect: '音效', sound_effect: '音效', trim: '裁剪', speed: '变速', smart_split: '智能分段', custom_split: '自定义分段' },
  length: { short: '精简', standard: '标准长度', long: '长篇' },
  creativity: { precise: '严谨', balanced: '均衡', creative: '创意' }
};

function kindOf(kind) {
  return Object.hasOwn(DEFAULTS, kind) ? kind : 'text';
}

function cloneList(value) {
  return Array.isArray(value) ? value.map(item => ({ ...item })) : [];
}

function normalizedParameters(kind, value = {}) {
  const type = kindOf(kind);
  const nested = value.parameters && typeof value.parameters === 'object' ? value.parameters : {};
  const parameters = { ...DEFAULTS[type] };
  for (const key of PARAMETER_KEYS[type]) {
    if (nested[key] !== undefined) parameters[key] = nested[key];
    else if (value[key] !== undefined) parameters[key] = value[key];
  }
  if (type === 'image') {
    if (!['16:9', '9:16', '1:1'].includes(parameters.ratio)) parameters.ratio = IMAGE_DEFAULTS.ratio;
    if (!['low', 'standard', 'high'].includes(parameters.quality)) parameters.quality = IMAGE_DEFAULTS.quality;
    if (!['1K', '2K', '4K'].includes(parameters.resolution)) parameters.resolution = IMAGE_DEFAULTS.resolution;
    parameters.count = [1, 2, 3, 4].includes(Number(parameters.count)) ? Number(parameters.count) : 1;
  } else if (type === 'video') {
    if (!['16:9', '9:16', '1:1'].includes(parameters.ratio)) parameters.ratio = VIDEO_DEFAULTS.ratio;
    parameters.resolution = String(parameters.resolution || VIDEO_DEFAULTS.resolution).toUpperCase();
    parameters.durationSeconds = Math.max(1, Number(parameters.durationSeconds) || VIDEO_DEFAULTS.durationSeconds);
    parameters.count = [1, 2, 3, 4].includes(Number(parameters.count)) ? Number(parameters.count) : 1;
    parameters.generateAudio = parameters.generateAudio === true;
  } else if (type === 'audio') {
    parameters.speed = Number(parameters.speed) > 0 ? Number(parameters.speed) : 1;
    parameters.pitch = Number(parameters.pitch) || 0;
    parameters.format = String(parameters.format || 'mp3').toLowerCase();
  }
  return parameters;
}

export function normalizeNodeConfig(kind, value = {}) {
  const type = kindOf(kind);
  const parameters = normalizedParameters(type, value);
  const modelId = String(value.modelId || value.model || '').trim();
  const config = {
    ...value,
    provider: String(value.provider || '').trim(),
    modelId,
    model: modelId,
    mode: String(value.mode || parameters.operation || '').trim(),
    prompt: String(value.prompt || ''),
    parameters,
    advanced: value.advanced && typeof value.advanced === 'object' ? { ...value.advanced } : {},
    manualReferences: cloneList(value.manualReferences),
    edgeReferences: cloneList(value.edgeReferences),
    referenceAssets: cloneList(value.referenceAssets),
    autoReferenceAssetIDs: Array.isArray(value.autoReferenceAssetIDs) ? [...value.autoReferenceAssetIDs] : []
  };
  for (const key of PARAMETER_KEYS[type]) config[key] = parameters[key];
  return config;
}

export function patchNodeConfig(kind, current = {}, patch = {}) {
  const type = kindOf(kind);
  const oldConfig = normalizeNodeConfig(type, current);
  const flatParameterPatch = {};
  for (const key of PARAMETER_KEYS[type]) {
    if (patch[key] !== undefined) flatParameterPatch[key] = patch[key];
  }
  return normalizeNodeConfig(type, {
    ...oldConfig,
    ...patch,
    parameters: { ...oldConfig.parameters, ...flatParameterPatch, ...(patch.parameters || {}) },
    advanced: { ...oldConfig.advanced, ...(patch.advanced || {}) }
  });
}

export function formatNodeParameterSummary(kind, value = {}) {
  const type = kindOf(kind);
  const parameters = normalizeNodeConfig(type, value).parameters;
  if (type === 'image') {
    return `${parameters.ratio} · ${LABELS.quality[parameters.quality] || parameters.quality} · ${parameters.resolution} · ${parameters.count}张`;
  }
  if (type === 'video') {
    return `${parameters.ratio} · ${parameters.resolution} · ${parameters.durationSeconds}s · ${parameters.count}个${parameters.generateAudio ? ' · 含声音' : ''}`;
  }
  if (type === 'audio') {
    return `${LABELS.operation[parameters.operation] || parameters.operation} · ${parameters.speed}x · ${String(parameters.format).toUpperCase()}`;
  }
  return `${LABELS.operation[parameters.operation] || parameters.operation} · ${LABELS.length[parameters.length] || parameters.length} · ${LABELS.creativity[parameters.creativity] || parameters.creativity}`;
}
