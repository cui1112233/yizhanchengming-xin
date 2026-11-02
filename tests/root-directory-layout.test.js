const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const root = path.resolve(__dirname, '..');
const exists = (relative) => fs.existsSync(path.join(root, relative));

test('repository uses the documented project structure', () => {
  for (const directory of [
    'backend',
    'frontend',
    'services/121-browser-worker',
    'executors/local-executor',
    'executors/giant-material-executor',
    'scripts',
    'docs',
    'tests',
    'test',
    'data',
    'docs/PROJECT_FILE_STRUCTURE.md',
  ]) {
    assert.equal(exists(directory), true, `missing consolidated path: ${directory}`);
  }

  assert.equal(exists('deploy/v88-direct'), false, 'retired Node deployment path must not be restored');
  assert.equal(exists('server.js'), false, 'legacy root Node server must stay removed');
  assert.equal(exists('app.js'), false, 'legacy root Node app must stay removed');
});
