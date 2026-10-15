const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { AccountStore } = require('../src/account-store');

test('missing account registry loads as an empty list', () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'yizhan-accounts-'));
  const store = new AccountStore({ filePath: path.join(dir, 'accounts.json') });
  assert.deepEqual(store.load(), []);
});

test('account registry persists only non-sensitive local account metadata', () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'yizhan-accounts-'));
  const filePath = path.join(dir, 'accounts.json');
  const store = new AccountStore({ filePath });
  store.save([
    { id: 'a1', name: '豆包账号 1', state: 'available', jobId: 'job-x', cookie: 'must-not-save', password: 'must-not-save' }
  ]);
  const text = fs.readFileSync(filePath, 'utf8');
  assert.equal(text.includes('must-not-save'), false);
  assert.deepEqual(store.load(), [{ id: 'a1', name: '豆包账号 1', state: 'available' }]);
});
