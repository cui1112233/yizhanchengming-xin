const test = require('node:test');
const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const { resolve } = require('node:path');

const root = resolve(__dirname, '..');
const read = path => readFileSync(resolve(root, path), 'utf8');

test('Agent 工作区由 React 源码实现并通过 Go 内嵌脚本发布', () => {
  const page = read('frontend/src/user/pages/AgentPageV2.jsx');
  const release = read('scripts/deploy-go-embedded.ps1');

  assert.match(page, /AgentWorkspaceHeader/);
  assert.match(page, /AgentCanvas/);
  assert.match(release, /同步前端产物到 Go embed 目录/);
  assert.match(release, /Invoke-Checked 'go' @\('build'/i);
  assert.doesNotMatch(release, /deploy\/v88-direct|node-stage|node-cutover/i);
});
