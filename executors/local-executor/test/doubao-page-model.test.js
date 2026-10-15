const test = require('node:test');
const assert = require('node:assert/strict');
const {
  classifyPageState,
  detectCapabilities,
  selectRequestedOptions
} = require('../src/doubao-page-model');

test('classifies login, human verification, quota and ready states explicitly', () => {
  assert.equal(classifyPageState({ visibleText: '请登录后继续使用 豆包' }), 'auth_required');
  assert.equal(classifyPageState({ visibleText: '安全验证 请完成验证后继续' }), 'human_verification');
  assert.equal(classifyPageState({ visibleText: '今日视频生成额度已用完' }), 'quota_exhausted');
  assert.equal(classifyPageState({ visibleText: 'Seedance 2.0 生成视频' }), 'available');
});

test('detects live page capabilities from semantic controls rather than CSS class names', () => {
  const capabilities = detectCapabilities({
    controls: [
      { role: 'button', text: 'Seedance 2.0' },
      { role: 'button', text: '5秒' },
      { role: 'button', text: '10秒' },
      { role: 'button', text: '9:16' },
      { role: 'button', text: '16:9' },
      { role: 'button', text: '生成视频' }
    ],
    promptInputs: [{ kind: 'contenteditable' }],
    fileInputs: [{ accept: 'image/png,image/jpeg' }]
  });
  assert.equal(capabilities.promptInput, true);
  assert.equal(capabilities.imageUpload, true);
  assert.deepEqual(capabilities.durations, [5, 10]);
  assert.deepEqual(capabilities.ratios, ['9:16', '16:9']);
  assert.equal(capabilities.submit, true);
  assert.equal(capabilities.videoGeneration, true);
  assert.deepEqual(capabilities.models, ['Seedance 2.0']);
});

test('ordinary Doubao chat input cannot be mistaken for video generation', () => {
  const capabilities = detectCapabilities({
    visibleText: '豆包 对话',
    controls: [{ role: 'button', text: '发送' }],
    promptInputs: [{ kind: 'contenteditable' }],
    fileInputs: []
  });
  assert.equal(capabilities.videoGeneration, false);
  assert.throws(
    () => selectRequestedOptions(capabilities, { prompt: '不要发到普通聊天框' }),
    error => error?.code === 'VIDEO_MODE_UNAVAILABLE'
  );
});

test('prompt-only jobs do not require image upload support', () => {
  const selected = selectRequestedOptions(
    { promptInput: true, imageUpload: false, durations: [5, 10], ratios: ['16:9'], models: ['Seedance 2.0'], submit: true },
    { prompt: '雨夜城市镜头缓缓推进', images: [], duration: 10, ratio: '16:9', model: 'Seedance 2.0' }
  );
  assert.equal(selected.duration, 10);
  assert.equal(selected.needsImageUpload, false);
});

test('image jobs fail instead of silently dropping unsupported references', () => {
  assert.throws(() => selectRequestedOptions(
    { promptInput: true, imageUpload: false, durations: [5, 10], ratios: ['16:9'], models: ['Seedance 2.0'], submit: true },
    { prompt: '人物转身', images: ['a.png'], duration: 5, ratio: '16:9', model: 'Seedance 2.0' }
  ), error => error?.code === 'REFERENCE_IMAGES_UNSUPPORTED');
});

test('unsupported duration ratio or model fails instead of changing the request', () => {
  const capabilities = { promptInput: true, imageUpload: true, durations: [5, 10], ratios: ['16:9'], models: ['Seedance 2.0'], submit: true };
  assert.throws(() => selectRequestedOptions(capabilities, { prompt: 'x', images: [], duration: 15 }), error => error?.code === 'DURATION_UNSUPPORTED');
  assert.throws(() => selectRequestedOptions(capabilities, { prompt: 'x', images: [], ratio: '9:16' }), error => error?.code === 'RATIO_UNSUPPORTED');
  assert.throws(() => selectRequestedOptions(capabilities, { prompt: 'x', images: [], model: 'Seedance Mini' }), error => error?.code === 'MODEL_UNSUPPORTED');
});
