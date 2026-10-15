const test = require('node:test');
const assert = require('node:assert/strict');
const { DeviceStore } = require('../src/device-store');

function memoryFs() {
  const files = new Map();
  return {
    files,
    mkdirSync() {},
    writeFileSync(path, content) { files.set(path, String(content)); },
    renameSync(from, to) { files.set(to, files.get(from)); files.delete(from); },
    readFileSync(path) {
      if (!files.has(path)) { const error = new Error('ENOENT'); error.code = 'ENOENT'; throw error; }
      return files.get(path);
    }
  };
}

const safeStorage = {
  isEncryptionAvailable: () => true,
  encryptString: value => Buffer.from(`encrypted:${value}`),
  decryptString: buffer => buffer.toString().replace(/^encrypted:/, '')
};

test('saved device file never contains plaintext executor token', () => {
  const fs = memoryFs();
  const store = new DeviceStore({ safeStorage, fs, filePath: '/data/device.json' });
  store.save({ baseUrl: 'https://v78.example.com', executorId: 'lex1', token: 'super-secret-token' });
  const raw = fs.files.get('/data/device.json');
  assert.equal(raw.includes('super-secret-token'), false);
  assert.equal(store.load().token, 'super-secret-token');
});

test('save fails closed when OS encryption is unavailable', () => {
  const fs = memoryFs();
  const store = new DeviceStore({ safeStorage: { isEncryptionAvailable: () => false }, fs, filePath: '/data/device.json' });
  assert.throws(() => store.save({ baseUrl: 'x', executorId: 'y', token: 'z' }), /encryption/i);
  assert.equal(fs.files.size, 0);
});
