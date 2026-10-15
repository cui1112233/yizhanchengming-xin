const test = require('node:test');
const assert = require('node:assert/strict');
const { normalizeVideoInput } = require('../src/video-input');

test('prompt only is valid video input', () => {
  assert.deepEqual(normalizeVideoInput({ prompt: '夜晚城市追逐，电影感' }), {
    prompt: '夜晚城市追逐，电影感',
    images: []
  });
});

test('prompt plus images keeps image references', () => {
  assert.deepEqual(normalizeVideoInput({ prompt: '人物回头', images: [{ url: 'https://example.com/a.png' }] }), {
    prompt: '人物回头',
    images: [{ url: 'https://example.com/a.png' }]
  });
});

test('image is optional but prompt is required', () => {
  assert.throws(() => normalizeVideoInput({ images: [{ url: 'https://example.com/a.png' }] }), /video prompt is required/);
});
