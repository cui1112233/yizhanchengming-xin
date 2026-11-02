import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const source = fs.readFileSync(path.join(root, 'backend/internal/httpapi/batch_factory_v12_intake.go'), 'utf8');

test('V12 project summary uses the lightweight cover index when the store provides it', () => {
  assert.match(source, /if lister, ok := store\.\(batchSummaryIndexLister\); ok \{/);
  assert.match(source, /batches, err = lister\.ListBatchSummaryIndex\(r\.Context\(\), owner\)/);
  assert.match(source, /"kind": "video", "url": batch\.ProjectCoverMediaURL/);
});
