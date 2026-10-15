const { normalizeVideoInput } = require('./video-input');

class DoubaoCapabilityError extends Error {
  constructor(code, message) {
    super(message);
    this.name = 'DoubaoCapabilityError';
    this.code = code;
  }
}

function classifyPageState(snapshot = {}) {
  const text = normalizeText([
    snapshot.visibleText,
    ...(snapshot.controls || []).map(control => `${control.text || ''} ${control.aria || ''}`)
  ].join(' '));

  if (/安全验证|人机验证|请完成验证|完成验证后继续|拖动.{0,8}滑块|验证后继续/.test(text)) {
    return 'human_verification';
  }
  if (/今日.{0,12}(视频|生成).{0,12}(额度|次数).{0,8}(已用完|用尽|不足)|已达.{0,12}(今日|每日).{0,8}(上限|限制)|今日可用.{0,6}0/.test(text)) {
    return 'quota_exhausted';
  }
  if (/请登录|登录后继续|扫码登录|手机号登录|登录\s*\/\s*注册/.test(text)) {
    return 'auth_required';
  }
  return 'available';
}

function detectCapabilities(snapshot = {}) {
  const controls = Array.isArray(snapshot.controls) ? snapshot.controls : [];
  const promptInputs = Array.isArray(snapshot.promptInputs) ? snapshot.promptInputs : [];
  const fileInputs = Array.isArray(snapshot.fileInputs) ? snapshot.fileInputs : [];
  const texts = controls.map(control => String(control.text || control.aria || '').trim()).filter(Boolean);
  const fullText = normalizeText([snapshot.visibleText || '', ...texts].join(' '));

  const durations = uniqueSortedNumbers(texts.flatMap(extractDurations));
  const ratios = unique(texts.flatMap(extractRatios));
  const models = unique(texts.filter(text => /seedance/i.test(text)).map(cleanModelText));
  const imageUpload = fileInputs.some(input => /image/i.test(String(input.accept || '')))
    || texts.some(text => /参考图|参考图片|上传图片|添加图片/.test(text));
  const submit = texts.some(text => /^(生成|发送|开始制作|立即生成|生成视频|开始生成)$/.test(normalizeText(text)));
  const videoGeneration = models.length > 0
    || durations.length > 0
    || ratios.length > 0
    || /Seedance|视频生成|生成视频|参考图.{0,8}(视频|生成)/i.test(fullText);

  return {
    promptInput: promptInputs.length > 0,
    imageUpload,
    durations,
    ratios,
    models,
    submit,
    videoGeneration
  };
}

function selectRequestedOptions(capabilities = {}, payload = {}) {
  const input = normalizeVideoInput(payload);
  const videoGeneration = capabilities.videoGeneration === true
    || (Array.isArray(capabilities.models) && capabilities.models.length > 0)
    || (Array.isArray(capabilities.durations) && capabilities.durations.length > 0)
    || (Array.isArray(capabilities.ratios) && capabilities.ratios.length > 0);
  if (!videoGeneration) {
    throw new DoubaoCapabilityError('VIDEO_MODE_UNAVAILABLE', 'Doubao is not confirmed to be in video generation mode');
  }
  if (!capabilities.promptInput) {
    throw new DoubaoCapabilityError('PROMPT_INPUT_UNAVAILABLE', 'Doubao prompt input is unavailable');
  }
  if (!capabilities.submit) {
    throw new DoubaoCapabilityError('SUBMIT_UNAVAILABLE', 'Doubao video submit control is unavailable');
  }
  if (input.images.length > 0 && !capabilities.imageUpload) {
    throw new DoubaoCapabilityError('REFERENCE_IMAGES_UNSUPPORTED', 'Doubao page does not currently expose reference image upload');
  }

  const duration = normalizeDuration(payload.duration);
  if (duration !== null && !includesNumber(capabilities.durations, duration)) {
    throw new DoubaoCapabilityError('DURATION_UNSUPPORTED', `requested duration ${duration}s is not available on the live Doubao page`);
  }

  const ratio = normalizeOptionalString(payload.ratio || payload.aspectRatio);
  if (ratio && !includesText(capabilities.ratios, ratio)) {
    throw new DoubaoCapabilityError('RATIO_UNSUPPORTED', `requested ratio ${ratio} is not available on the live Doubao page`);
  }

  const model = normalizeOptionalString(payload.model);
  if (model && !includesText(capabilities.models, model)) {
    throw new DoubaoCapabilityError('MODEL_UNSUPPORTED', `requested model ${model} is not available on the live Doubao page`);
  }

  return {
    prompt: input.prompt,
    images: input.images,
    needsImageUpload: input.images.length > 0,
    duration,
    ratio,
    model
  };
}

function extractDurations(text) {
  const values = [];
  const re = /(?:^|\D)(\d{1,2})\s*秒/g;
  let match;
  while ((match = re.exec(text))) values.push(Number(match[1]));
  return values;
}

function extractRatios(text) {
  const matches = String(text).match(/\b\d{1,2}:\d{1,2}\b/g);
  return matches || [];
}

function cleanModelText(text) {
  const match = String(text).match(/Seedance\s*[\w.-]+(?:\s*[\w.-]+)?/i);
  return (match?.[0] || String(text)).replace(/\s+/g, ' ').trim();
}

function normalizeDuration(value) {
  if (value === undefined || value === null || value === '' || value === 'auto') return null;
  const number = Number(value);
  if (!Number.isFinite(number) || number <= 0) {
    throw new DoubaoCapabilityError('DURATION_INVALID', 'video duration is invalid');
  }
  return number;
}

function normalizeOptionalString(value) {
  const text = String(value || '').trim();
  return text || null;
}

function normalizeText(value) {
  return String(value || '').replace(/\s+/g, ' ').trim();
}

function includesNumber(values, wanted) {
  return (Array.isArray(values) ? values : []).some(value => Number(value) === wanted);
}

function includesText(values, wanted) {
  const needle = normalizeText(wanted).toLowerCase();
  return (Array.isArray(values) ? values : []).some(value => normalizeText(value).toLowerCase() === needle);
}

function unique(values) {
  return [...new Set(values.filter(Boolean))];
}

function uniqueSortedNumbers(values) {
  return [...new Set(values.filter(Number.isFinite))].sort((a, b) => a - b);
}

module.exports = {
  DoubaoCapabilityError,
  classifyPageState,
  detectCapabilities,
  selectRequestedOptions,
  extractDurations,
  extractRatios
};
