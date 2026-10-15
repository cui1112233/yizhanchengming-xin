const test = require('node:test');
const assert = require('node:assert/strict');
const { validateMp4Buffer } = require('../src/mp4-validator');

function mp4(size = 1024) {
  const buffer = Buffer.alloc(size);
  buffer.writeUInt32BE(24, 0);
  buffer.write('ftyp', 4, 'ascii');
  buffer.write('isom', 8, 'ascii');
  return buffer;
}

test('accepts a plausible MP4 with an ftyp box', () => {
  assert.deepEqual(validateMp4Buffer(mp4(), { minBytes: 100 }), { ok: true, size: 1024, brand: 'isom' });
});

test('rejects html/error responses even if the filename would be mp4', () => {
  assert.throws(() => validateMp4Buffer(Buffer.from('<html>login expired</html>'), { minBytes: 1 }), error => error?.code === 'INVALID_MP4_HEADER');
});

test('rejects suspiciously small downloads', () => {
  assert.throws(() => validateMp4Buffer(mp4(32), { minBytes: 100 }), error => error?.code === 'MP4_TOO_SMALL');
});
