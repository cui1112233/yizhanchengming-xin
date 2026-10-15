import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = path.dirname(fileURLToPath(import.meta.url));
const html = fs.readFileSync(path.join(root, 'index.html'), 'utf8');
const app = fs.readFileSync(path.join(root, 'app.js'), 'utf8');

test('小说获取不再暴露会互相覆盖的局部保留设置', () => {
  assert.doesNotMatch(html, /storageCleanupEnabled|storageRetentionDays/);
});

test('小说获取不再读写局部 storage.retention_days', () => {
  assert.doesNotMatch(app, /storage\.cleanup_enabled|storage\.retention_days|storageCleanupEnabled|storageRetentionDays/);
});
