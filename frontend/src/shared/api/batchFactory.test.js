import test from 'node:test';
import assert from 'node:assert/strict';
import { bf11Path, bindBatchFactoryAgentProject } from './batchFactory.js';

test('canonical Batch Factory facade uses the V12 namespace', () => {
  assert.equal(bf11Path('batches/demo'), '/api/batch-factory/v12/batches/demo');
  assert.match(String(bindBatchFactoryAgentProject), /\$\{BASE\}\/batches/);
});
