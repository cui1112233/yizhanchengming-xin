const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

test('Node app does not register the retired batch factory runtime', () => {
  const source = fs.readFileSync(path.join(__dirname, '..', 'app.js'), 'utf8');
  assert.doesNotMatch(source, /createBatchFactoryV11Router/);
  assert.doesNotMatch(source, /createBatchFactoryV12Router/);
  assert.doesNotMatch(source, /createBatchFactoryV11Scheduler/);
});
