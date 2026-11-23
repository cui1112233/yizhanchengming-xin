import assert from 'node:assert/strict';
import test from 'node:test';
import * as settings from './batchFactoryBookConfigRegions.js';

const { bookConfigRegionStatus, bookSensitiveModeUpdate, normalizeSensitiveMode, normalizeSensitiveSettings, SENSITIVE_MODE_OPTIONS } = settings;

test('sensitive settings use the three canonical modes and default to local processing', () => {
  assert.equal(typeof normalizeSensitiveMode, 'function');
  assert.deepEqual(SENSITIVE_MODE_OPTIONS, [
    { value: 'ai', label: 'AI优化' },
    { value: 'system', label: '系统优化' },
    { value: 'off', label: '关闭' }
  ]);
  assert.equal(normalizeSensitiveMode({}), 'system');
  assert.equal(normalizeSensitiveMode({ sensitiveMode: 'system', sensitiveAIEnabled: true }), 'system');
  assert.equal(normalizeSensitiveMode({ sensitiveMode: 'off', sensitiveAIEnabled: true }), 'off');
  assert.equal(normalizeSensitiveMode({ sensitiveMode: 'ai', sensitiveAIEnabled: false }), 'ai');
});

test('legacy sensitive AI support recognizes only an explicit true flag', () => {
  assert.equal(typeof normalizeSensitiveMode, 'function');
  assert.equal(normalizeSensitiveMode({ sensitiveAIEnabled: true }), 'ai');
  assert.equal(normalizeSensitiveMode({ sensitiveAIEnabled: false }), 'system');
  assert.equal(normalizeSensitiveMode({ sensitiveMode: 'unknown', sensitiveAIEnabled: true }), 'ai');
  assert.equal(normalizeSensitiveMode({ sensitiveMode: true, sensitiveAIEnabled: 'true' }), 'system');
});

test('loading or saving a preset keeps the legacy choice without forwarding its old flag', () => {
  assert.equal(typeof normalizeSensitiveSettings, 'function');
  const oldPreset = { sensitiveAIEnabled: true, textModelId: 'example' };
  assert.deepEqual(normalizeSensitiveSettings(oldPreset), { sensitiveMode: 'ai', textModelId: 'example' });
  assert.deepEqual(oldPreset, { sensitiveAIEnabled: true, textModelId: 'example' });
  assert.deepEqual(normalizeSensitiveSettings({ sensitiveAIEnabled: true, sensitiveMode: 'off' }), { sensitiveMode: 'off' });
});

test('saving a preset without sensitive settings explicitly writes the system default', () => {
  assert.deepEqual(normalizeSensitiveSettings({ textModelId: 'example' }), {
    textModelId: 'example', sensitiveMode: 'system'
  });
});

test('saving a preset with an invalid sensitive mode writes the effective valid mode', () => {
  assert.deepEqual(normalizeSensitiveSettings({ sensitiveMode: 'unknown' }), { sensitiveMode: 'system' });
  assert.deepEqual(normalizeSensitiveSettings({ sensitiveMode: 'unknown', sensitiveAIEnabled: true }), { sensitiveMode: 'ai' });
});

test('single-book sensitive settings save only the changed policy and restore inheritance sparsely', () => {
  assert.equal(typeof bookSensitiveModeUpdate, 'function');
  assert.deepEqual(bookSensitiveModeUpdate({}, {}, {}), { patch: {}, restoreKeys: [] });
  assert.deepEqual(bookSensitiveModeUpdate({}, {}, { sensitiveMode: 'off' }), {
    patch: { sensitiveMode: 'off' }, restoreKeys: []
  });
  assert.deepEqual(bookSensitiveModeUpdate({}, { sensitiveMode: 'off' }, { sensitiveMode: 'system' }), {
    patch: {}, restoreKeys: ['sensitiveMode']
  });
  assert.deepEqual(bookSensitiveModeUpdate({ sensitiveMode: 'off' }, {}, { sensitiveMode: 'system' }), {
    patch: { sensitiveMode: 'system' }, restoreKeys: []
  });
});

test('an unrelated book save preserves legacy fields and an explicit restore removes a conflicting legacy flag', () => {
  assert.equal(typeof bookSensitiveModeUpdate, 'function');
  const oldBook = { sensitiveMode: 'system', sensitiveAIEnabled: true, other: 'retained' };
  assert.deepEqual(bookSensitiveModeUpdate({}, oldBook, { ...oldBook, textModelId: 'new' }), {
    patch: {}, restoreKeys: []
  });
  assert.deepEqual(bookSensitiveModeUpdate({}, oldBook, {}, true), {
    patch: {}, restoreKeys: ['sensitiveMode', 'sensitiveAIEnabled']
  });
  assert.deepEqual(bookSensitiveModeUpdate({}, { sensitiveAIEnabled: true }, { sensitiveAIEnabled: true }), {
    patch: {}, restoreKeys: []
  });
});

test('single-book engine status accounts for sensitive policy and keeps absent settings inherited', () => {
  const book = patch => ({ settingsState: { patch } });
  assert.equal(bookConfigRegionStatus(book({ sensitiveMode: 'off' }), 'engine').tone, 'overridden');
  assert.equal(bookConfigRegionStatus(book({ sensitiveMode: 'off' }), 'engine', { sensitiveMode: 'off' }).tone, 'pushed');
  assert.equal(bookConfigRegionStatus(book({ sensitiveAIEnabled: true }), 'engine').tone, 'overridden');
  assert.equal(bookConfigRegionStatus(book({}), 'engine', { sensitiveMode: 'off' }).tone, 'inherited');
});
