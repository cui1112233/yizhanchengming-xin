const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const root = path.resolve(__dirname, '..');
const exists = (relative) => fs.existsSync(path.join(root, relative));

test('repository uses the consolidated business directories', () => {
  for (const directory of [
    'backend',
    'frontend',
    'services/121-browser-worker',
    'executors/local-executor',
    'executors/giant-material-executor',
    'deploy',
    'scripts',
    'docs',
    'tests',
    'data',
  ]) {
    assert.equal(exists(directory), true, `missing consolidated path: ${directory}`);
  }

  assert.equal(exists('node_modules'), false, 'root Node dependencies must not be installed');
  assert.equal(exists('server.js'), false, 'legacy root Node server must stay removed');
  assert.equal(exists('app.js'), false, 'legacy root Node app must stay removed');
});
