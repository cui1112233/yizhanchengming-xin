import test from 'node:test';
import assert from 'node:assert/strict';
import {
  appendImageCandidate,
  appendImageCandidateAndSelectSingle,
  buildReferenceAssetGenerationPayload
} from './scriptEntityImages.js';

test('builds a character generation request from the current editor fields', () => {
  const payload = buildReferenceAssetGenerationPayload({
    type: 'characters',
    entity: { id: 'char-lin' },
    fields: { 角色名称: '林溪', 身份: '记者', 外形: '短发，深色风衣', imageUrls: ['https://old.example/a.png'] },
    novelText: '林溪在雨夜赶到车站。',
    extractionPreset: 'modern'
  });

  assert.equal(payload.asset_type, 'character');
  assert.equal(payload.asset_id, 'char-lin');
  assert.equal(payload.character.name, '林溪');
  assert.match(payload.description, /记者/);
  assert.match(payload.description, /短发/);
  assert.match(payload.context, /雨夜/);
  assert.match(payload.generation_guidance, /人物/);
  assert.doesNotMatch(JSON.stringify(payload), /api[_-]?key|token|secret/i);
});

test('builds a scene generation request without inventing a character payload', () => {
  const payload = buildReferenceAssetGenerationPayload({
    type: 'scenes',
    entity: { id: 'scene-station' },
    fields: { 场景名称: '旧车站', 时段: '夜晚', 氛围: '潮湿压抑' },
    novelText: '旧车站的雨幕遮住了远处的灯。'
  });

  assert.equal(payload.asset_type, 'scene');
  assert.equal(payload.asset_id, 'scene-station');
  assert.equal(payload.character, undefined);
  assert.match(payload.description, /旧车站/);
  assert.match(payload.generation_guidance, /场景/);
});

test('appending a generated result keeps explicit main-image selection untouched', () => {
  const current = ['https://img.example/existing.png'];
  const next = appendImageCandidate(current, '/api/reference-assets/file/character/char-lin/main');

  assert.deepEqual(next, [
    'https://img.example/existing.png',
    '/api/reference-assets/file/character/char-lin/main'
  ]);
  assert.deepEqual(appendImageCandidate(next, next[1]), next);
});

test('the first image is automatically selected as the main image', () => {
  assert.deepEqual(
    appendImageCandidateAndSelectSingle([], '', 'https://img.example/only.png'),
    {
      imageUrls: ['https://img.example/only.png'],
      mainImageUrl: 'https://img.example/only.png'
    }
  );
});

test('adding another image keeps the existing main image', () => {
  assert.deepEqual(
    appendImageCandidateAndSelectSingle(
      ['https://img.example/main.png'],
      'https://img.example/main.png',
      'https://img.example/second.png'
    ),
    {
      imageUrls: ['https://img.example/main.png', 'https://img.example/second.png'],
      mainImageUrl: 'https://img.example/main.png'
    }
  );
});

test('multiple images without a selection remain unselected', () => {
  assert.deepEqual(
    appendImageCandidateAndSelectSingle(
      ['https://img.example/first.png', 'https://img.example/second.png'],
      '',
      'https://img.example/third.png'
    ),
    {
      imageUrls: [
        'https://img.example/first.png',
        'https://img.example/second.png',
        'https://img.example/third.png'
      ],
      mainImageUrl: ''
    }
  );
});
