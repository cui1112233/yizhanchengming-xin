const test = require('node:test');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { LocalNovelBodyStore } = require('../src/novel-body-store');

function sha256(text) {
  return crypto.createHash('sha256').update(Buffer.from(text, 'utf8')).digest('hex');
}

function walkFiles(root) {
  const out = [];
  for (const entry of fs.readdirSync(root, { withFileTypes: true })) {
    const full = path.join(root, entry.name);
    if (entry.isDirectory()) out.push(...walkFiles(full));
    else out.push(full);
  }
  return out;
}

test('local novel bodies are account isolated, hashed, revisioned and never exported as txt implicitly', t => {
  const rootDir = fs.mkdtempSync(path.join(os.tmpdir(), 'yizhan-novel-bodies-'));
  t.after(() => fs.rmSync(rootDir, { recursive: true, force: true }));
  const store = new LocalNovelBodyStore({ rootDir });

  const aliceV1 = store.put({
    owner: 'alice@example.com',
    bookId: '123',
    versionId: 'ai3',
    content: '第一版正文🙂',
    sourceRevision: 7
  });
  const bobV1 = store.put({
    owner: 'bob@example.com',
    bookId: '123',
    versionId: 'ai3',
    content: 'Bob 的正文',
    sourceRevision: 2
  });

  assert.equal(aliceV1.revision, 1);
  assert.equal(aliceV1.sourceRevision, 7);
  assert.equal(aliceV1.contentHash, sha256('第一版正文🙂'));
  assert.equal(aliceV1.charCount, [...'第一版正文🙂'].length);
  assert.equal(bobV1.revision, 1);
  assert.equal(store.get({ owner: 'alice@example.com', bookId: '123', versionId: 'ai3' }).content, '第一版正文🙂');
  assert.equal(store.get({ owner: 'bob@example.com', bookId: '123', versionId: 'ai3' }).content, 'Bob 的正文');

  const aliceV2 = store.put({
    owner: 'alice@example.com',
    bookId: '123',
    versionId: 'ai3',
    content: '第二版正文',
    sourceRevision: 8
  });
  assert.equal(aliceV2.revision, 2);
  assert.equal(aliceV2.sourceRevision, 8);
  assert.equal(store.get({ owner: 'alice@example.com', bookId: '123', versionId: 'ai3' }).content, '第二版正文');
  assert.equal(store.get({ owner: 'bob@example.com', bookId: '123', versionId: 'ai3' }).content, 'Bob 的正文');

  const files = walkFiles(rootDir);
  assert.ok(files.some(file => file.endsWith('.body.gz')));
  assert.ok(files.some(file => file.endsWith('.meta.json')));
  assert.equal(files.some(file => file.endsWith('.txt')), false);
  assert.equal(files.some(file => file.endsWith('.tmp')), false);
  assert.equal(files.some(file => file.includes('alice@example.com') || file.includes('bob@example.com')), false);
});

test('local novel body store rejects unsafe identifiers and preserves existing body on invalid writes', t => {
  const rootDir = fs.mkdtempSync(path.join(os.tmpdir(), 'yizhan-novel-bodies-'));
  t.after(() => fs.rmSync(rootDir, { recursive: true, force: true }));
  const store = new LocalNovelBodyStore({ rootDir });

  store.put({ owner: 'alice', bookId: 'safe-book', versionId: 'ai1', content: 'safe', sourceRevision: 1 });
  assert.throws(() => store.put({ owner: 'alice', bookId: '../escape', versionId: 'ai1', content: 'bad' }), /book/i);
  assert.throws(() => store.put({ owner: 'alice', bookId: 'safe-book', versionId: '../ai1', content: 'bad' }), /version/i);
  assert.equal(store.get({ owner: 'alice', bookId: 'safe-book', versionId: 'ai1' }).content, 'safe');
});
